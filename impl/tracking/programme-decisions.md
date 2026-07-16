---
id: TRACK-PROGRAMME-DECISIONS
type: decision
title: Programme-level decision records (not scoped to a single story)
status: active
created_at: 2026-07-16
updated_at: 2026-07-16
derived: false
---

# Programme-level decisions

Per `governance/templates/decision-template.md` and mandate §11.8: architectural and
implementation decisions, including unresolved ones, must be recorded, not buried in prose. This
file holds decisions that are **programme-level** rather than scoped to one story. Summary rows
are also carried in `impl/tracking/decision-register.md`.

Both records below were created 2026-07-16 by **autopsy remediation R-1**, following findings in
`impl/reports/implementation-autopsy-report-2026-07-16.md`. Both are **`proposed`** — human
ratification pending — per mandate §11.8's requirement that unresolved decisions still be
recorded, not silently deferred.

---

## DEC-PROG-001 — Interim coverage floor acknowledgment and ratchet plan

### Decision ID

DEC-PROG-001

### Title

Interim coverage floor acknowledgment and ratchet plan

### Status

proposed

### Context

Commit `e8cda6b` lowered `COVERAGE_FLOOR` from 90.0% to 84.0% with no deviation or decision record
(see **DEV-PROG-001**, `impl/tracking/programme-deviations.md`). Measured coverage regressed
92.3% → 84.5% over an unchanged measurement scope — a genuine regression, not a scope change.
This decision must resolve what the programme's coverage floor actually is going forward and how
the regression gets closed, rather than leaving the gate silently weakened.

### Options considered

1. **Ratify 84.0% as the new permanent floor.** Cheapest, but formalizes a real quality regression
   without addressing it.
2. **Immediately restore 90.0% by build-failing until coverage recovers.** Correct in principle,
   but would block all further work on unrelated packages until the regression is fully closed,
   with no interim path.
3. **Acknowledge 84.0% as an explicit, dated interim floor with a ratchet plan back to 90.0%,
   raising the floor incrementally as specific regression packages are covered.** Chosen option —
   makes the regression visible and dated rather than silently absorbed, without freezing all
   other work.

### Decision

**Interim:** the 84.0% floor is acknowledged as a real regression from the 90.0% baseline, not
ratified as a new permanent target. **Ratchet plan:** raise the floor incrementally as coverage
recovers, with a target of restoring 90.0%. As of 2026-07-16, `kernel/tracing` and `kernel/safety`
— the two packages autopsy finding M-4 flagged at 0% coverage despite real logic — have been
brought to 100% coverage; the next floor increment and measurement are recorded as of
2026-07-16. **This decision is `proposed`, not `ratified`** — it requires human ratification
before the ratchet schedule is binding.

### Rationale

A silent floor drop in the same commit that claims programme finalization is exactly the failure
mode mandate §2.6/§8.9 exist to prevent. Making the regression explicit and dated, with a
concrete recovery step already taken (two packages to 100%), is more honest than either silently
keeping 84.0% or blocking all work pending a full restore no one has scoped.

### Consequences

Coverage-gated CI continues to pass at 84.0% today; readers of `Makefile`/CI config must consult
this decision (not just the number) to know it is an acknowledged interim state, not a target.
Future PRs into the previously-0%-coverage packages must not regress them.

### Related source items

Autopsy finding H-1; DEV-PROG-001; autopsy remediation R-4; AC-W00-03.

### Date

2026-07-16 (measurement recorded this date).

### Deciders

Proposed by autopsy remediation R-1 (conductor); pending human/acceptance-authority ratification.

---

## DEC-PROG-002 — Disposition of built-but-not-wired AR-01/AR-02 and unimplemented SEC-04

### Decision ID

DEC-PROG-002

### Title

Disposition of built-but-not-wired AR-01/AR-02 and unimplemented SEC-04

### Status

proposed

### Context

`kernel/appmodel` and `kernel/port` (AR-01/AR-02, ~765 LOC, tested) exist on `main` with zero
non-test imports — the framework boot sequence still constructs registries directly, not through
these types. SEC-04 (bounded/epoch authz cache) is not implemented: the existing authz cache is
unbounded TTL, and the `authz_epoch` migration is orphaned (no Go code reads it). (Autopsy finding
H-6; see **DEV-PROG-002** for the tracking-lifecycle side of this same finding.) A decision is
needed on whether this code gets wired in during W05 as originally planned, or is formally
deferred.

### Options considered

1. **Wire in during W05.** AR-01/AR-02 and the authz_epoch-backed bounded cache get connected to
   the real boot sequence as W05-E01/E02/E05 (and a new/updated SEC-04 story) execute for real.
   Matches the original plan; requires W05 to actually execute, which depends on W03 acceptance
   (currently unaccepted — see DEV-PROG-003).
2. **Formal deferral.** Record AR-01/AR-02 and SEC-04 as intentionally deferred (mandate-compliant
   deferral, not silent abandonment), with a target milestone and reopen trigger, and remove them
   from W05's critical path so W05 can close without them.
3. **Delete the orphaned code.** Rejected outright — the code is real, tested, and represents
   completed engineering work; deleting it would destroy value for no governance benefit.

### Decision

**Unresolved — deferred to the Wave 05 execution owner.** This decision record exists to make the
open question explicit (mandate §11.8: unresolved decisions must still be recorded), not to
resolve it here. The Wave 05 execution owner chooses between option 1 (wire-in-W05) and option 2
(formal deferral) once W03's independent-review prerequisite (autopsy remediation R-3) is
satisfied and W05 can honestly begin. Reopen triggers, if option 2 is chosen, follow
`impl/tracking/deferred-items-register.md`'s existing conventions (target milestone + explicit
reopen condition, not silent drop).

### Rationale

Neither option can be soundly chosen before W05's own entry gate (W03 acceptance) is honestly
satisfied — choosing now would repeat the sequencing-gate-bypass pattern this remediation pass is
correcting (DEV-PROG-003).

### Consequences

`kernel/appmodel`/`kernel/port` remain unwired and the `authz_epoch` migration remains orphaned
until this decision is ratified one way or the other. SEC-04 remains not implemented in the
interim — the existing unbounded-TTL authz cache continues in production.

### Related source items

Autopsy findings H-6, H-7; DEV-PROG-002; DEV-PROG-003; autopsy remediation R-7;
`impl/tracking/deferred-items-register.md`.

### Date

2026-07-16.

### Deciders

Proposed by autopsy remediation R-1 (conductor); decision explicitly deferred to the Wave 05
execution owner, pending human/acceptance-authority ratification of that assignment.
