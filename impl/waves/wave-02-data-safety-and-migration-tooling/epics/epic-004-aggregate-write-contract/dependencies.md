---
id: W02-E04-DEPS
type: epic-dependencies
epic: W02-E04
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E04 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00** (full wave) — per `../../dependencies.md` (wave-level), W02 depends on W00's exit gate.
  This epic has no additional upstream dependency beyond that gate — DATA-06 has no D-0N ADR
  dependency and no dependency on any W01 finding.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W03-E04-S001 (DATA-07 T3) | W02-E04-S001 (T2, the `registrar_pg.go` actor-attribution fix) | PLAN's own PF-DATA cross-cutting note (2): "`kernel/resource/registrar_pg.go`'s nil-actor placeholder is one fix claimed by two findings (DATA-06 T2, DATA-07 T3) — one owner, not two PRs." `impl/analysis/wave-allocation-detail.md`'s W03-E04 row confirms: "T3 consumed from DATA-06 T2." This epic's T2 is the single owner of that fix; DATA-07 T3 (W03-E04-S001, hard-dependent on SEC-01/W03-E01 per PLAN's own note) reuses this epic's T2 mechanism directly rather than reimplementing it. This epic's closure should leave the T2 fix at a stable, documented location so W03-E04-S001 can cite it without re-deriving the design. |
| PROD-level: wowsociety `committeeseat.go` migration | W02-E04-S001 (T1, T3 — reference implementation) | PLAN's own wowsociety-impact note: "**Sequencing:** follow wowapi's T1/T3 (reference implementation proven first); not urgent, current pattern still functions." Tracked as a product-level coordination item, not a framework-implementation dependency — see `requirement-inventory.md` §D for the PROD-level item this maps to (no dedicated PROD-NN row exists for DATA-06 specifically; it is noted here for completeness per mandate §2.3's requirement to record excluded product-level work with rationale). |

## Internal (within this epic)

Single story (S001); no internal epic-level story dependency to record. Within S001, T2 depends on
T1 (same helper), T3 depends on T1+T2 (migrate reference handler onto the completed helper), and T4
depends on T1 (docs describe the implemented contract) — per PLAN DATA-06's own Depends-on column.
See `stories/story-001-aggregate-write-contract/story.md` "Dependencies" for the story-scoped
statement.

## Cross-wave dependencies

W03-E04-S001 (see downstream table above) is the only cross-wave dependency this epic creates. No
cross-wave dependency runs the other direction — this epic does not depend on any W03, W04, W05,
W06, or W07 item.

## External dependencies

None new. This epic's helper is built on the existing PostgreSQL/pgx transaction-management
machinery already used by `kernel/resource`, `kernel/audit`, and `kernel/outbox`.

## Repository dependencies

wowsociety's `committeeseat.go` (`internal/modules/identity/committeeseat.go:69-70`) uses "the exact
manual pattern DATA-06 targets" per PLAN's evidence — a real, moderate-severity product-level
exposure, but PLAN's own sequencing note states it is "not urgent, current pattern still functions"
and should follow this epic's reference implementation, not precede or block it. No wowapi-side
blocking dependency on wowsociety for this epic's own closure.

## Tooling dependencies

None beyond the existing Go/PostgreSQL toolchain.

## Decision dependencies

None. See `epic.md` "Required decisions."
