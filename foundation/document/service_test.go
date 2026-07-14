package document_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/testkit"
)

func sum(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

// harness bundles a service over a fresh DB with an in-memory store.
type harness struct {
	h     *testkit.DBHandle
	store *storage.Memory
	svc   *document.Service
	tn    uuid.UUID
	actor uuid.UUID // owner capacity
	ctx   context.Context
}

func newHarness(t *testing.T, class document.Class) *harness {
	t.Helper()
	h := testkit.NewDB(t)
	reg := document.NewRegistry()
	reg.Register("core", class)
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	store := storage.NewMemory()
	svc := document.New(reg, store, nil, outbox.NewWriter(model.UUIDv7()), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h).ID
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tn), actor)
	return &harness{h: h, store: store, svc: svc, tn: tn, actor: actor, ctx: ctx}
}

// uploadVersion creates a document and confirms one version with the given bytes.
func (a *harness) uploadVersion(t *testing.T, class string, sens document.Sensitivity, body []byte, mime string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	var docID, verID uuid.UUID
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = a.svc.Create(ctx, db, document.CreateInput{Class: class, Title: "T", Sensitivity: sens})
		if e != nil {
			return e
		}
		sess, e := a.svc.InitiateUploadChecksum(ctx, db, docID, sum(body))
		if e != nil {
			return e
		}
		a.store.Put(sess.StorageKey, body)
		verID, e = a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: mime,
		})
		return e
	})
	if err != nil {
		t.Fatalf("uploadVersion: %v", err)
	}
	return docID, verID
}

func (a *harness) download(t *testing.T, actor uuid.UUID, docID uuid.UUID) (document.Download, error) {
	t.Helper()
	var out document.Download
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		out, e = a.svc.Download(ctx, db, authz.Actor{CapacityID: actor, UserID: actor, TenantID: a.tn}, document.DownloadInput{DocumentID: docID})
		return e
	})
	return out, err
}

func TestIntegrationUploadRoundTrip(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", AllowedMIME: nil, MaxBytes: 1 << 20})
	body := []byte("hello, framework")
	docID, verID := a.uploadVersion(t, "core.doc", document.SensitivityInternal, body, "text/plain; charset=utf-8")
	if verID == uuid.Nil {
		t.Fatal("no version id")
	}
	// Owner downloads (internal sensitivity → pending scan does not block).
	out, err := a.download(t, a.actor, docID)
	if err != nil {
		t.Fatalf("owner download: %v", err)
	}
	if out.URL.Method != http.MethodGet || out.VersionNo != 1 {
		t.Fatalf("bad download: %+v", out)
	}
}

func TestIntegrationConfirmVerifiesBytes(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", MaxBytes: 1 << 20})
	body := []byte("payload")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, e := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		if e != nil {
			return e
		}
		sess, e := a.svc.InitiateUploadChecksum(ctx, db, docID, sum(body))
		if e != nil {
			return e
		}
		a.store.Put(sess.StorageKey, body)
		// Declared checksum is wrong → rejected at confirm.
		_, e = a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum([]byte("different")), DeclaredMIME: "text/plain",
		})
		if kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("checksum mismatch must be rejected, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationMIMEMismatchRejected(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("just some text, not a png")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUploadChecksum(ctx, db, docID, sum(body))
		a.store.Put(sess.StorageKey, body)
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "image/png",
		})
		if kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("text sniffed as image/png declared must be rejected, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationScanGateBlocksConfidential(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", DefaultSensitivity: document.SensitivityConfidential})
	body := []byte("secret bytes")
	docID, verID := a.uploadVersion(t, "core.doc", document.SensitivityConfidential, body, "text/plain; charset=utf-8")

	// Pending scan blocks a confidential download.
	if _, err := a.download(t, a.actor, docID); kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("pending scan must block confidential download, got %v", err)
	}
	// Clean scan (platform op) then it flows.
	if err := a.svc.UpdateScanStatus(context.Background(), a.h.PlatformTxM, a.tn, verID, "clean"); err != nil {
		t.Fatalf("scan clean: %v", err)
	}
	if _, err := a.download(t, a.actor, docID); err != nil {
		t.Fatalf("clean confidential download: %v", err)
	}
}

func TestIntegrationInfectedNeverServes(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", DefaultSensitivity: document.SensitivityInternal})
	docID, verID := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("virus"), "text/plain; charset=utf-8")
	if err := a.svc.UpdateScanStatus(context.Background(), a.h.PlatformTxM, a.tn, verID, "infected"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.download(t, a.actor, docID); kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("infected version must never serve, got %v", err)
	}
}

func TestIntegrationAccessGrant(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("shared"), "text/plain; charset=utf-8")
	stranger := uuid.New()

	// A stranger cannot download.
	if _, err := a.download(t, stranger, docID); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("stranger must be forbidden, got %v", err)
	}
	// Owner grants the stranger read.
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.Grant(ctx, db, document.GrantInput{
			DocumentID: docID, GranteeKind: "capacity", GranteeRef: stranger.String(), Access: "read",
		})
		return e
	})
	if err != nil {
		t.Fatalf("grant: %v", err)
	}
	// Now the stranger can.
	if _, err := a.download(t, stranger, docID); err != nil {
		t.Fatalf("granted download: %v", err)
	}
}

func TestIntegrationRetentionSweep(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", Retention: time.Hour})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("keepsafe"), "text/plain; charset=utf-8")

	// Sweep well before expiry → nothing.
	if n, err := a.svc.SweepRetention(context.Background(), a.h.PlatformTxM, a.tn, time.Now()); err != nil || n != 0 {
		t.Fatalf("premature sweep: n=%d err=%v", n, err)
	}
	// Sweep past expiry → the version and document are voided.
	n, err := a.svc.SweepRetention(context.Background(), a.h.PlatformTxM, a.tn, time.Now().Add(2*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("retention sweep: n=%d err=%v", n, err)
	}
	if _, err := a.download(t, a.actor, docID); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("voided document must be gone, got %v", err)
	}
}

func TestIntegrationLegalHoldBlocksSweep(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", Retention: time.Hour})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("held"), "text/plain; charset=utf-8")
	// Put the document under legal hold.
	if _, err := a.h.Admin.Exec(context.Background(),
		`UPDATE documents SET legal_hold = true WHERE id = $1`, docID); err != nil {
		t.Fatal(err)
	}
	n, err := a.svc.SweepRetention(context.Background(), a.h.PlatformTxM, a.tn, time.Now().Add(2*time.Hour))
	if err != nil || n != 0 {
		t.Fatalf("legal hold must block the sweep: n=%d err=%v", n, err)
	}
	// Still downloadable.
	if _, err := a.download(t, a.actor, docID); err != nil {
		t.Fatalf("held document should remain: %v", err)
	}
}

// TestIntegrationLegalHoldRaceSurvivesSweep is the R6 regression: a legal hold
// applied concurrently with a running sweep must win — the document survives. The
// hold transaction takes the documents row lock first, so the sweep's FOR UPDATE
// blocks; when the hold commits, the sweep re-checks legal_hold and skips the
// document. Before the fix (no FOR UPDATE, unguarded void UPDATE) the sweep saw
// legal_hold=false in its snapshot and voided the row anyway.
func TestIntegrationLegalHoldRaceSurvivesSweep(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", Retention: time.Hour})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("racy"), "text/plain; charset=utf-8")

	ctx := context.Background()
	// Hold transaction: set legal_hold and KEEP the row lock (do not commit yet).
	tx, err := a.h.Admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE documents SET legal_hold = true WHERE id = $1`, docID); err != nil {
		t.Fatal(err)
	}

	type result struct {
		n   int
		err error
	}
	done := make(chan result, 1)
	go func() {
		n, err := a.svc.SweepRetention(ctx, a.h.PlatformTxM, a.tn, time.Now().Add(2*time.Hour))
		done <- result{n, err}
	}()

	// Let the sweep reach and block on the locked row, then let the hold win.
	time.Sleep(300 * time.Millisecond)
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	res := <-done
	if res.err != nil {
		t.Fatalf("sweep errored: %v", res.err)
	}
	if res.n != 0 {
		t.Fatalf("a hold applied mid-sweep must save the document, but %d version(s) were voided", res.n)
	}
	if _, err := a.download(t, a.actor, docID); err != nil {
		t.Fatalf("held document must survive the race: %v", err)
	}
}

// TestIntegrationDownloadInReadOnlyTx is the ARCH-65 regression: Download is a
// read and must succeed inside a WithTenantRO transaction (it emits no event).
func TestIntegrationDownloadInReadOnlyTx(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("ro-read"), "text/plain; charset=utf-8")
	err := a.h.TxM.WithTenantRO(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		out, e := a.svc.Download(ctx, db, authz.Actor{CapacityID: a.actor, UserID: a.actor, TenantID: a.tn}, document.DownloadInput{DocumentID: docID})
		if e != nil {
			return e
		}
		if out.URL.Method != http.MethodGet {
			t.Fatalf("bad download in RO tx: %+v", out)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("download must work in a read-only tx: %v", err)
	}
}

// TestIntegrationMIMEEssenceMismatch is the ARCH-69 regression: text/html bytes
// declared as text/plain (same top-level type) must still be rejected.
func TestIntegrationMIMEEssenceMismatch(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("<!DOCTYPE html><html><body>hi</body></html>")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		sess, _ := a.svc.InitiateUploadChecksum(ctx, db, docID, sum(body))
		a.store.Put(sess.StorageKey, body)
		_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if kerr.KindOf(e) != kerr.KindValidation {
			t.Fatalf("html declared as text/plain must be rejected, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestIntegrationDistinctUploadKeys is the ARCH-66 regression: two upload
// sessions for the same document get distinct storage keys (no clobber).
func TestIntegrationDistinctUploadKeys(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, _ := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		s1, _ := a.svc.InitiateUploadChecksum(ctx, db, docID, sum(nil))
		s2, _ := a.svc.InitiateUploadChecksum(ctx, db, docID, sum(nil))
		if s1.StorageKey == s2.StorageKey {
			t.Fatalf("two upload sessions must not share a storage key: %q", s1.StorageKey)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestIntegrationGrantRLSBlocksNonOwner is the SEC-41 regression: a module
// (app_rt) cannot INSERT a document access grant for a document it does not own,
// even bypassing the service — the restrictive ownership RLS policy blocks it.
func TestIntegrationGrantRLSBlocksNonOwner(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("owned"), "text/plain; charset=utf-8")

	other := uuid.New() // NOT the document owner
	octx := database.WithActorID(testkit.TenantCtx(a.tn), other)
	err := a.h.TxM.WithTenant(octx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx,
			`INSERT INTO document_access_grants (id, tenant_id, document_id, grantee_kind, grantee_ref, access, created_by)
			 VALUES ($1, app_tenant_id(), $2, 'capacity', $3, 'write', $4)`,
			uuid.New(), docID, other.String(), other)
		return e
	})
	if err == nil {
		t.Fatal("a non-owner must not be able to insert a grant (SEC-41 RLS backstop)")
	}
}

// TestIntegrationLegalHoldColumnProtected is the SEC-44 regression: app_rt has no
// UPDATE privilege on documents.legal_hold / status, so a module cannot clear a
// legal hold or void a document to dodge retention.
func TestIntegrationLegalHoldColumnProtected(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("held"), "text/plain; charset=utf-8")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx, `UPDATE documents SET legal_hold = false WHERE id = $1`, docID)
		return e
	})
	if err == nil {
		t.Fatal("app_rt must not be able to UPDATE documents.legal_hold (SEC-44)")
	}
}

// TestIntegrationRevokeRequiresWrite is the SEC-43 regression: a non-owner cannot
// revoke a grant.
func TestIntegrationRevokeRequiresWrite(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("shared"), "text/plain; charset=utf-8")
	stranger := uuid.New()
	var grantID uuid.UUID
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		grantID, e = a.svc.Grant(ctx, db, document.GrantInput{
			DocumentID: docID, GranteeKind: "capacity", GranteeRef: stranger.String(), Access: "read",
		})
		return e
	}); err != nil {
		t.Fatalf("owner grant: %v", err)
	}
	// The stranger (read grant only, not owner) cannot revoke.
	sctx := database.WithActorID(testkit.TenantCtx(a.tn), stranger)
	err := a.h.TxM.WithTenant(sctx, func(ctx context.Context, db database.TenantDB) error {
		return a.svc.Revoke(ctx, db, grantID)
	})
	if kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("non-owner revoke must be forbidden, got %v", err)
	}
}

func TestIntegrationTenantIsolation(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	docID, _ := a.uploadVersion(t, "core.doc", document.SensitivityInternal, []byte("mine"), "text/plain; charset=utf-8")

	// A second tenant cannot see the document at all.
	other := testkit.CreateTenant(t, a.h).ID
	octx := database.WithActorID(testkit.TenantCtx(other), a.actor)
	err := a.h.TxM.WithTenant(octx, func(ctx context.Context, db database.TenantDB) error {
		_, e := a.svc.Download(ctx, db, authz.Actor{CapacityID: a.actor, UserID: a.actor, TenantID: other}, document.DownloadInput{DocumentID: docID})
		if kerr.KindOf(e) != kerr.KindNotFound {
			t.Fatalf("cross-tenant read must be NotFound, got %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
