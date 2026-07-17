# Validation & Error Handling

wowapi has one error model end to end: handlers return a typed `kernel/errors.Error`, and the HTTP layer
renders it as **RFC 9457 `application/problem+json`** with the right status. Validation failures are just
one error *kind* with structured field details. This page shows the taxonomy, the wire shape, and how to
produce each from a handler. (`kernel/errors/`, `kernel/httpx/`, `kernel/validation/`.)

## Request decoding + validation in one call

Handlers decode and validate a JSON body with a single generic helper:

```go
req, err := httpx.BindAndValidate[CreateRequest](r, h.val, 64*1024)
if err != nil {
    httpx.WriteError(ctx, w, err)   // → 400 problem+json with field errors
    return
}
```

`httpx.BindAndValidate[T](r, v, maxBytes)` (`kernel/httpx/decode.go`):

- Decodes JSON **strictly** — unknown fields are rejected, the body is size-capped at `maxBytes`, and only
  a single JSON value is accepted (trailing data is an error).
- Runs struct-tag validation on the decoded value.
- On failure returns a `*errors.Error` with `Kind = KindValidation` carrying per-field details.

Validation uses **`go-playground/validator/v10`** via `kernel/validation`. Get the shared validator from
`module.Context.Validator()` and tag your DTOs:

```go
type CreateRequest struct {
    Title string `json:"title" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Count int    `json:"count" validate:"gte=0,lte=100"`
}
```

Field errors report the **JSON path** (from the `json` tag) and a stable code:

| Tag | Field-error `code` |
|---|---|
| `required` | `required` |
| `email`, `uuid` | `invalid_format` |
| `min`, `gte` | `min` |
| `max`, `lte` | `max` |
| `len` | `length` |
| `oneof` | `invalid_value` |

You can also validate a struct directly: `validation.New().Struct(&v)` returns the same
`KindValidation` error.

## The error taxonomy

Every domain error is a `kernel/errors.Error` with a `Kind`. Each `Kind` maps to a fixed HTTP status and a
default stable machine `code` (`kernel/errors/errors.go`):

| Kind | HTTP | Default code | When |
|---|---|---|---|
| `KindValidation` | 400 | `validation_failed` | Bad input; carries field errors. |
| `KindUnauthenticated` | 401 | `unauthenticated` | No/invalid credentials. |
| `KindForbidden` | 403 | `permission_denied` | Authenticated but not permitted. |
| `KindNotFound` | 404 | `not_found` | Resource doesn't exist. |
| `KindTenantIsolation` | 404 | `tenant_mismatch` | Cross-tenant access — **masked as 404** so existence doesn't leak. |
| `KindConflict` | 409 | `conflict` | Uniqueness / state conflict. |
| `KindIdempotencyInFlight` | 409 | `retry_later` | A duplicate request is still processing. |
| `KindWorkflowState` | 409 | `invalid_transition` | Illegal workflow transition. |
| `KindVersionConflict` | 412 | `version_conflict` | Optimistic-concurrency mismatch. |
| `KindRuleViolation` | 422 | `rule_violation` | A configurable business rule rejected the request. |
| `KindRateLimited` | 429 | `rate_limited` | Too many requests. |
| `KindExternal` | 502 | `upstream_error` | A downstream dependency failed. |
| `KindInternal` | 500 | `internal` | Unexpected — **message is never exposed to clients**. |

## The `Error` type

```go
type Error struct {
    Kind   Kind          // → HTTP status + default code
    Code   string        // stable machine code (defaults to Kind.DefaultCode())
    Msg    string        // user-facing, safe (NEVER shown for KindInternal)
    Op     string        // operation name — logs only
    Fields []FieldError  // validation field details
    Err    error         // wrapped cause — logs only
}

type FieldError struct {
    Field   string // JSON path, e.g. "contacts[0].email"
    Code    string // "required", "max_length", "invalid_format", …
    Message string // safe for users
}
```

Construct one with `errors.E`:

```go
import kerr "github.com/qatoolist/wowapi/v2/kernel/errors"

// simple
return kerr.E(kerr.KindNotFound, "not_found", "request not found")

// wrapping a cause (the cause is logged, never sent to the client)
return kerr.E(kerr.KindConflict, "conflict", "title already exists")
```

The **`Op`** and wrapped **`Err`** are for logs only; **`KindInternal` messages are never rendered to the
client** — a 500 returns a generic problem document while the real cause is logged with the request ID.

## The wire shape: `application/problem+json`

`httpx.WriteError(ctx, w, err)` maps any `Error` to a `ProblemError` and writes it with
`Content-Type: application/problem+json` (`kernel/httpx/errors.go`):

```json
{
  "type": "https://errors.wowapi.dev/validation_failed",
  "title": "Validation failed",
  "status": 400,
  "detail": "one or more fields are invalid",
  "code": "validation_failed",
  "request_id": "01J…",
  "errors": [
    { "field": "title", "code": "required", "message": "title is required" },
    { "field": "email", "code": "invalid_format", "message": "email must be a valid email" }
  ]
}
```

Field notes:

- `type` — a stable URI derived from the machine `code`.
- `code` — the same stable machine code apps should switch on (not the human `title`).
- `request_id` — correlate the client-visible error with server logs.
- `errors[]` — present for `KindValidation` (and any error carrying `Fields`); omitted otherwise.
- `detail` — user-safe; **omitted entirely for `KindInternal`**.

## Success responses

Handlers wrap successful payloads with `httpx.OK` and write with `httpx.WriteJSON`:

```go
httpx.WriteJSON(w, http.StatusCreated, httpx.OK(dto))
```

This keeps success and error envelopes consistent across every module.

## Localizing responses (i18n)

By default every error is English. To serve localized `title`, `detail`, and field messages, wire the
`kernel/i18n` catalog and the `httpx.Locale` middleware — **machine `code`s and field paths never
change with the locale**, only the human-facing text, so clients keep switching on the stable `code`.

**What the framework ships.** `kernel/i18n` ships its own **English** catalog as the first bundle:
one problem `title` per error `Kind`, one message per validator tag, and one problem `detail` for the
framework's own well-known codes with a stable message (currently `validation_failed`) — stored under
the reserved `kernel.` namespace and keyed by the *stable machine identifier* (the kind's code / the
tag name / the error's code), never the English text. English is the default locale and the ultimate
fallback: a missing translation falls back to English, and a missing key falls back to the key itself
— a translation gap can never break a response. Internal logs stay technical English regardless of the
request locale.

**The `detail` contract.** `title` and validation field messages always localize (the framework
guarantees a catalog entry for every `Kind`/tag). `detail` is different: it localizes **only when a
`kernel.detail.<code>` catalog entry exists** for the error's machine `code`, via `i18n.KeyDetail(code)`
— which is `reservedPrefix + "detail." + code`, i.e. always under the framework's own `kernel.`
namespace. `KeyDetail` currently only names **framework-owned** codes (an `errors.Kind`'s
`DefaultCode()`, e.g. `validation_failed`); it has no module-qualified form, so it cannot name a
product-specific error code. Most codes (framework or product) carry a producer-supplied,
already-appropriate `Msg` and have no catalog entry — for those, `detail` falls back to that `Msg`
verbatim, unchanged from before. Internal errors never expose `detail` at all, translated or not.

**Registering product/module translations.** A module contributes a bundle per locale during
`Register`, under its own `<module>.` prefix (the framework owns `kernel.`, and `Register` rejects any
key under that reserved prefix or outside the module's own prefix — a module's bundle can only add its
*own* keys, e.g. an order-status label, never a `kernel.*` title/validation/detail key):

```go
func (m *Module) Register(mc module.Context) error {
    mc.I18n(i18n.Bundle{Locale: "mr", Messages: map[string]string{
        // Must be prefixed "<module>." — a bare kernel.* key here is rejected
        // at boot (Registry.Register's ownership check).
        "orders.status.shipped": "पाठवले",
    }})
    return nil
}
```

`app.Boot` merges every module's bundles with the framework catalog and returns the result as
`Booted.I18n`. Ownership is validated at boot: a module registering a key outside its prefix (or under
the reserved `kernel.` namespace) fails boot with the other registry checks.

### Catalog sources, precedence, and the file-based workflow (GAP-001B)

Translations do not have to live in Go maps. `wowapi init` scaffolds a **`locales/` tree** and an
**`i18n:` config section**, and the generated `cmd/api`, `cmd/worker`, and `cmd/migrate` binaries load
the configured sources through **one lifecycle** before boot completes — no product-authored loader
code. The framework loads four first-class source kinds, merged in a fixed **precedence** order:

1. **framework defaults** — the framework's own English `kernel.*` strings, embedded per-locale YAML
   (`kernel/i18n/locales/<locale>/kernel.yaml`); always present, lowest precedence.
2. **product framework-override files** — `locales/<locale>/kernel.yaml` in your repo. A `kernel.*` key
   here **overrides** the embedded framework default for that locale. You may *retranslate* a framework
   key but **not invent** a new `kernel.*` key.
3. **product/module catalog files** — `locales/<locale>/*.yaml` and `locales/<locale>.json` (or
   `locales/*.json`) under **your own `<name>.` namespace**. YAML for hand-authoring, JSON for
   tooling/large text-as-key catalogs.
4. **compiled Go bundles** — `internal/i18n/catalogs` returning `[]i18n.RawBundle`, for translations you
   want the compiler to own; highest static precedence.

A **DB overlay** is reserved as a future opt-in source (precedence would sit last, after validation) but
is **not** built today (catalogs freeze at boot — see below).

**Rules the loader enforces (and `wowapi i18n validate` checks in CI):** duplicate keys *within one
layer* fail; a later layer overriding an earlier layer's key is allowed (that is precedence); only the
framework-defaults and a sanctioned override layer may write `kernel.*`; every key must exist in the
default locale (so a lookup always has a fallback).

The canonical config the scaffold emits:

```yaml
i18n:
  default_locale: en
  supported_locales: [en, mr]
  locales_dir: locales   # loaded as a sanctioned framework-override source
  go_bundles: true       # load internal/i18n/catalogs
```

**Sanctioned in-code override path.** If you prefer Go over a `locales/<locale>/kernel.yaml` file, the
registry exposes a guarded `RegisterFrameworkLocale(bundle)` — the replacement for raw `Catalog.Add`. It
validates that every key is an existing `kernel.*` key (retranslate, don't invent) and records
violations at boot like every other registry. Raw `Catalog.Add` still exists but is a **no-op after the
catalog is frozen** (see below); use the config-driven `locales/` files or `RegisterFrameworkLocale`.

**Validation.** Run `wowapi i18n validate --dir locales --supported en,mr` (add `--strict` to require
every supported locale to define every key rather than relying on fallback). It fails on the four defect
classes: missing coverage, intra-layer duplicates, `kernel.*` ownership violations, and placeholder
drift. Wire it into your product CI.

**Freeze-at-boot (Decision 3).** After boot merges every source and module bundle, the catalog is
**frozen**: request-time reads never race a write, and a post-boot `Add` is silently ignored. Hot-reload
and the DB overlay are a separate opt-in concern, not part of this contract.

**Interpolation / pluralization — v1 is static strings only.** The catalog stores and returns **static
strings**; it has no message-template engine, named placeholders, or plural-form selection. The one
supported parameter mechanism is the framework's own `%s`-style validation messages (`min`, `max`,
`len`, `oneof`, `gte`, `lte`), whose single argument is filled by `kernel/validation` at render time —
**not** by the catalog. A translation of one of those messages must keep the same `%`-verb count as the
English template; `wowapi i18n validate` fails on a mismatch (placeholder drift). If your product needs
rich interpolation or pluralization, format the final string in your handler and store only the static
fragments in the catalog. This is a deliberate v1 scope limit, stated (not silently omitted) so you can
plan around it.

**Wiring the middleware.** Pass the merged catalog to `httpx.Locale`, placed after `RequestID` and
before your routes:

```go
h := httpx.Chain(mux,
    httpx.RequestID(),
    httpx.Locale(booted.RuntimeI18n()), // negotiates Accept-Language, sets Content-Language
    // …edge + auth middleware…
)
```

`httpx.Locale` parses `Accept-Language` (RFC 9110 q-values; a supported `mr` matches an offered
`mr-IN`), binds the negotiated locale to the request context, and sets `Content-Language` on the
response. `httpx.WriteError` then localizes the problem `title` and (where a `detail.<code>` entry
exists) `detail`, and `httpx.BindAndValidate` localizes field messages — no handler change required.
**Passing no catalog (or `nil`) is a valid zero-config setup: responses stay English, byte-for-byte
identical to a framework with no i18n.**

Example: a request with `Accept-Language: mr-IN,mr;q=0.9,en;q=0.8` against a catalog that supports
Marathi gets `Content-Language: mr`, a Marathi `title`/field message (and a Marathi `detail` if a
`detail.<code>` entry was registered for that error's code), and the **same** `code`/`field` as the
English response. An unsupported locale (e.g. `fr-FR`) falls back deterministically to English.

**Testing.** `testkit` provides `AssertNegotiatedLocale`, `NewLocaleRequest`, and
`AssertLocalizedProblem` to assert negotiation and that a problem localizes its title while keeping its
machine code stable.

## Patterns & pitfalls

| Situation | Do this |
|---|---|
| Bad input | Return `KindValidation` (or just let `BindAndValidate` produce it). Never hand-roll a 400. |
| Not found | `kerr.E(kerr.KindNotFound, "not_found", …)` — map `pgx.ErrNoRows` to it explicitly. |
| Cross-tenant access | Use `KindTenantIsolation` — it renders as **404**, not 403, to avoid leaking existence. |
| Optimistic concurrency | `KindVersionConflict` (412) when a `version` check fails. |
| Wrapping a DB error | `kerr.E(kind, code, safeMsg)` and let the cause be logged — don't echo driver text to clients. |
| Anything unexpected | Return the raw error; it becomes `KindInternal` → generic 500, real cause logged. |

Next: [Auth](auth.md) · [Testing](testing.md).
