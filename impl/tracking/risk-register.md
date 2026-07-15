---
id: TRACK-RISK-REGISTER
type: register
title: Risk register — programme-level and cross-cutting risks
status: active
created_at: 2026-07-12
updated_at: 2026-07-12
derived: true
---

# Risk register

DERIVED VIEW. Per mandate §11.7: Risk ID | Description | Likelihood | Impact | Severity |
Affected items | Mitigation | Contingency | Owner | Status | Residual risk. Canonical source =
REVIEW §T Risk register (`docs/implementation/fable5-final-architecture-review-2026-07-11.md`
lines 372–383), PLAN §7 cross-cutting risks (`docs/implementation/premier-framework-implementation-plan.md`
lines 776–808), and `impl/index.md` "Programme-level risks". Likelihood/Owner/Status/Residual-risk
columns are not present verbatim in the source tables — populated here using engineering judgment
per mandate instruction, grounded in each source's own description. Status = `open` for all rows:
nothing has executed under this programme yet.

Duplicates across sources are merged into one row with both sources cited (e.g. "kernel surface
locks in" appears in both REVIEW §T and `impl/index.md`).

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-001 | Kernel surface locks in wrong before FBL-01 lands (9 packages remain outside `foundation/`, shape hardens under real usage before the re-home) | Medium | High | High | FBL-01, W05, all consumers of the 9 kernel packages | Do FBL-01 before any v1-stabilisation claim; blast radius already proven small (REVIEW §P) | If surface has hardened undesirably, accept the debt and re-home with a documented breaking-change/migration note rather than delaying further | framework (Fable 5) | open | Low once W05-E05 (FBL-01) closes; monitored via blast-radius check before/after |
| RISK-002 | SEC-01: JWT-trusted session state is a live security gap until server-side grant resolution ships (P0) | High | High | High (P0) | SEC-01, W03-E01, DATA-07, PROD-04 | Server-side grant resolver (SEC-01 T1–T7); DEC-Q1 safe default unblocks build without waiting on IdP contract | If IdP claim contract slips, ship framework-side grant table with conservative default expiry and revisit product cutover timing | framework (Fable 5), cutover coordination = product (wowsociety) | open | Medium until SEC-01 closes and PROD-04 cutover completes |
| RISK-003 | DATA-01: FK integrity gap across tenants (missing composite tenant FKs) risks cross-tenant data corruption | Medium | High | High (P0) | DATA-01, W02-E02, PROD-01 | Composite FK + online migration protocol (DATA-09 precedes DATA-01 T4/T5) | If migration proves unsafe at scale, stage via DATA-09's expand/backfill/validate/contract phases with a rollback point at each phase | framework (Fable 5); wowsociety's own instance = product (PROD-01) | open | Low once DATA-01 lands over the DATA-09 protocol |
| RISK-004 | DATA-08: compliance-evidence metadata/tx_id fields unhashed, weakening audit-trail integrity guarantees | Medium | High | High | DATA-08, W04-E04, PROD-05 | hash_version discriminator column (D-04) so historical rows remain verifiable after the hash contract widens | Staging drill (PROD-05) before any version bump touching live audit rows; roll back the migration if the drill surfaces unverifiable rows | framework (Fable 5); staging drill = product (wowsociety) | open | Low once W04-E04 closes and PROD-05 drill passes |
| RISK-005 | PF-9: no production seed-sync path exists — deployment-blocking gap | Medium | High | High (prod-blocking) | FBL-02, W02-E05 | FBL-02 implements the seed-sync path; CS-21 acceptance bar fixed in MATRIX | If design proves more complex than scoped, split into its own investigation task before implementation | framework (Fable 5) | open | Low once W02-E05 (FBL-02) closes |
| RISK-006 | e2e concurrency flake masks real test failures, undermining confidence in the full-suite gate | Medium | Medium | Medium | T-TEST-01, W01-E04-S003, REL-04 | T-TEST-01 reproduces the flake first (original shared-DB diagnosis withdrawn) before attempting a fix | If not reproducible, quarantine the flaky test with a tracked ticket rather than leaving it silently red/green | framework (Fable 5) | open | Low-medium; residual until root cause is confirmed and fixed |
| RISK-007 | yaml.v3 dependency cadence — supply-chain risk if the upstream module's maintenance slows or stops | Low | Low | Low | all yaml.v3 consumers (config, i18n loaders) | Monitor community fork activity; no action required unless cadence degrades further | Vendor or fork the dependency if upstream becomes unmaintained | framework (Fable 5) | open | Low; monitoring-only risk |
| RISK-008 | Reference-performance-environment ownership is an unscheduled prerequisite — no owner or timeline exists for the dedicated Linux amd64 reference runner that PERF-02..05 absolute SLOs require | Medium | High | High | PERF-02, PERF-03, PERF-04, PERF-05, W07-E01, DEC-Q9 | Provisional default set (GH runner + reference json) unblocks relative/container benchmarking now (REVIEW §12); absolute SLO gated on DEC-Q9 | If no dedicated runner is ever provisioned, keep relative/container benchmarks as the permanent acceptance bar and formally waive absolute SLO | framework (Fable 5); environment provisioning = product/infra owner | open | Medium until DEC-Q9 resolves or the provisional default is ratified as permanent |
| RISK-009 | SEC-05 / Wave-6-class penetration test needs a named security lead and an external vendor — neither identified in any source document | Low | High | Medium | SEC-05, W07-E02-S001 | Track as its own scheduling dependency; do not let it silently block W07 exit if vendor engagement is still pending | Substitute an internal, narrower security review if no external vendor can be engaged in time, with an explicit deviation record | framework (Fable 5); vendor engagement = product/business owner | open | Medium until a vendor or internal substitute is confirmed |
| RISK-010 | GitHub repo-admin actions for REL-01/REL-02 are blocked — no protected `release` environment, no branch/tag protection exist today (confirmed live via `gh api`) | High | High | High (P0) | REL-01, REL-02, W06-E03, DEC-Q10 | ~85% of REL-01/REL-02 is buildable without admin actions now; DEC-Q10 (safe-default provisional posture) unblocks the remaining build | Track admin-gated work as a separate ticket ("PF-REL-ADMIN-01") so agent-completable YAML work isn't silently gated on this | framework (Fable 5); repo-admin action = human (repo owner) | open | Medium-high until DEC-Q10 resolves (merge-queue rulesets unavailable on a user-owned repo) |
| RISK-011 | SEC-01 cross-repo cutover coordination (PROD-04): wowsociety's impersonation/whoami flow must migrate in lockstep with SEC-01's framework-side grant table, risking a breaking cutover if timed wrong | Medium | High | High | SEC-01, PROD-04, W03-E01 | Coordinated rollout plan as part of SEC-01 T1/T5; framework owns validity/expiry/revocation, wowsociety keeps UX/audit-trail (REVIEW §7 item 2 recommendation) | If cutover cannot be synchronised, ship SEC-01 behind a compatibility shim until wowsociety's migration lands | framework (Fable 5) for SEC-01; product (wowsociety) for cutover execution | open | Medium until PROD-04 cutover completes and shim (if used) is retired |
| RISK-012 | DATA-08 W6 migration could break live audit rows in production if the hash-widening change is not validated against real data first | Medium | High | High | DATA-08, PROD-05, W04-E04 | PROD-05 staging drill re-verifies compliance evidence against real staging data before the version bump ships | Roll back the W6-T1 migration and re-scope the hash_version discriminator if the staging drill fails | framework (Fable 5) for migration; product (wowsociety) for staging drill execution | open | Low once PROD-05 drill passes cleanly (duplicate of RISK-004's mitigation, tracked separately because it is programme-level per `impl/index.md`) |
| RISK-013 | Three human decisions (DEC-Q1, DEC-Q9, DEC-Q10) stall final activation of SEC-01, PERF-02..05 absolute SLOs, and REL-01/REL-02 respectively if left unresolved past their target waves | Medium | Medium | Medium | SEC-01, PERF-02..05, REL-01, REL-02, W03, W06, W07 | Safe defaults (PA-01/PA-02/PA-03 in `decision-register.md`) keep build unblocked regardless of when the human decisions land | Escalate to programme owner if a decision remains unresolved by the time its dependent wave reaches exit gate | framework (Fable 5) tracks; human decision-makers own resolution | open | Low — provisional defaults absorb the risk; residual is only the final-activation delay |
| RISK-014 | Single-maintainer bandwidth across 8 waves — the entire programme currently has one effective maintainer, risking schedule slip or quality shortcuts under load | High | Medium | Medium | all waves | Dependency-aware sequencing (this programme) lets independent stories run in parallel within a wave; conductor-worker delegation (per project CLAUDE.md) spreads mechanical work to cheaper workers | If bandwidth is insufficient, slip wave exit dates rather than skip evidence/verification steps | framework (Fable 5) | open | Medium — structural risk that persists for the programme's duration |

## Duplicate-source notes

- RISK-001 (kernel surface locks in) is stated in both REVIEW §T and `impl/index.md` programme
  risks — merged into one row, both sources cited.
- RISK-002 (SEC-01 JWT-trusted session state) is stated in both REVIEW §T and (as PROD-04
  cross-repo cutover coordination) `impl/index.md` — the underlying security gap is RISK-002; the
  cutover-coordination angle is tracked separately as RISK-011 because `impl/index.md` frames it
  distinctly (coordination risk, not the security gap itself).
- RISK-004 and RISK-012 both concern DATA-08 W6, sourced from REVIEW §T ("DATA-08 metadata/tx_id
  unhashed") and `impl/index.md` ("DATA-08 W6 breaking live audit rows / PROD-05 staging drill")
  respectively — kept as two rows because they describe different failure modes (data-integrity
  gap vs. migration-breaks-production-data), both converging on the same mitigation (D-04 +
  PROD-05 drill).
- RISK-013 (three human decisions stall final activation) is stated in `impl/index.md` programme
  risks and elaborated by PLAN §7 items 1, 8, 10 and REVIEW §F Q1 — merged into one row.

15 distinct risk rows (within the ~12–15 target range), covering all items named in REVIEW §T (7),
PLAN §7 (3 named: items 8, 9, 10), and `impl/index.md` programme risks (5), with de-duplication
noted above.
