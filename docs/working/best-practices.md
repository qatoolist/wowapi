# wowapi — Best Practices Guide

How work is done here. Rules are grounded in this codebase's conventions and in the concrete mistakes
found by three review passes ([review-learning-register.md](review-learning-register.md)). Follow them;
deviations from a blueprint/convention go in `docs/implementation/decisions.md` first.

## Understand before implementing
- Read the goal/roadmap item literally and enumerate every sub-requirement before touching code.
- Read the relevant blueprint section (`docs/blueprint/`), the existing package, and the nearest
  decision entries. Prefer `mcp__lumen__semantic_search`, then Grep/Read.
- If a design choice is load-bearing or security-critical (authz, audit, tenancy), decide it explicitly
  and record it in `decisions.md` **before** the code.

## Inspect existing code before adding new code
- Search for an existing package/type/helper that already does it. The kernel has ~30 packages; new
  functionality usually extends one (e.g. add to `kernel/retention`, don't make a parallel package).
- Match the existing pattern: registry + shared pointer for boot-time registration; service methods take
  `(ctx, db database.TenantDB, …)`; errors via `kerr.E`/`kerr.Wrapf`; append-only via grants; RLS policy
  on every tenant table.
- Never invent an API, field, config key, or migration column that isn't already supported without
  checking it exists. Anti-hallucination is a hard rule.

## Reuse conventions & preserve terminology
- Roles: `app_rt` / `app_platform` / `app_migrate`. Sinks/sweeps/relay run cross-tenant as app_platform.
- Naming: migrations `NNNNN_snake.sql`; permissions `module.resource.verb` (closed verb set); events/jobs
  `module.resource.verb`; decisions `D-00NN`; evidence bundles per phase.
- Keep the same words the blueprint uses (tenant, capacity, actor, RouteMeta, TenantDB, secretref). Don't
  introduce synonyms.

## Don't invent unsupported logic
- No dynamic SQL over arbitrary tables — the framework is allowlist-only (filter/sort DSL, retention
  callbacks). If a feature needs product data access, use a registry+callback (see `retention.Engine`),
  not generated table names.
- No disabling switch for a core security guarantee (RLS, deny-by-default, redaction).

## Avoid duplicate tests / duplicate implementation
- Before writing a test, check for an existing one covering the case (`miscellaneous/find_duplicate_tests.sh`).
- Before adding a helper, grep for an equivalent. Extend, don't fork.
- One canonical place per concept (e.g. `actorFromCtx`, `nullStr`, cursor encode) — reuse it.

## Write meaningful tests
- **TDD**: failing test first, watch it fail, implement. For subtle fixes (concurrency, off-by-one,
  security), prove the test catches the bug by reverting the fix (as done for the R6 legal-hold race and
  the migration reversibility drill).
- **Real integration tests over mocks** against real Postgres via `testkit`; DB tests must FAIL not skip
  under `WOWAPI_REQUIRE_DB=1`. A green suite that skips the meaningful test is not proof.
- Cover boundaries and adversarial input: page boundaries, concurrency (parallel allocations), rollback,
  RLS isolation (cross-tenant must see nothing), append-only denial, expiry/revocation, injection (fuzz).

## Run regression checks (the authoritative gate)
- `make ci` (host: vet, lint, boundaries, unit, race, perf budgets, build) AND `make ci-container`
  (authoritative — DB/integration tests forced via `WOWAPI_REQUIRE_DB=1`). Both must be 0 FAIL / 0 SKIP.
- `make lint-boundaries` after any new package/import. `make test-fuzz` for parser changes.
- Confirm pre-existing tests still pass (no regressions), and `gofmt -l` is clean.

## Document decisions & maintain traceability
- `decisions.md` entry (context → options → decision → tradeoffs → affected files) for every deviation,
  **before** the code.
- Evidence bundle per phase/goal (`proof-bundle.md` + `review-findings.md` + `command-log.md` +
  `acceptance-map.md`); "reviewed/tested/verified" without an entry counts as not done.
- Update `CHANGELOG.md` `[Unreleased]`. Keep claims honest — never write "complete" next to a deferral.

## Handle third-party & internal review findings
- Treat every finding as real until disproved; verify against the code, not the reviewer's or your own
  summary. Classify severity + impact, locate the affected requirement/module/test/artifact, fix per
  conventions, add/strengthen a test, re-run the gate, then re-run the review gate.
- Log the learning in [review-learning-register.md](review-learning-register.md); if the class recurs,
  promote it to a checklist rule.

## Prevent repeat mistakes (the six recurring patterns — check every time)
1. **Built-but-not-wired** — trace entry→effect; is the new code actually called at runtime?
2. **Deferred-claimed-as-done** — every sub-requirement Done/Partial/Missing; Partial ≠ Done.
3. **Green-but-hollow tests** — no skips masking coverage; DB tests actually run.
4. **Artifact-doesn't-actually-work** — parse/run/boot every generated/rendered artifact.
5. **Missing-required-infra** — did the feature need a container/config/migration/grant? Deliver it.
6. **Local-not-production** — check production config path, locked-down roles, secrets.

## Keep work production-ready
- Fail-closed defaults (DenyAllAuthenticator, fail-closed `app_tenant_id()`); best-effort side paths must
  never block the main decision (durable audit, last-used bump).
- Provide the supporting infra with the feature (a container in compose, a config knob, a migration + the
  grant). "Works in the test DB" is not done.
- No stray artifacts committed (built binaries, dumps); `git status` clean per commit; commit per coherent
  unit with a decision reference.
