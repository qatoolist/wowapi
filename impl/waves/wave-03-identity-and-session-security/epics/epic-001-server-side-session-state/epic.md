---
id: W03-E01
type: epic
title: Server-side session state
status: in-progress
wave: W03
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - SEC-01
  - D-01
  - DEC-Q1
  - CS-07
depends_on: []
stories:
  - W03-E01-S001
  - W03-E01-S002
  - W03-E01-S003
  - W03-E01-S004
decisions:
  - ADR-W00-E02-S003-001
risks:
  - RISK-W03-001
  - RISK-W03-002
  - RISK-W03-004
  - RISK-W03-005
---

# W03-E01 — Server-side session state

## Epic objective

Resolve SEC-01 in full (PLAN §5.2 T1–T7): build the server-side `identity_grant` table, make
tenant-membership verification unconditional, add server-side capacity selection, replace direct
JWT-claim trust for impersonation/break-glass with a verified grant-table resolver, bind assurance
freshness (`auth_time`/`acr`/`amr`) to step-up, distinguish credential schemes explicitly, and
produce the documentation/coordination plan for wowsociety's two-repo cutover of the resulting
breaking change.

## Problem being solved

`requirement-inventory.md` row SEC-01 records: "Server-side tenant/privileged session state
(T1–T7)" — class IMPL, priority P0, disposition `planned`, target `W03-E01-S001..S004`, notes
"D-01 ratified; DEC-Q1 safe default unblocks; BREAKING wowsociety." MATRIX CS-07 independently
confirms this is "the top-ranked security risk" in the whole review (§A). The defect, per PLAN
§5.2's own evidence citation: `Verifier.Actor` (`kernel/auth/auth.go:181-208`) validates membership
only when `CapacityID != uuid.Nil` — a capacity-less actor gets zero membership check — and
`TenantID`/`ImpersonatorUserID`/`BreakGlass` are copied straight from JWT claims with no
server-side verification. The target `user_tenant_access` table already exists in migrations
(`00002_core_identity.sql:54-83`) but, per the same evidence, **no Go code queries it**. No
break-glass or impersonation grant table exists at all — PLAN's own words: "genuinely greenfield
schema."

The consequence, per MATRIX CS-07: "server trusts client-presented session/impersonation state; a
validly-signed token with stale/forged tenant or impersonation claims is honoured," leading to
"tenant-isolation bypass via stale membership; unauditable impersonation."

## Scope

- T1 — `identity_grant` migration (break-glass + impersonation activation records: status, tenant,
  actor, impersonated user, approver, reason, activation/expiry/revocation, opaque grant ID); RLS
  FORCE per blueprint 03 §1; unique partial index enforcing one active grant per actor;
  `app_platform`-only grant writes (S001).
- T2 — extend `PrincipalStore` with `ActiveTenantAccess(ctx, userID, tenantID) error` against the
  existing `user_tenant_access` table; call unconditionally in `Verifier.Actor` (S001).
- T3 — reject zero/unknown tenant claims before opening a tenant transaction (S001).
- T4 — require explicit server-side-validated capacity choice when more than one capacity is
  active (S002).
- T5 — privileged-session resolver replacing the direct claim copy of `ImpersonatorUserID`/
  `BreakGlass` with a T1 grant-table lookup by opaque grant ID (S002).
- T6 — bind `auth_time`/`acr`/`amr` into assurance; enforce freshness for step-up (S003).
- T7 — distinguish user/API-key/webhook/internal credential schemes explicitly (S003).
- S004 — sequencing, staging-validation, and rollback coordination plan for wowsociety's two-repo
  cutover (PROD-04) — documentation/verification only, no product code.

## Out of scope

- Resolving DEC-Q1 (the IdP `grant_id` claim contract) itself — this epic proceeds against its
  documented safe default (REVIEW §F row 1 / MATRIX CS-07), it does not resolve the human decision.
- Any wowsociety code change (product-level, PROD-04) — S004 produces coordination documentation
  only; the actual `startImpersonation`/`stopImpersonation` rework, `identity_impersonation_session`
  schema change, and `whoami.go` rewrite happen in wowsociety's own repository, outside this
  programme's framework-implementation scope (mandate §2.3).
- DATA-07's party-subject relationship evaluation — that is W03-E04's scope, which hard-depends on
  this epic's acceptance but is a separate epic with its own story.
- SEC-04's cache-bounding/epoch work — that is W05-E04's scope; this epic's resolver does not
  introduce or modify an authorization cache.
- SEC-05's versioned security verification profile — that is W07-E02's scope; this epic supplies
  test evidence that profile will later consume, it does not build the profile itself.

## Source requirements

SEC-01 (PLAN §5.2 T1–T7). Cross-referenced: D-01 (`ADR-W00-E02-S003-001` — framework owns grant
validity/expiry/revocation), DEC-Q1 (IdP claim contract, human-blocked, safe default applied),
MATRIX CS-07 (closure spec, fail-first test-class list).

## Architectural context

This epic is the framework's principal identity-trust boundary correction. Today, per PLAN's
evidence, `Verifier.Actor` treats JWT claims as authoritative for tenant membership (conditionally),
impersonation, and break-glass state — the server performs no independent verification against
persisted state for these fields. This epic inverts that: the server becomes the source of truth,
consulting `user_tenant_access` (T2) and the new `identity_grant` table (T1, T5) on every request
that carries these claims, rather than trusting the claims themselves. This is a foundational
contract change — mandate §2.2's "foundational contracts before adapters" principle is why this
epic is sequenced early in W03 and why W05 (the ApplicationModel/registrar work) explicitly waits
for this epic's acceptance (`impl/analysis/wave-allocation-detail.md`'s cross-wave note: "actor
model stability").

The affected layers are `kernel/auth/` (`Verifier.Actor`, the new privileged-session resolver),
`kernel/principal` or equivalent (`PrincipalStore.ActiveTenantAccess`), the database layer (new
`identity_grant` migration, RLS policy), and — indirectly, via S004's coordination plan only —
wowsociety's `internal/modules/identity/impersonation.go` and `whoami.go`.

## Included stories

- **W03-E01-S001 — grant-schema-and-membership** (SEC-01 T1, T2, T3): the `identity_grant`
  migration, unconditional `ActiveTenantAccess` membership verification, and zero/unknown-tenant
  rejection. D-01 enacted; DEC-Q1 safe default recorded in the story's assumptions.
- **W03-E01-S002 — capacity-and-privileged-resolver** (SEC-01 T4, T5): server-side capacity
  selection and the privileged-session resolver.
- **W03-E01-S003 — assurance-and-credential-schemes** (SEC-01 T6, T7): assurance freshness and
  explicit credential-scheme distinction.
- **W03-E01-S004 — cross-repo-cutover-plan** (PROD-04 coordination): sequencing, staging
  validation, and rollback plan — documentation/verification story, no product code.

## Dependencies

Within this epic: S002 depends on S001 (PLAN: T4/T5 both depend on T2; T5 depends on T1). S003
depends on S001 (T6/T7 depend on T2). S004 depends on S001 and S002 being substantially planned
(the coordination plan needs T1's grant-table shape and T5's resolver contract to sequence
against), though S004 can be drafted in parallel and refined as S001/S002 firm up. This epic
depends on W00-E02-S003's ADR-ification of D-01 (`ADR-W00-E02-S003-001`) and on W01/W02 at wave
scope (validation seam, DATA-09 protocol) — see `../../dependencies.md`.

## Risks

RISK-W03-001 (DEC-Q1 remaining unresolved), RISK-W03-002 (wowsociety two-repo cutover cannot be
completed unilaterally), RISK-W03-004 (SEC-01 T2's "every session has a live `user_tenant_access`
row" precondition may not hold against real data), RISK-W03-005 (SEC-01 T4's capacity-selection
requirement may break a currently-working capacity-less flow) — all inherited from `../../risks.md`
(wave-level); see `risks.md` (epic-level) for epic-scoped elaboration.

## Required decisions

None new. This epic enacts already-ratified `ADR-W00-E02-S003-001` (D-01: framework owns grant
validity/expiry/revocation), see W00-E02-S003. DEC-Q1 is a separate, still-open human decision this
epic proceeds against via its documented safe default — this epic does not resolve DEC-Q1, and does
not author any new ADR for it. Accordingly, only W03-E01-S001 carries a `decisions/` directory
(referencing `ADR-W00-E02-S003-001`); the other three stories carry `decisions: []`.

## Epic acceptance criteria

- **AC-W03-E01-01**: `identity_grant` exists with RLS FORCE, a unique partial index for
  one-active-grant-per-actor, and `app_platform`-only write access; `PrincipalStore
  .ActiveTenantAccess` is called unconditionally in `Verifier.Actor`; a zero/unknown tenant claim
  is rejected before `WithTenantID` is called. Traces to W03-E01-S001.
- **AC-W03-E01-02**: A capacity-less actor with >1 active capacity is rejected pending explicit
  server-side-validated choice; `ImpersonatorUserID`/`BreakGlass` are populated only from a
  verified `identity_grant` lookup by opaque grant ID, never trusted off the JWT. Traces to
  W03-E01-S002.
- **AC-W03-E01-03**: A stale `auth_time` with valid `amr` fails step-up; a permission scoped to
  `CredentialUser` rejects a valid API-key actor. Traces to W03-E01-S003.
- **AC-W03-E01-04**: Sequencing, staging-validation, and rollback plans exist and are reviewed for
  the wowsociety cutover; no product code is written by this epic. Traces to W03-E01-S004.
- **AC-W03-E01-05**: All four stories have passed independent review per mandate §14, with
  S001/S002/S003 specifically checked against the full SEC-01 required test-class list (token
  substitution, zero-tenant, stale membership, revoked capacity, expired step-up, issuer/audience/
  key rotation, JWKS failure).

## Closure conditions

All four stories reach `accepted`; AC-W03-E01-01 through AC-W03-E01-05 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date; no
unresolved regression from the unconditional-membership or capacity-selection enforcement changes.

## Status update (2026-07-16)

`status: in-progress` — S001 and S002 accepted; S003 verified-pending-human (formal
product-security-lead sign-off outstanding, see `impl/tracking/deferred-items-register.md` DEF-07);
S004 implemented (cross-repo wowsociety reviewer sign-off unverifiable in-repo, acceptance
deferred). Epic cannot reach `accepted` until S003/S004 resolve.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
