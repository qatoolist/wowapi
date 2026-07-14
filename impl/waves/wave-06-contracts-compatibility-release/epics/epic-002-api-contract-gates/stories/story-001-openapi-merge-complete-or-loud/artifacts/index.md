---
id: W06-E02-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W06-E02-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E02-S001 — Artifacts index

Artifacts are implemented in the registered source, fixture, decision, and workflow paths.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W06-E02-S001-001 | Expanded merge struct with per-field policy | source-code change | implementation | Complete-or-loud policies for every OpenAPI 3.1 top-level and components field | DX-06 | W06-E02-S001-T001 | `internal/cli/openapi_merge.go` | produced |
| ART-W06-E02-S001-002 | Fixture-driven per-field test suite | test suite | implementation | Field-policy, validation, additive, and breaking fixtures | DX-06 | W06-E02-S001-T001 | `internal/cli/openapi_contract_test.go`; `internal/cli/testdata/openapi-diff/` | produced |
| ART-W06-E02-S001-003 | Structural validator wiring and decision record | source + decision | implementation | OpenAPI 3.1.1/JSON Schema 2020-12 structural validation and dependency review | DX-06 | W06-E02-S001-T002 | `internal/cli/openapi_merge.go`; `evidence/security/validator-dependency-review.md` | produced |
| ART-W06-E02-S001-004 | Semantic-diff CI gate | CI configuration | implementation | libopenapi semantic breaking classifier exercised by adversarial fixtures | DX-06 | W06-E02-S001-T003 | `internal/cli/openapi_diff.go`; `.github/workflows/compatibility-gates.yml` | produced |
