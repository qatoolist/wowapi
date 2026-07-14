---
id: PLAN-W01-E01-S003
type: plan
parent_story: W01-E01-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W01-E01-S003

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information.

## Proposed architecture

No architectural change. This story is pure CI-workflow and git-hook configuration: one new CI step
(`go mod verify`), one new or extended CI step (license scanning), one confirmation activity (nightly
fuzz schedule), and one small script fix (pre-push hook). No new package, interface, service, or
runtime contract is introduced.

## Implementation strategy

1. Re-read `.github/workflows/ci.yml`, `.github/workflows/security-scan.yml`, and
   `.githooks/pre-push` fresh, at this story's actual start commit, to confirm or correct the exact
   line numbers and job/step structure cited in `story.md`'s "Current-state assessment" — this is the
   fail-first/confirm-first step for a story whose gaps are largely already known but whose exact
   locations may have drifted.
2. Add a `go mod verify` step to `ci.yml`'s build/test pipeline (exact job placement — e.g. alongside
   the existing `go vet`/build steps — to be determined at implementation time from the pipeline's
   actual job graph).
3. Implement the license-signal choice (see "License signal decision" below): add `license` to the
   `trivy` job's `scanners:` list in `security-scan.yml` (the planned default choice), or add a new
   `go-licenses` step if the fresh re-read in step 1 changes the picture.
4. Confirm the nightly fuzz schedule: re-inspect `ci.yml`'s `schedule:`/`cron:` trigger and the job
   graph it reaches, and — where practically feasible within this story's execution window — observe an
   actual scheduled or manually-triggered (`workflow_dispatch`, if available) run to confirm the fuzz
   seed-corpus replay step actually executes under that trigger. Produce a confirmation/audit record
   (see "Task breakdown" grouping decision below) stating what was confirmed and explicitly restating
   the `-fuzz=` coverage-guided-generation gap as W07 scope.
5. Fix `.githooks/pre-push`: change the `go test ./...` invocation (or its surrounding logic) so that
   DB-gated tests failing to run because `WOWAPI_REQUIRE_DB` (or equivalent) is unset produces a loud,
   non-zero-exit failure rather than a silent pass — exact mechanism to be determined at implementation
   time (see "Unresolved questions").
6. Re-run each changed CI step and the modified hook to confirm fail-before/pass-after behavior per the
   acceptance criteria.

## License signal decision

Per mandate §18 ("do not silently resolve ambiguous architecture decisions... record assumptions
explicitly"), this is a genuine judgment call this story must document, not silently resolve.

**Confirmed facts (as of this story's writing, 2026-07-12, re-confirm at implementation time):**

- Trivy's `filesystem scan (trivy)` job in `security-scan.yml` configures `scanners: vuln,secret,misconfig`
  — the `license` scanner is not enabled. Trivy natively supports a `license` scanner mode
  (`scanners: license` or as an addition to the existing list) that would enumerate dependency licenses
  as part of the same existing job, with no new Action, no new pinned SHA, and no new external tool to
  vet.
- `dependency-review-action`'s `license-check: true` setting exists in the same file's
  `dependency-review` job, but that job only runs `if: github.event_name == 'pull_request' && ...` — a
  property of the Action's own design (it diffs a PR's dependency graph against the base branch; it has
  no meaningful "scan everything" mode outside a PR event), not a consequence of the repository's
  visibility. The repository has been public since 2026-07-03, so "the repo is private" is not the
  reason this signal is dormant outside PR events — it is dormant on non-PR events (direct pushes,
  scheduled runs) purely because that is how the Action is designed to run. This distinction matters:
  fixing repository visibility would not change this dormancy; only running the check somewhere that is
  not gated to `pull_request` events would.
- `go-licenses` (`github.com/google/go-licenses`) is a separate, standalone Go tool that walks a
  module's dependency graph and reports each dependency's detected license — it is not currently
  referenced anywhere in this repository's CI configuration.

**Planned choice: enable Trivy's `license` scanner, by adding `license` to the existing `trivy` job's
`scanners:` list in `security-scan.yml`.**

**Rationale for this planned choice:**

- Zero new tooling surface: the `trivy` job, the pinned `aquasecurity/trivy-action` SHA, and the job's
  existing `exit-code: "0"` (informational, non-blocking) posture are already present and already
  vetted for this repository. Adding `license` to an existing `scanners:` list is a one-line
  configuration change with no new Action to pin, no new SHA to audit, and no new binary dependency —
  directly consistent with this epic's stated framing (`epic.md`: "every task is 'flip a config flag'... no new binary, no new dependency, no new external service").
- `go-licenses` would be a genuinely new external tool: a new Go binary to install (or `go run` invoke)
  in CI, a new step, a new pin/version to track, and a tool whose output format and failure modes are
  not yet exercised anywhere in this repository's pipeline. It is a reasonable alternative and is not
  rejected outright — it remains a valid fallback if Trivy's license-scanning coverage or output quality
  proves inadequate at implementation time (see "Unresolved questions") — but it carries strictly more
  net-new surface than extending an already-present, already-informational-by-design Trivy job.
- Both options satisfy FBL-07's own text ("license signal (Trivy license scanner or go-licenses while
  dependency-review is visibility-dormant)") equally as a matter of requirement compliance — the choice
  between them is a pure implementation-cost/surface-area judgment call, not a compliance question.

**This choice may be revisited at implementation time** if the fresh re-read in "Implementation
strategy" step 1 finds that `security-scan.yml`'s exact state has changed materially (e.g. the `trivy`
job has since been restructured, or Trivy's license-scanning mode for this repository's dependency mix
proves to have materially worse SPDX-detection coverage than `go-licenses` would). If the choice is
revised at implementation time, that revision is recorded as a deviation in `deviations.md`, not by
silently editing this plan after the fact (mandate §2.6).

## Nightly-fuzz-confirmation scope boundary

This is the second genuine judgment call this story documents rather than silently resolves.

**Confirmed facts (session delta SD-02, `requirement-inventory.md` §E, and this story's own direct
re-read as of 2026-07-12):** the nightly schedule already exists since PR #24 — `ci.yml` has a
`schedule:`/`cron: "17 3 * * *"` trigger, with header comments describing a test job that runs "fuzz
seed corpus" replay "on main pushes and the nightly schedule." The remaining gap, per FBL-07's own
disposition note, is that "the fuzz portion [is] still seed-replay only" — meaning the nightly run
replays existing saved seed-corpus inputs, not real `-fuzz=` coverage-guided generation that would
explore genuinely new inputs.

**Scope boundary, stated explicitly:** this story's task is narrowly to *confirm* — by direct
inspection of `ci.yml`'s actual job graph, and by observing an actual run where feasible — that the
nightly schedule (a) really exists, (b) really fires on a nightly cadence, and (c) really reaches and
executes the fuzz-seed-corpus-replay step, not merely that the header comments claim it does. This
story does **not** implement the `-fuzz=` flag or any coverage-guided fuzzing capability — that is
`REL-04` T8 and `PERF-06` T3/T4, both explicitly assigned W07 scope with shared ownership to "PF-REL"
per `premier-framework-implementation-plan.md`. The boundary is stated this explicitly so that:

- No one assumes W07 already covers "confirm the nightly schedule exists and is wired correctly" and
  skips verifying it (the schedule confirmation would then never actually happen, since W07's own scope
  is about adding `-fuzz=`, not about auditing the pre-existing schedule).
- No one implementing this story accidentally scope-creeps into adding the `-fuzz=` flag itself, which
  would silently duplicate work against W07's task list and blur REL-04/PERF-06's ownership.

## Task breakdown grouping decision

Per the epic's own instruction (`epic.md` and the calling context establishing this story): S003 splits
into 3 tasks (go-mod-verify CI step; license-signal enablement; pre-push hook fix) at minimum, with the
nightly-fuzz-schedule confirmation either folded into one of the three or given its own 4th task — a
call this plan makes explicitly.

**Decision: give nightly-fuzz-schedule confirmation its own 4th task
(`W01-E01-S003-T004`), separate from the `go mod verify` task (`W01-E01-S003-T001`).**

**Rationale**, per mandate §12's decomposition criteria ("split when materially different risk/
evidence/ownership; don't over-fragment"):

- **Evidence type differs materially.** `go mod verify`'s evidence is a straightforward CI run log
  showing a new step executing and passing — a standard "before/after" artifact identical in kind to
  every other CI-step-addition task in this epic. The nightly-fuzz-schedule confirmation's evidence is
  an audit/confirmation note: a record of what was inspected and, where feasible, what an actual
  scheduled/triggered run showed — closer in kind to a review record than to a code-change verification
  log. Folding a confirmation-type evidence artifact into a code-addition task's evidence record would
  blur two different evidence types under one task ID, working against mandate §10's evidence-type
  clarity.
- **Risk profile differs.** `go mod verify`'s risk is essentially zero (a well-understood, standard Go
  toolchain command). The nightly-fuzz confirmation's risk is different in kind: the risk is not "does
  the change work" (there is no change) but "does the confirmation actually verify real behavior,
  rather than re-stating the header comment's claim as fact" — a scope-boundary risk (see "Nightly-fuzz-
  confirmation scope boundary" above) that is easy to get wrong by under-verifying (accepting the
  comment at face value) or over-verifying (drifting into implementing `-fuzz=` itself). Keeping this as
  its own task makes that specific risk independently trackable and independently reviewable, consistent
  with mandate §14's requirement that reviewers check "no source requirement has been silently dropped"
  and this epic's own acceptance criterion AC-W01-E01-04 calling out this exact scope-boundary check by
  name.
- **Not excessive fragmentation.** Four tasks for four gaps that are already individually distinct in
  the source material (go mod verify; license signal; nightly-fuzz confirmation; pre-push hook) is the
  natural grain — collapsing nightly-fuzz confirmation into `go mod verify`'s task would save one task
  ID at the cost of blurring two different evidence types and risk profiles under one task, which
  mandate §12 explicitly warns against ("split when... have materially different risks").

The alternative (folding nightly-fuzz confirmation into the `go mod verify` task, since both are
`ci.yml`-focused) was considered and rejected for the reasons above, not overlooked.

## Expected package or module changes

No Go package changes. Workflow-configuration changes: `.github/workflows/ci.yml`,
`.github/workflows/security-scan.yml`. Script change: `.githooks/pre-push`.

## Expected file changes where determinable

- `.github/workflows/ci.yml` — add a `go mod verify` step; no change expected for the nightly-fuzz
  schedule itself unless the confirmation task (T004) finds it is not actually correctly wired, in which
  case a fix becomes a deviation from this plan, recorded in `deviations.md`, not silently absorbed here.
- `.github/workflows/security-scan.yml` — add `license` to the `trivy` job's `scanners:` list (planned
  default choice; see "License signal decision").
- `.githooks/pre-push` — modify the `go test ./...` invocation or its surrounding shell logic so DB-test
  skip requires `WOWAPI_REQUIRE_DB` (or fails loudly if the DB is unavailable) — exact mechanism to be
  determined at implementation time (see "Unresolved questions").
- `.githooks/pre-commit` — explicitly **not** touched by this story (see `story.md` "Out of scope").

## Contracts and interfaces

None. No public Go interface, API, or contract is affected by this story.

## Data structures

None.

## APIs

None affected.

## Configuration changes

The pre-push hook fix may introduce or formalize `WOWAPI_REQUIRE_DB` as an expected local-environment
variable if it does not already exist as a convention elsewhere in the repository (to be confirmed at
implementation time — `requirement-inventory.md`'s FBL-07 note names `WOWAPI_REQUIRE_DB` directly,
suggesting it may already be an established convention used by other DB-gated tests; this plan does not
assume it is a brand-new variable without confirming first).

## Persistence changes

None.

## Migration strategy

Not applicable — no schema or data migration.

## Concurrency implications

None. CI-workflow and shell-script changes have no runtime concurrency implications.

## Error-handling strategy

`go mod verify`'s failure mode is the tool's own standard non-zero exit on checksum mismatch — no custom
error handling needed beyond letting the CI step fail normally. The license-scan step's failure mode
follows whichever mechanism (Trivy `exit-code`, or `go-licenses`'s own exit behavior) is chosen; per
`story.md`'s "Residual-risk expectations," this story enables the signal but does not commit to a
specific block-on-violation policy beyond what the chosen tool's default reporting behavior provides —
whether the license scan step is configured to fail the build on a real violation, or remains
informational like the existing `trivy` job's other scanners (`exit-code: "0"`), is a decision to make
explicit at implementation time, not left ambiguous in the merged workflow file. The pre-push hook's new
failure mode (loud failure on missing `WOWAPI_REQUIRE_DB` or unreachable DB) is the entire point of
`AC-W01-E01-S003-04` — it must produce a clear, actionable error message, not a bare non-zero exit with
no explanation.

## Security controls

`go mod verify` is itself a supply-chain-integrity security control. The license-scanning signal is a
license-compliance governance control. Neither introduces a new attack surface; both are read-only
inspection steps.

## Observability changes

None. This story does not add new metrics, logs, or dashboards beyond standard CI job output and the
nightly-fuzz confirmation's own audit-note evidence record.

## Testing strategy

- `go mod verify`: verified by direct execution in CI against the current, presumably-clean module
  cache — this is itself the "test." No separate unit test is applicable.
- License-scanning signal: verified by direct execution in CI, confirming the step runs and produces a
  license report or equivalent output (a real violation is not expected to be manufactured merely to
  test the fail path, per mandate §13's guidance against creating tests that don't validate meaningful
  behavior — a genuine violation, if any exists in the current dependency set, would be a real finding to
  triage, not a test fixture).
- Nightly-fuzz-schedule confirmation: verified by direct inspection of the workflow file's job graph and,
  where feasible, by observing an actual triggered run (manual `workflow_dispatch` if the workflow
  supports it, or by waiting for/observing the next scheduled 03:17 UTC run if a manual trigger is not
  available — to be determined at implementation time).
- Pre-push hook fix: verified by a fail-before/pass-after demonstration — run the hook (or its DB-test
  segment) without `WOWAPI_REQUIRE_DB` set and without a reachable DB before the fix (documents the
  current silent-skip/silent-pass behavior) and after the fix (documents the new loud-failure behavior),
  then again with a DB available and `WOWAPI_REQUIRE_DB` set to confirm the hook still passes normally
  when a DB is genuinely available.

## Regression strategy

`go mod verify` and the license-scan step, once added, are themselves ongoing regression guards (any
future module-cache corruption or newly-introduced non-compliant-license dependency would now be caught
in CI where it previously would not). The pre-push hook fix is a local-only regression guard (catches a
developer accidentally believing DB tests ran when they did not) and has no CI-side regression
implication.

## Compatibility strategy

`go mod verify` and the license-scan step are additive CI steps with no effect on build output or
runtime behavior — fully backward-compatible. The pre-push hook fix intentionally changes local
developer-facing behavior (see `story.md` "Compatibility considerations") — this is the story's purpose,
not an unintended break, and is called out for documentation at implementation time.

## Rollout strategy

Single PR/commit per the epic's stated pattern (each task's change is small and independently
mergeable); no phased rollout required for any of the four items — none is a runtime-behavior change
requiring gradual exposure.

## Rollback strategy

Each of the four changes can be reverted independently: remove the `go mod verify` step, remove `license`
from Trivy's `scanners:` list (or remove the `go-licenses` step, whichever was chosen), revert the
pre-push hook script change. No cross-dependency between the four means a revert of one does not require
reverting the others.

## Implementation sequence

The four tasks (T001-T004, see "Task breakdown" below) are independent of each other (disjoint files: two
touch `ci.yml`, one touches `security-scan.yml`, one touches `.githooks/pre-push` — even the two `ci.yml`
tasks, T001 and T004, touch different, non-overlapping sections of that file) and may be executed in any
order or in parallel. Step 1 ("Implementation strategy" above — the fresh re-read of all three files) must
occur before any of the four tasks' fixes begin, since it is the shared fail-first/confirm-first step all
four tasks depend on for accurate current-state grounding.

## Task breakdown

- **W01-E01-S003-T001** — `go mod verify` CI step addition.
- **W01-E01-S003-T002** — License-scanning signal enablement (Trivy `license` scanner, planned choice).
- **W01-E01-S003-T003** — Pre-push hook DB-silent-skip fix.
- **W01-E01-S003-T004** — Nightly fuzz-schedule confirmation (own task; see "Task breakdown grouping
  decision" above for the rationale).

## Expected artifacts

Updated `.github/workflows/ci.yml` (go mod verify step); updated `.github/workflows/security-scan.yml`
(license scanner); updated `.githooks/pre-push`; a nightly-fuzz-schedule confirmation/audit note (not a
code diff).

## Expected evidence

CI run log for `go mod verify`; CI run log for the license-scanning step; nightly-fuzz-schedule
confirmation evidence (workflow inspection plus an observed run where feasible); fail-before/pass-after
demonstration for the pre-push hook fix.

## Unresolved questions

- Exact mechanism for the pre-push hook fix: require `WOWAPI_REQUIRE_DB` and fail if unset, vs. actively
  probe DB reachability and fail if unreachable regardless of the env var, vs. some combination — to be
  determined at implementation time from how `WOWAPI_REQUIRE_DB` (or its equivalent) is already used
  elsewhere in the repository's DB-gated test infrastructure, if it already exists as a convention.
- Whether the license-scanning signal, once enabled, should be configured to actually fail CI on a
  detected violation or remain informational (matching the existing `trivy` job's `exit-code: "0"`
  posture for its other scanners) — not resolved by this plan; to be made an explicit, documented choice
  at implementation time rather than left ambiguous.
- Whether a manual trigger (`workflow_dispatch`) is available or addable to `ci.yml` to observe the
  nightly-fuzz job's behavior without waiting for the actual 03:17 UTC cron — to be determined at
  implementation time; if not available and not addable within this story's bounded scope, the
  confirmation record states this limitation explicitly (see `story.md` "Assumptions").
- Whether the exact `security-scan.yml` line numbers and `trivy`/`dependency-review` job structure cited
  in "Current-state assessment" (this story's own direct 2026-07-12 read) still hold at implementation
  time — to be re-confirmed, not assumed.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above are answered by a
fresh re-read of `ci.yml`, `security-scan.yml`, and `.githooks/pre-push` at story start, (b) the
license-signal choice (Trivy license scanner, as planned, or a revised choice with recorded rationale) is
confirmed or revised based on that fresh re-read, and (c) the owner and reviewer are assigned.
