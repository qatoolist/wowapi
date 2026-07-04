package jobs_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationSchedulerRunsDueTask proves a registered task runs when due and
// reports a non-negative lag.
func TestIntegrationSchedulerRunsDueTask(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	var runs int32
	s := jobs.NewScheduler(h.Platform, nil)
	s.Register("test.sched.run", time.Hour, func(context.Context) error {
		atomic.AddInt32(&runs, 1)
		return nil
	})
	if err := s.Ensure(ctx); err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	// next_run_at defaults to now(), so the task is immediately due.
	s.Tick(ctx)
	if got := atomic.LoadInt32(&runs); got != 1 {
		t.Fatalf("task ran %d times, want 1", got)
	}
	// A second immediate Tick must NOT re-run it (next_run_at advanced by 1h).
	s.Tick(ctx)
	if got := atomic.LoadInt32(&runs); got != 1 {
		t.Fatalf("task re-ran after claim advanced the schedule: %d", got)
	}
}

// TestIntegrationSchedulerLeaderSafe is the R3 regression: N replicas all ticking
// a due task must run it exactly ONCE per interval (the atomic claim elects a
// single winner), never once-per-replica.
func TestIntegrationSchedulerLeaderSafe(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	var runs int32
	const replicas = 6
	scheds := make([]*jobs.Scheduler, replicas)
	for i := range scheds {
		s := jobs.NewScheduler(h.Platform, nil)
		s.Register("test.sched.leader", time.Hour, func(context.Context) error {
			atomic.AddInt32(&runs, 1)
			return nil
		})
		scheds[i] = s
	}
	// One replica creates the (due) schedule row; all share it.
	if err := scheds[0].Ensure(ctx); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	// All replicas tick the same due task at once.
	var wg sync.WaitGroup
	wg.Add(replicas)
	for _, s := range scheds {
		go func() { defer wg.Done(); s.Tick(ctx) }()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&runs); got != 1 {
		t.Fatalf("due task ran %d times across %d replicas; must be exactly 1 (leader-safe)", got, replicas)
	}
}
