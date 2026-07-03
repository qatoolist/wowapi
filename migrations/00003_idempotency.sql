-- Blueprint 002-equivalent (pulled forward from the 03 §5 "009" batch, D-0031):
-- the idempotency_keys table backing kernel/database.IdemStore and
-- kernel/httpx.WithIdempotency. Tenant-scoped with RLS, so a stored response
-- for one tenant can never be replayed for another.

-- +goose Up

CREATE TABLE idempotency_keys (
    tenant_id       uuid NOT NULL,
    actor_scope     text NOT NULL,          -- capacity id or system actor identifier
    idem_key        text NOT NULL,          -- client-supplied Idempotency-Key
    request_hash    text NOT NULL,          -- hash of method+path+body; mismatch => conflict
    status          text NOT NULL DEFAULT 'in_progress'
                        CHECK (status IN ('in_progress','completed')),
    response_status int,                     -- populated when completed
    response_body   bytea,                   -- stored response replayed BYTE-EXACT on retry
                                             -- (bytea, not jsonb: jsonb reformats whitespace and
                                             --  would change the replayed bytes vs the original)
    created_at      timestamptz NOT NULL DEFAULT now(),
    expires_at      timestamptz NOT NULL,    -- swept after this instant
    PRIMARY KEY (tenant_id, actor_scope, idem_key)
);

-- Tenant isolation: identical convention to every tenant-scoped table (03 §1).
ALTER TABLE idempotency_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE idempotency_keys FORCE ROW LEVEL SECURITY;
CREATE POLICY idempotency_keys_tenant_isolation ON idempotency_keys
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());

-- Sweep index for the expiry job (Phase 6+).
CREATE INDEX idempotency_keys_expiry ON idempotency_keys (expires_at);

-- Runtime role needs full row lifecycle here (rows are created, completed, and
-- swept); still no schema-wide grants.
GRANT SELECT, INSERT, UPDATE, DELETE ON idempotency_keys TO app_rt;

-- +goose Down

DROP TABLE IF EXISTS idempotency_keys;
