---
id: W03-E05-DEPS
type: epic-dependencies
epic: W03-E05
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E05 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W03-E01-S001** (hard, for T5 only) — `impl/analysis/wave-allocation-detail.md` E05 row: "T5
  durable audit (grant-ID field dep on E01 S001)." PLAN SEC-02 T5's own Depends-on column: "T1, T3,
  T4; benefits from SEC-01 T1." T4 (ratification) has no dependency on W03-E01.

## Downstream (epics/waves that depend on this epic)

None recorded — no other epic in `impl/analysis/wave-allocation-detail.md` or
`requirement-inventory.md` names a dependency on SEC-02's remainder.

## Internal (within this epic)

Single story (S001) — no internal cross-story dependency. Within S001 itself: T5 depends on T1, T3
(already executed, Wave 0) and T4 (this epic's own ratification task) per PLAN's own Depends-on
column for SEC-02 T5: "T1, T3, T4; benefits from SEC-01 T1."

## Cross-wave dependencies

None beyond the W03-E01-S001 dependency stated above.

## External dependencies

None.

## Repository dependencies

wowapi-internal (`kernel/workflow`). Per PLAN's own wowsociety-impact note: "Not affected. Zero
occurrences of `workflow.NewRuntime`, `workflow.Runtime`, `.Override(`, or any `kernel/workflow`
import anywhere in wowsociety. No required changes, no sequencing constraint." — this epic carries
zero wowsociety compatibility risk, unlike W03-E01/E03.

## Tooling dependencies

None.

## Decision dependencies

None new. T4's "reject or implement" choice is a story-level decision (see `epic.md` "Required
decisions"), not a decision dependency on an external ADR.
