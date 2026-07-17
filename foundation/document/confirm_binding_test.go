package document_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/v2/foundation/document"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// F-05 regressions (adversarial-framework-review-2026-07-17): a pending and
// unexpired upload session may confirm only the document, version, and storage
// object it reserved.

func sessionStatus(t *testing.T, a *harness, sessionID uuid.UUID) string {
	t.Helper()
	var status string
	if err := a.h.Admin.QueryRow(context.Background(),
		`SELECT status FROM document_upload_sessions WHERE id = $1`, sessionID).Scan(&status); err != nil {
		t.Fatalf("read session status: %v", err)
	}
	return status
}

func versionCount(t *testing.T, a *harness, docID uuid.UUID) int {
	t.Helper()
	var n int
	if err := a.h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM document_versions WHERE document_id = $1`, docID).Scan(&n); err != nil {
		t.Fatalf("count versions: %v", err)
	}
	return n
}

// Cross-document substitution: confirming document A's session against
// document B must fail entirely — B gets no version, A's session stays pending.
func TestIntegrationConfirmUploadCrossDocumentRejected(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("cross-document payload")
	var sessID, docA, docB uuid.UUID
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docA, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "A"})
		if e != nil {
			return e
		}
		docB, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "B"})
		if e != nil {
			return e
		}
		sess, e := a.svc.InitiateUploadChecksum(ctx, db, docA, sum(body))
		if e != nil {
			return e
		}
		sessID = sess.SessionID
		a.store.Put(sess.StorageKey, body)

		_, confirmErr := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docB, VersionNo: sess.VersionNo,
			StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
			DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if confirmErr == nil {
			t.Fatal("session for document A confirmed against document B")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := versionCount(t, a, docB); n != 0 {
		t.Fatalf("document B gained %d version(s) from another document's session", n)
	}
	if st := sessionStatus(t, a, sessID); st != "pending" {
		t.Fatalf("session status = %q after rejected substitution, want pending", st)
	}
}

// An expired-but-unswept session must not confirm.
func TestIntegrationConfirmUploadExpiredSessionRejected(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("expired payload")
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
		if _, err := db.Exec(ctx,
			`UPDATE document_upload_sessions SET expires_at = now() - interval '1 minute' WHERE id = $1`,
			sess.SessionID); err != nil {
			return err
		}
		_, confirmErr := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo,
			StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
			DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if confirmErr == nil {
			t.Fatal("expired session confirmed (sweeper had not run yet)")
		}
		if kerr.KindOf(confirmErr) != kerr.KindConflict {
			t.Fatalf("expired confirm kind = %v, want KindConflict (err=%v)", kerr.KindOf(confirmErr), confirmErr)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Wrong version or storage key must not settle the session; the correct
// confirmation still succeeds exactly once afterward.
func TestIntegrationConfirmUploadWrongIdentityDoesNotSettle(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc"})
	body := []byte("identity payload")
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

		if _, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo + 7,
			StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
			DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		}); e == nil {
			t.Fatal("confirm with wrong version succeeded")
		}
		var st string
		if err := db.QueryRow(ctx,
			`SELECT status FROM document_upload_sessions WHERE id = $1`, sess.SessionID).Scan(&st); err != nil {
			return err
		}
		if st != "pending" {
			t.Fatalf("session settled by a wrong-version confirm: %q", st)
		}

		if _, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo,
			StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
			DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		}); e != nil {
			t.Fatalf("correct confirm after rejected attempts: %v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
