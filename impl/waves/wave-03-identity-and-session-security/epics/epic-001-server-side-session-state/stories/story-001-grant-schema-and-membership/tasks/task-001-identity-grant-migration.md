---
id: W03-E01-S001-T001
type: task
title: identity_grant migration (SEC-01 T1)
status: done
parent_story: W03-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W03-E01-S001-01
artifacts:
  - ART-W03-E01-S001-001
  - ART-W03-E01-S001-002
evidence:
  - EV-W03-E01-S001-001
  - EV-W03-E01-S001-002
---

# W03-E01-S001-T001 — identity_grant migration (SEC-01 T1)

## Task Definition

### Task objective

Author the `identity_grant` migration: break-glass and impersonation activation records (status,
tenant, actor, impersonated user, approver, reason, activation/expiry/revocation, opaque grant ID),
RLS FORCE per blueprint 03 §1, a unique partial index enforcing one active grant per actor, and
`app_platform`-only write access.

### Parent story

W03-E01-S001 — Grant schema and unconditional membership enforcement.

### Owner

unassigned

### Status

complete

### Dependencies

None.

### Detailed work

1. Confirm the current highest migration number and the existing schema conventions (column
   typing, RLS policy authoring pattern, role-grant pattern) by reading a recent comparable
   migration (e.g. `00002_core_identity.sql`) at this task's actual start commit.
2. Author the `identity_grant` migration: table definition with the named columns, RLS FORCE
   policy, unique partial index (predicate finalized against the chosen status representation),
   and `GRANT INSERT, UPDATE ... TO app_platform` (or equivalent) with no grant to `app_rt`.
3. Write the migration's down path (clean drop).
4. Route the migration through W02-E01's DATA-09 protocol per this wave's entry-criteria
   dependency; confirm which DATA-09 phases materially apply to a wholly-new, empty table.
5. Write a migration up/down test.
6. Write an RLS-catalog-extension test confirming FORCE is set and only `app_platform` can write.
7. Write a concurrency test proving the unique partial index rejects a second concurrent
   grant-activation attempt for the same actor.
8. Get security-lead sign-off before merge, per PLAN SEC-01 T1's own risk note ("Schema is
   genuinely new — get security-lead sign-off before merge").

### Expected files or components affected

A new migration file (path/numbering TBD at implementation time); RLS policy definitions;
role-grant statements.

### Expected output

`identity_grant` exists in the schema with RLS FORCE, the unique partial index, and
`app_platform`-only write access; migration up/down test and RLS catalog extension test both pass.

### Required artifacts

ART-W03-E01-S001-001 (migration up/down), ART-W03-E01-S001-002 (RLS policy).

### Required evidence

EV-W03-E01-S001-001 (migration test report), EV-W03-E01-S001-002 (RLS catalog report).

### Related acceptance criteria

AC-W03-E01-S001-01.

### Completion criteria

Migration applies and reverts cleanly; RLS catalog reflects FORCE and correct role grants; the
unique partial index is proven under a concurrency test; security-lead sign-off obtained.

### Verification method

Direct migration execution (up/down) plus a dedicated RLS-catalog query test and a concurrency
test, logged output retained as evidence.

### Risks

Schema is genuinely new (PLAN's own risk note) — requires security-lead sign-off before merge, not
a purely mechanical change.

### Rollback or recovery considerations

The down-migration cleanly drops `identity_grant`; since no other table references it within this
story's scope, rollback carries no cascading-dependency risk at this stage.

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

*Not yet implemented.*

### Security changes

*Not yet implemented — anticipated: RLS FORCE + `app_platform`-only write restriction.*

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
