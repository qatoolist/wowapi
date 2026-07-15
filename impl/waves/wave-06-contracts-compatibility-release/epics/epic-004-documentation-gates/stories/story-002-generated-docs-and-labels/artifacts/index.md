---
id: W06-E04-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W06-E04-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E04-S002 — Artifacts index

Per mandate §9.2, both planned artifacts were produced and registered at their repository paths.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W06-E04-S002-001 | Reference-doc-generation pipeline | source-code + generated docs | implementation | Generates and byte-checks a reference table from W05 AR-03's authoritative `appmodel.GenerateProjections` export | AR-05 | W06-E04-S002-T001 | `internal/tools/docexamples/reference.go`; `docs/reference/application-model.md` | produced |
| ART-W06-E04-S002-002 | Future-state-labeling lint | source-code tool | implementation | Fails future/planned/target design blocks without the `Target, not implemented` label in blueprints and `*-target-design.md` documents | AR-05 | W06-E04-S002-T002 | `internal/tools/docexamples/future.go`; `internal/tools/docexamples/testdata/future-*.md` | produced |
