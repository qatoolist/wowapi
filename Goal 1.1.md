------- GOAL ACCOMPLISHED - DO NOT REWORK --------

# Goal 1.1: Refine WowAPI As A Third-Party Framework Dependency

You are the architect for `wowapi`.

Review the existing `Goal.md` and every file under `docs/blueprint/`. The current blueprint is strong for a reusable, domain-agnostic backend framework, but it still reads in places like product/domain modules live inside the same repository as the framework. Make small, surgical refinements so the architecture is explicitly designed for this usage model:

`wowapi` will be developed, versioned, and consumed as a third-party Go framework dependency by separate product projects.

Product-specific applications, such as a future housing society product, must live in their own repositories and import `wowapi`. The framework core must remain isolated, domain-neutral, and reusable without deleting or moving product code.

## Required Clarification

Make the blueprint explicit that:

1. `wowapi` is the framework repository and Go module.
2. Product applications use `wowapi` as a dependency, for example with `go get github.com/qatoolist/wowapi@vX.Y.Z`.
3. Product modules live in the consuming application repository, not in the `wowapi` framework repository.
4. `wowapi` must expose a stable public module SDK, kernel API surface, testkit, migration/seed tooling, and generator tooling that product repositories can import or execute.
5. Anything under Go `internal/` cannot be imported by external projects, so the blueprint must not place consumer-facing APIs only under `/internal/kernel`, `/internal/platform`, or `/internal/testkit`.
6. The framework repository may still use `/internal/...` for private implementation details, but public extension contracts must live in importable packages.

## Refactor The Blueprint Where Needed

Do not redesign the framework. Preserve the existing architecture and only adjust the parts needed to make dependency-based usage correct.

Update the relevant blueprint sections so they distinguish clearly between:

- The `wowapi` framework repository.
- A product application repository that imports `wowapi`.
- Public packages exposed by `wowapi`.
- Private implementation packages inside `wowapi/internal`.
- Product modules owned by the consuming app.
- Non-contractual standalone examples that may live under `wowapi/examples`.
- Private contract-test fixtures that should live somewhere non-public such as `wowapi/internal/testmodules` or `wowapi/testdata`, never as real product modules inside the framework core.

## Expected Package / Repository Guidance

Recommend a practical Go layout for `wowapi` as a reusable framework dependency.

Address whether the public API should look like this or a better equivalent:

```text
/kernel/...          # public framework primitives and service interfaces safe for modules to import
/module/...          # public Module, Context, registries, lifecycle contracts
/app/...             # public composition helpers for building api/worker/migrate binaries
/adapters/...        # optional public adapters: postgres, s3, oidc, smtp, etc.
/testkit/...         # public testing utilities for consumer product modules
/codegen/...         # public generator library only if useful
/cmd/wowapi          # CLI shipped out of the box
/internal/...        # private implementation details not imported by product apps
/internal/testmodules/... # private neutral fixtures for framework contract tests
/examples/...        # standalone neutral examples, not framework dependencies or public API
/docs/blueprint/...  # architecture docs
```

If you choose a different layout, explain why it is better. The key requirement is that consuming applications can import the needed packages without depending on `internal` packages.

## Product Application Usage Model

Add a recommended usage flow for a new product project.

It should answer:

1. How a product app initializes a new backend that depends on `wowapi`.
2. How product modules register with the framework.
3. How product modules embed their own migrations, seeds, OpenAPI fragments, workflows, rule points, permissions, roles, jobs, and event handlers.
4. How the app builds its own `cmd/api`, `cmd/worker`, and `cmd/migrate` binaries using `wowapi`.
5. How the product app keeps domain code isolated from the framework dependency.
6. How version pinning and upgrades should work.

Include a short example such as:

```go
import (
    "github.com/qatoolist/wowapi/app"
    "github.com/qatoolist/wowapi/kernel/config"

    "example.com/society/internal/modules/notices"
    "example.com/society/internal/modules/units"
)
```

Do not make the housing society product the main example. Use neutral modules such as `requests`, `assets`, or `approvals`, and mention society only as a future consuming product.

## CLI And Generators Must Ship Out Of The Box

The current blueprint already mentions generators and scaffolding. Strengthen it so these tools are available directly from the `wowapi` dependency, without cloning or copying the framework repository.

Specify a normal installable CLI as the preferred workflow:

```text
go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z

wowapi init --module example.com/acme-ops --wowapi-version vX.Y.Z
wowapi new-module requests
wowapi gen crud --module requests --resource request
wowapi migrate create --module requests --name create_requests
wowapi seed validate
wowapi openapi merge
wowapi lint boundaries
```

The CLI should be installable and runnable on macOS, Linux, and Windows. It should also be distributable through release binaries for teams that do not want every machine to build from source.

Mention that `go run github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z ...` is still acceptable as a no-install fallback or for tightly pinned CI jobs, but it should not be the primary developer experience.

Also support optional Makefile wrappers in product apps, but the source of truth should be the framework CLI.

Clarify that:

- Generator templates are bundled in the framework using embedded assets.
- The installed CLI performs all generator, scaffold, validation, OpenAPI, migration helper, and boundary-lint jobs.
- Generators write into the consuming product repository, not into `wowapi`.
- Generated code is committed, reviewed, and editable.
- Business logic is never generated.
- Regeneration is optional and diffable.
- The installed CLI version should normally match the product app's `wowapi` dependency version, and the CLI should warn on version mismatch when possible.
- The CLI should be usable in CI for seed validation, OpenAPI merging, and boundary checks.

## Migrations, Seeds, And Module Assets

Refine the module SDK so external product modules can provide embedded assets through public contracts.

The architect should define how modules expose:

- Embedded migrations.
- Embedded seeds.
- Embedded OpenAPI fragments.
- Workflow definitions.
- Rule point defaults.
- Permission and role catalogs.
- Notification templates.
- Test fixtures if applicable.

The migration runner in a consuming app must be able to combine:

1. Kernel migrations shipped by `wowapi`.
2. Product module migrations embedded in the product app.
3. Ordered module migrations based on declared dependencies.

## Boundary Rules

Add explicit dependency-boundary rules:

- `wowapi` must not import any product application package.
- Product modules may import only public `wowapi` packages.
- Product modules must not import `wowapi/internal/...`.
- Product modules must not import another product module's internals.
- Product modules communicate through declared ports, registered events, and public module contracts.
- Framework package names and docs must avoid product-domain terms.
- Standalone examples in the framework repo must be clearly marked as non-contractual examples.
- Contract-test fixtures must live in a private or otherwise non-public location such as `internal/testmodules` or `testdata`.

Add lint or contract-test requirements that enforce these rules.

## Documentation Updates To Make

Update or add blueprint content in the smallest useful set of files. At minimum review these:

- `docs/blueprint/00-overview.md`
- `docs/blueprint/04-project-and-primitives.md`
- `docs/blueprint/06-module-sdk.md`
- `docs/blueprint/08-testing-and-tooling.md`
- `docs/blueprint/10-delivery.md`
- `docs/blueprint/README.md`

Add a new section or file if needed, such as:

```text
docs/blueprint/11-framework-distribution-and-consumption.md
```

That section should explain how external product projects consume `wowapi`.

## Acceptance Criteria

The refined blueprint is successful when it clearly states and enables all of the following:

1. A blank product repository can add `wowapi` as a Go dependency and build an API binary.
2. A product repository can generate a new module using the `wowapi` CLI without copying framework files manually.
3. A product module can import the public module SDK and register routes, permissions, rules, workflows, jobs, events, seeds, migrations, and OpenAPI fragments.
4. Product modules live outside the framework repository.
5. The framework repository contains no real product modules, only standalone non-contractual examples or private test fixtures.
6. The public API surface is clearly separated from private framework internals.
7. No consumer-facing contract depends on Go `internal` packages.
8. The testkit is available to external product modules.
9. Kernel migrations and product module migrations can run together from a consuming app.
10. Boundary linting catches product-domain leakage into the framework and illegal imports in product apps.
11. Generator commands are available out of the box from `cmd/wowapi`.
12. The framework can be reused by multiple products without modifying or forking the core.

## Tone And Scope

Keep the changes practical and implementation-oriented.

This is not a request to rebuild the architecture. Treat it as a framework distribution, package-boundary, and developer-experience refinement pass.

Preserve the existing choices unless they conflict with third-party dependency usage.

Be explicit where the current blueprint must change because `/internal` packages are not importable by external Go modules.
