---
id: W00-E01
type: epic
title: Executed-slice verification
status: planned
wave: W00
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - SEC-02
  - AR-04
  - AR-06
  - PERF-01
  - PERF-06
  - DATA-08
  - REL-04
depends_on: []
stories:
  - W00-E01-S001
  - W00-E01-S002
  - W00-E01-S003
decisions: []
risks:
  - RISK-W00-001
  - RISK-W00-002
  - RISK-W00-003
---

# W00-E01 — Executed-slice verification

## Epic objective

Re-run, at the current repository HEAD, the exact named test files/commands that
`premier-framework-implementation-plan.md` (PLAN) and `fable5-final-architecture-review-2026-07-11.md`
(REVIEW) claim were already executed for eight finding-slices spanning security, performance, data
integrity, and CI infrastructure — and register mandate-§10-conformant evidence for each. This epic
converts prose "EXECUTED" claims cited against a past commit SHA into re-proven, currently-pinned
evidence inside `impl/`'s own traceability structure.

## Problem being solved

`impl/analysis/requirement-inventory.md` §A records eight PLAN findings with disposition `partial` or
`implemented-needs-verification` (INV) for their executed portions: SEC-02 (T1-T3), AR-04 (T1), AR-06
(T1), PERF-01, PERF-06 (T1), DATA-08 (W0 slice), and REL-04 (T1-T4). Each is described in PLAN/REVIEW
as already implemented and tested, but the evidence backing that claim lives only as prose test
descriptions and `git log` commit-SHA citations in documents that predate `impl/`'s own evidence
register (mandate §10). None of these eight slices has an evidence record that identifies the
execution command, the tested revision, the environment, and a reviewer in the format this programme
requires. Until that gap is closed, no downstream wave can safely treat these slices as a proven
"before" state — W01's AR-04/AR-06 remainder tasks, W03's SEC-02 ratification work, and W07's
REL-04/PERF-06 fuzz work all build directly on top of these slices being genuinely intact today, not
merely intact at the SHA the review documents cite.

## Scope

- Re-running the exact test files/commands PLAN §5/§8/§9 and REVIEW §D name for each of the eight
  executed slices, at the epic's own closing commit SHA.
- Registering mandate-§10 evidence records for each re-run (command, commit SHA, environment, tool
  versions, date/time, result, reviewer).
- Confirming the three MATRIX verify-outcome rows most directly entangled with this epic's slices
  (CS-03 config fail-closed, CS-19 i18n freeze, CS-24 SSRF dial-time guard) still hold, via evidence
  pointer, as part of S003's scope (grouped there because REL-04's CI-pipeline-state confirmation
  task is the natural home for a "does the current repository state still match what MATRIX claims"
  check — see S003 for the precise boundary).
- Confirming session-delta facts SD-01/SD-02/SD-03 (CI parallelization, bench recalibration) are
  correctly reflected in the current repository state where they bear directly on a slice being
  re-verified (SD-03 for PERF-01/S002; SD-01/SD-02 for REL-04/S003).

## Out of scope

- Any code change, fix, or remediation — this epic is verification-only. If a regression is found,
  the story that surfaced it stays open (does not move to `accepted`) and a new remediation task is
  opened under the owning future-wave story per the finding's `requirement-inventory.md` target
  (e.g. a SEC-02 regression would need a new task, not a silent fix folded into this epic).
- The *unexecuted* remainder of each partial finding (SEC-02 T4/T5 ratification design, AR-04 T2-T5,
  AR-06 T2/T3 lint+audit, PERF-06 T3/T4 fuzz, DATA-08 W6 hash widening, REL-04 T5-T8) — these are
  `planned` disposition in `requirement-inventory.md`, targeted at their own future-wave stories
  (W03-E05-S001, W05-E03-S002, W05-E04-S001, W07-E02-S002, W04-E04-S001..S002, W07-E02-S002
  respectively), not this epic.
- AR-05 (T1/T2 "Composition/doc drift removal") — `requirement-inventory.md` §A canonically targets
  AR-05 to **W06-E04-S002**, not W00-E01. `wave.md`'s prose and `epics/index.md`'s summary line both
  state W00-E01 covers "AR-05 (T1/T2)," which conflicts with the canonical allocation and with this
  epic's own detailed task-issuer content (S001 was specified to cover only SEC-02/AR-04/AR-06). This
  epic does **not** include an AR-05 story or task — see `risks.md` and the epic-creation report for
  this conflict, which is flagged rather than silently resolved per mandate §18.
- Quantitative baseline capture (coverage %, lint hit-counts, bench-budget snapshot, CI wall-clock,
  dependency inventory) and D-01..D-09 ADR-ification — these are W00-E02's scope, not W00-E01's.

## Source requirements

SEC-02, AR-04, AR-06, PERF-01, PERF-06, DATA-08, REL-04 (each at the `partial`/`INV` disposition and
target listed in `impl/analysis/requirement-inventory.md` §A). Session-delta facts SD-01, SD-02,
SD-03 (§E) bear on S002 and S003's scope as described above.

## Architectural context

The eight slices span four unrelated framework layers, which is why they are grouped into three
stories by *layer* rather than kept as one undifferentiated epic-wide task list:

- **Workflow/runtime and boot/authz composition** (SEC-02, AR-04, AR-06) — `kernel/workflow/`,
  `app/boot.go`, `kernel/kernel.go`, `kernel/authz/`. All three findings concern fail-closed behavior
  and construction-time correctness of the kernel's composition root.
- **Rate limiting and the build's own quality gates** (PERF-01, PERF-06) — `kernel/httpx/ratelimit.go`
  and `internal/tools/benchbudget/`. Both concern the token-bucket sweep implementation and the CI
  mechanism that enforces its benchmark budget, respectively.
- **Attachment/notification durability and CI test infrastructure** (DATA-08 W0, REL-04 T1-T4) —
  `kernel/attachment/`, `kernel/notify/`, `.github/workflows/ci.yml`, `deployments/compose.yaml`,
  `Makefile`. Grouped together because REL-04's S3/TOTP/CI-pipeline wiring is the test infrastructure
  DATA-08's DB-gated tests (and, more broadly, this whole epic's re-runs) depend on being correctly
  configured.

## Included stories

- **W00-E01-S001 — verify-workflow-and-boot-slices**: re-verify SEC-02 (T1-T3), AR-04 (T1), and AR-06
  (T1) at current HEAD.
- **W00-E01-S002 — verify-performance-slices**: re-verify PERF-01 and PERF-06 (T1) at current HEAD,
  including confirming the #25 sweep-bench recalibration (SD-03) is reflected in `bench-budgets.txt`.
- **W00-E01-S003 — verify-data-and-integration-slices**: re-verify DATA-08 (W0-T1/W0-T2) and REL-04
  (T1-T4) at current HEAD, including confirming SD-01/SD-02's parallel-CI pipeline state is reflected
  in `.github/workflows/ci.yml`.

## Dependencies

- No dependency on any other epic or wave (W00-E01 is entry-point work; see `wave.md` "Dependencies").
- Internal to the epic: S001, S002, and S003 target disjoint packages/files and disjoint test commands
  (`kernel/workflow`+`app`+`kernel` vs. `kernel/httpx`+`internal/tools/benchbudget` vs.
  `kernel/attachment`+`kernel/notify`+CI config) and can execute in any order or in parallel — see
  `dependencies.md` for the full statement.
- All three stories share one external dependency: a working `docker compose`/Postgres/MinIO test
  environment (`make ci-container`), required for S003's DB-gated and S3-gated tests specifically, and
  assumed available for the others' unit/race runs.

## Risks

RISK-W00-001 (a claimed-executed slice fails to re-verify — regression since the reviewed SHA),
RISK-W00-002 (test infrastructure unavailable, producing a false-negative regression), RISK-W00-003
(bench-budget baseline captured against stale, pre-#25 values). Full register: `risks.md` (epic-level)
and `../../risks.md` (wave-level, source of these three IDs).

## Required decisions

None. This epic ratifies no new architecture decision — D-01..D-09 ADR-ification is W00-E02-S003's
scope. The one judgment call this epic's creation surfaced (the AR-05 scope conflict, see "Out of
scope" above) is a documentation conflict to be resolved by the acceptance authority, not an
architecture decision this epic can make unilaterally.

## Epic acceptance criteria

- **AC-W00-E01-01**: All nine tasks across S001/S002/S003 (3 per story) have been executed at the
  epic's closing commit SHA, each producing a `pass` result with a registered evidence ID, OR a
  `failed`-status evidence record plus an open follow-up task if a regression was found (no task may
  be silently marked done with an unresolved failure).
- **AC-W00-E01-02**: Every story's `verification.md` post-execution record is complete (actual result,
  evidence ID, execution date, commit SHA, environment, reviewer) for every acceptance criterion in
  that story's `story.md`.
- **AC-W00-E01-03**: The AR-05 scope conflict identified above has been explicitly resolved by the
  acceptance authority — RESOLVED 2026-07-12 by the programme author: AR-05 executed T1/T2 re-verification is IN scope for W00-E01-S001 (AC-04 + task added; the 8-slice wave exit gate requires it); AR-05 T3-T5 remain at W06-E04. before this
  epic may move to `accepted`.
- **AC-W00-E01-04**: No evidence record in any of the three stories' `evidence/index.md` is missing a
  commit SHA, execution command, or result field (per `evidence-policy.md`).

## Closure conditions

All three stories reach `accepted` (each satisfying its own `closure.md`); AC-W00-E01-01 through
AC-W00-E01-04 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; no evidence gap or unresolved regression remains open against any of
the nine tasks.
