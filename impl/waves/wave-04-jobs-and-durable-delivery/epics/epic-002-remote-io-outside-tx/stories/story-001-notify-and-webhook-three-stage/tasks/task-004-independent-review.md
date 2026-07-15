---
id: W04-E02-S001-T004
type: task
title: Independent review
status: done
parent_story: W04-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S001-T001
  - W04-E02-S001-T002
  - W04-E02-S001-T003
acceptance_criteria:
  - AC-W04-E02-S001-01
  - AC-W04-E02-S001-02
  - AC-W04-E02-S001-03
artifacts: []
evidence: []
---

# W04-E02-S001-T004 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance
criteria are proven with valid evidence; the shared primitive was genuinely reused (not copied);
the self-documented "should move outside tx" comment was genuinely resolved, not merely deleted
without the underlying gap being closed; the webhook current-row-state check genuinely moved to the
claim stage; no source requirement (DATA-03 T1, T2, T3) was silently dropped or narrowed.

### Parent story

W04-E02-S001 — Notify and webhook three-stage remote-I/O protocol.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E02-S001-T001, W04-E02-S001-T002, W04-E02-S001-T003 (review requires all three to be
implemented first).

### Detailed work

1. Confirm T001's lease-column migration matches W04-E01's shared primitive's own schema exactly —
   not a parallel or bespoke implementation with a similar but distinct column set.
2. Confirm T002's three-stage protocol matches PLAN DATA-03 T2's acceptance criterion ("No
   `sender.Send` call while a DB tx is open") and that the self-documented comment deletion/update
   genuinely reflects a closed gap, not a comment removal papering over an unresolved issue.
3. Confirm T003's three-stage protocol matches PLAN DATA-03 T3's acceptance criterion ("No
   DNS/secret-resolve/POST call while a tx is open") and that the current-row-state check is
   genuinely in the claim stage, confirmed by direct code inspection, not merely by task-record
   claim.
4. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item —
   evidence without this must not be treated as final proof (mandate §10).
5. Confirm this story's `story.md` acceptance criteria are not narrower than PLAN DATA-03 T1/T2/T3's
   own acceptance-criteria columns.
6. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record itself is captured in this task's own Verification Record.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E02-S001-01, AC-W04-E02-S001-02, AC-W04-E02-S001-03 (confirms all three, does not itself
prove any new one).

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence, or lists
findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001/T002/T003's evidence.

### Risks

None beyond the review itself missing a genuine gap — mitigated by requiring the review to
specifically re-check the "genuinely, not merely claimed" points above (shared-primitive reuse,
comment resolution, claim-stage relocation) rather than trusting T001/T002/T003's own self-reported
completion.

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
| AC-W04-E02-S001-01 | Independent review against mandate §14 checklist | Code review | Confirmed: shared primitive genuinely reused, not copied | review report | unassigned |
| AC-W04-E02-S001-02 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: no send-while-tx-open, comment genuinely resolved | review report | unassigned |
| AC-W04-E02-S001-03 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: no network-call-while-tx-open, check genuinely in claim stage | review report | unassigned |

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
