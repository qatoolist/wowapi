---
id: W04-E01-S003-T004
type: task
title: Evidence aggregation (consolidated bundle)
status: done
parent_story: W04-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S003-T001
  - W04-E01-S003-T002
  - W04-E01-S003-T003
acceptance_criteria:
  - AC-W04-E01-S003-01
  - AC-W04-E01-S003-02
  - AC-W04-E01-S003-03
artifacts:
  - ART-W04-E01-S003-004
evidence:
  - EV-W04-E01-S003-004
---

# W04-E01-S003-T004 — Evidence aggregation (consolidated bundle)

## Task Definition

### Task objective

Consolidate T001/T002/T003's individual test outputs (the duplicate-effect/registration-rejection
test, the effect-ledger-vs-fencing test, and the named chaos test) into one consolidated evidence
bundle, registered in `evidence/index.md` with every mandate-§10 required field (commit SHA,
execution command, environment, result) per constituent test.

**Why this task exists (recorded per the wave-planning brief's instruction to document the
reasoning when adding an evidence-collection task):** this story's evidence is produced by three
separate tasks with materially different evidence types — a registration/unit test, an integration
test, and a multi-boundary chaos test — and nothing in PLAN DATA-02's own T5–T7 rows owns assembling
them into the single consolidated record this story's closure and the epic's AC-W04-E01-03 both
consume. This differs from a case like this epic's own W04-E01-S002, whose three tasks' evidence
remained naturally separable and small enough not to warrant aggregation. Here, a genuine,
separately-producible aggregation artifact exists, and — distinctly from W02-E01-S003's own
rationale — this bundle carries an additional, story-specific purpose: W04-E02's and W04-E03's own
reviewers, before trusting their reuse of T003's chaos harness, need a single place to confirm this
story's own chaos-test evidence genuinely passed all three named boundaries, rather than hunting
across three separate task files. T004's own existence is itself part of making the harness's
"shared, not reimplemented" intent operationally real for those downstream consumers.

### Parent story

W04-E01-S003 — Worker idempotency contract and the shared duplicate-worker chaos harness.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S003-T001, W04-E01-S003-T002, W04-E01-S003-T003 (aggregation requires all three test
outputs to exist).

### Detailed work

1. Collect the three test outputs (EV-W04-E01-S003-001/-002/-003).
2. Assemble the consolidated bundle: per-test result, commit SHA, execution command, environment,
   and date, cross-referenced to the three DATA-02 T5/T6/T7 rows each test proves.
3. Register the bundle in `evidence/index.md` (EV-W04-E01-S003-004) and in this story's
   `artifacts/index.md` (ART-W04-E01-S003-004, lifecycle stage post-implementation).
4. Verify no constituent evidence item is missing a mandate-§10 required field — an item without a
   commit SHA must not be treated as final proof.

### Expected files or components affected

This story's `evidence/` and `artifacts/` directories (first real content triggers subdirectory
creation per naming-conventions Adaptation 2); no application code.

### Expected output

One consolidated, revision-identified 3-test evidence bundle.

### Required artifacts

ART-W04-E01-S003-004 (the consolidated bundle, as a registered post-implementation artifact).

### Required evidence

EV-W04-E01-S003-004 (the bundle itself, registered as the story's consolidated evidence record).

### Related acceptance criteria

AC-W04-E01-S003-01, AC-W04-E01-S003-02, AC-W04-E01-S003-03 (jointly with T001/T002/T003).

### Completion criteria

The bundle exists, aggregates all three constituent evidence items with complete §10 fields, and is
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
| AC-W04-E01-S003-01 through -03 (bundle) | Index inspection + §10 field-completeness check | Documentation review | Bundle aggregates all three constituent items; every item revision-identified | consolidated evidence bundle | unassigned |

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
