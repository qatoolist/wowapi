# Hardening H5 — Evidence-layer primitives — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). H5 builds the compliance evidence primitives,
delivered as individually QA-gated commits:

- **E3 — gap-free per-tenant sequence allocator** — DONE (D-0066).
- **E6 — bulk-operation framework** — DONE (D-0068).
- E2 — generalized retention / disposition + DSR — pending.
- E4 — snapshot/artifact pipeline — pending.

## E6 — bulk-operation framework

| Verdict | Fix |
|---|---|
| real (P0) — jobs were per-item only; no chunking, progress, partial-failure ledger, or resumability | `kernel/bulk.Service` over `bulk_operations` + `bulk_items` (migration 00016, RLS). `Start` creates the op + one pending item per payload; `Process(txm, tenant, id, limit, fn)` runs up to `limit` pending items (chunked, resumable — picks up only pending), each in its own tenant tx; `Progress` reports Total/Done/Failed/Pending/Status |

Per-item isolation: on success, `fn`'s work commits ATOMICALLY with the `done` mark (one tx); on
failure, that tx rolls back and a second tx records `failed` + the error — so a partial write never
lingers, one item's failure never stops the run (partial-failure ledger), and a crash resumes from the
remaining pending items. Item work must be idempotent (at-least-once, like a job worker).

Tests (`kernel/bulk/bulk_test.go`): all-succeed → completed; **partial-failure ledger** (2 of 5 fail →
Done3/Failed2, errors recorded, and a scratch-table check proves the failed items' writes rolled back
while the 3 successes persisted); **chunked/resumable** (2+2+1 across separate `Process` calls); empty
op → immediately completed. Gate: 0 FAIL, 0 SKIP, 78 packages; boundary lint + 00016 reversibility pass.

## E3 — gap-free per-tenant sequence allocator

| Verdict | Fix |
|---|---|
| real (P0) — no numbered-series primitive; products would hand-roll `MAX()+1` (the wowsociety.app failure) | `kernel/sequence.Allocator` over `sequences` + `sequence_allocations` (migration 00015). `Allocate` increments a per-(tenant,series) counter row **inside the caller's tenant tx** via `INSERT … ON CONFLICT DO UPDATE … RETURNING`, so a number is consumed only on commit (gap-free) and concurrent callers serialize on the row lock (race-free). `Void` records an audited void without renumbering (gaps are intentional and traceable). `Peek` reads the last issued value. |

Deliberately NOT a Postgres sequence: `nextval()` does not roll back, so it leaves gaps — unacceptable
for statutory numbering. The cost (allocations on one series serialize) is inherent to gap-free numbering.

Tests (`kernel/sequence/sequence_test.go`, all against real Postgres + RLS):
- sequential 1,2,3 + `Peek`;
- **gap-free on rollback** — an allocation in a rolled-back tx frees the number (next reuses 1);
- **concurrent no-gaps-no-dupes** — 12 parallel committed allocations yield exactly 1..12;
- void → already-voided is `KindConflict`, unallocated is `KindNotFound`, and a void never renumbers
  (next allocation skips the voided value);
- tenant isolation — each tenant's series is independent under RLS.

Usable today: a module allocates with its own `database.TenantDB` (`sequence.New(idgen).Allocate(ctx, db,
series)`); a `module.Context` convenience accessor is a small follow-up.

Gate: `make ci` + `make ci-container` green — 0 FAIL, 0 SKIP, 76 packages, DB tests forced; boundary
lint + migration reversibility (00015 Down) pass.
</content>
