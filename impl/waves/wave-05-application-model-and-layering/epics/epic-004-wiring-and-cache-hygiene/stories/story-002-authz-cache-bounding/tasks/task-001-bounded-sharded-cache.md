---
id: W05-E04-S002-T001
type: task
title: Bounded, sharded cache
status: todo
parent_story: W05-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E04-S002-01
artifacts:
  - ART-W05-E04-S002-001
evidence:
  - EV-W05-E04-S002-001
---

# W05-E04-S002-T001 — Bounded, sharded cache

## Task Definition

### Task objective

Replace `kernel/authz/caching.go`'s unbounded map+mutex with a `hashicorp/golang-lru/v2`-backed
bounded, sharded cache, sized by config, swapped behind the existing `Store` interface.

### Parent story

W05-E04-S002 — Bounded, epoch-invalidated authorization cache.

### Owner

unassigned

### Status

todo

### Dependencies

None (this story's own first task; independent per MATRIX CS-17's own note).

### Detailed work

1. Re-read `kernel/authz/caching.go` at this task's start commit to confirm the current unbounded
   state.
2. Swap the cache backend for `golang-lru/v2`, sized by config, behind the existing `Store`
   interface.
3. Write the insert->max-keys test and a race test, producing `SEC-04/bounded-cache-tests.md`.
4. Document the bounded-cache configuration surface.

### Expected files or components affected

`kernel/authz/caching.go`.

### Expected output

A bounded cache that never exceeds its configured maximum under adversarial cardinality.

### Required artifacts

ART-W05-E04-S002-001.

### Required evidence

EV-W05-E04-S002-001.

### Related acceptance criteria

AC-W05-E04-S002-01.

### Completion criteria

The insert->max-keys test and race test both pass.

### Verification method

Direct execution of the tests producing `SEC-04/bounded-cache-tests.md`, including under `-race`.

### Risks

Low-moderate, per PLAN T1's own risk column — "swap behind existing `Store` interface."

### Rollback or recovery considerations

Revert if the swap breaks the existing `Store` interface contract for any current caller.

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
| AC-W05-E04-S002-01 | Run the bounded-cache insert/race tests | Local dev or CI, Go toolchain (`-race`) | Cache never exceeds configured max under adversarial cardinality | test report | unassigned |

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
