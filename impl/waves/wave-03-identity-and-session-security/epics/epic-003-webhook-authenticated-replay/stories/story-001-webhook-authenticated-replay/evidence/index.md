---
id: W03-E03-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W03-E03-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E03-S001 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` are created on first real content, not pre-populated
empty. All entries below are `not yet produced`.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W03-E03-S001-001 | unit test report | W03-E03-S001-T001 | AC-W03-E03-S001-01 | `go test ./kernel/webhook -run 'TestUnit(HMACVerifier|FakeVerifier)' -v` | TBD | PASS | produced |
| EV-W03-E03-S001-002 | targeted test report | W03-E03-S001-T002 | AC-W03-E03-S001-02 | `go test ./kernel/webhook -run 'TestIntegrationHandleInbound_TimestampManipulationImmune' -v` | TBD | PASS | produced |
| EV-W03-E03-S001-003 | adversarial tamper-matrix test report | W03-E03-S001-T003 | AC-W03-E03-S001-03 | `go test ./kernel/webhook -run 'TestIntegrationHandleInbound_TamperMatrix' -v` | TBD | PASS | produced |
| EV-W03-E03-S001-004 | review report | W03-E03-S001-T005 | AC-W03-E03-S001-01, AC-W03-E03-S001-02, AC-W03-E03-S001-03, AC-W03-E03-S001-04 | Independent review checklist per mandate §14 | TBD | TBD | not yet produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.

Note: T004 (the provider-verifier contract document) produces no test-execution evidence of its own;
its verification is a documentation review, recorded directly in `../verification.md`'s
post-execution section rather than as a numbered `EV-` entry, consistent with its "document review
record" evidence type in that file's planned-verification table.
