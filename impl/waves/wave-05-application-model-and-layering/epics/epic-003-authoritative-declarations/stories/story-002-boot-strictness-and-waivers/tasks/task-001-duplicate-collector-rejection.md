---
id: W05-E03-S002-T001
type: task
title: Duplicate-collector rejection
status: todo
parent_story: W05-E03-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E03-S002-01
artifacts:
  - ART-W05-E03-S002-001
evidence:
  - EV-W05-E03-S002-001
---

# W05-E03-S002-T001 — Duplicate-collector rejection

## Task Definition

### Task objective

Make every collector reject a second write to the same identity (replacing today's
last-writer-wins behavior), while explicitly preserving legitimate multi-locale accumulation.

### Parent story

W05-E03-S002 — Boot-time strictness and the shared no-op-adapter waiver mechanism.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T002, T003); depends on W05-E01 (AR-01 T1) at story
scope.

### Detailed work

1. Audit every collector type's current last-writer-wins behavior.
2. Implement duplicate-write rejection, explicitly excepting the legitimate
   multi-locale-accumulation pattern.
3. Write `AR-04/duplicate_collector_rejection_test.go`: one adversarial fixture per collector type,
   plus a positive fixture proving multi-locale accumulation is not falsely rejected.
4. Document the rejection rule and its legitimate exception.

### Expected files or components affected

The collector implementations across the registration surface (exact list TBD by the audit).

### Expected output

Duplicate writes rejected per collector type; legitimate multi-locale accumulation preserved.

### Required artifacts

ART-W05-E03-S002-001.

### Required evidence

EV-W05-E03-S002-001.

### Related acceptance criteria

AC-W05-E03-S002-01.

### Completion criteria

All adversarial fixtures pass (duplicates rejected); the positive fixture confirms no false-positive
rejection of legitimate accumulation.

### Verification method

Direct execution of `AR-04/duplicate_collector_rejection_test.go`.

### Risks

Medium, per PLAN T2's own risk column — "distinguish illegitimate duplicate from legitimate
multi-locale accumulation."

### Rollback or recovery considerations

If the positive fixture reveals false-positive rejection of legitimate accumulation, fix before
proceeding — do not ship an overly-aggressive check.

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
| AC-W05-E03-S002-01 | Run `AR-04/duplicate_collector_rejection_test.go` | Local dev or CI, Go toolchain | Duplicates rejected; legitimate accumulation not falsely rejected | adversarial-test report | unassigned |

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
