---
id: ADR-W00-E02-S003-003
type: decision
title: Post-seal mutation errors in production, panics only under an explicit dev/test build tag
status: ratified
context: AR-01/AR-04 post-seal mutation handling — should a mutation attempt after seal error or panic in production builds?
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
related_source_items:
  - D-03
  - W05-E01
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-003 — Post-seal mutation errors in production, panics only under an explicit dev/test build tag

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-003.

## Title

Post-seal mutation errors in production, panics only under an explicit dev/test build tag.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §F row 4 (Q4); this ADR
file's own creation/registration is tracked separately by task W00-E02-S003-T001's own `status:
todo`→`done` lifecycle (see `../story.md` "Status discipline").

## Context

REVIEW §F question 4 asks: when code attempts to mutate an already-sealed application-model
component (AR-01/AR-04's "post-seal mutation" case), should that attempt return an error, or should
it panic — and should that behavior differ between production and non-production builds? This is a
concurrency/lifecycle decision that AR-01 T8 and AR-04 T4 implement directly.

## Options considered

- **Unconditional panic on post-seal mutation, in all builds including production** — rejected.
  REVIEW §F row 4 frames the question as "error vs panic in prod" and resolves it against panic in
  production specifically, stating "A framework must not convert a benign retained-handle into a
  prod crash." (Wave-00-added clarification, cross-sourced: the identification of wowsociety's
  `s.rulesReg` retention as the concrete benign retained-handle case comes from
  `premier-framework-implementation-plan.md` §7 item 4 and MATRIX CS-06, not from REVIEW §F row 4
  itself.)
- **Error in production builds; panic only under an explicit `dev`/test build tag** — chosen. See
  Decision below.

## Decision

**Error in production builds; panic only under an explicit `dev`/test build tag.** (REVIEW §F row
4, quoted verbatim.)

### Safe default

No distinct safe-default stated beyond the decision itself — REVIEW §F row 4 states an
unconditional resolution ("resolved"), not a recommendation with a separate fallback path.

## Rationale

REVIEW §F row 4, quoted verbatim: "A framework must not convert a benign retained-handle into a
prod crash." wowsociety's harmless `s.rulesReg` retention is the concrete case motivating this (per
`premier-framework-implementation-plan.md` §7 item 4 — "panic-in-prod would convert wowsociety's
currently-harmless `s.rulesReg` retention into a crash risk" — a Wave-00-added cross-source
attribution, since REVIEW §F row 4 states the principle without naming the case): a product component that retains a handle to a registrar past its seal point, without
actually attempting a disallowed mutation through it, is a common and benign pattern — panicking
unconditionally on any post-seal interaction (rather than only on an actual disallowed mutation
attempt) would turn ordinary, safe code into a production outage. Development and test builds can
still panic (via the explicit build tag) so that developers get a loud, immediate signal during
development, without that same loudness becoming a production liability.

## Consequences

- AR-01 T8 and AR-04 T4 implement error-returning behavior for post-seal mutation attempts in
  default (production) builds.
- A `dev`/test build tag path exists (implemented as part of the same tasks) where the same
  condition panics instead, for fast developer feedback.
- wowsociety's existing `s.rulesReg` retention pattern remains safe under this design — no product-
  level remediation is required as a consequence of this decision (contrast with D-01, which does
  require product-level remediation).

## Related source items

D-03; downstream epic W05-E01 (AR-01) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies." AR-04 T4 (W05-E03-S002, per
`requirement-inventory.md`'s AR-04 row) also implements this decision.

## Date

2026-07-12.

## Deciders

Fable 5.
