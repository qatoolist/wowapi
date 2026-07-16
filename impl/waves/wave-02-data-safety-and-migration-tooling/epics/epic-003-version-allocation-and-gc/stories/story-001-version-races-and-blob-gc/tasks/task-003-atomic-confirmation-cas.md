---
id: W02-E03-S001-T003
type: task
title: Atomic confirmation CAS
status: done
parent_story: W02-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E03-S001-T001
  - W02-E03-S001-T002
acceptance_criteria:
  - AC-W02-E03-S001-03
artifacts:
  - ART-W02-E03-S001-003
evidence:
  - EV-W02-E03-S001-003
---

# W02-E03-S001-T003 — Atomic confirmation CAS

## Task Definition

### Task objective

Make confirmation CAS the upload session's status and the version allocation atomically, so that of
two racing confirmation calls against the same session, exactly one succeeds.

### Parent story

W02-E03-S001 — Version-allocation races and upload-blob GC.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E03-S001-T001, W02-E03-S001-T002 (per PLAN DATA-05 T3's own Depends-on column: "T1, T2" —
confirmation has nothing to CAS without T1's version-allocation mechanism and T2's session record
both existing).

### Detailed work

1. Implement the atomic CAS confirmation path: confirmation reads the session, checks its current
   status, and updates the session's status and the version allocation together in one
   transaction/CAS operation.
2. Ensure a losing racer's confirmation attempt is rejected outright by the CAS (no partial update to
   either the session or the version).
3. Write the concurrency test: two racing confirmation calls against the same session, confirming
   exactly one succeeds and the other is cleanly rejected.

### Expected files or components affected

`kernel/document`'s confirmation code path (exact file/line TBD, expected adjacent to
`InitiateUpload`'s existing code).

### Expected output

An atomic CAS confirmation mechanism, proven by the racing-confirmation concurrency test.

### Required artifacts

ART-W02-E03-S001-003 (the atomic CAS confirmation logic).

### Required evidence

EV-W02-E03-S001-003 (racing-confirmation concurrency test output).

### Related acceptance criteria

AC-W02-E03-S001-03.

### Completion criteria

Of two racing confirmation calls against the same session, exactly one succeeds and the other is
cleanly rejected with no partial state change — proven by a passing concurrency test.

### Verification method

Direct execution of the racing-confirmation concurrency test against a live PostgreSQL instance.

### Risks

An incorrect CAS implementation that allows both racers to partially succeed would silently corrupt
either the session state or the version allocation — this is a correctness-critical task, not
merely a convenience feature.

### Rollback or recovery considerations

If the CAS implementation is found to deadlock or produce false-positive rejections against
legitimate, non-racing confirmations, escalate for redesign rather than loosening the atomicity
guarantee, which would reintroduce the exact race this task exists to close.

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

*Not applicable — this task adds confirmation logic against the existing session table from T002; it
introduces no new table of its own.*

### Security changes

*Not applicable.*

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

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E03-S001-03 | Run the racing-confirmation concurrency test | Local dev or CI, PostgreSQL instance | Exactly one of two racing confirms succeeds; the other is cleanly rejected | concurrency-test report | unassigned |

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
