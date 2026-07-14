---
id: W01-E02-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W01-E02-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E02-S002 — Artifacts index

Per mandate §9.2. All artifacts are produced as working-tree source at HEAD
`0a31186cada5c275a588c74081cf977adf346e61` (conductor owns commits); repository paths below are the
canonical locations.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Status |
|---|---|---|---|---|---|---|---|
| ART-W01-E02-S002-001 | pgx.QueryTracer implementation | source-code package | implementation | `kernel/database/query_tracer.go` — `queryTracer` implementing `pgx.QueryTracer` over the observability port (`tracing.Tracer`, = `observability.Tracer` by alias, per DEV-W01-E02-S001-001): span-per-query named `db.<VERB>` from a closed verb set, `db.statement` (trimmed/truncated parameterized SQL, never `Args`), `db.rows_affected` on success, `RecordError` on failure, span carried start→end via a private context key | FBL-06 T3, D-08 | W01-E02-S002-T001 | produced |
| ART-W01-E02-S002-002 | Pool-config wiring diff (new Option) | source-code package | implementation | `WithQueryTracer(tr) Option` in the same file, following the `WithSetRole`/`WithConnRLSGuard` convention: sets `pc.ConnConfig.Tracer`; nil/NoOpTracer leaves the config untouched (zero-cost disabled path) | FBL-06 T3 | W01-E02-S002-T001 | produced |

## Notes

Supporting artifact: `kernel/database/query_tracer_test.go` (trace-tree, attrs, error-marking,
literal-leakage, root-span, sampling-inheritance integration tests). This story does not own the
D-08 ADR file (W00-E02-S003's artifact); its ratified wording was confirmed before implementation —
see `../implementation.md`.
