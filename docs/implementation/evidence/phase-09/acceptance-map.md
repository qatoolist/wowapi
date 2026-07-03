# Phase 9 — Acceptance Map

Phase 9 exit criteria (Goal 2 Phase 9 + phase-plan row 9 + blueprint 07 §5/§6) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | Notification templates (Go text/template, allowlisted vars, seed-validated) | `kernel/notify/registry.go` TemplateSpec + `ValidateBody` (unknown var rejected at register); `TestValidateBody*` |
| 2 | Template resolution tenant→platform + locale fallback | `notify.Service` template lookup; `TestTemplateResolutionTenantOverridesBeatsPlatform`, `TestTemplateLocaleWithFallback` |
| 3 | **Send writes notification + one delivery per channel in the business tx** | `Service.Send`; `TestSendWritesNotificationAndDeliveries`, `TestSendAtomicRollback` (rollback → nothing persists) |
| 4 | Channel resolution (explicit ChannelDest; in-app destination = party) | `Message.Channels []ChannelDest`; in-app auto-destination |
| 5 | ChannelSender port + async delivery | `ChannelSender` interface + FakeSender; `Service.SendPending` (app_platform, tenant-bound) |
| 6 | **Delivery retry + dead-letter** | `SendPending` claims queued/failed, advances status, dead-letters at maxAttempts; `TestSendPendingRetriesAndDeadLetters` |
| 7 | In-app inbox list | `Service.ListForParty`; `TestListForParty` |
| 8 | Webhook inbound: verify signature + replay + persist + async | `webhook.Service.HandleInbound` (verifier registry, replay via UNIQUE + timestamp window); `ProcessInbound` (handler + retry/DLQ) |
| 9 | **Bad signature rejected, body not logged, signature_ok recorded** | `TestHandleInbound*` bad-sig → KindUnauthenticated; signature_ok=false persisted |
| 10 | Replay protection (idempotent) | UNIQUE(endpoint_id, external_event_id) + ±window; `TestHandleInbound*Replay` |
| 11 | Webhook outbound: fan events → endpoints, HMAC-signed | `Service.DispatchOutbound` — X-Signature/X-Timestamp/X-Event-Id headers; matching subscribed_events; `TestDispatchOutbound*` |
| 12 | **Circuit breaker per endpoint** | `kernel/webhook/breaker.go` (open after N failures, half-open after cooldown, injectable clock); `TestBreaker*` |
| 13 | Retry/backoff/DLQ for outbound + inbound | next_attempt_at backoff, dead-letter at ceiling |
| 14 | Integration provider registry + per-tenant config/credential refs | `kernel/integration/` Registry + Store; `Resolve` (tenant→platform), `Upsert`; `TestIntegrationResolvePlatformAndOverride` |
| 15 | **Credentials only as secret refs (never plaintext)** | `Upsert` rejects a non-secretref credential; `TestIntegrationUpsertRejectsPlaintextCredential`; `Resolve` via secrets.Provider |
| 16 | Provider health checks (non-fatal, for readiness) | `Store.HealthChecks`; `TestIntegrationHealthChecks` |
| 17 | Anti-corruption boundary (provider payloads translated at the adapter) | `integration.Provider` port — provider types never cross into services |
| 18 | Privilege boundary (config app_platform-written; deliveries/events append-only to app_rt) | migration grants: templates/providers/endpoints app_rt SELECT-only; deliveries/events app_rt SELECT+INSERT, status by app_platform |
| 19 | Tenant isolation across all six tables | RLS (hybrid app_tenant_id_or_null for templates/providers; strict app_tenant_id for the rest); isolation tests in each package |
| 20 | Module.Context accessors wired + boot gates | NotifyTemplates/Notify/Webhooks/IntegrationProviders/Integrations on `module.Context`; boot gates NotifyTemplates().Err() + IntegrationProviders().Err() |
| 21 | Container-first verification | host `make ci` + `make test-integration`; `make ci-container` |
| 22 | Evidence bundle + parallel review | this directory; review-findings.md (security + architecture agents) |

Carried forward: legal-importance delivery audit event (app_platform lacks INSERT on events_outbox from
migration 00007 — deferred to the audit_logs writer); real smtp/sms/whatsapp/push ChannelSender adapters
and a real Verifier per provider (the framework ships the ports + HMAC verifier + fakes); moving network
I/O outside the claim tx for production throughput. Graphify `extract` blocked on LLM key (R11).
