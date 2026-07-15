---
id: W04-E03-RISKS
type: epic-risks
epic: W04-E03
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03 — Risks

The wave-level risk register (`../../risks.md`) carries four risks (RISK-W04-001 through
RISK-W04-004); none targets DATA-04 or W04-E03 by name. This is recorded explicitly rather than
silently omitted, per mandate §18 ("record assumptions explicitly") — DATA-04's own source material
(the T1–T6 row table reproduced in `epic.md`) names risk-relevant notes only at the individual task
level (e.g. T2's "Additive," T3's "Preserve `runItem`'s existing idempotent completion CAS guard,"
T5's "Larger scope — schedule in the full P1/Wave-3 slice, not the fast-track stopgap"), not a
standalone epic-level or wave-level risk entry. Two epic-scoped risks are derived below from those
task-level notes, since they carry genuine severity and mitigation content worth tracking
independently rather than leaving buried in a task table's "Risk" column.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W04-E03-001 | S001's stopgap (advisory lock or CAS at the `Service` API boundary) and S002's T2 lease-column migration are two independent single-processor-enforcement mechanisms that must be correctly sequenced: if S002's lease columns land while S001's advisory-lock/CAS code path is still active and not cleanly superseded, the two mechanisms could either conflict (a claim rejected by one but not the other) or leave a window where neither is actually enforcing exclusivity | Low-medium | Medium — a sequencing gap between the two mechanisms could reintroduce exactly the duplicate-processing race this epic exists to close, even though each mechanism individually behaves correctly in isolation | Medium | W04-E03-S001, W04-E03-S002 (T2) | S002's `plan.md` must record an explicit supersession step (S001's stopgap is deliberately superseded by S002's T2 lease-column mechanism, mirroring this wave's own `W04-E01-S001`-supersedes-`W02-E01-S002` pattern per `wave.md` "Assumptions") — not a silent parallel coexistence of both mechanisms | If S002's T2 is delayed, continue relying on S001's stopgap alone and record the extended interim period as an accepted, time-bounded technical-debt item, not a silent permanent fork | unassigned | open | Low once S002's `plan.md` records the explicit supersession step and its own concurrency test (AC-W04-E03-02) is passing |
| RISK-W04-E03-002 | S002's T3 (the atomic `SKIP LOCKED` leased claim) must preserve `runItem`'s existing idempotent completion CAS guard while replacing the plain unlocked `SELECT` claim path underneath it — per the source's own risk note on T3: "Preserve `runItem`'s existing idempotent completion CAS guard." An incorrect rewrite could silently drop or weaken that pre-existing guard while appearing to add a new, stronger claim-side lock | Low-medium | Medium-high — `runItem`'s completion CAS guard is the last line of defense against a duplicate completion write; losing it while believing the new claim-side `SKIP LOCKED` path alone is sufficient would be a regression disguised as a fix | Medium | W04-E03-S002 (T3) | T3's task record and its independent-review pass must explicitly confirm the completion CAS guard is unchanged (not merely "not obviously broken"), with a test asserting the guard still rejects a duplicate completion write | If the guard is found weakened during review, block story acceptance until it is restored and re-tested — do not accept T3 on the strength of the new claim-side lock alone | unassigned | open | Low once the independent-review task (S002-T005) explicitly confirms the guard is unchanged |

## Residual risk after mitigation

Both epic-scoped risks above are expected to reduce to low residual risk once their respective
mitigations — S002's explicit supersession step in `plan.md`, and the independent-review task's
explicit confirmation that `runItem`'s completion CAS guard is unchanged — are executed as planned.
No wave-level risk applies to this epic; none is fabricated here to fill a gap the source material
does not itself support.
