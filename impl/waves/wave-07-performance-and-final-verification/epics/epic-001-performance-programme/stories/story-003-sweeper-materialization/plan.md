---
id: PLAN-W07-E01-S003
type: plan
parent_story: W07-E01-S003
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Plan — W07-E01-S003

Per mandate §8.5. T5's own implementation strategy is written to explicitly consume W04's own
already-accepted DATA-02/DATA-03 primitives, per PLAN's own "do not attempt in isolation" instruction.
Confirmed facts, planned changes, and assumptions are distinguished explicitly below.

## Proposed architecture

Bounded-batch query patterns for `SweepSLA` and `RetryOutbound`; a partial index on `remind_after`; a
leased-state-machine rework of outbox claim/dispatch, built directly on W04-E01/E02's own DATA-02/DATA-03
primitives rather than a parallel fencing mechanism; queue-lag/batch-duration metrics; bounded-batch
benchmarks.

## Implementation strategy

1. Add `LIMIT` to both `SweepSLA` queries; implement looping via job re-invocation rather than in-memory
   materialization of the full due-row set (T1).
2. Convert per-row UPDATE+load+emit to set-based/batched operations where semantically possible,
   preserving idempotency guards (T2).
3. Add a partial index on `remind_after`, following DATA-09's expand-only protocol (T3).
4. Batch-load endpoints in `RetryOutbound` via one `IN (...)` query per invocation (T4).
5. Re-confirm W04-E01/E02's DATA-02/DATA-03 lease primitives' actual shape before beginning T5's own
   implementation (per RISK-W07-E01-001's own mitigation); rework outbox claim/dispatch into a leased
   state machine consuming those primitives, preserving per-aggregate ordering (T5).
6. Add queue-lag/batch-duration metrics for sweeper/webhook/outbox timing (T6).
7. Build bounded-batch benchmarks at due-row cardinality tiers, landing budget entries in the same PR
   (T7).
8. Publish before/after evidence against `perf/reference-v1.json` (T8).

## Expected package or module changes

`kernel/workflow` (or wherever `SweepSLA` lives); `kernel/webhook`; `kernel/outbox`; new database
migration for the `remind_after` partial index.

## Expected file changes where determinable

- The `SweepSLA` implementation file(s) — bounded-batch conversion.
- The reminder-query implementation and its new partial index migration.
- `webhook.RetryOutbound`'s implementation file — batch-load conversion.
- `kernel/outbox`'s claim/dispatch implementation — leased-state-machine rework.
- New metric-emission code for T6.
- New benchmark files and `bench-budgets.txt` entries for T7.

## Contracts and interfaces

T5's own leased-state-machine rework consumes W04's own DATA-02/DATA-03 lease/fencing interface directly
— no new fencing contract is defined here.

## Data structures

None new beyond the `remind_after` partial index (T3).

## APIs

None affected — these are internal job/worker mechanisms, not a public API surface.

## Configuration changes

Possible new configuration for bounded-batch sizes (T1, T4) — exact keys TBD at implementation time.

## Persistence changes

The `remind_after` partial index (T3), following DATA-09's own online-migration protocol since
`workflow_tasks` is a live shared table.

## Migration strategy

T3's index addition uses `CREATE INDEX CONCURRENTLY` via DATA-09's own expand-only protocol (W02-E01),
consistent with this programme's own established pattern for live-table schema changes.

## Concurrency implications

T5's own leased-state-machine rework directly concerns concurrency — it must correctly handle concurrent
outbox dispatch attempts using W04's own fencing guarantees, not a parallel, independently-derived
mechanism.

## Error-handling strategy

T1/T2's own bounded-batch conversion must not silently drop a due row on a batch-boundary edge case — the
job re-invocation loop must guarantee eventual coverage of every due row.

## Security controls

None new beyond what W04's own DATA-02/DATA-03 primitives already establish.

## Observability changes

T6's own queue-lag/batch-duration metrics are the primary observability addition.

## Testing strategy

- T1: fixed-query-count/memory tests at 10/1k/100k due rows.
- T2: query-count assertion tests confirming set-based conversion, idempotency-guard preservation tests.
- T3: `EXPLAIN` plan tests confirming index scan.
- T4: query-count assertion test (N rows / M endpoints → 1 query).
- T5: inherited crash/duplicate-worker chaos tests from DATA-02/DATA-03's own gate.
- T6: metric-emission tests.
- T7: bounded-batch benchmarks at cardinality tiers, budget entries in the same PR.
- T8: before/after comparison against `perf/reference-v1.json`.

## Regression strategy

T1/T2's own idempotency-guard tests and T5's own inherited chaos tests become the ongoing regression
guard against a future change reintroducing a double-remind race or a fencing violation.

## Compatibility strategy

Not applicable beyond the idempotency-guard and per-aggregate-ordering preservation requirements already
stated.

## Rollout strategy

T1/T2/T3 may proceed in parallel (disjoint from T4/T5); T4 is independent; T5 begins only after
re-confirming W04's lease primitives' shape; T6 and T7 follow once T1-T5 exist; T8 publishes last.

## Rollback strategy

If T5's leased-state-machine rework proves incompatible with W04's own primitives in a way this story's
own re-confirmation step did not anticipate, halt T5 and escalate to the performance/SRE lead — do not
silently hand-roll a parallel fencing mechanism to work around the incompatibility.

## Implementation sequence

T1/T2/T3 (parallel) and T4 (independent) → T5 (after re-confirming W04's primitives) → T6, T7 → T8,
matching PLAN PERF-04's own dependency structure (T1 independent; T2 depends on T1; T3 independent; T4
independent; T5 hard-depends on DATA-02/DATA-03; T6 depends on T1-T5; T7 depends on T1-T3; T8 depends on
PERF-02 T1).

## Task breakdown

- **W07-E01-S003-T001** — Bounded batch claiming for `SweepSLA` (T1).
- **W07-E01-S003-T002** — Set-based/batched operation conversion (T2).
- **W07-E01-S003-T003** — Partial index on `remind_after` (T3).
- **W07-E01-S003-T004** — Batch-loaded `RetryOutbound` endpoints (T4).
- **W07-E01-S003-T005** — Leased-state-machine outbox rework (T5).
- **W07-E01-S003-T006** — Queue-lag/batch-duration metrics (T6).
- **W07-E01-S003-T007** — Bounded-batch benchmarks + budget entries (T7).
- **W07-E01-S003-T008** — Publication against `perf/reference-v1.json` (T8).
- **W07-E01-S003-T009** — Independent review.

## Expected artifacts

Bounded-batch `SweepSLA` code; set-based/batched operation conversions; the `remind_after` partial
index migration; batch-loaded `RetryOutbound` code; the leased-state-machine outbox rework; queue-lag/
batch-duration metric emission; bounded-batch benchmarks + budget entries; the published before/after
comparison.

## Expected evidence

Fixed-query-count/memory test output; query-count assertion test output (T2, T4); `EXPLAIN` output;
inherited chaos test output; metric-emission test output; bounded-batch benchmark output; the published
comparison report.

## Unresolved questions

- The exact bounded-batch size configuration for T1/T4 (hardcoded constant or config key) — to be
  decided at implementation time.
- Whether W04's own DATA-02/DATA-03 primitives require an adaptation layer for T5's own consumption
  needs — genuinely unknown until T5's own re-confirmation step is performed.

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned; T5's own
implementation must not begin until its own re-confirmation step (checking W04's lease primitives'
actual shape) is performed, per RISK-W07-E01-001's own mitigation.
