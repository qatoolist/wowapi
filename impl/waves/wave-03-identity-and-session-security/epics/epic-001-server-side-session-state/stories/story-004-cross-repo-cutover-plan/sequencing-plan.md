# Sequencing plan — wowsociety impersonation-flow cutover (SEC-01 T1/T5)

## Scope

This document governs the two-repo coordinated cutover of wowsociety's impersonation flow onto
the wowapi framework's new server-side `identity_grant` table and privileged-session resolver
(W03-E01-S001 / W03-E01-S002). It is a coordination artifact only: no wowapi or wowsociety product
code is modified by this document.

## Authority decision

Per REVIEW §U decision D-01 and `docs/tracking/decision-register.md`: **the framework owns grant
validity/expiry/revocation.** wowsociety's existing `identity_impersonation_session` table remains
a product-layer record for UX and audit but is no longer the authority for whether an
impersonation is valid. The framework `identity_grant` row is the authority.

## Pre-conditions

- W03-E01-S001 accepted: `identity_grant` table exists in wowapi with RLS FORCE, partial index,
  and adversarial membership tests passing.
- W03-E01-S002 accepted: `PrincipalStore.ResolveGrant` contract and `Verifier.Actor` grant-ID
  resolver are stable; `Claims.GrantID` is the supported claim.
- wowsociety production `user_tenant_access` data audit completed (RISK-W03-004 / S001 deferred
  work) and confirms no active session belongs to a revoked/inactive membership.
- wowsociety engineering owner identified and committed to the cutover window.

## Three-phase sequence

### Phase 1 — wowapi ships T1 + T5

1. Merge W03-E01-S001 and W03-E01-S002 to wowapi `main`.
2. Tag/release wowapi version containing:
   - `identity_grant` migration (`adapters/auth/pgprincipal` migration).
   - `PrincipalStore.ActiveCapacityCount` and `ResolveGrant`.
   - `Verifier.Actor` grant-ID resolver.
   - `testkit.WithGrantID` and adversarial grant tests.
3. Run wowapi `make ci` and `make ci-container` 0-FAIL 0-SKIP.
4. Deploy wowapi to wowsociety staging with the new migration applied.

**wowsociety impact during Phase 1:** none. wowsociety continues to read
`actor.ImpersonatorUserID` from the JWT claim directly, because its tokens carry no `grant_id`
claim yet. The framework resolver ignores the claim when `GrantID` is absent, so the cutover is
backward compatible at this stage.

### Phase 2 — wowsociety adopts `grant_id`

1. In wowsociety, add `grant_id uuid` (nullable, indexed) to
   `identity_impersonation_session`.
2. Update `internal/modules/identity/impersonation.go`:
   - `startImpersonation` mints a framework `identity_grant` row via the framework API/port and
     stores the returned `grant_id` in `identity_impersonation_session`.
   - `stopImpersonation` revokes both the framework grant row and the wowsociety session row.
3. Update the IdP/token-issuance path (or a post-issuance enrichment step) to include
   `grant_id` in the JWT for impersonated sessions.
4. Update `internal/modules/identity/whoami.go:39,51` to trust
   `actor.ImpersonatorUserID` only when `actor` was populated from a verified grant (i.e., always,
   because the framework now guarantees it). Remove or comment the "trusts the claim without a DB
   re-check" note.
5. Rewrite the load-bearing test fixtures:
   - `internal/modules/identity/abac_test.go:52-94`
   - `internal/modules/identity/whoami_impersonation_test.go:31-56`
   Use `testkit.WithGrantID` and seed matching `identity_grant` rows instead of constructing
   `authz.Actor{ImpersonatorUserID: ...}` literals.
6. Run wowsociety unit/integration tests locally.
7. Deploy wowsociety to staging with the schema change and code changes behind a feature flag or
   dark-launch gate if possible.

### Phase 3 — coordinated cutover

1. **Staging validation drill** (see `staging-validation-plan.md`):
   - Validate S001's T2 unconditional membership enforcement against wowsociety staging data.
   - Validate S002's T5 resolver against wowsociety staging impersonation sessions.
   - Re-run `abac_test.go`, `whoami_impersonation_test.go`, and `rls_test.go`.
2. Declare go/no-go (see staging-validation plan for criteria).
3. Cut over production:
   - Option A (preferred): feature-flag `use_framework_grant_id` enabled for a small cohort, then
     ramped to 100%.
   - Option B: hard cutover at a scheduled maintenance window if wowsociety lacks feature-flag
     infrastructure.
4. Monitor wowsociety impersonation success rate and framework grant-table lookup rate for the
   duration of the cutover window.

## Named wowsociety files/tests requiring rework

| File / test | Why it changes | Owner |
|---|---|---|
| `internal/modules/identity/whoami.go:39,51` | Stops trusting raw JWT claim; relies on framework `Actor.ImpersonatorUserID` populated from grant row. | wowsociety identity team |
| `internal/modules/identity/impersonation.go` | Mints framework grant row, stores `grant_id`, revokes grant on stop. | wowsociety identity team |
| `internal/modules/identity/abac_test.go:52-94` | Fixtures build `authz.Actor{}` literals directly; must use `WithGrantID` + seeded grant rows. | wowsociety identity team |
| `internal/modules/identity/whoami_impersonation_test.go:31-56` | Same fixture issue. | wowsociety identity team |

## Coordination checklist

- [ ] wowapi S001/S002 merged and released.
- [ ] wowsociety engineering owner assigned.
- [ ] wowsociety `identity_impersonation_session` schema change designed (follow DATA-09 online
      migration discipline).
- [ ] IdP `grant_id` claim path confirmed (DEC-Q1 safe default: framework-owned grant record).
- [ ] Staging validation plan reviewed and signed off.
- [ ] Rollback plan reviewed and signed off.
- [ ] Cutover go/no-go criteria agreed.

## Out of scope

- The actual wowsociety code changes (product-level work, PROD-04).
- Resolving DEC-Q1's final IdP claim contract — this plan assumes the safe default
  (framework-owned grant record keyed by `grant_id`).
- Executing the cutover — this document only plans it.

## Review status

- wowapi-side reviewer: self-review as part of W03-E01-S004 acceptance.
- wowsociety-side reviewer: to be assigned by wowsociety engineering owner.
