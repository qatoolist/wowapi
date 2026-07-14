---
id: ANALYSIS-WAVE-ALLOC
type: analysis
title: Story-level allocation detail for waves W02–W07 (companion to requirement-inventory)
status: complete
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# W02–W07 story allocation (canonical groupings)

Story groupings ratified by the programme author. Wave-tree generators must follow these
exactly; allocation changes require a change-log entry. Task content comes from the PLAN §5
T-rows and MATRIX CS specs named per story.

## W02 data-safety-and-migration-tooling (5 epics)
- **E01 online-migration-protocol (DATA-09):** S001 manifest-and-lock-budget (T1, T2); S002 expand-backfill-validate (T3, T4, T5 — T4 backfill harness reuses DATA-02 T1's lease primitive: forward-dependency, so S002 builds a minimal checkpoint lease and W04-E01 replaces it — record as planned deviation-risk); S003 canary-switch-contract-drills (T6, T7, T8, T9).
- **E02 tenant-fk-integrity (DATA-01):** S001 parent-indexes-scanner-gate (T1, T2, T6); S002 audit-fk-validate-negatives (T3, T4, T5, T7, T8 — T4/T5 gated on E01 S001/S002 acceptance).
- **E03 version-allocation-and-gc (DATA-05):** S001 all T1–T5 (single reviewer domain).
- **E04 aggregate-write-contract (DATA-06):** S001 T1–T4 (T2 owns the shared registrar_pg.go actor fix; DATA-07 T3 consumes it later).
- **E05 production-seed-sync (FBL-02):** S001 design+implement per CS-21 acceptance bar (prod boot on empty catalog DB reaches readiness only after seed-sync; readiness reports seed/catalog hash). Contains explicit design-investigation task (catalog manifest format) before implementation tasks.

## W03 identity-and-session-security (5 epics)
- **E01 server-side-session-state (SEC-01):** S001 grant-schema-and-membership (T1, T2, T3; D-01 enacted; DEC-Q1 safe default recorded in story assumptions); S002 capacity-and-privileged-resolver (T4, T5); S003 assurance-and-credential-schemes (T6, T7); S004 cross-repo-cutover-plan (coordination artifact for PROD-04: sequencing, staging validation, rollback — documentation/verification story, no product code).
- **E02 outbound-security-governance (SEC-06):** S001 all tasks (D-07 enacted).
- **E03 webhook-authenticated-replay (SEC-03):** S001 T1–T4 (breaking Verifier interface — compat notes mandatory).
- **E04 relationship-semantics (DATA-07):** S001 T1, T2, T4 (T3 consumed from DATA-06 T2; hard dep W03-E01 accepted; SEC-04 epoch dep noted soft — cache-invalidation AC deferred-linked to W05-E04-S002).
- **E05 workflow-privileged-completion (SEC-02 remainder):** S001 T4 ratification design+implement (or documented reject-interim posture) + T5 durable audit (grant-ID field dep on E01 S001).

## W04 jobs-and-durable-delivery (4 epics)
- **E01 lease-fencing-primitive-and-jobs (DATA-02):** S001 shared-primitive (T1 — replaces W02-E01-S002's minimal checkpoint lease; migration note); S002 jobs-lease-and-finalize (T2, T3, T4); S003 idempotency-and-chaos (T5, T6, T7 chaos harness — harness shared with E02/E03).
- **E02 remote-io-outside-tx (DATA-03):** S001 notify-and-webhook-three-stage (T1, T2, T3); S002 inbound-two-phase-and-contracts (T4, T5, T6, T8; T7 already done — cross-ref DATA-08 W0); S003 retry-adoption (FBL-04: cenkalti/backoff parity + fault injection).
- **E03 bulk-multi-worker-safety (DATA-04):** S001 T1 stopgap (can start at wave entry); S002 T2–T6 leased claims + lifecycle + chaos.
- **E04 compliance-and-readiness (DATA-08 W6 + DX-07):** S001 audit-hash-widening (W6-T1; D-04 enacted; dep W02-E01 protocol; PROD-05 staging-drill coordination noted); S002 anchor-dsr-hold (W6-T2, T3, T4, T5); S003 readiness-truthfulness (DX-07 T1, T2, T3; T4 deferred-linked to W05 AR-04 T5 waiver).

## W05 application-model-and-layering (5 epics)
- **E01 application-model (AR-01):** S001 model-and-registrar-capability (T1, T2; D-02/D-03 enacted as story ADR refs); S002 registry-ownership (T3, T4, T5, T6); S003 snapshots-hash-race (T7, T9, T10 + T8 post-seal rejection); S004 legacy-adapter (T11 — compatibility story).
- **E02 typed-ports (AR-02):** S001 port-api-and-forge-proofs (T1, T2); S002 graph-validation-and-profiles (T3, T4, T5); S003 lifecycle-manifest-retirement + legacy shim (T6, T7).
- **E03 authoritative-declarations (AR-03 + AR-04 remainder):** S001 manifest-and-projections (AR-03 T1, T3, T4, T5; T2 = DX-06-owned, cross-ref only); S002 boot-strictness-and-waivers (AR-04 T2, T3, T4, T5 — T5 builds the shared waiver mechanism consumed by SEC-06/DX-07).
- **E04 wiring-and-cache-hygiene:** S001 constructor-bypass-closure (AR-06 T2, T3); S002 authz-cache-bounding (SEC-04 all tasks per CS-17: golang-lru + epoch table D-06; DATA-07 T4 cache-invalidation AC closes here).
- **E05 kernel-re-home (FBL-01):** S001 foundation-move-and-shims (CS-01 mechanics: git mv 9 pkgs, mfa forwarding shim, depguard+boundaries extension); S002 re-home-verification (kernel package-count AC, wowsociety identity-suite green on shim — PROD-02 coordination).

## W06 contracts-compatibility-release (4 epics)
- **E01 consumer-and-dsl:** S001 module-dsl-design (DX-03 — DESIGN INVESTIGATION story: outputs design doc + decision, no code); S002 golden-consumer-matrix (DX-04; dep W01-E04-S001 harness).
- **E02 api-contract-gates:** S001 openapi-merge-complete-or-loud (DX-06 T1–T3; owns AR-03 T2 scope; validator dependency decision task); S002 compat-gates-buildable-now (REL-03a: T1, T2, T4, T6, T8, T9); S003 compat-gates-unblocked (REL-03b: T3, T5, T7 — entry criteria reference their unblocking stories).
- **E03 release-gating:** S001 exact-commit-release-pipeline (REL-01 T1–T8 buildable set); S002 protection-activation (REL-01 remainder + DEC-Q10 — human-gated story, explicit blocked status allowed); S003 blocking-security-scans (REL-02: Trivy exit-code flip, waiver schema, visibility-guard review).
- **E04 documentation-gates:** S001 doc-example-compile-gate (CS-22/AR-05 T3 spec); S002 generated-docs-and-labels (AR-05 T4, T5 — dep E02/W05-E03 manifest).

## W07 performance-and-final-verification (4 epics)
- **E01 performance-programme (PERF-02..05 + CS-16):** S001 request-benchmarks-real-pg (PERF-02 relative); S002 rules-resolution-sql (PERF-03); S003 sweeper-materialization (PERF-04); S004 checksum-behaviour-and-bench-coverage (PERF-05 + CS-16's 7 hot-path package benchmarks + budgets). DEC-Q9 tracked at epic level; absolute-SLO ACs conditional.
- **E02 verification-hardening:** S001 security-verification-profile (SEC-05); S002 coverage-truthfulness-completion (REL-04 T5 fail-not-skip, T6 skip manifest, T7 race-integration schedule, T8 real fuzz — owns PERF-06 T3/T4 scope).
- **E03 product-alignment-verification:** S001 wowsociety-readiness-check (verify PROD-01..05 coordination artifacts exist and product upgrade path documented; framework-side only).
- **E04 programme-closure:** S001 final-verification-gate (re-run REVIEW §30-style gate across the programme; traceability-matrix completeness; disposition audit); S002 closure-and-claim-decision (programme closure report; production-readiness claim upgrade decision package for the human authority).

## Cross-wave sequencing notes
W02-E01-S002's minimal checkpoint lease is intentionally superseded by W04-E01-S001 (recorded
here so the deviation is planned, not silent). DATA-01 T4/T5 must not start before W02-E01
S001+S002 acceptance. W03-E01 may start against DEC-Q1's safe default. W05 entry requires
W03-E01 acceptance (actor model stability). W06-E02-S003 legs unblock individually as DX-06 /
W05-E03 / W06-E01-S002 land.
