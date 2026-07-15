---
id: W04-E03-S002-T002
type: task
title: Atomic leased claim, bounded batch
status: done
parent_story: W04-E03-S002
owner: W04BulkSafety
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E03-S002-T001
acceptance_criteria:
  - AC-W04-E03-S002-02
artifacts:
  - ART-W04-E03-S002-002
evidence:
  - EV-W04-E03-S002-002
---

# W04-E03-S002-T002 — Atomic leased claim, bounded batch

## Task Definition

### Task objective

Replace the plain unlocked `SELECT ... LIMIT 1` claim path with an atomic
`UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`
statement, bounded to a configured batch size, and prove it uses `SKIP LOCKED`.

### Status

done.

### Completion criteria

- Claim SQL uses `FOR UPDATE SKIP LOCKED`.
- Batch size is configurable (default 10).
- Concurrent N>1 claimers receive disjoint batches.
- `runItem`'s pre-existing idempotent completion CAS guard is preserved.

## Implementation Record

- Added `claimSQL` constant using CTE + `FOR UPDATE SKIP LOCKED` + `UPDATE ... FROM ... RETURNING`.
- Added `claimBatch` method that leases up to `Service.batchSize` items atomically.
- Added `WithBatchSize` option (default 10).
- Added `ExplainClaimPlan` helper for EXPLAIN assertion.
- Replaced `Service.next` with `claimBatch`; removed S001 stopgap CAS lock.
- Preserved `runItem` completion guard: UPDATE checks `status='running'`, `lease_token`,
  `lease_generation`, and `lease_expires_at > now()`.

## Verification Record

- `TestIntegrationBulkExplainUsesSkipLocked`: EXPLAIN shows `LockRows` node.
- `TestIntegrationBulkConcurrentClaimers`: 4 workers process 10 items with no duplicates.
- `TestIntegrationBulkAllSucceed`, `TestIntegrationBulkChunkedResumable`: existing behavior preserved.

### Pass or fail

Pass.

### Evidence identifier

EV-W04-E03-S002-002.

## Deviations Record

None.
