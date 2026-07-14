---
id: W02-E01-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W02-E01-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W02-E01-S001-001 | Migration manifest schema definition | schema | implementation | Defines the required manifest fields and their validation rules | DATA-09 | W02-E01-S001-T001 | TBD at implementation time | not yet produced |
| ART-W02-E01-S001-002 | Manifest-schema CI validator | source-code package | implementation | Tool that validates every migration's manifest entry in CI | DATA-09 | W02-E01-S001-T001 | TBD at implementation time | not yet produced |
| ART-W02-E01-S001-003 | Lock-timeout enforcement mechanism | source-code package | implementation | Wraps online-classified DDL execution with a 2-second lock-timeout budget and bounded abort-and-retry | DATA-09 | W02-E01-S001-T002 | TBD at implementation time | not yet produced |
| ART-W02-E01-S001-004 | Manifest-schema and lock-budget documentation | documentation | post-implementation | Documents the manifest schema, required fields, lock-timeout budget, and retry ceiling | DATA-09 | W02-E01-S001-T001, W02-E01-S001-T002 | TBD at implementation time | not yet produced |
