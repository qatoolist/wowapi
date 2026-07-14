---
id: W07-E01-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W07-E01-S003
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S003 — Evidence index

Per mandate §10, complete machine-readable records live under `evidence/performance/`;
checksums pin each cited repository artifact.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W07-E01-S003-001 | cardinality test report (10/1k/100k due rows) | W07-E01-S003-T001 | AC-W07-E01-S003-01 | focused real-Postgres workflow package | working tree based on `733ef3e` | PASS: fixed 100-row ceiling and bounded loads | produced |
| EV-W07-E01-S003-002 | query-count + idempotency report | W07-E01-S003-T002 | AC-W07-E01-S003-02 | focused traced workflow contracts | working tree based on `733ef3e` | PASS: fixed statements, reinvocation, no double remind | produced |
| EV-W07-E01-S003-003 | real PostgreSQL EXPLAIN plan report | W07-E01-S003-T003 | AC-W07-E01-S003-03 | focused EXPLAIN integration test | working tree based on `733ef3e` | PASS: Index Scan using `wft_remind_after` | produced |
| EV-W07-E01-S003-004 | RetryOutbound query-count report | W07-E01-S003-T004 | AC-W07-E01-S003-04 | 10-delivery/3-endpoint traced test | working tree based on `733ef3e` | PASS: exactly one batch endpoint query | produced |
| EV-W07-E01-S003-005 | leased relay chaos report | W07-E01-S003-T005 | AC-W07-E01-S003-05 | outbox race + inherited W04 chaos | working tree based on `733ef3e` | PASS: commit boundary, fencing, ordering | produced |
| EV-W07-E01-S003-006 | metric-emission report | W07-E01-S003-T006 | AC-W07-E01-S003-06 | workflow/webhook/outbox metric tests | working tree based on `733ef3e` | PASS: lag gauge + duration histogram | produced |
| EV-W07-E01-S003-007 | bounded benchmark + budget report | W07-E01-S003-T007 | AC-W07-E01-S003-07 | focused benchmark→budget gate | working tree based on `733ef3e` | PASS: all three tiers within same-change budgets | produced |
| EV-W07-E01-S003-008 | published relative comparison | W07-E01-S003-T008 | AC-W07-E01-S003-07 | same-host before/after publication | working tree based on `733ef3e` | PASS relative; absolute pending DEC-Q9 | produced |

Fresh passing evidence uses index status `produced`; failed/superseded/retested/resolved and
accepted-exception remain available if a future rerun replaces any record.
