---
id: W04-E03-S002-T004
type: task
title: Pause/resume/cancel lifecycle controls
status: done
parent_story: W04-E03-S002
owner: W04BulkSafety
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E03-S002-T002
  - W04-E03-S002-T003
acceptance_criteria:
  - AC-W04-E03-S002-04
artifacts:
  - ART-W04-E03-S002-004
evidence:
  - EV-W04-E03-S002-004
---

# W04-E03-S002-T004 — Pause/resume/cancel lifecycle controls

## Task Definition

### Task objective

Add operation-level pause, resume, and cancel controls to `Service`.

### Status

done.

### Completion criteria

- `Pause` marks operation `paused`; `Process` stops claiming new batches.
- `Resume` marks operation `running`; `Process` continues.
- `Cancel` marks operation `cancelled`; `Process` stops and cancels pending items.
- No duplicate effects on pause/resume/cancel cycles.

## Implementation Record

- Added `Service.Pause(ctx, txm, tenantID, bulkID) error`.
- Added `Service.Resume(ctx, txm, tenantID, bulkID) error`.
- Added `Service.Cancel(ctx, txm, tenantID, bulkID) error`.
- `Process` reads operation status at start; if `paused`/`cancelled`/`completed`, returns.
- `Process` checks status inside the batch loop; if operation transitions to `paused`, returns
  processed count; if `cancelled`, cancels pending items and returns.

## Verification Record

- `TestIntegrationBulkPauseResumeCancel`: pauses mid-operation, resumes, cancels, verifies counts.

### Pass or fail

Pass.

### Evidence identifier

EV-W04-E03-S002-004.

## Deviations Record

None.
