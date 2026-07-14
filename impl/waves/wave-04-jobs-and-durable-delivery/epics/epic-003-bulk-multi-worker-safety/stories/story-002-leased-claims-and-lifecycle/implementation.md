---
id: IMPL-W04-E03-S002
type: implementation-record
parent_story: W04-E03-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W04-E03-S002

## What was actually implemented

- Migration `00044_bulk_items_lease_and_lifecycle.sql` adds lease columns (`lease_token`,
  `lease_generation`, `lease_expires_at`), `idempotency_key`, and `max_attempts` to
  `bulk_operations`; extends status constraints for pause/resume/cancel; drops superseded S001
  stopgap columns.
- `kernel/bulk/bulk.go` rewritten to use an atomic leased-claim SQL statement:
  `UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`.
- `Service` options added: `WithBatchSize`, `WithLeaseTTL`, `WithMaxAttempts`, `WithLogger`.
- `Item` struct now carries `Lease` and `IdempotencyKey` and is passed to worker functions.
- Finalize fencing implemented in `runItem` and `recordFailure`: success/failure UPDATEs check
  `status='running'`, `lease_token`, `lease_generation`, and `lease_expires_at > now()`.
- Operation-level lifecycle controls: `Pause`, `Resume`, `Cancel`; `Process` respects operation
  status mid-run.
- Named chaos test at `kernel/bulk/chaos/duplicate_worker_test.go`.

## Components changed

- `kernel/bulk`
- `migrations`
- `kernel` (constructor option wiring for logger)

## Files changed

- `migrations/00044_bulk_items_lease_and_lifecycle.sql`
- `kernel/bulk/bulk.go`
- `kernel/bulk/bulk_test.go`
- `kernel/bulk/bulk_cov_test.go`
- `kernel/bulk/chaos/duplicate_worker_test.go` (new)
- `kernel/kernel.go`

## Interfaces introduced or changed

- `bulk.ItemFunc` signature changed to `(context.Context, database.TenantDB, bulk.Item) error`.
- `bulk.Item` struct added.
- New `bulk.Option` constructors.

## Configuration changes

None.

## Schema or migration changes

Migration `00044`.

## Security changes

None beyond correctness fencing.

## Observability changes

`WithLogger` option wired; claim and finalize errors logged.

## Tests added or modified

- `TestIntegrationBulkLeaseColumnsExist`
- `TestIntegrationBulkExplainUsesSkipLocked`
- `TestIntegrationBulkConcurrentClaimers`
- `TestIntegrationBulkFencedFinalizeRejectsStaleWorker`
- `TestIntegrationBulkIdempotencyKeyPassedToWorker`
- `TestIntegrationBulkPauseResumeCancel`
- `TestIntegrationBulkDuplicateWorkerChaos`
- Existing bulk tests updated for new `ItemFunc` signature.

## Commits

Working tree changes.

## Pull requests

None in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

- Named chaos test is self-contained; should be migrated to the shared `W04-E01-S003` harness once
  it lands.

## Follow-up items

- Migrate `kernel/bulk/chaos/duplicate_worker_test.go` onto shared chaos harness.

## Relationship to the approved plan

Matches plan, with one deviation recorded in `deviations.md` (shared chaos harness unavailable).
