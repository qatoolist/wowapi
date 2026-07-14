---
id: W03-E01-S001-T003
type: task
title: Zero/unknown-tenant rejection (SEC-01 T3)
status: complete
parent_story: W03-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E01-S001-T002
acceptance_criteria:
  - AC-W03-E01-S001-03
artifacts:
  - ART-W03-E01-S001-006
evidence:
  - EV-W03-E01-S001-005
---

# W03-E01-S001-T003 — Zero/unknown-tenant rejection (SEC-01 T3)

## Task Definition

### Task objective

Reject a zero or unknown tenant claim before a tenant transaction is opened (before any
`WithTenantID` call), rather than allowing it to reach a downstream check.

### Parent story

W03-E01-S001 — Grant schema and unconditional membership enforcement.

### Owner

unassigned

### Status

complete

### Dependencies

W03-E01-S001-T002 — PLAN's own Depends-on column for T3 names T2; the rejection logic sits
alongside/ahead of the membership-check call path T002 modifies.

### Detailed work

1. Identify the exact call path from claim extraction to `WithTenantID` invocation at this task's
   actual start commit.
2. Add a check rejecting a zero-UUID or otherwise-malformed tenant claim before that call path
   reaches `WithTenantID`.
3. Write a negative test proving a zero/garbage-UUID tenant claim is rejected pre-transaction.

### Expected files or components affected

`kernel/auth/auth.go` (or the specific call-path file identified in step 1 — to be confirmed at
implementation time).

### Expected output

A zero or garbage-UUID tenant claim is rejected before any tenant transaction opens.

### Required artifacts

ART-W03-E01-S001-006 (zero/unknown-tenant rejection logic).

### Required evidence

EV-W03-E01-S001-005 (negative test report).

### Related acceptance criteria

AC-W03-E01-S001-03.

### Completion criteria

The negative test proves rejection occurs before `WithTenantID` is called, not merely that the
request eventually fails somewhere downstream.

### Verification method

Direct negative-test execution, logged output retained as evidence.

### Risks

Low — PLAN's own risk column for T3 records "Low."

### Rollback or recovery considerations

Revert the added check if it is found to reject a legitimate tenant-claim shape not anticipated by
this task (e.g. a valid tenant ID that happens to look malformed under an overly strict check) —
low risk given the narrow, additive nature of the change.

## Implementation Record

### What was actually implemented

Completed per the parent story's `implementation.md`.

### Components changed

See `implementation.md` §Components changed.

### Files changed

See `implementation.md` §Files changed.

### Interfaces introduced or changed

See `implementation.md` §Interfaces introduced or changed.

### Configuration changes

None.

### Schema or migration changes

See `implementation.md` §Schema or migration changes.

### Security changes

See `implementation.md` §Security changes.

### Observability changes

None beyond existing error taxonomy.

### Tests added or modified

See `implementation.md` §Tests added or modified.

### Commits

Working-tree revision `1626b11`.

### Pull requests

None yet — tracked in working tree.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

See `implementation.md` §Known limitations.

### Follow-up items

See `implementation.md` §Follow-up items.

### Relationship to the approved plan

Matches `plan.md`; no deviations recorded in `deviations.md`.



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

*Not applicable.*

### Security changes

*Not yet implemented — anticipated: closes a zero/unknown-tenant pass-through gap.*

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

### Actual result

Pass.

### Pass or fail

Pass.

### Evidence identifier

See `evidence/index.md`.

### Execution date

2026-07-13.

### Commit or revision

Working-tree revision `1626b11`.

### Environment

Local Docker Postgres 16; `WOWAPI_TEST_DSN=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`; Go 1.26.5.

### Reviewer

Independent review completed for T004.

### Findings

None.

### Retest status

Initial pass.

### Final conclusion

Acceptance criterion satisfied.



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

*No deviations recorded.*



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
