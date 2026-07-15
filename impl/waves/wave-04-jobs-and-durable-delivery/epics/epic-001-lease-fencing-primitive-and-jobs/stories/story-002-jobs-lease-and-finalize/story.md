---
id: W04-E01-S002
type: story
title: Jobs lease columns, fenced finalize, and fenced reclaim
status: accepted
wave: W04
epic: W04-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-02
depends_on:
  - W04-E01-S001
blocks:
  - W04-E01-S003
acceptance_criteria:
  - AC-W04-E01-S002-01
  - AC-W04-E01-S002-02
  - AC-W04-E01-S002-03
artifacts: []
evidence: []
decisions: []
risks: []
---

# W04-E01-S002 — Jobs lease columns, fenced finalize, and fenced reclaim

## Story ID

W04-E01-S002

## Title

Jobs lease columns, fenced finalize, and fenced reclaim

## Objective

Add lease columns to `jobs_queue`, make claim SQL assign a fresh lease token and incremented
generation per claim, make the finalize paths (`complete`/`fail`) compare lease token/generation and
reject a stale mismatch, and make `ReclaimStalled` bump `lease_generation` on reclaim — closing the
confirmed race where a reclaimed worker's late finalize silently overwrites the reclaiming worker's
outcome.

## Value to the framework

This story is where W04-E01-S001's shared primitive first gets applied to a real consumer and
proves itself: PLAN DATA-02 T2/T3/T4's own acceptance criteria are the epic's central correctness
guarantee ("A statement exceeding budget aborts cleanly" — no, that is W02's DATA-09 T2; here, the
guarantee is "stale finalize affects 0 rows, observably rejected" and "reclaimed row is a provably
new lease epoch"). Until this story lands, `jobs_queue`'s claim SQL "returns no lease token/
generation; completion/failure match only `id`; `ReclaimStalled` blind-resets every stale row with
no per-row fencing check" (PLAN DATA-02 evidence) — the confirmed race (worker A stalls, gets
reclaimed by B, B completes, A's eventual finalize silently overwrites B's outcome) remains open
until this story closes it. This story is also the proof that S001's primitive is genuinely reusable
as designed, not merely theoretically so — the epic's AC-W04-E01-01 requires the primitive be "proven
reused by this epic's own jobs-queue application," and this story is that proof.

## Problem statement

PLAN DATA-02's task table gives the exact acceptance bar this story must satisfy:

- T2: "Add lease columns to `jobs_queue`; claim SQL assigns fresh token + `generation+1` | T1 |
  `claimedJob` carries lease context | Migration + unit | `DATA-02/jobs-lease-migration/` | Reuse
  existing timeout-floor logic, don't introduce a second inconsistent timeout source."
- T3: "Finalize paths compare lease token/generation, reject mismatch | T2 | Stale finalize affects
  0 rows, observably rejected | See T7 chaos test | `DATA-02/finalize/` | Must not regress the
  at-least-once recovery path."
- T4: "`ReclaimStalled` bumps `lease_generation` on reclaim | T2 | Reclaimed row is a provably new
  lease epoch | Same test as T3, asserting generation delta | `DATA-02/reclaim/` | —."

MATRIX CS-11 confirms the current honest framing: "the current at-least-once posture is explicitly
documented as an accepted idempotent-worker tradeoff (`kernel/jobs/runner.go:437-438,108-113`)" —
this story's job is "make the documented assumption enforceable," not fix a previously-unacknowledged
defect. No lease columns, fencing comparison, or generation-bump-on-reclaim exist in `jobs_queue`
today.

## Source requirements

DATA-02 (T2, T3, T4).

## Current-state assessment

Per PLAN's own DATA-02 evidence (to be re-confirmed at this story's own execution commit):
`jobs_queue`'s claim SQL returns no lease token or generation; `claimedJob`'s struct carries no
lease context; the completion and failure (finalize) code paths match a row by `id` only, with no
token/generation comparison; `ReclaimStalled` performs an unconditional reset of every stale row,
with no generation bump and no per-row fencing check. This story's own re-confirmation step (per
this programme's fail-first convention, e.g. W02-E01-S001) is to read `kernel/jobs`'s claim,
finalize, and `ReclaimStalled` SQL/code at this story's actual start commit and confirm these facts
still hold before implementing fencing.

## Desired state

`jobs_queue` carries lease columns backed by W04-E01-S001's shared primitive. Claim SQL assigns a
fresh `lease_token` and `lease_generation+1` on every claim, and the resulting `claimedJob` carries
that lease context forward. The `complete`/`fail` finalize paths compare the caller's supplied
token/generation against the row's current lease state and reject a mismatch — a stale finalize
affects zero rows, and the rejection is observable, not silent. `ReclaimStalled` bumps
`lease_generation` on every row it reclaims, producing a provably new lease epoch distinguishable
from the epoch the stalled worker was operating under.

## Scope

- Adding lease columns to `jobs_queue` (token, generation, expiry — per S001's primitive schema).
- Claim SQL: assign a fresh lease token and `generation+1` per claim; `claimedJob` carries the lease
  context forward.
- Finalize paths (`complete`, `fail`): compare lease token/generation against the row's current
  state; reject mismatch with an observable rejection, not a silent no-op.
- `ReclaimStalled`: bump `lease_generation` on every reclaimed row.
- Reusing the existing timeout-floor logic for the lease-column migration (per T2's own risk note:
  "don't introduce a second inconsistent timeout source").
- A test (or test pair) proving: a stale finalize (from a since-reclaimed lease epoch) affects zero
  rows and is observably rejected; a reclaimed row's `lease_generation` has provably incremented.

## Out of scope

- **The named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`** — W04-E01-S003's
  scope (PLAN DATA-02 T7). This story's own T3/T4 tests prove the fencing/generation-bump mechanics
  directly; S003's chaos test exercises the full stall→reclaim→finalize race end-to-end across all
  three named boundaries.
- **The worker idempotency-declaration contract and the effect-ledger-survives-fencing test** —
  W04-E01-S003's scope (PLAN DATA-02 T5, T6).
- **Applying the shared primitive to `kernel/notify`/`kernel/webhook` or bulk processing** —
  W04-E02's and W04-E03's scope respectively.
- **Designing the shared primitive itself** — W04-E01-S001's scope; this story is a consumer of it,
  not its author.

## Assumptions

- The existing timeout-floor logic referenced by T2's risk note ("reuse existing timeout-floor
  logic") is assumed to be a pre-existing mechanism in `kernel/jobs` governing claim/lease
  expiration timing; this story's plan records the exact reused mechanism once confirmed at
  implementation-time re-read, not invented here.
- The exact lease-column set added to `jobs_queue` mirrors S001's primitive schema
  (`lease_token`, `lease_generation`, `lease_expires_at`, optional heartbeat field) — this story
  does not redesign the schema, it applies S001's already-locked design.
- Whether the finalize paths' rejection surfaces as a returned error, a logged event, or both is not
  specified by the source beyond "observably rejected" (T3's own acceptance criterion) — recorded as
  an implementation-time decision in `plan.md`.

## Dependencies

Depends on W04-E01-S001 (the shared primitive must exist and be locked before `jobs_queue` can
carry lease columns assigned and compared against it). Blocks W04-E01-S003 (the idempotency
contract and chaos harness both operate on the fenced claim/finalize/reclaim chain this story
builds).

## Affected packages or components

`kernel/jobs` — the `jobs_queue` schema (new migration adding lease columns), claim SQL, the
`claimedJob` struct, the `complete`/`fail` finalize code paths, and `ReclaimStalled`.

## Compatibility considerations

The lease-column migration must not break any in-flight job claimed under the pre-fencing schema at
deploy time — per T2's own risk note about reusing existing timeout-floor logic rather than
introducing a second, inconsistent timeout source. This story's plan should consider whether the
migration requires an in-flight-job drain/compatibility window, to be resolved in `plan.md` given no
explicit source guidance beyond the timeout-floor-reuse note.

## Security considerations

The fenced finalize path is itself the security-relevant control this story delivers — rejecting a
stale worker's finalize attempt is what prevents a reclaimed worker's outcome from being silently
overwritten. T3's own risk note ("Must not regress the at-least-once recovery path") is a required
constraint, not optional care: the fencing must reject genuinely stale finalizes without breaking
the legitimate at-least-once retry/recovery path for a worker that was never actually superseded.

## Performance considerations

None separately mandated by the source beyond reusing existing timeout-floor logic rather than
introducing a second timeout mechanism (T2's own risk note) — this is itself a performance/
consistency concern the story must honor, not a separate performance requirement to invent.

## Observability considerations

Finalize-path rejections (a stale token/generation mismatch) should be observable per T3's own
acceptance criterion wording ("observably rejected") — at minimum, distinguishable in logs/metrics
from a successful finalize, so an operator can see the fencing mechanism actually operating rather
than silently dropping stale finalizes with no trace.

## Migration considerations

This story requires a schema migration adding lease columns to `jobs_queue` — per T2's own required
artifact path `DATA-02/jobs-lease-migration/`. The migration itself should follow this framework's
general migration discipline; whether it specifically routes through W02-E01's online-migration
protocol (DATA-09) is not mandated by the source for this particular migration (unlike
W04-E04-S001's DATA-08 W6-T1, which explicitly does) — recorded as an implementation-time decision
in `plan.md`.

## Documentation requirements

Document the lease-column schema, the claim/finalize/reclaim fencing behavior, and the rejection
semantics, so a future `kernel/jobs` contributor understands the fencing contract without re-reading
this story's own planning documents.

## Acceptance criteria

- **AC-W04-E01-S002-01**: `jobs_queue` carries lease columns backed by W04-E01-S001's shared
  primitive; claim SQL assigns a fresh lease token and `generation+1` per claim; `claimedJob` carries
  the resulting lease context forward — proven by a migration + unit test pair.
- **AC-W04-E01-S002-02**: The `complete`/`fail` finalize paths compare lease token/generation and
  reject a mismatch; a stale finalize (from a superseded lease epoch) affects zero rows and is
  observably rejected — proven by a test simulating a stale finalize attempt, without regressing the
  legitimate at-least-once recovery path for a non-superseded worker.
- **AC-W04-E01-S002-03**: `ReclaimStalled` bumps `lease_generation` on every reclaimed row, producing
  a provably new lease epoch — proven by the same test as AC-W04-E01-S002-02, additionally asserting
  the generation delta.

## Required artifacts

- The `jobs_queue` lease-column migration.
- The fenced claim/finalize/reclaim code.
- Documentation of the lease-column schema and fencing behavior.
See `artifacts/index.md`.

## Required evidence

- Migration + unit-test output for lease-column claim assignment.
- Stale-finalize-rejection test output (zero rows affected, observable rejection).
- Reclaim generation-delta test output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W04-E01-S001
recorded, owner/reviewer assignment pending, unresolved questions (timeout-floor reuse mechanism,
rejection-surfacing mechanism, migration-protocol routing) explicitly recorded rather than silently
assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the finalize fencing does not regress the at-least-
once recovery path (T3's own risk note) and that lease columns genuinely reuse S001's shared
primitive rather than a bespoke copy.

## Risks

None distinct at this story's own scope beyond the epic-level risks already tracked (RISK-W04-001,
RISK-W04-003) — this story does not itself carry a new risk-register entry; its central correctness
constraint (not regressing the at-least-once recovery path) is captured as an acceptance-criterion
condition (AC-W04-E01-S002-02) rather than a separate risk, consistent with T3's own risk-column
framing being a design constraint, not an open uncertainty.

## Residual-risk expectations

Residual risk is expected to be low once the stale-finalize and reclaim-generation tests
(AC-W04-E01-S002-02/-03) pass and independent review confirms no regression to the at-least-once
recovery path — this is a well-bounded consumer story applying an already-locked primitive (S001) to
a single, well-understood code surface.

## Plan

See `plan.md`.
