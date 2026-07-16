---
id: TRACK-PROGRAMME-DEVIATIONS
type: deviation
title: Programme-level deviation records (not scoped to a single story)
status: active
created_at: 2026-07-16
updated_at: 2026-07-16
derived: false
---

# Programme-level deviations

Per `governance/templates/deviation-template.md` and mandate §2.6/§8.9: the approved plan is never
rewritten to look right — actual divergences are recorded here as separate deviation records. This
file holds deviations that are **programme-level** (not attributable to a single story's own
`deviations.md`) — coverage-gate changes, cross-wave sequencing bypasses, and commit-message
overclaims. Summary rows for these are also carried in `impl/tracking/deviation-register.md`.

All records below were created 2026-07-16 by **autopsy remediation R-1**, following findings in
`impl/reports/implementation-autopsy-report-2026-07-16.md`.

---

## DEV-PROG-001 — Coverage floor lowered 90.0% → 84.0% without a deviation/decision record

### Deviation ID

DEV-PROG-001

### Approved plan

The programme's coverage floor, per `AC-W00-03` and the baseline established in W00, was 90.0%
(`Makefile:324`, `COVERAGE_FLOOR`, pre-`e8cda6b`). No deviation or decision record authorized
changing it.

### Actual implementation

Commit `e8cda6b` ("feat: finalize wowapi implementation programme (Waves 00-07)") lowered
`COVERAGE_FLOOR` from 90.0 to 84.0 in the same commit that declared the programme finalized, with
`COVER_PKGS`/`COVER_EXCLUDE` (measurement scope) left unchanged. Measured coverage regressed
92.3% → 84.5% between the pre-`e8cda6b` baseline and `e8cda6b`, over the same measurement scope —
a genuine coverage regression, not a scope change, absorbed by weakening the gate rather than
recorded and fixed. (Autopsy finding **H-1**.)

### Reason

Not recorded at the time — no deviation or decision record exists for this change; it was
discovered only by the 2026-07-16 autopsy's direct diff of `Makefile` across `e8cda6b^` and
`e8cda6b`.

### Impact

The quality gate that is supposed to prevent exactly this kind of regression was itself weakened,
silently, in the commit claiming programme completion. `kernel/tracing` and `kernel/safety`
carried real logic at 0% coverage under the lowered floor (autopsy finding M-4).

### Risks

Untested code paths in security/observability-adjacent packages (`kernel/tracing`,
`kernel/safety`) ship without coverage; the lowered floor could mask further regression if not
ratcheted back up.

### Approval

Not originally approved by any recorded decision. Retroactively acknowledged 2026-07-16 via
**DEC-PROG-001** (status: proposed — human ratification pending), which records the interim floor
and a ratchet plan back to 90.0%.

### Compensating controls

`kernel/tracing` and `kernel/safety` brought to 100% coverage 2026-07-16 (measured against the
real DB, same measurement scope as the regression). DEC-PROG-001 sets the ratchet plan for the
remaining gap.

### Follow-up work

Autopsy remediation R-4: restore aggregate coverage to ≥90% (target the regression packages) or
formally ratify a lower floor via a fully-approved decision record (DEC-PROG-001, currently
`proposed`).

---

## DEV-PROG-002 — FBL-01 kernel re-home (and AR-01/AR-02/authz_epoch) executed outside tracked story lifecycle

### Deviation ID

DEV-PROG-002

### Approved plan

FBL-01's kernel re-home (moving 9 kernel packages under `foundation/`) is W05-E05's scope
(`W05-E05-S001`, `W05-E05-S002`). Per the programme's plan, this work executes only after
`W05-E05-S001` moves through its own lifecycle (`planned` → `ready` → `in-progress` → ... →
`accepted`), tracked in that story's own front matter and roll-ups.

### Actual implementation

The full 9-package re-home was executed on `main` in commit `e8cda6b`, with compatibility shims in
place, while `W05-E05-S001`'s `story.md` front matter remained `planned` and every task remained
`todo` (autopsy finding **H-7**). The same pattern applies to AR-01/AR-02 (`kernel/appmodel`,
`kernel/port`, ~765 LOC, tested, zero non-test imports — built but not wired; boot still
constructs registries directly) and to the `authz_epoch` migration, which is orphaned (no Go code
reads it) — both landed as real code with no corresponding story-lifecycle progression (autopsy
finding **H-6**). W05-E01-S001, W05-E02-S001, and W05-E03-S001 show the same
already-real-code-vs-`planned`-tracking contradiction pattern (autopsy verdict `contradictory` for
all four).

### Reason

Not recorded at the time. The work appears to have been executed as part of the same
finalization push as commit `e8cda6b`, ahead of and outside the sequencing the programme's own
plan defines for W05.

### Impact

The programme's own status ledger cannot be trusted to reflect what code actually exists: `W05`
reads `planned`/8-stories-missing while substantial, tested code for 3 of its stories is already on
`main`. This is a governance/tracking failure, not a security defect in the code itself (Fable
adjudicated this dispute in the autopsy as "High-governance rather than Critical-security").

### Risks

Future agents or reviewers trusting `W05`'s `planned` status could re-implement or conflict with
already-landed code; AR-01/AR-02's built-but-not-wired state risks bit-rot or divergence from
whatever eventually wires it in; the orphaned `authz_epoch` migration risks being forgotten
entirely.

### Approval

Not originally approved by any recorded decision. Disposition (wire in during W05 vs. formal
deferral) is recorded as an open item in **DEC-PROG-002** (status: proposed), deferred to the
Wave 05 execution owner.

### Compensating controls

This deviation record plus the per-story notes added 2026-07-16 to `W05-E01-S001`, `W05-E02-S001`,
`W05-E03-S001`, and `W05-E05-S001` (pointing back to this record) restore traceability without
altering those stories' honest `planned` tracking status.

### Follow-up work

Autopsy remediation R-7: decide and execute AR-01/AR-02 wiring + SEC-04 cache work (or formally
defer with records), then finish W05 per plan — tracked via **DEC-PROG-002**.

---

## DEV-PROG-003 — Sequencing-gate bypasses across W02/W03/W04/W05

### Deviation ID

DEV-PROG-003

### Approved plan

The programme's wave/epic entry criteria (mandate §6, wave-level `wave.md` entry-criteria
sections) require predecessor waves/epics to be accepted (or explicitly waived) before dependent
work begins — e.g. `W05`'s entry gate depends on `W03-E01` acceptance; `W04` depends on `W02`
closure.

### Actual implementation

Multiple sequencing gates were bypassed without a recorded waiver: `W04` executed while `W02` was
still unclosed (`W02`'s closure gate was, per **C-4**, falsely claimed passed — see the `W02`
closure-report.md correction note); `W05-E04-S001` reached `ready-for-review` despite `W05`'s hard
entry gate on `W03-E01` acceptance, while `W03` remains unaccepted (zero `W03` stories validly
accepted per autopsy findings C-3/H-5). (Autopsy finding **H-7**, cross-referenced with
**M-5**-adjacent findings.)

### Reason

Not recorded at the time. No waiver or deviation record authorized proceeding past these entry
gates.

### Impact

Work products from later waves rest on foundations (`W02`, `W03`) whose own acceptance is not
valid, compounding the traceability problem this remediation pass addresses. `RISK-001` (the
programme's own risk register) explicitly warned that this class of sequencing bypass could occur.

### Risks

If `W02` or `W03`'s underlying implementation is later found materially defective during real
review (R-3), dependent `W04`/`W05` work built on top may need rework.

### Approval

Not originally approved by any recorded decision or waiver.

### Compensating controls

`W02` closure-report.md status reverted to `verification` (this remediation pass); `W03` stories'
false acceptances reverted (`W03-E03-S001`, `W03-E02-S001`); real independent reviews for `W02`
(6 stories) and `W03` (all stories) are being scheduled per autopsy remediation R-3, which is a
prerequisite for re-validating anything built on top.

### Follow-up work

Autopsy remediation R-3 (execute the missing independent reviews) is a hard prerequisite before
any further acceptance activity proceeds against `W04`/`W05` work that depended on `W02`/`W03`.

---

## DEV-PROG-004 — Commit `e8cda6b` message overclaims programme finalization

### Deviation ID

DEV-PROG-004

### Approved plan

A commit claiming programme finalization ("feat: finalize wowapi implementation programme (Waves
00-07)") should reflect an actually-executed and reviewed programme closure gate (`W07-E04`), per
mandate §6/§8.9's truthful-status-reporting requirements.

### Actual implementation

Commit `e8cda6b`'s message claims Waves 00–07 finalized. The programme's own artifacts directly
contradict this: `W07` was `in-progress` (not closed), `SEC-05` story `blocked`, `RISK-W07-002`
open and unaccepted, and the status register un-regenerated at the time of the claim. The
programme closure gate (`W07-E04`) was never executed. (Autopsy finding **H-2**.)

### Reason

Not recorded at the time.

### Impact

An unsupported completion claim exists at the VCS level, permanently, since the commit is already
pushed to `main` and immutable without a history rewrite this remediation is explicitly not
authorized to perform.

### Risks

Future readers relying on `git log` alone (rather than `impl/`'s own front matter) would be misled
about actual programme state — exactly the failure mode this autopsy and remediation exist to
correct.

### Approval

Not applicable — this deviation documents an already-immutable historical fact, not a change
requiring approval.

### Compensating controls

The commit message itself cannot be edited (already pushed, and rewriting history is out of this
remediation's scope and explicitly against the "plans are never rewritten to look right"
principle applied at the ledger level). Compensating controls are this deviation record, the full
autopsy report (`impl/reports/implementation-autopsy-report-2026-07-16.md`), and this entire
reconciliation pass, which together give any reader outside `git log` the true state.

### Follow-up work

Autopsy remediation R-9: run the programme closure gate (`W07-E04`) for real once R-1..R-8
complete; only then would any future finalization claim (commit or otherwise) be truthful.
