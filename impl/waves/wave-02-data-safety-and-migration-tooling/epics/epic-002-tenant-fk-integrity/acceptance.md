---
id: W02-E02-ACCEPTANCE
type: epic-acceptance
epic: W02-E02
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E02 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W02-02 there maps onto this epic).

## AC-W02-E02-01 — Parent unique indexes, catalog scanner, and CI gate in place

`UNIQUE (tenant_id, id)` exists, built `CONCURRENTLY`, on `parties` (parent of persons/
legal_entities/party_contacts/acting_capacities), `organizations` (parent of `resources`), and
`documents`/`document_versions` (parents of document_versions/document_access_grants/attachments).
The tenant-FK catalog scanner enumerates exactly the 8 known FKs with zero silent gaps, keyed off
the existing RLS-tagged tenant-table matrix rather than a hand-maintained list (PLAN T2's own risk
note). The scanner is wired as a permanent CI gate: a new migration adding a single-column tenant FK
fails CI, proven by a negative fixture migration. Traces to W02-E02-S001.

## AC-W02-E02-02 — Mismatch audit clean and composite FKs validated

The mismatch audit (S002-T3) reports zero cross-tenant mismatches against staging/prod-shaped data,
using a platform-role connection to bypass RLS for the scan — or, if a mismatch was found, a
documented and resolved remediation decision exists per RISK-W02-002 before proceeding. All 8 edges
carry a composite FK added `NOT VALID` (T4) and subsequently `VALIDATE CONSTRAINT`-clean (T5) under
a concurrent-writer-load test, with both steps started only after W02-E01-S001 and W02-E01-S002
reach `accepted`. Traces to W02-E02-S002.

## AC-W02-E02-03 — Cross-tenant negative tests pass under both roles

A seeded cross-tenant insert fails under both `app_rt` and `app_platform` roles, confirmed by a new
catalog-driven RLS matrix test — explicitly confirming the platform role does not bypass the new FK
constraints, per PLAN T7's own risk note not to assume this. Traces to W02-E02-S002.

## AC-W02-E02-04 — Independent review passed

Both stories (S001, S002) have passed independent review per mandate §14. S002's review specifically
confirms the T4/T5 gate on W02-E01 acceptance was genuinely honored — checked against timestamps/
commit history, not merely asserted — and that the mismatch-audit outcome is honestly recorded
whichever way it resolved (zero-mismatch, or a documented remediation).

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA).
