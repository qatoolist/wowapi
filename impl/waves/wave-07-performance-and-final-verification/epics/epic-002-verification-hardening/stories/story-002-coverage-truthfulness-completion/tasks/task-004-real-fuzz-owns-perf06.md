---
id: W07-E02-S002-T004
type: task
title: Real time-bounded coverage-guided fuzzing (owns PERF-06 T3/T4)
status: done
parent_story: W07-E02-S002
owner: W07-E02-S002 executor
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E02-S002-04
artifacts:
  - ART-W07-E02-S002-004
evidence:
  - EV-W07-E02-S002-004
---

# W07-E02-S002-T004 — Real time-bounded coverage-guided fuzzing (owns PERF-06 T3/T4)

## Task Definition

### Task objective

Wire real -fuzz=<Name> -fuzztime=Ns execution into PR and scheduled CI runs, owning PERF-06 T3/T4's identical scope per CONFLICT-02.

### Parent story

W07-E02-S002

### Owner

unassigned

### Status

todo

### Dependencies

None — but this task's own scope explicitly subsumes PERF-06 T3/T4; confirm no duplicate implementation exists elsewhere in the repository under PERF-06's own name before or after this task lands.

### Detailed work

1. Wire real -fuzz=<Name> -fuzztime=Ns execution into PR CI (short duration).
2. Wire a separate scheduled job with longer duration and corpus retention across runs.
3. Fix make test-fuzz's own wiring if it exists un-wired, per MATRIX CS-13's own evidence.
4. Write a fuzz-duration/corpus-mtime test confirming non-zero fuzzing time beyond seed replay.
5. Confirm no separate, duplicate fuzz-wiring implementation exists anywhere else in the repository
   under PERF-06's own name — this task is the single owner of this scope.

### Expected files or components affected

New CI workflow configuration for PR and scheduled fuzz jobs; make test-fuzz's own wiring confirmed/fixed.

### Expected output

Real coverage-guided fuzzing on PR and scheduled runs, single-owned, no duplicate implementation.

### Required artifacts

ART-W07-E02-S002-004 (real-fuzz PR + scheduled CI job configuration).

### Required evidence

EV-W07-E02-S002-004 (fuzz-duration/corpus-mtime test output).

### Related acceptance criteria

AC-W07-E02-S002-04.

### Completion criteria

Fuzz artifacts prove non-zero time beyond seed replay; corpus retained across scheduled runs; no duplicate PERF-06-named implementation exists.

### Verification method

Direct execution of the fuzz-duration/corpus-mtime test; explicit search for any duplicate PERF-06-named fuzz-wiring implementation.

### Risks

Medium — CI runtime budget impact; needs a time-bound decision the directive doesn't specify, per PLAN T8's own risk note; shared scope with PERF-06, coordinate ownership before implementing (this task IS that coordination, per CONFLICT-02's resolution).

### Rollback or recovery considerations

If a duplicate PERF-06-named implementation is found elsewhere, remove it and consolidate onto this task's own single-owned implementation, recording the consolidation as a deviation if it was already landed independently.

## Implementation Record

### What was actually implemented

`fuzzproof` runs all three native Go fuzz targets with an absolute persistent GOCACHE, rejects output
that contains only seed replay, records positive elapsed executions, snapshots generated corpus file
count/latest mtime, and emits JSON plus raw logs. CI has separate 10s-per-target non-scheduled and
1m-per-target scheduled jobs. Unique save keys and stable restore prefixes retain generated corpus.

### Components and files changed

`.github/workflows/ci.yml`, `Makefile`, `.gitignore`, `internal/tools/fuzzproof/`, and the repository
PR/scheduled evidence reports and raw output.

### Interfaces and configuration

`make test-fuzz` now means real PR fuzzing; explicit `test-fuzz-pr` and `test-fuzz-scheduled` targets
accept `FUZZTIME`, `FUZZ_CACHE`, and `FUZZ_OUTPUT`. CI proof retention is 14/30 days.

### Security and observability

Coverage-guided generation continuously explores filter/sort/cursor inputs. Reports expose duration,
executions, corpus before/after, retention state, and corpus mtime.

### Tests

Parser tests reject seed replay, accept seconds/minutes progress, and inventory corpus files. Actual
profiles ran 10s then 1m per target against the same persistent cache.

### Revision, date, debt, and plan relationship

Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus scoped shared-worktree provenance; implemented
2026-07-14; no debt. The durations and Actions-cache mechanism resolve planned open choices.

## Verification Record

| Acceptance criterion | Actual result | Result | Evidence | Reviewer |
|---|---|---|---|---|
| AC-W07-E02-S002-04 | PR ran 11s/target and corpus 0→520; scheduled restored 520, ran through 1m/target, and grew to 761. | PASS | EV-W07-E02-S002-004 | W05ReviewGateFinal: PASS |

Single-ownership search/review remains part of T005. No second CI implementation was intentionally
added under PERF-06. Environment: Darwin arm64, Go 1.26.5. Actual fuzz evidence completed
2026-07-14. Final task conclusion: verified and artifact/evidence registered.

## Deviations Record

No plan divergence; exact durations and cache retention resolve explicit plan questions.
