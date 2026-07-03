# Phase 9 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-04.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `make migrate` (00011 on fresh schema) | 0 | migration 11 applied: notification_templates/notifications/notification_deliveries, integration_providers, webhook_endpoints/webhook_events + RLS (hybrid for templates/providers, strict for the rest) + grants |
| 2 | `go build ./kernel/notify/ ./kernel/webhook/ ./kernel/integration/` | 0 | three new packages compile |
| 3 | `DATABASE_URL=… go test ./kernel/notify/` | 0 | template register validation, ValidateBody/RenderBody (allowlist + missingkey=error), Send writes notification+deliveries atomically (rollback → none), tenant-override + locale-fallback resolution, SendPending delivers via fake sender, retry→dead-letter at ceiling, ListForParty, tenant isolation (23 tests) |
| 4 | `DATABASE_URL=… go test ./kernel/webhook/` | 0 | inbound verify+persist, bad signature → unauthenticated + signature_ok recorded, replay idempotent, out-of-window rejected, ProcessInbound handler + retry/DLQ, DispatchOutbound HMAC-sign + POST via fake sender, non-matching skipped, circuit breaker open/half-open with injected clock, tenant isolation (12 tests) |
| 5 | `DATABASE_URL=… go test ./kernel/integration/` | 0 | registry validation, platform + tenant-override resolution, secret-ref credential resolution, plaintext-credential rejected, health checks (skip unconfigured / probe configured), not-configured → NotFound |
| 6 | `go build ./...` | 0 | full module builds after wiring notify/webhook/integration into kernel + module.Context + app boot |
| 7 | `make ci` (host) | 0 | vet, boundary lint, unit, race, build green |
| 8 | `sh scripts/lint_boundaries.sh` | 0 | OK — notify/webhook/integration import kernel/* only; domain-neutral (no leaked domain terms) |
| 9 | (review pass) migration 00011 amended: partial dedup index (SEC-49), notifdel index covers queued+failed (ARCH-73), `next_attempt_at` column (ARCH-75), restrictive hybrid policy (SEC-53), `GRANT INSERT ON events_outbox TO app_platform`; `make migrate` | 0 | version 11 re-applies cleanly |
| 10 | (review pass) `go test ./kernel/integration/` after ARCH-71 fix | 0 | + `TestIntegrationUpsertReturnsPersistedID` (conflict upsert returns the existing, real id) |
| 11 | (review pass) `go test ./kernel/notify/` after ARCH-75/SEC-51/ARCH-77 fixes | 0 | 24 tests: + email HTML-escape, incomplete-vars rejected synchronously, backoff gates failed-delivery re-claim |
| 12 | (review pass) `go test ./kernel/webhook/` after ARCH-70/72 + SEC-49/50/52 fixes | 0 | 15 tests: + RetryOutbound redelivers failed, id-less dedup, failed-sig doesn't block valid, timestamped signature, breaker clears degraded |
| 13 | `make ci` (host, post-fix) | 0 | vet, boundary lint, unit, race, build green |
| 14 | `docker compose run --rm tools go test -p 1 ./...` | 0 | full suite green in-container, serial (warms the migration-hash-changed template) |
| 15 | `make ci-container` (warm template) | 0 | parallel container CI green |
