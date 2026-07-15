---
id: W07-E02-S002-TASKS-INDEX
type: tasks-index
parent_story: W07-E02-S002
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E02-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E02-S002-T001](task-001-fail-not-skip-e2e.md) | Fail-not-skip E2E prerequisites | W07-E02-S002 executor | done | none | Non-zero exit on unmet prerequisite | AC-W07-E02-S002-01 | implemented | verified (EV-001) |
| [W07-E02-S002-T002](task-002-machine-checked-skip-manifest.md) | Machine-checked skip manifest | W07-E02-S002 executor | done | T001 | Skip manifest enforcing approval | AC-W07-E02-S002-02 | implemented | verified (EV-002) |
| [W07-E02-S002-T003](task-003-race-tests-integration.md) | Race tests over integration-relevant packages | W07-E02-S002 executor | done | none | -race catching a seeded data race in CI | AC-W07-E02-S002-03 | implemented | verified (EV-003) |
| [W07-E02-S002-T004](task-004-real-fuzz-owns-perf06.md) | Real time-bounded coverage-guided fuzzing (owns PERF-06 T3/T4) | W07-E02-S002 executor | done | none | Single-owned real-fuzz wiring, PR + scheduled | AC-W07-E02-S002-04 | implemented | verified (EV-004) |
| [W07-E02-S002-T005](task-005-independent-review.md) | Independent review | W05ReviewGateFinal | done | T001-T004 | Independent-review record per mandate §14 | AC-W07-E02-S002-01 .. AC-W07-E02-S002-04 | review-only | PASS; no open issues |

## Grouping rationale

Per mandate §12: T001-T004 follow PLAN REL-04's own T5-T8 task table exactly, kept as four
separate tasks because each targets a genuinely disjoint CI-truthfulness mechanism (E2E fail-closed,
skip manifest, race testing, fuzz testing) with its own separately-evidenced acceptance bar. T004 (the
real-fuzz task) is explicitly the single-ownership resolution point for CONFLICT-02 — its own task
description states this ownership explicitly rather than silently assuming a reader already knows. T005
adds an independent-review task per mandate §14, specifically scoped to re-verify T004's single-
ownership claim via an explicit repository-wide duplicate search, since a silently-duplicated
implementation would be a genuine programme-integrity defect (the exact class of waste CONFLICT-02 exists
to prevent).
