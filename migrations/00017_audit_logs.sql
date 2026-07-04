-- Durable field-level audit trail (roadmap E1). A standardized, append-only,
-- queryable record of who changed what: entity, field, before/after, actor,
-- capacity, impersonator, request id — written inside the business transaction
-- so an audit row commits iff the change does. Append-only is enforced at the
-- grant level: app_rt may INSERT and SELECT but NOT UPDATE or DELETE, so the
-- runtime cannot rewrite history. (Cryptographic tamper-evidence — hash-chaining
-- per S6 — layers on top of this table later.)

-- +goose Up

CREATE TABLE audit_logs (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    occurred_at     timestamptz NOT NULL DEFAULT now(),
    actor_id        uuid,               -- app.actor_id (the acting capacity)
    actor_kind      text,               -- user | system | webhook
    impersonator_id uuid,               -- support impersonation (double-logged, 01 §3)
    request_id      text,               -- correlation id (X-Request-Id)
    action          text NOT NULL,      -- 'document.download', 'receipt.void', …
    entity_type     text,               -- 'document', 'receipt', …
    entity_id       uuid,
    field           text,               -- changed field; NULL for whole-entity actions
    old_value       text,               -- redactable
    new_value       text,               -- redactable
    reason          text,
    metadata        jsonb NOT NULL DEFAULT '{}'
);

CREATE INDEX audit_logs_entity ON audit_logs (tenant_id, entity_type, entity_id, occurred_at DESC);
CREATE INDEX audit_logs_actor  ON audit_logs (tenant_id, actor_id, occurred_at DESC);

ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs FORCE ROW LEVEL SECURITY;
CREATE POLICY audit_logs_tenant_isolation ON audit_logs
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());

-- Append-only: INSERT + SELECT, never UPDATE/DELETE for the runtime role.
GRANT SELECT, INSERT ON audit_logs TO app_rt;

-- +goose Down

DROP TABLE IF EXISTS audit_logs;
