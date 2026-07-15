---
id: W05-E05-S002-T003
type: task
title: Independent review
status: todo
parent_story: W05-E05-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E05-S002-T001
  - W05-E05-S002-T002
acceptance_criteria:
  - AC-W05-E05-S002-01
  - AC-W05-E05-S002-02
artifacts: []
evidence: []
---

# W05-E05-S002-T003 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: T001's
package-count and lint verification genuinely ran against the actual post-move state (not merely
asserted); T002's wowsociety identity-suite verification genuinely ran the FULL suite (not a
narrowed mfa-scoped subset), per REVIEW §P's own explicit instruction, with both repositories'
commit SHAs genuinely recorded; this epic's own closure claim ("FBL-01 re-homed and verified") is
fully supported by this story's evidence, not asserted beyond what was actually tested.

### Parent story

W05-E05-S002 — Kernel package-count and wowsociety identity-suite verification.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E05-S002-T001, W05-E05-S002-T002.

### Detailed work

1. Independently re-run (or re-inspect the CI output of) T001's `go list` count and lint checks.
2. Independently confirm T002's wowsociety test run genuinely covered the FULL identity/authz suite,
   not a narrowed subset — inspect the actual test-run manifest/output, not merely a pass/fail
   summary.
3. Independently confirm both repositories' commit SHAs are recorded and correspond to the actual
   commits tested.
4. Confirm this story's `story.md` acceptance criteria are not narrower than MATRIX CS-01's own
   acceptance bar.
5. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted` — and, transitively, before this epic (and this whole wave's own kernel-
   layering claim) can be considered closed.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W05-E05-S002-01, AC-W05-E05-S002-02.

### Completion criteria

The review record confirms both acceptance criteria are proven with valid evidence, or lists
findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, with independent re-execution of both T001's and
T002's verification steps as the specific "genuinely, not merely claimed" checks given this epic's
own "largest single architectural correction" and auth-critical content.

### Risks

RISK-W05-003, RISK-W05-004 — mitigated by requiring the reviewer to independently re-run both
verification steps, particularly confirming the FULL wowsociety identity/authz suite was genuinely
exercised, not trusted from a summary result alone.

### Rollback or recovery considerations

Not applicable — a review task has no code to roll back; a failed review blocks story acceptance
(and thus this epic's own closure) until its findings are resolved.

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
| AC-W05-E05-S002-01 | Independent re-execution of package-count/lint checks | Documentation + re-run | Confirmed: count and lint results genuine | review report | unassigned |
| AC-W05-E05-S002-02 | Independent inspection of wowsociety's full test-run manifest | Cross-repo review + test-manifest inspection | Confirmed: full identity/authz suite genuinely run, commit SHAs correct | review report | unassigned |

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
