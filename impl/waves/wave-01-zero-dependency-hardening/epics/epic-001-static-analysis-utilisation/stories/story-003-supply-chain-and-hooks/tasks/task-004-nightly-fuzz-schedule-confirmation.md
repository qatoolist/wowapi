---
id: W01-E01-S003-T004
type: task
title: Nightly fuzz-schedule confirmation
status: done
parent_story: W01-E01-S003
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S003-03
artifacts:
  - ART-W01-E01-S003-004
evidence:
  - EV-W01-E01-S003-003
---

# W01-E01-S003-T004 — Nightly fuzz-schedule confirmation

## Task Definition

### Task objective

Confirm — by direct inspection of `.github/workflows/ci.yml`'s job graph, and by observing an actual
scheduled or manually-triggered run where feasible — that the nightly schedule already known to exist
since PR #24 (session delta SD-02) genuinely fires on a nightly cadence and genuinely reaches and
executes the fuzz-seed-corpus-replay step. This task does **not** implement real `-fuzz=` coverage-
guided generation; that gap is explicitly out of scope and belongs to REL-04 T8 / PERF-06 T3/T4 (W07,
shared ownership "PF-REL").

### Parent story

W01-E01-S003 — Close supply-chain and pre-push hook hygiene gaps.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001/T002/T003. Shares `ci.yml` with T001 but touches a disjoint section of the
file (the `schedule:`/fuzz-replay job graph vs. T001's new `go mod verify` step) and produces a
different artifact type (a confirmation/audit note vs. a code diff).

### Detailed work

1. Re-read `.github/workflows/ci.yml` fresh, at this task's actual start commit, focusing on the
   `schedule:`/`cron:` trigger (observed at 2026-07-12 as `cron: "17 3 * * *"`) and the job(s) it
   reaches, including the header comments describing "test — test-unit (DB+S3) + fuzz seed corpus" and
   "on main pushes and the nightly schedule" (observed around lines 17, 25, 42-46, 130, 239 as of
   2026-07-12 — re-confirm exact structure, do not assume unchanged).
2. Trace the job graph from the `schedule:` trigger through to the step(s) that actually invoke fuzz
   targets, confirming the invocation is in seed-corpus-replay mode (i.e., no `-fuzz=` flag present) —
   consistent with FBL-07's own disposition note ("fuzz portion still seed-replay only").
3. Where feasible, observe an actual run under the nightly trigger: either wait for/inspect the next
   scheduled 03:17 UTC run's logs, or use a manual trigger (`workflow_dispatch`, if the workflow
   supports it or can be added within this task's bounded scope) to force an on-demand run that
   exercises the same job path. If neither is practically feasible within this task's execution window,
   record that limitation explicitly in the confirmation note rather than fabricating an observed run.
4. Produce a confirmation/audit note (ART-W01-E01-S003-004) recording: what was inspected, what was
   observed (including any limitation from step 3), and an explicit restatement that the coverage-
   guided `-fuzz=` gap remains open and is W07 scope (REL-04 T8 / PERF-06 T3/T4) — not silently closed
   by this task and not silently duplicated (this task must not itself add a `-fuzz=` flag anywhere).
5. If the confirmation finds the nightly schedule is *not* actually correctly wired (e.g. the schedule
   trigger exists but does not actually reach the fuzz-replay step, contrary to the header comments),
   record this as a deviation from the expected state in `deviations.md` and determine, in coordination
   with the story owner, whether a minimal fix falls within this task's bounded scope or must be
   escalated as a separate follow-up item — do not silently expand this task into a larger fix without
   recording that expansion.

### Expected files or components affected

None expected to change if the nightly schedule is confirmed correctly wired as-is (this is primarily a
confirmation activity, not an implementation activity) — `.github/workflows/ci.yml` only if step 5's
contingency applies.

### Expected output

A confirmation/audit note stating the nightly schedule's verified state and explicitly restating the
`-fuzz=` scope boundary.

### Required artifacts

ART-W01-E01-S003-004 (nightly fuzz-schedule confirmation note).

### Required evidence

EV-W01-E01-S003-003 (CI execution record + confirmation/audit note).

### Related acceptance criteria

AC-W01-E01-S003-03.

### Completion criteria

The confirmation note is produced, stating the nightly schedule was inspected and (where feasible)
observed to run, explicitly reaffirming the `-fuzz=` coverage-guided-generation gap as W07 scope.

### Verification method

Direct workflow-file inspection plus, where feasible, an observed scheduled or manually-triggered run;
logged output and the confirmation note retained as evidence per `evidence/index.md`.

### Risks

The primary risk is scope-boundary risk, not implementation risk: under-verifying (accepting the
existing header comments' claims about the nightly schedule at face value without actually tracing the
job graph or observing a run) would produce a false-positive confirmation; over-verifying (drifting into
implementing the `-fuzz=` flag itself) would silently duplicate W07's scope. Both failure modes are
addressed explicitly in "Detailed work" steps 2-4.

### Rollback or recovery considerations

Not applicable in the code-rollback sense (this task, absent step 5's contingency, produces no code
diff). If step 5's contingency applies and a fix is made, that fix follows the same rollback approach as
any other small `ci.yml` change (revert the specific fix, recorded as a deviation).

## Implementation Record

Implemented 2026-07-13 by W01Lint.

### What was actually implemented

No code change (wiring correct as found). Produced the confirmation/audit note
`artifacts/nightly-fuzz-confirmation.md`: (a) cron `17 3 * * *` present (`ci.yml:42-46`);
(b) schedule event → `changes` job fail-safe `code=true` → `gate` test leg; (c) the leg runs
`make ci-container-test` → `go test ./kernel/filtering/ ./kernel/pagination/ -run "^Fuzz" -count=1`
— seed-corpus replay (in-code `f.Add` seeds), no `-fuzz=`; (d) OBSERVED scheduled runs on
consecutive nights (29182363356 on 2026-07-12, 29229288699 on 2026-07-13, both success) with the
seed-replay invocation visible in the gate-test job log. The `-fuzz=` coverage-guided gap is
explicitly restated as W07 scope (REL-04 T8 / PERF-06 T3/T4) — neither closed nor duplicated.

### Files changed

None (audit note artifact only).

### Commits

Conductor owns commits; delivered as a working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`.

### Pull requests

None (conductor owns wave integration).

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

## Verification Record

| AC | Actual result | Pass/fail | Evidence |
|---|---|---|---|
| AC-W01-E01-S003-03 | Schedule exists, genuinely nightly, reaches and executes seed replay — proven by file-chain inspection plus an observed scheduled run's job log; GH schedule delay (~3h after cron mark) noted as expected platform behavior | pass | EV-W01-E01-S003-003 (`evidence/logs/nightly-fuzz-observed-run.log`, `artifacts/nightly-fuzz-confirmation.md`) |

### Retest status

Not required.

### Final conclusion

AC satisfied; confirmation grounded in an observed run, not header comments.

## Deviations Record

None.
