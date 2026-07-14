---
id: PLAN-W00-E01-S001
type: plan
parent_story: W00-E01-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Plan — W00-E01-S001 — Verify workflow and boot composition slices at current HEAD

This is a re-verification plan, not an implementation plan. Per mandate §8.5 and this story's
disposition context: **no code changes are expected or in scope**. The "proposed architecture" and
"expected package changes" that a normal `plan.md` would describe do not apply here in the usual
sense — instead this plan describes the verification/evidence-collection approach: which commands
will be re-run, in what environment, what "pass" means for each, and how evidence will be
registered.

Per mandate §8.5, verbatim: "Do not invent precise code changes where the repository does not yet
provide enough information. Clearly distinguish confirmed facts, planned changes, and
implementation assumptions." This plan follows that discipline throughout.

## Proposed architecture

Not applicable. No architectural change is proposed. The "architecture" this story engages with is
the *existing* composition-root behavior of `kernel/workflow/runtime.go`, `app/boot.go`, and
`kernel/kernel.go` — re-confirmed, not redesigned.

## Implementation strategy

Not applicable in the code-change sense. The strategy is a verification strategy: for each of the
three finding-slices, (1) confirm the relevant source file still contains the behavior the
plan/review documents describe, (2) re-run the exact named test command(s) at the story's closing
commit SHA, (3) record the result as a mandate-§10-conformant evidence record, (4) if the result is
not `pass`, preserve the `failed` evidence record and open a follow-up remediation task under the
finding's canonical target story rather than fixing it inside this story.

## Expected package or module changes

None. Confirmed fact: no production code in `kernel/workflow/`, `app/`, `kernel/`, or
`kernel/authz/` is expected to change as part of this story.

## Expected file changes where determinable

None. This story reads and re-tests the following files without modifying them:

- `kernel/workflow/runtime.go`, `kernel/workflow/runtime_extra_test.go`,
  `kernel/workflow/runtime_lifecycle_test.go`, `kernel/workflow/runtime_test.go`,
  `testkit/workflowsim_cov_test.go`.
- `app/boot.go`, `app/boot_extra_test.go`.
- `kernel/kernel.go`, `kernel/authz/caching_internal_test.go`, `kernel/kernel_rules_test.go`.

If Task 001, 002, or 003 discovers a regression, the *fix* file changes belong to a new task under
the finding's canonical target story (`W03-E05-S001` / `W05-E03-S002` / `W05-E04-S001`
respectively), not to this plan.

## Contracts and interfaces

None changed. No interface change is proposed.

## Data structures

None changed.

## APIs

None changed.

## Configuration changes

None. AR-04 T1's subject matter (rejecting unknown `modules.<typo>` config namespaces) is itself a
configuration-validation *behavior* being re-verified, not a configuration change this plan makes.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None introduced. `-race` is used on the re-run commands for SEC-02 (`kernel/workflow/...`) and
AR-06 (`kernel/authz/...`, `kernel/...`) specifically because those are concurrency-sensitive
composition paths already covered by the existing race-detector test suite; this plan does not add
new concurrency behavior.

## Error-handling strategy

Not applicable — no new error-handling code is introduced. AR-04 T1's "deterministic named error"
for unknown config namespaces is the *subject under test*, confirmed to still exist, not authored
here.

## Security controls

Not applicable in the sense of introducing new controls. SEC-02's fail-closed behavior (workflow
privileged operations deny by default) is the security control under re-verification; this plan
introduces no new control and changes no existing one.

## Observability changes

None.

## Testing strategy

Re-run the existing test suites exactly as named, using `-race` where the plan/review documents
specify it:

- **SEC-02 (Task 001)**: `go test ./kernel/workflow/... -race`.
- **AR-04 T1 (Task 002)**: `go test ./app/... -run Boot` (or the specific boot-namespace-rejection
  test, to be identified by name during task execution), plus a full `go test ./...` to confirm no
  unrelated regression.
- **AR-06 T1 (Task 003)**: `go test ./kernel/authz/... -race` and
  `go test ./kernel/... -run TestKernelRules -race` (or the equivalent command covering
  `kernel_rules_test.go`'s sentinel-store-injection assertion — exact `-run` pattern to be confirmed
  during task execution since the precise test function name is not yet confirmed from this plan).

No new tests are written by this story. If a regression is found, whether a new test is needed to
pin the fix is a decision for the remediation task under the finding's canonical target story, not
this plan.

## Regression strategy

The regression-detection mechanism *is* this story: re-running the named commands at current HEAD
is how a regression since the reviewed SHA would be caught. If a regression is found, it is recorded
as `failed` evidence (never silently retried until green, never deleted) and escalated via a new
task per "Rollback strategy" below.

## Compatibility strategy

Not applicable — no code change, so no compatibility question arises from this story itself.

## Rollout strategy

Not applicable. Verification-only story; nothing is rolled out.

## Rollback strategy

Not applicable to this story's own work (no change is deployed). If a regression is discovered:
the affected story does not move to `accepted`; a new remediation task is opened under the
finding's canonical target story (`W03-E05-S001` for SEC-02, `W05-E03-S002` for AR-04,
`W05-E04-S001` for AR-06) per `requirement-inventory.md`'s target column; the `failed` evidence
record stays in `evidence/index.md` permanently, superseded only by a later `retested` record once
the remediation lands.

## Implementation sequence

The three tasks are independent and disjoint in package/file scope (`kernel/workflow` vs. `app` vs.
`kernel`+`kernel/authz`) and may execute in any order or in parallel — no task's evidence depends on
another task's output. Suggested sequence, purely for reviewer convenience (not a dependency):
Task 001 (SEC-02) → Task 002 (AR-04) → Task 003 (AR-06).

## Task breakdown

- **W00-E01-S001-T001** — Re-verify SEC-02 workflow fail-closed behavior. Related AC:
  AC-W00-E01-S001-01.
- **W00-E01-S001-T002** — Re-verify AR-04 T1 boot-time unknown-namespace rejection. Related AC:
  AC-W00-E01-S001-02.
- **W00-E01-S001-T003** — Re-verify AR-06 T1 `authzStore` composition (no duplicate
  `authz.NewStore()` call). Related AC: AC-W00-E01-S001-03.

## Expected artifacts

Three test-execution-log artifacts (post-implementation lifecycle stage — this story's
"implementation" is running verification commands), one per task, registered in
`artifacts/index.md`. No pre-implementation or implementation-stage artifact is expected since no
design or code change occurs.

## Expected evidence

Three evidence records, `EV-W00-E01-S001-01` through `EV-W00-E01-S001-03` (planned ID format — not
yet existing), one per acceptance criterion, registered in `evidence/index.md` once each command has
actually been executed. Evidence type: test-execution log (race-detector log for AC-01/AC-03;
test-execution log plus full-suite green-check log for AC-02).

## Unresolved questions

- The exact `-run` pattern/test function name for AR-04 T1's boot-namespace-rejection test is not
  yet confirmed from this plan (referenced in the calling instructions as "the specific boot
  namespace test," name not given) — must be identified from `app/boot_extra_test.go` during Task
  002 execution.
- The exact test function name inside `kernel_rules_test.go` / `kernel/authz/caching_internal_test.go`
  that performs the AR-06 sentinel-store-injection assertion is not yet confirmed — must be
  identified during Task 003 execution.
- Whether `testkit/workflowsim_cov_test.go` (SEC-02) or the `kernel/authz`/`kernel` DB-backed tests
  (AR-06) require a live Postgres instance via `make ci-container` is not yet confirmed — must be
  determined during task execution; if required and unavailable, this is a RISK-W00-002-class
  environment issue, not a code regression.
- Whether any code drift has occurred between the plan/review documents' review commit and this
  story's actual closing commit SHA is unknown until the tasks are executed — this plan assumes
  "no drift" only as a starting hypothesis to be confirmed or refuted, not as a fact.

## Approval conditions

This plan is considered approved and ready for implementation (i.e., for task execution to begin)
once: an owner is assigned to the story; the story has moved to `ready` per
`definition-of-ready.md`; and the unresolved questions above are either resolved or explicitly
acknowledged as "to be resolved during task execution" by the assigned owner, consistent with this
plan's stated intent not to invent specifics the repository does not yet confirm.
