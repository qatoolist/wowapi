---
id: W01-E02-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E02-S002
status: recorded
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E02-S002 — Evidence index

Per mandate §10. All three expected records are now produced; raw transcripts live in `tests/` and
`regression/`. Every record pins revision `0a31186cada5c275a588c74081cf977adf346e61` (HEAD;
conductor owns commits, so the tested tree is HEAD plus this story's uncommitted change, whose file
set is listed in `../implementation.md`). All integration runs executed against a real Postgres
(postgres:16-alpine, compose stack) with the otel adapter's in-memory span exporter as the trace
fixture.

| Evidence ID | Title | Type | Acceptance criteria proven | Status |
|---|---|---|---|---|
| [EV-W01-E02-S002-001](tests/EV-W01-E02-S002-001.md) | Trace-tree export test output (pgx span as child of parent span, fail-first before/after) | test | AC-W01-E02-S002-01 | recorded — pass (fail-first pair preserved) |
| [EV-W01-E02-S002-002](tests/EV-W01-E02-S002-002.md) | Statement-summary/rows-affected/error-marking test output | test | AC-W01-E02-S002-02 | recorded — pass |
| [EV-W01-E02-S002-003](tests/EV-W01-E02-S002-003.md) | Literal-parameter-leakage negative test output | test | AC-W01-E02-S002-02 | recorded — pass |

## Notes

Supporting regression evidence: `regression/touched-packages-race.txt` (shared sweep with S001).
Reviewer fields are `pending` until the conductor's independent-review gate (mandate §14) runs;
workers do not self-accept. Independent review should specifically re-confirm no OTel type in
`kernel/database`'s import graph (RISK-W01-E02-003) — checked at implementation time:
`grep -rn 'go.opentelemetry' kernel/database/*.go kernel/tracing/*.go` (excluding _test files)
returns nothing.
