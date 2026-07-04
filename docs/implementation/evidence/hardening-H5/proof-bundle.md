# Hardening H5 — Evidence-layer primitives — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). H5 builds the compliance evidence primitives,
delivered as individually QA-gated commits:

- **E3 — gap-free per-tenant sequence allocator** — DONE (D-0066).
- **E6 — bulk-operation framework** — DONE (D-0068).
- **E2 — data lifecycle: generalized legal hold + DSR ledger** — DONE (D-0072).
- **E4 — snapshot/artifact pipeline** — DONE (D-0076). **H5 complete.**

## E4 — snapshot / artifact pipeline

| Verdict | Fix |
|---|---|
| real (P0) — no immutable versioned-artifact primitive; a compliance product would hand-roll receipt/certificate snapshots | `kernel/artifact`: `Generate` turns product-rendered bytes into an immutable, per-(tenant,kind) versioned `artifacts` row (migration 00021) with `sha256(content)`, a structured sidecar, content-type, template version + effective date; `Get`/`List`/`Verify`. `Templates` resolves the template version effective at a date. |

Framework owns immutability (append-only grants — app_rt has no UPDATE/DELETE, tested), per-(tenant,kind)
versioning, content hashing, tamper-verification, and template-by-effective-date; the **product supplies
the rendered bytes** (e.g. a PDF/A from its own renderer), so no document-format library enters the
kernel — mirroring the storage-port layering. Content is stored in-row (bounded compliance artifacts) so
an artifact is atomic and self-verifying.

Tests (`kernel/artifact/*_test.go`): generate→get round-trips content/hash/sidecar/template; `Verify`
passes clean and **detects an out-of-band content mutation** (hash mismatch); versions increment per kind
(receipt 1,2,3; certificate 1); append-only (app_rt UPDATE+DELETE denied); template resolution picks the
version effective at a date (and rejects pre-effective / unknown kinds). Gate: 0 FAIL, 0 SKIP, 86
packages; boundary lint + 00021 reversibility pass.

## E2 — data lifecycle (generalized legal hold + DSR)

| Verdict | Fix |
|---|---|
| real (P0) — legal hold was a per-document flag; no generalized hold, no DSR primitive, no statutory-override | `kernel/retention` over `legal_holds` + `dsr_requests` (migration 00020). **Holds** (`Place`/`Release`/`IsHeld`/`List`) generalize hold to any `(entity_type, entity_id)` — at most one active hold per entity (partial unique index), consultable by any retention sweep. **DSR** ledger (`Open`/`Complete`/`Reject`/`Get`) tracks export/erasure requests with a **statutory-override reason** for refusing an erasure a retention obligation forbids. |

Scope note: the two concrete, framework-owned primitives are fully implemented. Per-record-class
disposition *over arbitrary product tables* is delivered as an orchestration pattern (the H2 scheduler
drives it; products register per-class dispose/export/erase callbacks — no dynamic-table SQL, keeping the
framework's allowlist-only discipline). The registry+callback wiring is a documented follow-up; the
data-integrity primitives (holds, DSR request lifecycle) are done and tested.

Tests (`kernel/retention/retention_test.go`): hold lifecycle (place→held, duplicate-active conflict,
release→not-held, double-release not-found, re-place after release); tenant isolation; DSR (export
open→complete, re-complete conflict, erasure reject requires+records an override reason). Gate: 0 FAIL,
0 SKIP, 84 packages; boundary lint + 00020 reversibility pass.

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
