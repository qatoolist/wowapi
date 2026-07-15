---
id: W03-E02-ACCEPTANCE
type: epic-acceptance
epic: W03-E02
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E02 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" as a standalone,
independently-referenceable record, consistent with the wave-level `../../acceptance.md` pattern
(AC-W03-05 there maps onto this epic).

## AC-W03-E02-01 — Fingerprint scope confirmed

`SharedFingerprint()`'s scope is confirmed (or extended, if the confirmation test finds a gap) to
cover the outbound allowlist (`AllowedHosts`/`AllowedCIDRs`); a fingerprint-diff regression test
passes.

## AC-W03-E02-02 — Boot-time egress report

A boot-time (readiness or startup log) report enumerates every enabled egress exception; no
credential value is exposed in the report output, proven by a test asserting the report's field set
excludes any credential-shaped field.

## AC-W03-E02-03 — Allowlist change-audit trail

A configuration change touching the outbound allowlist produces an audit-visible record, proven by
a test that changes the allowlist config and asserts the audit record's presence/content.

## AC-W03-E02-04 — JWKS-client governance gate (D-07)

A `prod`-profile boot with a custom JWKS `*http.Client` injected and no declared trusted-issuer
allowlist config fails readiness, proven by a negative fixture test.

## AC-W03-E02-05 — Never-tenant-controlled fitness check

A static/fitness check asserts that allowlist and JWKS-client construction code paths never read
request-scoped or tenant-scoped data — codifying an already-true invariant per PLAN's own framing.

## AC-W03-E02-06 — Independent review passed

W03-E02-S001 has passed independent review per mandate §14, with specific confirmation that D-07's
JWKS-client governance gate exactly matches the ADR's stated design (declared, fingerprinted config
field; reject in `prod` without it) and does not silently narrow or widen the gate's scope.

## Acceptance authority

Product-security lead (PLAN §5.2's stated accountable role for PF-SEC).
