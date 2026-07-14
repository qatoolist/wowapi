---
id: W01-E03-S001
type: story
title: Server timeouts and body bounds
status: accepted
wave: W01
epic: W01-E03
owner: W01Http
reviewer: unassigned
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-09
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W01-E03-S001-01
  - AC-W01-E03-S001-02
  - AC-W01-E03-S001-03
  - AC-W01-E03-S001-04
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W01-003
  - RISK-W01-E03-001
---

# W01-E03-S001 — Server timeouts and body bounds

## Story ID

W01-E03-S001

## Title

Server timeouts and body bounds

## Objective

Configure all four `http.Server` connection-level timeouts (read/write/idle/header) from
`config`, with safe non-zero defaults, in the product scaffold template that `wowapi init` renders;
reject an explicit zero-value timeout in the prod profile; and defensively bound the CSRF
middleware's unbounded `r.FormValue` read (gosec G120).

## Value to the framework

A generated product's HTTP server currently inherits Go's infinite defaults for connection-level
read, write, and idle timeouts. This is a real, exploitable resource-exhaustion surface (a
Slowloris-response-side variant: a client that opens a connection and reads the response arbitrarily
slowly, or leaves a connection idle, is never disconnected by the server). Closing this gap in the
**scaffold template** — not a one-off product patch — means every future `wowapi init` invocation
across every downstream product inherits the fix, which is the generic-framework-first way to deliver
this (mandate §2.3): the fix belongs in the thing that generates products, not in each product
individually.

## Problem statement

MATRIX CS-09 (via `requirement-inventory.md` row FBL-09) records: no `http.Server{}` literal in
wowapi's own generated output sets `ReadTimeout`, `WriteTimeout`, or `IdleTimeout`. Verified directly
against the current scaffold template
(`internal/cli/templates/init/cmd_api_main.go.tmpl:314-317`):

```text
srv := &http.Server{
    Addr:              cfg.HTTP.Addr,
    Handler:           handler,
    ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
}
```

Only `ReadHeaderTimeout` is set. `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` are left at Go's
zero value, which `net/http` documents as "no timeout" for each. The generated product's own
`cmd/api/main.go` (e.g. wowsociety's, at the equivalent lines) inherits this gap unchanged.

This is a distinct problem from the framework's existing per-request `httpx.Timeout` middleware
(`kernel/httpx/edge.go:168-182`), which bounds *handler execution time* by cancelling the request
context and writing a 503 — it does not bound the underlying TCP connection's read/write/idle time.
A slow-reading or slow-writing client can hold a connection open indefinitely regardless of how fast
the handler itself completes, because `http.TimeoutHandler` (which backs `httpx.Timeout`) buffers the
handler's response but does not control how long the runtime spends flushing bytes onto a slow
socket, nor how long it waits to read a slow request body before the handler even starts.

Separately, `kernel/httpx/csrf.go:118`'s `r.FormValue(p.FieldName)` call parses the request body with
no explicit size bound of its own. The framework's `BodyLimit` middleware
(`kernel/httpx/edge.go:157-166`) already wraps `r.Body` in an `http.MaxBytesReader` earlier in the
default chain — but CSRF's chain position is app-controlled (the scaffold template's own comments
note CSRF is placed "last (innermost, right before mux/the auth gate)" only under the opt-in browser
security profile). A product's own middleware ordering choice could place `CSRFProtect` before
`BodyLimit`, in which case `FormValue`'s body read is genuinely unbounded. gosec's G120 rule flags
exactly this pattern. `W01-E01-S002` (judged-linter-set) explicitly excludes this specific hit from
its own scope, cross-referencing it to this story.

## Source requirements

FBL-09 (`requirement-inventory.md` row FBL-09, target `W01-E03-S001`). Cross-referenced: MATRIX CS-09
(closure spec including the four default timeout values); W01-E01-S002's judged-linter-set story
(cross-references gosec G120 to this story rather than fixing it itself).

## Current-state assessment

Confirmed directly against current repository state (2026-07-12):

- `internal/cli/templates/init/cmd_api_main.go.tmpl:314-317` constructs `http.Server{Addr, Handler,
  ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout}`. `ReadTimeout`, `WriteTimeout`, `IdleTimeout` are
  absent from the literal — Go's `net/http` zero-value defaults (no timeout) apply.
- `kernel/config/config.go:104-114` defines `HTTP` with `Addr`, `ReadHeaderTimeout`,
  `RequestTimeout`, `MaxBodyBytes`, `CORSAllowedOrigins`, `RateLimit` — **no** `ReadTimeout`,
  `WriteTimeout`, or `IdleTimeout` field exists yet on this struct.
- `kernel/config/config.go:189-200` (`Framework.Validate()`) unconditionally rejects
  `HTTP.ReadHeaderTimeout <= 0`, `HTTP.RequestTimeout <= 0`, and `HTTP.MaxBodyBytes <= 0` — this
  rejection is **not** prod-gated; it fires in every environment. This is a materially different
  existing pattern from the SSRF-disable prod-only rejection this story's task brief cites as
  precedent (`kernel/config/config.go:261-263`, inside the `if f.Environment.IsProd()` block). Both
  precedents exist in the same file; this story must choose (and record the choice) rather than
  silently assume which one the new keys follow — see "Unresolved questions" in `plan.md`.
- `kernel/httpx/edge.go:157-166` (`BodyLimit`) and `kernel/httpx/edge.go:168-182` (`Timeout`, backed
  by `HTTP.RequestTimeout`) are both implemented and already wired into the scaffold template's
  default middleware chain (RequestID→Recover→Locale→SecureHeaders→CORS→Trace→metrics→AccessLog→
  RateLimit→BodyLimit→Timeout, per the template's own ordering comments at
  `cmd_api_main.go.tmpl:280-309`). These are not being re-implemented by this story.
- `kernel/httpx/csrf.go:114-119` shows the unsafe-methods branch reading `r.Header.Get(p.HeaderName)`
  first, falling back to `r.FormValue(p.FieldName)` only when `p.FieldName != ""` and the header was
  empty. The `FormValue` call carries no defensive size bound of its own.
- No response compression exists anywhere in the framework's HTTP layer — confirmed absent, matching
  `wave.md`'s framing that this is a reverse-proxy concern and explicitly out of scope.

## Desired state

The scaffold template's `http.Server{}` literal sets all four timeouts (`ReadHeaderTimeout`,
`ReadTimeout`, `WriteTimeout`, `IdleTimeout`) from new `config.HTTP` fields, each with a safe non-zero
default (read 30s, write 60s, idle 120s, header 10s — MATRIX CS-09's stated defaults). Every future
`wowapi init` invocation generates a `cmd/api/main.go` with all four timeouts configured. A prod-
profile config carrying an explicit zero value for any of the new keys fails `config.Validate` at
boot. `kernel/httpx/csrf.go`'s `FormValue` call is wrapped in a defensive `http.MaxBytesReader` so the
CSRF middleware is self-protecting regardless of its position in a product's own middleware chain.

## Scope

- Add `HTTP.ReadTimeout`, `HTTP.WriteTimeout`, `HTTP.IdleTimeout`, `HTTP.HeaderTimeout` fields to
  `kernel/config.HTTP` with MATRIX CS-09's stated safe defaults (30s/60s/120s/10s).
- Wire all four into the scaffold template's `http.Server{}` literal
  (`internal/cli/templates/init/cmd_api_main.go.tmpl`).
- Add a template-render test asserting all four timeout fields are present in generated output.
- Add prod-profile rejection of an explicit zero-value timeout for the new keys in
  `config.Validate`/`Framework.Validate()`, following the codebase's existing prod-gated-rejection
  pattern (SSRF-disable precedent) — see `plan.md` for the reconciliation with the *other* existing
  pattern (unconditional rejection of the three pre-existing HTTP timeout keys).
- Wrap `kernel/httpx/csrf.go:118`'s `r.FormValue(p.FieldName)` call in a defensive
  `http.MaxBytesReader` bound (gosec G120 fix).

## Out of scope

- `HTTP.ReadHeaderTimeout`, `HTTP.RequestTimeout`, `HTTP.MaxBodyBytes` — already exist, already
  validated. This story does not modify their existing (unconditional) validation behavior.
- Response compression — reverse-proxy concern, per `wave.md`.
- The wowsociety-side backport of the four scaffold-template timeout lines into its already-committed
  `cmd/api/main.go` — tracked as **PROD-03** (`requirement-inventory.md` §D). This story enables
  PROD-03 by fixing the template; it does not touch the wowsociety repository (framework/product
  boundary, mandate §2.3).
- W01-E01-S002's judged-linter-set enablement itself (gosec, errorlint, etc.) — this story only fixes
  the one named G120 hit that story's scope excludes; it does not enable the linter.

## Assumptions

- MATRIX CS-09's stated default values (read 30s / write 60s / idle 120s / header 10s) are treated as
  the specified defaults to implement, per the task brief's explicit instruction, not values to
  re-derive from first principles.
- The exact validation policy for the four new keys (unconditional rejection, matching the three
  pre-existing HTTP timeout keys, vs. prod-profile-only rejection, matching the SSRF-disable
  precedent) is an open implementation-time decision — see `plan.md` "Unresolved questions." This
  story's task brief specifies prod-profile-only; `story.md`'s current-state assessment above records
  that the more directly analogous existing HTTP-timeout-key pattern is actually unconditional. Both
  are legitimate existing precedents in the same file; the choice is not silently resolved here.
- `HeaderTimeout` is treated as a fourth, independent config key distinct from the already-existing
  `ReadHeaderTimeout`, per the task brief's naming. Whether these should instead be unified into one
  key (since `ReadHeaderTimeout` already serves the same purpose Go's `http.Server.ReadHeaderTimeout`
  field serves) is flagged as a question for implementation-time / review-time resolution in
  `plan.md`, since introducing a second key for what may be the same concept risks exactly the
  confusion RISK-W01-E03-001 names.

## Dependencies

None — see `../../dependencies.md` "Internal" section. No dependency on S002 or on any other W01
epic.

## Affected packages or components

- `kernel/config` (`config.go`: `HTTP` struct, `Framework.Validate()`, `Defaults()`).
- `internal/cli/templates/init/cmd_api_main.go.tmpl` (scaffold template).
- `internal/cli/templates/init/configs_base.yaml.tmpl` and/or `configs_local.yaml.tmpl` (if the new
  keys need example values in the generated config — to be determined at implementation time).
- `kernel/httpx/csrf.go` (`CSRFProtect` unsafe-method branch).
- Whatever template-render test harness already exists for scaffold-output assertions (shared
  primitive per DX-01 T5, referenced at wave level — to be located and reused, not reinvented, during
  implementation).

## Compatibility considerations

The new config keys are additive with non-zero safe defaults, so an existing (already-generated)
product's config that does not set them explicitly is unaffected at the config-loading layer — no
existing deployment's config file has to change. The *scaffold template output* changes for any
*future* `wowapi init` run, which is the intended fix surface. Existing generated products (like
wowsociety's already-committed `cmd/api/main.go`) do not receive the fix automatically; PROD-03 tracks
that backport as wowsociety's own work.

## Security considerations

This story directly closes a resource-exhaustion (denial-of-service) gap: unbounded connection
read/write/idle time. It also closes gosec G120 (unbounded `FormValue` read in CSRF middleware),
making the CSRF middleware defensively self-bounding regardless of a product's own middleware
ordering choice.

## Performance considerations

None expected beyond the timeout values themselves. The chosen defaults (30s/60s/120s/10s) are
generous enough not to affect legitimate slow-but-valid clients under normal network conditions —
this is the same judgment MATRIX CS-09 already made in specifying them.

## Observability considerations

None required by this story specifically; a connection terminated by a new server-level timeout
surfaces through Go's standard `net/http` server error logging, which is unchanged by this story.

## Migration considerations

None — additive config keys with safe defaults; no schema or data migration involved.

## Documentation requirements

The scaffold template's generated `README.md.tmpl` and/or `configs_base.yaml.tmpl` comments should
document the four timeout keys and their defaults, consistent with the existing documentation style
for `HTTP.ReadHeaderTimeout`/`RequestTimeout`/`MaxBodyBytes` in the same files — exact locations to be
confirmed at implementation time.

## Acceptance criteria

- **AC-W01-E03-S001-01**: `kernel/config.HTTP` has `ReadTimeout`, `WriteTimeout`, `IdleTimeout`,
  `HeaderTimeout` fields with defaults 30s/60s/120s/10s respectively, matching MATRIX CS-09.
- **AC-W01-E03-S001-02**: A template-render test asserts all four timeout fields are present, wired
  from `cfg.HTTP.*`, in the scaffold template's generated `http.Server{}` literal. The test fails
  against the current template (pre-fix) and passes after the fix (fail-first, mandate §13).
- **AC-W01-E03-S001-03**: A prod-profile `Framework` config carrying an explicit zero value for any of
  the four new timeout keys fails `config.Validate`, with a test proving this.
- **AC-W01-E03-S001-04**: `kernel/httpx/csrf.go`'s `r.FormValue(p.FieldName)` call is wrapped in a
  defensive `http.MaxBytesReader`; a re-run of gosec (once W01-E01-S002 enables it) shows the
  `csrf.go:118` G120 hit resolved.

## Required artifacts

- Scaffold-template diff (`cmd_api_main.go.tmpl`).
- Config schema addition (`kernel/config.HTTP` new fields).
- See `artifacts/index.md`.

## Required evidence

- Template-render assertion (fail-first: fails before, passes after).
- Prod-profile zero-timeout `config.Validate` rejection test.
- gosec G120 hit resolution (re-run once the linter is enabled by W01-E01-S002).
- See `evidence/index.md`.

## Definition of ready

Per `governance/definition-of-ready.md` — confirmed: specific (one coherent capability: connection-
level timeouts + one named CSRF hit), bounded (scope/out-of-scope both stated above), traceable
(`source_requirements: [FBL-09]`), measurable AC (4 numbered criteria above), dependencies identified
(none), assumptions recorded (validation-policy and `HeaderTimeout`-naming questions above),
`plan.md` drafted alongside this file.

## Definition of done

Per `governance/definition-of-done.md` — this story is `accepted` only once all 8 completion
requirements (mandate §2.5) are satisfied: implementation, required artifacts registered, tests
proving the 4 ACs, evidence registered with commit-pinned IDs, acceptance-criteria verification
recorded in `verification.md`, independent review passed (checklist in `definition-of-done.md`),
documentation requirement satisfied, and `deviations.md` either empty or fully accounted.

## Risks

RISK-W01-003 (prod-profile zero-timeout rejection interacting with an unset-config deployment —
mitigated by shipping safe non-zero defaults) and RISK-W01-E03-001 (new-key naming collision /
validation-policy inconsistency risk against the three pre-existing HTTP timeout keys). See
`../../risks.md` (epic-level).

## Residual-risk expectations

Even after acceptance, connection-level timeouts do not fully close every exhaustion vector (e.g. an
attacker opening many connections within the timeout window is a separate concern addressed by
rate-limiting and infrastructure-level connection caps, not this story). This residual risk is
expected and tracked as inherent to the scope boundary ("HTTP transport hardening," not "complete
DoS protection"), not a gap this story must additionally close.

## Plan

See `plan.md` (sibling file) for the full §8.5 plan content.
