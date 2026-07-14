---
id: W03-E03-DEPS
type: epic-dependencies
epic: W03-E03
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E03 — Dependencies

## Upstream (epics/waves this epic depends on)

None beyond the wave-level W00/W01/W02 entry criteria. No ADR, no other W03 epic.

## Downstream (epics/waves that depend on this epic)

None recorded — no other epic in `impl/analysis/wave-allocation-detail.md` or
`requirement-inventory.md` names a dependency on SEC-03.

## Internal (within this epic)

Single story (S001) — no internal cross-story dependency.

## Cross-wave dependencies

None.

## External dependencies

None.

## Repository dependencies

wowapi-internal (`kernel/webhook`). Per PLAN's own wowsociety-impact note: "Zero `kernel/webhook`
import, zero custom `Verifier` implementation anywhere in wowsociety" — to be re-confirmed at this
epic's own execution time (fresh grep), not merely trusted from the cited snapshot, per
RISK-W03-006's mitigation.

## Tooling dependencies

None.

## Decision dependencies

None.
