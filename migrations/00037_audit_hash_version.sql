-- DATA-08 W6-T1 (W04-E04-S001): add hash_version discriminator to audit_logs so the
-- widened chainHash (now covering canonicalized metadata and tx_id) can coexist with
-- historical rows that were hashed under the original 15-field scheme. D-04 reserves
-- hash_version=1 for the historical scheme; new rows are written as hash_version=2.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM audit_logs WHERE hash_version IS NULL
-- rollback_plan: goose Down drops the hash_version column; historical rows remain verifiable under the v1 scheme.
-- +wowapi:end

-- +goose Up

SET LOCAL lock_timeout = '2s';
SET LOCAL statement_timeout = '5s';

ALTER TABLE audit_logs ADD COLUMN hash_version smallint NOT NULL DEFAULT 1;

-- +goose Down

ALTER TABLE audit_logs DROP COLUMN hash_version;
