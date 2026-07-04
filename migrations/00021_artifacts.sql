-- Snapshot / artifact pipeline (roadmap E4): a dataset rendered to an IMMUTABLE,
-- versioned artifact (content + sha256 + structured sidecar), with the template
-- version and effective date it was produced under. The framework owns
-- immutability, per-(tenant,kind) versioning, hashing, and tamper-verification;
-- the product supplies the rendered bytes (e.g. a PDF/A its own renderer emits),
-- so no document-format library enters the kernel. Content is stored in-row
-- (compliance artifacts are bounded-size) so an artifact is atomic and
-- self-verifying. Append-only: app_rt has INSERT+SELECT, never UPDATE/DELETE.

-- +goose Up

CREATE TABLE artifacts (
    id               uuid PRIMARY KEY,
    tenant_id        uuid NOT NULL,
    kind             text NOT NULL,            -- 'receipt', 'certificate', …
    version          int  NOT NULL,            -- per (tenant, kind), 1-based
    content_hash     text NOT NULL,            -- sha256(content) hex
    content          bytea NOT NULL,           -- the immutable rendered bytes
    content_type     text NOT NULL DEFAULT 'application/pdf',
    sidecar          jsonb NOT NULL DEFAULT '{}',  -- structured metadata alongside the artifact
    template_version text,                      -- template used
    effective_date   timestamptz,              -- the date the template was effective for
    created_at       timestamptz NOT NULL DEFAULT now(),
    created_by       uuid,
    UNIQUE (tenant_id, kind, version)
);

CREATE INDEX artifacts_kind ON artifacts (tenant_id, kind, version DESC);

ALTER TABLE artifacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE artifacts FORCE ROW LEVEL SECURITY;
CREATE POLICY artifacts_tenant_isolation ON artifacts
    USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT ON artifacts TO app_rt;  -- append-only: no UPDATE/DELETE

-- +goose Down

DROP TABLE IF EXISTS artifacts;
