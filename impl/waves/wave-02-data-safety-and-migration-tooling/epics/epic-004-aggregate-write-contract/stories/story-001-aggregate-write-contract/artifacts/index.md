---
id: W02-E04-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W02-E04-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E04-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W02-E04-S001-001 | Typed aggregate repository/unit-of-work helper | source-code package | implementation | Bundles aggregate write + mirror upsert + audit + outbox atomically | DATA-06 | W02-E04-S001-T001 | TBD at implementation time (expected within `kernel/resource`) | not yet produced |
| ART-W02-E04-S001-002 | `registrar_pg.go` actor-attribution fix | source-code change | implementation | Sources `created_by` from context; rejects missing actor for user-initiated writes | DATA-06 | W02-E04-S001-T002 | `kernel/resource/registrar_pg.go` | not yet produced |
| ART-W02-E04-S001-003 | Migrated reference handler | source-code change | implementation | Reference handler calls the new helper instead of two independent statements | DATA-06 | W02-E04-S001-T003 | TBD at implementation time | not yet produced |
| ART-W02-E04-S001-004 | Updated `kernel/resource` documentation | documentation | post-implementation | Describes the implemented mandatory-mirror contract | DATA-06 | W02-E04-S001-T004 | TBD at implementation time (expected within `kernel/resource`'s package doc) | not yet produced |
