---
id: W05-E02
type: epic
title: Typed ports
status: planned
wave: W05
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-02
depends_on:
  - W05-E01
stories:
  - W05-E02-S001
  - W05-E02-S002
  - W05-E02-S003
decisions: []
risks: []
---

# W05-E02 — Typed ports

## Epic objective

Build a typed port-key API (`port.Key[T]`) and a compiled, boot-time-validated provider graph on
top of W05-E01's `Registrar` capability type, so that module dependency wiring is expressed through
generic, type-safe declarations rather than ad hoc constructor calls, with zero reflection at
request-resolve time and boot-time rejection of duplicate providers, missing requirements,
undeclared edges, cycles, and invalid scope/lifetime edges.

## Problem being solved

`requirement-inventory.md` row AR-02 states: "Typed port keys + compiled provider graph (T1–T7) |
IMPL | P1 | planned | W05-E02-S001..S003 | Depends AR-01 T1/T2." PLAN's own AR-02 directive
requirement: "`port.Key[T]` + `Define`/`Provide`/`Require`/`Resolve` generic functions bound to an
owner-bound `Registrar`; compiler builds a heterogeneous provider graph, type-erasing only at
compile time (never on request hot paths); rejects duplicate providers, missing requirements, type
mismatches, undeclared dependencies, cycles, invalid scope/lifetime edges before any process
starts." Today, no typed port-key API and no compiled provider graph exist — wiring is hand-copied
per profile (API/worker/migrate), and the existing `kernel/lifecycle` manifest is hand-maintained
rather than generated.

## Scope

- `port.Key[T]` and the four generic free functions (`Define`/`Provide`/`Require`/`Resolve`) bound
  to W05-E01's `Registrar` (S001, PLAN AR-02 T1).
- The internal compiler factory minting registrars with immutable owner identity, proven by an
  adversarial compile-fail fixture that capability confusion is impossible if AR-01/AR-02 share one
  `Registrar` type (S001, PLAN AR-02 T2).
- A type-erased provider graph with zero reflection on request hot paths, proven by benchmark and
  lint (S002, PLAN AR-02 T3).
- Boot-time graph validation: duplicate providers, missing requirements, undeclared edges, cycles,
  invalid scope/lifetime edges — one adversarial fixture per failure class (S002, PLAN AR-02 T4).
- Compiling API/worker/migrate profiles as three projections of one graph, with no hand-copied
  wiring template remaining (S002, PLAN AR-02 T5).
- Retirement of the hand-maintained `kernel/lifecycle` manifest in favor of the generated graph,
  preserving existing lint-failure classes now data-driven (S003, PLAN AR-02 T6).
- The legacy port adapter (`ProvidePort`/`Port` shim onto the typed graph), confirmed to have zero
  external callers today (S003, PLAN AR-02 T7).

## Out of scope

- **W05-E01's `ApplicationModel`/`Registrar` themselves** — already built; this epic reuses them.
- **AR-03's manifest and derived-projection tooling** — a separate epic (W05-E03) that depends on
  this epic's own S002-T5 (the three-profile projection) having landed, per PLAN AR-02 T5's own
  dependency row citing AR-03.
- **FBL-01's kernel package re-home** — W05-E05's scope, sequenced after both this epic and W05-E01.

## Source requirements

AR-02 (T1-T7).

## Architectural context

AR-02 directly reuses W05-E01's `Registrar` capability type (T1's own dependency row: "AR-01 T1,
T2"), making this epic entirely dependent on W05-E01 at wave scope. Within this epic, T1-T2 (the
port-key API and registrar-forge safety) are the foundation; T3-T4 (zero-reflection graph, boot-time
validation) build on T1-T2; T5 (three-profile projection) depends on T1-T4 and on AR-03's manifest
shape being fixed, per PLAN's own note "sequence after AR-03's manifest shape is fixed" — recorded
as a forward-looking sequencing note, not a hard blocking dependency, since this epic's own T5
delivers the projection mechanism AR-03 later consumes, not the reverse. T6 (retiring the
hand-maintained lifecycle manifest) depends on T1-T5; T7 (the legacy port adapter) depends on T1-T6.
This epic's three stories group these by phase-cluster per `impl/analysis/wave-allocation-detail.md`:
S001 (T1, T2) is the port-API foundation; S002 (T3, T4, T5) is the graph-validation-and-profiles
closure; S003 (T6, T7) is the lifecycle-manifest retirement plus legacy shim.

## Included stories

- **W05-E02-S001 — port-api-and-forge-proofs** (PLAN AR-02 T1, T2): the typed port-key API and the
  registrar-forge safety proof.
- **W05-E02-S002 — graph-validation-and-profiles** (PLAN AR-02 T3, T4, T5): zero-reflection provider
  graph, boot-time validation, three-profile projection.
- **W05-E02-S003 — lifecycle-manifest-retirement** (PLAN AR-02 T6, T7): retiring the hand-maintained
  lifecycle manifest; the legacy port adapter.

## Dependencies

Depends on W05-E01 (full epic — T1's own PLAN dependency row cites AR-01 T1, T2 specifically, and
the capability-confusion safety proof T2 requires presupposes AR-01's `Registrar` type exists in its
final form). Downstream: W05-E03 (AR-03) depends on this epic's S002-T5 (three-profile projection);
W05-E05 (FBL-01) depends on this epic in full, alongside W05-E01, per MATRIX CS-01.

## Risks

No dedicated epic-level risk beyond the general "depends on AR-01 landing correctly" transitive
risk, already tracked at W05-E01's own risk entries. This epic's own task-level risk values (PLAN's
own Medium/High/Medium/Medium/Medium/Low-medium/Low for T1-T7) are notable at T2 (High — "verify
capability confusion is impossible if AR-01/AR-02 share one `Registrar` type") but are addressed
directly by T2's own adversarial compile-fail fixture, which this epic's S001 carries an
independent-review task to confirm.

## Required decisions

None. AR-02 has no D-0N architecture-decision dependency in the source — confirmed by scanning
`requirement-inventory.md` §B for any D-0N row targeting AR-02: none exists.

## Epic acceptance criteria

- **AC-W05-E02-01**: `port.Key[T]` and the four generic free functions compile and resolve correctly
  bound to W05-E01's `Registrar`; the internal compiler factory mints registrars with immutable owner
  identity, and an adversarial compile-fail fixture proves capability confusion is impossible given
  AR-01/AR-02 share one `Registrar` type.
- **AC-W05-E02-02**: The provider graph performs zero `reflect.*` calls at `Resolve` time, proven by
  benchmark and lint; boot-time validation rejects duplicate providers, missing requirements,
  undeclared edges, cycles, and invalid scope/lifetime edges, one adversarial fixture per failure
  class; API/worker/migrate profiles compile as three projections of one graph with no hand-copied
  wiring template remaining.
- **AC-W05-E02-03**: The hand-maintained `kernel/lifecycle` manifest is retired in favor of the
  generated graph, with existing lint-failure classes still passing, now data-driven; the legacy
  port adapter compiles/resolves unchanged for any existing caller (confirmed zero external callers
  today).
- **AC-W05-E02-04**: All three stories have passed independent review per mandate §14, with S001
  specifically checked given T2's High-risk capability-confusion proof.

## Closure conditions

All three stories reach `accepted` (each satisfying its own `closure.md`); AC-W05-E02-01 through
AC-W05-E02-04 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date.
