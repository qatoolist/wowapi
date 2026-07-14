---
id: W03-E05-ACCEPTANCE
type: epic-acceptance
epic: W03-E05
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E05 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" as a standalone,
independently-referenceable record, consistent with the wave-level `../../acceptance.md` pattern
(AC-W03-08 there maps onto this epic).

## AC-W03-E05-01 — Ratification implemented or interim posture documented

Either: ratification is implemented as a real definition field and state transition, proven by
three tests (override-then-ratify happy path; pending-not-yet-effective; rejection reverts); or:
`ratify_by`-declaring definitions are explicitly rejected with a documented interim posture,
proven by a test that a `ratify_by`-declaring definition is rejected at the appropriate boundary
(definition-registration time or override time, to be determined by the story's own design
decision). The choice made and its rationale are recorded in `story.md`/`plan.md`.

## AC-W03-E05-02 — Durable, complete override audit

Every override produces a complete audit row — actor, impersonator, grant ID (from
W03-E01-S001), source/target states, reason, ratification outcome — written in the same
transaction as the state jump. A fault-injection test proves an injected audit-write failure rolls
back the override, leaving zero effect from the attempted override.

## AC-W03-E05-03 — No regression to Wave-0 fail-closed behavior

T1–T3's already-executed and verified fail-closed behavior (mandatory evaluator; no public API
accepting a nil `authz.Evaluator`; unconditional `Override` permission check) remains intact,
confirmed by re-running or re-reviewing their existing test coverage as part of this epic's own
review, not assumed unchanged.

## AC-W03-E05-04 — Independent review passed

W03-E05-S001 has passed independent review per mandate §14, with specific confirmation the
audit-write-failure-rollback fault-injection test (AC-W03-E05-02) is genuinely adversarial, not a
happy-path test relabeled.

## Acceptance authority

Product-security lead (PLAN §5.2's stated accountable role for PF-SEC).
