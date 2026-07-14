---
id: W03-E01-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W03-E01-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E01-S002 — Evidence index

Per mandate §10.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W03-E01-S002-001 | functional test report | W03-E01-S002-T001 | AC-W03-E01-S002-01 | `go test -v ./kernel/auth/...` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 (with working-tree changes) | Pass: `TestActor_NoCapacitySingleCapacityAllowed`, `TestActor_NoCapacityMultipleCapacitiesRejected`, `TestActor_ExplicitCapacityValidatedServerSide` pass; plus `TestActiveCapacityCount` in `./adapters/auth/pgprincipal/...` | accepted |
| EV-W03-E01-S002-002 | adversarial test report | W03-E01-S002-T002 | AC-W03-E01-S002-02 | `go test -v ./kernel/auth/... ./adapters/auth/pgprincipal/...` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 (with working-tree changes) | Pass: all six rejection conditions (expired, revoked, wrong-tenant, wrong-actor, forged-ID, unauthorized-approver) rejected in both unit (`kernel/auth`) and integration (`adapters/auth/pgprincipal`) suites; happy-path impersonation and break-glass grants accepted | accepted |
| EV-W03-E01-S002-003 | review report | W03-E01-S002-T003 | AC-W03-E01-S002-01, AC-W03-E01-S002-02 | Independent review checklist per mandate §14 | — | Pending independent review | pending |

Evidence status vocabulary (per mandate §10): `accepted` for produced evidence that passed; `pending` for the outstanding independent-review evidence.

## Detailed evidence notes

### EV-W03-E01-S002-001 — Multi-capacity / capacity-selection tests

Executed:

```bash
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
WOWAPI_REQUIRE_DB=1 WOWAPI_REQUIRE_S3=1 \
go test -v ./kernel/auth/... ./adapters/auth/pgprincipal/...
```

Relevant passing tests:

- `TestActor_NoCapacitySingleCapacityAllowed` — capacity-less token with exactly one active capacity is accepted.
- `TestActor_NoCapacityMultipleCapacitiesRejected` — capacity-less token with two active capacities is rejected with `KindValidation`.
- `TestActor_ExplicitCapacityValidatedServerSide` — explicit `CapacityID` claim is validated server-side and accepted.
- `TestActor_CapacityNotYoursForbidden` — explicit `CapacityID` that does not belong to the actor is rejected with `KindForbidden`.
- `TestActiveCapacityCount` — real DB integration proving 0/1/2 active capacities counted correctly and cross-tenant capacities invisible under RLS.

### EV-W03-E01-S002-002 — Adversarial privileged-session tests

Executed with the same command. Relevant passing tests:

From `kernel/auth/auth_test.go` (unit-level via `fakePrincipalStore`):

- `TestActor_PrivilegedSessionResolvedFromGrant` — happy path: grant resolved, `ImpersonatorUserID`/`BreakGlass` populated.
- `TestActor_DirectImpersonationClaimIgnoredWithoutGrantID` — legacy claim fields ignored when no `grant_id`.
- `TestActor_ForgedGrantIDRejected` — `GrantRejectionNotFound`.
- `TestActor_ExpiredGrantRejected` — `GrantRejectionExpired`.
- `TestActor_RevokedGrantRejected` — `GrantRejectionRevoked`.
- `TestActor_WrongTenantGrantRejected` — `GrantRejectionWrongTenant`.
- `TestActor_WrongActorGrantRejected` — `GrantRejectionWrongActor`.
- `TestActor_UnauthorizedApproverGrantRejected` — `GrantRejectionUnauthorizedApprover`.

From `adapters/auth/pgprincipal/pgprincipal_test.go` (real DB integration):

- `TestResolveGrant_ImpersonationSuccess` — active impersonation grant accepted.
- `TestResolveGrant_BreakGlassSuccess` — active break-glass grant accepted.
- `TestResolveGrant_ExpiredRejection` — expired grant rejected.
- `TestResolveGrant_RevokedRejection` — revoked grant rejected.
- `TestResolveGrant_WrongTenantRejection` — cross-tenant grant rejected.
- `TestResolveGrant_WrongActorRejection` — grant for a different actor rejected.
- `TestResolveGrant_NotFoundRejection` — forged/unknown grant ID rejected.
- `TestResolveGrant_UnauthorizedApproverRejection` — missing, self-, and foreign-non-member approvers rejected.
