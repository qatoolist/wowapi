# wowapi — Existing Test Review (non-duplication baseline)

Purpose: record what the existing 477-test suite already covers **well**, so this QA effort adds only
missing/weak coverage. Grounded in the actual `_test.go` files.

## Strong existing coverage — DO NOT duplicate

| Area | Existing tests (representative) | Assessment |
|---|---|---|
| Tenant isolation / RLS | `testkit.AssertRLSIsolation` catalog sweep; per-package `TestIntegration*TenantIsolation` (document, comment, attachment, notify, outbox, jobs, authz, rules) | Strong — the isolation invariant is well-covered per tenant table |
| Deny-by-default authz | `kernel/authz` evaluator suite (RBAC/ReBAC/ABAC deny-first, scope coverage, unresolved-attr fail-closed), `escalation_test` (no self-grant) | Strong |
| Privilege escalation | `kernel/authz/escalation_test` (self-grant blocked at DB privilege level), document grant RLS (SEC-41/42), legal-hold column lockdown (SEC-44) | Strong |
| Outbox / jobs reliability | outbox atomicity, relay dispatch+inbox dedup, per-aggregate ordering under retry, DLQ; jobs enqueue-atomic, retry→DLQ, backoff, reclaim, tenant isolation, worker end-to-end | Strong |
| Rules engine | resolution precedence, historical + superseded-window (`ARCH-60`), write-time schema validation, approval gating | Strong (store/resolver happy + key paths) |
| Workflow definition | parse-strict, validate (orphan/dangling/unreachable/unknown-action/resolver), fail-closed gating, approval-completeness, linear approval, reject, optimistic lock, SLA sweep, WorkflowSim | Strong on DEFINITION; **weak on runtime task lifecycle** (see gaps) |
| Documents / storage | upload round-trip, byte verification, MIME sniff/essence, scan gate, infected-never-serves, access grant, retention sweep, legal hold, RO-tx download, distinct keys, tenant isolation; storage memory adapter | Strong |
| Notify / webhook / integration | template validation, send-in-tx atomicity, locale fallback, backoff/dead-letter, html-escape (SEC-51); inbound verify/replay/window, ProcessInbound/RetryOutbound, HMAC signing, circuit breaker, id-less dedup, failed-sig isolation; provider resolve/credential/health | Strong |
| Config / secrets | layered load, redaction (logs/CLI/dumps), fingerprint, per-knob unsafe-config prod matrix, shared-section drift, env-provider | Strong |
| HTTP primitives | problem-details, decode, etag, idempotency middleware, listing, pagination keyset, filtering sort/keyset, health endpoints | Good |
| Module SDK / boot | `RunModuleContract`, boot gates (perms/resources/router/events/jobs/rules/workflows/docs/notify), external scratch-consumer, widgets contract | Strong |
| Observability / perf | metrics port + RED middleware + AccessLog + NoOp, prometheus adapter, 24 hot-path benchmarks + budget gate | Strong |
| Delivery | migrations fresh+idempotent+ordering, E2E scaffold→build→migrate→/healthz, CLI commands | Strong |

## Available helpers/fixtures (reuse — do not reinvent)

- `testkit.NewDB(t)` → `DBHandle{Admin, Runtime, Platform pools; TxM, PlatformTxM}`; per-test cloned DB.
- Fixtures: `CreateTenant`, `CreateUser`, `CreateCapacity`, `CreateOrg`, `CreatePermission`,
  `CreateRole`, `GrantRole`, `CreateResourceType[AndResource]`, `CreateResource`, `TenantCtx`.
- Asserts: `AssertRLSIsolation`; `testkit/authz_asserts`. Sim: `WorkflowSim`. Fakes: `testkit/fakes`.
- `database.WithTenantID`/`WithActorID`; `WithTenant`/`WithTenantRO`; `PlatformTxM.WithTenant` (bound).

## Weak / missing coverage → addressed by this effort

Cross-checked against the above; the following are NOT covered and are the targets (see 01 §4 G1–G8):
RLS connection guard (security), workflow runtime task lifecycle + gateway + SLA parse, relationship
write path, resource registry contract, DB-layer idempotency claim, kernel.New wiring, benchbudget parser.

No existing test is shallow-to-the-point-of-misleading; no flaky tests remain after the Phase-11
testkit role-provisioning fix. The suite is well-structured; the gaps are specific behaviors, not whole
areas.
