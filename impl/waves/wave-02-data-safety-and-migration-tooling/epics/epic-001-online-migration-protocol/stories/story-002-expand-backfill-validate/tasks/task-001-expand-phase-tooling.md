---
id: W02-E01-S002-T001
type: task
title: Expand-phase tooling
status: done
parent_story: W02-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E01-S001
acceptance_criteria:
  - AC-W02-E01-S002-01
artifacts:
  - ART-W02-E01-S002-001
  - ART-W02-E01-S002-004
evidence:
  - EV-W02-E01-S002-001
---

# W02-E01-S002-T001 — Expand-phase tooling

## Task Definition

### Task objective

Implement expand-phase tooling — nullable/default-safe columns, new tables/indexes/compatibility
views, `NOT VALID` constraints, non-transactional `CREATE INDEX CONCURRENTLY` — such that expand
migrations don't block traffic and both old and new application readers accept the expanded schema,
per PLAN DATA-09 T3's acceptance criterion.

### Parent story

W02-E01-S002 — Expand-phase tooling, resumable backfill harness, and validation-phase tooling.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S001 (the manifest schema this task's expand migrations are classified against — PLAN T3's
own "Depends-on" column names T1).

### Detailed work

1. Confirm PLAN T3's own risk note first: test whether the current migration tooling supports
   issuing statements outside the wrapping transaction — a hard prerequisite for
   `CREATE INDEX CONCURRENTLY`, which cannot run inside a transaction block. If it does not, extend
   the migration-execution mechanism to support non-transactional statements before building the
   expand helpers on top.
2. Implement expand-phase helpers: nullable/default-safe column addition, new table/index/
   compatibility-view creation, `NOT VALID` constraint addition, non-transactional
   `CREATE INDEX CONCURRENTLY` issuance.
3. Write the old-reader-compatibility test: apply an expand migration and confirm both an
   old-version reader and a new-version reader accept the expanded schema during the migration
   window, and that the expand statements do not block concurrent traffic.
4. Document the expand-phase tooling's supported schema-change classes.

### Expected files or components affected

New expand-phase tooling package (exact location TBD per `plan.md`); possibly the migration-
execution mechanism established/extended by W02-E01-S001, if non-transactional statement support
must be added (step 1).

### Expected output

Expand-phase tooling whose migrations don't block traffic, proven by the old-reader-compatibility
test.

### Required artifacts

ART-W02-E01-S002-001 (expand-phase tooling), ART-W02-E01-S002-004 (documentation, shared with
T002/T003).

### Required evidence

EV-W02-E01-S002-001 (old-reader-compatibility test output).

### Related acceptance criteria

AC-W02-E01-S002-01.

### Completion criteria

`CREATE INDEX CONCURRENTLY` and `NOT VALID` constraints are issued without blocking traffic; the
old-reader-compatibility test passes with both reader versions accepting the expanded schema —
evidenced by a logged run against a named commit SHA.

### Verification method

Direct execution of the old-reader-compatibility test against a live PostgreSQL instance, logged
output retained as evidence per `evidence/index.md`.

### Risks

PLAN T3's own risk note: the current tooling may not support issuing statements outside the wrapping
transaction — if confirming this reveals a larger-than-expected gap in the migration-execution
mechanism, the extra work is recorded (and, if it threatens task boundedness per mandate §12,
escalated) rather than silently absorbed.

### Rollback or recovery considerations

Expand-phase changes are additive by design (that is the expand/contract discipline's core
property); reverting the tooling itself is a plain code revert with no data impact, since this task
executes no real production migration.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable — this task builds tooling; it executes no real production migration.*

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
| AC-W02-E01-S002-01 | Old-reader-compatibility test after an expand migration | Local dev or CI, PostgreSQL, two reader versions (or equivalent simulation) | Both readers accept; no traffic blocked by expand DDL | compatibility-test report | unassigned |

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
