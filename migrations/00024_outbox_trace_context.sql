-- Carry the originating request's distributed-trace context on each outbox event
-- (roadmap O1/CA-9) so the relay/worker can continue the SAME trace when it
-- dispatches the event asynchronously. trace_context is an opaque W3C
-- `traceparent` string (empty/NULL when tracing is disabled); it is provenance,
-- not business data, and is never part of any hash chain.

-- +goose Up

ALTER TABLE events_outbox ADD COLUMN trace_context text;

-- +goose Down

ALTER TABLE events_outbox DROP COLUMN trace_context;
