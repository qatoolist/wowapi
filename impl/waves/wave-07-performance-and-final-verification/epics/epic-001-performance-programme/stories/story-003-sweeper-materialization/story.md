---
id: W07-E01-S003
type: story
title: Sweeper and worker materialization — bounded batches, leased outbox, N+1 removal
status: accepted
wave: W07
epic: W07-E01
owner: W07-Scoping-Dispatch.W07E01S003
reviewer: W07-Scoping-Dispatch.W07E01S003ReviewR
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - PERF-04
depends_on:
  - W04-E01
  - W04-E02
blocks: []
acceptance_criteria:
  - AC-W07-E01-S003-01
  - AC-W07-E01-S003-02
  - AC-W07-E01-S003-03
  - AC-W07-E01-S003-04
  - AC-W07-E01-S003-05
  - AC-W07-E01-S003-06
  - AC-W07-E01-S003-07
artifacts:
  - ART-W07-E01-S003-001
  - ART-W07-E01-S003-002
  - ART-W07-E01-S003-003
  - ART-W07-E01-S003-004
  - ART-W07-E01-S003-005
  - ART-W07-E01-S003-006
  - ART-W07-E01-S003-007
  - ART-W07-E01-S003-008
evidence:
  - EV-W07-E01-S003-001
  - EV-W07-E01-S003-002
  - EV-W07-E01-S003-003
  - EV-W07-E01-S003-004
  - EV-W07-E01-S003-005
  - EV-W07-E01-S003-006
  - EV-W07-E01-S003-007
  - EV-W07-E01-S003-008
decisions: []
risks:
  - RISK-W07-E01-001
---

# W07-E01-S003 — Sweeper and worker materialization — bounded batches, leased outbox, N+1 removal

## Story ID

W07-E01-S003

## Title

Sweeper and worker materialization — bounded batches, leased outbox, N+1 removal

## Objective

Remove N+1 and unbounded materialization from `SweepSLA` and the reminder query (T1, T2, T3); batch-load
webhook-retry endpoints (T4); rework outbox claim/dispatch into a leased state machine consuming W04's
own DATA-02/DATA-03 lease primitives, preserving per-aggregate ordering (T5); add queue-lag/batch-
duration metrics (T6); build bounded-batch benchmarks with same-PR budget entries (T7); and publish
before/after evidence against `perf/reference-v1.json` (T8).

## Value to the framework

PLAN's own PERF-04 evidence, all 4 citations confirmed exactly: "`SweepSLA` loads ALL due rows unbounded
(no `LIMIT`), then does 1 UPDATE + 1 load + emit per row. The reminder query has no matching index
(`wft_due` only covers `due_at`, the query filters `remind_after`). `webhook.RetryOutbound` loads
endpoints per-delivery (bounded batch of 10, but still N queries). Outbox's outer claim transaction spans
the entire per-subscriber dispatch loop, including nested per-event tenant transactions." This story
converts a sweep/dispatch cost that scales with the number of due rows or subscribers into one bounded
by explicit batch size, and moves the outbox's own dispatch loop off a single long-held transaction —
directly preventing the pool-exhaustion-under-provider-latency failure mode T5's own hard dependency
(W04's DATA-02/DATA-03) was built to close.

## Problem statement

PLAN's own PERF-04 task table gives T5 (the outbox rework) an explicit, unusually strong dependency
statement: "**Hard dependency on PF-DATA's Wave-3 DATA-02/DATA-03 lease primitives — cannot start before
those land**" with risk classification "**High — cross-work-package dependency, do not attempt in
isolation, flag explicitly rather than re-deriving the lease design inside PF-PERF**." This story honors
that dependency exactly: W04 (this programme's own W04-E01/E02) has already built and accepted those
primitives by the time this wave begins (per this wave's own all-prior-waves entry gate), so T5 consumes
them, it does not re-derive fencing/lease logic from scratch inside PF-PERF's own scope.

## Source requirements

PERF-04 (T1–T8).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit): `SweepSLA` loads all
due rows with no `LIMIT`, then performs 1 UPDATE + 1 load + 1 emit per row sequentially. The reminder
query has no matching index — `wft_due` covers only `due_at`, but the query filters `remind_after`.
`webhook.RetryOutbound` loads endpoints per-delivery — bounded to a batch of 10, but still N queries, one
per delivery. Outbox's outer claim transaction spans the entire per-subscriber dispatch loop, including
nested per-event tenant transactions — a long-held transaction across potentially many network calls.

## Desired state

`SweepSLA`'s two queries both carry an explicit `LIMIT`, with the sweep looping via job re-invocation
rather than materializing the full due-row set in memory — fixed query count and memory across due-row
cardinalities (10/1k/100k). Per-row UPDATE+load+emit is converted to set-based/batched operations where
semantically possible (set-based UPDATE for guard flips; batch-load by ID set), preserving existing
idempotency guards with no reintroduced double-remind race. A partial index on `remind_after` matches
the query predicate, shown via `EXPLAIN` to use an index scan. `webhook.RetryOutbound` batch-loads
endpoints via one `IN (...)` query per invocation, not per-delivery. Outbox's claim/dispatch is reworked
into a leased state machine (consuming W04's DATA-02/DATA-03 primitives) with no outer transaction
spanning tenant handlers, preserving per-aggregate ordering, passing the inherited chaos tests from
DATA-02/DATA-03's own gate. Queue-lag and batch-duration metrics exist for sweeper/webhook/outbox
timing. Bounded-batch benchmarks exist at due-row cardinality tiers, with budget entries landed in the
same PR (per PERF-06's own fail-closed policy). Before/after evidence is published against
`perf/reference-v1.json`.

## Scope

- **T1** — Bounded batch claiming for `SweepSLA` (both queries get `LIMIT`, loop via job re-invocation).
- **T2** — Convert per-row UPDATE+load+emit to set-based/batched operations where semantically possible.
- **T3** — Partial index on `remind_after` matching the query predicate.
- **T4** — Batch-load endpoints in `RetryOutbound` (one `IN (...)` query per invocation, not per-row).
- **T5** — Rework outbox claim/dispatch into a leased state machine, preserving per-aggregate ordering —
  consuming W04's own DATA-02/DATA-03 lease primitives (hard dependency, already satisfied by this
  wave's own entry gate).
- **T6** — Queue lag and batch duration metrics for sweeper/webhook/outbox timing.
- **T7** — Bounded-batch benchmarks at due-row cardinality tiers, budget entries landed same-PR.
- **T8** — Publish before/after evidence against `perf/reference-v1.json`.

## Out of scope

- **DATA-02/DATA-03's own lease/fencing primitive design** — already built and accepted at W04-E01/E02;
  this story consumes those primitives for T5, it does not re-derive or modify their own design.
- **PERF-02's own reference-environment build** — W07-E01-S001's own scope; this story's T8 consumes
  that environment, it does not rebuild it.
- **Emit()/escalation logic's own per-instance nature** — PLAN's own T2 risk note acknowledges this is
  "inherently per-instance"; this story does not force emit/escalation into a set-based operation where
  the underlying semantics genuinely require per-instance handling.

## Assumptions

- W04-E01/E02's own DATA-02/DATA-03 lease primitives are assumed to directly fit T5's own consumption
  needs without requiring an adaptation layer — RISK-W07-E01-001 (epic-scoped) tracks the possibility
  this assumption proves wrong, requiring T5's own re-confirmation step before implementation begins.
- The exact set-based/batched conversion approach for T2 (which specific operations convert to
  set-based UPDATE, which remain per-instance for emit/escalation) is not fully specified by any source
  document beyond PLAN's own framing — this story's own T2 design work determines the exact split.

## Dependencies

**Hard dependency on W04-E01 (DATA-02 shared lease/fencing primitive) and W04-E02 (DATA-03 remote-I/O-
outside-tx primitives)** for T5 specifically — PLAN's own explicit framing: "cannot start before those
land." Both are satisfied by this wave's own all-prior-waves entry gate, since W04 is a prior wave. No
dependency within W07-E01 beyond T8's own consumption of W07-E01-S001's `perf/reference-v1.json`.

## Affected packages or components

`kernel/workflow` (or wherever `SweepSLA` lives); `kernel/webhook` (`RetryOutbound`); `kernel/outbox`
(the claim/dispatch rework); new database indexes (`remind_after` partial index).

## Compatibility considerations

T1/T2's own idempotency-guard-preservation requirement is a strict correctness constraint: the bounded-
batch conversion must not reintroduce a double-remind race that today's (inefficient but correct)
per-row loop avoids. T5's own per-aggregate-ordering-preservation requirement is the equivalent
constraint for the outbox rework.

## Security considerations

Not directly applicable beyond what W04's own DATA-02/DATA-03 primitives already establish for fencing/
lease correctness — this story consumes those guarantees, it does not weaken them.

## Performance considerations

This story IS the performance optimization; see "Objective" and "Desired state" above.

## Observability considerations

T6's own queue-lag/batch-duration metrics are this story's own primary observability addition.

## Migration considerations

T3's own partial-index addition is an additive schema change — PLAN's own T3 risk note explicitly frames
it as following "DATA-09's expand-only protocol since `workflow_tasks` is a live shared table," meaning
this story's own T3 task should use W02-E01's own online-migration protocol for the index addition, not
an ad hoc migration.

## Documentation requirements

Document the leased-state-machine outbox rework's own design (T5), so a future maintainer understands
how it relates to W04's own lease primitives; document the bounded-batch conversion's own idempotency-
preservation reasoning (T1/T2).

## Acceptance criteria

- **AC-W07-E01-S003-01**: `SweepSLA`'s both queries carry a fixed `LIMIT`; fixed query count and memory hold
  across due-row cardinalities (10/1k/100k), tested at each tier.
- **AC-W07-E01-S003-02**: Set-based UPDATE for guard flips and batch-load by ID set are used where semantically
  possible; existing idempotency guards are preserved with no reintroduced double-remind race.
- **AC-W07-E01-S003-03**: `EXPLAIN` shows index-scan access for the reminder query against the new partial
  index on `remind_after`.
- **AC-W07-E01-S003-04**: `RetryOutbound` issues one `IN (...)` query per invocation, not per-delivery, proven
  by a query-count assertion test (N rows / M endpoints → 1 query).
- **AC-W07-E01-S003-05**: The leased-state-machine outbox rework passes its inherited crash/duplicate-worker
  chaos tests from DATA-02/DATA-03's own gate; no outer transaction spans tenant handlers; per-aggregate
  ordering is preserved.
- **AC-W07-E01-S003-06**: Queue-lag and batch-duration metrics are emitted for sweeper/webhook/outbox timing.
- **AC-W07-E01-S003-07**: Bounded-batch benchmarks exist at due-row cardinality tiers with budget entries landed
  in the same PR; before/after evidence is published against `perf/reference-v1.json`.

## Required artifacts

- Bounded-batch `SweepSLA` code (T1).
- Set-based/batched operation conversions (T2).
- The `remind_after` partial index migration (T3).
- Batch-loaded `RetryOutbound` code (T4).
- The leased-state-machine outbox rework (T5).
- Queue-lag/batch-duration metric emission (T6).
- Bounded-batch benchmarks + budget entries (T7).
- The published before/after comparison (T8).
See `artifacts/index.md`.

## Required evidence

- Fixed-query-count/memory test output at 10/1k/100k due rows (T1).
- Query-count assertion test output confirming set-based conversion (T2).
- `EXPLAIN` output confirming index scan (T3).
- Query-count assertion test output for `RetryOutbound` (T4).
- Inherited chaos test output from DATA-02/DATA-03's own gate (T5).
- Metric-emission test output (T6).
- Bounded-batch benchmark output + budget-entry confirmation (T7).
- The published comparison report (T8).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all seven acceptance criteria numbered and measurable, T5's hard dependency on
W04-E01/E02 recorded explicitly, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all seven acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T5's own consumption of W04's lease primitives is
genuine (not a re-derived, parallel fencing mechanism) and that T1/T2's idempotency guards are genuinely
preserved, not merely asserted preserved.

## Risks

RISK-W07-E01-001 (T5 may discover W04's lease primitives require an adaptation layer beyond simple
consumption) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once T5's own re-confirmation step (checking the lease primitives' actual shape before implementation)
is honored and all seven acceptance criteria are verified, residual risk is expected to be low.

## Plan

See `plan.md`.
