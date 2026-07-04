# Hardening H1 — Edge & pagination hardening — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). Decision: [D-0061](../../decisions.md).
Scope: the self-contained, no-heavy-schema P0/P1 gaps — S7, S8, R7, S5, R6, O4.

## Findings addressed

| Item | Pri | Verdict | Fix |
|---|---|---|---|
| S7 | P0 | real — blueprint's `SecureHeaders→CORS→BodyLimit→Timeout` chain never implemented; `HTTP.MaxBodyBytes`/`RequestTimeout` were dead config | `kernel/httpx.SecureHeaders/CORS/BodyLimit/Timeout`; wired into generated api; `http.cors_allowed_origins` config; reference nginx + smoke.sh + deployment checklist |
| S8 | P0 | real — zero fuzz/property tests on the two untrusted parsers | `FuzzFilterParse`, `FuzzParseSort`, `FuzzDecodeCursor` (seed corpus in CI, `make test-fuzz` for deep runs); 1.7M+ execs clean |
| R7 | P1 | real — cursor bound column *values* but not sort direction/order → silent wrong pages | optional sort-spec signature in the cursor; `Sort.Signature()`; `filtering.NextCursor`; loud `KeysetClause` mismatch |
| S5 | P1 | real — `expires_at` + reclaim existed, nothing ever DELETED → unbounded growth | `IdemStore.SweepExpired` (cross-tenant, app_platform); migration 00012 platform sweep policy+grant |
| R6 | P0 | real — retention sweep checked `legal_hold` once, unlocked → hold applied mid-sweep was voided | `FOR UPDATE` on candidate select (EvalPlanQual re-check) + `legal_hold=false` guard on the void UPDATE |
| O4 | P1 | real — `/readyz` fingerprint exposed, nothing consumes it | config-drift alerting convention + reference Prometheus rule in the deployment checklist |

Verified NOT gaps (roadmap claims inaccurate; recorded so they don't reopen): **S4** — integration
credentials are already `credential_ref` (secret-provider key, never plaintext) with compiler-enforced
`config.Secret` redaction (`migrations/00011:67`, `kernel/config/secret.go`). **R2/R8** — advisory lock
and per-endpoint breaker are correct as designed.

## Implementation inventory

New: `kernel/httpx/edge.go` (+`edge_test.go`); `kernel/pagination/cursor_sig_test.go`,
`fuzz_test.go`; `kernel/filtering/fuzz_test.go`; `migrations/00012_idempotency_sweep.sql`;
`deployments/reference/{nginx.conf,smoke.sh}`; `docs/operations/deployment-checklist.md`;
`docs/implementation/hardening-plan.md`.
Changed: `kernel/config/config.go` (`HTTP.CORSAllowedOrigins`); `kernel/pagination/cursor.go`
(signed envelope, `EncodeCursorWithSig`, `Sig()`); `kernel/filtering/{sort.go,keyset.go}`
(`Signature`, `NextCursor`, mismatch guard); `kernel/database/idempotency.go` (`SweepExpired`);
`kernel/document/service.go` (R6 fix); generated `cmd/api/main.go.tmpl` chain; `Makefile`
(`test-fuzz`); `migrations/migrations_test.go`; two `!=`→`reflect.DeepEqual` fixes for the now
slice-bearing `config.HTTP`.

## Acceptance

All P0/P1 items in scope closed with tests. `make ci` + `make ci-container` green: **0 FAIL, 0 SKIP,
74 packages**, DB tests forced (`WOWAPI_REQUIRE_DB=1`). R6 fix proven by revert (buggy → "1 version(s)
were voided"; fixed → pass). Fuzzers ran 1.7M (filter) / 478K (cursor) execs with no crash.
