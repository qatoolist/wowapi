---
id: W05-E01-RISKS
type: epic-risks
epic: W05-E01
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W05-001 and
RISK-W05-002 originate at wave scope and are reproduced/elaborated here because they land entirely
within this epic's S002.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W05-001 | S002's T5 (owner-bound registrar wrapper for `authz.Registry` permission registration) closes the framework's only zero-ownership-check registration surface — PLAN's own risk column: "High — only registry with zero existing ownership check," "the actual security boundary" | Medium | High — an incomplete fix leaves permission registration open to cross-module forgery | High | W05-E01-S002 | T5's acceptance criterion requires the named adversarial test (`AR-01/authz_ownership_adversarial_test.go`); S002 adds an independent-review task scoped specifically to T5 | Block story acceptance if the adversarial test surfaces a bypass | unassigned | open | Low once genuinely proven |
| RISK-W05-002 | S002's T6 (owner-bound wrappers for ~9+ remaining declaration classes) carries PLAN's own explicit risk note: "easy to under-scope" | Medium-high | Medium — an under-scoped T6 leaves a subset of declaration classes unowned | Medium | W05-E01-S002 | T6's task record requires a table-driven adversarial suite enumerating every declaration class explicitly, checked against AR-01's own acceptance-gate class list at review time | Treat a missing class found at review as a task-scope correction in `deviations.md`, not a silent narrowing | unassigned | open | Low once the enumeration is explicit and reviewed |
| RISK-W05-E01-003 | S004's legacy adapter (T11) is itself a trust boundary — PLAN's own risk note: "the adapter is itself a trust boundary" — an adapter that derives owner incorrectly, or that bypasses T2-T6's ownership checks for convenience, would silently reintroduce the exact unowned-registration gap this epic exists to close, just hidden behind a compatibility shim | Low-medium | High if it occurs — a bypassing adapter defeats the entire epic's security property for every module that boots through it (i.e. every existing module, until each migrates off the legacy path) | Medium-high | W05-E01-S004 | T11's acceptance criterion explicitly requires the adapter to derive owner from `Module.Name()` and route through the same owner-bound registrars as the non-legacy path — "it must not bypass T2-T6"; S004's integration test asserts this by running the adversarial fixtures from T2-T6 through the legacy path, not only the non-legacy path | If the legacy path is found to bypass an ownership check, treat as a blocking defect, not a documented limitation | unassigned | open | Low once the adversarial fixtures are confirmed to run identically through both paths |

## Residual risk after mitigation

RISK-W05-001 and RISK-W05-E01-003 are expected to reduce to low residual risk once their respective
adversarial proofs are genuinely executed — both are exactly why this epic's independent-review
tasks exist. RISK-W05-002 reduces to low once T6's declaration-class enumeration is explicit and
reviewed against AR-01's own acceptance-gate class list.
