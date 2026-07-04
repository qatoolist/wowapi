-- Machine authentication: API keys / service principals (roadmap S1). Non-human
-- callers (gate devices, integrations) authenticate without a user token. Each
-- key is scoped (an explicit permission set), rotatable, revocable, and
-- expirable; only the sha256 of the secret is stored (never the secret). The
-- public key_prefix is the lookup handle. Verification is cross-tenant (a key is
-- presented before any tenant is known), so app_platform reads by prefix via a
-- permissive policy (like events_outbox); key management is tenant-scoped.

-- +goose Up

CREATE TABLE api_keys (
    id           uuid PRIMARY KEY,
    tenant_id    uuid NOT NULL,
    name         text NOT NULL,
    key_prefix   text NOT NULL UNIQUE,        -- public lookup handle
    key_hash     text NOT NULL,               -- sha256(secret) hex; secret never stored
    scopes       text[] NOT NULL DEFAULT '{}',-- granted permissions (authz machine scope)
    expires_at   timestamptz,                 -- NULL = no expiry
    revoked_at   timestamptz,                 -- NULL = active
    last_used_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    created_by   uuid
);

CREATE INDEX api_keys_tenant ON api_keys (tenant_id, created_at DESC);

ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys FORCE ROW LEVEL SECURITY;
-- Tenant-scoped management (issue/list/revoke as app_rt).
CREATE POLICY api_keys_tenant_isolation ON api_keys
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
-- Cross-tenant verification (pre-auth lookup by prefix as app_platform).
CREATE POLICY api_keys_platform_all ON api_keys
    TO app_platform USING (true) WITH CHECK (true);

GRANT SELECT, INSERT, UPDATE ON api_keys TO app_rt;        -- manage own tenant's keys
GRANT SELECT, UPDATE ON api_keys TO app_platform;          -- verify + bump last_used_at

-- +goose Down

DROP TABLE IF EXISTS api_keys;
