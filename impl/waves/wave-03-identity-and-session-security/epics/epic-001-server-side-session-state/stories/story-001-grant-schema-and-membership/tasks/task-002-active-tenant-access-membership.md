---
id: W03-E01-S001-T002
type: task
title: ActiveTenantAccess + unconditional membership check (SEC-01 T2)
status: done
parent_story: W03-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W03-E01-S001-02
artifacts:
  - ART-W03-E01-S001-003
  - ART-W03-E01-S001-004
  - ART-W03-E01-S001-005
evidence:
  - EV-W03-E01-S001-003
  - EV-W03-E01-S001-004
---

# W03-E01-S001-T002 — ActiveTenantAccess + unconditional membership check (SEC-01 T2)

## Task Definition

### Task objective

Extend `PrincipalStore` with `ActiveTenantAccess(ctx, userID, tenantID) error` against the existing
`user_tenant_access` table, and call it unconditionally in `Verifier.Actor` for every actor
carrying a `TenantID` — removing today's `CapacityID != uuid.Nil` gate that currently allows a
capacity-less actor to bypass the membership check entirely.

### Parent story

W03-E01-S001 — Grant schema and unconditional membership enforcement.

### Owner

unassigned

### Status

complete

### Dependencies

None hard (this task's Go-level work does not require `identity_grant` to exist yet — that table is
consumed by S002's resolver, not this task). Soft ordering with T001 recommended so the story's
migration and code changes land together, but not a blocking dependency.

### Detailed work

1. Read `kernel/auth/auth.go:181-208` (`Verifier.Actor`) and the principal-store implementation
   (PLAN cites `pgprincipal.Store`) at this task's actual start commit, confirming exact current
   line numbers and method set.
2. Run the `user_tenant_access` data-audit step: confirm (or characterize the gap in) "every
   existing valid session has a live `user_tenant_access` row," per PLAN's own risk note and
   RISK-W03-004's mitigation. Produce the audit report as ART-W03-E01-S001-005.
3. Implement `PrincipalStore.ActiveTenantAccess(ctx, userID, tenantID) error` against
   `user_tenant_access`, returning a distinguishable "no live membership" error versus a genuine
   infrastructure error.
4. Change `Verifier.Actor`'s call site to invoke `ActiveTenantAccess` unconditionally for every
   actor carrying a `TenantID`, removing the `CapacityID != uuid.Nil` gate on this specific check.
5. If the data-audit step (2) found a material gap, stage the unconditional-enforcement rollout
   behind a profile flag per RISK-W03-004's contingency, rather than enabling it unconditionally
   against incomplete data — document this decision explicitly if taken.
6. Write the adversarial test suite: revoked/absent/foreign-tenant membership rejected with a
   validly signed token; a capacity-less actor is now membership-checked (previously bypassed).

### Expected files or components affected

`kernel/auth/auth.go`; the principal-store implementation file.

### Expected output

`PrincipalStore.ActiveTenantAccess` implemented; `Verifier.Actor` calls it unconditionally; the
data-audit report; adversarial test suite passing.

### Required artifacts

ART-W03-E01-S001-003 (`ActiveTenantAccess` implementation), ART-W03-E01-S001-004 (`Verifier.Actor`
call-site change), ART-W03-E01-S001-005 (data-audit report).

### Required evidence

EV-W03-E01-S001-003 (adversarial test report), EV-W03-E01-S001-004 (data-audit report as evidence).

### Related acceptance criteria

AC-W03-E01-S001-02.

### Completion criteria

`ActiveTenantAccess` is called unconditionally in `Verifier.Actor`; the adversarial test suite
proves revoked/absent/foreign-tenant membership rejected with a validly signed token; the data-audit
step has run and its result (clean or staged-rollout decision) is documented.

### Verification method

Adversarial test execution against a testkit DB seeded with fixture `user_tenant_access` rows;
direct review of the data-audit report.

### Risks

PLAN's own risk note: "Every existing valid session must have a live `user_tenant_access` row —
audit production data first." See RISK-W03-004.

### Rollback or recovery considerations

If unconditional enforcement rejects a significant volume of currently-valid sessions post-rollout,
stage behind a profile flag per RISK-W03-004's contingency rather than reverting the change
outright.

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

*Not yet implemented — anticipated: `PrincipalStore` gains `ActiveTenantAccess`.*

### Configuration changes

*Not yet implemented — anticipated: possible profile-flag addition if the data audit finds a gap.*

### Schema or migration changes

*Not applicable — this task consumes `user_tenant_access`, an existing table; no migration.*

### Security changes

*Not yet implemented — anticipated: closes the capacity-less membership-check bypass (MATRIX
CS-07's top-ranked risk).*

### Observability changes

*Not yet implemented.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated, unless the data audit forces a staged-rollout flag, which would itself be
tracked as a follow-up item to later remove once data gaps are closed.*

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
