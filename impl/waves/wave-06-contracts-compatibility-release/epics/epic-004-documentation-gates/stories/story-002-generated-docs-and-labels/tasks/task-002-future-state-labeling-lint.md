---
id: W06-E04-S002-T002
type: task
title: Future-state-labeling lint
status: done
parent_story: W06-E04-S002
owner: W06E04Impl
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W06-E04-S002-02
artifacts:
  - ART-W06-E04-S002-002
evidence:
  - EV-W06-E04-S002-002
---

# W06-E04-S002-T002 — Future-state-labeling lint

## Task Definition

### Task objective

Build a lint over docs/blueprint/ that fails on an unlabeled normative-sounding future-state block.

### Parent story

W06-E04-S002

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001/W05-E03.

### Detailed work

1. Review docs/blueprint/ for the current set of future-state design blocks and their labeling
   status.
2. Build a lint detecting a normative-sounding future-state block lacking the "target, not implemented"
   (or equivalent) label.
3. Write fixture tests: an unlabeled block fails; a correctly-labeled block passes.

### Expected files or components affected

A new lint tool (exact location TBD) plus its fixture tests.

### Expected output

A lint that fails on an unlabeled future-state block and passes on a correctly-labeled one.

### Required artifacts

ART-W06-E04-S002-002 (future-state-labeling lint).

### Required evidence

EV-W06-E04-S002-002 (lint fixture test report).

### Related acceptance criteria

AC-W06-E04-S002-02.

### Completion criteria

Unlabeled block fails the lint; correctly-labeled block passes.

### Verification method

Direct execution of the lint against both fixture types.

### Risks

Low, per PLAN T5's own risk classification.

### Rollback or recovery considerations

If the lint produces false positives against a legitimate non-future-state block, revise the detection logic; do not silently narrow scope without recording why.

## Implementation Record

Implemented future-state lint for `docs/blueprint/*.md` plus documentation named
`*-target-design.md`. Future/planned/target headings require the immediately-following
`Target, not implemented` label; fenced code is excluded. Added adversarial unlabeled/labeled
fixtures and labeled the two existing blueprint future headings.

- **Files changed:** `internal/tools/docexamples/{future.go,main.go,main_test.go,testdata/future-*.md}`,
  `docs/blueprint/00-overview.md`, and `docs/blueprint/01-domain-model.md`.
- **Tests:** negative unlabeled fixture; positive labeled/current-state/code-fence fixtures; full
  repository gate including W06-E01's `module-dsl-target-design.md`.
- **Implementation date:** 2026-07-13.
- **Commits/PRs:** none; conductor owns integration.
- **Plan relationship:** matched `plan.md`; the documented optional extension to named target-design
  docs was adopted without broadening to all `impl/` records.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E04-S002-02 | focused fixture tests and repository gate | macOS arm64, Go 1.26.5 | Unlabeled fails; labeled/current passes | lint fixture report | pending W06-E04-S002-T003 |

- **Actual result:** PASS — unlabeled fixture failed at line 3 with the required label in the
  diagnostic; labeled/current/code-fence cases passed; repository gate linted 15 documents.
- **Evidence identifier:** EV-W06-E04-S002-002.
- **Execution date/revision:** 2026-07-13; `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus shared changes.
- **Environment:** macOS Darwin 25.5.0 arm64; Go 1.26.5.
- **Retest status:** focused tests and combined gate passed.
- **Final conclusion:** implementation proof passed; independent review remains the acceptance gate.

## Deviations Record

No task-level deviation. See story `deviations.md` for T4-only lifecycle deviation.
