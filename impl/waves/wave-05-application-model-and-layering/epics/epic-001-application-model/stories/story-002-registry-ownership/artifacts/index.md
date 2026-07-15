---
id: W05-E01-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W05-E01-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories are created on first real content, not pre-populated empty. All entries
below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W05-E01-S002-001 | resource.Registry owner-bound wrapper | source-code package | implementation | Ownership-structural wrapper over `kernel/resource` | AR-01 | W05-E01-S002-T001 | TBD at implementation time | not yet produced |
| ART-W05-E01-S002-002 | rules.Registry owner-bound wrapper | source-code package | implementation | Ownership-structural wrapper over `kernel/rules` | AR-01 | W05-E01-S002-T001 | TBD at implementation time | not yet produced |
| ART-W05-E01-S002-003 | authz.Registry owner-bound permission-registration wrapper | source-code package | implementation | Closes the framework's widest zero-ownership-check registration gap | AR-01 | W05-E01-S002-T002 | TBD at implementation time | not yet produced |
| ART-W05-E01-S002-004 | Owner-bound wrappers for remaining ~9+ declaration classes | source-code package | implementation | Events, jobs, workflow actions, providers, templates, health checks, migrations, seeds, OpenAPI | AR-01 | W05-E01-S002-T003 | TBD at implementation time | not yet produced |
| ART-W05-E01-S002-005 | Declaration-class enumeration/audit record | documentation | pre-implementation | Confirms the exact declaration-class list T6 wraps, resolving PLAN's own "~9+" estimate | AR-01 | W05-E01-S002-T003 | TBD at implementation time | not yet produced |
