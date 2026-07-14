---
id: W04-E03-S002-T005
type: task
title: Named multi-worker chaos test
status: done
parent_story: W04-E03-S002
owner: W04BulkSafety
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E03-S002-T002
  - W04-E03-S002-T003
  - W04-E03-S002-T004
acceptance_criteria:
  - AC-W04-E03-S002-05
artifacts:
  - ART-W04-E03-S002-005
evidence:
  - EV-W04-E03-S002-005
---

# W04-E03-S002-T005 — Named multi-worker chaos test

## Task Definition

### Task objective

Create the named chaos test `DATA-04/chaos/duplicate_worker_test.go` that proves multi-worker
safety under adversarial concurrency.

### Status

done.

### Completion criteria

- Test path matches required name.
- ≥2 processors concurrently claim/retry/pause/resume/cancel the same operation.
- No duplicate effects.
- No stale finalization.

## Implementation Record

- Created `kernel/bulk/chaos/duplicate_worker_test.go`.
- Spawns 4 worker goroutines that loop calling `Process` with retry on error.
- A separate lifecycle toggler randomly pauses, resumes, and cancels.
- An effect ledger (`bulk_marks`) records each successful item exactly once, protected by a unique
  constraint on `(bulk_id, seq)`.
- Final assertion: `ledger count == done count` and no duplicate seq entries.

## Verification Record

- `TestIntegrationBulkDuplicateWorkerChaos` passes.

### Pass or fail

Pass.

### Evidence identifier

EV-W04-E03-S002-005.

## Deviations Record

The shared chaos harness from `W04-E01-S003` was not landed when this task executed. A self-contained
chaos test was authored instead, with the same required scenario and path name, and structured to be
migrated onto the shared harness once it lands. Recorded in `deviations.md`.
