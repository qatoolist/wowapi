-- Tenant-isolation hardening for the job queue (finding M2). jobs_queue and
-- job_runs shipped in 00007 as GLOBAL, grant-only tables: app_rt holds INSERT on
-- jobs_queue (so a module can enqueue in its business tx) but the table had NO
-- row-level security. Because jobs.Enqueue supplies tenant_id from app_tenant_id()
-- that is safe on the happy path, but a hostile app_rt session bound to tenant A
-- could still INSERT a row with an arbitrary tenant_id=B — a write-integrity /
-- defense-in-depth gap (no read leak: app_rt has no SELECT grant here). This
-- migration closes it by FORCEing RLS with a WITH CHECK that pins every app_rt
-- write to its own tenant, exactly mirroring events_outbox (00007). The runner /
-- relay keep working: they connect as app_platform and get a permissive policy so
-- they read + write cross-tenant, and the constant-true qual folds away the
-- app_tenant_id() call (app_platform runs the claim scan with no bound tenant).
--
-- Global (tenant-less) rows are written by EnqueueGlobal, which runs as
-- app_platform and is covered by the permissive *_platform_all policy below — so
-- the app_rt-facing tenant policy does NOT admit tenant_id IS NULL. That keeps it
-- exactly as strict as events_outbox: an app_rt session can only enqueue for its
-- own bound tenant, never a NULL/global sentinel or another tenant, even via raw SQL.
--
-- app_rt grants are UNCHANGED: INSERT-only on jobs_queue, nothing on job_runs.
-- This migration adds no SELECT grant to app_rt (grant-only isolation preserved).

-- +goose Up

ALTER TABLE jobs_queue ENABLE ROW LEVEL SECURITY;
ALTER TABLE jobs_queue FORCE ROW LEVEL SECURITY;
-- app_rt may only enqueue for its OWN tenant; the WITH CHECK rejects a cross-tenant
-- (or NULL/global) INSERT even though app_rt holds the grant. Mirrors events_outbox.
CREATE POLICY jobs_queue_tenant_isolation ON jobs_queue
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
-- The runner (app_platform) claims/completes ALL tenants' jobs cross-tenant.
CREATE POLICY jobs_queue_platform_all ON jobs_queue TO app_platform USING (true) WITH CHECK (true);

ALTER TABLE job_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE job_runs FORCE ROW LEVEL SECURITY;
-- job_runs has no app_rt grant at all; the tenant policy is defence-in-depth so
-- that if a grant is ever added the same tenant pin applies.
CREATE POLICY job_runs_tenant_isolation ON job_runs
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
-- The runner (app_platform) writes the reporting mirror cross-tenant.
CREATE POLICY job_runs_platform_all ON job_runs TO app_platform USING (true) WITH CHECK (true);

-- +goose Down

DROP POLICY IF EXISTS job_runs_platform_all ON job_runs;
DROP POLICY IF EXISTS job_runs_tenant_isolation ON job_runs;
ALTER TABLE job_runs NO FORCE ROW LEVEL SECURITY;
ALTER TABLE job_runs DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS jobs_queue_platform_all ON jobs_queue;
DROP POLICY IF EXISTS jobs_queue_tenant_isolation ON jobs_queue;
ALTER TABLE jobs_queue NO FORCE ROW LEVEL SECURITY;
ALTER TABLE jobs_queue DISABLE ROW LEVEL SECURITY;
