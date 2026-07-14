---
id: W01-E01-S003
type: story
title: Close supply-chain and pre-push hook hygiene gaps
status: accepted
wave: W01
epic: W01-E01
owner: W01Lint
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-07
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W01-E01-S003-01
  - AC-W01-E01-S003-02
  - AC-W01-E01-S003-03
  - AC-W01-E01-S003-04
artifacts:
  - ART-W01-E01-S003-001
  - ART-W01-E01-S003-002
  - ART-W01-E01-S003-003
  - ART-W01-E01-S003-004
evidence:
  - EV-W01-E01-S003-001
  - EV-W01-E01-S003-002
  - EV-W01-E01-S003-003
  - EV-W01-E01-S003-004
decisions: []
risks: []
---

# W01-E01-S003 — Close supply-chain and pre-push hook hygiene gaps

## Story ID

W01-E01-S003

## Title

Close supply-chain and pre-push hook hygiene gaps

## Objective

Add a `go mod verify` step to `ci.yml`; enable a license-scanning signal (Trivy license scanner or
`go-licenses`) with the choice documented; confirm the nightly fuzz-schedule that already exists since
PR #24 is correctly wired in its seed-replay form, with the coverage-guided `-fuzz=` gap explicitly
recorded as out of scope; and fix the pre-push hook so it no longer silently skips DB-gated tests.

## Value to the framework

`go mod verify` is a zero-cost, high-value supply-chain check (it detects a corrupted or tampered
module cache against the recorded checksums in `go.sum`) that currently runs nowhere in CI — this is a
pure gap, not a degraded signal. A license-scanning signal closes a real blind spot: today, no CI job
actually inspects dependency licenses, since Trivy's own license scanner is off and
`dependency-review`'s `license-check: true` only ever fires on `pull_request` events. Confirming the
nightly fuzz schedule is correctly wired (rather than assuming PR #24 landed cleanly and never
rechecking) keeps the seed-corpus regression signal honest. Fixing the pre-push hook's silent DB-skip
converts a currently-misleading local signal ("pre-push passed" when DB tests never actually ran) into
an honest one, consistent with the repository's own skip-hygiene stance — a developer who has no local
DB should get a loud, actionable failure, not silent green.

## Problem statement

`requirement-inventory.md` row FBL-07 records: "Utilisation closure (gosec triage, go mod verify,
license signal, nightly fuzz, hook DB-skip)" — disposition `partial`, priority split P1/P2, target
`W01-E01-S002..S003`, with the explicit note: "Nightly ci schedule EXISTS since #24 (fuzz portion still
seed-replay only)." FBL-07 is a single requirement row spanning two stories: the gosec/errorlint/
exhaustive/forcetypeassert/usestdlibvars enablement-and-triage half is `W01-E01-S002` (a sibling story,
referenced here by ID only — this story does not depend on or duplicate its content). This story,
`W01-E01-S003`, is FBL-07's remainder half: the supply-chain and process-hygiene items that are CI- and
hook-configuration changes rather than static-analyzer triage.

Four distinct gaps make up this remainder, each independently confirmed against the current repository
state as of this story's writing (2026-07-12), and each to be re-confirmed fresh at actual
implementation time since the codebase moves between planning and execution:

1. **`go mod verify` is absent from `ci.yml` entirely.** Confirmed: no `go mod verify` invocation
   exists anywhere in `.github/workflows/ci.yml` as of this writing. This is a pure addition, not a
   fix to a misconfigured existing step.
2. **No license-scanning signal is actually enforced.** Confirmed as of this writing:
   `.github/workflows/security-scan.yml`'s `trivy` job (around line 57-75) configures
   `scanners: vuln,secret,misconfig` — the `license` scanner is not in that list. Separately,
   `dependency-review-action` (around line 77-93) sets `license-check: true`, but that job's own `if:`
   condition (`github.event_name == 'pull_request' && needs.guard.outputs.public == 'true'`) means it
   only ever runs on `pull_request` events — a fact of the GitHub Action's own design (dependency-review
   only supports PR-diff-based scanning), not a symptom of the repository's visibility. See "License
   signal decision" below for the important nuance the source material insists on getting right.
3. **The nightly fuzz schedule's existence needs confirmation, not re-implementation.** Confirmed as of
   this writing: `.github/workflows/ci.yml` has a `schedule:` trigger with `cron: "17 3 * * *"` (around
   line 42-46), and header comments (around lines 17, 25) describe a "test — test-unit (DB+S3) + fuzz
   seed corpus" job that runs "on main pushes and the nightly schedule." This story's job is to verify
   this wiring is real, actually nightly, and actually invokes fuzz targets in seed-replay mode — not to
   add real `-fuzz=` coverage-guided generation, which is out of scope (see "Out of scope").
4. **The pre-push hook silently skips DB-gated tests.** Confirmed as of this writing:
   `.githooks/pre-push` (around line 21-22) runs `go test ./...` with the hook's own comment stating
   "DB tests skip without a DSN" — i.e., if `WOWAPI_REQUIRE_DB` (or equivalent) is not set and no DB is
   reachable, DB-gated tests self-skip rather than failing, and the hook reports success regardless.
   `.githooks/pre-commit` is a separate, different hook and is explicitly out of this story's scope (see
   "Out of scope").

## Source requirements

FBL-07 (remainder half — go mod verify, license signal, nightly-fuzz confirmation, pre-push hook fix).
The gosec/errorlint/exhaustive/forcetypeassert/usestdlibvars enablement-and-triage half of FBL-07 is
`W01-E01-S002`, a sibling story under this same epic — not this story's scope.

## Current-state assessment

Confirmed by direct inspection of the repository at the time this story was written (2026-07-12); all
four items are to be re-confirmed fresh at implementation time since exact line numbers and job
conditions may have shifted:

- `go mod verify`: zero hits anywhere in `.github/workflows/ci.yml`.
- License scanning: Trivy's `scanners:` list in `security-scan.yml` does not include `license`;
  `dependency-review`'s `license-check: true` exists but is gated to `pull_request` events only, by the
  action's own design, and is further gated by `needs.guard.outputs.public == 'true'` in this
  workflow's `if:` condition.
- Nightly fuzz schedule: a `schedule:`/`cron: "17 3 * * *"` trigger exists in `ci.yml`, with comments
  describing a test job that includes "fuzz seed corpus" replay on the nightly run and on main pushes.
  Whether this is wired correctly end-to-end (schedule actually reaches the fuzz-seed-replay step, with
  no silent gating that prevents it from running) has not been independently re-verified as part of
  this planning pass — that confirmation is this story's actual task, not an assumed-true fact.
- Pre-push hook: `.githooks/pre-push` runs `go test ./...` and documents, in its own comment, that DB
  tests self-skip without a DSN — the hook does not currently require `WOWAPI_REQUIRE_DB` or otherwise
  fail loudly when DB tests cannot run.
- The repository has been **public since 2026-07-03**. This is a separate fact from
  `dependency-review`'s PR-only triggering behavior — the two must not be conflated. The
  `dependency-review` job's `if:` condition includes a `needs.guard.outputs.public == 'true'` check,
  which the now-public repository satisfies; the job's *dormancy* as a license gate is about it being a
  `pull_request`-event-only action by GitHub's own design, not about repository visibility blocking it.

## Desired state

`ci.yml` runs `go mod verify` as part of its normal build/test pipeline and fails the pipeline if it
fails. A license-scanning signal — either Trivy's `license` scanner added to the existing `trivy` job's
`scanners:` list, or a new `go-licenses` step — runs in CI and its choice and rationale are documented
in this story's `implementation.md` once implemented. The nightly fuzz schedule in `ci.yml` is confirmed
(by direct re-inspection at implementation time, plus a triggered/observed run if feasible) to exist,
run nightly, and correctly invoke fuzz targets in seed-corpus-replay mode — with the coverage-guided
`-fuzz=` gap recorded, not silently closed. `.githooks/pre-push` requires `WOWAPI_REQUIRE_DB` (or fails
loudly if a DB is unavailable) rather than silently allowing DB-gated tests to self-skip.

## Scope

- Add a `go mod verify` step to `.github/workflows/ci.yml`.
- Enable exactly one license-scanning signal — Trivy's `license` scanner or `go-licenses` — per the
  planned choice and rationale in "License signal decision" below, confirmed or revised at
  implementation time if the exact `security-scan.yml` state has changed.
- Confirm the nightly fuzz schedule in `ci.yml` exists, is genuinely nightly, and correctly invokes fuzz
  targets in seed-corpus-replay mode. Produce a confirmation/audit record as this task's evidence.
- Fix `.githooks/pre-push` so it no longer silently allows DB-gated tests to self-skip: require
  `WOWAPI_REQUIRE_DB` (or fail loudly if the DB is unavailable) rather than silently passing.

## Out of scope

- **`W01-E01-S002`'s gosec/errorlint/exhaustive/forcetypeassert/usestdlibvars enablement and triage** —
  the other half of FBL-07, a sibling story under this epic. Not implemented, duplicated, or referenced
  beyond citing it by ID.
- **REL-04 T8 / PERF-06 T3/T4's real `-fuzz=` coverage-guided generation wiring** — explicitly W07
  scope, shared ownership assigned to "PF-REL" per `premier-framework-implementation-plan.md`. This
  story confirms the nightly *schedule* exists and is correctly wired for seed-corpus replay; it does
  not add the `-fuzz=` flag or any coverage-guided fuzzing capability. This boundary is stated explicitly
  so it is neither silently dropped (someone assumes W07 already covers "confirm the schedule exists"
  and no one checks) nor silently duplicated (someone in this story accidentally implements the
  `-fuzz=` flag wiring that is actually W07's job).
- **`.githooks/pre-commit`** — a different hook entirely, currently shown as modified/uncommitted in
  this session's working tree, and explicitly out of this story's scope. This story touches
  `.githooks/pre-push` only.
- **The design rationale that hooks are a strict subset of full CI** (i.e., hooks intentionally run a
  lighter/faster check than the DB-backed `make ci-container` gate) — this is accepted, unchanged
  design and is not what this story fixes. Only the *silent* nature of the DB-test skip is fixed; the
  hook may still legitimately run without a DB available in some developer environments, but it must
  say so loudly rather than passing quietly.
- **Any exact-line-number claim in this document treated as authoritative without re-confirmation.**
  The line numbers cited in "Current-state assessment" reflect a direct read of the repository at
  story-writing time (2026-07-12) and are more precise than the approximate `security-scan.yml:71` and
  `security-scan.yml:80,93` citations in the underlying source material, but implementation must still
  re-read the actual file at its own start commit rather than trusting either citation blindly, since
  the file may have changed in the interim.

## Assumptions

- The choice between Trivy's license scanner and `go-licenses` (see "License signal decision") is a
  planned choice, not a locked-in one — it may be revisited at implementation time if a fresh read of
  `security-scan.yml`'s exact state changes the picture (e.g. if Trivy's license scanner turns out to
  have a materially different SPDX-detection quality than assumed here, or if `go-licenses` proves
  harder to wire into the existing job structure than expected).
- The pre-push hook's exact required change is assumed to be small (a one-line addition requiring
  `WOWAPI_REQUIRE_DB` or equivalent, or making the DB-test skip fail loudly) based on the hook's current
  ~35-line size and the narrow nature of the fix described in the source material. The precise mechanism
  (require an env var vs. probe DB reachability directly) is to be determined at implementation time.
- The nightly fuzz schedule's correctness (per "Desired state") is assumed to be verifiable by direct
  re-inspection of `ci.yml` plus, where feasible, observing an actual scheduled or manually-triggered
  run — not merely by re-reading the header comments describing intended behavior. If a triggered run is
  not practically feasible within this story's scope (e.g. requires waiting for the next 03:17 UTC
  cron), the confirmation record states this limitation explicitly rather than fabricating an observed
  run.

## Dependencies

None within W01-E01 — S001/S002/S003 target disjoint files (database config vs. CLI/app/auth/config/
workflow files vs. CI/hook configuration) and can proceed in any order or in parallel. Depends on W00's
exit gate at wave scope (baseline CI/lint state captured).

## Affected packages or components

`.github/workflows/ci.yml`; `.github/workflows/security-scan.yml`; `.githooks/pre-push`.

## Compatibility considerations

`go mod verify` is additive and non-breaking: it either passes (module cache matches `go.sum`, which it
should under normal operation) or fails loudly on a genuine integrity problem, which is the desired
behavior, not a regression. The license-scanning signal, however implemented, is additive and does not
change build output. The pre-push hook fix intentionally changes local developer-facing behavior — a
push that previously succeeded silently (DB tests skipped) may now fail loudly if `WOWAPI_REQUIRE_DB` is
required and unset; this is the story's explicit purpose, not an unintended compatibility break, but it
is a real behavior change for any developer relying on the current silent-skip convenience, so it is
worth documenting clearly at implementation time (e.g. in the hook's own inline comments and/or
`docs/user-guide/build-deploy.md`).

## Security considerations

`go mod verify` directly strengthens supply-chain integrity by catching a tampered or corrupted local
module cache before it can propagate into a build. A license-scanning signal is itself a supply-chain
governance control (detecting a dependency under an incompatible or unexpected license before it lands
on `main`). The pre-push hook fix has no direct security control effect but improves the "tests actually
ran" honesty of the local pre-push gate, which indirectly reduces the risk of a DB-dependent regression
reaching `main` undetected by a developer who believed their local push had exercised DB-backed tests.

## Performance considerations

`go mod verify` and a license scan both add wall-clock time to CI, though both are expected to be small
relative to the existing pipeline (module verification and dependency-license enumeration are typically
fast, sub-minute operations for a module of this size) — to be confirmed with an actual timed run at
implementation time, consistent with the CI-parallelization/wall-clock-tracking context established by
session-delta SD-01 (CI gate parallelized into 3 legs, per `requirement-inventory.md` §E). The pre-push
hook fix has no CI performance effect; if it causes the hook to fail loudly for developers without local
DB access, it may increase local iteration friction for those specific developers — an intentional
trade-off (honest failure over silent success), not a defect.

## Observability considerations

None beyond what already exists for CI job output. The nightly-fuzz-schedule confirmation task should
record what it actually observed (schedule trigger present, job graph reaching the fuzz-seed-replay
step) as evidence, which is itself a form of observability into the CI pipeline's own correctness, but
this story does not add new metrics, logs, or dashboards.

## Migration considerations

None. No schema, data, or configuration migration is involved — this story is CI-workflow and git-hook
configuration only.

## Documentation requirements

- Document the `go mod verify` step's addition in `implementation.md` once implemented (per the
  standard task/story implementation-record discipline; no separate user-facing doc is anticipated to be
  required, since this is an internal CI-pipeline change).
- Document the license-signal choice (Trivy license scanner vs. `go-licenses`) and its rationale, both in
  this story (`plan.md`, "License signal decision" below) before implementation and in
  `implementation.md` after implementation, including whether the choice made here was carried through
  unchanged or revised at implementation time.
- Document the nightly-fuzz-schedule confirmation outcome as an evidence record (an audit/confirmation
  note, not a code diff) referenced from `evidence/index.md`.
- Document the pre-push hook behavior change (no more silent DB-test skip) wherever local
  developer-workflow expectations are documented — likely `docs/user-guide/build-deploy.md` or
  equivalent, to be confirmed at implementation time — so developers are not surprised by a newly-loud
  failure.

## Acceptance criteria

- **AC-W01-E01-S003-01**: `.github/workflows/ci.yml` runs `go mod verify` as a distinct step, and that
  step's failure fails the overall CI job; evidenced by a run log showing the step executing and passing
  against a clean module cache.
- **AC-W01-E01-S003-02**: A license-scanning signal (Trivy `license` scanner or `go-licenses`, per the
  documented choice) is enabled and runs in CI; the choice and its rationale are recorded in
  `implementation.md`; evidenced by a CI run log showing the license-scan step executing and producing a
  license report or equivalent output.
- **AC-W01-E01-S003-03**: The nightly fuzz schedule in `ci.yml` is confirmed — by direct inspection of
  the workflow file at implementation-time HEAD, plus an observed scheduled or manually-triggered run
  where feasible — to exist, to be genuinely nightly (cron-triggered), and to correctly invoke fuzz
  targets in seed-corpus-replay mode; the confirmation record explicitly states the coverage-guided
  `-fuzz=` gap remains open and is W07 scope (REL-04 T8 / PERF-06 T3/T4), not silently closed by this
  story and not duplicated against W07's task list.
- **AC-W01-E01-S003-04**: `.githooks/pre-push` no longer allows DB-gated tests to silently self-skip:
  either `WOWAPI_REQUIRE_DB` (or equivalent) is required and the hook fails loudly with an actionable
  message when it is unset, or the hook otherwise fails loudly when the DB is unavailable rather than
  silently passing; evidenced by a fail-before/pass-after demonstration (hook silently passes without a
  DB before the fix; hook fails loudly and clearly without a DB, and passes with one, after the fix).
  `.githooks/pre-commit` is unmodified by this criterion's verification.

## Required artifacts

- Updated `.github/workflows/ci.yml` (`go mod verify` step; nightly-fuzz-schedule confirmation may not
  itself produce a file diff if the schedule is already correctly wired).
- Updated `.github/workflows/security-scan.yml` (license-scanning signal), if Trivy's scanner list is
  the chosen mechanism, or a new/updated workflow step if `go-licenses` is chosen instead.
- Updated `.githooks/pre-push`.
- A nightly-fuzz-schedule confirmation/audit note (not a code diff — see "Task breakdown" judgment call
  in `plan.md`).
See `artifacts/index.md`.

## Required evidence

- CI run log showing `go mod verify` executing and passing.
- CI run log showing the license-scanning step executing and producing output.
- Confirmation/audit evidence for the nightly fuzz schedule (workflow-file inspection plus an observed
  run where feasible).
- Fail-before/pass-after demonstration of the pre-push hook's DB-skip behavior change.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md` and
`plan.md` complete, acceptance criteria numbered and measurable, dependencies (none) recorded, owner/
reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, with the reviewer specifically checking that the nightly-fuzz scope boundary is
honestly stated (neither silently closed nor silently duplicated against W07) and that the license-
signal choice is documented with rationale rather than silently picked.

## Risks

No dedicated risk-register entry exists for this story at the time of writing (unlike S001's
RISK-W01-E01-002). The primary latent risk — that the exact `security-scan.yml` line numbers or gating
state cited in "Current-state assessment" have drifted between this story's writing and its
implementation — is addressed procedurally (re-confirm at implementation time, per "Out of scope") rather
than tracked as a separate risk-register entry, since it is a low-likelihood, low-impact drift (a few
line numbers, not a structural change) rather than a genuine implementation risk.

## Residual-risk expectations

Once all four acceptance criteria are verified, no residual risk is expected to remain open at
acceptance for the `go mod verify` and pre-push hook items (both are small, mechanical, low-risk
changes). The license-signal item carries a small residual-risk expectation: whichever of Trivy license
scanner or `go-licenses` is chosen, its output is a *signal*, not an enforcement gate, in this story's
scope — the story enables detection, it does not commit to blocking CI on a license violation, since
that policy decision (what license makes a build fail) is not itself specified in the source material and
is not invented here. This residual scope boundary should be carried into any later story that wants to
convert the signal into a hard gate.

## Plan

See `plan.md`.
