---
id: W05-E05-S001-T005
type: task
title: Independent review
status: todo
parent_story: W05-E05-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E05-S001-T001
  - W05-E05-S001-T002
  - W05-E05-S001-T003
  - W05-E05-S001-T004
acceptance_criteria:
  - AC-W05-E05-S001-01
  - AC-W05-E05-S001-02
  - AC-W05-E05-S001-03
artifacts: []
evidence: []
---

# W05-E05-S001-T005 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
9-package move is genuinely behaviour-preserving (T001, T002); the `kernel/mfa` shim's
behavioral-equivalence test is genuinely proven, not merely claimed, given its auth-critical status;
the depguard and boundaries-lint extensions genuinely enforce the new layering, not merely appear to
in a narrow test case; no source requirement (FBL-01, MATRIX CS-01's own 5-step mechanics) was
silently dropped or narrowed.

### Parent story

W05-E05-S001 — Foundation tree, package moves, and mfa forwarding shim.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E05-S001-T001, W05-E05-S001-T002, W05-E05-S001-T003, W05-E05-S001-T004.

### Detailed work

1. Independently confirm the full repository build passes post-move, and that `git log` shows
   preserved history for a sample of moved files across both T001 and T002's package sets.
2. Independently re-run (or re-inspect the CI output of) T002's `kernel/mfa` shim
   behavioral-equivalence test — specifically confirming every exported symbol the shim forwards is
   covered, not a partial subset, given the auth-critical status REVIEW §P assigns this package.
3. Independently re-attempt the depguard and boundaries-lint adversarial fixtures (T003, T004),
   confirming both are genuinely denied/failed, not merely asserted to be.
4. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item.
5. Confirm this story's `story.md` acceptance criteria are not narrower than MATRIX CS-01's own
   5-step mechanics and acceptance bar.
6. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W05-E05-S001-01, AC-W05-E05-S001-02, AC-W05-E05-S001-03.

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence, or lists
findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, with independent re-execution of the build, the
shim's equivalence test, and both lint adversarial fixtures as the specific "genuinely, not merely
claimed" checks given this story's own architectural-correction and auth-critical content.

### Risks

RISK-W05-004 — mitigated by requiring the reviewer to independently re-verify the shim's
behavioral-equivalence coverage, not merely trust T002's own self-reported completion.

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
| AC-W05-E05-S001-01 | Independent re-verification of build + git history | Documentation + build re-run | Confirmed: build passes, history preserved | review report | unassigned |
| AC-W05-E05-S001-02 | Independent re-execution of the shim equivalence test | Code review + test re-execution | Confirmed: shim genuinely equivalent, full symbol coverage | review report | unassigned |
| AC-W05-E05-S001-03 | Independent re-attempt of both lint adversarial fixtures | Code review + lint re-execution | Confirmed: both denial rules genuinely trigger | review report | unassigned |

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
