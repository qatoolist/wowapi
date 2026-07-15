---
id: W04-E04-S001
type: story
title: Audit hash-chain widening with hash_version discriminator
status: accepted
wave: W04
epic: W04-E04
owner: W04Compliance
reviewer: code-reviewer
priority: P0
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-08
depends_on:
  - W02-E01
blocks:
  - W04-E04-S002
acceptance_criteria:
  - AC-W04-E04-S001-01
  - AC-W04-E04-S001-02
  - AC-W04-E04-S001-03
artifacts: []
evidence: []
decisions:
  - D-04
risks:
  - RISK-W04-002
  - RISK-W04-E04-S001-001
---

# W04-E04-S001 — Audit hash-chain widening with hash_version discriminator

## Story ID

W04-E04-S001

## Title

Audit hash-chain widening with hash_version discriminator

## Objective

Widen `kernel/audit`'s `chainHash` to cover every persisted field — canonicalized `metadata`,
`tx_id`, all nullable fields, sequence, ID, timestamps, and previous hash — so that mutating any
declared field independently breaks tamper-evidence verification, and add a `hash_version smallint
NOT NULL DEFAULT 1` column in the same migration so verification branches by row version: historical
rows verify under v1, new rows verify under v2 (metadata + tx_id included), per D-04's already-
ratified design.

## Value to the framework

This is the single highest-risk task in the epic's entire scope. PLAN DATA-08's own risk column for
W6-T1 states directly: "**Single highest-risk task in PF-DATA's Wave-6 scope, and directly hits
wowsociety's live audit rows.**" Without this story, the framework's compliance-evidence chain has a
confirmed, real gap: `metadata` and `tx_id` can be tampered on an audit row without the chain
detecting it, which for compliance evidence "is close to none" of a guarantee (MATRIX CS-20). This
story closes that gap — but doing so is itself a breaking format change against a live-production
audit chain that wowsociety already writes real rows into today, which is exactly why the
`hash_version` discriminator (D-04) is not optional scope-creep but the mechanism that makes the fix
safe to ship at all.

## Problem statement

`kernel/audit/audit.go:130-179`'s `chainHash` covers 15 length-prefixed fields (prev_hash…reason) but
excludes `metadata` — documented at `:155-159` as excluded because of a jsonb round-trip
reformatting problem that makes the stored form unreproducible — and excludes `tx_id`, inserted via
`pg_current_xact_id()` at `:140` and never hashed at all. `Verify` (`:195-248`) recomputes with the
identical, incomplete field list; `Anchor`/`CheckAnchor` (`:253-311`) only guard against tail
truncation, not field-level tampering. No `hash_version` column exists anywhere in the schema today.
MATRIX CS-20 states the defect/consequence directly: "an attacker (or bug) can alter audit
`metadata`/`tx_id` on a row without breaking the chain — the tamper-evidence guarantee is partial,
which for compliance evidence is close to none." The documented jsonb-reformatting rationale for the
original exclusion is precisely why a naive fix (just add `metadata`/`tx_id` to the hash input) is
insufficient: it would reintroduce non-reproducibility unless the fix hashes a canonicalized
pre-serialization form, never the stored jsonb.

## Source requirements

DATA-08 (Wave-6 task W6-T1 only). D-04 (already ratified, enacted here, not authored here).

## Current-state assessment

Per PLAN DATA-08's own evidence (to be re-confirmed at this story's actual start commit): `chainHash`
covers 15 fields, excludes `metadata` and `tx_id`; no `hash_version` column exists; `Verify` recomputes
with the same incomplete field list; `Anchor`/`CheckAnchor` guards tail truncation only. This is a
confirmed, real, currently-live gap in a production compliance-evidence mechanism, not a partial or
disputed finding — MATRIX CS-20 states the exclusions are "deliberate and internally consistent —
which does not make them sufficient."

**DATA-08 Wave-0 tasks (W0-T1, W0-T2) — already executed and verified ×2 elsewhere in the programme,
NOT this story's scope, referenced here only as already-done context:**

- **W0-T1** (stop discarding the attachment outbox-write error; propagate so `Attach` fails and
  rolls back in the same transaction) — already executed. This story does not re-implement or
  re-verify W0-T1; it is cited only because `requirement-inventory.md`'s DATA-08 row notes "W0 slice
  EXECUTED (verified ×2)" as prior, separate work.
- **W0-T2** (remove the stale deferral comment in `notify/service.go`; implement legal-delivery
  audit write using migration `00011`'s already-granted `app_platform` INSERT permission on
  `events_outbox`) — already executed. Same status: prior, separate, already-verified work, not
  re-marked as this story's own.

Neither W0-T1 nor W0-T2 is a hard prerequisite for W6-T1 per PLAN's own dependency column ("W0-T1,
W0-T2 (sequenced first, not a hard blocker)") — they are sequenced first by convention, not gating.
This story's own re-confirmation step (per this programme's fail-first convention applied elsewhere)
is to read `kernel/audit/audit.go` at this story's actual start commit and confirm the field
exclusions, the `pg_current_xact_id()` call, and the absence of a `hash_version` column all still
hold before implementing the widening.

## Desired state

`chainHash` hashes a canonicalized pre-serialization form of every persisted field, including
`metadata` and `tx_id` — never the stored jsonb directly, which would reintroduce the original
non-reproducibility problem. A `hash_version smallint NOT NULL DEFAULT 1` column exists, added in the
same migration that widens field coverage. `Verify` branches by `hash_version`: rows with
`hash_version = 1` verify under the original 15-field scheme; rows with `hash_version = 2` (or
whatever version value the new scheme is assigned) verify under the widened scheme. A tamper test
mutating each declared field independently — `metadata`, `tx_id`, and every other field already in
scope — confirms every one now breaks verification.

## Scope

- Widening `chainHash`'s field coverage to include canonicalized `metadata` and `tx_id`, alongside
  the 15 fields already covered, plus confirming every other persisted field (all nullable fields,
  sequence, ID, timestamps, previous hash) is genuinely covered — not assumed covered because it was
  already in the pre-existing 15-field list.
- Adding the `hash_version smallint NOT NULL DEFAULT 1` column in the same migration as the widening,
  per D-04.
- Implementing version-branched verification in `Verify` (and any other verification-adjacent
  function, e.g. `CheckAnchor`, that recomputes or depends on the hash input set).
- Canonicalizing `metadata` into a reproducible pre-serialization form for hashing, distinct from the
  stored jsonb representation, so the hash input is deterministic across reads.
- The per-field tamper test: mutating each declared field independently and asserting every one fails
  verification.
- Shipping the migration through W02-E01's online-migration protocol (expand/backfill/validate/
  contract), not an ad hoc one-off migration, given the breaking-change risk against a live table.

## Out of scope

- **DATA-08 W0-T1 and W0-T2** — already executed elsewhere; not re-implemented or re-verified here
  (see "Current-state assessment").
- **DATA-08 W6-T2 through W6-T5** (external anchor verification, encrypted DSR export, central
  legal-hold wrapper, explicit per-class DSR status) — W04-E04-S002's scope. This story's widened,
  versioned hash chain is a prerequisite those tasks build on; it does not itself implement them.
- **DX-07's readiness/config diagnostics scope** — W04-E04-S003's scope, unrelated to this story.
- **PROD-05** (the wowsociety-side staging audit re-verification drill before `FRAMEWORK_VERSION` is
  bumped past this story's commit) — a product-level compliance drill (`requirement-inventory.md`
  §D), excluded from this story's framework-side closure per mandate §2.3. Recorded here as a noted,
  non-blocking coordination item, not implemented.
- **Choosing the exact `hash_version` integer value assigned to the new scheme** beyond "not 1" (v1
  is reserved for the historical scheme per D-04) — recorded as an implementation-time decision in
  `plan.md`, not invented here.

## Assumptions

- This story's migration is expected to ship through W02-E01's online-migration protocol
  (expand/backfill/validate/contract), per this epic's `dependencies.md` and the wave-level W02
  dependencies.md's own framing: "the audit hash-chain widening migration (a breaking format change
  touching wowsociety's live audit rows) is expected to ship via DATA-09's protocol, not ad hoc."
  This is a confirmed sequencing expectation from the source, not an invented one — `depends_on`
  below cites **W02-E01** (the epic as a whole), matching the grain at which both the wave-level and
  epic-level dependency records state the edge, not a narrower sub-story citation the source does not
  make.
- D-04's design (hash_version smallint, version-branched verification, canonicalized
  pre-serialization metadata hashing) is treated as already decided and not re-litigated by this
  story — this story's own contribution is implementation, evidenced by the tamper test, not
  re-deciding the discriminator design.
- The exact `hash_version` value for the new scheme (e.g. `2`) is not specified by D-04's own decision
  text beyond distinguishing it from `1`; this story's plan records the exact chosen value as an
  implementation-time decision.

## Dependencies

Depends on **W02-E01** (the DATA-09 online-migration protocol) — per this epic's `dependencies.md`
and the wave-level entry criteria in `../../../../wave.md`: "this is the one concrete predecessor
capability this wave's own stories require, and only for W04-E04-S001." Blocks W04-E04-S002 (DATA-08
W6-T2 through T5 build on this story's widened, versioned hash chain).

## Affected packages or components

`kernel/audit` (the `chainHash`, `Verify`, `Anchor`/`CheckAnchor` functions and their supporting
types); a new migration adding the `hash_version` column, shipped through W02-E01's protocol tooling.

## Compatibility considerations

This is a **breaking format change**, not an additive one. Per PLAN DATA-08 W6-T1's own risk column:
"No `hash_version` column exists today; widening makes every historical row unverifiable under
new-scheme verification unless a version discriminator is added in the same migration and
verification branches by row version." The `hash_version` column and version-branched verification
are the compatibility mechanism itself — this is not an optional hardening add-on, it is the
mechanism that makes the breaking change survivable. wowsociety structurally depends on
`kernel/audit`: `identity/service.go` and `policy/service.go` both hold `*kaudit.Writer` fields;
`impersonation.go` writes two `s.audit.Record(...)` calls for grant/revoke — "a load-bearing
compliance flow"; `cmd/api/main.go` wires `kaudit.New(...)` for API-key audit. wowsociety produces
real, live audit rows today, and changing `chainHash`'s input set changes the hash of every new row
after the change lands. No call-shape change is required on wowsociety's side (only internal hash
computation plus the new version column, per the risk note); wowsociety must re-run any audit-
verification tooling after upgrading and confirm historical rows still verify under the
backward-compatible v1 scheme. Zero effect on wowsociety until it bumps `FRAMEWORK_VERSION` past this
story's commit — a dedicated wowsociety-side staging verification pass is required before that bump
(PROD-05, tracked as coordination context, not this story's own scope).

## Security considerations

This story directly strengthens the tamper-evidence guarantee the audit chain exists to provide — the
per-field tamper test (mutating `metadata`, `tx_id`, and every other declared field independently) is
itself the acceptance-defining security control, not a supplementary hardening test. The
canonicalization requirement (hash a pre-serialization form of `metadata`, never the stored jsonb) is
also a security-relevant control: hashing the stored jsonb directly would be non-reproducible across
reads, which would make the hash unusable as a tamper-evidence mechanism rather than merely
suboptimal.

## Performance considerations

None identified beyond the migration's own execution profile, which is W02-E01's protocol's concern
(lock-timeout budget, backfill batching) rather than a separate performance consideration this story
introduces. The widened hash computation itself (a few additional fields) is not expected to be a
measurable runtime cost relative to the existing 15-field hash.

## Observability considerations

Verification failures under either the v1 or v2 branch should be distinguishable in logs/errors —
so an operator investigating a failed verification can tell which scheme was in play, not just that
verification failed. This is a reasonable implementation-time addition given the version-branch
design; not separately mandated by the source beyond the version-branching requirement itself.

## Migration considerations

This story is itself a schema migration — adding the `hash_version smallint NOT NULL DEFAULT 1`
column — and is expected to be a direct, live exercise of W02-E01's online-migration protocol given
its confirmed breaking-change risk against a live-production table. The migration must ship in the
same migration unit as the `chainHash` widening, per D-04's decision text ("Add a `hash_version
smallint NOT NULL DEFAULT 1` column in the same migration that widens `chainHash`'s field coverage"),
not as two separately-sequenced changes.

## Documentation requirements

Document the widened field list, the canonicalization approach for `metadata`, the `hash_version`
column and its version-branch semantics, and the exact `hash_version` value assigned to the new
scheme, so a future reader of `kernel/audit` understands why the discriminator exists and how
verification selects a branch.

## Acceptance criteria

- **AC-W04-E04-S001-01**: `chainHash` covers every persisted field, including canonicalized
  `metadata` and `tx_id` alongside the previously-covered fields (all nullable fields, sequence, ID,
  timestamps, previous hash). A tamper test mutates each declared field independently and asserts
  every one fails verification — not a single combined mutation, not a subset of fields.
- **AC-W04-E04-S001-02**: A `hash_version smallint NOT NULL DEFAULT 1` column exists, added in the
  same migration that widens `chainHash`'s field coverage, per D-04. `Verify` branches by
  `hash_version`: historical rows (`hash_version = 1`) verify correctly under the original 15-field
  scheme; new rows (the new version value) verify correctly under the widened scheme including
  metadata and tx_id.
- **AC-W04-E04-S001-03**: The migration ships through W02-E01's online-migration protocol
  (classified via its manifest schema, budgeted under its lock-timeout mechanism, and run through
  expand/backfill/validate/contract as applicable), not as an ad hoc one-off migration, given its
  confirmed breaking-change risk against a live-production table.

## Required artifacts

- The widened `chainHash` implementation and its metadata-canonicalization function.
- The `hash_version` column migration (via W02-E01's protocol).
- The version-branched `Verify` implementation.
- Documentation of the widened field list, canonicalization approach, and version-branch semantics.
See `artifacts/index.md`.

## Required evidence

- Per-field tamper test output (metadata, tx_id, and every other declared field independently).
- Version-branch verification test output (v1 historical-row branch; v2 new-row branch).
- Confirmation the migration was classified and shipped through W02-E01's protocol (manifest entry,
  lock-timeout budget compliance).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W02-E01 recorded,
owner/reviewer assignment pending, the exact `hash_version` new-scheme value and metadata-
canonicalization approach explicitly recorded as implementation-time decisions rather than silently
assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the per-field tamper test genuinely covers every
declared field independently and D-04's version-branch design was implemented exactly as ratified.

## Risks

RISK-W04-002 (wave-level: this task's confirmed highest-risk status in PF-DATA's Wave-6 scope,
breaking format change hitting wowsociety's live audit rows) and RISK-W04-E04-S001-001 (this story's
own elaboration: an incorrect canonicalization of `metadata` could reintroduce non-reproducibility
even after the widening) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Cannot be reduced to fully resolved within this story's own closure. PLAN DATA-08 W6-T1's own risk
column is explicit that this is the single highest-risk task in the epic's scope; even with D-04's
design implemented exactly as ratified and the per-field tamper test passing, final confidence
requires the product-side PROD-05 staging drill (wowsociety re-verifying its own live rows before
bumping `FRAMEWORK_VERSION`), which is outside this story's own closure authority. This story's
closure records that residual risk explicitly, not silently.

## Plan

See `plan.md`.
