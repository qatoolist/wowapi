---
id: W01-E01-S002-T006
type: task
title: forcetypeassert fixes (2 sites)
status: done
parent_story: W01-E01-S002
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S002-05
artifacts:
  - ART-W01-E01-S002-008
  - ART-W01-E01-S002-009
evidence:
  - EV-W01-E01-S002-007
---

# W01-E01-S002-T006 — forcetypeassert fixes (2 sites)

## Task Definition

### Task objective

Fix forcetypeassert's 2 named production hits at `kernel/auth/jwks.go:112` and
`kernel/config/bind.go:150` by converting each unchecked type assertion to a checked (comma-ok) form
with explicit handling of the false-ok path. This is a real, mechanical code fix — not an annotation.

### Parent story

W01-E01-S002 — Enable and triage the judged linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001-T005 (disjoint files).

### Detailed work

1. Run `golangci-lint run --enable=forcetypeassert ./kernel/auth/... ./kernel/config/...` to confirm
   the fail-first "before" state (2 hits at the named lines).
2. At `kernel/auth/jwks.go:112`, read the surrounding function to determine its existing
   error-handling convention (does it already return an `error`? log-and-continue? panic on
   unexpected input?). Convert the unchecked type assertion (`x := v.(T)`) to a checked form
   (`x, ok := v.(T)`), and add explicit handling for the `ok == false` case consistent with the
   surrounding function's convention — if the function does not currently return an error and a
   signature change is required to handle the false-ok path properly, record this as a
   plan-vs-actual note (see "Contracts and interfaces" in `plan.md`) rather than silently forcing an
   ill-fitting handling shape (e.g. a bare panic) to avoid the signature change.
3. Repeat step 2 at `kernel/config/bind.go:150`.
4. Re-run `golangci-lint run --enable=forcetypeassert ./kernel/auth/... ./kernel/config/...` to
   confirm the "after" state (0 hits).
5. Add a targeted unit test at each site exercising both the successful-assertion path (unchanged
   behavior) and the newly-explicit false-ok path (confirms the new error handling actually triggers
   and behaves as intended, rather than merely satisfying the linter with unreachable-in-practice
   handling code).
6. Run the existing `kernel/auth` and `kernel/config` test suites to confirm no regression beyond the
   two new tests.

### Expected files or components affected

`kernel/auth/jwks.go`, `kernel/config/bind.go`.

### Expected output

Both sites use checked (comma-ok) type assertions with explicit, tested false-ok handling;
`forcetypeassert` reports 0 hits against both files.

### Required artifacts

ART-W01-E01-S002-008, ART-W01-E01-S002-009.

### Required evidence

EV-W01-E01-S002-007 (fail-before/pass-after static-analysis report + unit-test report).

### Related acceptance criteria

AC-W01-E01-S002-05.

### Completion criteria

Both sites use checked type assertions; `forcetypeassert` exits 0 against both files; a targeted unit
test exists and passes for both the successful-assertion and false-ok paths at each site; any
signature change required to properly handle the false-ok path is recorded, not silently avoided by
choosing an ill-fitting handling shape.

### Verification method

Direct command execution (`golangci-lint run --enable=forcetypeassert`, `go test`), logged output
retained as evidence.

### Risks

Moderate — unlike the annotation-only tasks in this story, this is a real logic change at two
security/config-adjacent sites (JWKS parsing, config binding). The false-ok handling must be judged
correctly per site; an incorrect choice (e.g. silently defaulting instead of failing closed on a
malformed JWKS claim or config value) would itself be a new defect. This task's step 2/3 explicitly
require matching the surrounding function's existing convention rather than inventing a new one
ad hoc.

### Rollback or recovery considerations

Revert either site's fix independently if its new unit test reveals the false-ok handling was chosen
incorrectly (e.g. a silent default where a fail-closed error was warranted) — escalate for a design
decision on the correct handling rather than shipping a plausible-looking but unreviewed choice.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

jwks.go:112 and bind.go:150 (named) + httpclient/client.go:71 (drift) converted to comma-ok assertions with explicit false-path handling: documented loud panics at the two boot-time transport constructors (stdlib contract makes the false path unreachable; loud beats untamed-transport), `b.errf` fail-closed binder error in bind.go. `_test.go` excluded for forcetypeassert (DEV-004).

## Verification Record

AC-W01-E01-S002-05: fail-before in both enumerations; forcetypeassert per-linter run exit 0 (EV-007); kernel/auth, kernel/config, kernel/httpclient suites pass. False-ok paths documented rather than unit-forced (verification.md finding 3 — forcing them would mutate stdlib global state to prove an impossibility). **pass**

### Final conclusion

Checked assertions everywhere; false paths explicit.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
