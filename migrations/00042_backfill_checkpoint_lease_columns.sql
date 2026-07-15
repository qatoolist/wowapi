-- Replaces the interim checkpoint-lease code path with the shared lease primitive.
-- Existing checkpoints keep their last_key; generation starts at 0 and is bumped
-- on the next run. The shared lease primitive in kernel/lease owns the semantics.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM migration.backfill_checkpoint WHERE lease_generation IS NULL
-- rollback_plan: goose Down drops lease_token, lease_generation, and lease_expires_at; table is preserved if it already existed.
-- +wowapi:end

-- +goose Up

CREATE SCHEMA IF NOT EXISTS migration;

CREATE TABLE IF NOT EXISTS migration.backfill_checkpoint (
    job_id text PRIMARY KEY,
    tenant_id uuid,
    last_key bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    lease_token text,
    lease_generation bigint NOT NULL DEFAULT 0,
    lease_expires_at timestamptz
);

ALTER TABLE migration.backfill_checkpoint
    ADD COLUMN IF NOT EXISTS lease_token text,
    ADD COLUMN IF NOT EXISTS lease_generation bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lease_expires_at timestamptz;

-- +goose Down

ALTER TABLE migration.backfill_checkpoint
    DROP COLUMN IF EXISTS lease_token,
    DROP COLUMN IF EXISTS lease_generation,
    DROP COLUMN IF EXISTS lease_expires_at;
