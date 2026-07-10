# Evidence: wowsociety framework gap closure (GAP-001…GAP-008)

Spec: [`docs/implementation/wowsociety-framework-gap-analysis.md`](../../wowsociety-framework-gap-analysis.md)
Framework branch: `feat/wowsociety-framework-gaps` (`287abc3..be955c4`, 24 commits)
Product branch (wowsociety): `feat/consume-framework-gap-apis` (base `6ac9404`, commits `7488ba0..240fec9`)
Date: 2026-07-10. Neither branch pushed.

## What shipped (framework)

| Gap | Capability | Key commits | Landed where |
|---|---|---|---|
| GAP-001 | `kernel/i18n` (catalog/bundle/fallback), `httpx.Locale` middleware (Accept-Language → context, `Content-Language`), localized problem details + validation messages, testkit locale asserts | `97bba01`, `6128bf7` | `kernel/i18n`, `kernel/httpx`, `kernel/validation`, `testkit` |
| GAP-002 | Production S3/MinIO adapter (presign PUT/GET, Stat w/ checksum fast path, ranged Peek, idempotent Delete, KindNotFound mapping, bucket validation/auto-create) | `89f3e41`, `f1981d5`, `96ad0ae` | `adapters/storage/s3` |
| GAP-003 | Seed-sync lifecycle: `wowapi seed sync` CLI, generated migrate main runs migrations → `seeds.Sync`, `app.CatalogsSeeded` readiness check | `0c03569`, `e11c90b`, `b85be99` | `internal/cli`, templates, `app` |
| GAP-004 | `PermissionSeed.StepUp` (`step_up:` YAML) → boot → registry; migration 00029 `permissions.step_up` persisted by Sync; `auth.Claims.AMR` → `Actor.AMR`; `testkit.WithAMR`; e2e challenge test | `6868f0a`, `bb933d1`, `849d788` | `kernel/seeds`, `kernel/auth`, `app`, `migrations`, `testkit` |
| GAP-005 | `kernel/mfa`: TOTP/HOTP (RFC 4226/6238 vectors), OTP gen + salted constant-time hash/verify, TTL/attempt policy, SMS/email sender ports + test adapters | `8e8e7e0`, `c890996` | `kernel/mfa` |
| GAP-006 | `module.Context.Privileged()`: audited tenant-bound `Relationships().Grant/Revoke` + `Rules().ActivateTenant` with module-prefix ownership; no product SECURITY DEFINER needed | `15e9043`, `87b9607` | `kernel/privileged`, `module`, `app` |
| GAP-007 | `rules.SyncDefinitions` (registry → `rule_definitions`, idempotent) in the generated migrate lifecycle; schema validator now enforces min/max/exclusive bounds, min/maxLength, pattern, min/maxItems, required | `f304571`, `8162078`, `bc6dba6`, `6cee5ed` | `kernel/rules`, `internal/cli` |
| GAP-008 | Scaffold: storage (S3/MinIO) + OIDC/JWT config sections & wiring, i18n locale enablement, configcheck round-trip, generated migrate loads composed appcfg (one overlay serves api/worker/migrate) | `e487887`, `d2a4164` | `internal/cli/templates`, `internal/cli` |

## What shipped (product — wowsociety consumes, per-gap commits `7488ba0..240fec9`)

Removed/shrunk (verified by independent audit, greps + fresh tests): product S3 adapter dir (minio-go now indirect-only), manual seeds.Sync block, direct step-up registration + `authenticator.go` JWT-reparse wrapper (zero reparse remains), raw TOTP/HOTP/OTP implementations (thin wrappers over `kernel/mfa` remain), SECURITY DEFINER bridges unreferenced (dropped via new forward migrations identity 00006 / policy 00005 — applied history untouched), rule-definition SQL mirror retired (drift guard rewritten as a SyncDefinitions convergence proof), product locale negotiation deleted (`internal/i18n` reduced to product-owned message content). `FRAMEWORK_VERSION` = `be955c4`.

## Independent Review Gate (mandatory) — PASSED

Three independent dimensions, all executed against live Postgres + MinIO:

1. **Whole-branch correctness/security review** (fresh reviewer, executed the proving tests): SPEC COMPLIANCE ✅ all 8 gaps; QUALITY Approved; 0 Critical / 0 Important; 1 Minor doc-drift fixed (`be955c4`). Privileged-service model verified against live `pg_policies`; accepted Lows confirmed (MFA modulo bias ~2⁻³⁰; remaining uncovered lines are defensive DB-error wraps).
2. **Cross-repo Definition-of-Done audit** (fresh auditor): PASS, no findings — DoD bullets, per-gap Remove/Replace lists, and 7-step refactor checklist verified with evidence; applied-migration safety confirmed; no built-but-not-wired or deferred-claimed-as-done instances.
3. **Fresh authoritative gates** (independent runner, `-count=1`, strict env so skips are impossible):
   - wowapi: lint-new 0, full `make lint` 0, boundaries OK, `go test ./...` **1329 PASS / 0 FAIL / 0 SKIP** (52 pkgs), test-security 0 fail, coverage **91.8%** ≥ 90% floor, `review_gate.sh` clean.
   - wowsociety: `framework-verify` OK (`be955c4`), `go test -race -count=1 ./...` **91 PASS / 0 FAIL / 0 SKIP**, 0 data races, vet/build clean.

Per-task gates during the build: every gap had its own independent reviewer; findings fixed pre-acceptance included one **Critical** (kernel/mfa digits=10 uint32 modulus overflow, fixed `c890996`), one boundary-lint violation (kernel test importing an adapter, fixed `849d788`), a raw-key-leak on i18n catalog miss, a migrate-template overlay rejection (`d2a4164`), and assorted Lows.

## Traceability

- CHANGELOG `[Unreleased]` carries entries for all 8 gaps.
- User-guide sections added/extended: storage (build-deploy), seed-sync lifecycle, step-up/AMR, i18n, module development (privileged services), rules lifecycle, scaffold output.
- wowsociety upstream notes / PILOT-FINDINGS marked resolved against framework `be955c4` (historical record preserved).
- Full per-task reports + gate logs: session scratchpad `briefs/` (gap-00N-report.md, final-review-report.md, gate-dod-audit.md, gate-gates-run.md).
