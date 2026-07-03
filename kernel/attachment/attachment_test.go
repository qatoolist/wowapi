package attachment_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/attachment"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/testkit"
)

func newSvc() *attachment.Service {
	return attachment.New(model.UUIDv7(), nil)
}

// seedDocVersion inserts a documents row and a document_versions row via Admin
// (bypasses RLS; tenant_id set explicitly) and returns the document_version id.
func seedDocVersion(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID) uuid.UUID {
	t.Helper()
	docID := uuid.New()
	dvID := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO documents (id, tenant_id, document_class, title, created_by)
		 VALUES ($1, $2, 'test', 'test doc', $3)`,
		docID, tenantID, uuid.Nil); err != nil {
		t.Fatalf("seed document: %v", err)
	}
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO document_versions
		    (id, tenant_id, document_id, version_no, storage_key, mime_type, size_bytes, checksum_sha256, uploaded_by)
		 VALUES ($1, $2, $3, 1, 'key/test', 'application/octet-stream', 0,
		         'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855', $4)`,
		dvID, tenantID, docID, uuid.Nil); err != nil {
		t.Fatalf("seed document_version: %v", err)
	}
	return dvID
}

func TestAttachAndList(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	dvID := seedDocVersion(t, h, tn.ID)

	var aid uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		aid, e = svc.Attach(ctx, db, attachment.AttachIn{
			Resource:          ref,
			DocumentVersionID: dvID,
		})
		return e
	}); err != nil {
		t.Fatalf("Attach: %v", err)
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
		t.Errorf("id mismatch: got %s, want %s", atts[0].ID, aid)
	}
	if atts[0].Status != "active" {
		t.Errorf("status = %q, want active", atts[0].Status)
	}
	if atts[0].DocumentVersionID != dvID {
		t.Errorf("document_version_id mismatch")
	}
}

func TestAttachBogusDocumentVersionError(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Attach(ctx, db, attachment.AttachIn{
			Resource:          ref,
			DocumentVersionID: uuid.New(), // non-existent
		})
		return e
	})
	if errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("bogus document version should be KindNotFound, got kind=%v err=%v", errors.KindOf(err), err)
	}
}

func TestDetachVoids(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	ctx := database.WithActorID(testkit.TenantCtx(tn.ID), uuid.New()) // attacher = context actor
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

	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Detach(ctx, db, aid, 1)
	}); err != nil {
		t.Fatalf("Detach: %v", err)
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
		t.Fatalf("want 1 (voided) attachment, got %d", len(atts))
	}
	if atts[0].Status != "voided" {
		t.Errorf("status = %q, want voided", atts[0].Status)
	}
	if atts[0].Version != 2 {
		t.Errorf("version = %d, want 2", atts[0].Version)
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
	dvID := seedDocVersion(t, h, tnA.ID)

	if err := h.TxM.WithTenant(ctxA, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Attach(ctx, db, attachment.AttachIn{Resource: ref, DocumentVersionID: dvID})
		return e
	}); err != nil {
		t.Fatalf("Attach in tenant A: %v", err)
	}

	var atts []attachment.Attachment
	if err := h.TxM.WithTenantRO(ctxB, func(ctx context.Context, db database.TenantDB) error {
		var e error
		atts, e = svc.List(ctx, db, ref)
		return e
	}); err != nil {
		t.Fatalf("List from tenant B: %v", err)
	}
	if len(atts) != 0 {
		t.Fatalf("tenant B should see 0 attachments, got %d", len(atts))
	}
}

// TestDetachByNonCreatorForbidden is the SEC-46 regression: only the actor who
// created the attachment may detach it.
func TestDetachByNonCreatorForbidden(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	svc := newSvc()
	ref := resource.Ref{Type: "test.thing", ID: uuid.New()}
	dvID := seedDocVersion(t, h, tn.ID)
	creatorCtx := database.WithActorID(testkit.TenantCtx(tn.ID), uuid.New())

	var aid uuid.UUID
	if err := h.TxM.WithTenant(creatorCtx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		aid, e = svc.Attach(ctx, db, attachment.AttachIn{Resource: ref, DocumentVersionID: dvID})
		return e
	}); err != nil {
		t.Fatalf("Attach: %v", err)
	}
	otherCtx := database.WithActorID(testkit.TenantCtx(tn.ID), uuid.New())
	err := h.TxM.WithTenant(otherCtx, func(ctx context.Context, db database.TenantDB) error {
		return svc.Detach(ctx, db, aid, 1)
	})
	if errors.KindOf(err) != errors.KindForbidden {
		t.Fatalf("non-creator detach must be forbidden, got kind=%v err=%v", errors.KindOf(err), err)
	}
}
