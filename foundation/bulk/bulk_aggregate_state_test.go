package bulk_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/v2/foundation/bulk"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// F-04 regression (adversarial-framework-review-2026-07-17): an operation is
// complete only when NO pending or running item exists. Before the fix, worker
// B's empty claim (while worker A held the sole item's live lease) marked the
// aggregate 'completed' unconditionally; A's retryable failure then returned
// the item to pending under a terminal aggregate — permanently stranded.
func TestIntegrationBulkEmptyClaimWithLivePeerDoesNotComplete(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithBatchSize(1), bulk.WithMaxAttempts(3))
	tenant := uuid.New()
	tctx := database.WithTenantID(context.Background(), tenant)

	id := start(t, h, svc, tctx, payloads(item{Val: 1}))

	claimed := make(chan struct{})
	release := make(chan struct{})
	// Any Fatal below must still unblock worker A, or its open tenant tx wedges
	// the testkit teardown.
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	first := true
	workerA := make(chan error, 1)
	go func() {
		_, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, func(ctx context.Context, db database.TenantDB, it bulk.Item) error {
			if first {
				first = false
				close(claimed)
				<-release
				return errors.New("retryable business failure")
			}
			return nil
		})
		workerA <- err
	}()

	// Wait for A to own the item's live lease inside its callback — but never
	// block forever: if A's Process returns early (a claim error / connection
	// starvation under heavy CI load) it never closes `claimed`, so also select
	// on A's completion and a bounded deadline. An unguarded receive here would
	// hang the whole package to the 10-minute go-test timeout.
	select {
	case <-claimed:
	case err := <-workerA:
		t.Fatalf("worker A returned before claiming the item (err=%v); cannot exercise the live-peer race", err)
	case <-time.After(30 * time.Second):
		t.Fatal("worker A did not claim the item within 30s")
	}

	// Worker B: empty claim while A's running item is live. Must NOT complete.
	if n, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); err != nil || n != 0 {
		t.Fatalf("worker B: n=%d err=%v, want 0/nil", n, err)
	}
	if p := progress(t, h, svc, tctx, id); p.Status == "completed" {
		t.Fatalf("operation marked completed while a peer holds a live running item (progress %+v)", p)
	}

	select {
	case <-release:
	default:
		close(release)
	}
	if err := <-workerA; err != nil {
		t.Fatalf("worker A: %v", err)
	}

	// The failed item must be retryable — a later Process finishes the work and
	// only THEN does the aggregate complete.
	deadline := time.After(10 * time.Second)
	for {
		n, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, func(ctx context.Context, db database.TenantDB, it bulk.Item) error { return nil })
		if err != nil {
			t.Fatalf("retry process: %v", err)
		}
		p := progress(t, h, svc, tctx, id)
		if p.Status == "completed" {
			if p.Done != 1 {
				t.Fatalf("operation completed with the item stranded: %+v", p)
			}
			break
		}
		_ = n
		select {
		case <-deadline:
			t.Fatalf("item stranded: final progress %+v", p)
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// F-04: lifecycle transitions are compare-and-swap with legal source states;
// terminal states cannot be reopened; unknown operations are not-found.
func TestIntegrationBulkLifecycleTransitionMatrix(t *testing.T) {
	h := testkit.NewDB(t)
	svc := bulk.New(model.UUIDv7(), bulk.WithBatchSize(1))
	tenant := uuid.New()
	tctx := database.WithTenantID(context.Background(), tenant)

	newOp := func() uuid.UUID { return start(t, h, svc, tctx, payloads(item{Val: 1})) }
	complete := func(id uuid.UUID) {
		if _, err := svc.Process(context.Background(), h.TxM, tenant, id, 0, okFunc); err != nil {
			t.Fatalf("complete op: %v", err)
		}
		if p := progress(t, h, svc, tctx, id); p.Status != "completed" {
			t.Fatalf("setup: op not completed: %+v", p)
		}
	}

	t.Run("resume of completed is rejected and stays completed", func(t *testing.T) {
		id := newOp()
		complete(id)
		if err := svc.Resume(context.Background(), h.TxM, tenant, id); err == nil {
			t.Fatal("Resume of a completed operation returned nil")
		}
		if p := progress(t, h, svc, tctx, id); p.Status != "completed" {
			t.Fatalf("completed operation relabelled to %q by Resume", p.Status)
		}
	})

	t.Run("cancel of completed is rejected", func(t *testing.T) {
		id := newOp()
		complete(id)
		if err := svc.Cancel(context.Background(), h.TxM, tenant, id); err == nil {
			t.Fatal("Cancel of a completed operation returned nil")
		}
		if p := progress(t, h, svc, tctx, id); p.Status != "completed" {
			t.Fatalf("completed operation relabelled to %q by Cancel", p.Status)
		}
	})

	t.Run("pause of cancelled is rejected", func(t *testing.T) {
		id := newOp()
		if err := svc.Cancel(context.Background(), h.TxM, tenant, id); err != nil {
			t.Fatalf("Cancel: %v", err)
		}
		if err := svc.Pause(context.Background(), h.TxM, tenant, id); err == nil {
			t.Fatal("Pause of a cancelled operation returned nil")
		}
	})

	t.Run("resume of running (never paused) is rejected", func(t *testing.T) {
		id := newOp()
		// pending -> resume is also illegal: resume's only legal source is paused.
		if err := svc.Resume(context.Background(), h.TxM, tenant, id); err == nil {
			t.Fatal("Resume of a never-paused operation returned nil")
		}
	})

	t.Run("unknown operation id is not-found, not nil", func(t *testing.T) {
		ghost := uuid.New()
		for name, err := range map[string]error{
			"Pause":  svc.Pause(context.Background(), h.TxM, tenant, ghost),
			"Resume": svc.Resume(context.Background(), h.TxM, tenant, ghost),
			"Cancel": svc.Cancel(context.Background(), h.TxM, tenant, ghost),
		} {
			if err == nil {
				t.Fatalf("%s(unknown id) returned nil", name)
			}
			if kerr.KindOf(err) != kerr.KindNotFound {
				t.Fatalf("%s(unknown id) kind = %v, want KindNotFound (err=%v)", name, kerr.KindOf(err), err)
			}
		}
	})

	t.Run("legal pause and resume still work", func(t *testing.T) {
		id := newOp()
		if err := svc.Pause(context.Background(), h.TxM, tenant, id); err != nil {
			t.Fatalf("Pause(pending): %v", err)
		}
		if err := svc.Resume(context.Background(), h.TxM, tenant, id); err != nil {
			t.Fatalf("Resume(paused): %v", err)
		}
	})
}
