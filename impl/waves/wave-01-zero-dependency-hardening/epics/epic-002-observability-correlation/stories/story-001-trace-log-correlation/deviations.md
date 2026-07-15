---
id: DEV-W01-E02-S001
type: deviations-record
parent_story: W01-E02-S001
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Deviations — W01-E02-S001

One deviation recorded. Per mandate §2.6 the approved `plan.md` is left unmodified; this record is
the divergence trail.

The two decisions `plan.md` explicitly deferred to implementation time were resolved as planned and
are NOT deviations (recorded in `implementation.md`): span retrieval = package-level
`ContextWithSpan`/`SpanFromContext` helpers populated by real adapters' `StartSpan`; wiring point =
`logging.New` (approach (a), broader coverage).

## DEV-W01-E02-S001-001 — Span/Tracer port definition extracted to leaf package `kernel/tracing`

### Approved plan

`plan.md` "Expected package or module changes": widen `Span` in place in
`kernel/observability/tracing.go`; no new package.

### Actual implementation

The `Tracer`/`Span` interfaces, `NoOpTracer`, and the `ContextWithSpan`/`SpanFromContext` helpers
are DEFINED in a new stdlib-only leaf package `kernel/tracing`; `kernel/observability` re-exports
them via type aliases (`type Tracer = tracing.Tracer`, `type Span = tracing.Span`,
`var NoOpTracer Tracer = tracing.NoOpTracer`) and thin forwarding functions. Every existing
consumer keeps compiling and binding `observability.*` unchanged (alias identity, verified by the
unmodified pass of all pre-existing suites).

### Reason

Discovered at implementation time by W01-E02-S002: `kernel/database` cannot import
`kernel/observability` — pre-existing edges `kernel/observability → kernel/httpx → kernel/authz →
kernel/database` make it an import cycle (confirmed by `go vet`; also independently hit and
reported by the W01Http worker). The port a deep kernel package consumes must live below the
middleware stack. This condition was unknowable from the planning-time file inspection, which
looked at declarations, not the transitive import graph.

### Impact

Zero source impact on consumers (aliases); one new kernel package; S002's `WithQueryTracer`
signature uses `tracing.Tracer`, which is type-identical to `observability.Tracer`. The story's
outcome contract (ACs) is unchanged and fully proven.

### Risks

Two names for one port could confuse future readers — mitigated by doc comments in both packages
declaring `kernel/tracing` the defining package and `observability` the façade adapters bind to.

### Approval

Worker-recorded per mandate; pending conductor ratification at story acceptance (flagged for the
independent-review gate).

### Compensating controls

Compile-time alias identity; full `-race` regression sweep over all affected packages
(`evidence/regression/touched-packages-race.txt`); `go vet ./kernel/... ./adapters/...` clean.

### Follow-up work

None required. Optional future consolidation (moving the `Trace` middleware's docs to reference
`kernel/tracing`) can ride any later observability story.
