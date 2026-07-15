---
id: IMPL-W01-E04-S003
type: implementation-record
parent_story: W01-E04-S003
status: final
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E04-S003

This record aggregates the implementation reality of the story across all of its tasks. Do not
pre-populate implementation claims for work that has not yet occurred.

## What was actually implemented

Executed 2026-07-13 against pinned commit `0a31186cada5c275a588c74081cf977adf346e61`.

**T001 (reproduce + investigate):** the full reproduction protocol — 1 preflight, 4 invocations
of `go test -count=5 -parallel=4 ./internal/e2e/`, 3 targeted stress iterations (e2e `-count=2`
concurrent with `go test ./testkit/ ./internal/cli/` against the same base database), and
`go test -race -count=2` — with `WOWAPI_REQUIRE_DB=1` throughout. Result: **29/29 clean
executions; historical failure not reproduced.** The first 4 invocations (main working tree)
produced 16 failures fully attributed to sibling wave workers' in-flight edits compiling into
the suite via the product `replace` directive (see `deviations.md`); the protocol was
re-executed in a detached worktree pinned at the SHA above. Independently, direct code reading
determined `internal/e2e` uses its **own DB wiring** (raw `DATABASE_URL`; the scaffolded
product's migrate applies kernel migrations directly to the base database; no `testkit.NewDB`,
no `t.Parallel`). The withdrawn "shared-DB concurrency" cause is not re-asserted.

**T002 (conditional outcome):** monitoring-only branch (task-002 illustrative branch 3) — no
code fix; the historical failure is downgraded to a programme-level monitoring item with an
explicit triage protocol (`evidence/premier/T-TEST-01/diagnosis-note.md` §5-§6).

## Components changed

None — investigation + documentation only. `internal/e2e/` and `testkit/` were read, executed,
and stressed, never modified.

## Files changed

No production files. New governance/evidence files under this story directory only:
`evidence/premier/T-TEST-01/` (diagnosis-note.md, reproduction-runs.md, logs/ — 16 log files),
plus updates to this story's tasks/, indices, verification, deviations, and closure records.

## Interfaces introduced or changed

None yet.

## Configuration changes

None yet.

## Schema or migration changes

None yet.

## Security changes

None yet.

## Observability changes

None yet.

## Tests added or modified

None — existing tests executed repeatedly per the investigation plan.

## Commits

None yet.

## Pull requests

None yet.

## Implementation dates

2026-07-13 (single day).

## Technical debt introduced

None yet.

## Known limitations

The original failure's log was never preserved, so its cause is permanently unknowable; the
non-reproduction is bounded by this protocol's budget (29 executions, one host, one PG16
instance), not a proof of impossibility. Both limits are stated in the diagnosis note.

## Follow-up items

Programme-level monitoring item (diagnosis-note §6): on any future full-suite `internal/e2e`
failure, preserve the log before rerunning and classify the failing step (tree-compilation vs
runtime) before attributing cause.

## Relationship to the approved plan

Followed as approved. T001's budget (N=5, P=4, 4 invocations + stress + race) was fixed at
implementation time within the plan's bounded-budget guidance. T002 resolved to a branch
explicitly foreseen among the illustrative options (branch 3, monitoring-only). One
environment adaptation — isolated-worktree re-execution after sibling-worker tree
contamination — is recorded in `deviations.md`; the protocol itself was unchanged.
