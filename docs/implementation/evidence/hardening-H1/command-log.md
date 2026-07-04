# Hardening H1 — command log

| # | Command | Result |
|---|---|---|
| 1 | `go test ./kernel/httpx/ -run 'SecureHeaders\|CORS\|BodyLimit\|Timeout'` | ok — edge middleware (TDD: red → green) |
| 2 | `go test ./kernel/config/ ./app/` | ok — after `HTTP.CORSAllowedOrigins` slice broke two `!=` comparisons, fixed to `reflect.DeepEqual` |
| 3 | `go test ./kernel/pagination/ ./kernel/filtering/` | ok — R7 sig round-trip + loud mismatch |
| 4 | `go test ./kernel/filtering/ -fuzz=FuzzFilterParse -fuzztime=6s` | PASS — 1,707,799 execs, 0 crashes |
| 5 | `go test ./kernel/pagination/ -fuzz=FuzzDecodeCursor -fuzztime=6s` | PASS — 478,431 execs, 0 crashes |
| 6 | `make migrate` (local compose DB) | applied; source at version 12 (00012 idempotency sweep) |
| 7 | `go test ./kernel/document/ -run 'LegalHold\|Retention'` | ok — incl. `TestIntegrationLegalHoldRaceSurvivesSweep` |
| 8 | `go test ./testkit/ -run 'Idempotency'` | ok — incl. `TestIntegrationIdempotencySweepExpired` |
| 9 | R6 revert-test (drop `FOR UPDATE`+guard) → run race test | **FAIL** "1 version(s) were voided" → restore → **PASS** (proves the guard) |
| 10 | `gofmt -l` / `go build ./...` / `go vet ./...` | clean / OK / OK |
| 11 | `make ci` (host) | exit 0 — vet, lint, boundaries, unit, race, perf budgets, build |
| 12 | `make ci-container` (authoritative, `WOWAPI_REQUIRE_DB=1`) | exit 0 — **0 FAIL, 0 SKIP, 74 ok**; grep `FAIL\|permitted to log in\|REQUIRE_DB is set`=0, `SKIP`=0 |

Residual risk: the S7 reference nginx stack is smoke-tested via `deployments/reference/smoke.sh`
(a deploy-time/quarterly drill), not in core CI — adding nginx to the CI image was judged out of
proportion for this phase. The in-process header posture IS unit-tested (`kernel/httpx/edge_test.go`),
so CI still proves the application-layer contract.
</content>
