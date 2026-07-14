---
id: W04-E03-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E03-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03-S001 — Artifacts index

Per mandate §9.2.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|---|
| ART-W04-E03-S001-001 | Corrected migration `00016` header comment | documentation (in-repo) | implementation | Removes the false "safe across replicas" claim; states the actual single-processor-enforced property | DATA-04 | W04-E03-S001-T001 | `migrations/00016_bulk_operations.sql` | accepted |
| ART-W04-E03-S001-002 | Single-processor enforcement mechanism (CAS) | source-code package | implementation | `bulk_operations.processor_id` / `processor_started_at` CAS guard; rejects a second concurrent processor at the `Service` API boundary with `ErrConcurrentProcessor` | DATA-04 | W04-E03-S001-T001 | `kernel/bulk/bulk.go` (Process, acquireProcessor, releaseProcessor) | accepted |
| ART-W04-E03-S001-003 | Stopgap mechanism documentation | documentation | post-implementation | This index + code comments document the corrected claim and the CAS enforcement mechanism | DATA-04 | W04-E03-S001-T001 | `impl/waves/.../story-001-stopgap/artifacts/index.md` | accepted |

Mechanism chosen: **CAS against `bulk_operations` columns** rather than PostgreSQL advisory locks, because `Process` spans multiple per-item transactions and a session-scoped advisory lock cannot be held across `TxManager.WithTenant` calls without a dedicated connection. The CAS guard uses a 5-minute timeout so a crashed processor does not permanently block the operation.
