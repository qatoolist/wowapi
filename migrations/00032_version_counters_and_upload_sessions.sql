-- DATA-05 (W02-E03-S001): durable version allocation + upload-session GC.
-- Replaces the inline MAX(version)+1 / MAX(version_no)+1 races in
-- kernel/artifact.Generate and kernel/document.InitiateUpload with locked
-- per-scope counters, and makes upload sessions durable so a crash between
-- InitiateUpload and ConfirmUpload can be recovered and garbage-collected.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 10000
-- nn1_compatible: true
-- backfill_owner: app_platform
-- validation_query: SELECT count(*) FROM version_counters WHERE false; SELECT count(*) FROM document_upload_sessions WHERE false
-- rollback_plan: goose Down drops version_counters and document_upload_sessions; additive-only table creation, no data loss.
-- +wowapi:end

-- +goose Up

-- Per-(tenant, scope) locked counter. The upsert
--   INSERT ... ON CONFLICT (tenant_id, scope) DO UPDATE SET value = value + 1
-- serializes concurrent allocators on the row and returns a distinct version.
CREATE TABLE version_counters (
    tenant_id uuid    NOT NULL,
    scope     text    NOT NULL,
    value     int     NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, scope)
);

ALTER TABLE version_counters ENABLE ROW LEVEL SECURITY;
ALTER TABLE version_counters FORCE ROW LEVEL SECURITY;
CREATE POLICY version_counters_tenant_isolation ON version_counters
    USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id());

GRANT SELECT, INSERT, UPDATE ON version_counters TO app_rt;
GRANT SELECT, INSERT, UPDATE ON version_counters TO app_platform;

-- Durable upload sessions. A pending session reserves a version number and a
-- storage key; ConfirmUpload CASes pending->confirmed. Expired sessions are
-- garbage-collected by the platform sweep.
CREATE TABLE document_upload_sessions (
    id              uuid PRIMARY KEY,
    tenant_id       uuid        NOT NULL,
    document_id     uuid        NOT NULL REFERENCES documents(id),
    version_no      int         NOT NULL,
    storage_key     text        NOT NULL,
    status          text        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','confirmed','expired')),
    expires_at      timestamptz NOT NULL,
    checksum_sha256 text,
    size_bytes      bigint,
    mime_type       text,
    created_at      timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE document_upload_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE document_upload_sessions FORCE ROW LEVEL SECURITY;
CREATE POLICY document_upload_sessions_tenant_isolation ON document_upload_sessions
    USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id());

GRANT SELECT, INSERT, UPDATE ON document_upload_sessions TO app_rt;
GRANT SELECT, INSERT, UPDATE, DELETE ON document_upload_sessions TO app_platform;

-- Backfill counters so existing data does not produce duplicate versions.
-- artifacts -> scope 'artifact:'||kind; document_versions -> 'document:'||doc_id.
INSERT INTO version_counters (tenant_id, scope, value)
SELECT tenant_id, 'artifact:' || kind, COALESCE(MAX(version), 0)
FROM artifacts
GROUP BY tenant_id, kind
ON CONFLICT (tenant_id, scope) DO NOTHING;

INSERT INTO version_counters (tenant_id, scope, value)
SELECT tenant_id, 'document:' || document_id::text, COALESCE(MAX(version_no), 0)
FROM document_versions
GROUP BY tenant_id, document_id
ON CONFLICT (tenant_id, scope) DO NOTHING;

-- A document may have only one confirmed session per version_no.
CREATE UNIQUE INDEX document_upload_sessions_confirmed_version
    ON document_upload_sessions (document_id, version_no)
    WHERE status = 'confirmed';

-- GC query index: pending sessions that have expired.
CREATE INDEX document_upload_sessions_gc
    ON document_upload_sessions (tenant_id, expires_at)
    WHERE status = 'pending';

-- +goose Down

DROP TABLE IF EXISTS document_upload_sessions;
DROP TABLE IF EXISTS version_counters;
