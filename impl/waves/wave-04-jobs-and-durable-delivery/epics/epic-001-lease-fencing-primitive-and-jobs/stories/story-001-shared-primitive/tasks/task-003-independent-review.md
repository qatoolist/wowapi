---
id: W04-E01-S001-T003
type: task
title: Independent review
status: done
parent_story: W04-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S001-T001
  - W04-E01-S001-T002
acceptance_criteria:
  - AC-W04-E01-S001-01
  - AC-W04-E01-S001-02
  - AC-W04-E01-S001-03
artifacts: []
evidence: []
---

# W04-E01-S001-T003 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance criteria
are proven with valid evidence; the cross-consumer field-set review genuinely occurred against
DATA-03/DATA-04's stated needs (not merely claimed); the interim-checkpoint-lease migration
genuinely executed and evidenced (not silently skipped); no source requirement (DATA-02 T1) was
silently dropped or narrowed.

### Parent story

W04-E01-S001 — Shared lease/fencing primitive.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S001-T001, W04-E01-S001-T002 (review requires both to be implemented first).

### Detailed work

1. Confirm T001's primitive matches PLAN DATA-02 T1's acceptance criterion ("One primitive reused
   ≥3 times, not three independent copies") — specifically, that the cross-consumer field-set review
   (EV-W04-E01-S001-002) is genuine: dated, attributed, and actually checked against DATA-03's and
   DATA-04's own PLAN task rows, not a self-referential sign-off.
2. Confirm T002's interim-checkpoint-lease migration is genuinely complete: the interim lease code
   path no longer exists, and the migration test (EV-W04-E01-S001-003) genuinely proves no
   checkpoint-state loss or duplication, not merely that the migration ran without error.
3. Confirm RISK-W04-001 is correctly recorded as resolved (not silently left open with the interim
   lease still present) once T002 is complete.
4. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item —
   evidence without this must not be treated as final proof (mandate §10).
5. Confirm this story's `story.md` acceptance criteria are not narrower than PLAN DATA-02 T1's own
   acceptance-criteria column.
6. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record itself is captured in this task's own Verification Record, consistent with
the pattern in W02-E01-S001-T003.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E01-S001-01, AC-W04-E01-S001-02, AC-W04-E01-S001-03 (confirms all three, does not itself
prove any new one).

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence, or lists
findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001/T002's evidence.

### Risks

None beyond the review itself missing a genuine gap — mitigated by requiring the review to
specifically re-check the two named "genuinely, not merely claimed" points above rather than
trusting T001/T002's own self-reported completion.

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
| AC-W04-E01-S001-01 | Independent review against mandate §14 checklist | Test-assertion + code review | Confirmed: comparison semantics genuinely tested, not merely documented | review report | unassigned |
| AC-W04-E01-S001-02 | Independent review against mandate §14 checklist | Documentation review | Confirmed: cross-consumer review genuinely occurred against DATA-03/DATA-04's stated needs | review report | unassigned |
| AC-W04-E01-S001-03 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: interim lease genuinely removed; migration test genuinely proves no loss/duplication | review report | unassigned |

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
