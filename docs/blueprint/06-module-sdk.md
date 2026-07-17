# 06 — Module Starter Template, Registration Contract, DI/Bootstrap, Hooks

## 1. Module starter template

Modules live in the **consuming product repository** and are scaffolded there by
`wowapi new-module requests` (templates embedded in the CLI — see
[11-framework-distribution-and-consumption.md](11-framework-distribution-and-consumption.md)).
The framework repo keeps a private neutral fixture at `wowapi/internal/testmodules/requests` for
contract tests, plus optional standalone examples under `wowapi/examples/*`. Examples are
non-contractual and never imported by the framework core.

```text
<product-repo>/internal/modules/requests/   # neutral example module
  module.go            # Module impl: Name, DependsOn, Register(ctx). Wiring ONLY — no logic.
  domain/
    model.go           # aggregates + invariant methods (Request.Approve(actor, now) error). No SQL, no HTTP, no kernel service calls except model types.
    errors.go          # sentinel/domain errors built on kernel/errors kinds.
    events.go          # event type consts + payload structs (requests.request.approved).
    validation.go      # domain rule validation funcs (cross-field, state rules).
  app/
    commands.go        # command structs.        NOT: business logic.
    queries.go         # query structs.
    service.go         # orchestration (see 05). NOT: SQL, HTTP types.
    ports.go           # consumer-side interfaces this module needs (incl. other modules' capabilities).
  api/
    routes.go          # route table w/ RouteMeta. NOT: handler bodies.
    handlers.go        # thin handlers.           NOT: business rules, SQL.
    dto.go             # wire structs.            NOT: domain types serialized directly.
    mapper.go          # dto ↔ domain/command.    Boring by design; generated where possible.
  store/
    queries.sql        # ALL SQL for this module. NOT: other modules' tables.
    repository.go      # repo structs over sqlc gen; dynamic filter/sort via kernel builders.
    sqlc/              # generated; committed.
  seeds/
    permissions.yaml   # permission catalog entries (key, description, sensitive).
    roles.yaml         # role templates + permission lists.
    resource_types.yaml
    relationship_types.yaml
    workflows.yaml     # definitions (02-workflow-rules format).
    rules.yaml         # rule points + platform defaults.
    document_classes.yaml
    notification_templates.yaml
  migrations/          # goose files, prefix-ordered, ONLY this module's tables.
  tests/               # integration tests using wowapi/testkit (module contract test included).
```

`ports.go` is the *only* thing another module may import (`modules/x/app` port interfaces are
re-exported via a tiny `modules/x/port` package to keep the import surface explicit and lintable).

## 2. Module registration contract (`wowapi/module` — public)

The SDK is a public package: product repos import `github.com/qatoolist/wowapi/module` (plus the
`wowapi/kernel/*` contracts it references). The embedded-asset methods (`Migrations(fs)`,
`Seeds(fs)`, `OpenAPI(fragment)`) are public contracts precisely so *external* modules can hand
their `embed.FS` assets to the framework.

<!-- doc-example: illustrative -->
```go
// Package module — imported as wowapi/module; Context avoids stutter.
type Module interface {
    Name() string                 // "requests"
    DependsOn() []string          // module names; cycle → boot failure
    Register(ctx Context) error
}

// Context is capability-scoped: modules get registries and services, never raw pools.
// This listing matches module/module.go's Context interface method-for-method.
type Context interface {
    Logger() *slog.Logger                        // module-scoped structured logger
    Config() config.ModuleView                   // strict Decode of modules.<name>.* ONLY — no global framework config (see 12)

    // registration
    Routes() *httpx.Router                        // Handle(method, pattern, meta, h)
    Validator() *validation.Validator             // shared request validator (httpx.BindAndValidate)
    Permissions() *authz.Registry                // usually fed from seeds/permissions.yaml
    Resources() *resource.Registry
    Authz() authz.Evaluator                       // record-level checks + list filtering

    Tx() database.TxManager
    IDGen() model.IDGen

    Migrations(fsys fs.FS)                        // embedded goose dir
    Seeds(fsys fs.FS)                             // embedded yaml bundle
    OpenAPI(fragment []byte)
    I18n(bundle i18n.Bundle)                      // module's localized message bundle, one call per locale

    Health(name string, check func(context.Context) error)

    ProvidePort(name string, impl any)            // declare a port for dependents
    Port(name string) (any, error)                // fetch another module's declared port (checked at boot)

    Events() *outbox.HandlerRegistry              // Subscribe(eventType, handlerName, fn)
    Outbox() outbox.Writer                        // emit events in a business transaction

    Jobs() *jobs.Registry                         // RegisterKind(kind, worker, retryPolicy)
    RecurringJob(name string, every time.Duration, fn func(ctx context.Context, db database.TenantDB) error)

    Rules() *rules.Registry                       // rule-point registry
    RulesResolver() *rules.Resolver               // effective rule values
    Workflows() *workflow.Registry                // definitions, auto-actions, assignee resolvers
    WorkflowRuntime() *workflow.Runtime           // drive workflow instances

    RetentionClasses() *retention.Registry        // dispose/export/erase callbacks

    // Evidence-layer services, all operating in the caller's tenant transaction.
    Audit() *kaudit.Writer                        // field-level change capture + hash chain
    Sequence() *sequence.Allocator                // gap-free per-tenant numbered series
    Bulk() *bulk.Service                          // chunked, resumable bulk operations
    Artifacts() *artifact.Pipeline                // immutable versioned artifacts

    Privileged() *privileged.Services             // scoped, audited PLATFORM-privilege operations

    // Document / file framework.
    DocumentClasses() *document.Registry
    DocumentHooks() *document.Hooks               // OnFileUpload / OnDocumentAccess
    Documents() *document.Service                 // nil when no object-storage adapter is wired
    Comments() *comment.Service
    Attachments() *attachment.Service

    // Notification / webhook / integration framework.
    NotifyTemplates() *notify.Registry
    Notify() *notify.Service
    Webhooks() *webhook.Service
    IntegrationProviders() *integration.Registry
    Integrations() *integration.Store
}
```

<!-- doc-example: illustrative -->
```go
// <product-repo>/internal/modules/requests/module.go — complete wiring example
import "github.com/qatoolist/wowapi/module"

func (m Module) Register(mc module.Context) error {
    repo := store.NewRequestRepo()
    svc := app.NewService(repo, mc.RulesResolver(), mc.WorkflowRuntime(),
        mc.Authz(), mc.IDGen())
    h := api.NewHandlers(svc, mc.Tx(), validation.Default())

    api.MountRoutes(mc.Routes(), h)
    mc.Seeds(seedsFS); mc.Migrations(migrationsFS); mc.OpenAPI(openapiFragment)
    mc.Events().Subscribe("documents.document.uploaded", "requests.attach-scan", svc.OnDocumentUploaded)
    mc.Jobs().RegisterKind("requests.sla-sweep", app.NewSLASweeper(svc), jobs.DefaultRetry())
    mc.Workflows().RegisterAutoAction("requests.provision", svc.AutoProvision)
    mc.ProvidePort("requests.Lookup", app.NewLookupPort(svc))
    return nil
}
```

Module config: declare a typed struct with defaults + validation and decode it once in `Register`
(`mc.Config().Decode(&cfg)`); pass values into constructors. Decode errors, failed validation, or
unknown keys in the module's namespace fail boot. Modules never read env vars or global config
directly ([12-configuration-and-deployment.md](12-configuration-and-deployment.md) §2).

**Lifecycle:** `Register` (collect) → `Validate` (whole-graph checks: dup permission keys, routes
without meta, unknown workflow auto-actions, seed schema errors, unsatisfied ports, module config
decode/validation errors, dependency
cycles — all boot failures with precise messages) → `Migrate` (cmd/migrate only) → `SeedSync`
(idempotent upsert of catalogs/templates; tenant data never touched) → `Start` (HTTP or workers)
→ `Stop` (reverse order: stop intake, drain jobs/outbox with deadline, close pools).

## 3. DI / IoC / bootstrap — manual composition root

No container, no reflection, no service locator. Two structs:

<!-- doc-example: illustrative -->
```go
// kernel.Kernel: owns infrastructure + kernel services. Built once, explicit order.
type Kernel struct {
    Cfg     config.Framework
    Log     *slog.Logger
    DB      *pgxpool.Pool          // never leaves the kernel
    Tx      database.TxManager
    Authz   authz.Evaluator
    Rules   *rules.Engine
    WF      *workflow.Engine
    Audit   audit.Writer
    Outbox  *outbox.Outbox
    Jobs    *jobs.Runner
    Docs    document.Service
    Notify  *notify.Dispatcher
    Hooks   *hooks.Registry
    Health  *health.Registry
    …
}
func New(ctx context.Context, cfg config.Framework) (*Kernel, error) // ordered construction; any failure aborts boot

// app.App: kernel + modules. (Both public: wowapi/app composes wowapi/module + wowapi/kernel.)
type App struct { K *kernel.Kernel; modules []module.Module; registries *module.Registries }
func (a *App) Register(ms ...module.Module)
func (a *App) Validate() error
func (a *App) StartAPI(ctx) error / StartWorker(ctx) error / Shutdown(ctx) error
```

<!-- doc-example: illustrative -->
```go
// <product-repo>/cmd/api/main.go (pseudocode) — the composition root belongs to the PRODUCT app
import (
    "context"

    "github.com/qatoolist/wowapi/app"

    "example.com/acme-ops/internal/appcfg"       // product-owned Config type, scaffolded by `wowapi init`
    "example.com/acme-ops/internal/modules/assets"
    "example.com/acme-ops/internal/modules/requests"
)

func main() {
    cfg := appcfg.MustLoad()                     // product Config embeds config.Framework (see 12 §2)
    ctx := context.Background()                  // production main wraps this with SIGTERM/SIGINT handling
    die(app.RunAPI(ctx, cfg, requests.Module{}, assets.Module{})) // society.Module{} lives in ITS repo
}
```

Rules: constructors take interfaces they consume; `Kernel` fields are concrete where there's one
impl, interface where an adapter boundary exists (storage, mail, secrets). Circular deps are
blocked by the public package graph (`kernel` imports no `module`/`app`, `module` imports kernel
contracts, adapters implement kernel ports, `app` wires them); module↔module cycles are caught by
`DependsOn` topo-sort. Wire adoption is
optional later — the composition root is already the shape Wire generates. Testing: `testkit.NewApp(t)`
builds a real Kernel on a testcontainer DB with fake clock/idgen/providers injected via config.

Startup order: config → logger → DB pool (+ping) → migrations check (refuse to serve on drift) →
kernel services → module Register → Validate → seed sync → HTTP/workers → health=ready.
Shutdown order: health=not-ready → stop accepting (HTTP drain, job intake off) → wait in-flight
(deadline 25s) → stop relay/schedulers → close pool → flush logs.

## 4. Hook / interceptor system (`kernel/hooks`)

Purpose: let modules attach *cross-cutting side behavior* without kernel edits. Not a plugin bus for
business logic — business reactions belong in **event handlers**; hooks are synchronous, in-flow.

<!-- doc-example: illustrative -->
```go
type Registry struct{ … }
// Typed registration points (closed set — adding one is a kernel change, deliberately):
func OnRequest(phase Phase, h func(ctx, *RequestInfo) error)          // before|after HTTP
func OnCommand(phase Phase, h func(ctx, CommandInfo) error)           // before|success|failure
func OnWorkflowTransition(phase Phase, h func(ctx, TransitionInfo) error)
func OnRuleActivation(phase Phase, h func(ctx, RuleChangeInfo) error)
func OnDocumentAccess(phase Phase, h func(ctx, DocAccessInfo) error)  // watermark, extra checks
func OnFileUpload(phase Phase, h func(ctx, UploadInfo) error)         // malware scan slot
func OnAuditWrite(Before, h func(ctx, *audit.Entry) error)            // enrich only; cannot veto
func OnEventPublished(After, h func(ctx, outbox.Event))               // observe only
func OnJob(phase Phase, h func(ctx, JobInfo) error)
```

**Failure semantics are explicit per point (this is where hook systems rot, so it's normative):**
- `Before*` hooks on security-relevant points (upload, document access) may **veto** — error aborts the operation.
- `Before request/command` hooks may veto only with `errors.Kind` set; anything else = 500 + hook disabled-alarm metric.
- `After*`/observe hooks can never fail the operation: errors are logged + counted, never propagated.
- Ordering: registration order within a module; modules in dependency order; each hook runs with a
  100ms soft budget (exceeded → warn metric). No hook may start a goroutine (lint).
- Observability: every hook wrapped with otel span `hook.<point>.<module>` + duration/error metrics.
- Testing: testkit `HookRecorder` registers at every point and asserts invocation.

Dangerous and therefore disallowed: hooks that mutate command payloads (hidden logic), hooks on hot
read paths beyond request-level, hook-to-hook dependencies. If a "hook" starts making business
decisions, it must become an event handler or an explicit service step.

## Workflow callback context contract (V2)

Auto actions (`workflow.AutoInput.Context`) and assignee resolvers
(`workflow.ResolveInput.Context`) receive a **deep canonical copy** of the
instance context — mutating or retaining it never affects the framework, and
the returned output map is the only supported mutation channel. Values are
canonical JSON shapes: every **number is a `json.Number`** (never `float64`
or `int`). Assert numerics accordingly:

```go-illustrative
n, ok := in.Context["amount"].(json.Number) // NOT .(float64)
amount, err := n.Float64()
```

This is what makes gateway routing exact and identical before and after an
instance reloads from the database (see `docs/reference/invariant-ledger.md`).
