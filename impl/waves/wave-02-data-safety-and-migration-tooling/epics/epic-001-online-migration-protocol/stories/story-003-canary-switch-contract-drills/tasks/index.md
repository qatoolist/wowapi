---
id: W02-E01-S003-TASKS-INDEX
type: tasks-index
parent_story: W02-E01-S003
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1").

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W02-E01-S003-T001](task-001-canary-deploy-n-tooling.md) | Canary/deploy-N tooling | W02ProtoRerun | done | W02-E01-S002 | Canary tooling + soak metrics + named canary test (both legs) | AC-W02-E01-S003-01 | implemented | verified |
| [W02-E01-S003-T002](task-002-switch-phase-tooling.md) | Switch-phase tooling | W02ProtoRerun | done | T001 | Compatibility flag + dual-version support + named switch-rollback test | AC-W02-E01-S003-02 | implemented | verified |
| [W02-E01-S003-T003](task-003-contract-phase-gate.md) | Contract-phase gate | W02ProtoRerun | done | T002 | Evidenced-precondition gate + named contract-gate test (both properties) | AC-W02-E01-S003-03 | implemented | verified |
| [W02-E01-S003-T004](task-004-ci-drill-pipeline.md) | CI drill pipeline | W02ProtoRerun | done | T001, T002, T003 | Scheduled pipeline running all 6 directive-named drills + passing run artifact | AC-W02-E01-S003-04 | implemented | verified |
| [W02-E01-S003-T005](task-005-evidence-aggregation.md) | Evidence aggregation (consolidated 6-drill bundle) | W02ProtoRerun | done | T004 | Consolidated evidence bundle registered in `evidence/index.md` | AC-W02-E01-S003-04 | implemented | verified |
| [W02-E01-S003-T006](task-006-independent-review.md) | Independent review | W02ProtoReview | done | T001–T005 | Independent-review record per mandate §14 | AC-W02-E01-S003-01 through -04 | reviewed | verified |

## Grouping rationale

Per mandate §12: T001/T002/T003/T004 map one-to-one onto PLAN DATA-09's own T6/T7/T8/T9 rows, which
form a strict dependency chain (T7 depends on T6, T8 on T7, T9 on T1–T8) and each carry their own
explicitly-named test ("This is the test" in each row's Tests column) — separate outputs, separate
evidence, materially different risks (T6's soak-calibration gap; T7's "core safety property"; T8's
"most safety-critical piece"; T9's "largest single infra investment"), so no merging is warranted.
T005 (evidence aggregation) is added by this story's own judgment, per the wave-planning brief's
instruction to add an evidence-collection task where the T-rows don't naturally produce a
consolidated evidence artifact: T6/T7/T8 each produce an individual named-drill test output and T9
produces a pipeline run artifact, but nothing in the four T-rows owns assembling those into the one
consolidated 6-drill evidence bundle this story's AC-W02-E01-S003-04 and closure require — unlike
W01-E01-S001 (where the linter run itself was the complete evidence, so no aggregation task was
added there), this story's evidence is spread across four producers and a genuine aggregation step
exists. T005's own Task Definition records this reasoning. This story is P0, so T006 adds an
independent-review task per mandate §14, with specific attention to the soak-threshold judgment gap
being honestly recorded (epic-level AC-W02-E01-04's story-specific review focus).
