---
id: W01-E01-S003-T003
type: task
title: Pre-push hook DB-silent-skip fix
status: done
parent_story: W01-E01-S003
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S003-04
artifacts:
  - ART-W01-E01-S003-003
evidence:
  - EV-W01-E01-S003-004
---

# W01-E01-S003-T003 — Pre-push hook DB-silent-skip fix

## Task Definition

### Task objective

Fix `.githooks/pre-push` so that DB-gated tests can no longer silently self-skip when
`WOWAPI_REQUIRE_DB` is unset (or the DB is otherwise unavailable): the hook must fail loudly and
actionably instead of reporting success while the DB-gated tests were never actually exercised.
`.githooks/pre-commit` is a different hook and is explicitly out of this task's scope.

### Parent story

W01-E01-S003 — Close supply-chain and pre-push hook hygiene gaps.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001/T002/T004 (disjoint files: this task touches `.githooks/pre-push` only).

### Detailed work

1. Re-read `.githooks/pre-push` fresh, at this task's actual start commit, to confirm its current
   structure (per `story.md`'s "Current-state assessment," the hook was observed at 2026-07-12 to run
   `go test ./...` around line 21-22 with its own comment noting "DB tests skip without a DSN" — re-
   confirm, do not assume unchanged).
2. Determine whether `WOWAPI_REQUIRE_DB` (or an equivalent variable) is already an established
   convention used elsewhere in the repository's DB-gated test infrastructure (per `plan.md`'s
   "Configuration changes" / "Unresolved questions"). If it is, reuse that exact convention; if not,
   confirm the correct variable/mechanism from how DB-gated tests actually decide to skip themselves
   today (i.e., what condition their own `t.Skip(...)` calls check).
3. Modify the pre-push hook so that, when `WOWAPI_REQUIRE_DB` is unset or the DB is otherwise
   unreachable, the hook fails loudly with a clear, actionable message (e.g. explaining that DB-gated
   tests were skipped and how to run them, or how to explicitly opt out) rather than silently reporting
   success. The exact mechanism — requiring the env var outright, or probing DB reachability directly —
   is determined at this step from what is found in step 2.
4. Preserve the existing, accepted design rationale that the pre-push hook is intentionally a lighter,
   faster subset of the full DB-backed CI gate (`make ci-container`) — this task does not add a full
   local DB-backed test run to every push; it only removes the *silent* nature of the current skip.
5. Do not modify `.githooks/pre-commit` — confirm by inspection that this task's diff touches only
   `.githooks/pre-push`.

### Expected files or components affected

`.githooks/pre-push`.

### Expected output

An updated `.githooks/pre-push` that fails loudly and actionably when DB-gated tests cannot run,
instead of silently passing.

### Required artifacts

ART-W01-E01-S003-003 (updated `.githooks/pre-push`).

### Required evidence

EV-W01-E01-S003-004 (fail-before/pass-after execution log).

### Related acceptance criteria

AC-W01-E01-S003-04.

### Completion criteria

The hook fails loudly with an actionable message when run without `WOWAPI_REQUIRE_DB` set (or without a
reachable DB, per whichever mechanism step 3 lands on); the hook still passes normally when a DB is
genuinely available; `.githooks/pre-commit` remains untouched.

### Verification method

Direct execution of the hook script in both the "no DB" and "DB available" conditions, logged output
retained as evidence per `evidence/index.md`; a diff confirming `.githooks/pre-commit` is unmodified.

### Risks

Low implementation risk (a small, mechanical shell-script change), but a real local-developer-experience
change: developers without local DB access who relied on the current silent-skip convenience will now
see a hook failure on every push unless they explicitly opt in or out — an intentional trade-off (see
`story.md` "Compatibility considerations"), not a defect, but worth flagging for documentation so it
does not read as a surprising regression.

### Rollback or recovery considerations

Revert the hook script change if the loud-failure behavior proves to have an unintended false-positive
trigger (e.g. firing even when a DB genuinely is available and reachable) — escalate and fix rather than
silently reverting to the old silent-skip behavior without recording why.

## Implementation Record

Implemented 2026-07-13 by W01Lint.

### What was actually implemented

Replaced the hook's bare `go test ./...` with a required-DB invocation reusing the repository's
established convention (`testkit.RequireDB` / `WOWAPI_REQUIRE_DB=1`, DSN fallback mirroring the
Makefile's `TEST_DSN` compose default). On failure the hook prints an actionable guidance block
(start `make up` / export `DATABASE_URL` / `WOWAPI_PREPUSH_SKIP_DB=1` explicit skip /
`git push --no-verify`) and exits 1. Added `WOWAPI_PREPUSH_SKIP_DB=1` as a loudly-announced explicit
opt-out (per this task's step-3 "how to explicitly opt out"), preserving the hook's
lighter-than-full-CI design. `.githooks/pre-commit` untouched (`git diff --quiet` clean).

### Files changed

`.githooks/pre-push` (+24/−2).

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
| AC-W01-E01-S003-04 | Fail-before (unmodified hook, no DSN): `pre-push: OK`, exit 0 while DB-gated tests all SKIP. After fix: unreachable DB → loud actionable FAIL, exit 1; DB available → DB tests genuinely ran (kernel/database 15.8s vs 7.8s skip-mode) and hook passed; opt-out → loud WARNING + pass | pass | EV-W01-E01-S003-004 (4 logs under `evidence/logs/prepush-*`) |

All hook runs executed in a pristine `git archive HEAD` copy (siblings were editing the shared tree);
HEAD `0a31186cada5c275a588c74081cf977adf346e61`.

### Retest status

Not required.

### Final conclusion

AC satisfied; silent DB-skip eliminated with full fail-before/pass-after proof.

## Deviations Record

None — mechanism (env-var + DSN fallback + loud failure) is within the task's stated latitude.
