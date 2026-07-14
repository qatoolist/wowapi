-- identity_grant: server-side source of truth for break-glass and impersonation
-- activation state. Framework-owned (D-01 / ADR-W00-E02-S003-001), keyed by an
-- opaque grant ID. Tenant-scoped, FORCE-RLS, writable only by app_platform.
--
-- Scope: tenant-scoped per docs/blueprint/03 §1. The table carries no DELETE
-- grant — status lifecycle (active → revoked/expired) replaces hard deletes.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM identity_grant
-- rollback_plan: goose Down drops identity_grant, its RLS policies, and its partial unique index.
-- +wowapi:end

-- +goose Up

SET LOCAL lock_timeout = '2s';
SET LOCAL statement_timeout = '5s';

CREATE TABLE identity_grant (
    id                   uuid         PRIMARY KEY,
    status               text         NOT NULL
                                      CHECK (status IN ('active', 'revoked', 'expired')),
    tenant_id            uuid         NOT NULL,
    actor_id             uuid         NOT NULL,
    impersonated_user_id uuid,
    approver_id          uuid,
    reason               text,
    activated_at         timestamptz,
    expires_at           timestamptz,
    revoked_at           timestamptz
);

-- At most one active grant per actor. This is the concurrency guard: two
-- concurrent activation attempts for the same actor must result in exactly one
-- succeeding, the other rejected by the unique constraint.
CREATE UNIQUE INDEX identity_grant_one_active_per_actor
    ON identity_grant (actor_id)
    WHERE status = 'active';

-- Tenant isolation: app_rt has no grants here, but the policy is defence-in-depth
-- against any future grant addition. app_platform operates cross-tenant.
ALTER TABLE identity_grant ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity_grant FORCE ROW LEVEL SECURITY;

CREATE POLICY identity_grant_tenant_isolation ON identity_grant
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());

CREATE POLICY identity_grant_platform_all ON identity_grant TO app_platform
    USING (true) WITH CHECK (true);

-- Least-privilege: app_platform owns the grant record; app_rt gets nothing.
GRANT SELECT, INSERT, UPDATE ON identity_grant TO app_platform;

-- +goose Down

DROP POLICY IF EXISTS identity_grant_platform_all ON identity_grant;
DROP POLICY IF EXISTS identity_grant_tenant_isolation ON identity_grant;
ALTER TABLE identity_grant NO FORCE ROW LEVEL SECURITY;
ALTER TABLE identity_grant DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS identity_grant;
