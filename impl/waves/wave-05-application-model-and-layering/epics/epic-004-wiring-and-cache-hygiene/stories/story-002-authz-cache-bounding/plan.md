---
id: PLAN-W05-E04-S002
type: plan
parent_story: W05-E04-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E04-S002

Per mandate §8.5. D-06 is treated as confirmed, ratified fact; this plan implements the concretized
CS-17 closure spec (`golang-lru/v2` swap + per-tenant epoch table), not a re-derivation of the
architecture decision itself.

## Proposed architecture

`kernel/authz/caching.go`'s unbounded map is replaced with a `hashicorp/golang-lru/v2`-backed
bounded, sharded cache, sized by config. A new `authz_epoch` table holds one row per tenant (or a
global row, per D-06's own framing "per-tenant epoch integer"), incremented by every framework-side
mutation path in the same transaction as the mutation itself. The authz read path checks the epoch
on read; a stale local cache entry (epoch mismatch) is treated as a miss. Existing
`Invalidate*` methods remain for product-triggered cases. The TTL floor remains as a backstop.

## Implementation strategy

1. Re-read `kernel/authz/caching.go` at this story's start commit to confirm the current unbounded
   state.
2. Swap the map+mutex for `golang-lru/v2`, sized by config.
3. Write `SEC-04/bounded-cache-tests.md`'s producing tests: insert >max keys, confirm bound never
   exceeded; race test.
4. Implement eviction with a full admission/eviction metric set.
5. Write `SEC-04/eviction-metrics-tests.md`'s producing test.
6. Implement singleflight-collapse of concurrent misses.
7. Write `SEC-04/singleflight-tests.md`'s producing test: N concurrent misses → 1 DB load.
8. Design and migrate the `authz_epoch` table (per-tenant epoch integer).
9. Enumerate every known framework-side mutation path: role/permission assignment writes in
   `kernel/authz`, `kernel/seeds/seeds.go`, and SEC-01's grant-table writes (landed by this wave's
   entry gate).
10. Wire an epoch bump into each enumerated mutation path, in the same transaction as the mutation.
11. Wire the epoch check into the authz read path.
12. Write `SEC-04/cross-pod-epoch-tests.md`'s producing simulated cross-pod test.
13. Expose `CacheHit`/epoch-observed on `Decision`.
14. Write `SEC-04/decision-provenance-tests.md`'s producing test.
15. Add prod-config validation: explicit max-size + stale-allow bound required when cache enabled,
    boot fails without both.
16. Write `SEC-04/prod-config-gate-tests.md`'s producing negative config test.
17. Document all six properties, referencing D-06 for the epoch-table design.

## Expected package or module changes

`kernel/authz/caching.go`; a new `authz_epoch` migration; the mutation-path files across
`kernel/authz`, `kernel/seeds`, and SEC-01's grant-table write path; `config.go`.

## Expected file changes where determinable

`kernel/authz/caching.go` (rewritten cache backend); a new migration file for `authz_epoch`;
mutation-path files enumerated in step 9; `config.go` (prod validation extension); new test files as
named above (T1-T6).

## Contracts and interfaces

`Decision`'s own type, extended with `CacheHit`/epoch-observed fields (T5). The cache's own internal
`Store` interface, per MATRIX CS-17's own note: "swap behind existing `Store` interface" — this is a
confirmed existing seam to reuse, not a new interface this story invents.

## Data structures

The `authz_epoch` table: one row per tenant (or a global row), an incrementing integer epoch column.

## APIs

None externally facing — internal cache/authz-read-path change.

## Configuration changes

The prod-config gate (T6) requires new config keys for max-size and stale-allow bound, per MATRIX
CS-17's own note that "an established pattern already exists in `config.go`" for this kind of
prod-required-bound validation.

## Persistence changes

The new `authz_epoch` table (migration).

## Migration strategy

Author the `authz_epoch` migration per this programme's existing migration convention at this
story's own implementation time (using W02-E01-S001's manifest schema if landed, or the framework's
pre-DATA-09 convention otherwise — recorded as a note if a divergence occurs, not silently decided).

## Concurrency implications

The epoch bump must occur in the same transaction as the triggering mutation (MATRIX CS-17's own
explicit instruction: "framework-side mutation paths bump the epoch in the same tx"), so a reader
never observes a mutation without its corresponding epoch bump. The bounded cache itself must be
safe for concurrent reads/writes from multiple goroutines (a property `golang-lru/v2` provides
natively, subject to this story's own confirmation via the race test).

## Error-handling strategy

The prod-config gate's failure must name the missing bound(s) specifically (max-size, stale-allow,
or both), consistent with this programme's field-specific-error-message convention.

## Security controls

The epoch-table mechanism (T4) and the prod-config gate (T6) are both required security controls,
per this story's own "Security considerations."

## Observability changes

T2's eviction metrics and T5's decision-provenance metadata are both required observability
additions.

## Testing strategy

- `SEC-04/bounded-cache-tests.md`: insert >max keys, race test.
- `SEC-04/eviction-metrics-tests.md`: idle-entry eviction, full metric set.
- `SEC-04/singleflight-tests.md`: N concurrent misses → 1 DB load.
- `SEC-04/cross-pod-epoch-tests.md`: simulated cross-pod revocation visibility, exercising every
  enumerated mutation path.
- `SEC-04/decision-provenance-tests.md`: `Decision` metadata hit vs. miss/epoch-observed.
- `SEC-04/prod-config-gate-tests.md`: negative config test, prod + cache-enabled + missing bound(s).

## Regression strategy

All six named tests are permanent regression guards. The cross-pod-epoch test specifically guards
against a future framework-side mutation path being added without its own epoch bump — though this
requires the test to be extended whenever a new mutation path is added, a maintenance obligation this
story's own documentation should call out explicitly.

## Compatibility strategy

Existing `Invalidate*` methods remain for product-triggered cases, per MATRIX CS-17's own explicit
instruction — no removal, only an additive structural mechanism alongside them.

## Rollout strategy

Single story, landed as its own reviewable unit.

## Rollback strategy

Revert the epoch-table mechanism (T4) if the cross-pod test reveals a missed mutation path or a
transaction-boundary bug (epoch bump not atomic with the mutation) — escalate for a fix rather than
shipping a partially-correct epoch mechanism, since a partial fix could create false confidence in a
security-adjacent control.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-17). T1 (steps 2-3) is independent and can
land first, per MATRIX CS-17's own note: "T1 (LRU swap) is independent, land any time." T4 (steps
8-12) is the highest-risk item and should receive the most implementation care, given PLAN's own
"Highest-risk task" framing — even though D-06 resolves its architecture-decision component.

## Task breakdown

- **W05-E04-S002-T001** — Bounded, sharded cache (T1; steps 2-3 above).
- **W05-E04-S002-T002** — Eviction with admission/eviction metrics (T2; steps 4-5 above).
- **W05-E04-S002-T003** — Singleflight-collapse of concurrent misses (T3; steps 6-7 above).
- **W05-E04-S002-T004** — Per-tenant authz_epoch table and cross-pod epoch-bump wiring, enacting D-06
  (T4; steps 8-12 above).
- **W05-E04-S002-T005** — Decision provenance metadata and prod-config gate (T5, T6; steps 13-16
  above — grouped together given both are Low-risk, small, closely-related "expose state correctly"
  concerns per PLAN's own risk column).
- **W05-E04-S002-T006** — Independent review (per mandate §14, scoped to this story, given T4's
  "Highest-risk task" status and the DATA-07 T4 cross-wave AC-closure relationship).

## Expected artifacts

The bounded cache (code); eviction metrics (code); singleflight collapse (code); the epoch table and
wiring (code + migration); decision-provenance metadata (code); the prod-config gate (code).

## Expected evidence

The six named test-report outputs.

## Unresolved questions

- Exact `authz_epoch` table shape (per-tenant row vs. a single global row with per-tenant columns) —
  D-06's own text says "per-tenant epoch integer in a small `authz_epoch` table," which this plan
  reads as one row per tenant, but the exact schema (indexing, additional metadata columns) is this
  story's own implementation-time design work.
- Exact prod-config key names for max-size and stale-allow bound — to be chosen consistent with
  `config.go`'s existing naming conventions.

## Approval conditions

This plan is approved for implementation once: (a) the `authz_epoch` table schema is drafted, and
(b) the owner and reviewer are assigned.
