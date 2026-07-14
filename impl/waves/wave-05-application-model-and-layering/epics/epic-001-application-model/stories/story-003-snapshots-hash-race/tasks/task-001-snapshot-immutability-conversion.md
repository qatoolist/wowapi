---
id: W05-E01-S003-T001
type: task
title: Snapshot-immutability conversion
status: todo
parent_story: W05-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E01-S003-01
artifacts:
  - ART-W05-E01-S003-001
evidence:
  - EV-W05-E01-S003-001
---

# W05-E01-S003-T001 — Snapshot-immutability conversion

## Task Definition

### Task objective

Convert every exported registry reader (`Specs()`, `Points()`, and equivalents) across all
registries S002 wrapped to return cloned/immutable data, so no exported reader returns a backing
map/slice.

### Parent story

W05-E01-S003 — Snapshot immutability, post-seal rejection, model hash, and race safety.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T002 — disjoint concern); depends on W05-E01-S002 (T3-T6)
at story scope.

### Detailed work

1. Audit every S002-wrapped registry's exported reader methods, confirming which return a backing
   map/slice at this task's start commit.
2. Convert each such reader to return cloned/immutable data.
3. Write `AR-01/snapshot_immutability_test.go`: mutate a returned value, assert registry internal
   state unaffected, across all wrapped registries.
4. Document the immutability guarantee.

### Expected files or components affected

The registries wrapped in S002 (`kernel/resource`, `kernel/rules`, `kernel/authz`, and the remaining
~9+ declaration classes) — exact file paths TBD per `plan.md`.

### Expected output

No exported registry reader returns a backing map/slice — proven by the named test.

### Required artifacts

ART-W05-E01-S003-001.

### Required evidence

EV-W05-E01-S003-001.

### Related acceptance criteria

AC-W05-E01-S003-01.

### Completion criteria

Mutating a value returned by any wrapped registry's exported reader does not affect that registry's
internal state — proven by the named test passing.

### Verification method

Direct execution of `AR-01/snapshot_immutability_test.go`.

### Risks

Low-medium, per PLAN T7's own risk column.

### Rollback or recovery considerations

If the conversion is found to have missed a reader (e.g. at review time), extend the test with the
missed case and fix — record any such gap in `deviations.md`.

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
| AC-W05-E01-S003-01 | Run `AR-01/snapshot_immutability_test.go` | Local dev or CI, Go toolchain | Mutating a returned value does not affect registry internal state | unit-test report | unassigned |

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
