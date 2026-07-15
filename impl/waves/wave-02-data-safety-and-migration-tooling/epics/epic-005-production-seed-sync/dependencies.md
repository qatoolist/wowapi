---
id: W02-E05-DEPS
type: epic-dependencies
epic: W02-E05
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E05 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00** (full wave) — per `../../dependencies.md` (wave-level), W02 depends on W00's exit gate.
  This epic has no additional upstream dependency beyond that gate — FBL-02 has no D-0N ADR
  dependency and no dependency on any W01 finding.
- **Explicitly none on W02-E01 (online-migration-protocol).** Both epics introduce a "manifest"
  concept, but they are different artifacts: DATA-09's *migration manifest* classifies schema-change
  risk (online/maintenance, lock budgets, backfill ownership); FBL-02's *catalog manifest* declares
  versioned seed content for production catalogs. Neither `requirement-inventory.md` (row FBL-02:
  "CS-21 acceptance bar fixed; design detail = story investigation task" — no dependency note), nor
  MATRIX CS-21, nor `wave-allocation-detail.md`'s W02 section records any FBL-02→DATA-09 dependency.
  This is stated affirmatively here so a future reader does not infer a dependency from the surface
  similarity of the two manifest concepts.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| PROD-03 (wowsociety readiness/timeout backports) | S001's delivered capability | CS-21: wowsociety's "generated `cmd/api/main.go:240-243` has the identical readiness gap — backport after T1." Product-level coordination item (`requirement-inventory.md` §D), consuming this epic's framework capability; not gated inside this wave. |
| FBL-03 / W01-E04-S002 (wowsociety upstream-register reconciliation) | S001's closure | CS-21: "PF-9 is *its* finding, closing it closes the register entry (FBL-03)." The register update is W01-E04-S002's documentation scope; the substantive closure it records is this epic's. |
| W04-E04-S003 (DX-07 readiness-truthfulness) | Compatible readiness-payload surface | Not a hard dependency in either direction — DX-07's seed/rule/model-hash checks and this epic's seed/catalog-hash reporting touch the same readiness payload. This epic builds its hash reporting compatibly; DX-07 later extends the same surface. Sequencing note only, no blocker. |

## Internal (within this epic)

Single story (S001). Its internal task sequencing is the epic's only internal dependency structure:
T001 (design investigation) gates T002–T005 (implementation), which gate T006 (independent review).
See the story's `tasks/index.md` grouping rationale.

## Cross-wave dependencies

None beyond the W00→W02 entry dependency and the downstream table above. In particular, this epic
neither depends on nor blocks W02-E01/E02/E03/E04 — per `../../dependencies.md`: "W02-E03, W02-E04,
W02-E05 are independent of W02-E01, W02-E02, and of each other."

## External dependencies

None anticipated. The seed-sync path is expected to build on the existing CLI surface, pgx/
PostgreSQL toolchain, and (plausibly) the existing `kernel/audit` infrastructure. If the design
investigation (S001-T001) concludes a new external dependency is required, that is RISK-W02-004's
contingency territory — recorded as a plan revision and, if D-0N-caliber, escalated per `epic.md`'s
"Required decisions" process safeguard, not silently added.

## Repository dependencies

None cross-repo for this epic's own closure. wowsociety impact is real but product-level (PROD-03
backport, FBL-03 register closure — see downstream table), excluded from framework-side closure per
mandate §2.3.

## Tooling dependencies

None new anticipated. The readiness-wiring test requires the existing prod-profile boot path and a
PostgreSQL instance with empty catalogs — both available in the existing test infrastructure.

## Decision dependencies

None in the D-0N sense. The catalog-manifest-format design decision is produced *by* this epic
(S001-T001), not consumed from elsewhere.
