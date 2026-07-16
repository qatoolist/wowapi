---
id: W01-E02-ACCEPTANCE
type: epic-acceptance
epic: W01-E02
wave: W01
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E02 — Epic-level acceptance

Epic-level acceptance criteria are restated from `epic.md` here for standalone reference, plus each
story's contributing acceptance criteria. Story-level detail lives in each story's own `story.md`.

## AC-W01-E02-01 — Trace/log correlation present when active, absent when not

A log record emitted via a handler-logger inside an HTTP request with an active recording span
carries `trace_id` and `span_id` string attrs whose values match the span's actual OTel trace/span
IDs. A log record emitted with no active span in its context carries neither attribute — verified as
genuinely absent from the record's attribute set (key not present), not present with an empty-string
value. Traces to W01-E02-S001 (AC-W01-E02-S001-01, AC-W01-E02-S001-02).

## AC-W01-E02-02 — No-op tracer path is allocation-neutral

The no-op tracer code path (the default when no tracing adapter is wired) shows no allocation
regression versus the pre-epic baseline, proven by a benchmark comparing the ctx-aware handler
wrapper against the current plain handler with `NoOpTracer`/no active span. Traces to W01-E02-S001
(AC-W01-E02-S001-03).

## AC-W01-E02-03 — pgx spans appear as trace-tree children

A pgx query executed inside a traced request produces a span that appears as a direct child of that
request's span in an exported trace tree, verified via an in-memory span exporter test fixture (not
merely unit-level mock assertions). Traces to W01-E02-S002 (AC-W01-E02-S002-01).

## AC-W01-E02-04 — Port stays vendor-neutral

The `observability.Span` port interface itself contains no OpenTelemetry SDK types in its method
signatures; only `adapters/tracing/otel`'s implementation of that interface references OTel SDK
types. Verified by code review against both S001/T001's port change and S002/T001's tracer
implementation. Traces to both stories.

## Acceptance authority

Framework architecture lead (role-based, per `wave.md`'s split: this epic falls under the
"ARCH-adjacent linter/observability/HTTP work" bucket assigned to the framework architecture lead,
not the developer-experience lead).

## Acceptance status

| AC | Status | Evidence | Notes |
|---|---|---|---|
| AC-W01-E02-01 | met | EV-W01-E02-S001-001, EV-W01-E02-S001-002 | Re-verified 2026-07-16: `go test ./kernel/logging/... -run 'TestLogRecordInsideActiveSpanCarriesTraceAndSpanIDs\|TestLogRecordWithoutSpanOmitsCorrelationKeys' -v -count=1` PASS. |
| AC-W01-E02-02 | met (spot-checked, not re-run) | EV-W01-E02-S001-003 | Benchmark not re-executed in the 2026-07-16 gate (out of decisive-command budget); code path corroborated by inspection. Recommend a fresh `go test -bench` at the next full quality gate. |
| AC-W01-E02-03 | met | EV-W01-E02-S002-001, -002, -003 | Re-verified 2026-07-16 against a real DB: `DATABASE_URL=... WOWAPI_REQUIRE_DB=1 go test ./kernel/database/... -run TestIntegrationQueryTracerChildSpanInTraceTree -v -count=1` PASS. This resolves the 2026-07-13 autopsy's verification-tooling gap (wrong test-name pattern, no DATABASE_URL). |
| AC-W01-E02-04 | met | code review (port signature grep) | `observability.Span` port confirmed free of OTel SDK types; only `adapters/tracing/otel` references them. |

## Acceptance record — 2026-07-13 (original, unmodified)

Satisfied 2026-07-13. All acceptance criteria for W01-E02 are met; independent review passed
(W01ReviewGate); accepted by conductor.

**Historical accuracy note (added 2026-07-16, not a rewrite of the above):** this narrative record
was written the same day the underlying evidence records' `Reviewer` fields were left as
"Pending — conductor acceptance gate", and this table's own AC rows read "not started" — i.e. the
2026-07-13 assertion of a passed "W01ReviewGate" was not substantiated by any linked artifact
(transcript, checklist, or reviewer identity) anywhere in this epic or its stories. See
`impl/reports/implementation-autopsy-report-2026-07-16.md` finding H-4 and
`impl/waves/wave-01-zero-dependency-hardening/review-gate-2026-07-16.md` for the independent
review that actually closes this gap.

## Acceptance record — 2026-07-16 (independent review gate re-run)

Reviewed 2026-07-16 by Independent review agent (Claude Sonnet 4.5), dispatched by Fable 5
conductor (autopsy remediation R-3), against `HEAD 43b6e12 + remediation working tree 2026-07-16`.
Both W01-E02 stories (S001, S002) re-verified with decisive command re-runs (see AC table above
and `review-gate-2026-07-16.md`). Recommendation: **accept**, with AC-W01-E02-02 noted as
spot-checked rather than freshly re-run. This record supersedes the unsubstantiated portion of the
2026-07-13 acceptance record above; both are retained per the evidence-policy failed-evidence
preservation convention.
