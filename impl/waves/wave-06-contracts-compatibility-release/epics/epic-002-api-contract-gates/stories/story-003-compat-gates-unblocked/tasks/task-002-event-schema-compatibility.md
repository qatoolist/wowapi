---
id: W06-E02-S003-T002
type: task
title: Event/schema compatibility (blocked on W06-E01-S001 and W05-E03)
status: blocked
parent_story: W06-E02-S003
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W06-E01-S001
  - W05-E03
acceptance_criteria:
  - AC-W06-E02-S003-02
artifacts:
  - ART-W06-E02-S003-002
evidence:
  - EV-W06-E02-S003-002
---

# W06-E02-S003-T002 — Event/schema compatibility (blocked on W06-E01-S001 and W05-E03)

## Task Definition

### Task objective

Implement an event/schema compatibility check tied to a Compatibility mode. BLOCKED: cannot begin until both W06-E01-S001 (DX-03 design) and W05-E03 (AR-03 remainder) reach accepted.

### Parent story

W06-E02-S003

### Owner

unassigned

### Status

todo

### Dependencies

W06-E01-S001 (DX-03 design record) AND W05-E03 (AR-03 remainder, cross-wave) must both be `accepted` — the compatibility-mode concept this task requires does not exist in current source per MATRIX CS-15's own framing ('premature against today's stringly-typed event registry'). This task must not begin before both entry criteria are satisfied.

### Detailed work

1. Confirm both W06-E01-S001 and W05-E03 have reached `accepted`.
2. Determine, from AR-03's actual delivered shape, what compatibility-mode concept exists to tie this
   check to (this task's own scope is bounded by what AR-03 actually delivers, not by DX-03's
   design-only record).
3. Build the compatibility check against that concept.
4. Write a seeded breaking-event fixture and confirm the gate fails it when the compatibility mode is
   declared.

### Expected files or components affected

A new package/CI job (exact shape TBD pending AR-03's delivered form).

### Expected output

A gate failing a seeded breaking-event fixture when the compatibility mode is declared.

### Required artifacts

ART-W06-E02-S003-002 (event/schema compatibility-check mechanism).

### Required evidence

EV-W06-E02-S003-002 (seeded breaking-event-fixture test report).

### Related acceptance criteria

AC-W06-E02-S003-02.

### Completion criteria

The seeded breaking-event fixture fails the gate when the compatibility mode is declared.

### Verification method

Direct execution of the gate against the fixture, once unblocked.

### Risks

The primary risk is beginning this task before both entry criteria are genuinely satisfied, or implementing against an assumed DX-03 shape that does not match AR-03's actual delivered form — mitigated by this task's own explicit scope-boundary statement.

### Rollback or recovery considerations

If begun prematurely, or if implemented against an incorrect assumed AR-03 shape, halt and record a deviation.

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

*Not yet implemented.*

### Observability changes

*Not yet implemented.*

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

*Not yet implemented. Once implementation occurs, record whether it matched `plan.md`; if not,
reference the corresponding entry in `deviations.md`.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S003-02 | Run the gate against the seeded fixture, once both entry criteria are satisfied | CI | Breaking fixture fails the gate when the compatibility mode is declared | CI gate test report | unassigned |

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
