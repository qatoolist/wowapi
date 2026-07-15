---
id: PLAN-W07-E01-S004
type: plan
parent_story: W07-E01-S004
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Plan — W07-E01-S004

Per mandate §8.5. This plan covers two distinct source-finding sets folded into one story per this
wave's own canonical allocation: PERF-05 (T1-T5) and CS-16's own 7-package bench-coverage expansion.
Confirmed facts, planned changes, and assumptions are distinguished explicitly below.

## Proposed architecture

Required, universal checksum-metadata enforcement on the upload path, with the existing full-hash
fallback moved behind an explicit labeled repair invocation; dedicated fallback-invocation metrics; a
resumable async backfill mechanism for legacy objects. Separately: 7 new benchmark files, one per
CS-16-named package, each targeting its own specific hot path, with corresponding `bench-budgets.txt`
entries and a `BENCH_PKGS` extension in the `Makefile`.

## Implementation strategy

**PERF-05:**
1. Enumerate every current upload call site (T1); require checksum-metadata persistence at each; write
   an integration test confirming no `GetObject` call on a normal `Stat`.
2. Move the full-hash fallback behind an explicit, size/time-bounded, labeled repair invocation (T2);
   decide the exact API-surface shape (new `Stat` variant vs. `RepairChecksum` method); confirm other
   `storage.ObjectInfo`-implementing adapters still compile and behave correctly.
3. Add dedicated fallback-invocation metrics (T3).
4. Build or confirm the "legacy objects lacking checksum metadata" inventory mechanism, then implement
   the resumable async backfill (T4).
5. Publish before/after evidence against `perf/reference-v1.json` (T5).

**CS-16 bench-coverage expansion:**
6. Add a benchmark to `kernel/database` targeting tenant-tx open/commit.
7. Add a benchmark to `jobs` targeting the claim/finalize loop.
8. Add a benchmark to `outbox` targeting relay dispatch batch.
9. Add a benchmark to `workflow` targeting its own most request-relevant hot path (exact target TBD at
   implementation time).
10. Add a benchmark to `auth` targeting token verify.
11. Add a benchmark to `mfa` targeting TOTP derive.
12. Add a benchmark to `httpclient` targeting guarded dial.
13. Add `bench-budgets.txt` entries for all 7, in the same PR as their benchmarks (per PERF-06's own
    fail-closed policy); extend `BENCH_PKGS` in the `Makefile`.

## Expected package or module changes

`adapters/storage/s3` (checksum enforcement, bounded repair path); possibly the `storage.ObjectInfo`
port (T2's own API-surface decision); `kernel/database`, `kernel/jobs`, `kernel/outbox`,
`kernel/workflow`, `kernel/auth`, `kernel/mfa`, `kernel/httpclient` (new benchmark files); `bench-
budgets.txt`; `Makefile`.

## Expected file changes where determinable

- `adapters/storage/s3`'s implementation files — checksum-required enforcement, bounded repair path,
  metrics.
- Possibly `kernel/storage`'s `ObjectInfo` port definition (T2's own API-surface decision).
- A new resumable-backfill mechanism (exact location TBD).
- 7 new `*_bench_test.go` files, one per named package.
- `bench-budgets.txt` (7 new entries).
- `Makefile` (`BENCH_PKGS` extended, lines around `Makefile:206-214`).

## Contracts and interfaces

T2's own possible `storage.ObjectInfo` port change is the primary contract-surface consideration —
affecting every adapter implementing that port, per PLAN's own risk note.

## Data structures

The legacy-object inventory mechanism (T4) may require a new data structure or table to track which
objects lack checksum metadata.

## APIs

T2's own possible new `Stat` variant or `RepairChecksum` method is the primary API-surface addition.

## Configuration changes

None anticipated beyond T2's own size/time-bound configuration for the repair path.

## Persistence changes

T4's own inventory mechanism may require new persisted state (exact shape TBD).

## Migration strategy

If T4's inventory mechanism requires a new table, follow DATA-09's own online-migration protocol
(W02-E01) if applied to a live shared table.

## Concurrency implications

T4's own resumable backfill must be safe under concurrent invocation (interrupt-and-resume, per its own
acceptance criterion) — no duplicate work.

## Error-handling strategy

T1's own checksum-required enforcement must fail clearly if a call site cannot persist checksum
metadata, not silently proceed without it.

## Security controls

T1's own universal checksum enforcement is itself a data-integrity control.

## Observability changes

T3's own dedicated fallback-invocation metrics; the CS-16 expansion's own benchmark additions make
regressions in the 7 named packages visible for the first time.

## Testing strategy

- T1: integration test, upload via framework path, `Stat`, assert no `GetObject` call.
- T2: test that a legacy object triggers the fallback only via the labeled repair path.
- T3: metric-emission test.
- T4: interrupt/resume backfill test, no duplicate work, eventual completion.
- T5: before/after comparison against `perf/reference-v1.json`.
- CS-16: `make bench-budget` passing with all 7 new entries present.

## Regression strategy

T1's own integration test becomes the ongoing regression guard against a future upload path silently
skipping checksum persistence. The CS-16 expansion's own 7 new benchmarks become the ongoing regression
guard for their respective hot paths, per PERF-06's own fail-closed enforcement.

## Compatibility strategy

T2's own port-API-surface change must preserve compatibility for any other adapter implementing
`storage.ObjectInfo`, per PLAN's own risk note.

## Rollout strategy

T1 → T2 → T3, T4 in parallel → T5 (PERF-05's own sequence). The CS-16 expansion's own 7 benchmarks may
be added independently and in any order, each landing with its own budget entry in the same PR.

## Rollback strategy

If T2's port-API-surface change breaks another adapter, revise the API shape to preserve compatibility
rather than accepting the breakage; if a CS-16 benchmark proves too noisy for a stable budget, diagnose
and stabilize per systematic-debugging discipline before accepting a wide, low-value budget range.

## Implementation sequence

PERF-05's T1-T5 in the sequence above; the CS-16 expansion's 7 benchmarks may proceed in parallel with
PERF-05's own work, since the two finding sets target disjoint code (storage vs. 7 kernel packages).

## Task breakdown

- **W07-E01-S004-T001** — Required checksum enforcement + call-site audit (T1).
- **W07-E01-S004-T002** — Bounded repair path (T2).
- **W07-E01-S004-T003** — Fallback-invocation metrics (T3).
- **W07-E01-S004-T004** — Resumable async backfill (T4).
- **W07-E01-S004-T005** — Publication against `perf/reference-v1.json` (T5).
- **W07-E01-S004-T006** — CS-16 bench-coverage expansion (7 packages).
- **W07-E01-S004-T007** — Independent review.

## Expected artifacts

The checksum-required enforcement + call-site audit; the bounded repair path; fallback-invocation
metrics; the resumable backfill mechanism; the published comparison report; 7 new benchmark files with
bench-budget entries.

## Expected evidence

Integration test output (no `GetObject` on normal `Stat`); labeled-repair-path test output; metric-
emission test output; interrupt/resume backfill test output; the published comparison report; `make
bench-budget` passing output including all 7 new entries.

## Unresolved questions

- T2's own exact `storage.ObjectInfo` port API-surface decision (new `Stat` variant vs. `RepairChecksum`
  method).
- T4's own legacy-object inventory mechanism (does not obviously exist yet, per PLAN's own risk note).
- The `workflow` package's own exact benchmark target (MATRIX CS-16 names it less precisely than the
  other 6 packages).

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned.
