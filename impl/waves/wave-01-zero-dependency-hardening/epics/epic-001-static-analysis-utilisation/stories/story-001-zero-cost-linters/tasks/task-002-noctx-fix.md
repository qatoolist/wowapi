---
id: W01-E01-S001-T002
type: task
title: noctx fix (2 named production sites)
status: done
parent_story: W01-E01-S001
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S001-02
artifacts:
  - ART-W01-E01-S001-002
  - ART-W01-E01-S001-003
evidence:
  - EV-W01-E01-S001-002
---

# W01-E01-S001-T002 — noctx fix (2 named production sites)

## Task Definition

### Task objective

Fix `noctx`'s 2 named production hits at `internal/cli/config_delegate.go:34` and
`internal/cli/lint_cmd.go:129` by replacing `exec.Command` with `exec.CommandContext`.

### Parent story

W01-E01-S001 — Enable the zero-cost leak-detection linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001/T003/T004 (disjoint files).

### Detailed work

1. Run `golangci-lint run --enable=noctx ./internal/cli/...` to confirm the fail-first "before" state
   (2 hits at the named lines).
2. At `internal/cli/config_delegate.go:34`, replace `exec.Command(...)` with
   `exec.CommandContext(ctx, ...)`, threading an existing context from the calling function if one is
   already available in scope; if none is threaded through to this call site, use
   `context.Background()` (or `context.TODO()` if the call site is genuinely not yet
   context-aware) with a short comment noting this is a linter-compliance minimum, not full
   cancellation wiring — to be judged at implementation time based on the actual surrounding code.
3. Repeat step 2 at `internal/cli/lint_cmd.go:129`.
4. Re-run `golangci-lint run --enable=noctx ./internal/cli/...` to confirm the "after" state (0 hits).
5. Run the existing CLI test suite (`go test ./internal/cli/...`) to confirm no behavioral regression.

### Expected files or components affected

`internal/cli/config_delegate.go`, `internal/cli/lint_cmd.go`.

### Expected output

Both sites use `exec.CommandContext`; `noctx` reports 0 hits against `internal/cli/...`.

### Required artifacts

ART-W01-E01-S001-002, ART-W01-E01-S001-003.

### Required evidence

EV-W01-E01-S001-002 (fail-before/pass-after static-analysis report).

### Related acceptance criteria

AC-W01-E01-S001-02.

### Completion criteria

`noctx` exits 0 against both files; existing CLI tests still pass; the fail-before state was captured
before the fix.

### Verification method

Direct command execution (`golangci-lint run --enable=noctx`, `go test ./internal/cli/...`), logged
output retained as evidence.

### Risks

Low — a scoped, mechanical 2-site fix with no behavioral change expected beyond subprocess
cancellability.

### Rollback or recovery considerations

Revert the two call-site changes if the existing CLI test suite regresses; low risk given the change
is additive (context parameter) rather than logic-altering.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

Named sites `internal/cli/config_delegate.go:34` and `internal/cli/lint_cmd.go:129` now use `exec.CommandContext` (+`#nosec G204` justifications, shared with S002's gosec triage). noctx's one real non-test hit (`testkit/i18n.go:33`) fixed with `httptest.NewRequestWithContext`; `_test.go` excluded for noctx (documented). DRIFT: noctx v2.11.4 does not flag exec sites — see story `deviations.md` DEV-001/002.

## Verification Record

AC-W01-E01-S001-02: fail-before via gosec G204 (both triages) + code diff (`evidence/static-analysis/noctx-copyloopvar-site-fix.diff`); noctx per-linter run exit 0 (EV-002). **pass** (evidence mechanism substituted per DEV-001)

### Final conclusion

Sites fixed; drift honestly recorded.

## Deviations Record

DEV-001/DEV-002 (story-level): noctx does not detect exec.Command in v2.11.4; 146-hit reality dispositioned via test exclusion + testkit fix.
