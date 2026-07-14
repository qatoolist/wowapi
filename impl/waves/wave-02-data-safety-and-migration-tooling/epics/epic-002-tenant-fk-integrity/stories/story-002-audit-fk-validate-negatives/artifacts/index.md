---
id: W02-E02-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W02-E02-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E02-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories are created on first real content, not pre-populated empty. All entries
below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W02-E02-S002-001 | Mismatch-audit tool and report | tooling + report | pre-implementation | Platform-role-connected scan across all 8 edges; dated, inspectable zero-mismatch report or resolved remediation-decision record | DATA-01 | W02-E02-S002-T001 | TBD (adjacent to W02-E02-S001's catalog scanner) at implementation time | not yet produced |
| ART-W02-E02-S002-002 | Composite FK `NOT VALID` migrations (8 edges) | migration | implementation | Per-table `NOT VALID` composite FK add for each of the 8 tenant-scoped edges | DATA-01 | W02-E02-S002-T002 | TBD (migrations directory) at implementation time | not yet produced |
| ART-W02-E02-S002-003 | `VALIDATE CONSTRAINT` migrations/statements (8 edges) | migration | implementation | Per-table validation of each new composite FK | DATA-01 | W02-E02-S002-T003 | TBD (migrations directory) at implementation time | not yet produced |
| ART-W02-E02-S002-004 | Extended cross-tenant negative-test suite | test suite | implementation | Catalog-driven RLS matrix test extended with seeded cross-tenant insert cases under both `app_rt` and `app_platform` | DATA-01 | W02-E02-S002-T004 | TBD (existing RLS matrix test file(s)) at implementation time | not yet produced |
| ART-W02-E02-S002-005 | FK-removal migrations + consumer/rollback verification record (optional) | migration + report | post-implementation | Removal of the 8 redundant single-column FKs, only if pursued, with its grep/regression verification record | DATA-01 | W02-E02-S002-T005 | TBD at implementation time, if pursued | not yet produced |
