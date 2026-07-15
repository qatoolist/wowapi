---
id: W02-E04
type: epic
title: Aggregate write contract
status: accepted
wave: W02
owner: W02FKVerAgg
reviewer: W02ReviewGate
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-06
depends_on: []
stories:
  - W02-E04-S001
decisions: []
risks:
  - RISK-W02-E04-001
---

# W02-E04 — Aggregate write contract

## Epic objective

Make the resource-mirror write mandatory and framework-enforced: build a typed aggregate
repository/unit-of-work helper that bundles a module's business-row write with the resource-mirror
upsert, audit row, and outbox entry in one atomic transaction, so that a module can no longer write
its business row without the framework also writing the mirror — and source real actor attribution
(`created_by`) into that same helper, replacing today's `uuid.Nil` placeholder.

## Problem being solved

`requirement-inventory.md` row DATA-06 records: "Resource-mirror aggregate write contract (T1–T4) |
IMPL | P1 | planned | W02-E04-S001 | T2 shared fix w/ DATA-07 T3 (one owner)." PLAN's own DATA-06
evidence is exact: "`kernel/resource` package doc confirms a manual, comment-only contract — a
module owns its business table and separately upserts the mirror, with no framework enforcement.
`registrar_pg.go:38-58` passes `created_by` as `uuid.Nil` with a TODO. Even the reference handler
manually performs two independent statements." The gap this epic closes is that the mirror-write
contract exists today only as documentation prose — nothing in the framework prevents a module from
writing its business row and forgetting the mirror upsert, and every mirror row currently written
carries no real actor, only a placeholder `uuid.Nil` with an unresolved TODO.

## Scope

- A typed aggregate repository/unit-of-work helper bundling aggregate write + mirror upsert + audit
  + outbox atomically, proven via fault injection at each of 4 stages independently with full
  rollback at every stage (PLAN DATA-06 T1).
- Sourcing `created_by` from context in the same helper; rejecting a missing actor for user-
  initiated writes while leaving system-actor paths unaffected (PLAN DATA-06 T2) — this task is the
  **single owner** of the `registrar_pg.go` nil-actor placeholder fix; PLAN's own PF-DATA cross-
  cutting note (2) states plainly: "`kernel/resource/registrar_pg.go`'s nil-actor placeholder is one
  fix claimed by two findings (DATA-06 T2, DATA-07 T3) — one owner, not two PRs." This epic is that
  one owner; DATA-07 T3 (W03-E04-S001, out of this epic's scope) is a downstream consumer that
  reuses this fix directly rather than reimplementing it — see `dependencies.md`.
- Migrating the reference handler onto the new helper, so the reference pattern other modules copy
  is no longer the two-independent-statements pattern DATA-06 targets (PLAN DATA-06 T3).
- Updating `kernel/resource` documentation to describe the mandatory-mirror contract as implemented,
  not merely as a comment-only aspiration (PLAN DATA-06 T4).

## Out of scope

- **DATA-07's own relationship-semantics work** (`Checker.Has`'s party-subject-edge evaluation,
  the full `subject_kind` matrix, the mutation-audit-cache AC) — that is W03-E04-S001's scope,
  which has a hard dependency on SEC-01 (W03-E01) per PLAN's own note: "**Hard dependency on
  PF-SEC's SEC-01 — do not schedule before it lands**." This epic does not implement any DATA-07
  task; it only produces the T2 fix DATA-07 T3 will later reuse.
- **AR-03's own authoritative-declaration work** — PLAN DATA-06 T1's own risk note states "Overlaps
  AR-03 — coordinate to avoid a parallel one-off mechanism." AR-03 is W05-E03's scope, sequenced
  much later than this wave. This epic does not wait for AR-03 (there is no dependency edge in the
  source forcing that), but its T1 helper's design should be recorded as a known overlap requiring
  future coordination, not silently built as if AR-03 did not exist — see "Architectural context"
  below and `risks.md`.
- **wowsociety's own `committeeseat.go` migration onto the new helper** — PLAN's own wowsociety-
  impact note states this is "not urgent, current pattern still functions" and should "follow
  wowapi's T1/T3 (reference implementation proven first)." This is tracked as a product-level
  coordination item outside framework implementation scope per mandate §2.3, not this epic's
  responsibility.

## Source requirements

DATA-06 (T1, T2, T3, T4). Cross-referenced: DATA-07 T3 (W03-E04-S001, downstream consumer of this
epic's T2 fix, not this epic's scope).

## Architectural context

`kernel/resource`'s current contract is enforced only by documentation and by convention — PLAN's
own evidence: "a module owns its business table and separately upserts the mirror, with no
framework enforcement... Even the reference handler manually performs two independent statements."
This epic converts that convention into a structural guarantee via a typed aggregate repository/
unit-of-work helper: a module that wants to write its business row is given no code path that skips
the mirror write, because the helper is the code path. This is architecturally adjacent to AR-03's
future work (W05-E03, "one authoritative declaration, derived projections") in the sense that both
concern making an implicit multi-step contract explicit and framework-enforced — PLAN's own risk
note for T1 flags this overlap directly. This epic proceeds now (W02, P1, no blocking dependency)
rather than waiting for AR-03 (W05) because DATA-06's own priority and this wave's task brief place
it here; the overlap is a coordination risk to track, not a sequencing dependency to enforce.

The T2 actor-attribution fix touches the exact file (`registrar_pg.go`) that DATA-07 T3 will later
extend for relationship-mirror writes — this epic's T2 is deliberately scoped as the single owner of
that fix's mechanism (source actor from context, reject missing actor for user-initiated writes) so
that DATA-07 T3, when W03 reaches it, reuses the mechanism this epic builds rather than re-deriving
an independent, possibly-divergent implementation of the same nil-actor fix.

## Included stories

- **W02-E04-S001 — aggregate-write-contract** (PLAN DATA-06 T1–T4): the typed aggregate repository/
  unit-of-work helper; the actor-attribution fix (single owner of the shared `registrar_pg.go` fix
  surface); the reference-handler migration; the `kernel/resource` documentation update.

## Dependencies

No dependency on any other W02 epic — this epic's scope (`kernel/resource`) is disjoint from
W02-E01/E02's migration-protocol and tenant-FK work, W02-E03's `kernel/artifact`/`kernel/document`
scope, and W02-E05's seed-sync scope. This epic depends only on W00's exit gate per `wave.md`'s
entry criteria. Downstream: W03-E04-S001 (DATA-07 T3) consumes this epic's T2 fix once W03 begins —
see `dependencies.md`.

## Risks

RISK-W02-E04-001 (the AR-03 overlap creating rework if AR-03's eventual design conflicts with this
epic's T1 helper shape) — see `risks.md` for full detail and mitigation/contingency.

## Required decisions

None. DATA-06 has no D-0N architecture-decision dependency in the source — confirmed by scanning
`requirement-inventory.md` §B and REVIEW §F/§U for any D-0N row citing DATA-06; none exists. This
epic's story accordingly carries no `decisions/` directory.

## Epic acceptance criteria

- **AC-W02-E04-01**: The typed aggregate repository/unit-of-work helper makes it structurally
  impossible for a module to write its business row without the framework also writing the mirror
  in the same transaction, proven by fault injection at each of 4 stages independently with full
  rollback at every stage.
- **AC-W02-E04-02**: `created_by` is sourced from context in the same helper; a user-initiated write
  with no actor fails fast; system-actor paths are unaffected — proven by a test with/without actor,
  system vs. user path. This fix is recorded as the single owner of the `registrar_pg.go` nil-actor
  placeholder, with an explicit cross-reference for DATA-07 T3 to consume rather than reimplement.
- **AC-W02-E04-03**: The reference handler no longer manually calls both the business-row write and
  the mirror upsert as two independent statements — it is migrated onto the new helper, with
  existing reference tests passing.
- **AC-W02-E04-04**: `kernel/resource` documentation matches the implemented mandatory-mirror
  contract, confirmed by manual review.
- **AC-W02-E04-05**: The story has passed independent review per mandate §14, specifically
  confirming T2's actor-attribution fix does not break any legitimate system-actor call site and
  that the AR-03 overlap is recorded as a known coordination risk, not silently ignored.

## Closure conditions

The story reaches `accepted` (satisfying its own `closure.md`); AC-W02-E04-01 through AC-W02-E04-05
above are all satisfied; `closure-report.md` for this epic is completed with reviewer conclusion and
acceptance date; RISK-W02-E04-001 is recorded (not resolved — its resolution depends on W05-E03's
own AR-03 design, out of this epic's control) with a clear pointer for that future story to consult
this epic's T1 design before finalizing AR-03's own mechanism.
