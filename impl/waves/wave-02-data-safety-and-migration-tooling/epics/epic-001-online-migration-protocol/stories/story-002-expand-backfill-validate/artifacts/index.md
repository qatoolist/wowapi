---
id: W02-E01-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W02-E01-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W02-E01-S002-001 | Expand-phase tooling | source-code package | implementation | Nullable/default-safe columns, tables/indexes/compatibility views, `NOT VALID` constraints, non-transactional `CREATE INDEX CONCURRENTLY` | DATA-09 | W02-E01-S002-T001 | TBD at implementation time | not yet produced |
| ART-W02-E01-S002-002 | Backfill-job harness + interim checkpoint-lease mechanism | source-code package | implementation | Resumable, tenant-scoped, keyset-paginated, checkpointed harness with bounded batch/tx time and rate controls; interim lease scope-bounded to checkpoint/resumability pending W04-E01-S001 | DATA-09 | W02-E01-S002-T002 | TBD at implementation time | not yet produced |
| ART-W02-E01-S002-003 | Validation-phase tooling + artifact schema | source-code package / schema | implementation | `VALIDATE CONSTRAINT` orchestration, reconciliation queries, machine-checked report artifact schema | DATA-09 | W02-E01-S002-T003 | TBD at implementation time | not yet produced |
| ART-W02-E01-S002-004 | Expand/backfill/validate documentation (incl. interim-lease scope-boundary note) | documentation | post-implementation | Documents all three tooling layers; explicitly flags the interim lease as a bounded, temporary substitute with a W04-E01-S001 forward reference | DATA-09 | W02-E01-S002-T001, W02-E01-S002-T002, W02-E01-S002-T003 | TBD at implementation time | not yet produced |
