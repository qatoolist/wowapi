---
id: W06-E04-S002-TASKS-INDEX
type: tasks-index
parent_story: W06-E04-S002
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E04-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E04-S002-T001](task-001-generated-docs-byte-match.md) | Generated reference docs byte-matching AR-03's model export | W06E04Impl | complete with recorded entry deviation | delivered AR-03 export present; W05 lifecycle bookkeeping pending | Generated reference tables byte-matching the model export | AC-W06-E04-S002-01 | complete | PASS (EV-W06-E04-S002-001) |
| [W06-E04-S002-T002](task-002-future-state-labeling-lint.md) | Future-state-labeling lint | W06E04Impl | complete | none | Lint failing on unlabeled future-state blocks | AC-W06-E04-S002-02 | complete | PASS (EV-W06-E04-S002-002) |
| [W06-E04-S002-T003](task-003-independent-review.md) | Independent review | W06-E01-E04-Execution.W06E04ReviewR | complete | T001, T002 | Independent-review PASS, no issues | AC-W06-E04-S002-01, AC-W06-E04-S002-02 | complete | PASS (REV-W06-E04-S002-001) |

## Grouping rationale

Per mandate §12: T001 (AR-03-dependent) and T002 (independent) are kept as separate tasks precisely
because they have genuinely different entry criteria — collapsing them into one task would obscure that
T002 can proceed today while T001 cannot. T003 adds an independent-review task scoped specifically to
re-checking T001's entry criterion (if attempted) rather than merely re-running the standard mandate §14
checklist generically, mirroring the same pattern used in W06-E02-S003's own review task for the same
class of blocked-entry risk.
