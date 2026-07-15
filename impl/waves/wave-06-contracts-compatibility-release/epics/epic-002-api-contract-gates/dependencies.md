---
id: W06-E02-DEPS
type: epic-dependencies
epic: W06-E02
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E02 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W06-E01** (this wave) — S003's T5 leg depends on W06-E01-S001 (DX-03 design); S003's T7 leg depends
  on W06-E01-S002 (DX-04).
- **W05-E03** (cross-wave) — S003's T5 leg depends on AR-03's remainder (the "concept doesn't exist in
  current source" gap MATRIX CS-15 names for event/schema compatibility).
- **W05** (full wave, transitively) — per this wave's own entry gate.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W06-E03-S001 (REL-01) | This epic's S002 (REL-03a T9, SBOM/provenance-verify) | PLAN REL-03 T9's own framing: "Folds directly into REL-01 T8/T9 — not separate work, just REL-03's naming of a property REL-01 already builds." Shared evidence, not a blocking dependency in the other direction. |

## Internal (within this epic)

- **S003 depends on S001** for its T3 leg (OpenAPI semantic diff, MATRIX CS-15: "Blocked on DX-06 — a
  lossy merge can't be meaningfully diffed").
- **S002 has no internal dependency on S001 or S003** — REL-03a's six tasks (Go API diff, compile
  matrix, config compat, migration upgrade-drill, arch smoke, SBOM/provenance-verify) target disjoint
  surfaces from DX-06's OpenAPI-merge work and REL-03b's blocked legs; S002 may proceed independently
  once this epic's own entry gate (W06-E01, transitively W05) is satisfied.
- **S001 has no internal dependency on S002 or S003.**

## Cross-wave dependencies

W05-E03 (AR-03 remainder) is this epic's one cross-wave dependency beyond the W05 entry gate, landing
specifically on S003's T5 leg.

## External dependencies

DX-06 T2 introduces a candidate new external dependency (an OpenAPI 3.1 validator, `pb33f/libopenapi`
per MATRIX CS-15's evaluation candidate) — not yet approved; the decision is an implementation-time task
in S001. REL-03a T1 uses `golang.org/x/exp/apidiff`/`gorelease` — "the standard Go answers" per MATRIX
CS-15, already a low-risk, mature-tooling choice, not requiring the same security-review weight as a
new OpenAPI validator dependency.

## Repository dependencies

None cross-repo for this epic's own closure. wowsociety impact is real but non-blocking — see `wave.md`
"Repository dependencies" for the DX-06/REL-03 wowsociety notes (audit fragments once T1 ships; no
equivalent check of its own for REL-03's gates).

## Tooling dependencies

None new beyond the OpenAPI validator (S001, undecided) and `apidiff`/`gorelease` (S002, already
standard Go tooling).

## Decision dependencies

None in the D-0N sense. See `epic.md` "Required decisions."
