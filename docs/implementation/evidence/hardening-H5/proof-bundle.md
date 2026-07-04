# Hardening H5 — Evidence-layer primitives — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). H5 builds the compliance evidence primitives,
delivered as individually QA-gated commits:

- **E3 — gap-free per-tenant sequence allocator** — DONE (D-0066). This bundle.
- E2 — generalized retention / disposition + DSR — pending.
- E6 — bulk-operation framework — pending.
- E4 — snapshot/artifact pipeline — pending.

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
