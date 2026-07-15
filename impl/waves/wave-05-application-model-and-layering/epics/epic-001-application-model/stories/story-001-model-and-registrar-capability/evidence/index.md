---
id: W05-E01-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W05-E01-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01-S001 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` are created on first real content, not pre-populated
empty. All entries below are `not yet produced`.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W05-E01-S001-001 | unit-test report (`AR-01/lifecycle_test_output.txt`) | W05-E01-S001-T001 | AC-W05-E01-S001-01 | `go test -v ./kernel/appmodel` | 733ef3e | PASS | produced |
| EV-W05-E01-S001-002 | unit-test report (build-tag matrix, D-03 error/panic split) | W05-E01-S001-T001 | AC-W05-E01-S001-02 | `go test -v -tags=dev ./kernel/appmodel` | 733ef3e | PASS | produced |
| EV-W05-E01-S001-003 | compile-fail fixture report (`AR-01/registrar_capability_test_output.txt`) | W05-E01-S001-T002 | AC-W05-E01-S001-03 | `go build ./kernel/port/testdata/...` | 733ef3e | FAIL (expected) | produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
