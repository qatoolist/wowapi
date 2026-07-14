---
id: W05-E02-S002-TASKS-INDEX
type: tasks-index
parent_story: W05-E02-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E02-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E02-S002-T001](task-001-zero-reflection-provider-graph.md) | Type-erased provider graph, zero-reflection proof | unassigned | todo | none (depends on S001) | Provider graph + benchmark + lint | AC-W05-E02-S002-01 | not started | not started |
| [W05-E02-S002-T002](task-002-boot-time-graph-validation.md) | Boot-time graph validation | unassigned | todo | T001 | Validation for 5 failure classes | AC-W05-E02-S002-02 | not started | not started |
| [W05-E02-S002-T003](task-003-three-profile-projection.md) | Three-profile projection compiler | unassigned | todo | T001, T002 | API/worker/migrate projections | AC-W05-E02-S002-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (graph + zero-reflection), T002 (validation, depends on T001's graph
existing), and T003 (projection, depends on both) are kept as three sequential tasks matching PLAN's
own T3→T4→T5 dependency chain. No independent-review task — PLAN's own risk column reads Medium for
all three, materially lower than S001's High-risk T2, and each task's correctness is proven by its
own dedicated benchmark/lint/adversarial-suite mechanism.
