-- Blueprint 000 bootstrap: extensions, cluster-wide roles, app_tenant_id().
-- This migration must complete before any table migrations (00002+).
--
-- Role strategy (D-0023, revised for SEC-11/SEC-12):
--   app_rt        — runtime identity for all application queries; created NOLOGIN
--                   so NO password ships in migrations. Deployed processes MUST
--                   connect AS a non-superuser login mapped to app_rt (ops
--                   provision the credential); the runtime must NOT be a superuser
--                   doing SET ROLE, because a superuser login can escape any
--                   SET ROLE via RESET ROLE and defeat RLS. The testkit models
--                   this by granting app_rt a local-only LOGIN out-of-band (never
--                   committed) and connecting as it.
--   app_platform  — kernel/support-ops role; holds the grants on the global
--                   identity tables (which have no RLS). NOLOGIN; a dedicated
--                   platform pool authenticates as it once kernel identity
--                   services exist (Phase 4).
--   app_migrate   — NOT created here. The migration runner connects with whatever
--                   owner credentials ops provide (e.g. the superuser or a
--                   dedicated owner login); no separate role is needed in Phase 2.

-- +goose Up

-- Extensions are cluster-wide objects: CREATE IF NOT EXISTS is safe to run
-- repeatedly and across databases sharing the same cluster.
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Roles are cluster-wide: we use an idempotent DO block to avoid errors on
-- repeated runs or across databases sharing a cluster. ALTER ROLE …
-- NOLOGIN is a no-op if the role already carries that attribute.
-- +goose StatementBegin
DO $$
BEGIN
    -- Create each role NOLOGIN only when ABSENT; do NOT re-assert attributes on
    -- an existing role. The LOGIN attribute is owned OUT OF BAND — ops grant the
    -- runtime login in production, and the test kit grants a local login. A
    -- migration that reset LOGIN→NOLOGIN on every (re-)run would fight that,
    -- flipping the cluster-global role mid-run under parallel test packages.
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_rt') THEN
        CREATE ROLE app_rt NOLOGIN;
    END IF;

    -- app_platform: support-ops read-only role.
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_platform') THEN
        CREATE ROLE app_platform NOLOGIN;
    END IF;
EXCEPTION
    -- Roles are CLUSTER-GLOBAL. Concurrent migrations (parallel test packages
    -- each building a fresh template, or api+worker+migrate racing at deploy)
    -- can update the same pg_authid tuple in overlapping transactions, raising
    -- "tuple concurrently updated". The change is idempotent (the role ends up
    -- NOLOGIN regardless of which session wins), so this specific catalog race is
    -- benign and retried; anything else re-raises.
    WHEN OTHERS THEN
        IF SQLERRM LIKE '%tuple concurrently updated%' THEN
            NULL;
        ELSE
            RAISE;
        END IF;
END
$$;
-- +goose StatementEnd

-- app_tenant_id() returns the current session's tenant UUID.
--
-- SECURITY: current_setting is called WITHOUT the missing_ok=true second
-- argument. If app.tenant_id has not been set on this session the function
-- raises an ERROR — fail closed. Never pass missing_ok => true here; that
-- would silently return NULL and cause RLS policies to admit or reject all
-- rows non-deterministically instead of hard-failing.
CREATE OR REPLACE FUNCTION app_tenant_id() RETURNS uuid LANGUAGE sql STABLE AS
$$ SELECT current_setting('app.tenant_id')::uuid $$;

-- app_tenant_id_or_null() is the FORGIVING variant for the few hybrid tables
-- that hold BOTH platform-template rows (tenant_id IS NULL) and tenant rows
-- (roles, policies). It returns NULL when app.tenant_id is unset instead of
-- raising, so a platform/catalog connection (app_platform, no tenant bound) can
-- write NULL-tenant templates without the strict function aborting the
-- statement. Pure tenant tables keep using app_tenant_id() (fail-closed/loud).
CREATE OR REPLACE FUNCTION app_tenant_id_or_null() RETURNS uuid LANGUAGE sql STABLE AS
$$ SELECT nullif(current_setting('app.tenant_id', true), '')::uuid $$;

-- Schema-level grants: allow both runtime roles to see public objects.
-- Table-level grants are applied in each migration that creates tables (see
-- 00002_core_identity.sql). No blanket GRANT ALL — least-privilege per table.
GRANT USAGE ON SCHEMA public TO app_rt, app_platform;

-- +goose Down

-- Drop the session helper. Dependent RLS policies must already be gone (they
-- live in later migrations whose Down sections run first under goose rollback).
DROP FUNCTION IF EXISTS app_tenant_id_or_null();
DROP FUNCTION IF EXISTS app_tenant_id();

-- Roles and extensions are intentionally left in place on Down.
--
-- Reason — roles: pg roles are cluster-scoped. Dropping app_rt or app_platform
-- here would break any sibling database in the same cluster that also uses this
-- framework. goose Down for bootstrap is therefore best-effort schema cleanup
-- only; role lifecycle belongs to cluster-level DBA procedures.
--
-- Reason — extensions: citext and btree_gist may be in use by other schemas or
-- databases. Extension removal requires explicit DBA action.
REVOKE USAGE ON SCHEMA public FROM app_rt, app_platform;
