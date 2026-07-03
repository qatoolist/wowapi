-- Blueprint 001 core identity: global spine tables tenants, users,
-- user_tenant_access.
--
-- Scope: GLOBAL (docs/blueprint/03 §2 flag G). These tables hold cluster-wide
-- identity data and are accessed only by kernel services — never by product
-- modules directly. They deliberately carry no RLS (row-level security): they
-- are not tenant-scoped so RLS would provide no isolation benefit while adding
-- operational complexity. Tenant-scoped tables (migration 00003+) all carry
-- ENABLE + FORCE RLS per the blueprint convention (03 §1).
--
-- Delete semantics: no DELETE grants are issued. Status lifecycle
-- (active → suspended/closed for tenants; active → disabled for users)
-- replaces hard deletes. True erasure (GDPR) is handled by dedicated redaction
-- jobs, not handler-level DELETEs (docs/blueprint/03 §1 "Soft delete").

-- +goose Up

-- tenants — isolation root; every tenant-scoped row carries tenant_id FK here.
CREATE TABLE tenants (
    id               uuid         PRIMARY KEY,
    slug             text         NOT NULL UNIQUE
                                      CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
    display_name     text         NOT NULL,
    parent_tenant_id uuid         REFERENCES tenants(id),
    status           text         NOT NULL DEFAULT 'active'
                                      CHECK (status IN ('active', 'suspended', 'closed')),
    settings         jsonb        NOT NULL DEFAULT '{}',
    version          int          NOT NULL DEFAULT 1,
    created_at       timestamptz  NOT NULL DEFAULT now(),
    created_by       uuid         NOT NULL,
    updated_at       timestamptz,
    updated_by       uuid
);

-- users — global identity; resolved to a per-tenant acting capacity downstream.
CREATE TABLE users (
    id              uuid         PRIMARY KEY,
    idp_subject     text         NOT NULL UNIQUE,
    -- citext: case-insensitive comparison and unique index; requires the citext
    -- extension installed in 00001_bootstrap.sql.
    email           citext       NOT NULL UNIQUE,
    status          text         NOT NULL DEFAULT 'active'
                                     CHECK (status IN ('active', 'disabled')),
    -- person_party_id is a global hint; per-tenant resolution goes through
    -- acting_capacities (migration 00003+).
    person_party_id uuid,
    version         int          NOT NULL DEFAULT 1,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    created_by      uuid         NOT NULL,
    updated_at      timestamptz,
    updated_by      uuid
);

-- user_tenant_access — user↔tenant membership including cross-tenant grants.
-- Temporal validity: valid_to IS NULL means the grant is still open.
CREATE TABLE user_tenant_access (
    id         uuid         PRIMARY KEY,
    user_id    uuid         NOT NULL REFERENCES users(id),
    tenant_id  uuid         NOT NULL REFERENCES tenants(id),
    kind       text         NOT NULL DEFAULT 'member'
                                CHECK (kind IN ('member', 'support', 'federated_admin')),
    status     text         NOT NULL DEFAULT 'active',
    valid_from timestamptz  NOT NULL DEFAULT now(),
    valid_to   timestamptz,
    created_at timestamptz  NOT NULL DEFAULT now(),
    created_by uuid         NOT NULL
);

-- Partial unique index enforces one active grant per (user, tenant, kind).
-- The WHERE clause means closed grants (valid_to IS NOT NULL) are excluded,
-- allowing historical re-grants without violating uniqueness.
CREATE UNIQUE INDEX uta_active
    ON user_tenant_access (user_id, tenant_id, kind)
    WHERE valid_to IS NULL;

-- Grants — least-privilege, no DELETE (status lifecycle only, see file header).
-- Global identity tables carry NO RLS (03 §1: kernel-service access only), so
-- a grant to app_rt would let any module — which runs arbitrary SQL as app_rt
-- inside a tenant tx — read or tamper with the entire cross-tenant membership
-- graph (SEC-13). They are granted ONLY to app_platform; kernel identity
-- services run platform transactions under that role (pool wiring lands with
-- the first such service in Phase 4). app_rt gets nothing here.
GRANT SELECT, INSERT, UPDATE ON tenants, users, user_tenant_access TO app_platform;

-- +goose Down

-- Drop in reverse dependency order to avoid FK violations.
-- user_tenant_access → users + tenants; users has no FK deps here.
DROP TABLE IF EXISTS user_tenant_access;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;
