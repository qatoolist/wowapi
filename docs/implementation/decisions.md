# Decision Log

Format per entry: context → options → decision → tradeoffs → affected files/tests.
Blueprint deviations MUST land here before the code that implements them.

## D-0090 — B11/B12/B13 re-verification (2026-07-11): all three stay PARKED, evidence reproduced independently
- **Context:** `docs/implementation/framework-backlog-p2-decisions.md` (2026-07-10, commit `0e578e8`) parked
  B11/B12/B13 with measured evidence. A `/goal` run asked for an independent re-verification from source
  (not from the doc) before accepting the parked status, with a mandate to reopen and implement the smallest
  safe increment if evidence had shifted. No kernel/product code changed between the two passes.
- **B11 (radix router) — reproduced, decision unchanged.** Ran `BenchmarkDispatch` live (Apple M3 Max,
  `-benchtime=200000x`): 50/500/2000 routes → 571.2 / 590.4 / 629.6 ns/op, flat at 14 allocs/op — an even
  flatter curve than the recorded 579.9/607.3/656.2 ns/op (≈10% growth over 40× routes vs. the doc's ≈13%).
  Matches `bench-budgets.txt`; independently re-confirmed by the review-gate reviewer's own two additional runs
  (596-633 ns/op range, same flat 14-allocs shape). Confirmed `RouteMeta` (`kernel/httpx/router.go:20-45`) is
  still the mandatory, boot-validated contract (`meta.validate()` fails registration; `Router.Handle`
  accumulates errors that fail boot). **Correction (caught by the independent review gate):** an earlier draft
  of this entry claimed `Route` "backs OpenAPI/permission-sync" as if both were equally live — only
  permission-sync is: `Router.Permissions()` is consumed at `app/boot.go:254` to boot-gate every route's
  permission against the registry. OpenAPI generation is **not** wired to `Route`/`Router.Routes()` today —
  `internal/cli/openapi_cmd.go` only merges hand-authored fragments, and `Router.Routes()` has zero production
  callers in either repo (test-only). `router.go:46-47`'s own doc comment already hedges this correctly
  ("exposed for permission-sync and OpenAPI generation **(later phases)**"). No reopen trigger met — dispatch
  is still flat, not a bottleneck. **Stays PARKED.**
- **B12 (schema unification) — reproduced, decision unchanged.** Confirmed `BindAndValidate[T]`
  (`kernel/httpx/decode.go:52-55`) and `moduleContext.OpenAPI` (`app/context.go:324`) as the two hand-maintained
  sides; `internal/cli/openapi_cmd.go` still only merges fragments (`gatherFragments`/`mergeFragment`), no
  generator exists. Recounted wowsociety directly: `internal/modules/{identity,policy}/openapi.json` = 200 +
  234 = **434 lines** across **2 fragments**; **5** files carry `validate:"…"` tags — exact match to the prior
  doc. wowsociety has exactly **2** modules with request bodies (identity, policy), far below the ≈5–6-module
  reopen trigger; no traced OpenAPI-vs-handler drift defect found. `kernel/rules.RuleValueSchema`
  (`kernel/rules/schema.go:60-77`) confirmed as a separate, out-of-scope grammar (rule config values, not API
  DTOs). No reopen trigger met. **Stays PARKED** (generator remains overbuild for 2 modules; the
  drift-detection contract test remains the correct first increment only if/when this reopens).
- **B13 (hot-reloadable DB overlays) — reproduced, decision unchanged, with one clarifying finding.** i18n
  freeze-at-boot is real and enforced: `Catalog.frozen`/`Freeze()`/`Frozen()` (`kernel/i18n/catalog.go:64,98,106`),
  `boot.i18n.Freeze()` called in `app/boot.go:268` after layer merge + ownership validation. `KindDBOverlay` is
  a reserved-but-refused source kind (`kernel/i18n/loader.go:28-33`, `kernel/i18n/config.go:70`: "db_overlay
  source is not supported yet ... disable it") — the extension point exists, nothing wires it. Grepped
  `docs/` for a concrete operational requirement to mutate i18n/rules at runtime without redeploy: found none
  — only forward-looking design notes (`docs/implementation/wowsociety-framework-gap-design-review.md`,
  `docs/implementation/framework-competitive-architecture-benchmark.md`,
  `docs/user-guide/validation-errors.md:238`) that describe the freeze invariant and defer hot-reload, not a
  demonstrated need. **Clarifying finding (new, not in the prior doc):** rule *values* are already
  runtime-mutable without redeploy today — `Resolver.Resolve` (`kernel/rules/resolver.go:66,125-135`) reads
  `rule_versions` live from the DB per request; only the rule *definition/schema* registry is boot-synced
  (`kernel/rules/sync.go` `SyncDefinitions`). So the practical goal B13 names for rules — edit a value without
  a redeploy — is already met by the existing design; the gap, if any, is i18n-only, and no need has been
  shown there either. **Stays PARKED.**
- **W1 spot-check (out of scope for B11-13, noted for completeness):** wowsociety's Go-map i18n catalogs are
  gone, `locales/{en,mr}/product.yaml` + `locales/mr/kernel.yaml` exist (mr kernel.* coverage present), no
  `Catalog.Add`/manual `Register` remnants found. One acceptance clause is NOT met: `wowapi i18n validate` is
  not wired into wowsociety's CI (no reference in `.github/workflows/` or `Makefile`). Not fixed here — outside
  this goal's B11/B12/B13 scope; flagged for a future small W1 follow-up.
- **Decision:** no code or backlog-status changes. All three P2 items remain PARKED with their existing reopen
  triggers unchanged; this entry is the re-verification record so a future pass doesn't have to redo the
  reproduction work.
- **Affected:** docs/implementation/decisions.md (this entry), docs/implementation/framework-backlog-p2-decisions.md
  (re-verification appendix), docs/working/review-learning-register.md (process-learning entries). No
  kernel/product/test files changed.

## D-0089 — Pre-v1.0.0 close-out: pinned lint gate + full-tree enforcement + B-7 reference-stack smoke
- **Context:** the last three tracked pre-tag items, finished to enterprise standard: pin golangci-lint,
  promote full-tree `make lint` to the enforced gate, and close B-7 (reference-stack header smoke in CI).
- **Decisions:**
  - **Pinned golangci-lint** — a single-source-of-truth `GOLANGCI_VERSION` (`v2.11.4`) in the Makefile and
    `.github/workflows/ci.yml` (was `@latest` in both). The `tools` target reinstalls only if the installed
    version differs. Pinning is the prerequisite for enforcing full-tree lint: a new upstream release can no
    longer fail CI until the version is bumped deliberately.
  - **Full-tree `make lint` is now the enforced CI gate** — the `unit` job runs `make lint` (whole tree)
    instead of `make lint-new`. Safe now that B-1 is closed (D-0087) and the binary is pinned (deterministic 0).
    `make lint-new` stays as the fast local pre-push pre-check; docs (quality-gates.md) updated to match.
  - **B-7 (CA-6) reference-stack header smoke** — `make smoke-reference` +
    `scripts/smoke_reference_stack.sh` + `deployments/reference/smoke-compose.yaml` + a CI job `reference-smoke`.
    It scaffolds a product with `wowapi init`, builds its **linux static** api/migrate binaries
    (`CGO_ENABLED=0`), stands up postgres + a one-shot migrate + the api + the reference nginx (TLS, self-signed
    cert) via compose, then runs the existing `deployments/reference/smoke.sh` against `https://localhost` —
    asserting `X-Content-Type-Options`, `X-Frame-Options`, `Content-Security-Policy`, `Referrer-Policy`, and
    `Strict-Transport-Security` are delivered THROUGH the proxy. This exercises the proxy/TLS *wiring* — TLS
    termination + the app's headers forwarded unstripped — which the in-process `kernel/httpx/edge_test.go`
    doesn't cover (it tests `SecureHeaders` in isolation, incl. HSTS). The api/migrate connect as the postgres
    superuser + `SET ROLE app_rt`/`app_platform` (the local-dev pattern; `WithConnRLSGuard` still passes on the
    demoted effective role). **Verified end-to-end locally with Docker** before wiring CI.
- **Independent-gate fixes (before finalizing):** (1) the app already emits HSTS by default, so nginx's
  `add_header` produced a *duplicate* HSTS and the smoke's HSTS check couldn't attribute it to the edge — added
  `proxy_hide_header Strict-Transport-Security` so nginx owns HSTS authoritatively (one header; the app still
  sets its own for no-proxy/other-proxy deployments), and corrected the earlier "edge HSTS the in-process test
  can't cover" overclaim. (2) The nginx service had no healthcheck, so `compose up --wait` could release before
  :443 was listening — the script now polls the TLS endpoint before smoking (no CI flake). (3) Pinned actionlint
  (`ACTIONLINT_VERSION`) too, for the same determinism reason as golangci-lint.
- **Second cleanup pass (a follow-up review):** the actionlint pin was applied in CI but not the Makefile
  `actionlint` target (the same sibling-completeness gap) — pinned it there too (`ACTIONLINT_VERSION`, matching
  CI). `docs/working/quality-gates.md` still said the CI unit job runs `lint-new` and had a "why lint-new not
  full lint" section — updated to reflect full `make lint` is now the enforced gate (lint-new is the fast local
  pre-check). The other `@latest` installs are left deliberately and now documented as such: `govulncheck` must
  track the newest checks (a security scanner), and the authoritative goreleaser release is pinned in
  `release.yml` (the Makefile targets are local convenience).
- **Result:** every tracked backlog item (B-1…B-9) is now closed or rescoped; nothing outstanding before a
  `v1.0.0` tag.
- **Affected:** Makefile (GOLANGCI_VERSION, tools, smoke-reference), .github/workflows/ci.yml (golangci +
  actionlint pins, full lint, reference-smoke job), scripts/smoke_reference_stack.sh,
  deployments/reference/{smoke-compose.yaml, nginx.conf}, docs/{GOALS-TRACKER.md, working/quality-gates.md,
  operations/deployment-checklist.md}.

## D-0088 — Pre-v1.0.0 review: UTF-8-safe error truncation + tracker/doc accuracy
- **Context:** a review after D-0087 confirmed the B-1 parameter removals broke no behavior, but found a real
  (pre-existing) bug and honesty drift — including one overclaim D-0087 itself introduced.
- **Fixed:**
  - **Bug (Medium) — byte truncation could split a UTF-8 rune.** `kernel/jobs/runner.go` `truncate(s, n)` and
    `kernel/webhook/service.go` `truncate(s)` both did `s[:n]`, so an error message with a multibyte character at
    the cut could be stored as invalid UTF-8 — which a Postgres `text` update rejects. Both now back the cut up to
    a rune boundary (`utf8.RuneStart`). Added revert-sensitive tests: a `世` straddling the cut must yield valid
    UTF-8 with the straddling rune dropped (asserts `utf8.ValidString` AND exact content, so byte-slicing fails
    the test). New `kernel/jobs/internals_test.go` (`package jobs`) for the unexported jobs helper.
  - **Honesty — D-0087's tracker line overclaimed.** GOALS-TRACKER said "none of the tracked backlog items
    remain," but **B-7** (CA-6 reference-stack app-smoke in CI) is genuinely open. Corrected: B-7 is the one
    remaining item.
  - **Stale — B-9 was still marked active** though jobs/notify trace propagation shipped and is tested
    (`jobs.WithTracer` + `kernel/jobs/trace_test.go`; `notify.WithTracer` + `kernel/notify/trace_test.go`).
    Closed the B-9 row with that evidence.
  - **Stale doc wording:** `quality-gates.md` still called full `make lint` an advisory backlog (it's now clean);
    the enterprise-CI/CD design spec still said "implementation pending" (the workflows ship). Both corrected.
    (The hardening-P1 evidence bundle's now-stale trace note is left as a point-in-time append-only record.)
- **Affected:** kernel/jobs/{runner.go,internals_test.go}, kernel/webhook/{service.go,internals_test.go},
  docs/GOALS-TRACKER.md, docs/working/quality-gates.md,
  docs/superpowers/specs/2026-07-05-enterprise-cicd-design.md.

## D-0087 — B-1 lint backlog closed: `make lint` = 0 with no behavior change
- **Context:** the advisory `golangci-lint` backlog (B-1) had ~160 issues that `make lint-new` gated for new
  code but never blocked (full `make lint` was advisory). Goal: close it to 0 without introducing any bug.
- **Shape:** 150 of 160 were in `internal/cli`, all `fmt.Fprint*` (149) + one `tw.Flush` — best-effort writes
  to stdout/stderr `io.Writer`s. The other 10 were scattered (errcheck `f.Close`/Format writes, 3 `unparam`,
  2 dead test symbols, 1 `unconvert`).
- **Decision (user chose the resolution for the 150 CLI writes):**
  - **CLI terminal writes → one path-scoped `.golangci.yml` exclusion** (`path: internal/cli/` +
    `source: fmt\.Fprint`). A failed write to os.Stdout/os.Stderr has no recovery, so ignoring it is correct —
    the canonical errcheck-exempt case (mirrors the stdlib's own `fmt.Print*` exclusion). Scoped so genuine
    errcheck issues in the CLI (pool/file/exec errors) are still caught. This is the burn-down plan's sanctioned
    "scoped exclusion for a proven-safe stdlib call" — chosen over 150 `_ =` edits precisely to minimise churn
    and thus bug risk. The single `tw.Flush` got an explicit `_ =` with a reason.
  - **10 scattered → real code fixes, each behaviour-preserving:** read-only `defer f.Close()` →
    `defer func() { _ = f.Close() }()`; `Secret.Format` writes (a `fmt.Formatter` can't return an error) →
    `_, _ =`; `unparam` removed `webhook.truncate`'s always-500 `max` (const inside; 5 callers updated),
    `workflow.createTask`'s unused `def` (1 caller), `testkit.newPoolDB`'s always-nil `opts` (no caller change);
    `unused` removed a dead `type want` + `func run` in tests; `unconvert` dropped a redundant `int64()`.
- **Verification (the "no new bugs" bar):** `make lint` 0; `go build`/`go vet` clean; the full real-DB suite
  green incl. every touched package (webhook/workflow/testkit/model/config/benchbudget/buildinfo); coverage
  91.6% ≥ 90%.
- **Durability + follow-up:** `make lint-new` keeps changed code clean, so the tree stays at 0 for normal work.
  Promoting full-tree `make lint` to the enforced CI gate (burn-down step 3) is deferred as a recommended
  follow-up — CI installs golangci-lint `@latest`, so a full-tree enforced gate should be paired with a pinned
  version first to avoid breaking on a future upstream check (that pairing would itself be a gate change, out
  of scope for "close the backlog without new errors").
- **Affected:** .golangci.yml, internal/cli/config_cmd.go, internal/buildinfo/buildinfo.go,
  internal/tools/benchbudget/main.go, kernel/config/secret.go, kernel/webhook/{service,internals_test}.go,
  kernel/workflow/runtime.go, testkit/{db.go,consumer_test.go}, kernel/model/model_test.go, Makefile,
  docs/working/lint-backlog.md, docs/GOALS-TRACKER.md.

## D-0086 — Doc-drift sweep: user-facing docs re-grounded in the hardened runtime contract
- **Context:** a broad review found **no** Critical/High code defects (the scaffold/runtime hardening holds) but
  real Medium/Low doc drift: the D-0083/D-0084 runtime changes (platform_dsn now required; `migrate down` is a
  guarded full reset, not a stepwise rollback) never propagated into the framework-repo user docs, and the local
  `make ci` gate was described as running full lint when it runs vet + boundary lint only.
- **Fixed (proof-driven — the reviewer reproduced each against a fresh generated product):**
  - **platform_dsn required, not optional:** README quick-start, `getting-started.md`, and `configuration.md`
    (prose + config table) omitted `PLATFORM_URL` / called `platform_dsn` optional, yet api/worker fail closed
    without it. Added it to every getting-started env block and a `db.platform_dsn is required` troubleshooting row.
  - **migrate down semantics:** `database-migrations.md` said "roll back the last migration"; corrected to
    "FULL reset to version 0; refuses outside local/dev" (matches the generated runner + Makefile, D-0083 F2).
  - **`make ci` scope:** README (×2), `cli-reference.md`, `testing.md`, `best-practices.md`,
    `quality-gate-checklist.md`, and the Makefile `##` help all said `make ci` runs "lint"/"vet+lint"; it runs
    vet + boundary lint, with golangci-lint gated separately via `make lint-new` / hosted CI. Reworded all.
  - **Honesty/status:** GOALS-TRACKER CI line updated from "pending push" to "green on `a1ee245`"; local
    (git-ignored) `Goal 2.md` marker flipped from ACTIVE to ACCOMPLISHED to match `progress.md`.
- **Sibling completeness:** rather than fix only the reviewer's cited line numbers, grepped the whole docs tree
  for each drift pattern and fixed every instance (several the review did not enumerate); confirmed the actual
  product-dev harness (`deployments/product-dev.yaml`) already sets `PLATFORM_URL` (only the historical design
  spec omits it — left as a point-in-time artifact), and left append-only evidence logs untouched.
- **Not changed:** the ~154-item advisory `errcheck` backlog (B-1) stays open and tracked (per the F5 decision
  in D-0083); no code changed in this entry.
- **Affected:** README.md, docs/user-guide/{getting-started,configuration,database-migrations,testing,cli-reference,
  troubleshooting-faq}.md, docs/working/{best-practices,quality-gate-checklist}.md, docs/GOALS-TRACKER.md, Makefile.

## D-0085 — Third open-ended sweep (over the D-0084 commit itself): the sweep it claimed to have done
- **Context:** the user asked for one more unscoped sweep over the D-0084 commit `dd5085f` before trusting it.
  Two unscoped reviewers confirmed the four D-0084 fixes were real and revert-sensitive, but found that
  D-0084's own "swept every platform pool / covers both pools" story was itself incomplete — the same
  sibling-completeness failure, one level up.
- **Findings fixed:**
  - **C-1 (was the sharp one) — the api's `k.Platform` boot leg was dead code.** The generated api built its
    platform pool *after* `Boot` and never wired it into the kernel, so `app.Boot`'s new `k.Platform` check
    never ran for the api (only the worker wired it). The api's pool was still guarded by `WithConnRLSGuard`,
    so no live hole — but the framework backstop the M3 design exists to provide was absent, and D-0082's
    "both are protected" was misleading. Fix: build the api platform pool BEFORE `kernel.New` and pass
    `Platform: platformPool`, mirroring the worker — so the boot check genuinely covers both processes.
  - **F1 — `internal/cli/dlq_cmd.go` app_platform pool lacked `WithConnRLSGuard`.** The one platform-role CLI
    pool the D-0084 sweep missed (audit/apikey CLIs already had it). Added it.
  - **F2 — the 429 e2e proof was narrower than the claim.** It asserted only `X-Content-Type-Options`, not the
    CORS header the commit also claimed. Strengthened it: the manifest now sets an allowed origin and the test
    sends `Origin` and asserts the 429 carries **both** `X-Content-Type-Options` and `Access-Control-Allow-Origin`.
  - **I-1 — documented a real trade-off the reorder introduced.** With CORS outer to RateLimit, allowed-origin
    OPTIONS preflights are answered before the limiter, so they are not rate-limited/metered. Deliberate
    (rate-limiting preflights breaks browsers), now documented in the api template as an edge/WAF concern.
  - **F3 (honesty) — corrected D-0082's "both are protected" to describe the actual mechanism** (now accurate
    after C-1: both processes wire `Platform` before Boot).
- **Process note:** this is the pattern compounding — each unscoped sweep found the *previous* fix's missed
  sibling. The durable rule (now in the review-learnings log): when fixing one instance, grep every sibling of
  the same shape and fix or explicitly clear each, and boot the delivered artifact rather than the boot path.
- **Fourth sweep (over this D-0085 commit) — addendum:** it confirmed the pool sibling-sweep is finally
  complete (no unguarded serving pool remains; no accidental platform work leaks into the api process; every
  prose claim holds) and there is **no live hole**. It found one real gap: the C-1 api-wiring and F1 dlq-guard
  fixes had **no revert-sensitive test** — the e2e's superuser DSN is demoted by `SET ROLE app_platform`, so
  `AssertRLSEnforced`/`WithConnRLSGuard` pass whether or not the wiring is present (reverting the fix left CI
  green). A behavioral delivered-artifact test can't distinguish wired-from-unwired for exactly that reason, so
  added **structural** revert-sensitive guards: `TestInitWiresGuardedPlatformPool` (asserts the generated
  api/worker wire `Deps.Platform` *and* apply `WithConnRLSGuard`; **proven to FAIL** when `Platform` is
  un-wired) and `TestCLIPoolsApplyRLSGuard` (locks the guard on the dlq/audit/apikey CLI pools). The behavioral
  proof of the boot-check mechanism itself remains `app.TestIntegrationBootFailsOnRLSBypassPlatformPool`. The
  adversarial half of the same sweep found one more real gap: the `kernel.Kernel.Platform` / `kernel.Deps.Platform`
  doc comments still said "nil in api-only processes / required only for worker/migrate" — the API-contract
  surface a *custom* (non-scaffold) api main reads, which would tell that engineer to skip `Platform` and
  silently bypass the M3 check. Corrected both to state that the api and worker both wire it and that leaving it
  nil skips the backstop (nil only for migrate). This is the deliberate stopping point (per the agreed stopping
  rule): findings converged to a no-live-hole test-rigor + doc-contract gap, now closed.
- **Affected:** internal/cli/templates/init/cmd_api_main.go.tmpl, internal/cli/dlq_cmd.go, internal/e2e/e2e_test.go,
  internal/cli/platform_pool_wiring_test.go, kernel/kernel.go (Platform doc contract),
  docs/implementation/decisions.md (D-0082 correction).

## D-0084 — Open-ended-review follow-ups: what the confirm-scoped gates missed
- **Context:** after D-0082/D-0083 were declared done (each behind a review gate *scoped to confirm the fix I
  had just written*), two **open-ended** reviewers — told to find problems, not to confirm anything, with the
  target unchosen by me — found four more real issues in already-"done" code. That is the point of the entry:
  a confirm-shaped gate inherits the author's frame; only an unscoped adversarial pass reaches the sibling paths
  and the delivered-artifact behaviour the author never thought to check.
- **Findings fixed:**
  - **HIGH — deploy render omits `WOWAPI__DB__PLATFORM_DSN`.** api/worker fail closed without `db.platform_dsn`
    (CF-1), but `deploy render` (both `compose` and `env`) emitted only DSN + MIGRATE_DSN — so the rendered
    manifest crash-loops. Exact sibling of D-0083 F4 (which added MIGRATE_DSN and stopped there). Added the
    platform DSN to both formats; the deploy unit test now asserts all three, and the e2e now boots the api
    **from the rendered manifest** (mapped to real DSNs) so a missing required DSN fails the acceptance test.
  - **HIGH — generated api middleware order.** `SecureHeaders`/`CORS` sat *inner* to `RateLimit`, which returns
    429 without calling `next`; so 429 (and 413/timeout) responses shipped with no security or CORS headers — a
    browser cross-origin client that is rate-limited couldn't read the 429 or `Retry-After`. Moved
    `SecureHeaders`/`CORS` outermost (after Recover), keeping RateLimit/BodyLimit/Timeout inner.
  - **HIGH (defense-in-depth) — M3 boot check ignored the platform pool.** `AssertRLSEnforced` ran only on
    `k.Pool`, not `k.Platform`, which does all cross-tenant kernel work over FORCE-RLS tables. Extended the boot
    check to `k.Platform` (+ integration test with a superuser platform pool), and added `WithConnRLSGuard()` to
    the generated api/worker platform pools so the api's apikey platform pool (never wired into the kernel) is
    guarded at connect time too.
  - **Honesty — README onboarding + overstated claims.** The scaffold README's getting-started commands all
    failed out of the box (no `APP_ENV`/DSN exports against the fail-closed runtime) — rewrote it. Corrected the
    D-0082 "backstops a forgotten guard" overstatement (the check closes the over-privileged-role hole only; a
    forgotten guard is separately fail-closed by `app_tenant_id()` raising). Made the GOALS-TRACKER CI claim
    honest that commits after `329cc0e` are pending push + hosted CI.
- **Process change (durable):** independent review moves to the front and stays unscoped — run a reviewer whose
  job is "find what's wrong here," with the target not chosen by me, *before* declaring done, not as a closing
  stamp. Assert mechanisms not outcomes; boot the delivered artifact, not a hand-assembled env.
- **Affected:** internal/cli/deploy_cmd.go, internal/cli/cmds_test.go, internal/cli/templates/init/{cmd_api_main,
  cmd_worker_main,README.md}.go.tmpl, app/boot.go, app/rls_boot_guard_test.go, internal/e2e/e2e_test.go,
  docs/{GOALS-TRACKER.md, implementation/decisions.md (D-0082 correction)}.

## D-0083 — Generated-scaffold config/migrate/deploy correctness (consumer-facing review)
- **Context:** a consumer-facing review of the *generated product* (not the kernel) found six real
  scaffold/config/deployment/release-honesty gaps. The kernel, migrations, testkit, CI, and container gate
  were rated strong; these were correctness gaps in what `wowapi init` emits and in a couple of release claims.
- **Decisions (all fixed):**
  - **F1 (High):** the scaffolded `tools/configcheck` ignored `--env` and loaded only `APP_ENV`, so
    `wowapi config validate --env prod` validated the wrong environment (the prod CI gate was inert). Rewrote
    the checker to parse `--dir/--base/--env/--env-prefix`, select the `<env>.yaml` overlay (required when
    `--env` is given), wire the env secret provider, and **assert the composed environment equals `--env`**
    (fail-closed) — mirroring the framework-side `config_cmd.go` path. `--env` wins over `APP_ENV`.
  - **F2 (High):** the generated `cmd/migrate` was fail-open (config-load failure → compiled defaults) and
    ignored `up`/`down` (always migrated up), so `make migrate-down` was a no-op that looked like a rollback.
    Now: **fail-closed** config load (any error aborts); `up`/`down` subcommands; `down` runs
    `database.MigrateReset` (full reset to v0, modules-reverse-then-kernel) but **refuses unless environment
    is local/dev** — production schema change is forward-only expand-contract (blueprint 12). The framework
    exposes no stepwise down-one, so a guarded full reset is the honest primitive (chosen over adding a
    down-one API that would encourage the non-expand-contract path).
  - **F3 (Medium):** `config diff` was not delegated to the product checker, and the framework-side diff
    built loader options without a secret provider (so `secretref://env/...` configs errored). Wired
    `envprovider.New()` into `runConfigDiff`; `config diff` now delegates to `tools/configcheck` when present
    (added a `diff` mode there over the product's `appcfg.Config`), framework-only fallback otherwise.
  - **F4 (Medium):** `wowapi deploy render --format env` omitted `WOWAPI__DB__MIGRATE_DSN` (compose had it),
    so the migrate job couldn't run. Added the migrate DSN secret reference to the env template.
  - **F5 (Medium, release-honesty):** `make lint` is still red (~154 `errcheck`, backlog B-1). Qualified the
    "production-ready" claim in `docs/qa/07-closure-report.md` to scope it to the green `make ci` gate and
    name the tracked lint backlog as pending for a clean `v1.0.0` — rather than burn down 154 mechanical
    findings this round (user's call).
  - **F6 (Low):** refreshed `docs/GOALS-TRACKER.md` stale counts / CI SHA and the outdated "outstanding" line
    (B-2…B-5 are closed; only B-1 remains).
- **Tradeoffs:** F1/F2 convert generated-scaffold footguns into fail-closed defaults, consistent with the
  H1/H2/M2/M3/CF-1 "safe-by-default" line. `migrate down` deliberately offers only a guarded full reset, not
  stepwise rollback, to stay aligned with forward-only expand-contract.
- **Affected:** internal/cli/templates/init/{tools_configcheck_main.go.tmpl, cmd_migrate_main.go.tmpl,
  Makefile.tmpl}, internal/cli/{deploy_cmd.go, config_delegate.go, config_cmd.go}, internal/cli/cmds_test.go,
  internal/e2e/e2e_test.go (regression guards), docs/qa/07-closure-report.md, docs/GOALS-TRACKER.md.

## D-0082 — M3: boot-time RLS-enforcement check is safe-by-default (fail closed on an RLS-bypassing runtime pool)
- **Context:** the follow-up flagged in D-0081 ("M3 guard-by-default"). Postgres `FORCE ROW LEVEL SECURITY`
  does NOT apply to a superuser or a `BYPASSRLS` role. So a product that wires a data-serving pool over an
  over-privileged DSN would run tenant queries unfiltered, cross-tenant, with no signal. This was the last
  framework lever that stayed "safe only if the operator remembers."
- **Scope (precise, corrected per D-0084):** the check closes exactly the *over-privileged-role* hole — a
  superuser/BYPASSRLS effective role. It does NOT detect a *forgotten* per-connection/per-tx guard on an
  otherwise non-privileged role; that case is already fail-closed-and-loud elsewhere, because `app_tenant_id()`
  reads `current_setting('app.tenant_id')` with no `missing_ok`, so a tenant query with no bound tenant *raises*
  rather than leaking. Earlier wording here ("backstops a product that forgets the guards") conflated the two;
  the accurate claim is narrower.
- **Options:** (a) leave it on the product integration checklist (status quo — a plausible misconfig is a silent
  cross-tenant leak); (b) make the per-tx `WithRLSGuard` the default (per-query cost, and still bypassable by a
  pool that never enters the Manager); (c) assert enforcement once at boot, on the real pool, and fail closed.
  Chose (c): one check at the single chokepoint every deployment passes through, zero per-query cost.
- **Decision:**
  - `database.AssertRLSEnforced(ctx, pool)` probes a real pooled connection with
    `SELECT current_setting('is_superuser') = 'off' AND NOT rolbypassrls FROM pg_roles WHERE rolname = current_user`.
    Because it runs on a live connection, `current_user`/`is_superuser` reflect the EFFECTIVE role — including a
    `SET ROLE app_rt` applied in `AfterConnect` (`WithSetRole`) — not just the DSN login. The error names the fix.
  - `app.Boot` runs it after `validateAndOrder` for **both** `k.Pool` (runtime) and `k.Platform` (the
    cross-tenant kernel pool) when non-nil — **safe-by-default, no opt-in**. Both the scaffolded api and worker
    build their platform pool BEFORE Boot and wire it as `kernel.Deps.Platform`, so the boot check covers every
    serving pool in both processes; each pool's per-connection `WithConnRLSGuard` is defense-in-depth on top.
    (The platform-pool boot leg + the api/worker `WithConnRLSGuard` were added in D-0084; wiring the api's
    platform pool *before* Boot — so the `k.Platform` leg is actually reached in the api, not dead code — was
    added in D-0085 after a third open-ended review found the api built its platform pool after Boot.)
  - Escape hatch `app.SkipRLSEnforcementCheck()` (a `BootOption`) is applied to **exactly one** caller — the
    scaffolded migrate command — which boots solely to COLLECT module migration sets, connects with privileged DDL
    creds by design, and never serves tenant traffic. Minimal, framework-controlled, backward-compatible (existing
    callers pass no options).
- **Tradeoffs:** converts a delegated-responsibility item into "safe unless deliberately disabled," mirroring the
  CF-1 fail-closed-DSN philosophy (D-0078). The only opt-out is the migrate path, recorded here so the boundary is
  deliberate, not an untracked hole. Cost is one round-trip at boot; none at request time.
- **Affected:** kernel/database/database.go (`AssertRLSEnforced`), app/boot.go (`BootOption`,
  `SkipRLSEnforcementCheck`, the default check), internal/cli/templates/init/cmd_migrate_main.go.tmpl (opt-out),
  app/rls_boot_guard_test.go (integration test: superuser pool fails, app_rt pool boots clean).

## D-0081 — Third-review tenant-isolation footgun hardening (H1/H2/H3/M2/F-2)
- **Context:** an adversarial tenant-isolation review (live-Postgres probes) rated the framework PASS — no
  critical/high live cross-tenant leak — but flagged surfaces the framework delegates to the product where a
  plausible integration bug becomes a leak/DoS. Decision: harden those in the framework so a product can't get
  them wrong (convert "safe if careful" into "safe by default").
- **Decisions:**
  - **H1** rate-limit key: `KeyByActor` collapsed every nil-capacity caller (all API-key/system/webhook actors,
    all tenants) into one bucket `actor:00000000-…`. New key `t:<tenant>|<principal>` (capacity → api-key/system
    → subject → per-IP fallback), never buckets on `uuid.Nil`; the authz gate now binds the full actor.
  - **H2** webhook: `DispatchOutbound` now binds the tenant from `ev.TenantID` and fails closed on a mismatched
    passed tenant — closing the "A's payload delivered to B, signed with B's secret" path.
  - **H3** safe async: shipped leader-safe per-tenant pollers (`kernel.notify.send_pending`, `kernel.webhook.retry`)
    in `app/maintenance.go`, enumerating active tenants and binding the tenant from the enumeration — so products
    don't hand-roll the loop where H2 happens.
  - **M2** jobs RLS: migration `00028` enables RLS+FORCE on `jobs_queue`/`job_runs` with a strict tenant policy
    `WITH CHECK (tenant_id = app_tenant_id())` + a permissive `app_platform` policy, **exactly** mirroring
    `events_outbox` (write-integrity defense-in-depth; app_rt stays read-denied; global/NULL jobs are written by
    `EnqueueGlobal` as app_platform under the permissive policy, so app_rt can't enqueue NULL/global rows).
    Plus the F-3 grant-only test.
  - **F-2** regression guard: a table-driven `AssertRLSIsolation` census over 34 strict tenant tables + a
    self-maintaining completeness check (44 live = 34 probed + 10 excluded) that fails when a new tenant table
    isn't registered.
  - Also fixed a shutdown race the extra schedules exposed: `Scheduler.Run` treats a canceled-context `Ensure`
    as clean shutdown, not an error.
- **Not done (product responsibility, per the review's integration checklist):** M1 (JWT tenant claim trust),
  M3–M8, documented in the local tenant-isolation report; M3 (guard-by-default) was folded into the framework — see D-0082.
- **Affected:** kernel/httpx/{ratelimit,context,authz_gate}.go, kernel/webhook/service.go, app/{maintenance,worker}.go,
  migrations/00028_jobs_rls.sql, kernel/jobs/{scheduler,rls_test}.go, testkit/{asserts,rls_isolation_all_test}.go,
  kernel/config/config.go (doc-string).

## D-0079 — CA-2(b): authz-cache invalidation wired to the authorization-spine write
- **Context:** `authz.CachingStore` was wireable (`kernel.Deps.AuthzCacheTTL` → `Kernel.AuthzCache`) but
  its `Invalidate`/`InvalidateTenant` had ZERO non-test callers, so a role/permission change could be
  served stale for up to the TTL, and the original "invalidation hook on seed/spine writes" acceptance
  clause had been silently dropped. The review flagged this as an incompletely-closed hardening item.
- **What the cache actually holds (scoping the fix):** `CachingStore` caches ONLY `ActiveAssignments`,
  which pre-joins `role_permissions`. So exactly two write classes can be served stale: (1) an
  `actor_assignment` grant/revoke, and (2) a role's permission-set change (`roles` + `role_permissions`,
  i.e. a seed sync). The other Store reads (`OrgAncestors`/`OrgSubtree`/`Policies`/`ResourceOrg`) and the
  ReBAC `Checker.Has` are pass-through / checked directly each `Evaluate`, so **ABAC policy activation and
  granted_via relationship-edge changes are never cached and cannot go stale** — they need no invalidation
  (correct by construction, now documented on the `Kernel.AuthzCache` field).
- **Decision:**
  - **Seed / authorization-spine sync (framework-owned write) — REAL wiring.** Added
    `CachingStore.InvalidateAll()` (drops the whole cache) and an optional variadic `SpineInvalidator`
    hook on `seeds.Sync(ctx, db, bundle, invalidators...)`. After the spine writes commit, `Sync` calls
    `InvalidateAll` on any passed cache. Platform roles are GLOBAL (tenant_id NULL, assignable in any
    tenant), so the correct invalidation is broad (whole cache), not one tenant. Backward compatible:
    existing callers pass no invalidator; caching-off (the default) passes nothing → exact prior behavior.
    Products that run `seeds.Sync` with caching on pass `Kernel.AuthzCache`.
  - **actor_assignment grant/revoke (product-owned write) — exposed handle, not framework-wireable.** The
    framework ships NO assignment-management write path (only testkit fixtures and product modules write
    `actor_assignments`), so there is nothing in-framework to wire. `Kernel.AuthzCache` is exposed with a
    documented contract: call `Invalidate(tenant, capacity)` (or `InvalidateTenant`) right after the
    grant/revoke commits. This is a handle, not a TTL-only rescope — the TTL remains only the cross-pod
    bound.
- **Not a TTL-only rescope:** the one reachable framework spine write IS wired; the only unwired path is a
  write the framework does not own, for which the invalidation handle is exposed and documented.
- **Tests (real, DB-backed, non-skipping):** `kernel/seeds/sync_invalidate_test.go` proves a re-sync that
  prunes a role grant is reflected immediately when the cache is passed to `Sync` (and stays bounded-stale
  within a 1h TTL when it is not — the control arm isolates invalidation, not clock, as the cause) plus a
  caching-off arm; `kernel/authz/caching_pg_test.go` proves through the real `Evaluate` that an
  `actor_assignment` revoke is immediate after `Invalidate` fires (and immediate with caching off).
  Revert-check: neutering `InvalidateAll` fails the spine test at the stale-allow assertion.
- **Affected:** `kernel/authz/caching.go` (`InvalidateAll`), `kernel/seeds/seeds.go` (`SpineInvalidator` +
  `Sync` hook), `kernel/kernel.go` (`AuthzCache` field doc), `kernel/seeds/sync_invalidate_test.go`,
  `kernel/authz/caching_pg_test.go`.

## D-0078 — CF-1: fail-closed platform DSN (runtime/platform separation)
- **Context:** an independent review (CF-1) found the generated api/worker platform pool fell back to the
  runtime DSN + `SET ROLE app_platform` when `db.platform_dsn` was unset. That fallback structurally requires
  `app_rt` to be a (cluster-global) member of `app_platform`. The product-dev harness enacted exactly that
  membership on the shared compose cluster, which poisoned every database on it (incl. testkit clones) and made
  the framework's own self-grant/legal-hold guards fail.
- **Options:** (a) warn only; (b) fail closed in prod, dev opt-in; (c) fail closed always.
- **Decision (c):** the api/worker templates now **require** a distinct `db.platform_dsn` (a dedicated
  `app_platform` login) and return a boot error when it is unset — the `platformDSN = dsn` fallback is removed.
  The scaffold's `configs/local.yaml` wires `platform_dsn: secretref://env/PLATFORM_URL` by default. The
  product-dev harness connects the platform pool directly as an `app_platform` LOGIN (via `PLATFORM_URL`) and no
  longer runs `GRANT app_platform TO app_rt`. A regression guard,
  `authz.TestIntegrationRuntimeRoleNotMemberOfPlatform`, fails red if `app_rt` is ever a member of
  `app_platform`, converting environment poisoning into a caught test.
- **Tradeoffs:** consumers must configure two runtime DSNs (runtime + platform) instead of one; accepted as the
  correct security posture (privilege separation is the whole point of the two roles).
- **Affected:** internal/cli/templates/init/{cmd_api_main.go.tmpl,cmd_worker_main.go.tmpl,configs_local.yaml.tmpl},
  scripts/product-dev.sh, deployments/product-dev.yaml, kernel/authz/escalation_test.go,
  docs/operations/product-dev-container.md.

## D-0077 — Post-hardening review: close six deferrals
- **Context:** a review of the hardening pass found six items whose deferrals left roadmap gaps
  incompletely closed. All were verified real and fixed (commits `d22ff7f`, `54abec1`).
- **Decisions:**
  - **F1 durable authz-denial audit.** Denials were only WARN-logged. Since `Evaluate` runs in a
    read-only tx, added `kernel.durableAudit` (default when a TxManager is present) that writes an
    `authz.denied` `audit_logs` row in its own tenant tx — persisting even if the request rolls back.
  - **F2 data-lifecycle engine.** Implemented the per-record-class disposition + DSR export/erasure that
    was "future orchestration": `retention.Registry` (Dispose/Export/Erase callbacks — no dynamic-table
    SQL) + `retention.Engine`; wired into the kernel (`RetentionClasses`/`Retention`), the module
    `Context`, and a per-tenant scheduled disposition sweep.
  - **F3 end-to-end OTel.** Added `adapters/tracing/otel` (real OTel adapter, configurable ratio
    sampler, `NewOTLP` OTLP/HTTP exporter), extended the `Tracer` port with `Inject`/`Extract` (W3C
    traceparent) with the HTTP middleware continuing an inbound trace, and added the tracing INFRA — a
    Jaeger all-in-one service in the compose stack (OTLP + UI) and a deployment-checklist section.
  - **F4 channel preferences.** `notification_channel_prefs` (migration 00022) + `notify.SetChannelPref`;
    `Send` skips opted-out channels and fails loudly if all are opted out. (Failure surfacing to caller
    = the existing `Deliveries` receipts API.)
  - **F5 config CLI delegation.** `wowapi config validate|print|schema|doctor` now exec the product-local
    `tools/configcheck` when present (framework-only fallback); added `config diff` + expanded the
    generated checker.
  - **F6 worker platform DSN.** Added `config db.platform_dsn` (dedicated app_platform login); the
    generated worker uses it, falling back to the runtime DSN + SET ROLE for local.
  - **Prior issue (workflow vote/min_approval/self_approval):** confirmed expected — fail-closed at
    definition validation (D-0054). No change.
- **Affected:** `kernel/kernel.go`, `kernel/retention/{engine}.go`, `kernel/observability/tracing.go`,
  `adapters/tracing/otel/`, `kernel/notify/service.go`, `internal/cli/config_*.go`, generated worker +
  configcheck templates, `kernel/config/config.go`, `module/module.go`, `app/{context,boot,maintenance}.go`,
  `deployments/compose.yaml`, migrations 00022, `docs/operations/deployment-checklist.md`.

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

## D-0035 — Phase 4: migration numbering maps blueprint 002–004 to on-disk 00004–00006
- **Context:** blueprint 03 §5 numbers the identity/resource/authz migrations 002/003/004, but
  on-disk 00003 was taken by idempotency (D-0031). goose numbers are per-source and only need to be
  monotonic.
- **Decision:** 00004_org_party_capacity.sql (blueprint 002: organizations, parties, persons,
  legal_entities, party_contacts, acting_capacities), 00005_resource_relationship.sql (003:
  resource_types, resources, relationship_types, relationships), 00006_authz.sql (004: permissions,
  roles, role_permissions, actor_assignments, policies, policy_conditions). All tenant-scoped tables
  get ENABLE+FORCE RLS + app_rt grants; global registries (resource_types, relationship_types,
  permissions) get app_platform grants (kernel-service access, per SEC-13/D-0026).
- **Affected:** migrations/00004–00006; docs/blueprint/03 §5 note.

## D-0036 — Phase 4: authz evaluator is deny-by-default with a Store port; registry validated at boot
- **Context:** 01 §3 specifies the layered Evaluate algorithm (RBAC → ReBAC → ABAC, deny-first) and
  a permission registry where an unknown permission is a boot error, not a runtime 403.
- **Decision:** `kernel/authz` defines Actor/Target/Decision/Evaluator + a `Store` port (loads
  active assignments, role permissions, relationship grants, policies for an actor/target) so the
  evaluator is pure and unit-testable with a fake store; the pg-backed store lands beside it. The
  permission registry is a validated set built from module route permissions + seeded permissions;
  Evaluate on an unregistered permission returns an error (surfaced at boot when routes register,
  not per request). Filter returns a structured `ListFilter` (org/resource id constraints) the
  store translates to SQL — never load-then-filter.
- **Affected:** kernel/authz, kernel/policy, kernel/relationship, kernel/resource.

## D-0037 — Phase 4: OIDC verifier with an injectable JWKS source + local test issuer
- **Context:** 01 §3 / auth middleware needs an OIDC token verifier, but tests must mint tokens the
  verifier accepts without an external IdP.
- **Decision:** `kernel/auth` verifies JWTs against a `KeySource` port (JWKS by key id); production
  wires a caching JWKS-over-HTTPS adapter, tests wire a local RSA signer (`testkit.IssueToken`).
  The verifier maps validated claims → `authz.Actor` (user id, tenant, capacity) after resolving
  the user's active capacity in the tenant. Break-glass/impersonation carry explicit ctx markers
  and are audited.
- **Affected:** kernel/auth, testkit/auth.go, adapters/oidc (JWKS adapter, later).

## D-0038 — Phase 4 review: closed verb set extended with `ingest` and `activate` (ARCH-41)
- **Context:** the 01 §3 closed action verb set is `create|read|list|update|deactivate|restore|approve|
  reject|assign|export|admin`, but the blueprint's own matrix uses `payments.callback.ingest`
  (webhook ingest) and break-glass needs an `activate` verb.
- **Decision:** extend the closed set with `ingest` (inbound webhook/event ingestion) and `activate`
  (break-glass / feature activation). The set stays closed and small; both have concrete blueprint
  usages. 01 §3's list is updated to match so code and blueprint agree.
- **Affected:** kernel/authz/registry.go, docs/blueprint/01 §3.

## D-0039 — Phase 4 review: evaluator runs in the caller's tenant tx (ARCH-36); caching + list-ReBAC deferred
- **Context:** the pg Store/Checker each opened their own `WithTenantRO` tx per method, so one
  Evaluate spanned ~5 separate transactions — a different MVCC snapshot from the request's business
  tx (a just-written resources mirror row would be invisible), N round-trips, and second-connection
  deadlock risk on the hot path.
- **Decision:** the Store/Checker/Evaluator methods take the caller's `database.TenantDB` and run
  their reads on it — one snapshot, one connection, consistent with the request's writes. The pure
  evaluator is unchanged; only the seam moves. Per-request memoization and the 30s assignment
  snapshot cache (01 §3) are deferred to Phase 5/6 with the live wiring (a TODO on the evaluator);
  ReBAC list visibility (`ListFilter` from relationship-derived resource ids) needs a
  `Store.RelationshipResourceIDs` seam and is completed in Phase 5 when list endpoints ship
  (ARCH-37) — until then `Filter` covers RBAC scopes only, documented in code.
- **Affected:** kernel/authz (store.go, evaluator.go, store_pg.go), kernel/relationship,
  kernel/resource/registrar; phase-plan rows 4/5.

## D-0040 — Phase 5: Context accessor scope (which of 06 §2 ships now)
- **Context:** 06 §2's full Context references kernel packages that arrive in later phases (rules→7,
  workflow→7, outbox/jobs→6, document→8, notify/webhook→9). D-0006 grows Context per phase.
- **Decision:** Phase 5 ships the accessors whose kernel capabilities exist:
  Routes/Permissions/Roles/ResourceTypes/RelationshipTypes, Migrations/Seeds/OpenAPI, Tx/Authz/
  Logger/Config/IDGen/Clock/Health, and Port/ProvidePort (inter-module ports checked at boot). The
  later-phase accessors (Rules/Workflows/Events/Jobs/Documents/Notify/Webhooks) are added with their
  packages. Interface widening pre-v0.1.0 is an accepted breaking change (D-0006).
- **Affected:** module/module.go, app/context.go.

## D-0041 — Phase 5: Kernel + App composition root; boot wires the evaluator and gates on registries
- **Context:** Phase 4 left the evaluator, permission registry, and PrincipalStore dangling (ARCH-39,
  ARCH-44). Phase 5 is where the app boot builds them.
- **Decision:** `kernel.Kernel` (New(ctx, cfg, deps) → owns pool, Tx, Authz evaluator, Log, Health,
  Audit sink) and `app.App` (Register/Validate/StartAPI/StartWorker/Shutdown). Boot order per 06 §2
  lifecycle: construct kernel → per-module Register (collect into registries) → Validate (whole-graph:
  dup permissions, routes without meta, unknown deps/cycles, unsatisfied ports, module-config decode,
  seed-schema, **permission registry Err()**) → SeedSync (idempotent catalog upsert) → Start. The
  evaluator is built from the composed permission registry + PgStore + policy engine + relationship
  checker + audit sink and injected into every module.Context.Authz(). Boot aborts on any Validate
  error — the permission registry gate is now enforced (closes the Phase 4 deferral).
- **Affected:** kernel/ (new package), app/app.go + run.go + context.go.

## D-0042 — Phase 5: seed loader is declarative YAML → idempotent catalog upsert
- **Context:** modules ship `seeds/*.yaml` declaring permissions, roles (+role_permissions),
  resource_types, relationship_types; SeedSync upserts them idempotently (never touches tenant data).
- **Decision:** `kernel/seeds` parses a typed seed bundle (strict YAML, unknown keys fail) and
  SeedSync upserts into the global catalogs as app_platform (the catalogs are app_platform-writable,
  per SEC-13/D-0026). Seed permission/role keys feed the boot permission registry. Idempotent:
  ON CONFLICT DO UPDATE; running twice is a no-op diff. Contract-tested (run twice).
- **Affected:** kernel/seeds, migrations grants (already app_platform), testkit contract suite.

## D-0043 — Phase 5: scratch-consumer test builds a real external module in a tmpdir
- **Context:** the headline exit criterion — an external product repo can import wowapi, define a
  module, and pass the contract suite without framework edits.
- **Decision:** a `test-consumer` flow (host+container) scaffolds a tiny product module in
  t.TempDir(), `go mod init` + `go mod edit -replace github.com/qatoolist/wowapi/v2 => <repo>`, writes
  a module using only public packages, and runs `testkit.RunModuleContract`. Proves the public API
  surface is sufficient and import-direction-clean from outside the repo.
- **Affected:** testkit/contract.go, a consumer test under testkit or internal, Makefile test-consumer.

## D-0044 — Phase 5 review: seed ownership covers role grants + granted_via; grants reconciled
- **Context:** the seed prefix-ownership check validated declared keys but NOT the role grant-list or
  `granted_via` — so a module could grant itself a foreign permission (SEC-32, reproduced) or wire
  its permission to another module's relationship (SEC-34). Sync was also insert-only, so removed
  grants never pruned (ARCH-47).
- **Decision:** `seeds.validate` prefix-checks every `RoleSeed.Permissions` entry and `GrantedVia`,
  and requires `granted_via` to name a relationship type the same bundle declares. `Sync`
  reconciles each role's grants (deletes grants not in the seed) so a demoted role sheds
  privileges across redeploys. Regression tests in seeds_test.go.
- **Affected:** kernel/seeds/seeds.go.

## D-0045 — Phase 5 review: seeds run as app_platform; hybrid-table RLS uses a forgiving tenant fn
- **Context:** the contract ran `seeds.Sync` as superuser, never testing the SEC-13 grant boundary
  (SEC-33). Running as app_platform hit the roles/policies RLS `WITH CHECK`, which calls the strict
  `app_tenant_id()` (raises when unset) — a platform connection has no tenant, so NULL-template
  writes aborted.
- **Decision:** add `app_tenant_id_or_null()` (missing_ok → NULL) and use it ONLY in the
  roles/policies policies (`tenant_id IS NULL OR tenant_id = app_tenant_id_or_null()`), so a
  platform/catalog connection can read/write NULL-tenant templates while a tenant connection still
  sees only its rows + templates. Pure tenant tables keep the strict raising `app_tenant_id()`
  (loud fail-closed + AssertRLSIsolation unchanged). testkit provisions an `app_platform` login +
  Platform pool; the contract syncs seeds under it (SEC-33) and asserts effect-idempotency via a
  catalog checksum (ARCH-49). app_rt is still SELECT-only on roles/policies, so this does not widen
  it.
- **Affected:** migrations/00001, 00006; testkit/db.go (Platform pool), testkit/contract.go.

## D-0046 — Phase 5 review: contract RLS check is diff-based, not name-prefix (ARCH-48)
- **Context:** the RLS assertion matched tables by `<module>_` prefix — evadable by naming — and a
  module with zero conforming tables passed silently.
- **Decision:** the contract snapshots public tables before/after the module migrate, and asserts
  ENABLE+FORCE RLS on every table the migration actually created (excluding goose bookkeeping);
  a module that ran migrations but produced no RLS-forced table fails.
- **Affected:** testkit/contract.go.

## D-0047 — Phase 6: Postgres-backed job runner behind the interfaces, not River
- **Context:** Goal 2 says "River OR the selected Postgres-backed job runner behind framework
  interfaces". River is a large dependency with its own migration set and API surface; the module
  portability contract only depends on `jobs.Registry`/`Runner`/`Worker`.
- **Decision:** implement a focused Postgres job queue (`kernel/jobs`) behind those interfaces:
  a `jobs_queue` table, `FOR UPDATE SKIP LOCKED` claim, bounded fixed worker pool per queue,
  exponential backoff + jitter retry, DLQ (status=discarded mirrored to `job_runs`). Interfaces
  match the blueprint so a future River swap is internal. Keeps the dependency surface small and
  the retry/DLQ semantics ours to test precisely.
- **Affected:** kernel/jobs, migration 00007.

## D-0048 — Phase 6: outbox relay reads cross-tenant as app_platform; dispatches per-tenant
- **Context:** `events_outbox` is tenant-scoped (RLS) so modules write/read only their tenant's
  events in the business tx. The relay must dispatch ALL tenants' pending events.
- **Decision:** a role-scoped RLS policy grants `app_platform` (the relay/kernel role) SELECT+UPDATE
  across all outbox rows; the relay claims a batch with `FOR UPDATE SKIP LOCKED` as app_platform,
  then for each event RE-ENTERS a tenant transaction bound to the event's tenant_id (SET LOCAL) to
  run handlers under normal tenant RLS + the inbox dedup. Ordering is per-aggregate
  (`occurred_at` per resource). This keeps app_rt strictly tenant-isolated while giving the kernel
  relay the cross-tenant read it needs — mirrors the app_platform posture from Phase 5.
- **Affected:** migration 00007 (events_outbox policies), kernel/outbox relay.

## D-0049 — Phase 6: TenantDB.Outbox()/Events() + module.Context Events()/Jobs()
- **Context:** 05 §2 TenantDB carries `Outbox()`; 06 §2 Context carries `Events()`/`Jobs()`.
- **Decision:** `database.TenantDB` grows `Outbox() outbox.Writer` (same-tx event write); the
  per-tx writer is attached by the TxManager. module.Context grows `Events() outbox.HandlerRegistry`
  (Subscribe) and `Jobs() jobs.Registry` (RegisterKind). The worker process (`app.RunWorker`) starts
  the relay + job pools and drains gracefully on shutdown.
- **Affected:** kernel/database (TenantDB), kernel/outbox, kernel/jobs, module/module.go,
  app/context.go, app worker start.

## D-0050 — Phase 6 review: per-aggregate ordering enforced; event DLQ; job timeout/drain separation
- **Context:** the review reproduced that per-aggregate ordering was NOT actually held (the
  blueprint's advisory lock was absent; a transient handler failure reordered events, ARCH-53),
  failed events retried forever with an ineffective cooldown (ARCH-54/55), and the job runner
  conflated the shutdown drain with the per-job timeout (ARCH-56/57).
- **Decision:**
  - Relay: the claim only picks the earliest still-undispatched event per (tenant, resource) — a
    later event never overtakes an earlier pending/failed one — plus a tx-scoped
    `pg_advisory_xact_lock` per aggregate so concurrent relays serialize. Per-aggregate ordering is
    now real (regression test under retry).
  - Event DLQ: `events_outbox` gains `failed_at`, `max_attempts`, `last_error` and a `'dead'` status;
    a poison event dead-letters after max_attempts; `RequeueFailed` keys its cooldown on `failed_at`.
  - Jobs: a per-job `jobTimeout` (default 2m) separate from the shutdown `drainTimeout`; outcomes are
    written with a fresh short-lived context; `stalledTimeout` is floored above jobTimeout+drain so a
    live job can't be reclaimed and run concurrently (ARCH-58); `StartWorker` enforces a HARD drain
    cap so a ctx-ignoring worker can't hang shutdown (ARCH-57).
  - Semantics documented: jobs are at-least-once with NO framework dedup (workers with external side
    effects must carry their own idempotency key); event handlers get exactly-once DB effect via the
    inbox (ARCH-59).
- **Affected:** migrations/00007, kernel/outbox/relay.go, kernel/jobs/{runner,jobs}.go, app/worker.go.

## D-0051 — Phase 7: migrations 00008 (rules) + 00009 (workflow); custom Postgres engines
- **Context:** blueprint 02 §1.1 recommends a small custom Postgres-backed workflow engine over
  Temporal/Camunda (approval/state-machine shaped, tenant-editable, shares the business tx/RLS/audit/
  outbox). Rules likewise are a Postgres-backed versioned config engine.
- **Decision:** `kernel/rules` (rule-point registry + version storage + resolution) and
  `kernel/workflow` (definition model + runtime + SLA sweeper) as custom engines. Migration 00008
  = rule_definitions (global) + rule_versions (tenant+platform hybrid, temporal, exclusion
  constraint one-active-per-scope); 00009 = workflow_definitions (global+tenant) + workflow_instances
  + workflow_tasks + workflow_task_assignees (tenant-scoped RLS). Both engines share the tenant tx
  and emit outbox events + audit in the same transaction as state changes.
- **Affected:** kernel/rules, kernel/workflow, migrations/00008–00009.

## D-0052 — Phase 7: rule resolution is org-ancestry → tenant → platform → code default, historical by `at`
- **Decision:** `rules.Resolver.Resolve(key, tenant, org?, at)` picks the first active version
  (effective_from <= at < effective_to) walking org ancestry upward, then tenant, then platform,
  then the code-registered default; the value is JSON-Schema validated (defense in depth) and
  returned with provenance. Versions are immutable (never mutated, only superseded), so any
  historical `at` resolves deterministically. Approval-gated points require an `active` version to
  have passed approval; a draft/pending version never resolves. Resolution runs on the caller's
  TenantDB (one snapshot).
- **Affected:** kernel/rules resolver + tests.

## D-0053 — Phase 7: workflow step-type set is closed; definitions validated at boot
- **Decision:** closed step types (approval|task|auto|gateway|vote|terminal); assignee kinds
  (actor|role-at-scope|relationship|resource_owner|resolver). Definitions are validated at
  registration (graph connectivity, no orphan steps, terminals reachable, unknown auto-actions
  fail boot). Instances pin their definition version (immutable per version). Every transition
  re-checks the actor (assignee + `workflow.task.decide`), mutates with optimistic locking, and
  writes audit + outbox in the same tenant tx. testkit `WorkflowSim` drives definitions over a real
  test DB.
- **Affected:** kernel/workflow, testkit/workflowsim.

## D-0067 — Hardening P1 (R5): notification delivery receipts query
- **Context:** the roadmap called notifications "fire-and-forget," but the audit found delivery status
  IS tracked in `notification_deliveries` — what was missing is a query API to read it per notification.
- **Decision:** `notify.Service.Deliveries(ctx, db, notificationID) []DeliveryReceipt` returns the
  per-channel receipts (status, attempts, provider message id, last error, timestamps), RLS-scoped to
  the caller's tenant. No schema change — it reads existing columns.
- **Tradeoffs:** closes "delivery status queryable per notification; provider receipts stored". Per-user
  channel preferences (opt-out) is deferred — it needs a preferences table + send-path enforcement.
- **Affected:** `kernel/notify/service.go` (+`notify_test.go`), evidence/hardening-P1.

## D-0076 — Hardening H5 (E4): snapshot / artifact pipeline
- **Context:** no immutable versioned-artifact primitive; a compliance product would hand-roll
  receipt/certificate snapshots (roadmap E4).
- **Decision:** `kernel/artifact` over an `artifacts` table (migration 00021). `Generate` turns
  product-rendered bytes into an immutable per-(tenant,kind) versioned row with sha256(content), a
  structured sidecar, content-type, template version + effective date; `Get`/`List`/`Verify` (re-hash to
  detect tamper). A `Templates` registry resolves the version effective at a date. Content is stored
  in-row (bounded compliance artifacts) so an artifact is atomic + self-verifying; append-only grants
  (app_rt no UPDATE/DELETE).
- **Layering:** the framework owns immutability/versioning/hashing/verify/template-resolution; the product
  supplies the rendered bytes (its own PDF/A renderer) — no document-format library in the kernel,
  mirroring the storage-port split.
- **Affected:** `kernel/artifact/{artifact,templates}.go` (+`_test.go`), `migrations/00021_artifacts.sql`,
  evidence/hardening-H5.

## D-0075 — Hardening P1 (O1): distributed-tracing seam
- **Context:** only request-id propagation; no tracing (roadmap O1: "behind the metrics/observability
  port; zero-cost when disabled").
- **Decision:** a `kernel/observability.Tracer`/`Span` port + `NoOpTracer` (sibling of `Metrics`) + a
  `Trace` HTTP middleware opening a server span per request (route/method/status/request-id). Wired into
  the generated api chain with `NoOpTracer` — zero-cost when disabled. The OpenTelemetry SDK binding is a
  thin adapter (`adapters/tracing/otel`), keeping the kernel otel-free, exactly as metrics keeps
  prometheus in an adapter.
- **Tradeoffs:** cross-process trace propagation (injecting/extracting traceparent through outbox events
  and job payloads for API→relay→worker) is the follow-up; the port + HTTP spans + nesting are in place.
- **Affected:** `kernel/observability/tracing.go` (+`_test.go`), generated `cmd/api/main.go.tmpl`,
  evidence/hardening-P1.

## D-0074 — Hardening P1 (R1): authz decision caching
- **Context:** every `Evaluate` hit the DB for the actor's assignments (roadmap R1).
- **Decision:** `authz.CachingStore`, an OPT-IN `Store` decorator caching `ActiveAssignments` per
  `(tenant, actor)` for a short TTL (default 1s). Unwrapped = current behavior (zero risk).
  `Invalidate`/`InvalidateTenant` give immediate same-pod effect on a role change; the TTL bounds
  cross-pod staleness. Other reads pass through (narrow invalidation surface).
- **Correctness:** the R1 "no stale-allow after revocation" requirement is met by explicit invalidation
  (immediate) — tested: revoke bounded-stale within TTL, then denied right after `Invalidate`.
- **Tradeoffs:** in-process per pod (Redis is a later adapter behind the same seam). Read-replica routing
  (R1's other half) is a deployment seam — point the Manager's `WithTenantRO` at a replica pool; the
  evaluator already reads in that read-only tx.
- **Affected:** `kernel/authz/caching.go` (+caching_internal_test), evidence/hardening-P1.

## D-0073 — Hardening P1 (S3): step-up / MFA hooks
- **Context:** the token was the only factor; the authz layer could not demand elevated auth per
  permission (roadmap S3, blueprint 07 §1 "env.mfa conditions").
- **Decision:** `authz.Permission.StepUp` marks a permission MFA-required; `authz.Actor.AMR` carries the
  surfaced auth-methods-references; `Evaluate` turns an otherwise-allowed decision into a step-up
  challenge (`Decision.StepUpRequired`, reason `step_up_required`) when the AMR carries no strong factor
  (`mfa/otp/totp/hwk/sms/fpt/face`). `env.mfa` is surfaced as an ABAC attribute. The httpx gate maps
  `StepUpRequired` → `401` + `WWW-Authenticate: … step_up="mfa"`.
- **Tradeoffs:** step-up only gates an existing allow (never grants; a plain deny is not masked — tested).
  MFA remains the IdP's job; the framework gates on the surfaced amr. Generic TOTP-challenge issuance +
  dual-control-with-workflow composition are follow-ups.
- **Affected:** `kernel/authz/{registry,authz,evaluator}.go` (+step_up_test), `kernel/httpx/authz_gate.go`,
  evidence/hardening-P1.

## D-0072 — Hardening H5 (E2): data lifecycle — generalized legal hold + DSR ledger
- **Context:** legal hold was a per-document flag (R6); no generalized hold across entities, no DSR
  primitive, no statutory-override for refusing erasure (roadmap E2, DPDP Rules live 2026).
- **Decision:** `kernel/retention` over `legal_holds` + `dsr_requests` (migration 00020). Holds
  (`Place`/`Release`/`IsHeld`/`List`) generalize hold to any `(entity_type, entity_id)` — at most one
  active hold per entity via a partial unique index — consultable by any retention sweep. DSR ledger
  (`Open`/`Complete`/`Reject`/`Get`) tracks export/erasure with a required statutory-override reason on
  rejection. All tenant-scoped under RLS.
- **Scope/tradeoffs:** the two concrete data-integrity primitives are complete + tested. Per-record-class
  disposition over arbitrary product tables is left as a registry+callback pattern (the H2 scheduler
  orchestrates; products supply per-class dispose/export/erase callbacks — no dynamic-table SQL,
  preserving the allowlist-only discipline). Wiring that registry is a documented follow-up.
- **Affected:** `kernel/retention/{retention,dsr}.go` (+`_test.go`), `migrations/00020_retention_dsr.sql`,
  evidence/hardening-H5.

## D-0071 — Hardening H3 (S1): machine authentication (API keys / service principals)
- **Context:** only OIDC user JWTs existed; non-human callers had no credential (roadmap S1).
- **Decision:** `kernel/apikey` over an `api_keys` table (migration 00019): issuable, scoped, rotatable,
  revocable, expirable keys; only `sha256(secret)` stored, public prefix is the lookup handle. Management
  (Issue/Revoke/List) is tenant-scoped app_rt; Verify is cross-tenant app_platform (tenant unknown
  pre-auth) via a permissive platform policy. `apikey.Authenticator` satisfies the H1 `httpx.Authenticator`
  port and maps a verified key to an `ActorSystem` with the key's scopes.
- **Authz integration (the flagged decision):** chose a machine-scope fast-path over capacity coupling.
  `authz.Actor` gains `Scopes []string`; `Evaluate` allows a machine actor when the perm is in its scopes
  — placed after the RBAC loop so ABAC deny still overrides, deny-by-default preserved, and scopeless
  internal system actors are unaffected (tested). Minimal, additive change to the security-critical
  evaluator.
- **Security:** constant-time secret compare (`crypto/subtle`), hash compared even for unknown prefixes
  (no timing oracle), single non-specific `KindUnauthenticated` on any failure.
- **Tradeoffs:** rotation = issue-new + revoke-old (two calls); a `wowapi apikey` CLI and per-key rate
  limits are follow-ups.
- **Affected:** `kernel/apikey/apikey.go` (+`_test.go`), `kernel/authz/{authz,evaluator}.go` (+machine_scope_test),
  `migrations/00019_api_keys.sql`, evidence/hardening-H3.

## D-0070 — Hardening H4 (S6): audit tamper-evidence via hash-chaining
- **Context:** audit_logs was append-only by grant (E1/D-0069) but had no cryptographic proof against an
  owner/DBA who bypasses the runtime role (roadmap S6).
- **Decision:** migration 00018 adds `seq`/`row_hash`/`prev_hash` to audit_logs + a per-tenant
  `audit_chain(next_seq, head_hash)`. `Record` locks the tenant chain head, assigns a gap-free seq,
  computes `row_hash = sha256(prev_hash ‖ length-prefixed canonical row)`, inserts, and advances the head
  — atomically in the caller's tx. `Verify` recomputes the chain and reports the first break (a mutated
  row's hash mismatch, or a seq gap from deletion); `Anchor` exports the head (seq+hash) for external
  notarization.
- **Correctness:** timestamp truncated to microseconds so Record's hash matches Verify's read-back;
  metadata (jsonb reformats) excluded from the hash — the audited change is what's protected;
  length-prefixed encoding prevents field-boundary collisions.
- **Tradeoffs:** every audit write now serializes on the tenant chain head (correctness over throughput,
  acceptable for audit). Verify is O(rows); anchor-based partial verification is a follow-up.
- **Affected:** `kernel/audit/audit.go` (+`_test.go`), `migrations/00018_audit_chain.sql`,
  evidence/hardening-H4.

## D-0069 — Hardening H4 (E1): durable field-level audit trail
- **Context:** the only audit was `authz.AuditSink.AuthzDenial` (denial logging via a nil-safe sink);
  the kernel stubbed "durable audit_logs writer replaces it in Phase 6". No durable, field-level,
  queryable audit existed (roadmap E1).
- **Decision:** `kernel/audit.Writer` over an `audit_logs` table (migration 00017). `Record` appends an
  entry (entity/field/before/after/actor/actor-kind/impersonator/request-id/action/reason/metadata) in
  the caller's tenant tx (commits iff the change does). `Query(Filter)` reads it back (RLS-scoped,
  newest-first with a UUIDv7 id tiebreaker for same-tx rows). A `Redactor` hook masks sensitive field
  values pre-persist. Append-only is grant-enforced: app_rt gets SELECT+INSERT but NOT UPDATE/DELETE —
  proven by a test asserting both are denied.
- **Tradeoffs:** integrity via append-only grants now; cryptographic tamper-evidence (hash-chaining) is
  S6, layering on this table. Records are written explicitly by services (no automatic trigger capture
  yet); the `AuthzDenial` denial-sink bridge is deferred (its signature lacks a tx handle). Exposed as a
  constructable primitive; a `module.Context` accessor is a follow-up.
- **Affected:** `kernel/audit/audit.go` (+`_test.go`), `migrations/00017_audit_logs.sql`,
  evidence/hardening-H4.

## D-0068 — Hardening H5 (E6): bulk-operation framework
- **Context:** the job runner processed items one at a time; a compliance product needs chunked bulk
  operations with progress, a partial-failure ledger, and resumability (roadmap E6).
- **Decision:** `kernel/bulk.Service` over `bulk_operations` + `bulk_items` (migration 00016, RLS
  tenant-scoped). `Start` records the op + one pending item per payload in the caller's tx. `Process`
  runs up to `limit` pending items (chunked; resumable — it only ever touches still-pending items),
  each in its own tenant tx: on success `fn`'s work commits atomically with the `done` mark; on failure
  that tx rolls back and a second tx records `failed` + the error. So a partial write never lingers, one
  item's failure never stops the run, and a crash resumes from the pending remainder. `Progress` reports
  Total/Done/Failed/Pending/Status. Runs as app_rt tenant-bound (bulk items are tenant data).
- **Tradeoffs:** single-processor per operation (a `FOR UPDATE SKIP LOCKED` claim would fan out across
  workers — noted follow-up). Item work must be idempotent (at-least-once, like a job worker).
- **Affected:** `kernel/bulk/bulk.go` (+`_test.go`), `migrations/00016_bulk_operations.sql`,
  evidence/hardening-H5.

## D-0066 — Hardening H5 (E3): gap-free per-tenant sequence allocator
- **Context:** no framework primitive for statutory numbered series (receipts/vouchers/certificates);
  a product would hand-roll `MAX()+1`, which races and leaves gaps — the wowsociety.app failure (E3).
- **Decision:** `kernel/sequence.Allocator` over `sequences` (per-(tenant,series) counter) +
  `sequence_allocations` (audited ledger), migration 00015, RLS tenant-scoped. `Allocate` runs the
  `INSERT … ON CONFLICT DO UPDATE next_value+1 RETURNING` inside the CALLER's tenant tx, so the number
  commits/rolls back with the business write (gap-free) and concurrent callers serialize on the row lock
  (race-free). `Void` marks an allocation voided (audited) and never renumbers — a voided statutory
  number leaves a traceable gap. `Peek` reads the last issued value.
- **Tradeoffs:** deliberately not a Postgres sequence (`nextval()` doesn't roll back → gaps). Allocations
  on one series serialize — inherent to gap-free numbering; use distinct series keys to parallelize.
  Exposed as a constructable primitive; a `module.Context` accessor is a small follow-up.
- **Affected:** `kernel/sequence/sequence.go` (+`_test.go`), `migrations/00015_sequences.sql`,
  evidence/hardening-H5.

## D-0065 — Hardening H2 (E5, R3): recurring scheduler + leader-safe kernel sweeps
- **Context:** the workflow SLA sweeper and the idempotency-key sweep existed as methods but nothing ran
  them periodically, and nothing stopped N worker replicas from all firing at once (roadmap E5 + R3).
- **Decision:** a `jobs.Scheduler` over a new `schedules` table (migration 00014). Each registered task
  has a row; a due tick is claimed by an atomic conditional `UPDATE`/`SELECT … FOR UPDATE SKIP LOCKED`
  where `next_run_at <= now()`, then `next_run_at` advances by the interval — so exactly one replica runs
  a given task per interval, **without a separate leader election**. Tasks run outside the claim tx (a
  slow task never holds the row lock); a failed task retries next interval (tasks are idempotent). Wired
  as a third loop in `StartWorker` with two kernel tasks: the cross-tenant idempotency sweep (as
  app_platform) and the per-tenant workflow SLA sweep (fan-out over active tenants via `k.Tx.WithTenant`).
  Lag is surfaced via an `OnRun` hook (logged; wireable to observability — R3 "sweeper lag as a metric").
- **Tradeoffs:** interval-based recurrence, not cron expressions (covers the P0 sweep need; a cron parser
  is a later enhancement). Per-tenant SLA fan-out is sequential; fine at current scale, shardable later.
- **Affected:** `kernel/jobs/scheduler.go` (+`_test.go`), `app/maintenance.go`, `app/worker.go`,
  `migrations/00014_schedules.sql`, evidence/hardening-H2.

## D-0064 — Hardening P1 (S2): in-process rate limiting
- **Context:** rate limiting was proxy-delegated with only middleware hooks; no in-process limiter for
  per-principal / per-permission guardrails (roadmap S2, blueprint 07 §1).
- **Decision:** `kernel/httpx.RateLimit(limiter, keyFn)` middleware + a `TokenBucket` limiter
  (`NewTokenBucket(rate, burst)`), returning 429 + `Retry-After` + RFC 7807 (reusing the existing
  `KindRateLimited`). Key strategies `KeyByIP` (edge) and `KeyByActor` (after the authz gate); products
  supply a custom keyFn for per-permission buckets. Idle buckets are swept so the key map cannot grow
  unbounded. In-memory per pod; a shared (Redis) limiter is a later adapter behind `RateLimiter`.
- **Tradeoffs:** opt-in (limits are product-specific; a forced default could throttle legitimate
  traffic) — wiring documented in the deployment checklist. Per-pod counting, not global.
- **Affected:** `kernel/httpx/ratelimit.go` (+`_test.go`), `docs/operations/deployment-checklist.md`,
  evidence/hardening-P1.

## D-0063 — Hardening H2 (O2, O5): migration reversibility drill + backup/restore
- **Context:** migrations had structure tests but no forward/down drill (roadmap O2), and there was no
  backup/restore procedure or rehearsal (O5).
- **Decisions:**
  - **O2.** Added `database.MigrateReset` (goose Down-to-0) and `TestIntegrationMigrationsReversible`,
    which runs forward→down→forward on an isolated DB in `make ci-container`. It immediately found a real
    bug — migration 00010 created `app_actor_id()` but its Down did not drop it, breaking re-apply —
    fixed in the 00010 Down. Documented the zero-downtime expand/contract pattern in
    `docs/operations/migrations.md`. Rule enforced: every object an Up creates, its Down must drop; never
    drop cluster-scoped roles/extensions.
  - **O5.** `scripts/backup_restore_drill.sh` proves the dump→restore round-trip against a seeded
    instance (marker row + schema verified; the verify step is authoritative over non-fatal client/server
    version-skew warnings). Runbook `docs/operations/backup-restore.md` documents PITR + object-store
    restore order (DB ≤ object-store timestamp, never the reverse).
- **Tradeoffs:** the drill is a logical dump/restore, not provider PITR/WAL (rehearse that in staging per
  release). `MigrateReset` is test/ops-only and must never run in production.
- **Affected:** `kernel/database/migrate.go`, `migrations/reversible_test.go`, `migrations/00010_documents.sql`
  (Down fix), `scripts/backup_restore_drill.sh`, `docs/operations/{migrations,backup-restore,deployment-checklist}.md`.

## D-0062 — Hardening H2 (R4): dead-letter-queue operability
- **Context:** dead-lettering worked (jobs → `status='discarded'`, events → `dispatch_status='dead'`)
  but there was no inspect/replay/discard path (roadmap R4). Operators could not recover poison work.
- **Decision:** kernel admin functions on the platform pool — `jobs.{ListDead,ReplayDead,DiscardDead}`
  and `outbox.{ListDeadEvents,ReplayDeadEvent,DiscardDeadEvent}` — plus a `wowapi dlq` CLI
  (`<jobs|events> <list|inspect|replay|discard>`) that connects as app_platform via DATABASE_URL.
  Replay resets status/attempts; discard DELETEs. Migration 00013 grants DELETE on both tables to
  app_platform (it already had SELECT/UPDATE from 00007).
- **Safety:** replay is safe by construction — jobs are at-least-once + idempotent workers; events
  dedup via the `processed_events` inbox on re-dispatch.
- **Tradeoffs:** durable audit of the admin action lands with the audit subsystem (H4); for now the
  action is logged. An end-to-end CLI-through-DB test was dropped (testkit isolates per-test DBs while
  the CLI reads the base `DATABASE_URL`); kernel funcs carry the integration coverage.
- **Affected:** `kernel/jobs/dlq.go`, `kernel/outbox/dlq.go`, `internal/cli/dlq_cmd.go`, `cli.go`,
  `migrations/00013_dlq_admin.sql`, evidence/hardening-H2.

## D-0061 — Hardening H1: edge middleware, cursor sort-spec versioning, sweeps, legal-hold race
- **Context:** ROADMAP-wowapi.md hardening backlog. A three-track code audit verified each item's
  "current state" claim before any work; the H1 phase closes the self-contained P0/P1 gaps
  (plan: `docs/implementation/hardening-plan.md`).
- **Decisions:**
  - **Edge middleware (S7).** The blueprint's fixed chain lists `SecureHeaders → CORS → BodyLimit →
    Timeout`, but none existed and `HTTP.MaxBodyBytes`/`RequestTimeout` were dead config. Added them to
    `kernel/httpx` (kernel-owned so every product ships the posture) and a `HTTP.CORSAllowedOrigins`
    config field (deny-by-default). Generated api wires the chain. A reference nginx + smoke.sh + a
    deployment checklist cover the proxy/TLS layer; the in-process headers are unit-tested, the nginx
    stack is a deploy/quarterly drill (adding nginx to core CI was out of proportion).
  - **Cursor sort-spec versioning (R7).** `KeysetClause` already rejected a changed column *set* but not
    a direction flip or reorder (same keys, silently wrong pages). Cursors now optionally carry a sort
    signature (`EncodeCursorWithSig`; two-key `__s`/`__v` envelope, backward-compatible with flat
    cursors), minted via `filtering.NextCursor`, validated loudly in `KeysetClause`.
  - **Idempotency sweep (S5).** `expires_at` existed but nothing purged rows. Added
    `IdemStore.SweepExpired` running cross-tenant as app_platform (migration 00012 adds a permissive
    platform policy + DELETE grant, mirroring `outbox_relay_all`). Periodic scheduling lands in H2.
  - **Retention legal-hold race (R6).** `SweepRetention` checked `legal_hold` once in an unlocked
    SELECT; a hold committed before the void was ignored. Fixed with `FOR UPDATE` (EvalPlanQual
    re-checks the qual on the locked tuple under READ COMMITTED) + a `legal_hold=false` guard on the
    void UPDATE. Proven by revert-test.
  - **Adversarial fuzzing (S8).** Native Go fuzz targets for the filter DSL parser and cursor decoder;
    seed corpus runs in CI, `make test-fuzz` drives deep runs. 1.7M/478K execs clean.
  - **Config-drift convention (O4).** The `/readyz` fingerprint had no consumer; documented an alerting
    convention + reference Prometheus rule in the deployment checklist (no framework code needed).
  - **Not gaps (roadmap inaccurate):** S4 (creds already `credential_ref` + compiler-redacted), R2, R8.
- **Tradeoffs:** `config.HTTP` now holds a slice → non-comparable; two tests moved to
  `reflect.DeepEqual`. `Timeout` keeps stdlib `TimeoutHandler`'s plain-text 503 body for now.
- **Affected:** `kernel/httpx/edge.go`, `kernel/pagination/cursor.go`, `kernel/filtering/{sort,keyset}.go`,
  `kernel/database/idempotency.go`, `kernel/document/service.go`, `kernel/config/config.go`,
  generated `cmd/api/main.go.tmpl`, `migrations/00012_idempotency_sweep.sql`, `deployments/reference/*`,
  `docs/operations/deployment-checklist.md`, `Makefile` (`test-fuzz`), evidence/hardening-H1.

## D-0060 — Review-findings pass: runtime authz gate, deploy/config-scaffold fixes, CI DB gate
- **Context:** an external review reproduced six findings against the Goal-2 framework; five were real
  (one a false-premise-free but expected deferral). Fixed each with existing conventions + regression tests.
- **Decisions:**
  - **Runtime authz enforcement (High).** The RouteMeta permission gate was boot-validated but NEVER
    enforced per request — a deployed API served every route unauthenticated/unauthorized. Added
    `httpx.SecureHandler`/`gateRoute`: for each non-Public route, AuthN (via a pluggable `Authenticator`
    port — the product supplies OIDC/tenant strategy) → bind tenant+actor → AuthZ(permission) at tenant
    scope → serve; deny-by-default. The generated api wires it with `DenyAllAuthenticator` (fail-closed:
    business routes 401 until a real Authenticator is set). Fine-grained resource checks stay per-handler.
  - **Workflow pagination off-by-one (Medium).** `OpenTasksFor` encoded the cursor from the dropped
    lookahead row, skipping one task per page boundary; now encodes the last RETURNED item. Regression
    test proven by revert (skips 1 → paged 4/5).
  - **deploy render (High).** Defaulted `--env production` (invalid; valid is `prod`) and rendered
    `${WOWAPI_DB_DSN}` (config.DB.DSN is a Secret needing `secretref://`). Now defaults `prod`, validates
    `--env` via `config.Env.Valid()`, and renders `secretref://env/WOWAPI_DB_DSN` (+ MIGRATE_DSN).
  - **Product config scaffolding (Medium).** `wowapi init` now scaffolds `internal/appcfg` (product
    Config embedding config.Framework + Modules namespaces, D-0002) and `tools/configcheck` (D-0003); the
    generated api/worker load via `appcfg.Load` and pass `cfg.Modules` to `Boot` (was `nil`).
  - **CI DB-skip hygiene (Medium).** DB-backed tests SKIP without a DSN, so host `make ci` could be
    green-but-hollow. Added `testkit.RequireDB()` (WOWAPI_REQUIRE_DB=1) → FAIL not skip; `make ci-container`
    and `make test-integration` set it, so the authoritative gate cannot silently skip DB/E2E proofs.
  - **Deferrals (Lower) — no change.** Workflow vote/min_approvals>1/self_approval are fail-closed
    (D-0054), audit_logs is the logging sink, gen-crud emits honest TODO handlers — all already
    accurately documented as deferrals; verified no doc overclaims them complete.
- **Affected:** kernel/httpx/{authz_gate,router}.go, kernel/workflow/runtime.go, internal/cli/{deploy_cmd,
  init_cmd}.go + templates, testkit/db.go + consumer_test, internal/e2e, internal/testmodules/requests,
  Makefile; evidence/phase-12 acceptance-map (#18 now runtime-enforced).

## D-0059 — Phase 12: `wowapi init` produces a framework-wired product repo; E2E acceptance
- **Context:** Phase 12 (capstone) must prove a blank repo builds a WORKING API binary (AC #19) and runs
  kernel + module migrations from cmd/migrate (AC #22). The Phase-10 init mains were framework-import-free
  stubs — a gap.
- **Decisions:**
  - **The scaffolded mains wire the framework.** `wowapi init` now renders real `cmd/api|worker|migrate`
    mains: config load → pool (runtime AS app_rt + RLS guard; worker also a platform pool) → `kernel.New`
    → `app.New().Register(wire.Modules()...).Boot` → serve the router behind the observability middleware
    chain + `/healthz`//`/readyz`, graceful shutdown; worker runs `app.StartWorker`; migrate runs
    `migrations.Kernel()` then each module's migrations. Modules are registered via a generated
    `internal/wire/modules.go` (manual list — auto-append is a documented follow-up).
  - **Config scaffold uses secret references.** `configs/local.yaml` renders `secretref://env/DATABASE_URL`
    (raw/empty DSN strings fail `Secret.UnmarshalText` by design) — the secret-ref-only guarantee shows up
    in the scaffold itself.
  - **E2E test = acceptance through the real CLI.** `internal/e2e` runs `wowapi init`, replaces wowapi with
    the local tree, `go build`s the repo, and (with a DB) runs the migrate binary + curls the api binary's
    `/healthz` — following the consumer test's offline-skip discipline.
  - **Release notes + full acceptance sweep.** `CHANGELOG.md` (v0.1.0); the 28-criterion acceptance map.
- **Affected:** internal/cli/templates/init/* (cmd mains + internal/wire + config), internal/cli/init_cmd.go,
  internal/e2e/e2e_test.go, CHANGELOG.md; evidence/phase-12/. **Goal 2 complete (Phases 0–12).**

## D-0058 — Phase 11: observability + performance budgets + security suite + config drift
- **Context:** Phase 11 hardens the framework (blueprint 07 §1–2/§9; AC #17/#18/#26/#27) — observability
  wiring, perf budgets, a security gate, and cross-process config drift. Additive; no new domain tables.
- **Decisions:**
  - **Observability = ports + adapters:** `kernel/observability` defines a small `Metrics` port
    (ObserveRequest/IncCounter/SetGauge) + a NoOp default + RED and AccessLog middleware; the Prometheus
    client lives ONLY in `adapters/metrics/prometheus` (with a `/metrics` handler). The RED middleware
    labels by the matched route PATTERN (bounded cardinality). Full OTel span export is a product adapter.
  - **Health:** `kernel/httpx/health.go` — liveness runs NO checks (a failing dep must not trip a
    liveness probe); readiness runs checks → 200/503 and reports the redacted config fingerprint.
    `app.Readiness` assembles module `ctx.Health` + framework checks (DB ping / migrations-current,
    supplied by the composition root) + fingerprint.
  - **Performance budgets (#17):** 24 hot-path benchmarks + a pure-Go `internal/tools/benchbudget` gate
    reading piped `go test -bench` output against `bench-budgets.txt`, wired into `make ci`. Config field
    reads at 0.3 ns/op, 0 allocs prove the hot path is reflection/lookup-free.
  - **Security suite (#18/#26):** a curated `make test-security` gate over the existing RLS/authz/
    privilege/secret tests + new per-knob unsafe-config matrix + a structural-secret-redaction gap test.
    Audit found the core guarantees (deny-by-default, secret-ref-only, structural redaction, RLS,
    unsafe-config-fails-startup) have no disabling config key.
  - **Config drift (#27):** `kernel/config/shared.go` — `SharedFingerprint` covers env/schema/DB
    (excludes process-specific HTTP/Log); `CheckSharedDrift(expected)` fails a mis-deployed process.
- **Affected:** kernel/observability, adapters/metrics/prometheus, kernel/httpx/health.go,
  kernel/config/shared.go, app/health.go, internal/tools/benchbudget, bench-budgets.txt, Makefile
  (bench/bench-budget/test-security + bench-budget in ci), benchmarks + security tests; evidence/phase-11/.

## D-0057 — Phase 10: installable `wowapi` CLI (scaffolding, codegen, tooling) + review fixes
- **Context:** Phase 10 delivers the CLI command surface (blueprint 10 §2 E21): init, new-module,
  gen crud, migrate create, seed validate, openapi merge, lint boundaries, deploy render — plus the
  existing version/config. No new DB tables.
- **Decisions:**
  - **Dispatcher = one file per command:** `internal/cli/cli.go` switches to a `runX(args, stdout,
    stderr) int`; each command is its own file, buffer-testable. Enabled a conflict-free parallel build
    (lead: transform commands; agent: scaffolding).
  - **Generated Go is gofmt-clean:** `renderToFile` runs `go/format.Source` on `.go` output — formats
    AND fails generation loudly on an invalid-Go template (stronger than a parse-only check).
  - **Scaffold path safety:** module/resource/field names are `identRE`-validated before any path is
    built (no traversal); `--force` gates every overwrite.
  - **lint reuses the framework law:** `wowapi lint boundaries` ports the import-layering + module-
    isolation rules from `scripts/lint_boundaries.sh` as a pure, unit-tested `checkBoundaries`; the
    shell script remains the authoritative framework gate for vocabulary/Reveal/test-import checks.
  - **Review fixes (D-0057):** unknown `gen crud` field type rejected instead of emitting unbuildable Go
    (CLI-01); `openapi merge` rejects non-object fragments (CLI-02); `checkBoundaries` gained the missing
    adapters/cmd/internal-cli/internal-tools layer rules + hard testkit rule (CLI-06); usage-error exit
    codes normalized to 2 (CLI-03); `go list` stderr surfaced (CLI-04); stdout write errors propagated
    (CLI-05); derived package name validated (CLI-07).
- **Affected:** internal/cli/ (all command files, scaffold.go, templates/, tests), cmd/wowapi;
  evidence/phase-10/.

## D-0056 — Phase 9: notify / webhook / integration framework + review fixes
- **Context:** Phase 9 delivers the notification, webhook, and integration subsystems (migration 00011,
  blueprint 07 §5/§6). Two parallel review agents reproduced 13 defects (evidence/phase-09/review-findings.md).
- **Decisions:**
  - **Config tables are app_platform-written (SEC-13):** notification_templates, integration_providers,
    and webhook_endpoints are behavior-changing config (which channels/endpoints fire, which credentials
    sign) — app_rt SELECT-only. notifications is module-written in a business tx; notification_deliveries
    and webhook_events are append-only to app_rt with status advanced by the app_platform sender/relay.
  - **Notifications:** template registry (module-declared, allowlisted vars, `text/template` — but
    `html/template` for the email channel to auto-escape, SEC-51); `Send` writes the notification + one
    delivery per resolved channel in the caller's tenant tx and dry-run-renders each body so a missing
    var fails synchronously (ARCH-77); `SendPending` (app_platform) claims + delivers with a
    `next_attempt_at` backoff and a maxAttempts dead-letter (ARCH-75).
  - **Webhooks:** inbound `HandleInbound` verifies the provider signature (constant-time HMAC), enforces
    replay via a synthesized-or-provided dedup id over a PARTIAL unique index (SEC-49) and a ±5m window;
    a signature-failure audit row carries a NULL dedup id so it cannot block a real event (SEC-50);
    outbound signing covers `timestamp + "." + body` (SEC-52). `RetryOutbound` (app_platform) is the
    worker that actually drives outbound backoff/DLQ — DispatchOutbound alone gave one attempt (ARCH-70).
    A per-endpoint circuit breaker opens after N failures, half-opens after a cooldown, and clears the
    persisted `degraded` status on recovery (ARCH-72).
  - **Integrations:** a provider-adapter registry (anti-corruption boundary) + a store that resolves
    per-tenant/platform config and a credential from a secret REFERENCE (plaintext rejected); `Upsert`
    uses `RETURNING id` so the conflict path returns the real row id (ARCH-71); `HealthChecks` probes
    configured providers for readiness.
  - **Hybrid RLS backstop (SEC-53):** a RESTRICTIVE policy on the platform+tenant hybrid tables forbids
    a tenant-bound session from writing a NULL-tenant (platform) row.
  - **events_outbox INSERT for app_platform:** granted in 00011 so tenant-bound workers (inbound
    handlers, the delivery sender) can emit events; the relay's WITH CHECK admits it, the outbox Writer
    stamps the tenant.
- **Affected:** kernel/notify, kernel/webhook, kernel/integration,
  migrations/00011_notify_webhook_integration.sql, kernel/kernel.go, module/module.go,
  app/{context,boot}.go; evidence/phase-09/.

## D-0055 — Phase 8: document/file framework (storage port, append-only versions, grant RLS) + review fixes
- **Context:** Phase 8 delivers documents/versions/grants/comments/attachments (migration 00010,
  blueprint 07 §4). Two parallel review agents reproduced 13 defects (evidence/phase-08/review-findings.md).
- **Decisions:**
  - **Object storage is a port (`kernel/storage.Adapter`):** PresignPut/Get + Stat + Peek + Delete;
    blob bytes never transit the API process (client ↔ store via presigned URLs). A memory adapter
    backs tests + local dev; an S3/minio adapter implements the same five methods.
  - **Append-only versions + privilege split:** `document_versions` is INSERT-only to app_rt;
    scan-status settlement and retention voiding run as app_platform (tenant-bound via a PlatformTxM),
    so a module can neither rewrite an immutable file pointer nor clear an infected scan flag.
  - **Download authorization is deny-first + owner + capacity-grant:** an explicit deny policy from
    the authz evaluator is authoritative; otherwise the document owner, an authz role/policy allow, or
    a valid (windowed) capacity grant permits. Two kernel-owned permissions (`kernel.document.read`,
    `kernel.document.update`) are registered at boot.
  - **Grant writes are RLS-ownership-enforced (SEC-41/42):** a new `app_actor_id()` SQL function +
    a RESTRICTIVE policy pin every `document_access_grants` INSERT/UPDATE to a document the acting
    actor owns — a module cannot self-grant or redirect a grant even via raw SQL. Chosen over an
    app_platform-only grant path to keep grant creation composable in the module's business tx.
  - **Governance columns are app_platform-only (SEC-44):** app_rt gets column-level UPDATE on
    documents (title/sensitivity/version/updated_*) but NOT status/legal_hold/retention_until — a
    module cannot clear a legal hold or void a document to dodge retention.
  - **Download is a pure read (ARCH-65):** it emits NO outbox event (that INSERT broke read-only-tx
    callers); durable download audit is deferred to the audit_logs writer.
  - **Retention sweep ordering (SEC-48):** rows are tombstoned inside the tx; blobs are deleted only
    AFTER commit — a failure orphans a blob (safe) rather than leaving an active row over a deleted blob.
  - **Random storage keys (ARCH-66):** the upload key uses a UUID suffix, not the version number, so
    concurrent InitiateUpload calls never clobber each other's blob.
  - **Comment/attachment author guards (SEC-45/46):** Go-level author/creator checks (fail-closed on
    no actor) for edit/void/detach — the realistic user-vs-user protection; a trusted in-process
    module issuing raw SQL can still touch its own tenant's rows (accepted; DB-level protection is
    reserved for the cross-authorization/legal controls).
- **Affected:** kernel/storage, kernel/document, kernel/comment, kernel/attachment,
  migrations/00010_documents.sql, kernel/kernel.go, module/module.go, app/{context,boot}.go,
  testkit/db.go; evidence/phase-08/.

## D-0054 — Phase 7 review: temporal resolution, write-time schema, draft/activate split, workflow fail-closed
- **Context:** two parallel review agents (security + architecture) reproduced eight gaps in the
  rules + workflow slice (see evidence/phase-07/review-findings.md).
- **Decisions:**
  - **Historical resolution includes superseded (ARCH-60):** the resolver reads
    `status IN ('active','superseded')` within the temporal `effective_from/to` window, not
    `status='active'` — a value active in the past then superseded must still resolve for an `at`
    inside its old window rather than falling through to the code default.
  - **Write-time schema validation (SEC-40):** `Propose` validates the value against the point's
    `value_schema` (focused `type`+`enum` validator, `kernel/rules/schema.go`) before INSERT —
    defense in depth over read-path Decode. Full JSON Schema deferred.
  - **Draft/activate privilege split (SEC-13):** `Propose` inserts a DRAFT on app_rt (never
    resolves); `Activate` supersedes+activates on app_platform via a role-scoped
    `rule_versions_platform_all` policy. Activation changes runtime behavior, so it stays off the
    module role. `created_by` is the proposing actor from `ActorIDFrom(ctx)` (ARCH-62).
  - **Workflow fail-closed on unenforced gating (SEC-36/37/38):** the runtime does not yet tally
    votes, enforce `min_approvals > 1`, or exclude self-approval, so the definition validator
    REJECTS such definitions at boot rather than accepting and mis-enforcing them. `Policy.SelfApproval`
    is `*bool` to distinguish unset from explicit false. Approval steps must define both
    `on_approve.next` and `on_reject.next` (ARCH-64). Per R7, fail-closed is the acceptable posture
    for an unshipped control.
  - **Override authz gate (SEC-39):** `Runtime.Override(ctx, actor, id, to, reason)` evaluates
    `workflow.instance.override` on the instance resource before forcing a step; deny → `KindForbidden`.
  - **Test-suite fix:** `TestVerify_TamperedSignature` flipped the trailing base64url char of the
    JWT signature, which can carry only discarded padding bits → non-deterministic (passed on host,
    failed in-container). Now flips the first char (always 6 significant bits); 200× stable.
- **Affected:** kernel/rules/{resolver,store,schema}.go, kernel/workflow/{definition,runtime}.go,
  kernel/auth/auth_test.go; evidence/phase-07/.

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

## D-0031 — Phase 3: idempotency_keys migration (00003) ships now, out of 03 §5 order
- **Context:** phase-plan row 3 requires tested idempotency helpers; 05 §2's `IdemStore` needs the
  `idempotency_keys` table, which blueprint 03 §5 lists in migration 009 (a Phase 6 batch).
- **Decision:** pull the single `idempotency_keys` table forward into kernel migration
  `00003_idempotency.sql` (tenant-scoped, ENABLE+FORCE RLS, granted to app_rt) so the Phase 3
  idempotency store is real and integration-tested against RLS now. The remaining migration-009
  tables (outbox, processed_events, job_runs, audit_logs) still land in Phase 6. Migration numbers
  are per-source and monotonic, so pulling one table forward is safe.
- **Affected:** migrations/00003_idempotency.sql; kernel/database/idempotency.go (IdemStore + pg
  impl); kernel/httpx/idempotency.go (WithIdempotency); docs/blueprint/03 §5 note.

## D-0032 — Phase 3: module.Context gains Routes() and Validator()
- **Context:** D-0006 grows Context per phase; Phase 3 delivers httpx + validation, so modules can
  now register routes and validate input.
- **Decision:** add `Routes() *httpx.Router` and `Validator() *validation.Validator` to
  module.Context (and the app-side moduleContext). Route registration errors surface at boot via
  Router.Err(). Tx()/Authz()/etc. still arrive in their phases.
- **Affected:** module/module.go, app/context.go.

## D-0033 — Phase 3 review: the database layer may emit taxonomy Kinds (ARCH-30)
- **Context:** D-0024 kept `kernel/database` on exported sentinels mapped upstream. `IdemStore`
  naturally produces conflict / retry_later / in-flight outcomes that ARE taxonomy Kinds
  (KindConflict, KindIdempotencyInFlight); returning sentinels and re-mapping them in httpx would
  duplicate the taxonomy.
- **Decision:** `kernel/database` MAY import `kernel/errors` and return `*errors.Error` for
  outcomes that map cleanly to a Kind (idempotency, and version-conflict helpers may migrate to
  this too). The graph stays acyclic — `kernel/errors` imports only stdlib. Encoded a `depguard`
  rule in `.golangci.yml` (kernel must not import module/app/adapters/testkit) so the import law is
  machine-checked, not just documented.
- **Affected:** kernel/database/idempotency.go, .golangci.yml.

## D-0034 — Phase 3 review: idempotency review-finding resolutions
- **SEC-16/ARCH-27 (critical, reproduced):** the claim raced (SELECT-FOR-UPDATE cannot lock a
  non-existent row, so concurrent first-uses both went Fresh and the unconditional upsert clobbered
  a completed response). Rewritten to atomic `INSERT … ON CONFLICT DO NOTHING RETURNING` — only a
  real insert is Fresh; otherwise `SELECT … FOR UPDATE` and branch (completed→replay, hash
  mismatch→conflict, expired→re-claim, else in-flight). Concurrency regression test
  (`TestIntegrationIdempotencyConcurrent`, 8 goroutines, exactly-once, passes ×5 under `-race`).
- **SEC-18 (medium, reproduced):** `Recover` appended a problem body to already-written responses
  and swallowed `http.ErrAbortHandler`. Now tracks whether bytes were written (skips the problem
  body if so) and re-panics on ErrAbortHandler.
- **ARCH-32/SEC-23:** `WithIdempotency` now stores only 2xx responses; non-2xx claims are discarded
  (stay retryable) via the new `IdemStore.Discard`.
- **SEC-19:** `RequestHash` now includes the URL query string.
- **ARCH-29:** `DecodeJSON` rejects a literal `null` body like an empty one.
- **ARCH-31/SEC-22:** added `filtering.KeysetClause` (blueprint 05 §2, previously missing) with
  cursor-key allowlisting + `Sort.Terms()` accessors; columns come only from the sort allowlist,
  cursor supplies only bound values.
- **ARCH-34:** `RequireIfMatch` rejects `*` (optimistic concurrency requires a concrete version).
- **Accepted/deferred:** ARCH-28/SEC-21 (Router.Err() enforced at boot) → Phase 5 app wiring;
  ARCH-35 (ScopeExtractor `any` → authz.Target) → Phase 4; SEC-20 (duplicate JSON keys / no
  Content-Type check) → defense-in-depth noted, strict decode + domain validation suffice.
- **Affected:** kernel/database/idempotency.go, kernel/httpx/{idempotency,middleware,decode,etag}.go,
  kernel/filtering/{sort,keyset}.go; evidence/phase-03/review-findings.md.

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

## D-0080 — CA-12: reversibility schema-diff + PITR/object-storage restore legs are scripted in-repo; production PITR stays provider-owned (B-4/B-5)
- **Context:** CA-12's O2/O5 finishers had two gaps. B-4: the reversibility drill only checked the
  goose-version round-trip + one sentinel table, so a migration whose Down is *schema*-asymmetric
  (drops a table but leaves an index/policy/default/stray object) passed while the physical schema
  drifted. B-5: the PITR + object-storage restore legs were unscripted — only described in the runbook.
- **Decision:**
  - **B-4 (done):** `scripts/migration_reversibility_drill.sh` (`make drill-reversibility`) migrates
    up → down-to-0 → up on a throwaway DB and **diffs normalized `pg_dump --schema-only` snapshots**
    at head; any byte difference fails with a non-zero exit. This required a `reset` (down-to-0) mode
    on `internal/tools/migrate` so a shell drill can drive `MigrateReset`. Verified: passes on head,
    and catches an injected asymmetric Down that the goose-version round-trip does not.
  - **B-5 (done — real round-trips, NOT rescoped):** both legs are scripted against the local stack.
    `scripts/pitr_restore_drill.sh` (`make drill-pitr`) runs a genuine PITR — a throwaway
    `postgres:16-alpine` primary with `archive_mode=on`, physical `pg_basebackup`, then restore +
    `recovery_target_time` WAL replay + promote, asserting recovery stops exactly at the target.
    `scripts/object_storage_restore_drill.sh` (`make drill-object-storage`) runs a MinIO blob
    backup→loss→restore round-trip with a byte-identity + `storage_key`-resolves check.
  - **Scoping (the only thing NOT built in-repo):** *production* PITR — continuous WAL archiving,
    retention, and cross-region object replication — remains a **managed-provider capability**, not the
    ephemeral compose DB's job (whose `archive_mode` is intentionally off). The compose stack cannot
    stand up a production-representative WAL archive/retention tier, so that layer stays a staging/
    provider rehearsal per release train (backup-restore.md). The drills prove the *recovery procedure*
    against real WAL/blobs; the provider owns the durability substrate.
- **Rationale:** preferred real scripted round-trips over a rescope wherever the stack allowed it (both
  legs did). The residual provider-owned piece is a genuine deploy-environment concern, recorded here so
  it is a deliberate boundary, not an untracked gap.
- **Affected:** scripts/migration_reversibility_drill.sh, scripts/pitr_restore_drill.sh,
  scripts/object_storage_restore_drill.sh, internal/tools/migrate/main.go, Makefile (drill-* targets),
  docs/operations/{migrations,backup-restore}.md.

## D-0091 — Booted is opaque-by-construction (the V2 opacity decision)

**Date:** 2026-07-17 · **Context:** fourth/fifth adversarial closure audits;
V1 discarded before any production consumer existed. `app.Booted` is fully
OPAQUE in V2: no informational mirror fields, no exposed kernel aggregate —
every capability flows through accessors backed by the boot-validated runtime
view, and values not produced by `App.Boot` fail loudly (`ErrNotBooted` /
accessor panics). Rationale: informational mirrors created two sources of
truth (the F-10 defect class); a hand-constructed `Booted` never passed boot
validation, so silently operating on one was the vulnerability itself. The V2
contract fixtures (`internal/compat/v2_contract_test.go`) freeze
`app.Hook` and `document.UploadEvent` — the V2 positional-literal contract —
and intentionally exclude `app.Booted`, which is not constructible.
