package jobs_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/jobs"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// discardLogger returns a slog.Logger that swallows output (keeps test logs clean
// while still exercising the runner's internal error-logging branches).
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// constIDGen returns the same UUID for every New() — used to force a job_runs
// primary-key collision so the outcome-record error branches execute.
type constIDGen struct{ id uuid.UUID }

func (c constIDGen) New() uuid.UUID { return c.id }

// badPayloadJob has a field json.Marshal cannot encode (a channel), so Enqueue /
// EnqueueGlobal take their "payload is not JSON-encodable" branch.
type badPayloadJob struct {
	Ch chan int `json:"ch"`
}

func (badPayloadJob) Kind() string { return "test.jobs.badpayload" }

// countJobsByStatus counts jobs_queue rows of jobKind in a given status.
func countJobsByStatus(t *testing.T, h *testkit.DBHandle, status string) int {
	t.Helper()
	var n int
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT count(*) FROM jobs_queue WHERE kind = $1 AND status = $2`, jobKind, status).Scan(&n); err != nil {
		t.Fatalf("count jobs by status: %v", err)
	}
	return n
}

// jobRow reads status/attempts/run_at-in-future/last_error for a job.
func jobRow(t *testing.T, h *testkit.DBHandle, id int64) (status string, attempts int, future bool, lastErr string) {
	t.Helper()
	var le *string
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT status, attempts, run_at > now(), last_error FROM jobs_queue WHERE id = $1`, id).
		Scan(&status, &attempts, &future, &le); err != nil {
		t.Fatalf("read job row: %v", err)
	}
	if le != nil {
		lastErr = *le
	}
	return
}

func enqueueFor(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, opts ...jobs.Opt) {
	t.Helper()
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, testJob{N: "x"}, opts...)
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
}

// TestIntegrationEnqueueWithRunAtDelays proves WithRunAt sets a future run_at so
// the job is not yet claimable.
func TestIntegrationEnqueueWithRunAtDelays(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	enqueueFor(t, h, tenant.ID, jobs.WithRunAt(time.Now().Add(time.Hour)))

	id := singleJobID(t, h)
	if _, _, future, _ := jobRow(t, h, id); !future {
		t.Fatal("WithRunAt(future): run_at should be in the future")
	}
	r := jobs.NewRunner(h.Platform, h.TxM, jobs.NewRegistry())
	if n, err := r.ClaimOnce(context.Background()); err != nil || n != 0 {
		t.Fatalf("ClaimOnce = (%d, %v), want (0, nil) — delayed job not yet eligible", n, err)
	}

	// Rewind run_at to the past → now claimable (a fresh runner, no worker: it will
	// fail to unknown_kind, but claiming proves eligibility flipped).
	if _, err := h.Platform.Exec(context.Background(),
		`UPDATE jobs_queue SET run_at = now() - interval '1 minute' WHERE id = $1`, id); err != nil {
		t.Fatalf("rewind run_at: %v", err)
	}
	if n, err := r.ClaimOnce(context.Background()); err != nil || n != 1 {
		t.Fatalf("ClaimOnce after rewind = (%d, %v), want (1, nil)", n, err)
	}
}

// TestEnqueueRejectsInvalidJobs covers Enqueue's guard branches: nil job, empty
// kind, and an unmarshalable payload — all KindInternal, none enqueued.
func TestEnqueueRejectsInvalidJobs(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tenant.ID)

	call := func(j jobs.Job) error {
		return h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			return jobs.Enqueue(ctx, db, j)
		})
	}
	if err := call(nil); kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("nil job: KindOf = %v, want Internal", kerr.KindOf(err))
	}
	if err := call(emptyKindJob{}); kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("empty-kind job: KindOf = %v, want Internal", kerr.KindOf(err))
	}
	if err := call(badPayloadJob{Ch: make(chan int)}); kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("unmarshalable payload: KindOf = %v, want Internal", kerr.KindOf(err))
	}
}

// TestEnqueueGlobalRejectsUnmarshalablePayload covers EnqueueGlobal's marshal
// error branch (the invalid-kind branches are covered elsewhere).
func TestEnqueueGlobalRejectsUnmarshalablePayload(t *testing.T) {
	h := testkit.NewDB(t)
	r := jobs.NewRunner(h.Platform, h.TxM, jobs.NewRegistry())
	if err := r.EnqueueGlobal(context.Background(), badPayloadJob{Ch: make(chan int)}); kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("unmarshalable global payload: KindOf = %v, want Internal", kerr.KindOf(err))
	}
}

// TestIntegrationUnknownKindFailsAndReschedules proves a claimed job whose kind
// has no registered worker is recorded as a failure (unknown_kind) and — since it
// is below max_attempts — rescheduled with the default backoff (backoffFor's
// unregistered-kind fallback to ExpJitterBackoff pushes run_at into the future).
func TestIntegrationUnknownKindFailsAndReschedules(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	enqueueFor(t, h, tenant.ID, jobs.WithMaxAttempts(3))
	id := singleJobID(t, h)

	// Empty registry: no worker for jobKind.
	r := jobs.NewRunner(h.Platform, h.TxM, jobs.NewRegistry(), jobs.WithLogger(discardLogger()))
	if n, err := r.ClaimOnce(context.Background()); err != nil || n != 1 {
		t.Fatalf("ClaimOnce = (%d, %v), want (1, nil)", n, err)
	}

	status, attempts, future, lastErr := jobRow(t, h, id)
	if status != "available" || attempts != 1 || !future {
		t.Fatalf("after unknown-kind failure: status=%q attempts=%d future=%v, want available/1/true", status, attempts, future)
	}
	if !strings.Contains(lastErr, "no worker registered") {
		t.Fatalf("last_error = %q, want it to mention no worker registered", lastErr)
	}
	if c := countRuns(t, h, id, "failed"); c != 1 {
		t.Fatalf("job_runs failed rows = %d, want 1", c)
	}
}

// TestIntegrationLongErrorTruncated proves a worker error longer than 2000 bytes
// is truncated before it is stored in last_error.
func TestIntegrationLongErrorTruncated(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKindWithIdempotency(jobKind, func(context.Context, database.TenantDB, []byte) error {
		return errors.New(strings.Repeat("x", 5000))
	}, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.RetryPolicy{MaxAttempts: 3, Backoff: func(int) time.Duration { return 30 * time.Second }})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}
	enqueueFor(t, h, tenant.ID, jobs.WithMaxAttempts(3))
	id := singleJobID(t, h)

	r := jobs.NewRunner(h.Platform, h.TxM, reg, jobs.WithLogger(discardLogger()))
	if n, err := r.ClaimOnce(context.Background()); err != nil || n != 1 {
		t.Fatalf("ClaimOnce = (%d, %v), want (1, nil)", n, err)
	}
	var lastErr string
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT last_error FROM jobs_queue WHERE id = $1`, id).Scan(&lastErr); err != nil {
		t.Fatalf("read last_error: %v", err)
	}
	if len(lastErr) != 2000 {
		t.Fatalf("last_error length = %d, want 2000 (truncated)", len(lastErr))
	}
}

// TestIntegrationRunnerOptionsAndDeadHook exercises every RunnerOpt (pool size,
// job/reclaim/drain timeouts — the small reclaim timeout also trips NewRunner's
// safety floor — id gen, logger) and asserts the dead-letter hook fires once with
// the correct DeadJob when a job exhausts its attempts.
func TestIntegrationRunnerOptionsAndDeadHook(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKindWithIdempotency(jobKind, func(context.Context, database.TenantDB, []byte) error {
		return errors.New("always fails")
	}, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.RetryPolicy{MaxAttempts: 2, Backoff: func(int) time.Duration { return 0 }})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	var deadCalls int32
	var captured jobs.DeadJob
	r := jobs.NewRunner(h.Platform, h.TxM, reg,
		jobs.WithPoolSize(3),
		jobs.WithJobTimeout(time.Minute),
		jobs.WithReclaimTimeout(time.Second), // < floor(jobTimeout+drain+1m) → clamped
		jobs.WithReclaimInterval(time.Second),
		jobs.WithDrainTimeout(time.Second),
		jobs.WithIDGen(model.UUIDv7()),
		jobs.WithLogger(discardLogger()),
		jobs.WithDeadHook(func(_ context.Context, dj jobs.DeadJob) {
			atomic.AddInt32(&deadCalls, 1)
			captured = dj
		}),
	)

	enqueueFor(t, h, tenant.ID, jobs.WithMaxAttempts(2))
	id := singleJobID(t, h)

	ctx := context.Background()
	for i := 0; i < 6 && jobStatus(t, h, id) != "discarded"; i++ {
		if _, err := r.ClaimOnce(ctx); err != nil {
			t.Fatalf("ClaimOnce: %v", err)
		}
	}
	if s := jobStatus(t, h, id); s != "discarded" {
		t.Fatalf("job status = %q, want discarded", s)
	}
	if n := atomic.LoadInt32(&deadCalls); n != 1 {
		t.Fatalf("dead hook called %d times, want 1", n)
	}
	if captured.ID != id || captured.Kind != jobKind || captured.Attempts != 2 || captured.LastError == "" {
		t.Fatalf("DeadJob = %+v, want id=%d kind=%s attempts=2 non-empty error", captured, id, jobKind)
	}
	if captured.Tenant == nil || *captured.Tenant != tenant.ID {
		t.Fatalf("DeadJob.Tenant = %v, want %v", captured.Tenant, tenant.ID)
	}
	if c := countRuns(t, h, id, "dead"); c != 1 {
		t.Fatalf("job_runs dead rows = %d, want 1", c)
	}
}

// TestIntegrationRecordSuccessConflictLogged forces the success-record path's
// error branch: a constant IDGen makes the SECOND succeeded job_runs INSERT
// violate the primary key, so its outcome tx rolls back and the job is left
// 'running' (at-least-once) rather than crashing the runner.
func TestIntegrationRecordSuccessConflictLogged(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKindWithIdempotency(jobKind, func(context.Context, database.TenantDB, []byte) error {
		return nil // both jobs succeed
	}, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	enqueueFor(t, h, tenant.ID)
	enqueueFor(t, h, tenant.ID)

	r := jobs.NewRunner(h.Platform, h.TxM, reg,
		jobs.WithIDGen(constIDGen{id: uuid.New()}),
		jobs.WithLogger(discardLogger()))
	if n, err := r.ClaimOnce(context.Background()); err != nil || n != 2 {
		t.Fatalf("ClaimOnce = (%d, %v), want (2, nil)", n, err)
	}

	// Exactly one succeeded run persisted (the PK collision aborted the other);
	// its job is 'completed', the conflicted job stays 'running'.
	var succeeded int
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT count(*) FROM job_runs WHERE job_kind = $1 AND status = 'succeeded'`, jobKind).Scan(&succeeded); err != nil {
		t.Fatalf("count succeeded runs: %v", err)
	}
	if succeeded != 1 {
		t.Fatalf("succeeded job_runs = %d, want 1 (one INSERT hit the PK conflict)", succeeded)
	}
	if got := countJobsByStatus(t, h, "completed"); got != 1 {
		t.Fatalf("completed jobs = %d, want 1", got)
	}
	if got := countJobsByStatus(t, h, "running"); got != 1 {
		t.Fatalf("running jobs = %d, want 1 (conflicted outcome tx rolled back)", got)
	}
}

// TestIntegrationRecordFailureConflictLogged forces recordFailure's error branch
// via the same constant-IDGen PK collision: two failing jobs, only one 'failed'
// job_runs row lands; the conflicted job is left 'running', the other 'available'.
func TestIntegrationRecordFailureConflictLogged(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKindWithIdempotency(jobKind, func(context.Context, database.TenantDB, []byte) error {
		return errors.New("boom")
	}, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.RetryPolicy{MaxAttempts: 5, Backoff: func(int) time.Duration { return 30 * time.Second }})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	enqueueFor(t, h, tenant.ID)
	enqueueFor(t, h, tenant.ID)

	r := jobs.NewRunner(h.Platform, h.TxM, reg,
		jobs.WithIDGen(constIDGen{id: uuid.New()}),
		jobs.WithLogger(discardLogger()))
	if n, err := r.ClaimOnce(context.Background()); err != nil || n != 2 {
		t.Fatalf("ClaimOnce = (%d, %v), want (2, nil)", n, err)
	}

	var failed int
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT count(*) FROM job_runs WHERE job_kind = $1 AND status = 'failed'`, jobKind).Scan(&failed); err != nil {
		t.Fatalf("count failed runs: %v", err)
	}
	if failed != 1 {
		t.Fatalf("failed job_runs = %d, want 1 (PK conflict aborted the other)", failed)
	}
	if got := countJobsByStatus(t, h, "available"); got != 1 {
		t.Fatalf("available jobs = %d, want 1", got)
	}
	if got := countJobsByStatus(t, h, "running"); got != 1 {
		t.Fatalf("running jobs = %d, want 1 (conflicted outcome tx rolled back)", got)
	}
}

// TestRunnerRunExitsOnCancelledCtx covers Run's poll<=0 default and its top-of-
// loop early return when the context is already cancelled before any claim.
func TestRunnerRunExitsOnCancelledCtx(t *testing.T) {
	h := testkit.NewDB(t)
	r := jobs.NewRunner(h.Platform, h.TxM, jobs.NewRegistry(), jobs.WithLogger(discardLogger()))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx, 0) }() // poll <= 0 → default poll

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return promptly on a pre-cancelled ctx")
	}
}

// TestIntegrationRunnerRunReclaimsStalled proves Run's periodic reclaim sweep
// fires: a job left 'running' by a crashed worker (old lock) is reset to
// 'available' by the reclaim ticker, then claimed and completed — all via Run.
func TestIntegrationRunnerRunReclaimsStalled(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	ctx0 := context.Background()

	// A job a dead worker left 'running' 20 minutes ago (older than any timeout).
	if _, err := h.Platform.Exec(ctx0,
		`INSERT INTO jobs_queue (kind, tenant_id, payload, status, locked_at)
		 VALUES ($1, $2, '{}', 'running', now() - interval '20 minutes')`,
		jobKind, tenant.ID); err != nil {
		t.Fatalf("insert stalled job: %v", err)
	}
	id := singleJobID(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKindWithIdempotency(jobKind, func(context.Context, database.TenantDB, []byte) error {
		return nil
	}, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	// Small reclaim interval so the sweep fires quickly under Run.
	r := jobs.NewRunner(h.Platform, h.TxM, reg,
		jobs.WithReclaimInterval(10*time.Millisecond),
		jobs.WithLogger(discardLogger()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx, 10*time.Millisecond) }()

	// The job can only leave 'running' via the reclaim sweep, then get claimed and
	// completed by the worker.
	deadline := time.After(8 * time.Second)
	for jobStatus(t, h, id) != "completed" {
		select {
		case <-deadline:
			cancel()
			t.Fatalf("stalled job never reclaimed+completed (status=%s)", jobStatus(t, h, id))
		default:
			time.Sleep(10 * time.Millisecond)
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
	if c := countRuns(t, h, id, "succeeded"); c != 1 {
		t.Fatalf("job_runs succeeded rows = %d, want 1", c)
	}
}

// TestIntegrationRunnerRunGracefulDrain proves Runner.Run drives claims until
// cancel and drains an in-flight job: the worker starts, ctx is cancelled while it
// runs, yet the detached execution finishes and records success before Run returns.
func TestIntegrationRunnerRunGracefulDrain(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	started := make(chan struct{})
	var once sync.Once
	reg := jobs.NewRegistry()
	reg.RegisterKindWithIdempotency(jobKind, func(context.Context, database.TenantDB, []byte) error {
		once.Do(func() { close(started) })
		time.Sleep(200 * time.Millisecond) // in-flight when cancel arrives
		return nil
	}, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}
	enqueueFor(t, h, tenant.ID)
	id := singleJobID(t, h)

	r := jobs.NewRunner(h.Platform, h.TxM, reg,
		jobs.WithReclaimInterval(50*time.Millisecond),
		jobs.WithLogger(discardLogger()))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx, 20*time.Millisecond) }()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("worker never started")
	}
	cancel() // graceful shutdown mid-job

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel (drain hung)")
	}

	if s := jobStatus(t, h, id); s != "completed" {
		t.Fatalf("job status = %q, want completed (drain must finish the in-flight job)", s)
	}
	if c := countRuns(t, h, id, "succeeded"); c != 1 {
		t.Fatalf("job_runs succeeded rows = %d, want 1", c)
	}
}
