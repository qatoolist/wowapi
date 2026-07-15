---
id: W02-E04-S001-T004
type: task
title: kernel/resource documentation update
status: todo
parent_story: W02-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E04-S001-T001
acceptance_criteria:
  - AC-W02-E04-S001-04
artifacts:
  - ART-W02-E04-S001-004
evidence:
  - EV-W02-E04-S001-004
---

# W02-E04-S001-T004 — kernel/resource documentation update

## Task Definition

### Task objective

Update `kernel/resource`'s package documentation to describe the mandatory-mirror contract as
implemented by T1/T2, replacing the current manual, comment-only description.

### Parent story

W02-E04-S001 — Typed aggregate write contract with mandatory mirror, audit, and outbox.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E04-S001-T001 (the documentation describes the implemented helper, which must exist first).

### Detailed work

1. Re-read `kernel/resource`'s current package documentation at this task's actual start commit.
2. Rewrite the documentation to describe the new helper as the primary, enforced write path,
   describing the atomicity guarantee (T1) and the actor-attribution behavior (T2).
3. Note the continued availability of the low-level `Upsert` API (per PLAN's own compatibility
   note) so a reader understands both paths exist and why.
4. Submit the documentation update for manual review, per PLAN T4's own "Tests" column ("Manual
   review").

### Expected files or components affected

`kernel/resource`'s package documentation file (exact path TBD).

### Expected output

Documentation that accurately describes the implemented mandatory-mirror contract, reviewed and
confirmed to match the actual implementation.

### Required artifacts

ART-W02-E04-S001-004 (updated documentation).

### Required evidence

EV-W02-E04-S001-004 (documentation-review record).

### Related acceptance criteria

AC-W02-E04-S001-04.

### Completion criteria

Documentation matches the implemented contract, confirmed via a recorded manual review.

### Verification method

Manual review comparing the documentation's claims against the actual T1/T2 implementation.

### Risks

PLAN T4's own named risk: "Low, don't skip — stale docs created this defect class" — the original
defect (a manual, comment-only contract with no enforcement) was itself partly a documentation
problem; this task's own risk is repeating that pattern by documenting an aspiration rather than the
actual implementation.

### Rollback or recovery considerations

Not applicable — a documentation task has minimal rollback risk; if the review finds the
documentation inaccurate, it is corrected and re-reviewed, not reverted.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

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

*Not applicable — this task's verification method is manual review, not an automated test, per
PLAN T4's own "Tests" column.*

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

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E04-S001-04 | Manual review of documentation against implementation | Documentation review | Documentation accurately describes the implemented mandatory-mirror contract | review report | unassigned |

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
