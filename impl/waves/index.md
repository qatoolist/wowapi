---
id: WAVES-INDEX
type: waves-index
title: Implementation waves — roll-up index
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Waves index

Roll-up per mandate §16.1. Canonical status lives in each wave's own `wave.md` front matter;
this table is a derived view, updated when a wave's canonical status changes. Wave map source:
`../index.md` §"Wave map"; per-item allocation source: `../analysis/requirement-inventory.md`.

| Wave | Title | Objective | Status | Priority | Depends on | Epics | Stories | Progress | Evidence completeness | Planned exit gate |
|---|---|---|---|---|---|---|---|---|---|---|
| [W00](wave-00-baseline-and-verification/wave.md) | baseline-and-verification | Pin verification of the 8 executed finding-slices (SEC-02, PERF-01, PERF-06, DATA-08 W0, AR-04 T1, AR-05 T1/T2, AR-06 T1, REL-04 T1-T4) at current HEAD; capture coverage/lint/bench/CI baselines; ADR-ify D-01..D-09 | accepted | P0 (gate for all later waves) | — | 2 | 6 | 6/6 stories accepted (100%) | complete | All 8 executed slices re-verified with evidence registered at HEAD; baselines captured (coverage, lint, bench-budgets, CI wall-clock, dependency/toolchain inventory); D-01..D-09 ADRs ratified and registered |
| [W01](wave-01-zero-dependency-hardening/wave.md) | zero-dependency-hardening | Land every finding with no upstream framework dependency: zero-cost + judged static-analysis utilisation (FBL-05/07), supply-chain/hooks hygiene, OTel trace/log correlation + pgx query tracer (FBL-06), HTTP transport hardening (FBL-09) and central validation enforcement (FBL-08), generator correctness (DX-01/DX-02) + documentation reconciliation (T-DOC-01, DX-05 residual, FBL-03) + e2e flake diagnosis (T-TEST-01) | verification | P1 | W00 | 4 | 10 | 10/10 stories accepted (100%) | complete | All W01 stories accepted; `.golangci.yml` judged+zero-cost sets enabled with zero unexplained regressions; trace/log correlation + pgx tracer evidence registered; HTTP timeouts + CSRF body bound enforced in prod profile; RouteMeta validation enforcement live behind profile flag; generator-output-boots test green; T-DOC-01/DX-05/FBL-03 reconciled; T-TEST-01 diagnosis resolved or explicitly re-scoped further |
| [W02](wave-02-data-safety-and-migration-tooling/wave.md) | data-safety-and-migration-tooling | Build the DATA-09 online expand/backfill/validate/contract protocol; build DATA-01 composite tenant FKs over it; DATA-05 version-allocation/blob-GC; DATA-06 resource-mirror write contract; FBL-02 production seed-sync path | partially-accepted | P0 | W00 | 5 | 8 | 0/8 stories started (0%) | none captured yet | DATA-09 T1-T9 CI drill pipeline green; DATA-01 composite FKs validated with zero cross-tenant mismatches; DATA-05/DATA-06 fault-injection suites pass; FBL-02 prod-profile empty-catalog boot reaches readiness only post seed-sync |
| [W03](wave-03-identity-and-session-security/wave.md) | identity-and-session-security | SEC-01 server-side session/grant state (D-01, DEC-Q1 safe default); SEC-06 outbound-security escape-hatch governance (D-07); SEC-03 webhook replay binding; DATA-07 relationship semantics (hard dep SEC-01); SEC-02 remainder (ratification T4/T5) | in-progress | P0 | W01 (validation seam), W02 (grant-table migration reuses DATA-09) | 5 | 8 | 0/8 stories started (0%) | none captured yet | SEC-01 adversarial membership/grant test classes pass; wowsociety two-repo cutover plan documented (PROD-04); SEC-06/SEC-03/DATA-07/SEC-02 remainder accepted |
| [W04](wave-04-jobs-and-durable-delivery/wave.md) | jobs-and-durable-delivery | Shared lease/fencing primitive (DATA-02 T1) → DATA-02/03/04 full closure; FBL-04 retry-library adoption; DATA-08 Wave-6 audit-hash widening (D-04) + remaining W6 tasks; DX-07 truthful readiness diagnostics | in-progress | P0/P1 | W02 (DATA-09 for W6-T1 migration) | 4 | 11 | 0/11 stories started (0%) | none captured yet | Named chaos tests pass at every DATA-02/03/04 boundary; audit tamper-matrix proves every field breaks verification; DX-07 prod-capacity enforcement live behind waiver mechanism |
| [W05](wave-05-application-model-and-layering/wave.md) | application-model-and-layering | AR-01/AR-02 ownership-bound ApplicationModel + typed provider graph (D-02, D-03); AR-03/AR-04/AR-06 remainder; SEC-04 authz cache bounding (D-06); FBL-01 kernel re-home | planned | P1 (core) | W03 (actor model stabilises registrar security assumptions) | 5 | 13 | 0/13 stories started (0%) | none captured yet | Adversarial ownership tests pass across all declaration classes; model-hash determinism proven; kernel package count reduced per FBL-01 target list; wowsociety identity suite green on `foundation/mfa` |
| [W06](wave-06-contracts-compatibility-release/wave.md) | contracts-compatibility-release | DX-03 module DSL design (deferred/design-only); DX-04 golden consumer + upgrade matrix; DX-06 OpenAPI merge (single owner of AR-03 T2) + REL-03a/b compatibility gates; REL-01/REL-02 release gating (DEC-Q10 activation); doc-example CI gate (AR-05 T3/CS-22) | in-progress | P0/P1 | W05 (AR-03 unblocks REL-03b legs) | 4 | 10 | 0/10 stories started (0%) | none captured yet | REL-01 T1-T6,T8,T10 buildable-now scope closed; REL-03a legs green; DX-06 full-field merge fixtures pass; doc-example gate blocks a staled-example fixture |
| [W07](wave-07-performance-and-final-verification/wave.md) | performance-and-final-verification | PERF-02..05 relative/container performance programme (DEC-Q9 provisional reference env); SEC-05 versioned security profile; REL-04 remainder (real coverage-guided fuzz, shared w/ PERF-06 T3/T4); product-alignment verification; programme closure gate | in-progress | P1 | all prior waves | 4 | 9 | 5/9 stories accepted, 2 blocked, 2 planned (55.6%) | 40 produced evidence records; E01 complete, E02/E03 blocker evidence preserved, E04 has planned slots only | Reference-env relative benchmarks published for PERF-02..05; SEC-05 control map leaves zero open Critical/High; hosted fuzz runs real `-fuzz` on PR + nightly; final programme closure gate passes with no open critical finding |

## Notes

- All eight wave trees (W00–W07) are fully generated: wave → epics → stories → tasks with
  artifacts/evidence indexes. Totals: 33 epics, 75 stories, 297 task files.
- "Progress" and "Evidence completeness" columns are derived from each story's/task's own status
  front matter; at programme start (all stories `status: planned`, all tasks `status: todo`) they
  read 0% / none by construction, not by omission.
- Update history for this derived view: see `../tracking/change-log.md`.
