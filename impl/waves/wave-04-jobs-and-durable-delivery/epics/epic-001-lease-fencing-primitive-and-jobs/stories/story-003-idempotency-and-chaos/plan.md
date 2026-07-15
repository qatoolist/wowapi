---
id: PLAN-W04-E01-S003
type: plan
parent_story: W04-E01-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E01-S003

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information — in particular, this plan does **not** resolve T5's breaking
worker-signature-change coordination question; it records it as open per RISK-W04-003.

## Proposed architecture

A worker-registration-time contract layered onto `kernel/jobs`'s existing worker-registration
mechanism, requiring each registering worker to declare exactly one duplicate-safety mechanism.
A stable idempotency key and lease context threaded through to worker invocation (T5's breaking
signature change). A chaos-test harness — built as a standalone, reusable test-infrastructure
package/module, not embedded inline in a single test file — providing the pause-after-claim,
expire, reclaim-via-B, resume-A, multi-boundary-finalize-attempt mechanics as composable primitives
that W04-E02 and W04-E03 can parameterize for their own effect types (notify/webhook for DATA-03,
bulk-operation effects for DATA-04) rather than reimplement.

## Implementation strategy

1. Re-read `kernel/jobs`'s worker-registration mechanism and `kernel/outbox/relay.go`'s inbox-dedup
   logic at this story's actual start commit to confirm the current-state assessment still holds.
2. Confirm whether PF-ARCH's typed operation model exists in the repository as of this story's
   start; if not, design the idempotency-declaration contract as a runtime registration-time check
   (resolving `plan.md`'s "Unresolved questions" item on enforcement mechanism).
3. Design the stable job idempotency key's derivation and the lease-context-to-worker invocation
   shape; document that this is T5's confirmed-breaking worker-signature change (RISK-W04-003), with
   the wowsociety coordination note recorded explicitly, not resolved.
4. Implement the idempotency-declaration contract: registration fails without exactly one declared
   mechanism (inbox/effect ledger, domain CAS, provider idempotency key).
5. Write a duplicate-effect / registration-rejection test proving the contract is enforced.
6. Write a test proving fencing the queue row alone does not undo an already-committed stale-worker
   domain transaction — construct a scenario where S002's fencing succeeds but an idempotency-
   ignoring worker's domain transaction has already committed, and confirm the effect ledger (not
   the queue-row fencing) is what catches the duplicate.
7. Design the chaos harness's pause/expire/reclaim/resume mechanics as reusable, parameterizable
   test infrastructure.
8. Implement the named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`, exercising
   all three named boundaries (domain, external, finalize), using the harness built in step 7.
9. Document the harness's reuse contract for W04-E02/W04-E03.
10. Document the idempotency contract, the fencing/effect-ledger distinction, and the T5 coordination
    note.

## Expected package or module changes

`kernel/jobs`'s worker-registration mechanism (idempotency-declaration enforcement) and worker-
invocation signature (idempotency key + lease context, T5's breaking change). A new chaos-harness
package (exact location TBD — expected to structurally correspond to the source's own
`DATA-02/chaos/` path notation, adapted to this repository's actual test-package conventions at
implementation time).

## Expected file changes where determinable

- `kernel/jobs`'s worker-registration code (exact file path TBD, not yet confirmed by file/line
  pending this story's own start-commit re-read).
- `kernel/jobs`'s worker-invocation signature (exact file path TBD) — this is T5's breaking change.
- A new duplicate-effect / registration-rejection test.
- A new effect-ledger-vs-fencing test.
- A new chaos-harness package and the named test file
  `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` (or its repository-actual equivalent path).

## Contracts and interfaces

The worker-registration contract: a registering worker must supply exactly one of three declared
duplicate-safety mechanisms; the registration API rejects a worker declaring zero or more than one.
The worker-invocation signature gains the stable idempotency key and lease context (T5's breaking
change) — exact signature shape TBD, consuming S001/S002's already-built lease context.

## Data structures

No new persisted data structure beyond what the idempotency-declaration mechanism itself needs to
track a worker's declared mechanism (in-memory registration metadata, not necessarily a new
database table) — exact shape TBD at implementation time.

## APIs

The worker-registration API and worker-invocation signature both change — T5's confirmed-breaking
change. No other public API is affected.

## Configuration changes

None anticipated beyond whatever the idempotency-declaration contract itself requires (e.g. a
registration-time validation flag) — not confirmed by the source beyond T5's own scope.

## Persistence changes

None from this story directly, beyond whatever the effect-ledger-vs-fencing test (step 6) needs as
test fixtures — no new production schema is mandated by T5/T6/T7 beyond what already exists
(`kernel/outbox/relay.go`'s existing inbox-dedup mechanism, per MATRIX CS-11's own evidence
refinement).

## Migration strategy

Not applicable — no schema or data migration is required by this story.

## Concurrency implications

The chaos test itself is the concurrency proof: it must genuinely exercise the pause-after-claim,
expire, reclaim-via-B, resume-A race condition, not a serialized approximation of it. The harness's
pause mechanism (step 7) must produce a genuine window where B can claim and complete before A
resumes, not a mechanism that accidentally serializes the two workers.

## Error-handling strategy

A worker registering without a declared duplicate-safety mechanism must fail registration with a
clear, actionable error (not a silent default or a generic validation failure). A's stale writes at
each of the three named boundaries must be rejected observably (consuming S002's own
observable-rejection mechanism at the finalize boundary; domain and external boundaries need their
own rejection observability, defined by this story).

## Security controls

The idempotency-declaration contract is itself the required security/correctness control (see
`story.md` "Security considerations") — not optional hardening.

## Observability changes

The chaos test must be able to observe "exactly one logical effect recorded" and "A's writes
rejected" at each of the three named boundaries — this requires observability hooks at the domain
and external boundaries in addition to S002's existing finalize-boundary observability.

## Testing strategy

- Duplicate-effect / registration-rejection test: a worker without a declared mechanism cannot
  register.
- Effect-ledger-vs-fencing test: construct a stale-worker domain transaction that has already
  committed despite fencing; confirm the effect ledger (not queue-row fencing) is what catches the
  duplicate.
- The named chaos test itself: pause A after claim, expire, reclaim via B, B completes, resume A,
  attempt finalize at all three named boundaries — exactly one logical effect recorded, A's writes
  rejected at every boundary. This is, per PLAN T7's own "Tests" column, "This is the test" — no
  substitute or paraphrase is acceptable.

## Regression strategy

The chaos harness, once built, becomes the regression guard for W04-E02's and W04-E03's own chaos
work — any change to the shared primitive or its consumers that breaks the fencing guarantee should
be caught by this harness's own test (for jobs) and by E02/E03's parameterized reuse of it (for
notify/webhook and bulk).

## Compatibility strategy

T5's worker-signature change is confirmed breaking. This story's compatibility strategy is to record
the change honestly (RISK-W04-003) rather than attempt a compatibility shim the source does not
request — PLAN's own risk note treats this as a coordination problem, not a compatibility-engineering
problem, given wowsociety's current zero usage.

## Rollout strategy

Single story, landed as its own reviewable unit, as the epic's final story. The chaos harness's
reusability for W04-E02/W04-E03 should be validated (at minimum, structurally reviewed) before this
story is treated as complete, since those epics' own stories may already be in progress in parallel
and need a stable harness contract to consume.

## Rollback strategy

Revert the idempotency-declaration contract if it is found to reject a legitimate worker's valid
declaration due to a contract-parsing defect; escalate for redesign rather than silently loosening
the "exactly one declared mechanism" requirement. The chaos test itself has no runtime rollback
concern (test-only).

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–10). Steps 4–6 (idempotency contract,
duplicate-effect test, effect-ledger test) may proceed in parallel with steps 7–8 (chaos harness,
chaos test) since they touch different code surfaces, but step 8 (the chaos test) logically depends
on T5's idempotency key/lease context (step 3-4) being available to exercise realistic worker
invocation during the chaos scenario.

## Task breakdown

- **W04-E01-S003-T001** — Idempotency-declaration contract, key/lease-context threading, and
  duplicate-effect test (steps 2–5 above).
- **W04-E01-S003-T002** — Effect-ledger-vs-fencing test (step 6 above).
- **W04-E01-S003-T003** — Chaos harness and the named chaos test (steps 7–9 above).
- **W04-E01-S003-T004** — Evidence aggregation (consolidated bundle across T001–T003).
- **W04-E01-S003-T005** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The idempotency-declaration contract; the effect-ledger-vs-fencing test; the named chaos test and
its reusable harness; documentation of the contract, the fencing/effect-ledger distinction, the
harness reuse contract, and the T5 coordination note.

## Expected evidence

Duplicate-effect / registration-rejection test output; effect-ledger-vs-fencing test output; the
named chaos test's own passing-run output covering all three named boundaries; a consolidated
evidence bundle.

## Unresolved questions

- Whether PF-ARCH's typed operation model exists in the repository to enforce the idempotency
  declaration at compile time, or whether this story implements a runtime check instead — to be
  confirmed at this story's own start-commit re-read.
- The exact idempotency key's derivation scheme.
- The chaos harness's exact pause-after-claim mechanism (test hook, synchronization primitive, or
  controlled delay).
- **T5's worker-signature-change coordination with wowsociety** — explicitly NOT resolved by this
  plan. PLAN DATA-02 T5's own risk note states "worker signature change is breaking — coordinate
  with wowsociety even though it has zero current job usage." This plan records the coordination as
  an open, tracked item (RISK-W04-003) for the framework's own changelog/migration-guide process; it
  is not this story's role to unilaterally decide the coordination outcome.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above — most centrally,
the PF-ARCH enforcement-mechanism question — are answered, (b) the owner and reviewer are assigned,
and (c) W04-E01-S002 has reached at least `implemented` status (this story's own dependency). The T5
coordination note (RISK-W04-003) is explicitly NOT a precondition for this plan's approval — it
remains open through this story's own implementation and closure, tracked forward per the risk
register.
