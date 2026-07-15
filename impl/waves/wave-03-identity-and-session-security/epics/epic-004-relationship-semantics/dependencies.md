---
id: W03-E04-DEPS
type: epic-dependencies
epic: W03-E04
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E04 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W03-E01** (hard, blocking) — must be `accepted`, not merely started, before this epic's
  implementation work begins. PLAN §5.3 DATA-07 T1's own Depends-on column: "Hard dependency on
  PF-SEC's SEC-01 — do not schedule before it lands." `requirement-inventory.md` row DATA-07 records
  disposition "blocked→planned" for exactly this reason.
- **W02-E04-S001** (soft, cross-reference) — DATA-06 T2's actor-attribution mechanism in
  `registrar_pg.go` is consumed by this epic's T4, not reimplemented. If DATA-06 T2 has not landed,
  T4 is blocked on that specific input.

## Downstream (epics/waves that depend on this epic)

None recorded — no other epic in `impl/analysis/wave-allocation-detail.md` or
`requirement-inventory.md` names a dependency on DATA-07.

## Internal (within this epic)

Single story (S001) — no internal cross-story dependency. Within S001 itself: T2 depends on T1
(PLAN's own Depends-on column for DATA-07 T2: "T1"); T4 depends on T1-T3 and also on SEC-04's
cache-epoch work (PLAN: "T1-T3; also depends on SEC-04's cache-epoch work") — since T3 itself is
out of this epic's scope (cross-referenced to DATA-06), T4's practical dependency within this
epic's own task set is T1 and T2, plus the external DATA-06 T2 and W05-E04-S002 dependencies noted
above.

## Cross-wave dependencies

- W03-E04-S001-T4 ↔ W05-E04-S002 (SEC-04 epoch table, D-06) — deferred-link, not a blocking
  dependency for this epic's own acceptance; see `epic.md` "Dependencies" and `../../risks.md`
  RISK-W03-003.

## External dependencies

None beyond the DATA-06/SEC-04 dependencies stated above.

## Repository dependencies

wowapi-internal (`kernel/relationship`). Per `requirement-inventory.md` row DATA-07: "No confirmed
direct usage" in wowsociety ("`grep -rn 'kernel/relationship'` returns zero matches across
wowsociety... Re-verify at DATA-07 ship time" per PLAN's own note) — to be re-confirmed at this
epic's own execution time, not assumed from the cited snapshot.

## Tooling dependencies

None beyond what W03-E01 and W02-E04-S001 already establish.

## Decision dependencies

None new. This epic depends on SEC-01's principal model (W03-E01) as an implementation input, not
as a decision dependency in the ADR sense.
