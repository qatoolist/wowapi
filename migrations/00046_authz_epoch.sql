-- Per-tenant authz_epoch table for invalidating the sharded authorization cache (SEC-04 / D-06).
-- Each tenant has exactly one row; the epoch increment invalidates cached assignments.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 1000
-- bytes_estimate: 65536
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM authz_epoch
-- rollback_plan: goose Down drops the authz_epoch table.
-- +wowapi:end

-- +goose Up

CREATE TABLE authz_epoch (
    tenant_id uuid PRIMARY KEY,
    epoch     integer NOT NULL DEFAULT 1
);

ALTER TABLE authz_epoch ENABLE ROW LEVEL SECURITY;
ALTER TABLE authz_epoch FORCE ROW LEVEL SECURITY;
CREATE POLICY authz_epoch_tenant_isolation ON authz_epoch
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());

GRANT SELECT, INSERT, UPDATE ON authz_epoch TO app_rt;
GRANT SELECT, INSERT, UPDATE ON authz_epoch TO app_platform;

-- +goose Down

DROP TABLE IF EXISTS authz_epoch;
