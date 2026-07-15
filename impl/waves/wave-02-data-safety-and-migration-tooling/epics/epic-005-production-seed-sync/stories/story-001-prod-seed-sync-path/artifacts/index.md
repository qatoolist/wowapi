---
id: W02-E05-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W02-E05-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E05-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W02-E05-S001-001 | Design-investigation decision record | design document | pre-implementation | Documents the resolved catalog manifest format, versioning scheme, CLI shape, idempotency mechanism, RLS/role posture, dry-run format, and audit integration, each with rationale | FBL-02 | W02-E05-S001-T001 | TBD at implementation time | not yet produced |
| ART-W02-E05-S001-002 | Catalog manifest schema definition | schema | implementation | Defines the versioned catalog manifest's required fields and validation rules, per T001's resolved design | FBL-02 | W02-E05-S001-T002 | TBD at implementation time | not yet produced |
| ART-W02-E05-S001-003 | Seed-sync command/path | source-code package | implementation | The `wowapi seed sync --env prod`-shaped command implementing idempotent, RLS-respecting catalog sync | FBL-02 | W02-E05-S001-T002 | TBD at implementation time | not yet produced |
| ART-W02-E05-S001-004 | Dry-run and audit-record mechanism | source-code package | implementation | Dry-run reporting and per-run audit-record production | FBL-02 | W02-E05-S001-T003 | TBD at implementation time | not yet produced |
| ART-W02-E05-S001-005 | Readiness-check registration | source-code package | implementation | The named readiness check reporting whether seed-sync has run against the current manifest version | FBL-02 | W02-E05-S001-T004 | TBD at implementation time | not yet produced |
| ART-W02-E05-S001-006 | Readiness-payload seed/catalog-hash reporting | source-code package | implementation | Readiness payload field reporting the seed/catalog hash once seed-sync has run | FBL-02 | W02-E05-S001-T005 | TBD at implementation time | not yet produced |
| ART-W02-E05-S001-007 | Seed-sync and readiness documentation | documentation | post-implementation | Documents the manifest format, CLI shape, readiness check, and audit-record shape | FBL-02 | W02-E05-S001-T005 | TBD at implementation time | not yet produced |
