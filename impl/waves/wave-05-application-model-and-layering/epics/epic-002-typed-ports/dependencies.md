---
id: W05-E02-DEPS
type: epic-dependencies
epic: W05-E02
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E02 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W05-E01** (full epic) — PLAN AR-02 T1's own dependency row: "Depends-on: AR-01 T1, T2." T2's own
  capability-confusion safety proof requires AR-01's `Registrar` type to exist in its final,
  D-02-compliant form.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W05-E03 (AR-03, this wave) | W05-E02-S002 (T5, three-profile projection) | PLAN AR-03 T3's own dependency row: "T1, AR-01, AR-02" — the manifest-derived-projection tooling depends on AR-02's compiled graph existing. |
| W05-E05 (FBL-01, this wave) | W05-E02 (full epic) | MATRIX CS-01's own "Dependencies: AR-01/02 first." |

## Internal (within this epic)

S001 → S002 → S003 in dependency order, matching PLAN AR-02's own T-number chain: T3-T4 depend on
T1-T2; T5 depends on T1-T4 (and, per its own row, on AR-03's manifest shape being fixed — a
forward-looking sequencing note, not a hard block, since this epic delivers the projection mechanism
AR-03 later consumes); T6 depends on T1-T5; T7 depends on T1-T6.

## Cross-wave dependencies

None beyond the downstream table above.

## External dependencies

None new. Generic Go type-parameter machinery only.

## Repository dependencies

None. PLAN's own AR-02 wowsociety-impact note: "Not affected. Confirmed via repo-wide search: zero
call sites for `ProvidePort`/`Port(` anywhere in wowsociety. No breaking change, no required action,
no sequencing constraint." T7's legacy port adapter accordingly has no wowsociety-facing
compatibility risk.

## Tooling dependencies

None new.

## Decision dependencies

None. See `epic.md` "Required decisions."
