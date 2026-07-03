package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/outbox"
)

// WorkerConfigOpts tunes the worker loops.
type WorkerConfigOpts struct {
	RelayBatch    int
	RelayPoll     time.Duration
	JobPoll       time.Duration
	JobPoolSize   int
	ShutdownDrain time.Duration
}

// StartWorker runs the background worker process for a booted app: the outbox
// relay (dispatches events to subscribed handlers) and the job runner (executes
// enqueued jobs with retries/DLQ). It blocks until ctx is cancelled, then drains
// in-flight work within ShutdownDrain and returns. Requires the kernel's
// Platform pool (cross-tenant kernel work) — a worker/migrate posture, not
// api-only.
//
// Signal wiring (SIGINT/SIGTERM) belongs to the process main via
// signal.NotifyContext; StartWorker stays testable with a plain context.
func StartWorker(ctx context.Context, b *Booted, opts WorkerConfigOpts) error {
	k := b.Kernel
	if k.Platform == nil {
		return errNoPlatformPool
	}
	log := k.Log
	if log == nil {
		log = slog.Default()
	}
	if opts.RelayPoll <= 0 {
		opts.RelayPoll = time.Second
	}
	if opts.JobPoll <= 0 {
		opts.JobPoll = time.Second
	}
	if opts.ShutdownDrain <= 0 {
		opts.ShutdownDrain = 30 * time.Second
	}

	relay := outbox.NewRelay(k.Platform, k.Tx, b.Events, opts.RelayBatch)
	var runnerOpts []jobs.RunnerOpt
	if opts.JobPoolSize > 0 {
		runnerOpts = append(runnerOpts, jobs.WithPoolSize(opts.JobPoolSize))
	}
	runnerOpts = append(runnerOpts, jobs.WithDrainTimeout(opts.ShutdownDrain), jobs.WithLogger(log))
	runner := jobs.NewRunner(k.Platform, k.Tx, b.Jobs, runnerOpts...)

	// Both loops respect ctx cancellation and drain in-flight work themselves.
	// StartWorker blocks until ctx is cancelled and both have returned — but with
	// a HARD cap: if a loop does not drain within ShutdownDrain (e.g. a worker
	// ignoring ctx), StartWorker returns anyway rather than hanging the process
	// forever (review finding ARCH-57). Leaked in-flight work is logged; the DB
	// reclaim path recovers any job left 'running'.
	log.InfoContext(ctx, "worker starting", "relay_poll", opts.RelayPoll, "job_poll", opts.JobPoll)
	var wg sync.WaitGroup
	var relayErr, jobErr error
	wg.Add(2)
	go func() { defer wg.Done(); relayErr = relay.Run(ctx, opts.RelayPoll) }()
	go func() { defer wg.Done(); jobErr = runner.Run(ctx, opts.JobPoll) }()

	drained := make(chan struct{})
	go func() { wg.Wait(); close(drained) }()

	<-ctx.Done() // wait for shutdown signal
	stopCtx := context.WithoutCancel(ctx)
	select {
	case <-drained:
		log.InfoContext(stopCtx, "worker stopped (drained)")
	case <-time.After(opts.ShutdownDrain):
		log.WarnContext(stopCtx, "worker shutdown drain deadline exceeded; releasing with work possibly in flight",
			"drain", opts.ShutdownDrain)
		return errDrainTimeout
	}
	if relayErr != nil {
		return relayErr
	}
	return jobErr
}

// errDrainTimeout signals the hard shutdown-drain cap was hit.
var errDrainTimeout = workerErr("app: worker shutdown drain deadline exceeded")

// errNoPlatformPool is returned when the worker is started without a platform pool.
var errNoPlatformPool = workerErr("app: StartWorker requires the kernel Platform pool (worker/migrate posture)")

type workerErr string

func (e workerErr) Error() string { return string(e) }
