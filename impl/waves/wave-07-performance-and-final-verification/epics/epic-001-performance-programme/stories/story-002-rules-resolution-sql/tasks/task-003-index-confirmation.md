---
id: W07-E01-S002-T003
type: task
title: Index confirmation/addition
status: done
parent_story: W07-E01-S002
owner: W07-Scoping-Dispatch.W07E01S002
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S002-T002
acceptance_criteria:
  - AC-W07-E01-S002-03
artifacts:
  - ART-W07-E01-S002-003
evidence:
  - EV-W07-E01-S002-003
---

# W07-E01-S002-T003 — Index confirmation/addition

## Task Definition

### Task objective

Add or confirm indexes matching both current and historical predicates.

### Parent story

W07-E01-S002

### Owner

unassigned

### Status

todo

### Dependencies

W07-E01-S002-T001, W07-E01-S002-T002 (T0's audit and T1's query design inform which indexes are needed).

### Detailed work

1. Confirm existing indexes match the new query's predicates, or add new ones.
2. Confirm indexes also match historical query predicates (for any legacy code path still in use).
3. Run EXPLAIN (ANALYZE, BUFFERS) to confirm index access, not sequential scan.

### Expected files or components affected

New or confirmed index migration(s), following DATA-09's protocol if CREATE INDEX CONCURRENTLY is needed on a live table.

### Expected output

Indexes confirmed or added, proven via EXPLAIN to use index access.

### Required artifacts

ART-W07-E01-S002-003 (confirmed/added indexes).

### Required evidence

EV-W07-E01-S002-003 (EXPLAIN output confirming index access).

### Related acceptance criteria

AC-W07-E01-S002-03.

### Completion criteria

EXPLAIN shows index access, not sequential scan.

### Verification method

Direct execution of EXPLAIN (ANALYZE, BUFFERS) against the new query.

### Risks

Medium — wrong column order defeats the plan, per PLAN T2's own risk note.

### Rollback or recovery considerations

If an index fails to be used by the planner, diagnose the column-order/predicate mismatch and correct it.

## Implementation Record

### What was actually implemented

Confirmed the active-only GiST exclusion index from `00008_rules.sql` serves the current predicate.
Added `rule_versions_history_resolution_idx` for
`status IN ('active','superseded')` with equality columns for key/scope/tenant followed by descending
`effective_from`, and removed the obsolete narrower lookup index.

### Schema or migration changes

`migrations/00048_rule_versions_resolution_indexes.sql` uses the DATA-09 online pattern: manifest,
`NO TRANSACTION`, bounded session timeouts, and `CREATE INDEX CONCURRENTLY`.

### Tests added or modified

- `migrations/rules_resolution_indexes_test.go`
- `kernel/rules/resolver_perf_test.go` real EXPLAIN assertions

### Implementation dates

2026-07-14.

### Relationship to the approved plan

Matched T2. The audit-informed design confirmed the current constraint and added only the missing
historical index, avoiding a redundant active-only B-tree.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Actual result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S002-03 | Catalog test plus four real `EXPLAIN (ANALYZE, BUFFERS)` pairs | PostgreSQL 16.14 container | PASS — current and history use index access; zero `rule_versions` sequential scans | EV-W07-E01-S002-003 | pending story independent review |

### Final conclusion

Passed on 2026-07-14. `go test ./migrations -count=1 -v` also passed.

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
