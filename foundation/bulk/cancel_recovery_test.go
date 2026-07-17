package bulk_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/bulk"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/testkit"
)

// Second closure-audit regressions (2026-07-17, F-04): a terminal CANCELLED
// aggregate must never regain pending items through any recovery path — not a
// retryable worker failure (recordFailure) and not lease reclaim
// (ReclaimStalled). Both previously reset items to 'pending' unconditionally;
// with Process's initial cancelled check returning immediately, those items
// were stranded forever under a terminal aggregate no Cancel retry repairs.

func itemStatus(t *testing.T, h *testkit.DBHandle, ctx context.Context, bulkID uuid.UUID) string {
	t.Helper()
	var status string
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT status FROM bulk_items WHERE bulk_id = $1`, bulkID).Scan(&status)
	}); err != nil {
		t.Fatalf("item status: %v", err)
	}
	return status
}

// A worker holds the only item as running; Cancel commits (running items are
// intentionally left to finish); the worker then fails RETRYABLY under
// limit=1. The failure must land the item in 'cancelled' — not resurrect it as
// 'pending' — without any second Cancel call.
func TestIntegrationRetryableFailureAfterCancelLandsCancelled(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithLeaseTTL(30*time.Second), bulk.WithBatchSize(1))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))

	claimed := make(chan struct{})
	signalClaimed := sync.OnceFunc(func() { close(claimed) })
	release := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	t.Cleanup(unblock) // never leave the worker parked if the test dies early

	var wg sync.WaitGroup
	wg.Add(1)
	var procN int
	var procErr error
	go func() {
		defer wg.Done()
		procN, procErr = svc.Process(context.Background(), h.TxM, tenant, id, 1,
			func(_ context.Context, _ database.TenantDB, _ bulk.Item) error {
				signalClaimed()
				<-release
				return errors.New("transient worker failure") // retryable: attempts stay below the budget
			})
	}()

	select {
	case <-claimed:
	case <-time.After(30 * time.Second):
		t.Fatal("worker never claimed the item")
	}
	// Cancel while the item is running: the aggregate goes terminal, the
	// running item is left to finish under its lease.
	if err := svc.Cancel(context.Background(), h.TxM, tenant, id); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	unblock()
	wg.Wait()
	if procErr != nil {
		t.Fatalf("Process: %v", procErr)
	}
	_ = procN

	// The retryable failure must have observed the cancelled aggregate: the
	// item is cancelled, not pending — and no second Cancel was issued.
	if got := itemStatus(t, h, ctx, id); got != "cancelled" {
		t.Fatalf("item status after retryable failure under a cancelled aggregate = %q, want cancelled", got)
	}
	p := progress(t, h, svc, ctx, id)
	if p.Pending != 0 || p.Status != "cancelled" {
		t.Fatalf("progress = %+v, want Pending0/status cancelled", p)
	}
}

// A worker's lease expires after the aggregate is cancelled. ReclaimStalled
// must transition the expired running item to 'cancelled', never back to
// 'pending'.
func TestIntegrationReclaimAfterCancelYieldsCancelledNeverPending(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithLeaseTTL(50*time.Millisecond), bulk.WithBatchSize(1))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, ctx, payloads(item{Val: 1}))

	claimed := make(chan struct{})
	signalClaimed := sync.OnceFunc(func() { close(claimed) })
	release := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	t.Cleanup(unblock)

	var wg sync.WaitGroup
	wg.Add(1)
	var procErr error
	go func() {
		defer wg.Done()
		_, procErr = svc.Process(context.Background(), h.TxM, tenant, id, 1,
			func(_ context.Context, _ database.TenantDB, _ bulk.Item) error {
				signalClaimed()
				<-release
				return nil
			})
	}()

	select {
	case <-claimed:
	case <-time.After(30 * time.Second):
		t.Fatal("worker never claimed the item")
	}
	if err := svc.Cancel(context.Background(), h.TxM, tenant, id); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	// Poll until the lease has lapsed and reclaim transitions the item (n>=1
	// proves both the claim and the expiry — no fixed-sleep timing).
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
			t.Fatal("expired running item was never reclaimed")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if got := itemStatus(t, h, ctx, id); got != "cancelled" {
		t.Fatalf("reclaimed item under a cancelled aggregate = %q, want cancelled (never pending)", got)
	}

	// The stale worker's finalize is fenced (lease reclaimed) — and the item
	// must STILL be cancelled afterwards.
	unblock()
	wg.Wait()
	if procErr == nil {
		t.Fatal("stale worker finalized successfully, want lease-mismatch error")
	}
	if got := itemStatus(t, h, ctx, id); got != "cancelled" {
		t.Fatalf("item status after stale finalize attempt = %q, want cancelled", got)
	}
}

// waitForShareBlocked polls pg_stat_activity until a backend is lock-waiting
// on the recovery path's FOR SHARE read (i.e., blocked behind Cancel's
// uncommitted bulk_operations row lock). Without the FOR SHARE fix nothing
// ever blocks, the poll times out, and the final status assertion fails on the
// resurrected pending item — the test discriminates both ways.
func waitForShareBlocked(t *testing.T, h *testkit.DBHandle) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for {
		var waiting int
		if err := h.Admin.QueryRow(context.Background(),
			`SELECT count(*) FROM pg_stat_activity
			  WHERE wait_event_type = 'Lock' AND query ILIKE '%FOR SHARE%'`).Scan(&waiting); err != nil {
			t.Fatalf("pg_stat_activity: %v", err)
		}
		if waiting >= 1 {
			return
		}
		if time.Now().After(deadline) {
			t.Log("no FOR SHARE lock-waiter observed; recovery read did not serialize against the in-flight Cancel")
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Second closure-audit verification regression (2026-07-17, F-04): a retryable
// failure recorded DURING Cancel's commit window (aggregate row updated to
// cancelled but not yet committed, item sweep already past) must not resurrect
// the item as pending. recordFailure's aggregate read takes FOR SHARE, so it
// blocks on Cancel's row lock and lands the item in cancelled.
func TestIntegrationRetryableFailureDuringCancelCommitWindow(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithLeaseTTL(30*time.Second), bulk.WithBatchSize(1))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)
	id := start(t, h, svc, ctx, payloads(item{Val: 1}))

	claimed := make(chan struct{})
	signalClaimed := sync.OnceFunc(func() { close(claimed) })
	release := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	t.Cleanup(unblock)

	// Hold Cancel open INSIDE its transaction, AFTER the pending-item sweep and
	// before commit — the exact window the audit named: a pending write landing
	// here is invisible to the already-run sweep, so only the recovery path's
	// own FOR SHARE serialization can prevent the stranded item.
	cancelEntered := make(chan struct{})
	signalEntered := sync.OnceFunc(func() { close(cancelEntered) })
	cancelHold := make(chan struct{})
	releaseCancel := sync.OnceFunc(func() { close(cancelHold) })
	t.Cleanup(releaseCancel)
	bulk.SetCancelCommitInterceptor(svc, func(ctx context.Context, _ database.TenantDB) error {
		signalEntered()
		<-cancelHold
		return nil
	})

	var wg sync.WaitGroup
	wg.Add(1)
	var procErr error
	go func() {
		defer wg.Done()
		_, procErr = svc.Process(context.Background(), h.TxM, tenant, id, 1,
			func(_ context.Context, _ database.TenantDB, _ bulk.Item) error {
				signalClaimed()
				<-release
				return errors.New("transient failure inside the cancel commit window")
			})
	}()
	select {
	case <-claimed:
	case <-time.After(30 * time.Second):
		t.Fatal("worker never claimed the item")
	}

	var cancelErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		cancelErr = svc.Cancel(context.Background(), h.TxM, tenant, id)
	}()
	select {
	case <-cancelEntered:
	case <-time.After(30 * time.Second):
		t.Fatal("Cancel never reached its commit window")
	}

	// Cancel now holds the uncommitted 'cancelled' aggregate write. Release the
	// worker: its recordFailure must BLOCK on the aggregate row, not read
	// around the in-flight cancel.
	unblock()
	waitForShareBlocked(t, h)
	releaseCancel()
	wg.Wait()
	if cancelErr != nil {
		t.Fatalf("Cancel: %v", cancelErr)
	}
	if procErr != nil {
		t.Fatalf("Process: %v", procErr)
	}
	if got := itemStatus(t, h, ctx, id); got != "cancelled" {
		t.Fatalf("item recorded during the cancel commit window = %q, want cancelled (pending would be stranded forever)", got)
	}
	p := progress(t, h, svc, ctx, id)
	if p.Pending != 0 || p.Status != "cancelled" {
		t.Fatalf("progress = %+v, want Pending0/status cancelled", p)
	}
}

// Same commit window, reclaim path: ReclaimStalled running while Cancel is
// mid-commit must not resurrect the expired item as pending.
func TestIntegrationReclaimDuringCancelCommitWindow(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithLeaseTTL(50*time.Millisecond), bulk.WithBatchSize(1))
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)
	id := start(t, h, svc, ctx, payloads(item{Val: 1}))

	claimed := make(chan struct{})
	signalClaimed := sync.OnceFunc(func() { close(claimed) })
	release := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	t.Cleanup(unblock)

	cancelEntered := make(chan struct{})
	signalEntered := sync.OnceFunc(func() { close(cancelEntered) })
	cancelHold := make(chan struct{})
	releaseCancel := sync.OnceFunc(func() { close(cancelHold) })
	t.Cleanup(releaseCancel)
	bulk.SetCancelCommitInterceptor(svc, func(ctx context.Context, _ database.TenantDB) error {
		signalEntered()
		<-cancelHold
		return nil
	})

	var wg sync.WaitGroup
	wg.Add(1)
	var procErr error
	go func() {
		defer wg.Done()
		_, procErr = svc.Process(context.Background(), h.TxM, tenant, id, 1,
			func(_ context.Context, _ database.TenantDB, _ bulk.Item) error {
				signalClaimed()
				<-release
				return nil
			})
	}()
	select {
	case <-claimed:
	case <-time.After(30 * time.Second):
		t.Fatal("worker never claimed the item")
	}
	// Let the 50ms lease lapse so the item is reclaimable.
	time.Sleep(100 * time.Millisecond)

	var cancelErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		cancelErr = svc.Cancel(context.Background(), h.TxM, tenant, id)
	}()
	select {
	case <-cancelEntered:
	case <-time.After(30 * time.Second):
		t.Fatal("Cancel never reached its commit window")
	}

	reclaimDone := make(chan error, 1)
	go func() {
		_, err := svc.ReclaimStalled(context.Background(), h.TxM, tenant, id)
		reclaimDone <- err
	}()
	waitForShareBlocked(t, h)
	releaseCancel()
	if err := <-reclaimDone; err != nil {
		t.Fatalf("ReclaimStalled: %v", err)
	}
	if got := itemStatus(t, h, ctx, id); got != "cancelled" {
		t.Fatalf("item reclaimed during the cancel commit window = %q, want cancelled (never pending)", got)
	}
	unblock()
	wg.Wait()
	if cancelErr != nil {
		t.Fatalf("Cancel: %v", cancelErr)
	}
	if procErr == nil {
		t.Fatal("stale worker finalized successfully, want lease-mismatch error")
	}
	if got := itemStatus(t, h, ctx, id); got != "cancelled" {
		t.Fatalf("item status after stale finalize attempt = %q, want cancelled", got)
	}
}
