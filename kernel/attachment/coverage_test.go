package attachment_test

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/attachment"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/testkit"
)

// failingOutboxWriter is a fault-injecting outbox.Writer double: every Write
// call returns err. Used to prove Attach propagates an outbox-write failure
// instead of discarding it (DATA-08 W0-T1).
type failingOutboxWriter struct{ err error }

func (f failingOutboxWriter) Write(context.Context, database.TenantDB, outbox.Event) error {
	return f.err
}

// seedComment inserts a comments row via Admin (bypasses RLS) so it can satisfy
// the attachments.comment_id FK, and returns the new comment id.
func seedComment(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID, ref resource.Ref) uuid.UUID {
	t.Helper()
	cID := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO comments (id, tenant_id, resource_type, resource_id, author_capacity_id, body, created_by)
		 VALUES ($1, $2, $3, $4, $5, 'test comment', $6)`,
		cID, tenantID, ref.Type, ref.ID, uuid.Nil, uuid.Nil); err != nil {
		t.Fatalf("seed comment: %v", err)
	}
	return cID
}

// TestAttachValidationErrors covers the two input-validation guards in Attach:
// a zero resource ref and a nil document version id both fail closed with
// KindValidation before any DB work.
func TestAttachValidationErrors(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()

	t.Run("zero resource ref", func(t *testing.T) {
		err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			_, e := svc.Attach(ctx, db, attachment.AttachIn{
				Resource:          resource.Ref{}, // zero
				DocumentVersionID: uuid.New(),
			})
			return e
		})
		if errors.KindOf(err) != errors.KindValidation {
			t.Fatalf("zero resource must be KindValidation, got kind=%v err=%v", errors.KindOf(err), err)
		}
	})

	t.Run("nil document version id", func(t *testing.T) {
		err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			_, e := svc.Attach(ctx, db, attachment.AttachIn{
				Resource:          resource.Ref{Type: "test.thing", ID: uuid.New()},
				DocumentVersionID: uuid.Nil, // missing
			})
			return e
		})
		if errors.KindOf(err) != errors.KindValidation {
			t.Fatalf("nil doc version must be KindValidation, got kind=%v err=%v", errors.KindOf(err), err)
		}
	})
}

// TestAttachWithCommentIDStored exercises the non-nil branch of uuidPtrArg by
// attaching with a real (FK-valid) comment id and asserting it round-trips
// through the persisted row.
func TestAttachWithCommentIDStored(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	dvID := seedDocVersion(t, h, tn.ID)
	cID := seedComment(t, h, tn.ID, ref)

	var aid uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		aid, e = svc.Attach(ctx, db, attachment.AttachIn{
			Resource:          ref,
			DocumentVersionID: dvID,
			CommentID:         &cID,
		})
		return e
	}); err != nil {
		t.Fatalf("Attach with comment: %v", err)
	}

	var atts []attachment.Attachment
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		atts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(atts) != 1 {
		t.Fatalf("want 1 attachment, got %d", len(atts))
	}
	if atts[0].ID != aid {
		t.Errorf("id mismatch: got %s want %s", atts[0].ID, aid)
	}
	if atts[0].CommentID == nil || *atts[0].CommentID != cID {
		t.Errorf("comment_id = %v, want %s", atts[0].CommentID, cID)
	}
}

// TestAttachInsertErrorOnBadCommentFK drives the insert-error branch: a non-nil
// comment id that references no existing comment triggers a real FK violation
// on INSERT, which Attach wraps (defaulting to KindInternal).
func TestAttachInsertErrorOnBadCommentFK(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	dvID := seedDocVersion(t, h, tn.ID) // valid so pre-check passes
	bogusComment := uuid.New()          // no such comment row -> FK violation

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Attach(ctx, db, attachment.AttachIn{
			Resource:          ref,
			DocumentVersionID: dvID,
			CommentID:         &bogusComment,
		})
		return e
	})
	if err == nil {
		t.Fatal("expected FK-violation error on insert, got nil")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("insert FK violation should wrap to KindInternal, got kind=%v err=%v", errors.KindOf(err), err)
	}

	// And nothing was persisted (transaction rolled back).
	var atts []attachment.Attachment
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		atts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatalf("List after failed attach: %v", err)
	}
	if len(atts) != 0 {
		t.Fatalf("failed attach must persist nothing, got %d rows", len(atts))
	}
}

// TestAttachEmitsOutboxEvent wires a real outbox writer and asserts Attach
// writes an "attachment.created" event carrying the new attachment id. Covers
// the s.ob != nil emission branch.
func TestAttachEmitsOutboxEvent(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := attachment.New(model.UUIDv7(), outbox.NewWriter(model.UUIDv7()))
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	dvID := seedDocVersion(t, h, tn.ID)

	var aid uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		aid, e = svc.Attach(ctx, db, attachment.AttachIn{Resource: ref, DocumentVersionID: dvID})
		return e
	}); err != nil {
		t.Fatalf("Attach: %v", err)
	}

	var got int
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT count(*) FROM events_outbox
			  WHERE event_type = 'attachment.created'
			    AND payload->>'attachment_id' = $1`, aid.String()).Scan(&got)
	}); err != nil {
		t.Fatalf("query events_outbox: %v", err)
	}
	if got != 1 {
		t.Fatalf("want exactly 1 attachment.created outbox event for %s, got %d", aid, got)
	}
}

// TestAttachDocVersionCheckError covers the wrap of the pre-check QueryRow
// failure: a canceled context makes the SELECT EXISTS query fail, which Attach
// wraps before it can reach the insert.
func TestAttachDocVersionCheckError(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		cctx, cancel := context.WithCancel(ctx)
		cancel() // force QueryRow to fail
		_, e := svc.Attach(cctx, db, attachment.AttachIn{
			Resource:          ref,
			DocumentVersionID: uuid.New(),
		})
		return e
	})
	if err == nil {
		t.Fatal("expected pre-check query error on canceled context, got nil")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("canceled pre-check should wrap to KindInternal, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

// TestDetachNotFound covers the ErrNoRows path: detaching an id that does not
// exist returns KindNotFound.
func TestDetachNotFound(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := database.WithActorID(testkit.TenantCtx(tn.ID), uuid.New())
	svc := newSvc()

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Detach(ctx, db, uuid.New(), 1)
	})
	if errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("detach of unknown id must be KindNotFound, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

// TestDetachVersionConflict covers the optimistic-lock branch: a stale expected
// version updates zero rows and yields KindVersionConflict, and the attachment
// remains active.
func TestDetachVersionConflict(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := database.WithActorID(testkit.TenantCtx(tn.ID), uuid.New())
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	dvID := seedDocVersion(t, h, tn.ID)

	var aid uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		aid, e = svc.Attach(ctx, db, attachment.AttachIn{Resource: ref, DocumentVersionID: dvID})
		return e
	}); err != nil {
		t.Fatalf("Attach: %v", err)
	}

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Detach(ctx, db, aid, 99) // wrong version
	})
	if errors.KindOf(err) != errors.KindVersionConflict {
		t.Fatalf("stale version must be KindVersionConflict, got kind=%v err=%v", errors.KindOf(err), err)
	}

	// Row untouched: still active at version 1.
	var atts []attachment.Attachment
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		atts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(atts) != 1 || atts[0].Status != "active" || atts[0].Version != 1 {
		t.Fatalf("attachment should be unchanged (active,v1); got %+v", atts)
	}
}

// TestDetachLoadError covers the wrap of the created_by lookup failure (a
// non-ErrNoRows error) via a canceled context.
func TestDetachLoadError(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := database.WithActorID(testkit.TenantCtx(tn.ID), uuid.New())
	svc := newSvc()

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		return svc.Detach(cctx, db, uuid.New(), 1)
	})
	if err == nil {
		t.Fatal("expected load error on canceled context, got nil")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("canceled load should wrap to KindInternal, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

// TestListQueryError covers the wrap of the List query failure via a canceled
// context.
func TestListQueryError(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}

	err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		_, e := svc.List(cctx, db, ref)
		return e
	})
	if err == nil {
		t.Fatal("expected List query error on canceled context, got nil")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("canceled List query should wrap to KindInternal, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

// TestAttachOutboxWriteErrorRollsBack is the DATA-08 W0-T1 fault-injection
// regression: when the outbox write after the attachment INSERT fails, Attach
// must propagate the error (not discard it), and the caller's tenant
// transaction must roll back so the attachment row does NOT exist afterward —
// the whole operation rolls back, not just "Attach returned an error but the
// row is still there."
func TestAttachOutboxWriteErrorRollsBack(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	injected := stderrors.New("outbox backend unavailable")
	svc := attachment.New(model.UUIDv7(), failingOutboxWriter{err: injected})
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	dvID := seedDocVersion(t, h, tn.ID)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Attach(ctx, db, attachment.AttachIn{Resource: ref, DocumentVersionID: dvID})
		return e
	})
	if err == nil {
		t.Fatal("expected Attach to propagate the injected outbox-write error, got nil")
	}
	if !stderrors.Is(err, injected) {
		t.Fatalf("Attach error should wrap the injected outbox error, got %v", err)
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("outbox-write failure should wrap to KindInternal, got kind=%v err=%v", errors.KindOf(err), err)
	}

	// The whole operation must roll back: since WithTenant's callback returned a
	// non-nil error, the outer transaction never committed, so the attachment
	// row must not exist in a fresh transaction either.
	var count int
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT count(*) FROM attachments WHERE resource_type = $1 AND resource_id = $2`,
			ref.Type, ref.ID).Scan(&count)
	}); err != nil {
		t.Fatalf("query attachments: %v", err)
	}
	if count != 0 {
		t.Fatalf("failed Attach (outbox write error) must persist nothing, got %d attachment rows", count)
	}
}
