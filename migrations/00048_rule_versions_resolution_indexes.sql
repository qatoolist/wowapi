-- PERF-03 (W07-E01-S002): bounded rules resolution reads active and
-- superseded versions across org, tenant, and platform scopes. The active-only
-- exclusion index from 00008 covers the current predicate; the original lookup
-- index omitted scope_id/tenant_id and did not cover superseded history.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 600000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM pg_indexes WHERE schemaname='public' AND tablename='rule_versions' AND indexname='rule_versions_history_resolution_idx'
-- rollback_plan: goose Down recreates the original active-only lookup index before dropping the historical-resolution index; no data loss.
-- +wowapi:end

-- +goose NO TRANSACTION

-- +goose Up

SET lock_timeout = '2s';
SET statement_timeout = '10min';


CREATE INDEX CONCURRENTLY IF NOT EXISTS rule_versions_history_resolution_idx
    ON rule_versions (rule_key, scope_kind, scope_id, tenant_id, effective_from DESC)
    WHERE status IN ('active', 'superseded');

DROP INDEX CONCURRENTLY IF EXISTS rule_versions_lookup;

RESET lock_timeout;
RESET statement_timeout;

-- +goose Down

SET lock_timeout = '2s';
SET statement_timeout = '10min';

CREATE INDEX CONCURRENTLY IF NOT EXISTS rule_versions_lookup
    ON rule_versions (rule_key, scope_kind, effective_from)
    WHERE status = 'active';

DROP INDEX CONCURRENTLY IF EXISTS rule_versions_history_resolution_idx;

RESET lock_timeout;
RESET statement_timeout;
