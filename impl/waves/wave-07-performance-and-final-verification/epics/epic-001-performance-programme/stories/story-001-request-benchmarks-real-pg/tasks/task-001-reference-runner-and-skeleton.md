---
id: W07-E01-S001-T001
type: task
title: Reference runner + perf/reference-v1.json skeleton
status: complete
parent_story: W07-E01-S001
owner: W07-Phase-A-Execution.W07E01S001
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E01-S001-01
artifacts:
  - ART-W07-E01-S001-001
evidence:
  - EV-W07-E01-S001-001
---

# W07-E01-S001-T001 — Reference runner + perf/reference-v1.json skeleton

## Task Definition

### Task objective

Stand up a dedicated Linux amd64 reference runner (provisional per DEC-Q9's default) and a perf/reference-v1.json skeleton recording the full §14 field list.

### Parent story

W07-E01-S001

### Owner

W07-Phase-A-Execution.W07E01S001

### Status

complete

### Dependencies

None. This task's own output is the shared prerequisite for T002-T005 in this story and for W07-E01-S002/S003/S004's own publication tasks.

### Detailed work

1. Stand up the provisional Linux amd64 GitHub Actions reference runner.
2. Build the perf/reference-v1.json skeleton, recording CPU/runner digest, Go version, Postgres
   config, pool size, dataset cardinality, tenant distribution, workload seed, warm-up/measurement
   durations.

### Expected files or components affected

perf/reference-v1.json (new); new CI workflow configuration for the reference runner.

### Expected output

A working reference runner and a complete-field-list perf/reference-v1.json skeleton.

### Required artifacts

ART-W07-E01-S001-001 (perf/reference-v1.json + fixtures).

### Required evidence

EV-W07-E01-S001-001 (field-completeness confirmation report).

### Related acceptance criteria

AC-W07-E01-S001-01.

### Completion criteria

All named §14 fields are present and correctly recorded.

### Verification method

Direct inspection of the generated artifact against the field list.

### Risks

High — new CI infrastructure, no owner/timeline established anywhere in the directive, per PLAN T1's own risk note.

### Rollback or recovery considerations

If the reference runner proves unreliable, escalate to the performance/SRE lead; do not silently substitute an unrepresentative environment without recording why.

## Implementation Record

### What was actually implemented

Added the standalone Linux amd64 reference workflow, full-field reference JSON, deterministic fixture, and contract tests.

### Components and files changed

`.github/workflows/perf-reference.yml`, `perf/reference-v1.json`, `perf/fixtures/request-workloads-v1.json`, `perf/requestbench/reference_test.go`

### Interfaces, configuration, schema, and security

Benchmark-only additive interfaces/configuration. No schema migration or production API changed; runtime RLS remains enforced.

### Tests, revision, and date

Focused contracts and real-PostgreSQL execution passed on the working tree based on `1626b11`; implemented 2026-07-13 through 2026-07-14. No commit or pull request was created.

### Relationship to the approved plan

Implemented as planned. Absolute SLOs remain outside scope pending DEC-Q9.

## Verification Record

### Actual result and pass/fail

PASS. See `EV-W07-E01-S001-001` and the story-level `verification.md` for the exact command, environment, result, and checksum-pinned output.

### Execution date and revision

2026-07-14; working tree based on entry SHA `1626b11`.

### Environment

Exact required local PostgreSQL environment for focused contracts; pinned Linux/amd64 Go 1.26.5 + PostgreSQL 16.9 containers for publication.

### Reviewer and findings

Independent review by `W05ReviewGateFinal` passed with zero open actionable issues.

### Retest and conclusion

Retested after fixes; task completion criterion satisfied.

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
