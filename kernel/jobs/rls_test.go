package jobs_test

import (
	"context"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationJobsQueueRLSRejectsCrossTenantInsert is the M2 probe: with
// FORCE RLS on jobs_queue (migration 00028), an app_rt session bound to tenant A
// can no longer INSERT a row carrying tenant_id=B. app_rt holds the INSERT grant,
// so this is the WITH CHECK — not the grant — doing the rejecting. A same-tenant
// control INSERT proves the policy still admits legitimate enqueues (the RLS
// change did not blanket-break app_rt writes).
func TestIntegrationJobsQueueRLSRejectsCrossTenantInsert(t *testing.T) {
	h := testkit.NewDB(t)
	tenantA := testkit.CreateTenant(t, h)
	tenantB := testkit.CreateTenant(t, h)

	// Bound to A, try to plant a row under B — must be rejected by WITH CHECK.
	err := h.TxM.WithTenant(testkit.TenantCtx(tenantA.ID), func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx,
			`INSERT INTO jobs_queue (kind, tenant_id, payload) VALUES ($1, $2, '{}')`,
			jobKind, tenantB.ID)
		return e
	})
	if err == nil {
		t.Fatal("app_rt bound to A must NOT be able to INSERT jobs_queue row with tenant_id=B (RLS WITH CHECK)")
	}
	if !strings.Contains(err.Error(), "row-level security") {
		t.Fatalf("expected a row-level-security WITH CHECK violation, got: %v", err)
	}
	// Nothing landed under B.
	if n := countJobs(t, h, tenantB.ID); n != 0 {
		t.Fatalf("cross-tenant INSERT left %d rows under B, want 0", n)
	}

	// Control: bound to A, INSERT under app_tenant_id() (== A) succeeds — the
	// policy admits a legitimate same-tenant enqueue.
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenantA.ID), func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx,
			`INSERT INTO jobs_queue (kind, tenant_id, payload) VALUES ($1, app_tenant_id(), '{}')`,
			jobKind)
		return e
	}); err != nil {
		t.Fatalf("same-tenant INSERT must still succeed under RLS: %v", err)
	}
	if n := countJobs(t, h, tenantA.ID); n != 1 {
		t.Fatalf("same-tenant enqueue left %d rows under A, want 1", n)
	}
}

// TestIntegrationJobsGrantOnlyIsolation is finding F-3: jobs_queue and job_runs
// are grant-only isolated from app_rt — app_rt has NO SELECT grant on either
// table. FORCE RLS (00028) does not add one, so a tenant-bound app_rt SELECT is
// still denied at the grant layer. Previously untested; asserted here.
func TestIntegrationJobsGrantOnlyIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	for _, table := range []string{"jobs_queue", "job_runs"} {
		table := table
		t.Run(table, func(t *testing.T) {
			err := h.TxM.WithTenantRO(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
				var n int
				return db.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n)
			})
			if err == nil {
				t.Fatalf("app_rt must NOT be able to SELECT %s (grant-only isolation)", table)
			}
			if !strings.Contains(err.Error(), "permission denied") {
				t.Fatalf("expected a permission-denied error selecting %s, got: %v", table, err)
			}
		})
	}
}

// TestIntegrationJobsRunnerCrossTenantAfterRLS is the regression guard for M2:
// after FORCE RLS lands on jobs_queue/job_runs, the runner (app_platform pool)
// must still enqueue, claim, and run jobs across DIFFERENT tenants — the
// permissive app_platform policy keeps the cross-tenant claim scan and the
// job_runs writes working. Covers app_rt enqueue (WITH CHECK admits
// app_tenant_id()) and app_platform read/write cross-tenant in one flow.
func TestIntegrationJobsRunnerCrossTenantAfterRLS(t *testing.T) {
	h := testkit.NewDB(t)
	tenantA := testkit.CreateTenant(t, h)
	tenantB := testkit.CreateTenant(t, h)

	reg := jobs.NewRegistry()
	reg.RegisterKind(jobKind, func(context.Context, database.TenantDB, []byte) error {
		return nil // trivially succeeds; we only care the runner processes both tenants
	}, jobs.DefaultRetry())
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	// Enqueue one job for each tenant, each in its own app_rt business tx.
	for _, tn := range []testkit.TenantHandle{tenantA, tenantB} {
		if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
			return jobs.Enqueue(ctx, db, testJob{N: "cross"})
		}); err != nil {
			t.Fatalf("enqueue for %s: %v", tn.ID, err)
		}
	}

	// The runner (app_platform) claims both cross-tenant in one scan.
	r := jobs.NewRunner(h.Platform, h.TxM, reg)
	claimed, err := r.ClaimOnce(context.Background())
	if err != nil {
		t.Fatalf("ClaimOnce: %v", err)
	}
	if claimed != 2 {
		t.Fatalf("runner claimed %d jobs cross-tenant, want 2", claimed)
	}

	// Both jobs completed, and both have a 'succeeded' job_runs mirror row —
	// proving the app_platform pool read + wrote across both tenants under RLS.
	for _, tn := range []testkit.TenantHandle{tenantA, tenantB} {
		var id int64
		if err := h.Platform.QueryRow(context.Background(),
			`SELECT id FROM jobs_queue WHERE kind = $1 AND tenant_id = $2`, jobKind, tn.ID).Scan(&id); err != nil {
			t.Fatalf("read job id for %s: %v", tn.ID, err)
		}
		if s := jobStatus(t, h, id); s != "completed" {
			t.Fatalf("job for %s status = %q, want completed", tn.ID, s)
		}
		if c := countRuns(t, h, id, "succeeded"); c != 1 {
			t.Fatalf("job for %s: succeeded job_runs rows = %d, want 1", tn.ID, c)
		}
	}
}
