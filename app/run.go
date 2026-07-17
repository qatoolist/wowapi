// Startup/shutdown lifecycle skeleton (phase-plan Phase 1).
//
// RunHooks provides supervised start-then-stop sequencing for named components
// (HTTP server, outbox relay, …). Signal wiring (SIGINT/SIGTERM) belongs to the
// process main via signal.NotifyContext; RunHooks stays testable with a plain
// context.
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

// Hook is a component whose background work can die AFTER a successful
// Start (a listener that stopped serving, a loop that crashed) and report it
// through Failed. RunHooks treats a received error like a Start
// failure: it stops every started hook and returns the error — the process
// must never sit alive while a critical serving loop is dead (adversarial
// review 2026-07-17, F-02).
type Hook struct {
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

// RunHooks starts hooks in order and blocks until ctx is cancelled or a Start
// returns an error or a started component reports asynchronous failure. It
// blocks on every non-nil Failed channel and
// treats a received error like a Start failure — every started hook is stopped
// (in reverse order, each bounded by stopTimeout) and the error is returned.
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
			go func(h Hook) {
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
