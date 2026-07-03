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

## D-0020 — Phase 2: `kernel/model` ships complete now
- **Context:** phase-plan row 2 doesn't name kernel/model, but TenantDB helpers key on
  `model.TenantScoped`, testkit fixtures return typed handles, and migrations follow its column
  conventions — building database/testkit against ad-hoc types would create the partial
  implementations preflight rule 3 bans.
- **Decision:** implement 04 §3 verbatim in Phase 2: BaseFields/TenantScoped/Auditable/CreatedOnly/
  Versioned/Temporal/Statused + Ref value objects + `IDGen` port with a UUIDv7 default.
  Deps: google/uuid (v7 support), shopspring/decimal (Money).
- **Affected:** kernel/model; go.mod.

## D-0021 — Phase 2: DB DSNs validated at process-view narrowing, not by a required tag
- **Context:** blueprint 12 §2 sketches `DSN Secret validate:"required"`, but a tag-required DSN
  would make every Framework load (CLI schema/validate in the framework repo, config-only tests,
  Defaults()) fail without a database — and §7 says each process receives only what it needs.
- **Decision:** `config.DB` fields are optional at load; `app.NewAPIConfig`/`NewWorkerConfig`
  error when the runtime DSN is unset, `app.NewMigrateConfig` errors when the migrate DSN is
  unset. Raw (non-secretref) DSN strings remain structurally impossible (Secret.UnmarshalText).
- **Affected:** kernel/config/config.go (DB section), app/views.go, tests.

## D-0022 — Phase 2: integration tests use env-DSN + template-database clones, not testcontainers
- **Context:** test-strategy sketched testcontainers; the compose stack already provides Postgres
  both on the host (localhost:5432) and inside the tools container (DATABASE_URL), and
  testcontainers-go would be the largest dependency in the tree by far.
- **Decision:** testkit connects via `WOWAPI_TEST_DSN` (fallback `DATABASE_URL`); tests skip with
  a clear message when neither is set. Speed: kernel migrations run once per process into a
  template database; each test gets `CREATE DATABASE … TEMPLATE …` + drop on cleanup.
  Testcontainers can be layered later without API changes. test-strategy.md updated.
- **Affected:** testkit/db.go, Makefile test-integration, docs/implementation/test-strategy.md.

## D-0023 — Phase 2: runtime RLS identity is a non-superuser login (revised after SEC-11/SEC-12)
- **Context:** RLS must be enforced against a role that is non-owner, non-superuser, and lacks
  BYPASSRLS. The original decision (superuser admin login + `SET ROLE app_rt`) was reproduced by
  the Phase 2 security review to be escapable: a module running arbitrary SQL as designed can
  `RESET ROLE` back to the superuser login mid-transaction and read every tenant (SEC-11), and a
  pool wired against an over-privileged DSN silently disables RLS with no signal (SEC-12).
- **Decision:** deployed processes MUST authenticate as a **non-superuser login mapped to app_rt**;
  `SET ROLE` from a superuser is no longer an accepted production posture. Defense in depth, all
  shipped:
  1. `database.WithConnRLSGuard()` refuses, at connect, any pool whose effective role is superuser
     or BYPASSRLS (fail-closed pool construction).
  2. `database.Manager` `WithRole` re-asserts `SET LOCAL ROLE` per tenant tx (survives pool-state
     leaks across checkouts), and `WithRLSGuard` re-checks enforcement per tenant tx.
  3. `app_rt`/`app_platform` stay NOLOGIN in the committed migration — no password ships. The
     testkit grants `app_rt` a local-only LOGIN out-of-band (never committed) and connects as it,
     modelling production exactly; the SEC-11 escalation test passes only because the login is a
     genuine non-superuser.
- **Tradeoffs:** product deployment docs must state the non-superuser-login requirement plainly
  (Phase 10/12); `WithSetRole` is retained only as a session baseline for tooling, not a security
  boundary.
- **Affected:** migrations/00001_bootstrap.sql, kernel/database (pool guards, per-tx role),
  testkit/db.go, docs/blueprint/12 (deployment note, Phase 10).

## D-0026 — Phase 2 review: global identity tables granted to app_platform, not app_rt (SEC-13)
- **Context:** global tables carry no RLS (03 §1); granting them to `app_rt` let any module read or
  tamper with the whole cross-tenant membership graph via ordinary tenant-tx SQL.
- **Decision:** 00002 grants SELECT/INSERT/UPDATE on tenants/users/user_tenant_access to
  `app_platform` only. Kernel identity services run platform transactions under that role via a
  dedicated pool; that pool is wired when the first such service lands (Phase 4). In Phase 2 the
  runtime `app_rt` simply cannot touch the global spine — correct for now.
- **Affected:** migrations/00002_core_identity.sql; kernel/database.Manager.Platform (pool wiring
  deferred to Phase 4, tracked in phase-plan row 4).

## D-0027 — Phase 2 review: per-source migration history tables (ARCH-16)
- **Context:** goose derives a version from the leading filename digits and tracks one history
  table; kernel `00001..` and a module's `0001..` would collide, making the documented
  multi-source model impossible.
- **Decision:** `database.Migrate(ctx, pool, src, source)` uses a per-source history table
  (`goose_version_<source>`); the kernel source is `migrations.SourceName` ("wowapi"), each module
  supplies its own. Independently-numbered sources coexist. `Migrate` returns `MigrateResult{Version,
  Applied}` so idempotency (`Applied==0` on rerun) is assertable.
- **Affected:** kernel/database/migrate.go, migrations/migrations.go, internal/tools/migrate,
  testkit; docs/blueprint/03 §5 wording.

## D-0028 — Phase 2 review: ExpectOneRow distinguishes 0-row conflict from >1-row bug (ARCH-20)
- **Decision:** 0 rows → `ErrVersionConflict` (409/412); >1 row → a distinct internal error (500),
  never masked as a conflict — a too-broad WHERE on a versioned aggregate is a bug, not contention.
- **Affected:** kernel/database/errors.go.

## D-0029 — Phase 2 review: `config.Pool` sub-struct absorbs shared pool knobs (ARCH-17)
- **Decision:** pool knobs live in `config.Pool`, embedded in `config.DB` and in the app views'
  `RuntimeDB`/`MigrateDB`; new pool fields propagate to every narrowed view without editing the
  narrowing code, closing the silent-drop drift.
- **Affected:** kernel/config/config.go, app/views.go.

## D-0030 — Phase 2 review: actor binding stays optional until the actor model exists (ARCH-19)
- **Context:** 05 §2 says `WithTenant` binds `app.tenant_id` AND `app.actor_id` "error if absent".
  The Phase 2 TxManager hard-fails on missing tenant but binds actor only when present. There is no
  actor model, no audit triggers, and no `created_by` defaults reading `app.actor_id` until Phase 4.
- **Decision:** keep actor binding optional for Phase 2 (tenant remains fail-closed). When Phase 4
  introduces the actor/audit machinery that actually consumes `app.actor_id`, `WithTenant` (RW)
  will require it (fail-closed at the door), while `WithTenantRO` read paths stay actor-optional.
  Recorded now so the deviation from 05 §2 is explicit, not silent.
- **Affected:** kernel/database/txmanager.go; revisit at phase-plan row 4.

## D-0024 — Phase 2: TenantDB grows per-phase accessors; sentinel errors until kernel/errors
- **Context:** 05 §2's TenantDB carries Outbox()/Audit()/Resources(), owned by Phases 4/6; the
  error taxonomy arrives in Phase 3.
- **Decision:** Phase 2 TenantDB = DBTX only (D-0006 growth pattern; accessors land with their
  capabilities). Version-conflict/no-tenant failures are exported sentinel errors in
  kernel/database now and get mapped into the Phase 3 taxonomy when it exists.
- **Affected:** kernel/database; revisit notes in phase-plan rows 3/4/6.

## D-0025 — Phase 2: only kernel migrations 000–001 ship; RLS proven on probe tables
- **Context:** tenants/users/user_tenant_access (001) are GLOBAL tables — RLS-bearing kernel
  tables start at migration 002+ (later phases). Phase 2 must still prove the RLS mechanics.
- **Decision:** ship 000 (extensions, roles, `app_tenant_id()`), 001 (tenants/users/access) per
  phase plan; `testkit.AssertRLSIsolation` + integration tests create standard-convention probe
  tables (tenant_id + ENABLE/FORCE + policy) to prove SET LOCAL binding, isolation, WITH CHECK,
  and no-tenant-context failure. Each later migration adding tenant tables reuses the same
  assertion catalog-driven.
- **Affected:** migrations/, testkit/asserts.go, kernel/database integration tests.
