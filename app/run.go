// Startup/shutdown lifecycle skeleton (phase-plan Phase 1).
//
// RunHooks provides orderly start-then-stop sequencing for named components
// (HTTP server, outbox relay, …) wired in as Hook values. Signal wiring
// (SIGINT/SIGTERM) belongs to the process main via signal.NotifyContext;
// RunHooks stays testable with a plain context.
//
// Real server construction arrives in Phases 2–3; this skeleton establishes
// the lifecycle contract that those phases plug into.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Hook is one startable/stoppable component: an HTTP server, outbox relay,
// background sweeper, or any other process-lifetime service.
//
// Hook's field set is FROZEN at its v1 shape (Name, Start, Stop): consumers
// write unkeyed composite literals (app.Hook{"api", start, stop}), and adding
// any field — exported or not — is a source-incompatible change for them
// (closure review 2026-07-17, F-02). Components that need to report
// post-start background failure use SupervisedHook instead.
type Hook struct {
	// Name is used in log messages; should be short and unique within a run.
	Name string
	// Start launches background work and must return promptly (not block for
	// the component's lifetime). The ctx is the run context; components should
	// respect its cancellation on their own internal paths.
	Start func(ctx context.Context) error
	// Stop performs a graceful shutdown. nil means nothing to stop. Stop
	// receives a fresh context bounded by the stopTimeout, independent of
	// the (already-cancelled) run context.
	Stop func(ctx context.Context) error
}

// SupervisedHook is a Hook whose background work can die AFTER a successful
// Start (a listener that stopped serving, a loop that crashed) and report it
// through Failed. RunSupervisedHooks treats a received error like a Start
// failure: it stops every started hook and returns the error — the process
// must never sit alive while a critical serving loop is dead (adversarial
// review 2026-07-17, F-02). This is a separate type, not a field on Hook, so
// the frozen v1 Hook shape keeps compiling for unkeyed composite literals.
type SupervisedHook struct {
	// Name is used in log messages; should be short and unique within a run.
	Name string
	// Start launches background work and must return promptly.
	Start func(ctx context.Context) error
	// Failed, when non-nil, reports post-Start background death. Send at most
	// one error; nil channels are simply never selected.
	Failed <-chan error
	// Stop performs a graceful shutdown. nil means nothing to stop.
	Stop func(ctx context.Context) error
}

// Supervised adapts a plain Hook into a SupervisedHook with no failure signal,
// for mixing legacy hooks into a RunSupervisedHooks call.
func Supervised(h Hook) SupervisedHook {
	return SupervisedHook{Name: h.Name, Start: h.Start, Stop: h.Stop}
}

// RunHooks starts hooks in order and blocks until ctx is cancelled or a Start
// returns an error. It then stops the successfully-started hooks in reverse
// order, each bounded by stopTimeout.
//
// Errors are handled as follows:
//   - A Start error aborts the remaining starts and triggers immediate shutdown.
//   - Stop errors are all collected (never short-circuited) and joined with
//     any Start error via errors.Join.
//   - If every Start and Stop succeeds, RunHooks returns nil.
func RunHooks(ctx context.Context, logger *slog.Logger, stopTimeout time.Duration, hooks ...Hook) error {
	supervised := make([]SupervisedHook, len(hooks))
	for i, h := range hooks {
		supervised[i] = Supervised(h)
	}
	return RunSupervisedHooks(ctx, logger, stopTimeout, supervised...)
}

// RunSupervisedHooks is RunHooks for components that can also die after a
// successful Start: it additionally blocks on every non-nil Failed channel and
// treats a received error like a Start failure — every started hook is stopped
// (in reverse order, each bounded by stopTimeout) and the error is returned.
func RunSupervisedHooks(ctx context.Context, logger *slog.Logger, stopTimeout time.Duration, hooks ...SupervisedHook) error {
	var (
		started  []SupervisedHook
		startErr error
	)

	for _, h := range hooks {
		logger.Info("starting", "hook", h.Name)
		if err := h.Start(ctx); err != nil {
			startErr = err
			break
		}
		started = append(started, h)
	}

	// Block until the run context is done (normal shutdown), a Start failed,
	// or a started hook reports that its background work died (F-02: a process
	// must never sit alive while a critical serving loop is dead).
	if startErr == nil {
		asyncFail := make(chan error, 1)
		watchCtx, stopWatch := context.WithCancel(ctx)
		for _, h := range started {
			if h.Failed == nil {
				continue
			}
			go func(h SupervisedHook) {
				select {
				case err := <-h.Failed:
					if err != nil {
						select {
						case asyncFail <- fmt.Errorf("hook %q: background work failed: %w", h.Name, err):
						default:
						}
					}
				case <-watchCtx.Done():
				}
			}(h)
		}
		select {
		case <-ctx.Done():
		case err := <-asyncFail:
			logger.Error("hook background work failed; shutting down", "err", err)
			startErr = err
		}
		stopWatch()
	}

	// Stop in reverse order; Stop receives its own timeout-bounded context
	// because the run ctx is already cancelled at this point.
	var stopErrs []error
	for i := len(started) - 1; i >= 0; i-- {
		h := started[i]
		logger.Info("stopping", "hook", h.Name)
		if h.Stop == nil {
			continue
		}
		stopCtx, cancel := context.WithTimeout(context.Background(), stopTimeout)
		err := h.Stop(stopCtx)
		cancel()
		if err != nil {
			stopErrs = append(stopErrs, err)
		}
	}

	return errors.Join(append([]error{startErr}, stopErrs...)...)
}
