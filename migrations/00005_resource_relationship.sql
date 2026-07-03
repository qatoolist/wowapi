-- Blueprint 003 (D-0035): the resource registry (global resource_types +
-- tenant resources mirror) and the relationship graph (global relationship_types
-- + tenant relationships). Global registries are app_platform-only (no RLS,
-- kernel-service access, SEC-13); tenant tables get standard RLS + app_rt.

-- +goose Up

CREATE TABLE resource_types (            -- global registry, synced from module registration at boot
    key         text PRIMARY KEY,        -- 'requests.request'
    module      text NOT NULL,
    description text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE relationship_types (        -- global registry
    key          text PRIMARY KEY,       -- 'core.owner_of'
    module       text NOT NULL,
    subject_kind text NOT NULL CHECK (subject_kind IN ('party','resource','capacity')),
    object_kind  text NOT NULL CHECK (object_kind  IN ('party','resource')),
    cardinality  text NOT NULL DEFAULT 'many' CHECK (cardinality IN ('one','many')),
    description  text NOT NULL
);

CREATE TABLE resources (                 -- tenant mirror; id == owning module row id
    id            uuid PRIMARY KEY,
    tenant_id     uuid NOT NULL,
    resource_type text NOT NULL REFERENCES resource_types(key),
    org_id        uuid REFERENCES organizations(id),
    label         text NOT NULL,
    status        text NOT NULL DEFAULT 'active',
    version       int  NOT NULL DEFAULT 1,
    created_at    timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at    timestamptz, updated_by uuid
);
CREATE INDEX res_by_type ON resources (tenant_id, resource_type, status);

CREATE TABLE relationships (
    id           uuid PRIMARY KEY,
    tenant_id    uuid NOT NULL,
    rel_type     text NOT NULL REFERENCES relationship_types(key),
    subject_kind text NOT NULL,
    subject_id   uuid NOT NULL,
    object_kind  text NOT NULL,
    object_id    uuid NOT NULL,
    attributes   jsonb NOT NULL DEFAULT '{}',
    valid_from   timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
    version      int NOT NULL DEFAULT 1,
    created_at   timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at   timestamptz, updated_by uuid
);
-- Cover time-boxed-but-active edges (valid_to > now), not just open-ended ones
-- — a partial "WHERE valid_to IS NULL" would seq-scan temporary relationships
-- (review finding ARCH-42).
CREATE INDEX rel_obj ON relationships (tenant_id, object_kind, object_id, rel_type, valid_from, valid_to);
CREATE INDEX rel_sub ON relationships (tenant_id, subject_kind, subject_id, rel_type, valid_from, valid_to);

-- Global registries: kernel-service access only, no RLS (03 §1).
GRANT SELECT ON resource_types, relationship_types TO app_rt;   -- read-only lookups on hot paths
GRANT SELECT, INSERT, UPDATE ON resource_types, relationship_types TO app_platform;

-- Tenant tables: standard RLS on both.
-- +goose StatementBegin
DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['resources','relationships']
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('CREATE POLICY %I ON %I USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id())', t||'_tenant_isolation', t);
    END LOOP;
END
$$;
-- +goose StatementEnd

-- Grants — deliberately asymmetric (SEC-24):
--   resources: app_rt writes the mirror for its OWN aggregates via the
--     Registrar (INSERT/UPDATE). org_id transitions are audit-worthy but the
--     mirror write is the module's own; cross-tenant is blocked by RLS.
--   relationships: a granted_via edge is an AUTHORIZATION INPUT — an actor that
--     holds an edge is granted the mapped permission on its object. If the
--     shared app_rt role could write edges, any module could self-grant by
--     inserting an edge naming its own capacity. Edge creation is therefore a
--     kernel/platform capability: app_rt gets SELECT only; writes are
--     app_platform (an audited edge-management service, wired Phase 4+).
GRANT SELECT, INSERT, UPDATE ON resources TO app_rt;
GRANT SELECT ON relationships TO app_rt;
GRANT SELECT, INSERT, UPDATE ON relationships TO app_platform;
GRANT SELECT ON resources TO app_platform;

-- +goose Down

DROP TABLE IF EXISTS relationships;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS relationship_types;
DROP TABLE IF EXISTS resource_types;
