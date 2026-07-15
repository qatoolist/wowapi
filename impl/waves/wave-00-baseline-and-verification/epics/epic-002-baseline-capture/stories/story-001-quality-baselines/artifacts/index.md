---
id: ARTIFACTS-W00-E02-S001
type: artifact-index
parent_story: W00-E02-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Artifact index — W00-E02-S001

Per mandate §9.2. All four expected artifacts produced 2026-07-13 at commit
`0a31186cada5c275a588c74081cf977adf346e61`. Raw capture files live under this story's
`artifacts/` category subdirectories (created on first real content per Adaptation 2).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Repository path or storage location | Version | Checksum | Status | Reviewer | Retention requirement |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-W00-E02-S001-001 | Coverage report | generated report | pre-implementation | `make coverage-check` run log (per-package + `total: 92.3%`, floor pass). Authoritative profile: `coverage.out`/`coverage.html` at repo root at the execution commit (build artifacts, referenced per the no-duplication rule, not copied into `impl/`). | PERF-06, REL-04 | W00-E02-S001-T001 | `coverage/coverage-check-output.txt` (this dir); `coverage.out`/`coverage.html` at repo root | point-in-time @ 0a31186 | n/a | **produced** | unassigned | Life of programme; superseded (not deleted) by later re-capture. |
| ART-W00-E02-S001-002 | Lint report (25-analyzer, MATRIX CS-23 diff) | static-analysis report | pre-implementation | Authoritative raw `golangci-lint run` output (text+JSON, 991 issues) under the throwaway config; anomalous run-1 output preserved; committed-config control run (0 issues); the throwaway config itself for reproducibility. Drift table lives in EV-W00-E02-S001-002. | FBL-05 (reference), FBL-07 (reference) | W00-E02-S001-T002 | `static-analysis/lint-25-analyzer-raw.{txt,json}`, `static-analysis/lint-25-analyzer-run1-anomalous.{txt,json}`, `static-analysis/lint-committed-config-raw.txt`, `static-analysis/golangci.matrix-cs23.throwaway.yml` | point-in-time @ 0a31186 | n/a | **produced** | unassigned | Life of programme; superseded once FBL-05 permanently enables analyzers. |
| ART-W00-E02-S001-003 | Bench-budget snapshot | generated report | pre-implementation | `make bench-budget` run output (43/43 OK, exit 0) + `bench-budgets.txt` entry-count confirmation (43, post-#25 — no drift). Authoritative budget file: `bench-budgets.txt` at repo root (committed, referenced not duplicated). | PERF-01, PERF-06 | W00-E02-S001-T003 | `benchmarks/bench-budget-output.txt`, `benchmarks/bench-budgets-entry-count.txt` | point-in-time @ 0a31186 | n/a | **produced** | unassigned | Life of programme as the W00 bench baseline reference. |
| ART-W00-E02-S001-004 | CI timing log | operational log | pre-implementation | Per-leg wall-clock of hosted run 29229288699 (headSha = execution commit) + recent-run list showing the pre/post-#23/#24/#25 timing shift. Data source: hosted GitHub Actions run history (preferred; local fallback not used). References `.github/workflows/ci.yml` at the execution commit rather than duplicating it. | SD-01, SD-02 | W00-E02-S001-T003 | `ci-timing/ci-wallclock-run-29229288699.txt` | point-in-time @ 0a31186 | n/a | **produced** | unassigned | Life of programme; superseded (not deleted) by later re-capture. |
