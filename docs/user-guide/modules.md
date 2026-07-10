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
| `NotifyTemplates()`/`Notify()`/`Webhooks()`/`IntegrationProviders()`/`Integrations()` | comms subsystem | Notifications, webhooks (outbound delivery is SSRF-safe by default — see [Webhooks](webhooks.md)), integrations. |
| `Privileged()` | `*privileged.Services` | Scoped, audited platform-privilege operations — grant/revoke ReBAC edges, activate tenant rules — on keys the module **owns**. See below. |

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

## Privileged services — platform operations without a `SECURITY DEFINER` bridge

A few framework writes are deliberately **off-limits** to the shared `app_rt` role your module runs as,
because they are authorization inputs:

- **`relationships`** — a `granted_via` edge grants a permission on its object, so `app_rt` has `SELECT`
  only; edge writes are `app_platform`.
- **`rule_versions`** — *activating* a rule version changes runtime behaviour, so `app_rt` may only
  propose drafts; activation is `app_platform`.

When your module has a *valid* reason to perform one of these (grant a committee seat, activate a
tenant policy version), do **not** ship your own `SECURITY DEFINER` SQL function. Use `mc.Privileged()`:
a scoped, audited surface that runs the operation with platform privilege **but tenant-bound**, so the
same RLS and constraints still isolate you.

```go
// Grant a ReBAC edge of a type THIS module owns (prefix "committee.").
id, err := mc.Privileged().Relationships().Grant(ctx, privileged.GrantSpec{
    RelType:     "committee.seat_of",             // must be owned (module prefix or allow-list)
    SubjectKind: relationship.KindCapacity,        // defaults to capacity when empty
    SubjectID:   capacityID,                       // must be an active capacity in this tenant
    Object:      resource.Ref{Type: "committee.committee", ID: committeeID}, // must exist in this tenant
    ValidTo:     &expiry,                           // optional temporal window; nil = open-ended
    Actor:       actorID,                           // recorded as created_by + in the audit trail
})

// Revoke soft-closes the edge (valid_to = now) — history is preserved, not deleted.
err = mc.Privileged().Relationships().Revoke(ctx, id, actorID)

// Activate a TENANT-scope draft of a rule KEY this module owns.
err = mc.Privileged().Rules().ActivateTenant(ctx, versionID, approverID, privileged.ActivateOptions{
    // Optional: run a domain gate atomically before the state transition (e.g. "a verified
    // citation must cover the effective date"). Returning an error rolls the activation back.
    Gate: func(ctx context.Context, db database.TenantDB) error { return checkCitation(ctx, db) },
})
```

What the framework enforces for you (so you never re-implement it in SQL):

- **Tenant binding** — fails closed if `ctx` carries no tenant.
- **Ownership** — the relationship type / rule key must be prefixed with your module name, or be in
  the product's declared allow-list. You can never manage another module's edges or rules.
- **Existence** — the subject capacity and object resource must exist in *your* tenant (cross-tenant
  targets are invisible via RLS → rejected).
- **Scope** — only *tenant-scope* rule versions of the caller's tenant can be activated here;
  platform-scope activation stays platform-tooling-only.
- **Audit** — every grant/revoke/activation writes a row on the tenant's audit hash chain.
- **Races** — a double-revoke or double-activate is reported as a conflict; concurrent activations are
  arbitrated by the one-active-per-instant constraint, exactly as before.

> Availability: `mc.Privileged()` requires a process wired with the `app_platform` pool (`api`/`worker`;
> not the `migrate` process). The default surface is **prefix-ownership**; a product that must widen it
> (e.g. let a module manage a kernel `core.` type) constructs its own `privileged.Services` with a
> `Config` allow-list from its wiring.

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
