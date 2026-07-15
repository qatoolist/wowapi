---
id: W01-E01-S002-T004
type: task
title: errorlint fix (kernel/httpx/middleware.go:54)
status: done
parent_story: W01-E01-S002
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S002-03
artifacts:
  - ART-W01-E01-S002-005
evidence:
  - EV-W01-E01-S002-005
---

# W01-E01-S002-T004 — errorlint fix (kernel/httpx/middleware.go:54)

## Task Definition

### Task objective

Fix errorlint's 1 named production hit at `kernel/httpx/middleware.go:54` by replacing a `==`
comparison against `http.ErrAbortHandler` with `errors.Is`. This is a low-risk mechanical fix, not a
defect remediation — `net/http` documents `ErrAbortHandler` as a panicked sentinel value, so the
existing `==` comparison is technically defensible as written; `errors.Is` is harmless and more
idiomatic to adopt.

### Parent story

W01-E01-S002 — Enable and triage the judged linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001-T003/T005-T006 (disjoint files).

### Detailed work

1. Run `golangci-lint run --enable=errorlint ./kernel/httpx/...` to confirm the fail-first "before"
   state (1 hit at line 54).
2. Read the surrounding function in `kernel/httpx/middleware.go` to confirm the comparison is indeed
   against a recovered panic value being checked against `http.ErrAbortHandler`.
3. Replace the `==` comparison with `errors.Is(recovered, http.ErrAbortHandler)` (exact variable
   naming to match the existing code).
4. Re-run `golangci-lint run --enable=errorlint ./kernel/httpx/...` to confirm the "after" state
   (0 hits).
5. Run the existing `kernel/httpx` test suite (`go test ./kernel/httpx/...`), and specifically confirm
   any test exercising the panic-recovery / `ErrAbortHandler` path still passes; add a targeted test
   if no existing test exercises this comparison.

### Expected files or components affected

`kernel/httpx/middleware.go`.

### Expected output

Line 54 uses `errors.Is` in place of `==`; `errorlint` reports 0 hits against `kernel/httpx/...`.

### Required artifacts

ART-W01-E01-S002-005 (updated `kernel/httpx/middleware.go`).

### Required evidence

EV-W01-E01-S002-005 (fail-before/pass-after static-analysis report).

### Related acceptance criteria

AC-W01-E01-S002-03.

### Completion criteria

`errorlint` exits 0 against `kernel/httpx/middleware.go`; the panic-recovery test path (existing or
newly added) still passes; the fail-before state was captured before the fix.

### Verification method

Direct command execution (`golangci-lint run --enable=errorlint`, `go test ./kernel/httpx/...`),
logged output retained as evidence.

### Risks

Low — a single-site, mechanical comparator-function substitution with no behavioral difference
expected for the `http.ErrAbortHandler` sentinel (a package-level `var`, not a wrapped error, so
`errors.Is` and `==` are expected to behave identically for this specific comparison).

### Rollback or recovery considerations

Revert if the `kernel/httpx` test suite regresses; low risk given `errors.Is` against an unwrapped
sentinel error is behaviorally equivalent to `==`.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

middleware.go:54: `==` → error-type guard + `errors.Is` (recover() is `any`; non-error panics keep falling through). Drift sites also fixed: benchbudget main.go:114,118 and sibling-new init_version.go:162 (`%v`→`%w`). `_test.go` excluded for errorlint (documented, DEV-004).

## Verification Record

AC-W01-E01-S002-03: fail-before in both enumerations; errorlint per-linter run exit 0 after (EV-005); kernel/httpx suite ok. **pass**

### Final conclusion

Mechanical adoption of errors.Is; wrapped-sentinel tolerant, behavior-preserving.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
