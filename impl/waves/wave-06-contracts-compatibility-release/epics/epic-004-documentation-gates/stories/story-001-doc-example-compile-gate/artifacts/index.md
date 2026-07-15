---
id: W06-E04-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W06-E04-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E04-S001 — Artifacts index

Per mandate §9.2, every planned artifact was produced and registered at its repository path.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W06-E04-S001-001 | docexamples extractor tool | source-code package | implementation | Scans, classifies, extracts, and compiles normative Go doc examples in isolated throwaway packages | AR-05 | W06-E04-S001-T001 | `internal/tools/docexamples/` | produced |
| ART-W06-E04-S001-002 | Marker convention applied to existing examples | documentation change | implementation | `compile` tags identify normative standalone source; `illustrative` tags explicitly classify pseudo-code/signature excerpts | AR-05 | W06-E04-S001-T001 | `docs/blueprint/*.md` (convention documented in `11-framework-distribution-and-consumption.md`) | produced |
| ART-W06-E04-S001-003 | make docs-check target + CI wiring | build/CI configuration | implementation | Local and CI invocation of the same documentation contracts gate | AR-05 | W06-E04-S001-T002 | `Makefile`; `.github/workflows/ci.yml` unit job | produced |
| ART-W06-E04-S001-004 | Adversarial staled-example fixture | test fixture | implementation | Removed `app.RunAPI` symbol proves the gate rejects stale docs at the original Markdown location | AR-05 | W06-E04-S001-T002 | `internal/tools/docexamples/testdata/stale-example.md` | produced |
