package app_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/testkit"
)

// F-02 regression (adversarial-framework-review-2026-07-17): a worker whose
// critical child loops die (here: every DB call fails because the pools are
// closed) must NOT stay alive until the parent context is cancelled. The first
// unexpected child failure must cancel siblings, drain, and surface the error
// promptly while the parent context is still live.
func TestStartWorkerReturnsPromptlyWhenChildrenFail(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	// Break the platform pool AFTER boot: relay/runner/scheduler loops now fail
	// on their first poll, exactly like a database outage.
	h.Platform.Close()

	// Parent context stays live far longer than the assertion window: a return
	// before its deadline proves supervision, not cancellation.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.StartWorker(ctx, booted, app.WorkerConfigOpts{
			RelayPoll:     50 * time.Millisecond,
			JobPoll:       50 * time.Millisecond,
			SchedulerPoll: 50 * time.Millisecond,
			ShutdownDrain: 2 * time.Second,
		})
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("StartWorker returned nil after its critical children failed")
		}
		// Sibling cancellation, explicitly: StartWorker only returns after ALL
		// three children have exited (drain barrier) or the 2s drain cap fires.
		// The parent context is still live (30s), so a prompt full-drain return
		// entails that every child that had not itself failed exited via the
		// supervisor's sibling cancellation — a child waiting on the PARENT
		// context would hold the drain for the full 30s and trip the cap path
		// with errDrainTimeout in the joined error instead.
		if strings.Contains(err.Error(), "shutdown drain deadline exceeded") {
			t.Fatalf("children did not exit via sibling cancellation (drain cap hit): %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("StartWorker still blocked 10s after every critical child loop failed (parent context live) — child failures are unsupervised")
	}
}

// F-02 regression: a hook that starts successfully but whose serving loop dies
// later (an occupied listener, a crashed Serve) must surface through RunHooks
// without any external signal.
func TestRunHooksSurfacesAsyncHookFailure(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sentinel := errors.New("listener died")
	failed := make(chan error, 1)

	hook := app.Hook{
		Name: "flaky-listener",
		Start: func(ctx context.Context) error {
			go func() {
				time.Sleep(100 * time.Millisecond)
				failed <- sentinel
			}()
			return nil
		},
		Failed: failed,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- app.RunHooks(ctx, log, time.Second, hook) }()

	select {
	case err := <-done:
		if !errors.Is(err, sentinel) {
			t.Fatalf("RunHooks returned %v, want the async hook failure %v", err, sentinel)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("RunHooks still blocked 10s after the hook's serving loop failed — async hook death is unobserved")
	}
}

// Closure-review regressions (adversarial closure review 2026-07-17, F-02).

// After an async hook failure, RunHooks must still STOP every
// started hook (in reverse order) before returning the failure.
func TestRunHooksStopsAllHooksAfterAsyncFailure(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sentinel := errors.New("serving loop died")
	failed := make(chan error, 1)
	var stops []string
	var mu sync.Mutex
	recordStop := func(name string) func(context.Context) error {
		return func(context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			stops = append(stops, name)
			return nil
		}
	}
	hooks := []app.Hook{
		{Name: "first", Start: func(context.Context) error { return nil }, Stop: recordStop("first")},
		{Name: "flaky", Start: func(context.Context) error {
			go func() { failed <- sentinel }()
			return nil
		}, Failed: failed, Stop: recordStop("flaky")},
		{Name: "last", Start: func(context.Context) error { return nil }, Stop: recordStop("last")},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := app.RunHooks(ctx, log, time.Second, hooks...)
	if !errors.Is(err, sentinel) {
		t.Fatalf("RunHooks = %v, want the async failure", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(stops) != 3 || stops[0] != "last" || stops[1] != "flaky" || stops[2] != "first" {
		t.Fatalf("stops after async failure = %v, want [last flaky first] (all hooks, reverse order)", stops)
	}
}

// The generated API's exact hook shape (synchronous net.Listen in Start,
// Serve-death via Failed): an occupied address must surface as a Start error
// from the hook runner — the process exits instead of serving nothing.
func TestSupervisedListenerHookFailsStartOnOccupiedAddress(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Occupy a port.
	occupier, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = occupier.Close() }()
	addr := occupier.Addr().String()

	srv := &http.Server{Addr: addr}
	httpFailed := make(chan error, 1)
	hook := app.Hook{
		Name: "http",
		Start: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("http: bind %s: %w", addr, err)
			}
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					httpFailed <- err
				}
			}()
			return nil
		},
		Failed: httpFailed,
		Stop:   func(ctx context.Context) error { return srv.Shutdown(ctx) },
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = app.RunHooks(ctx, log, time.Second, hook)
	if err == nil || !strings.Contains(err.Error(), "bind") {
		t.Fatalf("RunHooks on an occupied address = %v, want a bind Start error", err)
	}
}
