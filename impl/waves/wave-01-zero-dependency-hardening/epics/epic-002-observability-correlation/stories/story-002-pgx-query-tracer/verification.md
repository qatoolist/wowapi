---
id: VER-W01-E02-S002
type: verification-record
parent_story: W01-E02-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W01-E02-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E02-S002-01 | Trace-fixture integration test: run a pgx query inside a context carrying a parent span (test tracer/in-memory span recorder), export the trace tree, assert the query span's parent-span ID matches the request span's ID | Local / CI, real Postgres instance (per project preference for real integration tests over mocks across a process boundary) | Query span appears as a direct child; fails before wiring, passes after | integration-test report | framework architecture lead |
| AC-W01-E02-S002-02 | Statement-summary/rows-affected/error test: assert attrs present on a successful query; assert `RecordError` invoked on a deliberately failing query; literal-leakage test with a sensitive-looking bound parameter | Local / CI, real Postgres instance | Attrs present and correct; span marked errored on failure; no literal parameter value present in statement-summary attr | integration-test report | framework architecture lead |

## Post-execution record

### Actual result

| Acceptance criterion | Actual result | Evidence |
|---|---|---|
| AC-W01-E02-S002-01 | PASS — `pool.Exec(ctx, "SELECT 1")` inside a parent span yields an exported `db.SELECT` span whose `Parent.SpanID()` equals the parent's `SpanID()` and whose trace ID matches. Fail-first pair captured: before wiring, "no query span parented under … (1 spans exported)". Tree-shape corollaries also pass: parentless query → root span; unsampled parent → zero exported spans (ParentBased inheritance). | EV-W01-E02-S002-001 |
| AC-W01-E02-S002-02 | PASS — `db.statement == "SELECT 1"`, `db.rows_affected == "1"` on success; failed query (missing table) → span status Error + `exception` event via `RecordError`; bound parameter `hunter2-super-secret-literal` absent from every string attr of every exported span while `db.statement` retains the `$1` placeholder. | EV-W01-E02-S002-002, EV-W01-E02-S002-003 |

### Pass or fail

PASS — both ACs.

### Evidence identifier

EV-W01-E02-S002-001, -002, -003 (`evidence/index.md`); regression sweep
`evidence/regression/touched-packages-race.txt` (full `kernel/database` suite incl. RLS-guard and
pool-construction tests, `-race`, real Postgres — pass).

### Execution date

2026-07-13, ~07:29–07:35 UTC.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (HEAD) + this story's uncommitted working change.

### Environment

Local dev machine, macOS Darwin 25.5.0 arm64, go1.26.5; real Postgres 16-alpine (compose stack
`wowapi-postgres-1`), DSN `postgres://wowapi:…@localhost:5432/wowapi`.

### Reviewer

W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13. Reviewer must re-confirm no OTel SDK
type in `kernel/database`'s public surface or import graph (RISK-W01-E02-003) — implementation-time
check: `go.opentelemetry` appears only in the `_test.go` fixture, never in production files.

### Findings

Import-cycle discovery (observability→httpx→authz→database) — resolved at source via the
`kernel/tracing` leaf package (S001 DEV-W01-E02-S001-001 / this story's reference deviation).

### Retest status

Final `-race` run after the port relocation: all six tracer tests plus the full package suite pass.

### Final conclusion

Both acceptance criteria verified with registered, revision-pinned evidence against a real
Postgres. Story advanced to `verified`; acceptance is the conductor's call.

> Note (conductor, 2026-07-13): review gate found the init scaffold templates did not wire `database.WithQueryTracer`; conductor wired it into both templates and re-ran `internal/cli` tests (ok 28.9s) before acceptance — see deviations.md.
