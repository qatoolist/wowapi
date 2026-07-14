---
id: W07-E01-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W07-E01-S002
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S002 — Artifacts index

Per mandate §9.2. All six required artifacts were produced from real implementation or real
PostgreSQL execution and accepted after a clean independent review.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W07-E01-S002-001 | Index-definition audit report | documentation | implementation | Confirms the active-only indexing claim from actual migration inspection | PERF-03 | W07-E01-S002-T001 | `impl/waves/wave-07-performance-and-final-verification/epics/epic-001-performance-programme/stories/story-002-rules-resolution-sql/artifacts/implementation/ART-W07-E01-S002-001-index-audit.md` | produced |
| ART-W07-E01-S002-002 | Set-based rules-resolution query | source-code | implementation | One statement replacing the per-ancestor loop | PERF-03 | W07-E01-S002-T002 | `kernel/rules/resolver.go` | produced and verified |
| ART-W07-E01-S002-003 | Current/historical indexes | schema migration | implementation | Confirms the active exclusion index and adds historical predicate coverage online | PERF-03 | W07-E01-S002-T003 | `migrations/00008_rules.sql`; `migrations/00048_rule_versions_resolution_indexes.sql` | produced and verified |
| ART-W07-E01-S002-004 | EXPLAIN fixture files | test fixtures | implementation | Real EXPLAIN ANALYZE BUFFERS for all four cardinality combinations | PERF-03 | W07-E01-S002-T004 | `perf/results/perf-03-explain-{shallow,deep}-{low,high}.json` | produced and verified |
| ART-W07-E01-S002-005 | Parity + SQL-count test suite | test suite | implementation | Real-PostgreSQL parity and 3/10/50 count comparison | PERF-03 | W07-E01-S002-T005 | `kernel/rules/resolver_perf_test.go` | produced and passing |
| ART-W07-E01-S002-006 | Published before/after comparison | documentation + data | post-implementation | DEC-Q9-honest comparison against the accepted reference policy and hash | PERF-03 | W07-E01-S002-T007 | `perf/results/perf-03-comparison.json` | produced and validated |
