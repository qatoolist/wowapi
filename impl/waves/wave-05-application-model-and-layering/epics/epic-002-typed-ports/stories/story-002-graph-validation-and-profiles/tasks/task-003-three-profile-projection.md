---
id: W05-E02-S002-T003
type: task
title: Three-profile projection compiler
status: todo
parent_story: W05-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E02-S002-T001
  - W05-E02-S002-T002
acceptance_criteria:
  - AC-W05-E02-S002-03
artifacts:
  - ART-W05-E02-S002-003
evidence:
  - EV-W05-E02-S002-003
---

# W05-E02-S002-T003 — Three-profile projection compiler

## Task Definition

### Task objective

Compile API/worker/migrate profiles as three projections of one provider graph, so no hand-copied
wiring template remains, proven by building all three from one fixture and asserting capability
subsets.

### Parent story

W05-E02-S002 — Zero-reflection provider graph, boot-time validation, and profile projection.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E02-S002-T001, W05-E02-S002-T002 (the validated graph this task projects).

### Detailed work

1. Design the projection mechanism: how a profile's capability subset is derived from the one graph.
2. Implement the projection compiler for the API, worker, and migrate profiles.
3. Write `AR-02/three_profile_projection_test.go`: build all three profiles from one fixture, assert
   correct capability subsets per profile.
4. Confirm no hand-copied wiring template remains in the codebase for any of the three profiles.
5. Document the projection mechanism, noting its forward relationship to AR-03's own manifest shape
   (W05-E03 scope, a later epic that consumes this mechanism).

### Expected files or components affected

A new projection-compiler package (exact location TBD).

### Expected output

API/worker/migrate profiles compiled as three projections of one graph, proven by the named test.

### Required artifacts

ART-W05-E02-S002-003.

### Required evidence

EV-W05-E02-S002-003.

### Related acceptance criteria

AC-W05-E02-S002-03.

### Completion criteria

All three profiles build from one fixture with correct capability subsets, and no hand-copied wiring
template remains.

### Verification method

Direct execution of `AR-02/three_profile_projection_test.go`.

### Risks

Medium, per PLAN T5's own risk column — "sequence after AR-03's manifest shape is fixed." This
story's own "Unresolved questions"/story-level assumption records the interpretation that this task
delivers the mechanism AR-03 later consumes, not a hard block on AR-03 completing first.

### Rollback or recovery considerations

If AR-03's later epic (W05-E03) reveals this task's projection mechanism needs a shape change to fit
the manifest, record as a cross-epic deviation at that time — do not silently redesign this task's
own output without recording why.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not yet implemented.*

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
| AC-W05-E02-S002-03 | Run `AR-02/three_profile_projection_test.go` | Local dev or CI, Go toolchain | All 3 profiles build from one fixture, correct capability subsets | integration-test report | unassigned |

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
