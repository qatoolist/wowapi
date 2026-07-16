---
id: W07-E01-S003-T002
type: task
title: Set-based/batched operation conversion
status: done
parent_story: W07-E01-S003
owner: W07-Scoping-Dispatch.W07E01S003
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W07-E01-S003-T001
acceptance_criteria:
  - AC-W07-E01-S003-02
artifacts:
  - ART-W07-E01-S003-002
evidence:
  - EV-W07-E01-S003-002
---

# W07-E01-S003-T002 — Set-based/batched operation conversion

## Task Definition

### Task objective

Convert per-row UPDATE+load+emit to set-based/batched operations where semantically possible.

### Parent story

W07-E01-S003

### Owner

W07-Scoping-Dispatch.W07E01S003

### Status

complete

### Dependencies

W07-E01-S003-T001 (conversion operates on T001's own bounded-batch structure).

### Detailed work

1. Identify which operations can be converted to set-based (guard-flip UPDATEs, batch-load by ID
   set).
2. Leave emit()/escalation logic per-instance where semantics genuinely require it.
3. Write query-count assertion tests confirming set-based conversion.
4. Confirm idempotency guards preserved.

### Expected files or components affected

The SweepSLA implementation file(s), extended.

### Expected output

Set-based operations where possible; per-instance where semantically required.

### Required artifacts

ART-W07-E01-S003-002 (set-based/batched operation conversions).

### Required evidence

EV-W07-E01-S003-002 (query-count assertion test output).

### Related acceptance criteria

AC-W07-E01-S003-02

### Completion criteria

Set-based conversion confirmed for guard flips and batch-loads; idempotency preserved.

### Verification method

Direct execution of query-count assertion tests.

### Risks

Medium-high — emit()/escalation logic is inherently per-instance, per PLAN T2's own risk note.

### Rollback or recovery considerations

If a set-based conversion is found to break emit/escalation semantics, revert that specific operation to per-instance handling and record why.

## Implementation Record

### What was actually implemented

Reminder and escalation eligibility are now claimed with two atomic `UPDATE ... FROM` statements
using `FOR UPDATE SKIP LOCKED LIMIT 100`. Claimed instance state is loaded with one `ANY(uuid[])`
query, and required definitions are resolved from the registry or one deduplicated database query.

### Components changed

Workflow SLA persistence and materialization.

### Files changed

`kernel/workflow/sweeper.go`, `kernel/workflow/sweeper_perf_test.go`.

### Interfaces introduced or changed

No public interface changed.

### Configuration changes

None.

### Schema or migration changes

None in this task.

### Security changes

Queries continue through the caller's tenant-bound transaction.

### Observability changes

None in this task.

### Tests added or modified

Query tracer assertions, safe reinvocation, concurrent no-double-remind, and exact reminder count tests.

### Commits

Working tree based on entry SHA `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

One instance and definition load occurs for each nonempty reminder/escalation phase; counts are
bounded by the fixed claim ceiling.

### Follow-up items

None.

### Relationship to the approved plan

Matches plan T2 without deviation.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-02 | Traced real-Postgres statements plus concurrent/reinvocation integration tests | PostgreSQL 16.9 | Set-based guard flips and bounded batch loads preserve idempotency | query-count report | W07-Scoping-Dispatch.W07E01S003ReviewR |

### Actual result

PASS: exactly two guard updates at every due-row tier, no per-row instance or definition query,
100/100/1/0 reinvocation, and one reminder per task under concurrent sweepers.

### Pass or fail

PASS.

### Evidence identifier

EV-W07-E01-S003-002.

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

Focused workflow package passed.

### Final conclusion

AC-02 accepted.
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
