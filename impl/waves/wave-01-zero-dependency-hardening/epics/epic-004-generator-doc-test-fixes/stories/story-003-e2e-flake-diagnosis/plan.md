---
id: PLAN-W01-E04-S003
type: plan
parent_story: W01-E04-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W01-E04-S003

**This is an investigation plan**, per mandate §8.5's explicit guidance for cases where implementation
details cannot yet be known: "state what must be determined during the story rather than inventing
specifics." Sections below that would normally describe a fix's architecture, contracts, or data
structures instead describe what T001's investigation will determine, and state plainly that T002's
plan is conditional and deferred until T001 completes. Do not read any section below as a commitment to
a particular fix mechanism.

## Proposed architecture

Not determinable yet. T001's investigation output is a *decision* about what (if anything) T002 should
build or change — there is no proposed architecture for a fix until that decision exists. What can be
stated now is the investigation's own architecture: a reproduction harness running `internal/e2e`'s full
suite repeatedly under `-count`+parallel settings, plus a direct code-reading pass over `internal/e2e`'s
test-setup/fixture code to determine its DB-wiring mechanism.

## Implementation strategy

**T001 (reproduce + investigate):**

1. Run `internal/e2e`'s full suite via `go test -count=N -parallel=P ./internal/e2e/...`, with N and P
   values to be fixed at implementation time within a reasonable CI/local time budget (not specified
   here — an unbounded or arbitrarily large N would itself violate mandate §2.1's doability principle).
   Repeat this run multiple times if the first attempt does not reproduce a failure, up to the planned
   budget.
2. Capture full logs for every run, pass or fail, so a failure (if any) has complete diagnostic context,
   and so a clean run record also has evidentiary value (it is evidence toward "not reproducible," not
   merely an absence of information).
3. Separately, read `internal/e2e`'s test-setup/fixture code directly to determine whether it calls
   `testkit.NewDB` (or equivalent) for per-test database cloning, or has its own, separate DB-wiring
   path. This is a direct code-reading determination, not an inference from the reproduction run's
   outcome — the two investigative steps are independent and both required.
4. Synthesize both findings into a diagnosis: if a failure reproduces, correlate it against the DB-
   wiring determination (e.g., if `internal/e2e` does NOT use `testkit.NewDB`'s isolation and a failure
   reproduces with symptoms consistent with shared-state interference, that would be new, actually-
   verified evidence — a materially different epistemic position than the original, unverified
   "shared-DB concurrency" claim). If no failure reproduces after the planned budget, record that
   outcome plainly, together with the DB-wiring determination (which still has standalone value even
   without a reproduced failure — it answers a real open question about the suite's isolation guarantees
   regardless of outcome).

**T002 (conditional fix):** Deferred. See "Task breakdown" below and `tasks/task-002-conditional-fix.md`
for the illustrative decision branches this task will resolve into once T001 completes.

## Expected package or module changes

T001: none (investigation only, no code change). T002: not determinable until T001 completes — could
range from no code change at all (if the outcome is "not reproducible, monitor") to a change in
`internal/e2e`'s test-setup code (if it is found to bypass `testkit.NewDB`) to a change in `testkit`
itself (if T001 finds a genuine defect in the cloning/cleanup mechanism, which is not currently
suspected but is not ruled out either).

## Expected file changes where determinable

None determinable at planning time beyond T001's own (non-code) reproduction-log and diagnosis-note
output.

## Contracts and interfaces

Not applicable to T001. Not determinable for T002 until T001 completes.

## Data structures

Not applicable.

## APIs

Not applicable.

## Configuration changes

None expected for T001. Not determinable for T002.

## Persistence changes

None expected — `testkit`'s existing per-test database cloning mechanism (if T002 ends up needing to
route `internal/e2e` through it) is itself already-existing infrastructure, not a new persistence
mechanism being introduced by this story.

## Migration strategy

Not applicable.

## Concurrency implications

The entire investigation is fundamentally about concurrency (parallel test execution, potential shared
resource contention) — T001's reproduction protocol is specifically designed to surface concurrency-
related failures if they exist, by running the suite repeatedly under parallel execution rather than
serially.

## Error-handling strategy

Not applicable to T001 (an investigation does not "handle errors," it records outcomes, including
failure outcomes, as data). T002's error-handling strategy, if any code change results, is not
determinable until the fix mechanism is known.

## Security controls

Not applicable.

## Observability changes

None to production systems. T001's own reproduction-run logs are the investigation's observability
artifact, not a production change.

## Testing strategy

T001's "testing strategy" is the investigation itself — running `internal/e2e`'s full suite repeatedly
under `-count`+parallel, capturing complete logs each run. This IS the fail-first evidence per mandate
§13: the reproduction attempt itself is the test, not a separately-authored adversarial fixture (this is
explicitly unusual relative to other stories in this programme, and is called out as such rather than
treated as a gap). T002's testing strategy is not determinable until T001's findings are known — whatever
fix (or non-fix) T002 implements will need its own test appropriate to that specific mechanism.

## Regression strategy

Not applicable to T001. T002's regression strategy is not determinable until the fix mechanism is known;
if T002 changes `internal/e2e`'s or `testkit`'s code, standard regression discipline (existing test
suite still passes) applies at that time.

## Compatibility strategy

Not applicable to T001. Not determinable for T002 until its mechanism is known.

## Rollout strategy

T001 has no rollout (an investigation produces a diagnosis note, not a deployed change). T002's rollout
strategy is not determinable until its mechanism is known.

## Rollback strategy

Not applicable to T001. T002's rollback strategy is not determinable until its mechanism is known — will
be recorded in T002's own task file once the fix branch is selected.

## Implementation sequence

T001 must complete in full — both the reproduction protocol and the DB-wiring determination — before
T002 can even be meaningfully scoped, let alone started. This is a hard sequential dependency, not a
parallelizable one: T002's very shape depends on T001's output.

## Task breakdown

- T001 — reproduce and investigate (produces a decision record: what T002's fix, if any, should be).
- T002 — conditional fix, implemented strictly according to T001's findings; see
  `tasks/task-002-conditional-fix.md` for the illustrative decision-branch space this task will resolve
  into (not a commitment to any one branch).

## Expected artifacts

Reproduction-run log collection (T001); the diagnosis note, functioning as a task-level decision record
(T001); conditionally, whatever code/test change or explicit no-change/monitoring note T002 produces.

## Expected evidence

Evidence at path `evidence/premier/T-TEST-01/`: the reproduction-run artifacts and the resulting
diagnosis note.

## Unresolved questions

- The exact `-count`/`-parallel` reproduction budget (N, P, and number of repeated executions) — to be
  fixed at implementation time within a reasonable CI/local time constraint, not specified here.
- Whether `internal/e2e` uses `testkit.NewDB` cloning or its own DB wiring — this is the central
  question T001 exists to answer; it is explicitly NOT assumed in either direction by this plan.
- Whether, if a failure does reproduce, its root cause will actually correlate with the DB-wiring
  determination, or turn out to be something else entirely (e.g., an unrelated resource-exhaustion
  condition, a timing issue unrelated to database state) — T001's plan does not pre-suppose the
  reproduced failure (if any) will be DB-related at all.

## Approval conditions

T001 is approved for implementation as planned above (its protocol is fully specified). T002 cannot be
approved for implementation until T001 has completed and its findings are reviewed — this plan's
approval for T002 is explicitly deferred, not granted in advance.
