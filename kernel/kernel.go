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
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/attachment"
	kaudit "github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/comment"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/document"
	"github.com/qatoolist/wowapi/kernel/integration"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/notify"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/relationship"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/secrets"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/kernel/webhook"
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

	// Notification / webhook / integration framework (Phase 9). NotifyTemplates
	// and IntegrationProviders are the shared registration pointers modules write
	// into during Register; Notify/Webhooks/Integrations are the runtime services.
	NotifyTemplates      *notify.Registry
	Notify               *notify.Service
	Webhooks             *webhook.Service
	IntegrationProviders *integration.Registry
	Integrations         *integration.Store

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
	// Secrets resolves secret references (webhook signing secrets, integration
	// credentials). Optional; when nil, resolving a ref errors at use time.
	Secrets secrets.Provider
	// WebhookSender delivers outbound webhooks. Optional; nil → the real HTTP sender.
	WebhookSender webhook.Sender
}

// New wires the kernel. cfg travels by value (immutable, 12 §6). The returned
// kernel's Perms/Resources registries are the shared pointers modules register
// into during boot; the evaluator reads them at decision time.
func New(cfg config.Framework, log *slog.Logger, deps Deps) (*Kernel, error) {
	idgen := model.UUIDv7()

	// Audit sink: unless a product injects one, denials are written DURABLY to
	// audit_logs (not just logged) whenever a runtime TxManager is available — the
	// evaluator runs in a read-only tx, so the durable sink writes in its own tenant
	// transaction (see durableAudit). Without a TxManager (rare api-only wiring) it
	// falls back to WARN logs.
	audit := deps.Audit
	if audit == nil {
		if deps.Tx != nil {
			audit = durableAudit{log: log, txm: deps.Tx, writer: kaudit.New(idgen, nil)}
		} else {
			audit = loggingAudit{log: log}
		}
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

	// Notifications: template registry (module-declared) + service. Channel sender
	// adapters (smtp/sms/…) are infra registered by the product on Notify.
	notifyReg := notify.NewRegistry()
	notifySvc := notify.New(notifyReg, idgen)

	// Webhooks: a service over a Sender (real HTTP by default) and a secret-ref
	// resolver adapting the kernel secrets provider. Modules register verifiers +
	// inbound handlers on it during Register.
	sender := deps.WebhookSender
	if sender == nil {
		sender = webhook.NewHTTPSender()
	}
	webhookSvc := webhook.New(sender, secretRefResolver{p: deps.Secrets}, idgen)

	// Integrations: provider adapter registry + config/credential store.
	intReg := integration.NewRegistry()
	intStore := integration.NewStore(intReg, deps.Secrets, idgen)

	return &Kernel{
		Cfg:                  cfg,
		Log:                  log,
		Pool:                 deps.Pool,
		Platform:             deps.Platform,
		Tx:                   deps.Tx,
		Authz:                eval,
		Perms:                perms,
		Resources:            resources,
		Rules:                ruleReg,
		RulesResolver:        ruleResolver,
		Workflows:            wfReg,
		WorkflowRuntime:      wfRuntime,
		DocumentClasses:      docClasses,
		DocumentHooks:        docHooks,
		Documents:            docSvc,
		Comments:             commentSvc,
		Attachments:          attachmentSvc,
		NotifyTemplates:      notifyReg,
		Notify:               notifySvc,
		Webhooks:             webhookSvc,
		IntegrationProviders: intReg,
		Integrations:         intStore,
		audit:                audit,
	}, nil
}

// secretRefResolver adapts the kernel secrets.Provider to the webhook package's
// string-ref SecretResolver port (an endpoint's secret_ref → the signing secret).
type secretRefResolver struct{ p secrets.Provider }

func (r secretRefResolver) Resolve(ctx context.Context, ref string) (string, error) {
	if r.p == nil {
		return "", fmt.Errorf("kernel: no secrets provider wired to resolve %q", ref)
	}
	parsed, err := secrets.ParseRef(ref)
	if err != nil {
		return "", err
	}
	return r.p.Resolve(ctx, parsed)
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

// durableAudit is the default AuditSink: it logs the denial AND writes a durable
// audit_logs row. The evaluator runs in a read-only transaction, so the durable
// write cannot happen inline — it runs in its own tenant write tx, which also
// correctly persists the denial even when the request transaction rolls back.
// Best-effort: a durable-write failure is logged, never blocking the decision.
type durableAudit struct {
	log    *slog.Logger
	txm    database.TxManager
	writer *kaudit.Writer
}

func (a durableAudit) AuthzDenial(ctx context.Context, actor authz.Actor, perm string, t authz.Target, reason string) {
	loggingAudit{log: a.log}.AuthzDenial(ctx, actor, perm, t, reason)
	if a.txm == nil || a.writer == nil || actor.TenantID == uuid.Nil {
		return
	}
	// Detach from the request context so a client disconnect after the 403 does
	// not drop the audit record; bind the actor for attribution.
	wctx := database.WithActorID(database.WithTenantID(context.WithoutCancel(ctx), actor.TenantID), actor.CapacityID)
	err := a.txm.WithTenant(wctx, func(ctx context.Context, db database.TenantDB) error {
		return a.writer.Record(ctx, db, kaudit.Entry{
			Action:    "authz.denied",
			Reason:    reason,
			ActorKind: string(actor.Kind),
			Metadata: map[string]any{
				"permission":   perm,
				"target_scope": string(t.Scope),
				"break_glass":  actor.BreakGlass,
			},
		})
	})
	if err != nil && a.log != nil {
		a.log.WarnContext(ctx, "authz denial: durable audit write failed", "err", err, "permission", perm)
	}
}
