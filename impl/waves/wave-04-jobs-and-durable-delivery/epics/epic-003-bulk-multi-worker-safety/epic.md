---
id: W04-E03
type: epic
title: Bulk multi-worker safety
status: accepted
wave: W04
owner: W04BulkSafety
reviewer: W04BulkSafety
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-04
depends_on: []
stories:
  - W04-E03-S001
  - W04-E03-S002
decisions: []
risks: []
---

# W04-E03 — Bulk multi-worker safety

## Epic objective

Close the confirmed contradiction between `kernel/bulk`'s migration comments and its actual
implementation, and make bulk multi-worker processing genuinely safe rather than documented-safe:
correct the false "safe across replicas" claim and enforce single-processor via an immediate
stopgap (S001), then reuse DATA-02's shared lease/fencing primitive to rebuild `bulk_items` claiming
on an atomic, `SKIP LOCKED`-honest, bounded-batch SQL statement, add item idempotency keys and
finalize fencing shared with DATA-02, add pause/resume/cancel lifecycle controls, and prove the
whole thing under a named multi-worker chaos test that reuses the shared chaos harness (S002).

## Problem being solved

`requirement-inventory.md` row DATA-04 states: "Bulk multi-worker safety (T1–T6) | IMPL | P1 |
planned | W04-E03-S001 | T1 stopgap can land early in-wave." The source's own evidence is a direct,
named contradiction between documentation and code: migration `00016`'s header claims the bulk
processing path is "safe across replicas" via `FOR UPDATE SKIP LOCKED`, while `Service.next`
(`kernel/bulk/bulk.go:123-144`) actually issues a plain unlocked `SELECT ... LIMIT 1`, and that
same function's own doc comment concedes "no lock — single processor per operation." This is not a
missing feature; it is a documentation claim that is actively false today, sitting alongside code
whose own comment already admits the true, more limited safety property. `wave.md`'s framing for
this wave situates the fix precisely: "a corrected, fenced, `SKIP LOCKED`-honest bulk multi-worker
claim path replacing today's documented-safe-but-actually-unsafe plain unlocked `SELECT`, plus
pause/resume/cancel lifecycle controls."

## Scope

- **T1 — Immediate stopgap** (S001): correct the false migration-comment claim; enforce
  single-processor via an advisory lock or CAS at the `Service` API boundary, so that a second
  concurrent processor is rejected rather than silently racing against the first.
- **T2 — Lease columns via the shared primitive** (S002): reuse DATA-02's shared lease/fencing
  primitive (`W04-E01-S001`) for `bulk_items`, additive to the existing schema.
- **T3 — Atomic leased claim** (S002): `UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT
  $batch) RETURNING ...`, bounded batch, provably using `SKIP LOCKED` via an `EXPLAIN`-plan
  assertion, preserving `runItem`'s existing idempotent completion CAS guard.
- **T4 — Item idempotency keys, finalize fencing, retry policy, cancellation** (S002): a fenced
  worker's finalize write must be rejected; shares finalize-fencing logic with DATA-02 T3
  (`W04-E01-S002`) — reuse, not reimplementation.
- **T5 — Pause/resume/cancel operation-level controls, bounded batch claims** (S002): correct
  mid-run lifecycle behavior.
- **T6 — Named chaos test** (S002): `DATA-04/chaos/duplicate_worker_test.go` — ≥2 processors
  concurrently claim/retry/pause/resume/cancel the same operation without duplicate effects or
  stale finalization, matching the Wave-3 exit gate wording verbatim; reuses the shared chaos
  harness built in `W04-E01-S003` (DATA-02 T7), not a reimplementation.

## Out of scope

- **DATA-02's shared lease/fencing primitive itself** (generations, heartbeats, the primitive's own
  API) — `W04-E01-S001`'s scope. This epic consumes that primitive for `bulk_items`; it does not
  build it.
- **DATA-02's finalize-fencing design** (`W04-E01-S002`, DATA-02 T3) — this epic's T4 explicitly
  shares that logic rather than designing its own fencing scheme from scratch.
- **The shared chaos-test harness's own construction** (`W04-E01-S003`, DATA-02 T7) — this epic's T6
  is a named consumer of that harness, not its author.
- **`kernel/jobs` or `kernel/notify`/`kernel/webhook` safety work** — `W04-E01`/`W04-E02` scope
  respectively. This epic is scoped to `kernel/bulk` only.
- **wowsociety-side bulk usage** — per the source's own wowsociety-impact note ("Not affected. Zero
  `kernel/bulk` import anywhere in wowsociety"), there is no product-side coordination in scope for
  this epic.

## Source requirements

DATA-04 (T1–T6). DATA-04 has no D-0N architecture-decision dependency — confirmed by `wave.md`'s
own "Assumptions" section, which names only `W04-E04-S001` (DATA-08 W6, D-04) as carrying a
decision dependency in this wave. Accordingly, no story in this epic carries a `decisions/`
directory.

## Architectural context

DATA-04 sits alongside DATA-02 and DATA-03 as the third of three consumers of this wave's keystone
build — the shared lease/fencing primitive (`W04-E01-S001`). `wave.md`'s "Framework capabilities
delivered" names this pattern explicitly: "A shared, reusable lease/fencing primitive ... used
identically by jobs, notify/webhook delivery, and bulk processing — not three independent copies."
Unlike DATA-02 and DATA-03, which build new safety mechanisms where none exist, DATA-04's problem is
a documentation/implementation mismatch on top of an already-existing (but weaker than claimed)
single-processor assumption — `Service.next`'s own doc comment already concedes "no lock — single
processor per operation," so this epic's first move (T1) is to make the code's actual behavior match
what it already honestly claims about itself, before the second move (T2–T6) upgrades that honest
but limited claim into a genuinely multi-worker-safe one. This two-step shape is why the epic is
split into exactly two stories rather than one: `wave-allocation-detail.md`'s canonical allocation
states it plainly — "S001 T1 stopgap (can start at wave entry); S002 T2–T6 leased claims +
lifecycle + chaos" — and `wave.md`'s own entry criteria confirm S001 has no dependency on `W04-E01`'s
primitive landing first ("E03-S001 (DATA-04 T1, the immediate stopgap) may start at wave entry
independent of E01's primitive landing").

## Included stories

- **W04-E03-S001 — stopgap** (DATA-04 T1): correct the false "safe across replicas" migration
  comment; enforce single-processor via advisory lock/CAS at the `Service` API boundary. No
  dependency on `W04-E01`. P0 given the source's own framing ("P1; P0 before advertising
  multi-worker") and its fast, independently-shippable role closing the false-documentation
  sub-issue before the full rewrite.
- **W04-E03-S002 — leased-claims-and-lifecycle** (DATA-04 T2–T6): the shared-primitive lease
  columns, the atomic `SKIP LOCKED` leased claim, item idempotency/finalize-fencing/retry/
  cancellation, pause/resume/cancel lifecycle controls, and the named multi-worker chaos test
  reusing the shared harness. Depends on `W04-E01-S001` and on this epic's own S001 (T1 as an
  interim measure until T2 lands).

## Dependencies

S001 has no dependency on any other epic or story in this wave and may start at wave entry. S002
depends on `W04-E01-S001` (the shared lease/fencing primitive) and on this epic's own S001 (T1's
stopgap as an interim single-processor enforcement measure until T2's lease columns land); S002's T4
additionally cites `W04-E01-S002` (DATA-02 T3's finalize-fencing logic, reused not reimplemented);
S002's T6 additionally cites `W04-E01-S003` (DATA-02 T7's shared chaos harness, reused not
reimplemented). See `dependencies.md` for the full statement.

## Risks

No epic-specific risk beyond the wave-level risk register (`../../risks.md`), which carries no
DATA-04-specific entry. This epic's own `risks.md` records that absence explicitly rather than
inventing a risk unsupported by the source.

## Required decisions

None. DATA-04 has no D-0N architecture-decision dependency (confirmed — see `wave.md`
"Assumptions"). Neither story in this epic carries a `decisions/` directory.

## Epic acceptance criteria

- **AC-W04-E03-01**: The false "safe across replicas" claim in migration `00016`'s header comment is
  removed or corrected; a second concurrent processor attempting to process the same `bulkID` is
  rejected, not silently racing, proven by a 2-processor concurrency test.
- **AC-W04-E03-02**: `bulk_items` gains lease columns via the shared primitive built in
  `W04-E01-S001`, proven by a migration test; the atomic leased-claim SQL
  (`UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`) provably uses
  `SKIP LOCKED` via an `EXPLAIN`-plan assertion and is exercised by a concurrent `N>1` claimer test;
  `runItem`'s existing idempotent completion CAS guard is preserved.
- **AC-W04-E03-03**: A fenced (stale) worker's finalize write to a `bulk_items` row is rejected,
  proven by reusing DATA-02's chaos pattern; item idempotency keys, retry policy, and cancellation
  behave correctly; pause/resume/cancel operation-level controls behave correctly mid-run, proven by
  lifecycle integration tests.
- **AC-W04-E03-04**: The named chaos test `DATA-04/chaos/duplicate_worker_test.go` passes — ≥2
  processors concurrently claim/retry/pause/resume/cancel the same operation without duplicate
  effects or stale finalization, matching the Wave-3 exit gate wording verbatim, and reusing (not
  reimplementing) the shared chaos harness from `W04-E01-S003`.
- **AC-W04-E03-05**: Both stories have passed independent review per mandate §14, with S002
  specifically checked for genuine reuse (not reimplementation) of `W04-E01-S002`'s finalize-fencing
  logic and `W04-E01-S003`'s chaos harness.

## Closure conditions

Both stories reach `accepted` (each satisfying its own `closure.md`); AC-W04-E03-01 through
AC-W04-E03-05 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; no source DATA-04 task (T1–T6) is left uncovered by a task in either
story's `tasks/index.md`.

## Status update (2026-07-16)

`status: accepted` (reconfirmed) — both stories independently reviewed and accepted per
`review-gate-2026-07-16.md`.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
