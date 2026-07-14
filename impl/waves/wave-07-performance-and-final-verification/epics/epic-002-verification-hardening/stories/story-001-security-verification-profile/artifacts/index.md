---
id: W07-E02-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W07-E02-S001
status: blocked
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E02-S001 — Artifacts index

Per mandate §9.2. The external report remains explicitly blocked; its commissioning status record is
registered separately and must never be substituted for the report.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Repository path / storage location | Version | Checksum | Status | Reviewer | Retention |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-W07-E02-S001-001 | Version-pinned control map bundle | documentation + data | implementation | Human map, canonical 412-entry JSON map, and pinned ASVS/OWASP/NIST source inventories | SEC-05 | W07-E02-S001-T001 | `SEC-05/control-map.md`; `SEC-05/control-map.json`; `SEC-05/sources/` | SEC-05 schema v1; ASVS 5.0.0; OWASP API 2023; NIST SP 800-63-4 final (2025-07) | JSON `890cacbcc71ce0c1b8fbbe8a24c0f618badea5b570559920f10a1479aa44bf1a`; Markdown `002923b2129b683592c67409d64db293d9d85437cbe47d0c03572c8ee92d2a44` | produced; focused validation passed; integration-commit retest pending | W05ReviewGateFinal — reviewed, no actionable issue | Permanent programme security record |
| ART-W07-E02-S001-002 | External assessment report | professional-services report | post-implementation | Genuine independent external assessment and findings disposition | SEC-05 | W07-E02-S001-T002 | Not available — no assessor/vendor, engagement, report, or report URI supplied | Not available | Not available | **blocked / not produced** | external assessor + product-security lead required | Permanent once produced |
| ART-W07-E02-S001-003 | Machine-check validator and regression tests | verification tooling | implementation | Enforces inventory digests/completeness, source-title integrity, version consistency, executable-test resolution, N/A rationale, and genuine waiver fields | SEC-05 | W07-E02-S001-T001 | `SEC-05/validate_control_map.py`; `SEC-05/test_validate_control_map.py`; `SEC-05/verify_prerequisites.py` | schema v1 | validator `7bf72732c1f66f71ade63dcd384941263337c3a3c4fde005b3c367e9a81a54de`; tests `3e92c9dc4a69cb5b544b98106056589260267cc3f11e5ebb8146061695c3eb1c`; prerequisite checker `7ebf68aab6a8ba422a6a06ae1279dde268163dc3ef77b255d0134ce68af2793f` | produced; tests passed | W05ReviewGateFinal — reviewed, no actionable issue | Retain with control map |
| ART-W07-E02-S001-004 | External-assessment commissioning status | blocker record | post-implementation | Exact human/vendor inputs absent and actions required to unblock AC-02 | SEC-05 | W07-E02-S001-T002 | `SEC-05/external-assessment-status.md` | 2026-07-14 status | `4d50555fa4d92300d5dedfac447b12b337ff3cff17b60f84e500a4f3c21b6e46` | produced; blocker open; not an assessment report | W05ReviewGateFinal — truthfulness reviewed | Retain until superseded, then preserve as history |
