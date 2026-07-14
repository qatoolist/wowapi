---
id: W06-E02-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E02-S003
status: blocked
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E02-S003 — Evidence index

The exact dependency-state inspection and per-leg blockers are registered in
`unblocking-status.txt`; no blocked gate result is claimed.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W06-E02-S003-001 | entry-criterion status record (OpenAPI gate) | W06-E02-S003-T001 | AC-W06-E02-S003-01 | dependency status inspection | working tree based on `733ef3e` | BLOCKED on S001 acceptance | produced blocker record |
| EV-W06-E02-S003-002 | entry-criterion status record (event/schema gate) | W06-E02-S003-T002 | AC-W06-E02-S003-02 | dependency status inspection | working tree based on `733ef3e` | BLOCKED on W06-E01-S001 + W05-E03 acceptance | produced blocker record |
| EV-W06-E02-S003-003 | entry-criterion status record (consumer upgrade) | W06-E02-S003-T003 | AC-W06-E02-S003-03 | dependency status inspection | working tree based on `733ef3e` | BLOCKED on W06-E01-S002 acceptance | produced blocker record |
| EV-W06-E02-S003-004 | independent blocked-leg review | W06-E02-S003-T004 | AC-W06-E02-S003-01 through AC-W06-E02-S003-03 entry conditions | review-only | working tree based on `733ef3e` | PASS, blockers honest | produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
