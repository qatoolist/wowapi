---
id: PLAN-W05-E05-S001
type: plan
parent_story: W05-E05-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E05-S001

Per mandate §8.5. MATRIX CS-01's own 5-step "Fix (mechanics, not just intent)" list is followed
directly as this story's implementation strategy — this plan does not invent an alternative
mechanism.

## Proposed architecture

A new `foundation/` top-level tree, populated by `git mv`-ing nine packages out of `kernel/`. A
thin forwarding shim remains at `kernel/mfa` for compatibility. `depguard` and
`scripts/lint_boundaries.sh` are extended to enforce the new layering going forward.

## Implementation strategy

Directly per MATRIX CS-01's own 5 steps:

1. Create the `foundation/` tree.
2. `git mv` each of the nine packages (`webhook, notify, document, artifact, attachment, comment,
   bulk, integration, mfa`), update import paths repo-wide (mechanical for 8 of 9, zero-consumer
   outside wowapi).
3. For `kernel/mfa` → `foundation/mfa` specifically: leave a deprecated forwarding shim (type aliases
   + var forwarding) at `kernel/mfa` for one minor version.
4. Extend `depguard`'s `.golangci.yml` kernel rule to deny `kernel → foundation` imports; add a
   `foundation` rule denying `foundation → app` imports.
5. Extend `scripts/lint_boundaries.sh`'s allowlist so a new kernel package addition fails CI without
   an explicit allowlist edit.

Additionally (this story's own required proof steps, not separately itemized by MATRIX CS-01's own
mechanics list but required by this story's acceptance criteria):

6. Confirm a full build succeeds post-move.
7. Write a behavioral-equivalence test for the `kernel/mfa` shim.
8. Write adversarial fixtures for the depguard extension (both denial rules) and the boundaries-lint
   extension.
9. Document all of the above.

## Expected package or module changes

`foundation/webhook`, `foundation/notify`, `foundation/document`, `foundation/artifact`,
`foundation/attachment`, `foundation/comment`, `foundation/bulk`, `foundation/integration`,
`foundation/mfa` (new locations); `kernel/mfa` (shim, retained); `.golangci.yml`;
`scripts/lint_boundaries.sh`.

## Expected file changes where determinable

Every file under the nine old `kernel/<pkg>` paths (moved via `git mv`); every file repo-wide
importing any of them (import-path updated); `kernel/mfa`'s new shim file(s); `.golangci.yml`;
`scripts/lint_boundaries.sh`.

## Contracts and interfaces

The nine packages' own public APIs are preserved unchanged (behaviour-preserving move); the
`kernel/mfa` shim preserves the pre-move public API surface exactly, forwarding to
`foundation/mfa`.

## Data structures

None new.

## APIs

None externally facing changed (internal Go import-path change only, plus the shim's own preserved
surface).

## Configuration changes

None application-level; `.golangci.yml` (lint configuration) is extended.

## Persistence changes

None.

## Migration strategy

This is itself a large-scale code migration (import-path level) via `git mv`, not a database
migration.

## Concurrency implications

None — a build-time/import-path change.

## Error-handling strategy

The depguard and boundaries-lint extensions must produce clear, actionable CI failure messages
naming the offending import.

## Security controls

The `kernel/mfa` shim's behavioral-equivalence proof is the required security-adjacent control for
this story, given the auth-critical nature of the package it forwards to.

## Observability changes

None material.

## Testing strategy

- Full build success, post-move, across the whole repository.
- `kernel/mfa` shim behavioral-equivalence test: calls through the shim behave identically to direct
  `foundation/mfa` calls.
- Depguard adversarial fixture: an attempted `kernel → foundation` import and an attempted
  `foundation → app` import, both denied.
- Boundaries-lint adversarial fixture: a new, un-allowlisted kernel package addition fails CI.

## Regression strategy

The depguard and boundaries-lint extensions are themselves permanent regression guards against any
future kernel-layering violation.

## Compatibility strategy

The `kernel/mfa` forwarding shim is this story's entire compatibility strategy — see "Compatibility
considerations" in `story.md`.

## Rollout strategy

Single story, landed as its own reviewable unit (though a large diff, given the scale of the move).
The `kernel/mfa` shim's own removal (one minor version later) is explicitly out of this story's own
scope — a future removal task, not performed here.

## Rollback strategy

If the full build fails post-move, or the shim's behavioral-equivalence test fails, revert the move
(or the specific broken import-path update) before proceeding — do not ship a broken build.

## Implementation sequence

As listed under "Implementation strategy" above (MATRIX CS-01's own steps 1-5, plus this story's own
proof steps 6-9). Step 3 (`kernel/mfa`'s shim) should be handled with the most care, consistent with
its own distinct risk profile relative to the other 8 mechanical moves.

## Task breakdown

- **W05-E05-S001-T001** — Foundation tree creation and the 8 zero-consumer package moves (step 1-2,
  excluding `mfa`; step 6's build-success check for these 8).
- **W05-E05-S001-T002** — `kernel/mfa` re-home and forwarding shim (step 2-3 for `mfa` specifically;
  step 7's behavioral-equivalence test).
- **W05-E05-S001-T003** — Depguard extension (step 4; step 8's depguard adversarial fixtures).
- **W05-E05-S001-T004** — Boundaries-lint allowlist extension (step 5; step 8's boundaries-lint
  adversarial fixture).
- **W05-E05-S001-T005** — Independent review (per mandate §14, scoped to this story, given FBL-01's
  "largest single architectural correction" status and the `kernel/mfa` shim's auth-critical
  nature).

## Expected artifacts

The `foundation/` tree (code, moved); the `kernel/mfa` shim (code); the extended depguard
configuration; the extended boundaries-lint allowlist.

## Expected evidence

Full build output; the shim behavioral-equivalence test output; the depguard and boundaries-lint
adversarial-fixture outputs.

## Unresolved questions

- Exact target-list count for `go list ./kernel/... | wc -l` post-move (MATRIX CS-01's own
  acceptance criterion references "target-list count" without stating the exact number in the
  excerpted text available to this plan) — to be confirmed at implementation time by counting the
  actual retained kernel packages (39 minus the 9 moved, or the equivalent, minus/plus any other
  change in flight).
- Exact `kernel/mfa` shim mechanics detail (which specific types/vars need aliasing/forwarding) —
  determined by the actual `kernel/mfa` package's own public surface at implementation time.

## Approval conditions

This plan is approved for implementation once: (a) the exact target kernel-package count is
confirmed, and (b) the owner and reviewer are assigned.
