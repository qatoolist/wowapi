---
id: W03-ACCEPTANCE
type: wave-acceptance
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03 — Wave-level acceptance

## AC-W03-01 — Server-side grant schema and unconditional membership enforcement live

The `identity_grant` table exists with RLS FORCE, a unique partial index enforcing one active grant
per actor, and `app_platform`-only grant writes; `PrincipalStore.ActiveTenantAccess` is queried
unconditionally in `Verifier.Actor` for every actor carrying a `TenantID`; a zero/unknown tenant
claim is rejected before a tenant transaction opens. Traces to W03-E01-S001.

## AC-W03-02 — Server-side capacity selection and privileged-session resolver live

A capacity-less actor with more than one active capacity is rejected, requiring explicit
server-side-validated capacity choice; `ImpersonatorUserID`/`BreakGlass` are populated only from a
verified `identity_grant` row looked up by opaque grant ID, never trusted off the JWT claim
directly. Traces to W03-E01-S002.

## AC-W03-03 — Assurance freshness and credential-scheme distinction live

A stale `auth_time` with an otherwise-valid `amr` fails step-up; user/API-key/webhook/internal
credential schemes are distinguished explicitly, such that a permission scoped to `CredentialUser`
rejects a valid API-key actor. Traces to W03-E01-S003.

## AC-W03-04 — Cross-repo cutover plan complete

Sequencing, staging-validation, and rollback plans exist for the wowsociety impersonation-flow
breaking change (PROD-04), each reviewed and accepted as a coordination artifact; no wowapi or
wowsociety product code is written by this story. Traces to W03-E01-S004.

## AC-W03-05 — Outbound-security escape-hatch governance live

`SharedFingerprint()` scope is confirmed (or extended) to cover the outbound allowlist with a
regression test; a boot-time report enumerates enabled egress exceptions with no credentials
exposed; allowlist configuration changes produce an audit-visible record; a `prod`-profile JWKS
client injection without a declared trusted-issuer allowlist fails readiness (D-07). Traces to
W03-E02-S001.

## AC-W03-06 — Webhook replay/dedup bound to authenticated data

The `Verifier` interface returns `(Envelope, error)`; `HMACVerifier` synthesizes `EventID`/
`OccurredAt` from authenticated body/receipt time only; `HandleInbound` sources replay-window/dedup
exclusively from `Envelope`; the adversarial tamper matrix (body/timestamp/event-ID/key-ID/
signature-version independently manipulated) passes. Traces to W03-E03-S001.

## AC-W03-07 — Relationship semantics complete for party-subject edges

`Checker.Has` evaluates party-subject edges (previously "not consulted yet" per its own code
comment) and every schema-enumerated `subject_kind`, with unsupported kinds failing closed; actor
attribution on `Relate`/mirror `Upsert` reuses DATA-06 T2's mechanism without reimplementing it.
Traces to W03-E04-S001.

## AC-W03-08 — Workflow ratification and durable privileged-operation audit complete

Ratification is implemented as a real definition field and state transition (override-then-ratify
happy path, pending-not-yet-effective, rejection reverts) or explicitly documented as an interim
reject posture for `ratify_by`-declaring definitions; every override produces a complete audit row
(actor, impersonator, grant ID, source/target states, reason, ratification outcome) in the same
transaction as the state jump, with audit-write failure rolling back the override. Traces to
W03-E05-S001.

## AC-W03-09 — Independent review passed

Every W03 story has passed independent review per mandate §14. E01's three implementation stories
(S001/S002/S003) are specifically checked against the full SEC-01 required test-class list (token
substitution, zero-tenant, stale membership, revoked capacity, expired step-up, issuer/audience/
key rotation, JWKS failure); E03 is specifically checked against the full tamper matrix; E04 is
specifically checked that it did not reimplement DATA-06 T2's actor-attribution mechanism; E05 is
specifically checked that the override-audit rollback behavior is proven under fault injection, not
merely asserted.

## Acceptance authority

Product-security lead (E01, E02, E03, E05); data/reliability lead jointly with product-security
lead (E04, given its hard SEC-01 dependency) — per `wave.md`'s accountable-role assignment.
