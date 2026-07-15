---
id: W02-E01-S001-TASKS-INDEX
type: tasks-index
parent_story: W02-E01-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W02-E01-S001-T001](task-001-manifest-schema-design-and-ci-validation.md) | Manifest schema design, external review, and CI validation | W02ProtoRerun | done | none | Manifest schema definition + CI validator + negative fixture test | AC-W02-E01-S001-01, AC-W02-E01-S001-02 | implemented | verified |
| [W02-E01-S001-T002](task-002-lock-timeout-enforcement.md) | Lock-timeout enforcement mechanism | W02ProtoRerun | done | none | Lock-timeout wrapper with bounded abort-and-retry | AC-W02-E01-S001-03 | implemented | verified |
| [W02-E01-S001-T003](task-003-independent-review.md) | Independent review | W02ProtoReview | done | T001, T002 | Independent-review record per mandate §14 | AC-W02-E01-S001-01, AC-W02-E01-S001-02, AC-W02-E01-S001-03 | reviewed | verified |

## Grouping rationale

Per mandate §12: T001 (schema design + external review + CI validation) and T002 (lock-timeout
mechanism) are kept separate because they produce unrelated outputs with separate evidence — T001's
evidence is a schema-validation fixture pair plus an external-review record; T002's evidence is a
concurrency test against a deliberately locked table. They also carry materially different risks
(T001's risk is design under-specification, per RISK-W02-E01-002; T002's risk is the DoS-adjacent
unbounded-retry concern named explicitly in PLAN T2). This story is P0 (DATA-09 as a whole is P0,
and this story is the protocol's foundation, gating all of S002/S003) per this wave's task brief, so
T003 adds an independent-review task per mandate §14, scoped to confirming both T001's external-
review step and T002's bounded-retry control were genuinely satisfied, not merely implemented in
code without the required review/bound. No separate evidence-collection task is added — T001 and
T002's own evidence (the fixture pair, the review record, the lock-timeout test output) is already a
consolidated, story-scope-sized record; a fourth aggregation task would add no tracking value for a
story this small (2 substantive tasks).
