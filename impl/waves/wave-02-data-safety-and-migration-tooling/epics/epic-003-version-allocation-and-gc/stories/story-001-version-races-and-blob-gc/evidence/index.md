---
id: W02-E03-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W02-E03-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W02-E03-S001 — Evidence index

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W02-E03-S001-001 | concurrency-test report (24 concurrent callers, lock wait measured) | W02-E03-S001-T001 | AC-W02-E03-S001-01 | `go test ./kernel/document/... -run TestIntegrationInitiateUploadConcurrentVersionAllocation -count=1 -v` | working tree on 1626b11 | PASS | accepted |
| EV-W02-E03-S001-002 | integration-test report (crash-simulation, upload-session durability) | W02-E03-S001-T002 | AC-W02-E03-S001-02 | `go test ./kernel/document/... -run TestIntegrationUploadSessionDurability -count=1 -v` | working tree on 1626b11 | PASS | accepted |
| EV-W02-E03-S001-003 | concurrency-test report (racing confirmations) | W02-E03-S001-T003 | AC-W02-E03-S001-03 | `go test ./kernel/document/... -run TestIntegrationConfirmUploadCAS -count=1 -v` | working tree on 1626b11 | PASS | accepted |
| EV-W02-E03-S001-004 | integration-test report (mixed confirmed/expired/pending GC sweep) | W02-E03-S001-T004 | AC-W02-E03-S001-04 | `go test ./kernel/document/... -run TestIntegrationSweepUploadSessionsAdversarial -count=1 -v` | working tree on 1626b11 | PASS | accepted |
| EV-W02-E03-S001-005 | concurrency-test report (dedicated `kernel/artifact.Generate` mirror test) | W02-E03-S001-T005 | AC-W02-E03-S001-05 | `go test ./kernel/artifact/... -run TestIntegrationGenerateConcurrentVersionAllocation -count=1 -v` | working tree on 1626b11 | PASS | accepted |

Evidence files are located under `evidence/tests/`.
