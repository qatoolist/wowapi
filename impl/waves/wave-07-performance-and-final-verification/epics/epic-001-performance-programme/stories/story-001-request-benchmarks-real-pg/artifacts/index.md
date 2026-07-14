---
id: W07-E01-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W07-E01-S001
status: implemented
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S001 — Artifacts index

Per mandate §9.2. All five produced artifacts are registered below.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W07-E01-S001-001 | perf/reference-v1.json + fixtures | configuration + fixtures | implementation | Full §14 reference environment and deterministic workload contract | PERF-02 | W07-E01-S001-T001 | `perf/reference-v1.json`; `perf/fixtures/request-workloads-v1.json` | produced |
| ART-W07-E01-S001-002 | DB-backed benchmark suite | test suite | implementation | 6 workload profiles through real HTTP/auth/authz/tenant-tx/PostgreSQL paths | PERF-02 | W07-E01-S001-T002 | `perf/requestbench/requests_bench_test.go` | produced |
| ART-W07-E01-S001-003 | Concurrency-matrix harness | test suite | implementation | Cold/warm × 1/10/100 tenant matrix and deterministic 100-tenant seed | PERF-02 | W07-E01-S001-T003 | `perf/requestbench/requests_bench_test.go`; `perf/requestbench/requests_contract_test.go` | produced |
| ART-W07-E01-S001-004 | Cost-breakdown instrumentation | source-code | implementation | OTel SQL count, phase timers, pool/lock wait, and plan hashes | PERF-02 | W07-E01-S001-T004 | `perf/requestbench/requests_bench_test.go` | produced |
| ART-W07-E01-S001-005 | Published comparison report | documentation + data | post-implementation | Pinned relative/container capture with DEC-Q9-conditional absolute framing | PERF-02 | W07-E01-S001-T005 | `perf/results/request-reference-v1.json`; `perf/results/request-reference-v1.txt` | produced |
