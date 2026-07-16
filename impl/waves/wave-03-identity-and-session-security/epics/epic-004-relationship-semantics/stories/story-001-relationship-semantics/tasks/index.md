---
id: W03-E04-S001-TASKS-INDEX
type: tasks-index
parent_story: W03-E04-S001
status: planned
derived: false
created_at: 2026-07-13
updated_at: 2026-07-13
---

# W03-E04-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation — each
file below contains its task definition, implementation record, verification record, and deviations
record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W03-E04-S001-T001](task-001-party-subject-evaluation.md) | Checker.Has party-subject evaluation (DATA-07 T1) | unassigned | done | none | `Checker.Has` correctly evaluates party-subject edges via the post-SEC-01 pri... | AC-W03-E04-S001-01 | implemented | verified |
| [W03-E04-S001-T002](task-002-subject-kind-matrix.md) | Checker.Has full subject-kind matrix (DATA-07 T2) | unassigned | done | W03-E04-S001-T001 | Every schema-enumerated, live-requirement `subject_kind` has an explicit eval... | AC-W03-E04-S001-02 | implemented | verified |
| [W03-E04-S001-T003](task-003-mutation-governance.md) | Mutation governance - ownership, attribution, audit, versioning (DATA-07 T4) | unassigned | done | W03-E04-S001-T001, W03-E04-S001-T002 | Every relationship-edge create/revoke mutation is ownership-checked, attribut... | AC-W03-E04-S001-03 | implemented | verified |
| [W03-E04-S001-T004](task-004-independent-review.md) | Independent review | unassigned | done | W03-E04-S001-T001, W03-E04-S001-T002, W03-E04-S001-T003 | A completed review report confirming the checklist above, recorded as evidence. | AC-W03-E04-S001-01, AC-W03-E04-S001-02, AC-W03-E04-S001-03 | reviewed | verified |

## Grouping rationale

Per `plan.md`: tasks follow the PLAN task breakdown for this story (Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation governance). Each task is
tracked separately because it produces distinct output with separate evidence. The final task is an
independent-review task per mandate §14 for this P0/P1 security or governance story.
