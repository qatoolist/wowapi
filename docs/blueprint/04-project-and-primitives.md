# 04 — Go Project Structure, Base Model / DTO / Error / Validation Primitives

## 1. Repository layout

```text
/cmd/api            # HTTP server binary: flags/env → kernel.New → app.Register(modules…) → serve
/cmd/worker         # outbox relay + job runner + schedulers; same composition root, no HTTP
/cmd/migrate        # goose runner (kernel + module migrations) + seed sync; only binary with app_migrate creds
/internal/kernel    # L1 platform kernel (package map below)
/internal/platform  # L2 extension layer: module SDK (Module, ModuleContext, registries, seed schema)
/internal/modules   # L3 product modules (requests/, assets/, … society/ later)
/internal/adapters  # L0: postgres/, s3/, smtp/, smsprovider/, oidc/, secrets/, scanner/
/internal/shared    # tiny pure helpers (ptr, slices, strcase). No kernel imports. Keep <500 LOC total.
/internal/testkit   # test fixtures, fakes, assertions (imports kernel; test-only)
/pkg                # ONLY code intended for external consumption (e.g. Go API client). Default: empty.
/migrations         # kernel goose migrations (modules embed their own)
/api/openapi        # base spec + merged output of module fragments
/configs            # config.example.env, per-env compose overrides. No secrets committed.
/deployments        # Dockerfile(s), compose.yaml, deploy manifests
/scripts            # dev scripts (db-reset, lint-boundaries)
/docs               # this blueprint, ADRs (docs/adr/NNNN-*.md)
/tools              # CLI: module generator, openapi merge, seed validate (module `tools/` with its own go.mod)
```

Import law (enforced by `make lint-boundaries`, e.g. go-arch-lint or a 40-line script):
`modules → platform → kernel → (adapter interfaces)`; `adapters → kernel` (implement kernel ports);
nothing imports `modules` except `cmd/*`; `shared` imports stdlib only; `testkit` never imported by
non-test code.

## 2. Kernel package map

| Package | Responsibility / key exports | Must not import | Modules may import |
|---|---|---|---|
| `kernel/model` | base embedded structs (below), `ID`, `Money`, `TimeRange`, `Ref` types | anything but stdlib+uuid | ✅ |
| `kernel/tenant` | `Context`, `FromContext`, resolver middleware, tenant service | modules | ✅ (read ctx) |
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
| `kernel/document` | document/file service, presign, scan hooks, grants | — | ✅ |
| `kernel/notify` | templates, dispatcher, channel adapters iface, preferences | — | ✅ (send API) |
| `kernel/webhook` | inbound verify/ingest, outbound deliver, replay protection | — | ✅ |
| `kernel/integration` | provider registry, credential refs, circuit breaker | — | ✅ |
| `kernel/httpx` | handler helpers, middleware chain, route metadata, server | modules | ✅ |
| `kernel/errors` | error taxonomy, codes, wrapping, HTTP mapping | http types beyond status codes | ✅ |
| `kernel/validation` | validator wrapper, field errors | — | ✅ |
| `kernel/pagination` | page/cursor types, keyset encoding | — | ✅ |
| `kernel/filtering` | allowlist filter/sort builders | — | ✅ |
| `kernel/database` | `TxManager`, `TenantDB`, RLS helpers, `IdemStore`, batch helpers | modules | ✅ |
| `kernel/config` | typed config structs, env loading | everything else | cmd only |
| `kernel/logging` / `kernel/observability` | slog setup, otel, metrics, health registry | — | ✅ |
| `kernel/seeds` | seed schema parsing + sync engine | modules | via platform |

## 3. Base model primitives (`kernel/model`) — composition, no god BaseModel

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
