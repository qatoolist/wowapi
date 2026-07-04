# Building & Extending Modules

A **module** is the unit of feature work in wowapi. You implement a small interface; the framework composes
your module with the kernel and other modules at boot. This page walks the real
[`requests`](../../internal/testmodules/requests) fixture module end to end — it's the canonical example the
framework's own contract suite runs against, so every snippet here is real, compiling code.

## The `Module` interface

Every module implements three methods (`module/module.go`):

```go
type Module interface {
    Name() string          // unique id: ^[a-z][a-z0-9_]{0,63}$  — prefixes perms, events, migrations
    DependsOn() []string   // module names this one needs; app topo-sorts, cycles fail boot
    Register(ctx Context) error  // wire routes/migrations/perms/… — NO I/O, NO business logic
}
```

`Register` must **only wire**. All runtime behavior lives in handlers/jobs invoked later. Registration
errors (bad config key, unsatisfied port, unregistered permission) surface **at boot**, not at request time.

## `module.Context` — the capability surface

`Register` receives a `Context`: capability-scoped accessors, never raw pools or global config. The full
set (`module/module.go`):

| Accessor | Returns | Use it for |
|---|---|---|
| `Logger()` | `*slog.Logger` | Module-tagged structured logging. |
| `Config()` | `config.ModuleView` | `.Decode(&cfg)` your `modules.<name>.*` namespace. |
| `Routes()` | `*httpx.Router` | Register HTTP routes with `RouteMeta`. |
| `Validator()` | `*validation.Validator` | Shared validator for `httpx.BindAndValidate`. |
| `Permissions()` | `*authz.Registry` | Declare permissions (unregistered = never authorizable). |
| `Resources()` | `*resource.Registry` | Declare resource types for record-level authz. |
| `Authz()` | `authz.Evaluator` | Fine-grained, resource-scoped checks inside handlers. |
| `Tx()` | `database.TxManager` | The **only** door to the DB — tenant-scoped. |
| `IDGen()` | `model.IDGen` | UUIDv7 id generation (injectable for tests). |
| `Migrations(fs)` / `Seeds(fs)` / `OpenAPI(b)` | — | Register embedded SQL / seed catalog / OpenAPI fragment. |
| `Health(name, fn)` | — | Named readiness check surfaced on `/readyz`. |
| `ProvidePort(name, impl)` / `Port(name)` | — / `(any, error)` | Inter-module contracts (the **only** way modules talk). |
| `Events()` / `Outbox()` | registries | Subscribe handlers / emit events in a business tx. |
| `Jobs()` | `*jobs.Registry` | Register job kinds + retry policy. |
| `Rules()` / `RulesResolver()` | rule registry/resolver | Configurable rule points + effective values. |
| `Workflows()` / `WorkflowRuntime()` | workflow registry/runtime | State machines. |
| `RetentionClasses()` | `*retention.Registry` | Dispose/export/erase callbacks (DSR). |
| `DocumentClasses()`/`DocumentHooks()`/`Documents()`/`Comments()`/`Attachments()` | document subsystem | Files, comments, attachments. |
| `NotifyTemplates()`/`Notify()`/`Webhooks()`/`IntegrationProviders()`/`Integrations()` | comms subsystem | Notifications, webhooks, integrations. |

> The Context grows one accessor per kernel capability. While wowapi is **v0**, widening this interface is
> an accepted breaking change (`module/module.go` package doc).

## Scaffold a module

```bash
wowapi new-module --name widgets       # → internal/modules/widgets/ (module.go, migrations/, seeds/, openapi.json)
```
Flags: `--name` (required), `--dir` (default `internal/modules`), `--force`.

Generate CRUD scaffolding for a resource:

```bash
wowapi gen crud --module internal/modules/widgets --resource widget --fields "title:string,count:int"
```
Flags: `--module` (required), `--resource` (required), `--fields` (comma-separated `name:type`), `--force`.

## Register it in the product

`wowapi init` generates `internal/wire/modules.go`; add your module to the slice:

```go
func Modules() []module.Module {
    return []module.Module{
        &widgets.Module{},
    }
}
```
Modules are registered in `DependsOn` topological order. `make build` re-links; the app **validates** the
graph at boot (unknown dependency or cycle → boot fails).

## Anatomy of a real module

Below is the actual `requests` fixture, annotated. It's a leaf module (no dependencies) exposing four
routes over one table.

### 1. The typed config namespace (`dto.go` / `module.go`)

```go
// Config is the module's namespaced config: modules.requests.*
type Config struct {
    SLAHours int `json:"sla_hours"`
}
```

### 2. `Register` — wiring only (`module.go`)

```go
func (m *Module) Name() string       { return "requests" }
func (m *Module) DependsOn() []string { return nil }

func (m *Module) Register(mc module.Context) error {
    cfg := Config{SLAHours: 48}                 // default first…
    if err := mc.Config().Decode(&cfg); err != nil {  // …overlay from modules.requests.*
        return err                              // unknown keys are rejected → boot fails
    }

    mc.Migrations(migrationsFS)                 // //go:embed migrations/*.sql
    mc.Seeds(seedsFS)                           // //go:embed seeds/...
    mc.OpenAPI(openapiFragment)                 // //go:embed openapi.json

    h := &Handlers{                             // inject capabilities into handlers
        tx:    mc.Tx(),
        authz: mc.Authz(),
        val:   mc.Validator(),
        idgen: mc.IDGen(),
    }

    r := mc.Routes()
    // Public route first: net/http 1.22 resolves /requests/healthz before /{id}.
    r.Handle("GET",  "/requests/healthz", httpx.RouteMeta{Public: true}, h.Healthz)
    r.Handle("POST", "/requests",         httpx.RouteMeta{Permission: "requests.request.create"}, h.Create)
    r.Handle("GET",  "/requests/{id}",    httpx.RouteMeta{Permission: "requests.request.read"},   h.Read)
    r.Handle("GET",  "/requests",         httpx.RouteMeta{Permission: "requests.request.list"},   h.List)

    mc.Health("db", func(_ context.Context) error { return nil })
    mc.ProvidePort("requests.Lookup", &lookupImpl{tx: mc.Tx()})   // consumable by other modules
    return nil
}
```

Every route declares a `RouteMeta`: either `Public: true` (explicit opt-out) or a `Permission`. There is no
third option — an un-annotated route is a boot error. See [Auth](auth.md).

### 3. The request DTO with validation tags (`dto.go`)

```go
type CreateRequest struct {
    Title string `json:"title" validate:"required"`
}
```

### 4. A handler — decode, validate, transact, respond (`handlers.go`)

```go
type Handlers struct {
    tx    database.TxManager
    authz authz.Evaluator      // for fine-grained, resource-scoped checks (route-level authz already ran)
    val   *validation.Validator
    idgen model.IDGen
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    req, err := httpx.BindAndValidate[CreateRequest](r, h.val, 64*1024)  // strict JSON, size-capped, validated
    if err != nil {
        httpx.WriteError(ctx, w, err)      // → problem+json, correct status
        return
    }
    id := h.idgen.New()
    if err := h.tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
        if _, err := db.Exec(ctx,
            `INSERT INTO requests_request (id, tenant_id, title, status, version, created_at, created_by)
             VALUES ($1, app_tenant_id(), $2, 'open', 1, now(), $3)`,
            id, req.Title, uuid.Nil); err != nil {
            return err
        }
        return resource.NewRegistrar().Bind(db).Upsert(ctx,
            resource.Ref{Type: "requests.request", ID: id}, nil, req.Title, "open")
    }); err != nil {
        httpx.WriteError(ctx, w, err)
        return
    }
    httpx.WriteJSON(w, http.StatusCreated, httpx.OK(RequestDTO{ID: id, Title: req.Title, Status: "open"}))
}
```

Note the discipline visible here:

- **Tenant comes from `ctx`**, and the row is written with `tenant_id = app_tenant_id()` — RLS does
  isolation; you never pass a tenant filter by hand.
- Read paths use `WithTenantRO` (a `BEGIN READ ONLY` tx) — see the `Read`/`List` handlers.
- Not-found is an explicit domain error: `kerr.E(kerr.KindNotFound, "not_found", "request not found")`.
- Responses go through `httpx.WriteJSON(w, status, httpx.OK(dto))`; errors through `httpx.WriteError`.
- **Domain types never serialize directly** — only the `RequestDTO` wire shape does.

## Inter-module communication: ports

Modules must not import each other. To expose functionality, **provide a port**; to consume one, **fetch it**:

```go
// provider module, in Register:
mc.ProvidePort("requests.Lookup", &lookupImpl{tx: mc.Tx()})

// consumer module, in Register (declare the dependency in DependsOn):
p, err := mc.Port("requests.Lookup")
if err != nil { return err }          // unsatisfied port fails boot
lookup := p.(requestsapi.Lookup)      // assert to the shared interface type
```

Both sides are checked at boot — an unsatisfied `Port` fails `Validate`. This keeps the dependency graph
explicit and `make lint-boundaries` green.

## Checklist for a new module

- [ ] `Name()` matches `^[a-z][a-z0-9_]{0,63}$` and prefixes your permissions/events/tables.
- [ ] `DependsOn()` lists every module whose port you consume.
- [ ] `Register` **only wires** — no DB calls, no HTTP, no business logic.
- [ ] Every route has a `RouteMeta` (`Public` or `Permission`); every `Permission` is declared in
      `Permissions()`.
- [ ] Migrations are RLS-forced with a tenant policy (see [Database & migrations](database-migrations.md)).
- [ ] DTOs carry `validate:"…"` tags; handlers use `httpx.BindAndValidate`.
- [ ] All DB work goes through `mc.Tx()` (`WithTenant`/`WithTenantRO`).
- [ ] Registered in the product's `internal/wire/modules.go`.
- [ ] Covered by a `testkit.RunModuleContract` test — see [Testing](testing.md).

Next: [Database & migrations](database-migrations.md) · [Auth](auth.md) · [Validation & errors](validation-errors.md).
