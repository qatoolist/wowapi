---
id: IMPL-W00-E02-S001
type: implementation-record
parent_story: W00-E02-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W00-E02-S001

Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (main). Per-task detail
in `tasks/task-001..003`; this record aggregates.

## What was actually implemented

Four baseline captures, registered as evidence EV-W00-E02-S001-001..004:

1. **Coverage** (T001): `make coverage-check` against real Postgres — **92.3% vs 90.0% floor, PASS**.
2. **Lint, 25-analyzer** (T002): throwaway golangci-lint v2.11.4 config (committed config +
   the 18 analyzers MATRIX CS-23 names verbatim — the "25" headcount is unsubstantiated in the
   source, flagged as DEV-W00-E02-S001-001) — **991 issues**, full per-analyzer drift table vs
   MATRIX; committed-config control run: **0 issues**. Zero-hit set confirmed; drift explicitly
   flagged on noctx, exhaustive, errorlint, forcetypeassert, gosec, wrapcheck, revive.
3. **Bench-budget** (T003a): `bench-budgets.txt` = **43 entries** (post-#25 confirmed, no drift);
   `make bench-budget` — **43/43 OK, exit 0**.
4. **CI wall-clock** (T003b): hosted GH Actions run 29229288699 (headSha = execution commit) —
   total **3m24s**; per-leg timing recorded; SD-01/SD-02 shape confirmed by direct `ci.yml` read.

## Components changed

None — read/measure only, as planned.

## Files changed

None in the committed source tree (confirmed: `git status` shows the committed `.golangci.yml`,
Makefile, workflows, and all Go files untouched). Writes confined to this story's `impl/` tree.
Generated build artifacts `coverage.out`/`coverage.html` at repo root (untracked/ignored).

## Interfaces introduced or changed

None.

## Configuration changes

None to the committed repository. The throwaway `.golangci.matrix-cs23.yml` existed at repo root
only for the duration of T002's runs, was never committed, and was deleted afterward; its content
is preserved at `artifacts/static-analysis/golangci.matrix-cs23.throwaway.yml`.

## Schema or migration changes

None.

## Security changes

None — gosec re-measured, not re-adjudicated. Security-relevant drift flagged in
EV-W00-E02-S001-002 (per story.md §Security considerations, not silently absorbed).

## Observability changes

None.

## Tests added or modified

None.

## Commits

None made by this story's execution (conductor owns commits). All evidence pinned to
`0a31186cada5c275a588c74081cf977adf346e61`.

## Pull requests

None.

## Implementation dates

2026-07-13 (single session).

## Technical debt introduced

None.

## Known limitations

- Point-in-time baseline; drifts as HEAD moves (inherent; see story.md residual-risk note).
- Concurrent sibling load present on the capture machine — noted in every evidence record's
  environment field; count-based results unaffected, local bench ns/op read as load-tolerant
  snapshot (budget gate passed), CI timing sourced from hosted runners (unaffected).
- MATRIX CS-23's "25 analyzers" reduced to the 18 recoverable names (DEV-W00-E02-S001-001).

## Follow-up items

Candidate new findings from T002's drift comparison, for FBL-05/FBL-07 disposition (flagged, not
resolved here — per story.md "Out of scope"): exhaustive +2 prod sites (`kernel/config/bind.go:326`,
`kernel/config/schema.go:95`); errorlint +2 prod sites (`internal/tools/benchbudget/main.go:114,118`);
forcetypeassert +1 prod site (`kernel/httpclient/client.go:71`); gosec G204/G301/G306 un-triaged
rule classes; noctx tool-behavior drift (MATRIX-named exec sites unreported by v2.11.4 though the
code is unchanged).

## Relationship to the approved plan

Execution followed `plan.md`'s implementation sequence (T001 → T002 → T003, bench run last with a
serialized quiet window) with one recorded deviation (DEV-W00-E02-S001-001, 18-vs-25 analyzer
list). Both `plan.md` unresolved questions that were resolvable at execution were resolved and
stated: hosted run history used for CI timing (not local approximation); golangci-lint v2 schema
accepts additive `enable` on `default: standard`.
