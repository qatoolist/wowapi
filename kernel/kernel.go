// Package kernel is wowapi's infrastructure composition root: it owns the
// database pool, the transaction manager, and the kernel services (the authz
// evaluator, and later outbox/jobs/documents/…). It is built once, in explicit
// order, by the product's app.App — never by a service locator (blueprint
// 06 §3).
//
// The authz evaluator is constructed over a SHARED permission-registry pointer.
// Modules register their permissions into that same registry during
// Module.Register; the registry is fully populated before any request is
// served, and the evaluator reads it at decision time. App gates boot on the
// registry's Err() so an unregistered/duplicate permission fails startup, not a
// runtime request (closes the Phase 4 deferral).
package kernel

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/relationship"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// Kernel owns infrastructure and kernel services. Fields are read-only after
// construction; the pool never leaves the kernel.
type Kernel struct {
	Cfg       config.Framework
	Log       *slog.Logger
	Pool      *pgxpool.Pool
	Platform  *pgxpool.Pool // app_platform pool for cross-tenant kernel work (relay, job runner, seed sync); may be nil in api-only processes
	Tx        database.TxManager
	Authz     authz.Evaluator
	Perms     *authz.Registry
	Resources *resource.Registry
	audit     authz.AuditSink
}

// Deps injects the pools/tx (built by the product main, or provided by testkit)
// so the kernel does not hard-code pool construction and stays testable.
type Deps struct {
	Pool     *pgxpool.Pool
	Platform *pgxpool.Pool // optional; required only for the worker/migrate processes
	Tx       database.TxManager
	Audit    authz.AuditSink // optional; nil → a logging sink
}

// New wires the kernel. cfg travels by value (immutable, 12 §6). The returned
// kernel's Perms/Resources registries are the shared pointers modules register
// into during boot; the evaluator reads them at decision time.
func New(cfg config.Framework, log *slog.Logger, deps Deps) (*Kernel, error) {
	audit := deps.Audit
	if audit == nil {
		audit = loggingAudit{log: log}
	}
	perms := authz.NewRegistry()
	resources := resource.NewRegistry()

	eval := authz.New(authz.Options{
		Store:         authz.NewStore(),
		Registry:      perms,
		Policies:      policy.New(),
		Relationships: relationship.NewChecker(),
		Audit:         audit,
	})

	return &Kernel{
		Cfg:       cfg,
		Log:       log,
		Pool:      deps.Pool,
		Platform:  deps.Platform,
		Tx:        deps.Tx,
		Authz:     eval,
		Perms:     perms,
		Resources: resources,
		audit:     audit,
	}, nil
}

// loggingAudit is the Phase 5 AuditSink: it logs authorization denials at WARN.
// The durable audit_logs writer replaces it in Phase 6.
type loggingAudit struct{ log *slog.Logger }

func (a loggingAudit) AuthzDenial(ctx context.Context, actor authz.Actor, perm string, t authz.Target, reason string) {
	if a.log == nil {
		return
	}
	a.log.WarnContext(ctx, "authz denial",
		"permission", perm,
		"reason", reason,
		"actor_user_id", actor.UserID.String(),
		"actor_capacity_id", actor.CapacityID.String(),
		"tenant_id", actor.TenantID.String(),
		"impersonating", actor.ImpersonatorUserID != actor.UserID && actor.ImpersonatorUserID.String() != "00000000-0000-0000-0000-000000000000",
		"break_glass", actor.BreakGlass,
		"target_scope", string(t.Scope),
	)
}
