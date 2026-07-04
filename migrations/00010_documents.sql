-- Blueprint 03 §5 "007" (on-disk 00010) + 07 §4: documents / versions / access
-- grants / comments / attachments. All are strict tenant-scoped tables (RLS on
-- app_tenant_id()). document_versions is append-only: app_rt may SELECT/INSERT
-- but NOT mutate; scan-status and retention voiding run as app_platform (the
-- worker/scan role), mirroring the relay posture — a module can never rewrite an
-- immutable file pointer or clear an infected flag.

-- +goose Up

-- app_actor_id() exposes the SET LOCAL app.actor_id to RLS policies (the acting
-- capacity/user). It RAISES when unset — a write that depends on actor identity
-- (e.g. creating a document access grant) must fail closed rather than evaluate
-- against a null actor. Mirrors app_tenant_id().
-- +goose StatementBegin
CREATE FUNCTION app_actor_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT current_setting('app.actor_id')::uuid
$$;
-- +goose StatementEnd

CREATE TABLE documents (
    id             uuid PRIMARY KEY,
    tenant_id      uuid NOT NULL,
    document_class text NOT NULL,               -- registered by modules
    resource_type  text, resource_id uuid,      -- optional anchor
    title          text NOT NULL,
    sensitivity    text NOT NULL DEFAULT 'internal'
        CHECK (sensitivity IN ('public','internal','confidential','restricted')),
    retention_until timestamptz,
    legal_hold     boolean NOT NULL DEFAULT false,
    status         text NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','voided')),
    version        int NOT NULL DEFAULT 1,
    created_at     timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at     timestamptz, updated_by uuid
);
CREATE INDEX doc_resource ON documents (tenant_id, resource_type, resource_id);
CREATE INDEX doc_class ON documents (tenant_id, document_class);
-- Retention sweep target: active docs past their retention with no legal hold.
CREATE INDEX doc_retention ON documents (tenant_id, retention_until)
    WHERE status = 'active' AND legal_hold = false;

CREATE TABLE document_versions (            -- append-only file pointers
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    document_id     uuid NOT NULL REFERENCES documents(id),
    version_no      int NOT NULL,
    storage_key     text NOT NULL,
    mime_type       text NOT NULL,
    size_bytes      bigint NOT NULL,
    checksum_sha256 text NOT NULL,
    scan_status     text NOT NULL DEFAULT 'pending'
        CHECK (scan_status IN ('pending','clean','infected','skipped')),
    -- status is the retention tombstone: voiding deletes the blob and flips this
    -- to 'voided' (void != hard erase; a redaction job does hard erasure).
    status          text NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','voided')),
    voided_at       timestamptz,
    uploaded_by     uuid NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (document_id, version_no)
);
CREATE INDEX docver_doc ON document_versions (tenant_id, document_id, version_no DESC);

CREATE TABLE document_access_grants (
    id           uuid PRIMARY KEY,
    tenant_id    uuid NOT NULL,
    document_id  uuid NOT NULL REFERENCES documents(id),
    grantee_kind text NOT NULL CHECK (grantee_kind IN ('capacity','role','relationship')),
    grantee_ref  text NOT NULL,
    access       text NOT NULL CHECK (access IN ('read','write')),
    valid_from   timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
    version      int NOT NULL DEFAULT 1,
    created_at   timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL
);
CREATE INDEX docgrant_doc ON document_access_grants (tenant_id, document_id);

CREATE TABLE comments (
    id                uuid PRIMARY KEY,
    tenant_id         uuid NOT NULL,
    resource_type     text NOT NULL, resource_id uuid NOT NULL,
    parent_comment_id uuid REFERENCES comments(id),
    author_capacity_id uuid NOT NULL, body text NOT NULL,
    status            text NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','edited','voided')),
    version           int NOT NULL DEFAULT 1,
    created_at        timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at        timestamptz, updated_by uuid
);
CREATE INDEX cmt_resource ON comments (tenant_id, resource_type, resource_id, created_at DESC);

CREATE TABLE attachments (
    id                  uuid PRIMARY KEY,
    tenant_id           uuid NOT NULL,
    resource_type       text NOT NULL, resource_id uuid NOT NULL,
    document_version_id uuid NOT NULL REFERENCES document_versions(id),
    comment_id          uuid REFERENCES comments(id),
    workflow_task_id    uuid REFERENCES workflow_tasks(id),
    status              text NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','voided')),
    version             int NOT NULL DEFAULT 1,
    created_at          timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL
);
CREATE INDEX att_resource ON attachments (tenant_id, resource_type, resource_id);

-- Strict tenant RLS on every table; app_rt gets SELECT/INSERT/UPDATE except on
-- the append-only document_versions (SELECT/INSERT only — scan + retention
-- mutations are app_platform).
-- +goose StatementBegin
DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['documents','document_versions','document_access_grants','comments','attachments']
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('CREATE POLICY %I ON %I USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id())', t||'_tenant_isolation', t);
    END LOOP;
END
$$;
-- +goose StatementEnd

-- documents: app_rt may INSERT and UPDATE only the mutable metadata columns.
-- status, legal_hold, and retention_until are governance controls the module
-- role must NOT rewrite (a module could otherwise clear a legal hold or void a
-- document to dodge retention — SEC-44). Those columns are app_platform-only.
GRANT SELECT, INSERT ON documents TO app_rt;
GRANT UPDATE (title, sensitivity, version, updated_at, updated_by) ON documents TO app_rt;
GRANT SELECT, UPDATE ON documents TO app_platform; -- retention sweep: status/legal_hold/retention_until

GRANT SELECT, INSERT              ON document_versions TO app_rt;   -- append-only
GRANT SELECT, INSERT, UPDATE      ON document_versions TO app_platform; -- scan + retention voiding

-- document_access_grants: app_rt may write, but a RESTRICTIVE policy pins every
-- INSERT/UPDATE to a grant on a document the ACTING actor owns — so a module
-- cannot self-grant on a document it does not own, nor redirect/escalate an
-- existing grant to another document (SEC-41/42). Tenant isolation is the
-- permissive policy from the loop above; this ANDs onto it.
CREATE POLICY document_access_grants_owner_write ON document_access_grants
    AS RESTRICTIVE FOR ALL
    USING (true)
    WITH CHECK (EXISTS (SELECT 1 FROM documents d
                         WHERE d.id = document_id AND d.created_by = app_actor_id()));
GRANT SELECT, INSERT, UPDATE ON document_access_grants TO app_rt;

GRANT SELECT, INSERT, UPDATE ON comments    TO app_rt;   -- edit keeps history; void != delete
GRANT SELECT, INSERT, UPDATE ON attachments TO app_rt;   -- void != delete

-- +goose Down

DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS document_access_grants;
DROP TABLE IF EXISTS document_versions;
DROP TABLE IF EXISTS documents;
-- Drop the function this migration created (its Up does CREATE FUNCTION, so a
-- rollback that leaves it behind makes a subsequent re-apply fail — caught by
-- the O2 reversibility drill). Safe: it is created here and only referenced by
-- the document policies dropped above; later migrations' Down runs first.
DROP FUNCTION IF EXISTS app_actor_id();
