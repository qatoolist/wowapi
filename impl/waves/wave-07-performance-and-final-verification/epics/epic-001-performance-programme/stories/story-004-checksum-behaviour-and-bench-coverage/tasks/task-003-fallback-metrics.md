---
id: W07-E01-S004-T003
type: task
title: Fallback-invocation metrics
status: done
parent_story: W07-E01-S004
owner: W07-Scoping-Dispatch.W07E01S004
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W07-E01-S004-T002
acceptance_criteria:
  - AC-W07-E01-S004-03
artifacts:
  - ART-W07-E01-S004-003
evidence:
  - EV-W07-E01-S004-003
---

# W07-E01-S004-T003 — Fallback-invocation metrics

## Task Definition

### Task objective

Dedicated metrics for fallback invocations: counter/histogram for hits, bytes, duration.

### Parent story

W07-E01-S004

### Owner

W07-Scoping-Dispatch.W07E01S004

### Status

complete

### Dependencies

W07-E01-S004-T002 (metrics instrument T002's own repair path).

### Detailed work

1. Add a counter for fallback hits.
2. Add a histogram for fallback bytes and duration.
3. Write a metric-emission test.

### Expected files or components affected

adapters/storage/s3's implementation files, extended.

### Expected output

Dedicated metrics for fallback hits, bytes, duration.

### Required artifacts

ART-W07-E01-S004-003 (fallback-invocation metrics).

### Required evidence

EV-W07-E01-S004-003 (metric-emission test output).

### Related acceptance criteria

AC-W07-E01-S004-03.

### Completion criteria

Metrics are emitted and tested.

### Verification method

Direct execution of the metric-emission test.

### Risks

Low, per PLAN T3's own risk classification.

### Rollback or recovery considerations

If a metric proves incorrectly labeled, correct it directly.

## Implementation Record

Extended `observability.Metrics` with `ObserveHistogram`, including NoOp and
Prometheus implementations, and wired repair hit, byte, and duration metrics
with the bounded repair label. Existing metrics fakes were updated for interface
compatibility. Implemented 2026-07-14 in the working tree based on `733ef3e`;
no schema/configuration change, PR, debt, or plan deviation.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S004-03 | focused S3, Prometheus, and observability tests | Local MinIO | hit counter and byte/duration histograms observed | integration/unit report | independent story reviewer |

**PASS**, 2026-07-14, working tree based on `733ef3e`.
EV-W07-E01-S004-003 records one hit, repaired bytes, and positive duration.
Independent review: correct, confidence 1, no findings.
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
