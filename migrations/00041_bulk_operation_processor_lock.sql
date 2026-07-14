-- Bulk-operation single-processor stopgap (DATA-04 T1). Until the leased-claim
-- rewrite lands, Process enforces a single active processor per operation via a
-- short-lived processor lock on bulk_operations. A crashed processor's lock
-- expires after a bounded timeout so the operation can be resumed.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM bulk_operations WHERE processor_id IS NOT NULL AND processor_started_at IS NULL
-- rollback_plan: goose Down drops processor_id and processor_started_at and revokes the app_rt column grant.
-- +wowapi:end

-- +goose Up

ALTER TABLE bulk_operations
    ADD COLUMN processor_id uuid,
    ADD COLUMN processor_started_at timestamptz;

GRANT UPDATE (processor_id, processor_started_at) ON bulk_operations TO app_rt;

-- +goose Down

ALTER TABLE bulk_operations
    DROP COLUMN IF EXISTS processor_id,
    DROP COLUMN IF EXISTS processor_started_at;
