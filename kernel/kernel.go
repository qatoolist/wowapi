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
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/foundation/artifact"
	"github.com/qatoolist/wowapi/foundation/attachment"
	"github.com/qatoolist/wowapi/foundation/bulk"
	"github.com/qatoolist/wowapi/foundation/comment"
	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/foundation/integration"
	"github.com/qatoolist/wowapi/foundation/notify"
	"github.com/qatoolist/wowapi/foundation/webhook"
	kaudit "github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpclient"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/relationship"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/retention"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/secrets"
	"github.com/qatoolist/wowapi/kernel/sequence"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/kernel/workflow"
)

// Kernel owns infrastructure and kernel services. Fields are read-only after
// construction; the pool never leaves the kernel.
type Kernel struct {
	Cfg       config.Framework
	Log       *slog.Logger
	Pool      *pgxpool.Pool
	Platform  *pgxpool.Pool // app_platform pool for cross-tenant kernel work; see Deps.Platform for the wiring contract (nil only for migrate)
	Tx        database.TxManager
	ModelHash string // deterministic hash of the booted application model (AR-01); empty until AR-01 lands
	// PlatformTx is a tenant-bindable TxManager over the app_platform pool
	// (WithRole app_platform + RLS guard). It backs the scoped privileged services
	// (kernel/privileged, GAP-006): platform write privilege for the protected
	// relationships/rule_versions writes, but tenant-bound so RLS still isolates.
	// Nil only when Deps.Platform is nil (the migrate process, which serves no
	// tenant traffic and wires no privileged services).
	PlatformTx database.TxManager
	// RuleStore persists + activates rule versions; the privileged Rules service
	// delegates the supersede+activate state machine to it.
	RuleStore       *rules.Store
	Authz           authz.Evaluator
	Perms           *authz.Registry
	Resources       *resource.Registry
	Rules           *rules.Registry
	RulesResolver   *rules.Resolver
	Workflows       *workflow.Registry
	WorkflowRuntime *workflow.Runtime

	// Data lifecycle (roadmap E2). RetentionClasses is the shared record-class
	// registry modules register their dispose/export/erase callbacks into during
	// boot; Retention is the engine that drives scheduled disposition and DSR
	// fulfilment over them.
	RetentionClasses *retention.Registry
	Retention        *retention.Engine

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

	// Metrics is the observability sink shared by kernel components (RED
	// middleware, scheduler lag, DLQ depth, webhook breaker, rate-limit drops).
	// Never nil — defaults to observability.NoOp when no adapter is wired.
	Metrics observability.Metrics

	// Tracer is the distributed-tracing port; never nil (NoOpTracer default). The
	// outbox writer captures its trace context; the relay continues it (CA-9).
	Tracer observability.Tracer

	// AuthzCache is the per-actor assignment cache wrapping the evaluator's store
	// when Deps.AuthzCacheTTL > 0; nil when caching is disabled. It caches
	// ActiveAssignments (which pre-join role_permissions), so BOTH an actor's
	// role grants/revokes AND a role's permission set can otherwise be served
	// stale up to the TTL. To keep it fresh (CA-2):
	//   - actor_assignment grant/revoke (product-owned write): call
	//     AuthzCache.Invalidate(tenant, capacity) — or InvalidateTenant for a
	//     bulk change — right after the write commits.
	//   - seed / authorization-spine sync (roles + role_permissions): pass this
	//     handle to seeds.Sync, which calls InvalidateAll after the writes commit.
	// ABAC policies and ReBAC relationship edges are NOT cached (they pass
	// through / are checked directly each Evaluate), so a policy activation or a
	// granted_via edge change is never stale and needs no invalidation.
	AuthzCache *authz.CachingStore

	// Evidence-layer services exposed to modules via module.Context (roadmap CA-11):
	// Audit (field-level audit + hash chain), Sequence (gap-free numbering), Bulk
	// (chunked resumable ops), Artifacts (immutable versioned artifacts).
	Audit     *kaudit.Writer
	Sequence  *sequence.Allocator
	Bulk      *bulk.Service
	Artifacts *artifact.Pipeline

	auditSink authz.AuditSink
}

// Deps injects the pools/tx (built by the product main, or provided by testkit)
// so the kernel does not hard-code pool construction and stays testable.
type Deps struct {
	Pool *pgxpool.Pool
	// Platform is the app_platform pool for cross-tenant kernel work (outbox relay,
	// job runner, seed sync) and, in the api, cross-tenant API-key verification. Both
	// the generated api and worker mains build it BEFORE Boot and wire it here, and
	// app.Boot's RLS-enforcement check (M3) validates it when non-nil. It is nil only
	// for the migrate process, which serves no tenant traffic and opts out of that
	// check via app.SkipRLSEnforcementCheck. A custom api/worker main MUST wire it
	// too — leaving it nil silently skips the M3 backstop for the platform pool.
	Platform *pgxpool.Pool
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
	// Metrics is the observability sink (Prometheus adapter in production).
	// Optional; nil → observability.NoOp so call sites never nil-check.
	Metrics observability.Metrics
	// Tracer is the distributed-tracing port (OTel adapter in production).
	// Optional; nil → observability.NoOpTracer. Wired into the outbox writer so
	// emitted events carry the request's trace context across the async boundary
	// (roadmap O1/CA-9).
	Tracer observability.Tracer
	// AuthzCacheTTL, when > 0, wraps the authorization store in a per-actor
	// ActiveAssignments cache with this TTL (roadmap R1/CA-2). DEFAULT OFF
	// (zero) — enabling it accepts up-to-TTL stale-allow after a revocation on
	// another pod, so keep the TTL short and call Kernel.AuthzCache.Invalidate
	// from your role grant/revoke paths for immediate effect on this pod.
	AuthzCacheTTL time.Duration
	// StepUpStrongFactors overrides the default set of AMR values that satisfy
	// a plain `step_up: true` permission (backlog B8). Empty/nil uses
	// authz.DefaultStrongFactors (mfa, otp, totp, hwk, fpt, face — "sms" is
	// EXCLUDED by default per Decision 5: SMS-based step-up is opt-in only). A
	// deployment re-adds "sms" — or narrows/widens the set further — by setting
	// this slice, with NO CODE CHANGES. Per-permission AMR requirements
	// (StepUpPolicy, e.g. "require hwk specifically") are declared in seed YAML
	// instead and are unaffected by this deployment-wide default.
	StepUpStrongFactors []string
	// StepUpDefaultChallenge overrides the factor/hint advertised in
	// WWW-Authenticate for a plain `step_up: true` permission. Empty uses "mfa".
	StepUpDefaultChallenge string
}

// newArtifactWriter builds the production DSR artifact writer from environment
// variables. The key is read from WOWAPI_DSR_ARTIFACT_KEY (hex, 32 bytes);
// when absent a deterministic test key is used so local/test boots succeed, but
// deployments must set the variable to avoid a shared fallback key.
func newArtifactWriter(log *slog.Logger, audit *kaudit.Writer) retention.ArtifactWriter {
	dir := os.Getenv("WOWAPI_ARTIFACT_DIR")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "wowapi-artifacts")
	}
	keyHex := os.Getenv("WOWAPI_DSR_ARTIFACT_KEY")
	var key []byte
	if keyHex != "" {
		var err error
		key, err = hex.DecodeString(keyHex)
		if err != nil || len(key) != 32 {
			log.WarnContext(context.Background(), "kernel: invalid WOWAPI_DSR_ARTIFACT_KEY; falling back to test key", "decode_err", err)
			key = retention.TestKey()
		}
	} else {
		log.WarnContext(context.Background(), "kernel: WOWAPI_DSR_ARTIFACT_KEY not set; using test key")
		key = retention.TestKey()
	}
	return retention.NewFileArtifactWriter(dir, key, audit)
}

// New wires the kernel. cfg travels by value (immutable, 12 §6). The returned
// kernel's Perms/Resources registries are the shared pointers modules register
// into during boot; the evaluator reads them at decision time.
func New(cfg config.Framework, log *slog.Logger, deps Deps) (*Kernel, error) {
	idgen := model.UUIDv7()

	// Metrics sink: NoOp unless a product wires an adapter (e.g. Prometheus),
	// so every emission call site is nil-safe.
	metrics := deps.Metrics
	if metrics == nil {
		metrics = observability.NoOp
	}
	// Tracer: NoOp unless a product wires an OTel adapter.
	tracer := deps.Tracer
	if tracer == nil {
		tracer = observability.NoOpTracer
	}

	// Shared audit writer: used both as the durable authz-denial sink below and
	// exposed to modules via Kernel.Audit / module.Context.Audit (roadmap CA-11).
	auditWriter := kaudit.New(idgen, nil)

	// Audit sink: unless a product injects one, denials are written DURABLY to
	// audit_logs (not just logged) whenever a runtime TxManager is available — the
	// evaluator runs in a read-only tx, so the durable sink writes in its own tenant
	// transaction (see durableAudit). Without a TxManager (rare api-only wiring) it
	// falls back to WARN logs.
	audit := deps.Audit
	if audit == nil {
		if deps.Tx != nil {
			audit = durableAudit{log: log, txm: deps.Tx, writer: auditWriter}
		} else {
			audit = loggingAudit{log: log}
		}
	}
	perms := authz.NewRegistry()
	resources := resource.NewRegistry()

	// Authorization store: the DB-backed store, optionally wrapped in the
	// per-actor assignment cache when a TTL is configured (roadmap R1/CA-2).
	// Off by default: caching accepts up-to-TTL stale-allow, so it is opt-in.
	var authzStore authz.Store = authz.NewStore()
	var authzCache *authz.CachingStore
	if deps.AuthzCacheTTL > 0 {
		authzCache = authz.NewCachingStore(authzStore, deps.AuthzCacheTTL)
		authzStore = authzCache
	}

	eval := authz.New(authz.Options{
		Store:            authzStore,
		Registry:         perms,
		Policies:         policy.New(),
		Relationships:    relationship.NewChecker(),
		Audit:            audit,
		StrongFactors:    deps.StepUpStrongFactors,
		DefaultChallenge: deps.StepUpDefaultChallenge,
	})

	writer := outbox.NewWriter(idgen, outbox.WithWriterTracer(tracer))

	// Rules: registry + resolver (org ancestry via the authz store) — the
	// registry is populated during module Register; the resolver reads it.
	// Reuses the composed authzStore (not a fresh authz.NewStore()) so this
	// path honors the same caching/decoration as the evaluator (AR-06 T1).
	ruleReg := rules.NewRegistry()
	orgAncestry := func(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error) {
		return authzStore.OrgAncestors(ctx, db, orgID)
	}
	ruleResolver := rules.NewResolver(ruleReg, orgAncestry)
	ruleStore := rules.NewStore(ruleReg, idgen)

	// Platform TxManager backing the scoped privileged services (kernel/privileged,
	// GAP-006): the app_platform pool, tenant-bindable (WithRole + RLS guard) so a
	// privileged write runs with platform grants yet stays tenant-isolated by RLS.
	// Nil when no platform pool is wired (migrate).
	var platformTx database.TxManager
	if deps.Platform != nil {
		platformTx = database.NewManager(deps.Platform, cfg.DB,
			database.WithRole("app_platform"), database.WithRLSGuard())
	}

	// Workflow: registry + runtime (shares the tx, evaluator, outbox writer).
	wfReg := workflow.NewRegistry()
	wfRuntime := workflow.NewRuntime(deps.Tx, wfReg, eval, writer, idgen, auditWriter, workflow.WithRuntimeMetrics(metrics))

	// Documents: the class registry + hook set modules register into, plus the
	// service (only when an object-storage adapter is wired). The two document
	// permissions are kernel-owned so the download gate can Evaluate them.
	// Data lifecycle: the record-class registry modules register into, and the
	// engine that drives disposition and DSR over it (roadmap E2).
	retClasses := retention.NewRegistry()
	retHolds := retention.NewHolds(idgen)
	retArtifacts := newArtifactWriter(log, auditWriter)
	retEngine := retention.NewEngine(retClasses, retention.NewDSR(idgen), retHolds, retArtifacts, auditWriter)

	docClasses := document.NewRegistry()
	docHooks := document.NewHooks()
	perms.Register(authz.Permission{Key: document.PermRead})
	perms.Register(authz.Permission{Key: document.PermWrite})
	var docSvc *document.Service
	if deps.Storage != nil {
		docSvc = document.New(docClasses, deps.Storage, eval, writer, docHooks, idgen, document.WithAudit(auditWriter))
	}
	commentSvc := comment.New(idgen, writer)
	attachmentSvc := attachment.New(idgen, writer)

	// Notifications: template registry (module-declared) + service. Channel sender
	// adapters (smtp/sms/…) are infra registered by the product on Notify.
	notifyReg := notify.NewRegistry()
	notifySvc := notify.New(notifyReg, idgen, notify.WithTracer(tracer), notify.WithOutbox(writer))

	// Webhooks: a service over a Sender (SSRF-safe HTTP by default, backlog B2)
	// and a secret-ref resolver adapting the kernel secrets provider. Modules
	// register verifiers + inbound handlers on it during Register.
	sender := deps.WebhookSender
	if sender == nil {
		out := cfg.Webhook.Outbound
		var senderOpts []webhook.HTTPSenderOption
		if out.SSRFProtectionDisabled {
			senderOpts = append(senderOpts, webhook.WithSSRFProtectionDisabled())
		} else {
			senderOpts = append(senderOpts, webhook.WithHTTPClientConfig(httpclient.Config{
				AllowedHosts: out.AllowedHosts,
				AllowedCIDRs: out.AllowedCIDRs,
			}))
		}
		sender = webhook.NewHTTPSender(senderOpts...)
	}

	// SEC-06 T2/T3: boot-time egress-exception report and allowlist change
	// audit. The report is credential-free by construction; the audit record
	// compares the loaded allowlist to the compiled defaults.
	if exceptions := cfg.EgressExceptions(); len(exceptions) > 0 {
		log.InfoContext(context.Background(), "egress_exceptions",
			slog.Any("exceptions", exceptions),
			slog.String("environment", string(cfg.Environment)),
		)
	}
	config.RecordAllowlistChange(config.Defaults().Webhook.Outbound, cfg.Webhook.Outbound, func(change config.AllowlistChange) {
		log.InfoContext(context.Background(), "config_change",
			slog.String("action", change.Action),
			slog.Any("old_hosts", change.OldHosts),
			slog.Any("new_hosts", change.NewHosts),
			slog.Any("old_cidrs", change.OldCIDRs),
			slog.Any("new_cidrs", change.NewCIDRs),
		)
	})

	webhookSvc := webhook.New(sender, secretRefResolver{p: deps.Secrets}, idgen, webhook.WithMetrics(metrics))

	// Integrations: provider adapter registry + config/credential store.
	intReg := integration.NewRegistry()
	intStore := integration.NewStore(intReg, deps.Secrets, idgen)

	return &Kernel{
		Cfg:                  cfg,
		Log:                  log,
		Pool:                 deps.Pool,
		Platform:             deps.Platform,
		Tx:                   deps.Tx,
		PlatformTx:           platformTx,
		RuleStore:            ruleStore,
		Authz:                eval,
		Perms:                perms,
		Resources:            resources,
		Rules:                ruleReg,
		RulesResolver:        ruleResolver,
		Workflows:            wfReg,
		WorkflowRuntime:      wfRuntime,
		RetentionClasses:     retClasses,
		Retention:            retEngine,
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
		Metrics:              metrics,
		Tracer:               tracer,
		AuthzCache:           authzCache,
		Audit:                auditWriter,
		Sequence:             sequence.New(idgen),
		Bulk:                 bulk.New(idgen, bulk.WithLogger(log)),
		Artifacts:            artifact.New(idgen),
		auditSink:            audit,
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
