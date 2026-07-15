---
id: W02-E02-RISKS
type: epic-risks
epic: W02-E02
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E02 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W02-002 originates
at wave scope and is reproduced/elaborated here because it lands entirely within this epic's S002.
One further epic-specific risk is added below.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W02-002 | The mismatch audit (S002-T3) finds real cross-tenant data — `child.tenant_id != parent.tenant_id` — on live or staging-shaped data, blocking `VALIDATE CONSTRAINT` (T5) until remediation, a decision this epic's own scope cannot make unilaterally | Low-medium — PLAN's own evidence describes the current state as structural ("nothing proves parent and child agree"), not as a confirmed-clean audit result; the audit has not yet run | High if it occurs — a confirmed cross-tenant mismatch is a live tenant-isolation breach requiring immediate remediation decision-making | High (conditional) | W02-E02-S002 | T3's acceptance criterion requires the audit against "staging/prod-shaped data" specifically so this surfaces before the wave's exit gate; the audit tool requires a platform-role connection specifically so RLS cannot mask a mismatch | If found: halt T4/T5 immediately, escalate to the acceptance authority (data/reliability lead) for a remediation decision, record in `deviations.md`, do not `VALIDATE CONSTRAINT` until a second zero-mismatch audit passes | unassigned | open | Cannot be reduced further within this epic's scope — the audit result is a fact to be discovered |
| RISK-W02-E02-002 | S002's T4/T5 (the composite-FK add and validation) are gated on W02-E01's acceptance, per `dependencies.md` — if that gate is not honored (e.g. under schedule pressure, T4/T5 are started against an incomplete or unaccepted online-migration protocol), the framework's highest-severity confirmed data-integrity fix would ship without the safety tooling PLAN's own cross-cutting note says it needs: "DATA-09 T1-T5 ahead of DATA-01 T4/T5... even though they're presented finding-by-finding" | Low (the dependency is explicit and recorded at both epic and story `depends_on` front matter, not merely in prose) | High if violated — a `NOT VALID` FK add or `VALIDATE CONSTRAINT` run without the lock-timeout budget/backfill-checkpoint/canary tooling this wave's W02-E01 provides risks exactly the maintenance-window-outage or partial-DDL failure DATA-09 exists to prevent | Medium (low likelihood, high impact) | W02-E02-S002 | The `depends_on` front-matter field on both `epic.md` and S002's `story.md` makes the gate machine-checkable, not merely documented in prose — a definition-of-ready check on S002's T4/T5 tasks should confirm W02-E01's acceptance status before those tasks move to `ready` | If T4/T5 are found to have started before W02-E01's acceptance, halt immediately, record the violation as a deviation, and re-run T4/T5 through the accepted protocol rather than treating the premature run as valid | unassigned | open | Low, contingent on the gate being honored as planned |

## Residual risk after mitigation

RISK-W02-E02-002 is expected to reduce to low residual risk as long as the `depends_on` gate is
honored — this is a process-discipline risk, not a technical uncertainty. RISK-W02-002 cannot be
pre-resolved by this epic's planning; its outcome is a fact about the actual data, discovered only
when the audit runs, and is tracked as a genuine blocking risk to S002's closure.
