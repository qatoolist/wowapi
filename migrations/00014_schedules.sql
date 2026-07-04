-- Recurring scheduler (roadmap E5) + leader-safe kernel sweeps (R3). The SLA
-- sweeper and the idempotency sweep existed as methods but nothing ran them on a
-- schedule, and nothing stopped N worker replicas from all firing at once. This
-- table drives a fixed-interval scheduler: each due tick is claimed by an atomic
-- conditional UPDATE on the row (only one replica wins per interval), so tasks
-- run leader-safe without a separate election. Global kernel table (no RLS),
-- owned by app_platform, like jobs_queue.

-- +goose Up

CREATE TABLE schedules (
    name             text PRIMARY KEY,                 -- stable task identity, e.g. 'kernel.workflow.sla'
    interval_seconds int  NOT NULL CHECK (interval_seconds >= 1),
    next_run_at      timestamptz NOT NULL DEFAULT now(), -- earliest next eligible run
    last_run_at      timestamptz,
    enabled          boolean NOT NULL DEFAULT true
);

GRANT SELECT, INSERT, UPDATE, DELETE ON schedules TO app_platform;

-- +goose Down

DROP TABLE IF EXISTS schedules;
