---
id: PLAN-W05-E03-S001
type: plan
parent_story: W05-E03-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E03-S001

Per mandate §8.5.

## Proposed architecture

A manifest schema describing each module's declarations (identity + projection inputs only, per
T1's own scope bound), from which deterministic tooling derives every downstream projection
(routes, permissions, resources, schema, lifecycle, profiles, tests, docs). A golden-fixture test
harness proves the derivation is complete and deterministic. A lint rule enforces that no
declaration is hand-duplicated or silently omitted from a projection.

## Implementation strategy

1. Audit the current scattered-declaration surface at this story's start commit.
2. Design the manifest schema, scoped to identity + projection inputs, explicitly excluding DX-03's
   full typed-operation DSL.
3. Write `AR-03/manifest_schema_fixture_test.go`: round-trip against ≥1 existing internal fixture
   module.
4. Implement route registration/metadata derivation from the manifest.
5. Write `AR-03/golden_declaration_delta_test.go`: a golden-fixture manifest change deterministically
   produces the expected full projection diff (route/permission/resource/schema/OpenAPI/lifecycle/
   profile/test/doc) with no other hand-edited file.
6. Implement the duplicate-identity/omitted-projection lint rule.
7. Write `AR-03/duplicate_omission_lint_test.go`: adversarial fixtures for both failure modes.
8. Extend projection derivation to documentation/test/manifest export output.
9. Write `AR-03/full_projection_golden_test.go`, extending T3's golden-delta coverage.
10. Document the schema, the derivation tooling, and the golden-delta gate's role.

## Expected package or module changes

A new manifest-schema package; route-derivation tooling; a new lint rule (exact tooling TBD);
extended projection-derivation tooling for docs/tests/manifest export.

## Expected file changes where determinable

New manifest-schema definition files; new derivation-tooling files; new lint-rule files; new test
files as named above (T1, T3, T4, T5).

## Contracts and interfaces

The manifest schema's own field contract (identity + projection inputs, exact shape TBD, traceable
1:1 to existing scattered declarations).

## Data structures

The manifest schema's own internal representation.

## APIs

None externally facing (build/boot-time tooling).

## Configuration changes

None anticipated.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None material — build/boot-time tooling.

## Error-handling strategy

The lint rule's failure messages should clearly identify the specific duplicate identity or omitted
projection, consistent with this programme's broader field-specific-error-message convention.

## Security controls

None new beyond this wave's existing capability-security posture.

## Observability changes

The golden-delta test's own diagnostic output should clearly identify which projection field
diverged, if any — important given PLAN's own framing of this test as the acceptance gate itself.

## Testing strategy

- `AR-03/manifest_schema_fixture_test.go`: schema round-trip.
- `AR-03/golden_declaration_delta_test.go`: the acceptance gate itself.
- `AR-03/duplicate_omission_lint_test.go`: adversarial lint fixtures.
- `AR-03/full_projection_golden_test.go`: extended golden-delta coverage for docs/tests/manifest
  export.

## Regression strategy

The golden-delta test and the lint rule are both permanent regression guards — any future change
that reintroduces hand-duplication or an omitted projection is caught by the lint rule; any change
that breaks the deterministic derivation is caught by the golden-delta test.

## Compatibility strategy

T1's own acceptance criterion (round-trips against an existing fixture module without requiring that
module's own declarations to change) is this story's primary compatibility concern.

## Rollout strategy

Single story, landed as its own reviewable unit.

## Rollback strategy

Revert if the golden-delta test proves unreliable (flaky, non-deterministic) — treat as a blocking
defect in the manifest/projection design itself, per RISK-W05-E03-001's own framing, not a
test-infrastructure inconvenience.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-10). Step 5 (T3's golden-delta test) is the
critical path — PLAN's own framing that "this test IS the acceptance gate" means step 5 cannot be
treated as merely one item among the sequence; it defines the story's own completion.

## Task breakdown

- **W05-E03-S001-T001** — Manifest schema definition and round-trip proof (T1; steps 2-3 above).
- **W05-E03-S001-T002** — Route derivation and the golden-declaration-delta acceptance gate (T3;
  steps 4-5 above).
- **W05-E03-S001-T003** — Duplicate-identity/omitted-projection lint rule (T4; steps 6-7 above).
- **W05-E03-S001-T004** — Documentation/test/manifest export projections (T5; steps 8-9 above).
- **W05-E03-S001-T005** — Independent review (per mandate §14, scoped to this story, given T3's
  "this test IS the acceptance gate" framing).

## Expected artifacts

The manifest schema (code); route-derivation tooling (code); the lint rule (code); extended
projection tooling (code).

## Expected evidence

The four named test outputs (T1, T3, T4, T5).

## Unresolved questions

- Exact manifest schema field list (beyond "identity + projection inputs") — this story's own T1
  design work, informed by the audit of existing scattered declarations, not pre-specified by the
  source beyond the scope bound itself.
- Exact lint tooling mechanism (likely `go/analysis`-based, consistent with AR-06 T2's own approach,
  but not confirmed by the source for this specific rule).

## Approval conditions

This plan is approved for implementation once: (a) the manifest schema's field list is drafted from
the declaration audit, and (b) the owner and reviewer are assigned.
