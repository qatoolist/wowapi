-- Blueprint 03 §4 + 07 §5/§6 (on-disk 00011): notification framework
-- (templates, notifications, deliveries), webhook framework (endpoints, events),
-- and the integration provider registry.
--
-- Config/registry tables (notification_templates, integration_providers,
-- webhook_endpoints) are platform+tenant hybrids written by app_platform — they
-- change behavior (which channels fire, which endpoints receive events, which
-- credentials are used), so per SEC-13 they stay off the module role, which only
-- SELECTs them. notifications is written by modules in a business tx (app_rt).
-- notification_deliveries and webhook_events are append-only to app_rt; their
-- status columns are advanced by the async sender/relay running as app_platform.

-- +goose Up

CREATE TABLE notification_templates (
    id          uuid PRIMARY KEY,
    tenant_id   uuid,                            -- NULL = platform/module default
    key         text NOT NULL,
    channel     text NOT NULL CHECK (channel IN ('inapp','email','sms','whatsapp','push')),
    locale      text NOT NULL DEFAULT 'en',
    subject     text,
    body        text NOT NULL,                   -- Go text/template with allowlisted vars
    status      text NOT NULL DEFAULT 'active',
    version     int  NOT NULL DEFAULT 1,
    created_at  timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at  timestamptz, updated_by uuid
);
CREATE UNIQUE INDEX notification_templates_key ON notification_templates
    (COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid), key, channel, locale);

CREATE TABLE notifications (
    id                 uuid PRIMARY KEY,
    tenant_id          uuid NOT NULL,
    template_key       text NOT NULL,
    recipient_party_id uuid NOT NULL,
    variables          jsonb NOT NULL DEFAULT '{}',
    resource_type      text, resource_id uuid,
    importance         text NOT NULL DEFAULT 'normal' CHECK (importance IN ('normal','important','legal')),
    status             text NOT NULL DEFAULT 'pending',
    created_at         timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL
);
CREATE INDEX notif_recipient ON notifications (tenant_id, recipient_party_id, created_at DESC);

CREATE TABLE notification_deliveries (          -- append-only; sender advances status
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    notification_id uuid NOT NULL REFERENCES notifications(id),
    channel         text NOT NULL,
    destination     text NOT NULL,
    status          text NOT NULL DEFAULT 'queued'
        CHECK (status IN ('queued','sent','delivered','failed','dead')),
    attempts        int NOT NULL DEFAULT 0,
    next_attempt_at timestamptz,                -- backoff gate for retrying failed deliveries
    provider_message_id text, last_error text,
    created_at      timestamptz NOT NULL DEFAULT now(), updated_at timestamptz
);
-- Claim index covers BOTH retryable statuses the sender polls (queued + failed).
CREATE INDEX notifdel_pending ON notification_deliveries (tenant_id, next_attempt_at)
    WHERE status IN ('queued','failed');

CREATE TABLE integration_providers (
    id            uuid PRIMARY KEY,
    tenant_id     uuid,                          -- NULL = platform-registered provider kind
    key           text NOT NULL,
    kind          text NOT NULL,                 -- payment|messaging|identity|storage|device
    config        jsonb NOT NULL DEFAULT '{}',   -- non-secret config
    credential_ref text,                         -- secret-provider key; NEVER plaintext
    status        text NOT NULL DEFAULT 'active',
    version       int  NOT NULL DEFAULT 1,
    created_at    timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at    timestamptz, updated_by uuid
);
CREATE UNIQUE INDEX integration_providers_key ON integration_providers
    (COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid), key);

CREATE TABLE webhook_endpoints (
    id               uuid PRIMARY KEY,
    tenant_id        uuid NOT NULL,
    direction        text NOT NULL CHECK (direction IN ('inbound','outbound')),
    provider_id      uuid REFERENCES integration_providers(id),
    url              text,                        -- outbound only
    secret_ref       text NOT NULL,               -- secret-provider key; NEVER plaintext
    signature_scheme text NOT NULL DEFAULT 'hmac-sha256',
    subscribed_events text[],                     -- outbound only
    status           text NOT NULL DEFAULT 'active',
    version          int  NOT NULL DEFAULT 1,
    created_at       timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at       timestamptz, updated_by uuid
);
CREATE INDEX whep_direction ON webhook_endpoints (tenant_id, direction, status);

CREATE TABLE webhook_events (                    -- append-only; processing advances status
    id               uuid PRIMARY KEY,
    tenant_id        uuid NOT NULL,
    endpoint_id      uuid NOT NULL REFERENCES webhook_endpoints(id),
    direction        text NOT NULL,
    external_event_id text,                       -- provider's id → replay protection
    event_type       text NOT NULL,
    payload          jsonb NOT NULL,
    signature_ok     boolean,
    received_at      timestamptz NOT NULL DEFAULT now(),
    delivery_status  text NOT NULL DEFAULT 'pending'
        CHECK (delivery_status IN ('pending','processed','delivered','failed','dead')),
    attempts         int NOT NULL DEFAULT 0, next_attempt_at timestamptz, last_error text
);
-- Replay dedup: a PARTIAL unique index so a NULL external_event_id (a provider
-- that omits an id) does not defeat the constraint (plain UNIQUE treats NULLs as
-- distinct — SEC-49). The webhook service synthesizes a body-hash id when none is
-- supplied, so real events are always deduped; this index is the DB backstop.
CREATE UNIQUE INDEX webhook_events_dedup ON webhook_events (endpoint_id, external_event_id)
    WHERE external_event_id IS NOT NULL;
CREATE INDEX whev_pending ON webhook_events (tenant_id, delivery_status, next_attempt_at)
    WHERE delivery_status IN ('pending','failed');

-- notification_templates + integration_providers: platform+tenant hybrids (a
-- tenant sees its own rows + platform defaults) — forgiving app_tenant_id_or_null.
-- +goose StatementBegin
DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['notification_templates','integration_providers']
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('CREATE POLICY %I ON %I USING (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null()) WITH CHECK (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null())', t||'_tenant', t);
        -- RESTRICTIVE backstop (SEC-53): a NULL-tenant (platform) row may be
        -- WRITTEN only from an UNBOUND session — a tenant-bound app_platform
        -- session (app_tenant_id_or_null() non-NULL) cannot forge a platform row.
        EXECUTE format('CREATE POLICY %I ON %I AS RESTRICTIVE USING (true) WITH CHECK (tenant_id IS NOT NULL OR app_tenant_id_or_null() IS NULL)', t||'_platform_write', t);
    END LOOP;
END
$$;
-- +goose StatementEnd

-- notifications / deliveries / endpoints / events: strict tenant RLS.
-- +goose StatementBegin
DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['notifications','notification_deliveries','webhook_endpoints','webhook_events']
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('CREATE POLICY %I ON %I USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id())', t||'_tenant_isolation', t);
    END LOOP;
END
$$;
-- +goose StatementEnd

-- Config/registry tables: module role reads; app_platform writes (SEC-13).
GRANT SELECT ON notification_templates TO app_rt;
GRANT SELECT, INSERT, UPDATE ON notification_templates TO app_platform;
GRANT SELECT ON integration_providers TO app_rt;
GRANT SELECT, INSERT, UPDATE ON integration_providers TO app_platform;
GRANT SELECT ON webhook_endpoints TO app_rt;
GRANT SELECT, INSERT, UPDATE ON webhook_endpoints TO app_platform;

-- notifications: modules enqueue in a business tx.
GRANT SELECT, INSERT ON notifications TO app_rt;
GRANT SELECT, UPDATE ON notifications TO app_platform; -- sender marks sent/failed

-- notification_deliveries: append-only to app_rt; the async sender (app_platform)
-- advances status/attempts.
GRANT SELECT, INSERT ON notification_deliveries TO app_rt;
GRANT SELECT, INSERT, UPDATE ON notification_deliveries TO app_platform;

-- webhook_events: append-only to app_rt (inbound receipt); the processing relay
-- (app_platform) advances delivery_status/attempts and dispatches outbound.
GRANT SELECT, INSERT ON webhook_events TO app_rt;
GRANT SELECT, INSERT, UPDATE ON webhook_events TO app_platform;

-- app_platform-run workers (webhook inbound handlers, delivery sender) may need
-- to EMIT events into the outbox — e.g. a legal-importance delivery audit event,
-- or a module's inbound-webhook handler raising a domain event. Migration 00007
-- granted app_platform only SELECT/UPDATE (relay marks dispatched); add INSERT so
-- a tenant-bound worker can append events (the outbox_relay_all WITH CHECK admits
-- it; the outbox Writer stamps tenant_id from app_tenant_id()).
GRANT INSERT ON events_outbox TO app_platform;

-- +goose Down

DROP TABLE IF EXISTS webhook_events;
DROP TABLE IF EXISTS webhook_endpoints;
DROP TABLE IF EXISTS integration_providers;
DROP TABLE IF EXISTS notification_deliveries;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_templates;
