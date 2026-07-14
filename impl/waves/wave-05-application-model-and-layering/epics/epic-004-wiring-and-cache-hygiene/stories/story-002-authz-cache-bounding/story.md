---
id: W05-E04-S002
type: story
title: Bounded, epoch-invalidated authorization cache
status: planned
wave: W05
epic: W05-E04
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - SEC-04
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W05-E04-S002-01
  - AC-W05-E04-S002-02
  - AC-W05-E04-S002-03
  - AC-W05-E04-S002-04
artifacts: []
evidence: []
decisions:
  - D-06
risks:
  - RISK-W05-005
---

# W05-E04-S002 — Bounded, epoch-invalidated authorization cache

## Story ID

W05-E04-S002

## Title

Bounded, epoch-invalidated authorization cache

## Objective

Replace the framework's unbounded authorization cache with a `golang-lru`-backed bounded, sharded
cache; add eviction with admission/eviction metrics; collapse concurrent misses via singleflight;
resolve cross-pod revocation staleness via a per-tenant `authz_epoch` table (D-06); expose
cache-hit/epoch-observed decision provenance; and require an explicit max-size + stale-allow bound
in `prod` config, failing boot without both.

## Value to the framework

This story closes SEC-04, graded P0 if the cache is enabled in production, and resolves what PLAN's
own risk column names "Highest-risk task" (T4, cross-pod invalidation) via D-06's ratified
architecture decision. It also closes DATA-07 T4's own cache-invalidation acceptance criterion, per
this wave's own cross-wave closure relationship — one story doing double duty for two requirement
IDs.

## Problem statement

`requirement-inventory.md` row SEC-04: "Bound authz staleness/memory | IMPL | P1 | planned |
W05-E04-S002 | CS-17: LRU (approved dep) + epoch table (D-06); P0 if cache prod-enabled." MATRIX
CS-17's own evidence: `kernel/authz/caching.go:29-36` is "a plain `map[string]cachedAssignments`
+mutex, unbounded, no LRU," with key `tenantID+"|c:/u:/s:"+id`, TTL default 1s;
`Invalidate`/`InvalidateTenant`/`InvalidateAll` exist but have "exactly one production caller
repo-wide (`kernel/seeds/seeds.go:278`, `InvalidateAll`)"; `kernel/kernel.go:118-121` documents
grant/revoke invalidation as "a product-owned obligation" — the framework performs the mutation but
delegates the cache consequence to the product's own memory. PLAN's own SEC-04 task table: T1 —
"Bounded, sharded cache | none | Never exceeds configured max under adversarial cardinality | Insert
>max keys; race test | `SEC-04/bounded-cache-tests.md` | Low-moderate — swap behind existing `Store`
interface." T2 — "Eviction with admission/eviction metrics | T1 | Idle entries evicted; full metric
set | Test | `SEC-04/eviction-metrics-tests.md` | Low." T3 — "Singleflight-collapse concurrent
misses | T1 | N concurrent misses → 1 DB load | Test | `SEC-04/singleflight-tests.md` | Low." T4 —
"Per-tenant/global authorization epoch or invalidation stream for cross-pod revocation | T1-T3 |
Revocation on pod A visible to pod B without a full TTL wait | Simulated cross-pod test |
`SEC-04/cross-pod-epoch-tests.md` | Highest-risk task — open architecture decision (LISTEN/NOTIFY
vs. epoch-row-poll), may overlap Wave 3's shared lease infrastructure." T5 — "Expose
`CacheHit`/epoch-observed on `Decision` | T1-T4 | Decision metadata differs hit vs. miss | Test |
`SEC-04/decision-provenance-tests.md` | Low." T6 — "Require explicit max-size + stale-allow bound in
prod config; fail boot without both | T1-T5 | Prod profile with cache enabled but no bound fails
validation | Negative config test | `SEC-04/prod-config-gate-tests.md` | Low — established pattern
already exists in `config.go`." MATRIX CS-17 concretizes: "T1 replace map with
`hashicorp/golang-lru/v2` (approved §L) sized by config; T2 per-tenant epoch column (D-06) checked
on read — framework-side mutation paths bump the epoch in the same tx, making invalidation
structural, not conventional; `Invalidate*` methods stay for product-triggered cases. TTL floor
stays as backstop."

## Source requirements

SEC-04 (T1-T6). D-06 (referenced, not authored — ratified in W00-E02-S003).

## Current-state assessment

Per MATRIX CS-17's own evidence, the cache today is unbounded (plain map + mutex), with a 1-second
default TTL as the only staleness bound, and cross-pod invalidation delegated entirely to the
product's own memory (an undocumented-in-practice obligation, since `kernel/kernel.go:118-121`
documents it but no enforcement exists). This story's own re-confirmation step is to re-read
`kernel/authz/caching.go` at this story's actual start commit and confirm this unbounded,
convention-dependent state still holds.

## Desired state

The cache is bounded via `golang-lru/v2`, sized by config, never exceeding its configured maximum
under adversarial cardinality. Idle entries are evicted with a full admission/eviction metric set.
N concurrent misses collapse to one DB load via singleflight. A per-tenant `authz_epoch` table
(D-06) is checked on read; framework-side mutation paths (role/permission assignment writes in
`kernel/authz`, seeds, and — since this wave enters after W03-E01 acceptance — SEC-01's grant-table
writes) bump the epoch in the same transaction, making cross-pod invalidation structural rather than
conventional. `Decision` metadata distinguishes cache-hit from cache-miss/epoch-observed. A `prod`
profile with the cache enabled but no explicit max-size/stale-allow bound fails boot validation.

## Scope

- The bounded, sharded `golang-lru/v2`-backed cache, sized by config (T1).
- Eviction with a full admission/eviction metric set (T2).
- Singleflight-collapse of concurrent misses (T3).
- The per-tenant `authz_epoch` table and epoch-bump wiring into every framework-side mutation path
  (role/permission assignment writes in `kernel/authz`, seeds, and SEC-01's grant-table writes) —
  resolving SEC-04 T4's "Highest-risk task" open architecture decision via D-06 (T4).
- `Decision` metadata exposing `CacheHit`/epoch-observed (T5).
- Explicit max-size + stale-allow bound required in `prod` config, boot fails without both (T6).
- The explicit acceptance criterion closing DATA-07 T4's cache-invalidation requirement.

## Out of scope

- **DATA-07's own relationship-semantics implementation** — W03-E04's scope, already landed by this
  wave's entry gate; this story closes DATA-07 T4's own AC by ID, it does not modify DATA-07's own
  files.
- **SEC-01's own grant-table mutation-path implementation** — W03-E01's scope, already landed; this
  story's T4 wires the epoch bump into SEC-01's existing grant-table write paths, it does not
  reimplement SEC-01 itself.
- **A message bus or LISTEN/NOTIFY-based invalidation transport as the primary correctness
  mechanism** — explicitly rejected per REVIEW §M ("custom message bus" rejected) and D-06's own
  resolution ("Postgres `LISTEN/NOTIFY` as an optional latency optimisation, not a correctness
  dependency"). This story implements the epoch-poll mechanism as the correctness mechanism;
  LISTEN/NOTIFY, if added at all, is an optional future latency optimization, not this story's
  required scope.

## Assumptions

- D-06's resolution ("per-tenant epoch integer in a small `authz_epoch` table, polled on the
  existing authz read path; Postgres `LISTEN/NOTIFY` as an optional latency optimisation, not a
  correctness dependency... avoids a new message bus in the kernel") is taken as ratified fact from
  `docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F item 7, referenced (not
  re-derived) via this story's `decisions/index.md`. This closes SEC-04 T4's own "open architecture
  decision" framing — this story's own planning treats the decision as resolved, not still open.
- MATRIX CS-17's own cross-CS sequencing note ("T2's epoch bumps must be added to the framework
  mutation paths that exist today ... and extended to SEC-01's grant table when it lands ... SEC-01's
  new mutation paths must adopt the epoch bump as part of their own acceptance") is taken as a
  confirmed source instruction — this story's own T4 task enumerates SEC-01's grant-table writes as
  one of its own mutation-path targets, since W05 enters only after W03-E01 (SEC-01) has already
  landed, per this wave's own entry gate.

## Dependencies

None within W05-E04 (independent of S001). Depends only on this wave's own entry gate (W03-E01
acceptance, since T4's epoch-bump wiring extends to SEC-01's grant-table mutation paths).

**Downstream cross-wave AC-closure relationship**: DATA-07 T4 (W03-E04-S001, already landed) —
per `impl/analysis/wave-allocation-detail.md`'s explicit note: "DATA-07 T4 cache-invalidation AC
closes here." This story's own AC-W05-E04-S002-04 records this relationship by ID; no DATA-07 file
is modified by this story.

## Affected packages or components

`kernel/authz/caching.go` (bounded cache, eviction, singleflight); a new `authz_epoch` table
migration; the epoch-bump wiring across `kernel/authz`'s role/permission assignment writes, seeds
(`kernel/seeds/seeds.go`), and SEC-01's grant-table writes; `config.go` (prod-config gate).

## Compatibility considerations

The existing `Invalidate`/`InvalidateTenant`/`InvalidateAll` methods stay for product-triggered
cases, per MATRIX CS-17's own explicit instruction — this story does not remove them, only adds the
structural epoch-based mechanism alongside them. wowsociety's own impact is explicitly
"strictly safer" per SEC-04's own wowsociety-impact note: "removes an undocumented obligation."

## Security considerations

This story's entire purpose is a security-adjacent correctness control: bounding cache memory
(preventing unbounded growth under adversarial tenant×principal cardinality) and closing the
cross-pod stale-authorization window (a revoked grant remaining effectively active on another pod
until a TTL expires). The prod-config gate (T6) is itself a required security control, not optional
hardening — an established pattern per PLAN's own note "already exists in `config.go`."

## Performance considerations

The singleflight collapse (T3) is itself a performance control (N concurrent misses → 1 DB load).
The epoch-table read (T4) adds a per-read check against the `authz_epoch` table — this story's own
implementation should keep this check cheap (e.g. an indexed, small-table lookup), consistent with
the authz read path's existing performance sensitivity.

## Observability considerations

T2's eviction metrics and T5's `CacheHit`/epoch-observed decision provenance are both required
observability additions, not optional.

## Migration considerations

The new `authz_epoch` table requires a migration — this story's own migration should be authored
per this programme's existing migration-manifest discipline (W02-E01-S001's own schema, if landed by
this point) or, if not yet landed, per the framework's existing pre-DATA-09 migration convention,
with the divergence recorded as a note if relevant.

## Documentation requirements

Document the bounded cache's configuration surface, the epoch-table mechanism and its role in
cross-pod invalidation (referencing D-06), the decision-provenance metadata, and the prod-config
gate's required fields.

## Acceptance criteria

- **AC-W05-E04-S002-01**: The cache never exceeds its configured maximum under adversarial
  cardinality (insert >max keys; race test) — proven by `SEC-04/bounded-cache-tests.md`. Idle entries
  are evicted with a full admission/eviction metric set — proven by `SEC-04/eviction-metrics-tests.md`.
- **AC-W05-E04-S002-02**: N concurrent misses collapse to 1 DB load via singleflight — proven by
  `SEC-04/singleflight-tests.md`. A simulated cross-pod revocation is visible on the second pod
  without a full TTL wait, via the D-06 per-tenant `authz_epoch` table, with epoch bumps wired into
  every enumerated framework-side mutation path (role/permission assignment writes in `kernel/authz`,
  seeds, SEC-01's grant-table writes) — proven by `SEC-04/cross-pod-epoch-tests.md`.
- **AC-W05-E04-S002-03**: `Decision` metadata differs cache-hit vs. cache-miss/epoch-observed —
  proven by `SEC-04/decision-provenance-tests.md`. A `prod` profile with the cache enabled but no
  explicit max-size/stale-allow bound fails boot validation — proven by
  `SEC-04/prod-config-gate-tests.md`.
- **AC-W05-E04-S002-04**: This story's implementation explicitly closes DATA-07 T4's
  cache-invalidation acceptance criterion, recorded by cross-reference ID in this story's
  `dependencies.md` and confirmed at closure.

## Required artifacts

- The bounded, sharded cache (code).
- Eviction and admission/eviction metrics (code).
- Singleflight miss-collapse (code).
- The `authz_epoch` table migration and epoch-bump wiring (code + migration).
- Decision-provenance metadata (code).
- The prod-config gate (code).
See `artifacts/index.md`.

## Required evidence

- `SEC-04/bounded-cache-tests.md`.
- `SEC-04/eviction-metrics-tests.md`.
- `SEC-04/singleflight-tests.md`.
- `SEC-04/cross-pod-epoch-tests.md`.
- `SEC-04/decision-provenance-tests.md`.
- `SEC-04/prod-config-gate-tests.md`.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, D-06 referenced (not
re-derived) via `decisions/index.md`, the DATA-07 T4 AC-closure relationship recorded, owner/reviewer
assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T4's epoch-bump wiring is complete across every
enumerated mutation path, given PLAN's own "Highest-risk task" framing.

## Risks

RISK-W05-005 (T4's epoch-bump wiring completeness) — see epic-level `risks.md` for full detail and
mitigation/contingency.

## Residual-risk expectations

Residual risk is expected to be low once the enumerated mutation-path list is confirmed complete and
independently re-checked by this story's own review task.

## Plan

See `plan.md`.
