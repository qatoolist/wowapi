---
id: W02-E02-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W02-E02-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E02-S002 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` are created on first real content, not pre-populated
empty. All entries below are `not yet produced`.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W02-E02-S002-001 | audit report | W02-E02-S002-T001 | AC-W02-E02-S002-01 | TBD (mismatch-audit tool run against staging/prod-shaped data) at implementation time | TBD | TBD | not yet produced |
| EV-W02-E02-S002-002 | integration-test report | W02-E02-S002-T001 | AC-W02-E02-S002-01 | TBD (seeded-mismatch integration test) at implementation time | TBD | TBD | not yet produced |
| EV-W02-E02-S002-003 | migration lock-duration report | W02-E02-S002-T002 | AC-W02-E02-S002-02 | TBD (per-table `NOT VALID` add, lock-duration measurement) at implementation time | TBD | TBD | not yet produced |
| EV-W02-E02-S002-004 | load-test report | W02-E02-S002-T003 | AC-W02-E02-S002-03 | TBD (concurrent-writer-load test during `VALIDATE CONSTRAINT`) at implementation time | TBD | TBD | not yet produced |
| EV-W02-E02-S002-005 | audit report (second confirmation) | W02-E02-S002-T003 | AC-W02-E02-S002-03 | TBD (second zero-mismatch confirmation) at implementation time | TBD | TBD | not yet produced |
| EV-W02-E02-S002-006 | RLS matrix test report | W02-E02-S002-T004 | AC-W02-E02-S002-04 | TBD (seeded cross-tenant insert under both `app_rt` and `app_platform`) at implementation time | TBD | TBD | not yet produced |
| EV-W02-E02-S002-007 | regression + grep sweep report | W02-E02-S002-T005 | AC-W02-E02-S002-05 | TBD (full regression + grep sweep for old FK name), if pursued, at implementation time | TBD | TBD | not yet produced |
| EV-W02-E02-S002-008 | review report | W02-E02-S002-T006 | AC-W02-E02-S002-01, AC-W02-E02-S002-02, AC-W02-E02-S002-03, AC-W02-E02-S002-04, AC-W02-E02-S002-05 | Independent review checklist per mandate §14 | TBD | TBD | not yet produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
