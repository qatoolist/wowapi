---
id: PLAN-W06-E02-S001
type: plan
parent_story: W06-E02-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E02-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

An expanded merge-target struct in `internal/cli/openapi_cmd.go` (or its logical successor) that
deserializes every OpenAPI 3.1 top-level field and every `components.*` field, applying an explicit
per-field merge policy function for each, followed by a structural-validation pass (3.1.1/2020-12)
before the merged document is accepted as output, followed by a semantic-diff CI job comparing the
merged output against the previous release's merged output, keyed to DX-05's v1 policy for what counts
as a breaking change.

## Implementation strategy

1. Enumerate every OpenAPI 3.1 top-level field and every `components.*` sub-field, cross-checking
   against the OpenAPI 3.1.1 specification itself (not merely the fields MATRIX CS-15 names as
   examples — `security`, `tags`, `servers`, `webhooks` are the named examples, not necessarily the
   complete list).
2. Design an explicit per-field merge policy for each field — union, identical-required, or
   reject-on-conflict, as appropriate to that field's own semantics — and document the rationale per
   field.
3. Implement the expanded merge struct and per-field policy logic.
4. Write the fixture-driven test suite: one fragment per field type, confirming each field's documented
   policy is honored (merges correctly or is explicitly rejected).
5. Evaluate the OpenAPI 3.1 validator dependency candidate (`pb33f/libopenapi` per MATRIX CS-15),
   including a security/licence review; select and wire it in, or select an alternative if the
   candidate fails review.
6. Implement structural validation of the merged document against 3.1.1/2020-12 using the selected
   validator; write a malformed-output negative fixture test.
7. Implement the semantic-diff gate, keyed to DX-05's already-ratified v1 policy; write a seeded
   intentional breaking-change fixture test.
8. Document the per-field merge policy, the validator choice and its review outcome, and the
   semantic-diff gate's behavior.

## Expected package or module changes

`internal/cli/openapi_cmd.go` (expanded merge struct, per-field policy, validation, semantic-diff
wiring); a new validator dependency in `go.mod` (pending T2's decision); a new semantic-diff CI job
(exact location TBD).

## Expected file changes where determinable

- `internal/cli/openapi_cmd.go` — expanded merge-target struct and per-field policy logic.
- New fixture test files, one fragment per OpenAPI 3.1 top-level/`components.*` field type.
- New structural-validation wiring and its negative fixture test.
- New semantic-diff CI job configuration and its seeded breaking-change fixture.
- `go.mod` — new validator dependency (pending T2's decision).

## Contracts and interfaces

The merge-target struct's own expanded field set is the primary contract change; no runtime API is
affected (this is a CLI/build-time tool).

## Data structures

The expanded merge-target struct itself; no application data model change.

## APIs

None affected — this story is CLI/build-tooling-internal.

## Configuration changes

None anticipated beyond the validator dependency's own configuration (if any) and the semantic-diff
gate's CI-job configuration.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None — this is a single-invocation CLI command, not a concurrent runtime path.

## Error-handling strategy

A fragment field that cannot be merged per its documented policy must be rejected with a clear,
field-specific error, not a generic "merge failed" message — consistent with the loud-on-collision
precedent already established for `paths`/`components.schemas`.

## Security controls

The `security` field's merge policy is itself a required security control, per `story.md` "Security
considerations" — a declared security requirement must never be silently dropped. The validator
dependency's own security/licence review (T2) is a required control before that dependency is trusted.

## Observability changes

The merge command should report, per field, whether it merged or was rejected (implementation-time
detail).

## Testing strategy

- T1: fixture-driven, one fragment per OpenAPI 3.1 top-level/`components.*` field type, confirming
  correct-merge or explicit-rejection per the documented policy.
- T2: structural-validation test against the selected validator, including a malformed-merged-output
  negative fixture.
- T3: seeded intentional-breaking-change fixture, confirming the semantic-diff gate fails it.

## Regression strategy

Once T1's fixture suite and T3's semantic-diff gate are wired into CI, they become the ongoing
regression guard against a future change silently reintroducing a field-drop or a breaking API change.

## Compatibility strategy

T3's semantic-diff gate is itself the compatibility-enforcement mechanism for the OpenAPI surface,
consuming DX-05's already-ratified v1 policy rather than defining a new compatibility model.

## Rollout strategy

Single story, landed as its own reviewable unit; T2's validator dependency addition should land with
its review record attached in the same change, not as a follow-up.

## Rollback strategy

If the expanded merge policy for a specific field proves wrong in practice (e.g. a legitimate use case
is rejected that should have merged), revise that field's policy directly and re-run the fixture suite
— revert the specific field's policy change, not the whole merge-struct expansion.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–8). Step 5 (validator security/licence review)
must occur before step 6 (structural validation implementation) trusts that dependency.

## Task breakdown

- **W06-E02-S001-T001** — Full-field merge struct and per-field policy (steps 1–4 above).
- **W06-E02-S001-T002** — Validator-dependency decision and structural validation (steps 5–6 above).
- **W06-E02-S001-T003** — Semantic-diff gate keyed to DX-05's v1 policy (step 7 above).
- **W06-E02-S001-T004** — Independent review.

## Expected artifacts

The expanded merge struct with per-field policy; the fixture-driven per-field test suite; the
structural-validation wiring; the semantic-diff CI gate; per-field merge-policy documentation.

## Expected evidence

Per-field-type fixture test output; structural-validation test output (including malformed-output
negative fixture); seeded-breaking-fixture semantic-diff test output; the validator-dependency
security/licence review record.

## Unresolved questions

- Exact per-field merge policy for each newly-covered field — to be designed at implementation time
  (T1).
- Validator dependency choice (`pb33f/libopenapi` or an alternative) — to be decided at implementation
  time with security/licence review (T2).
- Exact semantic-diff CI job location and invocation shape.

## Approval conditions

This plan is approved for implementation once: (a) the owner and reviewer are assigned, and (b) T2's
validator-dependency decision has been made with its security/licence review recorded (a precondition
for T2/T3's own implementation, not for the plan's approval to begin T1).
