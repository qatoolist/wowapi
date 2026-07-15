---
id: W05-E05-RISKS
type: epic-risks
epic: W05-E05
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E05 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W05-003 and
RISK-W05-004 originate at wave scope and are reproduced/elaborated here because they land entirely
within this epic's two stories.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W05-003 | FBL-01 is, per MATRIX CS-01, "the largest single architectural correction" and "must precede v1 stabilisation" — this epic's own schedule risk is inherited directly, since it is sequenced last within W05 (depends on E01+E02) | Low for the mechanical move itself (MATRIX CS-01: "behaviour-preserving moves," "import-path churn only") — medium for schedule risk, since any slip in E01/E02 delays this epic's start | High if delayed past this wave — the programme's own top risk item states "kernel surface locks in before FBL-01" | High (schedule-conditional) | W05-E05 (both stories), transitively every later wave assuming a stabilised kernel surface | S001's own sequencing starts as soon as E01+E02 land, not deferred to the end of this wave's closure window; the move itself is mechanical for 8 of 9 packages | If E01/E02 slip materially, escalate FBL-01's schedule risk to the acceptance authority | unassigned | open | Medium — irreducible schedule dependency on E01/E02 |
| RISK-W05-004 | The `kernel/mfa` re-home is the one auth-critical exception among the nine packages — REVIEW §P: "Bounded but not trivial ... a real, security-sensitive import-path + call-site migration across 5 identity files, not a mechanical zero-cost change" | Medium — the forwarding-shim mitigation is a well-understood Go pattern, but wowsociety's own migration timing is outside this epic's direct control | Medium-high if the shim is incorrectly implemented or the identity suite is not genuinely re-run — a broken TOTP/OTP path is an authentication-availability regression | Medium-high | W05-E05-S001, W05-E05-S002, PROD-02 (wowsociety coordination) | S001's shim task is scoped and tested independently of the other 8 packages' mechanical move; S002's acceptance requires wowsociety's full identity/authz suite (not just an mfa-scoped subset) to run green against the shim | If the shim breaks wowsociety's identity suite, do not proceed to remove the shim or advance PROD-02's migration timeline until root-caused and fixed | unassigned | open | Medium until wowsociety's identity/authz suite is confirmed green |

## Residual risk after mitigation

RISK-W05-003 is a schedule-conditional risk this epic's own internal sequencing (starting S001 as
soon as E01+E02 land) manages but cannot fully eliminate, since it depends on upstream epics landing
on time. RISK-W05-004 reduces to low once wowsociety's full identity/authz suite is confirmed green
against the shim, independently verified, not merely asserted.
