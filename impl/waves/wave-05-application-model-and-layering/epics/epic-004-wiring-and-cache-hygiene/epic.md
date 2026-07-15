---
id: W05-E04
type: epic
title: Wiring and cache hygiene
status: planned
wave: W05
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-06
  - SEC-04
depends_on: []
stories:
  - W05-E04-S001
  - W05-E04-S002
decisions:
  - D-06
risks:
  - RISK-W05-005
---

# W05-E04 — Wiring and cache hygiene

## Epic objective

Close the remaining hidden constructor-bypass surface in `kernel/kernel.go` (AR-06 remainder) and
bound the framework's authorization cache with epoch-based cross-pod invalidation, resolving what
PLAN's own risk column names SEC-04 T4's "Highest-risk task" open architecture decision via D-06
(SEC-04, per CS-17's consolidated closure spec).

## Problem being solved

`requirement-inventory.md` row AR-06: "Remove hidden constructor bypasses | IMPL | P1 | partial |
W05-E04-S001 | T1 EXECUTED; T2 lint + T3 audit planned." Row SEC-04: "Bound authz staleness/memory |
IMPL | P1 | planned | W05-E04-S002 | CS-17: LRU (approved dep) + epoch table (D-06); P0 if cache
prod-enabled." PLAN's own AR-06 evidence: `kernel/kernel.go:252-254`'s `orgAncestry` closure
originally called `authz.NewStore()` a second time instead of closing over the composed
`authzStore` — already fixed (T1, executed). T2 (a lint rule) and T3 (an audit) remain. MATRIX
CS-17's own evidence for SEC-04: `kernel/authz/caching.go:29-36`'s plain `map[string]
cachedAssignments`+mutex cache is unbounded, with `Invalidate*` methods existing but having exactly
one production caller repo-wide, and `kernel/kernel.go:118-121` documenting grant/revoke
invalidation as "a product-owned obligation" — a correctness-by-convention gap this epic closes
structurally.

## Scope

- The constructor-boundary lint tool: fails CI on any reintroduced ad hoc infrastructure constructor
  outside composition packages (S001, PLAN AR-06 T2).
- The `kernel/kernel.go` audit confirming or refuting whether the fixed closure-captures-a-fresh-
  instance pattern is isolated to the one already-fixed line (S001, PLAN AR-06 T3).
- The bounded, sharded `golang-lru`-backed authz cache replacing the unbounded map (S002, MATRIX
  CS-17 T1 / PLAN SEC-04 T1).
- Eviction with admission/eviction metrics (S002, PLAN SEC-04 T2).
- Singleflight-collapse of concurrent misses (S002, PLAN SEC-04 T3).
- Per-tenant authorization epoch table for cross-pod revocation, resolving MATRIX CS-17's "highest-
  risk" open architecture decision via D-06 (S002, PLAN SEC-04 T4, MATRIX CS-17 T2).
- `CacheHit`/epoch-observed decision-provenance metadata (S002, PLAN SEC-04 T5).
- Prod-config gating: explicit max-size + stale-allow bound required when the cache is enabled in
  `prod`, boot fails without both (S002, PLAN SEC-04 T6).

## Out of scope

- **AR-06 T1 (the `orgAncestry` closure fix itself)** — already executed. Not re-planned here.
- **DATA-07's own relationship-semantics work** — W03-E04's scope, already landed by this wave's own
  entry gate (W03-E01 acceptance); this epic's S002 closes DATA-07 T4's cache-invalidation
  acceptance criterion (an AC-level closure relationship, not a re-implementation of DATA-07 itself).
- **SEC-01's own grant-table mutation paths adopting the epoch bump** — MATRIX CS-17's own
  cross-CS sequencing note states SEC-01's new mutation paths must adopt the epoch bump "as part of
  their own acceptance," not this epic's own implementation responsibility (SEC-01 is W03-E01 scope,
  already landed).

## Source requirements

AR-06 (T2, T3 — T1 already executed). SEC-04 (T1-T6). D-06 (referenced, not authored — ratified in
W00-E02-S003).

## Architectural context

This epic's two stories are independent of each other — AR-06's constructor-bypass closure
(`kernel/kernel.go`'s composition-root discipline) and SEC-04's authz-cache bounding
(`kernel/authz/caching.go`'s bounded-and-invalidated cache) are disjoint concerns with no
task-level dependency between them. Neither depends on W05-E01, W05-E02, or W05-E03 — AR-06's
remainder and SEC-04's full scope are independent of the ownership-model rework this wave's earlier
epics deliver. SEC-04 T4's own "Highest-risk task" status (PLAN's own risk column) is resolved at
the architecture-decision level by D-06 (per-tenant epoch table, polled; LISTEN/NOTIFY optional
only) — this epic's S002 references, not re-decides, D-06.

## Included stories

- **W05-E04-S001 — constructor-bypass-closure** (PLAN AR-06 T2, T3): the constructor-boundary lint
  and the `kernel/kernel.go` audit.
- **W05-E04-S002 — authz-cache-bounding** (PLAN SEC-04 T1-T6, per CS-17: `golang-lru` + epoch table
  D-06): the bounded cache, eviction, singleflight, cross-pod epoch invalidation, decision
  provenance, and prod-config gating — closing DATA-07 T4's cache-invalidation AC.

## Dependencies

No dependency on any other W05 epic. Depends only on this wave's own entry gate (W03-E01
acceptance, since SEC-04 T4's epoch-bump wiring extends to SEC-01's grant-table mutation paths per
MATRIX CS-17's own cross-CS note, and SEC-01 is W03-E01 scope).

## Risks

RISK-W05-005 (SEC-04 T4's epoch-bump wiring completeness across every framework-side mutation path)
originates at wave scope and lands entirely within this epic's S002. See `risks.md` for the
epic-scoped elaboration.

## Required decisions

D-06 (per-tenant `authz_epoch` table, polled; LISTEN/NOTIFY optional only) — already ratified in
W00-E02-S003, referenced (not re-decided) in S002's `decisions/index.md`.

## Epic acceptance criteria

- **AC-W05-E04-01**: The constructor-boundary lint fails CI on any reintroduced ad hoc infrastructure
  constructor outside composition packages; the `kernel/kernel.go` audit explicitly confirms or
  refutes, with evidence, that the closure-captures-a-fresh-instance pattern is isolated to the one
  already-fixed line.
- **AC-W05-E04-02**: The authorization cache never exceeds its configured maximum under adversarial
  cardinality; idle entries are evicted with a full admission/eviction metric set; N concurrent
  misses collapse to one DB load via singleflight; a simulated cross-pod revocation is visible on the
  second pod without a full TTL wait, via the D-06 per-tenant epoch table; `Decision` metadata
  distinguishes cache-hit from cache-miss/epoch-observed; a `prod` profile with the cache enabled but
  no explicit max-size/stale-allow bound fails boot validation. This epic's own AC explicitly closes
  DATA-07 T4's cache-invalidation acceptance criterion per `impl/analysis/wave-allocation-detail.md`.
- **AC-W05-E04-03**: All stories have passed independent review per mandate §14, with S002
  specifically checked given SEC-04 T4's own "Highest-risk task" status and the epoch-bump
  wiring-completeness concern.

## Closure conditions

Both stories reach `accepted`; AC-W05-E04-01 through AC-W05-E04-03 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date; the
DATA-07 T4 cache-invalidation AC-closure relationship is confirmed recorded (not silently dropped) at
closure.
