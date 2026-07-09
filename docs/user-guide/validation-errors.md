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
import kerr "github.com/qatoolist/wowapi/kernel/errors"

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

By default every error is English. To serve localized `title` and field messages, wire the
`kernel/i18n` catalog and the `httpx.Locale` middleware — **machine `code`s and field paths never
change with the locale**, only the human-facing text, so clients keep switching on the stable `code`.

**What the framework ships.** `kernel/i18n` ships its own **English** catalog as the first bundle:
one problem `title` per error `Kind` and one message per validator tag, stored under the reserved
`kernel.` namespace and keyed by the *stable machine identifier* (the kind's code / the tag name),
never the English text. English is the default locale and the ultimate fallback: a missing translation
falls back to English, and a missing key falls back to the key itself — a translation gap can never
break a response. Internal logs stay technical English regardless of the request locale.

**Registering product/module translations.** A module contributes a bundle per locale during
`Register`, under its own `<module>.` prefix (the framework owns `kernel.`):

```go
func (m *Module) Register(mc module.Context) error {
    // Localize the framework's own strings for a new locale…
    mc.I18n(i18n.Bundle{Locale: "mr", Messages: map[string]string{
        i18n.KeyProblemTitle(errors.KindNotFound): "सापडले नाही",
        i18n.KeyValidationMessage("required"):     "हे फील्ड आवश्यक आहे",
        // …and your module's own keys (must be prefixed "<module>.").
        "orders.status.shipped": "पाठवले",
    }})
    return nil
}
```

`app.Boot` merges every module's bundles with the framework catalog and returns the result as
`Booted.I18n`. Ownership is validated at boot: a module registering a key outside its prefix (or under
the reserved `kernel.` namespace) fails boot with the other registry checks.

**Wiring the middleware.** Pass the merged catalog to `httpx.Locale`, placed after `RequestID` and
before your routes:

```go
h := httpx.Chain(mux,
    httpx.RequestID(),
    httpx.Locale(booted.I18n), // negotiates Accept-Language, sets Content-Language
    // …edge + auth middleware…
)
```

`httpx.Locale` parses `Accept-Language` (RFC 9110 q-values; a supported `mr` matches an offered
`mr-IN`), binds the negotiated locale to the request context, and sets `Content-Language` on the
response. `httpx.WriteError` then localizes the problem `title`, and `httpx.BindAndValidate` localizes
field messages — no handler change required. **Passing no catalog (or `nil`) is a valid zero-config
setup: responses stay English, byte-for-byte identical to a framework with no i18n.**

Example: a request with `Accept-Language: mr-IN,mr;q=0.9,en;q=0.8` against a catalog that supports
Marathi gets `Content-Language: mr`, a Marathi `title`/field message, and the **same** `code`/`field`
as the English response. An unsupported locale (e.g. `fr-FR`) falls back deterministically to English.

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
