---
id: W05-E02-S001-T001
type: task
title: port.Key[T] and generic free functions
status: todo
parent_story: W05-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E02-S001-01
artifacts:
  - ART-W05-E02-S001-001
  - ART-W05-E02-S001-003
evidence:
  - EV-W05-E02-S001-001
---

# W05-E02-S001-T001 — port.Key[T] and generic free functions

## Task Definition

### Task objective

Define `port.Key[T]` and implement `Define`, `Provide`, `Require`, `Resolve` as generic free
functions bound to a `Registrar`, proven by a happy-path define/provide/resolve round-trip test.

### Parent story

W05-E02-S001 — Typed port-key API and registrar-forge safety proof.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T002); depends on W05-E01-S001 at story scope.

### Detailed work

1. Design and implement `port.Key[T]`.
2. Implement the four generic free functions bound to a `Registrar`.
3. Write `AR-02/port_api_unit_test.go`: a happy-path define/provide/resolve round-trip.
4. Document the API.

### Expected files or components affected

A new `port` package (exact location TBD).

### Expected output

A working `port.Key[T]` API proven by the happy-path round-trip test.

### Required artifacts

ART-W05-E02-S001-001, ART-W05-E02-S001-003 (documentation, shared with T002).

### Required evidence

EV-W05-E02-S001-001.

### Related acceptance criteria

AC-W05-E02-S001-01.

### Completion criteria

The happy-path round-trip test passes.

### Verification method

Direct execution of `AR-02/port_api_unit_test.go`.

### Risks

Medium, per PLAN T1's own risk column — ergonomics review needed given Go's lack of
type-parameterized methods.

### Rollback or recovery considerations

If ergonomics review reveals the API is unusable in practice, revise the function signatures before
proceeding to S002's graph work, which builds on this API.

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
| AC-W05-E02-S001-01 | Run `AR-02/port_api_unit_test.go` | Local dev or CI, Go toolchain | Happy-path round-trip compiles and works | unit-test report | unassigned |

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
