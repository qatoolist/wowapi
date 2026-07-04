# Migration safety (O2)

Domain-neutral guidance for evolving the schema without downtime, and the CI drill that keeps every
migration reversible.

## Reversibility drill (CI)

`kernel/database.MigrateReset` rolls a source back to version 0 (goose Down, newest-first).
`TestIntegrationMigrationsReversible` (in `migrations/`) runs the full **forward → down → forward**
cycle on an isolated database and asserts the head version is reproduced. It runs in `make ci-container`
(DB tests forced), so a migration whose `-- +goose Down` block is missing or wrong fails CI.

This drill already caught a real defect: migration 00010 created the `app_actor_id()` function but its
Down did not drop it, so a re-apply failed with "function already exists". Rule of thumb it enforces:
**every object your Up creates (table, function, type, policy, index), your Down must drop** — and only
those; never drop cluster-scoped objects (roles, extensions) that sibling databases share.

## Zero-downtime expand/contract

Never rewrite a column or table in place while old and new code run side by side. Split every breaking
change into an **expand** migration (backward-compatible, deploy first), a **backfill/dual-write**
window, and a later **contract** migration (removes the old shape, deploy after all replicas run new
code). The journal-bearing tables (`events_outbox`, `jobs_queue`, `job_runs`, and any product ledger)
are especially sensitive — a relay/runner mid-batch must keep reading the old shape.

| Change | Expand (deploy 1) | Transition | Contract (deploy 2) |
|---|---|---|---|
| Add a column | `ADD COLUMN … NULL` (or with a default that does not rewrite the table) | code starts writing it; backfill in batches | add `NOT NULL`/constraint once backfilled |
| Rename a column | add the new column, dual-write both | backfill new from old; switch reads to new | drop the old column |
| Change a type | add a new column of the new type, dual-write | backfill + verify | drop the old column, rename |
| Drop a column | stop writing it (code first) | — | `DROP COLUMN` after no code references it |
| Add an index | `CREATE INDEX CONCURRENTLY` (outside a tx) | — | — |
| Add a NOT NULL | add `CHECK (col IS NOT NULL) NOT VALID`, backfill, `VALIDATE CONSTRAINT` | — | promote to `NOT NULL` |

Guidelines: keep each migration small and single-purpose; additive changes are always safe to deploy
before code; destructive changes always deploy after code that stopped using the old shape; prefer
`CONCURRENTLY` for indexes on hot tables; batch backfills so a single statement never locks a large
table. Every migration must have a correct Down (the drill enforces it).
</content>
