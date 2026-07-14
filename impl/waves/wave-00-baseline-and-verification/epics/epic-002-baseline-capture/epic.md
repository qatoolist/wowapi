---
id: W00-E02
type: epic
title: Baseline capture
status: planned
wave: W00
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - D-01
  - D-02
  - D-03
  - D-04
  - D-05
  - D-06
  - D-07
  - D-08
  - D-09
  - SD-01
  - SD-02
  - SD-03
  - SD-04
stories:
  - W00-E02-S001
  - W00-E02-S002
  - W00-E02-S003
decisions: []
risks:
  - RISK-W00-003
  - RISK-W00-004
  - RISK-W00-005
---

# W00-E02 — Baseline capture

## Epic objective

Capture, at the current repository HEAD, the quantitative and structural baselines that every
later wave's acceptance criteria will be measured against — unit-test coverage, full-tree lint
state (including the 25-analyzer hit-count inventory named in MATRIX CS-23), benchmark-budget
state, CI wall-clock timing, the dependency/toolchain inventory — and formalize the nine
architecture decisions (D-01 through D-09) already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U as durable ADR records
this programme can cite by stable ID instead of re-deriving from prose each time a downstream
story needs them.

## Problem being solved

Two distinct gaps, both blocking safe downstream execution:

1. **No current-HEAD quantitative baseline exists in the `impl/` traceability structure.** The
   plan/review/matrix documents describe coverage, lint, and bench state as of the SHAs they were
   written against (`main @ 0a31186` per `impl/index.md`). Session-delta facts SD-01..SD-04
   (`impl/analysis/requirement-inventory.md` §E) — CI parallelization (#23), bench path-scoping
   (#24), sweep-bench recalibration (#25), doc archival (#22) — postdate those documents. Without a
   fresh, evidence-recorded capture, every later wave's "did this regress / improve" claim has
   nothing registered to compare against.
2. **D-01..D-09 are real decisions, but only exist as prose rows in REVIEW §F/§U.** Nine
   downstream epics (see `dependencies.md`) each need one of these decisions as a design input
   before their own stories can be planned in detail. Leaving them as review-document prose means
   every consuming story would have to re-read and re-interpret §F/§U instead of citing a stable
   `ADR-...` ID — and mandate §11.8 explicitly requires decisions not be "buried only in prose."

## Scope

- Capturing coverage %, full-tree lint state (25-analyzer hit counts), bench-budget state, and CI
  wall-clock per leg as registered evidence artifacts (S001).
- Capturing the go.mod dependency inventory and pinned tool versions, cross-checked against REVIEW
  §L's approved-dependency register (S002).
- Writing nine ADR files (D-01 through D-09) that formalize the already-made decisions from REVIEW
  §F/§U, and registering them in the decision register (S003).

## Out of scope

- Re-verifying the 8 EXECUTED finding-slices (SEC-02, PERF-01, PERF-06, DATA-08 W0, AR-04 T1,
  AR-05 T1/T2, AR-06 T1, REL-04 T1-T4) — that is W00-E01's scope.
- Enabling the 25 currently-unenabled `golangci-lint` analyzers in the committed `.golangci.yml` —
  this epic only captures the *current* state with all 25 temporarily enabled via a throwaway
  config variant, for comparison purposes. Permanent enablement is FBL-05's job (W01-E01-S001).
- Making any new architecture decision. D-01..D-09's content is drawn verbatim from REVIEW §F/§U;
  this epic's S003 formalizes an already-made decision into the programme's ADR structure, it does
  not re-litigate or extend it (mandate §18).
- Resolving DEC-Q1/Q9/Q10 (the 3 genuine human decisions) — those remain tracked as open,
  human-blocked items per `impl/analysis/requirement-inventory.md` §B; they are not part of the
  D-01..D-09 set this epic ADR-ifies.
- Adopting `cenkalti/backoff/v5` or any other approved-but-unused dependency into actual code
  (FBL-04, W04-E02-S003) — S002 only inventories and cross-checks the approved-dependency register.

## Source requirements

D-01, D-02, D-03, D-04, D-05, D-06, D-07, D-08, D-09 (nine ratified architecture decisions —
`impl/analysis/requirement-inventory.md` §B, "D-01..D-09 — Nine ratified architecture decisions");
SD-01, SD-02, SD-03, SD-04 (session-delta facts — §E) fold into the S001 quality baseline capture.

## Architectural context

This epic touches no production code path. It reads and records the state of:

- The test/coverage toolchain (`go test -cover`, real-DB integration coverage per project history
  of a 92%-measured/90%-floor baseline — to be reconfirmed fresh, not assumed).
- The lint toolchain (`golangci-lint` v2.11.4 pinned at `Makefile:16` and `ci.yml:36`; the
  committed `.golangci.yml` enabling "standard + 4" analyzers; MATRIX CS-23's 25-analyzer
  inventory of what ships unenabled).
- The benchmark-budget toolchain (`internal/tools/benchbudget`, `make bench-budget`,
  `bench-budgets.txt`, post-#25 recalibration per MATRIX CS-16).
- The CI pipeline (`.github/workflows/ci.yml`, 3-leg parallelized per SD-01, path-scoped bench per
  SD-02).
- `go.mod`/`go.sum` and the approved/rejected dependency registers in REVIEW §L/§M.
- The nine decision points in REVIEW §F (rows 2-8 map to D-01..D-07) and §U (D-08, D-09), each of
  which is a design input another epic's stories will consume as a fixed premise rather than an
  open question.

## Included stories

- **W00-E02-S001 — quality-baselines**: capture coverage %, full-tree lint state (25-analyzer hit
  counts vs the MATRIX CS-23 snapshot), bench-budget state (post-#25, 43 entries), and CI
  wall-clock per leg as registered evidence.
- **W00-E02-S002 — dependency-and-toolchain-inventory**: capture the go.mod direct/indirect
  dependency list and pinned tool versions, cross-checked against REVIEW §L's approved register
  with zero unexplained drift.
- **W00-E02-S003 — adr-ification**: write nine ADR files formalizing D-01 through D-09, register
  them in the decision register.

## Dependencies

- None upstream — W00-E02 has no dependency on W00-E01 for its own execution (S001/S002/S003 can
  run independently of the finding-slice re-verification work), though `wave.md`'s rationale notes
  a **recommended, non-blocking sequencing**: S003 logically follows S001 (see this epic's own
  `dependencies.md` for the internal-sequencing rationale, not a hard blocker).
- Downstream: nine epics across W01, W03, W04, W05, W06 each depend on one of the D-01..D-09 ADRs
  produced by S003 — full table in `dependencies.md` (mirrors `wave.md`-level
  `../../dependencies.md`, scoped to this epic's stories as the producing unit).

## Risks

See `risks.md` for the full epic-level register. Summary: RISK-W00-003 (bench-budget baseline
captured against stale, pre-#25 budgets), RISK-W00-004 (ADR-ification silently adding design
content beyond what D-01..D-09's source states), RISK-W00-005 (CI/coverage baseline captured
without correctly accounting for SD-01/SD-02's pipeline changes).

## Required decisions

This epic does not require any decision to be made *before* it can proceed — S003 is itself the
mechanism that turns nine already-made decisions (REVIEW §F/§U) into this programme's durable ADR
records. There is no blocking "required decision" gating W00-E02's own start; the nine ADRs are an
epic *output*, not an epic *input*. The only decisions this epic's stories consume as fixed
premises are the D-01..D-09 content itself, which is not this epic's to re-decide (mandate §18).

## Epic acceptance criteria

See `acceptance.md` for the full numbered list. Summary:

- AC-W00-E02-01: coverage %, lint state, bench-budget state, and CI wall-clock are captured as
  registered evidence, commit-pinned (traces to S001).
- AC-W00-E02-02: dependency/toolchain inventory is captured and cross-checked against REVIEW §L
  with zero unexplained drift (traces to S002).
- AC-W00-E02-03: nine ADR files exist for D-01..D-09, each stating recommendation, safe default
  (where the source states one), and owner, registered in the decision register (traces to S003).
- AC-W00-E02-04: no story in this epic is accepted solely because its tasks are marked complete —
  each has passed independent review per mandate §14 (traces to S001/S002/S003 collectively).

## Closure conditions

- All three stories (S001, S002, S003) reach `accepted` per the story lifecycle
  (`impl/governance/lifecycle.md`), each with a complete `closure.md`.
- `closure-report.md` for this epic is complete: acceptance-criteria completion, story completion,
  artifact/evidence completeness, unresolved findings, accepted risks, reviewer conclusion,
  acceptance authority, closure date, final status (mandate §8.10 shape, applied at epic scope).
- No evidence gap flagged as unresolved; no ADR flagged as having silently added design content
  beyond its REVIEW §F/§U source (RISK-W00-004 resolved or explicitly accepted).
- The nine downstream epic dependencies listed in `dependencies.md` are confirmed unblocked (each
  ADR is in `ratified` status, not merely `proposed`).
