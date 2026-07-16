---
id: W06-E04-S001-T002
type: task
title: Wire into CI and make docs-check; adversarial staled-example fixture
status: done
parent_story: W06-E04-S001
owner: W06E04Impl
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W06-E04-S001-T001
acceptance_criteria:
  - AC-W06-E04-S001-02
  - AC-W06-E04-S001-03
artifacts:
  - ART-W06-E04-S001-003
  - ART-W06-E04-S001-004
evidence:
  - EV-W06-E04-S001-002
  - EV-W06-E04-S001-003
---

# W06-E04-S001-T002 — Wire into CI and make docs-check; adversarial staled-example fixture

## Task Definition

### Task objective

Wire the docexamples extractor as a CI step in the unit job and a make docs-check target; write an adversarial staled-example fixture proving fail-first.

### Parent story

W06-E04-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E04-S001-T001 (the extractor must exist before it can be wired in).

### Detailed work

1. Wire the extractor as a CI step in the unit job.
2. Add a make docs-check target invoking the same check locally.
3. Write an adversarial fixture: a deliberately staled example calling a removed symbol (resurrectable
   from the pre-AR-05 RunAPI text's git history per MATRIX CS-22's own suggestion).
4. Confirm the staled fixture fails the gate and the current, corrected docs pass.

### Expected files or components affected

Makefile (docs-check target); CI unit-job workflow configuration; a new adversarial fixture file.

### Expected output

A CI-enforced gate proven fail-first via the staled-example fixture.

### Required artifacts

ART-W06-E04-S001-003 (make docs-check target + CI wiring), ART-W06-E04-S001-004 (adversarial staled-example fixture).

### Required evidence

EV-W06-E04-S001-002 (make docs-check execution output), EV-W06-E04-S001-003 (staled-example fixture fail-before/pass-after report).

### Related acceptance criteria

AC-W06-E04-S001-02, AC-W06-E04-S001-03.

### Completion criteria

make docs-check exists and matches CI; the staled fixture fails, corrected docs pass.

### Verification method

Direct execution of make docs-check and the adversarial fixture test.

### Risks

None beyond standard CI-wiring integration risk.

### Rollback or recovery considerations

If the gate produces false positives once wired into real CI, diagnose and fix root cause; do not silently disable the gate.

## Implementation Record

Added `make docs-check`, wired that target into `.github/workflows/ci.yml`'s `unit` job, and added
the removed-symbol `app.RunAPI` fixture. The focused test observes the genuine Go compiler rejection
and asserts its Markdown file/line diagnostic; the corrected repository docs pass the same gate.

- **Files changed:** `Makefile`, `.github/workflows/ci.yml`,
  `internal/tools/docexamples/testdata/stale-example.md`, and `main_test.go`.
- **Tests:** `TestRemovedSymbolFixtureFailsAtDocumentationLocation` plus repository integration.
- **Implementation date:** 2026-07-13.
- **Commits/PRs:** none; conductor owns integration.
- **Plan relationship:** matched `plan.md`.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E04-S001-02 | `make docs-check`; inspect CI unit step | macOS arm64, Go 1.26.5; CI YAML | Local and CI invoke same gate | execution output | pending W06-E04-S001-T003 |
| AC-W06-E04-S001-03 | focused stale fixture test; current docs gate | macOS arm64, Go 1.26.5 | Stale fails; current passes | adversarial fixture report | pending W06-E04-S001-T003 |

- **Actual result:** PASS — Make invoked `go run ./internal/tools/docexamples -root .`; CI unit uses
  `make docs-check`. The stale fixture failed on undefined `app.RunAPI` at fixture line 7; current docs passed.
- **Evidence identifiers:** EV-W06-E04-S001-002 and EV-W06-E04-S001-003-R1 (retests/supersedes -003).
- **Execution date/revision:** 2026-07-13; `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus shared changes.
- **Retest status:** focused tests and Make target passed.
- **Final conclusion:** implementation proof passed; independent review remains the acceptance gate.

## Deviations Record

No task-level deviation. See story `deviations.md`.
