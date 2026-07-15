---
id: W07-E01-DEPS
type: epic-dependencies
epic: W07-E01
wave: W07
status: satisfied
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01 — Dependencies

## Upstream (epics/waves this epic depends on)

- **All prior waves (W00-W06)** — per `../../dependencies.md` (wave-level), W07 depends on all seven.
  This epic's own specific consumers: W04-E01 (DATA-02 shared lease primitive) and W04-E02 (DATA-03
  remote-I/O-outside-tx primitives) for S003's own T5 (the leased-state-machine outbox rework).
  The accepted S003 evidence confirms those primitives were consumed successfully.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W07-E04-S001 (final verification gate, this wave) | This epic (all 4 stories) | The final gate's own re-run scope spans the whole programme, including this epic's own closure state, as one of many inputs. |

## Internal (within this epic)

S001 (PERF-02) is the epic's own shared-prerequisite story: its T1 (the §14 reference-environment stand-
up) is named explicitly in this wave's task brief as "a shared prerequisite across all 4 PERF stories in
this epic." S002, S003, S004 each independently reference `perf/reference-v1.json` (built by S001's own
T1) for their own before/after publication tasks — this is a shared-artifact dependency, not a strict
story-completion-order dependency: S002/S003/S004's own non-publication tasks (the actual code fixes)
may proceed in parallel with S001, but their final publication tasks consume S001's T1 output.

## Cross-wave dependencies

W04-E01, W04-E02 (DATA-02/DATA-03 lease primitives) for S003's T5, as stated above.

## External dependencies

The provisional Linux/amd64 GitHub Actions/container reference workflow exists and supports
relative/container comparison. A dedicated reference-performance environment remains an open human
decision under DEC-Q9; it gates only future absolute-SLO acceptance, not this epic's evidenced scope.

## Repository dependencies

None cross-repo for this epic's own closure. PERF-02 through PERF-05 are all confirmed "No code change
required" or "Not affected" for wowsociety per each finding's own PLAN wowsociety-impact note (PERF-02:
"wowsociety's requests already go through the identical framework-owned `TxManager.WithTenant` path";
PERF-03: "confirmed absent — societies are single-org tenants in E0"; PERF-04: "confirmed absent by
direct grep... zero jobs/notify/webhook/bulk imports"; PERF-05: "Indirectly yes, no wowsociety code
changes needed... wowsociety inherits whatever PERF-05 establishes with zero call-site changes").

## Tooling dependencies

`benchstat` (already integrated per PERF-06 T2, W00-E01-S002) for the statistical regression-gate
consumption each story's own before/after publication task relies on.

## Decision dependencies

DEC-Q9 remains `open-human`; its provisional relative/container policy is active and non-blocking.
