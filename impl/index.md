---
id: IMPL-PROGRAMME
type: programme-index
title: wowapi Premier-Framework Implementation Programme
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# wowapi Implementation Programme

## Purpose

The authoritative, executable blueprint and tracking structure that takes `wowapi` from its
current verified state (strong pre-production foundation; NOT production-ready per the Fable 5
review) to the intended production-ready premier-framework state — executable story by story
without returning to the source planning documents.

## Source documents (authority order)

1. `docs/implementation/fable5-closure-depth-matrix-2026-07-11.md` (MATRIX — closure specs CS-01..25; latest)
2. `docs/implementation/fable5-final-architecture-review-2026-07-11.md` (REVIEW — verdict, FBL/D/T items)
3. `docs/implementation/premier-framework-implementation-plan.md` (PLAN — per-finding task tables; task-level source of record)
4. `docs/implementation/architecture-directive-2026-07-11.md` (directive — normative finding definitions)
5. Repository state at planning time: `main @ 0a31186` (post #22–#25 merges — see `analysis/requirement-inventory.md` §E session delta)

Where sources overlap, the later/stricter statement wins unless a decision record says
otherwise (see `analysis/conflict-resolution.md`).

## Scope

All 38 PLAN findings, all REVIEW FBL/T items and D-01..09 decisions, all MATRIX closure specs,
and the three human decisions (tracked, not blocking build). **Non-goals:** product
(wowsociety) code changes — recorded as PROD-01..05 coordination items; deferred items
DEF-01..06 (registered with reopen triggers); anything in REVIEW §M rejected register.

## Planning principles

Per the mandate: doability over theoretical completeness; dependency-aware sequencing;
framework-first scope (generic boundary preserved); full traceability chain (source →
requirement → wave → epic → story → AC → task → artifact → evidence → review → acceptance);
evidence-driven completion; plan-versus-actual deviation records (plans never rewritten).

## Wave map (8 waves; dependency-derived, not arbitrary)

| Wave | Title | Objective (one line) | Depends on | Epics |
|---|---|---|---|---|
| W00 | baseline-and-verification | Pin verification of the 8 executed finding-slices at current HEAD; capture coverage/lint/bench/CI baselines; ADR-ify D-01..09 | — | 2 |
| W01 | zero-dependency-hardening | Everything valuable with no upstream dependency: linter utilisation (FBL-05/07), OTel correlation (FBL-06), HTTP hardening (FBL-08/09), generator+doc+test fixes (DX-01/02, T-DOC-01, T-TEST-01, FBL-03) | W00 | 4 |
| W02 | data-safety-and-migration-tooling | DATA-09 online-migration protocol, then DATA-01 tenant FKs over it; DATA-05/06; FBL-02 prod seed-sync | W00 | 5 |
| W03 | identity-and-session-security | SEC-01 server-side session state (+D-01, DEC-Q1 safe default), SEC-06, SEC-03, DATA-07 (dep SEC-01), SEC-02 remainder | W01 (validation seam), W02 (grant-table migration uses DATA-09) | 5 |
| W04 | jobs-and-durable-delivery | Shared lease/fencing primitive → DATA-02/03/04; FBL-04 retry adoption; DATA-08 W6 audit integrity (D-04); DX-07 readiness truthfulness | W02 (DATA-09 for W6-T1 migration) | 4 |
| W05 | application-model-and-layering | AR-01/02 ownership model (+D-02/03), AR-03/AR-04 remainder, AR-06 remainder, SEC-04 cache (+D-06), FBL-01 kernel re-home | W03 (SEC-01 actor model stabilises registrar security assumptions) | 5 |
| W06 | contracts-compatibility-release | DX-03 design + DX-04 golden consumer; DX-06 merge + REL-03a/b diff gates; REL-01/REL-02 release gating (DEC-Q10 activation); doc-example gates (CS-22/AR-05) | W05 (AR-03 unblocks REL-03b legs) | 4 |
| W07 | performance-and-final-verification | PERF-02..05 relative programme (+DEC-Q9), SEC-05 profile, REL-04 remainder (real fuzz), product-alignment verification, programme closure gate | all prior | 4 |

Execution order: strictly W00→W07 for wave entry; independent stories inside a wave may run
in parallel per their own dependencies. A later wave must not start while a mandatory
predecessor capability is unaccepted (exception requires a deviation record).

## Programme-level risks (top; full register in `tracking/risk-register.md`)

kernel surface locks in before FBL-01 (mitigated: W05 before any v1-stabilisation claim) ·
SEC-01 cross-repo cutover coordination (PROD-04) · DATA-08 W6 breaking live audit rows
(PROD-05 staging drill) · three human decisions stall final activation (safe defaults keep
build unblocked) · single-maintainer bandwidth across 8 waves.

## Programme acceptance

All waves closed per their `closure-report.md`; requirement-traceability matrix shows every
`planned` item accepted/deferred-with-approval; the REVIEW §30-style final gate re-run passes;
no unexplained deviation; production-readiness claim upgrade is a separate, explicit decision.

## Structure

`governance/` (lifecycle, statuses, DoR/DoD, policies, templates) · `analysis/` (inventories,
dispositions, conflicts, duplicates, scope, assumptions) · `tracking/` (registers + matrices;
derived views marked as such) · `waves/` (the executable programme).

## Progress maintenance

Canonical status lives in each item's front matter (`story.md`, `epic.md`, `wave.md`).
Registers/indexes are derived roll-ups, updated when canonical files change; `tracking/change-log.md`
records every programme-structure change. Evidence/artifact registration rules:
`governance/evidence-policy.md`, `governance/artifact-policy.md`.
