---
id: W03-E02-RISKS
type: epic-risks
epic: W03-E02
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E02 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W03-E02-001 | T4's JWKS-client governance gate rejects a wowsociety production deployment that currently injects a custom JWKS client with no declaration path — PLAN's own "Genuine evidence gap flagged, not papered over: wowsociety's actual deployment config for allowlist entries or custom JWKS-client injection was not read in this pass" | Low (unconfirmed, not assumed present) | Medium — if present, a `prod`-profile wowsociety deployment would fail readiness until the trusted-issuer config is declared | Low-medium | W03-E02-S001 | Re-confirm wowsociety's actual JWKS-client construction pattern at this story's own execution time via a fresh config audit, rather than trusting PLAN's own flagged gap as either "safe" or "unsafe" | If a custom JWKS client with no declaration path is found, coordinate the declaration change with wowsociety before flipping the `prod`-profile gate to enforced | unassigned | open | Low once the fresh audit is performed |
| RISK-W03-E02-002 | T1's fingerprint-scope confirmation finds `SharedFingerprint()` does *not* already cover the outbound allowlist fields, contrary to PLAN's own "likely already covers these fields structurally, pending a direct scope-confirmation test" expectation | Low | Low — if the coverage gap is found, T1 becomes a genuine extension task rather than "add a regression test only," a larger but still bounded scope | Low | W03-E02-S001 | T1's own task explicitly performs the scope-confirmation test rather than assuming coverage | If a gap is found, extend `SharedFingerprint()`'s scope as part of T1 rather than treating it as an unplanned surprise | unassigned | open | Low |

## Residual risk after mitigation

Both risks reduce to Low residual risk once their respective re-confirmation steps are executed as
planned; neither is expected to block this epic's closure.
