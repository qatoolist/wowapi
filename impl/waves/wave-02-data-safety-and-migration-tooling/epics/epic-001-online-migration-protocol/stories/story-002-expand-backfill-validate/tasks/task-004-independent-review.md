---
id: W02-E01-S002-T004
type: task
title: Independent review
status: todo
parent_story: W02-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E01-S002-T001
  - W02-E01-S002-T002
  - W02-E01-S002-T003
acceptance_criteria:
  - AC-W02-E01-S002-01
  - AC-W02-E01-S002-02
  - AC-W02-E01-S002-03
artifacts: []
evidence: []
---

# W02-E01-S002-T004 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance criteria
are proven with valid evidence; and — the review's story-specific focus per epic-level
`acceptance.md` AC-W02-E01-04 — the interim-checkpoint-lease deviation is correctly recorded
(bounded scope, explicit forward reference to W04-E01-S001), not silently absorbed as if it were
DATA-02 T1's full primitive.

### Parent story

W02-E01-S002 — Expand-phase tooling, resumable backfill harness, and validation-phase tooling.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S002-T001, W02-E01-S002-T002, W02-E01-S002-T003 (review requires all three implemented
first).

### Detailed work

1. Confirm T001's expand-phase tooling matches PLAN DATA-09 T3's acceptance criterion ("Expand
   migrations don't block traffic; old and new readers both accept") and that the old-reader-
   compatibility evidence is genuine and revision-identified.
2. Confirm T002's interrupted/resumed backfill test genuinely proves "no reprocessing or skipping"
   (PLAN T4's explicitly-required test) — not a weaker proxy (e.g. a resume that merely completes
   without asserting row-level idempotency).
3. Confirm the interim checkpoint-lease's scope boundary is documented in code and story
   documentation, with the W04-E01-S001 forward reference present — the RISK-W02-001 mitigation
   genuinely executed, not merely claimed.
4. Confirm T003's validation report is genuinely machine-checked (artifact-schema-conformant), not
   free-form prose relabeled as an artifact.
5. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item
   (mandate §10), and that this story's acceptance criteria are not narrower than PLAN T3/T4/T5's
   own acceptance-criteria columns.
6. Record findings; any issue must be resolved or explicitly accepted before this story moves to
   `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record is captured in this task's own Verification Record, consistent with the
pattern in W02-E01-S001-T003.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W02-E01-S002-01, AC-W02-E01-S002-02, AC-W02-E01-S002-03 (confirms all three, does not itself
prove any new one).

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence and the
interim-lease deviation is honestly recorded, or lists findings that must be resolved before this
story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001/T002/T003's evidence.

### Risks

The review missing a weakened interrupted/resumed test (step 2's concern) — mitigated by requiring
the reviewer to read the test's assertions directly, not its name or summary.

### Rollback or recovery considerations

Not applicable — a review task has no code to roll back; a failed review blocks story acceptance
until its findings are resolved.

## Implementation Record

*Not applicable — this is a review task, not an implementation task.*

### What was actually implemented

*Not applicable.*

### Components changed

*Not applicable.*

### Files changed

*Not applicable.*

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

*Not applicable.*

### Pull requests

*Not applicable.*

### Implementation dates

*Not applicable.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not applicable.*

### Follow-up items

*Not yet executed.*

### Relationship to the approved plan

*Not applicable.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E01-S002-01 | Independent review against mandate §14 checklist | Documentation + code review | Confirmed: expand tooling non-blocking, compatibility evidence genuine | review report | unassigned |
| AC-W02-E01-S002-02 | Independent review against mandate §14 checklist | Code review + test-assertion inspection | Confirmed: interrupted/resumed test asserts row-level no-reprocess/no-skip; interim-lease scope boundary + W04 forward reference documented | review report | unassigned |
| AC-W02-E01-S002-03 | Independent review against mandate §14 checklist | Documentation + artifact inspection | Confirmed: validation report genuinely machine-checked | review report | unassigned |

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
