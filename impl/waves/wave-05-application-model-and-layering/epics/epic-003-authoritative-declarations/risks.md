---
id: W05-E03-RISKS
type: epic-risks
epic: W05-E03
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E03 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W05-E03-001 | AR-03 T3's golden-declaration-delta test is, per PLAN's own framing, "this test IS the acceptance gate" — a single, high-stakes integration test with no separate fallback proof mechanism for AR-03's central claim (a manifest change deterministically produces the expected full projection diff with no other hand-edited file) | Medium — golden-delta tests are inherently fragile to any non-deterministic input creeping into either the manifest or the projection tooling | High if the test cannot be made reliably deterministic — AR-03's entire acceptance bar rests on this one test passing meaningfully, not merely passing once | High | W05-E03-S001 | S001's own task record isolates T3 as a dedicated task with its own independent-review coverage, specifically checking the golden-delta test was genuinely run (not skipped or weakened) and genuinely covers the full projection surface (route/permission/resource/schema/OpenAPI/lifecycle/profile/test/doc) named in PLAN's own acceptance criterion | If the golden-delta test proves unreliable (flaky, non-deterministic), treat as a blocking defect in the manifest/projection design itself, not a test-infrastructure inconvenience to work around | unassigned | open | Low once the test is confirmed deterministic and independently re-run by review |

## Residual risk after mitigation

RISK-W05-E03-001 is expected to reduce to low residual risk once T3's golden-delta test is confirmed
deterministic and independently re-executed by S001's own review task, per PLAN's own framing that
this test is not merely one acceptance criterion among several but the acceptance gate itself.
