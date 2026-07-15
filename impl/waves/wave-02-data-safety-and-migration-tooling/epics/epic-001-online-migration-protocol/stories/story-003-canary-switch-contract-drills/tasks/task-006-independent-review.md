---
id: W02-E01-S003-T006
type: task
title: Independent review
status: todo
parent_story: W02-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E01-S003-T001
  - W02-E01-S003-T002
  - W02-E01-S003-T003
  - W02-E01-S003-T004
  - W02-E01-S003-T005
acceptance_criteria:
  - AC-W02-E01-S003-01
  - AC-W02-E01-S003-02
  - AC-W02-E01-S003-03
  - AC-W02-E01-S003-04
artifacts: []
evidence: []
---

# W02-E01-S003-T006 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all four acceptance criteria
are proven with valid, revision-identified evidence; and — the review's story-specific focus per
epic-level `acceptance.md` AC-W02-E01-04 — the soak-threshold judgment gap (RISK-W02-003) is
honestly recorded as an accepted residual risk, not silently resolved with an invented number.

### Parent story

W02-E01-S003 — Canary, switch, and contract-phase tooling with the full CI drill pipeline.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S003-T001 through -T005 (review requires all five completed first).

### Detailed work

1. Confirm T001's canary test genuinely covers both explicitly-required legs (N-1 on expanded N
   schema; N before/after backfill) — read the test's assertions, not its name.
2. Confirm soak duration/threshold parameters are genuinely configurable with no hardcoded guessed
   values presented as calibrated, and that documentation records calibration as a per-rollout
   human judgment (RISK-W02-003 honestly recorded).
3. Confirm T002's switch-rollback test asserts post-rollback behavioral correctness, not merely
   rollback completion, and that no destructive `Down` exists in any tooling-managed path.
4. Confirm T003's contract gate fails closed — the negative cases (missing/ambiguous evidence) are
   present in the test and pass — and forward recovery is exercised from every phase, not a subset.
5. Confirm T004's pipeline runs the directive-confirmed six-drill list (per T004 step 1's
   confirmation), with any divergence from the assumed list recorded in `deviations.md`.
6. Confirm T005's bundle aggregates all constituent evidence with complete mandate-§10 fields.
7. Confirm this story's acceptance criteria are not narrower than PLAN T6–T9's own acceptance-
   criteria columns, and no source requirement was silently dropped.
8. Record findings; any issue must be resolved or explicitly accepted before this story moves to
   `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record is captured in this task's own Verification Record, consistent with the
pattern in W02-E01-S001-T003 and W02-E01-S002-T004.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W02-E01-S003-01 through -04 (confirms all four, does not itself prove any new one).

### Completion criteria

The review record confirms all four acceptance criteria are proven with valid evidence and
RISK-W02-003 is honestly recorded, or lists findings that must be resolved before this story can
close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001–T005's evidence.

### Risks

The review accepting a weakened named test (steps 1/3/4's concern) — mitigated by requiring the
reviewer to read each test's assertions directly.

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
| AC-W02-E01-S003-01 | Independent review against mandate §14 checklist | Test-assertion + documentation review | Confirmed: both canary legs genuinely tested; soak parameters configurable; RISK-W02-003 honestly recorded | review report | unassigned |
| AC-W02-E01-S003-02 | Independent review against mandate §14 checklist | Test-assertion + code review | Confirmed: rollback correctness asserted; no destructive `Down` | review report | unassigned |
| AC-W02-E01-S003-03 | Independent review against mandate §14 checklist | Test-assertion review incl. negative cases | Confirmed: fail-closed proven; forward recovery from every phase | review report | unassigned |
| AC-W02-E01-S003-04 | Independent review against mandate §14 checklist | CI record + bundle inspection | Confirmed: six directive-confirmed drills run; bundle complete with §10 fields | review report | unassigned |

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
