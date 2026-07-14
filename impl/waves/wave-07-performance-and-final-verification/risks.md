---
id: W07-RISKS
type: wave-risks
wave: W07
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W07-001 | DEC-Q9 (reference-performance-environment ownership) remains unresolved past this wave's own closure, leaving every PERF-02..05 absolute-SLO acceptance criterion permanently conditional rather than eventually resolved | Medium-high (no owner/timeline established anywhere in the directive per PLAN's own cross-cutting note) | Medium — does not block this wave's own relative/container-comparison closure, but leaves the framework's own absolute performance guarantees permanently deferred if never resolved | Medium | W07-E01 (all four stories) | Every absolute-SLO acceptance criterion in this epic's stories is written explicitly as conditional on DEC-Q9, per this wave's own task-brief instruction, so the wave's own closure does not silently assume DEC-Q9 will resolve | Record DEC-Q9's continued open status honestly in this wave's own closure report and the final claim-upgrade decision package (W07-E04-S002) — the human authority reviewing that package must see this open item explicitly, not discover it was silently dropped | unassigned | open | Cannot be reduced further within this wave's own execution capacity — genuine infrastructure/ownership decision |
| RISK-W07-002 | SEC-05's external assessment surfaces an open Critical/High finding with no immediate remediation path available within this wave's own timeline | Low-medium | High if it occurs — SEC-05's own acceptance criterion requires "zero open Critical/High" or an approved waiver; an unremediable Critical finding would block this story's own closure | High (conditional on occurrence) | W07-E02-S001 | The control map itself, built against SEC-01-04's already-`accepted` implementation, is expected to find few genuinely new gaps (those findings were already closed by their own waves) — the primary risk is a finding specific to the *interaction* between controls, not a single control in isolation | If a Critical/High finding is found, escalate to the product-security lead for either an emergency remediation task or an approved, time-bounded waiver — do not silently close SEC-05 with an unaddressed Critical finding | unassigned | open | Low-medium — SEC-01-04's own prior acceptance reduces but does not eliminate this risk |
| RISK-W07-003 | W07-E04-S001 (the final verification gate) discovers an unresolved gap in an earlier wave's own closure that this wave cannot itself fix without reopening that wave's own scope | Medium | Medium-high — a genuine gap found this late in the programme has no clean remediation path within this wave's own story boundaries | Medium-high | W07-E04-S001, W07-E04-S002 | The gate re-run is explicitly designed to surface exactly this class of finding (per its own purpose: re-running REVIEW §30's gate, not merely restating its prior conclusions) — finding a real gap here is the gate working correctly, not a defect in this wave's own planning | Any gap found is recorded in the disposition audit and carried into the closure/claim-upgrade decision package (W07-E04-S002) as an explicit open item for the human authority, potentially recommending a follow-up story/wave rather than silently absorbing the fix into this wave's own already-defined scope | unassigned | open | Cannot be pre-resolved — this risk's entire nature is "a fact to be discovered," mirroring W02's own RISK-W02-002 framing for the DATA-01 mismatch audit |

## Residual risk after mitigation

RISK-W07-002 is expected to reduce to low residual risk given SEC-01-04's own prior acceptance across
earlier waves. RISK-W07-001 and RISK-W07-003 cannot be pre-resolved by this wave's own planning — their
outcomes are facts to be discovered (DEC-Q9's actual resolution timing; whether the final gate finds a
genuine unresolved gap), tracked honestly rather than silently assumed away.
