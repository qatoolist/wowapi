package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/lease"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/observability"
)

// enqueueConfig is the per-enqueue overrides set by Opts.
type enqueueConfig struct {
	runAt       *time.Time
	maxAttempts *int
	tracer      observability.Tracer
}

// Opt customizes a single Enqueue/EnqueueGlobal call.
type Opt func(*enqueueConfig)

// WithRunAt schedules the job to become eligible no earlier than t (delayed
// jobs). Without it the job is eligible immediately (run_at = now()).
func WithRunAt(t time.Time) Opt {
	return func(c *enqueueConfig) { c.runAt = &t }
}

// WithMaxAttempts overrides the number of executions before the job is discarded
// to the DLQ. Values <= 0 are ignored (the table default of 5 applies).
func WithMaxAttempts(n int) Opt {
	return func(c *enqueueConfig) {
		if n > 0 {
			c.maxAttempts = &n
		}
	}
}

// WithTracer wires a tracer so this enqueue captures the current request's W3C
// traceparent (roadmap O1/CA-9) into the job's trace_context; the runner
// continues that trace when it later executes the job. Default: NoOpTracer
// (empty trace context — no behavior change). Mirrors outbox.WithWriterTracer.
func WithTracer(tr observability.Tracer) Opt {
	return func(c *enqueueConfig) {
		if tr != nil {
			c.tracer = tr
		}
	}
}

// applyOpts folds the enqueue options over a config seeded with the caller's
// default tracer (NoOp for the tenant Enqueue free function; the Runner's tracer
// for EnqueueGlobal).
func applyOpts(defaultTracer observability.Tracer, opts []Opt) enqueueConfig {
	c := enqueueConfig{tracer: defaultTracer}
	for _, o := range opts {
		o(&c)
	}
	return c
}

// enqueueSQL inserts one job. tenant_id is supplied by the caller-specific
// expression ($tenantExpr): app_tenant_id() for a tenant enqueue (so it binds to
// the same tenant as the surrounding business tx and RLS/grants line up) or NULL
// for a global job. run_at/max_attempts fall back to the table defaults when the
// bound params are NULL. Casts pin the param types so pgx never fails inference.
const enqueueSQLTenant = `INSERT INTO jobs_queue (kind, tenant_id, payload, run_at, max_attempts, trace_context)
     VALUES ($1, app_tenant_id(), $2::jsonb, COALESCE($3::timestamptz, now()), COALESCE($4::int, 5), $5)`

const enqueueSQLGlobal = `INSERT INTO jobs_queue (kind, tenant_id, payload, run_at, max_attempts, trace_context)
     VALUES ($1, NULL, $2::jsonb, COALESCE($3::timestamptz, now()), COALESCE($4::int, 5), $5)`

// Enqueue inserts j into jobs_queue in the caller's tenant transaction. Because
// the INSERT rides the caller's tx (app_rt has INSERT on jobs_queue), the job is
// committed atomically with the business write: roll the tx back and the job
// never exists; commit it and the job is guaranteed queued. The tenant is taken
// from app_tenant_id() (the tx's SET LOCAL binding), not from Go.
func Enqueue(ctx context.Context, db database.TenantDB, j Job, opts ...Opt) error {
	if j == nil || j.Kind() == "" {
		return kerr.E(kerr.KindInternal, "invalid_job", "Enqueue requires a job with a non-empty Kind")
	}
	payload, err := json.Marshal(j)
	if err != nil {
		return kerr.E(kerr.KindInternal, "invalid_job", "job payload is not JSON-encodable")
	}
	c := applyOpts(observability.NoOpTracer, opts)
	// Capture the current distributed-trace context (W3C traceparent) so the
	// runner can continue the SAME trace when it executes the job asynchronously
	// (roadmap O1/CA-9). Empty string (NoOp / no active span) → NULL trace_context.
	if _, err := db.Exec(ctx, enqueueSQLTenant, j.Kind(), payload, c.runAt, c.maxAttempts, traceContext(ctx, c.tracer)); err != nil {
		return kerr.Wrapf(err, "jobs.Enqueue", "insert job %s", j.Kind())
	}
	return nil
}

// traceContext returns the injected W3C traceparent for the span active in ctx,
// or a nil any when tracing is disabled (so the DB stores NULL, not an empty
// string).
func traceContext(ctx context.Context, tr observability.Tracer) any {
	if tc := tr.Inject(ctx); tc != "" {
		return tc
	}
	return nil
}

// EnqueueGlobal inserts a tenant-less (global) job. It is a Runner method
// because it writes on the app_platform pool (there is no business tx to ride —
// a global job has no tenant), unlike Enqueue which rides the caller's tenant
// tx. The row's tenant_id is NULL; at execution the worker runs under the
// sentinel nil tenant (see execOne).
func (r *Runner) EnqueueGlobal(ctx context.Context, j Job, opts ...Opt) error {
	if j == nil || j.Kind() == "" {
		return kerr.E(kerr.KindInternal, "invalid_job", "EnqueueGlobal requires a job with a non-empty Kind")
	}
	payload, err := json.Marshal(j)
	if err != nil {
		return kerr.E(kerr.KindInternal, "invalid_job", "job payload is not JSON-encodable")
	}
	// Seed with the runner's tracer so a global enqueue captures trace context by
	// default (the tenant Enqueue free function defaults to NoOp); WithTracer still
	// overrides per call.
	c := applyOpts(r.tracer, opts)
	if _, err := r.pool.Exec(ctx, enqueueSQLGlobal, j.Kind(), payload, c.runAt, c.maxAttempts, traceContext(ctx, c.tracer)); err != nil {
		return kerr.Wrapf(err, "jobs.EnqueueGlobal", "insert global job %s", j.Kind())
	}
	return nil
}

// DeadJob describes a job that exhausted its attempts and landed in the DLQ
// (status=discarded). It is handed to the dead-letter hook (WithDeadHook) so a
// process can emit a metric or alert.
type DeadJob struct {
	ID        int64
	Kind      string
	Tenant    *uuid.UUID // nil for a global job
	Attempts  int
	LastError string
}

// Runner consumes jobs. It holds an app_platform pool (claim + status writes,
// which app_rt is not granted), the tenant TxManager (to execute each worker in
// a transaction bound to the job's tenant), and the Registry of workers. It runs
// a bounded fixed-size worker pool — never one goroutine per job (blueprint: no
// unbounded goroutines; the `go` keyword is permitted in kernel/jobs).
type Runner struct {
	pool   *pgxpool.Pool
	txm    database.TxManager
	reg    *Registry
	idgen  model.IDGen
	log    *slog.Logger
	tracer observability.Tracer

	poolSize       int           // max concurrent workers and claim batch size
	stalledTimeout time.Duration // running jobs older than this are reclaimable
	reclaimEvery   time.Duration // how often Run sweeps for stalled jobs
	drainTimeout   time.Duration // max time in-flight jobs get to finish on shutdown
	jobTimeout     time.Duration // per-job max runtime (independent of shutdown drain)

	onDead func(context.Context, DeadJob)
}

// RunnerOpt customizes a Runner.
type RunnerOpt func(*Runner)

// WithPoolSize sets the bounded worker-pool size (and the per-claim batch).
// Default 10.
func WithPoolSize(n int) RunnerOpt {
	return func(r *Runner) {
		if n > 0 {
			r.poolSize = n
		}
	}
}

// WithJobTimeout bounds a single job's worker runtime, independent of the
// shutdown drain budget. Default 2m. The reclaim floor is derived from this.
func WithJobTimeout(d time.Duration) RunnerOpt {
	return func(r *Runner) {
		if d > 0 {
			r.jobTimeout = d
		}
	}
}

// WithReclaimTimeout sets how old a 'running' job's lock must be before
// ReclaimStalled resets it to 'available' (crash recovery). Default 5m.
func WithReclaimTimeout(d time.Duration) RunnerOpt {
	return func(r *Runner) {
		if d > 0 {
			r.stalledTimeout = d
		}
	}
}

// WithReclaimInterval sets how often Run sweeps for stalled jobs. Default 1m.
func WithReclaimInterval(d time.Duration) RunnerOpt {
	return func(r *Runner) {
		if d > 0 {
			r.reclaimEvery = d
		}
	}
}

// WithDrainTimeout bounds how long in-flight jobs may finish after ctx is
// cancelled before the runner stops waiting. Default 30s.
func WithDrainTimeout(d time.Duration) RunnerOpt {
	return func(r *Runner) {
		if d > 0 {
			r.drainTimeout = d
		}
	}
}

// WithIDGen overrides the id generator used for job_runs primary keys (tests
// inject a deterministic sequence).
func WithIDGen(g model.IDGen) RunnerOpt {
	return func(r *Runner) {
		if g != nil {
			r.idgen = g
		}
	}
}

// WithDeadHook registers a callback invoked when a job is discarded to the DLQ
// (the "leave a hook for a metric" seam).
func WithDeadHook(fn func(context.Context, DeadJob)) RunnerOpt {
	return func(r *Runner) { r.onDead = fn }
}

// WithLogger overrides the slog.Logger for internal (non-worker) errors.
func WithLogger(l *slog.Logger) RunnerOpt {
	return func(r *Runner) {
		if l != nil {
			r.log = l
		}
	}
}

// WithRunnerTracer wires a tracer so the runner continues each job's originating
// request trace when it executes the job (roadmap O1/CA-9): it extracts the
// traceparent captured at enqueue and runs the worker under a child span, and it
// is also the default tracer for EnqueueGlobal. Default: NoOpTracer. Mirrors
// outbox.WithRelayTracer.
func WithRunnerTracer(tr observability.Tracer) RunnerOpt {
	return func(r *Runner) {
		if tr != nil {
			r.tracer = tr
		}
	}
}

// NewRunner wires a Runner. platformPool must authenticate as app_platform (the
// role granted claim/complete on jobs_queue + job_runs); txm runs worker
// transactions per tenant; reg supplies the workers.
func NewRunner(platformPool *pgxpool.Pool, txm database.TxManager, reg *Registry, opts ...RunnerOpt) *Runner {
	r := &Runner{
		pool:           platformPool,
		txm:            txm,
		reg:            reg,
		idgen:          model.UUIDv7(),
		log:            slog.Default(),
		tracer:         observability.NoOpTracer,
		poolSize:       10,
		stalledTimeout: 5 * time.Minute,
		reclaimEvery:   time.Minute,
		drainTimeout:   30 * time.Second,
		jobTimeout:     2 * time.Minute,
	}
	for _, o := range opts {
		o(r)
	}
	// Invariant (review finding ARCH-58): the stalled-reclaim timeout MUST
	// exceed the longest a job can run (jobTimeout) plus the shutdown drain, or a
	// still-executing job could be reclaimed and run concurrently by another
	// runner. Enforce a safe floor rather than trust configuration.
	if floor := r.jobTimeout + r.drainTimeout + time.Minute; r.stalledTimeout < floor {
		r.stalledTimeout = floor
	}
	return r
}

// claimedJob is one row claimed for execution.
type claimedJob struct {
	id          int64
	kind        string
	tenant      *uuid.UUID // nil for a global job
	payload     []byte
	attempts    int
	maxAttempts int
	trace       *string // W3C traceparent captured at enqueue (CA-9); nil when absent
	lease       lease.Lease
}

// claimSQL atomically selects up to $1 eligible jobs (available, run_at reached)
// in run_at order, skipping rows locked by a peer runner, and flips them to
// 'running' with a lock timestamp — all in one statement so the running state is
// durable before any worker executes. A crash after this commits leaves the job
// 'running'; ReclaimStalled recovers it.
const claimSQL = `WITH claimed AS (
        SELECT id
          FROM jobs_queue
         WHERE status = 'available' AND run_at <= now()
         ORDER BY run_at
         FOR UPDATE SKIP LOCKED
         LIMIT $1
    )
    UPDATE jobs_queue q
       SET status = 'running',
           locked_at = now(),
           lease_token = $2,
           lease_generation = lease_generation + 1,
           lease_expires_at = now() + make_interval(secs => $3::double precision)
      FROM claimed
     WHERE q.id = claimed.id
    RETURNING q.id, q.kind, q.tenant_id, q.payload, q.attempts, q.max_attempts, q.trace_context,
              q.lease_token, q.lease_generation, q.lease_expires_at`

// ClaimOnce claims up to poolSize available jobs (marking each 'running' in a
// committed statement), then executes them concurrently on the bounded worker
// pool, waiting for the batch to finish. It returns the number of jobs claimed.
// Per-job outcomes (completed / retry / DLQ) are written by the workers; a DB
// failure while writing an outcome is logged and leaves the job 'running' for
// ReclaimStalled — ClaimOnce only returns an error for a failure to claim.
func (r *Runner) ClaimOnce(ctx context.Context) (int, error) {
	leaseTTL := r.stalledTimeout.Seconds()
	rows, err := r.pool.Query(ctx, claimSQL, r.poolSize, lease.New(r.stalledTimeout).Token, leaseTTL)
	if err != nil {
		return 0, kerr.Wrapf(err, "jobs.ClaimOnce", "claim jobs")
	}
	var batch []claimedJob
	for rows.Next() {
		var jb claimedJob
		var expiresAt time.Time
		if err := rows.Scan(&jb.id, &jb.kind, &jb.tenant, &jb.payload, &jb.attempts, &jb.maxAttempts, &jb.trace, &jb.lease.Token, &jb.lease.Generation, &expiresAt); err != nil {
			rows.Close()
			return 0, kerr.Wrapf(err, "jobs.ClaimOnce", "scan claimed job")
		}
		jb.lease.ExpiresAt = expiresAt
		batch = append(batch, jb)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, kerr.Wrapf(err, "jobs.ClaimOnce", "iterate claimed jobs")
	}
	if len(batch) == 0 {
		return 0, nil
	}

	// Detach execution from ctx cancellation so an in-flight batch finishes on
	// graceful shutdown, then release. Tenant/actor context values are preserved
	// by WithoutCancel. Per-job runtime is bounded separately (execOne applies
	// jobTimeout) — drainTimeout is a SHUTDOWN budget only, not a per-job cap
	// (review finding ARCH-56).
	execCtx := context.WithoutCancel(ctx)

	sem := make(chan struct{}, r.poolSize)
	var wg sync.WaitGroup
	for _, jb := range batch {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			r.execOne(execCtx, jb)
		}()
	}
	wg.Wait()
	return len(batch), nil
}

// execOne runs one claimed job's worker in a transaction bound to the job's
// tenant, then records the outcome. A global (NULL-tenant) job runs under the
// sentinel nil tenant (blueprint permits a sentinel): jobs_queue and job_runs
// are un-RLS'd global tables, so a global worker touching only kernel tables is
// unaffected by the binding; a global worker must not read tenant-scoped tables.
func (r *Runner) execOne(ctx context.Context, jb claimedJob) {
	e, ok := r.reg.lookup(jb.kind)
	if !ok {
		r.recordFailure(ctx, jb, kerr.E(kerr.KindInternal, "unknown_kind",
			"no worker registered for kind "+jb.kind))
		return
	}

	tenant := uuid.Nil
	if jb.tenant != nil {
		tenant = *jb.tenant
	}

	// Continue the originating request's trace across the async boundary (roadmap
	// O1/CA-9): extract the traceparent captured at enqueue, then run the worker
	// under a child span (the job-runner span). Zero-cost with NoOpTracer.
	if jb.trace != nil && *jb.trace != "" {
		ctx = r.tracer.Extract(ctx, *jb.trace)
	}
	ctx, span := r.tracer.StartSpan(ctx, "jobs.run "+jb.kind)
	span.SetAttr("job.kind", jb.kind)
	span.SetAttr("job.id", strconv.FormatInt(jb.id, 10))
	defer span.End()

	// The worker runs under a per-job timeout. The outcome (success/failure)
	// is persisted with a SEPARATE fresh, short-lived context so a job whose
	// worker ctx expired can still record its status rather than being left
	// 'running' until reclaim (review finding ARCH-56).
	workerCtx, cancel := context.WithTimeout(ctx, r.jobTimeout)
	idempKey := fmt.Sprintf("%s:%d", jb.kind, jb.id)
	workerCtx = WithJobContext(workerCtx, jb.id, idempKey, jb.lease)
	tctx := database.WithTenantID(workerCtx, tenant)
	werr := r.txm.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		return e.worker(ctx, db, jb.payload)
	})
	cancel()

	outCtx, outCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer outCancel()
	if werr != nil {
		span.RecordError(werr)
		r.recordFailure(outCtx, jb, werr)
		return
	}
	r.recordSuccess(outCtx, jb)
}

// recordSuccess marks the job completed and mirrors a succeeded job_runs row, in
// a single app_platform transaction.
func (r *Runner) recordSuccess(ctx context.Context, jb claimedJob) {
	err := r.inPlatformTx(ctx, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx,
			`UPDATE jobs_queue
				   SET status = 'completed', finished_at = now(), locked_at = NULL,
				       lease_token = NULL, lease_expires_at = NULL
				 WHERE id = $1
				   AND lease_token = $2
				   AND lease_generation = $3
				   AND lease_expires_at > now()`,
			jb.id, jb.lease.Token, jb.lease.Generation)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return kerr.E(kerr.KindConflict, "lease_mismatch", "stale finalize rejected")
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO job_runs (id, tenant_id, job_kind, job_id, status, finished_at)
             VALUES ($1, $2, $3, $4, 'succeeded', now())`,
			r.idgen.New(), jb.tenant, jb.kind, jb.id)
		return err
	})
	if err != nil {
		// The worker already committed; failing to flip status leaves the job
		// 'running' for reclaim (at-least-once — workers are idempotent).
		r.log.Error("jobs: record success failed", "job_id", jb.id, "kind", jb.kind, "err", err)
	}
}

// recordFailure applies the retry/DLQ decision: attempts+1; if still under the
// job's max_attempts, reschedule 'available' at now()+Backoff(attempts) with the
// error; otherwise discard to the DLQ (status=discarded) and mirror a 'dead'
// job_runs row. A succeeded/failed run is always mirrored to job_runs.
func (r *Runner) recordFailure(ctx context.Context, jb claimedJob, cause error) {
	attempts := jb.attempts + 1
	msg := truncate(cause.Error(), 2000)
	dead := attempts >= jb.maxAttempts

	err := r.inPlatformTx(ctx, func(tx pgx.Tx) error {
		if dead {
			res, err := tx.Exec(ctx,
				`UPDATE jobs_queue
                    SET status = 'discarded', attempts = $2, last_error = $3,
                        finished_at = now(), locked_at = NULL,
                        lease_token = NULL, lease_expires_at = NULL
                  WHERE id = $1
                    AND lease_token = $4
                    AND lease_generation = $5
                    AND lease_expires_at > now()`,
				jb.id, attempts, msg, jb.lease.Token, jb.lease.Generation)
			if err != nil {
				return err
			}
			if res.RowsAffected() == 0 {
				return kerr.E(kerr.KindConflict, "lease_mismatch", "stale finalize rejected")
			}
			_, err = tx.Exec(ctx,
				`INSERT INTO job_runs (id, tenant_id, job_kind, job_id, status, finished_at, error)
                 VALUES ($1, $2, $3, $4, 'dead', now(), $5)`,
				r.idgen.New(), jb.tenant, jb.kind, jb.id, msg)
			return err
		}
		// Compute run_at on the DB clock (now() + backoff), not the app clock, so
		// eligibility is consistent with the claim query's now() even under
		// app/Postgres clock skew.
		backoffSecs := r.backoffFor(jb.kind, attempts).Seconds()
		res, err := tx.Exec(ctx,
			`UPDATE jobs_queue
                SET status = 'available', attempts = $2,
                    run_at = now() + make_interval(secs => $3::double precision),
                    last_error = $4, locked_at = NULL,
                    lease_token = NULL, lease_expires_at = NULL
              WHERE id = $1
                AND lease_token = $5
                AND lease_generation = $6
                AND lease_expires_at > now()`,
			jb.id, attempts, backoffSecs, msg, jb.lease.Token, jb.lease.Generation)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return kerr.E(kerr.KindConflict, "lease_mismatch", "stale finalize rejected")
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO job_runs (id, tenant_id, job_kind, job_id, status, finished_at, error)
             VALUES ($1, $2, $3, $4, 'failed', now(), $5)`,
			r.idgen.New(), jb.tenant, jb.kind, jb.id, msg)
		return err
	})
	if err != nil {
		r.log.Error("jobs: record failure failed", "job_id", jb.id, "kind", jb.kind, "err", err)
		return
	}
	if dead {
		r.log.Warn("jobs: job discarded to DLQ", "job_id", jb.id, "kind", jb.kind, "attempts", attempts, "err", msg)
		if r.onDead != nil {
			r.onDead(ctx, DeadJob{ID: jb.id, Kind: jb.kind, Tenant: jb.tenant, Attempts: attempts, LastError: msg})
		}
	}
}

// backoffFor resolves the registered backoff for a kind, defaulting to
// ExpJitterBackoff for an unregistered kind.
func (r *Runner) backoffFor(kind string, attempt int) time.Duration {
	if e, ok := r.reg.lookup(kind); ok && e.retry.Backoff != nil {
		return e.retry.Backoff(attempt)
	}
	return ExpJitterBackoff(attempt)
}

// inPlatformTx runs fn inside an app_platform transaction on the runner's pool
// (begin/commit/rollback owned here).
func (r *Runner) inPlatformTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ReclaimStalled resets 'running' jobs whose lock is older than olderThan back
// to 'available', so jobs a crashed worker left mid-flight are retried. It
// returns the number of jobs reclaimed. make_interval pins the unit
// unambiguously (a Go duration string like "5m0s" would be misread by Postgres
// interval parsing, where "m" means months).
func (r *Runner) ReclaimStalled(ctx context.Context, olderThan time.Duration) (int, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE jobs_queue
            SET status = 'available',
                locked_at = NULL,
                lease_token = NULL,
                lease_generation = lease_generation + 1,
                lease_expires_at = NULL
          WHERE status = 'running'
            AND locked_at < now() - make_interval(secs => $1::double precision)`,
		olderThan.Seconds())
	if err != nil {
		return 0, kerr.Wrapf(err, "jobs.ReclaimStalled", "reclaim stalled jobs")
	}
	return int(tag.RowsAffected()), nil
}

// Run drives the runner until ctx is cancelled: ClaimOnce back-to-back while
// there is work, then poll on the interval, sweeping stalled jobs periodically.
// Cancellation is graceful — the loop stops claiming new work, and the in-flight
// batch finishes (bounded by drainTimeout) before Run returns nil.
func (r *Runner) Run(ctx context.Context, poll time.Duration) error {
	if poll <= 0 {
		poll = time.Second
	}
	pollT := time.NewTicker(poll)
	defer pollT.Stop()
	reclaimT := time.NewTicker(r.reclaimEvery)
	defer reclaimT.Stop()

	for {
		if ctx.Err() != nil {
			return nil
		}
		n, err := r.ClaimOnce(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if n > 0 {
			continue // drain: more work may be waiting
		}
		select {
		case <-ctx.Done():
			return nil
		case <-pollT.C:
		case <-reclaimT.C:
			if _, err := r.ReclaimStalled(ctx, r.stalledTimeout); err != nil && ctx.Err() == nil {
				r.log.Warn("jobs: reclaim sweep failed", "err", err)
			}
		}
	}
}

// truncate caps s at n bytes so a pathological error string cannot bloat
// last_error / job_runs.error. It backs the cut up to a UTF-8 rune boundary so the
// stored string is never invalid UTF-8 — which a Postgres text column would reject.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
