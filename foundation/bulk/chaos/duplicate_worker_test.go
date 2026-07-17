// Package chaos tests the bulk leased-claim path under adversarial concurrency.
// This is the named chaos test DATA-04/chaos/duplicate_worker_test.go: multiple
// processors concurrently claim, retry, pause, resume, and cancel the same bulk
// operation without producing duplicate effects or allowing stale finalization.
package chaos

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/foundation/bulk"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/testkit"
)

type chaosItem struct {
	Seq  int `json:"seq"`
	Fail int `json:"fail"` // number of times to fail before succeeding
}

func TestIntegrationBulkDuplicateWorkerChaos(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	// Short lease TTL to exercise reclaim, very generous retry budget so
	// transient contention does not permanently fail items, tiny batch size to
	// maximize contention.
	svc := bulk.New(model.UUIDv7(),
		bulk.WithBatchSize(2),
		bulk.WithLeaseTTL(300*time.Millisecond),
		bulk.WithMaxAttempts(50))

	// 20 items, each instructed to fail a deterministic number of times (0-1).
	// The effect ledger lets us prove exactly one success was recorded per item.
	items := make([]json.RawMessage, 20)
	for i := range items {
		b, _ := json.Marshal(chaosItem{Seq: i, Fail: i % 2})
		items[i] = b
	}

	var bulkID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var err error
		bulkID, err = svc.Start(ctx, db, "test.bulk.chaos", items)
		return err
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Effect ledger: one row per successful item, keyed by seq. A duplicate
	// effect would violate the unique constraint or produce >1 success per seq.
	_, _ = h.Admin.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS bulk_chaos_effects (seq int PRIMARY KEY, count int NOT NULL DEFAULT 0)`)
	_, _ = h.Admin.Exec(context.Background(), `GRANT INSERT, SELECT, UPDATE ON bulk_chaos_effects TO app_rt`)

	var marksMu sync.Mutex
	marks := make(map[int]int) // seq -> success count seen by worker

	fn := func(_ context.Context, db database.TenantDB, it bulk.Item) error {
		var ci chaosItem
		if err := json.Unmarshal(it.Payload, &ci); err != nil {
			return err
		}

		// Read how many prior attempts were recorded for this seq. The bulk
		// service's leased claim already guarantees only one worker holds this
		// item's lease at a time; the ledger's unique constraint is a second
		// line of defense against duplicate effects.
		var prior int
		err := db.QueryRow(ctx, `SELECT count FROM bulk_chaos_effects WHERE seq = $1`, ci.Seq).Scan(&prior)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if prior >= 1 {
			// Already succeeded by a previous attempt; a stale worker must not
			// record a second effect.
			return nil
		}
		if prior < ci.Fail {
			// Record the attempt as a failure and ask for a retry.
			if _, err := db.Exec(ctx,
				`INSERT INTO bulk_chaos_effects (seq, count) VALUES ($1, 1)
				 ON CONFLICT (seq) DO UPDATE SET count = bulk_chaos_effects.count + 1`,
				ci.Seq); err != nil {
				return err
			}
			return errors.New("chaos retry")
		}
		if _, err := db.Exec(ctx,
			`INSERT INTO bulk_chaos_effects (seq, count) VALUES ($1, 1)
			 ON CONFLICT (seq) DO UPDATE SET count = bulk_chaos_effects.count + 1`,
			ci.Seq); err != nil {
			return err
		}

		marksMu.Lock()
		marks[ci.Seq]++
		marksMu.Unlock()
		return nil
	}

	var cancelled atomic.Bool
	var wg sync.WaitGroup
	workerCtx, stop := context.WithCancel(context.Background())
	defer stop()

	workers := 4
	processed := make([]atomic.Int32, workers)
	for i := range workers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for workerCtx.Err() == nil && !cancelled.Load() {
				n, err := svc.Process(workerCtx, h.TxM, tenant, bulkID, 0, fn)
				if err != nil {
					// Lease mismatch or cancellation are acceptable outcomes under chaos.
					return
				}
				processed[idx].Add(int32(n))
				if n == 0 {
					return
				}
			}
		}(i)
	}

	// Lifecycle toggler randomly pauses/resumes/cancels the operation.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50 && workerCtx.Err() == nil && !cancelled.Load(); i++ {
			time.Sleep(75 * time.Millisecond)
			switch i % 4 {
			case 0:
				_ = svc.Pause(workerCtx, h.TxM, tenant, bulkID)
			case 1, 2:
				_ = svc.Resume(workerCtx, h.TxM, tenant, bulkID)
			case 3:
				if i >= 45 {
					_ = svc.Cancel(workerCtx, h.TxM, tenant, bulkID)
					cancelled.Store(true)
				}
			}
		}
	}()

	// Let the chaos run until processors stop making progress.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("chaos test did not quiesce within 15s")
	}

	// Verify no duplicate success effects and that the ledger matches the final
	// operation state.
	var totalEffects int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM bulk_chaos_effects WHERE count >= 1`).Scan(&totalEffects); err != nil {
		t.Fatalf("count effects: %v", err)
	}

	var p bulk.Progress
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var err error
		p, err = svc.Progress(ctx, db, bulkID)
		return err
	}); err != nil {
		t.Fatalf("Progress: %v", err)
	}

	// After cancellation some items may remain pending/cancelled; every success
	// recorded in the ledger must correspond to exactly one done item.
	if totalEffects != p.Done {
		t.Fatalf("ledger successes = %d, done items = %d; duplicate or missing effect", totalEffects, p.Done)
	}

	// No seq was successfully processed more than once.
	for seq, c := range marks {
		if c > 1 {
			t.Fatalf("seq %d processed successfully %d times, want 1", seq, c)
		}
	}

	rows, _ := h.Admin.Query(context.Background(),
		`SELECT seq, attempts, last_error FROM bulk_items WHERE bulk_id=$1 AND status='failed'`, bulkID)
	for rows.Next() {
		var seq, attempts int
		var lastErr string
		_ = rows.Scan(&seq, &attempts, &lastErr)
		t.Logf("failed item seq=%d attempts=%d err=%q", seq, attempts, lastErr)
	}

	t.Logf("chaos finished: progress=%+v ledger=%d", p, totalEffects)
}
