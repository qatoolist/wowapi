---
id: W03-E03-RISKS
type: epic-risks
epic: W03-E03
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E03 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W03-006
originates at wave scope and lands entirely within this epic's single story.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W03-006 | A future or undiscovered custom `Verifier` implementation breaks silently at compile time only when the interface changes, since PLAN's cited "zero custom `Verifier` implementation anywhere in wowsociety" claim may have drifted since PLAN was written | Low | Low — a compile break is a safe failure mode, not a silent behavioral regression | Low | W03-E03-S001 | Re-confirm the zero-consumer claim at this story's own execution time (fresh grep), not merely trust the cited snapshot | If a consumer is found, document it in `deviations.md` and coordinate the interface change with that consumer before merge | unassigned | open | Low |

## Residual risk after mitigation

Reduces to Low residual risk once the fresh re-confirmation step is executed as planned; not
expected to block this epic's closure.
