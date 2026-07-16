---
id: W07-E01-S003-T003
type: task
title: Partial index on remind_after
status: done
parent_story: W07-E01-S003
owner: W07-Scoping-Dispatch.W07E01S003
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W07-E01-S003-03
artifacts:
  - ART-W07-E01-S003-003
evidence:
  - EV-W07-E01-S003-003
---

# W07-E01-S003-T003 — Partial index on remind_after

## Task Definition

### Task objective

Add a partial index on remind_after matching the reminder query's predicate.

### Parent story

W07-E01-S003

### Owner

W07-Scoping-Dispatch.W07E01S003

### Status

complete

### Dependencies

None.

### Detailed work

1. Design the partial index matching the query predicate.
2. Add it via CREATE INDEX CONCURRENTLY, following DATA-09's expand-only protocol (workflow_tasks is a
   live shared table).
3. Run EXPLAIN to confirm index-scan access.

### Expected files or components affected

A new migration for the remind_after partial index.

### Expected output

Index-scan access confirmed via EXPLAIN.

### Required artifacts

ART-W07-E01-S003-003 (remind_after partial index migration).

### Required evidence

EV-W07-E01-S003-003 (EXPLAIN plan report).

### Related acceptance criteria

AC-W07-E01-S003-03

### Completion criteria

EXPLAIN shows index scan, not sequential scan.

### Verification method

Direct execution of EXPLAIN against the reminder query.

### Risks

Low — additive, follows DATA-09's expand-only protocol since workflow_tasks is a live shared table, per PLAN T3's own risk note.

### Rollback or recovery considerations

If the index is not used by the planner, diagnose the predicate/column-order mismatch and correct it.

## Implementation Record

### What was actually implemented

Migration 00047 creates `wft_remind_after` concurrently on `(tenant_id, remind_after)` with the
partial predicate `status='open' AND remind_after IS NOT NULL`, matching the reminder query.
The migration follows the DATA-09 online manifest and reversible up/down contract.

### Components changed

Database migration and migration inventory.

### Files changed

`migrations/00047_perf04_sweeper_outbox_leases.sql`, `migrations/migrations_test.go`,
`kernel/workflow/sweeper_perf_test.go`.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

Adds `wft_remind_after`; the same migration also adds W04-compatible outbox lease fields required
by T5.

### Security changes

Tenant ID is the leading index key; RLS remains enabled and forced.

### Observability changes

None.

### Tests added or modified

Fresh migration contracts plus a real `EXPLAIN` assertion that rejects sequential scans.

### Commits

Working tree based on entry SHA `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

None.

### Follow-up items

None.

### Relationship to the approved plan

Matches plan T3 and DATA-09; no deviation.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-03 | Real PostgreSQL `EXPLAIN` plus migration package | PostgreSQL 16.9, migrated exclusive DB | Index scan using the partial index; migration contracts pass | EXPLAIN plan | W07-Scoping-Dispatch.W07E01S003ReviewR |

### Actual result

PASS: planner chose `Index Scan using wft_remind_after`, with tenant and due time in `Index Cond`;
no `enable_seqscan` override was used. Migration package passed.

### Pass or fail

PASS.

### Evidence identifier

EV-W07-E01-S003-003.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on entry SHA `733ef3e`.

### Environment

PostgreSQL 16.9 Docker service with real migration application.

### Reviewer

W07-Scoping-Dispatch.W07E01S003ReviewR.

### Findings

Independent review found no open issues.

### Retest status

Focused EXPLAIN and migration packages passed.

### Final conclusion

AC-03 accepted.
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
