---
id: TRACK-DEPENDENCY-REGISTER
type: register
title: Dependency register — cross-wave, story, and decision dependencies
status: active
created_at: 2026-07-12
updated_at: 2026-07-12
derived: true
---

# Dependency register

DERIVED VIEW. Per mandate §11.6: Dependency ID | Source item | Target item | Dependency type |
Description | Blocking status | Resolution. Canonical source = `impl/index.md` wave map
"Depends on" column + `requirement-inventory.md` notes columns + PLAN §7 cross-cutting
dependency notes (`premier-framework-implementation-plan.md` §7).

Blocking status legend: **hard** = build/story cannot proceed without resolution; **soft** = can
proceed with a documented assumption/safe default; **soft-blocked** = build unblocked by a
provisional value, final closure pending a human decision.

## (a) Cross-wave dependencies (source: `impl/index.md` wave map)

| Dependency ID | Source item | Target item | Type | Description | Blocking status | Resolution |
|---|---|---|---|---|---|---|
| DEP-001 | W01 | W00 | cross-wave | W01 (zero-dependency-hardening) cannot start until W00 (baseline-and-verification) closes | hard | Resolved when W00 closure-report.md is accepted |
| DEP-002 | W02 | W00 | cross-wave | W02 (data-safety-and-migration-tooling) cannot start until W00 closes | hard | Resolved when W00 closure-report.md is accepted |
| DEP-003 | W03 | W01 | cross-wave | W03 (identity-and-session-security) needs W01's validation seam (FBL-08 RouteMeta) | hard | Resolved when W01 closure-report.md is accepted |
| DEP-004 | W03 | W02 | cross-wave | W03 needs W02's grant-table migration built on the DATA-09 online-migration protocol | hard | Resolved when W02 closure-report.md is accepted |
| DEP-005 | W04 | W02 | cross-wave | W04 (jobs-and-durable-delivery) needs W02's DATA-09 protocol for W6-T1 migration | hard | Resolved when W02 closure-report.md is accepted |
| DEP-006 | W05 | W03 | cross-wave | W05 (application-model-and-layering) needs W03's SEC-01 actor model to stabilise registrar security assumptions | hard | Resolved when W03 closure-report.md is accepted |
| DEP-007 | W06 | W05 | cross-wave | W06 (contracts-compatibility-release) needs W05's AR-03 to unblock REL-03b legs | hard | Resolved when W05 closure-report.md is accepted |
| DEP-008 | W07 | W00..W06 | cross-wave | W07 (performance-and-final-verification) depends on all prior waves | hard | Resolved when W00–W06 closure-reports are all accepted |

## (b) Finding-level dependencies (source: `requirement-inventory.md` notes columns + PLAN §7)

| Dependency ID | Source item | Target item | Type | Description | Blocking status | Resolution |
|---|---|---|---|---|---|---|
| DEP-009 | AR-02 (W05-E02) | AR-01 (W05-E01) | story | AR-02 (typed port keys + compiled provider graph) depends on AR-01 T1/T2 (ownership-bound ApplicationModel) | hard | Structural within-wave sequencing; resolved when AR-01 T1/T2 complete |
| DEP-010 | AR-03 (W05-E03) | AR-01 (W05-E01) | story | AR-03 (one authoritative declaration) depends on AR-01 | hard | Structural within-wave sequencing |
| DEP-011 | AR-03 (W05-E03) | AR-02 (W05-E02) | story | AR-03 also depends on AR-02 | hard | Structural within-wave sequencing |
| DEP-012 | DATA-01 T4/T5 (W02-E02) | DATA-09 T1-T5 (W02-E01) | story | DATA-01's risky migration steps (T4/T5) depend on DATA-09's online expand/backfill/validate/contract protocol (T1–T5) | hard | Resolved when DATA-09 T1–T5 land in W02-E01 |
| DEP-013 | DATA-07 (W03-E04) | SEC-01 (W03-E01) | story | DATA-07 (relationship semantics + actor attribution) has a hard dependency on SEC-01 (server-side session state) | hard | Resolved when SEC-01 closes in W03-E01 |
| DEP-014 | DATA-07 (W03-E04) | SEC-04 (W05-E04) | story | DATA-07 has a soft (secondary) dependency on SEC-04 (bound authz staleness/memory) | soft | Can proceed with documented assumption; strengthened when SEC-04 lands in W05-E04 |
| DEP-015 | DATA-08 W6-T1 (W04-E04) | DATA-09 (W02-E01) | story | DATA-08's W6-T1 hash-widening migration depends on the DATA-09 online-migration protocol | hard | Resolved when DATA-09 lands in W02-E01 |
| DEP-016 | DX-04 (W06-E01) | DX-01 T5 (W01-E04) | story | DX-04 (golden consumer + upgrade matrix) depends on DX-01 T5 (scaffold harness) | hard | Resolved when DX-01 T5 lands in W01-E04-S001 |
| DEP-017 | REL-03b T3 (W06-E02) | DX-06 (W06-E02) | story | REL-03's b-leg T3 depends on DX-06 (OpenAPI merge complete-or-loud) | hard | Resolved when DX-06 lands in W06-E02-S001 |
| DEP-018 | REL-03b T5 (W06-E02) | AR-03 / DX-03 | story | REL-03's b-leg T5 depends on AR-03 (W05-E03) and DX-03 (W06-E01, deferred-design) | hard | Resolved when AR-03 lands in W05-E03 and DX-03's design story completes |
| DEP-019 | REL-03b T7 (W06-E02) | DX-04 (W06-E01) | story | REL-03's b-leg T7 depends on DX-04 (golden consumer + upgrade matrix) | hard | Resolved when DX-04 lands in W06-E01-S002 |
| DEP-020 | FBL-01 (W05-E05) | AR-01 (W05-E01) | story | FBL-01 (kernel re-home) depends on AR-01 | hard | Resolved when AR-01 lands in W05-E01 |
| DEP-021 | FBL-01 (W05-E05) | AR-02 (W05-E02) | story | FBL-01 also depends on AR-02 | hard | Resolved when AR-02 lands in W05-E02 |
| DEP-022 | SEC-02 T5 (W03-E05) | SEC-01 T1 (W03-E01) | story | SEC-02's T5 (audit) depends on SEC-01 T1 (server-side session state) | hard | Resolved when SEC-01 T1 lands in W03-E01 |
| DEP-023 | DX-07 T4 (W04-E04) | AR-04 T5 (W05-E03) | story | DX-07's T4 (readiness diagnostics) depends on AR-04 T5 (post-seal waiver mechanism) | hard | Resolved when AR-04 T5 lands in W05-E03-S002 |

## (c) Decision dependencies (source: `requirement-inventory.md` §B + REVIEW §F/§U)

| Dependency ID | Source item | Target item | Type | Description | Blocking status | Resolution |
|---|---|---|---|---|---|---|
| DEP-024 | SEC-01 (W03-E01) | DEC-Q1 | decision | SEC-01 is blocked pending the safe default per `planning-assumptions.md` (IdP grant_id claim contract) | soft-blocked — safe default/provisional value unblocks build, final activation pending human decision | Supersede provisional value with human decision when DEC-Q1 lands (see `decision-register.md` PA-01) |
| DEP-025 | PERF-02..05 absolute-SLO (W07-E01) | DEC-Q9 | decision | PERF-02..05's absolute-SLO acceptance criteria are blocked pending reference-perf-env ownership | soft-blocked — safe default/provisional value unblocks build, final activation pending human decision | Supersede provisional value with human decision when DEC-Q9 lands (see `decision-register.md` PA-02) |
| DEP-026 | REL-01 / REL-02 final activation (W06-E03) | DEC-Q10 | decision | REL-01/REL-02's final activation is blocked pending repo-admin protection setup | soft-blocked — safe default/provisional value unblocks build, final activation pending human decision | Supersede provisional value with human decision when DEC-Q10 lands (see `decision-register.md` PA-03) |

## Summary

26 dependency rows: 8 cross-wave + 15 finding-level (story) + 3 decision. No dependency edge
named in the source instructions was omitted.
