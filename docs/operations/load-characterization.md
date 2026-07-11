# Load Characterization — Per-Aggregate Advisory-Lock Ordering (R2 / CA-4)

Roadmap item **R2** requires a documented throughput envelope for the outbox relay's per-aggregate
ordering guarantee, plus a sub-sharding strategy for when that envelope is exceeded. This is that
document; the number below is produced by a repeatable test, not an estimate.

## The mechanism under test

The relay preserves **per-aggregate event ordering**: within a single aggregate (a `resource.Ref`), events
are dispatched strictly in emission order. It does this by taking a Postgres **transaction-scoped advisory
lock keyed on the aggregate** before dispatching each event
(`kernel/outbox/relay.go` → `pg_advisory_xact_lock(hashtextextended(<aggregate>, 0))`). Events for
*different* aggregates dispatch in parallel; events for the *same* aggregate serialize on that lock.

So the worst case for throughput is a **hot aggregate**: a burst of events all targeting one
`resource.Ref` (the classic shape is a bill-run posting many lines to one parent document).

## Measured envelope

Test: `kernel/outbox/outbox_test.go::TestIntegrationOutboxHotAggregateThroughput`. It emits 200 events onto
a single aggregate and drains them with 4 concurrent relay workers, measuring wall-clock throughput.

Reproduce:

```bash
make up   # local Postgres
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
  WOWAPI_REQUIRE_DB=1 go test ./kernel/outbox/ -run HotAggregateThroughput -v -count=1
```

**Observed:** ~**200 events/sec** for a single hot aggregate (local Postgres 16 in Docker, 4 relay
workers). Because the aggregate lock serializes dispatch, adding relay workers does **not** raise
single-aggregate throughput — it is bounded by per-event dispatch latency (handler + mark-dispatched
round-trips), not by relay parallelism. Multi-aggregate workloads scale out with worker count and DB
capacity; only same-aggregate contention hits this ceiling.

Treat ~200 events/sec/aggregate as an order-of-magnitude floor on modest hardware; production Postgres with
lower round-trip latency will exceed it, but the shape (flat in worker count for one aggregate) holds.

## When you exceed the envelope: sub-sharding

If a single aggregate must absorb a sustained emission rate above its per-aggregate ceiling, **sub-shard
the aggregate key** to trade strict per-aggregate ordering for throughput:

1. Choose a shard count `N` (e.g. 8) sized to your target rate ÷ the per-aggregate envelope.
2. Emit against a **derived aggregate key** that appends a deterministic shard suffix, e.g.
   `"<aggregate-id>#<shard>"` where `shard = hash(orderingSubkey) % N`. Pick the ordering subkey so that
   events which must stay ordered relative to each other land on the **same** shard (e.g. shard by line-item
   group, not by random), and events that are independent spread across shards.
3. Consumers that need a global order re-merge by a monotonic field (sequence number / `created_at`) after
   the fact; consumers that only need per-subkey order get it for free.

This preserves ordering **within** a shard while allowing up to `N` shards to dispatch in parallel, raising
the aggregate's effective ceiling to ~`N ×` the single-aggregate envelope until the DB or worker pool
saturates.

## Guidance

- Most aggregates never approach the ceiling — do **not** sub-shard preemptively; it complicates ordering.
- Sub-shard only the specific hot aggregates a product identifies under its own load (e.g. the bill-run
  parent), and record the choice of `N` and the ordering subkey alongside the module.
- Re-run the characterization test on production-class hardware to set your own per-aggregate budget before
  committing a shard count.

## Status

R2 is **closed** by this pass: the throughput envelope is measured by a repeatable test and the
sub-sharding strategy is documented (roadmap CA-4). It was previously declassified to "doc-only" in
`hardening-plan.md` (superseded; archived 2026-07-11 to the `wowapi2` archive, `archive/plans/`)
without the required evidence; that evidence now exists here.
