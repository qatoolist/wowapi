---
id: W07-E02-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W07-E02-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E02-S002 — Artifacts index

Per mandate §9.2. Repository source/configuration is retained for the life of the repository; generated
fuzz proof is retained by the workflow for 14 days (PR) or 30 days (scheduled).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Version | Checksum | Status | Reviewer | Retention |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-W07-E02-S002-001 | Fail-not-skip E2E + classification | CI/test configuration + classification record | implementation | Historical 22-site requirement re-scanned at execution: 39 found, flaky TOTP skip removed, all remaining 38 classified; required DB/S3/E2E branches fail closed | REL-04 | W07-E02-S002-T001 | `artifacts/implementation/skip-site-classification.md`; `.github/workflows/ci.yml`; `miscellaneous/check_required_test_prerequisites.sh` | 1 | not applicable (repository source) | reviewed | W05ReviewGateFinal | repository lifetime |
| ART-W07-E02-S002-002 | Machine-checked skip manifest | source code + configuration | implementation | Go-AST manifest validator, exhaustive owner/rationale manifest, positive and negative fixtures | REL-04 | W07-E02-S002-T002 | `miscellaneous/test-skip-manifest.json`; `internal/tools/testskipmanifest/`; `miscellaneous/check_test_skips.sh`; `miscellaneous/check_test_skip_fixtures.sh` | 1 | `sha256:563b9c33088bac783042a25523af67e832fe215f5131ac5b2dd43c7844ed5205` (manifest) | reviewed | W05ReviewGateFinal | repository lifetime |
| ART-W07-E02-S002-003 | DB/S3 integration race CI | CI configuration + negative fixture | implementation | Per-change integration race leg over DB/S3-backed packages, preceded by a seeded detector fixture | REL-04 | W07-E02-S002-T003 | `.github/workflows/ci.yml`; `Makefile`; `miscellaneous/check_race_detector.sh`; `internal/verificationfixtures/racefixture/` | 1 | not applicable (repository source) | reviewed | W05ReviewGateFinal | repository lifetime |
| ART-W07-E02-S002-004 | Real PR + scheduled fuzz CI | CI configuration + proof runner | implementation | Short PR and 1-minute-per-target scheduled coverage-guided profiles; generated corpus retained through versioned cache keys and proof artifacts | REL-04, PERF-06 | W07-E02-S002-T004 | `.github/workflows/ci.yml`; `Makefile`; `internal/tools/fuzzproof/` | 1 | not applicable (repository source) | reviewed | W05ReviewGateFinal | repository source lifetime; CI proof 14/30 days |
