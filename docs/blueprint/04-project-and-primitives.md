# 04 — Go Project Structure, Base Model / DTO / Error / Validation Primitives

## 1. Repository layout

wowapi is a **consumable Go module** ([11-framework-distribution-and-consumption.md](11-framework-distribution-and-consumption.md)):
everything a product module must import is a public package; `internal/` holds only private
implementation guts. (The earlier draft's all-`internal` layout is superseded — Go forbids external
modules from importing `internal/` packages, so consumer-facing contracts cannot live there.)

### Framework repository — `github.com/qatoolist/wowapi/v2`

```text
/kernel/...         # PUBLIC L1: primitives + service contracts (package map below)
/module             # PUBLIC L2: Module, Context, registries, lifecycle, seed schema
/app                # PUBLIC composition helpers: app.New, RunAPI, RunWorker, RunMigrate
/adapters/...       # PUBLIC L0: postgres/, s3/, smtp/, smsprovider/, oidc/, secrets/, scanner/
/testkit            # PUBLIC test fixtures, fakes, assertions, module contract suite
/migrations         # kernel goose migrations, exposed as embed.FS (migrations.Kernel())
/cmd/wowapi         # installable CLI (scaffolds, generators, seed validate, openapi merge, lint)
/internal/...       # PRIVATE impl guts: pg stores, engine internals, outbox relay, evaluator impl —
                    #   wired by /app; never a consumer-facing contract
/internal/testmodules/requests  # private neutral module fixture for contract tests
/examples/acme-ops  # optional standalone sample product app, with its own go.mod; non-contractual
/api/openapi        # base spec fragments the kernel contributes
/configs /deployments /scripts   # framework dev/test infra (compose: pg+minio+mailpit)
/docs               # this blueprint, ADRs (docs/adr/NNNN-*.md)
```

No `/pkg` wrapper directory: public packages at the repo root are the idiomatic shape for a
consumable framework (cf. chi, river). Shared micro-helpers fold into the relevant `kernel/*`
package or stay in `internal/` — no grab-bag `shared` package.

### Product application repository (e.g. `example.com/acme-ops`)

```text
go.mod              # require github.com/qatoolist/wowapi/v2 vX.Y.Z
/cmd/api            # thin main: config → app.RunAPI(cfg, modules…)
/cmd/worker         # app.RunWorker — outbox relay + job runner + schedulers
/cmd/migrate        # app.RunMigrate — kernel migrations (from wowapi) + embedded module migrations
/internal/modules   # L3 product modules (requests/, assets/, … society/ in its own product repo)
/api/openapi        # merged output (wowapi openapi merge)
/configs /deployments /scripts
```

Import law (enforced by `wowapi lint boundaries` + the Go compiler):

- Product modules → `wowapi/module` → `wowapi/kernel/*`; product modules may also import selected
  `wowapi/kernel/*` helpers directly.
- `wowapi/kernel/*` defines contracts and primitives; it must not import `wowapi/module`,
  `wowapi/app`, `wowapi/adapters`, `wowapi/testkit`, examples, or product code.
- `wowapi/module` may import `wowapi/kernel/*` contracts; it must not import `wowapi/app`,
  `wowapi/adapters`, examples, or product code.
- `wowapi/adapters/*` implement kernel ports and may import `wowapi/kernel/*`; they must not import
  `wowapi/module` or `wowapi/app`.
- `wowapi/app` is the composition root and may import `kernel`, `module`, `adapters`, and
  `migrations`; nothing in `kernel`, `module`, or `adapters` imports `app`.
- `wowapi/testkit` is test support and may compose `app`, `module`, `kernel`, adapters, and private
  test modules; production packages must not import it.
- `wowapi/internal/...` is compiler-blocked outside the framework repo; product modules import each
  other only via declared ports.
- `examples/*` are standalone sample apps or docs fixtures, preferably separate nested Go modules;
  they are non-contractual and never imported by `kernel`, `module`, `app`, or `adapters`.

This direction keeps the public package graph acyclic: `kernel` at the base, `module` above it,
`adapters` beside it implementing ports, and `app` at the top wiring everything together.

## 2. Kernel package map

All paths below are **public** packages under `github.com/qatoolist/wowapi/v2/` — importable by
product modules. Several heavyweight packages have been refactored into a `foundation/` layer with
thin v1 API stability shims in `kernel/` for backward compatibility (e.g., `kernel/webhook` re-exports
`foundation/webhook`). Where a package has implementation in `internal/`, it is wired by `app` and not
directly public to product modules.

| Package | Responsibility / key exports | Must not import | Modules may import |
|---|---|---|---|
| `kernel/model` | base embedded structs (below), `ID`, `Money`, `TimeRange`, `Ref` types | anything but stdlib+uuid | ✅ |
| `kernel/database` | `TxManager`, `TenantDB`, RLS helpers, `IdemStore`, batch helpers, tenant context binding via `SET LOCAL app.tenant_id` | modules | ✅ (read ctx) |
| `kernel/auth` | OIDC/JWT verification middleware, `Principal` | authz | ✅ (read ctx) |
| `kernel/authz` | `Evaluator`, `Actor`, `Decision`, assignment store, capacity resolution | modules | ✅ (interfaces) |
| `kernel/policy` | condition model + evaluation (used by authz) | http | ✅ read-only |
| `kernel/resource` | `Ref`, type registry, `Registrar` (upsert mirror rows) | modules | ✅ |
| `kernel/relationship` | edge store, `Checker` | modules | ✅ |
| `kernel/workflow` | runtime, registry, definitions, sweeper job | notify directly (emits events instead) | ✅ |
| `kernel/rules` | point registry, resolver, versions, flags sugar | workflow internals (calls service iface) | ✅ |
| `kernel/audit` | `Writer` iface + pg impl, audit middleware helpers | — | ✅ (Writer) |
| `kernel/outbox` | `Writer`, relay, dispatcher, `processed_events` inbox helper | — | ✅ (Writer) |
| `kernel/jobs` | `Runner`, `Registry`, worker pool wrapper over River, retry/backoff/DLQ | — | ✅ (register kinds) |
| `foundation/document` | document/file service, presign, scan hooks, grants (via `kernel/document` compat shim) | — | ✅ |
| `foundation/notify` | templates, dispatcher, channel adapters iface, preferences (via `kernel/notify` compat shim) | — | ✅ (send API) |
| `foundation/webhook` | inbound verify/ingest, outbound deliver, replay protection (via `kernel/webhook` compat shim) | — | ✅ |
| `foundation/integration` | provider registry, credential refs, circuit breaker (via `kernel/integration` compat shim) | — | ✅ |
| `kernel/httpx` | handler helpers, middleware chain, route metadata, server | modules | ✅ |
| `kernel/errors` | error taxonomy, codes, wrapping, HTTP mapping | http types beyond status codes | ✅ |
| `kernel/validation` | validator wrapper, field errors | — | ✅ |
| `kernel/pagination` | page/cursor types, keyset encoding | — | ✅ |
| `kernel/filtering` | allowlist filter/sort builders | — | ✅ |
| `kernel/secrets` | `Provider` port (env/cloud managers), `Ref` (`secretref://<provider>/<path>`) parsing/validation | everything (graph base: stdlib only) | ✅ (types; resolution happens at boot in `app`) |
| `kernel/config` | typed `Framework` config structs, layered loader, precedence, `Secret` redaction, `ModuleView` — see [12](12-configuration-and-deployment.md) | everything except `kernel/secrets` (near-base of the graph) | types only — *values* reach modules solely via `module.Context.Config()` |
| `kernel/logging` / `kernel/observability` | slog setup, otel, metrics, health registry | — | ✅ |
| `kernel/seeds` | seed schema parsing + sync engine | modules | via module SDK |
| `kernel/apikey` | API key provisioning, validation, revocation | — | ✅ |
| `kernel/appmodel` | request reference declarations, domain model bindings | — | ✅ |
| `kernel/httpclient` | SSRF-safe HTTP client, dial guard, allowlist escapes | — | ✅ |
| `kernel/i18n` | localization support | — | ✅ |
| `kernel/lease` | distributed lease/fencing mechanism | — | ✅ |
| `kernel/lifecycle` | module/provider lifecycle hooks | — | ✅ |
| `foundation/mfa` | multi-factor auth (TOTP enrolment/verification) (via `kernel/mfa` compat shim) | — | ✅ |
| `foundation/artifact` | immutable versioned artifacts (via `kernel/artifact` compat shim) | — | ✅ |
| `foundation/attachment` | attachment lifecycle over object storage (via `kernel/attachment` compat shim) | — | ✅ |
| `foundation/bulk` | bulk-operation framework (via `kernel/bulk` compat shim) | — | ✅ |
| `foundation/comment` | generic comment threads (via `kernel/comment` compat shim) | — | ✅ |
| `kernel/migration` | database migration runner, goose integration | — | ✅ |
| `kernel/port` | inter-module port registry (declared service boundaries) | — | via module SDK (`module.Context` port lookup) |
| `kernel/privileged` | privileged operation markers (audit, taints) | — | ✅ |
| `kernel/retry` | retry/backoff strategies | — | ✅ |
| `kernel/safety` | atomic/idempotent write guards | — | ✅ |
| `kernel/sequence` | sequential ID generation | — | ✅ |
| `kernel/storage` | object store port + presign API | — | ✅ |
| `kernel/tracing` | distributed tracing setup (OpenTelemetry) | — | ✅ |

## 3. Base model primitives (`kernel/model`) — composition, no god BaseModel

<!-- doc-example: illustrative -->
```go
// BaseFields: identity only. Embed in every persisted entity.
type BaseFields struct {
    ID uuid.UUID `db:"id"`
}

// TenantScoped: embed in every tenant-owned entity. Presence of this struct is what
// the repo helpers key on — an entity without it cannot use TenantDB write helpers.
type TenantScoped struct {
    TenantID uuid.UUID `db:"tenant_id"`
}

// Auditable: who/when. Embed in mutable entities. NOT on append-only rows (they use CreatedOnly).
type Auditable struct {
    CreatedAt time.Time  `db:"created_at"`
    CreatedBy uuid.UUID  `db:"created_by"`
    UpdatedAt *time.Time `db:"updated_at"`
    UpdatedBy *uuid.UUID `db:"updated_by"`
}
type CreatedOnly struct {
    CreatedAt time.Time `db:"created_at"`
    CreatedBy uuid.UUID `db:"created_by"`
}

// Versioned: optimistic locking. Embed in user-editable aggregates.
// Anti-pattern: embedding on append-only tables or high-frequency counters.
type Versioned struct {
    Version int `db:"version"`
}

// Temporal: validity window. Embed where history matters (assignments, relationships, grants).
type Temporal struct {
    ValidFrom time.Time  `db:"valid_from"`
    ValidTo   *time.Time `db:"valid_to"`
}
func (t Temporal) ActiveAt(at time.Time) bool { … }

// Statused: lifecycle instead of soft-delete booleans. Status vocab is per-entity (typed const).
type Statused[S ~string] struct {
    Status S `db:"status"`
}

// Refs — value objects, not embedded rows:
type ResourceRef struct { Type string; ID uuid.UUID }      // kernel-wide pointer to any domain object
type ActorRef    struct { Kind ActorKind; UserID, CapacityID uuid.UUID; System string }
type Money       struct { Amount decimal.Decimal; Currency string } // shopspring/decimal; DB: numeric+char(3)
type TimeRange   struct { From time.Time; To *time.Time }
type Metadata    map[string]any   // jsonb extension bag — for module-declared extras ONLY, never core logic
type ExternalRef struct { System, ID string }              // pointer into an external system
```

**Usage rules**
- Embed `BaseFields + TenantScoped + Auditable + Versioned + Statused` in a typical mutable aggregate.
  Append-only rows: `BaseFields + TenantScoped + CreatedOnly`.
- `Temporal` only where "as-of" queries are real requirements — don't sprinkle it (every temporal
  table pays query complexity forever).
- `Ownership` generic ownership is NOT a struct — it's a relationship (`core.owner_of`). Making it a
  column invites domain semantics into the kernel.
- `Metadata` is a pressure valve, not a schema strategy: modules may stash display extras; the moment
  code branches on a metadata key, promote it to a column.
- **Anti-pattern avoided:** one `BaseModel` with 15 fields forces every table to carry columns it
  doesn't need and every query to explain them. Composition keeps each entity honest and explicit,
  with ~4 short embed lines instead of copied fields.

## 4. DTO & API response primitives (`kernel/httpx` + `kernel/pagination`)

<!-- doc-example: illustrative -->
```go
// Success envelope. Data is the resource DTO; Meta optional.
type APIResponse[T any] struct {
    Data T          `json:"data"`
    Meta *Meta      `json:"meta,omitempty"`
}
type Meta struct {
    RequestID string     `json:"request_id"`
    Audit     *AuditMeta `json:"audit,omitempty"`
}
type AuditMeta struct {
    CreatedAt time.Time  `json:"created_at"`
    CreatedBy uuid.UUID  `json:"created_by"`
    UpdatedAt *time.Time `json:"updated_at,omitempty"`
    Version   int        `json:"version"` // clients echo via If-Match
}

// Offset page (admin/small lists) and cursor page (default for feeds/large lists).
type PageResponse[T any] struct {
    Items      []T   `json:"items"`
    Page       int   `json:"page"`
    PerPage    int   `json:"per_page"`
    TotalCount int64 `json:"total_count,omitempty"` // omitted when COUNT is too expensive
}
type CursorPage[T any] struct {
    Items      []T    `json:"items"`
    NextCursor string `json:"next_cursor,omitempty"` // opaque base64(keyset tuple)
    HasMore    bool   `json:"has_more"`
}

// RFC 9457 problem details — the ONLY error body shape the API emits.
type ProblemError struct {
    Type     string       `json:"type"`               // "https://errors.<platform>/validation"
    Title    string       `json:"title"`              // short, safe
    Status   int          `json:"status"`
    Detail   string       `json:"detail,omitempty"`   // safe for users; never internals
    Instance string       `json:"instance,omitempty"` // request path
    Code     string       `json:"code"`               // machine code: "validation_failed"
    RequestID string      `json:"request_id"`
    Errors   []FieldError `json:"errors,omitempty"`   // validation only
}
type FieldError struct {
    Field   string `json:"field"`    // JSON path: "contacts[0].email"
    Code    string `json:"code"`     // "required", "max_length", "invalid_format"
    Message string `json:"message"`
}

// Long-running / bulk / upload / webhook shapes.
type OperationResponse struct {      // 202 Accepted + Location: /v1/operations/{id}
    OperationID uuid.UUID `json:"operation_id"`
    Status      string    `json:"status"`   // pending|running|succeeded|failed
    Progress    *Progress `json:"progress,omitempty"`
    Result      any       `json:"result,omitempty"`
}
type BulkResponse struct {
    Succeeded int          `json:"succeeded"`
    Failed    int          `json:"failed"`
    Errors    []BulkError  `json:"errors,omitempty"` // index + ProblemError
}
type UploadSessionResponse struct {
    UploadID  uuid.UUID `json:"upload_id"`
    URL       string    `json:"url"`        // presigned PUT
    Headers   map[string]string `json:"headers"`
    ExpiresAt time.Time `json:"expires_at"`
}
type WebhookAck struct { Received bool `json:"received"`; EventID string `json:"event_id"` }
```

Handlers never build ad-hoc maps; they return DTO structs through `WriteJSON`/`WriteError` (see
[05-http-and-persistence.md](05-http-and-persistence.md)). DTOs live in module `api/dto.go`,
mapped from domain structs in `api/mapper.go` — domain types never serialize directly (protects
internal fields, decouples DB shape from wire shape).

## 5. Error & validation framework (`kernel/errors`, `kernel/validation`)

### Taxonomy → HTTP mapping (closed set)

| Kind (Go) | Code | HTTP | Notes |
|---|---|---|---|
| `KindValidation` | `validation_failed` | 400 | with `FieldError`s |
| `KindUnauthenticated` | `unauthenticated` | 401 | missing/invalid token |
| `KindForbidden` | `permission_denied` | 403 | audited when sensitive |
| `KindTenantIsolation` | `tenant_mismatch` | 404* | *masked as not-found: don't leak existence |
| `KindNotFound` | `not_found` | 404 | |
| `KindConflict` | `conflict` | 409 | uniqueness, idempotency-hash mismatch, workflow state |
| `KindVersionConflict` | `version_conflict` | 412 | optimistic lock / If-Match |
| `KindIdempotencyInFlight` | `retry_later` | 409 | same key still processing |
| `KindRuleViolation` | `rule_violation` | 422 | rule engine rejected value |
| `KindWorkflowState` | `invalid_transition` | 409 | |
| `KindRateLimited` | `rate_limited` | 429 | Retry-After |
| `KindExternal` | `upstream_error` | 502 | provider failures (circuit breaker) |
| `KindInternal` | `internal` | 500 | panic recovery included; detail NEVER exposed |

<!-- doc-example: illustrative -->
```go
type Error struct {
    Kind    Kind
    Code    string            // stable machine code
    Msg     string            // safe user-facing
    Op      string            // "requests.Service.Approve" — for logs only
    Fields  []FieldError
    Err     error             // wrapped cause (%w)
}
func E(kind Kind, code, msg string, args ...any) *Error
func (e *Error) Error() string; func (e *Error) Unwrap() error
// Handlers: kernelhttp.WriteError inspects errors.As(*Error) → ProblemError; unknown → 500 internal.
```

**Wrapping convention:** every layer wraps with `fmt.Errorf("op: %w", err)` or `errors.E`; the HTTP
layer logs the full chain at ERROR (5xx) / WARN (4xx sensitive) / DEBUG (routine 4xx) with request
id + tenant + actor; the client sees only `Msg`/`Code`. Panics: recover middleware → 500 + stack to
logs + `panic_total` metric — never a stack trace on the wire.

### Validation boundary
- **Shape validation** (required, formats, ranges): struct tags via wrapped `go-playground/validator`,
  invoked by `BindAndValidate` in the handler. Produces `FieldError`s with JSON paths.
- **Domain validation** (state rules, cross-field, rule-engine-driven): explicit funcs in module
  `domain/validation.go`, called by the service, returning `errors.E(KindValidation|KindRuleViolation…)`.
- Never duplicate: handlers do shape only; services do rules only; repos do neither (constraints are
  the DB's backstop, mapped to `KindConflict`).
