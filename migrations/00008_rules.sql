-- Blueprint 03 §5 "005" (on-disk 00008): the rule/configuration engine.
-- rule_definitions is the global registry mirror; rule_versions holds values at
-- a scope with temporal validity + approval status. rule_versions is a
-- platform+tenant hybrid (tenant_id NULL = platform), so it uses the forgiving
-- app_tenant_id_or_null() like roles/policies (D-0045).

-- +goose Up

CREATE TABLE rule_definitions (
    key              text PRIMARY KEY,          -- 'core.retention.audit_days'
    module           text NOT NULL,
    value_schema     jsonb NOT NULL,            -- RuleValueSchema (strict limited grammar: type/enum/bounds/lengths/pattern/items/required) for values
    default_value    jsonb NOT NULL,
    allowed_scopes   text[] NOT NULL DEFAULT '{platform,tenant,org}',
    requires_approval boolean NOT NULL DEFAULT false,
    description      text NOT NULL,
    created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE rule_versions (
    id             uuid PRIMARY KEY,
    rule_key       text NOT NULL REFERENCES rule_definitions(key),
    tenant_id      uuid,                          -- NULL for platform scope
    scope_kind     text NOT NULL CHECK (scope_kind IN ('platform','tenant','org')),
    scope_id       uuid,                          -- org id when scope_kind='org'
    value          jsonb NOT NULL,
    effective_from timestamptz NOT NULL,
    effective_to   timestamptz,
    status         text NOT NULL DEFAULT 'draft'
                       CHECK (status IN ('draft','pending_approval','active','superseded','rejected')),
    approved_by    uuid,
    workflow_instance_id uuid,
    created_at     timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    -- At most one ACTIVE version per (rule, scope) at any instant.
    EXCLUDE USING gist (
        rule_key WITH =, scope_kind WITH =,
        COALESCE(scope_id,'00000000-0000-0000-0000-000000000000'::uuid) WITH =,
        COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid) WITH =,
        tstzrange(effective_from, effective_to) WITH &&
    ) WHERE (status = 'active')
);
CREATE INDEX rule_versions_lookup ON rule_versions (rule_key, scope_kind, effective_from) WHERE status = 'active';

-- rule_definitions: global registry, kernel-managed.
GRANT SELECT ON rule_definitions TO app_rt;
GRANT SELECT, INSERT, UPDATE ON rule_definitions TO app_platform;

-- rule_versions: platform+tenant hybrid RLS. A tenant sees its own + platform
-- versions; writes are constrained to the tenant (platform versions are managed
-- by app_platform). Uses the forgiving tenant fn so a platform connection can
-- write NULL-tenant rows.
ALTER TABLE rule_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE rule_versions FORCE ROW LEVEL SECURITY;
CREATE POLICY rule_versions_tenant ON rule_versions
    USING (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null())
    WITH CHECK (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null());
-- app_platform activates/supersedes versions across all tenants (rule approval
-- is a cross-tenant kernel operation, like the outbox relay — D-0048). Policies
-- are OR'd, so this widens only app_platform, never app_rt.
CREATE POLICY rule_versions_platform_all ON rule_versions TO app_platform USING (true) WITH CHECK (true);
-- app_rt may propose (INSERT draft) + read; activation/approval writes are a
-- kernel/platform concern (app_platform). No DELETE (versions are immutable).
GRANT SELECT, INSERT ON rule_versions TO app_rt;
GRANT SELECT, INSERT, UPDATE ON rule_versions TO app_platform;

-- +goose Down

DROP TABLE IF EXISTS rule_versions;
DROP TABLE IF EXISTS rule_definitions;
