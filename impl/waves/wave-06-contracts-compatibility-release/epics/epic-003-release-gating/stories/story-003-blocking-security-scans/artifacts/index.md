---
id: W06-E03-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W06-E03-S003
status: implemented
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E03-S003 — Artifacts index

Implementation artifacts are registered at their repository paths below.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W06-E03-S003-001 | Blocking scoped Trivy | CI configuration | implementation | report-only baseline followed by exit-code 1 and scoped ignore | REL-02 | T001 | `.github/workflows/security-scan.yml`, `.trivyignore.yaml` | implemented |
| ART-W06-E03-S003-002 | Expiring waiver contract | schema + source | implementation | owner/rationale/expiry/remediation link and ignore synchronization | REL-02 | T002 | `ci/security-waivers.schema.json`, `ci/security-waivers.yaml`, `scripts/validation/security_contract.py` | implemented |
| ART-W06-E03-S003-003 | Visibility meta-check | CI workflow | implementation | public exact-SHA hosted proof; forced visibility tests | REL-02 | T003 | `.github/workflows/security-scan.yml` | implemented |
| ART-W06-E03-S003-004 | Private fallback | CI workflow | implementation | local actionlint/SAST/govulncheck/Trivy/posture without false parity | REL-02 | T004 | `.github/workflows/security-scan.yml`, `scripts/validation/security_contract.py` | implemented |
| ART-W06-E03-S003-005 | Gate/release wiring | CI configuration | post-implementation | one gate per scanner class plus required artifact/image security reports | REL-02 | T005 | `ci/release-gates.yaml`, `.github/workflows/required-gates.yml`, `.github/workflows/release.yml` | implemented |
