package document_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/testkit"
)

// Closure-review regressions (adversarial closure review 2026-07-17, F-05):
// (1) hooks must run ONLY for confirmations the session CAS accepts, and only
// with authoritative values — a rejected (cross-document, expired, replayed,
// wrong-key, wrong-version) confirmation must invoke zero hooks; (2) a voided
// document must never acquire a new active version, and a void→confirm attempt
// must invoke zero hooks and create zero versions.

type hookedHarness struct {
	h       *testkit.DBHandle
	store   *storage.Memory
	svc     *document.Service
	ctx     context.Context
	calls   *atomic.Int64
	lastDoc *atomic.Value // string: DocumentID seen by the hook
}

func newHookedHarness(t *testing.T) *hookedHarness {
	t.Helper()
	h := testkit.NewDB(t)
	reg := document.NewRegistry()
	reg.Register("core", document.Class{Key: "core.doc"})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	store := storage.NewMemory()
	calls := &atomic.Int64{}
	lastDoc := &atomic.Value{}
	hooks := &document.Hooks{}
	hooks.OnFileUpload(func(ctx context.Context, e document.UploadEvent) error {
		calls.Add(1)
		lastDoc.Store(e.DocumentID)
		return nil
	})
	svc := document.New(reg, store, nil, outbox.NewWriter(model.UUIDv7()), hooks, model.UUIDv7())
	tn := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(testkit.TenantCtx(tn), uuid.New())
	return &hookedHarness{h: h, store: store, svc: svc, ctx: ctx, calls: calls, lastDoc: lastDoc}
}

// prepared returns an active document with an uploaded, unconfirmed session.
func (a *hookedHarness) prepared(t *testing.T, db database.TenantDB, ctx context.Context, body []byte) (uuid.UUID, document.UploadSession) {
	t.Helper()
	docID, err := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	sess, err := a.svc.InitiateUploadChecksum(ctx, db, docID, sum(body))
	if err != nil {
		t.Fatalf("Initiate: %v", err)
	}
	a.store.Put(sess.StorageKey, body)
	return docID, sess
}

func TestIntegrationRejectedConfirmationsInvokeNoHooks(t *testing.T) {
	a := newHookedHarness(t)
	body := []byte("hook gating payload")
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, sess := a.prepared(t, db, ctx, body)
		otherDoc, err := a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "B"})
		if err != nil {
			return err
		}
		base := document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo,
			StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
			DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		}

		crossDoc := base
		crossDoc.DocumentID = otherDoc
		wrongVersion := base
		wrongVersion.VersionNo += 7
		wrongKey := base
		wrongKey.StorageKey = sess.StorageKey + "-forged"

		for name, in := range map[string]document.ConfirmInput{
			"cross-document": crossDoc,
			"wrong-version":  wrongVersion,
		} {
			if _, err := a.svc.ConfirmUpload(ctx, db, in); err == nil {
				t.Fatalf("%s confirmation succeeded", name)
			}
			if got := a.calls.Load(); got != 0 {
				t.Fatalf("%s confirmation invoked the upload hook %d time(s); rejected confirmations must invoke zero hooks", name, got)
			}
		}
		// wrong-key fails the object stat (forged key has no object) — also zero hooks.
		if _, err := a.svc.ConfirmUpload(ctx, db, wrongKey); err == nil {
			t.Fatal("wrong-key confirmation succeeded")
		}
		if got := a.calls.Load(); got != 0 {
			t.Fatalf("wrong-key confirmation invoked the upload hook %d time(s)", got)
		}

		// Expired session: also zero hooks.
		if _, err := db.Exec(ctx,
			`UPDATE document_upload_sessions SET expires_at = now() - interval '1 minute' WHERE id = $1`, sess.SessionID); err != nil {
			return err
		}
		if _, err := a.svc.ConfirmUpload(ctx, db, base); err == nil {
			t.Fatal("expired confirmation succeeded")
		}
		if got := a.calls.Load(); got != 0 {
			t.Fatalf("expired confirmation invoked the upload hook %d time(s)", got)
		}
		if _, err := db.Exec(ctx,
			`UPDATE document_upload_sessions SET expires_at = now() + interval '10 minutes' WHERE id = $1`, sess.SessionID); err != nil {
			return err
		}

		// The one VALID confirmation invokes the hook exactly once, with the
		// authoritative document id.
		if _, err := a.svc.ConfirmUpload(ctx, db, base); err != nil {
			t.Fatalf("valid confirmation: %v", err)
		}
		if got := a.calls.Load(); got != 1 {
			t.Fatalf("valid confirmation invoked the hook %d time(s), want exactly 1", got)
		}
		if got := a.lastDoc.Load().(string); got != docID.String() {
			t.Fatalf("hook saw document %s, want authoritative %s", got, docID)
		}

		// Replay of the settled session: rejected, and STILL exactly one call.
		if _, err := a.svc.ConfirmUpload(ctx, db, base); err == nil {
			t.Fatal("replayed confirmation succeeded")
		}
		if got := a.calls.Load(); got != 1 {
			t.Fatalf("replayed confirmation raised hook calls to %d, want 1", got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationVoidedDocumentCannotAcquireVersions(t *testing.T) {
	a := newHookedHarness(t)
	body := []byte("void gating payload")
	var docID uuid.UUID
	var sess document.UploadSession
	if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		docID, sess = a.prepared(t, db, ctx, body)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// Retention voids the document while the (unexpired) session is open — the
	// sweep's own terminal write reproduced deterministically with its
	// privileged posture (retention runs as app_platform, not the tenant role).
	if _, err := a.h.Admin.Exec(context.Background(),
		`UPDATE documents SET status = 'voided', updated_at = now() WHERE id = $1`, docID); err != nil {
		t.Fatal(err)
	}

	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		_, err := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo,
			StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
			DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		})
		if err == nil {
			t.Fatal("confirmation added a version beneath a voided document")
		}
		if kerr.KindOf(err) != kerr.KindConflict {
			t.Fatalf("void-confirm kind = %v, want KindConflict (err=%v)", kerr.KindOf(err), err)
		}
		if got := a.calls.Load(); got != 0 {
			t.Fatalf("void-confirm invoked the upload hook %d time(s), want 0", got)
		}
		var versions int
		if err := db.QueryRow(ctx,
			`SELECT count(*) FROM document_versions WHERE document_id = $1`, docID).Scan(&versions); err != nil {
			return err
		}
		if versions != 0 {
			t.Fatalf("voided document gained %d version(s)", versions)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Race: a confirmation and the REAL SweepRetention run concurrently on a
// retention-lapsed document. Both take the documents row lock first (retention
// via SELECT ... FOR UPDATE over candidates, confirmation via its authoritative
// FOR UPDATE read), so every interleaving must end in one of exactly two legal
// states — the new version lands first and retention voids it too, or the void
// lands first and the confirmation is rejected. Never an ACTIVE version
// beneath a voided document. Repeated to exercise both lock orders.
func TestIntegrationConfirmVersusRetentionRaceInvariant(t *testing.T) {
	a := newHookedHarness(t)
	platTxM := database.NewManager(a.h.Platform, config.DB{},
		database.WithRole("app_platform"), database.WithRLSGuard())
	tenantID, _ := database.TenantIDFrom(a.ctx)

	for round := range 6 {
		body := []byte(fmt.Sprintf("race payload %d", round))
		var docID uuid.UUID
		var sess document.UploadSession
		if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
			docID, sess = a.prepared(t, db, ctx, body)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		// Make the document retention-eligible so the REAL sweep voids it.
		if _, err := a.h.Admin.Exec(context.Background(),
			`UPDATE documents SET retention_until = now() - interval '1 hour' WHERE id = $1`, docID); err != nil {
			t.Fatal(err)
		}

		start := make(chan struct{})
		confirmDone := make(chan error, 1)
		sweepDone := make(chan error, 1)
		go func() {
			<-start
			confirmDone <- a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
				_, err := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
					SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo,
					StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
					DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
				})
				return err
			})
		}()
		go func() {
			<-start
			_, err := a.svc.SweepRetention(context.Background(), platTxM, tenantID, time.Now())
			sweepDone <- err
		}()
		close(start)
		confirmErr := <-confirmDone
		if err := <-sweepDone; err != nil {
			t.Fatalf("round %d: SweepRetention: %v", round, err)
		}

		var docStatus string
		var activeUnderVoided int
		if err := a.h.Admin.QueryRow(context.Background(),
			`SELECT status FROM documents WHERE id = $1`, docID).Scan(&docStatus); err != nil {
			t.Fatal(err)
		}
		if err := a.h.Admin.QueryRow(context.Background(),
			`SELECT count(*) FROM document_versions v JOIN documents d ON d.id = v.document_id
			  WHERE d.id = $1 AND d.status = 'voided' AND v.status = 'active'`, docID).Scan(&activeUnderVoided); err != nil {
			t.Fatal(err)
		}
		if activeUnderVoided != 0 {
			t.Fatalf("round %d: race left %d ACTIVE version(s) beneath a voided document (doc=%s confirm err=%v)",
				round, activeUnderVoided, docStatus, confirmErr)
		}
		// If the sweep won and voided first, the confirmation must have been
		// rejected — never silently succeeded against a voided document.
		if docStatus == "voided" && confirmErr == nil {
			var vs string
			if err := a.h.Admin.QueryRow(context.Background(),
				`SELECT status FROM document_versions WHERE document_id = $1`, docID).Scan(&vs); err != nil {
				t.Fatal(err)
			}
			if vs != "voided" {
				t.Fatalf("round %d: confirm succeeded, doc voided, but version status is %q", round, vs)
			}
		}
	}
}
