---
id: W01-ACCEPTANCE
type: wave-acceptance
wave: W01
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01 — Wave-level acceptance

## AC-W01-01 — Zero-cost linter set enabled

`sqlclosecheck`, `rowserrcheck`, `bodyclose`, `wastedassign`, `makezero`, `musttag`, `testifylint`
are enabled in `.golangci.yml` at zero hits; `noctx`'s 2 named prod hits and the 1 `copyloopvar` hit
are fixed; the enablement run itself is the fail-first evidence. Traces to W01-E01-S001.

## AC-W01-02 — Judged linter set enabled and triaged

`gosec` (with the CS-23 triage list: G704 JWKS annotation, G115 conversions reviewed/bounded/
annotated, G120 folded into FBL-09), `errorlint`, `exhaustive` (fail-closed default arms annotated),
`forcetypeassert` (checked assertions added at named sites), `usestdlibvars` are enabled with every
hit resolved (fixed or justified with an inline annotation). Traces to W01-E01-S002.

## AC-W01-03 — Supply-chain and hook hygiene

`go mod verify` runs in CI; a license-scanning signal exists (Trivy license scanner or `go-licenses`)
while dependency-review remains visibility-dormant; the pre-push hook no longer silently skips DB
tests. Traces to W01-E01-S003.

## AC-W01-04 — Trace/log correlation live

A log record emitted inside a traced (recording-span) request carries `trace_id`/`span_id` attrs;
absent without an active span (not empty-string noise); no-op tracer path is allocation-neutral per
benchmark. Traces to W01-E02-S001.

## AC-W01-05 — pgx query tracer live

Pgx query spans appear as children in the trace tree, attached via a thin in-kernel `pgx.QueryTracer`
consuming the existing observability `Tracer` port (D-08), not `otelpgx`. Traces to W01-E02-S002.

## AC-W01-06 — HTTP server timeouts enforced

All four timeouts (read/write/idle/header) are config-driven with safe defaults; prod profile rejects
zero-value timeout config; CSRF middleware applies `MaxBytesReader`. Traces to W01-E03-S001.

## AC-W01-07 — Central validation enforcement live

Boot rejects a POST/PUT/PATCH route with no declared `RouteMeta.Request` contract (behind a profile
flag first); an adversarial invalid-DTO POST to a route that does declare a contract returns 400 with
field errors. Traces to W01-E03-S002.

## AC-W01-08 — Generator emits valid, boot-passing output

`gen crud`'s emitted permission verb is in the closed set (`deactivate`, not `delete`); the
generator-output-boots test fails before the fix and passes after; `TestGenCRUDPermissionKeys` no
longer test-locks the bug. Traces to W01-E04-S001.

## AC-W01-09 — Documentation reconciled

PLAN's §6-vs-§9 DX-05 status inconsistency is fixed; DX-05's residual reconciliation items (per
`requirement-inventory.md` row DX-05) are resolved; the resolved wowsociety upstream findings (e.g.
PF-2 once DX-02 lands) are marked closed in FBL-03's target register. Traces to W01-E04-S002.

## AC-W01-10 — e2e flake diagnosed

T-TEST-01's reproduction is attempted under `-count`+parallel; the diagnosis (confirmed cause, or an
honest "not reproducible, downgraded to monitoring") is recorded — the withdrawn "shared-DB
concurrency" cause is not silently re-asserted. Traces to W01-E04-S003.

## AC-W01-11 — Independent review passed

Every W01 story has passed independent review per mandate §14, with the FBL-08/FBL-09 stories
specifically checked for compat-flag discipline and the DX-02 story specifically checked for the
test-lock fix (RISK-W01-005).

## Acceptance authority

Framework architecture lead / developer-experience lead (role-based, split by epic per `wave.md`).

## Acceptance record — 2026-07-13

Satisfied 2026-07-13. AC-W01-01 through AC-W01-11 met; independent review passed (W01ReviewGate);
accepted by conductor. See `closure-report.md` for evidence mapping and open items.
