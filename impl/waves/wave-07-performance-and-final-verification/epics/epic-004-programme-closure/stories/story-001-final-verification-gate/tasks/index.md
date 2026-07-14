---
id: W07-E04-S001-TASKS-INDEX
type: tasks-index
parent_story: W07-E04-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07-E04-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E04-S001-T001](task-001-rerun-capability-assessments.md) | Re-run the §H/§I-style capability assessments | unassigned | todo | W07-E01, E02, E03 accepted | Fresh capability reassessment report | AC-W07-E04-S001-01 | not started | not started |
| [W07-E04-S001-T002](task-002-traceability-completeness-check.md) | Traceability-completeness check | unassigned | todo | W07-E01, E02, E03 accepted | Every row confirmed to have a disposition | AC-W07-E04-S001-02 | not started | not started |
| [W07-E04-S001-T003](task-003-disposition-audit.md) | Disposition audit (sampled) | unassigned | todo | W07-E01, E02, E03 accepted | Sampled genuineness confirmation | AC-W07-E04-S001-03 | not started | not started |
| [W07-E04-S001-T004](task-004-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003 | Independent-review record per mandate §14 | AC-W07-E04-S001-01, AC-W07-E04-S001-02, AC-W07-E04-S001-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (capability reassessment), T002 (traceability completeness), and T003
(disposition audit) are kept as three separate tasks because each targets a genuinely distinct
verification surface with its own separately-evidenced output, and each may proceed in parallel once
this story's own entry gate is satisfied. T004 adds an independent-review task per mandate §14 — this
story is P0/critical (it is the programme's own terminal verification gate), and its own review task is
specifically scoped to catch RISK-W07-E04-001's restatement-not-re-verification failure mode, the single
highest-consequence way this story could fail silently.
