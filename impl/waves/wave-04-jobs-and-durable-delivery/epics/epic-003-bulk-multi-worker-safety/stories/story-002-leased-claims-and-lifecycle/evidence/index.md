---
id: W04-E03-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W04-E03-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03-S002 — Evidence index

Per mandate §10.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W04-E03-S002-001 | migration-test report (`bulk_items` lease columns) | W04-E03-S002-T001 | AC-W04-E03-S002-01 | `cd kernel/bulk && DATABASE_URL=... go test -run TestIntegrationBulkLeaseColumnsExist -count=1 -v .` | HEAD | `evidence/bulk_tests.log` | accepted |
| EV-W04-E03-S002-002 | EXPLAIN-plan `SKIP LOCKED` assertion + concurrent N>1 claimer test report | W04-E03-S002-T002 | AC-W04-E03-S002-02 | `cd kernel/bulk && DATABASE_URL=... go test -run 'TestIntegrationBulkExplainUsesSkipLocked|TestIntegrationBulkConcurrentClaimers' -count=1 -v .` | HEAD | `evidence/bulk_tests.log` | accepted |
| EV-W04-E03-S002-003 | fenced-finalize-rejection test report + idempotency/retry/cancellation test report | W04-E03-S002-T003 | AC-W04-E03-S002-03 | `cd kernel/bulk && DATABASE_URL=... go test -run 'TestIntegrationBulkFencedFinalizeRejectsStaleWorker|TestIntegrationBulkRetryThenFail|TestIntegrationBulkIdempotencyKeyPassedToWorker' -count=1 -v .` | HEAD | `evidence/bulk_tests.log` | accepted |
| EV-W04-E03-S002-004 | lifecycle integration-test report (pause/resume/cancel) | W04-E03-S002-T004 | AC-W04-E03-S002-04 | `cd kernel/bulk && DATABASE_URL=... go test -run TestIntegrationBulkPauseResumeCancel -count=1 -v .` | HEAD | `evidence/bulk_tests.log` | accepted |
| EV-W04-E03-S002-005 | named chaos-test report (`DATA-04/chaos/duplicate_worker_test.go`) | W04-E03-S002-T005 | AC-W04-E03-S002-05 | `cd kernel/bulk/chaos && DATABASE_URL=... go test -run TestIntegrationBulkDuplicateWorkerChaos -count=1 -v .` | HEAD | `evidence/duplicate_worker_chaos_test.log` | accepted |
| EV-W04-E03-S002-006 | full code diff for S002 | — | — | `git diff` | HEAD | `evidence/s002_changes.diff` | accepted |

All five acceptance criteria verified:
- AC-W04-E03-S002-01: `bulk_items` carries `lease_token`, `lease_generation`, `lease_expires_at`, and `idempotency_key`; `bulk_operations` carries `max_attempts`.
- AC-W04-E03-S002-02: `EXPLAIN` shows `LockRows` (FOR UPDATE evidence); `TestIntegrationBulkConcurrentClaimers` proves 4 workers receive disjoint batches.
- AC-W04-E03-S002-03: `TestIntegrationBulkFencedFinalizeRejectsStaleWorker` proves stale finalize rejected; retry and idempotency keys tested.
- AC-W04-E03-S002-04: `TestIntegrationBulkPauseResumeCancel` exercises pause/resume/cancel.
- AC-W04-E03-S002-05: `kernel/bulk/chaos/duplicate_worker_test.go` passes with ≥2 processors claiming/retrying/pausing/resuming/cancelling without duplicate effects or stale finalization.
