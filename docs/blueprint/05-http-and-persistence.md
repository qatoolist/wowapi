# 05 — Handler Helpers, Repository/Transaction Helpers, Service Conventions, CRUD Scaffolding

## 1. Handler pattern (`kernel/httpx`)

No generic controller. Handlers are small explicit funcs on a struct holding the module's service +
kernel helpers. The kernel provides *helpers*, the module keeps the *flow* visible.

```go
// kernel/httpx — the toolbox (signatures)
func DecodeJSON[T any](r *http.Request, maxBytes int64) (T, error)          // strict: unknown fields rejected
func BindAndValidate[T any](r *http.Request, v *validation.Validator) (T, error)
func WriteJSON[T any](w http.ResponseWriter, status int, body T)
func WriteError(ctx context.Context, w http.ResponseWriter, err error)     // → ProblemError
func ParseResourceID(r *http.Request, param string) (uuid.UUID, error)
func ParsePagination(r *http.Request, def page.Defaults) (page.Cursor, error)
func ParseFilters(r *http.Request, allow filtering.Allowlist) (filtering.Set, error)
func ParseSort(r *http.Request, allow filtering.SortAllowlist) (filtering.Sort, error)
func ETagFrom(version int) string; func RequireIfMatch(r *http.Request) (int, error)

// Route registration REQUIRES metadata — there is no method without it:
type RouteMeta struct {
    Permission string        // "" only if Public
    Public     bool          // explicit opt-out (health, webhooks pre-verify)
    Scope      ScopeExtractor // how to derive authz.Target from the request (org/resource id)
    Idempotent bool          // enables WithIdempotency for POST
    Sensitive  bool          // forces audit even on read
}
func (r *Router) Handle(method, pattern string, meta RouteMeta, h http.HandlerFunc)
// startup validation: meta.Permission unknown → boot failure; Public+Permission set → boot failure.
```

### Canonical module handler

```go
type Handlers struct {
    svc  *app.Service            // module service
    tx   database.TxManager
    val  *validation.Validator
}

func (h *Handlers) CreateRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()                                   // tenant+actor already resolved+authorized by middleware
    cmd, err := httpx.BindAndValidate[dto.CreateRequest](r, h.val)
    if err != nil { httpx.WriteError(ctx, w, err); return }

    var res domain.Request
    err = httpx.WithIdempotency(ctx, h.tx, r, func(ctx context.Context, db database.TenantDB) error {
        var err error
        res, err = h.svc.Create(ctx, db, cmd.ToCommand())  // service does authz-detail, rules, audit, outbox
        return err
    })
    if err != nil { httpx.WriteError(ctx, w, err); return }
    httpx.WriteJSON(w, http.StatusCreated, httpx.OK(dto.FromDomain(res)))
}
```

Division of labor: middleware = authn, tenant, capacity, coarse authz, rate limit; handler = decode,
validate shape, tx/idempotency wrapper, map to DTO; service = everything business. Handlers stay
10–25 lines. `HandleFileUpload` / `HandleWebhook` are prebuilt handler factories in the kernel
(upload: presign flow + validation hooks; webhook: signature verify + replay check + store + enqueue).

## 2. Persistence helpers (`kernel/database`)

```go
// TxManager is the ONLY door to the database for tenant work.
type TxManager interface {
    // WithTenant: BEGIN; SET LOCAL app.tenant_id/app.actor_id from ctx (error if absent); fn; COMMIT.
    WithTenant(ctx context.Context, fn func(ctx context.Context, db TenantDB) error) error
    // WithTenantRO: read-only tx (BEGIN READ ONLY) — list/get paths.
    WithTenantRO(ctx context.Context, fn func(ctx context.Context, db TenantDB) error) error
    // Platform: global tables only; kernel services only; not exposed on ModuleContext.
    Platform(ctx context.Context, fn func(ctx context.Context, db DB) error) error
}

// TenantDB is a facade over pgx.Tx that module sqlc code receives. It cannot outlive the tx,
// cannot be constructed elsewhere, and carries the per-tx service bundle:
type TenantDB interface {
    DBTX                                  // sqlc's interface: Exec/Query/QueryRow (with tx timeout)
    Outbox() outbox.Writer                // same-tx event writes
    Audit() audit.Writer                  // same-tx audit writes
    Resources() resource.Registrar        // same-tx mirror upsert
}

type UnitOfWork = TxManager // one tx per request IS the unit of work; no separate change-tracker (that's an ORM concept Go doesn't need)

// Optimistic locking helper (used inside generated/handwritten queries):
func ExpectOneRow(tag pgconn.CommandTag, entity string) error // 0 rows → KindVersionConflict

// Idempotency:
type IdemStore interface {
    Begin(ctx context.Context, db TenantDB, key, requestHash string, ttl time.Duration) (Replay, error)
    Complete(ctx context.Context, db TenantDB, key string, status int, body []byte) error
}
// httpx.WithIdempotency composes: no header → plain tx; header → Begin (replay? write stored response)
// → tx(fn) → Complete in same tx.

// Pagination / filtering / sorting (allowlist-driven; SQL injection impossible by construction):
type Keyset struct { Col string; Dir Dir }                 // encoded into opaque cursor
func KeysetClause(ks []Keyset, cur page.Cursor) (sql string, args []any)
type Allowlist map[string]filtering.FieldSpec              // "status": {Col: "status", Ops: eq|in}
// FilterBuilder/SortBuilder only ever emit columns from the allowlist; values always as args.

// Misc: BatchInsert (COPY), QueryTimeout (per-query context deadline, default 5s, hot paths 2s),
// TemporalActive(alias, at) → "alias.valid_from <= $n AND (alias.valid_to IS NULL OR …)".
```

### Repository conventions (module `store/`)
- One repository struct per aggregate, constructed with nothing (methods take `TenantDB`): 
  `func (r RequestRepo) GetByID(ctx, db database.TenantDB, id uuid.UUID) (domain.Request, error)`.
  Passing `db` in keeps repos stateless and makes the tx explicit at call sites.
- SQL lives in `store/queries.sql` compiled by sqlc; handwritten SQL only for dynamic
  filter/sort composition, and then only via the allowlist builders.
- **Forbidden (lint/review):** SQL outside `store/`; joins to another module's tables; any use of the
  raw pool; UPDATE on append-only tables; UPDATE without `AND version = $n` on Versioned aggregates;
  string-concatenated SQL.

## 3. Service / use-case conventions (module `app/`)

```go
// commands.go / queries.go — plain structs, no behavior:
type CreateRequestCommand struct { OrgID uuid.UUID; Title string; Body string; … }
type ListRequestsQuery    struct { Filter filtering.Set; Page page.Cursor }

// ports.go — what this module needs from others (consumer-side interfaces):
type Notifier interface { Send(ctx context.Context, n notify.Message) error }
type AssetLookup interface { AssetExists(ctx context.Context, id uuid.UUID) (bool, error) } // another module's port

// service.go
type Service struct {
    repo   store.RequestRepo
    rules  rules.Resolver
    wf     workflow.Runtime
    authz  authz.Evaluator
    idgen  model.IDGen
    clock  model.Clock
}
func NewService(…explicit deps…) *Service
```

**Service method shape (normative order):**
1. fine-grained authz (record-level / relationship checks the middleware couldn't do),
2. rule resolution (`rules.ResolveAs(ctx, "requests.sla.duration", now, &sla)`) — record version ids used,
3. domain validation + mutation (domain funcs on the model),
4. persistence via repo (optimistic lock),
5. `db.Resources().Upsert(...)` if aggregate is kernel-visible,
6. `db.Audit().Action(...)`, `db.Outbox().Publish(...)` — same tx,
7. optionally `wf.Start(...)` — same tx.

Naming: `CreateXCommand/UpdateXCommand/ListXQuery/XResult`, `XService`, `XRepo`, ports named for the
*capability* (`Notifier`), not the provider. Context flows through everything; services check
`ctx.Err()` before expensive stages; no goroutines in services (async = outbox/jobs).

What goes where: **handler** decode/validate-shape/respond · **service** orchestration+rules+authz-detail ·
**domain** invariants & state transitions (`request.Approve(actor, now) error`) · **repo** SQL only ·
**event handler** cross-module reactions · **job** long/bulk/scheduled work.

## 4. Generic CRUD scaffolding

Mechanism: **code generation, not runtime genericity.** `tools/gen crud --module assets --resource asset`
emits the full vertical slice (handlers with RouteMeta, DTOs+mappers, service with audit/outbox/rules
stubs, sqlc queries incl. keyset list + optimistic update + status lifecycle, OpenAPI fragment,
seeds for `asset.create|read|list|update|deactivate|restore` permissions, table-driven tests).
Generated code is committed, reviewed, owned, and freely editable — the generator is a starting
point, not a framework layer (no lock-in: regeneration is optional and diff-able).

| Use generated CRUD | Hand-write instead |
|---|---|
| Reference/catalog data (categories, tags, locations) | Anything with money movement or balances |
| Simple registries (contacts, notes-like records) | Workflow-driven aggregates (state machines) |
| Admin config screens | Security objects (roles, assignments, grants) |
| First cut of any new simple resource | Audit/compliance surfaces |
| | Cross-aggregate transactions |

Why: financial/workflow/security logic *is* the product — hiding it behind a generic engine creates a
weak low-code framework that fights every real requirement (the Goal.md anti-pattern). A runtime
`GenericCRUDController[T]` is explicitly rejected: it centralizes behavior that must vary, and Go's
generics can't express per-resource authz/rules/workflow variation without reflection or config-blob
programming.
