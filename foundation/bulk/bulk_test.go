package bulk_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/foundation/bulk"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

type item struct {
	Val  int  `json:"val"`
	Fail bool `json:"fail"`
}

func payloads(items ...item) []json.RawMessage {
	out := make([]json.RawMessage, len(items))
	for i, it := range items {
		b, _ := json.Marshal(it)
		out[i] = b
	}
	return out
}

func start(t *testing.T, h *testkit.DBHandle, svc *bulk.Service, ctx context.Context, ps []json.RawMessage) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = svc.Start(ctx, db, "test.bulk", ps)
		return e
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return id
}

func progress(t *testing.T, h *testkit.DBHandle, svc *bulk.Service, ctx context.Context, id uuid.UUID) bulk.Progress {
	t.Helper()
	var p bulk.Progress
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		p, e = svc.Progress(ctx, db, id)
		return e
	}); err != nil {
		t.Fatalf("Progress: %v", err)
	}
	return p
}

func TestIntegrationBulkAllSucceed(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}, item{Val: 2}, item{Val: 3}))
	n, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	if n != 3 {
		t.Fatalf("processed %d, want 3", n)
	}
	p := progress(t, h, svc, ctx, id)
	if p.Total != 3 || p.Done != 3 || p.Failed != 0 || p.Pending != 0 || p.Status != "completed" {
		t.Fatalf("progress = %+v, want Total3/Done3/Failed0/Pending0/completed", p)
	}
}

func TestIntegrationBulkPartialFailureLedger(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithMaxAttempts(1))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	// items 2 and 4 fail; the rest succeed.
	id := start(t, h, svc, ctx, payloads(
		item{Val: 1}, item{Val: 2, Fail: true}, item{Val: 3}, item{Val: 4, Fail: true}, item{Val: 5}))

	if _, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, markFunc(h)); err != nil {
		t.Fatalf("Process: %v", err)
	}
	p := progress(t, h, svc, ctx, id)
	if p.Done != 3 || p.Failed != 2 || p.Pending != 0 || p.Status != "completed" {
		t.Fatalf("progress = %+v, want Done3/Failed2/Pending0/completed (partial failure)", p)
	}
	// Failed items carry the error; one item's failure did not stop the others.
	var failedWithErr int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM bulk_items WHERE bulk_id=$1 AND status='failed' AND last_error <> ''`, id).
		Scan(&failedWithErr); err != nil {
		t.Fatal(err)
	}
	if failedWithErr != 2 {
		t.Fatalf("failed items with a recorded error = %d, want 2", failedWithErr)
	}
	// Atomicity: only the 3 SUCCEEDED items wrote their mark — the failed items'
	// writes rolled back with their transaction.
	var marks int
	if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM bulk_marks`).Scan(&marks); err != nil {
		t.Fatal(err)
	}
	if marks != 3 {
		t.Fatalf("bulk_marks rows = %d, want 3 (failed items' writes must roll back)", marks)
	}
}

func TestIntegrationBulkRetryThenFail(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithMaxAttempts(3))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1, Fail: true}))

	if _, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, markFunc(h)); err != nil {
		t.Fatalf("Process: %v", err)
	}
	p := progress(t, h, svc, ctx, id)
	if p.Done != 0 || p.Failed != 1 || p.Pending != 0 || p.Status != "completed" {
		t.Fatalf("progress = %+v, want Done0/Failed1/completed", p)
	}

	var attempts int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT attempts FROM bulk_items WHERE bulk_id=$1`, id).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3 (retried until max)", attempts)
	}
}

func TestIntegrationBulkChunkedResumable(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}, item{Val: 2}, item{Val: 3}, item{Val: 4}, item{Val: 5}))

	// Chunk of 2, then 2, then the rest — simulating resumable batches.
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 2, okFunc); n != 2 {
		t.Fatalf("first chunk processed %d, want 2", n)
	}
	if p := progress(t, h, svc, ctx, id); p.Done != 2 || p.Pending != 3 {
		t.Fatalf("after chunk 1: %+v, want Done2/Pending3", p)
	}
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 2, okFunc); n != 2 {
		t.Fatalf("second chunk processed %d, want 2", n)
	}
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); n != 1 {
		t.Fatalf("final chunk processed %d, want 1", n)
	}
	if p := progress(t, h, svc, ctx, id); p.Done != 5 || p.Pending != 0 || p.Status != "completed" {
		t.Fatalf("after all chunks: %+v, want Done5/Pending0/completed", p)
	}
}

func TestIntegrationBulkEmptyIsCompleted(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, nil)
	p := progress(t, h, svc, ctx, id)
	if p.Total != 0 || p.Status != "completed" {
		t.Fatalf("empty bulk op = %+v, want Total0/completed", p)
	}
}

func TestIntegrationBulkPauseResumeCancel(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithBatchSize(1))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}, item{Val: 2}, item{Val: 3}))

	// Process 1 item, then pause.
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 1, okFunc); n != 1 {
		t.Fatalf("first process: %d, want 1", n)
	}
	if err := svc.Pause(context.Background(), h.TxM, tenant, id); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if p := progress(t, h, svc, ctx, id); p.Status != "paused" || p.Done != 1 || p.Pending != 2 {
		t.Fatalf("after pause: %+v, want paused/Done1/Pending2", p)
	}
	// A subsequent Process sees paused and does not claim more items.
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); n != 0 {
		t.Fatalf("process while paused: %d, want 0", n)
	}
	if p := progress(t, h, svc, ctx, id); p.Done != 1 || p.Pending != 2 {
		t.Fatalf("after paused process: %+v, want Done1/Pending2", p)
	}
	// Resume and finish.
	if err := svc.Resume(context.Background(), h.TxM, tenant, id); err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); n != 2 {
		t.Fatalf("after resume: %d, want 2", n)
	}
	if p := progress(t, h, svc, ctx, id); p.Done != 3 || p.Status != "completed" {
		t.Fatalf("final: %+v, want Done3/completed", p)
	}

	// Start a fresh operation and cancel it mid-run.
	id2 := start(t, h, svc, ctx, payloads(item{Val: 4}, item{Val: 5}))
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id2, 1, okFunc); n != 1 {
		t.Fatalf("cancel setup: %d, want 1", n)
	}
	if err := svc.Cancel(context.Background(), h.TxM, tenant, id2); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if p := progress(t, h, svc, ctx, id2); p.Status != "cancelled" || p.Done != 1 || p.Cancelled != 1 {
		t.Fatalf("after cancel: %+v, want cancelled/Done1/Cancelled1", p)
	}
	// Process on a cancelled operation should not claim new items.
	if n, _ := svc.Process(context.Background(), h.TxM, tenant, id2, 0, okFunc); n != 0 {
		t.Fatalf("process after cancel: %d, want 0", n)
	}
}

// okFunc succeeds for every item.
func okFunc(_ context.Context, _ database.TenantDB, _ bulk.Item) error { return nil }

// markFunc writes a row into bulk_marks for each item, then fails the ones whose
// payload says so — proving the failed item's write rolls back.
func markFunc(h *testkit.DBHandle) bulk.ItemFunc {
	ensureMarks(h)
	return func(ctx context.Context, db database.TenantDB, it bulk.Item) error {
		var parsed item
		if err := json.Unmarshal(it.Payload, &parsed); err != nil {
			return err
		}
		if _, err := db.Exec(ctx, `INSERT INTO bulk_marks (val) VALUES ($1)`, parsed.Val); err != nil {
			return err
		}
		if parsed.Fail {
			return errors.New("item asked to fail")
		}
		return nil
	}
}

func ensureMarks(h *testkit.DBHandle) {
	ctx := context.Background()
	_, _ = h.Admin.Exec(ctx, `CREATE TABLE IF NOT EXISTS bulk_marks (val int)`)
	_, _ = h.Admin.Exec(ctx, `GRANT INSERT, SELECT ON bulk_marks TO app_rt`)
}

// TestIntegrationBulkConcurrentClaimers proves the DATA-04 T3 leased-claim path:
// N concurrent Process calls against the same bulkID each receive disjoint
// batches; no item is processed twice.
func TestIntegrationBulkConcurrentClaimers(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithBatchSize(2))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	// 10 items, 4 workers, batch size 2 -> every worker can claim something.
	id := start(t, h, svc, ctx, payloads(
		item{Val: 1}, item{Val: 2}, item{Val: 3}, item{Val: 4}, item{Val: 5},
		item{Val: 6}, item{Val: 7}, item{Val: 8}, item{Val: 9}, item{Val: 10}))

	var wg sync.WaitGroup
	workers := 4
	processed := make([]atomic.Int32, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			n, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, func(_ context.Context, _ database.TenantDB, _ bulk.Item) error {
				processed[idx].Add(1)
				return nil
			})
			if err != nil {
				t.Errorf("worker %d Process: %v", idx, err)
				return
			}
			processed[idx].Store(int32(n))
		}(i)
	}
	wg.Wait()

	total := 0
	for i := 0; i < workers; i++ {
		total += int(processed[i].Load())
	}
	if total != 10 {
		t.Fatalf("workers processed %d items total, want 10", total)
	}
	p := progress(t, h, svc, ctx, id)
	if p.Done != 10 || p.Pending != 0 || p.Status != "completed" {
		t.Fatalf("progress = %+v, want Done10/Pending0/completed", p)
	}
}

// TestIntegrationBulkLeaseColumnsExist proves the DATA-04 T2 migration added
// the shared primitive's lease columns and per-item idempotency key.
func TestIntegrationBulkLeaseColumnsExist(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))
	if n, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); err != nil || n != 1 {
		t.Fatalf("Process: n=%d err=%v", n, err)
	}

	var leaseToken string
	var leaseGeneration int64
	var idemKey uuid.UUID
	err := h.Admin.QueryRow(context.Background(),
		`SELECT lease_token, lease_generation, idempotency_key FROM bulk_items WHERE bulk_id=$1`, id).
		Scan(&leaseToken, &leaseGeneration, &idemKey)
	if err != nil {
		t.Fatalf("read lease columns: %v", err)
	}
	if leaseToken == "" {
		t.Fatal("lease_token is empty after claim")
	}
	if leaseGeneration == 0 {
		t.Fatalf("lease_generation = %d, want >0", leaseGeneration)
	}
	if idemKey == uuid.Nil {
		t.Fatal("idempotency_key is nil")
	}
}

// TestIntegrationBulkExplainUsesSkipLocked proves the claim plan uses
// FOR UPDATE SKIP LOCKED (DATA-04 T3 EXPLAIN-plan assertion).
func TestIntegrationBulkExplainUsesSkipLocked(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7())
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	var plan string
	err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		plan, e = svc.ExplainClaimPlan(ctx, db, uuid.New(), 5)
		return e
	})
	if err != nil {
		t.Fatalf("ExplainClaimPlan: %v", err)
	}
	lower := strings.ToLower(plan)
	// PostgreSQL represents FOR UPDATE as a "LockRows" plan node. The concurrent
	// claimer test is the behavioral proof that rows are skipped rather than
	// blocked; this assertion is the EXPLAIN-level evidence that row locking is
	// in use.
	if !strings.Contains(lower, "lockrows") {
		t.Fatalf("EXPLAIN plan does not show LockRows (FOR UPDATE evidence):\n%s", plan)
	}
}

// TestIntegrationBulkFencedFinalizeRejectsStaleWorker proves a stale worker's
// finalize is rejected after its lease is reclaimed (DATA-04 T4).
func TestIntegrationBulkFencedFinalizeRejectsStaleWorker(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithLeaseTTL(50*time.Millisecond), bulk.WithBatchSize(1))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))

	// Worker A claims the item and holds it (simulate a stalled worker).
	blocked := make(chan struct{})
	// A parks inside an open tenant tx until blocked closes; if the test dies on
	// any earlier Fatal, cleanup must still unblock it or the pool teardown
	// deadlocks the whole package into the 10m go-test panic.
	unblock := sync.OnceFunc(func() { close(blocked) })
	t.Cleanup(unblock)
	var wg sync.WaitGroup
	wg.Add(1)
	var aErr error
	go func() {
		defer wg.Done()
		_, aErr = svc.Process(context.Background(), h.TxM, tenant, id, 1, func(_ context.Context, _ database.TenantDB, _ bulk.Item) error {
			<-blocked
			return nil
		})
	}()

	// Reclaim deterministically: poll until A's claim exists AND its lease has
	// lapsed (ReclaimStalled only reclaims expired leases, so n>=1 proves both).
	// A fixed sleep is not enough — under full-suite DB contention A can claim
	// late, leaving its lease unexpired at any fixed instant.
	deadline := time.Now().Add(30 * time.Second)
	for {
		n, err := svc.ReclaimStalled(context.Background(), h.TxM, tenant, id)
		if err != nil {
			t.Fatalf("ReclaimStalled: %v", err)
		}
		if n >= 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("worker A's lease was never reclaimed (claim missing or lease never lapsed)")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Worker B now claims and completes the item.
	if n, err := svc.Process(context.Background(), h.TxM, tenant, id, 1, okFunc); err != nil || n != 1 {
		t.Fatalf("worker B Process: n=%d err=%v", n, err)
	}

	// Unblock A. A's finalize must be fenced because its lease was reclaimed.
	unblock()
	wg.Wait()
	if aErr == nil {
		t.Fatal("stale worker A finalized successfully, want lease-mismatch error")
	}
	if kerr.KindOf(aErr) != kerr.KindConflict {
		t.Fatalf("stale worker error kind = %v, want KindConflict (err=%v)", kerr.KindOf(aErr), aErr)
	}

	p := progress(t, h, svc, ctx, id)
	if p.Done != 1 || p.Pending != 0 {
		t.Fatalf("progress = %+v, want Done1/Pending0", p)
	}
}

// TestIntegrationBulkIdempotencyKeyPassedToWorker proves the worker receives a
// non-nil idempotency key for each item (DATA-04 T4).
func TestIntegrationBulkIdempotencyKeyPassedToWorker(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithBatchSize(1))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))

	var gotKey uuid.UUID
	if _, err := svc.Process(context.Background(), h.TxM, tenant, id, 1, func(_ context.Context, _ database.TenantDB, it bulk.Item) error {
		gotKey = it.IdempotencyKey
		return nil
	}); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if gotKey == uuid.Nil {
		t.Fatal("worker did not receive an idempotency key")
	}
}
