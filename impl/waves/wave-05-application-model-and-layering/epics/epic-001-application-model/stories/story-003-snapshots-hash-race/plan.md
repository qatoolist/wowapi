---
id: PLAN-W05-E01-S003
type: plan
parent_story: W05-E01-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E01-S003

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

Four independent-but-related correctness properties layered onto S001/S002's ownership-bound model:
clone-on-read immutability (T7); explicit-error post-seal rejection (T8, extending S001's D-03
mechanism to registrar retention specifically); a deterministic hash function over the sealed model
(T9); and a race-test suite proving the combination is safe under concurrency (T10).

## Implementation strategy

1. Audit every S002-wrapped registry's exported reader methods at this story's start commit,
   confirming which return a backing map/slice.
2. Convert each such reader to return cloned/immutable data.
3. Write `AR-01/snapshot_immutability_test.go`: mutate a returned value, assert registry internal
   state unaffected, across all wrapped registries.
4. Implement T8's post-seal Context/registrar retention rejection, building on S001's D-03
   error-not-panic mechanism, extended specifically to the retained-registrar-calls-post-boot
   scenario.
5. Write `AR-01/post_seal_mutation_rejection_test.go`: a fixture module retains a registrar, calls it
   post-boot, and receives an explicit error — plus a specific sub-test modeled on wowsociety's
   `s.rulesReg` pattern (retained, never read again) confirming rejection, and a second sub-test
   modeled on `s.rulesStore`/`s.rulesResolver` (built over the registry, used live) confirming no
   false-positive rejection.
6. Implement the deterministic model-hash function, excluding non-deterministic inputs (map order,
   timestamps).
7. Write `AR-01/model_hash_determinism_test.go`: two identical compiles produce a byte-identical
   hash; one changed declaration produces a different hash.
8. Wire the model hash into startup/readiness reporting.
9. Write the race-test suite (`AR-01/race_test_output.txt`'s producing test): concurrent legitimate
   reads under `go test -race`; an illegitimate write attempt fails via T8's rejection mechanism, not
   as an unguarded race.
10. Document all four properties.

## Expected package or module changes

The registries wrapped in S002; the `ApplicationModel`/`Compiler` from S001 (model-hash function,
post-seal rejection enforcement extended to registrar retention).

## Expected file changes where determinable

- Reader-method conversions across the S002-wrapped registry packages.
- A new model-hash function within or adjacent to the `ApplicationModel` type.
- New test files as named above (T7, T8, T9, T10).

## Contracts and interfaces

No new externally-facing contract beyond the model-hash function's own signature (exact shape TBD)
and the post-seal rejection error type (extending S001's own D-03 error type, not introducing a
second one).

## Data structures

None new beyond the model-hash's own internal accumulator (if any) — TBD at implementation time.

## APIs

None externally facing beyond the readiness-reporting integration point for the model hash (T9's own
"emitted at startup/readiness" requirement).

## Configuration changes

None anticipated.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

This is T10's own explicit scope: race tests proving no runtime mutation of the sealed model.
Concurrent legitimate reads must be race-free; an illegitimate write must fail cleanly via T8's
rejection mechanism (an error return), not manifest as an unguarded data race that `go test -race`
would flag as a race condition rather than a handled rejection.

## Error-handling strategy

T8's rejection error must be the same error-not-panic mechanism S001 established for D-03 (this
story extends it to registrar retention specifically, not introduces a parallel mechanism). The
error must be specific enough to distinguish "retained-and-mutated" from other post-seal error
conditions, so wowsociety's own two patterns (dead `s.rulesReg` vs. live `s.rulesStore`/
`s.rulesResolver`) can be correctly told apart by the fixture test.

## Security controls

T8's rejection is itself a security-adjacent availability control (see `story.md` "Security
considerations") — extending D-03's error-not-panic guarantee specifically to the
retained-registrar-calls-post-boot scenario, with a real, named production consumer (wowsociety's
`policy` module) as the validation target.

## Observability changes

The model hash (T9) is itself an observability/diagnosability artifact, emitted at
startup/readiness.

## Testing strategy

- `AR-01/snapshot_immutability_test.go`: mutate-returned-value-assert-internal-state-unaffected,
  across all wrapped registries (T7).
- `AR-01/post_seal_mutation_rejection_test.go`: fixture retains registrar/ctx, calls post-boot,
  receives explicit error — with the wowsociety-pattern-modeled sub-tests distinguishing dead
  retention from live use (T8).
- `AR-01/model_hash_determinism_test.go`: hash-determinism + hash-sensitivity (T9).
- `AR-01/race_test_output.txt`'s producing race test: `go test -race` clean on concurrent legitimate
  reads; illegitimate write fails via T8, not a race (T10).

## Regression strategy

All four test suites become permanent regression guards for their respective properties — a future
change that reintroduces a mutable reader, a silent-no-op or panicking post-seal path, a
non-deterministic hash, or a data race would be caught by the corresponding existing test.

## Compatibility strategy

T8's fixture test explicitly validates against wowsociety's own retained-registrar pattern
(`s.rulesReg`) as a real-world compatibility check — this is the story's primary compatibility
concern, and it is proactive (validated before landing), not reactive.

## Rollout strategy

Single story, landed as its own reviewable unit, sequenced after S002 per the dependency chain.

## Rollback strategy

Revert the T8 rejection mechanism if it is found to produce a false-positive rejection against
wowsociety's legitimately-used `s.rulesStore`/`s.rulesResolver` pattern — escalate for redesign of
the retained-vs-live distinction rather than silently loosening the rejection to avoid the
false positive (which would reopen the gap T8 exists to close).

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-10). T7 and T8 may proceed in parallel
(disjoint concerns — reader immutability vs. registrar retention); T9 depends on the fuller T1-T8
surface being stable; T10 depends on T1-T9.

## Task breakdown

- **W05-E01-S003-T001** — Snapshot-immutability conversion across all wrapped registries (T7).
- **W05-E01-S003-T002** — Post-seal Context/registrar retention rejection, validated against
  wowsociety's named pattern (T8).
- **W05-E01-S003-T003** — Deterministic model hash, emitted at startup/readiness (T9).
- **W05-E01-S003-T004** — Race-test suite proving no runtime mutation of the sealed model (T10).

No independent-review task is added for this story — per this wave's own task-brief guidance,
independent review is added where PLAN's own risk column names a High or explicitly-flagged risk;
this story's task-level risk values (Low-medium, Medium, Low, Low per PLAN's own T7/T8/T9/T10 risk
columns) are materially lower than S001/S002's High-risk items, and epic-level review coverage
(S001/S002's own independent-review tasks) already covers the model's core security-boundary
properties this story builds upon.

## Expected artifacts

Snapshot-immutability conversion (code); post-seal retention rejection (code); the model-hash
function (code); the race-test suite (code).

## Expected evidence

The four named test outputs (snapshot immutability, post-seal mutation rejection, model hash
determinism, race test).

## Unresolved questions

- Exact model-hash algorithm/serialization format (T9) — not specified by PLAN beyond "byte-identical
  hash" and "exclude non-deterministic inputs" — to be chosen at implementation time (e.g. a
  canonical-serialization-then-cryptographic-hash approach).
- Whether the readiness-reporting integration point for the model hash (T9's "emitted at
  startup/readiness") already exists in the framework's current readiness infrastructure or needs to
  be added — to be confirmed at implementation time.

## Approval conditions

This plan is approved for implementation once: (a) the model-hash algorithm is chosen and documented,
and (b) the owner and reviewer are assigned.
