---
id: W05-E01-S002-T004
type: task
title: Independent review
status: todo
parent_story: W05-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E01-S002-T001
  - W05-E01-S002-T002
  - W05-E01-S002-T003
acceptance_criteria:
  - AC-W05-E01-S002-01
  - AC-W05-E01-S002-02
  - AC-W05-E01-S002-03
artifacts: []
evidence: []
---

# W05-E01-S002-T004 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: T002's
authz-permission-registration fix genuinely closes the "only registry with zero existing ownership
check" gap (not merely claimed); T003's declaration-class enumeration is genuinely complete against
AR-01's own acceptance-gate class list, not silently under-scoped; T001's resource/rules wrappers
match PLAN's own acceptance criteria; no source requirement (AR-01 T3-T6) was silently dropped or
narrowed.

### Parent story

W05-E01-S002 — Owner-bound registry wrappers across all declaration classes.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E01-S002-T001, W05-E01-S002-T002, W05-E01-S002-T003 (review requires all three to be implemented
first).

### Detailed work

1. Confirm T001's `resource.Registry`/`rules.Registry` wrappers match PLAN AR-01 T3/T4's own
   acceptance criteria and that both adversarial tests genuinely reject a matching-key-prefix
   cross-module claim.
2. Confirm T002's `authz.Registry` fix genuinely closes the zero-ownership-check gap — independently
   re-run or re-inspect `AR-01/authz_ownership_adversarial_test.go`'s actual assertions, not merely
   its pass/fail result, given this is PLAN's own "actual security boundary" and "widest gap"
   language.
3. Confirm T003's declaration-class enumeration (ART-W05-E01-S002-005) is checked against the
   framework's actual registration surface, not merely against PLAN's own "~9+" approximate list —
   independently spot-check at least one declaration class not obviously covered by the audit's own
   list to probe for under-scoping.
4. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item.
5. Confirm this story's `story.md` acceptance criteria are not narrower than PLAN AR-01 T3-T6's own
   acceptance-criteria columns.
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

AC-W05-E01-S002-01, AC-W05-E01-S002-02, AC-W05-E01-S002-03.

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence, or lists
findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001/T002/T003's evidence, with
an independent spot-check of T003's declaration-class enumeration as the specific
"genuinely, not merely claimed" check given RISK-W05-002's under-scoping concern.

### Risks

RISK-W05-001, RISK-W05-002 — the review itself missing a genuine gap is mitigated by requiring the
reviewer to independently spot-check T003's enumeration and re-inspect T002's adversarial test's
actual assertions, rather than trusting the tasks' own self-reported completion.

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
| AC-W05-E01-S002-01 | Independent review against mandate §14 checklist | Documentation + code review | Confirmed: resource/rules wrappers genuinely reject cross-module claims | review report | unassigned |
| AC-W05-E01-S002-02 | Independent re-inspection of the authz adversarial test's assertions | Code review + test-output inspection | Confirmed: authz-registration gap genuinely closed | review report | unassigned |
| AC-W05-E01-S002-03 | Independent spot-check of the declaration-class enumeration | Code review + registration-surface audit | Confirmed: enumeration is complete, not under-scoped | review report | unassigned |

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
