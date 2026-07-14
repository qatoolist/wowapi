-- DATA-05 (W02-E03-S001): secondary index for looking up pending/confirmed upload
-- sessions by document. This migration is intentionally numbered 00033 to keep
-- the goose version sequence contiguous between 00032 and 00034.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 10000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM pg_indexes WHERE schemaname='public' AND tablename='document_upload_sessions' AND indexname='document_upload_sessions_document_id_idx'
-- rollback_plan: goose Down drops the index; additive-only index creation, no data loss.
-- +wowapi:end

-- +goose Up

CREATE INDEX IF NOT EXISTS document_upload_sessions_document_id_idx
    ON document_upload_sessions (document_id);

-- +goose Down

DROP INDEX IF EXISTS document_upload_sessions_document_id_idx;
