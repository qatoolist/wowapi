---
id: TASKS-W00-E02-S001
type: task-index
parent_story: W00-E02-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Task index — W00-E02-S001

Per mandate §16.4. Hand-maintained; must never disagree with each task file's own front matter
(`impl/governance/status-model.md` "Canonical source of truth" — if it ever does, the task file
wins and this index is stale).

| Task ID | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W00-E02-S001-T001](task-001-coverage-baseline.md) | Coverage baseline | W00E02S001 (worker) | done | none | Coverage-baseline evidence record | AC-W00-E02-S001-01 | done (capture executed 2026-07-13) | pass — EV-W00-E02-S001-001 |
| [W00-E02-S001-T002](task-002-lint-baseline.md) | Lint baseline (25-analyzer, MATRIX CS-23 drift) | W00E02S001 (worker) | done | none | Lint-baseline evidence record, incl. analyzer-by-analyzer drift comparison | AC-W00-E02-S001-02 | done (capture executed 2026-07-13; drift flagged; DEV-W00-E02-S001-001 recorded) | pass — EV-W00-E02-S001-002 |
| [W00-E02-S001-T003](task-003-bench-and-ci-baseline.md) | Bench-budget and CI wall-clock baseline | W00E02S001 (worker) | done | none | Bench-budget-baseline and CI-wall-clock evidence records | AC-W00-E02-S001-03, AC-W00-E02-S001-04 | done (captures executed 2026-07-13) | pass — EV-W00-E02-S001-003 / EV-W00-E02-S001-004 |

## Notes

- No task depends on another within this story — T001, T002, T003 can run in any order or in
  parallel; each captures an independent baseline.
- T003 covers two acceptance criteria (AC-03 and AC-04) because bench-budget confirmation and
  CI-timing capture are grouped as one task — see `../plan.md` "Task breakdown" for the grouping
  rationale.
