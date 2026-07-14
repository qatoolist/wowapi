---
id: W03-E01-S002
type: story
title: Capacity selection and privileged-session resolver
status: ready
wave: W03
epic: W03-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - SEC-01
depends_on:
  - W03-E01-S001
blocks:
  - W03-E01-S004
acceptance_criteria:
  - AC-W03-E01-S002-01
  - AC-W03-E01-S002-02
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W03-005
---

# W03-E01-S002 — Capacity selection and privileged-session resolver

## Story ID

W03-E01-S002

## Title

Capacity selection and privileged-session resolver

## Objective

Require explicit, server-side-validated capacity choice whenever an actor has more than one active
capacity, and replace the direct JWT-claim copy of `ImpersonatorUserID`/`BreakGlass` with a
privileged-session resolver that performs a T1 `identity_grant` lookup by opaque grant ID. This is
PLAN SEC-01 T4 and T5.

## Value to the framework

T4 closes a silent-default gap: today, a capacity-less actor with multiple active capacities is not
forced to make an explicit, validated choice — the framework has no mechanism to require or verify
one. T5 closes the second half of MATRIX CS-07's "top-ranked security risk": even after S001 makes
tenant membership unconditional, `ImpersonatorUserID` and `BreakGlass` are still, today, trusted
directly from the JWT claim with zero server-side verification. T5 replaces that trust with a
resolver that looks up a verified `identity_grant` row (built in S001) by opaque grant ID and
rejects every adversarial condition PLAN names explicitly: expired, revoked, wrong-tenant,
wrong-actor, forged-ID, and unauthorized-approver grants.

## Problem statement

PLAN §5.2 SEC-01's evidence, cited verbatim (shared with S001, focused here on the capacity and
claim-copy behavior specifically): "`Verifier.Actor` (`kernel/auth/auth.go:181-208`)... 
`TenantID`/`ImpersonatorUserID`/`BreakGlass` are copied straight from JWT claims." There is
currently no code path that requires an actor with multiple active capacities to make an explicit
choice, and no code path that verifies `ImpersonatorUserID`/`BreakGlass` against any persisted
record — both fields are accepted as-is from a validly-signed token, regardless of whether the
underlying grant is still active, belongs to the right tenant, or was ever legitimately approved.

## Source requirements

SEC-01 (T4, T5). Cross-referenced: MATRIX CS-07 (fail-first test-class list — this story is
responsible for the "revoked capacity" and (jointly with S003) "expired step-up"-adjacent classes);
DEC-Q1 (T5's grant-ID claim-contract dependency — safe default applied, see "Assumptions").

## Current-state assessment

Per PLAN §5.2's own evidence citation (to be re-confirmed at this story's own execution commit):

- `Verifier.Actor` performs no capacity-selection enforcement today — an actor with more than one
  active capacity is not required to make an explicit, server-validated choice among them.
- `ImpersonatorUserID` and `BreakGlass` fields on `Actor` are populated directly from JWT claims,
  with no lookup against any persisted grant record — because, per S001's own current-state
  assessment, no such record exists at all today.
- `pgprincipal.Store` exposes only `UserIDBySubject`/`ValidateCapacity` — no privileged-session
  resolution capability exists.

## Desired state

An actor with more than one active capacity is rejected unless it presents an explicit,
server-side-validated capacity choice. `ImpersonatorUserID` and `BreakGlass` are populated only by a
resolver that performs a lookup against S001's `identity_grant` table by opaque grant ID, rejecting
expired, revoked, wrong-tenant, wrong-actor, forged-ID, and unauthorized-approver grants — never
trusting the JWT claim's value directly.

## Scope

- T4 — server-side capacity-selection enforcement: reject a capacity-less actor with >1 active
  capacity pending explicit, validated choice.
- T5 — the privileged-session resolver: replace the direct claim copy of `ImpersonatorUserID`/
  `BreakGlass` with a grant-table lookup by opaque grant ID, covering the full adversarial matrix
  PLAN names (expired, revoked, wrong-tenant, wrong-actor, forged-ID, unauthorized-approver).

## Out of scope

- The `identity_grant` schema itself and unconditional tenant-membership verification — S001's
  scope, a hard prerequisite for this story.
- Assurance freshness (`auth_time`/`acr`/`amr`) and credential-scheme distinction — W03-E01-S003's
  scope.
- The wowsociety-side cutover — W03-E01-S004's coordination-documentation scope; no product code is
  written here.
- Resolving DEC-Q1 itself — this story's resolver is built against the documented safe default (the
  framework owns the grant record, keyed on grant-ID); the exact IdP claim-shape question remains
  open and human-blocked.

## Assumptions

- **DEC-Q1 safe default, T5-specific framing.** T5's own PLAN risk column states: "Breaking
  JWT-claim-contract change — needs a `grant_id` claim from the IdP; coordinate before merge,
  genuinely undecided today." This story builds the resolver against the safe default recorded in
  S001's assumptions (framework owns the grant record, looked up by grant-ID; if the IdP cannot yet
  emit a `grant_id` claim, the framework still owns the grant record and looks it up by session, per
  REVIEW §F row 1). The exact claim contract itself is not invented here — see `plan.md`'s
  "Unresolved questions."
- T4's capacity-selection mechanism (how a client presents its explicit choice — a new claim, a new
  request parameter, a new header) is not yet specified by PLAN beyond "require explicit capacity
  choice." This story records the mechanism as an implementation-time decision to be made and
  documented, not invented here, consistent with mandate §18.
- The exact product-side UX for capacity selection (PLAN's own risk note for T4: "needs a
  product-side UX") is assumed to be coordinated via W03-E01-S004, not designed within this story.

## Dependencies

Depends on W03-E01-S001 (PLAN: T4 depends on T2; T5 depends on T1 and T2 — both land in S001).
Blocks W03-E01-S004 at epic scope (the cutover plan needs T5's resolver contract to sequence
against).

## Affected packages or components

`kernel/auth/auth.go` (`Verifier.Actor` and the new privileged-session resolver); the
principal-store package (a new resolver method or a new resolver type consuming
`PrincipalStore`/the `identity_grant` table, exact shape to be determined at implementation time).

## Compatibility considerations

**T5 is a BREAKING change for wowsociety.** PLAN §5.2's wowsociety-impact prose, cited verbatim:
"`whoami.go:39,51` reads `actor.ImpersonatorUserID` directly off the framework `authz.Actor`,
populated from the unverified claim, by explicit design (comment: trusts the claim 'without a DB
re-check')." "Test files `abac_test.go:52-94`, `whoami_impersonation_test.go:31-56` construct
`authz.Actor{ImpersonatorUserID: ...}` directly — load-bearing test surface that will need
rewriting." The breaking-vs-compile-safe distinction depends entirely on whether the `Actor` struct
shape is preserved: "if the resolver preserves the `authz.Actor` struct shape but populates fields
more strictly, wowsociety compiles unchanged and gets a strict behavioral improvement... If fields
are renamed/removed (e.g. `BreakGlass` becomes a grant-status enum), `whoami.go`, `impersonation.go`,
and `whoami_impersonation_test.go:43` fail to compile." This story's own implementation preference,
per PLAN's stated recommendation, is to preserve the `Actor` struct shape. **Sequencing (PLAN's own
words): "two-repo coordinated cutover — wowapi ships T1+T5, wowsociety's auth flow adopts `grant_id`,
only then cut over; validate T2 against wowsociety staging data before making it unconditional."**
The full cutover sequencing is W03-E01-S004's coordination-plan scope; this story's obligation is to
ship T5 in a way that is compile-safe for wowsociety wherever the resolver's contract allows it.
T4's capacity-selection requirement similarly "breaks any currently-working capacity-less
multi-capacity flow — needs a product-side UX" (PLAN's own risk note) — see RISK-W03-005.

## Security considerations

T5 is the direct closure of MATRIX CS-07's "unauditable impersonation" consequence: after this
story, `ImpersonatorUserID`/`BreakGlass` cannot be forged by presenting a validly-signed token with
a manipulated claim value, because the value is no longer trusted directly — it is resolved against
a verified grant row. T4 closes a related but distinct gap: an actor silently defaulting to an
unintended capacity when multiple are active, rather than being forced to make (and have verified)
an explicit choice.

## Performance considerations

T5 adds a database round-trip (the grant-table lookup) to every request that carries an
impersonation or break-glass claim — a strictly bounded, low-volume code path relative to overall
request traffic, not expected to be a general performance concern, but not separately benchmarked
by this story unless implementation reveals otherwise.

## Observability considerations

Not mandated by this story's acceptance criteria. Recording resolver-rejection reasons (expired,
revoked, wrong-tenant, etc.) as structured log fields or metrics is a reasonable implementation-time
addition, useful for the eventual SEC-05 verification profile (W07) but not required scope here.

## Migration considerations

None beyond what S001 already introduces (`identity_grant`) — this story is Go-level logic
consuming that table, not a further schema change.

## Documentation requirements

Document the capacity-selection mechanism (once determined) and the privileged-session resolver's
full rejection matrix (expired, revoked, wrong-tenant, wrong-actor, forged-ID,
unauthorized-approver) in whatever documentation currently covers the identity/auth contract. This
documentation directly feeds W03-E01-S004's cutover plan and, eventually, SEC-01 T7's credential-
scheme documentation (S003).

## Acceptance criteria

- **AC-W03-E01-S002-01**: A capacity-less actor with more than one active capacity is rejected
  pending an explicit, server-side-validated capacity choice, proven by a multi-capacity test.
- **AC-W03-E01-S002-02**: `Actor` fields for impersonation/break-glass are populated only from a
  verified `identity_grant` row, never trusted off the JWT directly, proven by an adversarial test
  suite covering expired, revoked, wrong-tenant, wrong-actor, forged-ID, and unauthorized-approver
  grants — all rejected.

## Required artifacts

- Capacity-selection enforcement logic and its chosen mechanism.
- Privileged-session resolver implementation.
See `artifacts/index.md`.

## Required evidence

- Multi-capacity test output (AC-01).
- Adversarial privileged-session test output covering all six named rejection conditions (AC-02).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`:
`story.md` and `plan.md` complete, acceptance criteria numbered and measurable, dependency on
W03-E01-S001 recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically covering the full six-condition adversarial matrix for T5.

## Risks

RISK-W03-005 (T4's capacity-selection requirement may break a currently-working capacity-less
multi-capacity flow) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

T5's residual risk is structurally tied to DEC-Q1's eventual resolution (see RISK-W03-001 at
epic/wave scope) — this story's own scope reduces to low residual risk once its adversarial test
suite passes, but the broader claim-shape question remains open beyond this story's boundary. T4's
residual risk depends on product-side UX coordination via W03-E01-S004, tracked there.

## Plan

See `plan.md`.
