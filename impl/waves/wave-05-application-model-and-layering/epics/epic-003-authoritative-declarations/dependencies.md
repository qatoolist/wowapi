---
id: W05-E03-DEPS
type: epic-dependencies
epic: W05-E03
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E03 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W05-E01** (full epic) — PLAN AR-04 T2's own dependency row: "AR-01 T1"; PLAN AR-03 T3's own
  dependency row includes "AR-01."
- **W05-E02** (full epic) — PLAN AR-03 T3's own dependency row: "T1, AR-01, AR-02" — the
  manifest-derived-projection tooling requires AR-02's compiled provider graph.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W06-E02-S003 (REL-03b compatibility-gate legs) | W05-E03-S001 (AR-03 manifest) | `impl/index.md`'s wave map: W06 depends on "W05 (AR-03 unblocks REL-03b legs)." |
| W06-E04-S002 (AR-05 T4/T5 generated docs and labels) | W05-E03-S001 (AR-03 manifest) | `impl/analysis/wave-allocation-detail.md`'s W06-E04-S002 row: "dep E02/W05-E03 manifest." |
| SEC-06 (W03 scope) | W05-E03-S002 (AR-04 T5 waiver mechanism) | `impl/analysis/wave-allocation-detail.md`'s own note: "T5 builds the shared waiver mechanism consumed by SEC-06/DX-07." |
| DX-07 T4 (W04-E04-S003) | W05-E03-S002 (AR-04 T5 waiver mechanism) | Same note as above — DX-07 T4 is explicitly "dep AR-04 T5 waiver mechanism" per `requirement-inventory.md`'s own DX-07 row. |

## Internal (within this epic)

S001 (AR-03) and S002 (AR-04) are independent of each other — disjoint concerns (manifest/
projection tooling vs. boot-strictness/waiver mechanism), both depending only on the shared upstream
epics above, not on each other.

## Cross-wave dependencies

None beyond the downstream table above.

## External dependencies

None new.

## Repository dependencies

None new — AR-03's wowsociety-impact note states "Affected passively, low severity, not breaking...
No required change; opt-in adoption for new modules." AR-04's wowsociety-impact note states
"Affected, low risk, not breaking, already partially assumed." Neither imposes a blocking
cross-repo dependency on this epic's own closure.

## Tooling dependencies

None new.

## Decision dependencies

None. See `epic.md` "Required decisions."
