---
id: W05-E03
type: epic
title: Authoritative declarations
status: planned
wave: W05
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-03
  - AR-04
depends_on:
  - W05-E01
  - W05-E02
stories:
  - W05-E03-S001
  - W05-E03-S002
decisions: []
risks:
  - RISK-W05-E03-001
---

# W05-E03 — Authoritative declarations

## Epic objective

Establish the module manifest as the single authoritative declaration from which routes, permission/
resource catalogs, schema, and other projections are deterministically derived (AR-03 remainder:
T1, T3, T4, T5), and close the remaining boot-time silent-behaviour gaps — duplicate collectors,
empty required fragments, post-seal config rejection, and unwaived no-op adapters in `prod` — via a
shared waiver mechanism also consumed by SEC-06 and DX-07 (AR-04 remainder: T2-T5).

## Problem being solved

`requirement-inventory.md` row AR-03: "One authoritative declaration, derived projections (T1–T5) |
IMPL | P1 | planned | W05-E03-S001..S002 | T2 = DX-06 duplicate → single owner DX-06 (see
duplicate-analysis)." Row AR-04: "Eliminate boot-time silent behaviour | IMPL | P1 | partial |
W05-E03-S002 | T1 EXECUTED (verified ×2); T2–T5 planned, dep AR-01; T5 waiver shared w/ SEC-06/DX-07."
(Note: an initial requirement-inventory typo targeted a nonexistent S003 for AR-04; corrected 2026-07-12 to W05-E03-S002 — inventory, allocation detail, and this epic now agree. See tracking/change-log.md.)

PLAN's own AR-03 directive requirement: "module manifest becomes the authoritative
declaration; deterministic tooling derives routes, permission/resource catalogs, schema/OpenAPI,
event/job/workflow/rule identifiers, dependency/provider graphs, migration/seed/i18n/OpenAPI bundle
inventory, required-capability profiles, conformance tests, doc tables, and a machine-readable
manifest." Today, declarations are scattered and hand-duplicated across the codebase; AR-04's
remaining gaps (T2-T5) leave duplicate collectors, empty required fragments, and unwaived
misconfiguration silently accepted at boot rather than rejected.

## Scope

- The manifest schema definition, scoped to what this wave needs — identity + projection inputs, not
  DX-03's full typed-operation DSL (S001, PLAN AR-03 T1).
- Deriving route registration/metadata from the manifest, proven by a golden-fixture delta test that
  IS the acceptance gate (S001, PLAN AR-03 T3).
- A lint rule failing on hand-maintained duplicate identity or an omitted projection (S001, PLAN
  AR-03 T4).
- Documentation/test/manifest export projections, extending T3's golden-delta coverage (S001, PLAN
  AR-03 T5).
- Rejecting duplicate collectors — every collector rejects a second write to the same identity (S002,
  PLAN AR-04 T2).
- Rejecting empty required fragments (S002, PLAN AR-04 T3).
- Extending AR-01 T8's post-seal write rejection to config/namespace/collector state (S002, PLAN
  AR-04 T4).
- Explicit optional-capability declaration; `prod` readiness fails on required-but-no-op/missing
  adapter unless a policy-approved waiver exists — the shared waiver mechanism consumed by SEC-06 and
  DX-07 (S002, PLAN AR-04 T5).

## Out of scope

- **AR-03 T2 (the OpenAPI merge fix)** — explicitly single-owned by DX-06 per
  `requirement-inventory.md`'s own row: "T2 = DX-06 duplicate → single owner DX-06 (see
  duplicate-analysis)." This epic's S001 records T2 as an out-of-scope cross-reference; it is neither
  implemented nor skipped-without-mention here.
- **AR-04 T1 (unknown-namespace rejection at boot)** — already executed and verified twice per
  `requirement-inventory.md`'s AR-04 row. This epic's S002 does not re-plan or re-implement it.
- **W05-E01's `ApplicationModel`/`Registrar` and W05-E02's typed-port/provider-graph work
  themselves** — already built by their own epics; this epic depends on and consumes them.

## Source requirements

AR-03 (T1, T3, T4, T5 — T2 cross-ref only). AR-04 (T2, T3, T4, T5 — T1 already executed).

## Architectural context

AR-03's manifest-derived-projection tooling depends on W05-E01 (AR-01, the ownership-bound model)
and W05-E02 (AR-02, the compiled provider graph and its three-profile projection) both having
landed, per PLAN AR-03 T3's own dependency row: "T1, AR-01, AR-02." AR-04's remainder depends on
AR-01 per PLAN AR-04 T2's own dependency row: "AR-01 T1." This epic's two stories are independent of
each other in their own right (AR-03's manifest work and AR-04's boot-strictness work are disjoint
concerns), but both share the common upstream dependency on W05-E01 (and, for AR-03 specifically,
W05-E02 as well).

## Included stories

- **W05-E03-S001 — manifest-and-projections** (PLAN AR-03 T1, T3, T4, T5; T2 cross-ref only to
  DX-06): the manifest schema and its derived-projection tooling, including the golden-delta
  acceptance gate.
- **W05-E03-S002 — boot-strictness-and-waivers** (PLAN AR-04 T2, T3, T4, T5): duplicate-collector
  and empty-fragment rejection, post-seal config rejection, and the shared no-op-adapter waiver
  mechanism.

## Dependencies

Depends on W05-E01 (full epic) and W05-E02 (full epic, specifically for S001's AR-03 T3). No
dependency on W05-E04 or W05-E05. Downstream: W06-E02-S003 (REL-03b) and W06-E04-S002 (AR-05 T4/T5)
both depend on this epic's S001 (AR-03's manifest), per `impl/index.md`'s own wave-map note ("W05
(AR-03 unblocks REL-03b legs)") and `impl/analysis/wave-allocation-detail.md`'s own note ("dep
E02/W05-E03 manifest"). AR-04 T5's waiver mechanism (S002) is a forward-shared primitive consumed
by SEC-06 (W03 scope, already built or in-progress) and DX-07 T4 (W04-E04-S003's deferred-linked
item).

## Risks

RISK-W05-E03-001 (AR-03 T3's golden-delta test is, per PLAN's own framing, "this test IS the
acceptance gate" — a high-stakes single test with no fallback proof mechanism) — see `risks.md` for
full detail and mitigation/contingency.

## Required decisions

None. Neither AR-03 nor AR-04 has a D-0N architecture-decision dependency in the source — confirmed
by scanning `requirement-inventory.md` §B: no D-0N row targets AR-03 or AR-04 specifically.

## Epic acceptance criteria

- **AC-W05-E03-01**: The manifest schema is defined, scoped to identity + projection inputs; a
  golden-fixture manifest change deterministically produces the expected full projection diff with
  no other hand-edited file — the golden-delta test (PLAN's own "this test IS the acceptance gate")
  passes; a lint rule fails on hand-maintained duplicate identity or an omitted projection; AR-03 T2
  is correctly recorded as out-of-scope, single-owned by DX-06.
- **AC-W05-E03-02**: Every collector rejects a second write to the same identity; a module declaring
  a required-but-empty fragment fails boot; the post-seal error-not-panic contract (D-03) extends to
  config/namespace/collector state; a `prod` profile with a required-but-no-op/missing adapter and no
  waiver fails readiness by name, the same configuration under `local` succeeds, and a
  policy-approved waiver suppresses the failure with an audit record.
- **AC-W05-E03-03**: All stories have passed independent review per mandate §14, with S001
  specifically checked given AR-03 T3's "this test IS the acceptance gate" framing.

## Closure conditions

Both stories reach `accepted`; AC-W05-E03-01 through AC-W05-E03-03 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date; AR-03
T2's out-of-scope status is confirmed still correctly recorded (not silently absorbed) at closure.
