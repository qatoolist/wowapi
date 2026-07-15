---
id: ARTIFACTS-INDEX-W00-E01-S002
type: artifact-index
parent_story: W00-E01-S002
status: recorded
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Artifact index — W00-E01-S002

Per `artifact-policy.md` §9.2. All artifacts below were produced on 2026-07-13 at commit
`0a31186cada5c275a588c74081cf977adf346e61` and live directly under this story's `artifacts/`
directory (execution logs and inspection notes; no lifecycle-stage subdirectories were needed beyond
first real content landing here).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Repository path or storage location | Version | Checksum | Status | Reviewer | Retention requirement |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-W00-E01-S002-001 | `go test ./kernel/httpx/... -race` + `make bench-budget` output logs | execution log | post-implementation | Console output capturing PERF-01's race-test (exit 0, no races) and bench-budget-gate (exit 0, 43/43 OK) re-run results | PERF-01 | W00-E01-S002-T001 | `artifacts/T001-go-test-kernel-httpx-race.log`; `artifacts/T001-make-bench-budget.log` | 0a31186 | sha256 `7887dcc41ec89ea620bd337d9c662d4c88dda2ee9e4a8757d9227851eb7150fd`; `5fdd45bf8cc951e0e42e5e33c551f0b96145c80fc09baafcd37802c228880745` | produced | pending (conductor acceptance gate) | Retain per `evidence-policy.md` alongside EV-W00-E01-S002-01; not superseded until a later re-run replaces it |
| ART-W00-E01-S002-002 | Benchbudget missing-benchmark coverage-test + fail-first gate-check output logs | execution log | post-implementation | Console output capturing PERF-06 T1's subprocess exit-1 assertion (PASS) and the fail-first revert-proof ghost-entry gate check (gate exit 1, missing benchmark named), plus the scratch budgets file used | PERF-06 | W00-E01-S002-T002 | `artifacts/T002-go-test-benchbudget-missingbenchmark.log`; `artifacts/T002-failfirst-ghost-check.log`; `artifacts/T002-scratch-budgets-ghost.txt` | 0a31186 | sha256 `91836074fd16eac004952357bc3de2c974b68d22147b040483a68aff0a3f35d1`; `438bf32fad8952769cadcfffed9677f49c9c8a7531d38ac1f9b8e89a8c5303c8`; `45d7f4e1d2e7c8cb8a01a8b1f6af03fad4719615278e29d5acf0f89a522eca43` | produced | pending (conductor acceptance gate) | Retain per `evidence-policy.md` alongside EV-W00-E01-S002-02 |
| ART-W00-E01-S002-003 | `bench-budgets.txt` entry-count and spot-check inspection note | inspection note | post-implementation | Note recording the confirmed 43-entry count (manual + tool-reported reconciled), byte-identity with commit `0a31186`, and 3-entry spot-check, closing out RISK-W00-003 | SD-03 | W00-E01-S002-T003 | `artifacts/T003-bench-budgets-inspection-note.md` | 0a31186 | sha256 `649fadfb395ce4bef3686550660320fbec12a1751040a4459aad8ac057cb16c3` | produced | pending (conductor acceptance gate) | Retain per `evidence-policy.md` alongside EV-W00-E01-S002-03 |

No large generated or binary artifact was produced; each artifact above is itself the authoritative
output of its producing task's command (no-duplication rule satisfied).
