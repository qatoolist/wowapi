-- DLQ operability (roadmap R4). Dead-lettering already works — jobs land in
-- jobs_queue.status='discarded', events in events_outbox.dispatch_status='dead' —
-- but there was no way to inspect, replay, or purge them. The admin functions
-- (kernel/jobs, kernel/outbox) replay by resetting status; discard by DELETE.
-- app_platform already holds SELECT/UPDATE on both tables (00007); it needs
-- DELETE so an operator can permanently purge a terminal DLQ entry.

-- +goose Up

GRANT DELETE ON jobs_queue TO app_platform;
GRANT DELETE ON events_outbox TO app_platform;

-- +goose Down

REVOKE DELETE ON jobs_queue FROM app_platform;
REVOKE DELETE ON events_outbox FROM app_platform;
