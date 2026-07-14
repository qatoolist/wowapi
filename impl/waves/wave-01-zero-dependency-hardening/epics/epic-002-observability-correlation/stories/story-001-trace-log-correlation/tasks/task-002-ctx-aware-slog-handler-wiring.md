---
id: W01-E02-S001-T002
type: task
title: Ctx-aware slog.Handler wrapper + AccessLog wiring
status: done
parent_story: W01-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W01-E02-S001-T001
acceptance_criteria:
  - AC-W01-E02-S001-01
  - AC-W01-E02-S001-02
  - AC-W01-E02-S001-03
artifacts: []
evidence: []
---

# W01-E02-S001-T002 — Ctx-aware slog.Handler wrapper + AccessLog wiring

## Task Definition

*Per mandate §8.6. This section defines the task before work begins.*

### Task objective

Implement a context-aware `slog.Handler` wrapper that injects `trace_id`/`span_id` attrs into a log
record when a recording span is active in the record's context (using the `TraceID()`/`SpanID()`
accessors T001 adds), wire it into the runtime logging path (`AccessLog` and/or logger
construction), and prove both the positive case (attrs present) and the negative case (attrs
genuinely absent, not empty-string) with tests, plus an allocation-neutrality benchmark for the
no-op tracer path.

### Parent story

W01-E02-S001 — Trace/log correlation.

### Owner

Unassigned.

### Status

`done` (2026-07-13).

### Dependencies

W01-E02-S001-T001 (the `Span` port must already expose `TraceID()`/`SpanID()` before this task can
read them).

### Detailed work

1. Resolve the "how is the active span retrieved from `ctx`" open question from `plan.md`
   (candidate: a `SpanFromContext`-style helper on the `Tracer` port or as a package function in
   `kernel/observability`; alternative: constructing the wrapper with a reference to the specific
   `Tracer`). Record the chosen approach in this task's Implementation Record once decided.
2. Implement the handler wrapper type in `kernel/observability` (new file). `Handle(ctx, record)`
   checks for an active recording span; if found and its `TraceID()`/`SpanID()` are non-empty, clones
   the record and appends `trace_id`/`span_id` attrs before delegating; otherwise delegates
   unchanged. Implement the remaining `slog.Handler` interface methods (`Enabled`, `WithAttrs`,
   `WithGroup`) as thin delegations to the wrapped handler, preserving its behavior.
3. Write the fail-first positive-case test first: construct a context carrying a real (non-no-op)
   recording span, emit a record through the (not-yet-wired) plain handler, confirm the test fails
   (no `trace_id`/`span_id` present) — this documents the pre-fix failure per mandate §13's fail-first
   discipline.
4. Wire the wrapper into the logging path — either `logging.New` (wrapping its constructed handler)
   or `AccessLog` (attaching the attrs directly, mirroring `request_id`), per the decision made
   during T002 (see `plan.md` "Proposed architecture" — decision deferred to implementation).
5. Re-run the positive-case test; confirm it now passes.
6. Write the negative-case test: emit a record via `context.Background()` (no span) and separately
   via a `NoOpTracer`-produced span; assert the record's attribute set does not contain `trace_id`/
   `span_id` keys at all (iterate `record.Attrs` and assert absence of the keys, not merely check
   `attr.Value.String() == ""`).
7. Write the allocation-neutrality benchmark: `go test -bench` with `ReportAllocs()` comparing the
   wrapped handler (no-op tracer path) against the pre-existing plain handler; confirm no
   statistically meaningful allocation increase.
8. Confirm existing `kernel/logging` and `kernel/observability` test suites still pass (redaction,
   level/format parsing, `Trace`/`AccessLog` existing tests) — regression check per `plan.md`.

### Expected files or components affected

- New file under `kernel/observability/` (handler wrapper).
- `kernel/logging/logging.go` and/or `kernel/observability/middleware.go` (wiring, per the
  implementation-time decision).

### Expected output

A wired, tested, benchmarked ctx-aware logging correlation mechanism: `trace_id`/`span_id` present
on log records inside an active span, genuinely absent otherwise, no-op path allocation-neutral.

### Required artifacts

ART-W01-E02-S001-002 (handler wrapper source), ART-W01-E02-S001-003 (wiring diff),
ART-W01-E02-S001-004 (updated doc comments) — see `../../artifacts/index.md`.

### Required evidence

EV-W01-E02-S001-001 (fail-first test transcript), EV-W01-E02-S001-002 (negative-case test
transcript), EV-W01-E02-S001-003 (allocation benchmark) — see `../../evidence/index.md`.

### Related acceptance criteria

AC-W01-E02-S001-01, AC-W01-E02-S001-02, AC-W01-E02-S001-03 (all three are proven at this task).

### Completion criteria

All three acceptance criteria pass with registered evidence; existing test suites show no
regression; independent review confirms the negative-case test checks key-absence, not
empty-string (RISK-W01-E02-002).

### Verification method

`go test ./kernel/observability/... ./kernel/logging/...` (positive case, negative case, existing
regression suite); `go test -bench=. -benchmem ./kernel/observability/...` (allocation benchmark).

### Risks

RISK-W01-E02-002 (negative-case test implemented as empty-string check instead of key-absence
check) is the primary risk for this task specifically — mitigated by explicit test-design guidance
in this task's "Detailed work" step 6 and checked at independent review.

### Rollback or recovery considerations

Revert the commit(s). No persistent state is created — log records are ephemeral, so rollback has no
residual-state concern. If the wiring point chosen in step 4 causes an unexpected regression in
production log volume/format, the wrapper can be removed from the construction path without
affecting the underlying `Span` port change (T001 remains valid and useful for S002 regardless).

## Implementation Record

Executed 2026-07-13 by W01Obs.

### What was actually implemented

- **Span retrieval (step 1 decision):** package-level `ContextWithSpan(ctx, Span)` /
  `SpanFromContext(ctx) (Span, bool)` helpers (canonical in `kernel/tracing`, forwarded by
  `kernel/observability`). Real adapters store the port span in the context `StartSpan` returns —
  `adapters/tracing/otel.StartSpan` updated; `NoOpTracer` stores nothing (zero-alloc disabled path).
  Globally reusable across tracer instances (rejected the tracer-reference alternative).
- **Wrapper (step 2):** `observability.NewCorrelatingHandler(h slog.Handler) slog.Handler` in
  `kernel/observability/correlation.go`. Injects `trace_id`/`span_id` on a CLONED record only when
  a span with non-empty `TraceID()` is in ctx; pure delegation otherwise;
  `Enabled`/`WithAttrs`/`WithGroup` delegate and re-wrap.
- **Wiring (step 4 decision): `logging.New`** (approach (a)) — every process logger correlates.
  `AccessLog` unchanged; it inherits correlation via `InfoContext(r.Context())` inside `Trace(tr)`
  (chain order Trace→AccessLog per the scaffold's documented ordering), proven end-to-end by
  `TestAccessLogInsideTraceMiddlewareCarriesExportedSpanIDs`.

### Components changed

`kernel/observability`, `kernel/logging`, `adapters/tracing/otel`, `kernel/tracing`.

### Files changed

`kernel/observability/correlation.go` (new), `kernel/logging/logging.go`,
`adapters/tracing/otel/otel.go`, `kernel/tracing/tracing.go` (helpers). Tests:
`kernel/logging/correlation_test.go` (new), `kernel/observability/correlation_test.go` (new).

### Interfaces introduced or changed

Exported: `NewCorrelatingHandler`, `ContextWithSpan`, `SpanFromContext`.

### Configuration changes

None — correlation is structural, per plan.

### Schema or migration changes

None.

### Security changes

None new; redaction coexistence proven (`TestCorrelationCoexistsWithSecretRedaction`).

### Observability changes

This task is the correlation mechanism.

### Tests added or modified

Fail-first positive case (steps 3/5): written and captured FAILING before the `logging.New` wiring
(`evidence/tests/ev-001-fail-first-before.txt`), passing after (`…-after.txt`). Negative case
(step 6): key-ABSENCE assertions for background ctx, NoOpTracer ctx, and empty-TraceID span.
Benchmark (step 7): plain vs wrapped, no-span path, `-count=5` → 0 allocs/op both. Regression
(step 8): all touched suites `-race` green.

### Commits

None — conductor owns commits.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Injected attrs land inside any open `WithGroup` group (standard slog record-attr semantics);
unsampled-but-valid spans still correlate (intentional).

### Follow-up items

None.

### Relationship to the approved plan

Both deferred decisions resolved as the plan's stated preferences; the leaf-package extraction is
DEV-W01-E02-S001-001 (`../deviations.md`).

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E02-S001-01 | Fail-first unit test: positive-case correlation, real span active | Local / CI, Go test runner | Fails before wiring, passes after | unit-test report | framework architecture lead |
| AC-W01-E02-S001-02 | Negative-case unit test: no active span / no-op span, key-absence assertion | Local / CI, Go test runner | Keys genuinely absent | unit-test report | framework architecture lead |
| AC-W01-E02-S001-03 | Benchmark: `go test -bench -benchmem`, no-op tracer path, before/after comparison | Local / CI, Go benchmark runner | No allocation regression | benchmark result | framework architecture lead |

### Actual result

All tests and the benchmark pass; see story `verification.md` per-AC table.

### Pass or fail

Pass (all three ACs).

### Evidence identifier

EV-W01-E02-S001-001, -002, -003.

### Execution date

2026-07-13, ~07:26–07:35 UTC.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (HEAD) + uncommitted working change.

### Environment

Local dev machine, macOS Darwin 25.5.0 arm64, go1.26.5.

### Reviewer

Pending — conductor gate (must confirm key-absence assertion shape, RISK-W01-E02-002).

### Findings

None open.

### Retest status

Final `-race` sweep green (story `evidence/regression/`).

### Final conclusion

Task complete; all three story ACs proven at this task.

## Deviations Record

*Per mandate §8.9. Initially state that deviations are not yet known. The approved plan must not be
silently altered to hide deviations.*

One story-level deviation applies (DEV-W01-E02-S001-001, `../deviations.md`); no task-specific
deviation beyond it.

### Deviation ID

Not applicable.

### Approved plan

Not applicable.

### Actual implementation

Not applicable.

### Reason

Not applicable.

### Impact

Not applicable.

### Risks

Not applicable.

### Approval

Not applicable.

### Compensating controls

Not applicable.

### Follow-up work

Not applicable.
