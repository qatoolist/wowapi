---
id: DEV-W04-E04-S003
type: deviations-record
parent_story: W04-E04-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W04-E04-S003

## DEV-W04-E04-S003-001 — model_hash readiness detail is a placeholder pending AR-01

**Planned:** DX-07 T2 acceptance criterion AC-W04-E04-S003-02 requires readiness to report migration
version, seed/rule hash, and model hash.

**Actual:** `migration_version`, `seed_catalog_hash`, and `rule_hash` are fully implemented and
reported in the `/readyz` payload. `model_hash` is reported only when `kernel.Kernel.ModelHash` is
non-empty. As of this story's implementation, AR-01's deterministic application-model hash (W05-E01-
S003) has not landed, so `ModelHash` is always empty and the `model_hash` detail is omitted from the
payload.

**Impact:** The readiness payload does not yet satisfy the full model-hash portion of T2. The
omission is visible (the key is absent rather than silently blank) and the field will appear
automatically once AR-01 sets `Kernel.ModelHash`.

**Resolution plan:** No framework-side code change is required beyond the placeholder detail provider
already in `app/health.go`. When AR-01 lands, verify that a non-empty `model_hash` appears in the
readiness payload and update this deviation to `resolved`.

**Evidence:** `app/health_readiness_test.go` proves `migration_version`, `seed_catalog_hash`, and
`rule_hash` are present; it does not assert `model_hash` because the test kernel has no model hash.
