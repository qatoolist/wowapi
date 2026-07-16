---
id: W02-E02-S001-T001
type: task
title: Parent tenant-scoped unique indexes
status: done
parent_story: W02-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W02-E02-S001-01
artifacts:
  - ART-W02-E02-S001-001
evidence:
  - EV-W02-E02-S001-001
---

# W02-E02-S001-T001 — Parent tenant-scoped unique indexes

## Task Definition

### Task objective

Add or confirm `UNIQUE (tenant_id, id)` on every parent table referenced by the 8 confirmed
tenant-scoped child-table foreign keys (`parties`, `organizations`, `documents`,
`document_versions`), built `CONCURRENTLY` so no index build blocks concurrent traffic.

### Parent story

W02-E02-S001 — Parent tenant-scoped unique indexes, FK catalog scanner, and CI gate.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Query `pg_indexes` against `parties`, `organizations`, `documents`, and `document_versions` at
   this task's actual start commit to confirm which (if any) already carry a
   `UNIQUE (tenant_id, id)` index — PLAN DATA-01 T1's own acceptance criterion is "add/confirm," not
   merely "add."
2. For each parent lacking the index, write a `CONCURRENTLY`-built migration adding
   `UNIQUE (tenant_id, id)`, run non-transactionally per T1's own risk note ("`SHARE UPDATE
   EXCLUSIVE` lock — must run non-transactionally").
3. Write a migration test that queries `pg_indexes` post-migration and confirms all 4 parents carry
   the index.
4. Document which parents already had the index versus which were newly added, and why (this
   confirmation, not an invented assumption, is the source of truth for this task's own closure).

### Expected files or components affected

Up to 4 new migration files (one per parent lacking the index; fewer if some already have it), under
the existing migration directory structure.

### Expected output

`UNIQUE (tenant_id, id)` confirmed present, via `pg_indexes`, on all 4 parent tables.

### Required artifacts

ART-W02-E02-S001-001 (parent tenant-scoped unique-index migrations).

### Required evidence

EV-W02-E02-S001-001 (migration test report, `pg_indexes` query output).

### Related acceptance criteria

AC-W02-E02-S001-01.

### Completion criteria

`pg_indexes` confirms `UNIQUE (tenant_id, id)` exists on `parties`, `organizations`, `documents`, and
`document_versions`, proven by a passing migration test.

### Verification method

Direct execution of the migration test against a live PostgreSQL instance, querying `pg_indexes`
post-migration.

### Risks

`CONCURRENTLY` index builds take a `SHARE UPDATE EXCLUSIVE` lock that can conflict with certain other
concurrent DDL — must run non-transactionally per T1's own risk note; running it inside a transaction
would be a defect, not a valid implementation choice.

### Rollback or recovery considerations

If a `CONCURRENTLY` build fails partway, PostgreSQL leaves an invalid index behind — this must be
dropped and the migration retried, not silently left in place.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not yet implemented — expected: up to 4 `CONCURRENTLY` unique-index migrations.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E02-S001-01 | Query `pg_indexes` against all 4 parent tables post-migration | Local dev or CI, PostgreSQL instance | `UNIQUE (tenant_id, id)` present on all 4 parents | migration test report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

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
