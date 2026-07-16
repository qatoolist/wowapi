---
id: W02-E01-S001-T002
type: task
title: Lock-timeout enforcement mechanism
status: done
parent_story: W02-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W02-E01-S001-03
artifacts:
  - ART-W02-E01-S001-003
  - ART-W02-E01-S001-004
evidence:
  - EV-W02-E01-S001-003
---

# W02-E01-S001-T002 — Lock-timeout enforcement mechanism

## Task Definition

### Task objective

Implement a 2-second online-DDL lock-timeout enforcement mechanism with clean abort-and-retry and
an explicit, bounded retry ceiling, so that an online-classified DDL statement cannot hold a lock
beyond its budget or retry indefinitely.

### Parent story

W02-E01-S001 — Migration manifest schema and online-DDL lock budget.

### Owner

unassigned

### Status

todo

### Dependencies

None (parallel-safe with T001 — disjoint code surface; both may proceed independently, subject to
sharing the same story's `plan.md` design context).

### Detailed work

1. Implement a lock-timeout wrapper around DDL execution for online-classified migrations, budgeted
   at 2 seconds per PLAN DATA-09 T2's acceptance criterion.
2. Implement abort-and-retry behavior: on timeout, abort the statement cleanly (no partial DDL
   applied) and retry, bounded by an explicit retry ceiling (exact bound to be chosen and
   documented — "human-set retry ceiling" per PLAN T2).
3. Write a test against a deliberately concurrently-locked table (a held lock on the target table
   from a separate connection), confirming: the statement aborts cleanly within the 2-second
   budget, no partial DDL is applied, and the retry loop respects the bounded ceiling rather than
   retrying indefinitely.
4. Add observability (at minimum, a log line) for each abort/retry event, so an operator can
   distinguish "succeeded on retry N" from "retrying indefinitely."
5. Document the lock-timeout budget, abort-and-retry behavior, and retry ceiling.

### Expected files or components affected

A new lock-timeout enforcement mechanism (exact file path TBD, expected near the existing migration-
execution code, not yet confirmed by file/line per `plan.md`'s "Unresolved questions").

### Expected output

A lock-timeout wrapper enforcing the 2-second budget with bounded abort-and-retry, proven by a test
against a concurrently-locked table.

### Required artifacts

ART-W02-E01-S001-003 (lock-timeout enforcement mechanism), ART-W02-E01-S001-004 (documentation,
shared with T001).

### Required evidence

EV-W02-E01-S001-003 (concurrently-locked-table lock-timeout test output).

### Related acceptance criteria

AC-W02-E01-S001-03.

### Completion criteria

A DDL statement against a deliberately concurrently-locked table aborts cleanly within the 2-second
budget with no partial DDL, and the retry loop is bounded — proven by a passing test.

### Verification method

Direct execution of the concurrency test against a live PostgreSQL instance with a deliberately held
lock, logged output retained as evidence.

### Risks

The bounded retry ceiling is the required DoS-prevention control (PLAN T2's own risk note: "Bound
total retries — unbounded retry is a deploy-time DoS") — an incorrectly-unbounded implementation
would be a security-relevant defect, not merely a quality gap.

### Rollback or recovery considerations

Revert the lock-timeout wrapper if it produces false-positive aborts against a legitimately-fast DDL
statement under normal (non-adversarial) load; escalate for redesign rather than silently widening
the budget without recording why.

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

*Not yet implemented — the bounded retry ceiling is itself the security control; recorded here once
implemented.*

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
| AC-W02-E01-S001-03 | Run the lock-timeout mechanism against a deliberately concurrently-locked table | Local dev or CI, PostgreSQL instance with a held lock on the target table | Clean abort within 2s budget, no partial DDL, bounded retry | integration-test report | unassigned |

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
