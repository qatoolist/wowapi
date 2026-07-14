---
id: DEV-W01-E02-S002
type: deviations-record
parent_story: W01-E02-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Deviations — W01-E02-S002

One reference deviation recorded; the anticipated D-08-wording deviation did NOT materialize.

## D-08 wording — NO deviation

The ratified ADR (ADR-W00-E02-S003-008, confirmed before implementation) matches `plan.md`'s
working assumption exactly: thin in-kernel `pgx.QueryTracer` (~50 LOC) over the existing
observability port; `otelpgx` rejected to keep vendor types out of `kernel/database`. The
pre-registered "most likely deviation" is closed without divergence.

## DEV-W01-E02-S002-001 (reference) — port imported from `kernel/tracing`, not `kernel/observability`

### Approved plan

`plan.md` proposed `WithQueryTracer(tr observability.Tracer)` consuming `kernel/observability`.

### Actual implementation

`WithQueryTracer(tr tracing.Tracer)` consuming the new leaf package `kernel/tracing`, which now
DEFINES the port; `observability.Tracer` is a type ALIAS of `tracing.Tracer`, so the signature is
type-identical to the planned one and every composition-root caller passes the same values.

### Reason

`kernel/database` cannot import `kernel/observability`: the pre-existing chain
`observability → httpx → authz → database` makes it an import cycle (confirmed by `go vet` during
this story's implementation; this story was the discovery trigger). Primary record and full
analysis: S001's DEV-W01-E02-S001-001
(`../story-001-trace-log-correlation/deviations.md`).

### Impact / Risks / Compensating controls / Approval / Follow-up

See the primary record. For this story specifically: zero behavioral impact (alias identity), the
D-08 boundary is strengthened (the port package the tracer consumes is stdlib-only), and no OTel
type enters `kernel/database` (verified; RISK-W01-E02-003). Pending conductor ratification with
S001's record.

## DEV (conductor, 2026-07-13) — scaffold templates did not wire WithQueryTracer

DEV (conductor, 2026-07-13): reviewer found scaffold templates did not wire WithQueryTracer despite story compatibility note; conductor wired database.WithQueryTracer(tracer) into both init templates (api runtime pool; worker tracer block moved above pools) and re-ran internal/cli tests (ok 28.9s). Gap closed pre-acceptance.
