---
id: W02-E01-S002-TASKS-INDEX
type: tasks-index
parent_story: W02-E01-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1").

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W02-E01-S002-T001](task-001-expand-phase-tooling.md) | Expand-phase tooling | W02ProtoRerun | done | W02-E01-S001 | Non-blocking DDL tooling + old-reader-compatibility test | AC-W02-E01-S002-01 | implemented | verified |
| [W02-E01-S002-T002](task-002-backfill-harness-and-interim-lease.md) | Backfill-job harness and interim checkpoint-lease mechanism | W02ProtoRerun | done | T001 | Resumable backfill harness + interim lease + interrupted/resumed test | AC-W02-E01-S002-02 | implemented | verified |
| [W02-E01-S002-T003](task-003-validation-phase-tooling.md) | Validation-phase tooling | W02ProtoRerun | done | T002 | VALIDATE CONSTRAINT orchestration + reconciliation + artifact schema | AC-W02-E01-S002-03 | implemented | verified |
| [W02-E01-S002-T004](task-004-independent-review.md) | Independent review | W02ProtoReview | done | T001, T002, T003 | Independent-review record per mandate §14 | AC-W02-E01-S002-01, AC-W02-E01-S002-02, AC-W02-E01-S002-03 | reviewed | verified |

## Grouping rationale

Per mandate §12: T001/T002/T003 map one-to-one onto PLAN DATA-09's own T3/T4/T5 rows, which already
form a strict dependency chain (T4 depends on T3; T5 depends on T4) and produce unrelated outputs
with separate named tests (old-reader-compatibility; the named interrupted/resumed backfill test;
the artifact-schema test) — no further splitting or merging is warranted. T002 carries this story's
(and this epic's) single largest risk, RISK-W02-001, because it is where the interim checkpoint-
lease's deliberately-bounded scope is designed and implemented; it is kept as one task (not split
further into "design the lease" and "build the harness") because the two are tightly coupled design
decisions that a single reviewer must evaluate together — splitting them would let the lease's scope
be reviewed in isolation from the harness behavior it is meant to support, weakening the review's
ability to catch a scope mismatch. This story is P0 (DATA-09 as a whole is P0), so T004 adds an
independent-review task per mandate §14, with specific attention to confirming the interim-lease
deviation is recorded honestly (per this story's `story.md` "Definition of done"). No separate
evidence-collection task is added — each of T001/T002/T003's own named test is already its
consolidated evidence, and this story's `evidence/index.md` aggregates all three without needing a
fourth collection step.
