package document_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/testkit"
)

// hooks_fire_test.go — QA G10 (extension point / security): OnFileUpload and
// OnDocumentAccess are the framework's hook points — the canonical OnFileUpload
// enqueues a malware scan and can BLOCK the version; OnDocumentAccess is the
// watermark/deny slot. Every existing document test wires nil hooks, so the
// firing path (a hook aborting a confirm / denying a download) was untested.

func newHookedService(t *testing.T, hooks *document.Hooks) (*document.Service, *storage.Memory) {
	t.Helper()
	reg := document.NewRegistry()
	reg.Register("core", document.Class{Key: "core.doc"})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	store := storage.NewMemory()
	return document.New(reg, store, nil, outbox.NewWriter(model.UUIDv7()), hooks, model.UUIDv7()), store
}

func TestIntegrationUploadHookAbortsConfirm(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tn), actor)

	var gotClass, gotMIME string
	hooks := document.NewHooks()
	hooks.OnFileUpload(func(_ context.Context, e document.UploadEvent) error {
		gotClass, gotMIME = e.Class, e.MIME
		return errors.New("scan backend unavailable") // reject the version
	})
	svc, store := newHookedService(t, hooks)

	var docID uuid.UUID
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		if e != nil {
			return e
		}
		body := []byte("payload")
		sess, e := svc.InitiateUpload(ctx, db, docID, sum(body))
		if e != nil {
			return e
		}
		store.Put(sess.StorageKey, body)
		_, e = svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		// The hook aborted the confirm.
		if e == nil {
			t.Fatal("ConfirmUpload must fail when an OnFileUpload hook returns an error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// The hook received real event data.
	if gotClass != "core.doc" || gotMIME == "" {
		t.Fatalf("upload hook got wrong event: class=%q mime=%q", gotClass, gotMIME)
	}
	// And NO version row was committed (the abort left no immutable pointer).
	var versions int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM document_versions WHERE document_id=$1`, docID).Scan(&versions); err != nil {
		t.Fatal(err)
	}
	if versions != 0 {
		t.Fatalf("aborted upload left %d version rows, want 0", versions)
	}
}

func TestIntegrationAccessHookDeniesDownload(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tn), actor)

	denied := errors.New("watermark service down")
	var accessSeen bool
	hooks := document.NewHooks()
	hooks.OnDocumentAccess(func(_ context.Context, _ document.AccessEvent) error {
		accessSeen = true
		return denied
	})
	svc, store := newHookedService(t, hooks)

	// Upload a clean version (no upload hook) then attempt a download.
	var docID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		if e != nil {
			return e
		}
		body := []byte("hello, framework")
		sess, e := svc.InitiateUpload(ctx, db, docID, sum(body))
		if e != nil {
			return e
		}
		store.Put(sess.StorageKey, body)
		_, e = svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain; charset=utf-8",
		})
		return e
	}); err != nil {
		t.Fatal(err)
	}

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := svc.Download(ctx, db,
			authz.Actor{CapacityID: actor, UserID: actor, TenantID: tn},
			document.DownloadInput{DocumentID: docID})
		return e
	})
	if err == nil {
		t.Fatal("Download must fail when an OnDocumentAccess hook returns an error")
	}
	if !accessSeen {
		t.Fatal("the access hook was never invoked")
	}
}
