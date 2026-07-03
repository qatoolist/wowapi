# Decision Log

Format per entry: context → options → decision → tradeoffs → affected files/tests.
Blueprint deviations MUST land here before the code that implements them.

## D-0001 — Preflight: `kernel/secrets` added to the package map
- **Context:** 12-configuration-and-deployment referenced `kernel/secrets` types but the 04 package
  map never defined the package (Goal 2 preflight item).
- **Options:** (a) fold secrets into `kernel/config`; (b) separate `kernel/secrets` base package.
- **Decision:** (b). `kernel/secrets` = `Provider` port + `Ref` parsing, stdlib-only, graph base;
  `kernel/config` imports it. Adapters implement providers; `app` resolves refs at boot.
- **Tradeoffs:** one more public package; in exchange adapters don't import `kernel/config` and the
  graph stays layered.
- **Affected:** docs/blueprint/04 §2 (new row), 11 §2 kernel list, kernel/secrets (Phase 1 code).

## D-0002 — Preflight: config type naming standardized
- **Context:** blueprint mixed `config.Config`, `config.MustLoad()`, and `config.Framework`.
- **Decision:** framework-owned struct is **`config.Framework`** (in `wowapi/kernel/config`);
  the product-owned type is **`Config`** in the product's **`internal/appcfg`** package
  (scaffolded by `wowapi init`), embedding `config.Framework`, loaded via `appcfg.Load/MustLoad`.
  `kernel.Kernel.Cfg` is `config.Framework`.
- **Affected:** docs/blueprint/06 §3, 11 §3, 12 §2.

## D-0003 — Preflight: CLI config tooling never imports product packages
- **Context:** installed CLI is prebuilt; it cannot link product config types, but
  `config validate/schema/...` must operate on the *product's* composed config.
- **Options:** (a) CLI parses YAML against a generated JSON schema only; (b) generated
  product-local checker binary the CLI shells out to; (c) plugin loading (rejected: runtime magic).
- **Decision:** (b) with (a) as its transport: `wowapi init` scaffolds `tools/configcheck/main.go`
  (imports `internal/appcfg` + `wowapi/kernel/config`; emits schema/validation/redacted-effective
  JSON on stdout); the CLI runs `go run ./tools/configcheck` in the product repo and formats the
  result. Framework-repo fallback: `config.Framework` alone.
- **Tradeoffs:** requires Go toolchain for config commands in product repos (already required);
  in exchange, full typed validation with zero import-direction violations.
- **Affected:** docs/blueprint/12 §8.

## D-0004 — Preflight: CLI command listings use one command per line
- **Context:** 11 §5 used `wowapi config init | validate | …` which reads as a shell pipe.
- **Decision:** every doc lists each command on its own line. **Affected:** docs/blueprint/11 §5.

## D-0005 — Preflight: acyclicity re-verified with `kernel/secrets`
- **Decision:** graph remains acyclic: `kernel/secrets` (stdlib only) ← `kernel/config` ←
  other `kernel/*` (receive sub-structs by value; no config imports needed) ← `module` ← `app`;
  `adapters` → `kernel/*` only. Encoded in `scripts/lint_boundaries.sh` from Phase 0.

## D-0006 — Phase 0: walking-skeleton scope for `module.Context`
- **Context:** the full Context interface (06 §2) references many kernel packages that don't exist
  yet; stubbing them all would create broad partial implementations (banned by preflight rule 3).
- **Decision:** Phase 0 ships `module.Module` exactly as specified plus a **minimal** `Context`
  (Logger, Config→`config.ModuleView`) with the blueprint-documented growth path; each later phase
  adds its own accessor alongside the capability it delivers. Interface widening pre-v0.1.0 is an
  accepted breaking change (semver v0 rules).
- **Affected:** module/module.go; noted in evidence/phase-00/proof-bundle.md.

## D-0007 — Phase 0: Go toolchain version
- **Context:** blueprint says Go ≥ 1.23; local toolchain is 1.26.4.
- **Decision:** `go.mod` declares `go 1.26` (repo floor). CI pins the same; revisit at v1.
- **Affected:** go.mod.

## D-0008 — Phase 0: `wowapi version` implementation
- **Decision:** CLI version from `runtime/debug.ReadBuildInfo` (main module version when installed
  via `go install …@vX.Y.Z`; `(devel)` in-repo), with `-ldflags -X` override hook for goreleaser.
  Dependency-mismatch warning parses the nearest `go.mod` for the wowapi requirement.
- **Affected:** cmd/wowapi, internal/buildinfo.

## D-0009 — Phase 0: vocabulary denylist pragmatics in boundary lint
- **Context:** blueprint 00 §5 lists denylist words including over-generic ones (building, wing,
  flat, member) that would false-positive constantly in code ("building the request", struct
  members).
- **Decision:** the grep-based Phase 0 lint enforces the unambiguous terms (society, housing,
  chairman, treasurer, defaulter, conveyance, redevelopment, agm, maintenance_bill); generic terms
  are covered by code review until the Phase 5 AST-based lint can check identifiers only.
- **Affected:** scripts/lint_boundaries.sh; revisit at Phase 5.

## D-0010 — Phase 0→1: `environment` is fail-closed in deployed processes (SEC-1)
- **Context:** security review: `Defaults()` sets `environment=local`; a prod deploy that forgets
  to set it would silently validate under local (lenient) rules.
- **Decision:** the Phase 1 loader errors when `environment` is absent from every layer; the
  compiled `local` default serves only `Defaults()` in tests/local tooling. Blueprint 12 §4
  updated; Phase 1 exit criteria include a test for this.
- **Affected:** docs/blueprint/12 §4; kernel/config loader (Phase 1).
