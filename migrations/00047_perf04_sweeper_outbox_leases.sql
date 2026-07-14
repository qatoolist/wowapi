-- PERF-04 (W07-E01-S003): bound workflow SLA scans by the reminder predicate
-- and add W04 DATA-02-compatible lease/fencing columns to the outbox relay.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 600000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM pg_indexes WHERE schemaname='public' AND tablename='workflow_tasks' AND indexname='wft_remind_after'
-- rollback_plan: goose Down drops the additive reminder index and nullable outbox lease columns; no data rewrite is required.
-- +wowapi:end

-- +goose NO TRANSACTION

-- +goose Up

SET lock_timeout = '2s';
SET statement_timeout = '10min';

ALTER TABLE events_outbox
    ADD COLUMN IF NOT EXISTS lease_token text,
    ADD COLUMN IF NOT EXISTS lease_generation bigint NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lease_expires_at timestamptz;

CREATE INDEX CONCURRENTLY IF NOT EXISTS wft_remind_after
    ON workflow_tasks (tenant_id, remind_after, id)
    WHERE status = 'open'
      AND remind_after IS NOT NULL
      AND (last_reminded_at IS NULL OR last_reminded_at < remind_after);

RESET lock_timeout;
RESET statement_timeout;

-- +goose Down

SET lock_timeout = '2s';
SET statement_timeout = '10min';

DROP INDEX CONCURRENTLY IF EXISTS wft_remind_after;

ALTER TABLE events_outbox
    DROP COLUMN IF EXISTS lease_expires_at,
    DROP COLUMN IF EXISTS lease_generation,
    DROP COLUMN IF EXISTS lease_token;

RESET lock_timeout;
RESET statement_timeout;
