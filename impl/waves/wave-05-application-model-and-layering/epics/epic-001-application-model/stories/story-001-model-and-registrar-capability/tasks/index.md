---
id: W05-E01-S001-TASKS-INDEX
type: tasks-index
parent_story: W05-E01-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E01-S001-T001](task-001-applicationmodel-lifecycle-skeleton.md) | ApplicationModel/Compiler lifecycle skeleton and D-03 error/panic behavior | unassigned | todo | none | Lifecycle skeleton + state-machine tests + build-tag matrix test | AC-W05-E01-S001-01, AC-W05-E01-S001-02 | not started | not started |
| [W05-E01-S001-T002](task-002-registrar-capability-type.md) | Registrar capability type and typed-key mechanism per D-02 | unassigned | todo | none | Registrar type + typed-key mechanism + compile-fail fixture | AC-W05-E01-S001-03 | not started | not started |
| [W05-E01-S001-T003](task-003-independent-review.md) | Independent review | unassigned | todo | T001, T002 | Independent-review record per mandate §14 | AC-W05-E01-S001-01, AC-W05-E01-S001-02, AC-W05-E01-S001-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (lifecycle skeleton + D-03 behavior) and T002 (Registrar capability type +
D-02 typed-key mechanism + compile-fail fixture) are kept separate because they produce unrelated
outputs with separate evidence and separate ratified decisions to enact (T001 enacts D-03; T002
enacts D-02), matching PLAN AR-01's own T1/T2 task split. This story is P1-core and T2 is flagged
High risk ("the actual security boundary") per PLAN's own risk column, so T003 adds an
independent-review task per mandate §14, scoped to confirming both T001's D-03 build-tag split and
T002's compile-fail fixture were genuinely proven, not merely implemented in code without the
required negative-compilation proof. No separate evidence-collection task is added — T001 and T002's
own evidence (the transition tests, the build-tag matrix, the compile-fail fixture output) is
already a consolidated, story-scope-sized record.
