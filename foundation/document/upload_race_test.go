package document_test

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// TestIntegrationInitiateUploadConcurrentVersionAllocation is DATA-05's
// concurrency bar for kernel/document.InitiateUpload (W02-E03-S001 T1, AC-01):
// N concurrent callers must be issued N unique, monotonic version numbers with
// zero unexpected conflicts. Before the locked-counter fix, InitiateUpload
// computed the next version via an inline MAX(version_no)+1 read against
// document_versions — a table InitiateUpload never writes — so EVERY
// overlapping caller was handed the SAME version number and the race was only
// resolved much later, at confirm time, by orphaning the losers' blobs
// (fail-first evidence for this story). With the version_counters allocation
// the callers serialize on the per-document counter row and each receives a
// distinct version number.
//
// The 24 callers are concurrent goroutines; the testkit runtime pool caps
// DB-side concurrency at 4 in-flight transactions, which is enough overlap to
// reproduce the MAX()+1 race and to measure counter-row lock wait
// (RISK-W02-E03-001) after the fix.
func TestIntegrationInitiateUploadConcurrentVersionAllocation(t *testing.T) {
	const callers = 24

	a := newHarness(t, document.Class{Key: "core.doc", MaxBytes: 1 << 20})
	var docID uuid.UUID
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		return e
	})
	if err != nil {
		t.Fatalf("create document: %v", err)
	}

	var (
		start    = make(chan struct{})
		wg       sync.WaitGroup
		mu       sync.Mutex
		versions []int
		errs     []error
		waits    []time.Duration
	)
	wg.Add(callers)
	for range callers {
		go func() {
			defer wg.Done()
			<-start
			began := time.Now()
			var v int
			err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
				sess, e := a.svc.InitiateUpload(ctx, db, docID, sum(nil))
				v = sess.VersionNo
				return e
			})
			took := time.Since(began)
			mu.Lock()
			defer mu.Unlock()
			waits = append(waits, took)
			if err != nil {
				errs = append(errs, err)
				return
			}
			versions = append(versions, v)
		}()
	}
	close(start)
	wg.Wait()

	// Zero unexpected conflicts: every caller must succeed.
	for _, err := range errs {
		t.Errorf("concurrent InitiateUpload failed: %v", err)
	}
	if len(errs) > 0 {
		t.Fatalf("%d of %d concurrent callers failed — version allocation is not race-free", len(errs), callers)
	}

	// N callers → N unique monotonic versions: exactly the set 1..callers.
	sort.Ints(versions)
	if len(versions) != callers {
		t.Fatalf("got %d versions, want %d", len(versions), callers)
	}
	for i, v := range versions {
		if v != i+1 {
			t.Fatalf("version numbers not the contiguous set 1..%d: got %v", callers, versions)
		}
	}

	// Lock-wait measurement (RISK-W02-E03-001).
	var maxW, sum time.Duration
	for _, w := range waits {
		if w > maxW {
			maxW = w
		}
		sum += w
	}
	t.Logf("lock-wait under %d concurrent callers: max=%v avg=%v", callers, maxW, sum/time.Duration(len(waits)))
}

// TestIntegrationUploadSessionDurability proves that InitiateUpload leaves a
// durable pending session row, so a crash between initiate and confirm is
// recoverable and the reserved version number is not lost.
func TestIntegrationUploadSessionDurability(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", MaxBytes: 1 << 20})
	var docID, sessID uuid.UUID
	var versionNo int
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		if e != nil {
			return e
		}
		sess, e := a.svc.InitiateUpload(ctx, db, docID, sum(nil))
		if e != nil {
			return e
		}
		sessID = sess.SessionID
		versionNo = sess.VersionNo
		return nil
	})
	if err != nil {
		t.Fatalf("initiate upload: %v", err)
	}

	var status string
	if err := a.h.Admin.QueryRow(context.Background(),
		`SELECT status FROM document_upload_sessions WHERE id=$1`, sessID).Scan(&status); err != nil {
		t.Fatalf("read session row: %v", err)
	}
	if status != "pending" {
		t.Fatalf("want pending session row, got status=%q", status)
	}
	if versionNo != 1 {
		t.Fatalf("want version 1 for first session, got %d", versionNo)
	}
}

// TestIntegrationConfirmUploadCAS is DATA-05's concurrency bar for
// ConfirmUpload (W02-E03-S001 T4, AC-04): two goroutines confirming the SAME
// upload session must have exactly one winner. The database CAS
// (UPDATE ... WHERE status='pending') serializes the race; the loser sees no
// matching row and receives KindConflict.
func TestIntegrationConfirmUploadCAS(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", MaxBytes: 1 << 20})
	body := []byte("cas payload")
	var docID uuid.UUID
	var sess document.UploadSession
	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		if e != nil {
			return e
		}
		sess, e = a.svc.InitiateUpload(ctx, db, docID, sum(body))
		return e
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	a.store.Put(sess.StorageKey, body)

	var (
		start = make(chan struct{})
		wg    sync.WaitGroup
		mu    sync.Mutex
		ok    int
		conf  int
	)
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			<-start
			err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
				_, e := a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
					SessionID: sess.SessionID, DocumentID: docID, VersionNo: sess.VersionNo, StorageKey: sess.StorageKey,
					DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
				})
				return e
			})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok++
			case kerr.KindOf(err) == kerr.KindConflict:
				conf++
			default:
				t.Errorf("unexpected confirm error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if ok != 1 || conf != 1 {
		t.Fatalf("CAS race: got %d successes and %d conflicts, want exactly 1 of each", ok, conf)
	}
}

// TestIntegrationSweepUploadSessionsAdversarial proves that the GC sweep only
// touches pending sessions whose expiry has passed. Confirmed sessions and
// still-pending sessions with future expiry are left untouched.
func TestIntegrationSweepUploadSessionsAdversarial(t *testing.T) {
	a := newHarness(t, document.Class{Key: "core.doc", MaxBytes: 1 << 20})
	body := []byte("sweep me")

	var (
		docID        uuid.UUID
		confirmed    uuid.UUID
		expired      uuid.UUID
		stillPend    uuid.UUID
		confirmedKey string
		expiredKey   string
		stillPendKey string
	)

	err := a.h.TxM.WithTenant(a.ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		docID, e = a.svc.Create(ctx, db, document.CreateInput{Class: "core.doc", Title: "T"})
		if e != nil {
			return e
		}

		// Confirmed session: full upload flow.
		s1, e := a.svc.InitiateUpload(ctx, db, docID, sum(body))
		if e != nil {
			return e
		}
		a.store.Put(s1.StorageKey, body)
		if _, e = a.svc.ConfirmUpload(ctx, db, document.ConfirmInput{
			SessionID: s1.SessionID, DocumentID: docID, VersionNo: s1.VersionNo, StorageKey: s1.StorageKey,
			DeclaredSize: int64(len(body)), DeclaredChecksum: sum(body), DeclaredMIME: "text/plain",
		}); e != nil {
			return e
		}
		confirmed = s1.SessionID
		confirmedKey = s1.StorageKey

		// Expired session: pending row, but we age its expiry so the sweep picks it up.
		s2, e := a.svc.InitiateUpload(ctx, db, docID, sum(body))
		if e != nil {
			return e
		}
		a.store.Put(s2.StorageKey, body)
		expired = s2.SessionID
		expiredKey = s2.StorageKey

		// Still-pending session: future expiry.
		s3, e := a.svc.InitiateUpload(ctx, db, docID, sum(body))
		if e != nil {
			return e
		}
		a.store.Put(s3.StorageKey, body)
		stillPend = s3.SessionID
		stillPendKey = s3.StorageKey

		return nil
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Age the expired session's expiry to the past. Use Admin to bypass RLS.
	if _, err := a.h.Admin.Exec(context.Background(),
		`UPDATE document_upload_sessions SET expires_at = now() - interval '1 second' WHERE id=$1`, expired); err != nil {
		t.Fatalf("age expired session: %v", err)
	}

	n, err := a.svc.SweepUploadSessions(context.Background(), a.h.PlatformTxM, a.tn, time.Now())
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if n != 1 {
		t.Fatalf("sweep must remove exactly 1 expired session, got %d", n)
	}

	assertStatus := func(id uuid.UUID, want string) {
		var got string
		if err := a.h.Admin.QueryRow(context.Background(),
			`SELECT status FROM document_upload_sessions WHERE id=$1`, id).Scan(&got); err != nil {
			t.Fatalf("read session %s: %v", id, err)
		}
		if got != want {
			t.Fatalf("session %s: want status %q, got %q", id, want, got)
		}
	}
	assertStatus(confirmed, "confirmed")
	assertStatus(expired, "expired")
	assertStatus(stillPend, "pending")

	// The expired session's blob was deleted; the others remain.
	if a.store.Has(expiredKey) {
		t.Fatal("expired session blob must be deleted")
	}
	if !a.store.Has(confirmedKey) {
		t.Fatal("confirmed session blob must remain")
	}
	if !a.store.Has(stillPendKey) {
		t.Fatal("pending session blob must remain")
	}
}
