---
id: W07-E01-S003-T004
type: task
title: Batch-loaded RetryOutbound endpoints
status: complete
parent_story: W07-E01-S003
owner: W07-Scoping-Dispatch.W07E01S003
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E01-S003-04
artifacts:
  - ART-W07-E01-S003-004
evidence:
  - EV-W07-E01-S003-004
---

# W07-E01-S003-T004 — Batch-loaded RetryOutbound endpoints

## Task Definition

### Task objective

Batch-load endpoints in RetryOutbound via one IN (...) query per invocation, not per-delivery.

### Parent story

W07-E01-S003

### Owner

W07-Scoping-Dispatch.W07E01S003

### Status

complete

### Dependencies

None.

### Detailed work

1. Convert RetryOutbound's per-delivery endpoint query to a single IN (...) query per invocation.
2. Write a query-count assertion test (N rows / M endpoints → 1 query).

### Expected files or components affected

webhook.RetryOutbound's implementation file.

### Expected output

One IN (...) query per invocation, not per-delivery.

### Required artifacts

ART-W07-E01-S003-004 (batch-loaded RetryOutbound code).

### Required evidence

EV-W07-E01-S003-004 (query-count test output).

### Related acceptance criteria

AC-W07-E01-S003-04

### Completion criteria

The query-count assertion test confirms exactly one query per invocation.

### Verification method

Direct execution of the query-count assertion test.

### Risks

Low — directive suggests caching immutable endpoints by version too, per PLAN T4's own note (an optional further optimization, not required for this task's own acceptance).

### Rollback or recovery considerations

If batch-loading proves incompatible with an existing caller assumption, escalate rather than silently reintroducing per-delivery queries.

## Implementation Record

### What was actually implemented

`RetryOutbound` collects endpoint IDs from the claimed delivery batch, loads all endpoints once with
`WHERE id = ANY($1)`, builds an ID map, and performs delivery without endpoint reads in the loop.

### Components changed

Webhook outbound retry worker.

### Files changed

`foundation/webhook/service.go`, `foundation/webhook/retry_perf_test.go`.

### Interfaces introduced or changed

No public interface changed.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

Endpoint loading remains inside the existing tenant transaction.

### Observability changes

Covered by T6.

### Tests added or modified

Real-Postgres query tracing with 10 deliveries across 3 endpoints.

### Commits

Working tree based on entry SHA `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

The existing delivery-attempt transaction boundary is unchanged by PERF-04.

### Follow-up items

None.

### Relationship to the approved plan

Matches plan T4.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-04 | pgx query trace over 10 deliveries / 3 endpoints | PostgreSQL 16.9 | Exactly one batch endpoint load and no per-delivery endpoint load | query-count report | W07-Scoping-Dispatch.W07E01S003ReviewR |

### Actual result

PASS: one `webhook_endpoints WHERE id = ANY(uuid[])` query and zero `WHERE id = $1` endpoint queries.

### Pass or fail

PASS.

### Evidence identifier

EV-W07-E01-S003-004.

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

Focused webhook package passed.

### Final conclusion

AC-04 accepted.
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
