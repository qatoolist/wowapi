package app

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestRunHooks_StartOrder verifies hooks start in declaration order.
func TestRunHooks_StartOrder(t *testing.T) {
	var mu sync.Mutex
	var order []string
	record := func(name string) {
		mu.Lock()
		order = append(order, name)
		mu.Unlock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so RunHooks returns after all starts

	hooks := []Hook{
		{Name: "a", Start: func(context.Context) error { record("a"); return nil }},
		{Name: "b", Start: func(context.Context) error { record("b"); return nil }},
		{Name: "c", Start: func(context.Context) error { record("c"); return nil }},
	}
	if err := RunHooks(ctx, slog.Default(), time.Second, hooks...); err != nil {
		t.Fatal(err)
	}
	if strings.Join(order, ",") != "a,b,c" {
		t.Errorf("start order = %v, want a,b,c", order)
	}
}

// TestRunHooks_ReverseStopOrder verifies hooks stop in reverse start order.
func TestRunHooks_ReverseStopOrder(t *testing.T) {
	var mu sync.Mutex
	var stopOrder []string
	record := func(name string) {
		mu.Lock()
		stopOrder = append(stopOrder, name)
		mu.Unlock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	hooks := []Hook{
		{
			Name:  "a",
			Start: func(context.Context) error { return nil },
			Stop:  func(context.Context) error { record("a"); return nil },
		},
		{
			Name:  "b",
			Start: func(context.Context) error { return nil },
			Stop:  func(context.Context) error { record("b"); return nil },
		},
		{
			Name:  "c",
			Start: func(context.Context) error { return nil },
			Stop:  func(context.Context) error { record("c"); return nil },
		},
	}
	if err := RunHooks(ctx, slog.Default(), time.Second, hooks...); err != nil {
		t.Fatal(err)
	}
	if strings.Join(stopOrder, ",") != "c,b,a" {
		t.Errorf("stop order = %v, want c,b,a", stopOrder)
	}
}

// TestRunHooks_CtxCancelTriggerShutdown verifies that cancelling the run
// context triggers a clean shutdown of all started hooks.
func TestRunHooks_CtxCancelTriggerShutdown(t *testing.T) {
	var mu sync.Mutex
	var stopped []string

	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	hooks := []Hook{
		{
			Name: "svc",
			Start: func(context.Context) error {
				close(started)
				return nil
			},
			Stop: func(context.Context) error {
				mu.Lock()
				stopped = append(stopped, "svc")
				mu.Unlock()
				return nil
			},
		},
	}

	done := make(chan error, 1)
	go func() { done <- RunHooks(ctx, slog.Default(), time.Second, hooks...) }()

	<-started // ensure Start was called before we cancel
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHooks returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunHooks did not return after context cancel")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(stopped) != 1 || stopped[0] != "svc" {
		t.Errorf("expected svc to be stopped, got: %v", stopped)
	}
}

// TestRunHooks_StartFailureStopsOnlyStartedHooks verifies that when a hook's
// Start fails, only the already-started hooks are stopped (not the failed one
// or any hooks that were never started).
func TestRunHooks_StartFailureStopsOnlyStartedHooks(t *testing.T) {
	var mu sync.Mutex
	var stopped []string
	boom := errors.New("start exploded")

	ctx := context.Background()
	hooks := []Hook{
		{
			Name:  "first",
			Start: func(context.Context) error { return nil },
			Stop:  func(context.Context) error { mu.Lock(); stopped = append(stopped, "first"); mu.Unlock(); return nil },
		},
		{
			Name:  "second",
			Start: func(context.Context) error { return boom },
			Stop:  func(context.Context) error { mu.Lock(); stopped = append(stopped, "second"); mu.Unlock(); return nil },
		},
		{
			Name:  "third",
			Start: func(context.Context) error { return nil },
			Stop:  func(context.Context) error { mu.Lock(); stopped = append(stopped, "third"); mu.Unlock(); return nil },
		},
	}

	err := RunHooks(ctx, slog.Default(), time.Second, hooks...)
	if !errors.Is(err, boom) {
		t.Fatalf("RunHooks() error = %v, want to contain %v", err, boom)
	}

	mu.Lock()
	defer mu.Unlock()
	// Only "first" should be stopped; "second" failed to start; "third" never started.
	if len(stopped) != 1 || stopped[0] != "first" {
		t.Errorf("stopped = %v, want [first]", stopped)
	}
}

// TestRunHooks_StopErrorsJoined verifies that all Stop errors are collected
// and joined into the returned error, not short-circuited.
func TestRunHooks_StopErrorsJoined(t *testing.T) {
	errA := errors.New("stop-a failed")
	errB := errors.New("stop-b failed")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	hooks := []Hook{
		{
			Name:  "a",
			Start: func(context.Context) error { return nil },
			Stop:  func(context.Context) error { return errA },
		},
		{
			Name:  "b",
			Start: func(context.Context) error { return nil },
			Stop:  func(context.Context) error { return errB },
		},
	}

	err := RunHooks(ctx, slog.Default(), time.Second, hooks...)
	if err == nil {
		t.Fatal("RunHooks() should return an error when stops fail")
	}
	if !errors.Is(err, errA) {
		t.Errorf("returned error should contain errA: %v", err)
	}
	if !errors.Is(err, errB) {
		t.Errorf("returned error should contain errB: %v", err)
	}
}

// TestRunHooks_StopTimeout verifies that RunHooks respects the stopTimeout
// and does not block indefinitely when a Stop is slow. The Stop respects its
// context (no goroutine leak).
func TestRunHooks_StopTimeout(t *testing.T) {
	const stopTimeout = 80 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	hook := Hook{
		Name: "slow",
		Start: func(context.Context) error {
			close(started)
			return nil
		},
		Stop: func(ctx context.Context) error {
			// Block until the timeout ctx expires — tests that RunHooks
			// does not wait beyond stopTimeout. No goroutine is leaked
			// because Stop returns as soon as ctx is done.
			<-ctx.Done()
			return ctx.Err()
		},
	}

	done := make(chan error, 1)
	go func() { done <- RunHooks(ctx, slog.Default(), stopTimeout, hook) }()

	<-started
	cancel()

	deadline := time.After(5 * stopTimeout)
	select {
	case <-done:
		// returned within the expected window — pass
	case <-deadline:
		t.Fatal("RunHooks blocked past stopTimeout")
	}
}

// TestRunHooks_NilStop verifies that a nil Stop is silently skipped without
// panicking.
func TestRunHooks_NilStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	hook := Hook{
		Name:  "no-stop",
		Start: func(context.Context) error { return nil },
		Stop:  nil,
	}
	if err := RunHooks(ctx, slog.Default(), time.Second, hook); err != nil {
		t.Fatalf("RunHooks() with nil Stop = %v, want nil", err)
	}
}
