---
id: W03-E05-S001-TASKS-INDEX
type: tasks-index
parent_story: W03-E05-S001
status: done
derived: false
created_at: 2026-07-13
updated_at: 2026-07-13
---

# W03-E05-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation — each
file below contains its task definition, implementation record, verification record, and deviations
record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W03-E05-S001-T001](task-001-ratification-decision-and-implementation.md) | Ratification decision + implementation (SEC-02 T4) | unassigned | done | none | `RatifyBy` field added; non-empty `ratify_by` rejected at validation time; reject decision recorded | AC-W03-E05-S001-01 | done | done |
| [W03-E05-S001-T002](task-002-durable-override-audit.md) | Durable override audit (SEC-02 T5) | unassigned | done | W03-E05-S001-T001 | Every override produces a complete, transactional audit row; an injected audit-write failure rolls back the override | AC-W03-E05-S001-02 | done | done |
| [W03-E05-S001-T003](task-003-independent-review.md) | Independent review | unassigned | pending | W03-E05-S001-T001, W03-E05-S001-T002 | A completed review report confirming the checklist above, recorded as evidence. | AC-W03-E05-S001-01, AC-W03-E05-S001-02, AC-W03-E05-S001-03 | pending | pending |

## Grouping rationale

Per `plan.md`: tasks follow the PLAN task breakdown for this story (Workflow privileged completion — ratification and durable override audit). Each task is
tracked separately because it produces distinct output with separate evidence. The final task is an
independent-review task per mandate §14 for this P0/P1 security or governance story.
