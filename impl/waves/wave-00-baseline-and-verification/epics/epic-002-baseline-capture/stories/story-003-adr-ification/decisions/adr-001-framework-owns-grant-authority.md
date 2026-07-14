---
id: ADR-W00-E02-S003-001
type: decision
title: Framework owns grant validity/expiry/revocation authority
status: ratified
context: wowsociety's identity_impersonation_session table vs the framework's grant-authority table — which owns validity/expiry/revocation?
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
  - product/security-lead (D-01 tuning — IdP claim shape only)
related_source_items:
  - D-01
  - W03-E01
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-001 — Framework owns grant validity/expiry/revocation authority

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-001.

## Title

Framework owns grant validity/expiry/revocation authority.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §F row 2 (Q2); this ADR
file's own creation/registration is tracked separately by task W00-E02-S003-T001's own `status:
todo`→`done` lifecycle (see `../story.md` "Status discipline").

## Context

REVIEW §F question 2 asks: which system has authority over grant validity/expiry/revocation —
wowsociety's `identity_impersonation_session` table, or a server-side framework grant table? This
is the framework-boundary question WOW-Review §1 raised: which dependency direction is correct.
Resolving it is a precondition for SEC-01 (server-side tenant/privileged session state, W03-E01) to
be designed in detail — SEC-01's design assumes a specific answer to this question.

## Options considered

- **wowsociety owns grant authority** (via its existing `identity_impersonation_session` table) —
  rejected. REVIEW §F row 2 states this is the wrong dependency direction ("This is the correct
  dependency direction (WOW-Review §1)" — stated of the chosen option, implying the alternative,
  product-owns-authority, is the incorrect direction being rejected).
- **Framework owns grant authority; wowsociety's table becomes product UX/audit-only** — chosen.
  See Decision below.

REVIEW does not name a third alternative (e.g. a shared/joint-ownership model); none is invented
here.

## Decision

**Framework owns grant validity/expiry/revocation; wowsociety keeps its
[`identity_impersonation_session`] table for product UX/audit only.** (REVIEW §F row 2, quoted
verbatim; the bracketed table name is an editorial insertion — it appears in the same row's
question column, not in the decision cell itself.)

### Safe default

Adjacent safe default, stated in REVIEW §F row 1 (Q1, the related but distinct genuine-human
decision `DEC-Q1` — out of scope for this ADR but sharing the same framework-owns-the-grant-record
premise): "build the server-side `identity_grant` table + resolver now, keyed on grant-ID, and have
`Verifier.Actor` consult it. If the IdP cannot emit `grant_id`, the framework still owns the grant
record and looks it up by session — the JWT only carries a stable subject." This is the practical
default that makes D-01's decision buildable immediately, independent of how `DEC-Q1`'s IdP-claim
question is eventually resolved. It is recorded here as an adjacent fact — a Wave-00-added clarification, not D-01's own content —
since D-01 itself (REVIEW §F row 2) is stated as an unconditional resolution, not a
recommendation-with-fallback.

## Rationale

REVIEW §F row 2 states this is "the correct dependency direction (WOW-Review §1)" — i.e., the
framework, as the platform kernel, must be the authority for session/grant validity so that any
product built on the framework (not only wowsociety) can rely on a single, framework-owned
authority rather than each product re-implementing its own grant-validity logic. wowsociety
retaining its own table for product UX/audit purposes is compatible with this: it is a read-side/
audit-trail concern, not an authority concern.

## Consequences

- SEC-01 (W03-E01, server-side tenant/privileged session state) can proceed with a framework-owned
  grant table as a fixed design premise.
- wowsociety's `identity_impersonation_session` table is retained but its role changes to product
  UX/audit only — this is a **breaking change for wowsociety** (flagged in
  `impl/analysis/requirement-inventory.md` SEC-01 row: "BREAKING wowsociety"), tracked as product-
  level coordination item `PROD-04` ("SEC-01 impersonation cutover") per that inventory's §D.
  Implementing this cutover is SEC-01's job, not this ADR's.
- D-01's tuning (exact IdP claim shape) remains open as `DEC-Q1`, a genuine human
  (product/security-lead) decision — but per the safe default above, that openness does not block
  building the server-side grant table now.

## Related source items

D-01; downstream epic W03-E01 (SEC-01) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies."

## Date

2026-07-12.

## Deciders

Fable 5 (framework architecture lead role); D-01 tuning (IdP claim shape) = product/security-lead
(REVIEW §U: "owner = Fable 5 (framework) except D-01 tuning = product/security-lead").
