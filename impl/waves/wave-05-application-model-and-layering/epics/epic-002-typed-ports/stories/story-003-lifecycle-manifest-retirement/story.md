---
id: W05-E02-S003
type: story
title: Lifecycle manifest retirement and legacy port adapter
status: planned
wave: W05
epic: W05-E02
owner: unassigned
reviewer: unassigned
priority: medium
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-02
depends_on:
  - W05-E02-S002
blocks:
  - W05-E05-S001
acceptance_criteria:
  - AC-W05-E02-S003-01
  - AC-W05-E02-S003-02
artifacts: []
evidence: []
decisions: []
risks: []
---

# W05-E02-S003 — Lifecycle manifest retirement and legacy port adapter

## Story ID

W05-E02-S003

## Title

Lifecycle manifest retirement and legacy port adapter

## Objective

Retire the hand-maintained `kernel/lifecycle` manifest in favor of the generated provider graph,
preserving existing lint-failure classes now as data-driven checks, and build the legacy port
adapter (`ProvidePort`/`Port` shim onto the typed graph) for any existing caller — confirmed today
to have zero external callers.

## Value to the framework

This story completes AR-02's own arc: from S001's typed API through S002's compiled, validated,
projected graph, to this story's retirement of the last hand-maintained wiring artifact
(`kernel/lifecycle`'s manifest) and the compatibility shim ensuring no theoretical existing caller
breaks.

## Problem statement

`requirement-inventory.md` row AR-02 groups this story's scope: "S003
lifecycle-manifest-retirement + legacy shim (T6, T7)." PLAN's own AR-02 task table: T6 — "Retire
hand-maintained `kernel/lifecycle` manifest in favor of the generated graph | T1-T5 |
`lifecycle.go`/`manifest.go` deleted or generated; existing 5 lint failure classes still pass, now
data-driven | Regression: existing lifecycle-lint classes pass |
`AR-02/lifecycle_lint_generated_test_output.txt` | Low-medium." T7 — "Legacy port adapter
(`ProvidePort`/`Port` shim onto the typed graph) | T1-T6 | Existing calls (none in wowsociety;
possibly wowapi-internal fixtures) compile/resolve unchanged | Integration |
`AR-02/legacy_port_adapter_compat_test_output.txt` | Low — confirmed zero external callers."

## Source requirements

AR-02 (T6, T7).

## Current-state assessment

`kernel/lifecycle`'s manifest (`lifecycle.go`/`manifest.go`) is currently hand-maintained, with 5
existing lint-failure classes checked against it. No legacy port adapter exists yet, but PLAN's own
evidence confirms zero external callers of `ProvidePort`/`Port(` in wowsociety today — this story's
own re-confirmation step is to re-run that repo-wide search at this story's actual start commit
(across both wowapi-internal and wowsociety) before concluding the adapter has no real consumer to
validate against beyond wowapi-internal fixtures, if any.

## Desired state

`lifecycle.go`/`manifest.go` are deleted or generated from S002's provider graph; the existing 5
lint-failure classes still pass, now data-driven rather than hand-maintained, proven by regression
test. The legacy port adapter compiles/resolves unchanged for any existing caller (wowapi-internal
fixtures, if any; confirmed none in wowsociety), proven by integration test.

## Scope

- Retirement of `kernel/lifecycle`'s hand-maintained manifest, replaced by generation from S002's
  provider graph (T6).
- Regression proof that the existing 5 lint-failure classes still pass, now data-driven (T6).
- The legacy port adapter (`ProvidePort`/`Port` shim onto the typed graph) (T7).
- Confirmation, via re-run repo-wide search, that zero external callers exist beyond any
  wowapi-internal fixtures (T7).

## Out of scope

- **The provider graph, boot-time validation, and three-profile projection themselves** — S002's
  scope, already built; this story consumes them.
- **Any change to the 5 existing lint-failure classes' own semantics** — this story preserves them
  as-is, converting only their data source from hand-maintained to generated.

## Assumptions

- PLAN T7's own confirmed-zero-external-callers finding ("none in wowsociety; possibly wowapi-
  internal fixtures") is taken as a fact requiring re-confirmation at this story's own start commit,
  not as a permanently-fixed fact — a re-run repo-wide search is this story's own required first
  step for T7, consistent with this programme's fail-first re-confirmation convention.

## Dependencies

Depends on W05-E02-S002 (T6, T7 both depend on T1-T5 / T1-T6 respectively — the full preceding
graph, validation, and projection surface). Blocks W05-E05-S001 (FBL-01's kernel re-home is
sequenced after this epic completes in full).

## Affected packages or components

`kernel/lifecycle` (`lifecycle.go`/`manifest.go` — deleted or regenerated); a new legacy port
adapter (exact location TBD).

## Compatibility considerations

T6's regression proof (existing lint-failure classes still pass) and T7's compatibility proof
(existing calls compile/resolve unchanged) are both this story's central compatibility concerns.

## Security considerations

None material beyond this epic's existing capability-security posture.

## Performance considerations

None material.

## Observability considerations

None beyond existing lint/CI reporting.

## Migration considerations

None (code-only change, no schema/data migration).

## Documentation requirements

Document the generated-manifest replacement for `kernel/lifecycle` and the legacy port adapter's own
compatibility guarantee and confirmed-zero-external-caller status.

## Acceptance criteria

- **AC-W05-E02-S003-01**: `lifecycle.go`/`manifest.go` are deleted or generated; the existing 5
  lint-failure classes still pass, now data-driven — proven by
  `AR-02/lifecycle_lint_generated_test_output.txt`.
- **AC-W05-E02-S003-02**: Existing calls to `ProvidePort`/`Port` (none confirmed in wowsociety;
  possibly wowapi-internal fixtures) compile/resolve unchanged through the legacy adapter — proven
  by `AR-02/legacy_port_adapter_compat_test_output.txt`.

## Required artifacts

- The generated `kernel/lifecycle` manifest replacement (code).
- The legacy port adapter (code).
See `artifacts/index.md`.

## Required evidence

- `AR-02/lifecycle_lint_generated_test_output.txt`.
- `AR-02/legacy_port_adapter_compat_test_output.txt`.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on S002 recorded,
owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed.

## Risks

None beyond this epic's general "depends on the preceding story landing correctly" transitive risk.
PLAN's own risk column values (Low-medium for T6, Low for T7) are the lowest in this epic.

## Residual-risk expectations

Residual risk is expected to be low given both tasks' own Low/Low-medium PLAN risk ratings and their
regression/compatibility-proof-driven acceptance criteria.

## Plan

See `plan.md`.
