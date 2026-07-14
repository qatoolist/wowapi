---
id: W04-E03-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E03-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03-S002 — Artifacts index

Per mandate §9.2.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E03-S002-001 | `bulk_items` lease-column and lifecycle migration | migration | implementation | Adds `lease_token`, `lease_generation`, `lease_expires_at`, `idempotency_key` to `bulk_items`; adds `max_attempts` to `bulk_operations`; extends status constraints for pause/resume/cancel; drops superseded stopgap columns | DATA-04 | W04-E03-S002-T001 | `migrations/00044_bulk_items_lease_and_lifecycle.sql` | accepted |
| ART-W04-E03-S002-002 | Atomic leased-claim SQL implementation | source-code package | implementation | `UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`, bounded batch, with `ExplainClaimPlan` helper for EXPLAIN assertion | DATA-04 | W04-E03-S002-T002 | `kernel/bulk/bulk.go` (`claimBatch`, `claimSQL`) | accepted |
| ART-W04-E03-S002-003 | Item idempotency keys, finalize fencing, retry policy, cancellation | source-code package | implementation | `Item` carries `Lease` and `IdempotencyKey`; `runItem` finalizes with token/generation/expiry fencing; retry policy via `max_attempts`; cancellation via `Cancel` + operation-status check | DATA-04 | W04-E03-S002-T003 | `kernel/bulk/bulk.go` (`runItem`, `recordFailure`, `Cancel`) | accepted |
| ART-W04-E03-S002-004 | Pause/resume/cancel lifecycle-control API | source-code package | implementation | `Service.Pause`, `Service.Resume`, `Service.Cancel`; `Process` respects paused/cancelled state mid-run | DATA-04 | W04-E03-S002-T004 | `kernel/bulk/bulk.go` (`Pause`, `Resume`, `Cancel`, `Process`) | accepted |
| ART-W04-E03-S002-005 | Named multi-worker chaos test | test code | implementation | `kernel/bulk/chaos/duplicate_worker_test.go` — ≥2 processors concurrently claim/retry/pause/resume/cancel the same operation without duplicate effects or stale finalization | DATA-04 | W04-E03-S002-T005 | `kernel/bulk/chaos/duplicate_worker_test.go` | accepted |
| ART-W04-E03-S002-006 | Leased-claim, fencing, and lifecycle-control documentation | documentation | post-implementation | This index + code comments document the lease schema, claim SQL, idempotency scheme, fencing behavior, retry policy, cancellation path, and lifecycle-control API | DATA-04 | W04-E03-S002-T001 through -T005 | `impl/waves/.../story-002-leased-claims-and-lifecycle/artifacts/index.md` | accepted |

Note on ART-W04-E03-S002-005: the shared chaos harness from `W04-E01-S003` was not landed at the time this story implemented the named chaos test (see `deviations.md` and coordination log with `W04LeaseLeasePrimitive`). The test is therefore self-contained but is written to be a drop-in consumer of the harness once it lands; the test file path matches the source's required name verbatim and the scenario matches the Wave-3 exit-gate wording.
