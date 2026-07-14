---
id: W00-E01-ACCEPTANCE
type: epic-acceptance
epic: W00-E01
wave: W00
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E01 — Epic-level acceptance criteria

Numbered per mandate §5 pattern (`AC-W00-E01-NN`), each traceable to the story that proves it. This
epic cannot be marked `accepted` on the strength of "all 9 tasks complete" alone — every criterion
below requires actual, reviewed evidence (mandate §7, §14).

**Satisfaction record (2026-07-13):** all four criteria below are **satisfied**. AC-W00-E01-01,
-02, and -04 are proven by the three stories' verification and evidence records; AC-W00-E01-03 was
resolved by the conductor's DEV-W00-E01-S001-002 adjudication (AC-04 re-scoped to the executed
T1/T2 slice; future-state references routed to AR-05 T5 / W06-E04-S002). Independent review gate
passed 2026-07-13 (reviewer W00ReviewGate; conductor concurs).

## AC-W00-E01-01 — All nine re-verification tasks executed with a registered outcome

Every task across S001 (T001-T003), S002 (T001-T003), and S003 (T001-T003) has been executed at the
epic's closing commit SHA and has produced either:
- a `pass` result with a registered evidence ID in the owning story's `evidence/index.md`, or
- a `failed`-status evidence record plus an open follow-up remediation task, with the owning story
  not moved to `accepted` until the regression is resolved or explicitly accepted as a residual risk.

Traces to: W00-E01-S001, W00-E01-S002, W00-E01-S003 (each story's own AC-...-01 through -03).

## AC-W00-E01-02 — Verification records complete for every acceptance criterion

Every story's `verification.md` post-execution record (actual result, pass/fail, evidence
identifier, execution date, commit/revision, environment, reviewer, findings, retest status, final
conclusion) is filled in for every acceptance criterion listed in that story's `story.md` front
matter `acceptance_criteria`. No acceptance criterion may be left with an empty verification row.

Traces to: W00-E01-S001, W00-E01-S002, W00-E01-S003.

## AC-W00-E01-03 — AR-05 scope conflict resolved

The conflict between `wave.md`/`epics/index.md` (stating W00-E01 covers AR-05 T1/T2) and
`impl/analysis/requirement-inventory.md` (canonically targeting AR-05 to W06-E04-S002) — see
`epic.md` "Out of scope" and `risks.md` RISK-W00-E01-004 — has been explicitly ruled on by the
acceptance authority: either the wave-level documents are corrected to remove the AR-05 reference
from W00-E01's scope, or a fourth story is added to this epic with its own full planning structure.
This epic may not move to `accepted` while this conflict is open.

Traces to: epic-level (no single story owns this — it is a cross-document consistency finding).

## AC-W00-E01-04 — Evidence records are complete per policy

No evidence record registered under any of the three stories' `evidence/index.md` is missing a
commit SHA, execution command, environment, date/time, or result field, per
`impl/governance/evidence-policy.md`'s required-field list. Spot-checked by independent review before
epic acceptance (mandate §14).

Traces to: W00-E01-S001, W00-E01-S002, W00-E01-S003.

## Acceptance authority

Framework architecture lead (role-based; see `../../wave.md` "Acceptance authority" — no named human
DRI assigned yet, per `impl/index.md`'s scope-discipline note).
