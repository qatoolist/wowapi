---
id: W01-E01-S003-T001
type: task
title: go mod verify CI step
status: done
parent_story: W01-E01-S003
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S003-01
artifacts:
  - ART-W01-E01-S003-001
evidence:
  - EV-W01-E01-S003-001
---

# W01-E01-S003-T001 — go mod verify CI step

## Task Definition

### Task objective

Add a `go mod verify` step to `.github/workflows/ci.yml`'s build/test pipeline so that a corrupted or
tampered local module cache is caught in CI, closing a gap where zero hits for this command exist
anywhere in the workflow today.

### Parent story

W01-E01-S003 — Close supply-chain and pre-push hook hygiene gaps.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T002/T003/T004 (T002 touches `security-scan.yml`, T003 touches
`.githooks/pre-push`, T004 is a confirmation activity against a different section of `ci.yml`).

### Detailed work

1. Re-read `.github/workflows/ci.yml` fresh, at this task's actual start commit, to confirm `go mod
   verify` is still absent (per `story.md`'s "Current-state assessment") and to identify the correct
   job/step placement (expected alongside existing `go vet`/build steps, exact location to be
   determined from the pipeline's actual job graph).
2. Add a `go mod verify` step at the identified location.
3. Run the updated workflow (or a local equivalent — `go mod verify` on the shell against the current
   `go.sum`) to confirm it passes against the current, presumably-clean module cache.
4. Confirm the step's failure would fail the overall CI job (i.e., it is not marked
   `continue-on-error` or otherwise made non-blocking) — this is a supply-chain integrity check and
   should be a hard gate, consistent with the story's stated purpose.

### Expected files or components affected

`.github/workflows/ci.yml`.

### Expected output

An updated `ci.yml` with a `go mod verify` step that runs and passes as part of the normal pipeline.

### Required artifacts

ART-W01-E01-S003-001 (updated `ci.yml`).

### Required evidence

EV-W01-E01-S003-001 (CI execution record / command-execution log).

### Related acceptance criteria

AC-W01-E01-S003-01.

### Completion criteria

`go mod verify` runs as a distinct CI step and passes; a run log against a named commit SHA is
retained as evidence; the step is confirmed to be blocking (not `continue-on-error`).

### Verification method

Direct command execution (`go mod verify` locally, and the corresponding CI run), logged output
retained as evidence per `evidence/index.md`.

### Risks

Near-zero — a standard, well-understood Go toolchain command with a single, unambiguous pass/fail
outcome.

### Rollback or recovery considerations

Revert the added step if it produces an unexpected false-positive failure (e.g. due to a CI-runner-
specific module cache quirk) that cannot be resolved within this task's bounded scope; escalate rather
than silently removing the step without recording why.

## Implementation Record

Implemented 2026-07-13 by W01Lint.

### What was actually implemented

Added a distinct `go mod verify (supply-chain integrity — module cache vs go.sum)` step to
`ci.yml`'s `unit` job, immediately after the `go.mod tidy check` step (the job that already owns
module hygiene). Step failure fails the job (no `continue-on-error`).

### Files changed

`.github/workflows/ci.yml` (+2 lines).

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
| AC-W01-E01-S003-01 | Local `go mod verify` at HEAD `0a31186cada5c275a588c74081cf977adf346e61`: `all modules verified`, exit 0; `actionlint` clean over the edited workflow | pass (in-CI run log pending conductor push — see story `verification.md` carry-forward note) | EV-W01-E01-S003-001 (`evidence/logs/gomodverify-and-actionlint.log`) |

### Retest status

Wave gate's first CI run to be registered as `retested` evidence.

### Final conclusion

AC satisfied; step wired and proven against a clean module cache.

## Deviations Record

None — matched plan.
