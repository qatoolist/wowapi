---
id: TASKS-INDEX-W00-E01-S002
type: task-index
parent_story: W00-E01-S002
status: recorded
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Task index — W00-E01-S002

Per mandate §16.4. This index is a roll-up view; the canonical status for each task lives in that
task file's front matter.

| Task ID | Title | Owner | Status | Dependencies | Output | Related ACs | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| W00-E01-S002-T001 | Re-verify PERF-01: token-bucket sweep fix | W00E01S002 (wave-00 verification worker) | done | none | Both commands exit 0 at `0a31186`; evidence EV-W00-E01-S002-01 registered | AC-W00-E01-S002-01 | executed 2026-07-13 | pass |
| W00-E01-S002-T002 | Re-verify PERF-06 T1: fail-closed benchbudget missing-benchmark gate | W00E01S002 (wave-00 verification worker) | done | none | `TestMainMissingBenchmarkFails` PASS + fail-first ghost-entry check (gate exit 1) at `0a31186`; evidence EV-W00-E01-S002-02 registered | AC-W00-E01-S002-02 | executed 2026-07-13 | pass |
| W00-E01-S002-T003 | Confirm SD-03: #25 bench-budget recalibration reflected in `bench-budgets.txt` | W00E01S002 (wave-00 verification worker) | done | none | 43 entries confirmed (both counting methods); file byte-identical to `0a31186`; evidence EV-W00-E01-S002-03 registered | AC-W00-E01-S002-03 | executed 2026-07-13 | pass |
