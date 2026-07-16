---
id: W03
type: wave
title: Identity and session security
status: in-progress
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
included_epics:
  - W03-E01
  - W03-E02
  - W03-E03
  - W03-E04
  - W03-E05
depends_on:
  - W01
  - W02
blocks:
  - W05
source_requirements:
  - SEC-01
  - SEC-06
  - SEC-03
  - DATA-07
  - SEC-02
  - D-01
  - D-07
  - DEC-Q1
  - CS-07
---

# W03 — Identity and session security

## Objective

Resolve the framework's top-ranked security risk (`fable5-closure-depth-matrix-2026-07-11.md`
CS-07: "This is the top-ranked security risk (§A)") by moving tenant membership, break-glass, and
impersonation state out of client-presented JWT claims and into a server-side, framework-owned
grant record; govern the explicit outbound-security escape hatches (JWKS client injection,
allowlisted-host egress); bind webhook replay-window/dedup controls to provider-authenticated data
instead of caller-supplied fields; complete relationship-semantics actor attribution and
party-subject evaluation on top of the new principal model; and close the two remaining P0 tasks
of workflow privileged-operation fail-closed behavior (ratification, durable audit).

## Rationale

`impl/index.md`'s wave map assigns W03 exactly this bundle: "SEC-01 server-side session state
(+D-01, DEC-Q1 safe default), SEC-06, SEC-03, DATA-07 (dep SEC-01), SEC-02 remainder." This is the
programme's security-hardening wave, sequenced after W01 (so SEC-01's new grant-table endpoints
are built against the `RouteMeta` central-validation seam W01-E03-S002 establishes, not before it)
and after W02 (so SEC-01's `identity_grant` migration — "Schema is genuinely new" per PLAN SEC-01
T1 — is authored and rolled out using the DATA-09 online expand/backfill/validate/contract
protocol W02-E01 delivers, rather than a one-off unsafe migration). `requirement-inventory.md`
records SEC-01 as P0, disposition `planned`; MATRIX CS-07 independently confirms it as "the
top-ranked security risk" with a dedicated closure spec. DATA-07 carries a documented hard
dependency on SEC-01 (PLAN §5.3: "Hard dependency on PF-SEC's SEC-01 — do not schedule before it
lands") and is therefore grouped into this same wave as a dependent epic, not deferred to a later
one, per mandate §2.2's "foundational contracts before adapters" sequencing principle.

## Framework capabilities delivered

- A server-side `identity_grant` table (RLS FORCE, one-active-grant-per-actor partial index,
  `app_platform`-only grants) that becomes the sole source of truth for break-glass and
  impersonation state — PLAN SEC-01 T1.
- Unconditional tenant-membership verification in `Verifier.Actor` against the existing
  `user_tenant_access` table (today queried by zero Go code) — PLAN SEC-01 T2/T3.
- Server-side capacity selection and a privileged-session resolver that populates `Actor` fields
  only from a verified grant-table row, never trusted off the JWT — PLAN SEC-01 T4/T5.
- `auth_time`/`acr`/`amr`-bound step-up freshness and explicit user/API-key/webhook/internal
  credential-scheme distinction — PLAN SEC-01 T6/T7.
- A documented, coordinated wowsociety cutover plan for the impersonation-flow breaking change
  (PROD-04) — this wave's E01-S004.
- Outbound-security escape-hatch governance: `SharedFingerprint()` scope confirmation over the
  egress allowlist, a boot-time egress-exception report, an allowlist change-audit trail, and (D-07)
  a declared, fingerprinted trusted-issuer config gate for JWKS client injection in `prod` — PLAN
  SEC-06 T1–T5.
- Webhook replay/dedup controls bound exclusively to provider-authenticated data via a new
  `Verifier` interface returning `(Envelope, error)` — PLAN SEC-03 T1–T4 (breaking interface
  change).
- `Checker.Has` relationship evaluation extended to party-subject edges and every schema-enumerated
  `subject_kind`, with shared actor-attribution sourced from DATA-06 T2's mechanism — PLAN DATA-07
  T1/T2/T4.
- Workflow ratification implemented as a real definition field and state transition (or a
  documented interim reject posture) plus a durable, grant-ID-attributed override audit record —
  PLAN SEC-02 T4/T5, closing the P0 finding SEC-02 left open after its Wave-0 slice (T1–T3, already
  executed).

## Included epics

- **W03-E01 — server-side-session-state** (SEC-01): grant schema and membership enforcement
  (S001), capacity selection and the privileged-session resolver (S002), assurance and credential
  schemes (S003), and a documentation/coordination-only cross-repo cutover plan for wowsociety's
  breaking impersonation-flow change (S004).
- **W03-E02 — outbound-security-governance** (SEC-06): fingerprint-scope confirmation, egress
  reporting, allowlist change-audit, and the D-07 JWKS-client governance gate.
- **W03-E03 — webhook-authenticated-replay** (SEC-03): the breaking `Verifier` interface change and
  the rewired `HandleInbound` replay/dedup path.
- **W03-E04 — relationship-semantics** (DATA-07): party-subject evaluation, subject-kind matrix
  completion, and shared actor attribution — hard-dependent on E01's acceptance.
- **W03-E05 — workflow-privileged-completion** (SEC-02 remainder): ratification design/
  implementation (or documented interim reject posture) and durable audit with grant-ID
  attribution.

## Entry criteria

- W01's exit gate satisfied: the central-validation (`RouteMeta`) seam is live, so this wave's new
  grant-table endpoints (SEC-01 T1/T2) and webhook `Envelope`-consuming endpoints (SEC-03) are built
  against a stable contract-enforcement pattern rather than a moving target.
- W02's exit gate satisfied: the DATA-09 online expand/backfill/validate/contract protocol exists
  and is proven (its own CI drill pipeline, PLAN DATA-09 T9, has run at least once), so SEC-01 T1's
  "genuinely new" `identity_grant` migration is authored and rolled out through that protocol rather
  than a one-off unsafe migration.
- W00-E02-S003's ADR-ification story has ratified D-01 (`ADR-W00-E02-S003-001`) and D-07
  (`ADR-W00-E02-S003-007`) — this wave's E01-S001 and E02-S001 reference these ADRs as already-made
  design premises, not decisions this wave makes itself.

## Exit criteria

- `Verifier.Actor` unconditionally consults the grant table and `user_tenant_access`; a capacity-less
  actor is no longer trusted by default; zero/unknown tenant claims are rejected before a tenant
  transaction opens; `ImpersonatorUserID`/`BreakGlass` are populated only from a verified grant-table
  lookup by opaque grant ID.
- The required SEC-01 test classes (PLAN §6 SEC-05, mandatory: token substitution, zero-tenant,
  stale membership, revoked capacity, expired step-up, issuer/audience/key rotation, JWKS failure)
  all pass adversarially, where today (per MATRIX CS-07) they "currently *pass wrongly* or are
  untestable."
- SEC-06's D-07 gate is enforced: a `prod`-profile JWKS client injection without a declared
  trusted-issuer allowlist fails readiness.
- SEC-03's `Verifier` interface returns `Envelope`; `HandleInbound` sources replay-window/dedup
  exclusively from authenticated envelope fields, proven by the tamper matrix (body/timestamp/
  event-ID/key-ID/signature-version independently manipulated).
- DATA-07's `Checker.Has` evaluates party-subject edges and every schema-enumerated `subject_kind`;
  actor attribution on `Relate`/mirror `Upsert` reuses DATA-06 T2's mechanism without reimplementing
  it.
- SEC-02's ratification is either implemented as a real state transition or explicitly rejects
  `ratify_by`-declaring definitions with a documented interim posture; override audit rows persist
  actor, impersonator, grant ID, source/target states, and reason in the same transaction as the
  state jump.
- E01-S004's sequencing, staging-validation, and rollback coordination documents exist and are
  reviewed, satisfying PROD-04 without any wowapi or wowsociety product code change originating from
  this wave.

## Dependencies

Depends on W01 (central-validation seam — `impl/index.md` wave map) and W02 (DATA-09 protocol the
grant-table migration uses — `impl/index.md` wave map). Internally: W03-E04 (DATA-07) has a hard
dependency on W03-E01 being **accepted**, not merely started (PLAN §5.3, verbatim: "Hard dependency
on PF-SEC's SEC-01 — do not schedule before it lands"). W03-E04's T4 acceptance criterion also
carries a secondary, soft dependency on W05-E04-S002's authz-epoch work (SEC-04, D-06) for
cross-pod cache invalidation — PLAN's own cross-cutting note: "do not assume PF-SEC delivers on
PF-DATA's timeline," read at this wave's scope as "do not assume W05 delivers the epoch table on
W03's timeline." W03-E05's T5 (durable audit) depends on W03-E01-S001's grant-ID field. See
`dependencies.md` for the full register.

## Assumptions

- DEC-Q1 (the IdP `grant_id` claim-contract shape) remains an open, human-blocked decision at this
  wave's start. Per REVIEW §F row 1 and MATRIX CS-07: "Safe default per review §F Q1: framework
  owns the grant record keyed by grant-ID; IdP claim shape is tuning, not a blocker." This wave's
  E01-S001 proceeds against that safe default explicitly, not by resolving DEC-Q1 itself.
- D-01 (`ADR-W00-E02-S003-001`) and D-07 (`ADR-W00-E02-S003-007`) are ratified by W00-E02-S003
  before this wave's E01-S001 and E02-S001 begin detailed implementation; if not yet ratified when
  this wave is picked up, that is documented as a blocking gap rather than re-decided here.
- wowsociety's staging environment and `identity_impersonation_session` data are available for the
  two-repo coordinated cutover E01-S004 plans against — PLAN SEC-01's wowsociety-impact prose states
  this validation step is required ("validate T2 against wowsociety staging data before making it
  unconditional") but this wave does not assume a specific staging-environment timeline; E01-S004
  records what must be determined, per mandate §18.

## Risks

See `risks.md`. Headline risks: DEC-Q1 remaining unresolved past this wave's start, forcing a
later rework of E01-S001's grant-ID claim-shape assumption; the wowsociety two-repo coordinated
cutover for impersonation (PROD-04) being a real, security-sensitive breaking change that cannot be
completed unilaterally by this wave; DATA-07 T4's cross-work-package dependency on W05's SEC-04
epoch-table timeline.

## Quality gates

- SEC-01's adversarial test classes (mandate §13's "negative tests... security tests") are the
  fail-first evidence for E01, not merely "the happy path now uses the grant table."
- SEC-03's tamper matrix (mandate §13's "compatibility tests... security tests") is the fail-first
  evidence for E03 — a manipulated `InboundIn.Timestamp`/`ExternalEventID` must be provably inert
  after the fix, proven inert-with-manipulation before the fix.
- SEC-06's D-07 gate is proven with a negative fixture (a `prod`-profile boot with a custom JWKS
  client and no declared trusted-issuer allowlist fails readiness), not merely documented as
  intended behavior.
- Every P0 story in this wave (E01-S001/S002/S003, E05-S001) and every story carrying independent
  hard-dependency exposure (E04-S001) receives an explicit independent-review task per mandate §14,
  scoped to the adversarial test classes named above.

## Required artifacts

- `identity_grant` migration (up/down), RLS catalog extension, unique partial index definition.
- `PrincipalStore.ActiveTenantAccess` implementation and its call site in `Verifier.Actor`.
- Privileged-session resolver replacing the direct claim copy of `ImpersonatorUserID`/`BreakGlass`.
- `SharedFingerprint()` scope-confirmation regression test and its diff; boot-time egress-exception
  report; allowlist change-audit trail; JWKS trusted-issuer config-gate implementation.
- New `Verifier` interface (`Envelope` type) and both implementations (`HMACVerifier`,
  `FakeVerifier`); rewired `HandleInbound`.
- Extended `Checker.Has` subject-kind evaluation branches; shared actor-attribution call into
  DATA-06 T2's mechanism.
- Ratification definition field/state-transition implementation (or documented interim reject
  posture); durable override-audit schema/write path.
- E01-S004's sequencing plan, staging-validation plan, and rollback plan documents (coordination
  artifacts, not product code).

## Required evidence

- SEC-01 adversarial-negative test logs for every required test class (token substitution,
  zero-tenant, stale membership, revoked capacity, expired step-up, issuer/audience/key rotation,
  JWKS failure).
- SEC-06 fingerprint-diff test output; egress-report test output; allowlist-change-audit test
  output; JWKS-client-governance negative-fixture test output.
- SEC-03 tamper-matrix test output (5 independently manipulated fields, all provably inert
  post-fix).
- DATA-07 party-subject-eval test output (seeded party-subject edge, previously-false now true);
  subject-kind matrix test output.
- SEC-02 ratification happy-path/pending/rejection test output; override-audit fault-injection test
  output (audit write failure rolls back the override in the same transaction).
- E01-S004's coordination-plan review record (no executable test — this is a documentation-review
  evidence type per mandate §10's "review reports").

## Expected implementation outcome

A framework where tenant membership, break-glass, and impersonation state cannot be forged by
presenting a validly-signed but stale or manipulated JWT; where outbound HTTP escape hatches
(custom JWKS clients, allowlisted hosts) are governed, fingerprinted, and audited rather than
silent constructor parameters; where webhook replay/dedup decisions are immune to caller-supplied
timestamp/event-ID manipulation; where relationship checks correctly evaluate party-subject edges
instead of silently treating them as unconsulted; and where a workflow privileged-operation
override always produces a complete, durable, actor-and-grant-attributed audit row.

## Acceptance authority

Product-security lead (PLAN §5.2's stated accountable role for PF-SEC) for E01/E02/E03/E05;
data/reliability lead (PLAN §5.3's stated accountable role for PF-DATA) jointly with the
product-security lead for E04, given DATA-07's hard SEC-01 dependency.

## Closure conditions

All exit criteria satisfied; all five epics' `closure-report.md` accepted; `waves/index.md`'s W03
row updated to reflect `accepted` status; DEC-Q1 either resolved or explicitly re-confirmed as
non-blocking for this wave's closure (per its own safe-default framing); PROD-04's coordination
artifacts (E01-S004) reviewed and accepted by both a wowapi and a wowsociety-side reviewer, even
though no wowsociety code change is made by this wave itself.

## Status update (2026-07-16)

`status: in-progress` — 7 of 8 in-scope stories independently reviewed and accepted per
`review-gate-2026-07-16.md`; E02/E03/E04/E05 all `accepted`. E01 remains `in-progress`: S001/S002
accepted, S003 verified-pending-human (product-security-lead sign-off, DEF-07), S004 implemented
(cross-repo wowsociety sign-off unverifiable in-repo, acceptance deferred). Wave cannot reach
`accepted` until E01's S003/S004 resolve.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
