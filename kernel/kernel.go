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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/attachment"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/comment"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/document"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/relationship"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/kernel/workflow"
)

// Kernel owns infrastructure and kernel services. Fields are read-only after
// construction; the pool never leaves the kernel.
type Kernel struct {
	Cfg             config.Framework
	Log             *slog.Logger
	Pool            *pgxpool.Pool
	Platform        *pgxpool.Pool // app_platform pool for cross-tenant kernel work (relay, job runner, seed sync); may be nil in api-only processes
	Tx              database.TxManager
	Authz           authz.Evaluator
	Perms           *authz.Registry
	Resources       *resource.Registry
	Rules           *rules.Registry
	RulesResolver   *rules.Resolver
	Workflows       *workflow.Registry
	WorkflowRuntime *workflow.Runtime

	// Document / file framework (Phase 8). DocumentClasses + DocumentHooks are the
	// shared registration pointers modules write into during Register; Documents is
	// nil when no storage adapter is provided (an api-only process may run without
	// object storage). Comments/Attachments need no storage.
	DocumentClasses *document.Registry
	DocumentHooks   *document.Hooks
	Documents       *document.Service
	Comments        *comment.Service
	Attachments     *attachment.Service

	audit authz.AuditSink
}

// Deps injects the pools/tx (built by the product main, or provided by testkit)
// so the kernel does not hard-code pool construction and stays testable.
type Deps struct {
	Pool     *pgxpool.Pool
	Platform *pgxpool.Pool // optional; required only for the worker/migrate processes
	Tx       database.TxManager
	Audit    authz.AuditSink // optional; nil → a logging sink
	// Storage is the object-storage adapter backing the document framework.
	// Optional: when nil, Kernel.Documents is nil and modules that require it fail
	// boot only if they actually register a document class (checked by app.Boot).
	Storage storage.Adapter
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

	idgen := model.UUIDv7()
	writer := outbox.NewWriter(idgen)

	// Rules: registry + resolver (org ancestry via the authz store) — the
	// registry is populated during module Register; the resolver reads it.
	ruleReg := rules.NewRegistry()
	orgAncestry := func(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error) {
		return authz.NewStore().OrgAncestors(ctx, db, orgID)
	}
	ruleResolver := rules.NewResolver(ruleReg, orgAncestry)

	// Workflow: registry + runtime (shares the tx, evaluator, outbox writer).
	wfReg := workflow.NewRegistry()
	wfRuntime := workflow.NewRuntime(deps.Tx, wfReg, eval, writer, idgen)

	// Documents: the class registry + hook set modules register into, plus the
	// service (only when an object-storage adapter is wired). The two document
	// permissions are kernel-owned so the download gate can Evaluate them.
	docClasses := document.NewRegistry()
	docHooks := document.NewHooks()
	perms.Register(authz.Permission{Key: document.PermRead})
	perms.Register(authz.Permission{Key: document.PermWrite})
	var docSvc *document.Service
	if deps.Storage != nil {
		docSvc = document.New(docClasses, deps.Storage, eval, writer, docHooks, idgen)
	}
	commentSvc := comment.New(idgen, writer)
	attachmentSvc := attachment.New(idgen, writer)

	return &Kernel{
		Cfg:             cfg,
		Log:             log,
		Pool:            deps.Pool,
		Platform:        deps.Platform,
		Tx:              deps.Tx,
		Authz:           eval,
		Perms:           perms,
		Resources:       resources,
		Rules:           ruleReg,
		RulesResolver:   ruleResolver,
		Workflows:       wfReg,
		WorkflowRuntime: wfRuntime,
		DocumentClasses: docClasses,
		DocumentHooks:   docHooks,
		Documents:       docSvc,
		Comments:        commentSvc,
		Attachments:     attachmentSvc,
		audit:           audit,
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
