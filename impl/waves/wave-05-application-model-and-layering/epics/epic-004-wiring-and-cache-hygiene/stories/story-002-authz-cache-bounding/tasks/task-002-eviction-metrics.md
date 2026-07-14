---
id: W05-E04-S002-T002
type: task
title: Eviction with admission/eviction metrics
status: todo
parent_story: W05-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E04-S002-T001
acceptance_criteria:
  - AC-W05-E04-S002-01
artifacts:
  - ART-W05-E04-S002-002
evidence:
  - EV-W05-E04-S002-002
---

# W05-E04-S002-T002 — Eviction with admission/eviction metrics

## Task Definition

### Task objective

Ensure idle cache entries are evicted, and expose a full admission/eviction metric set.

### Parent story

W05-E04-S002 — Bounded, epoch-invalidated authorization cache.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E04-S002-T001 (the bounded cache this task instruments).

### Detailed work

1. Confirm `golang-lru/v2`'s own eviction behavior meets the idle-entry-eviction requirement, or add
   a TTL-based supplementary eviction if needed.
2. Implement admission and eviction metrics (counts, rates, or equivalent, consistent with this
   framework's existing observability conventions).
3. Write the eviction-metrics test, producing `SEC-04/eviction-metrics-tests.md`.
4. Document the metric set.

### Expected files or components affected

`kernel/authz/caching.go` (extended).

### Expected output

Idle entries evicted; a full admission/eviction metric set exposed.

### Required artifacts

ART-W05-E04-S002-002.

### Required evidence

EV-W05-E04-S002-002.

### Related acceptance criteria

AC-W05-E04-S002-01.

### Completion criteria

The eviction-metrics test confirms both idle-entry eviction and the full metric set.

### Verification method

Direct execution of the test producing `SEC-04/eviction-metrics-tests.md`.

### Risks

Low, per PLAN T2's own risk column.

### Rollback or recovery considerations

If metrics are found incomplete, extend before proceeding to T003.

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

*Not yet implemented — the admission/eviction metric set; recorded here once implemented.*

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
| AC-W05-E04-S002-01 | Run the eviction-metrics test | Local dev or CI, Go toolchain | Idle entries evicted; full metric set present | test report | unassigned |

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
