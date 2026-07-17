package jobs_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/jobs"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationSchedulerOnRunAndTaskError proves the OnRun observer fires after
// a claimed run with the observed lag and the task's own error (Tick's
// task-failed branch), and that a task error is non-fatal.
func TestIntegrationSchedulerOnRunAndTaskError(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	boom := errors.New("task boom")
	var calls int32
	var gotName string
	var gotLag time.Duration
	var gotErr error

	s := jobs.NewScheduler(h.Platform, nil)
	s.OnRun(func(name string, lag time.Duration, err error) {
		atomic.AddInt32(&calls, 1)
		gotName, gotLag, gotErr = name, lag, err
	})
	s.Register("test.sched.err", time.Hour, func(context.Context) error { return boom })
	if err := s.Ensure(ctx); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	s.Tick(ctx) // due immediately (next_run_at = now())

	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Fatalf("OnRun called %d times, want 1", n)
	}
	if gotName != "test.sched.err" {
		t.Fatalf("OnRun name = %q, want test.sched.err", gotName)
	}
	if !errors.Is(gotErr, boom) {
		t.Fatalf("OnRun err = %v, want task boom", gotErr)
	}
	if gotLag < 0 {
		t.Fatalf("OnRun lag = %v, want non-negative", gotLag)
	}
}

// TestIntegrationSchedulerRegisterClampsInterval proves Register clamps a
// sub-second interval up to 1s (persisted as interval_seconds = 1 by Ensure).
func TestIntegrationSchedulerRegisterClampsInterval(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	s := jobs.NewScheduler(h.Platform, nil)
	s.Register("test.sched.clamp", 10*time.Millisecond, func(context.Context) error { return nil })
	if err := s.Ensure(ctx); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	var secs int
	if err := h.Platform.QueryRow(ctx,
		`SELECT interval_seconds FROM schedules WHERE name = $1`, "test.sched.clamp").Scan(&secs); err != nil {
		t.Fatalf("read interval_seconds: %v", err)
	}
	if secs != 1 {
		t.Fatalf("interval_seconds = %d, want 1 (clamped from 10ms)", secs)
	}
}

// TestSchedulerTickCancelledCtxDoesNotRun proves Tick returns without running any
// task once the context is already cancelled.
func TestSchedulerTickCancelledCtxDoesNotRun(t *testing.T) {
	h := testkit.NewDB(t)

	var runs int32
	s := jobs.NewScheduler(h.Platform, nil)
	s.Register("test.sched.cancelled", time.Hour, func(context.Context) error {
		atomic.AddInt32(&runs, 1)
		return nil
	})
	if err := s.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.Tick(ctx)

	if n := atomic.LoadInt32(&runs); n != 0 {
		t.Fatalf("task ran %d times under a cancelled ctx, want 0", n)
	}
}

// TestSchedulerRunDefaultPollThenCancel covers Scheduler.Run's poll<=0 default
// (falls back to the 30s poll) and its clean return on cancel after the first
// immediate Tick.
func TestSchedulerRunDefaultPollThenCancel(t *testing.T) {
	h := testkit.NewDB(t)

	var runs int32
	s := jobs.NewScheduler(h.Platform, nil)
	s.Register("test.sched.defaultpoll", time.Hour, func(context.Context) error {
		atomic.AddInt32(&runs, 1)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx, 0) }() // poll <= 0 → default 30s

	// The first Tick runs synchronously before the loop blocks on the ticker.
	deadline := time.After(5 * time.Second)
	for atomic.LoadInt32(&runs) == 0 {
		select {
		case <-deadline:
			cancel()
			t.Fatal("first Tick never ran")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

// TestIntegrationSchedulerRunUntilCancel drives Scheduler.Run in a goroutine
// (Ensure + poll loop), waits for at least one run, then cancels and asserts Run
// returns nil.
func TestIntegrationSchedulerRunUntilCancel(t *testing.T) {
	h := testkit.NewDB(t)

	var runs int32
	s := jobs.NewScheduler(h.Platform, nil)
	s.Register("test.sched.runloop", time.Hour, func(context.Context) error {
		atomic.AddInt32(&runs, 1)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx, 10*time.Millisecond) }()

	deadline := time.After(5 * time.Second)
	for atomic.LoadInt32(&runs) == 0 {
		select {
		case <-deadline:
			cancel()
			t.Fatal("scheduler task never ran")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
	if n := atomic.LoadInt32(&runs); n < 1 {
		t.Fatalf("task ran %d times, want >= 1", n)
	}
}
