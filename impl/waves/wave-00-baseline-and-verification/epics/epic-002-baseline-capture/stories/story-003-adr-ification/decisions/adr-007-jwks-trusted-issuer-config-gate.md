---
id: ADR-W00-E02-S003-007
type: decision
title: Trusted-issuer/egress config as a declared fingerprinted field; custom JWKS client gated in prod
status: ratified
context: SEC-06 JWKS-client governance model — what governs a custom JWKS *http.Client escape hatch, especially in production?
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
related_source_items:
  - D-07
  - W03-E02
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-007 — Trusted-issuer/egress config as a declared fingerprinted field; custom JWKS client gated in prod

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-007.

## Title

Trusted-issuer/egress config as a declared fingerprinted field; custom JWKS client gated in prod.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §F row 8 (Q8); this ADR
file's own creation/registration is tracked separately by task W00-E02-S003-T002's own `status:
todo`→`done` lifecycle (see `../story.md` "Status discipline").

## Context

REVIEW §F question 8 asks: SEC-06 allows an outbound-security "escape hatch" — a custom JWKS
`*http.Client` a caller can supply to override the framework's default JWKS-fetching behavior. What
governance model prevents this escape hatch from becoming an ungoverned egress/trust-boundary risk,
particularly in production?

## Options considered

- **Permit a custom JWKS `*http.Client` in `prod` profile without any additional gate** — rejected
  (implicit in the decision's own framing as a governance requirement being imposed): an ungoverned
  custom client would let any caller silently redirect JWKS trust decisions to an arbitrary,
  undeclared endpoint/egress path in production.
- **Require trusted-issuer/egress config to be a declared, fingerprinted `config` field; reject a
  custom JWKS `*http.Client` in `prod` profile unless the trusted-issuer allowlist is set** —
  chosen. See Decision below.

## Decision

**Require trusted-issuer/egress config to be a declared, fingerprinted `config` field; reject a
custom JWKS `*http.Client` in `prod` profile unless the trusted-issuer allowlist is set.** (REVIEW
§F row 8, quoted verbatim.)

### Safe default

No distinct safe-default stated beyond the decision itself — REVIEW §F row 8 states an
unconditional resolution ("resolved"), not a recommendation with a separate fallback path.

## Rationale

REVIEW §F row 8 classifies this as "Fable 5 decision (security)." Making the trusted-issuer/egress
configuration a declared, fingerprinted field (rather than an implicit or undeclared property of
whatever `*http.Client` a caller happens to supply) gives the framework a single, auditable point
where production trust boundaries are established. Rejecting a custom client in `prod` profile
unless that allowlist is explicitly set closes the gap where a caller could otherwise silently
introduce an unreviewed egress path into a production JWKS-fetching flow — the escape hatch remains
available (SEC-06's original intent), but only inside a governed boundary in production.

## Consequences

- SEC-06 (W03-E02, per `requirement-inventory.md`'s SEC-06 row: "Outbound-security escape-hatch
  governance... D-07 ratified") implements the fingerprinted trusted-issuer config field and the
  prod-profile rejection gate.
- Non-production profiles are not stated by REVIEW as being subject to the same rejection gate —
  this ADR does not extend the gate to non-prod profiles beyond what REVIEW states, since REVIEW's
  decision text specifically scopes the rejection to "`prod` profile."
- Any caller wanting to use a custom JWKS client in production must first declare and fingerprint
  its trusted-issuer/egress configuration — this is a new configuration-surface requirement SEC-06
  introduces, not a runtime behavior change to existing default (non-custom-client) JWKS fetching.

## Related source items

D-07; downstream epic W03-E02 (SEC-06) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies."

## Date

2026-07-12.

## Deciders

Fable 5 (security decision).
