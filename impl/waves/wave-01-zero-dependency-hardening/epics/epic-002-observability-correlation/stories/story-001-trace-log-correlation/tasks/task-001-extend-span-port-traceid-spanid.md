---
id: W01-E02-S001-T001
type: task
title: Extend Span port with TraceID()/SpanID()
status: done
parent_story: W01-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E02-S001-01
  - AC-W01-E02-S001-02
artifacts: []
evidence: []
---

# W01-E02-S001-T001 — Extend Span port with TraceID()/SpanID()

## Task Definition

*Per mandate §8.6. This section defines the task before work begins.*

### Task objective

Widen the `observability.Span` port interface with two new vendor-neutral accessors —
`TraceID() string` and `SpanID() string` — and implement them on both existing implementations
(`noopSpan` in `kernel/observability/tracing.go`, `otelSpan` in `adapters/tracing/otel/otel.go`),
so a consumer (the S001-T002 log handler wrapper, and S002's pgx query tracer) has a vendor-neutral
way to read the active trace/span identifiers.

### Parent story

W01-E02-S001 — Trace/log correlation.

### Owner

Unassigned.

### Status

`done` (2026-07-13).

### Dependencies

None. This is the first task in the story and has no upstream dependency within the epic. It is,
however, a dependency *for* W01-E02-S002-T001 (recorded there, not here).

### Detailed work

1. Add `TraceID() string` and `SpanID() string` to the `Span` interface in
   `kernel/observability/tracing.go:33-39`, with doc comments describing the no-op-vs-real-adapter
   contract (empty string for no-op or invalid trace context; canonical string form otherwise).
2. Implement both methods on `noopSpan` (same file, `tracing.go:54-58`), each returning `""`.
3. Implement both methods on `otelSpan` (`adapters/tracing/otel/otel.go:99-111`), returning
   `s.span.SpanContext().TraceID().String()` and `s.span.SpanContext().SpanID().String()`
   respectively.
4. Confirm no other package in the repository implements `observability.Span` independently (a
   compile-time check across the whole module after the change is sufficient evidence — if the
   module builds, every implementer has the new methods or the build fails, which is itself the
   confirmation).
5. Update the doc comment block on the `Span` interface (`tracing.go:32`) to reflect the widened
   contract.

### Expected files or components affected

- `kernel/observability/tracing.go`
- `adapters/tracing/otel/otel.go`

### Expected output

A compiling module where `observability.Span` has four methods (`End`, `SetAttr`, `RecordError`,
plus the two new accessors), both existing implementations satisfy it, and the compile-time
assurance line already present in `otel.go:114` (`var _ observability.Tracer = (*Tracer)(nil)`)
continues to pass (an equivalent `var _ observability.Span = otelSpan{}` assertion may be added if
not already implicitly covered).

### Required artifacts

ART-W01-E02-S001-001 (extended Span port interface diff) — see `../../artifacts/index.md`.

### Required evidence

Compile/build confirmation is implicit infrastructure, not registered evidence on its own; the
registered evidence for this task's contribution is folded into T002's fail-first/negative-case
tests, since T001 in isolation has no independently observable behavioral acceptance criterion
beyond "the module still builds and the two new methods return the documented values" — verified as
part of T002's tests, which exercise `TraceID()`/`SpanID()` indirectly through the handler wrapper.
A minimal direct unit test for `otelSpan.TraceID()`/`SpanID()` returning the expected values against
a known span may additionally be added here if useful for isolating a failure to this task rather
than T002 — left to implementation-time judgment, recorded in `implementation.md` if added.

### Related acceptance criteria

AC-W01-E02-S001-01, AC-W01-E02-S001-02 (both are ultimately proven at T002, but T001 is the
prerequisite that makes them provable at all).

### Completion criteria

The module builds; `Span` has the two new methods; both implementations satisfy them per the
documented contract; no other `Span` implementer in the repository is broken by the widening.

### Verification method

`go build ./...` across the whole module (confirms no broken implementer); a direct unit test of
`otelSpan.TraceID()`/`SpanID()` against a span with a known, non-empty `SpanContext()` (optional, at
implementation-time judgment); `noopSpan.TraceID()`/`SpanID()` returning `""` (trivial, may be
folded into T002's negative-case test instead of a standalone test here).

### Risks

Low — this is an additive interface widening confirmed safe against the only two known implementers
at planning time (see `story.md` "Compatibility considerations"). The only risk is discovering a
third, previously-unknown `Span` implementer during the build — mitigated by the build itself
surfacing that immediately as a compile error, not a silent gap.

### Rollback or recovery considerations

Revert the commit. No state is created; rollback is a pure code revert with no residual-data
concern.

## Implementation Record

Executed 2026-07-13 by W01Obs.

### What was actually implemented

`Span` widened with `TraceID()`/`SpanID()` exactly per "Detailed work", with one structural
deviation: the port definition (Tracer/Span/NoOpTracer + context helpers) moved to the new
stdlib-only leaf package `kernel/tracing`, re-exported by alias from `kernel/observability` —
required because `kernel/database` (S002's consumer) cannot import `kernel/observability`
(pre-existing import chain observability→httpx→authz→database would cycle). See story
`deviations.md` DEV-W01-E02-S001-001. `noopSpan` returns `""` for both; `otelSpan` returns the
canonical hex strings guarded by `HasTraceID()`/`HasSpanID()`.

### Components changed

`kernel/tracing` (new), `kernel/observability`, `adapters/tracing/otel`.

### Files changed

`kernel/tracing/tracing.go` (new), `kernel/observability/tracing.go` (aliases; middleware
unchanged), `adapters/tracing/otel/otel.go`. Test fakes widened: `kernel/observability/tracing_test.go`,
`kernel/outbox/outbox_test.go`, `kernel/jobs/trace_test.go`, `kernel/notify/trace_test.go`.

### Interfaces introduced or changed

`tracing.Span` (= `observability.Span`) +2 methods; `tracing.Tracer`/`Span`/`NoOpTracer` aliases.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None.

### Observability changes

The port extension itself.

### Tests added or modified

No standalone T001 test (per this task's own "Required evidence" folding note): the accessors are
proven through T002's tests — logged-ID equality against `span.TraceID()`/`SpanID()` and the
middleware end-to-end assertion against the exported span context. Whole-module build confirms no
third implementer broke (the four test fakes were the only fixups, all in-repo).

### Commits

None — conductor owns commits.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

None.

### Follow-up items

None.

### Relationship to the approved plan

Matches, except the leaf-package extraction (DEV-W01-E02-S001-001, story-level `deviations.md`).

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E02-S001-01 (prerequisite portion) | `go build ./...`; optional direct unit test of `otelSpan.TraceID()`/`SpanID()` | Local / CI, Go toolchain | Module builds; accessors return documented values | build log / unit-test report | framework architecture lead |
| AC-W01-E02-S001-02 (prerequisite portion) | `noopSpan.TraceID()`/`SpanID()` return `""` | Local / CI, Go toolchain | Both return empty string | unit-test report | framework architecture lead |

### Actual result

`go build ./...` clean; `go vet ./kernel/... ./adapters/...` clean; both accessor contracts proven
via T002's suites (EV-W01-E02-S001-001/-002).

### Pass or fail

Pass.

### Evidence identifier

Folded into EV-W01-E02-S001-001/-002 per this task's "Required evidence".

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (HEAD) + uncommitted working change.

### Environment

Local dev machine, macOS Darwin 25.5.0 arm64, go1.26.5.

### Reviewer

Pending — conductor gate.

### Findings

None beyond the recorded deviation.

### Retest status

Included in the final `-race` regression sweep (story `evidence/regression/`).

### Final conclusion

Task complete; port widened without breaking any implementer.

## Deviations Record

*Per mandate §8.9. Initially state that deviations are not yet known. The approved plan must not be
silently altered to hide deviations.*

One deviation, recorded at story level: DEV-W01-E02-S001-001 (port definition extracted to leaf
package `kernel/tracing` with observability aliases; import-cycle forcing condition). Full record
in `../deviations.md`; fields below reference it rather than duplicating.

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
