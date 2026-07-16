---
id: W03-E02-S001
type: story
title: Outbound-security escape-hatch governance
status: accepted
wave: W03
epic: W03-E02
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - SEC-06
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W03-E02-S001-01
  - AC-W03-E02-S001-02
  - AC-W03-E02-S001-03
  - AC-W03-E02-S001-04
  - AC-W03-E02-S001-05
artifacts: []
evidence: []
decisions:
  - ADR-W00-E02-S003-007
risks: []
---

# W03-E02-S001 — Outbound-security escape-hatch governance

## Story ID

W03-E02-S001

## Title

Outbound-security escape-hatch governance

## Objective

Confirm/extend `SharedFingerprint()`'s scope to cover the outbound allowlist; add a boot-time report
enumerating enabled egress exceptions; add an explicit change-audit trail for allowlist
configuration changes; extend equivalent governance to the JWKS `Client` injection path per D-07 (a
`prod`-profile custom JWKS client must declare a trusted-issuer config or fail readiness); and
codify "never tenant/user-controlled data populates allowlists/JWKS clients" as a static fitness
check. This is PLAN SEC-06 T1 through T5 in full.

## Value to the framework

The framework has two deliberate, legitimate outbound-security escape hatches — a caller-injectable
JWKS `*http.Client` and exact-match allowlisted egress hosts — but today neither is governed: their
scope is not confirmed as fingerprinted, their presence is not visible at boot, changes to them are
not audited, and the JWKS injection path in particular has "zero config surface" and "cannot be
fingerprinted/audited today" (PLAN's own words). This story makes both escape hatches observable,
change-tracked, and — for the JWKS path — gated behind a declared configuration field rather than a
bare constructor parameter, closing the specific governance gap without removing the legitimate
capability itself.

## Problem statement

PLAN §5.2 SEC-06's evidence, cited verbatim: "`JWKSConfig.Client *http.Client`
(`kernel/auth/jwks.go:59`) is caller-injectable and bypasses the default client's proxy-disabling;
an injected client gets no private-IP dial guard, by design and self-documented.
`httpclient/client.go:142` — an exact-match allowlisted hostname skips IP-class checking entirely."
On configuration provenance: "`AllowedHosts`/`AllowedCIDRs` come from static deployment config,
boot-validated — not tenant/user-controlled. `SharedFingerprint()` likely already covers these
fields structurally, pending a direct scope-confirmation test." T4's evidence specifically: the
JWKS client injection path is "currently pure Go constructor param, zero config surface, cannot be
fingerprinted/audited today."

## Source requirements

SEC-06 (T1, T2, T3, T4, T5). Cross-referenced: D-07 (`ADR-W00-E02-S003-007`, referenced not
authored — see "Decisions" below).

## Current-state assessment

Per PLAN §5.2's own evidence citation (to be re-confirmed at this story's own execution commit):

- `JWKSConfig.Client *http.Client` at `kernel/auth/jwks.go:59` is a caller-injectable field that
  bypasses the default client's proxy-disabling; an injected client receives no private-IP dial
  guard, "by design and self-documented" (i.e. this is an intentional, documented escape hatch, not
  an accidental gap).
- `httpclient/client.go:142` implements an exact-match allowlisted-hostname exception that skips
  IP-class checking entirely for that host.
- `AllowedHosts`/`AllowedCIDRs` are static deployment configuration, boot-validated — not
  tenant/user-controlled — which materially lowers this finding's risk triage relative to a
  request-time-controlled attack surface.
- `SharedFingerprint()`'s exact scope relative to these fields is, per PLAN's own words, "likely
  already covers these fields structurally, pending a direct scope-confirmation test" — i.e. this is
  an unconfirmed assumption, not yet a verified fact.
- The JWKS `Client` injection path has zero configuration surface today: it is a plain Go
  constructor parameter, with no way to fingerprint, audit, or gate it.

## Desired state

`SharedFingerprint()`'s scope is confirmed (or extended) to cover the outbound allowlist. A
boot-time report enumerates every enabled egress exception with no credentials exposed. An allowlist
configuration change produces an audit-visible record. A `prod`-profile boot with a custom JWKS
client injected and no declared trusted-issuer allowlist fails readiness (D-07's enactment). A
static fitness check asserts that allowlist/JWKS-client construction never reads request- or
tenant-scoped data.

## Scope

- T1 — confirm/extend `SharedFingerprint()` scope to cover the outbound allowlist; regression test.
- T2 — boot-time startup report enumerating enabled egress exceptions, no credentials exposed.
- T3 — explicit change-audit trail for allowlist configuration changes.
- T4 — extend equivalent governance to the JWKS `Client` injection path per D-07: a `prod`-profile
  custom JWKS client must declare a trusted-issuer config or fail readiness.
- T5 — codify "never tenant/user-controlled data populates allowlists/JWKS clients" as a lint/
  fitness check.

## Out of scope

- SEC-01's grant-table/resolver work (W03-E01) — a separate identity-trust concern.
- SEC-03's webhook `Verifier` interface (W03-E03) — a separate, distinct finding.
- Changing the default `http.Client`'s own proxy-disabling or private-IP dial-guard behavior — this
  story governs the escape hatches around that default, verified sufficiently strong by MATRIX
  CS-24, not the default itself.
- A full wowsociety deployment-config audit for its own allowlist/JWKS-client usage — PLAN's own
  words: "wowsociety's actual deployment config for allowlist entries or custom JWKS-client
  injection was not read in this pass — needs a follow-up config audit." This story does not invent
  or assume that audit's outcome; see "Assumptions."

## Assumptions

- **`SharedFingerprint()`'s coverage is unconfirmed, not assumed correct.** PLAN's own words: "likely
  already covers these fields structurally, pending a direct scope-confirmation test." T1's own work
  is exactly that confirmation — this story does not assume the fingerprint already covers the
  allowlist; it proves it (or extends it if the proof fails).
- **wowsociety's deployment-config evidence gap is honestly recorded, not papered over.** Per PLAN's
  own admission, quoted above, wowsociety's actual allowlist/JWKS-client deployment configuration was
  not read as part of PLAN's own evidence-gathering pass. This story does not invent what that
  configuration looks like; T4's breaking-change risk to wowsociety is recorded as "unconfirmed" in
  "Compatibility considerations" below, not asserted either way.
- D-07 (`ADR-W00-E02-S003-007`) is ratified before this story's T4 implementation begins; if not yet
  ratified, that is a blocking gap to be documented, not worked around.

## Dependencies

Depends on W00-E02-S003's ADR-ification of D-07 (`ADR-W00-E02-S003-007`). No dependency on any other
W03 epic's story — SEC-06 is architecturally independent of SEC-01/SEC-03/DATA-07/SEC-02.

## Affected packages or components

`kernel/auth/jwks.go` (`JWKSConfig.Client`); `httpclient/client.go` (the allowlist); the config layer
(`SharedFingerprint()`'s scope, the new trusted-issuer config field for T4); the readiness/boot-
reporting layer (T2's egress-exception report); a new lint/fitness-check mechanism (T5).

## Compatibility considerations

T1, T2, T3, and T5 are additive — a regression test, a boot-time report, an audit trail, and a
static fitness check do not change existing runtime behavior for any current consumer. **T4 is
"Breaking only for T4, only if wowsociety currently injects a custom JWKS client with no declaration
path (unconfirmed)"** — PLAN's own words, cited verbatim. This is explicitly recorded as unconfirmed,
not assumed safe: wowsociety "never constructs `httpclient.New`/`auth.JWKSConfig` directly (wired by
wowapi's `kernel.New`), and configures OIDC/JWKS purely via static YAML (`configs/stage.yaml:59,63`)"
per PLAN's own evidence — which suggests low risk, but PLAN itself flags this as a "genuine evidence
gap... not papered over," needing "a follow-up config audit." This story does not resolve that
follow-up audit; it records the gap honestly and proceeds with T4's governance gate regardless, since
the gate itself (declare-or-fail-readiness) is the correct behavior independent of whether
wowsociety happens to trigger it today.

## Security considerations

This story is itself a security-governance story: it does not change the framework's outbound
attack surface, it makes the framework's existing, deliberate escape hatches from that surface
visible, fingerprinted, audited, and (for JWKS) gated. T5's fitness check is a durable, mechanically
enforced invariant preventing a future regression where request- or tenant-scoped data leaks into
allowlist/JWKS-client construction.

## Performance considerations

None material. Boot-time reporting (T2) and configuration-change auditing (T3) are not
request-path-latency-sensitive operations.

## Observability considerations

T2's boot-time egress-exception report and T3's change-audit trail are themselves observability
deliverables — this story adds visibility where none existed, rather than consuming existing
observability infrastructure.

## Migration considerations

None. No schema change unless T4's trusted-issuer config field requires persistence beyond
in-process configuration (to be confirmed at implementation time — likely a static config field, not
a database table).

## Documentation requirements

Document the confirmed/extended `SharedFingerprint()` scope, the egress-exception report's format,
the allowlist change-audit record's format, the trusted-issuer config field's declaration syntax and
its `prod`-profile enforcement behavior, and the fitness check's exact assertion, in whatever
documentation currently covers the outbound-HTTP/JWKS configuration contract.

## Acceptance criteria

- **AC-W03-E02-S001-01**: A fingerprint-diff regression test proves `SharedFingerprint()`'s scope
  covers the outbound allowlist.
- **AC-W03-E02-S001-02**: A boot-time report enumerates every enabled egress exception with no
  credentials exposed in the output.
- **AC-W03-E02-S001-03**: An allowlist configuration change produces an audit-visible record, proven
  by a test.
- **AC-W03-E02-S001-04**: A `prod`-profile boot with a custom JWKS client injected and no declared
  trusted-issuer allowlist fails readiness, proven by a negative fixture.
- **AC-W03-E02-S001-05**: A static fitness check asserts that allowlist/JWKS-client construction
  never reads request- or tenant-scoped data.

## Required artifacts

- `SharedFingerprint()` scope confirmation/extension and its regression test.
- Boot-time egress-exception report implementation.
- Allowlist change-audit trail implementation.
- JWKS trusted-issuer config-gate implementation (D-07 enactment).
- The no-tenant-controlled-allowlist fitness check.
See `artifacts/index.md`.

## Required evidence

- Fingerprint-diff regression test output (AC-01).
- Egress-exception report output sample, confirmed credential-free (AC-02).
- Allowlist-change-audit test output (AC-03).
- JWKS-client-governance negative-fixture test output (AC-04).
- Fitness-check test output (AC-05).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`:
`story.md` and `plan.md` complete, acceptance criteria numbered and measurable, D-07 dependency
recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all five acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14.

## Risks

None registered at wave/epic scope specifically for this story. T4's own PLAN risk column ("open
design decision, not yet made") is resolved by D-07's ratification, not a residual risk for this
story to carry.

## Residual-risk expectations

Low residual risk expected once all five acceptance criteria are met. The wowsociety
deployment-config evidence gap (see "Assumptions") remains an open follow-up item beyond this
story's own boundary, tracked honestly rather than silently closed.

## Plan

See `plan.md`.
