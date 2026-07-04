# wowapi — Hardening Roadmap

> Companion to [ROADMAP.md](ROADMAP.md) (§6 framework discovery, §14 generic improvements). Scope: what is required to take wowapi v0.1.0 — all 12 implementation phases completed ~2026-07-04, zero production exposure — to a state a compliance-grade product can safely stand on. Everything here is **domain-neutral**; nothing society-specific belongs in the framework.
>
> Two categories throughout: **(A) trust hardening** — making existing capabilities production-trustworthy under real load and adversarial pressure; **(B) capability gaps** — things a compliance-grade product cannot ship without (cross-referenced to ROADMAP.md §14 numbering).

Priority key: **P0** = before any product MVP ships on the framework · **P1** = before post-MVP phases depend on it · **P2** = quality-of-life / scale-driven.

---

## 1. Security hardening

| # | Item | Cat | Pri | Current state | Acceptance criteria |
|---|---|---|---|---|---|
| S1 | **Machine authentication (API keys / service principals)** | B (§14.8) | P0 | Only OIDC user JWTs; no non-human credential exists | Scoped, rotatable, revocable machine principals; per-key permission sets; issuance/rotation audited; gate-device and integration callers authenticate without user tokens |
| S2 | **Rate limiting** | B (§14.8) | P1 | Middleware hooks only; delegated to reverse proxy | Per-principal and per-permission limits enforceable in-process (e.g., PII-export endpoints); 429 with RFC 7807 body; metrics per limited route |
| S3 | **Step-up auth / MFA hooks** | B (§14.9) | P1 | Token is the only factor | Authz layer can demand elevated auth per permission; generic challenge interface (TOTP first); dual-control composes with workflow approval steps |
| S4 | **Encrypt integration credentials at app level** | A | P0 | Provider secrets stored as-is; relies on DB-level encryption | Envelope encryption for stored credentials; key rotation procedure; secrets never appear in logs/dumps (extend existing redaction tests) |
| S5 | **Idempotency-key expiry/archival** | A (§14.7) | P1 | Keys stored forever; unbounded growth | Configurable TTL + archival job; replay after expiry returns a defined error, not silent re-execution |
| S6 | **Audit tamper-evidence** | B (§14.2) | P0 | Audit rows append-only; no cryptographic proof | Hash-chaining per tenant per period; exportable anchors; verification tool that detects any mutation/deletion in the chain |
| S7 | **Reference deployment for proxy-delegated concerns** | A | P0 | CORS, security headers, TLS, compression assumed to be the reverse proxy's job — undocumented, untested | A tested reference deployment (nginx/gateway config) shipping the assumed headers/TLS posture; deployment checklist updated; CI smoke test against the reference stack |
| S8 | **Adversarial testing program** | A | P0–P1 | Good built-in security tests (RLS isolation, deny-by-default, privilege escalation, redaction); no external pressure | Fuzzing of the filter DSL parser (user input → SQL WHERE; highest-value target) and pagination cursor decoding; property tests on authz scope-covering; third-party pen test once a product sits on top |

## 2. Reliability and scale hardening

| # | Item | Cat | Pri | Current state | Acceptance criteria |
|---|---|---|---|---|---|
| R1 | **Authz decision caching + read-replica routing** | B (§14.11) | P1 | Every `Evaluate` hits the DB; single database | Cache with explicit invalidation on authorization-spine writes; correctness proven by tests (no stale-allow after revocation); `WithTenantRO` routable to replicas |
| R2 | **Advisory-lock contention characterization** | A | P0 | Per-aggregate event ordering serializes to one event at a time per aggregate | Load test simulating bulk emission against hot aggregates (bill-run shape); documented throughput envelope; sub-sharding strategy for aggregate keys if the envelope is exceeded |
| R3 | **SLA sweeper: configurable + multi-replica-safe** | A | P0 | Interval hardcoded (~5 min); single-runner assumption | Configurable interval; leader election or advisory-lock guard so N worker replicas don't double-fire; sweeper lag exposed as a metric |
| R4 | **DLQ operability (events + jobs)** | A | P0 | Dead-lettering works; no inspection/replay tooling | Admin API/CLI to list, inspect, replay, and discard DLQ entries with audit; replay is idempotent-safe; DLQ depth alerting metric |
| R5 | **Notification delivery evidence** | B (§14.10) | P0 | Fire-and-forget; no delivery-status query, no provider receipts | Delivery status queryable per notification; provider receipts stored; per-user channel preferences; failures surface to caller, not just logs |
| R6 | **Retention-sweep legal-hold race** | A | P0 | Hold status checked once per sweep; hold applied mid-sweep can be missed | Hold re-checked at deletion time inside the deleting transaction; test that races a hold against a sweep and proves the document survives |
| R7 | **Keyset cursor versioning** | A | P1 | Sort-column changes mid-pagination silently yield wrong pages (documented as caller's problem) | Cursors carry a sort-spec version; mismatch fails loudly with a defined error kind |
| R8 | **Webhook outbound resilience granularity** | A | P2 | Circuit breaker is per-endpoint (all deliveries fail together) | Documented as intended behavior + per-delivery retry budget; endpoint health surfaced via metrics |

## 3. Operational hardening

| # | Item | Cat | Pri | Current state | Acceptance criteria |
|---|---|---|---|---|---|
| O1 | **Distributed tracing (OTel)** | B (§14, P2→raise) | P1 | Request-ID propagation only | Traces span API → outbox relay → worker → notification behind the existing metrics/observability port; sampling configurable; zero-cost when disabled |
| O2 | **Migration safety harness** | B (§14.13) | P0 | Goose + raw SQL; no expand/contract helpers; no migration tests | CI runs every migration forward (and down where defined) against seeded template DBs; snapshot diffing; documented zero-downtime expand/contract pattern for the journal-bearing tables |
| O3 | **v0.x upgrade discipline** | A | P0 | Breaking changes expected until v1.0 (error kinds, config schema, testkit declared fair game) | Product pins exact versions; module contract-test suite is the upgrade tripwire in CI; framework publishes a CHANGELOG-driven deprecation policy before product Phase 2 |
| O4 | **Config-drift alerting convention** | A | P1 | Fingerprint drift detection exists at `/readyz`, nothing consumes it | Reference monitoring rule (alert on fingerprint change without deploy); documented in the deployment checklist |
| O5 | **Backup/restore drill support** | A | P0 | Nothing framework-specific | Documented PITR + object-storage restore procedure validated against a testkit-seeded instance; drill scriptable so products can run it quarterly |

## 4. Evidence-layer capability gaps (P0 for compliance-grade products)

Not defects — but any compliance product on today's wowapi would hand-roll these, badly (wowsociety.app's failures are the cautionary tale):

1. **Field-level audit trail + query API** (ROADMAP §14.1) — framework-managed change capture (entity, field, before/after, actor, capacity, impersonator, request-ID, tx-ID); standardized append-only schema; queryable; per-module redaction hooks. Today: row-level who/when, no query surface, schema varies by module.
2. **Data lifecycle & retention engine** (§14.4) — per-record-class retention policies, scheduled disposition, legal hold generalized beyond documents, DSR export/erasure primitives with statutory-override reasons. DPDP Rules are live in 2026; not deferrable.
3. **Gap-free per-tenant sequence allocator** (§14.5) — transactional numbered series (receipts/vouchers/certificates) with audited voids. The framework currently offers nothing here, which is exactly how wowsociety.app ended up with `MAX()+1` races on statutory document numbers.
4. **Snapshot/artifact pipeline** (§14.3) — dataset → immutable versioned artifact (PDF/A + structured sidecar + hash) atop the document framework, templates versioned by effective date.
5. **Scheduler** (§14.12) — cron-style recurring job registration; today only on-demand enqueue plus the hardcoded sweeper.
6. **Bulk-operation framework** (§14.15) — chunked jobs with progress reporting, partial-failure ledger, resumability.

## 5. Hardening method — how to actually do it

The design is right; only sustained abuse proves the implementation. Sequence:

1. **Phase-0 pilot module** (the single highest-leverage step, per ROADMAP §13.3): one throwaway module exercising every `Context` capability end-to-end — route → authz → tx → outbox → job → document → notification — against testkit and the compose stack.
2. **Load & soak:** relay and job-runner soak tests; concurrent bill-run-shaped emission against advisory locks (feeds R2); authz Evaluate under member-portal read load (feeds R1).
3. **Chaos:** kill the worker mid-job and mid-relay batch; verify at-least-once semantics and DLQ behavior (feeds R4); race legal holds against retention sweeps (feeds R6).
4. **Adversarial:** fuzz the filter DSL and cursor decoding; attempt cross-tenant and privilege-escalation attacks beyond the built-in suite (feeds S8).
5. **Operational drills:** restore-from-backup on a seeded instance (O5); migration forward/back on realistic data (O2); upgrade the pinned framework version with contract tests as tripwire (O3).
6. **Upstream loop:** every issue shaken out lands in wowapi as a domain-neutral fix or in this backlog — never as a product-side workaround (framework-purity gate, ROADMAP §13.3).

**Exit gate (framework "hardened enough" for product MVP):** all P0 rows above closed; pilot module green under load/chaos/adversarial suites; a reference deployment exists and is smoke-tested in CI; backup/restore and migration drills documented and rehearsed once.
