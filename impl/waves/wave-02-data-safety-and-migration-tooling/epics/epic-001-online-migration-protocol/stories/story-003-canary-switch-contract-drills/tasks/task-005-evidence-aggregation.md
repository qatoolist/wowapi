---
id: W02-E01-S003-T005
type: task
title: Evidence aggregation (consolidated 6-drill bundle)
status: todo
parent_story: W02-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E01-S003-T004
acceptance_criteria:
  - AC-W02-E01-S003-04
artifacts:
  - ART-W02-E01-S003-005
evidence:
  - EV-W02-E01-S003-005
---

# W02-E01-S003-T005 — Evidence aggregation (consolidated 6-drill bundle)

## Task Definition

### Task objective

Consolidate T001/T002/T003's individual named-drill test outputs and T004's pipeline passing-run
artifact into one consolidated 6-drill evidence bundle, registered in `evidence/index.md` with
every mandate-§10 required field (commit SHA, execution command, environment, result) per
constituent drill.

**Why this task exists (recorded per the wave-planning brief's instruction to document the
reasoning when adding an evidence-collection task):** this story's evidence is produced by four
separate tasks — three individual named drills plus a pipeline run — and nothing in PLAN DATA-09's
own T6–T9 rows owns assembling them into the single consolidated record this story's
AC-W02-E01-S003-04, its `closure.md` evidence-completeness check, and the epic's AC-W02-E01-03 all
consume. This differs from cases like W01-E01-S001, where a single linter run was itself the
complete story evidence and no aggregation task was warranted — here a genuine, separately-produced
aggregation artifact exists, so it gets its own task per mandate §12 ("tasks must be decomposed
when they... need separate evidence").

### Parent story

W02-E01-S003 — Canary, switch, and contract-phase tooling with the full CI drill pipeline.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S003-T004 (aggregation requires all drill outputs and the pipeline artifact to exist).

### Detailed work

1. Collect the three named-drill test outputs (EV-W02-E01-S003-001/-002/-003) and the pipeline
   passing-run artifact (EV-W02-E01-S003-004).
2. Assemble the consolidated bundle: per-drill result, commit SHA, execution command, environment,
   and date, cross-referenced to the six directive-named drills confirmed in T004 step 1.
3. Register the bundle in `evidence/index.md` (EV-W02-E01-S003-005) and in this story's
   `artifacts/index.md` (ART-W02-E01-S003-005, lifecycle stage post-implementation).
4. Verify no constituent evidence item is missing a mandate-§10 required field — an item without a
   commit SHA must not be treated as final proof.

### Expected files or components affected

This story's `evidence/` and `artifacts/` directories (first real content triggers subdirectory
creation per naming-conventions Adaptation 2); no application code.

### Expected output

One consolidated, revision-identified 6-drill evidence bundle.

### Required artifacts

ART-W02-E01-S003-005 (the consolidated bundle, as a registered post-implementation artifact).

### Required evidence

EV-W02-E01-S003-005 (the bundle itself, registered as the story's consolidated evidence record).

### Related acceptance criteria

AC-W02-E01-S003-04 (jointly with T004).

### Completion criteria

The bundle exists, aggregates all four constituent evidence items with complete §10 fields, and is
registered in both indexes.

### Verification method

Index inspection plus field-completeness check across all constituent items.

### Risks

Low — pure aggregation. The only failure mode is a constituent item missing a required field
(commit SHA, command), which this task's step 4 exists to catch before closure rather than at
closure review.

### Rollback or recovery considerations

Not applicable — a documentation/evidence assembly task with no code or data impact.

## Implementation Record

*Not applicable in the code sense — this is an evidence-assembly task. Its execution record is
captured below once performed.*

### What was actually implemented

*Not yet executed.*

### Components changed

*Not applicable.*

### Files changed

*Not yet executed — expected: this story's `evidence/`/`artifacts/` content and indexes.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable.*

### Commits

*Not yet executed.*

### Pull requests

*Not yet executed.*

### Implementation dates

*Not yet executed.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not applicable.*

### Follow-up items

*Not yet executed.*

### Relationship to the approved plan

*Not yet executed.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E01-S003-04 (bundle half) | Index inspection + §10 field-completeness check | Documentation review | Bundle aggregates all four constituent items; every item revision-identified | consolidated evidence bundle | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
