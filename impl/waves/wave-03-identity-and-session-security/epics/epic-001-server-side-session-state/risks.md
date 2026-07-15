---
id: W03-E01-RISKS
type: epic-risks
epic: W03-E01
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E01 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). All four risks below
originate at wave scope and land entirely within this epic's stories.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W03-001 | DEC-Q1 remains unresolved through this epic's execution | Medium | Medium | Medium | W03-E01-S001, W03-E01-S002 | Build strictly against the documented safe default (framework owns the grant record keyed by grant-ID) | Record a deviation and schedule a follow-up story if DEC-Q1 resolves to a materially different claim shape after acceptance | unassigned | open | Low |
| RISK-W03-002 | The wowsociety two-repo coordinated cutover (PROD-04) cannot be completed unilaterally by this epic | High (structural) | High | High | W03-E01-S004 | S004 produces the coordination plan only; explicitly states the two-repo requirement | Record the gap in the programme's deferred-items register if wowsociety-side coordination cannot begin promptly | unassigned | open | Medium |
| RISK-W03-004 | SEC-01 T2's "every existing valid session has a live `user_tenant_access` row" precondition may not hold against real data | Medium | High | High | W03-E01-S001 | Data-audit step against `user_tenant_access` before enabling unconditional enforcement, per PLAN's own risk note | Stage enforcement behind a profile flag if the audit finds gaps | unassigned | open | Medium |
| RISK-W03-005 | SEC-01 T4's capacity-selection requirement may break a currently-working capacity-less multi-capacity flow | Medium | Medium | Medium | W03-E01-S002 | Record as a compatibility consideration requiring product-side coordination, tracked alongside S004 | Flag to product/wowsociety as part of S004's coordination scope if an active capacity-less flow is found | unassigned | open | Medium |

## Residual risk after mitigation

RISK-W03-001 and RISK-W03-004 reduce to Low/Medium residual risk once their respective
re-confirmation and audit steps are actually executed. RISK-W03-002 and RISK-W03-005 have
irreducible structural residual risk (Medium/High) because their resolution genuinely depends on
wowsociety-side work outside this epic's unilateral control — they are tracked, not eliminated, by
this epic's own scope.
