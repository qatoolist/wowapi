---
id: W01-E03
type: epic
title: HTTP hardening
status: planned
wave: W01
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - FBL-09
  - FBL-08
  - CS-09
  - CS-08
depends_on: []
stories:
  - W01-E03-S001
  - W01-E03-S002
decisions: []
risks:
  - RISK-W01-002
  - RISK-W01-003
---

# W01-E03 — HTTP hardening

## Epic objective

Close two independent HTTP-transport-layer enforcement gaps the framework's own tooling already
supports but does not yet apply: unbounded connection-level server timeouts (Slowloris-response-side
exhaustion), and discretionary, opt-in request validation with no boot-time safety net for a mutating
route that forgets to call it.

## Problem being solved

`requirement-inventory.md` §B records two review findings targeting this epic:

- **FBL-09** — "HTTP server timeouts + CSRF body bound" — disposition `planned`, priority P1, target
  `W01-E03-S001`. MATRIX CS-09 identifies that no `http.Server{}` literal in wowapi's own generated
  output sets `ReadTimeout`/`WriteTimeout`/`IdleTimeout`: the product scaffold template
  (`internal/cli/templates/init/cmd_api_main.go.tmpl:314-317`) constructs `http.Server{Addr, Handler,
  ReadHeaderTimeout}` only, so Go's infinite connection-level defaults apply to slow-write and
  idle-connection exhaustion. This is distinct from the existing per-request `httpx.Timeout` handler
  timeout (`kernel/httpx/edge.go:168-182`, driven by the already-present `HTTP.RequestTimeout` config
  key) — a handler-time bound does not bound connection read/write/idle time, and the gap is real.
- **FBL-08** — "Central validation enforcement (RouteMeta seam)" — disposition `planned`, priority P1,
  target `W01-E03-S002`. This was a **new finding**: the architecture review graded validation A-/Ready
  by library choice (validator/v10 is present and well-integrated), not by enforcement. The
  closure-depth pass downgraded it once it was observed that `kernel/httpx/router.go`'s `RouteMeta`
  (the boot-validated metadata every route already carries — see FBL-08's own architectural insight
  below) has zero binding/validation references, and `httpx.BindAndValidate[T]`
  (`kernel/httpx/decode.go:52-67`) is an opt-in helper a handler author must remember to call. A
  handler that skips it gets zero validation with no framework safety net — "relies on package
  presence rather than effective behaviour."

Both findings describe the same underlying shape: a capability the framework already built
(`http.Server`'s own timeout fields; `RouteMeta`'s boot-time validation seam) is not being fully used
to close a real gap — this is why `wave.md` groups both stories under one epic even though they are
technically independent (timeouts vs. validation-enforcement): "HTTP layer hygiene closing an
enforcement gap the framework's own tooling already supports."

## Scope

- **S001**: adding `HTTP.ReadTimeout`, `HTTP.WriteTimeout`, `HTTP.IdleTimeout`, `HTTP.HeaderTimeout`
  config keys with safe non-zero defaults (read 30s / write 60s / idle 120s / header 10s, per MATRIX
  CS-09's own stated defaults); wiring all four into the scaffold template's `http.Server{}` literal;
  a prod-profile `config.Validate` rejection of an explicit zero-value timeout, following the existing
  SSRF-disable prod-rejection precedent (`kernel/config/config.go:261-263`); and folding in gosec
  G120's fix at `kernel/httpx/csrf.go:118` (`r.FormValue` unbounded read) as a defensive
  `http.MaxBytesReader` inside the CSRF middleware itself, cross-referenced in from W01-E01-S002's
  judged-linter-set story which explicitly excludes G120 from its own scope in favor of this story.
- **S002**: adding a `RouteMeta.Request` contract-declaration field (or equivalent type-token /
  `Validate bool` shape — exact field layout to be determined at implementation time, see the story's
  `plan.md`); a boot-time check that fails any POST/PUT/PATCH route with no declared request contract,
  with a waiver field explicitly forward-compatible with AR-04 T5's waiver mechanism (W05 scope, not
  yet built); a `BindAndValidate`-calling generic handler adaptor so declaring the type IS wiring the
  validation, not a separate manual call; and crud/scaffold template migration to the new adaptor.

## Out of scope

- **S001** does not touch `HTTP.RequestTimeout`/`HTTP.ReadHeaderTimeout`/`HTTP.MaxBodyBytes` — these
  three already exist in `kernel/config.HTTP` with their own always-on (not prod-gated)
  `Framework.Validate()` rejection of `<= 0` (`kernel/config/config.go:192-200`). S001 adds four **new**
  keys for the connection-level timeouts Go's `http.Server` itself exposes and that are currently
  unset; it does not re-derive or duplicate the existing request-level timeout.
- **S001** does not implement response compression — `wave.md`'s source framing records this as a
  reverse-proxy concern, explicitly out of scope.
- **S001** does not perform wowsociety's own backport of the four scaffold-template timeout lines into
  its already-committed, hand-edited `cmd/api/main.go`. That is tracked as **PROD-03** in
  `requirement-inventory.md` §D (product-level items) — this story enables PROD-03 by fixing the
  template that generates future scaffolds; it does not itself touch the wowsociety repository, per
  the framework/product boundary (mandate §2.3).
- **S002** does not implement AR-03 (W05 scope: RouteMeta-derived projections) or AR-04 T5 (W05 scope:
  the general boot-time-silent-behaviour waiver mechanism). S002's `RouteMeta.Request` field and its
  waiver field are built now, forward-compatible with both future mechanisms per MATRIX CS-08's
  explicit coordination note — this is a prose coordination note, not a blocking dependency (see
  `dependencies.md`).
- **S002** does not flip enforcement to default-on for wowsociety. Per FBL-08's own compat note
  ("profile-flag first"), the boot rejection ships behind a profile flag for at least one version;
  wowsociety auditing its handlers and flipping the flag to enforced-by-default is a downstream
  coordination note, not this story's execution step.

## Source requirements

FBL-09, FBL-08. Cross-referenced constraints: CS-09 (HTTP timeout closure spec, including its stated
default values), CS-08 (central validation enforcement closure spec, including the RouteMeta-seam
architectural insight and the "compat: profile-flag first" note).

## Architectural context

Both stories operate at the `kernel/httpx` transport layer, but at different seams:

- **S001** operates below the framework's request-handling middleware chain, at the raw
  `net/http.Server` construction seam. That construction does not live in wowapi itself — it lives in
  the **product scaffold template** (`internal/cli/templates/init/cmd_api_main.go.tmpl`), which
  `wowapi init` renders into a product's `cmd/api/main.go`. This means S001's fix changes what future
  `wowapi init` invocations generate, not a wowapi-internal runtime code path — the distinction matters
  for scope (framework boundary, mandate §2.3) and for delivery model (a template fix does not
  retro-apply to already-generated, hand-edited product code; see DX-07 T1 for the same delivery-model
  precedent). The existing mitigations already present and NOT being re-implemented by S001: `BodyLimit`
  via `http.MaxBytesReader` (`kernel/httpx/edge.go:157-166`), the per-request `httpx.Timeout` handler
  timeout (`kernel/httpx/edge.go:168-182`, backed by `HTTP.RequestTimeout`), and the full middleware
  chain (RequestID→Recover→Locale→SecureHeaders→CORS→Trace→metrics→AccessLog→RateLimit→BodyLimit→
  Timeout, per the scaffold template's own ordering comments).
- **S002** operates at the `RouteMeta` boot-time metadata seam (`kernel/httpx/router.go:18-45`).
  `RouteMeta` is already mandatory: `Router.Handle` calls `meta.validate()` and accumulates a
  registration error (surfaced via `Router.Err()`, checked at app boot) if a route is neither `Public`
  nor carries a non-empty `Permission` — "There is deliberately no registration path without it"
  (router.go:18-20 doc comment). S002's key architectural insight is that this exact mechanism —
  boot-validated, mandatory route metadata — already exists and already fails boot on a missing
  invariant; it is the natural seam to add a second invariant ("mutating routes must declare a request
  contract") rather than inventing a new enforcement mechanism. `BindAndValidate[T]`
  (`kernel/httpx/decode.go:52-67`) already composes `DecodeJSON[T]` (strict body decode, unknown-field
  rejection, size cap) with `Validator.StructCtx` (`kernel/validation/validation.go:97-128`, which
  already produces localized `KindValidation` field errors) — S002 does not need a new validation
  mechanism, only a way to make the existing one non-optional.

## Included stories

- **W01-E03-S001 — server-timeouts-and-body-bounds** (FBL-09): four new connection-level HTTP timeout
  config keys with safe defaults, scaffold-template wiring, prod-profile zero-timeout rejection, CSRF
  `MaxBytesReader` defensive bound (gosec G120).
- **W01-E03-S002 — central-validation-enforcement** (FBL-08): `RouteMeta.Request` contract field, boot-
  time rejection of undeclared mutating routes (behind a profile flag), a `BindAndValidate`-calling
  handler adaptor, crud/scaffold template migration.

## Dependencies

No dependency between S001 and S002 — they touch disjoint files (scaffold `http.Server` construction
and CSRF middleware vs. `RouteMeta`/router/decode/validation and crud templates) and can proceed in any
order or in parallel. This epic depends only on W00's exit gate, per `wave.md`'s entry criteria — no
epic-specific blocking dependency beyond that. See `dependencies.md` for the forward-compatibility
coordination notes with AR-03 (W05) and AR-04 T5 (W05), which are explicitly not `depends_on` blocking
relationships.

## Risks

RISK-W01-002 (FBL-08's boot-time rejection breaking an existing route that currently works only
because validation was silently skipped) and RISK-W01-003 (FBL-09's prod-profile zero-timeout
rejection interacting with a deployment that has not set timeouts explicitly) are inherited from
`../../risks.md` (wave-level). See `risks.md` (epic-level) for the epic-scoped elaboration.

## Required decisions

None recorded as blocking. S002's exact `RouteMeta.Request` field shape (a DTO prototype vs. a
`Validate bool` flag vs. a type token) is left as an implementation-time determination in
`story-002-central-validation-enforcement/plan.md`, per mandate §8.5's instruction not to invent
precise code changes the repository does not yet provide enough information to fix — this is recorded
as an unresolved question in that story's plan, not a required ADR blocking this epic. No
`decisions/` directory exists for either story (see each story's `story.md` front matter,
`decisions: []`).

## Epic acceptance criteria

- **AC-W01-E03-01**: All four new HTTP timeout config keys (`ReadTimeout`, `WriteTimeout`,
  `IdleTimeout`, `HeaderTimeout`) are present in the scaffold-rendered `http.Server{}` literal with the
  MATRIX CS-09 default values (30s/60s/120s/10s); a template-render test asserts all four are present
  in generated output; a prod-profile config with an explicit zero-value timeout fails
  `config.Validate`; `kernel/httpx/csrf.go`'s `r.FormValue` call is wrapped in a defensive
  `http.MaxBytesReader`.
- **AC-W01-E03-02**: Boot rejects a fixture POST/PUT/PATCH route with no declared `RouteMeta.Request`
  contract (behind the profile flag); an adversarial invalid-DTO POST to a route that does declare a
  contract, via the new handler adaptor, returns 400 with field errors; the fixture-route boot-passes-
  today / boot-fails-after-T1 fail-first sequence is evidenced.
- **AC-W01-E03-03**: Both stories have passed independent review per mandate §14, with S001 checked for
  the "safe defaults, not zero" framing (RISK-W01-003) and S002 checked for profile-flag compat
  discipline (RISK-W01-002) and for not silently duplicating a conflicting waiver design against the
  not-yet-built AR-04 T5.

## Closure conditions

Both stories reach `accepted` (each satisfying its own `closure.md`); AC-W01-E03-01 through
AC-W01-E03-03 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; no unresolved regression against either story.
