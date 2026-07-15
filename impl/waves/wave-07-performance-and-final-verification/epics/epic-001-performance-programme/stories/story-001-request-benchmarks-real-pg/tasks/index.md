---
id: W07-E01-S001-TASKS-INDEX
type: tasks-index
parent_story: W07-E01-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E01-S001-T001](task-001-reference-runner-and-skeleton.md) | Reference runner + perf/reference-v1.json skeleton | W07-Phase-A-Execution.W07E01S001 | complete | none | Reference runner + complete skeleton | AC-W07-E01-S001-01 | implemented | verified |
| [W07-E01-S001-T002](task-002-db-backed-benchmarks.md) | DB-backed benchmarks, all 6 profiles | W07-Phase-A-Execution.W07E01S001 | complete | T001 | Real-Postgres benchmark suite | AC-W07-E01-S001-02 | implemented | verified |
| [W07-E01-S001-T003](task-003-concurrency-matrix.md) | Concurrency-matrix variants | W07-Phase-A-Execution.W07E01S001 | complete | T002 | Cold/warm × 1/10/100-tenant matrix | AC-W07-E01-S001-03 | implemented | verified |
| [W07-E01-S001-T004](task-004-cost-breakdown-attribution.md) | Cost-breakdown attribution | W07-Phase-A-Execution.W07E01S001 | complete | T002 | Per-component cost attribution | AC-W07-E01-S001-04 | implemented | verified |
| [W07-E01-S001-T005](task-005-publication-dec-q9-conditional.md) | Publication against perf/reference-v1.json, DEC-Q9-conditional | W07-Phase-A-Execution.W07E01S001 | complete | T001, T002, T003, T004 | Published comparison report | AC-W07-E01-S001-05 | implemented | verified |
| [W07-E01-S001-T006](task-006-independent-review.md) | Independent review | W05ReviewGateFinal | complete | T001-T005 | Independent-review record per mandate §14 | AC-W07-E01-S001-01 .. AC-W07-E01-S001-05 | review-only | PASS, no open issues |

## Grouping rationale

Per mandate §12: T001-T005 follow PLAN PERF-02's own T1-T5 task table exactly, in the same
dependency order. T001 is the epic's own shared prerequisite, explicitly flagged "High risk — new CI
infrastructure, no owner/timeline established" by PLAN itself, warranting its own dedicated task rather
than being folded into T002. T006 adds an independent-review task per mandate §14, specifically to
re-check the DEC-Q9-conditionality framing this whole story's own acceptance bar depends on getting
right.
