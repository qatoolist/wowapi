---
id: W02-E02-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W02-E02-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E02-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W02-E02-S001-001 | Parent tenant-scoped unique-index migrations | migration | implementation | 4 `CONCURRENTLY`-built `UNIQUE (tenant_id, id)` migrations on `parties`, `organizations`, `documents`, `document_versions` | DATA-01 | W02-E02-S001-T001 | TBD at implementation time | not yet produced |
| ART-W02-E02-S001-002 | Tenant-FK catalog scanner | source-code package | implementation | Tool enumerating every tenant-table FK and flagging any not composite on `(tenant_id, …)` | DATA-01 | W02-E02-S001-T002 | TBD at implementation time | not yet produced |
| ART-W02-E02-S001-003 | CI gate wiring | configuration | implementation | Extends CI to invoke the scanner as a permanent gate | DATA-01 | W02-E02-S001-T003 | TBD at implementation time | not yet produced |
| ART-W02-E02-S001-004 | Negative fixture migration | test fixture | implementation | A migration adding a single-column, non-composite tenant FK, used to prove the CI gate rejects it | DATA-01 | W02-E02-S001-T003 | TBD at implementation time | not yet produced |
| ART-W02-E02-S001-005 | Scanner and CI-gate documentation | documentation | post-implementation | Documents the scanner's purpose, matrix-keying mechanism, and CI gate failure behavior | DATA-01 | W02-E02-S001-T002, W02-E02-S001-T003 | TBD at implementation time | not yet produced |
