# Implementation-Readiness Checklist (Goal 2 preflight item 2)

| Item | Status | Evidence |
|---|---|---|
| Public package graph finalized | ✅ | blueprint 04 §1–2 (incl. `kernel/secrets`, D-0001/D-0005); encoded in `scripts/lint_boundaries.sh` |
| Config contracts finalized | ✅ | blueprint 12 §2 + D-0002/D-0003 (`config.Framework` / product `internal/appcfg.Config` / `ModuleView` / generated `tools/configcheck`) |
| External consumer test strategy defined | ✅ | test-strategy.md §External-consumer row (scratch repo via `wowapi init` + `go mod edit -replace`, from Phase 5) |
| Container workflow defined | ✅ | Dockerfile + deployments/compose.yaml (pg/minio/mailpit/tools runner); test-strategy.md §Container execution. NOTE: live run deferred at Phase 0 — Docker daemon unavailable on author machine (risk R4, command-log) |
| Makefile target list finalized | ✅ | Makefile (Goal 2 §Makefile list, all targets present; unbuilt phases fail with explicit guidance) |
| First walking-skeleton scope defined | ✅ | phase-plan.md Phase 0 row + D-0006/D-0008; deep features explicitly deferred (preflight rule 5) |
| Blueprint inconsistencies from Goal 2 preflight resolved | ✅ | D-0001…D-0005; blueprint diffs in commit |
