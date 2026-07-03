# Phase 9 — Proof Bundle

Scope (phase-plan row 9): notification framework (templates, Send-in-tx, async delivery), webhook
framework (inbound verify/replay + outbound HMAC delivery + circuit breaker), integration provider
registry (config + secret-ref credentials + health), migration 00011. Date: 2026-07-04.

## 1. Decision evidence
D-0056 (Phase 9: notify/webhook/integration design + review fixes — config tables app_platform-written
per SEC-13, deliveries/events append-only, deny-first credential secret-refs, outbound retry worker,
partial-index replay dedup, restrictive hybrid RLS, email HTML-escape).

## 2. Discussion evidence
- **SEC-13 posture for config**: notification_templates, integration_providers, and webhook_endpoints
  change behavior (which channels/endpoints fire, which credentials sign), so they are app_platform-
  written and app_rt-SELECT-only — the same posture as rules/document governance columns.
- **Append-only delivery/event tables**: notification_deliveries and webhook_events are INSERT-only to
  the module role; status/attempts are advanced by the async sender/relay running as app_platform.
- **Outbound retry (the critical review fix)**: DispatchOutbound alone gave each event a single attempt;
  a dedicated `RetryOutbound` worker (mirroring notify's `SendPending` and outbox's relay) is what
  actually drives the backoff/DLQ. This is the honest shape — retry is a worker loop, not a side effect
  of dispatch.
- **Replay dedup**: a provider that omits an event id would defeat a plain UNIQUE (NULLs are distinct),
  so the service synthesizes a body-hash id and the DB carries a PARTIAL unique index — both layers.
- **Secrets**: credentials and webhook signing secrets are ONLY secret references (secretref://…),
  resolved through the secrets.Provider; a plaintext credential is rejected at write.

## 3. Critique/review evidence
`review-findings.md`: 13 reproduced defects — 1 critical (ARCH-70 outbound never retried), 1 high
(ARCH-71 phantom Upsert id), 8 med (replay bypass, audit DoS, email XSS, missing backoff/index, degraded
never cleared, incomplete vars), 3 low (unsigned timestamp, hybrid RLS write, audit rollback). All fixed
with regression tests or documented+enforced. Two parallel review agents; privilege boundary, tenant
isolation, constant-time HMAC, and the breaker state machine verified solid.

## 4. Implementation evidence
New: `kernel/notify/` (registry, sender, service), `kernel/webhook/` (webhook, service, verifier, sender,
breaker), `kernel/integration/` (integration, store), migration `00011_notify_webhook_integration.sql`.
Changed: `kernel/kernel.go` (wire Notify/Webhooks/Integrations + secretRefResolver adapter + Secrets/
WebhookSender deps), `module/module.go` + `app/context.go` + `app/boot.go` (accessors + boot gates),
`migrations/migrations_test.go`.
Team: 2 implementation agents (notify; webhook) + lead (integration, migration, wiring, all migration-
level review fixes + integration ARCH-71); the notify + webhook review fixes were driven back through
their original authoring agents. 2 review agents (security, architecture).

## 5. Verification evidence
`command-log.md`: notify integration (template validation, Send-in-tx atomicity, resolution/locale,
render escaping, backoff + dead-letter, incomplete-vars), webhook integration (inbound verify/replay/
timestamp-window, ProcessInbound + RetryOutbound retry/DLQ, HMAC-signed outbound, breaker open/half-open
+ degraded clear, id-less dedup, failed-sig isolation), integration (resolve precedence, secret-ref
credential, plaintext rejection, health, ARCH-71 id). Full `make ci` host + `go test -p 1` and
`make ci-container` in-container green.

## 6. Acceptance evidence
`acceptance-map.md`: all 22 Phase 9 exit criteria mapped to named tests. Carried forward: real channel-
sender + provider-verifier adapters, network I/O outside the claim tx for throughput, legal-delivery
audit event wiring (grant now in place), and the SEC-53 restrictive-policy retrofit on earlier hybrid
tables. Graphify `extract` blocked on LLM key (R11).
