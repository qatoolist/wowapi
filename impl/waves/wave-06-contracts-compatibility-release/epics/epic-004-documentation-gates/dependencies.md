---
id: W06-E04-DEPS
type: epic-dependencies
epic: W06-E04
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E04 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W05-E03** (cross-wave) — S002's T4 depends on AR-03's authoritative model export existing (AR-05
  T4's own dependency row: "dep AR-03 T1, T5"). W05-E03 is this wave's own upstream wave (W06 depends on
  W05).
- **W05** (full wave, transitively) — per this wave's own entry gate.

## Downstream (epics/waves that depend on this epic)

None identified — no other epic or wave in this programme's own scope depends on this epic's
documentation-gate closure.

## Internal (within this epic)

S001 (doc-example-compile-gate) has no dependency on S002 — T3's compile-gate mechanics are entirely
self-contained, operating on whatever normative doc examples currently exist, independent of AR-03's own
manifest work. S002 (generated-docs-and-labels) depends on W05-E03 cross-wave, not on S001.

## Cross-wave dependencies

W05-E03 (AR-03 remainder) is this epic's one cross-wave dependency, landing specifically on S002's T4.

## External dependencies

None new. T3's extractor tool is described in MATRIX CS-22 as "a small extractor tool (~150 LOC) on
stdlib (`go/parser` not even needed — build failure is the check); no new dependency."

## Repository dependencies

None cross-repo for this epic's own closure. MATRIX CS-22 confirms: "wowsociety: none (wowapi docs
only); pattern reusable there later."

## Tooling dependencies

None beyond the already-available Go toolchain (`go build`) and `make`.

## Decision dependencies

None. See `epic.md` "Required decisions."
