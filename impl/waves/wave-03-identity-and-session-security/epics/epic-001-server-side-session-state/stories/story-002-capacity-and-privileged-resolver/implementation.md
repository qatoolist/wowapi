---
id: IMPL-W03-E01-S002
type: implementation-record
parent_story: W03-E01-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W03-E01-S002

## What was actually implemented

- **T4 — capacity-selection enforcement.** `PrincipalStore` gained `ActiveCapacityCount`. `Verifier.Actor` now counts the actor's active capacities when the token carries no explicit `CapacityID`; if the count is greater than one, the request is rejected with `KindValidation` (`explicit capacity selection required`). When `CapacityID` is present, it is validated server-side via `ValidateCapacity` as before.
- **T5 — privileged-session resolver.** `Claims` gained an optional `GrantID`. `PrincipalStore` gained `ResolveGrant`, returning a `ResolvedGrant` or one of six typed `GrantRejection` reasons. `Verifier.Actor` now populates `ImpersonatorUserID`/`BreakGlass` only from a verified grant row; the legacy claim fields are ignored when no `GrantID` is present.
- **Typed rejection reasons.** `GrantRejection` constants (`grant_expired`, `grant_revoked`, `grant_wrong_tenant`, `grant_wrong_actor`, `grant_not_found`, `grant_unauthorized_approver`) and `IsGrantRejection` allow the adversarial test suite to assert the exact rejection condition.
- **Postgres implementation.** `pgprincipal.Store` implements `ActiveCapacityCount` under RLS and `ResolveGrant` on the platform manager (because `identity_grant` is `app_platform`-only). The resolver validates status, expiry, tenant, actor, and approver authority.

## Components changed

- `kernel/auth` — `Claims`, `PrincipalStore`, `Verifier.Actor`, new `ResolvedGrant`/`GrantRejection` types.
- `adapters/auth/pgprincipal` — `ActiveCapacityCount` and `ResolveGrant` implementations.
- `testkit` — `WithGrantID` token option.

## Files changed

- `kernel/auth/auth.go`
- `kernel/auth/auth_test.go`
- `adapters/auth/pgprincipal/pgprincipal.go`
- `adapters/auth/pgprincipal/pgprincipal_test.go`
- `testkit/auth.go`

## Interfaces introduced or changed

- `auth.PrincipalStore` extended with:
  - `ActiveCapacityCount(ctx, userID, tenantID uuid.UUID) (int, error)`
  - `ResolveGrant(ctx, userID, tenantID, grantID uuid.UUID) (*ResolvedGrant, error)`
- New types in `kernel/auth`:
  - `ResolvedGrant`
  - `GrantRejection`
  - `IsGrantRejection(err error, r GrantRejection) bool`

## Configuration changes

None.

## Schema or migration changes

None — this story reads S001's `identity_grant` table without altering it.

## Security changes

- Closes the silent multi-capacity default gap (T4).
- Closes MATRIX CS-07's unauditable-impersonation consequence by replacing direct JWT-claim trust for impersonation/break-glass with a verified grant-table lookup (T5).

## Observability changes

None required by this story. Each rejection reason is a distinct structured-error code, ready for future metrics/logging.

## Tests added or modified

- `kernel/auth/auth_test.go`:
  - `TestActor_NoCapacitySingleCapacityAllowed`
  - `TestActor_NoCapacityMultipleCapacitiesRejected`
  - `TestActor_ExplicitCapacityValidatedServerSide`
  - `TestActor_PrivilegedSessionResolvedFromGrant`
  - `TestActor_DirectImpersonationClaimIgnoredWithoutGrantID`
  - `TestActor_ForgedGrantIDRejected`
  - `TestActor_ExpiredGrantRejected`
  - `TestActor_RevokedGrantRejected`
  - `TestActor_WrongTenantGrantRejected`
  - `TestActor_WrongActorGrantRejected`
  - `TestActor_UnauthorizedApproverGrantRejected`
- `adapters/auth/pgprincipal/pgprincipal_test.go`:
  - `TestActiveCapacityCount`
  - `TestResolveGrant_ImpersonationSuccess`
  - `TestResolveGrant_BreakGlassSuccess`
  - `TestResolveGrant_ExpiredRejection`
  - `TestResolveGrant_RevokedRejection`
  - `TestResolveGrant_WrongTenantRejection`
  - `TestResolveGrant_WrongActorRejection`
  - `TestResolveGrant_NotFoundRejection`
  - `TestResolveGrant_UnauthorizedApproverRejection`

## Commits

Working-tree changes on top of `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

## Pull requests

Not created in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

None anticipated.

## Known limitations

- The exact IdP `grant_id` claim contract remains pending DEC-Q1. The implementation consumes `Claims.GrantID` per the documented safe default (framework owns the grant record).
- The unauthorized-approver authority model is interim: a privileged grant must name a distinct approver with active tenant membership. This may be refined when DEC-Q1 resolves.
- `identity_grant` has no explicit `break_glass` column; break-glass is inferred from `impersonated_user_id IS NULL`.

## Follow-up items

- W03-E01-S004: coordinate wowsociety cutover to framework `grant_id`.
- DEC-Q1 resolution: finalize IdP claim shape and approver authority model.

## Relationship to the approved plan

Implementation matches `plan.md`. The open questions from `plan.md` (capacity-selection mechanism, resolver interface shape, unauthorized-approver authority model) were resolved at implementation time as documented above and in the code comments; no deviation from the plan's constraints was required.
