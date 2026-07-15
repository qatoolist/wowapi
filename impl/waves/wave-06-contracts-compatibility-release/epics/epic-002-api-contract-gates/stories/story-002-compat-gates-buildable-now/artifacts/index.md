---
id: W06-E02-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W06-E02-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W06-E02-S002 — Artifacts index

All six REL-03a artifacts are produced. The architecture smoke runs inside the no-publish candidate
job, and the supply-chain property reuses REL-01's clean published-release verifier.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W06-E02-S002-001 | Go API diff CI job | script + CI | implementation | Pinned `apidiff` classifies oldest-supported vs current exported API | REL-03 | W06-E02-S002-T001 | `scripts/check_go_api_compat.sh`; `.github/workflows/compatibility-gates.yml` | produced |
| ART-W06-E02-S002-002 | Module compile matrix | CI configuration | implementation | Supported Go patch versions with locked dependencies and explicit exclusions | REL-03 | W06-E02-S002-T002 | `.github/workflows/compatibility-gates.yml` | produced and executed |
| ART-W06-E02-S002-003 | Config schema compatibility gate | CLI + fixtures + CI | implementation | Recursive breaking-config detection and additive optional acceptance | REL-03 | W06-E02-S002-T003 | `internal/compat/config_schema.go`; `cmd/compatcheck`; `.github/workflows/compatibility-gates.yml` | produced |
| ART-W06-E02-S002-004 | Oldest-supported migration drill | integration test | implementation | Rebuilds v1.0.0 head, seeds data, upgrades, reverses, and reconstructs | REL-03 | W06-E02-S002-T004 | `migrations/reversible_test.go`; `kernel/database/migrate.go` | produced |
| ART-W06-E02-S002-005 | Container architecture smoke | script + CI | implementation | Exact-OCI-layout amd64/arm64 CLI boot smoke before publish | REL-03 | W06-E02-S002-T005 | `scripts/smoke_candidate_arch.sh`; `scripts/smoke_candidate_oci.sh`; `.github/workflows/release.yml` | produced and executed |
| ART-W06-E02-S002-006 | REL-03 T9 cross-reference | shared release verification | post-implementation | Reuses REL-01 T8/T9 cryptographic signature, SBOM, provenance, hash, platform, and version verification | REL-03 | W06-E02-S002-T006 | `scripts/validation/verify_release.sh`; `scripts/validation/release_contract.py`; `.github/workflows/release.yml` | produced and executed |
