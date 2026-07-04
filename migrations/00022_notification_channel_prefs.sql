-- Per-user notification channel preferences (roadmap R5). A recipient can opt out
-- of a channel (e.g. email or SMS); Send skips channels a party has disabled.
-- Absence of a row means enabled (opt-out, not opt-in), so existing behaviour is
-- unchanged until a preference is set. Tenant-scoped under RLS.

-- +goose Up

CREATE TABLE notification_channel_prefs (
    tenant_id  uuid NOT NULL,
    party_id   uuid NOT NULL,
    channel    text NOT NULL CHECK (channel IN ('inapp','email','sms','whatsapp','push')),
    enabled    boolean NOT NULL DEFAULT true,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, party_id, channel)
);

ALTER TABLE notification_channel_prefs ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_channel_prefs FORCE ROW LEVEL SECURITY;
CREATE POLICY notification_channel_prefs_tenant_isolation ON notification_channel_prefs
    USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT, UPDATE, DELETE ON notification_channel_prefs TO app_rt;

-- +goose Down

DROP TABLE IF EXISTS notification_channel_prefs;
