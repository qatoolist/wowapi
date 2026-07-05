package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/outbox"
)

// registerMaintenance wires the kernel's periodic sweeps onto the scheduler
// (roadmap R3 + S5): the cross-tenant idempotency-key expiry sweep and the
// per-tenant workflow SLA sweep. Both are idempotent, so the scheduler's
// leader-safe at-most-once-per-interval guarantee is sufficient — N worker
// replicas will not double-fire.
func registerMaintenance(sched *jobs.Scheduler, k *kernel.Kernel, slaEvery, idemEvery, dlqEvery, anchorEvery time.Duration) {
	// Idempotency-key expiry: one cross-tenant DELETE as app_platform. The
	// k.Platform pool already connects AS app_platform, so Platform() runs with
	// the cross-tenant sweep policy (migration 00012).
	platTxM := database.NewManager(k.Platform, config.DB{}, database.WithRole("app_platform"), database.WithRLSGuard())
	idem := database.NewIdemStore()
	sched.Register("kernel.idempotency.sweep", idemEvery, func(ctx context.Context) error {
		_, err := idem.SweepExpired(ctx, platTxM, time.Now())
		return err
	})

	// Workflow SLA timers: fan out one tenant-bound sweep per active tenant. One
	// tenant's failure is logged and does not block the others.
	sched.Register("kernel.workflow.sla", slaEvery, func(ctx context.Context) error {
		tenants, err := activeTenants(ctx, k)
		if err != nil {
			return err
		}
		for _, tid := range tenants {
			tctx := database.WithTenantID(ctx, tid)
			if err := k.Tx.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
				_, _, serr := k.WorkflowRuntime.SweepSLA(ctx, db, time.Now())
				return serr
			}); err != nil {
				k.Log.WarnContext(ctx, "scheduler: sla sweep failed for tenant", "tenant", tid, "err", err)
			}
		}
		return nil
	})

	// Data-lifecycle disposition: fan out per active tenant, running each
	// registered record class's Dispose (roadmap E2). A no-op until a product
	// registers record classes; one tenant's failure never blocks the rest.
	sched.Register("kernel.retention.disposition", slaEvery, func(ctx context.Context) error {
		tenants, err := activeTenants(ctx, k)
		if err != nil {
			return err
		}
		for _, tid := range tenants {
			tctx := database.WithTenantID(ctx, tid)
			if err := k.Tx.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
				_, serr := k.Retention.SweepDisposition(ctx, db, time.Now())
				return serr
			}); err != nil {
				k.Log.WarnContext(ctx, "scheduler: disposition sweep failed for tenant", "tenant", tid, "err", err)
			}
		}
		return nil
	})

	// DLQ depth: export dead-lettered jobs + outbox events as the dlq_depth
	// gauge (roadmap CA-1 / backlog B-8). Runs on the leader-safe scheduler, so
	// the depth is counted once per interval across replicas rather than
	// double-counted per replica. k.Metrics is NoOp unless an adapter is wired.
	sched.Register("kernel.dlq.depth", dlqEvery, func(ctx context.Context) error {
		if err := jobs.PublishDLQDepth(ctx, k.Platform, k.Metrics); err != nil {
			return err
		}
		return outbox.PublishDLQDepth(ctx, k.Platform, k.Metrics)
	})

	// Audit anchor-export: durably persist each tenant's audit-chain head into the
	// append-only audit_anchors table (roadmap CA-11). One cross-tenant
	// INSERT..SELECT as app_platform, so anchors are written once per interval
	// across replicas — an offline verifier can later detect tail-truncation of
	// the hash chain (which Verify alone cannot) by checking the live chain still
	// contains the last anchored (seq, hash).
	sched.Register("kernel.audit.anchor", anchorEvery, func(ctx context.Context) error {
		_, err := audit.ExportAnchors(ctx, k.Platform)
		return err
	})
}

// registerModuleRecurring wires each module-registered recurring job onto the
// scheduler (roadmap E5/CA-5). Like the kernel sweeps, each runs leader-safe and
// fans out per active tenant, giving the module's callback a tenant-bound DB in
// that tenant's transaction. One tenant's failure is logged and does not block
// the others.
func registerModuleRecurring(sched *jobs.Scheduler, k *kernel.Kernel, recurring []RecurringJob) {
	for _, rj := range recurring {
		rj := rj
		sched.Register(rj.Name, rj.Every, func(ctx context.Context) error {
			tenants, err := activeTenants(ctx, k)
			if err != nil {
				return err
			}
			for _, tid := range tenants {
				tctx := database.WithTenantID(ctx, tid)
				if err := k.Tx.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
					return rj.Run(ctx, db)
				}); err != nil {
					k.Log.WarnContext(ctx, "scheduler: module recurring job failed for tenant",
						"job", rj.Name, "tenant", tid, "err", err)
				}
			}
			return nil
		})
	}
}

// activeTenants lists tenant ids eligible for per-tenant maintenance. Read on the
// platform pool (app_platform holds the tenants-catalog grant).
func activeTenants(ctx context.Context, k *kernel.Kernel) ([]uuid.UUID, error) {
	rows, err := k.Platform.Query(ctx, `SELECT id FROM tenants WHERE status = 'active'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
