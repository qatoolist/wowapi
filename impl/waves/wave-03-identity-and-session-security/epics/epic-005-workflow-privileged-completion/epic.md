---
id: W03-E05
type: epic
title: Workflow privileged completion
status: planned
wave: W03
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - SEC-02
depends_on:
  - W03-E01
stories:
  - W03-E05-S001
decisions: []
risks: []
---

# W03-E05 — Workflow privileged completion

## Epic objective

Close SEC-02's two remaining tasks after its Wave-0 fail-closed slice (T1–T3, already executed):
implement ratification as a real definition field and state transition, or explicitly document an
interim reject posture for `ratify_by`-declaring definitions; and persist actor, impersonator,
grant ID, source/target states, reason, and ratification outcome in a durable audit record written
in the same transaction as the state jump. This is PLAN §5.2's SEC-02, T4 and T5.

## Problem being solved

`requirement-inventory.md` row SEC-02 records: "Workflow privileged ops fail closed" — class IMPL,
priority P0, disposition `partial` ("T1–T3 EXECUTED (verified ×2); T4 ratification design + T5
audit remain"), target `W03-E05-S001`. PLAN §5.2's own evidence, with what it calls "a materially
important blast-radius correction": `NewRuntime`'s nil-guard, `Override`'s unconditional permission
check, and ratification's implementation status. T1–T3 (mandatory evaluator, test-only constructor,
unconditional `Override` check) were Wave-0 items and are already executed and verified per the
requirement inventory. This epic's scope is exactly T4 (ratification, "a bare `TODO` comment with
zero implementation" today) and T5 (durable audit, which benefits from but per PLAN's own dependency
notation does not strictly require SEC-01's grant-ID field — though this wave's story-allocation
explicitly grounds T5's grant-ID attribution in W03-E01-S001's output).

## Scope

- T4 — implement ratification as a real definition field and state transition (override-then-ratify
  happy path; pending-not-yet-effective; rejection reverts), **or** explicitly reject
  `ratify_by`-declaring definitions as an interim, Wave-0-compatible posture, per the directive's own
  "reject or implement" allowance.
- T5 — persist actor, impersonator, grant ID (from W03-E01-S001), source/target states, reason, and
  ratification outcome in a durable audit record, written in the same transaction as the state jump;
  audit-write failure rolls back the override.

## Out of scope

- T1, T2, T3 — already executed and verified in Wave 0 (per `requirement-inventory.md`); this epic
  does not re-implement or re-plan them, though its independent-review task should confirm they
  remain intact (no regression) as part of this epic's own closure.
- SEC-01 itself (W03-E01) — a dependency for T5's grant-ID field, not this epic's own implementation
  scope.
- Any change to `Override`'s permission-check logic beyond what T1-T3 already established — this
  epic adds ratification and audit, it does not revisit the fail-closed permission check itself.

## Source requirements

SEC-02 (T4, T5). Cross-referenced: T1–T3 (already executed, Wave 0), for continuity context only.

## Architectural context

SEC-02's Wave-0 slice closed the most acute fail-open risk (a nil evaluator silently skipping the
permission check entirely). This epic closes the two remaining gaps that keep a privileged override
from being a fully governed, auditable operation: ratification (today literally a `TODO` comment —
"zero implementation") and durable audit (today absent entirely for override operations). T5's
audit-write-failure-rolls-back-the-override requirement is the epic's most safety-critical
acceptance criterion — it means a privileged override that cannot be durably audited must not be
allowed to take effect, a fail-closed posture consistent with the rest of this wave's security
findings.

The affected layer is `kernel/workflow/runtime.go` (`Override`, and wherever ratification's state
machine and the audit write are implemented).

## Included stories

- **W03-E05-S001 — workflow-privileged-completion** (SEC-02 T4 ratification design+implement or
  documented reject-interim posture, plus T5 durable audit — single story per
  `impl/analysis/wave-allocation-detail.md`: "S001 T4 ratification design+implement (or documented
  reject-interim posture) + T5 durable audit (grant-ID field dep on E01 S001)").

## Dependencies

**Hard** (for T5's grant-ID field specifically, not for T4): W03-E01-S001's `identity_grant` table
must exist and have a stable grant-ID shape before T5's audit record can attribute a grant ID. T4
(ratification) has no dependency on W03-E01 and can proceed independently of it if sequencing
requires. See `dependencies.md`.

## Risks

No epic-specific risk beyond what is captured in `risks.md` — the "reject vs. implement" choice for
T4 itself carries design risk (genuinely greenfield state-machine work if "implement" is chosen),
elaborated at story scope.

## Required decisions

None new at the epic level. T4's own "reject or implement" choice is a story-level design decision
this epic's single story makes and records (in `story.md`/`plan.md`), not a program-level ADR — the
directive explicitly permits either resolution as valid, so this is not an "ambiguous architecture
decision" requiring escalation under mandate §18, it is a bounded implementation choice with two
directive-sanctioned outcomes.

## Epic acceptance criteria

- **AC-W03-E05-01**: Ratification is either implemented as a real definition field and state
  transition (override-then-ratify happy path, pending-not-yet-effective, rejection reverts) or
  `ratify_by`-declaring definitions are explicitly rejected with a documented interim posture — one
  of the two, chosen and recorded, not left as the current `TODO`.
- **AC-W03-E05-02**: Every override produces a complete audit row (actor, impersonator, grant ID,
  source/target states, reason, ratification outcome) in the same transaction as the state jump; a
  fault-injection test proves an audit-write failure rolls back the override.
- **AC-W03-E05-03**: T1–T3's already-executed fail-closed behavior remains intact (no regression),
  confirmed as part of this epic's own review, not merely assumed.
- **AC-W03-E05-04**: The story has passed independent review per mandate §14.

## Closure conditions

W03-E05-S001 reaches `accepted`; AC-W03-E05-01 through AC-W03-E05-04 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date.
