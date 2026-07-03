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
	"log/slog"
	"time"
)

// Hook is one startable/stoppable component: an HTTP server, outbox relay,
// background sweeper, or any other process-lifetime service.
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
	var (
		started  []Hook
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

	// Block until the run context is done (normal shutdown or start failure).
	if startErr == nil {
		<-ctx.Done()
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
