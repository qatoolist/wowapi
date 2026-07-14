---
id: W05-E03-S002-T002
type: task
title: Empty-required-fragment rejection
status: todo
parent_story: W05-E03-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E03-S002-01
artifacts:
  - ART-W05-E03-S002-002
evidence:
  - EV-W05-E03-S002-001
---

# W05-E03-S002-T002 — Empty-required-fragment rejection

## Task Definition

### Task objective

Make a module declaring a required-but-empty fragment fail boot.

### Parent story

W05-E03-S002 — Boot-time strictness and the shared no-op-adapter waiver mechanism.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T001, T003); depends on W05-E01 (AR-01, and T001/T002's
own preceding surface per PLAN's own "T1-T2" dependency row) at story scope.

### Detailed work

1. Identify the framework's "required fragment" declaration mechanism.
2. Implement boot-failure rejection for a required-but-empty fragment.
3. Write `AR-04/empty_required_fragment_test.go`: adversarial fixture.
4. Document the rejection rule.

### Expected files or components affected

Required-fragment validation logic (exact location TBD).

### Expected output

A module declaring a required-but-empty fragment fails boot, proven by the named test.

### Required artifacts

ART-W05-E03-S002-002.

### Required evidence

EV-W05-E03-S002-001 (combined with T001's evidence record per the story's own evidence index).

### Related acceptance criteria

AC-W05-E03-S002-01.

### Completion criteria

The adversarial fixture confirms boot failure for a required-but-empty fragment.

### Verification method

Direct execution of `AR-04/empty_required_fragment_test.go`.

### Risks

Low-medium, per PLAN T3's own risk column.

### Rollback or recovery considerations

If the fixture reveals a false-negative (empty fragment not rejected), fix before proceeding.

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
| AC-W05-E03-S002-01 | Run `AR-04/empty_required_fragment_test.go` | Local dev or CI, Go toolchain | Required-but-empty fragment fails boot | adversarial-test report | unassigned |

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
