---
id: PLAN-W05-E02-S003
type: plan
parent_story: W05-E02-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E02-S003

Per mandate §8.5.

## Proposed architecture

`kernel/lifecycle`'s manifest is replaced by generation from S002's compiled provider graph — either
by deleting `lifecycle.go`/`manifest.go` outright (if the graph fully supersedes their function) or
by generating them from the graph (if some consumer still expects those specific files to exist). A
legacy port adapter shims `ProvidePort`/`Port` calls onto the typed graph for any existing caller.

## Implementation strategy

1. Determine whether `lifecycle.go`/`manifest.go` can be deleted outright or must be generated
   (based on whether any consumer depends on their specific file/symbol existing).
2. Implement the chosen approach.
3. Write `AR-02/lifecycle_lint_generated_test_output.txt`'s producing regression test: the existing
   5 lint-failure classes still pass, now against the generated/data-driven source.
4. Re-run a repo-wide search (wowapi-internal and wowsociety) for `ProvidePort`/`Port(` call sites,
   confirming PLAN's own "zero external callers" finding still holds at this story's start commit.
5. Implement the legacy port adapter shimming any confirmed caller (wowapi-internal fixtures, if
   any) onto the typed graph.
6. Write `AR-02/legacy_port_adapter_compat_test_output.txt`'s producing integration test.
7. Document both changes.

## Expected package or module changes

`kernel/lifecycle` (deleted or regenerated); a new legacy port adapter package (exact location TBD).

## Expected file changes where determinable

`kernel/lifecycle/lifecycle.go`, `kernel/lifecycle/manifest.go` (deleted or regenerated); new
regression and integration test files as named above.

## Contracts and interfaces

The legacy port adapter preserves the existing `ProvidePort`/`Port` call signatures unchanged.

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

None new beyond what S002's graph already establishes.

## Error-handling strategy

None new beyond what the underlying graph/validation already provides.

## Security controls

None new.

## Observability changes

None new.

## Testing strategy

- `AR-02/lifecycle_lint_generated_test_output.txt`: regression proof for the 5 existing lint-failure
  classes.
- `AR-02/legacy_port_adapter_compat_test_output.txt`: integration proof of unchanged compile/resolve
  behavior for any existing caller.

## Regression strategy

Both named tests are themselves the regression guards for this story's own changes.

## Compatibility strategy

T7's entire purpose is compatibility for any existing `ProvidePort`/`Port` caller.

## Rollout strategy

Single story, landed as its own reviewable unit, sequenced after S002 and as this epic's final
story.

## Rollback strategy

Revert if the lint-class regression test or the legacy-adapter compat test fails — do not ship a
change that breaks existing lint enforcement or existing (even if currently zero) callers.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-7).

## Task breakdown

- **W05-E02-S003-T001** — Lifecycle manifest retirement, generated from the provider graph (T6;
  steps 1-3 above).
- **W05-E02-S003-T002** — Legacy port adapter (T7; steps 4-6 above).

No independent-review task is added for this story — both PLAN risk column values (Low-medium, Low)
are the lowest in this epic, and both tasks are proven by dedicated regression/compatibility tests.

## Expected artifacts

The generated `kernel/lifecycle` manifest replacement (code); the legacy port adapter (code).

## Expected evidence

The two named test outputs.

## Unresolved questions

- Whether `lifecycle.go`/`manifest.go` are deleted outright or generated — depends on whether any
  consumer expects those specific files/symbols to exist, to be confirmed at implementation time.

## Approval conditions

This plan is approved for implementation once: (a) the delete-vs-generate decision for
`kernel/lifecycle` is made, and (b) the owner and reviewer are assigned.
