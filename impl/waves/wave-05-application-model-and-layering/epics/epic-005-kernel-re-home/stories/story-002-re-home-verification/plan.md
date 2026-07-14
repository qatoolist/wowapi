---
id: PLAN-W05-E05-S002
type: plan
parent_story: W05-E05-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E05-S002

Per mandate §8.5.

## Proposed architecture

A verification-only story: no new production code, only test/verification execution and
evidence-recording against S001's own output.

## Implementation strategy

1. Confirm S001 has landed (its own acceptance criteria satisfied) before beginning this story's
   verification.
2. Run `go list ./kernel/... | wc -l` against the post-move state, recording the actual count against
   S001-T004's own final target-list enumeration.
3. Run the depguard and boundaries-lint suites, confirming green status.
4. Coordinate with wowsociety (PROD-02) to run its build and full identity/authz test suite against
   the `kernel/mfa` shim (or `foundation/mfa` directly, if wowsociety has already migrated).
5. Record all results, with commit SHAs for both wowapi and wowsociety, in this story's own
   verification record and evidence index.
6. Document the verification outcome.

## Expected package or module changes

None — verification-only.

## Expected file changes where determinable

A verification-results document (exact path TBD); no source-code file changes.

## Contracts and interfaces

None new.

## Data structures

None new.

## APIs

None.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

If any verification step fails (package count too high, lint red, wowsociety suite red), record the
failure as a finding requiring S001's own follow-up — this story does not silently patch S001's
implementation to make its own verification pass.

## Security controls

None new — this story verifies security-adjacent properties (the `kernel/mfa` shim's correctness via
wowsociety's suite) but introduces no new control of its own.

## Observability changes

None.

## Testing strategy

- The `go list` count check.
- The depguard/boundaries-lint green-run check.
- wowsociety's full build + identity/authz suite run.

## Regression strategy

Not applicable — this story is itself a one-time verification, though its own evidence becomes part
of this epic's permanent closure record.

## Compatibility strategy

This story's own wowsociety-suite-green check is the compatibility proof for S001's shim.

## Rollout strategy

Single story, executed once S001 has landed.

## Rollback strategy

Not applicable — a verification-only story has nothing of its own to roll back; if verification
fails, the finding routes back to S001 for a fix, recorded as a deviation if S001 has already been
marked `accepted` (which it should not be until this story's verification confirms success, per this
wave's own S001→S002 dependency).

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-6).

## Task breakdown

- **W05-E05-S002-T001** — Kernel package-count and lint verification (steps 2-3 above).
- **W05-E05-S002-T002** — wowsociety identity-suite verification (steps 4 above).
- **W05-E05-S002-T003** — Independent review (per mandate §14, scoped to this story, given FBL-01's
  own "largest single architectural correction" status).

## Expected artifacts

The verification-results document.

## Expected evidence

The `go list` count output; the lint green-run output; wowsociety's full suite run output with both
repositories' commit SHAs.

## Unresolved questions

- Exact mechanism for coordinating wowsociety's own suite run (cross-repo CI trigger, manual
  coordination, or another approach) — to be determined at implementation time, consistent with
  however this programme handles other wowsociety-facing verification steps (e.g. AR-01-S004).

## Approval conditions

This plan is approved for implementation once: (a) S001 has landed, and (b) the owner and reviewer
are assigned.
