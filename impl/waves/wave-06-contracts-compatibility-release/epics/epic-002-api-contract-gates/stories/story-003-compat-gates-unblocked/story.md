---
id: W06-E02-S003
type: story
title: Compatibility gates unblocked — REL-03b (OpenAPI diff, event/schema compat, generated-consumer upgrade)
status: blocked
wave: W06
epic: W06-E02
owner: W06E02Impl
reviewer: W06-E01-E04-Execution.W06E02ReviewFinal
priority: P2
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - REL-03
depends_on:
  - W06-E02-S001
  - W06-E01-S001
  - W06-E01-S002
blocks: []
acceptance_criteria:
  - AC-W06-E02-S003-01
  - AC-W06-E02-S003-02
  - AC-W06-E02-S003-03
artifacts: []
evidence:
  - EV-W06-E02-S003-001
  - EV-W06-E02-S003-002
  - EV-W06-E02-S003-003
  - EV-W06-E02-S003-004
decisions: []
risks:
  - RISK-W06-E02-002
---

# W06-E02-S003 — Compatibility gates unblocked — REL-03b

## Story ID

W06-E02-S003

## Title

Compatibility gates unblocked — REL-03b (OpenAPI diff, event/schema compat, generated-consumer upgrade)

## Objective

Implement the three REL-03 compatibility-gate tasks that PLAN's own text describes as "hard-blocked on
Wave 1/4 architecture work": OpenAPI semantic diff (T3, blocked on DX-06), event/schema compatibility
(T5, blocked on DX-03/AR-03), and generated-consumer upgrade check (T7, hard-blocked on DX-04). **This
story cannot begin any given leg until that leg's specific unblocking story has reached `accepted`** —
each leg's entry criterion is stated explicitly below and must not be silently bypassed.

## Value to the framework

PLAN's own text is explicit about why these three tasks must be tracked, not silently dropped, even
though none is buildable today: "Recommend splitting into REL-03a (buildable now...) and REL-03b
(hard-blocked on Wave 1/4 architecture work: OpenAPI diff needs DX-06 first, event/schema compat needs
DX-03/AR-03's typed model first, generated-consumer upgrade needs DX-04 first) — do not schedule as one
monolithic P1 item, or 5 of 9 sub-tasks silently block the other 4." This story is the honest
counterpart to W06-E02-S002: it exists precisely so that REL-03's full nine-task scope is tracked to
completion (or explicit, non-silent deferral) rather than allowing the "buildable now" majority to
create the false impression that REL-03 as a whole is closed.

## Problem statement

`requirement-inventory.md` row REL-03 states: "b=T3(DX-06),T5(AR-03/DX-03),T7(DX-04)." MATRIX CS-15
confirms: "REL-03b (blocked) = T3 (on DX-06), T5 (on DX-03/AR-03), T7 (on DX-04)." PLAN's own per-task
evidence: T3 — "Blocked on DX-06 (a lossy merge can't be meaningfully diffed)"; T5 — "Blocked on
DX-03/AR-03 — the concept doesn't exist in current source... premature against today's stringly-typed
event registry"; T7 — "Hard-blocked on DX-04... cannot exist before DX-04." Each of these three
dependencies is itself a story in this program (T3 on this epic's own S001; T5 on W06-E01-S001's DX-03
design plus W05-E03's AR-03 remainder; T7 on W06-E01-S002's DX-04), meaning this story's own entry is
gated per-leg on other stories in this same wave (and one in W05) reaching `accepted`.

## Source requirements

REL-03 (T3, T5, T7 — the REL-03b blocked-legs subset).

## Current-state assessment

Per PLAN's own evidence: none of the three mechanisms exists today, and none *can* exist today, because
each depends on a prerequisite that either does not yet exist in the codebase (DX-06's merge-
completeness, DX-04's golden consumer) or is explicitly deferred to future design work outside this
programme's implementation scope (DX-03's typed DSL — this programme only produces a design record for
DX-03, not an implementation, per W06-E01-S001's own scope boundary). T5 in particular is premature
against today's actual event-registry implementation, per PLAN's own framing: "premature against
today's stringly-typed event registry."

## Desired state

Each of the three legs is implemented once — and only once — its specific unblocking story has reached
`accepted`:

- **T3 (OpenAPI semantic diff)** — entry criterion: **W06-E02-S001 (DX-06) `accepted`.** Once DX-06's
  merge-completeness closure exists, T3 implements a semantic diff classifying breaking changes per
  DX-06's own 3.1/2020-12 baseline, failing a seeded breaking-OpenAPI fixture.
- **T5 (event/schema compatibility)** — entry criterion: **both W06-E01-S001 (DX-03 design record)
  `accepted` AND W05-E03 (AR-03 remainder) `accepted`.** Even once both land, T5's own scope in this
  programme is bounded by DX-03 remaining design-only (per W06-E01-S001's own scope) — T5 can only be
  implemented against whatever compatibility-mode concept AR-03's typed model actually delivers, not
  against DX-03's undelivered implementation. This is recorded as a genuine open question below, not
  silently resolved.
- **T7 (generated-consumer upgrade check)** — entry criterion: **W06-E01-S002 (DX-04) `accepted`.** Once
  DX-04's golden-consumer fixture and its upgrade-replay mechanism exist, T7 reuses that same drill
  (PLAN's own framing: "Reuses DX-04's drill") rather than building a second one.

## Scope

- **T3** — OpenAPI semantic diff, classifying breaking changes per DX-06's 3.1/2020-12 baseline; a
  seeded breaking-OpenAPI fixture fails the gate. Entry-gated on W06-E02-S001.
- **T5** — Event/schema compatibility check tied to a `Compatibility` mode; an incompatible bump fails
  when `CompatibilityBackward` is declared; a seeded breaking-event fixture fails the gate. Entry-gated
  on W06-E01-S001 AND W05-E03.
- **T7** — Generated-consumer upgrade check: golden consumer at N-1, upgraded to N, contracts re-pass,
  reusing DX-04's own drill. Entry-gated on W06-E01-S002.

## Out of scope

- **T1, T2, T4, T6, T8, T9 (REL-03a)** — W06-E02-S002's own scope, not duplicated here.
- **DX-06, DX-03, DX-04's own implementation** — each is its own story elsewhere in this wave; this
  story only consumes their `accepted` output, it does not implement any part of them.
- **Implementing T5 against an assumed future DX-03 implementation** — this story's own T5 scope is
  bounded to whatever AR-03's typed model (once accepted) actually delivers for a `Compatibility` mode
  concept; it does not implement against DX-03's design record as if that record were itself executable
  — DX-03 remains design-only throughout this programme's scope (see W06-E01-S001).

## Assumptions

- Each leg's entry criterion is stated as this story's own `depends_on` front matter and restated here
  in prose, per this story's own explicit design requirement (from the task brief: "this story's entry
  criteria must explicitly reference the stories that unblock each of its three legs").
- T5's exact implementable scope, once its two entry criteria are satisfied, is genuinely uncertain
  until AR-03's own typed-model shape (W05-E03) is known — PLAN's own T5 acceptance criterion
  ("Incompatible bump fails when `CompatibilityBackward` declared") presupposes a `Compatibility` mode
  concept that must come from AR-03's actual delivered shape, not from DX-03's design-only record. This
  is recorded as a genuine open question, not resolved by this planning document.
- This story's own status may legitimately remain `planned` (not `ready`) for an extended period if its
  entry criteria are not yet satisfied when the rest of this wave otherwise closes — this is expected
  and should be recorded honestly in `closure.md` as deferred-with-restated-unblocking-condition, per
  `governance/definition-of-done.md`'s "partially-accepted" framing at epic/wave scope.

## Dependencies

**T3 depends on W06-E02-S001 (this epic's own sibling story) reaching `accepted`.** **T5 depends on both
W06-E01-S001 (DX-03 design record) and W05-E03 (AR-03 remainder, cross-wave) reaching `accepted`.** **T7
depends on W06-E01-S002 (DX-04) reaching `accepted`.** No leg of this story may begin implementation
before its own specific entry criterion is satisfied — a leg beginning early is itself a deviation from
this story's own plan and must be recorded as such in `deviations.md`, not silently absorbed.

## Affected packages or components

T3: an extension to this epic's own S001 semantic-diff work, scoped to OpenAPI specifically. T5: a new
event/schema compatibility-check mechanism, exact location TBD pending AR-03's own delivered shape. T7:
reuses W06-E01-S002's DX-04 drill, adding a REL-03-specific invocation of it — no new fixture, no new
harness.

## Compatibility considerations

All three legs are themselves compatibility-enforcement mechanisms. T5 specifically depends on a
compatibility-*mode* concept (`CompatibilityBackward` or equivalent) that does not exist in the
framework today and whose exact shape is AR-03's own delivered responsibility, not this story's to
invent.

## Security considerations

Not directly applicable beyond what each unblocking story (DX-06, AR-03, DX-04) already addresses in
its own scope.

## Performance considerations

Not applicable — these are CI-time gates.

## Observability considerations

Each gate, once implemented, should report clearly which specific incompatibility was detected, per the
same pattern established in W06-E02-S002's six gates.

## Migration considerations

Not applicable.

## Documentation requirements

For each leg, once implemented: document the gate's purpose, invocation, and failure-diagnosis
guidance. For any leg still blocked at this story's own closure: document the exact unblocking
condition restated, so a future reader knows precisely what must happen before that leg can begin.

## Acceptance criteria

- **AC-W06-E02-S003-01**: T3 (OpenAPI semantic diff) — once W06-E02-S001 (DX-06) is `accepted`, a seeded
  breaking-OpenAPI fixture fails the diff gate, classified per DX-06's own 3.1/2020-12 baseline. If
  W06-E02-S001 is not yet `accepted` at this story's own closure attempt, this AC is recorded as
  deferred with the unblocking condition restated, not silently marked failed or silently dropped.
- **AC-W06-E02-S003-02**: T5 (event/schema compatibility) — once both W06-E01-S001 (DX-03 design) and W05-E03
  (AR-03 remainder) are `accepted`, a seeded breaking-event fixture fails when `CompatibilityBackward`
  (or AR-03's actual equivalent concept) is declared. If either entry criterion is unmet at this story's
  own closure attempt, this AC is recorded as deferred with both unblocking conditions restated.
- **AC-W06-E02-S003-03**: T7 (generated-consumer upgrade check) — once W06-E01-S002 (DX-04) is `accepted`, the
  golden consumer at N-1 upgraded to N re-passes contracts, reusing DX-04's own drill rather than a
  second implementation. If W06-E01-S002 is not yet `accepted` at this story's own closure attempt, this
  AC is recorded as deferred with the unblocking condition restated.

## Required artifacts

- The OpenAPI semantic-diff gate for T3 (once unblocked).
- The event/schema compatibility-check mechanism for T5 (once unblocked).
- The generated-consumer upgrade check invocation for T7 (once unblocked), reusing DX-04's drill.
See `artifacts/index.md`. Any artifact whose leg remains blocked at story-closure time is recorded as
"not yet produced — blocked," not silently omitted from the index.

## Required evidence

- Seeded breaking-OpenAPI-fixture test output (T3).
- Seeded breaking-event-fixture test output (T5).
- Generated-consumer N-1-to-N upgrade re-pass output (T7).
See `evidence/index.md`.

## Definition of ready

This story has a non-standard readiness posture: it may remain in `planned` status, correctly and
non-defectively, for as long as its per-leg entry criteria remain unsatisfied. Per
`governance/definition-of-ready.md`'s "dependencies identified" requirement, this story's DoR is
satisfied for planning purposes (its front matter `depends_on` correctly lists all three unblocking
stories, and this `story.md` states each leg's exact entry criterion) even while none of its three legs
can yet begin. A leg individually becomes `ready` for implementation only once its own specific entry
criterion is satisfied — this story does not require all three entry criteria to be satisfied
simultaneously before any leg may begin.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted` (in full) or
`partially-accepted` (if any leg remains blocked): every leg whose entry criterion was satisfied during
this story's execution window is implemented, evidenced, and independently reviewed; every leg still
blocked at closure attempt time is recorded in `closure.md` as deferred-with-restated-unblocking-
condition, not silently dropped or falsely marked complete.

## Risks

RISK-W06-E02-002 (this story's three legs may remain blocked past this epic's own closure attempt if
their unblocking stories are delayed) — see epic-level `risks.md` for full detail and mitigation/
contingency.

## Residual-risk expectations

Residual risk cannot be fully eliminated within this story's own scope — it is inherently dependent on
three other stories' own completion timing, two of them cross-epic and one cross-wave. The mitigation is
honest tracking (per-leg entry criteria, partial-acceptance status), not risk elimination.

## Plan

See `plan.md`.
