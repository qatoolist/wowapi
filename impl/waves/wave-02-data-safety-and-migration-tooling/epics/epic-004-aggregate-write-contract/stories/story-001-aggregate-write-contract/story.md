---
id: W02-E04-S001
type: story
title: Typed aggregate write contract with mandatory mirror, audit, and outbox
status: accepted
wave: W02
epic: W02-E04
owner: W02FKVerAgg
reviewer: W02ReviewGate
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-06
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W02-E04-S001-01
  - AC-W02-E04-S001-02
  - AC-W02-E04-S001-03
  - AC-W02-E04-S001-04
artifacts:
  - ART-W02-E04-S001-001
  - ART-W02-E04-S001-002
  - ART-W02-E04-S001-003
  - ART-W02-E04-S001-004
evidence:
  - EV-W02-E04-S001-001
  - EV-W02-E04-S001-002
  - EV-W02-E04-S001-003
  - EV-W02-E04-S001-004
decisions: []
risks:
  - RISK-W02-E04-001
---

# W02-E04-S001 — Typed aggregate write contract with mandatory mirror, audit, and outbox

## Story ID

W02-E04-S001

## Title

Typed aggregate write contract with mandatory mirror, audit, and outbox

## Objective

Build a typed aggregate repository/unit-of-work helper bundling a module's business-row write with
the resource-mirror upsert, audit row, and outbox entry atomically; source real `created_by` actor
attribution into that helper, rejecting missing actors for user-initiated writes; migrate the
reference handler onto the new helper; and update `kernel/resource` documentation to describe the
mandatory-mirror contract as implemented.

## Value to the framework

Today the resource-mirror contract exists only as documentation prose — PLAN's own evidence: "a
module owns its business table and separately upserts the mirror, with no framework enforcement...
Even the reference handler manually performs two independent statements." A module author who
forgets the second statement silently produces a business row with no corresponding mirror entry,
and every mirror row currently carries `uuid.Nil` as its `created_by` — an unattributed write with
an unresolved TODO in the code itself. This story converts an easy-to-forget, two-step convention
into a single, structurally-enforced write path: a module cannot get the mirror-write step wrong by
omission, because there is no longer a code path that lets it.

## Problem statement

`requirement-inventory.md` row DATA-06: "Resource-mirror aggregate write contract (T1–T4) | IMPL |
P1 | planned | W02-E04-S001 | T2 shared fix w/ DATA-07 T3 (one owner)." PLAN's DATA-06 evidence:
"`kernel/resource` package doc confirms a manual, comment-only contract — a module owns its business
table and separately upserts the mirror, with no framework enforcement. `registrar_pg.go:38-58`
passes `created_by` as `uuid.Nil` with a TODO. Even the reference handler manually performs two
independent statements." Four tasks close this gap: T1 builds the atomic helper; T2 fixes the
nil-actor placeholder inside that same helper; T3 migrates the reference handler so the pattern
other modules copy is the new, enforced one; T4 brings the documentation into agreement with the
implementation.

## Source requirements

DATA-06 (T1, T2, T3, T4).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit, per this
programme's fail-first convention):

- `kernel/resource`'s package documentation describes the mirror-write contract as a manual,
  comment-only obligation — a module writes its own business row, then separately calls the
  registrar's `Upsert` to write the mirror. Nothing in the framework enforces that both statements
  happen, or happen atomically.
- `registrar_pg.go:38-58` passes `created_by` as `uuid.Nil`, with a TODO comment marking this as
  known-incomplete — every mirror row written today carries no real actor attribution.
- Even the framework's own reference handler (the example other module authors are expected to
  copy) performs the business-row write and the mirror upsert as two independent statements, not a
  single atomic operation — meaning the pattern being propagated by example is the exact pattern
  this story exists to retire.
- `internal/modules/identity/committeeseat.go:69-70` in wowsociety independently reproduces this
  same manual pattern (PLAN's wowsociety-impact note) — confirming the gap is a real, copied defect
  class, not a single isolated instance.

This story's own re-confirmation step is to read `kernel/resource`'s current documentation,
`registrar_pg.go:38-58`, and the reference handler's current implementation at this story's actual
start commit before building the new helper, since line numbers and exact wording may have shifted
since PLAN was written.

## Desired state

A typed aggregate repository/unit-of-work helper exists such that calling it is the only way a
module writes an aggregate's business row — the helper itself writes the mirror upsert, an audit
row, and an outbox entry in the same transaction, and a fault injected at any one of the four stages
(business write, mirror upsert, audit write, outbox write) rolls back the entire transaction, not
just the failed stage. `created_by` is sourced from request/operation context inside the helper; a
user-initiated write with no resolvable actor fails fast (does not silently proceed with a
placeholder); a system-actor path (e.g. a scheduled job) is unaffected and continues to function.
The reference handler calls the new helper instead of performing two independent statements.
`kernel/resource`'s documentation accurately describes this mandatory-mirror contract.

## Scope

- The typed aggregate repository/unit-of-work helper (PLAN DATA-06 T1).
- Actor-attribution sourcing and enforcement inside the same helper (PLAN DATA-06 T2) — this task
  is the single owner of the `registrar_pg.go` nil-actor placeholder fix shared with DATA-07 T3 (see
  "Dependencies" and the epic-level `dependencies.md`); DATA-07 T3 is explicitly out of this story's
  scope and is expected to consume this story's T2 mechanism when W03 reaches it.
- Migrating the reference handler onto the new helper (PLAN DATA-06 T3).
- Updating `kernel/resource` documentation to match (PLAN DATA-06 T4).

## Out of scope

- DATA-07's own relationship-semantics implementation — W03-E04-S001's scope, hard-dependent on
  SEC-01 (W03-E01) per PLAN's own note. This story produces the T2 fix DATA-07 T3 will consume; it
  does not implement any DATA-07 task itself.
- AR-03's own authoritative-declaration/projection mechanism (W05-E03) — PLAN T1's own risk note
  flags an overlap to coordinate, not a dependency to block on. See RISK-W02-E04-001.
- wowsociety's `committeeseat.go` migration onto the new helper — product-level, tracked separately,
  "not urgent" per PLAN's own sequencing note; this story's reference-handler migration (T3) is the
  prerequisite proof-of-pattern, not the wowsociety-side migration itself.
- Any module beyond the reference handler being migrated onto the new helper — T3's scope is
  specifically the reference handler; migrating every other existing module consumer is not named
  in PLAN's DATA-06 task table and is not assumed here.

## Assumptions

- The exact shape of the "4 stages" fault-injection test (business write, mirror upsert, audit
  write, outbox write, in that order) is inferred from PLAN's own evidence ordering
  ("business table... separately upserts the mirror... audit... outbox" — the natural write-path
  ordering implied by `kernel/resource`'s existing contract description) — the precise internal
  ordering is confirmed at implementation time against the actual current write path, not invented
  here.
- `registrar_pg.go`'s exact current line numbers (PLAN cites `:38-58`) are assumed to be
  approximately, not exactly, stable since PLAN was written — this story's T2 task re-confirms the
  exact current location before editing, per this programme's fail-first convention.
- The reference handler's identity (which specific handler in the codebase is "the" reference
  handler PLAN refers to) is assumed to be discoverable and unambiguous at implementation time;
  if more than one candidate exists, T3's task record documents which one was chosen and why.

## Dependencies

None within W02-E04 (single-story epic). Internal task dependency, per PLAN DATA-06's own
Depends-on column: T2 depends on T1 (same helper); T3 depends on T1+T2 (migrate reference handler
onto the completed, actor-attributed helper); T4 depends on T1 (docs describe the implemented
contract). Depends on W00's exit gate at wave scope. No dependency on any other W02 story.

## Affected packages or components

`kernel/resource` (the new helper, the registrar's existing `Upsert` API, package documentation);
`kernel/resource/registrar_pg.go` (the actor-attribution fix); the reference handler (exact package
location to be confirmed at implementation time — PLAN's own text refers to "the reference handler"
without a fully-qualified path); `kernel/audit` and `kernel/outbox` (consumed by the new helper for
the audit-row and outbox-entry writes, not modified themselves).

## Compatibility considerations

PLAN's own wowsociety-impact note states the low-level `Upsert` API should remain available
alongside the new helper: "**Not breaking near-term** if wowapi keeps the low-level `Upsert` API
available alongside the new helper." This story's T1 accordingly adds the new helper as an
additional, preferred code path rather than removing the existing low-level `Upsert` API — a module
(including wowsociety's `committeeseat.go`) that has not yet migrated continues to function
unchanged. T3's reference-handler migration proves the new path works without breaking existing
reference tests; it does not require every other consumer to migrate immediately.

## Security considerations

T2's actor-attribution fix is a real, security-relevant fix, not cosmetic cleanup: replacing
`uuid.Nil` with a real, context-sourced actor and rejecting missing actors for user-initiated writes
closes an accountability gap where a mirror row could be written with no traceable author. This
directly supports the framework's compliance/audit posture (mirror rows feed into the same audit
trail `kernel/audit` and the DATA-08 work package are concerned with proving complete and durable).
T2's own risk note requires this fix "must not break legitimate system-actor call sites" — a
scheduled job or other system-initiated write must continue to succeed with its own system-actor
identity, not be forced through the user-initiated-actor-required path.

## Performance considerations

Bundling the mirror upsert, audit write, and outbox write into the same transaction as the business
write adds transactional scope compared to the current two-independent-statements pattern, but does
not introduce new I/O beyond what the framework's existing contract already implies (the mirror,
audit, and outbox writes already happen today, just not atomically) — no new performance concern
beyond what atomicity itself costs (holding a transaction open slightly longer to cover all four
writes), which is the correctness trade-off this story exists to make.

## Observability considerations

None mandated beyond what already exists. A fault-injection test failure at any of the 4 stages
should be clearly attributable to that stage in test output, so a future maintainer can diagnose
which write failed — a reasonable implementation-time expectation, not separately mandated by the
source.

## Migration considerations

None — this story adds application-layer tooling (the helper) and fixes an application-layer bug
(the nil-actor placeholder); it does not itself add or change any database schema or require a data
migration.

## Documentation requirements

Update `kernel/resource`'s package documentation to describe the mandatory-mirror contract as
implemented by T1/T2 — replacing the current manual, comment-only description with an accurate
statement of the enforced, atomic contract (T4's own scope).

## Acceptance criteria

- **AC-W02-E04-S001-01**: The typed aggregate repository/unit-of-work helper bundles aggregate
  write, mirror upsert, audit write, and outbox write atomically; fault injection at each of the 4
  stages independently causes full rollback at every stage, proven by a dedicated fault-injection
  test suite.
- **AC-W02-E04-S001-02**: `created_by` is sourced from context inside the same helper; a test with
  no actor present on a user-initiated write confirms the write fails fast; a test with a
  system-actor path confirms it remains unaffected and succeeds.
- **AC-W02-E04-S001-03**: The reference handler is migrated onto the new helper and no longer
  manually performs two independent statements; existing reference tests pass unmodified in
  observable behavior.
- **AC-W02-E04-S001-04**: `kernel/resource` documentation is updated to describe the implemented
  mandatory-mirror contract, confirmed via manual review against the actual implementation.

## Required artifacts

- The typed aggregate repository/unit-of-work helper (source code).
- The actor-attribution fix to `registrar_pg.go`.
- The migrated reference handler.
- Updated `kernel/resource` documentation.
See `artifacts/index.md`.

## Required evidence

- Fault-injection test output covering all 4 stages independently.
- Actor-attribution test output (with/without actor, system vs. user path).
- Reference-handler regression test output (existing tests still pass).
- Manual documentation-review record.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none within this
epic; internal task ordering T1→T2→T3/T4) recorded, owner/reviewer assignment pending, the
`registrar_pg.go` current-line-number and reference-handler-identity assumptions explicitly recorded
rather than silently assumed as fact.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T2 does not break any legitimate system-actor call
site and that the DATA-07 T3 cross-reference is documented clearly enough to be found and reused
without re-derivation.

## Risks

RISK-W02-E04-001 (the AR-03 overlap creating rework if AR-03's eventual design conflicts with this
story's T1 helper shape) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once T1's design decisions are documented explicitly (per RISK-W02-E04-001's mitigation) and T2's
system-actor-path test confirms no regression, residual risk is expected to be low for this story's
own closure — the AR-03 overlap risk itself remains open beyond this story's closure, tracked
forward to W05-E03, not resolved here.

## Plan

See `plan.md`.
