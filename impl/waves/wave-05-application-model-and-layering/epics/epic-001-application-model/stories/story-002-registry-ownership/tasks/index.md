---
id: W05-E01-S002-TASKS-INDEX
type: tasks-index
parent_story: W05-E01-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1").

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E01-S002-T001](task-001-resource-and-rules-ownership-wrappers.md) | resource.Registry and rules.Registry owner-bound wrappers | unassigned | todo | none (depends on S001) | Two owner-bound registry wrappers + adversarial tests | AC-W05-E01-S002-01 | not started | not started |
| [W05-E01-S002-T002](task-002-authz-permission-ownership-wrapper.md) | authz.Registry permission-registration owner-bound wrapper | unassigned | todo | none (depends on S001) | Owner-bound authz permission-registration API + adversarial test | AC-W05-E01-S002-02 | not started | not started |
| [W05-E01-S002-T003](task-003-remaining-declaration-classes.md) | Owner-bound wrappers for remaining ~9+ declaration classes | unassigned | todo | none (depends on S001) | Wrappers for all remaining classes + table-driven adversarial suite | AC-W05-E01-S002-03 | not started | not started |
| [W05-E01-S002-T004](task-004-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003 | Independent-review record per mandate §14 | AC-W05-E01-S002-01, AC-W05-E01-S002-02, AC-W05-E01-S002-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (resource + rules, "same shape as T3" per PLAN T4's own wording) are grouped
as the reference-pattern pair; T002 (authz permission registration) is kept separate given its
materially different risk profile (PLAN's own "High — only registry with zero existing ownership
check," an API-signature change rather than an additive wrapper); T003 (the remaining ~9+
declaration classes) is kept separate given its own distinct under-scoping risk and table-driven
test shape. This story is P1-core with T002 flagged High risk and T003 flagged an explicit
under-scoping risk, so T004 adds an independent-review task per mandate §14, scoped to confirming
both risks were genuinely closed, not merely implemented.
