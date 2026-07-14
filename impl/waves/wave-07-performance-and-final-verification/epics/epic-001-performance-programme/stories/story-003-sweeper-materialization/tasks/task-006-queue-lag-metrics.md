---
id: W07-E01-S003-T006
type: task
title: Queue-lag and batch-duration metrics
status: complete
parent_story: W07-E01-S003
owner: W07-Scoping-Dispatch.W07E01S003
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S003-T001
  - W07-E01-S003-T005
acceptance_criteria:
  - AC-W07-E01-S003-06
artifacts:
  - ART-W07-E01-S003-006
evidence:
  - EV-W07-E01-S003-006
---

# W07-E01-S003-T006 — Queue-lag and batch-duration metrics

## Task Definition

### Task objective

Add queue-lag and batch-duration metrics for sweeper/webhook/outbox timing.

### Parent story

W07-E01-S003

### Owner

W07-Scoping-Dispatch.W07E01S003

### Status

complete

### Dependencies

W07-E01-S003-T001, W07-E01-S003-T005 (metrics instrument the bounded-batch and leased-state-machine mechanisms).

### Detailed work

1. Add queue-lag metrics for sweeper/webhook/outbox timing.
2. Add batch-duration metrics for the same three mechanisms.
3. Write metric-emission tests.

### Expected files or components affected

New metric-emission code across kernel/workflow, kernel/webhook, kernel/outbox.

### Expected output

Queue-lag and batch-duration metrics emitted for all three mechanisms.

### Required artifacts

ART-W07-E01-S003-006 (queue-lag/batch-duration metric emission).

### Required evidence

EV-W07-E01-S003-006 (metric-emission test output).

### Related acceptance criteria

AC-W07-E01-S003-06

### Completion criteria

Metrics for sweeper/webhook/outbox timing are emitted and tested.

### Verification method

Direct execution of metric-emission tests.

### Risks

Low, per PLAN T6's own risk classification.

### Rollback or recovery considerations

If a metric proves incorrectly labeled or scoped, correct it directly.

## Implementation Record

### What was actually implemented

SweepSLA, RetryOutbound, and outbox relay now emit `worker_queue_lag_seconds` and
`worker_batch_duration_seconds`. Queue lag is the oldest claimed due/event age; batch duration
uses the histogram-capable metrics port. Labels are fixed to the `worker` dimension with three values.

### Components changed

Workflow runtime, webhook service, outbox relay, observability port, and production wiring.

### Files changed

`kernel/observability/metrics.go`, `kernel/workflow/runtime.go`, `kernel/workflow/sweeper.go`,
`foundation/webhook/service.go`, `kernel/outbox/relay.go`, `kernel/kernel.go`, `app/worker.go`.

### Interfaces introduced or changed

`observability.Metrics` gains `ObserveHistogram`; workflow and relay add metrics options.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

Metrics carry no tenant, endpoint, resource, or event identifiers.

### Observability changes

Adds bounded timing gauge/histogram emission for the three PERF-04 workers.

### Tests added or modified

Recording sinks assert both metric names and the exact bounded worker labels.

### Commits

Working tree based on entry SHA `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Queue lag is sampled from the oldest row claimed by the current bounded batch.

### Follow-up items

None.

### Relationship to the approved plan

Matches plan T6.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-06 | Recording metric sinks and focused production-wiring tests | Real PostgreSQL for worker paths | Lag gauge and duration histogram with bounded labels | metric-emission report | W07-Scoping-Dispatch.W07E01S003ReviewR |

### Actual result

PASS: all three worker values emitted both timing metrics; tests reject missing names or unexpected
label values; kernel/worker wiring passes.

### Pass or fail

PASS.

### Evidence identifier

EV-W07-E01-S003-006.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on entry SHA `733ef3e`.

### Environment

darwin/arm64, Go 1.26.5, PostgreSQL 16.9 Docker service.

### Reviewer

W07-Scoping-Dispatch.W07E01S003ReviewR.

### Findings

Independent review found no open issues.

### Retest status

Focused workflow, webhook, outbox, kernel, and app tests passed.

### Final conclusion

AC-06 accepted.
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
