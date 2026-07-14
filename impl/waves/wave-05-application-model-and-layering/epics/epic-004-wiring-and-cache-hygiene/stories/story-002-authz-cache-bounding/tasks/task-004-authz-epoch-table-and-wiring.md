---
id: W05-E04-S002-T004
type: task
title: authz_epoch table and cross-pod epoch-bump wiring
status: todo
parent_story: W05-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E04-S002-T001
  - W05-E04-S002-T002
  - W05-E04-S002-T003
acceptance_criteria:
  - AC-W05-E04-S002-02
artifacts:
  - ART-W05-E04-S002-004
evidence:
  - EV-W05-E04-S002-004
---

# W05-E04-S002-T004 — authz_epoch table and cross-pod epoch-bump wiring

## Task Definition

### Task objective

Build a per-tenant `authz_epoch` table (D-06), checked on the authz read path, with an epoch bump
wired into the same transaction as every enumerated framework-side mutation path (role/permission
assignment writes in `kernel/authz`, seeds, and SEC-01's grant-table writes), so a revocation on one
pod is visible on another without a full TTL wait — resolving SEC-04 T4's own "Highest-risk task"
open architecture decision, per D-06's ratified resolution.

### Parent story

W05-E04-S002 — Bounded, epoch-invalidated authorization cache.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E04-S002-T001, W05-E04-S002-T002, W05-E04-S002-T003 (PLAN T4's own dependency row: "T1-T3").

### Detailed work

1. Design and migrate the `authz_epoch` table: per-tenant epoch integer, per D-06's own resolution.
2. Enumerate every known framework-side mutation path: role/permission assignment writes in
   `kernel/authz`, `kernel/seeds/seeds.go`, and SEC-01's grant-table writes (landed by this wave's
   entry gate, per MATRIX CS-17's own cross-CS sequencing note).
3. Wire an epoch bump into each enumerated mutation path, in the same transaction as the mutation.
4. Wire the epoch check into the authz read path: a stale local cache entry (epoch mismatch) is
   treated as a miss.
5. Write a simulated cross-pod test, producing `SEC-04/cross-pod-epoch-tests.md`: exercise a
   revocation via each enumerated mutation path, confirm visibility on a second simulated pod
   without a full TTL wait.
6. Document the mechanism, referencing D-06, and the enumerated mutation-path list explicitly (so a
   future contributor adding a new mutation path knows to also add its epoch bump).

### Expected files or components affected

A new `authz_epoch` migration; the mutation-path files enumerated in step 2; the authz read path in
`kernel/authz/caching.go`.

### Expected output

A revocation via any enumerated mutation path is visible cross-pod without a full TTL wait.

### Required artifacts

ART-W05-E04-S002-004.

### Required evidence

EV-W05-E04-S002-004.

### Related acceptance criteria

AC-W05-E04-S002-02.

### Completion criteria

The cross-pod test confirms visibility for every enumerated mutation path, with the epoch bump
confirmed atomic (same transaction) with each mutation.

### Verification method

Direct execution of the simulated cross-pod test against a live PostgreSQL instance.

### Risks

Highest-risk task in this epic, per PLAN's own risk column — see RISK-W05-005 in epic-level
`risks.md`. The architecture-decision component is resolved by D-06; the remaining risk is
mutation-path-enumeration completeness.

### Rollback or recovery considerations

If the cross-pod test reveals a missed mutation path or a non-atomic epoch bump, fix before
proceeding — do not ship a partially-correct epoch mechanism.

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

*Not yet implemented — the new `authz_epoch` table migration; recorded here once implemented.*

### Security changes

*Not yet implemented — this task's entire purpose is a security-adjacent correctness control;
recorded here once implemented.*

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
| AC-W05-E04-S002-02 | Run the simulated cross-pod epoch test | Local dev or CI, Go toolchain, PostgreSQL instance | Revocation visible cross-pod without a full TTL wait, for every enumerated mutation path | simulated cross-pod test report | unassigned |

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
