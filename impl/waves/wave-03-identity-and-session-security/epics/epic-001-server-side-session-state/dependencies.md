---
id: W03-E01-DEPS
type: epic-dependencies
epic: W03-E01
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E01 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W01-E03-S002** (central-validation seam) — per `../../dependencies.md` (wave-level), SEC-01's
  new grant-table endpoints (T1/T2) should be built against the `RouteMeta` contract-enforcement
  pattern this story establishes.
- **W02-E01** (DATA-09 online migration protocol) — per `../../dependencies.md`, the
  `identity_grant` migration (T1) is "genuinely new schema" (PLAN's own words) and should roll out
  through the online expand/backfill/validate/contract protocol rather than a one-off migration.
- **W00-E02-S003** — this epic's S001 references `ADR-W00-E02-S003-001` (D-01) as an already-
  ratified design premise.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W05 (entire wave entry) | This epic's acceptance | `impl/analysis/wave-allocation-detail.md` "Cross-wave sequencing notes": "W05 entry requires W03-E01 acceptance (actor model stability)." |
| W03-E04-S001 (DATA-07) | This epic's acceptance | PLAN §5.3 DATA-07 T1: "Hard dependency on PF-SEC's SEC-01 — do not schedule before it lands." |
| W03-E05-S001-T5 (SEC-02 durable audit) | This epic's S001 (grant-ID field) | `impl/analysis/wave-allocation-detail.md` E05 row: "T5 durable audit (grant-ID field dep on E01 S001)." |

## Internal (within this epic)

| Story | Depends on | Type | Notes |
|---|---|---|---|
| S001 | none (within this epic) | — | T1/T2/T3 form the foundational grant-schema-and-membership slice; no prerequisite within W03-E01 itself. |
| S002 | S001 | Hard | PLAN: T4 depends on T2 (S001); T5 depends on T1 and T2 (both S001). |
| S003 | S001 | Hard | PLAN: T6 and T7 both depend on T2 (S001). |
| S004 | S001, S002 (substantially planned) | Soft | The coordination plan needs T1's grant-table shape and T5's resolver contract to sequence realistically against; S004 can be drafted early and refined as S001/S002 firm up rather than strictly gated on their acceptance. |

## Cross-wave dependencies

None beyond the W00/W01/W02 upstream dependencies stated above and the W05/W03-E04/W03-E05
downstream dependencies stated above.

## External dependencies

- IdP claim-contract confirmation (DEC-Q1) — human-blocked; this epic proceeds against the
  documented safe default rather than waiting.
- wowsociety staging environment and `identity_impersonation_session` data — needed for S001's T2
  data-audit step and for S004's staging-validation plan.

## Repository dependencies

wowapi-internal for S001/S002/S003 (all touch `kernel/auth/`, the database migration, and
`PrincipalStore`). S004 is the sole story in this epic with a cross-repository dimension, and it
produces coordination documentation only — no wowsociety code.

## Tooling dependencies

DATA-09's online-migration tooling (W02-E01) for S001's `identity_grant` migration rollout.

## Decision dependencies

- D-01 (`ADR-W00-E02-S003-001`) — from W00-E02-S003, referenced by S001, not re-authored.
- DEC-Q1 — human-blocked, tracked at wave/programme scope; S001's story records the safe default it
  proceeds against.
