---
id: W04-E03-S002-T003
type: task
title: Item idempotency keys, finalize fencing, retry policy, cancellation
status: done
parent_story: W04-E03-S002
owner: W04BulkSafety
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E03-S002-T002
acceptance_criteria:
  - AC-W04-E03-S002-03
artifacts:
  - ART-W04-E03-S002-003
evidence:
  - EV-W04-E03-S002-003
---

# W04-E03-S002-T003 — Item idempotency keys, finalize fencing, retry policy, cancellation

## Task Definition

### Task objective

Add per-item idempotency keys, reuse DATA-02's finalize-fencing logic, implement retry policy and
operation-level cancellation.

### Status

done.

### Completion criteria

- Each `bulk_item` carries an `idempotency_key` populated by `Start`.
- `runItem` finalizes with lease token/generation/expiry fencing.
- Retry policy respects `bulk_operations.max_attempts`.
- Cancellation marks pending items cancelled and stops the operation.

## Implementation Record

- `bulk.Item` includes `IdempotencyKey uuid.UUID`.
- `Start` generates a deterministic idempotency key per item (`uuidv5(bulkID, seq)`).
- `runItem` uses `Service.mark` and `recordFailure`:
  - Success path: `UPDATE ... SET status='done' ... WHERE status='running' AND lease_token=$1 AND
    lease_generation=$2 AND lease_expires_at > now()`.
  - Failure path: retries if `attempts < maxAttempts`; dead after exhaustion.
- `Cancel` marks operation `cancelled` and pending items `cancelled`.
- `Process` checks operation status before each batch and exits early on cancelled.

## Verification Record

- `TestIntegrationBulkFencedFinalizeRejectsStaleWorker`: stale finalize rejected.
- `TestIntegrationBulkRetryThenFail`: retries until `maxAttempts` then fails.
- `TestIntegrationBulkIdempotencyKeyPassedToWorker`: worker receives expected idempotency key.
- `TestIntegrationBulkPartialFailureLedger`: failure accounting correct.

### Pass or fail

Pass.

### Evidence identifier

EV-W04-E03-S002-003.

## Deviations Record

None.
