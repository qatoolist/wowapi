---
id: W01-E02-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E02-S001
status: recorded
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E02-S001 — Evidence index

Per mandate §10. All three expected records are now produced; raw transcripts live in `tests/`,
`benchmarks/`, and `regression/` beside the records. Every record pins revision
`0a31186cada5c275a588c74081cf977adf346e61` (HEAD; conductor owns commits, so the tested tree is
HEAD plus this story's uncommitted change, whose file set is listed in `../implementation.md`).

| Evidence ID | Title | Type | Acceptance criteria proven | Status |
|---|---|---|---|---|
| [EV-W01-E02-S001-001](tests/EV-W01-E02-S001-001.md) | Fail-first correlation test transcript (fails before fix, passes after) | test | AC-W01-E02-S001-01 | recorded — pass (fail-first pair preserved) |
| [EV-W01-E02-S001-002](tests/EV-W01-E02-S001-002.md) | Negative-case "attrs genuinely absent" test transcript | test | AC-W01-E02-S001-02 | recorded — pass (key-absence assertion) |
| [EV-W01-E02-S001-003](benchmarks/EV-W01-E02-S001-003.md) | Allocation-neutrality benchmark (no-op tracer path, before/after) | benchmark | AC-W01-E02-S001-03 | recorded — pass (0 allocs/op both sides) |

## Notes

Supporting regression evidence: `regression/touched-packages-race.txt` — full `-race` pass over
kernel/tracing, kernel/observability, kernel/logging, kernel/database, adapters/tracing/otel,
kernel/outbox, kernel/jobs, kernel/notify. Reviewer fields are `pending` until the conductor's
independent-review gate (mandate §14) runs; workers do not self-accept.
