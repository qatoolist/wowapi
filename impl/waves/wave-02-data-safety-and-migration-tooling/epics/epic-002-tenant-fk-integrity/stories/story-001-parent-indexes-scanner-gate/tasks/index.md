---
id: W02-E02-S001-TASKS-INDEX
type: tasks-index
parent_story: W02-E02-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E02-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W02-E02-S001-T001](task-001-parent-tenant-unique-indexes.md) | Parent tenant-scoped unique indexes | unassigned | todo | none | `UNIQUE (tenant_id, id)` on `parties`, `organizations`, `documents`, `document_versions` | AC-W02-E02-S001-01 | not started | not started |
| [W02-E02-S001-T002](task-002-tenant-fk-catalog-scanner.md) | Tenant-FK catalog scanner | unassigned | todo | none | Scanner enumerating all tenant-table FKs, flagging non-composite ones | AC-W02-E02-S001-02 | not started | not started |
| [W02-E02-S001-T003](task-003-ci-gate-wiring.md) | CI gate wiring | unassigned | todo | T002 | Scanner wired as permanent CI gate + negative fixture migration | AC-W02-E02-S001-03 | not started | not started |
| [W02-E02-S001-T004](task-004-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003 | Independent-review record per mandate §14 | AC-W02-E02-S001-01, AC-W02-E02-S001-02, AC-W02-E02-S001-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (parent unique indexes) and T002 (catalog scanner) are kept separate because
they produce unrelated outputs with separate evidence — T001's evidence is a `pg_indexes` migration
test; T002's evidence is a fixture-schema enumeration test. Both are parallel-safe (disjoint code
surface, no dependency between them per PLAN's own Depends-on column: T2 depends only on T1's
*existence* as a source concept, not on T1's own migrations having landed first — PLAN's Depends-on
column lists T2 as depending on T1, so this task index preserves that ordering as a recommended
sequencing rather than a hard code-level blocker; T003 is kept separate from T002 because it has its
own distinct output (a CI-wired gate plus a negative fixture) and its own distinct risk (a
false-positive CI rejection blocking a legitimate migration) from T002's own risk (an incomplete FK
enumeration). This story is P0 (DATA-01 as a whole is P0) per this wave's task brief, so T004 adds an
independent-review task per mandate §14, scoped to confirming the scanner's enumeration has zero
silent gaps and the CI gate genuinely rejects a non-composite tenant FK, not merely claims to. No
separate evidence-collection task is added — T001/T002/T003's own evidence (the `pg_indexes` report,
the fixture-schema test, the negative-fixture CI run) is already a consolidated, story-scope-sized
record; a fifth aggregation task would add no tracking value for a story this small (3 substantive
tasks).
