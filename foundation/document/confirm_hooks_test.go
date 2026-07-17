package document_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
		// A VALID object exists under the forged key (same bytes), so wrong-key
		// passes every object check (stat, size, checksum, MIME) and the session
		// CAS itself is the discriminator — the historical pre-CAS hook location
		// would have fired here (second closure audit 2026-07-17: the earlier
		// missing-object variant exited during Stat, before either hook location).
		a.store.Put(wrongKey.StorageKey, body)

		for name, in := range map[string]document.ConfirmInput{
			"cross-document": crossDoc,
			"wrong-version":  wrongVersion,
			"wrong-key":      wrongKey,
		} {
			if _, err := a.svc.ConfirmUpload(ctx, db, in); err == nil {
				t.Fatalf("%s confirmation succeeded", name)
			}
			if got := a.calls.Load(); got != 0 {
				t.Fatalf("%s confirmation invoked the upload hook %d time(s); rejected confirmations must invoke zero hooks", name, got)
			}
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

	confirmWon, sweepWon := 0, 0
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
		// Record which lock order this round actually took; the deterministic
		// BothLockOrders subtests force each order regardless of what the racy
		// rounds happened to produce.
		if confirmErr == nil {
			confirmWon++
		} else {
			sweepWon++
		}
	}
	t.Logf("racy rounds outcome: confirm-first=%d sweep-first=%d (both orders forced deterministically in TestIntegrationConfirmVersusRetentionBothLockOrders)", confirmWon, sweepWon)
}

// Second closure-audit regression (2026-07-17, F-05): the hook runs inside the
// confirming transaction, which can fail AFTER the hook returns; the reserved
// upload is then retryable. The event's Tx and DeliveryID must make hook
// effects safe on that path — a Tx-bound effect (the canonical outbox scan
// enqueue) is never delivered before commit and lands exactly once, and an
// external effect is re-delivered only with an IDENTICAL DeliveryID so an
// idempotent consumer deduplicates it.
func TestIntegrationHookEffectsAtomicOrDeduplicatedAcrossRetry(t *testing.T) {
	h := testkit.NewDB(t)
	reg := document.NewRegistry()
	reg.Register("core", document.Class{Key: "core.doc"})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	store := storage.NewMemory()
	scanWriter := outbox.NewWriter(model.UUIDv7())

	var mu sync.Mutex
	var deliveries []string // external (non-Tx) effect log — survives rollback
	hooks := &document.Hooks{}
	hooks.OnFileUpload(func(ctx context.Context, _ document.UploadEvent) error {
		d, ok := document.UploadDeliveryFromContext(ctx)
		if !ok {
			return errors.New("hook invoked without a delivery context")
		}
		mu.Lock()
		deliveries = append(deliveries, d.DeliveryID)
		mu.Unlock()
		// Tx-bound effect: enqueue the scan through the confirming transaction.
		return scanWriter.Write(ctx, d.Tx, outbox.Event{
			Type:    "document.scan.requested",
			Payload: map[string]any{"delivery_id": d.DeliveryID},
		})
	})
	svc := document.New(reg, store, nil, outbox.NewWriter(model.UUIDv7()), hooks, model.UUIDv7())
	tn := testkit.CreateTenant(t, h).ID
	ctx := database.WithActorID(testkit.TenantCtx(tn), uuid.New())

	body := []byte("atomic effect payload")
	var docID uuid.UUID
	var sess document.UploadSession
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var err error
		docID, err = svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		if err != nil {
			return err
		}
		sess, err = svc.InitiateUploadChecksum(ctx, db, docID, sum(body))
		if err != nil {
			return err
		}
		store.Put(sess.StorageKey, body)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	in := document.ConfirmInput{
		SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo,
		StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
		DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
	}
	scanEvents := func() int {
		t.Helper()
		var n int
		if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			return db.QueryRow(ctx,
				`SELECT count(*) FROM events_outbox WHERE event_type = 'document.scan.requested'`).Scan(&n)
		}); err != nil {
			t.Fatal(err)
		}
		return n
	}

	// Attempt 1: the confirmation (and hook) succeed, then a post-hook failure
	// aborts the transaction before commit.
	injected := errors.New("post-hook failure before commit")
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, err := svc.ConfirmUpload(ctx, db, in); err != nil {
			t.Fatalf("ConfirmUpload (attempt 1): %v", err)
		}
		return injected
	})
	if !errors.Is(err, injected) {
		t.Fatalf("expected the injected failure, got %v", err)
	}
	if got := scanEvents(); got != 0 {
		t.Fatalf("Tx-bound hook effect visible after rollback: %d scan events, want 0 (nothing delivered before commit)", got)
	}
	if len(deliveries) != 1 {
		t.Fatalf("external effect log after failed attempt = %d entries, want 1", len(deliveries))
	}

	// Retry the same reserved upload: it succeeds and commits.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, err := svc.ConfirmUpload(ctx, db, in)
		return err
	}); err != nil {
		t.Fatalf("ConfirmUpload (retry): %v", err)
	}
	if got := scanEvents(); got != 1 {
		t.Fatalf("committed scan events = %d, want exactly 1", got)
	}
	if len(deliveries) != 2 {
		t.Fatalf("external effect log after retry = %d entries, want 2", len(deliveries))
	}
	if deliveries[0] != deliveries[1] {
		t.Fatalf("DeliveryID not stable across retry: %q vs %q — external consumers cannot deduplicate", deliveries[0], deliveries[1])
	}
	if deliveries[0] != sess.SessionID.String() {
		t.Fatalf("DeliveryID = %q, want the durable session identity %s", deliveries[0], sess.SessionID)
	}
}

// Second closure-audit evidence fix (2026-07-17): the racy rounds above make
// both lock orders LIKELY; these two subtests FORCE each order
// deterministically and assert its one legal terminal state.
func TestIntegrationConfirmVersusRetentionBothLockOrders(t *testing.T) {
	a := newHookedHarness(t)
	platTxM := database.NewManager(a.h.Platform, config.DB{},
		database.WithRole("app_platform"), database.WithRLSGuard())
	tenantID, _ := database.TenantIDFrom(a.ctx)

	prepare := func(t *testing.T, body []byte) (uuid.UUID, document.UploadSession) {
		t.Helper()
		var docID uuid.UUID
		var sess document.UploadSession
		if err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
			docID, sess = a.prepared(t, db, ctx, body)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := a.h.Admin.Exec(context.Background(),
			`UPDATE documents SET retention_until = now() - interval '1 hour' WHERE id = $1`, docID); err != nil {
			t.Fatal(err)
		}
		return docID, sess
	}
	confirm := func(docID uuid.UUID, sess document.UploadSession, body []byte) error {
		return a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
			_, err := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
				SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo,
				StorageKey: sess.StorageKey, DeclaredSize: int64(len(body)),
				DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
			})
			return err
		})
	}

	t.Run("sweep-first: void lands, confirmation is rejected", func(t *testing.T) {
		body := []byte("order sweep-first")
		docID, sess := prepare(t, body)
		if _, err := a.svc.SweepRetention(context.Background(), platTxM, tenantID, time.Now()); err != nil {
			t.Fatalf("SweepRetention: %v", err)
		}
		err := confirm(docID, sess, body)
		if err == nil {
			t.Fatal("confirmation succeeded against a document retention had voided")
		}
		if kerr.KindOf(err) != kerr.KindConflict {
			t.Fatalf("kind = %v, want KindConflict (err=%v)", kerr.KindOf(err), err)
		}
		var versions int
		if err := a.h.Admin.QueryRow(context.Background(),
			`SELECT count(*) FROM document_versions WHERE document_id = $1`, docID).Scan(&versions); err != nil {
			t.Fatal(err)
		}
		if versions != 0 {
			t.Fatalf("voided document gained %d version(s)", versions)
		}
	})

	t.Run("confirm-first: version lands, then retention voids document and version", func(t *testing.T) {
		body := []byte("order confirm-first")
		docID, sess := prepare(t, body)
		if err := confirm(docID, sess, body); err != nil {
			t.Fatalf("confirm on a still-active document: %v", err)
		}
		if _, err := a.svc.SweepRetention(context.Background(), platTxM, tenantID, time.Now()); err != nil {
			t.Fatalf("SweepRetention: %v", err)
		}
		var docStatus, verStatus string
		if err := a.h.Admin.QueryRow(context.Background(),
			`SELECT d.status, v.status FROM documents d JOIN document_versions v ON v.document_id = d.id
			  WHERE d.id = $1`, docID).Scan(&docStatus, &verStatus); err != nil {
			t.Fatal(err)
		}
		if docStatus != "voided" || verStatus != "voided" {
			t.Fatalf("after confirm-then-sweep: doc=%q version=%q, want both voided", docStatus, verStatus)
		}
	})
}

// Compile-time compatibility contract (third closure audit 2026-07-17):
// UploadEvent's v1 field set is FROZEN. External consumers write unkeyed
// composite literals like the one below; if this stops compiling, a field was
// added or reordered — a source-incompatible change for a stable post-v1.0
// API. Transactional delivery metadata travels on the context
// (UploadDeliveryFromContext), never as new event fields.
func TestUploadEventUnkeyedLiteralCompatibility(t *testing.T) {
	e := document.UploadEvent{
		"doc-id", "core.doc", 1, "storage/key", "text/plain", int64(42), document.SensitivityInternal,
	}
	if e.DocumentID != "doc-id" || e.VersionNo != 1 || e.SizeBytes != 42 {
		t.Fatalf("positional literal mapped unexpectedly: %+v", e)
	}
}
