---
id: W01-E01-RISKS
type: epic-risks
epic: W01-E01
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E01 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W01-001 originates
at wave scope and is reproduced/elaborated here because it lands entirely within this epic's stories.
Two further epic-specific risks are added below, scoped narrowly to this epic's own stories.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W01-001 | Judged-linter enablement (gosec/errorlint/exhaustive/forcetypeassert) surfaces more hits at this epic's actual start commit than MATRIX CS-23's snapshot recorded, because the codebase has moved since the matrix pass | Medium | Medium — increases S002's scope beyond the planned triage list | Medium | W01-E01-S002 | Re-run the full analyzer set fresh at story start rather than trusting the matrix's cached counts; treat any new hit as an in-scope triage item, not a surprise that blocks the story | If new-hit volume threatens story boundedness (mandate §12), split into a follow-up task rather than silently expanding scope | unassigned | open | Low-medium |
| RISK-W01-E01-002 | S001's zero-cost-linter enablement surfaces a hit that MATRIX CS-23's zero-hit snapshot did not record, because the 26-site inventory it cites was captured against a prior commit | Low | Low-medium — a genuine leak site would need a real code fix, not just a config flip, expanding S001 beyond its stated 4-task scope | Low | W01-E01-S001 | S001's own plan explicitly re-runs sqlclosecheck/rowserrcheck fresh (not trusting the cited snapshot) as its own fail-first verification step, per the mandate's re-confirmation requirement | If a genuine new hit is found, treat it as a 5th task under S001 rather than silently absorbing it into an existing task's scope, and flag it in `deviations.md` | unassigned | open | Low |
| RISK-W01-E01-003 | S003's license-scanning-signal choice (Trivy license scanner vs. `go-licenses`) is made without re-confirming `security-scan.yml`'s exact current line numbers and dependency-review's exact gating state, since the repository visibility changed (public since 2026-07-03) between when MATRIX CS-23 was written and when this story executes | Low | Low — a wrong assumption about dependency-review's current gating state could lead to a redundant or misapplied license signal | Low | W01-E01-S003 | S003's task explicitly re-confirms the exact `security-scan.yml` line numbers and dependency-review's `license-check: true` gating state at implementation time rather than assuming the matrix's citation is unchanged | If the visibility-dormancy premise no longer holds (e.g. dependency-review is now active), document the corrected premise in `deviations.md` and adjust the license-signal choice accordingly, not silently proceed on a stale assumption | unassigned | open | Low |

## Residual risk after mitigation

All three risks reduce to Low or Low-medium residual risk once each story's own fail-first
re-confirmation step (rather than trusting a cited snapshot) is executed as planned. None of these
risks is expected to block this epic's closure; they are tracked to ensure the re-confirmation
actually happens rather than being silently skipped.
