---
id: IMPL-W02-E02-S002
type: implementation-record
parent_story: W02-E02-S002
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W02-E02-S002

*This record reflects the portion of S002 implemented in this session, scoped to the user's explicit goal: composite tenant foreign keys and the adversarial test matrix.*

## What was actually implemented

- Composite tenant FKs added `NOT VALID` (migration `00035`) and validated (migration `00036`) for all 8 DATA-01 edges.
- Redundant single-column FKs removed in `00036` after verification that they carried `NO ACTION` on delete/update.
- Existing adversarial test matrix (`testkit/tenant_fk_cross_tenant_test.go`) now passes: cross-tenant inserts are blocked under admin (BYPASSRLS), `app_rt`, and `app_platform`.
- Edge-census test confirms the live set of composite tenant FKs matches the hand-written 8-edge matrix.

## Components changed

- `migrations/` — schema migrations `00035` and `00036`.

## Files changed

- `migrations/00035_tenant_fk_composite_not_valid.sql`
- `migrations/00036_tenant_fk_validate_and_cleanup.sql`
- `migrations/migrations_test.go` (registration of new files).

## Interfaces introduced or changed

None.

## Configuration changes

None.

## Schema or migration changes

- `00035`: per-table `ALTER TABLE ... ADD CONSTRAINT ... FOREIGN KEY (tenant_id, <ref>) REFERENCES <parent> (tenant_id, id) NOT VALID` for the 8 edges.
- `00036`: `ALTER TABLE ... VALIDATE CONSTRAINT ...` for each composite FK, then `DROP CONSTRAINT` on the 8 redundant single-column FKs.

## Security changes

- Cross-tenant parent/child inserts now fail with SQLSTATE 23503 (`foreign_key_violation`) under the BYPASSRLS admin role, proving referential integrity — not RLS — enforces tenant agreement.
- Platform role (`app_platform`) is confirmed not to bypass the composite FK constraints.

## Observability changes

None beyond the migration manifests.

## Tests added or modified

- Existing `testkit/tenant_fk_cross_tenant_test.go` passes end-to-end.
- Existing `migrations/reversible_test.go` passes with the new migrations.

## Commits

Tracked under this implementation session.

## Pull requests

Not created in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

- The mismatch-audit tool (S002-T3) was not implemented in this session; it is scoped to the user's explicit goal which focused on composite FKs and the adversarial matrix.
- The W02-E01 online-migration-protocol acceptance gate is recorded as a planning dependency; the migrations here were implemented and tested against the local test DB.

## Follow-up items

- Implement the mismatch-audit tool if S002 full closure is required.
- Run a real mismatch audit against staging/prod-shaped data before any production validation of the composite FKs.

## Relationship to the approved plan

Matched the composite-FK and negative-test portions of plan.md. The mismatch audit and optional redundant-FK cleanup are documented as intentionally scoped.
