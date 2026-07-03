package comment_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/comment"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/testkit"
)

func newSvc() *comment.Service {
	return comment.New(model.UUIDv7(), nil)
}

func TestCreateAndList(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tn.ID, userID)
	ctx = database.WithActorID(ctx, capID) // author acts as the context actor

	var cid uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{
			Resource:         ref,
			AuthorCapacityID: capID,
			Body:             "hello world",
		})
		return e
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var cmts []comment.Comment
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cmts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(cmts) != 1 {
		t.Fatalf("want 1 comment, got %d", len(cmts))
	}
	if cmts[0].ID != cid {
		t.Errorf("id mismatch: got %s, want %s", cmts[0].ID, cid)
	}
	if cmts[0].Status != "active" {
		t.Errorf("status = %q, want active", cmts[0].Status)
	}
	if cmts[0].Version != 1 {
		t.Errorf("version = %d, want 1", cmts[0].Version)
	}
}

func TestParentMismatchRejected(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref1 := resource.Ref{Type: "test.thing", ID: uuid.New()}
	ref2 := resource.Ref{Type: "test.thing", ID: uuid.New()}
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tn.ID, userID)
	ctx = database.WithActorID(ctx, capID) // author acts as the context actor

	var parentID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		parentID, e = svc.Create(ctx, db, comment.CreateIn{
			Resource:         ref1,
			AuthorCapacityID: capID,
			Body:             "root comment on ref1",
		})
		return e
	}); err != nil {
		t.Fatalf("Create parent: %v", err)
	}

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Create(ctx, db, comment.CreateIn{
			Resource:         ref2,
			ParentID:         &parentID,
			AuthorCapacityID: capID,
			Body:             "reply with wrong resource",
		})
		return e
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("parent resource mismatch should be KindValidation, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestEditChangesBodyStatusVersion(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tn.ID, userID)
	ctx = database.WithActorID(ctx, capID) // author acts as the context actor

	var cid uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: capID, Body: "original"})
		return e
	}); err != nil {
		t.Fatal(err)
	}

	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Edit(ctx, db, cid, 1, "updated body")
	}); err != nil {
		t.Fatalf("Edit: %v", err)
	}

	var cmts []comment.Comment
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cmts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatal(err)
	}
	if len(cmts) != 1 {
		t.Fatalf("want 1 comment, got %d", len(cmts))
	}
	c := cmts[0]
	if c.Body != "updated body" {
		t.Errorf("body = %q, want updated body", c.Body)
	}
	if c.Status != "edited" {
		t.Errorf("status = %q, want edited", c.Status)
	}
	if c.Version != 2 {
		t.Errorf("version = %d, want 2", c.Version)
	}
	if c.UpdatedAt == nil {
		t.Error("updated_at must be set after Edit")
	}
}

func TestEditStaleVersionConflict(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tn.ID, userID)
	ctx = database.WithActorID(ctx, capID) // author acts as the context actor

	var cid uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: capID, Body: "v1"})
		return e
	}); err != nil {
		t.Fatal(err)
	}

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Edit(ctx, db, cid, 99, "new body") // wrong version
	})
	if errors.KindOf(err) != errors.KindVersionConflict {
		t.Fatalf("stale version should be KindVersionConflict, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestVoidThenEditConflict(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tn.ID, userID)
	ctx = database.WithActorID(ctx, capID) // author acts as the context actor

	var cid uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: capID, Body: "v1"})
		return e
	}); err != nil {
		t.Fatal(err)
	}

	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Void(ctx, db, cid, 1)
	}); err != nil {
		t.Fatalf("Void: %v", err)
	}

	// After void the version is 2; Edit should detect voided status, not version mismatch.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Edit(ctx, db, cid, 2, "new body after void")
	})
	if errors.KindOf(err) != errors.KindConflict {
		t.Fatalf("editing voided comment should be KindConflict, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	tnA := testkit.CreateTenant(t, h)
	tnB := testkit.CreateTenant(t, h)
	ctxA := testkit.TenantCtx(tnA.ID)
	ctxB := testkit.TenantCtx(tnB.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	userID := testkit.CreateUser(t, h)
	capIDA := testkit.CreateCapacity(t, h, tnA.ID, userID)

	if err := h.TxM.WithTenant(ctxA, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: capIDA, Body: "tenant A comment"})
		return e
	}); err != nil {
		t.Fatalf("Create in tenant A: %v", err)
	}

	var cmts []comment.Comment
	if err := h.TxM.WithTenantRO(ctxB, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cmts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatalf("List from tenant B: %v", err)
	}
	if len(cmts) != 0 {
		t.Fatalf("tenant B should see 0 comments, got %d", len(cmts))
	}
}

// TestEditByNonAuthorForbidden is the SEC-45 regression: only the author may
// edit; a different tenant actor is forbidden.
func TestEditByNonAuthorForbidden(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tn.ID, userID)
	authorCtx := database.WithActorID(testkit.TenantCtx(tn.ID), capID)

	var cid uuid.UUID
	if err := h.TxM.WithTenant(authorCtx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		cid, e = svc.Create(ctx, db, comment.CreateIn{Resource: ref, AuthorCapacityID: capID, Body: "mine"})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	// A different actor tries to edit.
	otherCtx := database.WithActorID(testkit.TenantCtx(tn.ID), uuid.New())
	err := h.TxM.WithTenant(otherCtx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Edit(ctx, db, cid, 1, "tampered")
	})
	if errors.KindOf(err) != errors.KindForbidden {
		t.Fatalf("non-author edit must be forbidden, got kind=%v err=%v", errors.KindOf(err), err)
	}
}
