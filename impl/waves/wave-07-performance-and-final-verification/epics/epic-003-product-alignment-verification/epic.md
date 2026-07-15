---
id: W07-E03
type: epic
title: Product alignment verification
status: blocked
wave: W07
owner: W07-Phase-A-Execution.W07E03S001
reviewer: W05ReviewGateFinal
priority: medium
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements: []
depends_on: []
stories:
  - W07-E03-S001
decisions: []
risks:
  - RISK-W07-E03-001
---

# W07-E03 — Product alignment verification

## Epic objective

Verify, framework-side only, that the five wowsociety-coordination items (PROD-01..05) each have a
documented coordination artifact confirming the enabling framework capability exists and that a product
upgrade path is documented — without implementing, modifying, or requiring any wowsociety-repository
code change.

## Problem being solved

`impl/analysis/requirement-inventory.md` §D lists five product-level items explicitly excluded from
this programme's own framework-side implementation scope, per mandate §2.3's framework/product boundary:
"PROD-01 | wowsociety `policy_override` composite FK | Product schema fix | DATA-01 T1 (parent unique
index) + DATA-09 protocol"; "PROD-02 | wowsociety `kernel/mfa` import migration (5 identity files) |
Product code migration | FBL-01 re-home ships deprecated forwarding shim"; "PROD-03 | wowsociety
readiness/timeout backports to committed main.go | Product hand-edit | DX-07 T1 + FBL-09 fix the
templates"; "PROD-04 | SEC-01 impersonation cutover (whoami/impersonation/tests) | Product auth flow
rework | SEC-01 T1/T5 grant contract + coordinated rollout plan"; "PROD-05 | DATA-08 W6 staging audit
re-verification before version bump | Product compliance drill | hash_version branch verification
(D-04)." Each row's own "Enabling framework capability" column names a framework-side deliverable
already built by an earlier wave in this programme. This epic's own job is to verify each of those five
enabling capabilities genuinely exists and is documented as consumable by wowsociety — not to perform
the wowsociety-side consumption itself.

## Scope

- Verify DATA-01 T1 + DATA-09's protocol exist and are documented as consumable for PROD-01
  (`policy_override` composite FK).
- Verify FBL-01's deprecated forwarding shim exists and is documented as consumable for PROD-02 (`kernel/
  mfa` import migration).
- Verify DX-07 T1 + FBL-09's template fixes exist and are documented as consumable for PROD-03
  (readiness/timeout backports).
- Verify SEC-01 T1/T5's grant contract exists and a coordinated rollout plan is documented for PROD-04
  (impersonation cutover).
- Verify D-04's `hash_version` branch verification exists and is documented as the mechanism PROD-05's
  own staging-drill re-verification would use.

## Out of scope

- **Any wowsociety-repository code change** — explicitly out of scope per mandate §2.3's framework/
  product boundary; this epic's own acceptance criteria are about documentation/coordination-artifact
  existence, not wowsociety code.
- **Performing the actual wowsociety-side cutover, migration, or drill** (e.g. actually running
  wowsociety's own staging audit re-verification for PROD-05, actually migrating wowsociety's own 5
  identity files for PROD-02) — these are wowsociety-repository actions, tracked but not performed here.

## Source requirements

None directly (this epic verifies existing framework-side deliverables from DATA-01, DATA-09, FBL-01,
DX-07, FBL-09, SEC-01, D-04 — each already implemented and accepted in an earlier wave). The five
PROD-01..05 rows themselves are `requirement-inventory.md` §D items, explicitly classified as
product-level, excluded from framework implementation.

## Architectural context

This epic exists because mandate §2.3's own framework/product boundary means this programme cannot
implement a wowsociety-repository change under any circumstance, yet the programme's own traceability
discipline (mandate §1.2, "do not duplicate the same work... document conflicts") requires that these
five coordination items not simply vanish from tracking once their framework-side enabling capability
ships. `impl/analysis/wave-allocation-detail.md`'s own W07-E03 grouping states this exactly: "S001
wowsociety-readiness-check (verify PROD-01..05 coordination artifacts exist and product upgrade path
documented; framework-side only)." This is a single-story epic because all five PROD-0N items share the
identical verification shape (confirm the framework-side capability exists; confirm a coordination
artifact documents the upgrade path) — there is no basis for splitting them into separate epics or
multiple stories.

## Included stories

- **W07-E03-S001 — wowsociety-readiness-check**: framework-side verification that all five PROD-01..05
  coordination artifacts exist.

## Dependencies

Depends cross-wave on DATA-01/DATA-09 (W02), FBL-01 (W05-E05), DX-07/FBL-09
(W04-E04-S003/W01-E03-S001), SEC-01 (W03-E01), and D-04 (W04-E04-S001,
W00-E02-S003). The all-prior-waves entry gate established record status, not capability truth.
This epic's direct checks found DATA-01's parent key absent and SEC-01's rollout artifact stale, so
those dependencies are not consumable for PROD-01/04 despite their earlier accepted records.

## Risks

`RISK-W07-E03-001` is realized: direct verification found a missing database prerequisite and a
security-relevant stale rollout artifact. Both are open blockers; see `risks.md`.

## Required decisions

None. This epic verifies existing decisions/capabilities, it does not make new ones.

## Epic acceptance criteria

- **AC-W07-E03-01**: Every PROD-01..05 row has a documented framework-side coordination artifact
  confirming the enabling framework capability exists (via re-verification, not merely trusting the
  originating wave's own closure claim) and a documented product upgrade path.
- **AC-W07-E03-02**: No wowsociety-repository code change is performed by this epic under any
  circumstance.
- **AC-W07-E03-03**: The story has passed independent review per mandate §14.

## Closure conditions

The story reaches `accepted`; AC-W07-E03-01 through AC-W07-E03-03 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date.
