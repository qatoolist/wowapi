package app_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
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
