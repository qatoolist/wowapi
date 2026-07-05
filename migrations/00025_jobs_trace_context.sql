-- Carry the originating request's distributed-trace context on each queued job
-- (roadmap O1/CA-9) so the runner/worker can continue the SAME trace when it
-- executes the job asynchronously. trace_context is an opaque W3C `traceparent`
-- string (empty/NULL when tracing is disabled); it is provenance, not business
-- data, and is never part of any hash chain. Mirrors 00024 (events_outbox).

-- +goose Up

ALTER TABLE jobs_queue ADD COLUMN trace_context text;

-- +goose Down

ALTER TABLE jobs_queue DROP COLUMN trace_context;
