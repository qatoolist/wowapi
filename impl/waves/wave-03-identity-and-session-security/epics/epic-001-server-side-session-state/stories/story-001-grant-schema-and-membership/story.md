---
id: W03-E01-S001
type: story
title: Grant schema and unconditional membership enforcement
status: accepted
wave: W03
epic: W03-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - SEC-01
  - CS-07
depends_on: []
blocks:
  - W03-E01-S002
  - W03-E01-S003
  - W03-E04-S001
acceptance_criteria:
  - AC-W03-E01-S001-01
  - AC-W03-E01-S001-02
  - AC-W03-E01-S001-03
artifacts: []
evidence: []
decisions:
  - ADR-W00-E02-S003-001
risks:
  - RISK-W03-004
---

# W03-E01-S001 — Grant schema and unconditional membership enforcement

## Story ID

W03-E01-S001

## Title

Grant schema and unconditional membership enforcement

## Objective

Build the `identity_grant` table (break-glass and impersonation activation records, RLS FORCE, one
active grant per actor, `app_platform`-only writes), extend `PrincipalStore` with an unconditional
`ActiveTenantAccess` membership check called from `Verifier.Actor` for every actor carrying a
`TenantID`, and reject zero/unknown tenant claims before a tenant transaction opens. This is PLAN
SEC-01 T1, T2, and T3.

## Value to the framework

Per MATRIX CS-07, this is "the top-ranked security risk" in the entire architecture review. Today,
per PLAN's own evidence, a capacity-less actor receives zero membership check, and the
`user_tenant_access` table — which already exists in migrations — is queried by no Go code at all.
This story closes both gaps: it gives the framework a genuine, server-verified record of grant
state (the currently-nonexistent `identity_grant` table) and makes tenant-membership verification
unconditional rather than contingent on an actor happening to carry a non-nil `CapacityID`. This is
the foundational slice the rest of W03-E01 (capacity selection, the privileged resolver, assurance
freshness) and W03-E04 (relationship semantics) build on.

## Problem statement

PLAN §5.2 SEC-01's evidence, cited verbatim: "`Verifier.Actor` (`kernel/auth/auth.go:181-208`)
validates membership only when `CapacityID != uuid.Nil`; a capacity-less actor gets zero membership
check. `TenantID`/`ImpersonatorUserID`/`BreakGlass` are copied straight from JWT claims.
`pgprincipal.Store` exposes only `UserIDBySubject`/`ValidateCapacity` — no membership, break-glass,
or impersonation grant lookup exists. The target `user_tenant_access` table already exists in
migrations (`00002_core_identity.sql:54-83`) but no Go code queries it. No break-glass/
impersonation grant table exists at all — genuinely greenfield schema." MATRIX CS-07 restates the
consequence: "server trusts client-presented session/impersonation state; a validly-signed token
with stale/forged tenant or impersonation claims is honoured," producing "tenant-isolation bypass
via stale membership; unauditable impersonation."

## Source requirements

SEC-01 (T1, T2, T3). Cross-referenced: MATRIX CS-07 (fail-first test-class list, "top-ranked
security risk" framing); D-01 (`ADR-W00-E02-S003-001`, referenced not authored — see "Decisions"
below); DEC-Q1 (safe default applied, see "Assumptions" below).

## Current-state assessment

Per PLAN §5.2's own evidence citation (to be re-confirmed at this story's own execution commit,
consistent with the wave-01 precedent of re-running fail-first checks rather than trusting a cited
snapshot blindly):

- `Verifier.Actor` at `kernel/auth/auth.go:181-208` performs membership validation only when
  `CapacityID != uuid.Nil` — a capacity-less actor is not membership-checked at all.
- `TenantID`, `ImpersonatorUserID`, and `BreakGlass` fields on `Actor` are populated directly from
  JWT claims with no server-side verification.
- `pgprincipal.Store` exposes only `UserIDBySubject` and `ValidateCapacity` — no method exists for
  membership lookup, break-glass lookup, or impersonation-grant lookup.
- The `user_tenant_access` table exists in `00002_core_identity.sql:54-83` but is queried by no Go
  code anywhere in the codebase today.
- No break-glass or impersonation grant table exists anywhere in the schema — this is, per PLAN's
  own words, "genuinely greenfield schema," not an extension of an existing table.

## Desired state

`identity_grant` exists as a new table with: status, tenant, actor, impersonated user, approver,
reason, activation/expiry/revocation timestamps, and an opaque grant ID, under RLS FORCE per
blueprint 03 §1, with a unique partial index enforcing at most one active grant per actor, and
writable only by the `app_platform` role. `PrincipalStore.ActiveTenantAccess(ctx, userID,
tenantID) error` exists and is called unconditionally — not gated on `CapacityID != uuid.Nil` — from
`Verifier.Actor` for every actor carrying a `TenantID`. A zero or unknown tenant claim is rejected
before any `WithTenantID` call, not merely detected downstream.

## Scope

- The `identity_grant` migration (up/down), including its RLS policy and unique partial index.
- `PrincipalStore.ActiveTenantAccess` implementation against `user_tenant_access`.
- The call-site change in `Verifier.Actor` making membership verification unconditional.
- Zero/unknown-tenant rejection logic ahead of `WithTenantID`.
- A data-audit step confirming (or characterizing gaps in) "every existing valid session has a live
  `user_tenant_access` row" before unconditional enforcement is enabled, per PLAN's own risk note
  and this wave's RISK-W03-004.

## Out of scope

- Server-side capacity selection and the privileged-session resolver (SEC-01 T4/T5) — that is
  W03-E01-S002's scope, which depends on this story.
- Assurance freshness and credential-scheme distinction (SEC-01 T6/T7) — W03-E01-S003's scope.
- The wowsociety-side cutover itself — W03-E01-S004 produces the coordination plan; the actual
  product-repo rework is out of this programme's framework-implementation scope entirely.
- Resolving DEC-Q1 — this story proceeds against its documented safe default; see "Assumptions."
- DATA-07's party-subject relationship evaluation — W03-E04's scope, which depends on this story's
  acceptance but is implemented separately.

## Assumptions

- **DEC-Q1 safe default.** Per REVIEW §F row 1 and MATRIX CS-07: "Safe default per review §F Q1:
  framework owns the grant record keyed by grant-ID; IdP claim shape is tuning, not a blocker." This
  story builds the `identity_grant` table and its grant-ID-keyed lookup path now, against this safe
  default. **DEC-Q1 itself — the exact IdP `grant_id` claim contract, and who approves break-glass —
  remains an open, human-blocked decision** (`requirement-inventory.md` §B, disposition "blocked
  (human)", target "W03-E01 (tracked)"). This story does not resolve DEC-Q1; it records that
  implementation proceeds against the safe default, and that the human decision, when made, only
  tunes claim shape rather than requiring a redesign — per REVIEW §F row 1's own framing: "the
  human decision only tunes claim shape."
- The "every existing valid session must have a live `user_tenant_access` row" precondition (PLAN's
  own risk note for T2) is assumed true in principle but is explicitly *not* assumed true in
  practice without a data audit — see RISK-W03-004 and the "Scope" section's data-audit step.
- `user_tenant_access`'s existing schema (`00002_core_identity.sql:54-83`) is assumed sufficient for
  `ActiveTenantAccess`'s query needs without a schema change; to be confirmed by reading the actual
  migration at implementation time.

## Dependencies

None within this story's own prerequisites — it is the foundational slice of W03-E01. Depends on
W00-E02-S003's ADR-ification of D-01 (`ADR-W00-E02-S003-001`), and at wave scope on W01 (validation
seam) and W02 (DATA-09 protocol for the migration rollout). Blocks W03-E01-S002, W03-E01-S003
(both depend on T2's membership-check plumbing), and — at epic scope — W03-E04-S001 (DATA-07's hard
dependency on this epic's acceptance).

## Affected packages or components

`kernel/auth/auth.go` (`Verifier.Actor`); the principal-store package exposing `PrincipalStore`
(exact package path to be confirmed at implementation time — PLAN cites `pgprincipal.Store`); a new
migration file for `identity_grant`; the RLS policy/catalog extension covering the new table.

## Compatibility considerations

**This is a breaking change for wowsociety**, per PLAN §5.2's wowsociety-impact prose: "Affected,
HIGH severity, BREAKING for impersonation." `internal/modules/identity/impersonation.go:1-21`
states explicitly: "What the framework does NOT provide: a session/grant record... This file is
that product-side layer" — wowsociety has already built its own workaround
(`identity_impersonation_session` table, `startImpersonation`/`stopImpersonation`, audited via
`kaudit.Entry`). `whoami.go:39,51` reads `actor.ImpersonatorUserID` directly off the framework
`authz.Actor`, "populated from the unverified claim, by explicit design (comment: trusts the claim
'without a DB re-check')." Test files `abac_test.go:52-94` and `whoami_impersonation_test.go:31-56`
construct `authz.Actor{ImpersonatorUserID: ...}` directly — "load-bearing test surface that will
need rewriting." This story's own T2 change (unconditional membership check) is, per PLAN, a
"strict behavioral improvement" if the `Actor` struct shape is preserved — wowsociety compiles
unchanged but gets a runtime behavior change (a previously-trusted-but-invalid state now correctly
rejected) that some current callers may be relying on. The full breaking-vs-compile-safe cutover
sequencing is W03-E01-S004's scope; this story's own compatibility obligation is to build T1/T2/T3
in a way that keeps the `Actor` struct shape stable wherever possible, per PLAN's own preference.

## Security considerations

This story directly closes MATRIX CS-07's "top-ranked security risk": tenant-isolation bypass via
stale/unchecked membership. RLS FORCE on `identity_grant` and the `app_platform`-only write
restriction are both explicit blueprint-mandated controls (blueprint 03 §1), not optional hardening.
The zero/unknown-tenant rejection (T3) closes a distinct adjacent gap: a garbage or missing tenant
claim reaching `WithTenantID` at all.

## Performance considerations

`ActiveTenantAccess` adds a database round-trip to every request carrying a `TenantID` claim that
did not previously incur one (the capacity-less path). This is an accepted, required cost of
closing the security gap — not separately optimized in this story unless the fresh-run/data-audit
step surfaces an unacceptable latency regression, in which case it is recorded as a finding, not
silently absorbed.

## Observability considerations

The story does not itself add new metrics beyond what a reasonable implementation would emit for a
security-critical database call (e.g., a counter for membership-check outcomes) — left as an
implementation-time judgment call, not a required scope item, consistent with wave-01's pattern for
similarly-scoped stories.

## Migration considerations

The `identity_grant` migration is, per PLAN's own risk column for T1, "genuinely new — get
security-lead sign-off before merge." It should be authored and rolled out through W02-E01's
DATA-09 online expand/backfill/validate/contract protocol per this wave's entry criteria, even
though as a wholly new table (not an alteration of an existing one) its risk profile differs from
DATA-01's composite-FK-on-existing-table work — the exact DATA-09 phases this migration needs (it
has no backfill target since the table starts empty) are to be determined at implementation time,
not pre-specified here.

## Documentation requirements

Document the `identity_grant` schema (columns, RLS policy, unique partial index) and the
`ActiveTenantAccess` contract in whatever documentation currently covers the identity/auth schema.
Record the zero/unknown-tenant rejection behavior as an explicit, testable contract in the same
documentation.

## Acceptance criteria

- **AC-W03-E01-S001-01**: A migration creates `identity_grant` with RLS FORCE, a unique partial
  index enforcing at most one active grant per actor, and `app_platform`-only write access; a
  migration up/down test and an RLS-catalog-extension test both pass.
- **AC-W03-E01-S001-02**: `PrincipalStore.ActiveTenantAccess(ctx, userID, tenantID) error` is
  implemented and called unconditionally in `Verifier.Actor` for every actor carrying a `TenantID`
  (not gated on `CapacityID != uuid.Nil`); an adversarial test proves a revoked, absent, or
  foreign-tenant membership is rejected even with a validly signed token.
- **AC-W03-E01-S001-03**: A zero or garbage-UUID tenant claim is rejected before `WithTenantID` is
  called, proven by a negative test.

## Required artifacts

- `identity_grant` migration (up/down files).
- RLS policy/catalog extension for the new table.
- `PrincipalStore.ActiveTenantAccess` implementation.
- `Verifier.Actor` call-site change (unconditional membership check).
- Zero/unknown-tenant rejection logic.
See `artifacts/index.md`.

## Required evidence

- Migration up/down test output; RLS catalog extension test output.
- Adversarial membership test output (revoked/absent/foreign-tenant membership rejected with a
  validly signed token).
- Zero/unknown-tenant negative test output.
- The `user_tenant_access` data-audit report (per RISK-W03-004's mitigation).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`:
`story.md` and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (D-01
ADR reference) recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three
acceptance criteria verified with evidence in `evidence/index.md`; `closure.md` completed;
independent review passed per mandate §14, specifically covering the adversarial test classes named
in `plan.md`'s testing strategy.

## Risks

RISK-W03-004 (the "every session has a live `user_tenant_access` row" precondition may not hold
against real data) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once the data-audit step (mitigating RISK-W03-004) is executed and either confirms the precondition
or characterizes and stages around any gap found, no residual risk beyond the structural DEC-Q1/
wowsociety-cutover risks tracked at epic/wave scope is expected to remain open against this story
specifically.

## Plan

See `plan.md`.
