---
id: W05-E04-S002-T003
type: task
title: Singleflight-collapse of concurrent misses
status: todo
parent_story: W05-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E04-S002-T001
acceptance_criteria:
  - AC-W05-E04-S002-02
artifacts:
  - ART-W05-E04-S002-003
evidence:
  - EV-W05-E04-S002-003
---

# W05-E04-S002-T003 — Singleflight-collapse of concurrent misses

## Task Definition

### Task objective

Collapse N concurrent cache misses for the same key into a single DB load.

### Parent story

W05-E04-S002 — Bounded, epoch-invalidated authorization cache.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E04-S002-T001 (the bounded cache this task extends).

### Detailed work

1. Implement singleflight-style miss collapsing (e.g. via `golang.org/x/sync/singleflight` or an
   equivalent mechanism).
2. Write a test proving N concurrent misses for the same key produce exactly 1 DB load, producing
   `SEC-04/singleflight-tests.md`.
3. Document the mechanism.

### Expected files or components affected

`kernel/authz/caching.go` (extended).

### Expected output

N concurrent misses collapse to 1 DB load.

### Required artifacts

ART-W05-E04-S002-003.

### Required evidence

EV-W05-E04-S002-003.

### Related acceptance criteria

AC-W05-E04-S002-02.

### Completion criteria

The singleflight test confirms exactly 1 DB load for N concurrent misses.

### Verification method

Direct execution of the test producing `SEC-04/singleflight-tests.md`.

### Risks

Low, per PLAN T3's own risk column.

### Rollback or recovery considerations

If the test reveals more than 1 DB load under concurrency, fix before proceeding.

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
| AC-W05-E04-S002-02 | Run the singleflight test | Local dev or CI, Go toolchain | N concurrent misses → 1 DB load | test report | unassigned |

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
