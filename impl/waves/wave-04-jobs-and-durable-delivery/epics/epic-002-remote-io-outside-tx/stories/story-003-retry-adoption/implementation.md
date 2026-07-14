---
id: IMPL-W04-E02-S003
type: implementation-record
parent_story: W04-E02-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W04-E02-S003

*This record aggregates the implementation reality of the story across all of its tasks. No
implementation has occurred yet — this file must not be pre-populated with implementation claims per
mandate §8.7.*

## What was actually implemented

- Added `kernel/retry`, a thin shared wrapper over `cenkalti/backoff/v5` that exposes
  attempt-indexed backoff durations (`Schedule.Next`) and a `SequenceBackOff`
  implementation for exact schedule parity.
- Replaced the hand-rolled `backoff` function and `backoffSchedule` table in
  `kernel/notify/service.go` with `notifyBackoff`, a `retry.Schedule` over the same
  `[30s, 2m, 10m]` durations.
- Replaced the hand-rolled `backoff` function in `kernel/webhook/service.go` with
  `webhookBackoff`, a `retry.Schedule` over the same `[1s, 5s, 30s, 2m, 5m]` durations.
- Added `kernel/retry/retry_test.go` with schedule-parity and fault-injection tests.
- Updated `kernel/notify/internal_test.go` and `kernel/webhook/internals_test.go` to
  exercise the new library-backed schedules.

## Components changed

- New `kernel/retry` shared retry/backoff component.
- `kernel/notify` and `kernel/webhook` now consume the shared component instead of
  maintaining duplicate hand-rolled backoff functions.

## Files changed

- `kernel/retry/retry.go` (new)
- `kernel/retry/retry_test.go` (new)
- `kernel/notify/service.go`
- `kernel/notify/internal_test.go`
- `kernel/webhook/service.go`
- `kernel/webhook/internals_test.go`
- `go.mod` / `go.sum` (`cenkalti/backoff/v5` promoted to direct dependency)

## Interfaces introduced or changed

*Not yet implemented.*

## Configuration changes

None. Retry schedules remain compile-time constants per call site, matching the
previous hand-rolled behavior.

## Schema or migration changes

*Not applicable — this story is a code-level dependency swap; it has no schema or data migration of
its own (see `story.md` "Migration considerations").*

## Security changes

*Not yet implemented.*

## Observability changes

*Not yet implemented.*

## Tests added or modified

- `kernel/retry/retry_test.go` — `TestScheduleSequenceParity`,
  `TestScheduleExponentialBackOff`, `TestRetryFaultInjection`,
  `TestRetryPermanentError`.
- `kernel/notify/internal_test.go` — `TestBackoffClamps` updated to test
  `notifyBackoff.Next`.
- `kernel/webhook/internals_test.go` — `TestBackoff` updated to test
  `webhookBackoff.Next`.

## Commits

Pending final commit/PR for the W04 wave.

## Pull requests

*Not yet implemented.*

## Implementation dates

*Not yet implemented.*

## Technical debt introduced

*Not yet implemented — none anticipated by the plan.*

## Known limitations

*Not yet implemented.*

## Follow-up items

*Not yet implemented.*

## Relationship to the approved plan

Implementation followed the approved plan: both hand-rolled backoff implementations
(notify and webhook) were located, their schedules documented as parity baselines,
`cenkalti/backoff/v5` was promoted to a direct dependency, and both were replaced
with library-backed schedules. No deviations from `plan.md`.
