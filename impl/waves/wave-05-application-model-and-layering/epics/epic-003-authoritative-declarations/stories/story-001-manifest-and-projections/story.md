---
id: W05-E03-S001
type: story
title: Manifest schema and derived-projection tooling
status: planned
wave: W05
epic: W05-E03
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-03
depends_on:
  - W05-E01
  - W05-E02
blocks: []
acceptance_criteria:
  - AC-W05-E03-S001-01
  - AC-W05-E03-S001-02
  - AC-W05-E03-S001-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W05-E03-001
---

# W05-E03-S001 — Manifest schema and derived-projection tooling

## Story ID

W05-E03-S001

## Title

Manifest schema and derived-projection tooling

## Objective

Define the manifest schema (scoped to identity + projection inputs, not DX-03's full typed-operation
DSL), derive route registration/metadata from it with a golden-fixture delta test as the acceptance
gate itself, add a lint rule failing on hand-maintained duplicate identity or an omitted projection,
and extend the golden-delta coverage to documentation/test/manifest export projections.

## Value to the framework

This story is AR-03's own substance: converting the framework's scattered, hand-duplicated
declarations into a single authoritative manifest from which every projection (routes, permissions,
resources, schema, lifecycle, profiles, tests, docs) is deterministically derived — closing the
drift risk that hand-duplication inherently creates.

## Problem statement

`requirement-inventory.md` row AR-03 groups this story's scope: "S001 manifest-and-projections
(AR-03 T1, T3, T4, T5; T2 = DX-06-owned, cross-ref only)." PLAN's own AR-03 task table: T1 — "Define
the manifest schema — scoped to what Wave 1 needs (identity + projection inputs), not DX-03's full
typed-operation DSL (Wave 4) | AR-01 T1 | Manifest fields traceable 1:1 to existing scattered
declarations, no new parallel metadata system introduced ahead of the model | Unit: manifest
round-trips against ≥1 existing internal fixture module | `AR-03/manifest_schema_fixture_test.go` |
Medium — scope-creep risk into DX-03 territory." T3 — "Derive route registration/metadata from the
manifest | T1, AR-01, AR-02 | A golden-fixture manifest change deterministically produces the
expected full projection diff (route/permission/resource/schema/OpenAPI/lifecycle/profile/test/doc)
with no other hand-edited file | Golden-delta test | `AR-03/golden_declaration_delta_test.go` | High
— this test IS the acceptance gate." T4 — "Lint rule failing on hand-maintained duplicate identity
or omitted projection | T1-T3 | Duplicate-identity and omitted-projection fixtures both fail lint |
Adversarial lint fixtures | `AR-03/duplicate_omission_lint_test.go` | Medium." T5 — "Documentation/
test/manifest export projections | T1-T4 | Extend T3's golden-delta to cover doc-table/manifest-
export output | Integration | `AR-03/full_projection_golden_test.go` | Low-medium — share fixtures
with AR-05."

## Source requirements

AR-03 (T1, T3, T4, T5). T2 (OpenAPI merge) is out of scope — see below.

## Current-state assessment

Per PLAN's own directive requirement, declarations today are scattered across the codebase with no
single authoritative manifest; routes, permissions, resources, schema, and other projections are
each independently maintained rather than derived from one source. This story's own re-confirmation
step is to audit the current declaration surface at this story's actual start commit and confirm
this scattered state still holds before designing the manifest schema against it.

## Desired state

The manifest schema's fields are traceable 1:1 to existing scattered declarations, with no new
parallel metadata system introduced ahead of the model, and it round-trips against at least one
existing internal fixture module. A golden-fixture manifest change deterministically produces the
expected full projection diff (route/permission/resource/schema/OpenAPI/lifecycle/profile/test/doc)
with no other hand-edited file. A lint rule fails on hand-maintained duplicate identity or an
omitted projection. The golden-delta coverage extends to documentation/test/manifest export output.

## Scope

- The manifest schema definition, scoped to identity + projection inputs (T1).
- Route registration/metadata derivation from the manifest, proven by the golden-fixture delta test
  (T3).
- The duplicate-identity/omitted-projection lint rule (T4).
- Documentation/test/manifest export projections, extending T3's golden-delta coverage (T5).

## Out of scope

- **AR-03 T2 (the OpenAPI merge fix — preserving every OpenAPI 3.1 top-level field)**. Per
  `requirement-inventory.md`'s explicit note: "T2 = DX-06 duplicate → single owner DX-06 (see
  duplicate-analysis)." PLAN's own T2 row independently confirms this: "duplicates DX-06's identical
  closure contract; assign single ownership." This story does not implement T2, and does not skip
  mentioning it — it is DX-06's scope (W06-E02-S001), recorded here as an explicit cross-reference.
- **DX-03's full typed-operation DSL** — PLAN's own T1 acceptance criterion explicitly excludes this
  ("not DX-03's full typed-operation DSL (Wave 4)"); DX-03 is W06-E01-S001's own design-investigation
  story, not this story's scope.
- **W05-E01's `ApplicationModel` and W05-E02's provider graph themselves** — already built by their
  own epics; this story consumes them.

## Assumptions

- T1's own risk note ("scope-creep risk into DX-03 territory") is a confirmed source-flagged risk,
  not an invented concern — this story's own plan must bound the manifest schema's scope explicitly
  against DX-03's excluded full DSL.
- T5's "share fixtures with AR-05" note (PLAN's own T5 risk column) is recorded as a forward
  coordination opportunity, not a hard dependency — AR-05 is W06-E04's own scope, and this story does
  not require AR-05's own stories to exist to complete its own T5 acceptance criterion.

## Dependencies

Depends on W05-E01 (full epic — AR-01's ownership-bound model) and W05-E02 (full epic — AR-02's
compiled provider graph, per T3's own dependency row: "T1, AR-01, AR-02"). No dependency within
W05-E03 (independent of S002).

## Affected packages or components

A new manifest-schema package (exact location TBD); route-registration/metadata derivation tooling;
a new lint rule (exact tooling TBD — likely `go/analysis`-based, consistent with this programme's
other lint tooling, e.g. AR-06 T2).

## Compatibility considerations

T1's own acceptance criterion requires the manifest to round-trip against an existing internal
fixture module without requiring that module's own declarations to change — the manifest schema is
designed to describe existing declarations, not to force their rewrite.

## Security considerations

None material beyond this wave's existing capability-security posture (built by W05-E01/E02).

## Performance considerations

None material — this is boot-time/build-time tooling, not a request-hot-path concern.

## Observability considerations

None beyond the golden-delta test's own diagnostic output (which projection field diverged, if any).

## Migration considerations

None.

## Documentation requirements

Document the manifest schema's fields and their 1:1 traceability to existing declarations; document
the golden-delta test's role as the acceptance gate itself, so a future contributor understands why
it cannot be weakened or skipped without escalation.

## Acceptance criteria

- **AC-W05-E03-S001-01**: The manifest schema's fields are traceable 1:1 to existing scattered
  declarations, no new parallel metadata system is introduced ahead of the model, and it round-trips
  against at least one existing internal fixture module — proven by
  `AR-03/manifest_schema_fixture_test.go`.
- **AC-W05-E03-S001-02**: A golden-fixture manifest change deterministically produces the expected
  full projection diff (route/permission/resource/schema/OpenAPI/lifecycle/profile/test/doc) with no
  other hand-edited file — proven by `AR-03/golden_declaration_delta_test.go`. This test IS the
  acceptance gate, per PLAN's own explicit framing — it must genuinely run, not be skipped or
  weakened.
- **AC-W05-E03-S001-03**: A lint rule fails on hand-maintained duplicate identity or an omitted
  projection, proven by `AR-03/duplicate_omission_lint_test.go`; the golden-delta coverage extends to
  documentation/test/manifest export output, proven by `AR-03/full_projection_golden_test.go`.

## Required artifacts

- The manifest schema definition (code).
- Route-derivation tooling (code).
- The duplicate-identity/omitted-projection lint rule (code).
- Documentation/test/manifest export projection tooling (code).
See `artifacts/index.md`.

## Required evidence

- `AR-03/manifest_schema_fixture_test.go` output.
- `AR-03/golden_declaration_delta_test.go` output.
- `AR-03/duplicate_omission_lint_test.go` output.
- `AR-03/full_projection_golden_test.go` output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W05-E01/E02
recorded, AR-03 T2's out-of-scope status explicitly recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the golden-delta test (T3) genuinely ran and genuinely
covers the full named projection surface, given PLAN's own "this test IS the acceptance gate"
framing.

## Risks

RISK-W05-E03-001 (T3's golden-delta test as the sole acceptance gate) — see epic-level `risks.md`
for full detail and mitigation/contingency.

## Residual-risk expectations

Residual risk is expected to be low once T3's golden-delta test is confirmed deterministic and
independently re-run by this story's own review task.

## Plan

See `plan.md`.

## Note (autopsy remediation R-1, 2026-07-16)

Status is unchanged — this story remains genuinely unexecuted as tracked (`planned`, all tasks
`todo`). However, the implementation-autopsy report
(`impl/reports/implementation-autopsy-report-2026-07-16.md`, §4 row W05-E03-S001, independent
verdict **contradictory**) found a repo-root `./AR-03/` directory (package `ar03_test`) already
exists, containing exactly the four test files this story's claimed artifacts name, while tracking
still says `planned`/`todo`. This is code landed outside this story's execution (autopsy H-6/H-7).
See deviation **DEV-PROG-002** in `impl/tracking/programme-deviations.md` for the full record.
— autopsy remediation R-1, 2026-07-16.
