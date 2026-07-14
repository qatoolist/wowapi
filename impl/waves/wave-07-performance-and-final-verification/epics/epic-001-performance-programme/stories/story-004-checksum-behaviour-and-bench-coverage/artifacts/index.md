---
id: W07-E01-S004-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W07-E01-S004
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S004 — Artifacts index

Per mandate §9.2. Source paths below are the produced, reviewable artifacts; the
machine-readable inventory and comparison are published under `perf/results/`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W07-E01-S004-001 | Checksum-required enforcement + call-site audit | source-code + audit report | implementation | Every upload call path persists checksum metadata | PERF-05 | W07-E01-S004-T001 | `foundation/document/service.go`; `kernel/storage/storage.go`; `perf/results/perf-05-checksum-inventory-v1.json` | produced |
| ART-W07-E01-S004-002 | Bounded repair path | source-code | implementation | Full-hash fallback reachable only via labeled invocation | PERF-05 | W07-E01-S004-T002 | `adapters/storage/s3/s3.go`; `adapters/storage/s3/checksum_repair_test.go` | produced |
| ART-W07-E01-S004-003 | Fallback-invocation metrics | source-code | implementation | Counter/histogram for hits, bytes, duration | PERF-05 | W07-E01-S004-T003 | `kernel/observability/metrics.go`; `adapters/metrics/prometheus/prometheus.go`; `adapters/storage/s3/checksum_repair_test.go` | produced |
| ART-W07-E01-S004-004 | Resumable backfill mechanism | source-code | implementation | Interrupt/resume-safe legacy-object backfill | PERF-05 | W07-E01-S004-T004 | `adapters/storage/s3/backfill.go`; `adapters/storage/s3/checksum_repair_test.go` | produced |
| ART-W07-E01-S004-005 | Published before/after comparison | documentation + data | post-implementation | Comparison against perf/reference-v1.json | PERF-05 | W07-E01-S004-T005 | `perf/results/perf-05-comparison-v1.json` | produced |
| ART-W07-E01-S004-006 | 7 new benchmark files + bench-budget entries | test suite + configuration | implementation | kernel/database, jobs, outbox, workflow, auth, mfa, httpclient | CS-16 | W07-E01-S004-T006 | `kernel/{database,jobs,outbox,workflow,auth,mfa,httpclient}/*bench_test.go`; `bench-budgets.txt`; `Makefile` | produced |
