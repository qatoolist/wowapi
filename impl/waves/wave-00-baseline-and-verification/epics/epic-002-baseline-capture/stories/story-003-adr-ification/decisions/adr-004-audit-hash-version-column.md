---
id: ADR-W00-E02-S003-004
type: decision
title: Audit hash_version smallint column, version-branched verification
status: ratified
context: DATA-08 W6 audit-hash discriminator design — how to widen chainHash's field coverage without breaking historical-row verification?
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
related_source_items:
  - D-04
  - W04-E04
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-004 — Audit hash_version smallint column, version-branched verification

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-004.

## Title

Audit hash_version smallint column, version-branched verification.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §F row 5 (Q5); this ADR
file's own creation/registration is tracked separately by task W00-E02-S003-T002's own `status:
todo`→`done` lifecycle (see `../story.md` "Status discipline").

## Context

REVIEW §F question 5 asks: DATA-08's Wave-6 work must widen `chainHash`'s field coverage (to
include metadata and tx_id) for compliance-evidence completeness — but existing historical audit
rows were hashed under the narrower, pre-widening field set. What discriminator design lets
verification correctly distinguish and validate both old and new rows without breaking historical
verification?

## Options considered

REVIEW §F row 5 classifies this as "answerable by technical analysis" rather than framing it as a
choice between named competing options; no explicit rejected alternative is stated in the source.
This ADR does not invent one. The chosen design (below) is presented in REVIEW as the resolution of
the technical-analysis question, not as a selection among enumerated alternatives.

## Decision

**Add a `hash_version smallint NOT NULL DEFAULT 1` column in the same migration that widens
`chainHash`'s field coverage; verification branches by version — historical rows verify under v1,
new rows under v2 (metadata + tx_id included).** (REVIEW §F row 5 — "Add a `hash_version smallint
NOT NULL DEFAULT 1` column in the same migration; verification branches on it. Historical rows
verify under v1; new rows under v2 (metadata + tx_id included)." — combined with MATRIX CS-20's
phrasing of the same decision, "add `hash_version smallint NOT NULL DEFAULT 1` in the same
migration; ... verification branches by version"; not a single verbatim quote.)

### Safe default

No distinct safe-default stated beyond the decision itself — REVIEW §F row 5 states an
unconditional resolution ("resolved"), not a recommendation with a separate fallback path.

## Rationale

REVIEW §F row 5, quoted verbatim: "Standard append-only-log versioning." Adding a version
discriminator column, defaulted to the old scheme's version number for all pre-existing rows,
lets verification code branch on that column rather than requiring a rewrite/re-hash of every
historical row (which, for a compliance audit log, would be both expensive and itself a form of
retroactive tampering with the historical record). New rows, created after the migration, get the
new (v2) version number and are verified under the widened field set (including metadata and
tx_id) from the start.

## Consequences

- DATA-08's Wave-6 migration (W04-E04, "hash_version branch verification") adds this column and
  the version-branching verification logic in the same migration.
- Historical audit rows remain verifiable exactly as they were hashed, under v1 semantics — no
  retroactive re-hash, no loss of the original evidentiary chain.
- New rows carry stronger evidence completeness (metadata + tx_id included in the hash) going
  forward, closing the compliance gap DATA-08 identified.
- `PROD-05` (wowsociety compliance-drill re-verification before version bump, per
  `requirement-inventory.md` §D) is the product-level consequence: wowsociety must re-verify its
  own staging audit trail against this version-branch logic before adopting the version bump in
  production — tracked as a product-level coordination item, not this ADR's own implementation
  concern.

## Related source items

D-04; downstream epic W04-E04 (DATA-08 Wave-6 hash-version work) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies."

## Date

2026-07-12.

## Deciders

Fable 5.
