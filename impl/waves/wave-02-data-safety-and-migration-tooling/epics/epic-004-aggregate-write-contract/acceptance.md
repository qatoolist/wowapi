---
id: W02-E04-ACCEPTANCE
type: epic-acceptance
epic: W02-E04
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E04 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W02-04 there maps onto this epic).

## AC-W02-E04-01 — Aggregate write atomicity enforced

The typed aggregate repository/unit-of-work helper bundles aggregate write, mirror upsert, audit,
and outbox writes atomically; fault injection at each of the 4 stages independently, with full
rollback confirmed at every stage, proves a module cannot write its business row without the
framework also writing the mirror in the same transaction. Traces to W02-E04-S001 (T1).

## AC-W02-E04-02 — Actor attribution sourced and enforced; single-owner fix recorded

`created_by` is sourced from context in the same helper; a test with/without actor confirms a
user-initiated write with no actor fails fast while system-actor paths remain unaffected. This fix
is documented as the single owner of the `registrar_pg.go` nil-actor placeholder shared with DATA-07
T3 (PLAN's own cross-cutting note 2), with an explicit pointer for W03-E04-S001 to consume rather
than reimplement. Traces to W02-E04-S001 (T2).

## AC-W02-E04-03 — Reference handler migrated

The reference handler no longer manually performs two independent statements (business-row write,
mirror upsert) — it is migrated onto the new helper, with existing reference tests passing
unmodified in behavior. Traces to W02-E04-S001 (T3).

## AC-W02-E04-04 — Documentation matches implementation

`kernel/resource` documentation is updated to describe the mandatory-mirror contract as implemented
by T1/T2, confirmed via manual review — stale documentation (the same class of gap PLAN's evidence
cites as having "created this defect class") is not left in place. Traces to W02-E04-S001 (T4).

## AC-W02-E04-05 — Independent review passed

The story has passed independent review per mandate §14, specifically confirming: T2's actor-
attribution fix does not break any legitimate system-actor call site (per PLAN T2's own risk note);
the AR-03 overlap (RISK-W02-E04-001) is recorded, not silently ignored; and the DATA-07 T3
cross-reference is documented clearly enough for a W03 implementer to find it without re-deriving
the design.

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA).
