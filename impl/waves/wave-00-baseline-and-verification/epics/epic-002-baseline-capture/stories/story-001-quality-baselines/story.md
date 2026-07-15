---
id: W00-E02-S001
type: story
title: Quality baselines
status: accepted
wave: W00
epic: W00-E02
owner: W00E02S001 (worker)
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - PERF-01
  - PERF-06
  - SD-01
  - SD-02
  - SD-03
  - FBL-05
  - FBL-07
  - REL-04
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W00-E02-S001-01
  - AC-W00-E02-S001-02
  - AC-W00-E02-S001-03
  - AC-W00-E02-S001-04
artifacts:
  - ART-W00-E02-S001-001
  - ART-W00-E02-S001-002
  - ART-W00-E02-S001-003
  - ART-W00-E02-S001-004
evidence:
  - EV-W00-E02-S001-001
  - EV-W00-E02-S001-002
  - EV-W00-E02-S001-003
  - EV-W00-E02-S001-004
decisions: []
risks:
  - RISK-W00-003
  - RISK-W00-005
---

# W00-E02-S001 — Quality baselines

## Story ID

W00-E02-S001

## Title

Quality baselines

## Objective

Capture, as registered evidence at the current repository HEAD, four quantitative baselines —
unit-test coverage percentage, full-tree lint state (including a fresh count for all 25 analyzers
named in the closure-depth matrix's CS-23 spec), benchmark-budget state post-#25 recalibration, and
CI wall-clock timing per leg — so every later wave's "did this regress or improve" claim has a
commit-pinned reference point registered in this programme's own traceability structure, not merely
cited from the prior review/matrix documents.

## Value to the framework

The framework's later waves (W01 lint-analyzer enablement, W07 performance-budget hardening, and
every wave in between that touches tested code) need an honest "before" state to measure against.
Without a fresh, evidence-recorded baseline, a later wave's claim of "coverage improved" or "no
lint regression" has nothing authoritative to cite — the prior review/matrix documents describe
state as of their own SHA, and three session-delta facts (CI parallelization, bench path-scoping,
sweep-bench recalibration) postdate those documents entirely. This story is infrastructure for
every subsequent wave's acceptance-criteria measurement, not a feature in itself — it is the kind
of baseline-capture work mandate §15 anticipates for a dedicated Wave 00.

## Problem statement

Two gaps, both blocking safe downstream measurement:

1. No coverage/lint/bench/CI baseline for the current repository state exists as a registered
   evidence record in `impl/`. Project history records a prior baseline of measured coverage at
   approximately 92% against a 90% floor, but that number was established at an earlier commit and
   must not be assumed still true — it must be reconfirmed fresh at this story's own execution
   commit.
2. The full 25-analyzer `golangci-lint` hit-count inventory that MATRIX CS-23 queried was captured
   at the MATRIX document's own SHA, with the committed `.golangci.yml` enabling only "standard + 4"
   analyzers (errcheck, govet, ineffassign, staticcheck, unused, plus depguard, misspell, unconvert,
   unparam). All 25 analyzers CS-23 queried ship unenabled in the committed config. A fresh run with
   all 25 temporarily enabled is needed to confirm the MATRIX-time counts still hold before FBL-05
   (a separate epic, W01-E01-S001) permanently enables any of them.

## Source requirements

- **PERF-01** — token-bucket sweep fix; disposition `implemented-needs-verification` per
  `impl/analysis/requirement-inventory.md` §A. Session delta SD-03 (sweep-bench O(n²)+empty-map fix,
  budgets recalibrated, #25) changed PERF-01's evidence basis — this story captures the
  post-recalibration bench-budget baseline that PERF-01's continued-validity claim rests on.
- **PERF-06** — fail-closed performance gates; disposition INV per §A. T1 executed; this story's
  bench-budget baseline capture is the quantitative half of re-confirming that gate's current state
  (T3/T4 fuzz scope is a separate item, REL-04 T8, out of this story's scope).
- **SD-01** — CI gate parallelized into 3 legs, toolbox image GHA-cached, docs-only diff skip,
  landed via PR #23 (`impl/analysis/requirement-inventory.md` §E). This story's CI-wall-clock
  baseline must reflect this shape, not a pre-#23 serial pipeline.
- **SD-02** — bench path-scoped on PRs, nightly schedule, `merge_group` support, landed via PR #24
  (§E). Folds into this story's bench-budget and CI-timing baselines.
- **SD-03** — sweep-bench O(n²)+empty-map fix, budgets recalibrated (§E). This story's bench-budget
  baseline must be captured against the post-#25 budgets, not the pre-#25 values.
- **FBL-05** (reference only, not implemented by this story) — enable zero-cost leak linters
  permanently in `.golangci.yml`; disposition `planned`, target `W01-E01-S001` per §B. This story
  captures the *current* state with all 25 analyzers temporarily enabled via a throwaway config
  variant for comparison purposes only — it does not edit the committed `.golangci.yml`. That
  permanent enablement remains FBL-05's job in a different epic.
- **FBL-07** (reference only) — utilisation closure (gosec triage, go mod verify, license signal,
  nightly fuzz, hook DB-skip); disposition `partial` per §B, target `W01-E01-S002..S003`. The
  gosec/nilerr/exhaustive/errorlint triage adjudications this story's lint baseline re-confirms are
  the same adjudications FBL-07 tracks to closure; this story does not perform that closure work,
  only re-measures the current hit counts against the MATRIX CS-23 snapshot.
- **REL-04** (context) — truthful integration coverage; disposition `partial` per §A, T1-T4
  executed. This story's coverage baseline is measured against the same real-DB integration
  coverage posture REL-04 T1-T4 established (`WOWAPI_REQUIRE_DB=1`, real Postgres, not a mocked
  subset) — REL-04's own re-verification is W00-E01-S003's scope, not this story's.

## Current-state assessment

Confirmed from direct inspection of the repository at this story's planning time:

- `Makefile:16` pins `GOLANGCI_VERSION ?= v2.11.4`; `.github/workflows/ci.yml:62` pins the same
  version in CI (`GOLANGCI_VERSION: "v2.11.4"`) — both confirmed by direct read.
- The committed `.golangci.yml` sets `linters.default: standard` (errcheck, govet, ineffassign,
  staticcheck, unused) plus `enable: [depguard, misspell, unconvert, unparam]` — "standard + 4",
  confirmed by direct read.
- `.github/workflows/ci.yml` currently defines: a `changes` job (docs-only / bench-relevant
  classification), `workflow-lint` (actionlint), `unit` (no DB — fmt/vet/lint/tidy/boundaries/
  test-unit/build), a `gate` job matrixed over `[test, race]` legs (container, DB+S3 required,
  conditioned on `needs.changes.outputs.code == 'true'`), a `gate-bench` job (path-scoped on PRs via
  `needs.changes.outputs.bench`, unconditional on main push/nightly), `reference-smoke`, and
  `coverage` (profile + floor, `WOWAPI_REQUIRE_DB=1`). This matches SD-01 (3-leg parallelized gate,
  docs-only skip) and SD-02 (bench path-scoping, nightly schedule via `cron: "17 3 * * *"`,
  `merge_group` support) — both confirmed present in the file as currently committed.
- `Makefile:240` sets `COVERAGE_FLOOR ?= 90.0`; `make coverage-check` (Makefile:241-246) runs
  coverage against the real DB (`WOWAPI_REQUIRE_DB=1`) and fails below the floor. This confirms a
  90% floor is currently enforced in the committed Makefile.
- `Makefile:208-227` defines `BENCH_PKGS` (8 kernel packages) and `bench-budget`, which pipes
  `go test -bench` output through `internal/tools/benchbudget` against `bench-budgets.txt`.

Not yet confirmed, and explicitly not to be assumed by this story's plan or tasks:

- The actual current coverage percentage. Project history records a prior baseline of approximately
  92% measured coverage against the 90% floor — this is prior history, not this story's own
  measurement, and **must be measured fresh at this story's execution commit** via
  `go test -cover`/`make coverage-check` before being cited as current fact.
- The actual current 25-analyzer hit counts. MATRIX CS-23 recorded specific counts at its own SHA
  (zero-hit set: sqlclosecheck, rowserrcheck, bodyclose, wastedassign, makezero, musttag,
  testifylint; near-zero: noctx 2 production hits, copyloopvar 1 production hit, gocritic
  `exitAfterDefer` 1 hit; gosec 38 hits with a named triage list — G704 JWKS taint, G120 unbounded
  form parse, G115 int-overflow set, G304 buildinfo file read; nilerr/exhaustive/errorlint hits
  adjudicated as deliberate, not gaps). These are the MATRIX-time snapshot to compare against, not
  an assumed-current fact.
- The exact `bench-budgets.txt` entry count and values at this story's execution commit. Per
  `impl/waves/wave-00-baseline-and-verification/risks.md` RISK-W00-003, the expected post-#25 count
  is 43 entries — this must be confirmed by inspection, not assumed.
- The actual CI wall-clock per leg. No prior wall-clock timing has been captured against the
  current 3-leg parallelized shape in this programme's evidence structure.

## Desired state

Four evidence records exist under this story's `evidence/` structure (populated on first real
content per Adaptation 2, `impl/governance/naming-conventions.md`), each commit-pinned, each citing
the exact command executed, tool versions, environment, date, and result:

1. A coverage-baseline evidence record stating the actual measured coverage percentage at the
   execution commit, measured against the real Postgres test DB.
2. A lint-baseline evidence record stating the fresh 25-analyzer hit-count inventory (via a
   throwaway config variant enabling all 25), compared analyzer-by-analyzer against the MATRIX
   CS-23 snapshot, with any drift explicitly flagged.
3. A bench-budget-baseline evidence record confirming the post-#25 recalibrated budget count and
   values.
4. A CI-wall-clock evidence record stating the current per-leg timing of `.github/workflows/ci.yml`
   as currently shaped (SD-01/SD-02 reflected).

All four are registered in `evidence/index.md` and referenced by this story's `verification.md`
against the four acceptance criteria below.

## Scope

- Measuring and recording current unit-test coverage percentage against the real DB.
- Running `golangci-lint run` with all 25 MATRIX CS-23 analyzers temporarily enabled via a
  throwaway config variant (not the committed `.golangci.yml`), and recording the fresh hit counts
  compared against the MATRIX CS-23 snapshot.
- Running `make bench-budget` and confirming the post-#25 budgeted-entry count and values.
- Reading and recording the current `.github/workflows/ci.yml` shape and its per-leg wall-clock
  timing.
- Registering all four as evidence records with commit SHA, command, environment, tool versions,
  date, and result.

## Out of scope

- Permanently enabling any of the 25 analyzers in the committed `.golangci.yml` — that is FBL-05's
  job, tracked at `W01-E01-S001` in a different epic. This story's throwaway config variant is
  explicitly temporary and is not committed as the project's lint configuration.
- Fixing any lint finding, coverage gap, or bench-budget violation this baseline capture surfaces.
  If a fresh finding contradicts the MATRIX CS-23 snapshot (drift), this story's job is to flag it,
  not to resolve it — resolution is tracked as a new finding requiring its own disposition, per
  `impl/governance/artifact-policy.md`/`evidence-policy.md` drift-handling norms.
- Re-verifying PERF-01/PERF-06's own executed-slice test files/commands — that re-verification is
  W00-E01-S002's scope (a sibling epic). This story only captures the quantitative bench-budget
  baseline that PERF-01/PERF-06's continued validity is measured against.
- Re-verifying REL-04 T1-T4 (S3/TOTP wiring) — that is W00-E01-S003's scope. This story only
  measures coverage against the same real-DB posture REL-04 established.
- Coverage/lint/bench measurement for `wowsociety` or any downstream product repository — this
  story is scoped to the `wowapi` framework kernel repository only (mandate §2.3 framework-first
  scope).

## Assumptions

- The `docker compose` / Postgres test infrastructure referenced by `make ci-container` /
  `TEST_DSN` is available in the execution environment at the time this story is actually executed;
  if unavailable, the coverage-baseline task cannot produce a real-DB-measured result and this must
  be recorded as a blocker, not worked around with a mocked/partial measurement.
- `golangci-lint` v2.11.4 (the pinned version) is installable in the execution environment; the
  25-analyzer baseline must use this exact pinned version, not a newer or older one, so the drift
  comparison against MATRIX CS-23 is apples-to-apples.
- The prior ~92%/90%-floor coverage figure cited in project history remains a reasonable
  expectation of the current state's order of magnitude, but is explicitly not asserted as this
  story's own finding until re-measured.

## Dependencies

- **None (blocking)** — this story has no `depends_on` entry. Per `epic.md`'s "Dependencies"
  section, S001 can execute independently of W00-E01 (executed-slice re-verification) and of
  W00-E02-S002/S003.
- **Non-blocking internal sequencing recommendation** — per the epic's `dependencies.md`, S003
  (ADR-ification) is recommended to run *after* this story so its authors work from a
  freshly-confirmed baseline snapshot; this is not a hard dependency and does not appear in this
  story's `depends_on`.

## Affected packages or components

This story reads and measures; it does not modify production code. Components read:

- `Makefile` (`coverage`, `coverage-check`, `bench-budget`, `lint` targets and their pinned
  versions).
- `.golangci.yml` (read only — a separate, uncommitted throwaway variant is used for the 25-analyzer
  run).
- `.github/workflows/ci.yml` (read only, for CI shape and per-leg timing).
- `bench-budgets.txt` and `internal/tools/benchbudget` (read/run only).
- The full `./...` package tree, for coverage and lint measurement purposes only.

## Compatibility considerations

Not applicable — this story produces no code or configuration change to the committed repository
state. The throwaway `golangci-lint` config variant used for the 25-analyzer run is not committed;
it exists only for the duration of the measurement task.

## Security considerations

The gosec 38-hit named triage list (G704 JWKS taint, G120 unbounded form parse, G115 int-overflow
set, G304 buildinfo file read) is re-measured, not re-adjudicated, by this story. If the fresh count
differs from MATRIX CS-23's 38, that drift must be flagged explicitly in the evidence record —
silently absorbing a security-relevant analyzer's count change into an aggregate number would
understate a potential new finding.

## Performance considerations

This story's bench-budget baseline is the reference point W07's performance-budget stories will
measure improvement or regression against. Per the epic's RISK-W00-003, the baseline must be
captured against confirmed **post-#25** budgets (expected 43 entries), not stale pre-recalibration
values — this story's task explicitly confirms the entry count before treating the capture as
authoritative.

## Observability considerations

Not applicable — no logging, metrics, or tracing changes. The CI-timing evidence record is itself a
form of operational observability data about the pipeline, not a change to the framework's runtime
observability surface.

## Migration considerations

Not applicable — no schema, data, or configuration migration involved.

## Documentation requirements

None beyond the evidence records and this story's own planning documents. This story does not
update `docs/` — it produces `impl/`-scoped planning and evidence artifacts only.

## Acceptance criteria

- **AC-W00-E02-S001-01**: Coverage report captured and registered as evidence, citing the exact
  commit SHA and the exact `go test -cover` / `make coverage-check` command used, measured against
  the real Postgres test DB (`WOWAPI_REQUIRE_DB=1`), stating the actual measured coverage
  percentage as a fresh fact (not the prior ~92% figure asserted without re-measurement).
- **AC-W00-E02-S001-02**: Full-tree lint state captured with all 25 MATRIX CS-23 analyzers
  temporarily enabled via a throwaway `golangci-lint` config variant (pinned v2.11.4, committed
  `.golangci.yml` left unmodified), registered as evidence, with an analyzer-by-analyzer comparison
  against the MATRIX CS-23 snapshot (zero-hit set, near-zero set, gosec 38-hit triage list,
  nilerr/exhaustive/errorlint adjudications) and any drift explicitly flagged in the evidence
  record rather than silently absorbed.
- **AC-W00-E02-S001-03**: Bench-budget state captured via `make bench-budget`, registered as
  evidence, confirming the post-#25 recalibrated entry count (expected 43) and citing the exact
  command, commit SHA, and environment (real DB required).
- **AC-W00-E02-S001-04**: CI wall-clock per leg captured from the current
  `.github/workflows/ci.yml` shape (3-leg parallelized `gate` matrix, path-scoped `gate-bench`,
  `unit`, `workflow-lint`, `reference-smoke`, `coverage` jobs), registered as evidence, explicitly
  noting the SD-01 (parallelization, #23) and SD-02 (bench path-scoping, #24) session-delta facts
  this baseline reflects.

## Required artifacts

- Coverage report (raw `coverage.out` / `coverage.html` generation command and output summary;
  authoritative path per the no-duplication rule in `impl/governance/artifact-policy.md`).
- Lint report, including the 25-analyzer diff against the MATRIX CS-23 snapshot.
- Bench-budget snapshot (`bench-budgets.txt` state confirmation plus the fresh `make bench-budget`
  run output).
- CI timing log (per-leg wall-clock observations against the current `ci.yml` shape).

See `artifacts/index.md` for the full registered list (status "not yet produced" at story creation
time, per Adaptation 2).

## Required evidence

- Coverage-baseline evidence record (proves AC-01).
- Lint-baseline evidence record, including the analyzer-by-analyzer drift comparison (proves
  AC-02).
- Bench-budget-baseline evidence record (proves AC-03).
- CI-wall-clock evidence record (proves AC-04).

See `evidence/index.md` for the full registered list (status "not yet produced" at story creation
time, per Adaptation 2).

## Definition of ready

This story satisfies `impl/governance/definition-of-ready.md`'s story checklist: it is specific
(four named baselines, not an "improve quality" theme), bounded (scope/out-of-scope stated above),
implementable (exact commands identified in `plan.md`), independently reviewable and verifiable
(each AC has its own evidence record, none depends on another story's completion), traceable
(`source_requirements` above), has measurable AC (four numbered ACs, each producing a pass/fail or
quantitative artifact), states dependencies (`depends_on: []`, confirmed above), records assumptions
explicitly (see "Assumptions"), has a drafted `plan.md`, and states required artifacts/evidence
(above). It is not yet `ready` — that transition requires reviewer confirmation per
`impl/governance/lifecycle.md`.

## Definition of done

This story will satisfy `impl/governance/definition-of-done.md` when: all four tasks (T001-T003;
T003 covers both bench-budget and CI-timing) are `done`; all four AC have a `pass` verification
entry with a registered evidence ID; `artifacts/index.md` and `evidence/index.md` list every
produced item with the required fields; `deviations.md` states "no deviations" or lists every
actual deviation from `plan.md`; `closure.md` is complete; and the independent-review checklist
(mandate §14, reproduced in `definition-of-done.md`) has passed clean — including confirming no
drift finding from AC-02 was silently absorbed without being flagged.

## Risks

- **RISK-W00-003** (shared with epic/wave register) — bench-budget baseline captured against stale
  pre-#25 budgets if the recalibration is not correctly reflected at the execution commit. Mitigated
  by explicitly confirming the `bench-budgets.txt` entry count (expected 43) before treating the
  capture as authoritative.
- **RISK-W00-005** (shared with epic/wave register) — CI wall-clock/coverage baseline captured
  without correctly accounting for SD-01/SD-02's pipeline changes, misdescribing the current CI
  shape. Mitigated by reading the current `.github/workflows/ci.yml` directly (not from memory or
  the prior review documents) at execution time.
- **Story-specific**: the 25-analyzer throwaway config run could itself introduce noise if the
  throwaway config diverges from the committed config's `exclusions`/`settings` blocks in an
  unintended way (e.g. accidentally dropping the `_test.go` errcheck/unparam exclusion, which would
  inflate hit counts with test-file noise not present in MATRIX CS-23's own measurement). Mitigation
  is addressed in `plan.md`'s task breakdown for T002.

## Residual-risk expectations

Some residual risk remains even after this story's acceptance: a baseline captured at one commit
can drift the moment the next commit lands (this is inherent to any point-in-time baseline, not a
gap in this story's execution). This residual risk is tracked implicitly by every downstream story
that cites this baseline — each such story is expected to note the baseline's commit SHA and confirm
it is still the relevant comparison point, or re-capture if materially stale, per the revision-
pinning rule in `impl/governance/evidence-policy.md`.
