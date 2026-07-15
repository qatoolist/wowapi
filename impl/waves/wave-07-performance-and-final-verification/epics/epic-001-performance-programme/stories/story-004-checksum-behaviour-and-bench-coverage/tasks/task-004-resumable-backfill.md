---
id: W07-E01-S004-T004
type: task
title: Resumable async backfill
status: complete
parent_story: W07-E01-S004
owner: W07-Scoping-Dispatch.W07E01S004
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S004-T002
acceptance_criteria:
  - AC-W07-E01-S004-04
artifacts:
  - ART-W07-E01-S004-004
evidence:
  - EV-W07-E01-S004-004
---

# W07-E01-S004-T004 — Resumable async backfill

## Task Definition

### Task objective

Build a resumable async backfill for legacy objects, surviving an interrupt-and-resume cycle with no duplicate work.

### Parent story

W07-E01-S004

### Owner

W07-Scoping-Dispatch.W07E01S004

### Status

complete

### Dependencies

W07-E01-S004-T002 (the backfill uses T002's own repair path as its per-object mechanism).

### Detailed work

1. Build or confirm the inventory mechanism for "legacy objects lacking checksum metadata" (not
   confirmed to exist yet, per PLAN's own risk note).
2. Build the resumable async backfill, consuming the inventory and T002's own repair path per object.
3. Write an interrupt/resume test confirming no duplicate work and eventual completion.

### Expected files or components affected

A new resumable-backfill mechanism (exact location TBD); possibly a new inventory table/mechanism.

### Expected output

A resumable backfill surviving interrupt/resume with no duplicate work.

### Required artifacts

ART-W07-E01-S004-004 (resumable backfill mechanism).

### Required evidence

EV-W07-E01-S004-004 (interrupt/resume backfill test output).

### Related acceptance criteria

AC-W07-E01-S004-04.

### Completion criteria

The interrupt/resume cycle succeeds with no duplicate work and reaches completion.

### Verification method

Direct execution of the interrupt/resume test.

### Risks

Medium — needs an inventory mechanism for 'legacy objects lacking checksum metadata,' which doesn't obviously exist yet, per PLAN T4's own risk note.

### Rollback or recovery considerations

If the inventory mechanism proves to require significant new infrastructure, scope it as its own sub-task and record the expanded scope, rather than silently absorbing an unbounded amount of new work.

## Implementation Record

Added `BackfillChecksums`, a context-cancellable, bounded batch operation for
background/async invocation. `ListObjects` uses a stable `StartAfter` cursor;
each object is classified by HEAD metadata, canonical objects are skipped, and
legacy objects reuse the bounded labeled repair path. The returned cursor is the
last successfully classified/repaired key, so interruption resumes without
reprocessing or duplicates.

Files: `adapters/storage/s3/backfill.go`,
`adapters/storage/s3/checksum_repair_test.go`, and the published inventory.
Implemented 2026-07-14, working tree based on `733ef3e`; no schema/configuration
change, PR, debt, or plan deviation.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S004-04 | interrupt/resume integration test | Local MinIO | no duplicates and all three legacy objects complete | integration test report | independent story reviewer |

**PASS**, 2026-07-14, working tree based on `733ef3e`.
EV-W07-E01-S004-004 proves a three-object real MinIO inventory completes over
interrupt/resume with exactly three repairs. Independent review: correct,
confidence 1, no findings.
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
