---
id: W07-E01-S003-T001
type: task
title: Bounded batch claiming for SweepSLA
status: complete
parent_story: W07-E01-S003
owner: W07-Scoping-Dispatch.W07E01S003
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E01-S003-01
artifacts:
  - ART-W07-E01-S003-001
evidence:
  - EV-W07-E01-S003-001
---

# W07-E01-S003-T001 — Bounded batch claiming for SweepSLA

## Task Definition

### Task objective

Add LIMIT to both SweepSLA queries; loop via job re-invocation rather than in-memory materialization.

### Parent story

W07-E01-S003

### Owner

W07-Scoping-Dispatch.W07E01S003

### Status

complete

### Dependencies

None.

### Detailed work

1. Add LIMIT to both SweepSLA queries.
2. Implement looping via job re-invocation rather than materializing the full due-row set.
3. Preserve existing idempotency guards, no reintroduced double-remind race.
4. Test at 10/1k/100k due rows for fixed query count and memory.

### Expected files or components affected

The SweepSLA implementation file(s).

### Expected output

Fixed query count and memory across due-row cardinalities.

### Required artifacts

ART-W07-E01-S003-001 (bounded-batch SweepSLA code).

### Required evidence

EV-W07-E01-S003-001 (cardinality test output).

### Related acceptance criteria

AC-W07-E01-S003-01

### Completion criteria

Fixed query count and memory hold at 10/1k/100k due rows.

### Verification method

Direct execution of the cardinality test suite.

### Risks

Medium — must preserve existing idempotency guards, no reintroduced double-remind race, per PLAN T1's own risk note.

### Rollback or recovery considerations

If a double-remind race is found, revert to the unbounded loop while re-diagnosing; do not ship a faster-but-racy version.

## Implementation Record

### What was actually implemented

`SweepSLA` now atomically claims at most 100 reminders and 100 escalations per invocation.
The registered maintenance job drains additional work by normal scheduler re-invocation rather than
materializing an unbounded due set in one call.

### Components changed

Workflow SLA sweeper and its registered maintenance job.

### Files changed

`kernel/workflow/sweeper.go`, `app/maintenance.go`, `kernel/workflow/sweeper_perf_test.go`.

### Interfaces introduced or changed

No public interface changed. `SweepSLA` still returns reminder and escalation counts for this bounded
invocation.

### Configuration changes

None.

### Schema or migration changes

Covered by W07-E01-S003-T003.

### Security changes

Tenant-bound `database.TenantDB` and RLS scope are preserved.

### Observability changes

Covered by W07-E01-S003-T006.

### Tests added or modified

Real-Postgres cardinality tiers, safe 100/100/1/0 reinvocation, and concurrent no-double-remind tests.

### Commits

Working tree based on entry SHA `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13 through 2026-07-14.

### Technical debt introduced

None.

### Known limitations

Batch size is intentionally fixed at 100; the job scheduler controls repeated progress.

### Follow-up items

None within PERF-04.

### Relationship to the approved plan

Matches plan T1; no deviation was required.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-01 | Real-Postgres cardinality, reinvocation, concurrency tests and benchmark tiers | PostgreSQL 16.9, `WOWAPI_REQUIRE_DB=1` | Fixed 100-row ceiling and bounded allocations at 10/1k/100k | test + benchmark report | W07-Scoping-Dispatch.W07E01S003ReviewR |

### Actual result

PASS: all three due-row tiers processed at most 100 rows; traced guard and batch-load statements stayed
fixed; median allocations stayed about 9.8k at 1k and 100k; reinvocation drained without duplicates.

### Pass or fail

PASS.

### Evidence identifier

EV-W07-E01-S003-001.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on entry SHA `733ef3e`.

### Environment

darwin/arm64, Go 1.26.5, real PostgreSQL 16.9 Docker service.

### Reviewer

W07-Scoping-Dispatch.W07E01S003ReviewR.

### Findings

Independent review found no open issues.

### Retest status

Focused workflow package and benchmark budget gate passed.

### Final conclusion

AC-01 accepted.
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
