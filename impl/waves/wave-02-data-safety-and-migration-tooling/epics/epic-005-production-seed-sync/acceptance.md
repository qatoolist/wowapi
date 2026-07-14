---
id: W02-E05-ACCEPTANCE
type: epic-acceptance
epic: W02-E05
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E05 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W02-05 there maps onto this epic).

## AC-W02-E05-01 — Design investigation complete before implementation

Every design question named in S001's `plan.md` "Unresolved questions" — catalog manifest
schema/format, versioning scheme, CLI command shape, idempotency mechanism, seed/catalog-hash
computation and readiness-payload placement, RLS/role posture, dry-run output format, and
audit-record integration — has a documented decision with rationale, recorded in T001's output
design document before any implementation task (T002–T005) began. Any decision of D-0N-caliber
significance was escalated for ADR treatment per `epic.md`'s process safeguard, not silently made
in-story. Traces to W02-E05-S001 (T001).

## AC-W02-E05-02 — Seed-sync path delivered with all five named properties

The seed-sync path exists and each of CS-21's five named properties is proven by its own test:
idempotent (a second run against an already-synced database converges with no spurious writes),
RLS-respecting (per T001's documented role posture, verified rather than bypassed), versioned
catalog manifests (sync consumes the ratified manifest format and records the manifest version),
dry-run (produces a change plan without writing), and audit (each run produces a durable audit
record per T001's documented integration decision). Traces to W02-E05-S001 (T002, T003, T005).

## AC-W02-E05-03 — CS-21's fixed acceptance bar holds, fail-first proven

Verbatim per MATRIX CS-21: "a prod-profile boot on an empty catalog DB reaches readiness only after
seed-sync has run, and the readiness payload reports the seed/catalog hash." The fail-first half is
also proven: before the fix, a prod-profile boot with empty catalogs silently reaches a
deny-everything ready state (CS-21: "currently silently deny-everything"); after, it returns a
named readiness failure until seed-sync has run. Traces to W02-E05-S001 (T004).

## AC-W02-E05-04 — Independent review passed

S001 (P0-prod) has passed independent review per mandate §14. The review specifically confirms:
the T001 design decisions were documented before implementation began (not backfilled after the
fact); the RLS posture decision (RISK-W02-E05-002) is genuinely justified rather than a silent
superuser bypass; and no design question from `plan.md`'s "Unresolved questions" was resolved
without a recorded rationale. Traces to W02-E05-S001 (T006).

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA, applied uniformly across this wave including FBL-02 — see `wave.md`:
"FBL-02, though sourced from the REVIEW/MATRIX rather than PLAN's PF-DATA table, is a deployment-
readiness/seed-sync concern that shares the same data/reliability accountability").
