---
id: IMPL-W02-E02-S001
type: implementation-record
parent_story: W02-E02-S001
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W02-E02-S001

## What was actually implemented

- Four `CONCURRENTLY`-built unique indexes on the parent tables referenced by the 8 DATA-01 edges.
- A tenant-FK catalog scanner (`internal/tools/tenantfk`) keyed off the live RLS-tagged tenant-table matrix (public tables with RLS enabled and a `tenant_id` column).
- A CI gate (`make tenantfk-gate`) that fails any post-DATA-01 migration adding a non-composite tenant FK.
- A negative fixture migration proving the gate rejects a single-column tenant FK.

## Components changed

- `migrations/` — 3 new SQL migration files.
- `internal/tools/tenantfk/` — new scanner tool and tests.
- `migrations/migrations_test.go` — registered new migration files.
- `Makefile` — added `tenantfk-gate` target.
- `.github/workflows/ci.yml` — added `tenantfk-gate` job.

## Files changed

- `migrations/00034_tenant_fk_parent_indexes.sql`
- `migrations/00035_tenant_fk_composite_not_valid.sql`
- `migrations/00036_tenant_fk_validate_and_cleanup.sql`
- `migrations/migrations_test.go`
- `internal/tools/tenantfk/main.go`
- `internal/tools/tenantfk/scanner.go`
- `internal/tools/tenantfk/parse.go`
- `internal/tools/tenantfk/scanner_test.go`
- `internal/tools/tenantfk/testdata/bad_fk_migration.sql`
- `Makefile`
- `.github/workflows/ci.yml`

## Interfaces introduced or changed

- New CLI: `tenantfk enumerate --dsn=...` and `tenantfk gate --dsn=... --migrations=... [--since=N]`.
- New make target: `make tenantfk-gate`.

## Configuration changes

None.

## Schema or migration changes

- `00034`: adds `UNIQUE (tenant_id, id)` indexes on `parties`, `organizations`, `documents`, `document_versions` via `CREATE UNIQUE INDEX CONCURRENTLY`.
- `00035`: adds 8 composite FKs `NOT VALID` for the DATA-01 edges.
- `00036`: validates the 8 composite FKs and drops the now-redundant single-column FKs.

All three migrations carry validated `+wowapi:manifest` blocks.

## Security changes

- Composite tenant FKs enforce that child and parent rows agree on `tenant_id`, closing the CS-18 cross-tenant reference gap.
- The CI gate prevents future migrations from reintroducing single-column tenant FKs on tenant-scoped tables.

## Observability changes

- `tenantfk enumerate` prints a tabular report of tenant-scoped FKs and composite status.
- `tenantfk gate` prints file/constraint/reason for each violation.

## Tests added or modified

- `internal/tools/tenantfk/scanner_test.go`: parser unit test, fixture-schema enumerate test (8 edges, zero gaps), negative gate fixture test.
- Existing `testkit/tenant_fk_cross_tenant_test.go` now passes against the migrated schema.
- Existing `migrations/reversible_test.go` passes with the new migrations.
- Existing `migrations/migrations_test.go` updated to expect the 3 new files.

## Commits

Tracked under this implementation session; no separate commit list maintained.

## Pull requests

Not created in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

- The scanner parses migration SQL with a purpose-built tokenizer; hand-authored exotic DDL may require parser updates, but all current kernel migration patterns are covered.
- The gate defaults to checking migrations with version > 36 (`--since=36`) so existing pre-DATA-01 migrations are not retroactively flagged.

## Follow-up items

- Monitor CI behavior of the new `tenantfk-gate` job on the first few PRs.
- W02-E02-S002-T3 mismatch audit tool was not implemented in this session; it remains scoped to the sibling story if required.

## Relationship to the approved plan

Matched plan.md. Build-readiness note: at the start of the session `kernel/document/service.go` was observed not to build, blocking testkit compilation. Before any patch was applied, the file was already fixed (import aliases present), so no deviation was required.
