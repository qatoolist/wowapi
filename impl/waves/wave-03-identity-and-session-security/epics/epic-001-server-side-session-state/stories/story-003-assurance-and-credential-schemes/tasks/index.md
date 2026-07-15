---
id: W03-E01-S003-TASKS-INDEX
type: tasks-index
parent_story: W03-E01-S003
status: accepted
derived: false
created_at: 2026-07-13
updated_at: 2026-07-13
---

# W03-E01-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation — each
file below contains its task definition, implementation record, verification record, and deviations
record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W03-E01-S003-T001](task-001-assurance-freshness.md) | Assurance freshness and step-up enforcement (SEC-01 T6) | self | done | none | A stale `auth_time` with an otherwise-valid `amr` fails step-up. | AC-W03-E01-S003-01 | complete | verified |
| [W03-E01-S003-T002](task-002-credential-scheme-distinction.md) | Credential-scheme distinction (SEC-01 T7) | self | done | W03-E01-S003-T001 | A permission scoped to `CredentialUser` rejects a valid API-key actor; the me... | AC-W03-E01-S003-02 | complete | verified |
| [W03-E01-S003-T003](task-003-independent-review.md) | Independent review | self | done | W03-E01-S003-T001, W03-E01-S003-T002 | A completed review report confirming the checklist above, recorded as evidence. | AC-W03-E01-S003-01, AC-W03-E01-S003-02 | complete | verified |

## Grouping rationale

Per `plan.md`: tasks follow the PLAN task breakdown for this story (Assurance freshness and credential-scheme distinction). Each task is
tracked separately because it produces distinct output with separate evidence. The final task is an
independent-review task per mandate §14 for this P0/P1 security or governance story.
