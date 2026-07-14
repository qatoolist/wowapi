---
id: W06-E02-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E02-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W06-E02-S002 — Evidence index

Focused gate, compile-matrix, migration, real OCI architecture, shared release-verifier, and
independent-review output is registered below and in the two evidence files.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W06-E02-S002-001 | CI gate fixture report (Go API additive/breaking) | W06-E02-S002-T001 | AC-W06-E02-S002-01 | focused `go test ./internal/compat` | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S002-002 | compile-matrix run output | W06-E02-S002-T002 | AC-W06-E02-S002-02 | exact Go 1.26.0 and 1.26.5 compile-only runs | working tree based on `733ef3e` | PASS: 68 packages each | produced |
| EV-W06-E02-S002-003 | CI gate fixture report (config additive/breaking) | W06-E02-S002-T003 | AC-W06-E02-S002-03 | focused `go test ./internal/compat ./internal/compatcli` | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S002-004 | oldest-supported migration integration report | W06-E02-S002-T004 | AC-W06-E02-S002-04 | supplied-env focused migration test | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S002-005 | exact candidate architecture smoke | W06-E02-S002-T005 | AC-W06-E02-S002-05 | real OCI archive copied without rebuild; digest-run amd64 and arm64 | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S002-006 | shared REL-01 T8/T9 verifier evidence | W06-E02-S002-T006 | AC-W06-E02-S002-06 | 12-test golden suite, actionlint, production workflow review | working tree based on `733ef3e` | PASS | produced |
| EV-W06-E02-S002-007 | independent review report | W06-E02-S002-T007 | AC-01 through AC-06 | final independent rerun | working tree based on `733ef3e` | PASS, no open issues | produced |

