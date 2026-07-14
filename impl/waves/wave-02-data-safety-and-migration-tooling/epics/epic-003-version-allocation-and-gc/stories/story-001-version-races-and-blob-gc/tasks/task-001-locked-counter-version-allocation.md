---
id: W02-E03-S001-T001
type: task
title: Locked-counter/sequence-row version allocation (both packages)
status: todo
parent_story: W02-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W02-E03-S001-01
artifacts:
  - ART-W02-E03-S001-001
  - ART-W02-E03-S001-005
evidence:
  - EV-W02-E03-S001-001
---

# W02-E03-S001-T001 — Locked-counter/sequence-row version allocation (both packages)

## Task Definition

### Task objective

Replace the inline `MAX(version)+1` read in both `kernel/artifact.Generate` and
`kernel/document.InitiateUpload` with a locked parent counter or dedicated per-aggregate sequence
row, so that N concurrent callers produce N unique, monotonic versions with zero unexpected
conflicts.

### Parent story

W02-E03-S001 — Version-allocation races and upload-blob GC.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Re-read `kernel/artifact.Generate` and `kernel/document.InitiateUpload` at this task's actual
   start commit to confirm both still compute version via inline `MAX(version)+1` with no locking or
   sequence primitive (resolving `plan.md`'s current-state re-confirmation step).
2. Draft the counter/sequence mechanism's options (locked parent-row read via
   `SELECT ... FOR UPDATE`, a PostgreSQL `SEQUENCE`, or a dedicated per-aggregate counter table) with
   trade-offs, particularly around lock-wait under concurrent load; select one and document the
   rationale (resolves `plan.md`'s "Unresolved questions" item on mechanism form).
3. Implement the chosen mechanism in `kernel/artifact.Generate`, replacing its inline `MAX()+1` read.
4. Implement the same mechanism in `kernel/document.InitiateUpload`, replacing its own inline
   `MAX()+1` read.
5. Write the concurrency test: at least 20 concurrent callers against both packages' version
   allocation, confirming N unique, monotonic versions with zero unexpected conflicts.
6. Measure and record lock wait as part of the concurrency test's evidence, per RISK-W02-E03-001.
7. Document the counter/sequence mechanism and its concurrency guarantee.

### Expected files or components affected

`kernel/artifact`'s `Generate` version-allocation code path; `kernel/document`'s `InitiateUpload`
version-allocation code path (exact file/line to be re-confirmed at this task's actual start commit);
a new schema migration for the counter/sequence mechanism.

### Expected output

A locked-counter/sequence-row version-allocation mechanism applied to both packages, proven race-free
under at least 20 concurrent callers, with lock-wait measured and recorded.

### Required artifacts

ART-W02-E03-S001-001 (the counter/sequence mechanism, code), ART-W02-E03-S001-005 (documentation,
shared with T002/T004).

### Required evidence

EV-W02-E03-S001-001 (concurrency test output, ≥20 concurrent callers, both packages, with measured
lock wait).

### Related acceptance criteria

AC-W02-E03-S001-01.

### Completion criteria

Both `kernel/artifact.Generate` and `kernel/document.InitiateUpload` allocate versions via the new
mechanism; the concurrency test passes with zero unexpected conflicts across at least 20 concurrent
callers; lock wait is measured and recorded as part of the test's evidence.

### Verification method

Direct execution of the concurrency test against a live PostgreSQL instance, with lock-wait
measurements captured in the test's logged output and retained as evidence.

### Risks

RISK-W02-E03-001 (counter-row contention is the new serialization point — PLAN T1's own risk note:
"measure lock wait") — see epic-level `risks.md`.

### Rollback or recovery considerations

Revert to the prior inline `MAX()+1` behavior only as an emergency measure if the new mechanism
produces false-positive allocation failures under normal (non-adversarial) load; this would
reintroduce the known race, so any such rollback must be paired with an explicit, time-bounded
remediation plan, not treated as an acceptable steady state.

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

*Not yet implemented — expected: a new counter/sequence mechanism (table or `SEQUENCE` object).*

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
| AC-W02-E03-S001-01 | Run the concurrency test (≥20 concurrent callers) against both packages' version allocation | Local dev or CI, PostgreSQL instance | N unique, monotonic versions, zero unexpected conflicts; lock wait measured | concurrency-test report | unassigned |

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
