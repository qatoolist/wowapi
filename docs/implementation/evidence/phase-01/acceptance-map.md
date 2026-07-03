# Phase 1 — Acceptance Map

Phase 1 exit criteria (Goal 2 Phase 1 + phase-plan row 1 + blueprint 12 §11 items in scope) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | Layered loading: defaults → base.yaml → <env>.yaml → env vars → secret refs → local flags | `kernel/config/load.go`; tests `TestLoadPrecedence`, `TestLoadEnvFileBeatsBase`, `TestLoadScalarConversions` |
| 2 | Strict validation, ALL errors accumulated (unknown keys, missing values, bad ranges, schema version) | `TestLoadErrorsAccumulate`, `TestLoadUnknownKeyViaEnvVarTypo`, `TestLoadSchemaVersionBounds`, `TestLoadRequiredFieldMissing`, `TestLoadTypeMismatchReported`; CLI smoke (command-log #11) shows 3 errors reported at once |
| 3 | SEC-1 fail-closed environment: missing/unknown env fails; prod safety fails closed; unsafe knobs rejected in prod | `TestLoadEnvironmentFailClosed`, `TestLoadUnknownEnvironmentRejected`, `TestLoadProdSafetyFloor`, `TestLoadFlagsRefusedInProd`, `TestLoadUnsafeKnobMatrix` (per-knob prod/stage/dev matrix, D-0015) |
| 4 | Secret refs resolved at boot; no raw secret via fmt/JSON/text/slog/errors/diagnostics | `TestLoadSecretResolution`, `TestLoadRawSecretValueRejected` (raw value never echoed), `TestLoadedConfigNeverPrintsSecrets` (%v/%+v/%#v/JSON), `kernel/config/secret_test.go` (Phase 0), `TestConfigSecretStructuralRedaction` (through slog) |
| 5 | Module config isolation: only `modules.<name>.*` reachable | `TestLoadModuleNamespaces` (catch-all decode sees only own keys), app `context_test.go` isolation test; no traversal API exists on ModuleView |
| 6 | Product config composition; `config.Framework` framework-owned | `prodConfig` test type embeds Framework (blueprint 12 §2 shape) — `TestLoadPrecedence`/`TestLoadUnsafeKnobMatrix` run through it; D-0002 naming honored |
| 7 | API/worker/migrate narrowed config views | `app/views.go`; `app/views_test.go` reflection assertions (WorkerConfig has no HTTP; MigrateConfig has no HTTP/Modules); per-view fingerprints |
| 8 | kernel/logging: slog JSON, redaction, startup config fingerprint | `kernel/logging/logging.go`; 11 tests incl. `TestLogStartupEmitsRequiredFields`, redaction suite |
| 9 | CLI: `config validate`, `config print --redacted`, `config doctor` skeleton, `config schema` | `internal/cli/config_cmd.go`; 19 tests; live smoke in command-log #11 |
| 10 | Startup/shutdown skeleton | `app/run.go` RunHooks (ordered start, reverse stop, timeout-bounded, error joining); 7 tests |
| 11 | No package cycles; import law holds | `scripts/lint_boundaries.sh` → `boundary lint: OK` (command-log #4, #10) |
| 12 | Evidence bundle + reviews | this directory; review-findings.md (security/config + architecture agents) |
| 13 | `make ci` green; containerized CI (R4) | command-log #10 (host) and #12 (container) |
