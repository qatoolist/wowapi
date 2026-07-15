---
id: W05-E02-S001
type: story
title: Typed port-key API and registrar-forge safety proof
status: planned
wave: W05
epic: W05-E02
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-02
depends_on:
  - W05-E01-S001
blocks:
  - W05-E02-S002
acceptance_criteria:
  - AC-W05-E02-S001-01
  - AC-W05-E02-S001-02
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W05-E02-001
---

# W05-E02-S001 — Typed port-key API and registrar-forge safety proof

## Story ID

W05-E02-S001

## Title

Typed port-key API and registrar-forge safety proof

## Objective

Define `port.Key[T]` and the four generic free functions (`Define`/`Provide`/`Require`/`Resolve`)
bound to W05-E01's `Registrar` capability type, and prove — via an adversarial compile-fail fixture —
that capability confusion is impossible even though AR-01 and AR-02 share the one `Registrar` type
by design (D-02).

## Value to the framework

This story is AR-02's own foundation: every later task in this epic (the compiled provider graph,
boot-time validation, profile projection, lifecycle-manifest retirement, legacy adapter) is built on
`port.Key[T]` and the registrar-forge safety guarantee this story establishes.

## Problem statement

`requirement-inventory.md` row AR-02 groups this story's scope: "S001 port-api-and-forge-proofs
(T1, T2)." PLAN's own AR-02 task table: T1 — "Define `port.Key[T]`, reuse AR-01 T2's `Registrar`,
and the four generic free functions | AR-01 T1, T2 | Happy-path define/provide/resolve round-trip
compiles and works | Unit | `AR-02/port_api_unit_test.go` | Medium — Go's lack of
type-parameterized methods forces a first-class-argument API; ergonomics review needed." T2 —
"Internal compiler factory mints registrars with immutable owner identity | T1, AR-01 T1 | Module
code cannot manufacture a `Registrar` from a bare string | Adversarial compile-fail fixture |
`AR-02/registrar_forge_compile_fail_fixture/` | High — verify capability confusion is impossible if
AR-01/AR-02 share one `Registrar` type."

## Source requirements

AR-02 (T1, T2).

## Current-state assessment

No `port.Key[T]` type and no typed port API exist in the framework today — the current wiring
approach is ad hoc constructor calls, hand-copied per profile. This story's own re-confirmation step
is to re-read the current wiring mechanism at this story's actual start commit and confirm this
absence still holds.

## Desired state

A happy-path define/provide/resolve round-trip compiles and works using `port.Key[T]` and the four
generic free functions, bound to W05-E01's `Registrar`. Module code cannot manufacture a `Registrar`
from a bare string — proven by an adversarial compile-fail fixture — and, critically, this holds even
though AR-01 and AR-02 share the single `Registrar` type by design (D-02), which is exactly the
scenario T2's own risk note calls out as needing explicit verification.

## Scope

- `port.Key[T]` type definition.
- The four generic free functions: `Define`, `Provide`, `Require`, `Resolve`, each bound to a
  `Registrar`.
- The internal compiler factory minting registrars with immutable owner identity, for AR-02's own
  port-key registration flow (reusing, not duplicating, AR-01 T2's minting mechanism).
- The adversarial compile-fail fixture proving capability confusion is impossible.
- The happy-path round-trip unit test.

## Out of scope

- **The compiled provider graph itself, zero-reflection dispatch, boot-time validation** — S002's
  scope, built on this story's port-key API.
- **The three-profile projection, lifecycle-manifest retirement, legacy port adapter** — S002/S003's
  scope.

## Assumptions

- "Go's lack of type-parameterized methods forces a first-class-argument API" (PLAN T1's own risk
  note) is taken as a confirmed Go-language constraint informing this story's own API design, not an
  invented detail.

## Dependencies

Depends on W05-E01-S001 (AR-01 T1, T2 — the `ApplicationModel` and `Registrar` capability type this
story reuses). Blocks W05-E02-S002 (the compiled provider graph consumes this story's `port.Key[T]`
API).

## Affected packages or components

New: a `port` package (or equivalent, exact location TBD per `plan.md`) defining `Key[T]` and the
four generic functions.

## Compatibility considerations

None — this is new API surface with no existing consumer (AR-02's own wowsociety-impact note: "Not
affected... zero call sites for `ProvidePort`/`Port(` anywhere in wowsociety").

## Security considerations

T2's registrar-forge safety proof is this story's central security property — verifying that sharing
one `Registrar` type across AR-01 and AR-02 (per D-02's design) does not introduce a
capability-confusion vulnerability where a `Registrar` minted for one subsystem's purpose could be
misused for another's.

## Performance considerations

None material at this story's own scope — hot-path performance is S002's own T3 concern
(zero-reflection at `Resolve` time), building on this story's API shape.

## Observability considerations

None beyond the epic's own boot-time logging conventions.

## Migration considerations

None.

## Documentation requirements

Document `port.Key[T]`'s API shape and usage pattern, and the registrar-forge safety guarantee this
story establishes (referencing D-02's shared-`Registrar`-type design and why it remains safe).

## Acceptance criteria

- **AC-W05-E02-S001-01**: A happy-path define/provide/resolve round-trip compiles and works —
  proven by `AR-02/port_api_unit_test.go`.
- **AC-W05-E02-S001-02**: Module code cannot manufacture a `Registrar` from a bare string, and
  capability confusion is impossible given AR-01/AR-02 share one `Registrar` type — proven by
  `AR-02/registrar_forge_compile_fail_fixture/`.

## Required artifacts

- The `port.Key[T]` type and four generic free functions (code).
- The internal compiler factory extension for port-key minting (code).
- Port-API documentation.
See `artifacts/index.md`.

## Required evidence

- `AR-02/port_api_unit_test.go` output.
- `AR-02/registrar_forge_compile_fail_fixture/` output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W05-E01-S001
recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T2's compile-fail fixture genuinely proves capability
confusion is impossible given the shared-`Registrar`-type design.

## Risks

RISK-W05-E02-001 (T2's capability-confusion safety proof, PLAN's own High-risk column) — see
epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Residual risk is expected to be low once T2's adversarial fixture is genuinely proven and
independently re-confirmed by this story's own review task.

## Plan

See `plan.md`.
