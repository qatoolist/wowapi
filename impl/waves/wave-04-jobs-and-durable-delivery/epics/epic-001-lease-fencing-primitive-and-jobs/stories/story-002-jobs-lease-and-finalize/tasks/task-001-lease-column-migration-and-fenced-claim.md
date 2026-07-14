---
id: W04-E01-S002-T001
type: task
title: Lease-column migration and fenced claim SQL
status: done
parent_story: W04-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S001
acceptance_criteria:
  - AC-W04-E01-S002-01
artifacts:
  - ART-W04-E01-S002-001
  - ART-W04-E01-S002-004
evidence:
  - EV-W04-E01-S002-001
---

# W04-E01-S002-T001 — Lease-column migration and fenced claim SQL

## Task Definition

### Task objective

Add lease columns to `jobs_queue`, backed by W04-E01-S001's shared lease/fencing primitive, and
extend claim SQL to assign a fresh lease token and `generation+1` per claim, with `claimedJob`
carrying the resulting lease context forward.

### Parent story

W04-E01-S002 — Jobs lease columns, fenced finalize, and fenced reclaim.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S001 (the shared primitive must exist and be locked before `jobs_queue` can carry lease
columns assigned against it).

### Detailed work

1. Re-read `kernel/jobs`'s claim SQL and `claimedJob` struct at this task's actual start commit to
   confirm no lease columns or lease-context fields currently exist.
2. Design the `jobs_queue` lease-column migration, mirroring W04-E01-S001's primitive schema;
   confirm and document the existing timeout-floor logic being reused, per PLAN DATA-02 T2's own
   risk note ("Reuse existing timeout-floor logic, don't introduce a second inconsistent timeout
   source").
3. Implement the migration.
4. Extend claim SQL to assign a fresh lease token and `generation+1` atomically with the claim;
   extend `claimedJob` to carry the lease context.
5. Write a migration + unit test proving claim assignment behavior.
6. Document the lease-column schema (this task's share of ART-W04-E01-S002-004).

### Expected files or components affected

A new migration file adding lease columns to `jobs_queue`; `kernel/jobs`'s claim SQL and
`claimedJob` struct (exact file paths TBD per `plan.md`).

### Expected output

`jobs_queue` carrying lease columns; claim SQL assigning a fresh token + `generation+1`; a passing
migration + unit test.

### Required artifacts

ART-W04-E01-S002-001 (lease-column migration), ART-W04-E01-S002-004 (documentation, shared with
T002/T003).

### Required evidence

EV-W04-E01-S002-001 (migration + unit-test report).

### Related acceptance criteria

AC-W04-E01-S002-01.

### Completion criteria

`jobs_queue` has lease columns; claim SQL assigns a fresh token + `generation+1`; `claimedJob`
carries lease context — proven by a passing migration + unit test.

### Verification method

Direct execution of the migration against a test database; direct execution of the unit test
asserting claim-time lease assignment.

### Risks

Reusing (not duplicating) the existing timeout-floor logic — PLAN T2's own risk note. Introducing a
second, inconsistent timeout source would be a design defect, not merely a quality gap.

### Rollback or recovery considerations

Revert the migration if it breaks any in-flight job claimed under the pre-fencing schema at deploy
time; resolve the compatibility handling (see `story.md`/`plan.md`) before re-applying.

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

*Not yet implemented — this task adds lease columns to `jobs_queue`; recorded here once executed.*

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
| AC-W04-E01-S002-01 | Run migration + unit test for claim lease assignment | Local dev or CI, PostgreSQL instance | Claim assigns fresh token + `generation+1`; `claimedJob` carries lease context | migration + unit-test report | unassigned |

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
