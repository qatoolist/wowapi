# Hardening H1 — acceptance map

Roadmap acceptance criteria → code/test/command evidence.

| Item | Acceptance criterion | Evidence |
|---|---|---|
| S7 | assumed headers/TLS posture shipped + tested; deployment checklist | `kernel/httpx/edge.go` `SecureHeaders`; `edge_test.go::TestSecureHeadersDefaults`; `deployments/reference/{nginx.conf,smoke.sh}`; `docs/operations/deployment-checklist.md` §1 |
| S7 | in-process body/timeout caps (blueprint chain) enforced | `BodyLimit`/`Timeout` wired from `HTTP.MaxBodyBytes`/`RequestTimeout`; `TestBodyLimitCapsRequestBody`, `TestTimeoutFiresOnSlowHandler`; generated `cmd/api/main.go.tmpl` |
| S7 | CORS allowlist per env config | `config.HTTP.CORSAllowedOrigins`; `httpx.CORS`; `TestCORSAllowedOriginEchoed`, `TestCORSPreflightShortCircuits`, `TestCORSCredentialsNeverWildcard` |
| S8 | fuzz the filter DSL parser and cursor decoding | `kernel/filtering/fuzz_test.go` (`FuzzFilterParse`,`FuzzParseSort`), `kernel/pagination/fuzz_test.go` (`FuzzDecodeCursor`); `make test-fuzz`; 1.7M/478K execs clean |
| R7 | cursors carry a sort-spec version; mismatch fails loudly | `pagination.EncodeCursorWithSig`/`Cursor.Sig`; `filtering.Sort.Signature`/`NextCursor`; `KeysetClause` guard; `TestKeysetClauseRejectsSortSpecChange`, `TestSortSignatureStableAndOrderSensitive` |
| S5 | TTL + archival job; no unbounded growth | `expires_at` (existing) + `IdemStore.SweepExpired`; migration `00012_idempotency_sweep.sql`; `TestIntegrationIdempotencySweepExpired` (cross-tenant) |
| R6 | hold re-checked at deletion time inside the deleting tx; race test proves survival | `FOR UPDATE` + `legal_hold=false` guard in `SweepRetention`; `TestIntegrationLegalHoldRaceSurvivesSweep` (proven by revert) |
| O4 | reference monitoring rule (alert on fingerprint change w/o deploy) | `docs/operations/deployment-checklist.md` §2 + `WowapiConfigDrift` Prometheus rule |

Gate: `make ci` + `make ci-container` exit 0 — 0 FAIL, 0 SKIP, 74 packages, DB tests forced.
