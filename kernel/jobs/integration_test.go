package jobs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/testkit"
)

// testJob is the generic payload used across integration tests; each test
// registers its own worker for the kind on its own Registry.
type testJob struct {
	N string `json:"n"`
}

func (testJob) Kind() string { return "test.jobs.run" }

const jobKind = "test.jobs.run"

// jobStatus reads a job's status via the app_platform pool.
func jobStatus(t *testing.T, h *testkit.DBHandle, id int64) string {
	t.Helper()
	var s string
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT status FROM jobs_queue WHERE id = $1`, id).Scan(&s); err != nil {
		t.Fatalf("read job status: %v", err)
	}
	return s
}

// singleJobID returns the id of the sole queued job of jobKind.
func singleJobID(t *testing.T, h *testkit.DBHandle) int64 {
	t.Helper()
	var id int64
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT id FROM jobs_queue WHERE kind = $1`, jobKind).Scan(&id); err != nil {
		t.Fatalf("read job id: %v", err)
	}
	return id
}

// countJobs counts jobs_queue rows for a kind + tenant via app_platform.
func countJobs(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID) int {
	t.Helper()
	var n int
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT count(*) FROM jobs_queue WHERE kind = $1 AND tenant_id = $2`, jobKind, tenant).Scan(&n); err != nil {
		t.Fatalf("count jobs: %v", err)
	}
	return n
}

// countRuns counts job_runs rows for a job id in a given status.
func countRuns(t *testing.T, h *testkit.DBHandle, jobID int64, status string) int {
	t.Helper()
	var n int
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT count(*) FROM job_runs WHERE job_id = $1 AND status = $2`, jobID, status).Scan(&n); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	return n
}

// countOutbox counts events_outbox rows of an event_type visible to a tenant
// (through the RLS-enforced runtime TxManager).
func countOutbox(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, evType string) int {
	t.Helper()
	var n int
	err := h.TxM.WithTenantRO(testkit.TenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx, `SELECT count(*) FROM events_outbox WHERE event_type = $1`, evType).Scan(&n)
	})
	if err != nil {
		t.Fatalf("count outbox as tenant: %v", err)
	}
	return n
}

// TestIntegrationJobsEnqueueAtomic proves the enqueue INSERT rides the caller's
// business tx: a tx that fails leaves no job; a tx that commits leaves the job.
func TestIntegrationJobsEnqueueAtomic(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tenant.ID)

	// (1) enqueue inside a tx that then FAILS → rolled back with the business tx.
	boom := errors.New("business write failed")
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if e := jobs.Enqueue(ctx, db, testJob{N: "rolled-back"}); e != nil {
			return e
		}
		return boom // abort the tx after enqueuing
	})
	if !errors.Is(err, boom) {
		t.Fatalf("WithTenant error = %v, want boom", err)
	}
	if n := countJobs(t, h, tenant.ID); n != 0 {
		t.Fatalf("after rolled-back tx: %d jobs queued, want 0 (enqueue must be atomic)", n)
	}

	// (2) enqueue inside a tx that COMMITS → job present.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, testJob{N: "committed"})
	}); err != nil {
		t.Fatalf("commit enqueue: %v", err)
	}
	if n := countJobs(t, h, tenant.ID); n != 1 {
		t.Fatalf("after committed tx: %d jobs queued, want 1", n)
	}
}

// TestIntegrationJobsWorkerSucceeds proves a successful run: job completed,
// job_runs succeeded, and the worker executed under the job's tenant RLS (it
// reads a row only visible to that tenant and its write lands under that tenant).
func TestIntegrationJobsWorkerSucceeds(t *testing.T) {
	h := testkit.NewDB(t)
	tenantA := testkit.CreateTenant(t, h)
	tenantB := testkit.CreateTenant(t, h)

	// Seed a row visible only to tenant A (Admin bypasses RLS to set tenant_id).
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO events_outbox (id, tenant_id, event_type, created_by) VALUES ($1, $2, 'seed.for.tenant', $3)`,
		uuid.New(), tenantA.ID, uuid.Nil); err != nil {
		t.Fatalf("seed tenant row: %v", err)
	}

	reg := jobs.NewRegistry()
	reg.RegisterKind(jobKind, func(ctx context.Context, db database.TenantDB, payload []byte) error {
		// The worker must see exactly tenant A's seeded row (RLS scoped).
		var seen int
		if err := db.QueryRow(ctx, `SELECT count(*) FROM events_outbox WHERE event_type = 'seed.for.tenant'`).Scan(&seen); err != nil {
			return err
		}
		if seen != 1 {
			return fmt.Errorf("worker saw %d seeded rows, want 1 (wrong tenant binding)", seen)
		}
		// Write a marker under app_tenant_id() — lands under the job's tenant.
		_, err := db.Exec(ctx,
			`INSERT INTO events_outbox (id, tenant_id, event_type, created_by) VALUES ($1, app_tenant_id(), 'job.ran.marker', $2)`,
			uuid.New(), uuid.Nil)
		return err
	}, jobs.DefaultRetry())
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	// Enqueue for tenant A in a committed business tx.
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenantA.ID), func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, testJob{N: "ok"})
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	id := singleJobID(t, h)

	r := jobs.NewRunner(h.Platform, h.TxM, reg)
	n, err := r.ClaimOnce(context.Background())
	if err != nil {
		t.Fatalf("ClaimOnce: %v", err)
	}
	if n != 1 {
		t.Fatalf("ClaimOnce claimed %d, want 1", n)
	}

	if s := jobStatus(t, h, id); s != "completed" {
		t.Fatalf("job status = %q, want completed", s)
	}
	if c := countRuns(t, h, id, "succeeded"); c != 1 {
		t.Fatalf("job_runs succeeded rows = %d, want 1", c)
	}
	// Tenant-awareness: the marker is visible to A, invisible to B.
	if c := countOutbox(t, h, tenantA.ID, "job.ran.marker"); c != 1 {
		t.Fatalf("tenant A sees %d marker rows, want 1", c)
	}
	if c := countOutbox(t, h, tenantB.ID, "job.ran.marker"); c != 0 {
		t.Fatalf("tenant B sees %d marker rows, want 0 (RLS leak)", c)
	}
}

// TestIntegrationJobsRetryToDLQ proves a permanently-failing job exhausts its
// attempts and is discarded to the DLQ with a matching 'dead' job_runs row.
func TestIntegrationJobsRetryToDLQ(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKind(jobKind, func(context.Context, database.TenantDB, []byte) error {
		return errors.New("always fails")
	}, jobs.RetryPolicy{MaxAttempts: 3, Backoff: func(int) time.Duration { return 0 }})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, testJob{N: "doomed"}, jobs.WithMaxAttempts(3))
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	id := singleJobID(t, h)

	r := jobs.NewRunner(h.Platform, h.TxM, reg)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		if jobStatus(t, h, id) == "discarded" {
			break
		}
		n, err := r.ClaimOnce(ctx)
		if err != nil {
			t.Fatalf("ClaimOnce: %v", err)
		}
		if n == 0 && jobStatus(t, h, id) != "discarded" {
			t.Fatalf("job not claimed and not discarded (status=%s)", jobStatus(t, h, id))
		}
	}

	if s := jobStatus(t, h, id); s != "discarded" {
		t.Fatalf("job status = %q, want discarded", s)
	}
	// attempts incremented to the ceiling; last_error recorded.
	var attempts int
	var lastErr *string
	if err := h.Platform.QueryRow(ctx,
		`SELECT attempts, last_error FROM jobs_queue WHERE id = $1`, id).Scan(&attempts, &lastErr); err != nil {
		t.Fatalf("read attempts: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	if lastErr == nil || *lastErr == "" {
		t.Fatalf("last_error not recorded")
	}
	// DLQ mirror: one 'dead' run and two 'failed' runs (attempts 1 & 2).
	if c := countRuns(t, h, id, "dead"); c != 1 {
		t.Fatalf("job_runs dead rows = %d, want 1", c)
	}
	if c := countRuns(t, h, id, "failed"); c != 2 {
		t.Fatalf("job_runs failed rows = %d, want 2", c)
	}
}

// TestIntegrationJobsBackoffReschedules proves a retryable failure reschedules
// the job to a future run_at (backoff) rather than discarding it.
func TestIntegrationJobsBackoffReschedules(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKind(jobKind, func(context.Context, database.TenantDB, []byte) error {
		return errors.New("transient")
	}, jobs.RetryPolicy{MaxAttempts: 5, Backoff: func(int) time.Duration { return 30 * time.Second }})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, testJob{N: "retry"})
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	id := singleJobID(t, h)

	r := jobs.NewRunner(h.Platform, h.TxM, reg)
	ctx := context.Background()
	if n, err := r.ClaimOnce(ctx); err != nil || n != 1 {
		t.Fatalf("first ClaimOnce = (%d, %v), want (1, nil)", n, err)
	}

	// Available again, one attempt recorded, run_at pushed into the future.
	var status string
	var attempts int
	var future bool
	if err := h.Platform.QueryRow(ctx,
		`SELECT status, attempts, run_at > now() FROM jobs_queue WHERE id = $1`, id).Scan(&status, &attempts, &future); err != nil {
		t.Fatalf("read job: %v", err)
	}
	if status != "available" || attempts != 1 || !future {
		t.Fatalf("after 1 failure: status=%q attempts=%d future=%v, want available/1/true", status, attempts, future)
	}
	if c := countRuns(t, h, id, "failed"); c != 1 {
		t.Fatalf("job_runs failed rows = %d, want 1", c)
	}
	// A second claim finds nothing — the job is not yet eligible (backoff).
	if n, err := r.ClaimOnce(ctx); err != nil || n != 0 {
		t.Fatalf("second ClaimOnce = (%d, %v), want (0, nil) — backoff should defer it", n, err)
	}
}

// TestIntegrationJobsReclaimStalled proves a job stuck 'running' (crashed worker)
// is reset to 'available' once its lock is older than the timeout.
func TestIntegrationJobsReclaimStalled(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	ctx := context.Background()

	// Simulate a job a dead worker left 'running' 10 minutes ago.
	if _, err := h.Platform.Exec(ctx,
		`INSERT INTO jobs_queue (kind, tenant_id, payload, status, locked_at)
         VALUES ($1, $2, '{}', 'running', now() - interval '10 minutes')`,
		jobKind, tenant.ID); err != nil {
		t.Fatalf("insert stalled job: %v", err)
	}
	id := singleJobID(t, h)

	r := jobs.NewRunner(h.Platform, h.TxM, jobs.NewRegistry())
	reclaimed, err := r.ReclaimStalled(ctx, 5*time.Minute)
	if err != nil {
		t.Fatalf("ReclaimStalled: %v", err)
	}
	if reclaimed != 1 {
		t.Fatalf("reclaimed %d jobs, want 1", reclaimed)
	}
	if s := jobStatus(t, h, id); s != "available" {
		t.Fatalf("stalled job status = %q, want available", s)
	}
	var lockedAt *time.Time
	if err := h.Platform.QueryRow(ctx, `SELECT locked_at FROM jobs_queue WHERE id = $1`, id).Scan(&lockedAt); err != nil {
		t.Fatalf("read locked_at: %v", err)
	}
	if lockedAt != nil {
		t.Fatalf("locked_at = %v, want NULL after reclaim", lockedAt)
	}

	// A fresh 'running' job (lock just now) is NOT reclaimed by the same timeout.
	if _, err := h.Platform.Exec(ctx,
		`UPDATE jobs_queue SET status='running', locked_at=now() WHERE id=$1`, id); err != nil {
		t.Fatalf("re-lock: %v", err)
	}
	if reclaimed, err := r.ReclaimStalled(ctx, 5*time.Minute); err != nil || reclaimed != 0 {
		t.Fatalf("ReclaimStalled on fresh lock = (%d, %v), want (0, nil)", reclaimed, err)
	}
}

// TestIntegrationJobsTenantIsolation proves a job enqueued for tenant A executes
// with app.tenant_id = A: its write is visible to A and invisible to B.
func TestIntegrationJobsTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	tenantA := testkit.CreateTenant(t, h)
	tenantB := testkit.CreateTenant(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKind(jobKind, func(ctx context.Context, db database.TenantDB, payload []byte) error {
		_, err := db.Exec(ctx,
			`INSERT INTO events_outbox (id, tenant_id, event_type, created_by) VALUES ($1, app_tenant_id(), 'isolation.marker', $2)`,
			uuid.New(), uuid.Nil)
		return err
	}, jobs.DefaultRetry())
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	if err := h.TxM.WithTenant(testkit.TenantCtx(tenantA.ID), func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, testJob{N: "for-A"})
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	r := jobs.NewRunner(h.Platform, h.TxM, reg)
	if n, err := r.ClaimOnce(context.Background()); err != nil || n != 1 {
		t.Fatalf("ClaimOnce = (%d, %v), want (1, nil)", n, err)
	}

	if c := countOutbox(t, h, tenantA.ID, "isolation.marker"); c != 1 {
		t.Fatalf("tenant A sees %d marker rows, want 1", c)
	}
	if c := countOutbox(t, h, tenantB.ID, "isolation.marker"); c != 0 {
		t.Fatalf("tenant B sees %d marker rows, want 0 (job ran under wrong tenant)", c)
	}
}
