---
id: W06-E04-S001-T001
type: task
title: Build the docexamples extractor tool and tag existing normative examples
status: complete
parent_story: W06-E04-S001
owner: W06E04Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W06-E04-S001-01
artifacts:
  - ART-W06-E04-S001-001
  - ART-W06-E04-S001-002
evidence:
  - EV-W06-E04-S001-001
---

# W06-E04-S001-T001 — Build the docexamples extractor tool and tag existing normative examples

## Task Definition

### Task objective

Build the internal/tools/docexamples extractor tool per MATRIX CS-22's mechanics spec, and tag existing normative Go examples in docs/blueprint/*.md and README.md.

### Parent story

W06-E04-S001

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Review the current normative doc set for existing Go code examples.
2. Judge, per example, whether it is normative (should compile) or illustrative pseudo-code.
3. Tag normative examples with `<!-- doc-example: compile -->`.
4. Build the internal/tools/docexamples extractor tool: scan for tagged blocks, write each into a
   generated throwaway package, go build it.

### Expected files or components affected

internal/tools/docexamples/ (new package); docs/blueprint/*.md, README.md (markers added).

### Expected output

An extractor tool that compiles every tagged normative example.

### Required artifacts

ART-W06-E04-S001-001 (docexamples extractor tool), ART-W06-E04-S001-002 (marker convention applied to existing examples).

### Required evidence

EV-W06-E04-S001-001 (extractor-run report).

### Related acceptance criteria

AC-W06-E04-S001-01.

### Completion criteria

Every currently-tagged normative example compiles via the extractor.

### Verification method

Direct execution of the extractor against the current doc set.

### Risks

None beyond the general judgment-call risk in tag-vs-untag decisions per example.

### Rollback or recovery considerations

If a tagged example is later found to be genuinely illustrative pseudo-code mistakenly tagged, untag it and record why in `deviations.md`.

## Implementation Record

Implemented `internal/tools/docexamples` with exact adjacent-marker parsing, mandatory
compile/illustrative fence classification, per-example throwaway packages, and `//line`-mapped
Markdown diagnostics. Classified all existing blueprint Go fences and added one complete normative
current-API example.

- **Files changed:** `internal/tools/docexamples/{examples.go,main.go,main_test.go,testdata/*}` and
  classified `docs/blueprint/*.md` fences.
- **Tests:** parser adjacency/classification, isolated builds, no leaked binaries, stale-symbol
  location, and repository integration in `main_test.go`.
- **Implementation date:** 2026-07-13.
- **Commits/PRs:** none; conductor owns the shared working-tree integration.
- **Plan relationship:** matched `plan.md`; mandatory illustrative classification strengthens the
  visible opt-out requirement without deviating from scope.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E04-S001-01 | `go run ./internal/tools/docexamples -root .` | macOS arm64, Go 1.26.5 | Every tagged example compiles | extractor-run report | pending W06-E04-S001-T003 |

- **Actual result:** PASS — 1 tagged normative example compiled; every other Go fence was explicitly
  classified illustrative; 15 future-state documents and the reference check also passed.
- **Evidence identifier:** EV-W06-E04-S001-001.
- **Execution date/revision:** 2026-07-13; `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus shared
  W05/W06 working-tree changes.
- **Environment:** macOS Darwin 25.5.0 arm64; Go 1.26.5.
- **Retest status:** focused package suite passed.
- **Final conclusion:** implementation proof passed; independent review remains the acceptance gate.
## Deviations Record

No task-level deviation. See story `deviations.md`.
