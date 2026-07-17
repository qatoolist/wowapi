package app

import (
	"context"
	"errors"
	"fmt"
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
func registerMaintenance(sched *jobs.Scheduler, k *kernel.Kernel, slaEvery, idemEvery, dlqEvery, anchorEvery, notifyEvery, webhookEvery, uploadSessionEvery time.Duration) {
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
		return forEachActiveTenant(ctx, k, "kernel.workflow.sla", func(ctx context.Context, tid uuid.UUID) error {
			tctx := database.WithTenantID(ctx, tid)
			return k.Tx.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
				_, _, serr := k.WorkflowRuntime.SweepSLA(ctx, db, time.Now())
				return serr
			})
		})
	})

	// Data-lifecycle disposition: fan out per active tenant, running each
	// registered record class's Dispose (roadmap E2). A no-op until a product
	// registers record classes; one tenant's failure never blocks the rest.
	sched.Register("kernel.retention.disposition", slaEvery, func(ctx context.Context) error {
		return forEachActiveTenant(ctx, k, "kernel.retention.disposition", func(ctx context.Context, tid uuid.UUID) error {
			tctx := database.WithTenantID(ctx, tid)
			return k.Tx.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
				_, serr := k.Retention.SweepDisposition(ctx, db, time.Now())
				return serr
			})
		})
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

	// Notification send/retry: drive queued and retriable-failed deliveries per
	// active tenant (Phase 9). Ships the async fan-out so products no longer
	// hand-roll per-tenant polling (H3). The tenant is bound from the enumeration
	// and passed straight to SendPending — never a caller-supplied param — and
	// SendPending runs each tenant in its own app_platform tx. Leader-safe: the
	// scheduler fires this once per interval across replicas. Guarded on Notify
	// being wired (nil in api-only postures that skip the notification framework).
	if k.Notify != nil {
		sched.Register("kernel.notify.send_pending", notifyEvery, func(ctx context.Context) error {
			return forEachActiveTenant(ctx, k, "kernel.notify.send_pending", func(ctx context.Context, tid uuid.UUID) error {
				_, serr := k.Notify.SendPending(ctx, platTxM, tid, time.Now())
				return serr
			})
		})
	}

	// Webhook retry + inbound processing: re-drive failed outbound deliveries and
	// run handlers for pending inbound events per active tenant (Phase 9, H3).
	// Closes H2 by construction: the tenant comes from the enumeration and binds
	// the whole dispatch, so no decoupled tenant param can reach the signing/
	// endpoint-lookup path. One tenant's failure is logged and never blocks the
	// rest. Guarded on Webhooks being wired.
	if k.Webhooks != nil {
		sched.Register("kernel.webhook.retry", webhookEvery, func(ctx context.Context) error {
			return forEachActiveTenant(ctx, k, "kernel.webhook.retry", func(ctx context.Context, tid uuid.UUID) error {
				var errs []error
				if rerr := k.Webhooks.RetryOutbound(ctx, platTxM, tid, time.Now()); rerr != nil {
					errs = append(errs, fmt.Errorf("retry_outbound: %w", rerr))
				}
				if perr := k.Webhooks.ProcessInbound(ctx, platTxM, tid, time.Now()); perr != nil {
					errs = append(errs, fmt.Errorf("process_inbound: %w", perr))
				}
				return errors.Join(errs...)
			})
		})
	}

	// Upload session GC: fan out per active tenant, expiring pending sessions that
	// never reached ConfirmUpload and deleting their orphaned blobs. Guarded on
	// Documents being wired (nil in api-only postures without storage).
	if k.Documents != nil {
		sched.Register("kernel.document.upload_session_sweep", uploadSessionEvery, func(ctx context.Context) error {
			return forEachActiveTenant(ctx, k, "kernel.document.upload_session_sweep", func(ctx context.Context, tid uuid.UUID) error {
				n, serr := k.Documents.SweepUploadSessions(ctx, platTxM, tid, time.Now())
				if serr != nil {
					return serr
				}
				if n > 0 {
					k.Log.InfoContext(ctx, "scheduler: expired upload sessions swept", "tenant", tid, "count", n)
				}
				return nil
			})
		})
	}
}

// registerModuleRecurring wires each module-registered recurring job onto the
// scheduler (roadmap E5/CA-5). Like the kernel sweeps, each runs leader-safe and
// fans out per active tenant, giving the module's callback a tenant-bound DB in
// that tenant's transaction. One tenant's failure is logged and does not block
// the others.
func registerModuleRecurring(sched *jobs.Scheduler, k *kernel.Kernel, recurring []RecurringJob) {
	for _, rj := range recurring {
		sched.Register(rj.Name, rj.Every, func(ctx context.Context) error {
			return forEachActiveTenant(ctx, k, rj.Name, func(ctx context.Context, tid uuid.UUID) error {
				tctx := database.WithTenantID(ctx, tid)
				return k.Tx.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
					return rj.Run(ctx, db)
				})
			})
		})
	}
}

// forEachActiveTenant fans a maintenance function out across every active
// tenant, continuing past per-tenant failures so one tenant never blocks the
// rest — but returning the JOINED per-tenant errors instead of swallowing them
// (adversarial review 2026-07-17, F-09: the observer and
// scheduler_task_errors_total previously reported success even when every
// tenant failed). Failed tenants retry at the task's next interval; the
// schedule advances regardless.
func forEachActiveTenant(ctx context.Context, k *kernel.Kernel, task string, fn func(ctx context.Context, tid uuid.UUID) error) error {
	tenants, err := activeTenants(ctx, k)
	if err != nil {
		return err
	}
	var errs []error
	for _, tid := range tenants {
		if err := fn(ctx, tid); err != nil {
			k.Log.WarnContext(ctx, "scheduler: task failed for tenant",
				"task", task, "tenant", tid, "err", err)
			errs = append(errs, fmt.Errorf("tenant %s: %w", tid, err))
		}
	}
	return errors.Join(errs...)
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
