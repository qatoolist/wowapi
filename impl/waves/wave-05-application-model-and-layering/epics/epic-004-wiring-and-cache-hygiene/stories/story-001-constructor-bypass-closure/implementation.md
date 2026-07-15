---
id: IMPL-W05-E04-S001
type: implementation-record
parent_story: W05-E04-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W05-E04-S001

## What was actually implemented

- A `go/analysis` analyzer that reports guarded framework infrastructure constructors outside the
  explicit composition packages.
- An alias-resistant adversarial fixture and exact-`kernel` composition-root control.
- A `singlechecker` executable wired into `make lint-constructors`; the existing
  `lint-boundaries` CI target depends on it.
- A full-file `kernel/kernel.go` constructor/closure audit.

## Components changed

`internal/tools/constructorlint`, `Makefile`, the Go module dependency graph, and story evidence.

## Files changed

- `internal/tools/constructorlint/analyzer.go`
- `internal/tools/constructorlint/cmd/constructorlint/main.go`
- `internal/tools/constructorlint/constructor_boundary_lint_test.go`
- `internal/tools/constructorlint/testdata/src/...`
- `Makefile`, `go.mod`, `go.sum`
- this story's task, artifact, evidence, verification, and closure records

## Interfaces introduced or changed

No runtime API. The new developer command is `make lint-constructors`.

## Configuration changes

`make lint-boundaries` now runs `lint-constructors` first, so the existing CI boundaries job
enforces AR-06.

## Schema or migration changes

Not applicable.

## Security changes

The lint prevents a module from silently bypassing composed/decorated infrastructure instances.

## Observability changes

Not applicable.

## Tests added or modified

`constructor_boundary_lint_test.go` uses `analysistest` to prove an aliased
`authz.NewStore` bypass is rejected while the kernel composition root remains allowed.

## Commits

No commit was created; evidence references baseline `733ef3e` plus the W05 working-tree diff.

## Pull requests

None.

## Implementation dates

2026-07-13.

## Technical debt introduced

None. The explicit guarded-name set makes future constructor categories an intentional review.

## Known limitations

Same-package constructors and third-party value constructors are deliberately outside the rule;
the bypass class is a cross-package framework infrastructure construction.

## Follow-up items

None for AR-06.

## Relationship to the approved plan

Matches the approved T2/T3 plan. No reusable AR-02 analyzer existed, so this story established the
`go/analysis` convention independently as permitted.
