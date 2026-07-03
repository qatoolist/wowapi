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
- **Amendment (Phase 1, ARCH-13):** "receive sub-structs by value" still requires importing
  `kernel/config` for the *types* (e.g. `kernel/logging` imports `config.Log`/`config.Fingerprint`).
  That is a types-only kernel→kernel edge, cycle-free and consistent with 04 §2; what stays
  forbidden is other packages *loading* config or reading stores at runtime.

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

## D-0011 — Phase 1: first third-party dependency, `gopkg.in/yaml.v3`
- **Context:** the layered loader must parse `configs/*.yaml`; blueprint 12 §2 already assumes YAML
  (product `Modules map[string]yaml.Node` example). Repo had zero deps.
- **Options:** (a) hand-rolled YAML subset (rejected: config parsing is exactly where correctness
  bugs hide); (b) `gopkg.in/yaml.v3` (stable, no transitive deps); (c) JSON-only config (rejected:
  blueprint mandates YAML overlays).
- **Decision:** (b). The "kernel/config imports only stdlib + kernel/secrets" rule in 12 §2 governs
  the *internal package graph* (acyclicity), not third-party libs; yaml.v3 keeps the graph acyclic.
- **Affected:** go.mod, kernel/config loader.

## D-0012 — Phase 1: binder scope — `conf`/`default`/`required` tags + `Validate()` hook
- **Context:** blueprint 12 §2 shows a full tag DSL (`conf`, `default`, `validate:"min=…,max=…"`,
  `unsafe`, `redact`, `doc`); Phase 0 shipped hand-written `Framework.Validate()` with accumulated
  errors; risk R5 warns against a reflection-heavy config system.
- **Decision:** ONE audited binder implementing: `conf` key mapping (embedded structs flatten),
  `default:"…"` tags, `required:"true"`, strict unknown-key rejection, scalar conversion
  (string/bool/ints/floats/duration/Env/Secret/slices), `unsafe:"true"` prod refusal (stage warns),
  and `doc` tags (feed `config schema`). Range/cross-field/enum checks stay in code via a
  `Validate() error` hook (already accumulates all errors) — no min/max tag mini-language.
  A drift-guard test asserts tag defaults reproduce `Defaults()`.
- **Tradeoffs:** two places express constraints (tags for shape, code for ranges); in exchange the
  binder stays small enough to audit and R5 stays contained.
- **Affected:** kernel/config (bind/load/schema), config_test.go.

## D-0013 — Phase 1: env secret provider lives at `adapters/secrets/envprovider`
- **Context:** D-0001 put the `Provider` port in `kernel/secrets` with implementations in adapters;
  blueprint 04 §1 lists `adapters/secrets/`.
- **Decision:** first provider is `adapters/secrets/envprovider` (`secretref://env/<VAR>` →
  process environment), with an injectable lookup func for tests. Cloud providers follow the same
  layout later (`adapters/secrets/<name>provider`).
- **Affected:** adapters/secrets/envprovider, app boot wiring, CLI config commands.

## D-0014 — Phase 1: loader API is `Load[T]` (blueprint signature) + `LoadDetailed[T]`
- **Context:** blueprint 12 §2 fixes `Load[T any](opts Options) (T, Fingerprint, error)`, but
  `config doctor` needs per-key provenance and stage-unsafe warnings need a channel out.
- **Decision:** keep the blueprint signature as the primary API; add
  `LoadDetailed[T any](opts Options) (Loaded[T], error)` where `Loaded` carries Config,
  Fingerprint, Provenance (key → layer) and Warnings. `Load` delegates to `LoadDetailed`.
  Fingerprint = SHA-256 of the canonical *redacted* effective config JSON (structural `Secret`
  redaction makes this safe by construction).
- **Affected:** kernel/config/load.go, internal/cli (validate/print/doctor), app views.

## D-0015 — Phase 1: `unsafe` knob mechanism ships now; first framework knob later
- **Context:** 12 §4 requires a per-knob prod-refusal matrix, but every listed dev convenience
  (fake token issuer, SQL echo, public pprof, permissive CORS) belongs to a later-phase component;
  adding a dead config field now would be a partial implementation (banned by preflight rule 3).
- **Decision:** the binder's `unsafe:"true"` handling (prod=error, stage=warning) is implemented
  and matrix-tested in Phase 1 against test-local structs (the binder is generic, so the tests are
  real end-to-end loader tests); `AllowFlags`-style CLI flags refused in prod is the one live
  production rule now. Each later phase adds its real knobs with `unsafe:"true"` + a matrix entry.
- **Affected:** kernel/config loader + tests; later phases' config sections.

## D-0016 — Phase 1 review: `config.Options` final shape (supersedes blueprint 12 §2 sketch)
- **Context:** review finding ARCH-12 — the implemented Options diverged from the blueprint sketch
  (`AllowFlags bool` dropped; `Environ []string` and `Flags map[string]string` added).
- **Decision:** keep the implemented shape. `Flags` presence + the prod refusal rule subsumes
  `AllowFlags` (an empty map IS "flags not allowed"); `Environ` makes the env layer hermetic in
  tests instead of mutating the process environment. Blueprint 12 §2 updated to match.
- **Affected:** kernel/config/load.go, docs/blueprint/12 §2.

## D-0017 — Phase 1 review: the environment gate is not overridable downward (SEC-5)
- **Context:** security review reproduced two downgrades: an env var could flip a committed
  `environment: prod` to `local` (disabling every prod check), and a flag setting `environment`
  escaped the flags-refused-in-prod guard by lowering the value the guard reads.
- **Decision:** trust rules in the loader: (1) `environment` may never come from the flag layer;
  (2) an env var may *supply* `environment` only when no config file sets it — any mismatch with a
  file value is an error, not an override; (3) prod checks and the flag guard key off the
  file-layer value when present. The blueprint §1 table's "env vars set `environment`" reading is
  narrowed accordingly (12 §4 updated).
- **Tradeoffs:** a platform can no longer "promote" an image whose files say `dev` by env var —
  intentional; environment changes ship as config changes.
- **Affected:** kernel/config/load.go; tests TestLoadEnvironmentNotDowngradableByEnvVar,
  TestLoadEnvironmentNeverFromFlags, TestLoadFlagDowngradeStillRefusedInProd; docs/blueprint/12 §4.

## D-0018 — Phase 1 review: module namespaces are file-layer only (for now) (ARCH-8)
- **Context:** env-var/flag values reach the tree as strings; a module's strict typed Decode would
  fail with a confusing per-module JSON error at boot (`"4"` into an int field).
- **Decision:** the loader rejects `modules.*` keys sourced from the env-var or flag layers with a
  clear error at load time. Lifted when module config decoding learns scalar string coercion
  (revisit at Phase 5 with the module SDK).
- **Affected:** kernel/config/bind.go (namespaces case); TestLoadModuleNamespaceViaEnvVarRejected;
  docs/blueprint/12 §3.

## D-0019 — Phase 1 review: unsafe knobs are judged on final bound values (SEC-3/SEC-4)
- **Context:** security review reproduced two fail-open holes: an unsafe knob whose unsafe value
  is its compiled default was never checked (check lived on the "value present in tree" path), and
  unsafe tags on struct/Secret/slice/pointer fields were silently unenforced.
- **Decision:** enforcement moved to a post-bind pass over the fully bound struct: any
  `unsafe:"true"` field with a non-zero final value refuses prod / warns stage, regardless of
  which layer (or default tag) produced the value and regardless of field kind.
- **Affected:** kernel/config/bind.go (enforceUnsafe), load.go; tests
  TestLoadUnsafeDefaultRefusedInProd, TestLoadUnsafeStructKnobRefusedInProd.
