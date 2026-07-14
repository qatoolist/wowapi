---
id: W04-E01-DEPS
type: epic-dependencies
epic: W04-E01
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W04's own entry gate** (`../../wave.md`) — the strict W00→W07 wave-entry ordering, satisfied by
  W00's exit criteria. Per `../../dependencies.md`, DATA-02 (this epic) has **no** dependency on
  W02's online-migration protocol — that dependency is narrowly scoped to W04-E04-S001 (DATA-08
  W6-T1) only. This epic may enter as soon as W04's own entry gate is satisfied, independent of
  W02-E01's acceptance status.
- **W02-E01-S002** (informational, not blocking) — this epic's S001 supersedes W02-E01-S002's
  interim checkpoint lease. This is a downstream-replaces-upstream-artifact relationship, not a
  dependency that gates this epic's own entry: S001 does not need W02-E01-S002 to still be
  `in-progress` or unaccepted to begin; it needs W02-E01-S002's interim-lease code and any
  checkpoint state it may have written to exist, so the migration step (RISK-W04-001) has something
  concrete to migrate.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W04-E02 (DATA-03, remote-io-outside-tx) | W04-E01-S001 (the primitive itself) | `../../dependencies.md`: "W04-E02, W04-E03 depend on W04-E01 for the shared lease/fencing primitive and the shared chaos harness." DATA-03 T1's own dependency column states "DATA-02 T1." |
| W04-E03 (DATA-04, bulk-multi-worker-safety) | W04-E01-S001 (primitive) for its full rewrite (S002); no dependency for its S001 stopgap | DATA-04 T2's dependency column states "DATA-02 T1; T1 as interim" — the stopgap (E03-S001) may land before this epic's primitive exists, but the full leased-claim rewrite (E03-S002) depends on it. |
| W04-E02-S001/S002/S003, W04-E03-S002 (chaos work) | W04-E01-S003 (the chaos harness, T7) | `wave-allocation-detail.md`: "S003 idempotency-and-chaos (T5, T6, T7 chaos harness — harness shared with E02/E03)." DATA-03's 6-boundary chaos test and DATA-04's chaos test reuse — not reimplement — this epic's harness. |
| W02-E01-S002 (interim checkpoint lease, forward reference) | W04-E01-S001 (supersession) | This epic's S001 is the item RISK-W02-001 in `../../../wave-02-data-safety-and-migration-tooling/risks.md` forward-references as the eventual replacement for W02-E01-S002's interim lease; the dependency direction is this epic providing the replacement, not depending on W02-E01-S002 remaining unaccepted. |

## Internal (within this epic)

S001 → S002 → S003 form a strict build-sequence dependency, matching PLAN DATA-02's own task
dependency chain (T2, T3, T4 all depend on T1; T5 depends on T2; T6 depends on T3, T5; T7 depends on
T3–T5). Concretely:

- S002 depends on S001 (T2/T3/T4 all require T1's primitive to exist before `jobs_queue` can carry
  lease columns assigned and compared against it).
- S003 depends on S002 (T5's idempotency contract and T6's effect-ledger test both require S002's
  fenced finalize paths to exist as the boundary they test against; T7's chaos test exercises the
  full claim→stall→reclaim→finalize chain S002 built).

Unlike W01-E01's three internally-parallel stories, this epic's three stories are **not**
independently orderable — they must execute S001 → S002 → S003 in sequence, because each stage's
tooling is a genuine prerequisite for the next stage to have anything to fence.

## Cross-wave dependencies

W04-E01-S001 supersedes W02-E01-S002's interim checkpoint lease (RISK-W04-001, mirroring
RISK-W02-001) — see "Downstream" above. No other cross-wave dependency exists for this epic.

## External dependencies

None new. This epic's primitive operates on the existing PostgreSQL/pgx toolchain already used by
`kernel/jobs`. No new external service, queue, or third-party library is introduced.

## Repository dependencies

None cross-repo for this epic's own closure. PLAN's own wowsociety-impact note for DATA-02: "Not
affected. Zero `kernel/jobs` import, zero job registration anywhere in wowsociety. Would become
breaking (worker signature change, T5) the moment wowsociety registers a job — flag for roadmap."
Tracked as a forward-looking coordination note in S003's `plan.md`, not a blocking dependency today.

## Tooling dependencies

None beyond the already-available Go/PostgreSQL toolchain. S003-T7's chaos harness extends the
existing Go test toolchain rather than introducing a new one.

## Decision dependencies

None. See `epic.md` "Required decisions."
