---
id: W06-E03-S002-TASKS-INDEX
type: tasks-index
parent_story: W06-E03-S002
status: blocked
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E03-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E03-S002-T001](task-001-repo-admin-activation.md) | Repo-admin activation | repo-administrator | blocked | DEC-Q10 | Active branch/environment/tag protection | AC-01 | readiness authored; live controls absent | failed/blocking probe |
| [W06-E03-S002-T002](task-002-post-activation-verification.md) | Post-activation verification | release/security lead | blocked | T001 | Real protected release proof | AC-02 | authored | not executable |

## Grouping rationale

Per mandate §12: T001 (the human-only activation itself) and T002 (the post-activation
verification, performable by any programme worker) are kept as two separate tasks because they have
genuinely different actors and genuinely different completion criteria — T001 can only be completed by a
human with repo-admin access and has no code-level verification of its own beyond live API confirmation;
T002 is a standard verification task that any worker can perform once T001's precondition is satisfied.
No independent-review task (a T003, per the pattern used elsewhere for P0/critical stories) is added
here: this story's own DoD already requires independent review before `accepted` per
`governance/definition-of-done.md`, and per mandate §14's own framing, adding a third task purely to
re-state that requirement would not add tracking value beyond what T002's own verification and this
story's closure-time review already provide for a two-task story this small.
