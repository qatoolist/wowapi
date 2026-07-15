---
id: W05-E03-S002
type: story
title: Boot-time strictness and the shared no-op-adapter waiver mechanism
status: planned
wave: W05
epic: W05-E03
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-04
depends_on:
  - W05-E01
blocks: []
acceptance_criteria:
  - AC-W05-E03-S002-01
  - AC-W05-E03-S002-02
  - AC-W05-E03-S002-03
artifacts: []
evidence: []
decisions: []
risks: []
---

# W05-E03-S002 — Boot-time strictness and the shared no-op-adapter waiver mechanism

## Story ID

W05-E03-S002

## Title

Boot-time strictness and the shared no-op-adapter waiver mechanism

## Objective

Reject duplicate collectors and empty required fragments at boot; extend AR-01 T8's post-seal
error-not-panic contract to config/namespace/collector state; and build the explicit
optional-capability waiver mechanism so a `prod` profile with a required-but-no-op/missing adapter
and no waiver fails readiness by name — the same mechanism SEC-06 and DX-07 later consume, per this
wave's own forward-dependency note.

## Value to the framework

This story closes AR-04's remaining boot-time silent-behaviour gaps and delivers a load-bearing
shared primitive: the waiver mechanism T5 builds is explicitly not scoped only to this story's own
consumer — `impl/analysis/wave-allocation-detail.md`'s own note states "T5 builds the shared waiver
mechanism consumed by SEC-06/DX-07." Building it once, correctly, here avoids three independent,
possibly-divergent waiver implementations across the framework.

## Problem statement

`requirement-inventory.md` row AR-04 states (target column W05-E03-S002 — this story; an initial
inventory typo pointing at a nonexistent S003 was corrected on 2026-07-12, see
`tracking/change-log.md`): "Eliminate boot-time silent
behaviour | IMPL | P1 | partial | ... | T1 EXECUTED (verified ×2); T2–T5 planned, dep AR-01; T5
waiver shared w/ SEC-06/DX-07." PLAN's own AR-04 task table: T2 — "Reject duplicate collectors
(currently last-writer-wins) | AR-01 T1 | Every collector rejects a second write to the same identity
| One adversarial fixture per collector type | `AR-04/duplicate_collector_rejection_test.go` | Medium
— distinguish illegitimate duplicate from legitimate multi-locale accumulation." T3 — "Reject empty
required fragments | AR-01, T1-T2 | A module declaring a required-but-empty fragment fails boot |
Adversarial fixture | `AR-04/empty_required_fragment_test.go` | Low-medium." T4 — "Post-seal write
rejection reused from AR-01 T8 | AR-01 T8 | Same error-not-panic contract extended to config/
namespace/collector state | Regression re-run of AR-01 T8 suite | `AR-04/
post_seal_config_rejection_test.go` | Low." T5 — "Explicit optional-capability declaration; `prod`
readiness fails on required-but-no-op/missing adapter unless a policy-approved waiver exists | AR-01,
AR-02, T1-T4 | `prod` + no-op adapter + no waiver → readiness fails named; `local` + same config →
succeeds; waiver present → suppressed and audited | Integration matrix: profile × waiver ×
adapter-real/no-op | `AR-04/prod_noop_adapter_readiness_test.go` | Medium — shares scope with SEC-06
and DX-07's readiness closure contracts; build the waiver mechanism once."

## Source requirements

AR-04 (T2, T3, T4, T5). T1 is already executed — see below.

## Current-state assessment

Per `requirement-inventory.md`'s own AR-04 row: "T1 EXECUTED (verified ×2)." This story does not
re-plan or re-implement T1 (unknown-namespace rejection at boot). T2-T5 remain planned per the same
row. This story's own re-confirmation step for T2-T5 is to audit the current collector,
required-fragment, and post-seal-config behavior at this story's actual start commit, confirming
the gaps PLAN describes still hold.

## Desired state

Every collector rejects a second write to the same identity, distinguishing illegitimate duplicates
from legitimate multi-locale accumulation. A module declaring a required-but-empty fragment fails
boot. The error-not-panic contract from AR-01 T8 (D-03) extends to config/namespace/collector state.
A `prod` profile with a required-but-no-op/missing adapter and no waiver fails readiness by name; the
same configuration under `local` succeeds; a policy-approved waiver suppresses the failure with an
audit record.

## Scope

- Duplicate-collector rejection, with explicit handling of the legitimate-multi-locale-accumulation
  exception (T2).
- Empty-required-fragment rejection (T3).
- Post-seal config/namespace/collector rejection, reusing AR-01 T8's error-not-panic contract (T4).
- The explicit optional-capability waiver mechanism: `prod` readiness failure on
  required-but-no-op/missing adapter without a waiver; `local` succeeds; waiver present suppresses
  and audits (T5) — built once, as the shared primitive SEC-06 and DX-07 later consume.

## Out of scope

- **AR-04 T1 (unknown-namespace rejection at boot)** — already executed and verified twice per
  `requirement-inventory.md`. Not re-planned here.
- **SEC-06's own consumption of the waiver mechanism** (W03 scope, already built or in-progress) —
  this story builds the mechanism; SEC-06's own story is responsible for its own consumption, not
  modified here.
- **DX-07 T4's own consumption of the waiver mechanism** (W04-E04-S003's deferred-linked item) — same
  boundary; this story does not implement DX-07 T4.

## Assumptions

- T2's "legitimate multi-locale accumulation" exception is taken as a confirmed source-flagged
  distinction (PLAN's own risk column), not an invented nuance — the duplicate-collector rejection
  must specifically preserve this legitimate accumulation pattern while rejecting genuine duplicates.
- T5's waiver mechanism, though built here, is explicitly a forward-shared primitive per
  `impl/analysis/wave-allocation-detail.md`'s own note. This story's own acceptance criteria are
  scoped to proving the mechanism works for AR-04's own T5 scenario (the profile × waiver ×
  adapter-real/no-op integration matrix); confirming SEC-06's and DX-07's own successful consumption
  is each of those items' own responsibility, not re-verified here.

## Dependencies

Depends on W05-E01 (full epic — AR-01's `ApplicationModel`, `Registrar`, and T8's error-not-panic
contract). T5 additionally depends on AR-02 (W05-E02) per PLAN's own T5 dependency row: "AR-01, AR-02,
T1-T4." No dependency within W05-E03 (independent of S001).

## Affected packages or components

The collector implementations across the registration surface (T2); required-fragment validation
(T3); the post-seal rejection mechanism from AR-01 T8, extended (T4); a new waiver-mechanism package
(T5, exact location TBD).

## Compatibility considerations

T2's distinction between illegitimate duplicates and legitimate multi-locale accumulation is the
story's central compatibility concern — an overly-aggressive duplicate check would break the
legitimate pattern.

## Security considerations

T5's waiver mechanism is itself a security-adjacent control: a `prod` profile silently accepting a
no-op adapter for a required capability is exactly the kind of silent misconfiguration this story's
epic (and AR-04 as a whole) exists to close. The waiver's own audit record is a required control, not
optional.

## Performance considerations

None material — boot-time checks.

## Observability considerations

T5's waiver mechanism requires an audit record when a waiver suppresses a readiness failure — this
is a required observability/compliance artifact, not optional.

## Migration considerations

None.

## Documentation requirements

Document each of the four boot-strictness behaviors, with particular emphasis on T5's waiver
mechanism as a shared, cross-consumer primitive — future SEC-06/DX-07 implementers need a clear
reference for how to correctly consume it.

## Acceptance criteria

- **AC-W05-E03-S002-01**: Every collector rejects a second write to the same identity, without
  rejecting legitimate multi-locale accumulation — proven by
  `AR-04/duplicate_collector_rejection_test.go`. A module declaring a required-but-empty fragment
  fails boot — proven by `AR-04/empty_required_fragment_test.go`.
- **AC-W05-E03-S002-02**: The error-not-panic contract from AR-01 T8 (D-03) extends to
  config/namespace/collector state — proven by a regression re-run of the AR-01 T8 suite,
  `AR-04/post_seal_config_rejection_test.go`.
- **AC-W05-E03-S002-03**: A `prod` profile with a required-but-no-op/missing adapter and no waiver
  fails readiness by name; the same configuration under `local` succeeds; a policy-approved waiver
  suppresses the failure with an audit record — proven by the integration matrix (profile × waiver ×
  adapter-real/no-op), `AR-04/prod_noop_adapter_readiness_test.go`.

## Required artifacts

- Duplicate-collector rejection (code).
- Empty-required-fragment rejection (code).
- Post-seal config/namespace/collector rejection extension (code).
- The waiver mechanism (code).
See `artifacts/index.md`.

## Required evidence

- `AR-04/duplicate_collector_rejection_test.go` output.
- `AR-04/empty_required_fragment_test.go` output.
- `AR-04/post_seal_config_rejection_test.go` output.
- `AR-04/prod_noop_adapter_readiness_test.go` output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W05-E01 (and, for
T5, W05-E02) recorded, AR-04 T1's already-executed status explicitly recorded (not re-planned),
owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed.

## Risks

None beyond this epic's general "depends on the preceding epics landing correctly" transitive risk.
PLAN's own risk column values (Medium, Low-medium, Low, Medium) are moderate, not High, for this
story's four tasks.

## Residual-risk expectations

Residual risk is expected to be low once T2's legitimate-vs-illegitimate distinction and T5's
integration matrix are both confirmed by their own named tests.

## Plan

See `plan.md`.
