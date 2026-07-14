---
id: VER-W01-E03-S001
type: verification-record
parent_story: W01-E03-S001
status: executed
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W01-E03-S001

## Post-execution record (2026-07-13, SHA 0a31186cada5c275a588c74081cf977adf346e61, local darwin/arm64, go1.26.5)

| Acceptance criterion | Verification method executed | Result | Evidence |
|---|---|---|---|
| AC-W01-E03-S001-01 | `TestHTTPTimeoutDefaultsMatchCS09` asserts Defaults().HTTP = header 10s / read 30s / write 60s / idle 120s | PASS (note: header 10s lives on existing `ReadHeaderTimeout`, not a new `HeaderTimeout` — DEV-001) | EV-W01-E03-S001-002, -003 |
| AC-W01-E03-S001-02 | Template-render tests, fail-first pair: FAILED at pristine 0a31186cada5c275a588c74081cf977adf346e61, PASS after template+config fix | PASS (fail-first captured) | EV-W01-E03-S001-001 |
| AC-W01-E03-S001-03 | 3 new prod-gated matrix rows + non-prod-zero-tolerated test + prod-baseline-with-defaults validates; fail-first pair captured | PASS (fail-first captured) | EV-W01-E03-S001-003 |
| AC-W01-E03-S001-04 | `TestCSRFOversizedFormBodyRejected` fail-first pair + `TestCSRFCustomMaxFormBytesOverridesDefault` + scoped `gosec ./kernel/httpx/` (0 findings) | PASS (gosec rule-id caveat recorded in EV-004; definitive linter re-run lands with W01-E01-S002) | EV-W01-E03-S001-004, -005 |

- **Race detector**: `go test -race -count=1 ./kernel/httpx/ ./kernel/config/ ./app/ ./internal/cli/` — all ok.
- **Regression**: full `kernel/config`, full CSRF suite, and the rendered-product compile tests (`TestInitRenderedProductCompiles*`) all pass — no existing behavior changed for in-bound requests or unset configs.
- **Reviewer**: W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 (review gate passed).
