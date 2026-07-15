---
id: W07-E04-S002-TASKS-INDEX
type: tasks-index
parent_story: W07-E04-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07-E04-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E04-S002-T001](task-001-compile-closure-report.md) | Compile the programme closure report | unassigned | todo | W07-E04-S001 accepted | Complete programme closure report | AC-W07-E04-S002-01 | not started | not started |
| [W07-E04-S002-T002](task-002-compile-decision-package.md) | Compile the production-readiness claim-upgrade decision package | unassigned | todo | T001 | Separate decision package, no self-issued declaration | AC-W07-E04-S002-02 | not started | not started |
| [W07-E04-S002-T003](task-003-cross-check-open-items.md) | Cross-check both documents for open-item completeness | unassigned | todo | T001, T002 | Confirmed no open item silently dropped | AC-W07-E04-S002-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (closure report), T002 (decision package), and T003 (cross-check) are kept
as three separate tasks because they map directly onto this story's own core design decision — the
closure report and decision package must be genuinely separate documents (T001, T002), and their
consistency must be independently verified (T003), not assumed. Collapsing T003 into T002 would risk
the exact failure mode this story's whole structure exists to prevent: an open item quietly dropped
during compilation with no independent check to catch it. No further independent-review task (a T004)
is added beyond T003: T003 already performs the specific, targeted verification this story's own risk
profile calls for, and the standard closure-time review required by `governance/definition-of-done.md`
provides the remaining rigor for a 3-task story of this size.
