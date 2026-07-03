-- Blueprint 002 (D-0035): organizations, party supertype + subtypes, contacts,
-- acting capacities. All tenant-scoped: ENABLE + FORCE RLS with the standard
-- tenant_id = app_tenant_id() policy (03 §1). app_rt gets row CRUD minus DELETE
-- (status/temporal lifecycle, never hard delete).

-- +goose Up

CREATE TABLE organizations (
    id            uuid PRIMARY KEY,
    tenant_id     uuid NOT NULL REFERENCES tenants(id),
    parent_org_id uuid REFERENCES organizations(id),
    name          text NOT NULL,
    kind          text NOT NULL DEFAULT 'org',
    status        text NOT NULL DEFAULT 'active',
    version       int  NOT NULL DEFAULT 1,
    created_at    timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at    timestamptz, updated_by uuid,
    UNIQUE (tenant_id, parent_org_id, name)
);

CREATE TABLE parties (
    id           uuid PRIMARY KEY,
    tenant_id    uuid NOT NULL REFERENCES tenants(id),
    kind         text NOT NULL CHECK (kind IN ('person','legal_entity')),
    display_name text NOT NULL,
    status       text NOT NULL DEFAULT 'active',
    version      int  NOT NULL DEFAULT 1,
    created_at   timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at   timestamptz, updated_by uuid
);

CREATE TABLE persons (
    party_id    uuid PRIMARY KEY REFERENCES parties(id),
    tenant_id   uuid NOT NULL,
    given_name  text NOT NULL,
    family_name text,
    dob         date,
    locale      text
);

CREATE TABLE legal_entities (
    party_id        uuid PRIMARY KEY REFERENCES parties(id),
    tenant_id       uuid NOT NULL,
    legal_name      text NOT NULL,
    registration_no text,
    jurisdiction    text
);

CREATE TABLE party_contacts (
    id         uuid PRIMARY KEY,
    tenant_id  uuid NOT NULL,
    party_id   uuid NOT NULL REFERENCES parties(id),
    kind       text NOT NULL CHECK (kind IN ('email','phone','address','other')),
    value      text NOT NULL,
    is_primary boolean NOT NULL DEFAULT false,
    verified_at timestamptz,
    version    int NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at timestamptz, updated_by uuid,
    UNIQUE (tenant_id, party_id, kind, value)
);

CREATE TABLE acting_capacities (
    id         uuid PRIMARY KEY,
    tenant_id  uuid NOT NULL,
    user_id    uuid NOT NULL REFERENCES users(id),
    party_id   uuid REFERENCES parties(id),
    label      text NOT NULL,
    status     text NOT NULL DEFAULT 'active',
    valid_from timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
    version    int NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at timestamptz, updated_by uuid
);
CREATE UNIQUE INDEX cap_active ON acting_capacities (tenant_id, user_id, label) WHERE valid_to IS NULL;

-- RLS: every tenant-scoped table isolates on app_tenant_id() for USING and
-- WITH CHECK, and forces the policy on the table owner too (03 §1).
-- +goose StatementBegin
DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['organizations','parties','persons','legal_entities','party_contacts','acting_capacities']
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('CREATE POLICY %I ON %I USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id())', t||'_tenant_isolation', t);
        EXECUTE format('GRANT SELECT, INSERT, UPDATE ON %I TO app_rt', t);
    END LOOP;
END
$$;
-- +goose StatementEnd

CREATE INDEX org_parent ON organizations (tenant_id, parent_org_id);
CREATE INDEX contacts_party ON party_contacts (tenant_id, party_id);
CREATE INDEX cap_user ON acting_capacities (tenant_id, user_id) WHERE valid_to IS NULL;

-- +goose Down

DROP TABLE IF EXISTS acting_capacities;
DROP TABLE IF EXISTS party_contacts;
DROP TABLE IF EXISTS legal_entities;
DROP TABLE IF EXISTS persons;
DROP TABLE IF EXISTS parties;
DROP TABLE IF EXISTS organizations;
