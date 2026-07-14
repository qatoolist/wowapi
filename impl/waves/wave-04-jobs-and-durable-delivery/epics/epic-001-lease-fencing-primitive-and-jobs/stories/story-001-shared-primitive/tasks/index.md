---
id: W04-E01-S001-TASKS-INDEX
type: tasks-index
parent_story: W04-E01-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E01-S001-T001](task-001-primitive-design-and-cross-consumer-review.md) | Shared primitive design, implementation, and cross-consumer field-set review | unassigned | todo | none | Lease/fencing primitive package + unit tests + cross-consumer review record | AC-W04-E01-S001-01, AC-W04-E01-S001-02 | not started | not started |
| [W04-E01-S001-T002](task-002-interim-checkpoint-lease-migration.md) | Interim-checkpoint-lease migration | unassigned | todo | T001 | Migration tooling + test proving no checkpoint-state loss/duplication | AC-W04-E01-S001-03 | not started | not started |
| [W04-E01-S001-T003](task-003-independent-review.md) | Independent review | unassigned | todo | T001, T002 | Independent-review record per mandate §14 | AC-W04-E01-S001-01, AC-W04-E01-S001-02, AC-W04-E01-S001-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (primitive design, implementation, and cross-consumer review) and T002
(interim-lease migration) are kept separate because they produce unrelated outputs with separate
evidence — T001's evidence is a unit-test report plus a cross-consumer field-set review record; T002's
evidence is a migration test proving no checkpoint-state loss or duplication across the cutover. They
also carry materially different risks (T001's risk is RISK-W04-E01-001, design under-specification
across three consumer epics; T002's risk is RISK-W04-001, the migration-correctness risk on the
receiving side). T002 depends on T001 because the migration re-expresses interim-lease state under
the primitive's own schema, which must exist first. This story is P0 (DATA-02 as a whole is P0, and
this story is the epic's and wave's keystone, gating S002/S003 and both sibling epics W04-E02/
W04-E03) per this wave's task brief, so T003 adds an independent-review task per mandate §14, scoped
to confirming both T001's cross-consumer review and T002's interim-lease migration were genuinely
executed, not merely implemented in code without the required review/migration step. No separate
evidence-collection task is added — T001 and T002's own evidence (the unit-test report, the review
record, the migration-test report) is already a consolidated, story-scope-sized record; a fourth
aggregation task would add no tracking value for a story this small (2 substantive tasks).
