---
id: W02-E02-S002-T004
type: task
title: Seeded cross-tenant insert negative tests, both roles (DATA-01 T7)
status: done
parent_story: W02-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E02-S002-T003
acceptance_criteria:
  - AC-W02-E02-S002-04
artifacts:
  - ART-W02-E02-S002-004
evidence:
  - EV-W02-E02-S002-006
---

# W02-E02-S002-T004 — Seeded cross-tenant insert negative tests, both roles (DATA-01 T7)

## Task Definition

### Task objective

Prove a seeded cross-tenant insert fails under both `app_rt` and `app_platform` roles, via a new
catalog-driven RLS matrix test, with the platform-role result specifically confirmed rather than
assumed.

### Parent story

W02-E02-S002 — Cross-tenant mismatch audit, composite FK validation, and negative tests.

### Owner

unassigned

### Status

todo

### Dependencies

**W02-E02-S002-T003** (PLAN's own Depends-on column for T7: "T5" — T003 here is this story's T5
equivalent). Not subject to the W02-E01 cross-wave gate itself (it depends on T003's validated
constraints existing, which is where that gate's effect propagates).

### Detailed work

1. Extend the existing catalog-driven RLS matrix test with seeded cross-tenant insert cases across
   all 8 edges.
2. Run the seeded cross-tenant insert under `app_rt` and assert it fails (violates the composite FK
   or RLS, whichever applies first).
3. Run the seeded cross-tenant insert under `app_platform` and **explicitly assert** on the result —
   per PLAN's own T7 risk note: "Confirm platform role doesn't bypass FK constraints — don't assume."
   Do not treat an unverified pass as sufficient.
4. If the platform role is found to bypass the new FK constraints, record this as a new finding
   requiring escalation, not a silently accepted or discarded result.

### Expected files or components affected

The existing catalog-driven RLS matrix test file(s) (exact path TBD at implementation time).

### Expected output

A seeded cross-tenant insert fails under both `app_rt` and `app_platform`, with the platform-role
result specifically confirmed by an explicit test assertion.

### Required artifacts

ART-W02-E02-S002-004 (the extended cross-tenant negative-test suite).

### Required evidence

EV-W02-E02-S002-006 (cross-tenant negative-test output under both roles).

### Related acceptance criteria

AC-W02-E02-S002-04.

### Completion criteria

The test suite proves rejection under both roles; the platform-role case is backed by an explicit
assertion, not an assumption.

### Verification method

Direct test execution against a testkit or staging DB with both role connections available, logged
output retained as evidence.

### Risks

If `app_platform` is found to bypass the new FK constraints, that is itself a new security finding
requiring escalation beyond this task's own scope to resolve.

### Rollback or recovery considerations

Not applicable — test-only task, no schema or production code change of its own beyond the test
suite extension.

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

*Not applicable.*

### Security changes

*Not applicable — this task tests existing security controls, it does not change them.*

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
| AC-W02-E02-S002-04 | Run the extended catalog-driven RLS matrix test with seeded cross-tenant inserts under both `app_rt` and `app_platform` | CI or staging environment, both role connections available | Insert fails under both roles; platform-role result explicitly asserted, not assumed | RLS matrix test report | unassigned |

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
