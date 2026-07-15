-- FBL-02 (W02-E05-S001): durable audit/state record for catalog seed-sync runs.
-- Append-only global table: app_platform records every successful (applied/noop)
-- sync and best-effort failed runs. It is NOT tenant-scoped — seed-sync is a
-- global operation — so no RLS policy is applied; access is controlled by grants.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 10000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM seed_sync_runs WHERE false
-- rollback_plan: goose Down drops seed_sync_runs; additive-only table creation, no data loss.
-- +wowapi:end

-- +goose Up

CREATE TABLE seed_sync_runs (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    manifest_hash  text        NOT NULL,
    version_label  text        NOT NULL DEFAULT '',
    actor          text        NOT NULL DEFAULT '',
    outcome        text        NOT NULL CHECK (outcome IN ('applied', 'noop', 'failed')),
    counts         jsonb       NOT NULL DEFAULT '{}',
    error          text,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX seed_sync_runs_created_at_idx ON seed_sync_runs (created_at DESC);
CREATE INDEX seed_sync_runs_hash_idx ON seed_sync_runs (manifest_hash);

-- app_platform: append-only writer and reader of the sync audit log.
GRANT SELECT, INSERT ON seed_sync_runs TO app_platform;

-- app_rt: no access — runtime tenants have no business reading global sync audit.
-- (No GRANT to app_rt is intentional.)

-- +goose Down

DROP TABLE IF EXISTS seed_sync_runs;
