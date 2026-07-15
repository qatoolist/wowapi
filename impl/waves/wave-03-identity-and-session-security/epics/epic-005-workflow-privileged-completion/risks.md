---
id: W03-E05-RISKS
type: epic-risks
epic: W03-E05
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E05 — Risks

No wave-level risk entry names this epic directly. Two epic-specific risks are identified below,
scoped to this epic's own story.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W03-E05-001 | T4's "implement" path (a real ratification state machine) is, per the directive's own framing, "genuinely greenfield design work" — if chosen, this epic's story risks materially expanding beyond a bounded task if the state-machine design is not scoped tightly | Medium | Medium — could threaten story boundedness (mandate §12) if scope isn't controlled | Medium | W03-E05-S001 | The story's `plan.md` should record the reject-vs-implement decision explicitly and, if "implement" is chosen, bound the state machine to exactly the three named states (override-then-ratify happy path; pending-not-yet-effective; rejection reverts) per PLAN's own acceptance-criteria wording, not a broader ratification framework | If the "implement" path is found to be growing beyond the three named states during implementation, split further work into a follow-up story rather than silently expanding this one | unassigned | open | Low-medium once the scope is bounded as planned |
| RISK-W03-E05-002 | T5's audit-write-failure-rolls-back-the-override behavior is safety-critical (a privileged override must not take effect without a durable audit record) — an implementation bug here could either silently allow an unaudited override (worse than today's state) or incorrectly roll back a legitimate override on a transient audit-write blip | Low | High — this is exactly the kind of silent-failure risk mandate §14's independent review is designed to catch | Medium | W03-E05-S001 | T5's own fault-injection test explicitly proves the rollback behavior under an injected audit-write failure, not merely asserted in prose; independent review specifically checks this test exists and is genuinely adversarial | If the fault-injection test reveals the rollback does not work as intended, this is a blocking finding for the story's closure, not a deferred follow-up | unassigned | open | Low once the fault-injection test is proven to genuinely exercise the failure path |

## Residual risk after mitigation

Both risks reduce to Low/Low-medium residual risk once their respective scope-bounding (T4) and
fault-injection proof (T5) steps are executed as planned.
