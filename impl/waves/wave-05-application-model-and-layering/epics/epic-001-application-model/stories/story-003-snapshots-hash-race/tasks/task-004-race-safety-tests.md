---
id: W05-E01-S003-T004
type: task
title: Race-test suite
status: todo
parent_story: W05-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E01-S003-T001
  - W05-E01-S003-T002
  - W05-E01-S003-T003
acceptance_criteria:
  - AC-W05-E01-S003-03
artifacts:
  - ART-W05-E01-S003-004
evidence:
  - EV-W05-E01-S003-004
---

# W05-E01-S003-T004 — Race-test suite

## Task Definition

### Task objective

Write race tests proving no runtime mutation of the sealed model: `go test -race` is clean on
concurrent legitimate reads; an illegitimate write fails via T002's rejection mechanism, not as an
unguarded data race.

### Parent story

W05-E01-S003 — Snapshot immutability, post-seal rejection, model hash, and race safety.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E01-S003-T001, W05-E01-S003-T002, W05-E01-S003-T003 (PLAN T10's own dependency row: "T1-T9" —
the full preceding surface, including the model hash, should be stable).

### Detailed work

1. Write a race test exercising concurrent legitimate reads of the sealed model from multiple
   goroutines, run under `go test -race`.
2. Write a race test exercising an illegitimate write attempt (a retained registrar calling a
   mutation method) concurrently with legitimate reads, confirming the write fails cleanly via
   T002's rejection mechanism (an error return) rather than manifesting as an unguarded data race
   `go test -race` would flag.
3. Capture the race-test run output as `AR-01/race_test_output.txt`.
4. Document the race-safety guarantee.

### Expected files or components affected

Test files exercising the `ApplicationModel`/`Compiler` and the wrapped registries from S001-S003
(exact file paths TBD per `plan.md`).

### Expected output

`go test -race` clean on concurrent legitimate reads; illegitimate writes fail via the rejection
mechanism, not as an unguarded race.

### Required artifacts

ART-W05-E01-S003-004.

### Required evidence

EV-W05-E01-S003-004.

### Related acceptance criteria

AC-W05-E01-S003-03.

### Completion criteria

The race test suite passes clean under `go test -race`, with the illegitimate-write scenario
confirmed to fail via the rejection mechanism (not surfaced as a race by the detector).

### Verification method

Direct execution of the race-test suite under `go test -race`; the output is retained as
`AR-01/race_test_output.txt`.

### Risks

Low, per PLAN T10's own risk column.

### Rollback or recovery considerations

If `go test -race` flags a genuine data race (not the expected rejected-write scenario), treat as a
blocking defect in the sealed model's concurrency safety — do not ship until resolved.

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
| AC-W05-E01-S003-03 | Run the race-test suite under `go test -race` | Local dev or CI, Go toolchain (`-race`) | Clean on concurrent legitimate reads; illegitimate write fails via rejection mechanism, not as a race | race-test report | unassigned |

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
