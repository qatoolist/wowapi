-- Lease/fencing columns for jobs_queue (DATA-02 T2). These back the shared
-- lease primitive in kernel/lease so claim/finalize/reclaim can fence stale
-- workers. lease_token is NULL for rows that have never been claimed under the
-- new primitive (pre-fencing rows or fully reset rows); lease_generation is 0
-- for those rows and advances on every reclaim, producing a provably new epoch.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM jobs_queue WHERE lease_generation IS NULL
-- rollback_plan: goose Down drops lease_token, lease_generation, and lease_expires_at.
-- +wowapi:end

-- +goose Up

ALTER TABLE jobs_queue
    ADD COLUMN lease_token text,
    ADD COLUMN lease_generation bigint NOT NULL DEFAULT 0,
    ADD COLUMN lease_expires_at timestamptz;

-- +goose Down

ALTER TABLE jobs_queue
    DROP COLUMN IF EXISTS lease_token,
    DROP COLUMN IF EXISTS lease_generation,
    DROP COLUMN IF EXISTS lease_expires_at;
