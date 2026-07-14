---
id: W05-E05-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W05-E05-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E05-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2". All
entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W05-E05-S001-001 | foundation/ tree (8 mechanical package moves) | source-code package | implementation | webhook, notify, document, artifact, attachment, comment, bulk, integration moved from kernel/ | FBL-01 | W05-E05-S001-T001 | TBD at implementation time | not yet produced |
| ART-W05-E05-S001-002 | foundation/mfa move + kernel/mfa forwarding shim | source-code package | implementation | mfa moved; deprecated shim retained at kernel/mfa for one minor version | FBL-01 | W05-E05-S001-T002 | TBD at implementation time | not yet produced |
| ART-W05-E05-S001-003 | Extended depguard configuration | configuration | implementation | Denies kernel→foundation and foundation→app imports | FBL-01 | W05-E05-S001-T003 | TBD at implementation time | not yet produced |
| ART-W05-E05-S001-004 | Extended lint_boundaries.sh allowlist | configuration | implementation | Fails CI on new un-allowlisted kernel package | FBL-01 | W05-E05-S001-T004 | TBD at implementation time | not yet produced |
