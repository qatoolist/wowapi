---
id: W01-E01-S001-T003
type: task
title: copyloopvar fix (1 named production site)
status: done
parent_story: W01-E01-S001
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S001-03
artifacts:
  - ART-W01-E01-S001-004
evidence:
  - EV-W01-E01-S001-003
---

# W01-E01-S001-T003 — copyloopvar fix (1 named production site)

## Task Definition

### Task objective

Fix `copyloopvar`'s 1 named production hit at `app/maintenance.go:148` by removing the dead
pre-1.22 loop-variable-capture idiom (e.g. `v := v` inside the loop body, now unnecessary given the
module's Go version, which scopes loop variables per-iteration since Go 1.22).

### Parent story

W01-E01-S001 — Enable the zero-cost leak-detection linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001/T002/T004 (disjoint files).

### Detailed work

1. Run `golangci-lint run --enable=copyloopvar ./app/...` to confirm the fail-first "before" state
   (1 hit at `app/maintenance.go:148`).
2. Confirm the module's `go.mod` `go` directive is >= 1.22 (the precondition for this idiom being
   dead code, not merely stylistic).
3. Remove the redundant loop-variable-capture statement at `app/maintenance.go:148`.
4. Re-run `golangci-lint run --enable=copyloopvar ./app/...` to confirm the "after" state (0 hits).
5. Run `go test ./app/...` to confirm no behavioral regression (particularly around any goroutine or
   closure capturing the loop variable within `maintenance.go`, where a real pre-1.22 bug would have
   manifested — confirm the removal is genuinely inert given the current Go version).

### Expected files or components affected

`app/maintenance.go`.

### Expected output

`app/maintenance.go:148` no longer contains the pre-1.22 capture idiom; `copyloopvar` reports 0 hits.

### Required artifacts

ART-W01-E01-S001-004.

### Required evidence

EV-W01-E01-S001-003 (fail-before/pass-after static-analysis report).

### Related acceptance criteria

AC-W01-E01-S001-03.

### Completion criteria

`copyloopvar` exits 0 against `app/maintenance.go`; `go test ./app/...` passes; the fail-before state
was captured before the fix.

### Verification method

Direct command execution (`golangci-lint run --enable=copyloopvar`, `go test ./app/...`), logged
output retained as evidence.

### Risks

Low — single-site, mechanical removal of dead code, contingent on confirming the Go version
precondition in step 2.

### Rollback or recovery considerations

Revert if `go test ./app/...` regresses (would indicate the loop-variable capture was not actually
dead, i.e. a goroutine/closure still depends on the old per-loop-iteration variable identity in a way
Go 1.22's semantics don't cover as expected) — in that case, escalate rather than silently
re-introducing the capture idiom without investigation.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

Deleted the dead `rj := rj` capture at `app/maintenance.go:148` (named site) and the 6 equivalent captures in test files (see story `implementation.md` §3).

## Verification Record

AC-W01-E01-S001-03: fail-before in both triage enumerations (7 hits); copyloopvar per-linter run exit 0 after (EV-003). **pass**

### Final conclusion

All copyloopvar hits eliminated.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
