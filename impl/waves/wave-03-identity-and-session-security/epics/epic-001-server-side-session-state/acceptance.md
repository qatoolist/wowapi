---
id: W03-E01-ACCEPTANCE
type: epic-acceptance
epic: W03-E01
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E01 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" as a standalone,
independently-referenceable record, consistent with the wave-level `../../acceptance.md` pattern
(AC-W03-01 through AC-W03-04 there map onto this epic).

## AC-W03-E01-01 — Grant schema and unconditional membership enforcement

`identity_grant` exists with RLS FORCE, a unique partial index for one-active-grant-per-actor, and
`app_platform`-only write access. `PrincipalStore.ActiveTenantAccess(ctx, userID, tenantID) error`
is implemented against `user_tenant_access` and called unconditionally in `Verifier.Actor` for
every actor carrying a `TenantID` (not only when `CapacityID != uuid.Nil`). A zero/unknown tenant
claim is rejected before `WithTenantID` is invoked. Traces to W03-E01-S001.

## AC-W03-E01-02 — Capacity selection and privileged-session resolver

A capacity-less actor with more than one active capacity is rejected pending explicit
server-side-validated capacity choice. `ImpersonatorUserID`/`BreakGlass` fields on `Actor` are
populated only via a resolver that performs a T1 grant-table lookup by opaque grant ID and rejects
expired/revoked/wrong-tenant/wrong-actor/forged-ID/unauthorized-approver conditions — never a
direct copy from the JWT claim. Traces to W03-E01-S002.

## AC-W03-E01-03 — Assurance freshness and credential schemes

`auth_time`/`acr`/`amr` are bound into the assurance model; a stale `auth_time` with an otherwise
valid `amr` still fails step-up. User/API-key/webhook/internal credential schemes are distinguished
explicitly at the permission-check layer, such that a permission scoped to `CredentialUser` rejects
a valid API-key actor. Traces to W03-E01-S003.

## AC-W03-E01-04 — Cross-repo cutover plan

A sequencing plan (which repo ships what, in what order), a staging-validation plan (validating T2
against wowsociety staging data before unconditional enforcement), and a rollback plan exist,
reviewed by both a wowapi-side and (where practicable) a wowsociety-side reviewer. No wowapi or
wowsociety product code is written by this story. Traces to W03-E01-S004.

## AC-W03-E01-05 — Independent review passed

S001, S002, and S003 have each passed independent review per mandate §14, specifically confirming
every SEC-01 required test class (token substitution, zero-tenant, stale membership, revoked
capacity, expired step-up, issuer/audience/key rotation, JWKS failure) is exercised somewhere
across the three stories' evidence, with no class silently unaddressed. S004's review confirms it
produced only coordination documentation, with no product-code scope creep.

## Acceptance authority

Product-security lead (PLAN §5.2's stated accountable role for PF-SEC).
