# CLI & Make Reference

Complete reference for the `wowapi` CLI (`internal/cli/`) and the root `Makefile`. Every command and flag
below is taken from the source — if a subcommand you expect isn't here, it isn't implemented (see
[Gaps](#gaps--not-yet-implemented)).

## `wowapi` CLI

Get the CLI: `go install github.com/qatoolist/wowapi/cmd/wowapi@latest` (or build from a clone — see
[Getting Started](getting-started.md#step-1--get-the-wowapi-cli)).

### Global

| Command | Flags | Purpose |
|---|---|---|
| `wowapi version` | — | Print CLI version + the `wowapi` dependency version. |
| `wowapi help` / `-h` / `--help` | — | Usage. |

### Scaffolding

| Command | Flags | Purpose |
|---|---|---|
| `wowapi init` | `--module` **(req)**, `--name`, `--dir` (`.`), `--force` | Scaffold a product repository. |
| `wowapi new-module` | `--name` **(req)**, `--dir` (`internal/modules`), `--force` | Scaffold a module package. |
| `wowapi gen crud` | `--module` **(req)**, `--resource` **(req)**, `--fields` (`name:type,…`), `--force` | Generate CRUD scaffolding. |
| `wowapi gen` subsystem commands | `rule`, `workflow`, `event-handler`, `recurring-job`, `document-flow`, `notification`, or `webhook`; each takes `--module`, `--name`, `--force` | Generate a boot-wired subsystem declaration. |

### Config

| Command | Flags | Purpose |
|---|---|---|
| `wowapi config validate` | `--dir` (`configs`), `--base`, `--env`, `--env-prefix` (`WOWAPI__`) | Load + validate; exit 0 OK / 1 invalid (prints every problem). |
| `wowapi config print` | `--dir`, `--base`, `--env`, `--env-prefix`, `--redacted` **(req)** | Print effective config as JSON, secrets redacted. |
| `wowapi config doctor` | `--dir`, `--base`, `--env`, `--env-prefix` | Per-key provenance table + fingerprint. |
| `wowapi config diff` | `--from` **(req)**, `--to` **(req)**, `--dir`, `--env-prefix` | Redacted effective-config diff between two environments. |
| `wowapi config schema` | — | Emit JSON Schema derived from struct tags. |
| `wowapi config capacity` | `--dir`, `--base`, `--env`, `--env-prefix` | Check the concurrency capacity budget; exit 0 within budget/not configured, exit 1 oversubscribed. |

### Migrations, seeds, OpenAPI

| Command | Flags | Purpose |
|---|---|---|
| `wowapi migrate create` | `--dir` (`migrations`), `--name` **(req)** | Scaffold the next-numbered goose migration. |
| `wowapi seed validate` | `--dir` (`seeds`), `--module` **(req)** | Load + validate a module seed bundle (no database). |
| `wowapi seed sync` | `--module name=dir` **(req, repeatable)**, `--dry-run` | Load one or more modules' seed bundles and upsert them into a real database (`DATABASE_URL`, connects as `app_platform`). Idempotent; computes a content hash so re-runs with an unchanged manifest are true no-ops. `--dry-run` prints a change plan without writing. See [Database & Migrations § Seeds](database-migrations.md#seeds-declarative-yaml-catalogs). |
| `wowapi i18n validate` | `--dir` (`locales`), `--default-locale` (`en`), `--supported` (`en`), `--strict` | Load + validate a product's locale catalogs (no database): coverage, `kernel.*` ownership, intra-layer duplicates, placeholder drift. Exit 0 OK / 1 with every problem listed. See [Validation & error handling § Localizing responses](validation-errors.md#localizing-responses-i18n). |
| `wowapi openapi merge` | `--dir` (`.`), `--title` (`wowapi API`), `--version` (`0.0.0`), `--out` | Merge OpenAPI 3.1 fragments into one document. |

> Applying migrations at runtime is the **product** `cmd/migrate` (`go run ./cmd/migrate up` / `make
> migrate-up`), not the `wowapi` CLI. The CLI's `migrate` subcommand only *creates* migration files. The
> generated `cmd/migrate up` also runs seed sync automatically (GAP-003) — `wowapi seed sync` is the
> standalone equivalent for re-syncing catalogs without a full migrate run.

### Boundaries & deployment

| Command | Flags | Purpose |
|---|---|---|
| `wowapi lint boundaries` | `--pkgs` (`./...`) | Module isolation + layering (import-law) check. |
| `wowapi lint lifecycle` | — | Print + lint the static provider/lifecycle manifest (`kernel/lifecycle`): catches scope leaks, raw pools reaching modules, tenant-scoped values escaping their transaction, migrate-only services wired into API/worker runtime, and missing providers/cycles. Exit 0 clean / 1 with every violation listed. |
| `wowapi deploy render` | `--format` (`compose`\|`env`), `--name` (`app`), `--image` (`app:latest`), `--env` (`local`\|`dev`\|`stage`\|`prod`), `--out` | Render a deployment manifest. |

### Dead-letter queue operations

Inspect and recover failed async work (`internal/cli/dlq_cmd.go`):

| Command | Flags/args | Purpose |
|---|---|---|
| `wowapi dlq jobs list` | `--limit` (50) | List discarded jobs. |
| `wowapi dlq jobs inspect <id>` | — | Full payload of one discarded job. |
| `wowapi dlq jobs replay <id>` | — | Requeue a discarded job. |
| `wowapi dlq jobs discard <id>` | — | Permanently delete a discarded job. |
| `wowapi dlq events list` | `--limit` (50) | List dead events. |
| `wowapi dlq events inspect <uuid>` | — | Full payload of one dead event. |
| `wowapi dlq events replay <uuid>` | — | Re-dispatch a dead event. |
| `wowapi dlq events discard <uuid>` | — | Permanently delete a dead event. |

## `Makefile` targets (framework repo)

### Setup & infra

| Target | Purpose |
|---|---|
| `make setup` | One-time dev setup (install tools + `go mod download`). |
| `make tools` | Install host dev tools (golangci-lint). |
| `make up` / `make down` / `make reset` | Start / stop / stop-and-wipe the local stack. |
| `make logs` | Tail infra logs. |
| `make shell` | Shell into the toolbox container (repo mounted at `/src`). |
| `make db-shell` | `psql` into the local postgres. |
| `make migrate` | Apply kernel migrations to the local compose DB. |

### Format, lint, boundaries

| Target | Purpose |
|---|---|
| `make fmt` | `gofmt` all Go files. |
| `make lint` | `golangci-lint` (falls back to `go vet`). |
| `make lint-boundaries` | Import-law + vocabulary + `Reveal()` boundary lint. |
| `make lint-lifecycle` | Static provider/lifecycle manifest lint (`wowapi lint lifecycle`; backlog B9). |

### Test

| Target | Purpose |
|---|---|
| `make test` | All currently-available suites. |
| `make test-unit` | Unit tests (no external services). |
| `make test-race` | Unit tests with the race detector. |
| `make test-integration` | Integration tests against real Postgres. |
| `make test-contract` | Module contract + scratch external-consumer suite. |
| `make test-security` | authz / RLS / secrets / redaction / unsafe-config. |
| `make test-fuzz` | Fuzz the filter-DSL parser and cursor decoder. |
| `make coverage` | Unit coverage report. |
| `make golden-consumer` | Install a versioned CLI, generate/boot the two-module eight-subsystem consumer, replay tagged v1.1.0→local release candidate, and verify the RLS census. |

### Bench, gen, build, CI

| Target | Purpose |
|---|---|
| `make bench` / `make bench-budget` | Hot-path benchmarks / enforce perf budgets. |
| `make gen` / `make gen-crud` / `make new-module` | Run generators / CRUD scaffold / module scaffold. |
| `make openapi` | Merge OpenAPI fragments. |
| `make config-validate` / `make config-doctor` | Validate config / show provenance. |
| `make build` | Build all packages + the CLI. |
| `make ci` | Full local CI: vet, boundary lint, lifecycle lint, unit, race, perf budgets, build (golangci-lint = `make lint-new`). |
| **`make ci-container`** | Run `make ci` inside the toolbox container — **the authoritative gate**. |

> Product repos get a **smaller** generated `Makefile` — `build`, `test`, `lint`, `migrate-up`
> (`go run ./cmd/migrate up`), `migrate-down`. See [Getting Started](getting-started.md).

## Gaps / not yet implemented

Documented honestly so you don't reach for something that isn't there:

- **`wowapi deploy validate`** — does not exist; only `deploy render`.
- **`make seed`** (framework repo) — placeholder; seeding is driven through modules + the app at boot.

Next: [Troubleshooting & FAQ](troubleshooting-faq.md).
