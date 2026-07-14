---
id: PLAN-W02-E02-S001
type: plan
parent_story: W02-E02-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W02-E02-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information.

## Proposed architecture

Two independent, additive pieces layered on the existing schema and CI infrastructure: (1) four
`CONCURRENTLY`-built unique-index migrations on the parent tables the composite FKs will eventually
reference, and (2) a tenant-FK catalog scanner tool wired into CI as a permanent gate. Neither piece
changes any existing table's data, RLS policy, or application code path — both are purely additive
schema/tooling changes.

## Implementation strategy

1. Confirm, per parent table (`parties`, `organizations`, `documents`, `document_versions`), whether
   a `UNIQUE (tenant_id, id)` index already exists, via `pg_indexes` at this story's actual start
   commit.
2. For each parent lacking the index, write a `CONCURRENTLY`-built migration adding
   `UNIQUE (tenant_id, id)`, run non-transactionally per T1's own risk note.
3. Write a migration test querying `pg_indexes` post-migration to confirm all 4 parents carry the
   index.
4. Investigate what constitutes "the existing RLS-tagged tenant-table matrix" (T2's own risk note)
   that the scanner should key off, rather than a hand-maintained list — record the finding.
5. Implement the tenant-FK catalog scanner: enumerate every tenant-scoped table's foreign keys,
   flagging any not composite on `(tenant_id, …)`, using the matrix identified in step 4.
6. Write a fixture-schema test confirming the scanner enumerates exactly the 8 known FKs with zero
   silent gaps.
7. Wire the scanner into CI as a permanent gate (extending the existing CI infrastructure, consistent
   with W02-E01-S001's own manifest-schema validation wiring).
8. Write a negative fixture migration (a migration adding a single-column, non-composite tenant FK)
   and confirm the CI gate actually fails it.
9. Document the scanner's purpose, its matrix-keying mechanism, and the CI gate's failure behavior.

## Expected package or module changes

Four new or modified migration files (one per parent table). A new tenant-FK catalog scanner tool
(exact package location TBD — expected near W02-E01-S001's manifest-schema validator, e.g. under
`internal/tools/` or a migration-tooling-adjacent package). Extended CI configuration to invoke the
scanner as a gate.

## Expected file changes where determinable

- 4 new `CONCURRENTLY` unique-index migrations under the existing migration directory.
- A new tenant-FK catalog scanner tool (exact file path TBD).
- CI configuration (e.g. `.github/workflows/`) extended to invoke the scanner.
- A new negative fixture migration for the CI gate test.

## Contracts and interfaces

The catalog scanner's own output contract: a list of tenant-scoped-table foreign keys, each flagged
as composite-compliant or not, keyed off the RLS-tagged tenant-table matrix identified in step 4
above. Exact typing/serialization format TBD at implementation time.

## Data structures

None beyond the scanner's internal FK-enumeration data structure (exact shape TBD). No application
data model change — the parent unique indexes add no new column, only an index.

## APIs

None affected — this story is schema/tooling-internal, not a runtime API change.

## Configuration changes

None anticipated for the unique-index migrations. The CI gate wiring may require a new CI
configuration entry (exact form TBD, consistent with however W02-E01-S001's manifest validation is
wired in).

## Persistence changes

Four new unique indexes on `parties`, `organizations`, `documents`, `document_versions` — additive
only, no column or row change.

## Migration strategy

Each of the 4 index migrations is built `CONCURRENTLY` and run non-transactionally, per T1's own risk
note ("`SHARE UPDATE EXCLUSIVE` lock — must run non-transactionally"). No backfill or data
transformation is required.

## Concurrency implications

`CONCURRENTLY` index builds do not block concurrent reads/writes on the parent tables, but they do
take a `SHARE UPDATE EXCLUSIVE` lock that conflicts with certain other concurrent DDL — this is the
reason T1's own risk note calls out non-transactional execution as a requirement, not an option.

## Error-handling strategy

If a `CONCURRENTLY` index build fails partway (e.g. due to a conflicting lock or a duplicate-key
violation surfaced during build), PostgreSQL leaves an invalid index behind that must be dropped and
retried — this story's migration tooling should surface that failure clearly rather than silently
leaving an invalid index in place. The CI gate must fail with a field/migration-specific error
message identifying exactly which FK triggered the rejection, not a generic failure.

## Security controls

None distinct from this story's own scope — the scanner's "zero silent gaps" requirement (T2's
acceptance criterion) is itself the primary correctness control this story delivers; an incomplete
enumeration would be a security-relevant defect (a future non-composite tenant FK could land
undetected), not merely a quality gap.

## Observability changes

The CI gate's failure output should identify the specific migration and FK that triggered rejection
(implementation-time addition, consistent with this story's own "Documentation requirements").

## Testing strategy

- Migration test: confirm `pg_indexes` shows `UNIQUE (tenant_id, id)` on all 4 parents post-migration.
- Fixture-schema test: confirm the scanner enumerates exactly the 8 known FKs with zero silent gaps
  against a fixture schema.
- Negative fixture migration test: confirm the CI gate actually fails a migration adding a
  single-column, non-composite tenant FK.

## Regression strategy

The CI gate itself, once wired, is the regression guard — any future migration reintroducing a
non-composite tenant FK fails CI going forward, preventing the DATA-01 gap from silently reopening.

## Compatibility strategy

Purely additive — no existing behavior changes. The CI gate's enforcement is immediate upon landing
(no transition period is described in the source for T6, unlike W02-E01-S001's own manifest-gate
timing question); this plan proceeds on that reading, to be confirmed at implementation time.

## Rollout strategy

Single story, landed as its own reviewable unit. T1 (parent indexes) and T2 (scanner) may proceed in
parallel — they are disjoint code surfaces; T6 (CI gate wiring) depends on T2 per PLAN's own
Depends-on column and should be sequenced promptly after T2 per PLAN's own risk note ("do first if
sequencing allows").

## Rollback strategy

The parent unique indexes can be dropped without touching any other data if a legitimate conflict is
found. The CI gate can be reverted (disabled) if it produces a false-positive rejection of a
legitimate migration — but any such reversion must be recorded as a deviation, not silently applied,
given the gate's durability requirement.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–9). Steps 1–3 (T1) and steps 4–6 (T2) may
proceed in parallel; steps 7–8 (T6) depend on step 5 (the scanner existing).

## Task breakdown

- **W02-E02-S001-T001** — Parent tenant-scoped unique indexes (PLAN DATA-01 T1; steps 1–3 above).
- **W02-E02-S001-T002** — Tenant-FK catalog scanner (PLAN DATA-01 T2; steps 4–6 above).
- **W02-E02-S001-T003** — CI gate wiring (PLAN DATA-01 T6; steps 7–9 above).
- **W02-E02-S001-T004** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The 4 parent unique-index migrations; the tenant-FK catalog scanner tool; the CI gate wiring; the
negative fixture migration; documentation of the scanner and gate.

## Expected evidence

`pg_indexes` migration test output; fixture-schema catalog-scanner test output; negative fixture
migration CI run output.

## Unresolved questions

- Which of the 4 parent tables (if any) already carry `UNIQUE (tenant_id, id)` — to be confirmed via
  `pg_indexes` at implementation time (T1 says "add/confirm," not merely "add").
- What exactly constitutes "the existing RLS-tagged tenant-table matrix" (T2's own risk note) the
  scanner should key off — to be investigated and recorded at implementation time, not invented here.
- Exact package location for the scanner tool.
- Exact CI configuration mechanism for wiring in the gate.

## Approval conditions

This plan is approved for implementation once: (a) the `pg_indexes` per-parent confirmation (step 1)
has been performed and its result recorded, and (b) the owner and reviewer are assigned.
