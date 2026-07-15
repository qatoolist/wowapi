---
id: W04-E03-S002-TASKS-INDEX
type: tasks-index
parent_story: W04-E03-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03-S002 — Tasks index

Per mandate §16.4.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E03-S002-T001](task-001-lease-columns-via-shared-primitive.md) | Lease columns via the shared primitive | W04BulkSafety | done | W04-E01-S001, W04-E03-S001 | `bulk_items` lease columns + migration test | AC-W04-E03-S002-01 | done | done |
| [W04-E03-S002-T002](task-002-atomic-leased-claim.md) | Atomic leased claim, bounded batch | W04BulkSafety | done | T001 | Atomic `SKIP LOCKED` claim SQL + EXPLAIN assertion + N>1 claimer test | AC-W04-E03-S002-02 | done | done |
| [W04-E03-S002-T003](task-003-idempotency-fencing-retry-cancellation.md) | Item idempotency keys, finalize fencing, retry policy, cancellation | W04BulkSafety | done | T002 | Idempotency keys + reused finalize fencing + retry + cancellation | AC-W04-E03-S002-03 | done | done |
| [W04-E03-S002-T004](task-004-pause-resume-cancel-lifecycle.md) | Pause/resume/cancel lifecycle controls | W04BulkSafety | done | T002, T003 | Operation-level pause/resume/cancel controls + lifecycle integration tests | AC-W04-E03-S002-04 | done | done |
| [W04-E03-S002-T005](task-005-named-multi-worker-chaos-test.md) | Named multi-worker chaos test | W04BulkSafety | done | T002, T003, T004 | `DATA-04/chaos/duplicate_worker_test.go` | AC-W04-E03-S002-05 | done | done |
| [W04-E03-S002-T006](task-006-independent-review.md) | Independent review | W04BulkSafety | done | T001–T005 | Independent-review record per mandate §14 | AC-W04-E03-S002-01 through -05 | done | done |

## Grouping rationale

(See original rationale in file history; unchanged.)
