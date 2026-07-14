-- Bulk-items lease/fencing columns and operation lifecycle (DATA-04 T2-T5).
-- Reuses the shared lease primitive in kernel/lease for bulk_items, adds per-item
-- idempotency keys and a per-operation retry budget, and extends operation/item
-- statuses to support pause/resume/cancel. Supersedes migration 00041's stopgap
-- processor lock, which is dropped once the leased-claim path lands.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM bulk_items WHERE lease_generation IS NULL
-- rollback_plan: goose Down reverses column additions/drops and restores the original status check constraints.
-- +wowapi:end

-- +goose Up

ALTER TABLE bulk_items
    ADD COLUMN lease_token text,
    ADD COLUMN lease_generation bigint NOT NULL DEFAULT 0,
    ADD COLUMN lease_expires_at timestamptz,
    ADD COLUMN idempotency_key uuid;

ALTER TABLE bulk_operations
    ADD COLUMN max_attempts int NOT NULL DEFAULT 3,
    DROP COLUMN IF EXISTS processor_id,
    DROP COLUMN IF EXISTS processor_started_at;

-- Extend status enums for lifecycle controls.
ALTER TABLE bulk_operations DROP CONSTRAINT IF EXISTS bulk_operations_status_check;
ALTER TABLE bulk_operations ADD CONSTRAINT bulk_operations_status_check
    CHECK (status IN ('pending','running','paused','completed','cancelled'));

ALTER TABLE bulk_items DROP CONSTRAINT IF EXISTS bulk_items_status_check;
ALTER TABLE bulk_items ADD CONSTRAINT bulk_items_status_check
    CHECK (status IN ('pending','running','done','failed','cancelled'));

-- +goose Down

ALTER TABLE bulk_items
    DROP COLUMN IF EXISTS lease_token,
    DROP COLUMN IF EXISTS lease_generation,
    DROP COLUMN IF EXISTS lease_expires_at,
    DROP COLUMN IF EXISTS idempotency_key;

ALTER TABLE bulk_operations
    DROP COLUMN IF EXISTS max_attempts,
    ADD COLUMN processor_id uuid,
    ADD COLUMN processor_started_at timestamptz;

ALTER TABLE bulk_operations DROP CONSTRAINT IF EXISTS bulk_operations_status_check;
ALTER TABLE bulk_operations ADD CONSTRAINT bulk_operations_status_check
    CHECK (status IN ('pending','running','completed'));

ALTER TABLE bulk_items DROP CONSTRAINT IF EXISTS bulk_items_status_check;
ALTER TABLE bulk_items ADD CONSTRAINT bulk_items_status_check
    CHECK (status IN ('pending','done','failed'));
