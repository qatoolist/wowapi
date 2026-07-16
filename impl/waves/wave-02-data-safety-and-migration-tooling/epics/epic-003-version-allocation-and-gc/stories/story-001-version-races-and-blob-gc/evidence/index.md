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
| EV-W02-E03-S001-006 | review report (independent review; this story has no dedicated independent-review task by this programme's own documented convention — see `tasks/index.md` "No dedicated independent-review task is added" rationale for P1 stories — so this review is recorded directly as evidence rather than via a task-NNN file) | (no independent-review task exists for this story) | AC-W02-E03-S001-01, AC-W02-E03-S001-02, AC-W02-E03-S001-03, AC-W02-E03-S001-04, AC-W02-E03-S001-05 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./... -run 'TestIntegrationInitiateUploadConcurrentVersionAllocation\|TestIntegrationUploadSessionDurability\|TestIntegrationConfirmUploadCAS\|TestIntegrationSweepUploadSessionsAdversarial\|TestIntegrationGenerateConcurrentVersionAllocation' -v -count=1 -tags=integration` | HEAD 43b6e12 + remediation working tree 2026-07-16 | pass (5/5 tests PASS; see review notes below re: `closure.md`'s phantom T006 reference) | produced |

Evidence files are located under `evidence/tests/`.

## Review notes (EV-006, independent review, 2026-07-16)

Reviewer: Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor
(autopsy remediation R-3). Environment: macOS (darwin/arm64), go1.26.5. Commit: HEAD 43b6e12 +
remediation working tree 2026-07-16 (kernel/artifact and foundation/document unmodified by the
uncommitted remediation diff).

Re-ran all 5 named acceptance tests: `TestIntegrationInitiateUploadConcurrentVersionAllocation`,
`TestIntegrationUploadSessionDurability`, `TestIntegrationConfirmUploadCAS`,
`TestIntegrationSweepUploadSessionsAdversarial`, `TestIntegrationGenerateConcurrentVersionAllocation`
— all 5 PASS. AC-01 through AC-05 CONFIRMED.

Two documentation findings, neither functional:
1. `closure.md`'s "Task completion" section states "W02-E03-S001-T006: complete (review gate
   W02ReviewGate)" and its "Reviewer conclusion" section states "Independent review passed
   (W02ReviewGate, 2026-07-13)" — but this story's own `tasks/index.md` explicitly states, by this
   programme's documented convention for P1 stories, "**No dedicated independent-review task is
   added**" and only 5 tasks (T001–T005) exist. `closure.md` references a T006 that was never
   created and a review event with no discoverable dated/attributed artifact predating this
   review — the same "claimed review, no corroborating record" pattern found across the other W02
   stories, except here it additionally cites a task ID that doesn't exist. This review (EV-006,
   dated and attributed) is the first genuinely evidenced review this story has received.
2. This evidence index's EV-001..005 execution commands reference `./kernel/document/...` and
   `./kernel/artifact/...`; the tests were found and re-run at `./foundation/document/...` and
   `./kernel/artifact/...` respectively during this review (package `document` now lives under
   `foundation/`, not `kernel/`) — a stale path in the recorded command, not a missing test. The
   named tests themselves exist and pass at their current location.

Recommendation: accept-with-conditions. All 5 ACs functionally confirmed. Condition: correct
`closure.md`'s phantom T006/W02ReviewGate reference to reflect that no dedicated review task was
ever created for this story (per its own documented convention) and that this EV-006 record is the
review of record; optionally correct the stale `kernel/document` path in EV-001..005.
