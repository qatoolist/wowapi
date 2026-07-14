---
id: W06-E03-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W06-E03-S001
status: implemented
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E03-S001 — Artifacts index

Implementation artifacts are registered at their repository paths below.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W06-E03-S001-001 | Manifest schema + validator | schema + source | implementation | strict gate catalog validation | REL-01 | T001 | `ci/release-gates.schema.json`, `scripts/validation/release_contract.py` | implemented |
| ART-W06-E03-S001-002 | Required gate entries | configuration | implementation | complete Wave-6 required gate catalog | REL-01 | T002 | `ci/release-gates.yaml` | implemented |
| ART-W06-E03-S001-003 | Reusable required gates | CI workflow | implementation | exact-SHA results and GitHub attestation | REL-01 | T003 | `.github/workflows/required-gates.yml` | implemented |
| ART-W06-E03-S001-004 | CI caller | CI workflow | implementation | PR/main shared path | REL-01 | T004 | `.github/workflows/ci.yml` | implemented |
| ART-W06-E03-S001-005 | Release verify | CI workflow | implementation | exact tag/SHA and hosted security barrier | REL-01 | T005 | `.github/workflows/release.yml` | implemented |
| ART-W06-E03-S001-006 | Immutable candidate | CI workflow | implementation | GoReleaser `--skip=publish`, OCI, scans, manifest | REL-01 | T006 | `.github/workflows/release.yml` | implemented with accepted deviation |
| ART-W06-E03-S001-007 | Exact-byte publisher | CI workflow | implementation | draft `gh`/ORAS publisher rejects unmanifested input | REL-01 | T007 | `.github/workflows/release.yml` | implemented |
| ART-W06-E03-S001-008 | Clean release verifier | script | implementation | fail-closed local/clean verification | REL-01 | T008 | `scripts/validation/verify_release.sh`, `scripts/validation/release_contract.py` | implemented |
| ART-W06-E03-S001-009 | Verify-published + guarantees | workflow + documentation | post-implementation | clean runner plus existing supply-chain verification guidance | REL-01 | T008 | `.github/workflows/release.yml`, `SECURITY.md` | implemented |
