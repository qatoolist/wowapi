---
id: PLAN-W05-E03-S002
type: plan
parent_story: W05-E03-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E03-S002

Per mandate §8.5.

## Proposed architecture

Four boot-time strictness checks layered on AR-01's ownership-bound model: duplicate-collector
rejection (with a legitimate-accumulation exception), empty-fragment rejection, post-seal
config-rejection extension (reusing D-03's error-not-panic mechanism), and an explicit waiver
mechanism gating `prod` readiness on required-capability adapter reality.

## Implementation strategy

1. Audit the current collector, required-fragment, and post-seal-config behavior at this story's
   start commit.
2. Implement duplicate-collector rejection, explicitly preserving the legitimate
   multi-locale-accumulation pattern.
3. Write `AR-04/duplicate_collector_rejection_test.go`: one adversarial fixture per collector type.
4. Implement empty-required-fragment rejection.
5. Write `AR-04/empty_required_fragment_test.go`.
6. Extend AR-01 T8's error-not-panic mechanism to config/namespace/collector state.
7. Write `AR-04/post_seal_config_rejection_test.go` as a regression re-run of the AR-01 T8 suite,
   extended to the new state categories.
8. Design the waiver mechanism: explicit optional-capability declaration; `prod` readiness fails on
   required-but-no-op/missing adapter unless a policy-approved waiver exists; waiver suppresses with
   an audit record.
9. Implement the waiver mechanism as a standalone, reusable primitive (not narrowly coupled to this
   story's own consumer), given its forward-shared-consumer status (SEC-06, DX-07).
10. Write `AR-04/prod_noop_adapter_readiness_test.go`: the integration matrix (profile × waiver ×
    adapter-real/no-op).
11. Document all four behaviors, with emphasis on T5's shared-primitive documentation.

## Expected package or module changes

Extensions to the collector implementations across the registration surface; a new
required-fragment-validation extension; an extension to the post-seal rejection mechanism from
AR-01; a new waiver-mechanism package.

## Expected file changes where determinable

Extensions to collector files across the registration surface (exact list TBD by the audit); new
fragment-validation files; new post-seal-rejection extension files; a new waiver-mechanism package;
new test files as named above.

## Contracts and interfaces

The waiver mechanism's own contract (exact shape TBD) must be designed as a reusable primitive from
the outset, since SEC-06 and DX-07 are named forward consumers — not retrofitted for reuse later.

## Data structures

The waiver mechanism's own internal representation (e.g. a waiver registry keyed by
capability/adapter identity, with policy-approval and audit metadata).

## APIs

The waiver mechanism likely exposes an internal API for declaring an optional capability and
checking waiver status at readiness time (exact shape TBD).

## Configuration changes

The waiver mechanism likely introduces a configuration surface for declaring waivers (exact shape
TBD, informed by this programme's existing config-fail-closed posture per CS-03).

## Persistence changes

None anticipated — waivers are likely configuration-level, not database-persisted, though this is an
implementation-time decision to confirm.

## Migration strategy

Not applicable.

## Concurrency implications

None material — boot-time checks.

## Error-handling strategy

T4's extension reuses D-03's error-not-panic contract exactly, not a parallel mechanism. T5's waiver
mechanism's readiness failure must name the specific missing/no-op adapter, consistent with this
programme's field-specific-error-message convention.

## Security controls

T5's waiver mechanism, including its audit-record requirement, is itself a required security/
compliance control — not optional hardening.

## Observability changes

T5's waiver audit record is a required observability/compliance artifact.

## Testing strategy

- `AR-04/duplicate_collector_rejection_test.go`: one adversarial fixture per collector type, plus a
  positive fixture proving legitimate multi-locale accumulation is not falsely rejected.
- `AR-04/empty_required_fragment_test.go`: adversarial fixture.
- `AR-04/post_seal_config_rejection_test.go`: regression re-run of the AR-01 T8 suite, extended.
- `AR-04/prod_noop_adapter_readiness_test.go`: integration matrix (profile × waiver ×
  adapter-real/no-op).

## Regression strategy

All four named tests are permanent regression guards for their respective properties.

## Compatibility strategy

T2's legitimate-accumulation exception is this story's central compatibility concern.

## Rollout strategy

Single story, landed as its own reviewable unit.

## Rollback strategy

Revert T2 if the legitimate-multi-locale-accumulation pattern is found to be falsely rejected;
revert T5 if the waiver mechanism's shape proves unsuitable for SEC-06/DX-07's later consumption
needs — escalate for redesign rather than shipping a narrow mechanism that would force a second,
divergent waiver system later.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-11). T2-T4 may proceed in parallel
(disjoint concerns); T5 is sequenced with awareness of its forward-shared-consumer status and should
receive the most design care in this story.

## Task breakdown

- **W05-E03-S002-T001** — Duplicate-collector rejection (T2; steps 2-3 above).
- **W05-E03-S002-T002** — Empty-required-fragment rejection (T3; steps 4-5 above).
- **W05-E03-S002-T003** — Post-seal config/namespace/collector rejection extension (T4; steps 6-7
  above).
- **W05-E03-S002-T004** — The shared waiver mechanism (T5; steps 8-10 above).

No independent-review task is added for this story — PLAN's own risk column values (Medium,
Low-medium, Low, Medium) are moderate, not High; T5's forward-shared-consumer status is addressed
through the plan's own design-care emphasis rather than a dedicated review task, consistent with
this wave's own task-brief guidance to use judgment scaled to risk for stories not explicitly named
as requiring review.

## Expected artifacts

Duplicate-collector rejection (code); empty-fragment rejection (code); post-seal-rejection
extension (code); the waiver mechanism (code).

## Expected evidence

The four named test outputs.

## Unresolved questions

- Exact waiver-mechanism configuration/persistence shape — to be decided at implementation time,
  informed by this programme's existing config-fail-closed posture (CS-03) and by anticipating
  SEC-06/DX-07's own consumption needs where knowable now.
- Exact set of "collector types" T2 must cover — to be confirmed by this story's own audit, not
  pre-enumerated by the source beyond "every collector."

## Approval conditions

This plan is approved for implementation once: (a) the waiver mechanism's shape is drafted with
explicit consideration of SEC-06/DX-07's anticipated consumption needs, and (b) the owner and
reviewer are assigned.
