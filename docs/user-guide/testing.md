# Testing

wowapi's test strategy is **real integration over mocks**: DB-backed tests run against a real PostgreSQL
with RLS active, each test on its own isolated database. This page covers the suites, the `testkit`
harness, the authoritative gate, and regression/fuzz testing. (`testkit/`, root `Makefile`.)

## The suites (Makefile targets)

| Target | What it runs |
|---|---|
| `make test-unit` | Unit tests, no external services. |
| `make test-race` | Unit tests under the race detector. |
| `make test-integration` | Integration tests against real Postgres. |
| `make test-contract` | Module contract suite + a scratch external-consumer build. |
| `make test-security` | Security-critical tests: authz, RLS, secrets, redaction, unsafe-config. |
| `make test-fuzz` | Fuzz the filter-DSL parser and cursor decoder. |
| `make bench` / `make bench-budget` | Hot-path benchmarks / enforce performance budgets. |
| `make coverage` | Unit coverage report. |
| `make test` | All currently-available suites. |

## The authoritative gate

Two commands, one authoritative:

```bash
make ci             # host CI: vet + boundary lint, unit, race, perf budgets, build (golangci-lint = make lint-new / hosted CI)
make ci-container   # runs `make ci` INSIDE the toolbox container — the authoritative gate
```

**`make ci-container` is the gate that counts.** It runs in the container where a real PostgreSQL is
available and sets `WOWAPI_REQUIRE_DB=1`, which turns DB-backed tests from *skip-if-no-DB* into
*fail-if-no-DB*. This closes the "green-but-hollow" hole where a suite passes only because every DB test
silently skipped.

```bash
# bring up local infra first if running DB tests on the host
make up                        # postgres + minio + mailpit + tools runner
WOWAPI_REQUIRE_DB=1 make test-integration
```

## `testkit` — the integration harness

### Isolated per-test databases

`testkit.NewDB(t)` gives each test its **own** database, cloned from a migrated template, with production-
shaped role separation (`testkit/db.go`):

```go
func TestWidgetCreate(t *testing.T) {
    h := testkit.NewDB(t)   // exclusive DB, auto-cleaned when the test ends

    // h.Admin    *pgxpool.Pool     — owner creds for fixtures/DDL
    // h.Runtime  *pgxpool.Pool     — SET ROLE app_rt (what production sees; RLS forced)
    // h.Platform *pgxpool.Pool     — SET ROLE app_platform (kernel background work)
    // h.TxM      database.TxManager — tenant tx manager over Runtime
    // h.PlatformTxM database.TxManager
    // h.Name     string            — the per-test database name

    ctx := database.WithTenantID(context.Background(), tenantID)
    err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
        _, err := db.Exec(ctx, `INSERT INTO widgets (id, tenant_id, title) VALUES ($1, app_tenant_id(), $2)`, id, "hi")
        return err
    })
    require.NoError(t, err)
}
```

Behavior when no database is available:

- **Default:** `NewDB` **skips** the test (so `go test ./...` works on a laptop without Postgres).
- **With `WOWAPI_REQUIRE_DB=1`:** it **fails** instead of skipping — this is what the container gate sets,
  so DB coverage can never silently evaporate.

Because `Runtime` connects as `app_rt` with forced RLS, tenant-isolation tests exercise the *real*
security boundary — a cross-tenant read fails in the test exactly as it would in production.

### Authenticated requests

`testkit.NewTokenIssuer()` mints real RS256 tokens from a local keypair, so handler tests go through the
actual auth gate (`testkit/auth.go`):

```go
ti  := testkit.NewTokenIssuer()
ks  := ti.KeySource()                                   // wire into your token verifier
tok := ti.Issue(subjectID, tenantID, capacityID,
    testkit.WithAudience("myapp"), testkit.WithExpiry(time.Hour))
req.Header.Set("Authorization", "Bearer "+tok)
```
Options: `WithIssuer`, `WithAudience`, `WithExpiry`, `WithGrantID`, `WithAMR`.
`WithImpersonator`/`WithBreakGlass` are retained for backwards compatibility but no longer drive
`authz.Actor` fields unless a matching `grant_id` is resolved server-side.

### The module contract suite

The single most valuable test for a new module — `testkit.RunModuleContract(t, m)` registers your module on
a fresh kernel and asserts the invariants every module must uphold (`testkit/contract.go`):

```go
func TestWidgetsContract(t *testing.T) {
    testkit.RunModuleContract(t, &widgets.Module{})
}
```

It verifies, among other things, that the module:

- **boots and validates** on an empty config namespace,
- has **idempotent** migrations and seeds (re-applying is a no-op),
- **enforces RLS** on its tables,
- **rejects invalid/unknown config keys**.

If your module passes the contract suite, it composes cleanly into any product.

## Writing a good test (TDD)

The house style is test-first with real integration:

1. Write the failing test against `testkit.NewDB`/`RunModuleContract`.
2. Watch it fail (`go test ./internal/modules/widgets/...`).
3. Implement the migration + handler until it passes.
4. Add a security test if the change touches tenancy/authz/secrets.
5. Gate with `make ci-container`.

Mock only across a boundary you genuinely can't stand up locally — the whole point of the isolated-DB
harness is that you rarely need to.

## Regression & fuzz testing

- **Regression:** when you fix a bug, add a test that reproduces it first (the reversibility drill caught a
  real kernel migration defect this way — see [Database & migrations](database-migrations.md#the-reversibility-drill-do-not-skip)).
- **Fuzzing:** `make test-fuzz` fuzzes the filter-DSL parser and cursor decoder — the two places that parse
  untrusted input. Extend the corpus when you touch either.
- **Security matrix:** `make test-security` is the regression home for authz/RLS/secret-redaction invariants.

## Common problems

| Symptom | Cause | Fix |
|---|---|---|
| DB tests "skipped" in CI | `WOWAPI_REQUIRE_DB` not set / no DB | Use `make ci-container`; it sets the flag and provides Postgres. |
| `NewDB` skips locally | no admin DSN configured | `make up`, then export the admin DSN, or just run non-DB suites. |
| Cross-tenant test "passes" but shouldn't isolate | queried via `Admin` (owner) not `Runtime` | Use `h.Runtime`/`h.TxM` — RLS only binds the `app_rt` role. |
| Contract test fails on config | module reads an unknown key | Only read your `modules.<name>.*` namespace; unknown keys are rejected. |
| Flaky isolation across tests | shared state | Each test must use its own `NewDB` handle; don't share pools across tests. |

Next: [Build & deploy](build-deploy.md) · [CLI reference](cli-reference.md).
