---
id: W04-E03-DEPS
type: epic-dependencies
epic: W04-E03
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W04-E01 (this wave) — story-scoped, not whole-epic.** `S002` (DATA-04 T2–T6) depends on
  `W04-E01-S001` (the shared lease/fencing primitive); PLAN DATA-04 T2's own dependency column
  states "DATA-02 T1; T1 as interim" — meaning S002 depends on both the shared primitive
  (`W04-E01-S001`) landing AND on this epic's own S001 (T1, the stopgap) as an interim measure
  bridging the gap until T2's lease columns exist. `S001` (T1, the stopgap) has **no** dependency on
  `W04-E01` at all — PLAN DATA-04 T1's own dependency column is "—", and
  `impl/analysis/wave-allocation-detail.md`'s W04-E03 row confirms: "S001 T1 stopgap (can start at
  wave entry)."
- **W04-E01-S002** — S002's T4 (item idempotency keys, finalize fencing, retry policy, cancellation)
  explicitly shares finalize-fencing logic with DATA-02 T3, which is covered by `W04-E01-S002`. PLAN
  DATA-04 T4's own risk note: "Shares finalize-fencing logic with DATA-02 T3 — reuse, don't
  reimplement." This is a reuse dependency, not merely a "builds on" relationship — T4 must not
  design its own independent fencing scheme.
- **W04-E01-S003** — S002's T6 (the named chaos test `DATA-04/chaos/duplicate_worker_test.go`)
  explicitly reuses the shared chaos harness built in `W04-E01-S003` (DATA-02 T7: "build as a
  reusable chaos harness shared with DATA-03/DATA-04"). This is a reuse dependency — T6 must not
  reimplement or redesign a chaos harness.
- **W00** (full wave) — per `../../dependencies.md` (wave-level), this wave's own entry is gated on
  W00's exit criteria. This epic has no additional upstream dependency beyond that gate for S001;
  S002's dependency on `W04-E01` is the one described above.

## Downstream (epics/waves that depend on this epic)

None confirmed. Neither `impl/index.md`'s wave map nor `wave-allocation-detail.md`'s cross-wave
sequencing notes name any epic or wave depending on W04-E03. Per the source's own wowsociety-impact
note for DATA-04 ("Not affected. Zero `kernel/bulk` import anywhere in wowsociety"), there is no
downstream product-side dependency either.

## Internal (within this epic)

S001 → S002 is a partial, not strict, sequencing dependency: S001 (T1) may complete and close
independently of S002, since S002's real gating dependency is `W04-E01-S001` (the shared primitive),
not S001 itself. S001 is consumed by S002 only as an *interim* measure — PLAN DATA-04 T2's
dependency column lists both "DATA-02 T1" and "T1 as interim," meaning S002's T2 work can begin
once either (a) `W04-E01-S001` has landed, using S001's stopgap as a bridge until then, or (b) S001
has landed and DATA-02 T1 has not yet, in which case S002's T2 is blocked on `W04-E01-S001` alone.
In no case does S002 depend on S001 having reached `accepted` status before S002 can start planning
or begin T2 design work — but S002's T2 code (the lease-column migration) supersedes S001's advisory
lock/CAS stopgap at the point T2 lands, per the same supersession pattern this wave uses elsewhere
(`W04-E01-S001` superseding `W02-E01-S002`'s interim checkpoint lease, per `wave.md` "Assumptions").

Within S002 itself: T2 → T3 → T4 → T5 form a dependency chain (T3 depends on T2's lease columns;
T4 depends on T3's atomic claim path existing before fencing/idempotency can be layered on top; T5
depends on T3/T4's claim and fencing paths before lifecycle controls can safely pause/resume/cancel
mid-run). T6 (the named chaos test) depends on T3, T4, and T5 all being in place, per PLAN's own
task table ("T3-T5" in T6's dependency column).

## Cross-wave dependencies

None. DATA-04 has no dependency on any wave other than this one's own internal `W04-E01` edge —
confirmed by `requirement-inventory.md`'s notes column for DATA-04, which cites no cross-wave
dependency, and by `wave.md`'s own "Dependencies" section, which names only `W04-E04-S001` as
carrying the wave's one W02 dependency.

## External dependencies

None new. This epic operates entirely within the existing `kernel/bulk` package and the existing
PostgreSQL/pgx persistence layer already used elsewhere in the framework.

## Repository dependencies

None cross-repo for this epic's own closure. Per DATA-04's own wowsociety-impact note (reproduced
in `epic.md` "Out of scope"): "Not affected. Zero `kernel/bulk` import anywhere in wowsociety." No
coordination is required for this epic's closure.

## Tooling dependencies

None beyond the already-available Go/PostgreSQL toolchain and the shared chaos-test infrastructure
built in `W04-E01-S003`, which this epic's S002-T6 consumes rather than extends.

## Decision dependencies

None. See `epic.md` "Required decisions."
