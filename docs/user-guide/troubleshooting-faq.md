# Troubleshooting & FAQ

Concrete symptoms, causes, and fixes — plus answers to the questions newcomers ask. Every entry maps to a
real behavior of the framework, not a hypothetical.

## Startup & configuration

| Symptom | Cause | Fix |
|---|---|---|
| `environment must be set` | overlay missing `environment:` | Add `environment: <env>` to `configs/<env>.yaml`. |
| `config: invalid configuration:` + a **list** | one or more invalid values (loader reports all at once) | Fix each line. Ranges: `db.max_conns` 2–200, `db.query_timeout` 100ms–60s. |
| `secretref resolution failed` | env var behind `secretref://env/<VAR>` not set | Export the named var. |
| Boot refuses with flags in prod | local flag overrides used with `environment: prod` | Remove flags; use env vars/overlays. |
| Unknown config key error | typo in a `modules.<name>.*` key | Fix the key — unknown keys are rejected by design. |
| `db.dsn required` | `DATABASE_URL`/`MIGRATE_URL` unset or wrong `APP_ENV` | Export them; set the right `APP_ENV`. |
| `db.platform_dsn is required` | `PLATFORM_URL` unset — api/worker fail closed without a dedicated `app_platform` login | Export `PLATFORM_URL` (never reuse the `app_rt` DSN). |

## Database & tenancy

| Symptom | Cause | Fix |
|---|---|---|
| `permission denied for table …` at runtime | connected as the wrong role, or missing `GRANT` | Use the `app_rt` DSN; add `GRANT … TO app_rt` in the migration. |
| Error mentioning `app_tenant_id` | query ran with no tenant context | Wrap DB work in `tx.WithTenant`/`WithTenantRO`; the tenant comes from `ctx` (`database.WithTenantID`). |
| `must be owner of table` during migrate | migrating as `app_rt` | Use the `MIGRATE_URL` (`app_migrate`) DSN for DDL. |
| Rows from other tenants visible | RLS not forced, or policy missing `WITH CHECK` | Add `FORCE ROW LEVEL SECURITY` + a `WITH CHECK` clause. |
| Reversibility gate fails | migration `Down` doesn't reverse `Up` | Fix `Down`, or declare the migration forward-only with a reason. |
| Tenant queries silently run RLS-off | over-privileged DSN / RLS guard disabled | Connect as `app_rt`; enable `WithRLSGuard()` (required in deployed processes). |

## Auth

| Symptom | Cause | Fix |
|---|---|---|
| Every business route → **401** | `DenyAllAuthenticator` (the secure default) still wired | Implement + wire a real `httpx.Authenticator` — [Auth](auth.md). |
| Valid user → **403** | permission not granted, or not declared in `Permissions()` | Grant it; ensure the permission is registered (unregistered ⇒ never authorizable). |
| Boot: "Public but also sets Permission" | route marked both | Pick one in `RouteMeta`. |
| Boot: non-public route has no permission | missing `Permission` | Add a `Permission` or `Public: true`. |
| Expected a re-auth prompt, got 403 | `Decision.StepUpRequired` unhandled by client | Handle the step-up challenge; populate `Actor.AMR`. |
| List shows rows a user shouldn't see | query not built from `Evaluator.Filter` | Apply the returned `ListFilter`. |

## HTTP & validation

| Symptom | Cause | Fix |
|---|---|---|
| `400 validation_failed` with `errors[]` | request body failed struct-tag validation | Read `errors[].field`/`code`; fix the payload. |
| `400` on a well-formed body | unknown JSON field (strict decoding) | Remove the extra field, or add it to the DTO. |
| `413`/body error | body exceeds `http.max_body_bytes` (or the handler's cap) | Reduce payload or raise the limit deliberately. |
| `409 conflict` / `412 version_conflict` | uniqueness or optimistic-concurrency clash | Re-read and retry with the current version. |
| `404` on a cross-tenant id | `KindTenantIsolation` masks existence as 404 | Working as intended — don't "fix" it to 403. |
| `500` with no detail | `KindInternal` never exposes its message | Look up the `request_id` in server logs for the real cause. |

## Async (outbox / jobs / scheduler)

| Symptom | Cause | Fix |
|---|---|---|
| Events/jobs never process | `cmd/worker` not running | Run the worker alongside the api. |
| A job keeps failing | poison message | Inspect + recover via `wowapi dlq jobs list/inspect/replay/discard`. |
| An event never delivered | it reached the dead-event table | `wowapi dlq events list/inspect/replay/discard`. |
| Scheduled work double-runs | expected safety is `FOR UPDATE SKIP LOCKED` | Confirm one worker per schedule tick; the kernel already guards this. |

## Tests & CI

| Symptom | Cause | Fix |
|---|---|---|
| DB tests "skipped" | no DB / `WOWAPI_REQUIRE_DB` unset | Use `make ci-container` (sets the flag + provides Postgres). |
| `testkit.NewDB` skips locally | no admin DSN | `make up` and export the admin DSN, or run non-DB suites. |
| Isolation test passes but shouldn't | queried via `h.Admin` (owner) not `h.Runtime` | Use `h.Runtime`/`h.TxM` — RLS binds only `app_rt`. |
| Boundaries lint fails | a module imports another module or crosses a layer | Route through a **port**; see [Modules](modules.md#inter-module-communication-ports). |

---

## FAQ

**What is wowapi, in one sentence?**
A domain-neutral, modular-monolith backend framework in Go that gives you multi-tenant PostgreSQL with
row-level security, deny-by-default authorization, a transactional async platform, and compliance
primitives behind a small module SDK.

**Is it a library or an application?**
Both, by role. You consume it as a **versioned Go dependency** and build your product in a separate repo
(`wowapi init`). The framework repo is where the kernel is developed. See [Architecture](architecture.md).

**What version is this, and is it stable?**
Pre-1.0 (`v0`). The module SDK (`module.Context`) still widens as kernel capabilities land — interface
changes are an accepted breaking change while v0. Pin an exact version in your `go.mod`.

**Do I write SQL, or is there an ORM?**
You write SQL through the tenant-scoped `TxManager` (`WithTenant`/`WithTenantRO`) over pgx. There's no ORM;
RLS + the DTO discipline replace the usual reasons for one. See [Modules](modules.md).

**How do I keep tenants isolated?**
You don't do it in application code — PostgreSQL RLS does, forced on the `app_rt` role, keyed off
`app_tenant_id()`. You just run DB work inside `WithTenant`. See [Database & migrations](database-migrations.md).

**Why does every route return 401 out of the box?**
Because the default authenticator is `DenyAllAuthenticator` — a safe default. Identity is deployment-
specific, so you supply the real `Authenticator`. See [Auth](auth.md).

**How do two modules talk to each other?**
Through **ports** (`ProvidePort`/`Port`), never by importing each other. `make lint-boundaries` enforces it.

**Where do secrets go?**
Never in YAML as plaintext. Config `Secret` fields accept only `secretref://env/<VAR>` (the `env` provider
is the one wired today); other providers are a documented seam. See [Configuration](configuration.md).

**What's the authoritative test command?**
`make ci-container`. It runs `make ci` in a container with Postgres and `WOWAPI_REQUIRE_DB=1`, so DB tests
can't silently skip. See [Testing](testing.md).

**How do I recover a stuck background job?**
`wowapi dlq jobs inspect <id>` then `replay` or `discard`; same for `dlq events`. See
[CLI reference](cli-reference.md).

**Is there a license?**
**No `LICENSE` file is present in the repository yet.** Until one is added, treat the code as
"all rights reserved" and confirm usage terms with the maintainers. (Documented as a gap, not a claim.)

**Where's the deeper design rationale?**
The [blueprint](../blueprint/README.md) explains the *why* behind every subsystem; the
[decisions log](../implementation/decisions.md) records specific trade-offs; the
[operations runbooks](../operations/deployment-checklist.md) cover running it.

Back to the [User Guide index](README.md) · the [project README](../../README.md).
