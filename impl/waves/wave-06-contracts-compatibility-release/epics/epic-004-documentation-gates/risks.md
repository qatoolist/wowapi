---
id: W06-E04-RISKS
type: epic-risks
epic: W06-E04
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E04 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W06-E04-001 | S002's T4 (generated-docs byte-matching AR-03's model export) cannot be completed until W05-E03 reaches `accepted`; if W05-E03 is delayed, S002 remains partially blocked | Low-medium | Low-medium — T5 (future-state labeling lint) can proceed independently of T4, so S002 is not fully blocked, only its T4 leg | Low | W06-E04-S002 | S002's own task breakdown separates T4 (AR-03-dependent) from T5 (independent), so partial progress remains possible even if T4 is delayed | If T4 remains blocked at this epic's closure attempt, record it as deferred with W05-E03's acceptance restated as the unblocking condition | unassigned | open | Low — this is a scheduling risk, not a design gap |

## Residual risk after mitigation

RISK-W06-E04-001 is expected to resolve naturally as W05-E03 lands in its own course; if it does not by
this epic's own closure attempt, the risk is tracked honestly via a partial-acceptance disposition for
S002, not silently dropped.
