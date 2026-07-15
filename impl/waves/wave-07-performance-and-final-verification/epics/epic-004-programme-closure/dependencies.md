---
id: W07-E04-DEPS
type: epic-dependencies
epic: W07-E04
wave: W07
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07-E04 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W07-E01, W07-E02, W07-E03** (this wave's own other three epics) — the final gate's own re-run scope
  cannot be meaningful until all three have reached their own closure state.
- **All prior waves (W00-W06)** — the gate re-runs across the whole programme.

## Downstream (epics/waves that depend on this epic)

None — this is the programme's own terminal epic; nothing downstream exists.

## Internal (within this epic)

**S002 depends on S001** — the closure report and claim-upgrade decision package cannot be honestly
assembled until the final gate re-run (S001) has actually produced its own verdict; S002 consumes S001's
own output directly.

## Cross-wave dependencies

All prior waves (W00-W06), as stated above.

## External dependencies

None new.

## Repository dependencies

None cross-repo — this epic's own re-run scope is the wowapi repository's own programme, though W07-E03's
own PROD-01..05 verification (a sibling epic this epic's S001 also consumes as an input) touches the
wowsociety-coordination question indirectly, without this epic itself requiring wowsociety repository
access.

## Tooling dependencies

None new beyond whatever tooling REVIEW's own original §30 gate used (documentation-and-evidence review,
not an automated tool).

## Decision dependencies

None new. See `epic.md` "Required decisions."
