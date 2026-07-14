---
id: W03-E01-S002-TASKS-INDEX
type: tasks-index
parent_story: W03-E01-S002
status: complete
created_at: 2026-07-13
updated_at: 2026-07-13
---

# W03-E01-S002 — Tasks index

Per mandate §16.4.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W03-E01-S002-T001](task-001-capacity-selection.md) | Capacity-selection enforcement (SEC-01 T4) | unassigned | complete | none | Capacity-selection enforcement logic; a passing multi-capacity test covering no-choice, valid-choice, and unentitled-assertion cases | AC-W03-E01-S002-01 | complete | pass |
| [W03-E01-S002-T002](task-002-privileged-session-resolver.md) | Privileged-session resolver (SEC-01 T5) | unassigned | complete | none | The privileged-session resolver, wired into `Verifier.Actor`; a passing adversarial test suite covering expired/revoked/wrong-tenant/wrong-actor/forged-ID/unauthorized-approver | AC-W03-E01-S002-02 | complete | pass |
| [W03-E01-S002-T003](task-003-independent-review.md) | Independent review | unassigned | pending | W03-E01-S002-T001, W03-E01-S002-T002 | A completed review report confirming the checklist above. | AC-W03-E01-S002-01, AC-W03-E01-S002-02 | not started | pending |

## Grouping rationale

Per `plan.md`: tasks follow the PLAN task breakdown for this story (Capacity selection and privileged-session resolver). Each task is tracked separately because it produces distinct output with separate evidence. The final task is an independent-review task per mandate §14 for this P0/P1 security or governance story.
