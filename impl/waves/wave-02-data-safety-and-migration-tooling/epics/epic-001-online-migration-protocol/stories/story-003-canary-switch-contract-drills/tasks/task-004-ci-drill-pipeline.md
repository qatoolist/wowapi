---
id: W02-E01-S003-T004
type: task
title: CI drill pipeline
status: todo
parent_story: W02-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E01-S003-T001
  - W02-E01-S003-T002
  - W02-E01-S003-T003
acceptance_criteria:
  - AC-W02-E01-S003-04
artifacts:
  - ART-W02-E01-S003-004
  - ART-W02-E01-S003-006
evidence:
  - EV-W02-E01-S003-004
---

# W02-E01-S003-T004 — CI drill pipeline

## Task Definition

### Task objective

Wire all six directive-named drills into a CI/scheduled pipeline producing a durable passing-run
artifact, per PLAN DATA-09 T9's acceptance criterion ("All six drills run in CI/scheduled
pipeline") — the step PLAN's own risk column calls "Largest single infra investment in PF-DATA."

### Parent story

W02-E01-S003 — Canary, switch, and contract-phase tooling with the full CI drill pipeline.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S003-T001, -T002, -T003 (PLAN T9's "Depends-on" column names T1–T8; within this story, the
three preceding tasks; S001/S002's tooling arrives transitively via the story-level dependency).

### Detailed work

1. Confirm the six-drill list against
   `docs/implementation/architecture-directive-2026-07-11.md`'s own drill naming (resolving
   `story.md`'s assumption: the six bolded requirements across PLAN T4–T8 — interrupted/resumed
   backfill; N-1 on expanded N schema; N before/after backfill; application rollback after switch;
   forward recovery from every failed phase; delayed-contract-after-absence-proven). If the
   directive's list differs, record the divergence in the story's `deviations.md` before wiring.
2. Decide the pipeline trigger (nightly schedule vs. per-PR vs. mixed subsets) against runner-cost
   constraints — `plan.md`'s "Unresolved questions."
3. Build the pipeline: a workflow under `.github/workflows/` (new or extending existing CI) running
   all six drills against fixture migrations with a real PostgreSQL instance.
4. Ensure each pipeline run produces a durable passing-run artifact (retained, revision-identified).
5. Document the pipeline's trigger, drill list, and artifact location. Record PLAN T9's own note
   that "which real migration is the first live exercise" is a human decision — "DATA-01's
   composite-FK rollout is the natural first candidate" (consumed by W02-E02, not by this task).

### Expected files or components affected

A new or extended workflow under `.github/workflows/`; drill fixture migrations.

### Expected output

A scheduled CI pipeline running all six drills, with a durable passing-run artifact.

### Required artifacts

ART-W02-E01-S003-004 (pipeline definition), ART-W02-E01-S003-006 (documentation, shared).

### Required evidence

EV-W02-E01-S003-004 (pipeline passing-run artifact).

### Related acceptance criteria

AC-W02-E01-S003-04 (jointly with T005's aggregation).

### Completion criteria

All six drills run and pass in the pipeline; the passing-run artifact exists and identifies the
tested revision.

### Verification method

Pipeline execution record inspection; drill-by-drill pass confirmation against the confirmed
six-drill list.

### Risks

PLAN T9's own risk column: "Largest single infra investment in PF-DATA" — if pipeline build cost
threatens task boundedness (mandate §12), the story-level contingency applies: land the six drills'
core wiring first, split pipeline hardening into a follow-up task, and never silently narrow which
drills run.

### Rollback or recovery considerations

The pipeline is a verification surface, not a runtime dependency — disabling or reverting it blocks
nothing at runtime, but doing so must be recorded (a silently-disabled drill pipeline would
invalidate the protocol's continuous-verification claim).

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not yet implemented — pipeline trigger/schedule configuration anticipated.*

### Schema or migration changes

*Not applicable — drill fixture migrations are test fixtures, not application schema.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented — the durable passing-run artifact is this task's observability deliverable.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E01-S003-04 (pipeline half) | Pipeline execution against the confirmed six-drill list | CI, PostgreSQL | All six drills pass; durable, revision-identified passing-run artifact | CI-execution record | unassigned |

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
