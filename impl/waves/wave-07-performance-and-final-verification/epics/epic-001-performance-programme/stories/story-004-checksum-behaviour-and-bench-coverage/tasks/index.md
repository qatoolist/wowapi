---
id: W07-E01-S004-TASKS-INDEX
type: tasks-index
parent_story: W07-E01-S004
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S004 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E01-S004-T001](task-001-required-checksum-enforcement.md) | Required checksum enforcement and call-site audit | W07-Scoping-Dispatch.W07E01S004 | complete | none | No body download on normal Stat | AC-W07-E01-S004-01 | complete | passed |
| [W07-E01-S004-T002](task-002-bounded-repair-path.md) | Bounded repair path | W07-Scoping-Dispatch.W07E01S004 | complete | T001 | Fallback reachable only via labeled path | AC-W07-E01-S004-02 | complete | passed |
| [W07-E01-S004-T003](task-003-fallback-metrics.md) | Fallback-invocation metrics | W07-Scoping-Dispatch.W07E01S004 | complete | T002 | Dedicated fallback metrics | AC-W07-E01-S004-03 | complete | passed |
| [W07-E01-S004-T004](task-004-resumable-backfill.md) | Resumable async backfill | W07-Scoping-Dispatch.W07E01S004 | complete | T002 | Interrupt/resume-safe legacy backfill | AC-W07-E01-S004-04 | complete | passed |
| [W07-E01-S004-T005](task-005-publication.md) | Publication against perf/reference-v1.json | W07-Scoping-Dispatch.W07E01S004 | complete | T001-T004; cross-story W07-E01-S001-T001 | Published before/after comparison | AC-W07-E01-S004-05 | complete | passed |
| [W07-E01-S004-T006](task-006-cs16-bench-coverage-expansion.md) | CS-16 bench-coverage expansion (7 packages) | W07-Scoping-Dispatch.W07E01S004 | complete | none | 7 new benchmarks + budget entries | AC-W07-E01-S004-06, AC-W07-E01-S004-07 | complete | passed |
| [W07-E01-S004-T007](task-007-independent-review.md) | Independent review | W07-Scoping-Dispatch.W07E01S004ReviewR | complete | T001-T006 | Independent-review record per mandate §14 | AC-W07-E01-S004-01 .. AC-W07-E01-S004-07 | complete | passed |

## Grouping rationale

Per mandate §12: T001-T005 follow PLAN PERF-05's own T1-T5 task table exactly. T006 folds in
MATRIX CS-16's own 7-package bench-coverage expansion as a single task rather than 7 separate tasks,
because all 7 share the identical mechanical shape (add one benchmark + one budget entry per package)
and the identical acceptance bar (make bench-budget passes) — splitting into 7 tasks would multiply file
count with no added tracking value, consistent with mandate §12's own "avoid excessive fragmentation"
instruction, distinct from cases elsewhere in this programme where genuinely disjoint risk profiles
justify separate tasks. T006 is independent of T001-T005 (disjoint code surface: storage vs. 7 kernel
packages) and may proceed in parallel. T007 adds an independent-review task per mandate §14, given the
easy-to-overclaim nature of "this benchmark targets the named hot path" without independent
verification.
