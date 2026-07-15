---
id: W02-E01-S001-T003
type: task
title: Independent review
status: todo
parent_story: W02-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E01-S001-T001
  - W02-E01-S001-T002
acceptance_criteria:
  - AC-W02-E01-S001-01
  - AC-W02-E01-S001-02
  - AC-W02-E01-S001-03
artifacts: []
evidence: []
---

# W02-E01-S001-T003 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance criteria
are proven with valid evidence; the manifest schema genuinely received external review before being
locked (not merely claimed); the lock-timeout retry ceiling is genuinely bounded (not merely
claimed); no source requirement (DATA-09 T1, T2) was silently dropped or narrowed.

### Parent story

W02-E01-S001 — Migration manifest schema and online-DDL lock budget.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S001-T001, W02-E01-S001-T002 (review requires both to be implemented first).

### Detailed work

1. Confirm T001's manifest schema matches PLAN DATA-09 T1's acceptance criterion ("Every migration
   has a validated manifest entry; missing fields fail CI") and that the external-review record
   (EV-W02-E01-S001-002) is genuine — dated, attributed, and predating enforcement.
2. Confirm T002's lock-timeout mechanism matches PLAN DATA-09 T2's acceptance criterion ("A
   statement exceeding budget aborts cleanly, no partial DDL") and that the retry ceiling is
   genuinely bounded, not merely documented as bounded while the code retries indefinitely.
3. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item —
   evidence without this must not be treated as final proof (mandate §10).
4. Confirm this story's `story.md` acceptance criteria are not narrower than PLAN DATA-09 T1/T2's
   own acceptance-criteria columns.
5. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record itself is captured in this task's own Verification Record / a review-report
evidence item, per the story's `evidence/index.md` pattern used elsewhere in this programme (this
story does not register a separate review-report evidence ID beyond EV-W02-E01-S001-002, which
covers T001's external-review specifically; this task's own review record is the story-level
independent review, recorded in this task file's Verification Record below).

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W02-E01-S001-01, AC-W02-E01-S001-02, AC-W02-E01-S001-03 (confirms all three, does not itself
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
| AC-W02-E01-S001-01 | Independent review against mandate §14 checklist | Documentation + code review | Confirmed: schema enforced, negative fixture genuinely fails CI | review report | unassigned |
| AC-W02-E01-S001-02 | Independent review against mandate §14 checklist | Documentation review | Confirmed: external review genuinely occurred before enforcement | review report | unassigned |
| AC-W02-E01-S001-03 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: retry ceiling genuinely bounded, abort genuinely clean | review report | unassigned |

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
