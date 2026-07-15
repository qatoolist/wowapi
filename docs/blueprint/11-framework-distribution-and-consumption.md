# 11 — Framework Distribution & Consumption

`wowapi` is developed, versioned, and consumed as a **third-party Go framework dependency**.
Product applications (a future housing society product, a school product, a facility product…)
live in **their own repositories**, add wowapi with `go get github.com/qatoolist/wowapi@vX.Y.Z`,
and register their domain modules against wowapi's public module SDK. The framework repository
contains no real product modules — only neutral standalone examples and private contract-test fixtures.

## 1. Distribution model

- **One Go module:** `github.com/qatoolist/wowapi`, semver-tagged. The repository is on the **stable
  v1 line** (`v1.0.0`, `v1.1.0` tagged): the public surface follows additive-only changes within v1.
  `v2+` uses `/v2` module paths per Go convention for any incompatible change. See
  [Versioning & stability](../../README.md#versioning--stability) and the
  [upgrade & deprecation policy](../operations/upgrade-and-deprecation-policy.md).
- **Stability contract:** public root packages (`kernel`, `module`, `app`, `adapters`, `testkit`,
  `migrations`, and `cmd/wowapi`) follow semver. Deprecations keep working for ≥1 minor release
  with a `// Deprecated:` notice and a changelog entry. `/internal` is private; `/examples` are
  non-contractual sample apps and preferably nested modules with their own `go.mod`.
- **What ships in the dependency:** kernel packages, module SDK, app composition helpers, adapters,
  testkit, embedded kernel migrations, embedded generator templates, and the `wowapi` CLI source.
- **Version pinning & upgrades (product side):** pin via `go.mod`; upgrade = bump version → read
  changelog → `go build ./...` → run `cmd/migrate` (new kernel migrations apply automatically) →
  seed sync runs at boot → re-run module contract tests. Kernel migrations are version-locked to
  the dependency, so schema and code can never drift apart.

## 2. Public vs private surface

| Path | Visibility | Contents |
|---|---|---|
| `wowapi/kernel/...` | **public** | primitives + service contracts modules import: `model`, `errors`, `validation`, `pagination`, `filtering`, `httpx`, `database` (TxManager/TenantDB), `tenant`, `auth`, `authz`, `policy`, `resource`, `relationship`, `workflow`, `rules`, `audit`, `outbox`, `jobs`, `document`, `notify`, `webhook`, `integration`, `secrets`, `config`, `health` |
| `wowapi/module` | **public** | `Module`, `Context`, registries, lifecycle contracts |
| `wowapi/app` | **public** | composition helpers: `App.New`, `App.Register`, `App.Boot(ctx, *kernel.Kernel, ...)` (product constructs the kernel via `kernel.New` and passes it in), plus the free function `StartWorker` |
| `wowapi/adapters/...` | **public** | postgres, s3, smtp, sms, push, oidc, secrets, scanner |
| `wowapi/testkit` | **public** | fixtures, fakes, asserts, `RunModuleContract` — importable by product test code |
| `wowapi/migrations` | **public** | kernel goose migrations as `embed.FS` via `migrations.Kernel()` |
| `wowapi/cmd/wowapi` | **public (installable)** | the CLI, with embedded generator templates |
| `wowapi/internal/...` | **private** | implementation guts (pg stores, engine internals, outbox relay, evaluator impl) — wired by `/app`, unreachable from product code (compiler-enforced) |
| `wowapi/internal/testmodules/...` | **private fixture** | neutral modules (`requests`) used by framework contract tests; not importable by consumers |
| `wowapi/examples/...` | **non-contractual examples** | standalone sample product apps or docs examples, preferably nested Go modules; never imported by `/kernel`, `/module`, `/app`, or `/adapters` |

The rule that forced this split: **Go forbids importing another module's `internal/` packages.**
Any type a product module must reference — interfaces, response envelopes, embedded structs,
`module.Context` — must live in a public package. `/internal` remains exactly for what consumers
must *not* couple to. (Public packages at the repo root is the idiomatic shape — cf. chi, river.)

## 3. Product application repository

```text
example.com/acme-ops                    # separate product repo (society/school/… follow the same shape)
  go.mod                                # require github.com/qatoolist/wowapi vX.Y.Z
  /cmd/api  /cmd/worker  /cmd/migrate   # thin mains over wowapi/app helpers
  /internal/modules/requests/           # product modules — the 06-module-sdk template, unchanged in shape
  /internal/modules/assets/
  /api/openapi                          # merged spec output (wowapi openapi merge)
  /configs  /deployments  /scripts
```

Product modules live under the *product's* `internal/` — correct, because only the product's own
`cmd/*` imports them. Domain code never enters the framework repo; the framework never imports a
product package (structurally impossible: separate repositories, dependency points one way).

### Usage flow for a new product backend

1. `go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z`
2. `wowapi init --module example.com/acme-ops` — scaffolds the repo
   above (go.mod with pinned wowapi, mains, compose, Makefile wrappers, CI stub). The command may
   offer prompts, but all inputs must also have flags so CI/bootstrap scripts are repeatable.
   A released CLI pins its own version in the scaffolded go.mod; to override, pass
   `--framework-version vX.Y.Z` (verified via `go list -m`), or `--local-framework /path/to/wowapi`
   for a dev-mode scaffold against a local checkout (emits a `replace` directive).
3. `wowapi new-module --name requests` — scaffolds `/internal/modules/requests` (template in [06-module-sdk.md](06-module-sdk.md)).
4. Implement domain/service/store code; embed assets (migrations, seeds, OpenAPI fragment) via
   `embed.FS` handed to `module.Context` in `Register` — the embedded-asset methods are public contracts.
5. Wire the mains:

<!-- doc-example: illustrative -->
```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/qatoolist/wowapi/app"
    "github.com/qatoolist/wowapi/kernel"

    "example.com/acme-ops/internal/appcfg"      // product-owned Config (embeds config.Framework), scaffolded by wowapi init
    "example.com/acme-ops/internal/modules/assets"
    "example.com/acme-ops/internal/modules/requests"
)

func main() {
    ctx := context.Background() // production main wraps this with SIGTERM/SIGINT handling
    log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    cfg := appcfg.MustLoad()    // loads + validates configs/{base,<env>}.yaml
    k, err := kernel.New(cfg.Framework, log, kernel.Deps{ /* adapters: metrics, tracer, secrets, storage, … */ })
    if err != nil { die(err) }

    a := app.New()
    a.Register(requests.Module{}, assets.Module{})
    booted, err := a.Boot(ctx, k, cfg.ModuleNamespaces() /* modules.* config, product-defined */)
    if err != nil { die(err) }
    // cmd/api serves booted.Router; cmd/worker calls app.StartWorker(ctx, booted, opts);
    // cmd/migrate applies booted.Migrations + booted.Seeds — same module list, same Boot call, every time.
}
```

The marker directly above each Go fence states its contract. A
`<!-- doc-example: compile -->` fence is normative current API and must be complete, standalone Go
source; `<!-- doc-example: illustrative -->` identifies signatures or product-specific pseudo-code
that is intentionally not compiled.

Run `make docs-check` before submitting documentation changes. When the authoritative AR-03
ApplicationModel projection changes intentionally, regenerate
`docs/reference/application-model.md` with
`go run ./internal/tools/docexamples -write-reference`, then run `make docs-check` to prove the
generated table byte-matches the export.

This minimal framework-only example is compile-checked:

<!-- doc-example: compile -->
```go
package main

import "github.com/qatoolist/wowapi/app"

func main() {
    application := app.New()
    if err := application.Validate(); err != nil {
        panic(err)
    }
}
```

6. Configure per environment: `wowapi init` seeds `configs/{base,local}.yaml`; add further
   overlays (`dev`/`stage`/`prod`.yaml) per environment (secret references only — never raw
   secrets); `wowapi config validate --env prod` runs in CI
   ([12-configuration-and-deployment.md](12-configuration-and-deployment.md)).
7. `go build ./...`, run `cmd/migrate`, start `cmd/api`/`cmd/worker`. A future society product is
   just another such repo registering `society.Module{}` — the framework is untouched.

## 4. Combined migrations (kernel + product modules)

The product's `cmd/migrate` calls `App.Boot` (same call as `cmd/api`/`cmd/worker`) and applies
`Booted.Migrations` — a `map[string]fs.FS` keyed by module name — composed in order:

1. **Kernel migrations** from `wowapi/migrations` (`migrations.Kernel()` embed.FS) — always first.
2. **Product module migrations** — each module's embedded goose dir, ordered by the module
   `DependsOn` graph (topo-sort), then by filename within a module.

Goose history rows are prefixed per source (`wowapi/000_bootstrap`, `requests/0001_create_requests`),
so kernel and module histories coexist in one `goose_db_version` table and re-runs are idempotent.
An upgraded wowapi that ships new kernel migrations applies them on the next migrate run, before
any module migration that might depend on them.

## 5. The `wowapi` CLI — out of the box

**Primary workflow (macOS, Linux, Windows):**

```text
go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z

wowapi init --module example.com/acme-ops          # scaffold a product repo; flags make it repeatable
wowapi new-module --name requests                  # scaffold a module
wowapi gen crud --module internal/modules/requests --resource request
wowapi migrate create --dir internal/modules/requests/migrations --name create_requests
wowapi seed validate --module requests --dir internal/modules/requests/seeds
wowapi openapi merge
wowapi lint boundaries
wowapi version

wowapi config validate --env prod
wowapi config doctor
wowapi config print --redacted
wowapi config diff --from dev --to prod
wowapi config schema
wowapi deploy render --env prod                    # see 12-configuration-and-deployment.md
```

- **Templates are embedded** (`embed.FS`) in the CLI binary — no cloning or copying of the
  framework repo, ever. Generators **write into the consuming product repository**; output is
  committed, reviewed, editable; regeneration is optional and diffable; business logic is never
  generated (the [05](05-http-and-persistence.md) §4 rules stand).
- **Version alignment:** the CLI reads the product's `go.mod`, compares the `wowapi` requirement to
  its own build version, and warns on mismatch (`wowapi version` prints both). Keep them equal so
  generated code matches the imported API.
- **Release binaries** (goreleaser) are published per tag for teams that don't build from source.
- **CI usage:** `wowapi seed validate`, `wowapi openapi merge`, `wowapi lint boundaries`
  run in product CI. For tightly pinned CI jobs — or as a no-install fallback —
  `go run github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z <cmd>` works, but it is not the primary
  developer experience.
- **Makefile wrappers** in product repos (`make new-module` → `wowapi new-module …`) are optional
  sugar; the CLI is the source of truth.

## 6. Boundary rules (dependency edition)

1. The public package graph is acyclic: `kernel` defines contracts/primitives and imports no
   `module`, `app`, `adapters`, `testkit`, examples, or product packages; `module` imports kernel
   contracts; `adapters` implement kernel ports; `app` sits at the top and wires everything.
2. `wowapi` never imports any product package (separate repos; dependency arrow points one way).
3. Product modules import **only public wowapi packages**; `wowapi/internal/...` is blocked by the
   Go compiler — and `wowapi lint boundaries` additionally AST-checks for it (catches `replace`
   -directive workarounds).
4. Product modules never import another module's internals — declared ports, registered events,
   and public module contracts only (`wowapi lint boundaries` enforces the module import rules
   inside product repos, configured via `wowapi.yaml`).
5. Framework packages and docs stay free of product-domain vocabulary — the denylist lint from
   [00-overview.md](00-overview.md) §5 runs in the framework repo's CI.
6. `/examples` are standalone, non-contractual sample apps and must not be imported by `/kernel`,
   `/module`, `/app`, or `/adapters` (framework-repo lint rule).
7. Private contract fixtures live under `/internal/testmodules`, not public examples.
8. `testkit.RunModuleContract` is the consumer-side guarantee: every product module passes the same
   contract suite the framework's private fixture passes.

## 7. Acceptance criteria for the distribution model

1. A blank product repo + `go get wowapi` + the §3 flow builds a working API binary.
2. `wowapi new-module` generates a module without copying framework files manually.
3. A product module imports only the public SDK and registers routes, permissions, rules,
   workflows, jobs, events, seeds, migrations, and OpenAPI fragments.
4. Product modules live outside the framework repository; the framework repo contains only private
   test fixtures and non-contractual standalone examples.
5. No consumer-facing contract lives under `wowapi/internal`; public vs private surface matches §2.
6. `wowapi/testkit` is importable and usable from an external product repo.
7. Kernel + product migrations run together from the product's `cmd/migrate` (§4 ordering test).
8. `wowapi lint boundaries` catches domain leakage into the framework and illegal imports in
   product repos.
9. Public-package dependency lint proves the framework graph is acyclic.
10. All generator/validation/merge/lint commands work from the installed CLI with embedded templates.
11. Multiple products can consume the same wowapi version without forking or modifying the core.
