---
id: PLAN-W05-E04-S001
type: plan
parent_story: W05-E04-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E04-S001

Per mandate §8.5.

## Proposed architecture

A `go/analysis`-based lint pass that walks the codebase for ad hoc infrastructure constructor calls
outside composition packages, and a standalone investigative audit of `kernel/kernel.go`.

## Implementation strategy

1. Check whether AR-02 T6's own lint tooling (if it has landed by this point) can be reused for this
   task's constructor-boundary check; if not, build independently.
2. Implement the constructor-boundary lint rule.
3. Write `AR-06/constructor_boundary_lint_test.go`: an adversarial fixture reintroducing an ad hoc
   constructor outside composition packages.
4. Audit `kernel/kernel.go` line by line for any other instance of the
   closure-captures-a-fresh-instance pattern.
5. Write `AR-06/kernel_constructor_audit.md`, documenting the search scope and findings explicitly.

## Expected package or module changes

A new lint tool (exact location TBD); no expected change to `kernel/kernel.go` unless the audit
finds a new instance of the pattern.

## Expected file changes where determinable

New lint-tool files; `AR-06/kernel_constructor_audit.md`; new adversarial fixture test file.

## Contracts and interfaces

None new beyond the lint tool's own internal analysis pass.

## Data structures

None new.

## APIs

None externally facing.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

The lint rule's failure message should clearly identify the offending constructor call and its
location.

## Security controls

None new — this is a composition-discipline lint, not a capability-security control.

## Observability changes

None beyond CI failure reporting.

## Testing strategy

- `AR-06/constructor_boundary_lint_test.go`: adversarial fixture.
- The audit report itself is the T3 "test" — an investigative document, not a code test.

## Regression strategy

The lint rule is itself the permanent regression guard.

## Compatibility strategy

Not applicable.

## Rollout strategy

Single story, landed as its own reviewable unit.

## Rollback strategy

Not applicable — no runtime behavior change expected unless the audit finds a new instance requiring
a fix, in which case that fix follows T1's own established pattern.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-5). T2 and T3 may proceed in parallel
(disjoint activities — building a lint tool vs. auditing a file).

## Task breakdown

- **W05-E04-S001-T001** — Constructor-boundary lint tool (T2; steps 1-3 above).
- **W05-E04-S001-T002** — `kernel/kernel.go` audit (T3; steps 4-5 above).

No independent-review task is added for this story — PLAN's own risk column values (Medium, Low) are
the lowest in this wave among stories not explicitly named for review in this wave's own task brief.

## Expected artifacts

The lint tool (code); the audit report.

## Expected evidence

The lint test output; the audit report itself.

## Unresolved questions

- Whether AR-02 T6's own lint tooling is available for reuse at this task's own implementation time
  — depends on relative scheduling, not resolved here.

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned.
