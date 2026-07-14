---
id: W00-E02-S001-T002
type: task
title: Lint baseline (25-analyzer, MATRIX CS-23 drift)
status: done
parent_story: W00-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W00-E02-S001-02
artifacts: []
evidence: []
---

# W00-E02-S001-T002 — Lint baseline (25-analyzer, MATRIX CS-23 drift)

## Task Definition

*Per mandate §8.6. Defines the task before work begins.*

### Task objective

Run `golangci-lint` with all 25 analyzers named in the closure-depth matrix's CS-23 spec
temporarily enabled via a throwaway config variant (explicitly not the committed `.golangci.yml`),
capture the fresh hit counts, and compare them analyzer-by-analyzer against the MATRIX CS-23
snapshot to confirm the prior counts still hold or flag drift.

### Parent story

W00-E02-S001 — Quality baselines.

### Owner

Unassigned.

### Status

`todo` (per `impl/governance/status-model.md` §7.3).

### Dependencies

None. This task can run independently of T001 and T003.

### Detailed work

1. Confirm the pinned `golangci-lint` version is installed: `golangci-lint version` should report
   v2.11.4 (matching `Makefile:16` and `.github/workflows/ci.yml:62`). Install via
   `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4` if not present.
2. Re-read the MATRIX CS-23 spec in its source document
   (`fable5-closure-depth-matrix-2026-07-11.md`, or its current archived/relocated location per
   session-delta SD-04 doc archival) and transcribe the verbatim list of all 25 analyzer names it
   queried. Do not rely on the category summary in `story.md`/`plan.md` ("zero-hit set,"
   "near-zero," "gosec," "adjudicated") as a substitute for the literal analyzer-name list — those
   are a paraphrase for planning purposes, not the authoritative source list.
3. Create a throwaway `golangci-lint` config variant (e.g. a local, uncommitted
   `.golangci.matrix-cs23.yml`): start from the committed `.golangci.yml` content, preserve its
   `exclusions` block verbatim (notably the `_test.go` errcheck/unparam exclusion at
   `.golangci.yml`'s `issues.exclusions.rules` — omitting this would inflate hit counts with
   test-file noise not present in the MATRIX-time measurement), and add all 25 analyzers from step 2
   to `linters.enable` (resolving during execution whether golangci-lint v2's schema supports an
   additive `enable` on top of `default: standard`, or requires an explicit full list — see
   `../plan.md` "Unresolved questions").
4. Run `golangci-lint run -c <throwaway-config> ./...` with the pinned v2.11.4 binary. A non-zero
   exit code is expected and acceptable (it reflects real lint findings) — it does not mean the
   task failed to execute.
5. Tabulate hit counts per analyzer from the output.
6. Compare each of the 25 analyzers' fresh count against its MATRIX CS-23 recorded count:
   - Zero-hit set (expect 0): sqlclosecheck, rowserrcheck, bodyclose, wastedassign, makezero,
     musttag, testifylint.
   - Near-zero set: noctx (expect 2 production hits), copyloopvar (expect 1 production hit),
     gocritic `exitAfterDefer` (expect 1 hit).
   - gosec (expect 38 hits total, with a named triage list: G704 JWKS taint, G120 unbounded form
     parse, G115 int-overflow set, G304 buildinfo file read — confirm the same finding sites recur,
     not merely the same aggregate count).
   - nilerr, exhaustive, errorlint (adjudicated at MATRIX time as deliberate, not gaps — confirm
     the same finding sites, not just an unchanged count).
   - Any of the 25 analyzers not explicitly named in the categories above (if MATRIX CS-23's full
     list is longer than the categories reproduced in `story.md`) must still be captured and
     compared; if MATRIX CS-23 did not record an expected count for a given analyzer, state that
     explicitly rather than inventing an expected value.
7. Flag any analyzer whose fresh count, or specific finding sites, differ from the MATRIX CS-23
   snapshot — explicitly, per-analyzer, not folded into a single aggregate "N issues found" number.
8. Register the result as an evidence record per `impl/governance/evidence-policy.md`, including the
   full drift table, and add an entry to `../evidence/index.md`.
9. Register the lint report (raw `golangci-lint run` output, and the throwaway config file content
   itself for reproducibility) as an artifact per `impl/governance/artifact-policy.md`, and add an
   entry to `../artifacts/index.md`.
10. Delete or leave untracked the throwaway config file — it must not be committed as part of this
    task's output (see `story.md` "Out of scope": permanent enablement is FBL-05's job).

### Expected files or components affected

None in the committed source tree. A throwaway, uncommitted config file is created and used for the
duration of this task only.

### Expected output

A lint-baseline evidence record with the full 25-analyzer hit-count table and an explicit
analyzer-by-analyzer drift comparison against MATRIX CS-23.

### Required artifacts

Lint report, including the 25-analyzer diff (raw `golangci-lint run` output + the throwaway config
used, for reproducibility).

### Required evidence

Lint-baseline evidence record (type: static-analysis report).

### Related acceptance criteria

AC-W00-E02-S001-02.

### Completion criteria

The lint-baseline evidence record exists in `../evidence/index.md`, states a count for all 25
named analyzers (not a subset), and states explicitly for each whether it matches or drifts from
the MATRIX CS-23 snapshot.

### Verification method

Per `../verification.md`'s AC-02 row: re-run (or review the run of) `golangci-lint run -c
<throwaway-config> ./...` with the pinned v2.11.4 binary, confirm the reported per-analyzer counts
match what is recorded in the evidence record, confirm the committed `.golangci.yml` was not
modified by this task (`git status`/`git diff` shows no change to it).

### Risks

- If the throwaway config diverges from the committed config's `exclusions` block (e.g. accidentally
  dropping the `_test.go` exclusion), hit counts would be inflated by test-file noise not comparable
  to MATRIX CS-23's own measurement. Mitigated by explicitly preserving `exclusions` verbatim (step
  3 above).
- If the verbatim 25-analyzer list is mistranscribed from the MATRIX source document, the comparison
  itself would be unreliable. Mitigated by requiring direct re-reading of the MATRIX document at
  execution time (step 2), not relying on the categorical paraphrase in `story.md`/`plan.md`.

### Rollback or recovery considerations

The throwaway config file is deleted or left untracked after the run — no rollback of committed
state is needed since none occurs.

## Implementation Record

*Per mandate §8.7. Not yet executed — no implementation claims are pre-populated.*

### What was actually implemented

Baseline capture executed: pinned golangci-lint v2.11.4 confirmed; MATRIX CS-23 re-read — it names 18 analyzers verbatim despite claiming 25 (flagged, see DEV-W00-E02-S001-001); throwaway config built from the committed `.golangci.yml` (settings/exclusions verbatim) + the 18 named analyzers (v2 schema accepts additive `enable` on `default: standard`, resolving plan unresolved question 3); full-tree run captured (991 issues) plus committed-config control (0 issues); per-analyzer drift table produced; throwaway config deleted from repo root (content preserved as artifact).

### Components changed

None.

### Files changed

None in the committed source tree. Evidence/artifact files written under this story directory only.

### Interfaces introduced or changed

None.

### Configuration changes

None to the committed `.golangci.yml` — confirmed by `git status` at task completion (file untouched).

### Schema or migration changes

None.

### Security changes

None — gosec re-measured, not re-adjudicated. Security-relevant drift flagged: aggregate 38 not reproducible (24 prod / 111 total); all four named triage classes recur at the named sites; G204/G301/G306 hits present but absent from the MATRIX triage list.

### Observability changes

None.

### Tests added or modified

None.

### Commits

None made by this task (verification-only; the conductor owns commits). Executed against `0a31186cada5c275a588c74081cf977adf346e61`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

MATRIX names only 18 of its claimed 25 analyzers — the comparison covers all recoverable names. Run-1 exhibited 11 nondeterministic staticcheck SA5011 hits under concurrent load (not reproducible in two cache-cleaned re-runs; raw output preserved).

### Follow-up items

Flagged drift = candidate new findings for FBL-05/FBL-07 disposition (not resolved here): exhaustive +2 prod sites (kernel/config/bind.go:326, schema.go:95); errorlint +2 prod sites (internal/tools/benchbudget/main.go:114,118 — post-#25 code); forcetypeassert +1 prod site (kernel/httpclient/client.go:71); gosec G204/G301/G306 un-triaged rule classes; noctx tool-behavior drift (named exec sites no longer reported by v2.11.4).

### Relationship to the approved plan

Followed `../plan.md` step 2 exactly, including the exclusions-preservation mitigation; one documented deviation (18-vs-25 analyzer list, DEV-W00-E02-S001-001).

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E02-S001-02 | Run `golangci-lint run -c <throwaway-config> ./...` (all 25 MATRIX CS-23 analyzers enabled, pinned v2.11.4); compare per-analyzer counts against MATRIX CS-23. | Local dev or CI container with `golangci-lint` v2.11.4 installed, full repository checked out. | Every one of the 25 analyzers has an explicit recorded count; drift (if any) is explicitly flagged, not silently absorbed. | Static-analysis report | unassigned |

### Actual result

991 issues under the throwaway config (exit 1, expected); 0 issues under the committed config. Zero-hit set: all 7 at zero (MATCH). Full per-analyzer drift table in EV-W00-E02-S001-002.

### Pass or fail

**Captured (pass as a capture task).** Every one of the 18 named analyzers has an explicit recorded count; all drift explicitly flagged per-analyzer.

### Evidence identifier

EV-W00-E02-S001-002.

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (main).

### Environment

Local dev workstation, macOS (Darwin 25.5.0) arm64, go1.26.5; real Postgres 16 via compose; concurrent sibling load present. golangci-lint v2.11.4 (pinned).

### Reviewer

Unassigned (conductor review gate pending).

### Findings

Drift flagged on: noctx (named prod sites unreported by v2.11.4 though code unchanged), exhaustive (+2 prod), errorlint (+2 prod), forcetypeassert (+1 prod), gosec (aggregate not reproducible; new un-triaged rule classes), wrapcheck/revive (≈50 not reproducible; noise-dominant verdict reinforced). Matches confirmed on: entire zero-hit set, copyloopvar, gocritic named site, nilerr exact, all named adjudication sites.

### Retest status

Not required — first capture, no failed run to retest.

### Final conclusion

AC-W00-E02-S001-02 satisfied; lint baseline with drift comparison registered; no drift silently absorbed.

## Deviations Record

*Per mandate §8.9. No deviations recorded yet.*

### Deviation ID

*Assign a stable deviation ID (`DEV-W00-E02-S001-T002-NNN`) if a deviation occurs.*

### Approved plan

*State what `../plan.md` said.*

### Actual implementation

*State what was actually implemented.*

### Reason

*State the reason for the deviation.*

### Impact

*State the impact of the deviation.*

### Risks

*State risks introduced by the deviation.*

### Approval

*State who approved the deviation and when.*

### Compensating controls

*State any compensating controls put in place.*

### Follow-up work

*State any follow-up work arising from the deviation.*
