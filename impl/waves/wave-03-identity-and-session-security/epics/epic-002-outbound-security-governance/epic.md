---
id: W03-E02
type: epic
title: Outbound-security governance
status: accepted
wave: W03
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - SEC-06
  - D-07
depends_on: []
stories:
  - W03-E02-S001
decisions:
  - ADR-W00-E02-S003-007
risks: []
---

# W03-E02 — Outbound-security governance

## Epic objective

Govern the framework's explicit outbound-security escape hatches — the caller-injectable JWKS
`*http.Client` and the exact-match allowlisted-hostname egress bypass — with fingerprint-scope
confirmation, boot-time reporting, change-audit, and (per D-07) a declared, fingerprinted
trusted-issuer config gate that rejects an ungoverned custom JWKS client in `prod`. This is PLAN
§5.2's SEC-06, T1 through T5 in full (single-story epic per `impl/analysis/wave-allocation-detail.md`).

## Problem being solved

`requirement-inventory.md` row SEC-06 records: "Outbound-security escape-hatch governance" — class
IMPL, priority P1, disposition `planned`, target `W03-E02-S001`, notes "D-07 ratified." PLAN §5.2's
own evidence: "`JWKSConfig.Client *http.Client` (`kernel/auth/jwks.go:59`) is caller-injectable and
bypasses the default client's proxy-disabling; an injected client gets no private-IP dial guard, by
design and self-documented. `httpclient/client.go:142` — an exact-match allowlisted hostname skips
IP-class checking entirely." The configuration provenance is favorable (PLAN notes: "come from
static deployment config, boot-validated — not tenant/user-controlled"), but the escape hatches
themselves are today ungoverned: no fingerprinting confirmation for the allowlist, no boot-time
visibility into what egress exceptions are active, no change-audit trail, and — for the JWKS client
injection path specifically — "currently pure Go constructor param, zero config surface, cannot be
fingerprinted/audited today" (T4's own evidence).

## Scope

- T1 — confirm/extend `SharedFingerprint()` scope to cover the outbound allowlist; regression test.
- T2 — boot-time startup report enumerating enabled egress exceptions, with no credentials exposed.
- T3 — explicit change-audit trail for allowlist configuration changes.
- T4 — extend equivalent governance to the JWKS `Client` injection path: a `prod`-profile custom
  JWKS client must either declare a trusted-issuer config or fail readiness (D-07's enactment).
- T5 — codify "never tenant/user-controlled data populates allowlists/JWKS clients" as a lint/
  fitness check.

## Out of scope

- SEC-01's grant-table/resolver work (W03-E01) — a separate epic with a separate identity-trust
  concern.
- SEC-03's webhook `Verifier` interface (W03-E03) — a separate outbound-adjacent but distinct
  finding.
- Any change to the default `http.Client`'s own proxy-disabling or private-IP dial-guard behavior —
  this epic governs the *escape hatches* around that default behavior, it does not change the
  default itself, which MATRIX CS-24 already verifies as sufficiently strong ("Verified strength;
  gosec G704 annotation task inside FBL-07" — that annotation task belongs to W01-E01-S002, not
  this epic).

## Source requirements

SEC-06 (T1–T5). Cross-referenced: D-07 (`ADR-W00-E02-S003-007` — JWKS trusted-issuer config gate).

## Architectural context

This epic operationalizes a specific security principle: an explicit, deliberate escape hatch
(caller-injected JWKS client, allowlisted egress host) is not inherently wrong — the codebase has
good reasons for both — but it must be *governed*: visible at boot, fingerprinted for drift
detection, audited on change, and (for the currently-ungoverned JWKS injection path specifically)
gated behind a declared config field rather than a bare constructor parameter. PLAN's own framing:
"Configuration provenance (changes risk triage): `AllowedHosts`/`AllowedCIDRs` come from static
deployment config, boot-validated — not tenant/user-controlled." This lowers the urgency relative to
a tenant-controlled attack surface, but does not eliminate the governance gap this epic closes.

The affected layers are `kernel/auth/jwks.go` (`JWKSConfig.Client`), `httpclient/client.go` (the
allowlist), the config layer (`SharedFingerprint()`'s scope, the new trusted-issuer config field),
and the readiness/boot-reporting layer (the new egress-exception report).

## Included stories

- **W03-E02-S001 — outbound-security-governance** (SEC-06 T1–T5, single story per
  `impl/analysis/wave-allocation-detail.md`: "S001 all tasks (D-07 enacted)").

## Dependencies

Depends on W00-E02-S003's ADR-ification of D-07 (`ADR-W00-E02-S003-007`). No dependency on W03-E01,
W03-E03, W03-E04, or W03-E05 — SEC-06 is architecturally independent of the other four findings in
this wave (no shared file surface, no design-premise dependency beyond D-07, which is itself
upstream of this wave, not intra-wave).

## Risks

No epic-specific risk beyond D-07's own "Highest-risk task" framing for T4 (PLAN: "open design
decision, not yet made" — now resolved by D-07 itself, so this epic enacts rather than makes that
decision). See `risks.md` for the epic-scoped elaboration.

## Required decisions

None new. This epic enacts already-ratified `ADR-W00-E02-S003-007` (D-07: require trusted-issuer/
egress config to be a declared, fingerprinted config field; reject a custom JWKS `*http.Client` in
`prod` profile unless the trusted-issuer allowlist is set), see W00-E02-S003. This epic's single
story accordingly carries a `decisions/` directory referencing (not authoring) this ADR.

## Epic acceptance criteria

- **AC-W03-E02-01**: `SharedFingerprint()`'s scope is confirmed (or extended) to cover the outbound
  allowlist, proven by a fingerprint-diff regression test.
- **AC-W03-E02-02**: A boot-time report enumerates every enabled egress exception with no
  credentials exposed in the report output.
- **AC-W03-E02-03**: An allowlist configuration change produces an audit-visible record.
- **AC-W03-E02-04**: A `prod`-profile boot with a custom JWKS client injected and no declared
  trusted-issuer allowlist fails readiness (D-07's enactment).
- **AC-W03-E02-05**: A static fitness check asserts that allowlist/JWKS-client construction never
  reads request- or tenant-scoped data.
- **AC-W03-E02-06**: The story has passed independent review per mandate §14.

## Closure conditions

W03-E02-S001 reaches `accepted`; AC-W03-E02-01 through AC-W03-E02-06 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date.

## Status update (2026-07-16)

`status: accepted` — W03-E02-S001 accepted; genuine independent review superseded the prior
self-review.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
