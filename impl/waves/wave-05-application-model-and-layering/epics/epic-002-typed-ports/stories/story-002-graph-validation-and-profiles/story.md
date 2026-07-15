---
id: W05-E02-S002
type: story
title: Zero-reflection provider graph, boot-time validation, and profile projection
status: planned
wave: W05
epic: W05-E02
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-02
depends_on:
  - W05-E02-S001
blocks:
  - W05-E02-S003
  - W05-E03-S001
acceptance_criteria:
  - AC-W05-E02-S002-01
  - AC-W05-E02-S002-02
  - AC-W05-E02-S002-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W05-E02-002
---

# W05-E02-S002 — Zero-reflection provider graph, boot-time validation, and profile projection

## Story ID

W05-E02-S002

## Title

Zero-reflection provider graph, boot-time validation, and profile projection

## Objective

Build a type-erased provider graph with zero reflection on request hot paths; boot-time validation
rejecting duplicate providers, missing requirements, undeclared edges, cycles, and invalid scope/
lifetime edges; and the compilation of API/worker/migrate profiles as three projections of one
graph, eliminating hand-copied wiring templates.

## Value to the framework

This story delivers AR-02's central runtime property (zero hot-path reflection) and its central
correctness property (boot-time graph validation), plus the projection mechanism that eliminates
the framework's current hand-copied per-profile wiring — the single biggest source of wiring drift
between API, worker, and migrate entry points today.

## Problem statement

`requirement-inventory.md` row AR-02 groups this story's scope: "S002
graph-validation-and-profiles (T3, T4, T5)." PLAN's own AR-02 task table: T3 — "Type-erased provider
graph with zero reflection on request hot paths | T1-T2 | Benchmark/static check proves zero
`reflect.*` calls at `Resolve` time | Benchmark + lint | `AR-02/hotpath_no_reflection_bench.txt` |
Medium — naive implementations reflect per-call." T4 — "Boot-time graph validation: duplicate
providers, missing requirements, undeclared edges, cycles, invalid scope/lifetime edges | T1-T3 |
One adversarial fixture per failure class; errors name both owners | Adversarial suite, reusing
`kernel/lifecycle`'s existing scope-rank ordering | `AR-02/boot_graph_validation_test.go` | Medium —
absorb/replace existing lifecycle scope logic, don't duplicate it." T5 — "Compile API/worker/migrate
profiles as three projections of one graph | T1-T4, AR-03 | No hand-copied wiring template remains |
Integration: build all three from one fixture, assert capability subsets |
`AR-02/three_profile_projection_test.go` | Medium — sequence after AR-03's manifest shape is fixed."

## Source requirements

AR-02 (T3, T4, T5).

## Current-state assessment

Per PLAN's own evidence, `kernel/lifecycle` already contains scope-rank ordering logic this task's
T4 must reuse, not duplicate — this is a confirmed existing asset, not a from-zero build. No
type-erased provider graph and no compiled three-profile projection exist today; wiring is hand-
copied per profile.

## Desired state

Zero `reflect.*` calls occur at `Resolve` time, proven by benchmark and static lint. Boot-time
validation rejects duplicate providers, missing requirements, undeclared edges, cycles, and invalid
scope/lifetime edges, with errors naming both owners involved, one adversarial fixture per failure
class, reusing `kernel/lifecycle`'s existing scope-rank ordering rather than duplicating it.
API/worker/migrate profiles compile as three projections of one graph, with no hand-copied wiring
template remaining, proven by building all three from one fixture and asserting capability subsets.

## Scope

- The type-erased provider graph implementation, with zero-reflection-on-hot-path design (T3).
- The hot-path-no-reflection benchmark and static lint check (T3).
- Boot-time graph validation covering all five named failure classes: duplicate providers, missing
  requirements, undeclared edges, cycles, invalid scope/lifetime edges (T4), reusing
  `kernel/lifecycle`'s existing scope-rank ordering.
- The three-profile projection compiler: API, worker, migrate as three projections of one graph (T5).

## Out of scope

- **The port-key API itself** — S001's scope, already built; this story consumes it.
- **Retiring the hand-maintained `kernel/lifecycle` manifest entirely** — S003's own T6 scope,
  though this story's T4 does begin reusing `kernel/lifecycle`'s scope-rank logic as a first step.
- **AR-03's own manifest schema** — this story's T5 is sequenced with awareness of "AR-03's manifest
  shape" per PLAN's own note, but does not itself define that shape; AR-03 (W05-E03) is a separate
  epic that later depends on this story's T5 output.

## Assumptions

- T5's PLAN dependency row ("T1-T4, AR-03") is read as a forward-looking coordination note (this
  story's projection mechanism is what AR-03 later consumes) rather than a hard block requiring
  AR-03's own epic to complete first — since `impl/analysis/wave-allocation-detail.md`'s own
  dependency direction states AR-03 depends on AR-02, not the reverse. This is recorded explicitly
  as an interpretation of a PLAN dependency row against the wave-allocation's own canonical
  sequencing, not silently resolved.

## Dependencies

Depends on W05-E02-S001 (T1, T2 — the port-key API and registrar-forge safety this story's graph
consumes). Blocks W05-E02-S003 (T6's lifecycle-manifest retirement builds on this story's completed
graph) and, at wave scope, W05-E03-S001 (AR-03's manifest-derived-projection tooling depends on this
story's T5 three-profile projection).

## Affected packages or components

A new provider-graph package (exact location TBD); `kernel/lifecycle` (T4 reuses its scope-rank
ordering logic, not yet removing it — that is S003's own T6 scope).

## Compatibility considerations

T4's reuse (not duplication) of `kernel/lifecycle`'s existing scope-rank ordering is itself a
compatibility-preserving design choice — avoiding two divergent scope-ordering implementations
existing simultaneously.

## Security considerations

None material beyond this epic's own general capability-security posture (established in S001);
this story's boot-time validation (T4) is a correctness, not a security, concern in the strict
capability-confusion sense.

## Performance considerations

T3's zero-hot-path-reflection requirement is this story's central performance property — proven by
dedicated benchmark, not merely functional correctness.

## Observability considerations

T4's validation errors must "name both owners" per PLAN's own acceptance criterion — a
diagnosability requirement, not merely a pass/fail signal.

## Migration considerations

None.

## Documentation requirements

Document the provider graph's zero-reflection design, the five validated failure classes, and the
three-profile projection mechanism.

## Acceptance criteria

- **AC-W05-E02-S002-01**: Zero `reflect.*` calls occur at `Resolve` time — proven by
  `AR-02/hotpath_no_reflection_bench.txt` (benchmark) and a static lint check.
- **AC-W05-E02-S002-02**: Boot-time validation rejects duplicate providers, missing requirements,
  undeclared edges, cycles, and invalid scope/lifetime edges, one adversarial fixture per failure
  class, errors naming both owners, reusing `kernel/lifecycle`'s existing scope-rank ordering —
  proven by `AR-02/boot_graph_validation_test.go`.
- **AC-W05-E02-S002-03**: API/worker/migrate profiles compile as three projections of one graph, no
  hand-copied wiring template remains — proven by `AR-02/three_profile_projection_test.go` building
  all three from one fixture and asserting capability subsets.

## Required artifacts

- The type-erased provider graph (code).
- Boot-time graph validation (code).
- The three-profile projection compiler (code).
See `artifacts/index.md`.

## Required evidence

- `AR-02/hotpath_no_reflection_bench.txt`.
- `AR-02/boot_graph_validation_test.go` output.
- `AR-02/three_profile_projection_test.go` output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on S001 recorded,
the T5-AR-03 dependency-direction interpretation recorded explicitly, owner/reviewer assignment
pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed.

## Risks

RISK-W05-E02-002 (T3's zero-hot-path-reflection claim requiring careful implementation) — see
epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Residual risk is expected to be low once T3's benchmark and lint are both genuinely clean.

## Plan

See `plan.md`.
