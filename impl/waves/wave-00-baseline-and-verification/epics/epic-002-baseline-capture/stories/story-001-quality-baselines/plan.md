---
id: PLAN-W00-E02-S001
type: story-plan
parent_story: W00-E02-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W00-E02-S001 — Quality baselines

Per mandate §8.5: "Do not invent precise code changes where the repository does not yet provide
enough information. Clearly distinguish confirmed facts, planned changes, and implementation
assumptions." This story performs no code change — its "implementation" is baseline-**capture**
work: run commands, record output, register evidence. The distinctions below are drawn between
facts already confirmed by direct repository inspection (at story-planning time), and facts that
must be freshly measured when this story is actually executed.

## Proposed architecture

Not applicable in the code-architecture sense — this story produces no new module, interface, or
component. Its "architecture" is a measurement procedure: four independent capture steps (coverage,
lint, bench-budget, CI-timing), each producing one evidence record, all four registered under this
story's `evidence/index.md` and cross-referenced from `verification.md`.

## Implementation strategy

Grouped into three tasks (not four, and not one) — see "Task breakdown" below for the grouping
rationale. Each task:

1. Runs the exact command(s) named in `story.md` "Required evidence" against the story's execution
   commit.
2. Captures raw output (redirected to a file or otherwise preserved, per `artifact-policy.md`'s
   no-duplication rule — large raw output is registered by path/command, not pasted wholesale into
   the story tree).
3. Summarizes the result into an evidence record per `evidence-policy.md`'s required-field list
   (evidence ID, type, story/task, AC proven, execution command, commit SHA, branch/tag,
   environment, tool versions, date/time, result, file/URI, reviewer).
4. Where the fresh result is compared against a prior snapshot (lint's MATRIX CS-23 comparison,
   bench-budget's post-#25 entry-count confirmation), the comparison and any drift is stated
   explicitly in the evidence record — never silently folded into a single pass/fail without detail.

## Expected package or module changes

None. This story is read-only against the framework's source tree.

## Expected file changes where determinable

None in the repository outside `impl/`. Within `impl/`, this story's own tree
(`story.md`, `plan.md`, `implementation.md`, `verification.md`, `deviations.md`, `closure.md`,
`tasks/`, `artifacts/`, `evidence/`) is created now (planning) and populated with evidence-record
content later (execution) — no other file in the repository is touched.

A throwaway `golangci-lint` config variant is created for T002's 25-analyzer run (e.g. a scratch
file such as `.golangci.matrix-cs23.yml`, not committed to the repository, deleted or left untracked
after the run). This is explicitly **not** a change to the committed `.golangci.yml` — see
`story.md` "Out of scope."

## Contracts and interfaces

Not applicable.

## Data structures

Not applicable.

## APIs

Not applicable.

## Configuration changes

None to the committed repository. T002 uses an uncommitted, throwaway `golangci-lint` config
variant — see "Expected file changes" above.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

If any capture command fails to run (e.g. Postgres unavailable for the coverage/bench-budget
commands), the task does not fabricate a result. Per `story.md` "Assumptions," this must be recorded
as a blocker on the affected task, and the evidence record (if partially produced) is marked
incomplete rather than presented as a pass. This mirrors mandate §10's failed-evidence preservation
rule: a failed capture attempt is itself recorded, not silently retried until it disappears.

## Security controls

Not applicable to this story's own execution. The gosec-related findings this story's lint baseline
surfaces are re-measured only, not re-adjudicated — see `story.md` "Security considerations."

## Observability changes

None.

## Testing strategy

Not applicable in the sense of "tests written by this story" — this story does not add or modify
test code. Its own "test" is the correctness of the capture commands and the accuracy of the
comparison against the MATRIX CS-23 snapshot, verified per `verification.md`'s planned procedure.

## Regression strategy

Not applicable — no code change, so no regression surface. The baseline this story produces is
itself the reference point later stories use to detect regression.

## Compatibility strategy

Not applicable.

## Rollout strategy

Not applicable — no rollout. Evidence records, once registered, are simply present in
`evidence/index.md`; there is no deployment step.

## Rollback strategy

Not applicable to the capture work itself. If an evidence record is later found to be inaccurate
(e.g. the wrong config variant was used for T002), the correction is a new evidence record marked
`superseded` over the old one, per `evidence-policy.md`'s failed/superseded/retested/resolved/
accepted-exception vocabulary — the old record is never deleted or silently edited.

## Implementation sequence

1. **T001 — coverage baseline**: confirm test infrastructure is up (`make up` / `TEST_DSN`
   reachable), run `make coverage-check` (or the equivalent `go test -coverprofile` +
   `go tool cover -func` sequence it wraps), capture the `total:` coverage percentage from
   `go tool cover -func=coverage.out` output, register as evidence.
2. **T002 — lint baseline (25-analyzer)**: create the throwaway `golangci-lint` config variant that
   starts from the committed `.golangci.yml` and additionally enables all 25 analyzers named in
   MATRIX CS-23 (sqlclosecheck, rowserrcheck, bodyclose, wastedassign, makezero, musttag,
   testifylint, noctx, copyloopvar, gocritic, gosec, nilerr, exhaustive, errorlint, and the
   remainder of CS-23's named set — the exact 25-item list must be transcribed verbatim from the
   MATRIX document by whoever executes this task, not re-derived from memory), preserving the
   committed config's `exclusions` block (notably the `_test.go` errcheck/unparam exclusion) so the
   comparison against MATRIX CS-23 is apples-to-apples and not inflated by test-file noise. Run
   `golangci-lint run -c <throwaway-config> ./...` with the pinned v2.11.4 binary, capture the
   per-analyzer hit counts, compare against the MATRIX CS-23 snapshot (zero-hit set, near-zero set,
   gosec 38-hit named triage list, nilerr/exhaustive/errorlint adjudications), flag any drift,
   register as evidence.
3. **T003 — bench-budget and CI-timing baseline**: run `make bench-budget`, confirm the
   `bench-budgets.txt` entry count matches the expected post-#25 value (43), capture the run output;
   separately, read the current `.github/workflows/ci.yml` and record its job/leg structure and
   (where obtainable — e.g. from the most recent actual workflow run history, if accessible in the
   execution environment, otherwise noted as "structure confirmed, wall-clock timing pending a live
   run observation") the per-leg wall-clock. Register as evidence.

## Task breakdown

- **W00-E02-S001-T001** — coverage baseline.
- **W00-E02-S001-T002** — lint baseline (25-analyzer, MATRIX CS-23 drift comparison).
- **W00-E02-S001-T003** — bench-budget and CI wall-clock baseline (grouped together; see rationale
  below).

**Grouping rationale (T003)**: bench-budget confirmation and CI-timing capture are grouped into one
task rather than split into two, because both are "read the current CI/bench configuration and
record its current numbers" — a single coherent unit of work sharing the same nature (inspect
committed config + run/observe, record numbers, no adjudication or drift-comparison step of the
depth T002's 25-analyzer comparison requires). Splitting them into separate tasks would not add
independent tracking value per mandate §12 ("avoid excessive fragmentation into trivial tasks that
provide no tracking value") — they have the same owner profile, the same risk profile (RISK-W00-003
and RISK-W00-005 respectively, both low-severity), and neither blocks the other. T001 (coverage) and
T002 (lint) are kept separate from each other and from T003 because each has materially different
risk (security-relevant gosec drift in T002 vs. a coverage-floor number in T001) and a materially
deeper comparison step (T002's full analyzer-by-analyzer MATRIX diff), which mandate §12 identifies
as grounds for decomposition ("have materially different risks").

## Expected artifacts

Coverage report, lint report (with 25-analyzer diff), bench-budget snapshot, CI timing log — see
`artifacts/index.md`.

## Expected evidence

Coverage-baseline, lint-baseline (with drift comparison), bench-budget-baseline, CI-wall-clock
evidence records — see `evidence/index.md`.

## Unresolved questions

- Whether the execution environment used to actually run this story's tasks has live access to a
  recent GitHub Actions run for `.github/workflows/ci.yml` to observe real per-leg wall-clock
  timing, or whether the wall-clock figure must instead be estimated/observed from a fresh local
  `make ci-container-test` / `make ci-container-race` / `make ci-container-bench` run (which
  approximates but is not identical to the hosted-runner timing, since GHA cache warm/cold state and
  runner hardware differ from local execution). This must be resolved and stated explicitly in the
  T003 evidence record — not silently assumed.
- The exact verbatim 25-item analyzer list from MATRIX CS-23 is referenced by category (zero-hit,
  near-zero, gosec, adjudicated) in this story's planning documents, but the story's execution must
  transcribe the literal 25 analyzer names from the MATRIX document itself before building the
  throwaway config — this plan does not reproduce that transcription because doing so here without
  re-reading the MATRIX document at execution time risks a stale or mistyped copy diverging from the
  authoritative source.
- Whether golangci-lint v2's config schema requires the throwaway variant to fully restate
  `linters.enable` (committed 4 + 21 more) or supports an additive `enable` list on top of `default:
  standard` without an explicit `disable` — this is a T002 implementation-time detail to resolve
  against the pinned v2.11.4 documentation, not a planning-time assumption.

## Approval conditions

This plan is considered approved and ready for implementation once: the story moves from `planned`
to `ready` per `impl/governance/definition-of-ready.md` (reviewer confirmation), and the three
unresolved questions above are either resolved or explicitly carried into the relevant task's
"Detailed work" section as an open item the task owner must resolve during execution (not silently
dropped).
