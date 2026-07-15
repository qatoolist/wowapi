---
id: W06-E01-RISKS
type: epic-risks
epic: W06-E01
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E01 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W06-E01-001 | DX-04 T2's subsystem-coverage requirement (resource, rule, workflow, event handler, recurring job, document flow, notification, webhook — each exercised at least once) spans a broad surface; PLAN's own risk note calls this "High — broad surface, many kernel subsystems must be generator-reachable" | Medium-high | Medium — a kernel subsystem that is not yet generator-reachable would require either extending the generator (scope creep beyond this epic's stated boundary) or narrowing the fixture's own claimed coverage | Medium | W06-E01-S002 | S002's own task breakdown scopes T2 to "exercise each subsystem at least once" without requiring exhaustive per-subsystem generator features — a minimal generated exercise of each subsystem, not a full feature-parity generator rewrite | If a subsystem proves not generator-reachable within this story's bounded scope, record the gap explicitly in `deviations.md` rather than silently narrowing the fixture's own acceptance criterion without saying so | unassigned | open | Low-medium once the minimal-exercise framing is honored |
| RISK-W06-E01-002 | DX-04 T4's upgrade-from-previous-version replay depends on DX-05's already-ratified v1/N-1 policy (W01, executed) to define what "previous supported version" means; if that policy's practical application to this specific fixture surfaces an ambiguity not anticipated at W01, T4 could stall on a policy question outside this epic's own authority to resolve | Low | Medium — a stalled T4 would delay S002's own AC-W06-E01-S002-04/05, though S002's earlier tasks (T1-T3) remain independently completable | Low-medium | W06-E01-S002 | T4's own task explicitly consumes DX-05's ratified policy rather than re-deriving it — the dependency is on an already-`accepted` piece of work, not an open question | If a genuine ambiguity surfaces, escalate to the developer-experience lead (DX-05's own accountable role) for a policy clarification, record the resolution in `deviations.md` | unassigned | open | Low — DX-05 is already closed and accepted, reducing the likelihood of a genuine gap |

## Residual risk after mitigation

Both risks are expected to reduce to low residual risk once their respective mitigations (minimal-
exercise framing for T2; consuming an already-ratified policy for T4) are honored as planned. Neither
risk is expected to block this epic's own closure if its mitigation is followed.
