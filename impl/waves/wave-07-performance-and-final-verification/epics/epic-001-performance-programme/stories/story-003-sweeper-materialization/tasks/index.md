---
id: W07-E01-S003-TASKS-INDEX
type: tasks-index
parent_story: W07-E01-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E01-S003-T001](task-001-bounded-batch-sweepsla.md) | Bounded batch claiming for SweepSLA | W07-Scoping-Dispatch.W07E01S003 | complete | none | Fixed query count/memory across cardinalities | AC-W07-E01-S003-01 | complete | accepted |
| [W07-E01-S003-T002](task-002-set-based-conversion.md) | Set-based/batched operation conversion | W07-Scoping-Dispatch.W07E01S003 | complete | T001 | Set-based operations and bounded loads | AC-W07-E01-S003-02 | complete | accepted |
| [W07-E01-S003-T003](task-003-remind-after-index.md) | Partial index on remind_after | W07-Scoping-Dispatch.W07E01S003 | complete | none | Real index-scan access confirmed | AC-W07-E01-S003-03 | complete | accepted |
| [W07-E01-S003-T004](task-004-batch-loaded-retryoutbound.md) | Batch-loaded RetryOutbound endpoints | W07-Scoping-Dispatch.W07E01S003 | complete | none | One `ANY(uuid[])` query per invocation | AC-W07-E01-S003-04 | complete | accepted |
| [W07-E01-S003-T005](task-005-leased-outbox-rework.md) | Leased-state-machine outbox rework | W07-Scoping-Dispatch.W07E01S003 | complete | W04-E01, W04-E02 accepted | Leased outbox passes inherited chaos/ordering | AC-W07-E01-S003-05 | complete | accepted |
| [W07-E01-S003-T006](task-006-queue-lag-metrics.md) | Queue-lag and batch-duration metrics | W07-Scoping-Dispatch.W07E01S003 | complete | T001, T005 | Bounded metrics for all three workers | AC-W07-E01-S003-06 | complete | accepted |
| [W07-E01-S003-T007](task-007-bounded-batch-benchmarks.md) | Bounded-batch benchmarks and budget entries | W07-Scoping-Dispatch.W07E01S003 | complete | T001, T002, T003 | Benchmarks with same-change budgets | AC-W07-E01-S003-07 | complete | accepted (relative; absolute conditional) |
| [W07-E01-S003-T008](task-008-publication.md) | Publication against perf/reference-v1.json | W07-Scoping-Dispatch.W07E01S003 | complete | T006, T007; W07-E01-S001-T001 | Truthful relative before/after comparison | AC-W07-E01-S003-07 | complete | accepted |
| [W07-E01-S003-T009](task-009-independent-review.md) | Independent review | W07-Scoping-Dispatch.W07E01S003ReviewR | complete | T001-T008 | Independent-review record per mandate §14 | AC-W07-E01-S003-01 .. AC-W07-E01-S003-07 | not applicable | PASS: no open issues |

## Grouping rationale

Per mandate §12: T001-T008 follow PLAN PERF-04's own T1-T8 task table exactly. T005 (the leased-
outbox rework) is PLAN's own explicitly-flagged highest-risk task ("High — cross-work-package
dependency, do not attempt in isolation") — kept as its own dedicated task rather than folded into T004
or T006, consistent with PLAN's own instruction to flag it explicitly. T009 adds an independent-review
task per mandate §14, specifically to re-check T005's genuine consumption of W04's own primitives (as
opposed to a parallel re-derivation) and T001/T002's genuine idempotency-guard preservation — both are
exactly the class of claim that is easy to assert and hard to verify without dedicated review.
