---
id: W07-E01-S004-T002
type: task
title: Bounded repair path
status: done
parent_story: W07-E01-S004
owner: W07-Scoping-Dispatch.W07E01S004
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W07-E01-S004-T001
acceptance_criteria:
  - AC-W07-E01-S004-02
artifacts:
  - ART-W07-E01-S004-002
evidence:
  - EV-W07-E01-S004-002
---

# W07-E01-S004-T002 — Bounded repair path

## Task Definition

### Task objective

Move the full-hash fallback to an explicit, size/time-bounded import/repair path.

### Parent story

W07-E01-S004

### Owner

W07-Scoping-Dispatch.W07E01S004

### Status

complete

### Dependencies

W07-E01-S004-T001 (the repair path is the fallback for objects T001's enforcement did not yet cover).

### Detailed work

1. Decide the exact storage.ObjectInfo port API-surface shape (new Stat variant vs. RepairChecksum
   method).
2. Move the full-hash fallback behind that labeled repair invocation.
3. Confirm other storage.ObjectInfo-implementing adapters still compile and behave correctly.
4. Write a test where a legacy object triggers the fallback only via the labeled path.

### Expected files or components affected

adapters/storage/s3's implementation files; possibly kernel/storage's ObjectInfo port.

### Expected output

The fallback reachable only via the labeled repair invocation, never ambient Stat.

### Required artifacts

ART-W07-E01-S004-002 (bounded repair path).

### Required evidence

EV-W07-E01-S004-002 (labeled-repair-path test output).

### Related acceptance criteria

AC-W07-E01-S004-02.

### Completion criteria

A legacy object triggers the fallback only via the labeled path.

### Verification method

Direct execution of the labeled-repair-path test.

### Risks

Medium — likely needs an API-surface decision affecting the storage.ObjectInfo port other adapters implement, per PLAN T2's own risk note.

### Rollback or recovery considerations

If the API-surface change breaks another adapter, revise the shape to preserve compatibility.

## Implementation Record

Added the optional `storage.ChecksumRepairer` capability and S3
`RepairChecksum`. Repair requires a non-empty operation label, positive byte
bound, and positive timeout. It rejects canonical, missing, malformed, or
oversized objects before GET; only the explicitly labeled path downloads,
hashes, and persists canonical metadata by server-side copy.

Files: `kernel/storage/storage.go`, `adapters/storage/s3/s3.go`, and
`adapters/storage/s3/checksum_repair_test.go`. The optional capability preserves
existing adapter source compatibility. Implemented 2026-07-14, working tree
based on `733ef3e`; no PR, debt, schema change, or plan deviation.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S004-02 | focused S3 integration tests | Local MinIO | only labeled size/time-bounded repair hashes legacy content | integration test report | independent story reviewer |

**PASS**, 2026-07-14, working tree based on `733ef3e`.
EV-W07-E01-S004-002 proves unlabeled and oversized work is rejected and bounded
repair succeeds. Independent review: correct, confidence 1, no findings.
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
