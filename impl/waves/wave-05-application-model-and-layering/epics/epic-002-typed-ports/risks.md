---
id: W05-E02-RISKS
type: epic-risks
epic: W05-E02
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E02 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W05-E02-001 | T2's internal compiler factory must prove capability confusion is impossible given AR-01 and AR-02 share one `Registrar` type (D-02's own design) — PLAN's own risk column: "High — verify capability confusion is impossible if AR-01/AR-02 share one `Registrar` type" | Medium | High if it occurs — a capability-confusion bug here would defeat both AR-01's and AR-02's security boundary simultaneously, since they share the one `Registrar` type by design | High | W05-E02-S001 | T2's own acceptance criterion requires the named adversarial compile-fail fixture (`AR-02/registrar_forge_compile_fail_fixture/`); S001 adds an independent-review task specifically scoped to re-confirming this fixture genuinely fails to compile | Block story acceptance if the fixture is found to compile (i.e. capability confusion is possible) | unassigned | open | Low once genuinely proven |
| RISK-W05-E02-002 | T3's zero-hot-path-reflection claim ("naive implementations reflect per-call" per PLAN's own risk note) requires careful implementation to avoid the most natural, but non-compliant, generic-dispatch approach | Medium | Medium — a reflection-based implementation would violate AR-02's own directive requirement ("type-erasing only at compile time, never on request hot paths") without necessarily being caught by a purely functional test | Medium | W05-E02-S002 | T3's own acceptance criterion requires a benchmark AND a static lint check, not a functional test alone — the lint specifically checks for `reflect.*` calls on the hot path, catching what a passing functional test might miss | If the benchmark or lint reveals hot-path reflection, redesign the dispatch mechanism before proceeding to T4/T5 | unassigned | open | Low once both the benchmark and lint are genuinely clean |

## Residual risk after mitigation

RISK-W05-E02-001 is expected to reduce to low residual risk once T2's adversarial fixture is
genuinely proven and independently re-confirmed by S001's review task. RISK-W05-E02-002 reduces to
low once T3's benchmark and lint are both clean, not merely the functional correctness tests.
