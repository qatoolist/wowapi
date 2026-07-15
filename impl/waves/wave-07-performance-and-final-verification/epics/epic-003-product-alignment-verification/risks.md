---
id: W07-E03-RISKS
type: epic-risks
epic: W07-E03
wave: W07
status: current
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E03 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W07-E03-001 | Direct verification found two consumability failures: DATA-01 lacks the `rule_versions(tenant_id,id)` unique parent key required by PROD-01, and W03-E01-S004's PROD-04 rollout material contradicts the current SEC-01 schema, claim authority, and safe rollback model | Occurred | High — PROD-01 cannot create its FK; executing PROD-04 as written can break privileged sessions or regress the verified grant authority boundary | High | W07-E03-S001 AC01/AC04; downstream W07-E04 final gate | Keep both rows blocked; add the parent key through DATA-09; correct the SEC-01 rollout artifacts and obtain wowapi/wowsociety security sign-off; reverify here | Do not use a product-only FK workaround or direct privileged-claim fallback; defer the affected product/framework version bump or disable privileged issuance until corrected | wowapi data/reliability (PROD-01); wowsociety identity + wowapi product-security (PROD-04) | realized/open | High until fixes land; expected low after focused retest and independent review |

## Residual risk after mitigation

Residual risk remains high while either blocker is open. Passing framework tests do not compensate
for the absent referenced key or an operational plan that contradicts the current security contract.
