---
id: W07-E02-S001-TASKS-INDEX
type: tasks-index
parent_story: W07-E02-S001
status: blocked
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E02-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E02-S001-T001](task-001-build-control-map.md) | Build the version-pinned control map | W07-Phase-A-Execution.W07E02S001 | blocked | SEC-01/03/04/06 accepted-state check fails | Complete 412-entry control map + validator | AC-W07-E02-S001-01 | implemented | focused pass; hard dependency + clean-commit retest pending |
| [W07-E02-S001-T002](task-002-commission-external-assessment.md) | Commission and record the external assessment | product-security lead | blocked | T001 + external professional-services engagement | Report not produced; exact blocker record produced | AC-W07-E02-S001-02 | not implemented | blocked/fail; no report |

## Grouping rationale

Per mandate §12: T001 (build the control map) and T002 (commission and record the external
assessment) are kept as two tasks because they have genuinely different actors and timelines — T001 is
performable by this programme's own workers; T002 depends on an external party's own engagement
timeline, outside this programme's direct control. No independent-review task is added: this story's
own DoD already requires independent review before `accepted`, and this story's entire purpose (an
external assessment) already provides an independent-verification mechanism stronger than an internal
review task would add on top of it for a 2-task story this size.
