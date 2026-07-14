---
id: DEV-W04-E01-S003
type: deviations-record
parent_story: W04-E01-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W04-E01-S003

## D1 — Breaking worker-registration API change recorded as open coordination item
**What changed.** `kernel/jobs.Registry.RegisterKind` now requires an
`Idempotency` value as a third argument:

```go
RegisterKind(kind string, w Worker, idem Idempotency, rp RetryPolicy)
```

This is a source-breaking change for any consumer that previously called
`RegisterKind` with only a worker and retry policy. The worker function
signature itself (`Worker func(ctx context.Context, db database.TenantDB, payload []byte) error`)
remains unchanged; the duplicate-safety contract is enforced at registration
time, and the stable idempotency key plus lease context are delivered through
`ctx` rather than by changing the function type.

**Why it diverges from plan.** PLAN-W04-E01-S003 explicitly left the T5
worker-signature-change coordination question unresolved (RISK-W04-003). The
implementation had to choose a concrete mechanism to enforce "exactly one
duplicate-safety mechanism"; adding the `Idempotency` parameter to
`RegisterKind` is that mechanism. It is therefore a real, breaking API change
relative to pre-W04 job registration.

**Coordination status.** This change is intentionally recorded here as an
open, tracked-forward item per RISK-W04-003. All in-framework call sites
(`kernel/jobs/*_test.go`, `kernel/jobs/chaos/*`, `kernel/webhook/service.go`)
have been updated. Downstream consumers (notably `wowsociety`) must update
their registrations when they adopt this framework version. No migration or
shim is provided because the previous signature had no idempotency information
to carry forward.
