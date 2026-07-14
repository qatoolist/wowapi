---
id: PLAN-W03-E04-S001
type: plan
parent_story: W03-E04-S001
status: ready
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W03-E04-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan does not invent T3's implementation (out of scope, owned by DATA-06 T2 /
W02-E04-S001) — it only describes how T4 consumes T3's output once it exists.

## Proposed architecture

No new package. `Checker.Has` in `kernel/relationship/relationship.go` gains an actor-resolution step
(actor → active capacity → optional party) consuming the post-SEC-01 authoritative principal model,
and its `subject_kind` evaluation logic is extended from a single hard-coded `capacity` branch to a
full matrix covering every schema-enumerated kind, with an explicit fail-closed default for any
unenumerated kind. Relationship-edge mutation call sites (`Relate`/revoke) gain an ownership check, an
audit write, and version-bump logic; attribution is sourced from DATA-06 T2's mechanism, consumed via
its shared file (`registrar_pg.go`), not reimplemented.

## Implementation strategy

1. **Checkpoint: confirm W03-E01 has reached `accepted`.** Do not begin substantive implementation
   before this checkpoint passes — per PLAN's own "do not schedule before it lands" language.
2. Confirm, at this story's actual start commit, the current state of `Checker.Has`
   (`kernel/relationship/relationship.go:42-66`) and the exact shape of the post-SEC-01 principal
   model (W03-E01's actual, accepted output — not the pre-SEC-01 claim-trusting `Actor` shape).
3. Implement T1: extend `Checker.Has` to resolve actor → active capacity → optional party through the
   principal model, enabling party-subject edge evaluation. Write the seeded party-subject-edge test
   (previously-false now true).
4. Implement T2: enumerate every schema-defined `subject_kind` value; confirm which are live
   requirements versus dead schema surface (per T2's own risk note) before writing evaluation
   branches for each; add an explicit fail-closed default branch for any unenumerated kind. Write the
   subject-kind matrix test.
5. Confirm DATA-06 T2 (W02-E04-S001) has landed and its actor-attribution mechanism in
   `registrar_pg.go` is available to consume. If not yet landed, T4 is blocked on this specific input
   — do not reimplement an independent attribution mechanism as a workaround.
6. Implement T4's non-cache-invalidation portions: ownership check on relationship-edge create/revoke;
   attribution via DATA-06 T2's consumed mechanism; audit-row write; version-bump on mutation. Write
   the mutation-governance test for these portions.
7. Check whether W05-E04-S002 (SEC-04's epoch table, D-06) has landed. If yes, implement and test the
   cache-invalidation sub-criterion against it. If no, record the cache-invalidation sub-criterion as
   explicitly deferred-linked in `story.md`/`closure.md` — not silently dropped, not silently assumed
   complete.
8. Independent review (T005), specifically confirming T3's cross-reference to DATA-06 was honored
   (not reimplemented) and the cache-invalidation deferral (if applicable) is honestly recorded.

## Expected package or module changes

`kernel/relationship` (`relationship.go`: `Checker.Has`'s actor-resolution and subject-kind-matrix
extensions; the mutation call sites gaining ownership-check/audit/versioning logic). No changes to
`kernel/auth` (T1 consumes its post-SEC-01 output, does not modify it) or to `registrar_pg.go` (T3's
file, owned by DATA-06 T2 / W02-E04-S001 — T4 consumes its attribution mechanism via a call, not a
modification of that file).

## Expected file changes where determinable

- `kernel/relationship/relationship.go:42-66` — `Checker.Has`'s actor-resolution and subject-kind
  evaluation logic.
- The relationship-edge mutation call site(s) (`Relate`/revoke) — exact file TBD at implementation
  time, gaining ownership-check/audit/versioning logic, and a call into DATA-06 T2's attribution
  mechanism.

## Contracts and interfaces

`Checker.Has`'s external signature is not expected to change — this story extends its internal
evaluation logic, not its public interface. No new public interface introduced by T1/T2. T4's
ownership-check/audit/versioning logic is additive to the existing mutation call sites, not a new
public contract, unless the cache-invalidation sub-criterion (once W05-E04-S002 exists) introduces a
new consumed interface for cache-invalidation signaling — not built here, only consumed once
available.

## Data structures

No new data structure for T1/T2 (evaluation-logic-only extensions). T4's audit-row write may reuse an
existing `kernel/audit` record shape or require a small extension — to be confirmed at implementation
time against the existing audit-writing convention, not invented here.

## APIs

None affected — this story is internal evaluation-logic and mutation-governance work within
`kernel/relationship`, not a runtime HTTP API change.

## Configuration changes

None anticipated.

## Persistence changes

None for T1/T2. T4's audit-write requirement may touch an existing audit table (extension) rather
than introduce a new one — to be confirmed at implementation time.

## Migration strategy

Not applicable for T1/T2. If T4's audit-write requirement needs a schema extension, it is a small,
additive migration, routed through W02-E01's DATA-09 protocol per this wave's entry criteria — not
assumed here without confirming the actual audit-writing gap first.

## Concurrency implications

T4's version-bump-on-mutation requirement must be safe under concurrent relationship-edge mutation
attempts for the same edge — a race where two concurrent mutations both attempt to version-bump the
same edge should not silently lose one mutation's audit trail. Exact mechanism (optimistic
concurrency check, row-level lock, or equivalent) to be determined at implementation time against the
existing pattern for similar versioned mutations elsewhere in the framework, not invented here.

## Error-handling strategy

T2's fail-closed default for an unenumerated `subject_kind` must return a distinguishable "denied,
unsupported kind" result, not a generic error indistinguishable from an infrastructure fault, so a
caller can correctly interpret the denial as intentional. T4's ownership-check failure and
audit-write failure are both treated as blocking the mutation from completing (fail closed), not as
advisory warnings.

## Security controls

T2's fail-closed default is itself a required security control, per its own acceptance criterion
wording ("unsupported/unenumerated kind fails closed"). T4's ownership check, audit write, and
versioning are the core mutation-governance controls this story establishes for relationship-edge
create/revoke.

## Observability changes

T4's audit-row write is itself the primary observability deliverable for this story. The
cache-invalidation sub-criterion's own "triggers observable cache invalidation" requirement is
deferred along with the rest of that sub-criterion pending W05-E04-S002.

## Testing strategy

- T1: a test seeding a party-subject edge, resolving an actor carrying a party through the post-SEC-01
  principal model, asserting the previously-false evaluation is now `true`.
- T2: a matrix test covering every schema-enumerated `subject_kind`, plus a fail-closed test for an
  unenumerated kind.
- T4: a mutation-governance test proving ownership-check enforcement, correct attribution (via
  DATA-06 T2's consumed mechanism), an audit row written, and a version bump on mutation. If
  W05-E04-S002 has landed by implementation time, an additional test proves cache invalidation is
  triggered; if not, this sub-criterion's test is explicitly recorded as pending/deferred rather than
  fabricated against a nonexistent epoch table.
- Fresh re-confirmation (per mandate's fail-first convention) that "no confirmed direct usage" of
  `kernel/relationship` in wowsociety still holds at this story's own execution commit.

## Regression strategy

The subject-kind matrix test (T2) is the durable regression guard against a future schema change
introducing a new `subject_kind` without a corresponding evaluation branch — a new, unhandled kind
should be caught by the matrix test's fail-closed assertion rather than silently passing.

## Compatibility strategy

Given zero confirmed `kernel/relationship` usage in wowsociety (re-confirmed fresh at this story's own
execution time, not merely trusted from PLAN's cited snapshot), this story carries low
compatibility risk. `Checker.Has`'s evaluation extensions are additive from any current consumer's
perspective — no previously-`true` result for an already-correctly-evaluated kind is expected to
change; only previously-`false`/skipped results for party-subject and newly-enumerated kinds change,
and only toward correctness.

## Rollout strategy

T1 and T2 land together (both extend the same `Checker.Has` evaluation logic and share the seeded-edge
test infrastructure). T4's non-cache-invalidation portions land in the same story once T1/T2 and
DATA-06 T2's consumed mechanism are both available; the cache-invalidation portion rolls out
separately, whenever W05-E04-S002 lands, per its deferred-link status.

## Rollback strategy

T1/T2's evaluation-logic extensions are revertible independently of T4's mutation-governance changes,
since they touch different code paths within the same file (`Checker.Has`'s read path versus the
mutation call sites' write path). If T4's ownership-check or audit-write logic is found to block a
legitimate mutation post-rollout, it can be reverted independently of T1/T2.

## Implementation sequence

Steps 1-8 under "Implementation strategy" above, with step 1 (the W03-E01-acceptance checkpoint) as a
hard, non-negotiable gate before any of steps 2-8 begin, and step 5 (confirming DATA-06 T2 has
landed) as a further gate specifically before step 6 (T4's implementation) begins.

## Task breakdown

- **W03-E04-S001-T001** — `Checker.Has` party-subject evaluation (DATA-07 T1) — **gated: cannot start
  before W03-E01 reaches `accepted`.**
- **W03-E04-S001-T002** — `Checker.Has` full subject-kind matrix (DATA-07 T2).
- **W03-E04-S001-T003** — Mutation governance: ownership check, attribution consumption (from DATA-06
  T2), audit write, versioning; cache-invalidation deferred-linked to W05-E04-S002 (DATA-07 T4).
- **W03-E04-S001-T004** — Independent review (mandate §14), specifically confirming T3's scope was
  correctly cross-referenced to DATA-06, not reimplemented, and the cache-invalidation deferral (if
  applicable) is honestly recorded.

## Expected artifacts

`Checker.Has`'s extended party-subject evaluation logic; the full subject-kind evaluation matrix with
fail-closed handling; the mutation-governance implementation (ownership check, attribution
consumption, audit write, versioning).

## Expected evidence

Party-subject-edge seeded test output; subject-kind matrix test output including the fail-closed
case; mutation-governance test output (ownership check, attribution, audit, versioning) — plus, if
W05-E04-S002 has landed by implementation time, cache-invalidation test output.

## Unresolved questions

- Which schema-enumerated `subject_kind` values are live requirements versus dead schema surface
  (T2's own risk note) — to be confirmed at implementation time, not invented here.
- Whether `PrincipalStore`/the post-SEC-01 principal model exposes the exact "active capacity →
  optional party" resolution path this story's T1 needs, or whether a small additive extension to
  that model is required — to be confirmed against W03-E01's actual, accepted implementation shape,
  not assumed here.
- Whether W05-E04-S002 (SEC-04's epoch table) will have landed by this story's own implementation
  window — genuinely unknown per RISK-W03-003; this plan does not assume either outcome.
- The exact audit-row shape for T4 (reuse of an existing `kernel/audit` record versus a small
  extension) — to be confirmed at implementation time.
- The exact concurrency-safety mechanism for T4's version-bump-on-mutation requirement — to be
  determined against existing framework patterns at implementation time.

## Approval conditions

This plan is approved for implementation once: (a) W03-E01 is confirmed `accepted`; (b) DATA-06 T2's
(W02-E04-S001) landing status is confirmed, with T4 explicitly blocked if it has not landed; (c) the
unresolved questions above are answered by implementation-time investigation; and (d) the owner and
reviewer are assigned.
