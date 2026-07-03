-- Blueprint 03 §5 "009" (pulled to on-disk 00007): the transactional outbox,
-- the consumer inbox, the job queue, and the job_runs reporting mirror.
-- events_outbox + processed_events are tenant-scoped (RLS); the relay reads
-- across tenants as app_platform (D-0048). jobs_queue + job_runs are global
-- (tenant travels in the payload) and kernel-only.

-- +goose Up

-- Transactional outbox: business writes and the event insert commit together.
CREATE TABLE events_outbox (
    id              uuid PRIMARY KEY,                 -- uuidv7 == event id
    tenant_id       uuid NOT NULL,
    event_type      text NOT NULL,
    schema_version  int  NOT NULL DEFAULT 1,
    resource_type   text,
    resource_id     uuid,
    actor           jsonb NOT NULL DEFAULT '{}',
    payload         jsonb NOT NULL DEFAULT '{}',
    occurred_at     timestamptz NOT NULL DEFAULT now(),
    dispatch_status text NOT NULL DEFAULT 'pending'
                        CHECK (dispatch_status IN ('pending','dispatched','failed','dead')),
    dispatched_at   timestamptz,
    failed_at       timestamptz,                      -- time of the last failed attempt (cooldown key)
    attempts        int NOT NULL DEFAULT 0,
    max_attempts    int NOT NULL DEFAULT 10,           -- poison ceiling → 'dead' (event DLQ)
    last_error      text,
    created_by      uuid NOT NULL
);
-- Per-aggregate dispatch order + a partial index for the relay's claim scan.
CREATE INDEX outbox_pending ON events_outbox (occurred_at) WHERE dispatch_status = 'pending';
CREATE INDEX outbox_aggregate ON events_outbox (tenant_id, resource_type, resource_id, occurred_at);

ALTER TABLE events_outbox ENABLE ROW LEVEL SECURITY;
ALTER TABLE events_outbox FORCE ROW LEVEL SECURITY;
-- Tenant isolation for module writes/reads in the business tx (app_rt).
CREATE POLICY outbox_tenant_isolation ON events_outbox
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
-- The relay (app_platform) reads + marks ALL tenants' events cross-tenant.
CREATE POLICY outbox_relay_all ON events_outbox TO app_platform USING (true) WITH CHECK (true);

GRANT SELECT, INSERT ON events_outbox TO app_rt;             -- write in business tx; no UPDATE (append-only for modules)
GRANT SELECT, UPDATE ON events_outbox TO app_platform;       -- relay marks dispatched

-- Consumer inbox: idempotent handlers dedup on (handler, event_id).
CREATE TABLE processed_events (
    handler      text NOT NULL,
    event_id     uuid NOT NULL,
    tenant_id    uuid NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (handler, event_id)
);
ALTER TABLE processed_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE processed_events FORCE ROW LEVEL SECURITY;
CREATE POLICY processed_events_tenant_isolation ON processed_events
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT ON processed_events TO app_rt;

-- Job queue (global; tenant in payload). Kernel-only — no module direct access.
CREATE TABLE jobs_queue (
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind         text NOT NULL,
    tenant_id    uuid,                                -- NULL for global jobs
    payload      jsonb NOT NULL DEFAULT '{}',
    status       text NOT NULL DEFAULT 'available'
                     CHECK (status IN ('available','running','completed','discarded')),
    attempts     int NOT NULL DEFAULT 0,
    max_attempts int NOT NULL DEFAULT 5,
    run_at       timestamptz NOT NULL DEFAULT now(),  -- earliest eligible time (backoff)
    locked_at    timestamptz,
    last_error   text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    finished_at  timestamptz
);
-- Claim scan: eligible, ordered by run_at.
CREATE INDEX jobs_available ON jobs_queue (run_at) WHERE status = 'available';

-- job_runs: reporting mirror (append + status), global.
CREATE TABLE job_runs (
    id          uuid PRIMARY KEY,
    tenant_id   uuid,
    job_kind    text NOT NULL,
    job_id      bigint,
    status      text NOT NULL CHECK (status IN ('running','succeeded','failed','dead')),
    started_at  timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    error       text,
    progress    jsonb
);
CREATE INDEX job_runs_kind ON job_runs (job_kind, started_at);

-- jobs_queue + job_runs are written by the kernel job runner, which connects as
-- app_platform (a privileged kernel role), never as app_rt. Modules enqueue via
-- the Runner API inside their business tx (app_rt) — the enqueue INSERT into
-- jobs_queue therefore needs an app_rt grant too, but ONLY insert.
GRANT INSERT ON jobs_queue TO app_rt;                        -- enqueue in business tx
GRANT SELECT, INSERT, UPDATE ON jobs_queue TO app_platform;  -- runner claim/complete
GRANT SELECT, INSERT, UPDATE ON job_runs TO app_platform;

-- +goose Down

DROP TABLE IF EXISTS job_runs;
DROP TABLE IF EXISTS jobs_queue;
DROP TABLE IF EXISTS processed_events;
DROP TABLE IF EXISTS events_outbox;
