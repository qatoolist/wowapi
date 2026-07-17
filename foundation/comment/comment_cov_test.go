package comment_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/comment"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/testkit"
)

// fixedIDGen always mints the same id. Used to force a duplicate primary-key
// insert so the Create error-wrap branch runs against a real DB constraint.
type fixedIDGen struct{ id uuid.UUID }

func (g fixedIDGen) New() uuid.UUID { return g.id }

// fixture spins up a tenant with a capacity-bound actor context and returns the
// pieces every comment test needs.
type fixture struct {
	h     *testkit.DBHandle
	tnID  uuid.UUID
	ctx   context.Context
	capID uuid.UUID
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tn.ID, userID)
	ctx := database.WithActorID(testkit.TenantCtx(tn.ID), capID)
	return fixture{h: h, tnID: tn.ID, ctx: ctx, capID: capID}
}

// -----------------------------------------------------------------------------
// Create validation error paths
// -----------------------------------------------------------------------------

func TestCreateResourceRequired(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Create(ctx, db, comment.CreateIn{
			AuthorCapacityID: f.capID,
			Body:             "no resource",
		})
		return e
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("missing resource must be KindValidation, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestCreateBodyRequired(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Create(ctx, db, comment.CreateIn{
			Resource:         ref,
			AuthorCapacityID: f.capID,
			Body:             "   \t\n ", // whitespace only
		})
		return e
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("blank body must be KindValidation, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestCreateAuthorRequired(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Create(ctx, db, comment.CreateIn{
			Resource: ref,
			Body:     "has body but no author",
		})
		return e
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("missing author must be KindValidation, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestCreateParentNotFound(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	ghost := uuid.New()
	err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Create(ctx, db, comment.CreateIn{
			Resource:         ref,
			ParentID:         &ghost,
			AuthorCapacityID: f.capID,
			Body:             "reply to a comment that does not exist",
		})
		return e
	})
	if errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("unknown parent must be KindNotFound, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

// -----------------------------------------------------------------------------
// Threading (valid reply) + outbox emission on Create
// -----------------------------------------------------------------------------

func TestReplyThreadingAndCreateOutbox(t *testing.T) {
	f := newFixture(t)
	// Real outbox writer so create emits events into events_outbox in-tx.
	svc := comment.New(model.UUIDv7(), outbox.NewWriter(model.UUIDv7()))
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}

	var parentID, replyID uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		parentID, e = svc.Create(ctx, db, comment.CreateIn{
			Resource: ref, AuthorCapacityID: f.capID, Body: "root",
		})
		if e != nil {
			return e
		}
		replyID, e = svc.Create(ctx, db, comment.CreateIn{
			Resource: ref, ParentID: &parentID, AuthorCapacityID: f.capID, Body: "reply",
		})
		return e
	}); err != nil {
		t.Fatalf("create parent+reply: %v", err)
	}

	// List must show the reply threaded under the parent.
	var cmts []comment.Comment
	if err := f.h.TxM.WithTenantRO(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cmts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(cmts) != 2 {
		t.Fatalf("want 2 comments, got %d", len(cmts))
	}
	var reply *comment.Comment
	for i := range cmts {
		if cmts[i].ID == replyID {
			reply = &cmts[i]
		}
	}
	if reply == nil {
		t.Fatal("reply not returned by List")
	}
	if reply.ParentID == nil {
		t.Fatal("reply.ParentID must be set (threading)")
	}
	if *reply.ParentID != parentID {
		t.Fatalf("reply.ParentID = %s, want %s", *reply.ParentID, parentID)
	}

	// Two comment.created events must have landed in the outbox.
	types := outboxTypesFor(t, f.h, f.ctx, ref)
	created := 0
	for _, ty := range types {
		if ty == "comment.created" {
			created++
		}
	}
	if created != 2 {
		t.Fatalf("want 2 comment.created events, got %d (all=%v)", created, types)
	}
}

// -----------------------------------------------------------------------------
// Create insert failure (duplicate PK) — real DB constraint hits the wrap path.
// -----------------------------------------------------------------------------

func TestCreateDuplicateIDInsertError(t *testing.T) {
	f := newFixture(t)
	dup := uuid.New()
	svc := comment.New(fixedIDGen{id: dup}, nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}

	var firstID uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		firstID, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "first"})
		return e
	}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if firstID != dup {
		t.Fatalf("first id = %s, want fixed %s", firstID, dup)
	}

	// Second create mints the same id → primary-key violation → wrapped error.
	var innerErr error
	_ = f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, innerErr = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "dup"})
		return innerErr
	})
	if innerErr == nil {
		t.Fatal("duplicate id insert must fail")
	}
	if errors.KindOf(innerErr) != errors.KindInternal {
		t.Fatalf("duplicate insert should wrap as KindInternal, got kind=%v err=%v", errors.KindOf(innerErr), innerErr)
	}
}

// -----------------------------------------------------------------------------
// Edit error paths + outbox emission
// -----------------------------------------------------------------------------

func TestEditBodyRequired(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	var cid uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "v1"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Edit(ctx, db, cid, 1, "   ")
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("blank edit body must be KindValidation, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestEditNotFound(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Edit(ctx, db, uuid.New(), 1, "body")
	})
	if errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("editing missing comment must be KindNotFound, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestEditEmitsOutboxEvent(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), outbox.NewWriter(model.UUIDv7()))
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}

	var cid uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "before"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Edit(ctx, db, cid, 1, "after")
	}); err != nil {
		t.Fatalf("edit: %v", err)
	}

	// Assert the edited event carries the before/after bodies for audit history.
	var editedPayload string
	if err := f.h.TxM.WithTenantRO(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT payload::text FROM events_outbox
			  WHERE resource_type=$1 AND resource_id=$2 AND event_type='comment.edited'`,
			ref.Type, ref.ID).Scan(&editedPayload)
	}); err != nil {
		t.Fatalf("load edited event: %v", err)
	}
	if !strings.Contains(editedPayload, "before") || !strings.Contains(editedPayload, "after") {
		t.Fatalf("edited event payload must carry previous+new body, got %q", editedPayload)
	}
	if !strings.Contains(editedPayload, cid.String()) {
		t.Fatalf("edited event payload must carry comment_id %s, got %q", cid, editedPayload)
	}
}

// TestEditLoadCancelledContext forces a non-NoRows load error (context cancelled)
// so the load error-wrap branch runs.
func TestEditLoadCancelledContext(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	var cid uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "v1"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	var innerErr error
	_ = f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		innerErr = svc.Edit(cctx, db, cid, 1, "body")
		return nil
	})
	if innerErr == nil {
		t.Fatal("edit under cancelled context must fail")
	}
	if errors.KindOf(innerErr) != errors.KindInternal {
		t.Fatalf("cancelled load should wrap as KindInternal, got kind=%v err=%v", errors.KindOf(innerErr), innerErr)
	}
}

// -----------------------------------------------------------------------------
// Void error paths
// -----------------------------------------------------------------------------

func TestVoidNotFound(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Void(ctx, db, uuid.New(), 1)
	})
	if errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("voiding missing comment must be KindNotFound, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestVoidByNonAuthorForbidden(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	var cid uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "mine"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	otherCtx := database.WithActorID(testkit.TenantCtx(f.tnID), uuid.New())
	err := f.h.TxM.WithTenant(otherCtx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Void(ctx, db, cid, 1)
	})
	if errors.KindOf(err) != errors.KindForbidden {
		t.Fatalf("non-author void must be KindForbidden, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestVoidStaleVersionConflict(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	var cid uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "v1"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	// Author voids with a stale expected version → optimistic-lock conflict.
	err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Void(ctx, db, cid, 99)
	})
	if errors.KindOf(err) != errors.KindVersionConflict {
		t.Fatalf("stale void must be KindVersionConflict, got kind=%v err=%v", errors.KindOf(err), err)
	}

	// The comment must remain active (the failed void changed nothing).
	var cmts []comment.Comment
	if err := f.h.TxM.WithTenantRO(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cmts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatal(err)
	}
	if len(cmts) != 1 || cmts[0].Status != "active" {
		t.Fatalf("comment must stay active after failed void, got %+v", cmts)
	}
}

func TestVoidLoadCancelledContext(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	var cid uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "v1"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	var innerErr error
	_ = f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		innerErr = svc.Void(cctx, db, cid, 1)
		return nil
	})
	if innerErr == nil {
		t.Fatal("void under cancelled context must fail")
	}
	if errors.KindOf(innerErr) != errors.KindInternal {
		t.Fatalf("cancelled load should wrap as KindInternal, got kind=%v err=%v", errors.KindOf(innerErr), innerErr)
	}
}

// TestCreateParentLoadCancelledContext exercises the parent-load error-wrap
// branch in Create (a non-NoRows failure while resolving the parent).
func TestCreateParentLoadCancelledContext(t *testing.T) {
	f := newFixture(t)
	svc := comment.New(model.UUIDv7(), nil)
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	var parentID uuid.UUID
	if err := f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		parentID, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: f.capID, Body: "root"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	var innerErr error
	_ = f.h.TxM.WithTenant(f.ctx, func(ctx context.Context, db database.TenantDB) error {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		_, innerErr = svc.Create(cctx, db, comment.CreateIn{
			Resource: ref, ParentID: &parentID, AuthorCapacityID: f.capID, Body: "reply",
		})
		return nil
	})
	if innerErr == nil {
		t.Fatal("create with cancelled parent-load must fail")
	}
	if errors.KindOf(innerErr) != errors.KindInternal {
		t.Fatalf("cancelled parent-load should wrap as KindInternal, got kind=%v err=%v", errors.KindOf(innerErr), innerErr)
	}
}

// outboxTypesFor returns the event_type of every outbox row anchored to ref.
func outboxTypesFor(t *testing.T, h *testkit.DBHandle, ctx context.Context, ref resource.Ref) []string {
	t.Helper()
	var types []string
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		rows, err := db.Query(ctx,
			`SELECT event_type FROM events_outbox WHERE resource_type=$1 AND resource_id=$2 ORDER BY occurred_at ASC`,
			ref.Type, ref.ID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var ty string
			if err := rows.Scan(&ty); err != nil {
				return err
			}
			types = append(types, ty)
		}
		return rows.Err()
	}); err != nil {
		t.Fatalf("query outbox: %v", err)
	}
	return types
}
