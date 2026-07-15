---
id: W02-EPICS-INDEX
type: epics-index
wave: W02
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02 — Epics index

| Epic | Title | Status | Stories | Objective |
|---|---|---|---|---|
| [W02-E01](epic-001-online-migration-protocol/epic.md) | online-migration-protocol | planned | 3 | Build the DATA-09 online expand/backfill/validate/contract migration protocol from zero: manifest schema, lock-timeout enforcement, backfill harness, validation/canary/switch/contract tooling, and a full CI drill pipeline |
| [W02-E02](epic-002-tenant-fk-integrity/epic.md) | tenant-fk-integrity | planned | 2 | Close the confirmed tenant-FK integrity gap (DATA-01): composite `(tenant_id, id)` foreign keys on all 8 affected tenant-scoped child tables, built on E01's protocol for the riskiest steps |
| [W02-E03](epic-003-version-allocation-and-gc/epic.md) | version-allocation-and-gc | planned | 1 | Replace racy `MAX(version)+1` allocation with a locked-counter/sequence approach for `kernel/artifact` and `kernel/document`, and garbage-collect orphaned upload blobs (DATA-05) |
| [W02-E04](epic-004-aggregate-write-contract/epic.md) | aggregate-write-contract | planned | 1 | Make the resource-mirror write mandatory and framework-enforced via a typed aggregate repository/unit-of-work helper, with real actor attribution (DATA-06) |
| [W02-E05](epic-005-production-seed-sync/epic.md) | production-seed-sync | planned | 1 | Build the production catalog seed-sync path that closes the empty-catalog-DB deny-everything gap (FBL-02), per MATRIX CS-21's fixed acceptance bar |
