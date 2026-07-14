---
id: W02-E03-S001-T005
type: task
title: kernel/artifact.Generate mirror fix and dedicated test
status: todo
parent_story: W02-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E03-S001-T001
acceptance_criteria:
  - AC-W02-E03-S001-05
artifacts:
  - ART-W02-E03-S001-001
evidence:
  - EV-W02-E03-S001-005
---

# W02-E03-S001-T005 — kernel/artifact.Generate mirror fix and dedicated test

## Task Definition

### Task objective

Confirm and independently prove that `kernel/artifact.Generate`'s version allocation, fixed as part
of T001's counter/sequence mechanism, meets the same concurrency bar as T001's own acceptance
criterion — via a dedicated mirror test scoped specifically to `kernel/artifact.Generate`, not merely
relying on T001's own (both-package) test coverage.

### Parent story

W02-E03-S001 — Version-allocation races and upload-blob GC.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E03-S001-T001 (per PLAN DATA-05 T5's own Depends-on column: "T1" — T5 mirrors T1's mechanism, it
does not introduce a new one).

### Detailed work

1. Confirm `kernel/artifact.Generate` uses the same counter/sequence mechanism T001 introduced (not a
   parallel, independently-drifted implementation).
2. Write a dedicated concurrency test for `kernel/artifact.Generate` specifically — at least 20
   concurrent callers, confirming N unique, monotonic versions with zero unexpected conflicts —
   independent of T001's own test (which may already exercise `kernel/artifact.Generate` as part of
   its "both packages" scope, but T005's own test must exist as a standalone, dedicated proof per
   PLAN T5's own "Tests" column: "Mirror test").
3. Confirm this dedicated test is not redundant busywork duplicating T001's test verbatim — per
   `epic.md`'s "Architectural context," the point of T5 is to ensure `kernel/artifact`'s own proof
   cannot be silently skipped if T001's test happens to focus on `kernel/document`; if T001's test
   already independently and adequately covers `kernel/artifact.Generate` at the required bar, T005's
   own test may be the same test file re-run in isolation against `kernel/artifact` alone, provided
   the isolation itself (proving `kernel/artifact.Generate` passes on its own, not merely as an
   incidental side effect of a combined test run) is genuine — record whichever approach is taken in
   this task's own Implementation Record, not silently.

### Expected files or components affected

`kernel/artifact`'s `Generate` version-allocation code path (already modified by T001; this task adds
no further code change to the mechanism itself, only its dedicated proof).

### Expected output

A dedicated concurrency test proving `kernel/artifact.Generate` independently meets the same
concurrency bar as T001's own acceptance criterion.

### Required artifacts

ART-W02-E03-S001-001 (the counter/sequence mechanism, shared with T001 — no new artifact introduced
by this task beyond its own test).

### Required evidence

EV-W02-E03-S001-005 (the dedicated `kernel/artifact.Generate` mirror concurrency test output).

### Related acceptance criteria

AC-W02-E03-S001-05.

### Completion criteria

A dedicated concurrency test against `kernel/artifact.Generate` alone passes at the same bar as
AC-W02-E03-S001-01 (≥20 concurrent callers, N unique monotonic versions, zero unexpected conflicts).

### Verification method

Direct execution of the dedicated mirror concurrency test against a live PostgreSQL instance,
independent of T001's own combined test run.

### Risks

None beyond T001's own RISK-W02-E03-001 (counter-row contention), which this task's test also
implicitly re-exercises for `kernel/artifact.Generate` specifically.

### Rollback or recovery considerations

Not applicable beyond T001's own rollback considerations — this task adds a test, not new production
code.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable — no interface change beyond T001's own.*

### Configuration changes

*Not applicable.*

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
| AC-W02-E03-S001-05 | Run the dedicated `kernel/artifact.Generate` mirror concurrency test | Local dev or CI, PostgreSQL instance | Same concurrency bar as AC-W02-E03-S001-01, proven independently | concurrency-test report | unassigned |

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
