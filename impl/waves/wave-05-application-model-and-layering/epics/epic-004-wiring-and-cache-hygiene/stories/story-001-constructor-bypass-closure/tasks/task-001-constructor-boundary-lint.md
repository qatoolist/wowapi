---
id: W05-E04-S001-T001
type: task
title: Constructor-boundary lint tool
status: done
parent_story: W05-E04-S001
owner: task
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W05-E04-S001-01
artifacts:
  - ART-W05-E04-S001-001
evidence:
  - EV-W05-E04-S001-001
---

# W05-E04-S001-T001 — Constructor-boundary lint tool

## Task Definition

### Task objective

Implement an AST/lifecycle lint forbidding ad hoc infrastructure constructors outside composition
packages, failing CI on a reintroduced instance.

### Parent story

W05-E04-S001 — Constructor-boundary lint and kernel.go audit.

### Owner

task

### Status

done

### Dependencies

None (parallel-safe with T002).

### Detailed work

1. Check whether AR-02 T6's own lint tooling can be reused; if not, build independently using
   `go/analysis`.
2. Implement the constructor-boundary lint rule.
3. Write `AR-06/constructor_boundary_lint_test.go`: an adversarial fixture reintroducing an ad hoc
   constructor outside composition packages.
4. Document the lint rule.

### Expected files or components affected

A new lint tool (exact location TBD).

### Expected output

A lint rule that fails CI on the adversarial fixture.

### Required artifacts

ART-W05-E04-S001-001.

### Required evidence

EV-W05-E04-S001-001.

### Related acceptance criteria

AC-W05-E04-S001-01.

### Completion criteria

The adversarial fixture fails lint.

### Verification method

Direct execution of `AR-06/constructor_boundary_lint_test.go`.

### Risks

Medium, per PLAN T2's own risk column — "new `go/analysis`-based tooling."

### Rollback or recovery considerations

If the fixture passes lint incorrectly, fix before proceeding.

## Implementation Record

Implemented `internal/tools/constructorlint` as a `go/analysis` analyzer with a
`singlechecker` command. It rejects cross-package framework constructors whose
`New*` names end in an infrastructure category (`Client`, `Manager`, `Pool`,
`Registry`, `Repository`, `Resolver`, `Runtime`, `Sender`, `Store`, or `Writer`)
outside the explicit composition roots. The analyzer uses type information, so import
aliases cannot evade it; it checks generated production files as well as handwritten
ones, while external-library constructors and `_test.go` files are outside its contract.

`make lint-constructors` runs the analyzer over `./...`, and `make lint-boundaries`
depends on it, so the existing CI boundaries job enforces the rule.

### Files changed

- `internal/tools/constructorlint/analyzer.go`
- `internal/tools/constructorlint/cmd/constructorlint/main.go`
- `internal/tools/constructorlint/constructor_boundary_lint_test.go`
- `internal/tools/constructorlint/testdata/src/...`
- `Makefile`
- `go.mod`, `go.sum`

### Tests added or modified

The `analysistest` fixture aliases `kernel/authz` and reintroduces `authz.NewStore`
from a non-composition package; it requires the diagnostic. A control fixture calls
the same constructor from the exact `kernel` composition root and requires no
diagnostic.

### Implementation dates

2026-07-13.

### Technical debt introduced

None. The guarded constructor-name set is intentionally explicit so additions are
reviewable and false positives from value constructors remain excluded.

### Known limitations

The analyzer governs cross-package constructor calls. Same-package constructors are
ordinary package internals and are not composition bypasses.

### Relationship to the approved plan

Implemented as planned. AR-02 T6 supplied no shared analyzer, so this story established
the first `go/analysis` package as the plan permits.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E04-S001-01 | `go test -v ./internal/tools/constructorlint` | Go 1.26.5 | adversarial fixture is diagnosed; composition control is accepted | adversarial-lint report | task |

### Actual result

The analyzer diagnosed the aliased non-composition `authz.NewStore` fixture and emitted
no diagnostic for the exact `kernel` composition root. The test passed.

### Pass or fail

Pass.

### Evidence identifier

EV-W05-E04-S001-001.

### Execution date

2026-07-13.

### Commit or revision

Baseline `733ef3e` plus W05 working-tree changes.

### Environment

Darwin arm64, Go 1.26.5.

### Reviewer

task.

### Findings

No open task-level findings.

### Retest status

Passed after the aliased-import and composition-root control fixtures were added.

### Final conclusion

AC-W05-E04-S001-01 is satisfied.

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
