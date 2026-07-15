---
id: W07-E04-RISKS
type: epic-risks
epic: W07-E04
wave: W07
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07-E04 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W07-003 originates at
wave scope and is reproduced/elaborated here because it lands entirely within this epic's two stories.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W07-003 | The final gate re-run (S001) discovers an unresolved gap in an earlier wave's own closure that this epic cannot itself fix without reopening that wave's own scope | Medium | Medium-high — a genuine gap found this late in the programme has no clean remediation path within this epic's own story boundaries | Medium-high | W07-E04-S001, W07-E04-S002 | The gate re-run is explicitly designed to surface exactly this class of finding — finding a real gap here is the gate working correctly | Any gap found is recorded in the disposition audit and carried into the closure/claim-upgrade decision package (S002) as an explicit open item for the human authority | unassigned | open | Cannot be pre-resolved — this risk's entire nature is a fact to be discovered |
| RISK-W07-E04-001 | S001's own gate re-run is performed as a restatement of REVIEW's own original 2026-07-11 conclusions rather than a genuine fresh re-verification against current HEAD, undermining the entire purpose of "re-running" the gate | Low-medium | High if it occurs — a rubber-stamped closure gate provides false confidence exactly where the programme most needs honest verification | Medium-high | W07-E04-S001 | S001's own task record requires evidence of genuine re-verification (e.g. actually re-checking capability-matrix rows against current code, not merely re-printing REVIEW's own §H/§I tables) — this is itself checked by S001's own independent-review task | If S001's own output is found to be a restatement rather than a re-verification, reject it and require genuine re-verification before this epic can close | unassigned | open | Low once S001's own independent-review task honors its explicit re-verification-not-restatement check |

## Residual risk after mitigation

RISK-W07-E04-001 is expected to reduce to low residual risk once S001's own independent-review task
genuinely confirms fresh re-verification occurred. RISK-W07-003 cannot be pre-resolved by this epic's own
planning — its outcome is a fact to be discovered, tracked honestly (via the claim-upgrade decision
package) rather than silently assumed away.
