-- Blueprint 004 (D-0035): the authorization spine — permissions catalog (global),
-- roles (platform templates + tenant rows), role_permissions, actor_assignments,
-- policies + policy_conditions. Deny-by-default is enforced in code (kernel/authz);
-- these tables only store grants. Roles/policies span global+tenant, so they use a
-- COALESCE(tenant_id, nil-uuid) uniqueness trick and RLS that admits the tenant's
-- rows plus platform (NULL tenant) templates.

-- +goose Up

CREATE TABLE permissions (               -- global catalog, synced from module registration
    key         text PRIMARY KEY,        -- 'document.read'
    module      text NOT NULL,
    description text NOT NULL,
    sensitive   boolean NOT NULL DEFAULT false   -- denials always audited when true
);

CREATE TABLE roles (
    id         uuid PRIMARY KEY,
    tenant_id  uuid REFERENCES tenants(id),      -- NULL = platform template
    key        text NOT NULL,
    name       text NOT NULL,
    is_system  boolean NOT NULL DEFAULT false,
    status     text NOT NULL DEFAULT 'active',
    version    int  NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at timestamptz, updated_by uuid
);
CREATE UNIQUE INDEX roles_key ON roles (COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid), key);

CREATE TABLE role_permissions (
    role_id        uuid NOT NULL REFERENCES roles(id),
    permission_key text NOT NULL REFERENCES permissions(key),
    PRIMARY KEY (role_id, permission_key)
);

CREATE TABLE actor_assignments (
    id             uuid PRIMARY KEY,
    tenant_id      uuid NOT NULL,
    capacity_id    uuid REFERENCES acting_capacities(id),
    system_actor   text,                          -- exactly one of capacity_id / system_actor
    role_id        uuid NOT NULL REFERENCES roles(id),
    scope_kind     text NOT NULL CHECK (scope_kind IN ('tenant','org','resource_type','resource')),
    scope_id       uuid,                          -- null for tenant scope
    scope_type     text,                          -- resource_type key when scope_kind='resource_type'
    valid_from     timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
    granted_by     uuid NOT NULL,
    delegated_from uuid REFERENCES actor_assignments(id),
    reason         text,
    version        int NOT NULL DEFAULT 1,
    created_at     timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    CHECK ((capacity_id IS NULL) <> (system_actor IS NULL)),
    -- Scope integrity (SEC-26/SEC-29): a resource_type-scoped grant MUST name a
    -- type, org/resource-scoped grants MUST name an id — a NULL there would let
    -- covers() over-grant via an empty-string / nil match.
    CHECK (scope_kind <> 'resource_type' OR scope_type IS NOT NULL),
    CHECK (scope_kind <> 'resource'      OR scope_id   IS NOT NULL),
    CHECK (scope_kind <> 'org'           OR scope_id   IS NOT NULL)
);
-- Indexes cover the temporal predicate the hot-path queries actually use
-- (valid_to IS NULL OR valid_to > now) — a partial "WHERE valid_to IS NULL"
-- would exclude time-boxed-but-active grants (delegation, +30d auditor) and
-- force a seq scan on exactly those rows (review finding ARCH-42). Both actor
-- selectors (capacity and system) are indexed.
CREATE INDEX asg_actor ON actor_assignments (tenant_id, capacity_id, valid_from, valid_to);
CREATE INDEX asg_system ON actor_assignments (tenant_id, system_actor, valid_from, valid_to);
CREATE INDEX asg_scope ON actor_assignments (tenant_id, scope_kind, scope_id, valid_to);

CREATE TABLE policies (
    id                       uuid PRIMARY KEY,
    tenant_id                uuid REFERENCES tenants(id),   -- NULL = platform policy
    key                      text NOT NULL,
    effect                   text NOT NULL CHECK (effect IN ('allow','deny')),
    applies_to_permission    text REFERENCES permissions(key),
    applies_to_resource_type text REFERENCES resource_types(key),
    priority                 int NOT NULL DEFAULT 100,
    status                   text NOT NULL DEFAULT 'active',
    version                  int NOT NULL DEFAULT 1,
    created_at               timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at               timestamptz, updated_by uuid
);

CREATE TABLE policy_conditions (
    id         uuid PRIMARY KEY,
    policy_id  uuid NOT NULL REFERENCES policies(id),
    attribute  text NOT NULL,      -- 'resource.status', 'actor.relationship', 'env.time_of_day'
    op         text NOT NULL CHECK (op IN ('eq','neq','in','not_in','contains','within','gte','lte')),
    value      jsonb NOT NULL
);

-- Permissions catalog: global, kernel-managed.
GRANT SELECT ON permissions TO app_rt;
GRANT SELECT, INSERT, UPDATE ON permissions TO app_platform;

-- roles/role_permissions/policies/policy_conditions span platform (tenant_id NULL)
-- + tenant rows. RLS admits the current tenant's rows AND platform templates for
-- reads; writes are constrained to the current tenant (platform rows are managed
-- by app_platform / migrations, never app_rt).
-- roles/policies hold platform templates (tenant_id NULL) + tenant rows. They
-- use the forgiving app_tenant_id_or_null() so a platform/catalog connection
-- (app_platform, no tenant bound) can read and write NULL-tenant templates
-- without the strict function aborting the statement, while a tenant connection
-- still sees only its own rows + platform templates. app_rt cannot write these
-- at all (SELECT-only grant), so allowing NULL-tenant writes here does not widen
-- app_rt.
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles FORCE ROW LEVEL SECURITY;
CREATE POLICY roles_tenant_read ON roles
    USING (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null())
    WITH CHECK (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null());

ALTER TABLE policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE policies FORCE ROW LEVEL SECURITY;
CREATE POLICY policies_tenant_read ON policies
    USING (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null())
    WITH CHECK (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null());

-- actor_assignments is strictly tenant-scoped.
ALTER TABLE actor_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE actor_assignments FORCE ROW LEVEL SECURITY;
CREATE POLICY actor_assignments_tenant_isolation ON actor_assignments
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());

-- PRIVILEGE-ESCALATION DEFENSE (critical): the authorization spine is the very
-- data the evaluator trusts. app_rt — the role arbitrary MODULE SQL runs as —
-- must NOT be able to write it, or a module could grant itself any role by
-- INSERTing an actor_assignment (or a role + role_permissions) and then be
-- authorized for it. app_rt therefore gets SELECT ONLY on the whole spine;
-- every write (role/assignment/policy management) is a kernel/platform
-- capability running as app_platform (a distinct pool, wired Phase 4+), or a
-- migration. RLS still scopes app_rt's reads to its tenant (+ platform templates).
GRANT SELECT ON roles TO app_rt;
GRANT SELECT ON role_permissions TO app_rt;
GRANT SELECT ON actor_assignments TO app_rt;
GRANT SELECT ON policies TO app_rt;
GRANT SELECT ON policy_conditions TO app_rt;

GRANT SELECT, INSERT, UPDATE ON roles TO app_platform;
GRANT SELECT, INSERT, UPDATE, DELETE ON role_permissions TO app_platform;
GRANT SELECT, INSERT, UPDATE ON actor_assignments TO app_platform;
GRANT SELECT, INSERT, UPDATE ON policies TO app_platform;
GRANT SELECT, INSERT, UPDATE, DELETE ON policy_conditions TO app_platform;

-- +goose Down

DROP TABLE IF EXISTS policy_conditions;
DROP TABLE IF EXISTS policies;
DROP TABLE IF EXISTS actor_assignments;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
