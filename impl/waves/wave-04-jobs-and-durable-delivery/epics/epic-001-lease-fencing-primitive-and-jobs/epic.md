---
id: W04-E01
type: epic
title: Lease-fencing primitive and jobs
status: accepted
wave: W04
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-02
depends_on: []
stories:
  - W04-E01-S001
  - W04-E01-S002
  - W04-E01-S003
decisions: []
risks:
  - RISK-W04-001
  - RISK-W04-003
---

# W04-E01 — Lease-fencing primitive and jobs

## Epic objective

Build, from zero, a shared lease/fencing primitive (`lease_token`, monotonic `lease_generation`,
`lease_expires_at`, optional heartbeat) as a reusable kernel building block, and apply it to close
the confirmed duplicate-effect race in `kernel/jobs`: today's claim SQL returns no lease token or
generation, completion/failure paths match only `id`, and `ReclaimStalled` blind-resets every stale
row with no per-row fencing check, so a stalled worker A that is reclaimed by worker B and later
wakes to finalize can silently overwrite B's outcome. This epic delivers the primitive once and
proves it fenced on the jobs queue specifically; `wave.md`'s W04-E02 (notify/webhook) and W04-E03
(bulk) consume — not copy — the same primitive this epic builds.

## Problem being solved

`requirement-inventory.md` row DATA-02 states: "Lease generations/fencing + idempotency (T1–T7) |
IMPL | P0 | planned | W04-E01-S001..S003 | T1 shared primitive is keystone." PLAN DATA-02's own
evidence is explicit: "claim SQL returns no lease token/generation; completion/failure match only
`id`; `ReclaimStalled` blind-resets every stale row with no per-row fencing check. Confirmed race:
A stalls, gets reclaimed by B, B completes, A's eventual finalize silently overwrites B's outcome."
MATRIX CS-11 frames the consequence directly: "duplicate side-effects (double notifications/webhook
posts), pool exhaustion under provider latency, silent lost updates... the at-least-once story has
no fencing, so it is at-least-once-with-overwrites." MATRIX CS-11 also supplies the honest framing
this epic must preserve: the current at-least-once posture is "explicitly documented as an accepted
idempotent-worker tradeoff (`kernel/jobs/runner.go:437-438,108-113`)" — the fix contract is "make the
documented assumption enforceable," not "fix an unacknowledged race." No lease/fencing concept
exists anywhere in `kernel/jobs` today; this epic builds one from zero and wires it through claim,
finalize, and reclaim.

## Scope

- A shared lease/fencing primitive — `lease_token`, monotonic `lease_generation`,
  `lease_expires_at`, optional heartbeat — designed and implemented as a reusable kernel building
  block, not a `kernel/jobs`-only type (S001, PLAN DATA-02 T1).
- Lease columns added to `jobs_queue`; claim SQL assigns a fresh token and `generation+1` per claim
  (S002, PLAN DATA-02 T2).
- Finalize paths (`complete`/`fail`) compare lease token/generation and reject a mismatch, so a
  stale finalize affects zero rows (S002, PLAN DATA-02 T3).
- `ReclaimStalled` bumps `lease_generation` on reclaim, producing a provably new lease epoch (S002,
  PLAN DATA-02 T4).
- A stable job idempotency key and lease context passed to workers, with each worker required to
  declare exactly one duplicate-safety mechanism: inbox/effect-ledger unique on
  `(job_id, effect_name)`, domain CAS, or provider idempotency key (S003, PLAN DATA-02 T5).
- A test proving fencing the queue row does not by itself undo an already-committed stale-worker
  domain transaction — the effect ledger, not the queue row, is the source of truth for whether an
  effect already happened (S003, PLAN DATA-02 T6).
- The named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`, built as a reusable
  chaos harness explicitly shared with DATA-03 (W04-E02) and DATA-04 (W04-E03) — not reimplemented
  by either (S003, PLAN DATA-02 T7).

## Out of scope

- **DATA-03's three-stage claim→effect-outside-tx→fenced-finalize protocol for `kernel/notify` and
  `kernel/webhook`** — W04-E02's scope. This epic's primitive is what E02 consumes; this epic does
  not itself move any remote I/O outside a transaction.
- **DATA-04's leased, `SKIP LOCKED`-honest bulk multi-worker claim path** — W04-E03's scope. Same
  relationship: E03 consumes this epic's primitive and chaos harness, this epic does not implement
  bulk processing.
- **Actually registering a job in wowsociety, or shipping the T5 worker-signature change as
  wowsociety-facing migration guidance** — PLAN's own wowsociety-impact note confirms "zero
  `kernel/jobs` import, zero job registration anywhere in wowsociety" today; T5's breaking signature
  change is recorded as a coordination note (S003 `plan.md`), not resolved or announced by this
  epic.
- **Retrofitting the shared primitive onto W02-E01-S002's already-built interim checkpoint lease
  beyond the migration step S001 itself requires** — S001's own supersession scope is bounded to
  what RISK-W04-001 and RISK-W02-001 name; it is not a general invitation to touch every W02-E01-S002
  code path.

## Source requirements

DATA-02 (T1–T7). No MATRIX CS-ID owns DATA-02 as a dedicated closure spec of its own; it is folded
into MATRIX CS-11 ("Jobs, outbox, lease/fencing, drain") alongside DATA-03 and DATA-04. DATA-02 has
no D-0N architecture-decision dependency — confirmed by `wave.md`'s "Assumptions" scan of
`requirement-inventory.md` §B.

## Architectural context

DATA-02 T1 is, per PLAN's own cross-cutting note, "the single highest-leverage build in this
package — staff and design-review it first," and per its own acceptance criterion the primitive
must be reused "≥3 times, not three independent copies." This epic is therefore the wave's keystone
epic: W04-E02 and W04-E03 both structurally depend on this epic's S001 output, and W04-E01-S003's
chaos harness (T7) is itself reused by DATA-03's 6-boundary chaos test and DATA-04's chaos test
rather than being reimplemented per epic. The three stories are grouped by build-sequence, not by
task count alone: S001 delivers the primitive itself in isolation (T1); S002 applies it to the jobs
queue's claim/finalize/reclaim paths (T2, T3, T4); S003 completes the worker-facing idempotency
contract and builds the shared chaos harness that proves the whole chain fenced (T5, T6, T7). This
grouping is fixed by `impl/analysis/wave-allocation-detail.md`'s canonical allocation ("S001
shared-primitive (T1 — replaces W02-E01-S002's minimal checkpoint lease; migration note); S002
jobs-lease-and-finalize (T2, T3, T4); S003 idempotency-and-chaos (T5, T6, T7 chaos harness — harness
shared with E02/E03)") and is not to be regrouped.

S001 carries an additional architectural responsibility beyond DATA-02 T1's own scope: it
supersedes W02-E01-S002's interim checkpoint lease (built there because this epic's primitive did
not yet exist when W02-E01 needed checkpoint safety for its backfill harness). This is a planned
transition, recorded consistently across `wave.md` ("Assumptions"), `dependencies.md`, `risks.md`
(RISK-W04-001, mirroring W02's own RISK-W02-001), and S001's own `story.md`/`plan.md` — not a silent
scope absorption.

## Included stories

- **W04-E01-S001 — shared-primitive** (PLAN DATA-02 T1): the shared lease/fencing primitive itself
  — the wave's keystone build — plus the planned supersession of W02-E01-S002's interim checkpoint
  lease.
- **W04-E01-S002 — jobs-lease-and-finalize** (PLAN DATA-02 T2, T3, T4): lease columns on
  `jobs_queue`; fenced finalize (reject stale token/generation); fenced `ReclaimStalled` (bump
  generation on reclaim).
- **W04-E01-S003 — idempotency-and-chaos** (PLAN DATA-02 T5, T6, T7): the worker idempotency-
  declaration contract; the effect-ledger-survives-fencing test; the named chaos test
  `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`, built as a chaos harness explicitly shared
  with W04-E02 and W04-E03.

## Dependencies

No dependency on any other W04 epic — this epic is the wave's keystone and W04-E02/W04-E03 both
depend on it (see `dependencies.md`). This epic depends only on W02's exit gate being satisfied at
wave-entry scope for W04 generally (`../../dependencies.md`), with the specific narrower fact that
DATA-02 itself has no dependency on W02's online-migration protocol — confirmed by
`requirement-inventory.md`'s notes column, which cites no W02 dependency for DATA-02.

## Risks

RISK-W04-001 (S001's supersession of W02-E01-S002's interim checkpoint lease carries a migration-
correctness risk on the receiving side) and RISK-W04-003 (S003's T5 worker-signature change is
confirmed breaking, wowsociety coordination required) both originate at wave scope and land entirely
within this epic's stories. See `risks.md` for the epic-scoped elaboration.

## Required decisions

None. DATA-02 has no D-0N architecture-decision dependency in the source (confirmed — see
`wave.md` "Assumptions"). This epic's stories accordingly carry no `decisions/` directory.

## Epic acceptance criteria

- **AC-W04-E01-01**: The shared lease/fencing primitive (`lease_token`, monotonic
  `lease_generation`, `lease_expires_at`, optional heartbeat) exists as one reusable kernel building
  block, proven reused by this epic's own jobs-queue application (S002) and structurally ready for
  W04-E02/W04-E03's independent consumption — not three independent copies.
- **AC-W04-E01-02**: `jobs_queue` carries lease columns; claim SQL assigns a fresh token and
  `generation+1`; finalize paths reject a stale token/generation mismatch (a stale finalize affects
  zero rows); `ReclaimStalled` bumps `lease_generation` on reclaim, producing a provably new lease
  epoch.
- **AC-W04-E01-03**: Every worker declares exactly one duplicate-safety mechanism and cannot
  register without doing so; a test proves fencing the queue row alone does not undo an
  already-committed stale-worker domain transaction; the named chaos test
  `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` proves exactly one logical effect recorded
  and the stale worker's writes rejected at all three named boundaries (domain, external, finalize).
- **AC-W04-E01-04**: All three stories have passed independent review per mandate §14, with S001
  specifically checked for the interim-checkpoint-lease migration being genuinely executed (not
  silently skipped) and S003 specifically checked for the worker-signature breaking change being
  honestly recorded as an open coordination note, not silently resolved.

## Closure conditions

All three stories reach `accepted` (each satisfying its own `closure.md`); AC-W04-E01-01 through
AC-W04-E01-04 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; RISK-W04-001's checkpoint-migration step is confirmed executed and
evidenced (not silently skipped) before this epic can close; RISK-W04-003's coordination note is
recorded as an accepted, tracked-forward item, not silently dropped.

## Status update (2026-07-16)

`status: accepted` (reconfirmed) — all three stories independently reviewed and accepted per
`review-gate-2026-07-16.md`; each story's previously-unfilled `closure.md` "Final status" template
now filled in.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
