---
id: IMPL-W03-E02-S001
type: implementation-record
parent_story: W03-E02-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W03-E02-S001

## What was actually implemented

- **T1 — Fingerprint-scope confirmation/extension**: `SharedSection` now
  includes `Security` and `Webhook`, so `SharedFingerprint()` covers
  `Webhook.Outbound.AllowedHosts`, `Webhook.Outbound.AllowedCIDRs`, and
  `Security.TrustedIssuers`. A fingerprint-diff regression test proves the
  fingerprint changes when the allowlist or trusted-issuer list changes.
- **T2 — Boot-time egress-exception report**: `Framework.EgressExceptions()`
  returns a redacted, credential-free list of enabled escape hatches. The
  kernel logs it at boot time (`kernel/kernel.go`).
- **T3 — Allowlist change-audit trail**: `RecordAllowlistChange(before, after,
  recorder)` emits a redacted `AllowlistChange` record whenever the outbound
  allowlist differs from a baseline. The kernel emits this at boot by comparing
  the loaded config to `Defaults()`.
- **T4 — JWKS trusted-issuer config gate (D-07)**: `Security.TrustedIssuers`
  is a declared, fingerprinted config field. `auth.JWKSConfig` carries
  `TrustedIssuers` and `Env`; `NewJWKSKeySource` fails closed in `prod` profile
  when a custom `Client` is injected with no trusted-issuer list. The generated
  api main template passes the values through.
- **T5 — No-tenant-controlled-allowlist fitness check**: An AST-based test in
  `kernel/config/egress_fitness_test.go` verifies that `httpclient.New`,
  `httpclient.Config`, `auth.NewJWKSKeySource`, and `auth.JWKSConfig`
  construction call sites never read `context.Context` or `*http.Request`
  data. A deliberate-violation sub-test proves the checker is effective.

## Components changed

- `kernel/config` — `SharedSection`, `Security`, `EgressExceptions`,
  `RecordAllowlistChange`.
- `kernel/auth` — `JWKSConfig` and `NewJWKSKeySource` governance gate.
- `kernel` — boot-time egress logging and allowlist change audit.
- `internal/cli/templates/init` — generated api main passes trusted issuers and
  environment to JWKS construction.

## Files changed

- `kernel/config/shared.go`
- `kernel/config/security.go`
- `kernel/config/egress.go` (new)
- `kernel/config/shared_test.go`
- `kernel/config/egress_report_test.go` (new)
- `kernel/config/allowlist_audit_test.go` (new)
- `kernel/config/egress_fitness_test.go` (new)
- `kernel/auth/jwks.go`
- `kernel/auth/jwks_governance_test.go` (new)
- `kernel/auth/jwks_test.go` (pre-existing test fix: `okTenant`)
- `kernel/kernel.go`
- `internal/cli/templates/init/cmd_api_main.go.tmpl`
- `app/boot_extra_test.go` (pre-existing signature fix for `RegisterKind`)

## Interfaces introduced or changed

- Added `Security.TrustedIssuers []string`.
- Added `auth.JWKSConfig.TrustedIssuers []string` and `auth.JWKSConfig.Env config.Env`.
- Added `Framework.EgressExceptions() []EgressException`.
- Added `config.RecordAllowlistChange(before, after WebhookOutbound, rec AllowlistChangeRecorder)`.
- `SharedSection` expanded; existing fingerprints that did not include
  `Security`/`Webhook` will now differ.

## Configuration changes

New `security.trusted_issuers` config field. Optional; required only in `prod`
when a custom JWKS `*http.Client` is injected.

## Schema or migration changes

None.

## Security changes

- `SharedFingerprint()` now reflects egress escape-hatch changes.
- Boot-time report surfaces all enabled egress exceptions without exposing
  credentials.
- Allowlist mutations are audit-visible.
- Prod-profile custom JWKS clients must declare trusted issuers or fail
  readiness.

## Observability changes

- Boot log emits `egress_exceptions` when any are configured.
- Boot log emits `config_change`/`webhook.outbound.allowlist_changed` when the
  loaded allowlist differs from defaults.

## Tests added or modified

See Files changed. All five acceptance criteria have dedicated tests.

## Commits

Working-tree changes on top of `1626b11`.

## Pull requests

None yet.

## Implementation dates

2026-07-13.

## Technical debt introduced

None anticipated.

## Known limitations

- The boot-time egress report cannot report an undeclared custom JWKS client
  (it only sees `Security.TrustedIssuers`). The prod gate prevents the
  undeclared case from booting in production.

## Follow-up items

- wowsociety deployment-config audit for existing allowlist entries / custom
  JWKS-client injection (out of scope per `story.md`).

## Independent review gate (self-review)

A fresh self-review was run against the independent-review-gate checklist:

1. Goal coverage — all five tasks map to concrete, tested deliverables.
2. Requirement coverage — all SEC-06 T1–T5 sub-requirements implemented; T006
   (external review) remains pending.
3. Built-but-not-wired — `EgressExceptions`, `RecordAllowlistChange`, and the
   JWKS gate are all invoked on real paths (`kernel.New` / `NewJWKSKeySource`).
4. Runtime enforcement — JWKS gate returns an error at construction time; the
   fitness check inspects source files.
5. Generated artifacts — template updated and compiles.
6. Tests are real — each AC has a dedicated, non-skipping test; negative
   fixtures fail before the fix.
7. No regressions — `go test ./kernel/config/... ./kernel/auth/... ./app/...
   ./kernel/httpclient/...` passes.
8. Required infra — none beyond existing config/log paths.
9. Production-ready — config field is declared, validated, and fingerprinted.
10. Consistency/correctness — follows existing style; no drive-by changes.
11. Docs/traceability — `implementation.md`, `verification.md`, `closure.md`,
    `deviations.md`, `evidence/index.md`, artifacts/tasks indices updated.
12. One-pass reviewer test — the main risk (SharedSection extension changing
    existing fingerprints) is an expected, additive fingerprint change.

No third-party-review-level issues remain open.

## Relationship to the approved plan

Implementation matches `plan.md`. The T3 audit sink chosen is a structured-log
record emitted by `kernel.New`, which is the only available sink at boot time
(no tenant transaction exists). This is recorded as a deliberate choice rather
than a deviation because `plan.md` left the exact sink TBD.
