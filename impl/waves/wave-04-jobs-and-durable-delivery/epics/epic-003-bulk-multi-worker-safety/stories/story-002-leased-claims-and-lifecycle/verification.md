---
id: VER-W04-E03-S002
type: verification-record
parent_story: W04-E03-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W04-E03-S002

## Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E03-S002-01 | Run the `bulk_items` lease-column migration test | Local dev or CI, PostgreSQL instance | Lease columns exist and match the shared primitive's schema contract | migration-test report | W04BulkSafety |
| AC-W04-E03-S002-02 | Run the `EXPLAIN`-plan assertion against the leased-claim SQL statement and the concurrent `N>1` claimer test | Local dev or CI, PostgreSQL instance | `EXPLAIN` shows row locking; no two concurrent claimers receive the same row; completion CAS guard unchanged | EXPLAIN-plan assertion + concurrency-test report | W04BulkSafety |
| AC-W04-E03-S002-03 | Run the fenced-finalize-rejection test plus idempotency-key/retry/cancellation tests | Local dev or CI, PostgreSQL instance | A fenced worker's finalize write is rejected; idempotency, retry, and cancellation behave correctly | test report | W04BulkSafety |
| AC-W04-E03-S002-04 | Run the lifecycle integration tests | Local dev or CI, PostgreSQL instance | Pause/resume/cancel controls behave correctly mid-run against bounded batch claims | integration-test report | W04BulkSafety |
| AC-W04-E03-S002-05 | Run the named chaos test `DATA-04/chaos/duplicate_worker_test.go` | Local dev or CI, PostgreSQL instance | ≥2 processors concurrently claim/retry/pause/resume/cancel the same operation with zero duplicate effects and no stale finalization | named chaos-test report | W04BulkSafety |

## Post-execution record

### Actual result

All five acceptance criteria verified by passing tests.

### Pass or fail

Pass.

### Evidence identifier

EV-W04-E03-S002-001 through EV-W04-E03-S002-006.

### Execution date

2026-07-13.

### Commit or revision

HEAD (working tree).

### Environment

Local PostgreSQL via `make up`; `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`.

### Reviewer

W04BulkSafety.

### Findings

None.

### Retest status

N/A.

### Final conclusion

Accepted.
