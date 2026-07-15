---
id: W07-E01-RISKS
type: epic-risks
epic: W07-E01
wave: W07
status: current
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W07-001 originates at
wave scope and is reproduced/elaborated here because it lands entirely within this epic's four stories.
One further epic-specific risk is added below.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W07-001 | DEC-Q9 remains unresolved after this epic's closure, leaving every absolute-SLO acceptance criterion across all 4 stories conditional | Occurred/open | Medium — does not block accepted relative/container evidence, but prevents an absolute performance guarantee | Medium | W07-E01 (all 4 stories) | Every absolute-SLO claim is explicitly conditional on DEC-Q9; the provisional Linux/amd64 container/GH-runner comparison policy is active | Carry the open decision into W07-E04-S002's claim-upgrade package and rerun the named benchmarks when a human-owned reference environment is approved | human infra/programme owner | open | Irreducible here: no approved owner, timeline, or dedicated reference environment |
| RISK-W07-E01-001 | W04's DATA-02/DATA-03 lease primitives might not fit PERF-04's batching and throughput requirements without adaptation | Did not occur | Medium if realized | Low-medium | W07-E01-S003 | S003 directly consumed the accepted W04 lease/fencing primitives and verified claim commit boundaries, fencing, duplicate-worker safety, and ordering | Reopen only if later evidence invalidates the accepted lease contract | performance/SRE lead | mitigated/closed | Low: external handler effects intentionally remain at-least-once |

## Residual risk after mitigation

RISK-W07-E01-001 did not realize and is closed by S003's accepted lease/chaos evidence.
RISK-W07-001 remains the exact open residual risk: DEC-Q9 has no approved owner, timeline, or dedicated
reference-performance environment, so this epic accepts relative/container evidence only and makes no
absolute latency, throughput, or SLO claim.
