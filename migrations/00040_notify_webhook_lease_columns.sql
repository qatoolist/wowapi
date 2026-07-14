-- Lease/fencing columns for notification_deliveries and webhook_events (DATA-03 T1).
-- These back the shared lease primitive in kernel/lease so the notify/webhook
-- three-stage claim/effect/finalize protocol can fence stale workers.
-- lease_token is NULL for rows never claimed under the new primitive;
-- lease_generation defaults to 0 and advances on every reclaim, producing a
-- provably new epoch.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM notification_deliveries WHERE lease_generation IS NULL UNION ALL SELECT count(*) FROM webhook_events WHERE lease_generation IS NULL
-- rollback_plan: goose Down drops lease_token, lease_generation, and lease_expires_at from notification_deliveries and webhook_events.
-- +wowapi:end

-- +goose Up

ALTER TABLE notification_deliveries
    ADD COLUMN lease_token text,
    ADD COLUMN lease_generation bigint NOT NULL DEFAULT 0,
    ADD COLUMN lease_expires_at timestamptz;

ALTER TABLE webhook_events
    ADD COLUMN lease_token text,
    ADD COLUMN lease_generation bigint NOT NULL DEFAULT 0,
    ADD COLUMN lease_expires_at timestamptz;

-- +goose Down

ALTER TABLE notification_deliveries
    DROP COLUMN IF EXISTS lease_token,
    DROP COLUMN IF EXISTS lease_generation,
    DROP COLUMN IF EXISTS lease_expires_at;

ALTER TABLE webhook_events
    DROP COLUMN IF EXISTS lease_token,
    DROP COLUMN IF EXISTS lease_generation,
    DROP COLUMN IF EXISTS lease_expires_at;
