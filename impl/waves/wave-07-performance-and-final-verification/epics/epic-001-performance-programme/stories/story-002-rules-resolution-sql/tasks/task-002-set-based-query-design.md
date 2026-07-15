---
id: W07-E01-S002-T002
type: task
title: Set-based query design
status: done
parent_story: W07-E01-S002
owner: W07-Scoping-Dispatch.W07E01S002
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S002-T001
acceptance_criteria:
  - AC-W07-E01-S002-02
artifacts:
  - ART-W07-E01-S002-002
evidence:
  - EV-W07-E01-S002-002
---

# W07-E01-S002-T002 — Set-based query design

## Task Definition

### Task objective

Design one set-based query over ancestry + tenant + platform fallback, preserving exact precedence semantics.

### Parent story

W07-E01-S002

### Owner

unassigned

### Status

todo

### Dependencies

W07-E01-S002-T001 (T0's own audit outcome informs this task's own design).

### Detailed work

1. Design the set-based query, informed by T001's own audit outcome.
2. Preserve exact nearest-ancestor-first → tenant → platform → code-default precedence.
3. Write result-parity unit tests against the old per-ancestor-loop implementation.

### Expected files or components affected

The rules-resolution query implementation file (exact path TBD).

### Expected output

A single SQL statement replacing the for loop, preserving exact precedence semantics.

### Required artifacts

ART-W07-E01-S002-002 (set-based query).

### Required evidence

EV-W07-E01-S002-002 (result-parity unit test output).

### Related acceptance criteria

AC-W07-E01-S002-02.

### Completion criteria

Result parity holds against the old implementation for every existing precedence scenario.

### Verification method

Direct execution of result-parity unit tests.

### Risks

Medium — must preserve nearest-ancestor-first → tenant → platform → code-default precedence and the schema-drift re-validation unchanged, per PLAN T1's own risk note.

### Rollback or recovery considerations

If precedence semantics diverge, treat as a correctness defect and revert to the old loop while re-diagnosing.

## Implementation Record

### What was actually implemented

Replaced the per-ancestor Go lookup loop with one SQL statement. `unnest(... WITH ORDINALITY)`
preserves self-first ancestry order; indexed LATERAL candidates choose the applicable version in each
org scope, followed by tenant and platform candidates. The existing current-schema validation remains
after the winner is scanned.

### Files changed

- `kernel/rules/resolver.go`
- `kernel/rules/resolver_perf_test.go`

### Tests added or modified

`TestIntegrationResolverSetBasedParity` compares the live statement to the legacy algorithm for
nearest org, ancestor org, tenant, platform, code default, and historical superseded windows.

### Implementation dates

2026-07-14.

### Relationship to the approved plan

Matched the plan; the exact SQL shape was resolved after T001 as an indexed LATERAL set.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Actual result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S002-02 | Real-PostgreSQL parity test | PostgreSQL 16.14 container, `WOWAPI_REQUIRE_DB=1` | PASS — all six precedence/history cases | EV-W07-E01-S002-002 | pending story independent review |

### Final conclusion

Passed on 2026-07-14 with no skipped cases.

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
