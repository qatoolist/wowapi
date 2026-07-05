-- Carry the originating request's distributed-trace context on each notification
-- delivery envelope (roadmap O1/CA-9) so the async sender can continue the SAME
-- trace when it delivers. trace_context is an opaque W3C `traceparent` string
-- (empty/NULL when tracing is disabled); it is provenance, not business data.
-- Mirrors 00024 (events_outbox) / 00025 (jobs_queue). app_rt writes it on Send
-- (INSERT grant already covers the new column); app_platform reads it on
-- SendPending (SELECT grant already covers it).

-- +goose Up

ALTER TABLE notification_deliveries ADD COLUMN trace_context text;

-- +goose Down

ALTER TABLE notification_deliveries DROP COLUMN trace_context;
