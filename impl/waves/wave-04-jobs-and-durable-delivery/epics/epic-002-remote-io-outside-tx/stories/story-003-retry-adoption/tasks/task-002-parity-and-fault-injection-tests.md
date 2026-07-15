---
id: W04-E02-S003-T002
type: task
title: Retry-schedule-parity and fault-injection tests
status: done
parent_story: W04-E02-S003
owner: W04-Rerun
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S003-T001
acceptance_criteria:
  - AC-W04-E02-S003-01
  - AC-W04-E02-S003-02
artifacts:
  - ART-W04-E02-S003-002
evidence:
  - EV-W04-E02-S003-001
  - EV-W04-E02-S003-002
---

# W04-E02-S003-T002 — Retry-schedule-parity and fault-injection tests

## Task Definition

### Task objective

Write and pass a retry-schedule-parity test proving `cenkalti/backoff/v5`'s configured behavior at
both replaced call sites matches or documented-ly improves on each prior hand-rolled schedule's
baseline, and a fault-injection test proving correct retry/backoff behavior under induced remote-call
failure — per REVIEW §O's own required test coverage: "Tests: retry-schedule parity + fault
injection."

### Parent story

W04-E02-S003 — Adopt cenkalti/backoff/v5 for duplicated retry logic.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E02-S003-T001 (both replacements must exist before their behavior can be tested).

### Detailed work

1. For each of the two replaced call sites, write a retry-schedule-parity test comparing the new
   library's observed behavior (attempt count, backoff timing/growth) against T001's documented
   baseline for that call site's original hand-rolled schedule.
2. For each of the two replaced call sites, write a fault-injection test that induces remote-call
   failure (transient and permanent failure modes) and confirms: correct attempt count, correct
   backoff timing between attempts, and correct terminal (give-up) behavior when retries are
   exhausted.
3. Confirm both test suites assert real, specific behavior (exact attempt counts, timing bounds, or
   terminal-state checks) rather than a superficial "does not error" assertion.
4. Record both test suites' output as evidence.

### Expected files or components affected

New test files for both replaced call sites (exact paths TBD, following T001's location discovery).

### Expected output

A passing retry-schedule-parity test and a passing fault-injection test for both replaced call
sites.

### Required artifacts

ART-W04-E02-S003-002 (retry-schedule-parity and fault-injection test suites).

### Required evidence

EV-W04-E02-S003-001 (retry-schedule-parity test report), EV-W04-E02-S003-002 (fault-injection test
report).

### Related acceptance criteria

AC-W04-E02-S003-01, AC-W04-E02-S003-02.

### Completion criteria

Both test suites pass for both replaced call sites, with meaningful assertions on schedule parity
and fault-injection behavior, not superficial checks.

### Verification method

Direct execution of both test suites against both replaced call sites.

### Risks

A parity test could pass while subtly mis-configuring the new library's backoff parameters relative
to the original schedule's actual intent, if that intent was itself under-documented before this
task's own baseline-recording step (T001, step 2) — mitigated by requiring the parity test to assert
specific, documented values, not a loose approximation.

### Rollback or recovery considerations

Not applicable — a failing test blocks this story's acceptance until the underlying configuration
(T001) is corrected; it does not itself get rolled back.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not applicable.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable.*

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
| AC-W04-E02-S003-01 | Run retry-schedule-parity test for both call sites | Local dev or CI, Go toolchain | New library's schedule matches or documented-ly improves on prior baseline | test report | unassigned |
| AC-W04-E02-S003-02 | Run fault-injection test for both call sites | Local dev or CI, Go toolchain, fault-injection harness | Correct attempt count, backoff timing, terminal behavior on exhausted retries | test report | unassigned |

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
