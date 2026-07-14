---
id: W03-E01-S002-T002
type: task
title: Privileged-session resolver (SEC-01 T5)
status: complete
parent_story: W03-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W03-E01-S002-02
artifacts:
  - ART-W03-E01-S002-002
  - ART-W03-E01-S002-003
evidence:
  - EV-W03-E01-S002-002
---

# W03-E01-S002-T002 — Privileged-session resolver (SEC-01 T5)

## Task Definition

### Task objective

Replace the direct claim copy of `ImpersonatorUserID`/`BreakGlass` with a resolver that performs a
lookup against S001's `identity_grant` table by opaque grant ID, rejecting expired, revoked,
wrong-tenant, wrong-actor, forged-ID, and unauthorized-approver grants.

### Parent story

W03-E01-S002 — Capacity selection and privileged-session resolver.

### Owner

unassigned

### Status

complete

### Dependencies

W03-E01-S001 must be `accepted` (T5 depends on T1 and T2, both S001's scope).

### Detailed work

1. Confirm the exact current claim-copy code path for `ImpersonatorUserID`/`BreakGlass` at this
   task's actual start commit.
2. Design the resolver's interface: input (opaque grant ID, sourced per DEC-Q1's safe default —
   framework owns the grant record, looked up by grant-ID or session), output (verified grant record
   or a typed rejection reason).
3. Implement the six-condition rejection matrix: expired, revoked, wrong-tenant, wrong-actor,
   forged/unknown grant ID, unauthorized-approver — each independently testable and distinguishable.
4. Wire the resolver into `Verifier.Actor`, replacing the direct claim copy, preserving the `Actor`
   struct shape wherever the resolver logic allows, per PLAN's stated compatibility preference.
5. Write the adversarial test suite: one test case per rejection condition, plus a happy-path case
   (valid, active, correctly-scoped grant accepted).
6. Document the resolver's contract and rejection matrix (ART-W03-E01-S002-003), feeding
   W03-E01-S004's cutover plan.

### Expected files or components affected

`kernel/auth/auth.go`; a new resolver file (exact path TBD at implementation time).

### Expected output

The privileged-session resolver, wired into `Verifier.Actor`; a passing adversarial test suite
covering all six named conditions plus the happy path.

### Required artifacts

ART-W03-E01-S002-002 (resolver implementation), ART-W03-E01-S002-003 (documentation).

### Required evidence

EV-W03-E01-S002-002 (adversarial test report).

### Related acceptance criteria

AC-W03-E01-S002-02.

### Completion criteria

All six adversarial conditions independently rejected with distinguishable reasons; the happy-path
case (valid grant) accepted; `Actor` struct shape preserved wherever possible per the compatibility
strategy.

### Verification method

Direct adversarial test execution against a testkit DB seeded with fixture `identity_grant` rows
covering all six conditions, logged output retained as evidence.

### Risks

PLAN's own risk note: "Breaking JWT-claim-contract change — needs a `grant_id` claim from the IdP;
coordinate before merge, genuinely undecided today." This is the direct T5-scoped instance of
RISK-W03-001 (DEC-Q1 unresolved) and RISK-W03-002 (wowsociety two-repo cutover), both tracked at
epic/wave scope. This task's own scope proceeds against the documented safe default; it does not
resolve DEC-Q1 itself.

### Rollback or recovery considerations

Reverting the resolver back to the direct claim-copy behavior reopens MATRIX CS-07's top-ranked
security risk — any such rollback must be treated as a security-incident response, not a routine
deployment rollback, per `plan.md`'s "Rollback strategy."

## Implementation Record

### What was actually implemented

- Added `Claims.GrantID` and `PrincipalStore.ResolveGrant`.
- Added `ResolvedGrant`, `GrantRejection`, and `IsGrantRejection` in `kernel/auth`.
- `pgprincipal.Store.ResolveGrant` queries `identity_grant` by grant ID, validates status/expiry,
  tenant, actor, and approver authority, and returns typed rejection reasons.
- `Verifier.Actor` populates `ImpersonatorUserID`/`BreakGlass` only from a resolved grant; legacy
  claim values are ignored when `GrantID` is absent.
- `testkit.TokenIssuer` gained `WithGrantID`.

### Components changed

`kernel/auth`; `adapters/auth/pgprincipal`; `testkit`.

### Files changed

- `kernel/auth/auth.go`
- `kernel/auth/auth_test.go`
- `adapters/auth/pgprincipal/pgprincipal.go`
- `adapters/auth/pgprincipal/pgprincipal_test.go`
- `testkit/auth.go`

### Interfaces introduced or changed

- `auth.PrincipalStore.ResolveGrant(ctx, userID, tenantID, grantID uuid.UUID) (*ResolvedGrant, error)` added.
- `auth.Claims.GrantID uuid.UUID` added.
- New types: `auth.ResolvedGrant`, `auth.GrantRejection`, `auth.IsGrantRejection`.

### Configuration changes

None.

### Schema or migration changes

None — reads S001's `identity_grant` table.

### Security changes

Closes MATRIX CS-07's unauditable-impersonation consequence.

### Observability changes

None required; rejection reasons are distinct structured-error codes.

### Tests added or modified

- `kernel/auth/auth_test.go`: `TestActor_PrivilegedSessionResolvedFromGrant`,
  `TestActor_DirectImpersonationClaimIgnoredWithoutGrantID`, `TestActor_ForgedGrantIDRejected`,
  `TestActor_ExpiredGrantRejected`, `TestActor_RevokedGrantRejected`,
  `TestActor_WrongTenantGrantRejected`, `TestActor_WrongActorGrantRejected`,
  `TestActor_UnauthorizedApproverGrantRejected`.
- `adapters/auth/pgprincipal/pgprincipal_test.go`: `TestResolveGrant_ImpersonationSuccess`,
  `TestResolveGrant_BreakGlassSuccess`, `TestResolveGrant_ExpiredRejection`,
  `TestResolveGrant_RevokedRejection`, `TestResolveGrant_WrongTenantRejection`,
  `TestResolveGrant_WrongActorRejection`, `TestResolveGrant_NotFoundRejection`,
  `TestResolveGrant_UnauthorizedApproverRejection`.

### Commits

Working-tree changes on top of `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Pull requests

Not created in this session.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

- Exact IdP `grant_id` claim contract pending DEC-Q1; implementation consumes `Claims.GrantID` per
  the safe default.
- Unauthorized-approver authority model is interim: distinct approver with active tenant membership.
- `identity_grant` has no explicit `break_glass` column; break-glass inferred from
  `impersonated_user_id IS NULL`.

### Follow-up items

- DEC-Q1 resolution: finalize claim shape and approver authority model.
- W03-E01-S004: wowsociety cutover plan.

### Relationship to the approved plan

Matches `plan.md`. The resolver is implemented inside `kernel/auth/auth.go` and
`adapters/auth/pgprincipal/pgprincipal.go` rather than a separate resolver file; this is a
file-organization judgment call allowed by the plan. The `Actor` struct shape is preserved.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S002-02 | Adversarial test suite: expired, revoked, wrong-tenant, wrong-actor, forged-ID, unauthorized-approver grants; happy-path valid grant | Local dev or CI, testkit DB seeded with fixture `identity_grant` rows | All six conditions rejected with distinguishable reasons; happy path accepted | adversarial test report | unassigned |

### Actual result

All adversarial tests pass; happy-path grants accepted.

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E01-S002-002.

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

AC-W03-E01-S002-02 passes at the implementation/verification level.

## Deviations Record

No deviations recorded.
