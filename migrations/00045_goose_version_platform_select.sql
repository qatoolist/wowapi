-- DX-07 T1 (W04-E04-S003): grant app_platform SELECT on goose_version_wowapi so
-- the generated readiness endpoint can verify migration currency without needing
-- migration-owner credentials at runtime.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM information_schema.table_privileges WHERE table_name = 'goose_version_wowapi' AND grantee = 'app_platform' AND privilege_type = 'SELECT'
-- rollback_plan: goose Down revokes the SELECT grant.
-- +wowapi:end

-- +goose Up

SET LOCAL lock_timeout = '2s';
SET LOCAL statement_timeout = '5s';

GRANT SELECT ON goose_version_wowapi TO app_platform;

-- +goose Down

REVOKE SELECT ON goose_version_wowapi FROM app_platform;
