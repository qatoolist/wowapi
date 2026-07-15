---
id: W01-E03-S002
type: story
title: Central validation enforcement
status: accepted
wave: W01
epic: W01-E03
owner: W01Http
reviewer: unassigned
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-08
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W01-E03-S002-01
  - AC-W01-E03-S002-02
  - AC-W01-E03-S002-03
  - AC-W01-E03-S002-04
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W01-002
---

# W01-E03-S002 — Central validation enforcement

## Story ID

W01-E03-S002

## Title

Central validation enforcement

## Objective

Make request-body validation for mutating routes non-discretionary by enforcing, at boot time, that
every POST/PUT/PATCH route declares a request contract on its `RouteMeta`, and by providing a
`BindAndValidate`-calling handler adaptor so declaring the contract is itself how a handler wires
validation — closing the gap where a handler author who forgets to call the existing opt-in
`BindAndValidate[T]` helper gets zero validation with no framework safety net.

## Value to the framework

The framework already has a strong validation library integration (`kernel/validation`, wrapping
`go-playground/validator/v10`, producing localized field errors) and a strong boot-time-mandatory
metadata pattern (`RouteMeta`, which already cannot be omitted — `Router.Handle` fails registration
if a route lacks a `Permission` or `Public` declaration). This story connects those two existing,
already-good primitives at their natural seam, rather than introducing a new validation mechanism —
this is precisely the "utilisation" principle other W01 stories (W01-E01) also apply: a capability
the framework already built is not yet fully used to close a real gap.

## Problem statement

`kernel/httpx/router.go:18-45`'s `RouteMeta` mandates `Permission`/`Public` at boot but has zero
awareness of request validation. `kernel/httpx/decode.go:52-67`'s `BindAndValidate[T]` composes
`DecodeJSON[T]` (strict decode, unknown-field rejection, size cap) with
`kernel/validation.Validator.StructCtx` (localized field-error production) — but it is a generic
function a handler author calls manually. `kernel/httpx/router.go` has zero references to binding or
validation anywhere in its own logic. A handler that is wired into the router with a valid
`RouteMeta` (satisfying the *existing* boot check) but that never calls `BindAndValidate` receives
zero validation, and nothing in the framework detects this — "relies on package presence rather than
effective behaviour," per the architecture review's own framing. This was a **new finding**: the
original review graded validation A-/Ready based on the library choice being sound, not on whether
its use was actually enforced anywhere; the closure-depth pass downgraded the grade once this
enforcement gap was specifically identified.

## Source requirements

FBL-08 (`requirement-inventory.md` row FBL-08, target `W01-E03-S002`). Cross-referenced: MATRIX CS-08
(closure spec, including the "compat: profile-flag first" note and the AR-03/AR-04-T5 coordination
notes).

## Current-state assessment

Confirmed directly against current repository state (2026-07-12):

- `kernel/httpx/router.go:18-33` — `RouteMeta` currently has exactly 5 fields: `Permission`,
  `Public`, `Scope`, `Idempotent`, `Sensitive`. No field relating to request-body validation exists.
- `kernel/httpx/router.go:36-45` — `RouteMeta.validate()` enforces exactly one invariant: `Public` XOR
  non-empty `Permission`. This is the pattern S002's new invariant will extend, not replace.
- `kernel/httpx/router.go:73-89` — `Router.Handle` calls `meta.validate()` and, on error, appends to
  `r.errs` (not a panic) — the router accumulates all registration errors so a module's whole route
  set is validated at once, surfaced later via `Router.Err()` at app boot. This is the exact mechanism
  a new "mutating route needs a declared contract" check should plug into.
- `kernel/httpx/decode.go:52-67` — `BindAndValidate[T](r, v, maxBytes)` already exists, already
  composes decode + validate correctly, already returns `KindValidation` with field errors on
  failure. It takes a `*validation.Validator` and a `maxBytes` value as explicit parameters — it is
  not currently invoked from anywhere inside `router.go` or any router-adjacent adaptor; it is called
  directly by handler code (confirmed: no call site inside `kernel/httpx` itself other than its own
  definition and tests).
- `kernel/validation/validation.go:53-74` (`New()`) registers only a `TagNameFunc` — confirmed zero
  custom validators registered anywhere in the repository, matching the task brief's framing.
- No `RouteMeta.Request` field, no boot-time mutating-route-contract check, and no generic handler
  adaptor exist anywhere in the codebase as of this assessment.
- AR-04 T5 (the general boot-time-silent-behaviour waiver mechanism this story's own waiver field must
  stay forward-compatible with) is W05 scope and **not yet built** — confirmed via
  `requirement-inventory.md` row AR-04 ("T2–T5 planned, dep AR-01... T5 waiver shared w/
  SEC-06/DX-07"). AR-03 (the RouteMeta-projection mechanism this story's `Request` field must remain a
  stable input for) is also W05 scope and not yet built.

## Desired state

`RouteMeta` gains a request-contract-declaration mechanism (exact shape — DTO prototype, `Validate
bool` flag, or type token — to be determined at implementation time per `plan.md`). Boot-time
validation (extending the existing `meta.validate()` pattern) fails any POST/PUT/PATCH route whose
metadata declares no request contract and no waiver, when a profile flag enables this enforcement. A
new generic handler adaptor calls `BindAndValidate` using the declared contract, so a handler that
uses the adaptor cannot accidentally skip validation — declaring the contract on `RouteMeta` and
wiring the handler through the adaptor are the same act. crud/scaffold templates are migrated to use
the new adaptor.

## Scope

- Add a `RouteMeta.Request` field (or equivalent contract-declaration mechanism) for mutating verbs.
- Add a boot-time check (extending `meta.validate()` / `Router.Handle`'s existing error-accumulation
  pattern) that fails a POST/PUT/PATCH route with no declared request contract and no waiver.
- Add a waiver field on `RouteMeta` for genuinely body-less mutations, designed additively/forward-
  compatible with AR-04 T5's future waiver mechanism (not yet built) — see
  `../../dependencies.md` "Forward-compatibility coordination notes."
- Add a `BindAndValidate`-calling generic handler adaptor so declaring the `Request` contract is
  itself how a handler wires validation.
- Migrate crud/scaffold templates to use the new adaptor.
- Ship the boot-time rejection behind a profile flag for at least one version (compat: profile-flag
  first, per FBL-08's own note), not enforced-by-default immediately.

## Out of scope

- AR-03 (W05 scope: RouteMeta-derived projections) and AR-04 T5 (W05 scope: the general waiver
  mechanism) — this story builds forward-compatibly with both but implements neither.
- Flipping the profile flag to enforced-by-default for wowsociety, or auditing wowsociety's existing
  handlers for missing `BindAndValidate` calls — a downstream coordination note, not this story's
  execution step.
- Any change to `kernel/validation`'s tag-to-code mapping, custom validator registration, or i18n
  message catalog — this story enforces the *use* of the existing validation library, it does not
  extend the library itself.
- Domain/cross-field/rule-engine validation (module-level `domain/validation.go`, per the two-layer
  validation strategy documented in `kernel/validation/validation.go`'s own package doc) — this story
  is about struct-tag shape validation enforcement only.

## Assumptions

- The exact `RouteMeta.Request` field shape is an open implementation-time question — see `plan.md`
  "Unresolved questions." This story's plan explicitly does not invent the precise Go type/field
  layout, per mandate §8.5.
- The profile-flag mechanism this story ships behind is assumed to be a new, story-specific flag
  (e.g. on `config.Security` or a new `config` subsection) rather than reuse of an existing flag,
  since no existing "validation enforcement" flag was found during current-state assessment — this is
  flagged as an assumption to confirm at implementation time, not a confirmed fact.
- AR-04 T5's eventual waiver mechanism shape is unknown (it is not yet designed). This story's own
  waiver field is therefore designed to be minimal and additive (e.g. a simple boolean or reason
  string) rather than anticipating AR-04 T5's exact eventual API, to minimize the surface AR-04 T5
  later has to reconcile.

## Dependencies

None blocking — see `../../dependencies.md` "Internal" section (no dependency on S001) and "Forward-
compatibility coordination notes" (AR-03, AR-04 T5 — explicitly not `depends_on` relationships).

## Affected packages or components

- `kernel/httpx/router.go` (`RouteMeta`, `meta.validate()`, `Router.Handle`).
- `kernel/httpx/decode.go` (new adaptor, likely alongside `BindAndValidate`).
- `kernel/validation` (consumed, not modified, by the new adaptor).
- crud/scaffold templates (`internal/cli/templates/...` — exact template files to be confirmed at
  implementation time; likely a `crud`-specific template set distinct from the `init` scaffold S001
  touches).
- `kernel/config` (if the profile flag lives in framework config rather than a product-level config
  section — to be determined).

## Compatibility considerations

The boot-time rejection is explicitly compat-gated: it ships behind a profile flag, defaulting to
**off** (advisory or no-op) for at least one version, per FBL-08's "compat: profile-flag first" note
and RISK-W01-002's mitigation. An existing route that currently boots successfully without a declared
contract continues to boot successfully while the flag is off. Only when a product explicitly
enables the flag does the boot-time rejection take effect — at which point that product is expected
to have already audited its own mutating routes (a downstream, product-side coordination step, not
this story's own work).

## Security considerations

This is a security-adjacent story (FBL-08's priority is explicitly framed this way in
`requirement-inventory.md`): closing a class of gap where a mutating endpoint silently accepts
unvalidated input because a handler author forgot one opt-in call. The adversarial 400-field-errors
test (AC-W01-E03-S002-02 below) is the direct proof this gap is closed for any route built through
the new adaptor.

## Performance considerations

None expected — the boot-time check runs once at startup (consistent with the existing
`meta.validate()` pattern's cost profile); the adaptor itself is a thin wrapper around the already-
existing `BindAndValidate`, adding no new runtime validation logic.

## Observability considerations

None required by this story specifically. A boot-time rejection surfaces through the existing
`Router.Err()` → app-boot-failure path, consistent with every other route-registration error.

## Migration considerations

None — no data or schema migration. The profile-flag mechanism is the compatibility/migration
strategy: existing routes are not forced to add a contract until a product opts into enforcement.

## Documentation requirements

The crud/scaffold template migration (T003) should update any generated documentation/comments that
currently reference the old (manual `BindAndValidate` call) pattern to reference the new adaptor
instead, consistent with the existing level of template documentation (see S001's equivalent
documentation-requirements note for the same "exact locations to be confirmed at implementation
time" caveat).

## Acceptance criteria

- **AC-W01-E03-S002-01**: A fixture route registering a POST handler with no declared
  `RouteMeta.Request` contract (and no waiver) **boots successfully today** (pre-fix) — this is the
  fail-first proof the defect is real (mandate §13).
- **AC-W01-E03-S002-02**: The same fixture route **fails at boot** after T1's check is implemented
  and the profile flag is enabled, with an error identifying the specific route and the missing
  contract.
- **AC-W01-E03-S002-03**: An adversarial test posts an invalid DTO (violating at least one
  `validate:` struct tag) to a fixture route built through the new adaptor and declaring a
  `RouteMeta.Request` contract, and receives HTTP 400 with field errors (matching the existing
  `KindValidation` → field-error shape `BindAndValidate` already produces).
- **AC-W01-E03-S002-04**: The waiver field exempts a genuinely body-less mutating route from the
  boot-time rejection, proven by a fixture route using the waiver that boots successfully with the
  enforcement flag enabled.

## Required artifacts

- `RouteMeta.Request` contract type (or resolved-shape equivalent).
- Handler adaptor (`kernel/httpx`).
- Updated crud/scaffold template(s).
- See `artifacts/index.md`.

## Required evidence

- Boot-rejection test output (fixture route, both pre-fix-boots and post-fix-fails states).
- Adversarial invalid-DTO 400 test output.
- Waiver-exemption boot-success test output.
- See `evidence/index.md`.

## Definition of ready

Per `governance/definition-of-ready.md` — confirmed: specific (one coherent capability: boot-time
mutating-route contract enforcement), bounded (scope/out-of-scope both stated above, explicitly
excluding AR-03/AR-04 T5 implementation), traceable (`source_requirements: [FBL-08]`), measurable AC
(4 numbered criteria above), dependencies identified (none blocking; forward-compatibility notes
recorded as prose, not `depends_on`), assumptions recorded (field shape, flag mechanism, waiver
design above), `plan.md` drafted alongside this file.

## Definition of done

Per `governance/definition-of-done.md` — this story is `accepted` only once all 8 completion
requirements (mandate §2.5) are satisfied, with particular attention (per this epic's
`acceptance.md` AC-W01-E03-03) to the independent reviewer confirming: profile-flag compat discipline
was honored (RISK-W01-002), and the waiver field design does not conflict with AR-04 T5's future
mechanism (a "no unsupported completion claims" and "architecture boundaries preserved" checklist
item, per `definition-of-done.md`'s independent-review checklist).

## Risks

RISK-W01-002 (boot-time rejection breaking an existing route that currently works only because
validation was silently skipped — mitigated by the profile-flag-first strategy). See `../../risks.md`
(epic-level).

## Residual-risk expectations

Even after acceptance and even with the profile flag eventually flipped to enforced-by-default in a
downstream product, this story does not guarantee every mutating route's validation *rules* are
correct — it only guarantees every mutating route declares *some* contract and that contract's rules
actually run. A handler that declares an overly permissive DTO (e.g. all fields optional) satisfies
this story's enforcement while still accepting effectively unvalidated input; that is a code-review
concern for the declared contract's own quality, not a gap this story's boot-time mechanism can
close. This residual risk is expected and out of scope.

## Plan

See `plan.md` (sibling file) for the full §8.5 plan content.
