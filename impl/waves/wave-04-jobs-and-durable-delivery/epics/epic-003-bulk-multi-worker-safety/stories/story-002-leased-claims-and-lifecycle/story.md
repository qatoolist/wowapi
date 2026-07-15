---
id: W04-E03-S002
type: story
title: Leased claims, finalize fencing, lifecycle controls, and the named multi-worker chaos test
status: accepted
wave: W04
epic: W04-E03
owner: W04BulkSafety
reviewer: W04BulkSafety
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-04
depends_on:
  - W04-E01-S001
  - W04-E03-S001
blocks: []
acceptance_criteria:
  - AC-W04-E03-S002-01
  - AC-W04-E03-S002-02
  - AC-W04-E03-S002-03
  - AC-W04-E03-S002-04
  - AC-W04-E03-S002-05
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W04-E03-002
---

# W04-E03-S002 — Leased claims, finalize fencing, lifecycle controls, and the named multi-worker chaos test

## Story ID

W04-E03-S002

## Title

Leased claims, finalize fencing, lifecycle controls, and the named multi-worker chaos test

## Objective

Reuse DATA-02's shared lease/fencing primitive for `bulk_items` (T2); replace the plain unlocked
`SELECT ... LIMIT 1` claim path with an atomic `UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED
LIMIT $batch) RETURNING ...` bounded-batch claim, provable via an `EXPLAIN`-plan assertion (T3); add
item idempotency keys, finalize fencing (shared with DATA-02's finalize-fencing logic), retry
policy, and cancellation (T4); add pause/resume/cancel operation-level lifecycle controls with
bounded batch claims (T5); and prove the entire rebuilt path under the named chaos test
`DATA-04/chaos/duplicate_worker_test.go`, reusing the shared chaos harness built in `W04-E01-S003`
(T6).

## Value to the framework

This story converts `kernel/bulk`'s honest-but-limited "single processor per operation" property
(post-`W04-E03-S001`'s stopgap) into a genuinely multi-worker-safe processing path — the actual
target state DATA-04's own header comment falsely claimed was already true. It is the second half of
this epic's two-step correction, and the half that delivers the real capability: multiple bulk
processors can now run concurrently against the same operation's items, claim disjoint batches via
`SKIP LOCKED`, have their finalize writes fenced against staleness, and be paused, resumed, or
cancelled mid-run without producing duplicate effects. This story also directly discharges two of
this wave's own stated reuse obligations (`wave.md`'s "Framework capabilities delivered": "A shared,
reusable lease/fencing primitive ... used identically by jobs, notify/webhook delivery, and bulk
processing — not three independent copies" and "A shared chaos-test harness (DATA-02 T7), reused —
not reimplemented — by DATA-03's 6-boundary chaos test and DATA-04's multi-worker chaos test").

## Problem statement

The source's own T-row table (reproduced in `epic.md`) states each task's problem directly: T2
("Reuse DATA-02's shared lease primitive for `bulk_items`"); T3 ("Atomic leased claim: `UPDATE ...
FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`, bounded batch" — replacing the
plain unlocked `SELECT ... LIMIT 1` this epic's parent finding identifies in `Service.next`); T4
("Item idempotency keys, finalize fencing, retry policy, cancellation" — "Shares finalize-fencing
logic with DATA-02 T3 — reuse, don't reimplement"); T5 ("Pause/resume/cancel operation-level
controls, bounded batch claims" — "Larger scope — schedule in the full P1/Wave-3 slice, not the
fast-track stopgap"); T6 ("Named chaos test: ≥2 processors concurrently claim/retry/pause/resume/
cancel the same operation without duplicate effects or stale finalization" — "Matches Wave-3 exit
gate wording verbatim" — "Reuse the shared chaos harness").

## Source requirements

DATA-04 (T2, T3, T4, T5, T6).

## Current-state assessment

Per this epic's own `W04-E03-S001` stopgap (a prerequisite interim state, not yet this story's own
starting point until S001 lands): `Service.next` enforces single-processor exclusivity via an
advisory lock or CAS, but still uses the plain unlocked `SELECT ... LIMIT 1` underneath that gate —
no lease columns exist on `bulk_items`, no atomic `SKIP LOCKED` claim exists, no finalize fencing
exists beyond `runItem`'s existing idempotent completion CAS guard, no item idempotency keys exist,
and no pause/resume/cancel lifecycle controls exist. This story's own re-confirmation step is to
read `kernel/bulk`'s current claim, finalize, and lifecycle-control (or absence thereof) code at
this story's actual start commit, and to read `W04-E01-S001`'s landed shared lease/fencing primitive
API, before beginning T2's integration work.

## Desired state

`bulk_items` carries lease columns via the shared primitive built in `W04-E01-S001`. The claim path
is an atomic `UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`
statement, bounded to a configured batch size, provably using `SKIP LOCKED` (not a plain `SELECT`)
via an `EXPLAIN`-plan assertion. `runItem`'s existing idempotent completion CAS guard is preserved
unchanged. Every item carries an idempotency key; a fenced (stale) worker's finalize write is
rejected using the same finalize-fencing logic as `W04-E01-S002` (DATA-02 T3); a defined retry
policy and cancellation path exist. Operation-level pause/resume/cancel controls behave correctly
mid-run, including against in-flight bounded batch claims. The named chaos test
`DATA-04/chaos/duplicate_worker_test.go` passes, reusing the shared chaos harness from
`W04-E01-S003`.

## Scope

- T2: Additive lease columns on `bulk_items` via the shared primitive (`W04-E01-S001`'s API), proven
  by a migration test.
- T3: The atomic leased-claim SQL statement, bounded batch, `EXPLAIN`-plan `SKIP LOCKED` assertion,
  concurrent `N>1` claimer test, preserving `runItem`'s existing completion CAS guard.
- T4: Item idempotency keys; finalize fencing reusing `W04-E01-S002`'s (DATA-02 T3's) logic; retry
  policy; cancellation. Proven by reusing DATA-02's chaos pattern for the fencing check.
- T5: Pause/resume/cancel operation-level controls; bounded batch claims exercised under lifecycle
  transitions. Proven by lifecycle integration tests.
- T6: The named chaos test `DATA-04/chaos/duplicate_worker_test.go`, reusing the shared chaos
  harness built in `W04-E01-S003` (DATA-02 T7).
- An independent-review task per mandate §14, given this story's P0/P1-boundary priority and its
  reuse obligations toward `W04-E01-S002`/`W04-E01-S003` (see `tasks/index.md` "Grouping rationale").

## Out of scope

- **The shared lease/fencing primitive's own construction** — `W04-E01-S001`'s scope. This story
  consumes that primitive's API for `bulk_items`; it does not design or build the primitive itself.
- **The finalize-fencing design itself** — `W04-E01-S002`'s scope (DATA-02 T3). T4 explicitly reuses
  that logic; it does not design an independent fencing scheme for `kernel/bulk`.
- **The shared chaos harness's own construction** — `W04-E01-S003`'s scope (DATA-02 T7). T6
  explicitly reuses that harness; it does not build or redesign a chaos-test harness.
- **`W04-E03-S001`'s stopgap mechanism itself** — already delivered by this epic's own S001. This
  story supersedes it (see RISK-W04-E03-001) but does not re-implement or re-review S001's own
  scope.
- **wowsociety-side bulk usage** — per DATA-04's own wowsociety-impact note ("Not affected. Zero
  `kernel/bulk` import anywhere in wowsociety"), there is no product-side coordination in scope.

## Assumptions

- `W04-E01-S001`'s shared lease/fencing primitive is assumed to expose an API this story's T2 can
  integrate against for `bulk_items` without requiring changes to the primitive itself — if T2's
  integration surfaces a gap in the primitive's API, that is recorded as a cross-story finding
  (potentially a deviation or a follow-up item against `W04-E01-S001`), not silently patched around
  in this story alone.
- The exact bounded-batch size for T3's leased claim (`LIMIT $batch`) is not specified numerically by
  the source beyond "bounded batch" — this story's plan records the exact default value and its
  configurability as an implementation-time decision, per mandate §18.
- T4's exact idempotency-key scheme (per-item UUID, content hash, or another scheme) is not specified
  by the source beyond "Item idempotency keys" — recorded as an implementation-time decision in
  `plan.md`.
- T5's exact pause/resume/cancel API shape (operation-level flags, a state-machine column, or a
  separate control table) is not specified by the source beyond "Pause/resume/cancel operation-level
  controls" — recorded as an implementation-time decision in `plan.md`.

## Dependencies

Depends on `W04-E01-S001` (the shared lease/fencing primitive) — T2's own dependency column per the
source is "DATA-02 T1; T1 as interim," meaning this story depends on both the primitive landing and
on this epic's own `W04-E03-S001` as the interim bridge until it does. Depends on `W04-E03-S001`
(this epic's stopgap) for the same reason. T4 additionally depends on `W04-E01-S002`'s
finalize-fencing logic by reuse (not merely by precedence — the logic itself is meant to be shared).
T6 additionally depends on `W04-E01-S003`'s shared chaos harness by reuse. See `dependencies.md` for
the full statement, including internal T2→T3→T4→T5→T6 sequencing.

## Affected packages or components

`kernel/bulk` (the `Service` type, `Service.next`'s replacement claim path, `runItem`'s finalize
path, and new lifecycle-control entry points); `bulk_items` (new lease columns via the shared
primitive, and any new idempotency-key/lifecycle-state columns T4/T5 require); a new migration or
migrations for the additive schema changes.

## Compatibility considerations

T3's atomic leased-claim SQL replaces `Service.next`'s existing unlocked `SELECT ... LIMIT 1` (as
already gated by `W04-E03-S001`'s stopgap enforcement) — this is a behavioral replacement of the
claim mechanism, not merely an addition alongside it; `W04-E03-S001`'s stopgap enforcement is
explicitly superseded at this point (RISK-W04-E03-001), not left running in parallel indefinitely.
Any existing caller of `Service.next` (or its equivalent) should observe the same external claim
contract (a caller successfully claims exactly one item or batch, or receives no work) even though
the underlying mechanism has changed.

## Security considerations

Finalize fencing (T4) is a security-adjacent data-integrity control: it must genuinely reject a
stale (fenced) worker's finalize write, not merely log a warning while still applying it — this
mirrors the same requirement `W04-E01-S002`'s own finalize-fencing logic already satisfies for jobs,
which is exactly why T4 reuses it rather than building an independently-reviewed (and potentially
independently-flawed) new mechanism.

## Performance considerations

T3's bounded batch size directly controls claim throughput and lock contention under `SKIP LOCKED`;
the exact default and its configurability are implementation-time decisions (see "Assumptions"
above), balancing throughput against the risk of one worker claiming an unfairly large batch.

## Observability considerations

Lease acquisition, claim batch size, fencing rejections (a stale worker's finalize write being
rejected), and pause/resume/cancel transitions should each be observable (logged, at minimum), so an
operator can distinguish normal multi-worker operation from a fencing rejection or an unexpected
lifecycle-control failure — a reasonable implementation-time addition given this story's safety-
critical nature, though not separately itemized by the source beyond the acceptance-criteria-level
requirements themselves.

## Migration considerations

T2's lease-column addition to `bulk_items` is an additive migration (new nullable/default-safe
columns via the shared primitive's own schema pattern, consistent with how `W04-E01-S001` itself
adds lease columns to `jobs_queue`). T4/T5 may require additional additive columns (idempotency key,
lifecycle-state) — exact schema TBD per `plan.md`'s "Unresolved questions." No destructive schema
change is anticipated in this story.

## Documentation requirements

Document: the lease-column schema and its relationship to the shared primitive; the atomic
leased-claim SQL statement and its bounded-batch parameter; the idempotency-key scheme; the
finalize-fencing behavior (with an explicit cross-reference to `W04-E01-S002`'s shared logic, not a
restatement implying independent design); the retry policy; the cancellation path; the pause/resume/
cancel lifecycle-control API.

## Acceptance criteria

- **AC-W04-E03-S002-01**: `bulk_items` gains lease columns via the shared primitive built in
  `W04-E01-S001`, proven by a migration test.
- **AC-W04-E03-S002-02**: The atomic leased-claim SQL statement
  (`UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`) provably uses
  `SKIP LOCKED` via an `EXPLAIN`-plan assertion, is bounded to a configured batch size, and is
  exercised by a concurrent `N>1` claimer test proving no two claimers receive the same row;
  `runItem`'s existing idempotent completion CAS guard is confirmed unchanged.
- **AC-W04-E03-S002-03**: A fenced (stale) worker's finalize write is rejected, proven by reusing
  DATA-02's chaos pattern (via the shared harness in `W04-E01-S003`); item idempotency keys, retry
  policy, and cancellation behave correctly under test.
- **AC-W04-E03-S002-04**: Pause/resume/cancel operation-level controls, exercised against bounded
  batch claims, behave correctly mid-run — proven by lifecycle integration tests.
- **AC-W04-E03-S002-05**: The named chaos test `DATA-04/chaos/duplicate_worker_test.go` passes: ≥2
  processors concurrently claim/retry/pause/resume/cancel the same operation without duplicate
  effects or stale finalization, matching the Wave-3 exit gate wording verbatim, and the test is
  built by reusing (not reimplementing) the shared chaos harness from `W04-E01-S003`.

## Required artifacts

- The `bulk_items` lease-column migration (via the shared primitive).
- The atomic leased-claim SQL implementation.
- The item idempotency-key, finalize-fencing, retry, and cancellation code.
- The pause/resume/cancel lifecycle-control API.
- The named chaos test `DATA-04/chaos/duplicate_worker_test.go`.
- Documentation of the lease schema, claim SQL, idempotency scheme, fencing behavior, retry policy,
  cancellation path, and lifecycle-control API.
See `artifacts/index.md`.

## Required evidence

- Lease-column migration test output.
- `EXPLAIN`-plan `SKIP LOCKED` assertion output plus concurrent `N>1` claimer test output.
- Fenced-finalize-rejection test output (reusing DATA-02's chaos pattern).
- Lifecycle integration-test output (pause/resume/cancel).
- The named chaos test's output (`DATA-04/chaos/duplicate_worker_test.go`).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (`W04-E01-S001`,
`W04-E03-S001`) recorded, owner/reviewer assignment pending, unresolved questions (bounded-batch
default, idempotency-key scheme, pause/resume/cancel API shape) explicitly recorded rather than
silently assumed, and `W04-E01-S001`/`W04-E01-S002`/`W04-E01-S003` confirmed landed (or landing in
lockstep) before T2/T4/T6 respectively can begin in earnest.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all five acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T4's finalize-fencing logic and T6's chaos harness
were genuinely reused (not reimplemented) from `W04-E01-S002`/`W04-E01-S003`, and confirming
`runItem`'s pre-existing completion CAS guard was not silently weakened by T3's rewrite.

## Risks

RISK-W04-E03-002 (epic-level `risks.md`) — T3's rewrite must preserve `runItem`'s existing
idempotent completion CAS guard, not silently drop or weaken it while adding the new claim-side
`SKIP LOCKED` lock. RISK-W04-E03-001 (epic-level `risks.md`) also applies to this story as the
receiving side of the supersession from `W04-E03-S001`'s stopgap.

## Residual-risk expectations

Once T3's completion-CAS-guard preservation is explicitly confirmed by the independent-review task
(RISK-W04-E03-002's mitigation) and T4/T6's genuine reuse of `W04-E01-S002`/`W04-E01-S003` is
confirmed by the same review, residual risk is expected to be low — the remaining uncertainty is
primarily execution risk in a multi-component rewrite, not open design risk, given the shared
primitive, fencing logic, and chaos harness are all built once (in `W04-E01`) and consumed here.

## Plan

See `plan.md`.
