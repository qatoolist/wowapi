# wowapi — Gaps, Recommendations & Fix Plan

## A. Gaps found and their disposition

| ID | Gap | Severity | Disposition |
|---|---|---|---|
| G1 | RLS pool guard (superuser/BYPASSRLS rejection) untested | HIGH (security) | **CLOSED** — `rls_guard_test.go` |
| G2 | Workflow runtime lifecycle (CompleteTask/Delegate/Override/gateway) untested | HIGH | **CLOSED** — `runtime_lifecycle_test.go` |
| G3 | SLA ISO-duration parse untested | MED | **CLOSED** — `sla_parse_test.go` |
| G4 | `relationship.Relate` write path untested | MED | **CLOSED** — `relationship_relate_test.go` |
| G5 | Resource-type registry contract untested | MED | **CLOSED** — `registry_test.go` |
| G6 | DB idempotency store | — | **NOT A GAP** — covered by `testkit/idempotency_test.go` (store/in-flight/concurrent); not duplicated |
| G7 | `kernel.New` wiring untested directly | LOW | **ACCEPTED** — exercised indirectly by app boot tests; a nil field fails those. Direct test = low marginal value |
| G8 | benchbudget CI-gate parser untested | LOW (tooling) | **CLOSED** — `main_test.go` |

## B. Design finding

### D1 — `relationship.Relate` is an exported, caller-less, previously-untested platform-only write

**Observed:** `relationship.Relate` (exported) inserts into `relationships` using `app_tenant_id()`, but
`app_rt` has only `SELECT` on that table (`app_platform` has `INSERT`; edge creation is a kernel/platform
capability, SEC-24). The function had **no callers anywhere** and **no test**. Called on an app_rt
`TenantDB` it fails with a permission error; it only works on a tenant-bound `app_platform` transaction.

**Severity:** LOW. Not a bug — the privilege split is intentional and correct. The risk was *latent*: an
exported write with no worked example invites incorrect (app_rt) use that fails only at runtime.

**Resolution (no code change):** the behavior is correct, so per the "no unrelated rewrites" rule no
production code was changed. Instead the contract is now pinned by `relationship_relate_test.go`:
- correct usage (tenant-bound `app_platform`) writes an edge that `Has` then reads;
- `app_rt` is denied (the SEC-24 boundary the schema comment asserted but nothing tested);
- tenant isolation holds.
The exported function is a legitimate part of the ReBAC edge-management surface a future
edge-management service (running as app_platform) will call; it now has a regression-protected,
documented contract.

## C. Recommendations (future test debt — not blocking)

1. **`parseBenchOutput(r *os.File)`** would be marginally more testable as `io.Reader`; not changed here
   to avoid an unrelated refactor (tested via a temp file instead). Consider it if the tool grows.
2. **OpenAPI strict CI-diff** (criterion #12's generated-vs-registered diff) remains an incremental
   harness beyond `openapi merge` — carried from Goal 2, not a regression gap.
3. **Durable audit_logs assertions** — audit currently flows through the logging sink; when the durable
   `audit_logs` writer lands, add data-integrity tests for the partitioned append-only table.
4. **Per-package coverage attribution** — `database`'s pool/tx plumbing is exercised by integration
   tests in *other* packages; if a per-package floor is ever enforced, measure with
   `go test -coverpkg=./...` to attribute cross-package coverage correctly.

## D. Fix plan summary

No framework defect required a code fix. The "fixes" delivered by this effort are the 21 regression
tests that convert previously-untested framework contracts (security guard, workflow lifecycle, ReBAC
write, registration, SLA parse, CI gate) into pinned, repeatable guarantees. Every closed gap has a
named test that fails if the behavior regresses (traceability in 03/05).
