---
id: W02-E03-RISKS
type: epic-risks
epic: W02-E03
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E03 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W02-E03-001 | The locked parent counter or dedicated per-aggregate sequence row that replaces `MAX(version)+1` becomes the new serialization point for concurrent version allocation — PLAN DATA-05 T1's own risk note states this explicitly: "Counter-row contention is the new serialization point — measure lock wait" | Medium | Medium — under high concurrent-write load against the same aggregate, callers now serialize on the counter row rather than racing on a read; if lock-wait grows unacceptably under realistic load, the mechanism needs revisiting before it can be considered closed | Medium | W02-E03-S001 (T1, T5) | T1's own concurrency test (≥20 concurrent callers per PLAN's "Tests" column) is required to measure and record lock-wait, not merely prove correctness — this epic's `plan.md` treats "no unexpected conflicts" and "acceptable lock-wait under the tested concurrency level" as two distinct things to evidence, not one | If measured lock-wait is unacceptable at the tested concurrency level, record it as an accepted residual risk with the measured figures (not silently smoothed over), and treat further optimization (e.g. sharding the counter, or a non-locking sequence primitive) as a follow-up item rather than blocking this story's own correctness acceptance criterion | unassigned | open | Medium until lock-wait is actually measured; expected to reduce once T1's concurrency test runs and produces real numbers |

## Residual risk after mitigation

RISK-W02-E03-001 cannot be resolved to zero within this epic's own scope — some amount of
serialization on the counter row is an inherent trade-off of moving from an unsafe `MAX()+1` read to
a safe locked counter, not a defect to eliminate. The epic's closure requires the risk to be
*measured and recorded*, not eliminated; residual risk is expected to remain open at closure unless
the measured lock-wait is negligible at the tested concurrency level, in which case it may be closed
as "measured, acceptable" rather than "mitigated."
