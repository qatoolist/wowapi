---
id: PLAN-W03-E01-S002
type: plan
parent_story: W03-E01-S002
status: ready
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W03-E01-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

A new privileged-session resolver component (exact package placement to be determined — likely
within `kernel/auth`, consuming `PrincipalStore` and the `identity_grant` table S001 introduces)
replaces the direct claim-copy of `ImpersonatorUserID`/`BreakGlass`. Capacity-selection enforcement
is added as a check within (or immediately alongside) `Verifier.Actor`.

## Implementation strategy

1. Confirm the exact current state of `Verifier.Actor` at this story's actual start commit,
   including how `CapacityID` is currently read and validated (`ValidateCapacity`).
2. Design the capacity-selection mechanism: how a client presents an explicit choice when more than
   one capacity is active. Candidate approaches (not pre-decided): a required request header/claim
   naming the chosen capacity, a dedicated capacity-selection endpoint issuing a scoped token, or an
   existing mechanism extended. Confirmed at implementation time, recorded in `deviations.md` if it
   diverges from whatever is assumed at plan-approval time.
3. Implement T4: reject an actor with >1 active capacity that has not made an explicit, validated
   choice; validate the choice server-side against the actor's actual active capacities
   (`ValidateCapacity` or equivalent), not merely accept a client-asserted capacity ID.
4. Design the privileged-session resolver's exact interface: a function/method taking an opaque
   grant ID (from wherever the claim eventually carries it, per DEC-Q1's safe default) and returning
   either a verified grant record or a typed rejection reason.
5. Implement the resolver's six-condition rejection matrix: expired grant, revoked grant,
   wrong-tenant grant, wrong-actor grant, forged/unknown grant ID, unauthorized-approver grant (i.e.
   the grant's `approver_id` did not have authority to approve it — exact authority check to be
   determined against whatever approval model S001's schema and this story's implementation settle
   on).
6. Wire the resolver into `Verifier.Actor`, replacing the direct claim copy for
   `ImpersonatorUserID`/`BreakGlass`, preserving the `Actor` struct shape per PLAN's stated
   compatibility preference.
7. Write the adversarial test suite covering all six rejection conditions plus the multi-capacity
   test for T4.

## Expected package or module changes

`kernel/auth` (`Verifier.Actor`, new resolver logic); possibly a new file within `kernel/auth` or a
closely adjacent package for the resolver itself, to keep `auth.go` from growing unboundedly (a
file-organization judgment call for implementation time, not specified here).

## Expected file changes where determinable

- `kernel/auth/auth.go` — `Verifier.Actor`'s capacity-selection and impersonation/break-glass
  population logic.
- A new resolver file (exact name/path TBD at implementation time).

## Contracts and interfaces

A new resolver type/function is introduced (exact signature TBD). `Actor`'s public field shape is
preserved wherever possible per PLAN's stated preference — this is a design constraint, not a
finalized interface spec, since the exact mechanism for surfacing rejection reasons (error type,
sentinel errors, or a result enum) is an implementation-time decision.

## Data structures

No new data structures beyond what S001 introduces (`identity_grant`). This story's own new types,
if any (e.g. a resolver-specific result type), are additive and internal.

## APIs

T4's capacity-selection mechanism may introduce a new request field, header, or endpoint — exact
shape not yet determined; see "Unresolved questions."

## Configuration changes

None anticipated unless the capacity-selection mechanism requires a new config toggle (e.g. to
stage the enforcement behind a flag, mirroring S001's T2 staged-rollout contingency) — to be
determined based on whether a currently-working capacity-less multi-capacity flow is found in
active use (RISK-W03-005).

## Persistence changes

None beyond what S001 introduces. This story only reads `identity_grant`, it does not alter its
schema.

## Migration strategy

Not applicable — no new migration in this story.

## Concurrency implications

None beyond what a standard database read implies. The resolver's grant lookup does not need to
coordinate with concurrent writers beyond standard read-committed semantics, since grant activation/
revocation is itself S001's `identity_grant` write path, not this story's.

## Error-handling strategy

The resolver returns distinguishable, testable rejection reasons for each of the six adversarial
conditions (expired, revoked, wrong-tenant, wrong-actor, forged-ID, unauthorized-approver) — not a
single generic "grant invalid" error — so the adversarial test suite can assert the correct
rejection reason per scenario, and so future observability work (not this story's scope) can
distinguish them.

## Security controls

The resolver is the direct implementation of "never trust the claim, always verify against the
grant record" — the core security control this story delivers. Capacity-selection server-side
validation (T4) similarly ensures a client cannot assert an unentitled capacity.

## Observability changes

Not mandated; see `story.md` "Observability considerations."

## Testing strategy

- Multi-capacity test (T4): an actor with >1 active capacity and no explicit choice is rejected; an
  actor with an explicit, valid choice is accepted; an actor asserting a capacity it does not
  actually hold is rejected (server-side validation, not merely client-trust).
- Adversarial privileged-session test suite (T5): expired grant rejected; revoked grant rejected;
  wrong-tenant grant rejected; wrong-actor grant rejected; forged/unknown grant ID rejected;
  unauthorized-approver grant rejected. Each as an independent test case, not a single combined
  fixture, so a regression in one condition cannot hide behind another passing.
- These map directly to MATRIX CS-07's "revoked capacity" required test class (T4) and contribute
  substantially to the overall SEC-01 adversarial coverage this epic is responsible for.

## Regression strategy

The adversarial test suite itself, run in CI, is the regression guard.

## Compatibility strategy

Preserve the `Actor` struct shape wherever the resolver's rejection-matrix logic allows — per PLAN's
own stated preference, this keeps wowsociety compile-safe even though its runtime behavior changes
(a previously-trusted claim is now verified). See `story.md` "Compatibility considerations" for the
full breaking-change framing.

## Rollout strategy

T5 ships together with S001's T1 per PLAN's own sequencing note ("wowapi ships T1+T5"); T4 may ship
independently once its capacity-selection mechanism is designed, but is grouped into this same story
for tracking simplicity since both close SEC-01's remaining `Verifier.Actor` gaps together.

## Rollback strategy

If T4's enforcement is found to break a currently-working capacity-less flow with no available
product-side UX yet (RISK-W03-005), stage T4 behind a profile flag rather than reverting outright.
T5's resolver can be reverted to the direct claim-copy behavior only as an emergency rollback (not a
planned fallback), since reverting T5 reopens MATRIX CS-07's top-ranked security risk — any such
rollback must be treated as a security incident response, not a routine deployment rollback.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-7). T4's mechanism design (step 2) and
T5's resolver interface design (step 4) should both be settled before their respective
implementation steps (3 and 5) begin.

## Task breakdown

- **W03-E01-S002-T001** — Capacity-selection enforcement (SEC-01 T4).
- **W03-E01-S002-T002** — Privileged-session resolver (SEC-01 T5).
- **W03-E01-S002-T003** — Independent review (mandate §14; P0 security story).

## Expected artifacts

Capacity-selection enforcement logic; privileged-session resolver implementation; the
capacity-selection mechanism's documentation.

## Expected evidence

Multi-capacity test log; adversarial privileged-session test log (six conditions).

## Unresolved questions

- The exact IdP `grant_id` claim contract remains pending DEC-Q1 — this story does not invent it;
  the resolver is built to consume a grant ID from wherever the claim eventually carries it, per the
  documented safe default. What must be determined during implementation: the exact claim field name
  and format, once DEC-Q1 resolves or a interim placeholder convention is agreed.
- T4's exact capacity-selection mechanism (header, claim, dedicated endpoint) is not yet determined
  — to be decided at implementation time in coordination with whatever product-side UX
  W03-E01-S004's cutover plan surfaces.
- The exact "unauthorized-approver" check's authority model (who is entitled to approve a grant) is
  not fully specified by PLAN's T5 row beyond naming the adversarial condition — to be determined
  against S001's finalized `identity_grant` schema and whatever approver-role model already exists
  in the framework's principal/authz layer.

## Approval conditions

This plan is approved for implementation once: (a) W03-E01-S001 has reached `accepted` (hard
dependency), (b) the capacity-selection mechanism and resolver interface unresolved questions above
are answered by design work at story start, and (c) the owner and reviewer are assigned.
