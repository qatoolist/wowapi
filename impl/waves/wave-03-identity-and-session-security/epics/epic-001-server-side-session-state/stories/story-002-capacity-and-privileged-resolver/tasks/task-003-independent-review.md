---
id: W03-E01-S002-T003
type: task
title: Independent review
status: todo
parent_story: W03-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E01-S002-T001
  - W03-E01-S002-T002
acceptance_criteria:
  - AC-W03-E01-S002-01
  - AC-W03-E01-S002-02
artifacts: []
evidence:
  - EV-W03-E01-S002-003
---

# W03-E01-S002-T003 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story per mandate §14, specifically confirming: the
multi-capacity test genuinely exercises the no-choice/valid-choice/unentitled-assertion cases; the
privileged-session resolver's adversarial test suite genuinely exercises all six named conditions
(expired, revoked, wrong-tenant, wrong-actor, forged-ID, unauthorized-approver) with distinguishable
rejection reasons; the `Actor` struct shape was preserved wherever the implementation allowed, per
the stated compatibility strategy; no source requirement (SEC-01 T4/T5) was silently narrowed.

### Parent story

W03-E01-S002 — Capacity selection and privileged-session resolver.

### Owner

unassigned

### Status

todo

### Dependencies

W03-E01-S002-T001, W03-E01-S002-T002.

### Detailed work

1. Confirm implementation matches `plan.md`, or that every divergence is recorded in
   `deviations.md`.
2. Confirm both acceptance criteria are each backed by passing tests with logged evidence,
   referencing the correct commit SHA.
3. Confirm all six adversarial rejection conditions for T5 are genuinely, independently tested — not
   collapsed into a single combined fixture that could mask a missing condition.
4. Confirm the `Actor` struct-shape compatibility strategy was honored wherever the implementation
   allowed, and that any necessary deviation (e.g. a field rename) is explicitly flagged as a
   breaking compile-time change in `deviations.md`, not silently introduced.
5. Confirm T4's capacity-selection mechanism choice is documented, not left implicit in code only.
6. Record findings; resolve or explicitly accept before this story can move to `accepted`.

### Expected files or components affected

None (review-only task).

### Expected output

A completed review report confirming the checklist above.

### Required artifacts

None.

### Required evidence

EV-W03-E01-S002-003 (review report).

### Related acceptance criteria

AC-W03-E01-S002-01, AC-W03-E01-S002-02.

### Completion criteria

Review report completed with no open finding, or every finding resolved or explicitly accepted.

### Verification method

Manual independent review against the checklist above, conducted by a reviewer who did not
implement T001/T002.

### Risks

None beyond the story's own inherited risks (RISK-W03-005) — this task's own risk is limited to the
review being performed superficially; mitigated by the explicit per-condition checklist above.

### Rollback or recovery considerations

Not applicable — a review-only task has no code to roll back.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not applicable — review-only task.*

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

*Not applicable.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S002-01, -02 | Independent review checklist per mandate §14 | N/A (documentation/code review) | No open finding, or every finding resolved/accepted | review report | unassigned |

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
