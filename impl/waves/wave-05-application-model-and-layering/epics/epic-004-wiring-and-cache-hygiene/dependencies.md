---
id: W05-E04-DEPS
type: epic-dependencies
epic: W05-E04
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E04 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W03-E01** (full-wave entry gate) — per `../../dependencies.md` (wave-level); SEC-04 T4's epoch-
  bump wiring extends to SEC-01's grant-table mutation paths, which land in W03-E01.
- No dependency on W05-E01, W05-E02, or W05-E03 — this epic's scope (AR-06 remainder, SEC-04) is
  disjoint from the ownership-model rework those epics deliver.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| DATA-07 T4 (W03-E04-S001, already landed) | W05-E04-S002 (SEC-04 cache-invalidation) | `impl/analysis/wave-allocation-detail.md`'s explicit note: "DATA-07 T4 cache-invalidation AC closes here." This is a cross-wave AC-closure relationship, not a code dependency — DATA-07 T4's own acceptance criterion is satisfied by this epic's work, recorded by ID, not by modifying DATA-07's own files. |

## Internal (within this epic)

S001 (AR-06) and S002 (SEC-04) are independent of each other — disjoint code surface
(`kernel/kernel.go` constructor closures vs. `kernel/authz/caching.go`).

## Cross-wave dependencies

None beyond the W03-E01 entry dependency and the DATA-07 T4 downstream AC-closure relationship
above.

## External dependencies

`hashicorp/golang-lru/v2`, named as an "approved dep" in MATRIX CS-17, for S002's T1.

## Repository dependencies

None. AR-06's wowsociety-impact note: "Not affected." SEC-04's wowsociety-impact note: "Not
affected... removes an undocumented obligation — strictly safer."

## Tooling dependencies

None new for S001 (extends existing lint infrastructure, consistent with AR-02 T6's own note: "may
share tooling with AR-02 T6").

## Decision dependencies

D-06, ratified in W00-E02-S003, referenced by S002. See `epic.md` "Required decisions."
