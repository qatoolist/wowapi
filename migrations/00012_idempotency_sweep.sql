-- Idempotency-key expiry sweep (roadmap S5). idempotency_keys carries an
-- expires_at column and reclaim-on-expiry logic, but nothing ever DELETED past
-- rows, so the table grew without bound. The sweep runs cross-tenant as
-- app_platform (one DELETE for all tenants), mirroring the events_outbox relay
-- pattern (00007): a permissive platform policy + the DELETE grant. app_rt keeps
-- its tenant-scoped row lifecycle; only app_platform may sweep across tenants.

-- +goose Up

-- Cross-tenant sweep policy for the platform role (matches outbox_relay_all).
CREATE POLICY idempotency_keys_platform_sweep ON idempotency_keys
    TO app_platform USING (true) WITH CHECK (true);

GRANT SELECT, DELETE ON idempotency_keys TO app_platform;

-- +goose Down

REVOKE SELECT, DELETE ON idempotency_keys FROM app_platform;
DROP POLICY IF EXISTS idempotency_keys_platform_sweep ON idempotency_keys;
