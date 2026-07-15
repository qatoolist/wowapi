---
id: W03-E04-ACCEPTANCE
type: epic-acceptance
epic: W03-E04
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E04 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" as a standalone,
independently-referenceable record, consistent with the wave-level `../../acceptance.md` pattern
(AC-W03-07 there maps onto this epic).

## AC-W03-E04-01 — Party-subject edge evaluation

`Checker.Has` resolves actor → active capacity → optional party through the post-SEC-01
authoritative principal model. A test seeds a party-subject edge, resolves an actor carrying a
party, and asserts the previously-false evaluation is now correctly `true`.

## AC-W03-E04-02 — Full subject-kind matrix

Every schema-enumerated `subject_kind` has an explicit evaluation branch in `Checker.Has`; a matrix
test confirms every enumerated kind; an unsupported/unenumerated kind fails closed, not silently
`true` or silently ignored.

## AC-W03-E04-03 — Mutation ownership, attribution, audit, versioning, cache invalidation

Every authorization-input mutation (edge create/revoke) is ownership-checked, attributed (via
DATA-06 T2's mechanism, consumed via its shared file, not reimplemented in this epic), writes an
audit row, and is versioned. The cache-invalidation portion of this criterion is explicitly
deferred-linked to W05-E04-S002 and tracked as such (not silently dropped, not silently assumed
complete) if W05-E04-S002 has not landed by this epic's own closure time.

## AC-W03-E04-04 — Independent review passed

W03-E04-S001 has passed independent review per mandate §14, with specific confirmation that T3's
scope was correctly cross-referenced to DATA-06 T2 (W02-E04-S001) rather than reimplemented in this
epic — checking for accidental duplicate ownership of the `registrar_pg.go` fix, per PLAN's own
"High duplication risk if staffed independently" warning.

## Acceptance authority

Data/reliability lead (PLAN §5.3's stated accountable role for PF-DATA) jointly with the
product-security lead, given this epic's hard dependency on SEC-01 — per `../../wave.md`'s
accountable-role assignment for W03-E04.
