---
id: W04-E03-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W04-E03-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03-S001 — Evidence index

Per mandate §10.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W04-E03-S001-001 | documentation-diff record (migration `00016` header correction) | W04-E03-S001-T001 | AC-W04-E03-S001-01 | Not applicable (documentation diff) | HEAD | `evidence/migration_and_code.diff` | accepted |
| EV-W04-E03-S001-002 | concurrency-test report (`DATA-04/stopgap/`) | W04-E03-S001-T001 | AC-W04-E03-S001-02 | `cd kernel/bulk && DATABASE_URL=... go test -run TestIntegrationBulkConcurrentProcessorRejected -count=1 -v .` | HEAD | `evidence/stopgap_concurrency_test.log` | accepted |

Both acceptance criteria verified:
- AC-W04-E03-S001-01: migration `00016` header no longer claims "safe across replicas" via `FOR UPDATE SKIP LOCKED`.
- AC-W04-E03-S001-02: `TestIntegrationBulkConcurrentProcessorRejected` passes with two goroutines processing the same `bulkID`; the second is rejected with `KindConflict` while the first completes all items.
