---
id: W06-E03-RISKS
type: epic-risks
epic: W06-E03
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E03 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W06-001 originates at
wave scope and is reproduced/elaborated here because it lands entirely within this epic's S002. One
further epic-specific risk is added below.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W06-001 | S002 cannot enter `ready`/`in-progress` until DEC-Q10 (branch protection, protected release Environment, tag protection ruleset) is resolved by a human with repo-admin access | High (confirmed today via `gh api` per REVIEW's own evidence) | Medium — does not block S001's own buildable-now work | Medium | W06-E03-S002 | S001's own scope is deliberately split from S002 so the ~85% buildable/testable work is not held hostage | Track DEC-Q10 as an explicit, separately-ticketed item (REVIEW's own recommendation: "PF-REL-ADMIN-01") | unassigned | open | Cannot be reduced further — genuine human-administration dependency |
| RISK-W06-E03-001 | S001's T6 (`build-candidate`/`publish` split via GoReleaser `--skip=publish`) may discover, at implementation time, that the pinned GoReleaser version does not behave as ADR-005 assumed — ADR-005's own "Consequences" section flags this as an unresolved caveat | Low-medium | Medium — if the split-mode support does not work as assumed, T6's implementation strategy would need to change, potentially requiring a hand-rolled pipeline the ADR explicitly rejected | Low-medium | W06-E03-S001 | T6's own task record requires the version-confirmation step (checking the pinned GoReleaser version's own documentation/changelog) before implementation is trusted as correct, per ADR-005's own stated responsibility | If the pinned version does not support the assumed behavior, record this as a deviation from ADR-005 and escalate to the release/security-engineering lead rather than silently hand-rolling a substitute | unassigned | open | Low once the version-confirmation step is executed as planned |

## Residual risk after mitigation

RISK-W06-E03-001 is expected to reduce to low residual risk once T6's version-confirmation step is
executed as planned. RISK-W06-001 cannot be resolved within this epic's own execution capacity — it is
a genuine, irreducible-within-this-epic human-administration dependency, tracked but not eliminated.
