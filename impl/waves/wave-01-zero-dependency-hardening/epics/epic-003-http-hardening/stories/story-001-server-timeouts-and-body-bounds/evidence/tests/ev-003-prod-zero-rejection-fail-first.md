# EV-W01-E03-S001-003 — prod-profile zero-timeout rejection (fail-first pair)

- **Evidence ID**: EV-W01-E03-S001-003
- **Evidence type**: unit-test report (fail-first pair)
- **Story / task**: W01-E03-S001 / W01-E03-S001-T002
- **Acceptance criteria proven**: AC-W01-E03-S001-03 (and AC-01 via TestHTTPTimeoutDefaultsMatchCS09 in the same run)
- **Execution command**: `go test ./kernel/config/ -run 'TestProdUnsafeConfigKnobMatrix|TestHTTPTimeoutDefaultsMatchCS09|TestConnectionTimeoutZeroToleratedOutsideProd|TestEnforceRouteContractsDefaultsOff' -count=1 -v`
- **Code revision / commit SHA**: 0a31186cada5c275a588c74081cf977adf346e61 (working tree on top of this HEAD; conductor owns the wave commit — the diff is the uncommitted working change, `git diff --stat` recorded in the story's implementation.md)
- **Branch**: main
- **Execution environment**: darwin/arm64 workstation (local)
- **Tool versions**: go1.26.5; gosec (dev build)
- **Date/time**: 2026-07-13 13:06 IST
- **Reviewer**: pending — W01 wave review gate (conductor)
- **Result**: pre-implementation the three new prod-gated rows FAILED ("knob is NOT gated in prod" — the defect is real); post-implementation all rows PASS, including `TestConnectionTimeoutZeroToleratedOutsideProd` proving the resolved policy is prod-gated (NOT unconditional) and the prod baseline with unset config (safe defaults) still validates (RISK-W01-003).

## Pre-implementation run (status: failed → resolved) — config fields existed, rejection not yet implemented

```
=== RUN   TestProdUnsafeConfigKnobMatrix
=== RUN   TestProdUnsafeConfigKnobMatrix/log.format=text_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/log.level=debug_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/webhook.outbound.ssrf_protection_disabled=true_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/http.read_timeout=0_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/http.write_timeout=0_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/http.idle_timeout=0_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/log.level=unknown_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/log.format=unknown_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.addr=empty_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.read_header_timeout=0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.request_timeout=0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.max_body_bytes=0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.max_body_bytes<0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/db.max_conns=1_(below_floor_2)_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/db.max_conns=201_(above_ceiling_200)_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/db.query_timeout=50ms_(below_floor_100ms)_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/db.query_timeout=61s_(above_ceiling_60s)_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/schema_version=0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/environment=invalid_rejected_everywhere
--- FAIL: TestProdUnsafeConfigKnobMatrix (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/log.format=text_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/log.level=debug_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/webhook.outbound.ssrf_protection_disabled=true_rejected_in_prod (0.00s)
    --- FAIL: TestProdUnsafeConfigKnobMatrix/http.read_timeout=0_rejected_in_prod (0.00s)
    --- FAIL: TestProdUnsafeConfigKnobMatrix/http.write_timeout=0_rejected_in_prod (0.00s)
    --- FAIL: TestProdUnsafeConfigKnobMatrix/http.idle_timeout=0_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/log.level=unknown_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/log.format=unknown_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.addr=empty_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.read_header_timeout=0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.request_timeout=0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.max_body_bytes=0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.max_body_bytes<0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/db.max_conns=1_(below_floor_2)_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/db.max_conns=201_(above_ceiling_200)_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/db.query_timeout=50ms_(below_floor_100ms)_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/db.query_timeout=61s_(above_ceiling_60s)_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/schema_version=0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/environment=invalid_rejected_everywhere (0.00s)
=== RUN   TestHTTPTimeoutDefaultsMatchCS09
--- PASS: TestHTTPTimeoutDefaultsMatchCS09 (0.00s)
=== RUN   TestConnectionTimeoutZeroToleratedOutsideProd
--- PASS: TestConnectionTimeoutZeroToleratedOutsideProd (0.00s)
FAIL
FAIL	github.com/qatoolist/wowapi/kernel/config	0.187s
FAIL
```

## Post-implementation run (status: passed)

```
=== RUN   TestEnforceRouteContractsDefaultsOff
--- PASS: TestEnforceRouteContractsDefaultsOff (0.00s)
=== RUN   TestProdUnsafeConfigKnobMatrix
=== RUN   TestProdUnsafeConfigKnobMatrix/log.format=text_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/log.level=debug_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/webhook.outbound.ssrf_protection_disabled=true_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/http.read_timeout=0_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/http.write_timeout=0_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/http.idle_timeout=0_rejected_in_prod
=== RUN   TestProdUnsafeConfigKnobMatrix/log.level=unknown_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/log.format=unknown_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.addr=empty_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.read_header_timeout=0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.request_timeout=0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.max_body_bytes=0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/http.max_body_bytes<0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/db.max_conns=1_(below_floor_2)_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/db.max_conns=201_(above_ceiling_200)_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/db.query_timeout=50ms_(below_floor_100ms)_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/db.query_timeout=61s_(above_ceiling_60s)_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/schema_version=0_rejected_everywhere
=== RUN   TestProdUnsafeConfigKnobMatrix/environment=invalid_rejected_everywhere
--- PASS: TestProdUnsafeConfigKnobMatrix (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/log.format=text_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/log.level=debug_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/webhook.outbound.ssrf_protection_disabled=true_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.read_timeout=0_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.write_timeout=0_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.idle_timeout=0_rejected_in_prod (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/log.level=unknown_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/log.format=unknown_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.addr=empty_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.read_header_timeout=0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.request_timeout=0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.max_body_bytes=0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/http.max_body_bytes<0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/db.max_conns=1_(below_floor_2)_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/db.max_conns=201_(above_ceiling_200)_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/db.query_timeout=50ms_(below_floor_100ms)_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/db.query_timeout=61s_(above_ceiling_60s)_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/schema_version=0_rejected_everywhere (0.00s)
    --- PASS: TestProdUnsafeConfigKnobMatrix/environment=invalid_rejected_everywhere (0.00s)
=== RUN   TestHTTPTimeoutDefaultsMatchCS09
--- PASS: TestHTTPTimeoutDefaultsMatchCS09 (0.00s)
=== RUN   TestConnectionTimeoutZeroToleratedOutsideProd
--- PASS: TestConnectionTimeoutZeroToleratedOutsideProd (0.00s)
ok  	github.com/qatoolist/wowapi/kernel/config	0.193s
```

## Reviewer completion addendum — 2026-07-16

**Reviewer**: Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3).
**Review date**: 2026-07-16.
**Commit revision reviewed against**: HEAD 43b6e12 + remediation working tree 2026-07-16.
**Disposition**: Verified (existence + autopsy corroboration). Same disposition as ev-001 in this story.

This addendum retroactively fills the evidence-policy-mandated "reviewer" field. The original
record above (including any "Pending — conductor acceptance gate" line) is left unmodified per
the failed-evidence preservation convention — this is an appended addendum, not a rewrite.
