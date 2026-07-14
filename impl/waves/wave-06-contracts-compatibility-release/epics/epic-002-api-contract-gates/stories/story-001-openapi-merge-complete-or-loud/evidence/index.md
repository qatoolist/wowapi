---
id: W06-E02-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E02-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E02-S001 — Evidence index

Raw execution output is registered in `openapi-focused-tests.txt`; the dependency decision and
security/licence evidence is registered in `security/validator-dependency-review.md`.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W06-E02-S001-001 | fixture test report (per-field-type) | W06-E02-S001-T001 | AC-W06-E02-S001-01 | `go test ./internal/cli -run 'OpenAPI' -count=1` | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S001-002 | structural-validation test report | W06-E02-S001-T002 | AC-W06-E02-S001-02 | same focused OpenAPI run | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S001-003 | CI gate test report (seeded breaking-change fixture) | W06-E02-S001-T003 | AC-W06-E02-S001-03 | same focused OpenAPI run | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S001-004 | dependency security/licence decision | W06-E02-S001-T002 | AC-W06-E02-S001-04 | module metadata and advisory/licence review | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S001-005 | independent review report | W06-E02-S001-T004 | AC-W06-E02-S001-01 through AC-W06-E02-S001-04 | review-only | working tree based on `733ef3e` | PASS, no open issues | produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
