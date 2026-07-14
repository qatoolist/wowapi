---
id: W07-E01-S002-T005
type: task
title: Parity and SQL-count-constant tests
status: done
parent_story: W07-E01-S002
owner: W07-Scoping-Dispatch.W07E01S002
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S002-T002
acceptance_criteria:
  - AC-W07-E01-S002-05
artifacts:
  - ART-W07-E01-S002-005
evidence:
  - EV-W07-E01-S002-005
---

# W07-E01-S002-T005 — Parity and SQL-count-constant tests

## Task Definition

### Task objective

Write result-parity + SQL-count-constant-with-depth tests, across 3/10/50-level ancestries.

### Parent story

W07-E01-S002

### Owner

unassigned

### Status

todo

### Dependencies

W07-E01-S002-T002 (tests operate on T002's own new query).

### Detailed work

1. Build query-counting instrumentation (not confirmed to exist yet, per PLAN T4's own risk note).
2. Write parametrized tests across 3/10/50-level ancestries.
3. Confirm result parity and constant SQL count with depth.

### Expected files or components affected

New parametrized test file(s).

### Expected output

SQL count stays constant across 3/10/50-level ancestries; result parity holds.

### Required artifacts

ART-W07-E01-S002-005 (parity + SQL-count test suite).

### Required evidence

EV-W07-E01-S002-005 (parametrized test output).

### Related acceptance criteria

AC-W07-E01-S002-05.

### Completion criteria

SQL count is constant across depths; parity holds.

### Verification method

Direct execution of the parametrized test suite.

### Risks

Low-medium — needs query-counting instrumentation, not confirmed to exist yet, per PLAN T4's own risk note.

### Rollback or recovery considerations

If query-counting instrumentation does not exist, build a minimal version scoped to this test's own need, not a general-purpose profiling tool.

## Implementation Record

### What was actually implemented

Added pgx query-count instrumentation scoped to the integration test and a reproducible legacy
reference implementation. The real-PostgreSQL test executes both algorithms at ancestry depths
3/10/50 and asserts exact result parity plus a constant set-based count.

### Tests added or modified

`kernel/rules/resolver_perf_test.go`:

- `TestIntegrationResolverQueryCountConstantWithDepth`
- `TestIntegrationResolverSetBasedParity`

### Measured result

Legacy total SQL statements: 11/18/58. Set-based total SQL statements: 8/8/8. The totals include
read-only transaction setup and one ancestry statement; the rules lookup itself is one statement.

### Implementation dates

2026-07-14.

### Relationship to the approved plan

Matched T4; the plan's previously unconfirmed query-count instrumentation was implemented locally in
the test.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Actual result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S002-05 | Parametrized legacy-versus-set-based integration test | PostgreSQL 16.14 container | PASS — parity holds; set-based count is 8/8/8 | EV-W07-E01-S002-005 | pending story independent review |

### Final conclusion

Passed on 2026-07-14 with no skipped depth.

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
