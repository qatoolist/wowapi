---
id: W03-E05-S001
type: story
title: Workflow privileged completion — ratification and durable override audit
status: accepted
wave: W03
epic: W03-E05
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - SEC-02
depends_on:
  - W03-E01-S001
blocks: []
acceptance_criteria:
  - AC-W03-E05-S001-01
  - AC-W03-E05-S001-02
  - AC-W03-E05-S001-03
artifacts: []
evidence: []
decisions: []
risks: []
---

# W03-E05-S001 — Workflow privileged completion — ratification and durable override audit

## Story ID

W03-E05-S001

## Title

Workflow privileged completion — ratification and durable override audit

## Objective

Implement ratification as a real definition field and state transition (override-then-ratify happy
path; pending-not-yet-effective; rejection reverts), **or** explicitly reject `ratify_by`-declaring
definitions as an interim, Wave-0-compatible posture — the directive's own "reject or implement"
allowance, a bounded design decision this story makes and records. Persist actor, impersonator, grant
ID (from W03-E01-S001), source/target states, reason, and ratification outcome in a durable audit
record, written in the same transaction as the state jump; an audit-write failure rolls back the
override. This is PLAN §5.2 SEC-02, T4 and T5.

## Value to the framework

SEC-02's Wave-0 slice (T1-T3, already executed and verified) closed the most acute fail-open
risk — a nil evaluator silently skipping the permission check entirely. This story closes the two
remaining gaps that keep a privileged override from being a fully governed, auditable operation.
Ratification is, per PLAN's own words, "a bare `TODO` comment with zero implementation" today —
this story either builds it as a real, bounded state machine, or explicitly documents why an interim
reject posture is the correct Wave-0-compatible choice instead. Durable audit is, today, absent
entirely for override operations — this story makes every override produce a complete, transactional
audit record, with the audit-write-failure-rolls-back-the-override behavior as its most
safety-critical guarantee: a privileged override that cannot be durably audited must not be allowed
to take effect.

## Problem statement

`requirement-inventory.md` row SEC-02 records: "Workflow privileged ops fail closed" — class IMPL,
priority P0, disposition `partial` ("T1–T3 EXECUTED (verified ×2); T4 ratification design + T5 audit
remain"), target `W03-E05-S001`. PLAN §5.2's own SEC-02 task table, T4 and T5 quoted exactly:

- "T4. Implement ratification as a real definition field + state transition (or explicitly reject
  `ratify_by`-declaring definitions as an interim Wave-0-compatible posture) | T1-T3 | Directive
  allows 'reject or implement' | Override-then-ratify happy path; pending-not-yet-effective;
  rejection reverts | `SEC-02/ratification-tests.md` | Genuinely greenfield design work | **Wave 2+,
  NOT Wave 0**."
- "T5. Persist actor, impersonator, grant ID (from SEC-01 T1), source/target states, reason,
  ratification outcome in durable audit | T1, T3, T4; benefits from SEC-01 T1 | Complete audit row in
  the same tx as the state jump; audit failure rolls back the override | Test: audit present/complete;
  write failure rolls back | `SEC-02/override-audit-tests.md` | Actor/reason/state portion is
  Wave-0-compatible; grant-ID field waits on SEC-01 | **Split across waves**."

Per `impl/analysis/wave-allocation-detail.md`'s own grouping instruction: "S001 T4 ratification
design+implement (or documented reject-interim posture) + T5 durable audit (grant-ID field dep on
E01 S001)."

## Source requirements

SEC-02 (T4, T5). Cross-referenced: T1–T3 (already executed and verified in Wave 0), for continuity
context only — not re-implemented or re-planned by this story.

## Current-state assessment

Per PLAN §5.2's own evidence and `requirement-inventory.md`'s disposition (to be re-confirmed at
this story's own execution commit, consistent with this programme's fail-first re-confirmation
convention):

- T1 (mandatory evaluator nil-guard), T2 (test-only constructor for nil-evaluator test call sites),
  and T3 (unconditional `Override` permission check) are already executed and independently reviewed
  in Wave 0 — this story does not re-implement them, though its own independent-review task confirms
  they remain intact (no regression).
- Ratification today is, per PLAN's own words, "a bare `TODO` comment with zero implementation" — no
  real definition field, no state transition exists.
- No durable audit record exists today for override operations — actor, impersonator, grant ID,
  source/target states, reason, and ratification outcome are not persisted anywhere in a transactional
  audit row.
- The grant-ID field this story's T5 must persist depends on W03-E01-S001's `identity_grant` table
  existing with a stable grant-ID shape — to be re-confirmed at this story's own start commit.

## Desired state

Either: ratification is implemented as a real definition field and state transition, covering three
proven cases (override-then-ratify happy path; pending-not-yet-effective; rejection reverts); or:
`ratify_by`-declaring definitions are explicitly rejected with a documented interim posture, proven by
a test that such a definition is rejected at the appropriate boundary. Whichever is chosen, the
decision and its rationale are recorded in this story's own `story.md`/`plan.md`. Every override
produces a complete audit row (actor, impersonator, grant ID, source/target states, reason,
ratification outcome) written in the same transaction as the state jump; a fault-injection test
proves an audit-write failure rolls back the override, leaving zero effect from the attempted
override.

## Scope

- T4 — implement ratification as a real definition field and state transition (three named cases), or
  explicitly reject `ratify_by`-declaring definitions as an interim, Wave-0-compatible posture. The
  choice is this story's own bounded design decision, per the directive's explicit "reject or
  implement" allowance.
- T5 — persist actor, impersonator, grant ID (from W03-E01-S001), source/target states, reason, and
  ratification outcome in a durable audit record, written in the same transaction as the state jump;
  audit-write failure rolls back the override.

## Out of scope

- T1, T2, T3 — already executed and verified in Wave 0; not re-implemented or re-planned by this
  story. This story's independent-review task confirms they remain intact as part of its own closure,
  but does not modify them.
- SEC-01 itself (W03-E01) — a dependency for T5's grant-ID field specifically, not this story's own
  implementation scope.
- Any change to `Override`'s permission-check logic beyond what T1-T3 already established — this
  story adds ratification and audit, it does not revisit the fail-closed permission check itself.

## Assumptions

- **T4's "reject or implement" choice is a bounded design decision this story makes and records, not
  an ambiguous architecture decision requiring escalation under mandate §18** — the directive
  explicitly sanctions either resolution as valid. This document does not pre-judge which choice is
  correct; the choice and its rationale belong in this story's own `plan.md`, made at implementation
  time against the actual scope pressure the "implement" path would create (per RISK-W03-E05-001,
  epic-level `risks.md`).
- **T5's grant-ID field depends on W03-E01-S001's `identity_grant` table having a stable shape.** T4
  itself has no dependency on W03-E01 and can proceed independently if sequencing requires — per
  PLAN's own T5 Depends-on column ("T1, T3, T4; benefits from SEC-01 T1"), the grant-ID field is
  specifically a T5 concern, not a T4 one.
- If the "implement" path is chosen for T4, the state machine is bounded to exactly the three named
  states (override-then-ratify happy path; pending-not-yet-effective; rejection reverts) per PLAN's
  own acceptance-criteria wording — this document does not assume a broader ratification framework is
  in scope.

## Dependencies

**Hard, for T5's grant-ID field only, not for T4**: W03-E01-S001's `identity_grant` table must exist
and have a stable grant-ID shape before T5's audit record can attribute a grant ID. T4 (ratification)
has no dependency on W03-E01 and may proceed independently of it if sequencing requires — see
epic-level `dependencies.md`.

## Affected packages or components

`kernel/workflow/runtime.go` (`Override`, and wherever ratification's state machine — if the
"implement" path is chosen — and the audit write are implemented).

## Compatibility considerations

Per PLAN's own wowsociety-impact note for SEC-02: "Not affected. Zero occurrences of
`workflow.NewRuntime`, `workflow.Runtime`, `.Override(`, or any `kernel/workflow` import anywhere in
wowsociety. No required changes, no sequencing constraint." This story carries zero wowsociety
compatibility risk, unlike W03-E01/E03 — to be re-confirmed fresh at this story's own execution time
per this programme's convention, not merely trusted from the cited snapshot, though the risk of
finding an unexpected consumer is materially lower here given the zero-occurrence citation covers
every relevant symbol (`NewRuntime`, `Runtime`, `.Override(`, the import itself).

## Security considerations

T5's audit-write-failure-rolls-back-the-override requirement is this story's most safety-critical
guarantee: a privileged override that cannot be durably audited must not be allowed to take effect —
a fail-closed posture consistent with the rest of this wave's security findings. Whichever T4 path is
chosen, the resulting behavior must not create a new fail-open surface: the "reject" path must
actually reject (not silently accept and ignore `ratify_by`), and the "implement" path's three named
states must not introduce a bypass (e.g. a pending-not-yet-effective override must not be treated as
already effective).

## Performance considerations

T5's audit write, occurring in the same transaction as the state jump, adds a database write to every
override operation that did not previously incur one for audit purposes. This is an accepted,
required cost of closing the audit gap — not separately optimized in this story unless
implementation surfaces an unacceptable latency regression, in which case it is recorded as a
finding, not silently absorbed.

## Observability considerations

T5's durable audit record is itself the primary observability deliverable of this story for override
operations. No further observability scope beyond what a reasonable implementation would emit for a
security-critical audit write (e.g. a metric for audit-write failures triggering rollback) is
required scope, left as an implementation-time judgment call.

## Migration considerations

If T4's "implement" path is chosen, a real ratification definition field requires a schema change
(a new column or table for the state machine's persisted state) — routed through W02-E01's DATA-09
protocol per this wave's entry criteria. T5's durable audit record requires a new or extended audit
table — also routed through DATA-09. If T4's "reject" path is chosen instead, no ratification-related
schema change is needed; only T5's audit-table change applies. Exact schema shape for either path is
not invented here — determined at implementation time against the chosen T4 path and the existing
`kernel/audit` conventions.

## Documentation requirements

Document whichever T4 path is chosen (the real ratification state machine's three named states, or
the interim reject posture and its rejection boundary) and the T5 durable-audit record's schema and
transactional guarantee (including the audit-write-failure-rolls-back-the-override behavior) in
whatever documentation currently covers the `kernel/workflow` module.

## Acceptance criteria

- **AC-W03-E05-S001-01**: Ratification is either implemented as a real definition field and state
  transition, proven by three tests (override-then-ratify happy path; pending-not-yet-effective;
  rejection reverts); or `ratify_by`-declaring definitions are explicitly rejected with a documented
  interim posture, proven by a test that a `ratify_by`-declaring definition is rejected at the
  appropriate boundary. The choice made and its rationale are recorded in `story.md`/`plan.md`.
- **AC-W03-E05-S001-02**: Every override produces a complete audit row (actor, impersonator, grant
  ID, source/target states, reason, ratification outcome) written in the same transaction as the
  state jump; a fault-injection test proves an injected audit-write failure rolls back the override,
  leaving zero effect from the attempted override.
- **AC-W03-E05-S001-03**: T1–T3's already-executed and verified fail-closed behavior (mandatory
  evaluator; no public API accepting a nil `authz.Evaluator`; unconditional `Override` permission
  check) remains intact, confirmed by re-running or re-reviewing their existing test coverage as part
  of this story's own review, not assumed unchanged.

## Required artifacts

- The ratification implementation (state machine, if "implement" is chosen) or the interim-reject
  implementation (if "reject" is chosen), plus the design-decision record.
- The durable override-audit-record implementation (schema + transactional write + rollback-on-
  failure logic).
See `artifacts/index.md`.

## Required evidence

- Ratification test output (three named cases, if "implement"; the rejection-boundary test, if
  "reject").
- Audit-present/complete test output; fault-injection test output proving audit-write failure rolls
  back the override.
- T1–T3 regression confirmation output (existing test suite re-run or re-reviewed).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies recorded
(`depends_on: [W03-E01-S001]` for T5's grant-ID field specifically), owner/reviewer assignment
pending, the T4 reject-vs-implement choice explicitly recorded (not left as the current `TODO`).

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, with specific confirmation the audit-write-failure-rollback fault-injection
test is genuinely adversarial, not a happy-path test relabeled.

## Risks

RISK-W03-E05-001 (T4's "implement" path risks materially expanding beyond a bounded task if not
scoped tightly) and RISK-W03-E05-002 (T5's audit-write-failure-rollback behavior is safety-critical;
an implementation bug could silently allow an unaudited override or incorrectly roll back a
legitimate one) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

RISK-W03-E05-001 is expected to reduce to low-medium residual risk once the state machine (if
"implement" is chosen) is bounded to exactly the three named states, per `plan.md`'s own scoping.
RISK-W03-E05-002 is expected to reduce to low residual risk once the fault-injection test is proven
to genuinely exercise the audit-write-failure path, not merely assert the happy path.

## Plan

See `plan.md`.

## Decision record

**T4 path chosen: reject.** At implementation time, the "implement" path for ratification (a real
state machine with override-then-ratify happy path, pending-not-yet-effective, and rejection reverts)
was scoped and found to require new persisted instance state, a ratification task lifecycle, and
role-based assignment plumbing that would expand this single story well beyond its Wave-0 security
remit. Per the directive's explicit "reject or implement" allowance, PLAN's RISK-W03-E05-001, and
the programme's bounded-scope mandate, this story instead adds a parsed-but-rejected `ratify_by`
field to `workflow.Definition` and `workflow.Step` and rejects any definition declaring it at
`Validate` time with a clear, fail-closed error. The durable override audit record (T5) documents
this interim posture by recording `ratification_outcome: rejected_interim` for every override. A
future story may implement the real ratification state machine when schedule and scope allow.
