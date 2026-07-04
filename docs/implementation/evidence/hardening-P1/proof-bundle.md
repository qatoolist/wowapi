# Hardening P1 — proof bundle

Self-contained P1 items from [../../hardening-plan.md](../../hardening-plan.md), each individually
QA-gated. (The large P1 items — R1 authz cache, O1 OTel — remain; this bundle accumulates the small,
self-contained ones.)

## S2 — Rate limiting (D-0064)

| Verdict | Fix |
|---|---|
| real (P1) — only middleware hooks, delegated to a proxy; no in-process limiter | `kernel/httpx.RateLimit` middleware + `TokenBucket` limiter (`NewTokenBucket`); 429 + `Retry-After` + RFC 7807 (`KindRateLimited`, already in the taxonomy); `KeyByIP` / `KeyByActor` key strategies; idle-bucket sweep bounds memory |

Enforceable in-process per the roadmap: per-principal (`KeyByActor`, after the authz gate),
per-permission (custom keyFn on expensive routes), per-IP (`KeyByIP`, edge). Opt-in — limits are
product-specific; wiring documented in `docs/operations/deployment-checklist.md` §5.

Tests (`kernel/httpx/ratelimit_test.go`): burst-then-limit, refill over time, independent keys, 429 body
+ Retry-After header, allow path, actor-key with IP fallback.

Gate: `make ci` + `make ci-container` green — 0 FAIL, 0 SKIP, 74 packages.
</content>
