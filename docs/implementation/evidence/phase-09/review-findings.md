# Phase 9 — Review Findings

Two parallel critique agents audited the notify / webhook / integration slice on 2026-07-04 with live
DB probes: one SECURITY-focused, one ARCHITECTURE/correctness-focused. They reproduced 13 defects —
1 critical, 1 high, 8 medium, 3 low — all fixed with regression tests. The SEC-13 privilege boundary,
tenant isolation, constant-time HMAC comparison, credential-ref enforcement, and the circuit-breaker
state machine were verified solid.

## Architecture findings

| ID | Sev | Finding (reproduced) | Resolution | Status |
|---|---|---|---|---|
| ARCH-70 | **critical** | outbound webhook deliveries in `'failed'` never retried — `DispatchOutbound` swallowed per-endpoint errors and returned nil, so the relay marked the event dispatched; the MaxAttempts/backoff table was dead code (one attempt per event, then languish) | new `RetryOutbound(ctx, plat, tenantID, now)` worker: claims failed outbound events past `next_attempt_at` (FOR UPDATE SKIP LOCKED), re-delivers, advances delivered/failed+backoff/dead; `TestIntegrationRetryOutbound_RedeliversFailed` | **fixed** |
| ARCH-71 | high | `integration.Store.Upsert` returned a freshly-generated UUID that does NOT exist when the ON CONFLICT → DO UPDATE path fired (the real row keeps its old id) → phantom FK targets | `RETURNING id` scanned into the result; `TestIntegrationUpsertReturnsPersistedID` | **fixed** |
| ARCH-72 | med | `webhook_endpoints.status` set to `'degraded'` when the breaker opened but never cleared on recovery → permanent divergence from reality | success branch UPDATEs `status='active' WHERE status='degraded'`; breaker test asserts the row returns to active | **fixed** |
| ARCH-73 | med | `notifdel_pending` index covered only `status='queued'`; `SendPending` claims `IN ('queued','failed')` → full scan for the failed half | index changed to `WHERE status IN ('queued','failed')` (mirrors the webhook `whev_pending` index) | **fixed** |
| ARCH-74 | med | (= SEC-49) NULL `external_event_id` bypassed inbound replay dedup | see SEC-49 | **fixed** |
| ARCH-75 | med | `notification_deliveries` had no backoff — `SendPending` re-claimed failed rows immediately, burning all 3 attempts in seconds during a transient outage | added `next_attempt_at` column + a monotonic `backoff(n)` (30s/2m/10m); claim gates on it; `TestSendPendingRetriesAndDeadLetters` extended | **fixed** |
| ARCH-76 | low | signature-failure audit row is rolled back in normal usage (returns an error → tx rollback) | documented as best-effort (persists only if the HTTP layer commits before the 401); durable failed-sig audit awaits the audit_logs writer | **documented** |
| ARCH-77 | low | `Send` did not check that all template-REFERENCED vars are supplied — an incomplete var map committed rows then failed at delivery (or silently for in-app) | `Send` dry-run renders each channel body against the supplied vars (channel-correct escaping) → `KindValidation` synchronously before any write; `TestSendRejectsIncompleteVariables` | **fixed** |

## Security findings

| ID | Sev | Finding (reproduced) | Resolution | Status |
|---|---|---|---|---|
| SEC-49 | med | a NULL/empty `external_event_id` defeated the UNIQUE replay constraint (Postgres NULLs are distinct) → unbounded inbound replay | webhook synthesizes `"sha256:"+hex(body)` when no id is supplied so every event is deduped; migration converts the inline UNIQUE to a PARTIAL unique index `WHERE external_event_id IS NOT NULL` (DB backstop); `TestIntegrationHandleInbound_IdlessDedup` | **fixed** |
| SEC-50 | med | on signature failure the audit row persisted the caller-supplied `external_event_id`, so a spoofed request could pre-claim a legitimate event's dedup slot (DoS) | sig-failure audit rows force `external_event_id = NULL`; `TestIntegrationHandleInbound_FailedSigDoesNotBlockValid` | **fixed** |
| SEC-51 | med | notification bodies rendered via `text/template` — an HTML-channel (email) body did not escape variable values → XSS in the recipient's mail client | `email` renders via `html/template` (contextual auto-escape); other channels stay `text/template`; `TestRenderBodyEmailEscapesHTML` | **fixed** |
| SEC-52 | low | outbound `X-Timestamp` was not covered by the HMAC signature → tamperable | sign `timestamp + "." + body` (Stripe/GitHub style); `TestUnitOutboundSignatureCoversTimestamp` | **fixed** |
| SEC-53 | low | hybrid RLS `WITH CHECK (tenant_id IS NULL OR …)` let a tenant-BOUND app_platform session write a platform (NULL-tenant) row | added a RESTRICTIVE policy `WITH CHECK (tenant_id IS NOT NULL OR app_tenant_id_or_null() IS NULL)` on notification_templates + integration_providers — a NULL-tenant write requires an unbound session | **fixed** |

Reviewer-verified solid (positive): SEC-13 boundary probed (app_rt SELECT-only on templates/providers/
endpoints; no UPDATE on deliveries/events — status is app_platform); strict tenant isolation on all six
tables (FORCE RLS); constant-time HMAC (`hmac.Equal`); empty signature rejected; plaintext credential
rejected at write; resolved secrets never logged/persisted; inbound body not logged on failure;
parameterized SQL throughout (`$1 = ANY(subscribed_events)`); breaker keyed per endpoint UUID;
dead-letter ceilings exact (no off-by-one) for both notify and webhook; template resolution precedence
+ terminating locale fallback; DispatchOutbound array-containment matching; Upsert conflict target
matches the unique index; boot gates NotifyTemplates/IntegrationProviders Err(); Notify/Webhooks/
Integrations never nil; migration down order FK-safe.

Enabling change made alongside the fixes: migration 00007 granted app_platform only SELECT/UPDATE on
events_outbox; `GRANT INSERT ON events_outbox TO app_platform` added in 00011 so tenant-bound workers
(inbound webhook handlers, the delivery sender) can EMIT events (e.g. a legal-importance delivery
audit event, a module's inbound-webhook domain event). The relay's WITH CHECK admits it; the outbox
Writer stamps tenant_id from app_tenant_id().

Residual / carried forward (honest):
- Real smtp/sms/whatsapp/push ChannelSender adapters and real per-provider Verifiers are the product's
  to supply — the framework ships the ports + an HMAC verifier + fakes.
- `RetryOutbound`/`ProcessInbound`/`SendPending` do the network I/O inside the claim tx for
  simplicity; a high-throughput deployment should move the I/O outside the tx (documented).
- The inbound `HMACVerifier` verifies body-only (external-provider scheme); our outbound signing now
  covers the timestamp — a wowapi→wowapi loopback verifier would use the timestamped scheme.
- Legal-importance delivery audit event emission is now UNBLOCKED by the events_outbox grant but not
  yet wired into SendPending (deferred with the audit_logs writer).
- The SEC-53 hybrid-RLS pattern exists on earlier-phase hybrid tables (roles/policies/rule_versions/
  workflow_definitions) too; retrofitting the restrictive backstop there is a noted follow-up (the Go
  stores already gate the tenant argument, so it is defense-in-depth, not an open exploit).
