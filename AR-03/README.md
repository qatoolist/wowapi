# AR-03 — Authoritative Declaration & Derived Projections

## What is this directory?

AR-03 is an architectural requirement for **one authoritative declaration of the framework's application model**, with derived automated projections and consistency checks. The directory contains Go tests that validate:

- **Manifest schema & structure** (test fixtures asserting the model's shape)
- **Golden declaration delta** (changes to the model vs. a baseline)
- **Full projection** (all derived views are consistent with the source declaration)
- **Duplicate detection** (linting for omission or collision in derived names)

Tests run as part of `make test` and enforce model consistency at compile/test time.

## Which requirements/stories reference this?

- **Wave 05, Epic 03** (`W05-E03-S001..S002`): the plan for executing this requirement
- **Wave 06, Epic 02, Story 001** (`W06-E02-S001`): compatibility gates consume AR-03 T2 scope (delegated to owner DX-06)
- **Wave 06, Epic 04, Story 002** (`W06-E04-S002`): composition/doc drift removal depends on AR-03

Cross-references in the implementation ledger: `impl/index.md` (wave allocation), `impl/analysis/requirement-inventory.md`, `impl/analysis/conflict-resolution.md` (AR-03 T2 vs DX-06 ownership).

## Why does AR-03 live at the repo root?

AR-03 tests are pinned to exact framework paths in their evidence records (`impl/waves/wave-05.../epic-003/.../evidence/index.md`). **Relocating this directory would break all pinned evidence references** in the implementation programme and require a complete evidence sweep and re-pin.

Relocation is tracked as a **future refactoring**, not a blocker. It requires:
1. Update all AR-03 artifact/evidence paths in `impl/waves/wave-05.../` 
2. Update cross-wave references (W06-E02-S001, W06-E04-S002 evidence pinning)
3. Re-run and re-pin evidence bundles

## Tests in this directory

Run with:
```bash
go test ./AR-03 -v
```

- `duplicate_omission_lint_test.go` — lints for collision or omission in derived projection names
- `full_projection_golden_test.go` — validates all derived projections remain consistent
- `golden_declaration_delta_test.go` — diffs model changes against a stored baseline
- `manifest_schema_fixture_test.go` — asserts schema structure of the authoritative declaration

All tests run with the real database (gated on `WOWAPI_REQUIRE_DB`).
