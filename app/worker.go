package app

import (
	"context"
	"errors"
	"fmt"
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
	// Scheduler (leader-safe kernel maintenance sweeps). Zero values use defaults.
	SchedulerPoll         time.Duration // how often to check for due tasks (default 30s)
	SLAInterval           time.Duration // workflow SLA sweep interval (default 1m)
	IdempotencyInterval   time.Duration // idempotency-key expiry sweep interval (default 1h)
	DLQDepthInterval      time.Duration // dlq_depth gauge refresh interval (default 1m)
	AuditAnchorInterval   time.Duration // audit-chain anchor-export interval (default 1h)
	NotifySendInterval    time.Duration // notify send/retry poll interval (default 1m)
	WebhookRetryInterval  time.Duration // webhook retry + inbound poll interval (default 1m)
	UploadSessionInterval time.Duration // document upload session GC interval (default 1h)
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
	// A Booted value App.Boot did not produce never passed validation; running
	// it would convert construction misuse into unvalidated operation (third
	// closure audit 2026-07-17). No fallback — fail closed.
	if b == nil || !b.runtime.set {
		return ErrNotBooted
	}
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
	if opts.SchedulerPoll <= 0 {
		opts.SchedulerPoll = 30 * time.Second
	}
	if opts.SLAInterval <= 0 {
		opts.SLAInterval = time.Minute
	}
	if opts.IdempotencyInterval <= 0 {
		opts.IdempotencyInterval = time.Hour
	}
	if opts.DLQDepthInterval <= 0 {
		opts.DLQDepthInterval = time.Minute
	}
	if opts.AuditAnchorInterval <= 0 {
		opts.AuditAnchorInterval = time.Hour
	}
	if opts.NotifySendInterval <= 0 {
		opts.NotifySendInterval = time.Minute
	}
	if opts.WebhookRetryInterval <= 0 {
		opts.WebhookRetryInterval = time.Minute
	}
	if opts.UploadSessionInterval <= 0 {
		opts.UploadSessionInterval = time.Hour
	}

	// The relay, runner, and scheduler consume the boot-validated runtime view;
	// reassigning Booted's exported fields cannot change what this worker runs
	// (second closure audit 2026-07-17, F-10).
	relay := outbox.NewRelay(k.Platform, k.Tx, b.RuntimeEvents(), opts.RelayBatch,
		outbox.WithRelayTracer(k.Tracer), outbox.WithRelayMetrics(k.Metrics))
	var runnerOpts []jobs.RunnerOpt
	if opts.JobPoolSize > 0 {
		runnerOpts = append(runnerOpts, jobs.WithPoolSize(opts.JobPoolSize))
	}
	// WithRunnerTracer continues each job's originating request trace across the
	// async boundary (roadmap O1/CA-9), mirroring the outbox relay tracer above.
	runnerOpts = append(runnerOpts, jobs.WithDrainTimeout(opts.ShutdownDrain), jobs.WithLogger(log), jobs.WithRunnerTracer(k.Tracer))
	runner := jobs.NewRunner(k.Platform, k.Tx, b.RuntimeJobs(), runnerOpts...)

	// Scheduler: leader-safe kernel maintenance sweeps (SLA timers, idempotency
	// expiry). Registered here so every worker replica participates; the schedules
	// table ensures each due task runs on exactly one replica per interval.
	sched := jobs.NewScheduler(k.Platform, log)
	sched.OnRun(func(name string, lag time.Duration, err error) {
		log.InfoContext(ctx, "scheduler ran maintenance task",
			"task", name, "lag_ms", lag.Milliseconds(), "ok", err == nil)
		// Export scheduler/sweeper lag as a gauge and task failures as a counter
		// (roadmap R3/CA-1). NoOp unless a metrics adapter is wired.
		k.Metrics.SetGauge("scheduler_lag_seconds", lag.Seconds(),
			map[string]string{"task": name})
		if err != nil {
			k.Metrics.IncCounter("scheduler_task_errors_total", 1,
				map[string]string{"task": name})
		}
	})
	registerMaintenance(sched, k, opts.SLAInterval, opts.IdempotencyInterval, opts.DLQDepthInterval, opts.AuditAnchorInterval, opts.NotifySendInterval, opts.WebhookRetryInterval, opts.UploadSessionInterval)
	registerModuleRecurring(sched, k, b.runtimeRecurring())

	// Both loops respect ctx cancellation and drain in-flight work themselves.
	// StartWorker blocks until ctx is cancelled and both have returned — but with
	// a HARD cap: if a loop does not drain within ShutdownDrain (e.g. a worker
	// ignoring ctx), StartWorker returns anyway rather than hanging the process
	// forever (review finding ARCH-57). Leaked in-flight work is logged; the DB
	// reclaim path recovers any job left 'running'.
	log.InfoContext(ctx, "worker starting", "relay_poll", opts.RelayPoll, "job_poll", opts.JobPoll)

	// Supervision (adversarial review 2026-07-17, F-02): the relay, runner, and
	// scheduler are this process's reason to exist. Each runs under a child
	// context; the FIRST unexpected child failure (a non-nil error while the
	// parent context is still live) cancels the siblings, drains, and returns
	// promptly — the worker must never sit falsely alive with its critical
	// loops dead, waiting for an external signal.
	childCtx, cancelChildren := context.WithCancel(ctx)
	defer cancelChildren()

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		childErr []error
		failed   = make(chan struct{})
		failOnce sync.Once
	)
	supervise := func(name string, run func(context.Context) error) {
		defer wg.Done()
		if err := run(childCtx); err != nil && !errors.Is(err, context.Canceled) {
			mu.Lock()
			childErr = append(childErr, fmt.Errorf("%s: %w", name, err))
			mu.Unlock()
			failOnce.Do(func() { close(failed) })
			cancelChildren() // fail fast: siblings stop instead of serving a half-dead process
		}
	}
	wg.Add(3)
	go supervise("relay", func(c context.Context) error { return relay.Run(c, opts.RelayPoll) })
	go supervise("job runner", func(c context.Context) error { return runner.Run(c, opts.JobPoll) })
	go supervise("scheduler", func(c context.Context) error { return sched.Run(c, opts.SchedulerPoll) })

	drained := make(chan struct{})
	go func() { wg.Wait(); close(drained) }()

	select {
	case <-ctx.Done(): // normal shutdown signal
	case <-failed: // a critical child died while the parent was still live
		log.ErrorContext(context.WithoutCancel(ctx),
			"worker critical loop failed; cancelling siblings and shutting down")
	}
	stopCtx := context.WithoutCancel(ctx)
	cancelChildren()
	select {
	case <-drained:
		log.InfoContext(stopCtx, "worker stopped (drained)")
	case <-time.After(opts.ShutdownDrain):
		log.WarnContext(stopCtx, "worker shutdown drain deadline exceeded; releasing with work possibly in flight",
			"drain", opts.ShutdownDrain)
		mu.Lock()
		defer mu.Unlock()
		return errors.Join(append([]error{error(errDrainTimeout)}, childErr...)...)
	}
	mu.Lock()
	defer mu.Unlock()
	return errors.Join(childErr...)
}

// errDrainTimeout signals the hard shutdown-drain cap was hit.
var errDrainTimeout = workerErr("app: worker shutdown drain deadline exceeded")

// errNoPlatformPool is returned when the worker is started without a platform pool.
var errNoPlatformPool = workerErr("app: StartWorker requires the kernel Platform pool (worker/migrate posture)")

type workerErr string

func (e workerErr) Error() string { return string(e) }
