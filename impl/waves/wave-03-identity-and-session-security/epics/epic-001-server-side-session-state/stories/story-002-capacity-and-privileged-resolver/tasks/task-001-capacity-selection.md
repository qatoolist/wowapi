---
id: W03-E01-S002-T001
type: task
title: Capacity-selection enforcement (SEC-01 T4)
status: done
parent_story: W03-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W03-E01-S002-01
artifacts:
  - ART-W03-E01-S002-001
evidence:
  - EV-W03-E01-S002-001
---

# W03-E01-S002-T001 — Capacity-selection enforcement (SEC-01 T4)

## Task Definition

### Task objective

Require an actor with more than one active capacity to present an explicit, server-side-validated
capacity choice, rejecting the case where no choice is made and the case where a client asserts a
capacity it does not actually hold.

### Parent story

W03-E01-S002 — Capacity selection and privileged-session resolver.

### Owner

unassigned

### Status

complete

### Dependencies

W03-E01-S001 must be `accepted` (unconditional membership verification is a prerequisite context
for capacity handling in `Verifier.Actor`).

### Detailed work

1. Confirm `ValidateCapacity`'s current behavior and `Verifier.Actor`'s current capacity handling at
   this task's actual start commit.
2. Design the capacity-selection mechanism (header, claim, or dedicated endpoint) — coordinate with
   whatever product-side UX guidance is available via W03-E01-S004's cutover-planning work; if none
   is yet available, choose the mechanism that keeps wowsociety's existing flows compile-safe
   wherever possible, and record the choice in `deviations.md` if it diverges from any assumption in
   `plan.md`.
3. Implement server-side validation of the presented capacity choice against the actor's actual
   active capacities — do not trust a client-asserted capacity ID without verification.
4. Reject an actor with >1 active capacity and no valid explicit choice.
5. Write the multi-capacity test: no-choice case rejected; valid-choice case accepted;
   unentitled-assertion case rejected.

### Expected files or components affected

`kernel/auth/auth.go` (or a new adjacent file for the capacity-selection logic, TBD at
implementation time).

### Expected output

Capacity-selection enforcement logic; a passing multi-capacity test covering all three named cases.

### Required artifacts

ART-W03-E01-S002-001 (capacity-selection enforcement logic).

### Required evidence

EV-W03-E01-S002-001 (functional test report).

### Related acceptance criteria

AC-W03-E01-S002-01.

### Completion criteria

The multi-capacity test proves: no-choice rejected, valid-choice accepted, unentitled-assertion
rejected.

### Verification method

Direct test execution against a testkit DB seeded with multi-capacity fixture actors, logged output
retained as evidence.

### Risks

PLAN's own risk note: "Breaks any currently-working capacity-less multi-capacity flow — needs a
product-side UX." See RISK-W03-005.

### Rollback or recovery considerations

Stage behind a profile flag if a currently-working capacity-less flow is found in active use with no
available product-side UX yet, per RISK-W03-005's contingency, rather than reverting outright.

## Implementation Record

### What was actually implemented

`PrincipalStore.ActiveCapacityCount` and a new branch in `Verifier.Actor` that rejects capacity-less
actors with more than one active capacity. The explicit choice mechanism is the existing optional
`CapacityID` claim in the JWT; when present, it is validated server-side via `ValidateCapacity`.

### Components changed

`kernel/auth`; `adapters/auth/pgprincipal`.

### Files changed

- `kernel/auth/auth.go`
- `kernel/auth/auth_test.go`
- `adapters/auth/pgprincipal/pgprincipal.go`
- `adapters/auth/pgprincipal/pgprincipal_test.go`

### Interfaces introduced or changed

- `auth.PrincipalStore.ActiveCapacityCount(ctx, userID, tenantID uuid.UUID) (int, error)` added.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

Closes the silent-capacity-default gap.

### Observability changes

None.

### Tests added or modified

- `kernel/auth/auth_test.go`: `TestActor_NoCapacitySingleCapacityAllowed`,
  `TestActor_NoCapacityMultipleCapacitiesRejected`, `TestActor_ExplicitCapacityValidatedServerSide`.
- `adapters/auth/pgprincipal/pgprincipal_test.go`: `TestActiveCapacityCount`.

### Commits

Working-tree changes on top of `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Pull requests

Not created in this session.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

None.

### Follow-up items

None.

### Relationship to the approved plan

Matches `plan.md`. The capacity-selection mechanism was chosen as the existing `CapacityID` claim
rather than a new header/endpoint; this is documented in code comments and the implementation
record. No deviation was required because `plan.md` left the mechanism open.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S002-01 | Multi-capacity test: no-choice, valid-choice, unentitled-assertion cases | Local dev or CI, testkit DB seeded with multi-capacity fixture actors | No-choice rejected; valid-choice accepted; unentitled-assertion rejected | functional test report | unassigned |

### Actual result

All relevant tests pass.

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E01-S002-001.

### Execution date

2026-07-13.

### Commit or revision

`733ef3e930cbb3f89f5bbc53d8f562c60e426513` (with working-tree changes).

### Environment

Local dev; Postgres at `postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`.

### Reviewer

Independent review pending (EV-W03-E01-S002-003).

### Findings

None.

### Retest status

No retest required.

### Final conclusion

AC-W03-E01-S002-01 passes at the implementation/verification level.

## Deviations Record

No deviations recorded.
