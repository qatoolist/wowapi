# wowapi — User Guide

The practical guide to installing, configuring, using, extending, testing, and troubleshooting `wowapi`.
Everything here is grounded in the actual code, CLI, `Makefile`, and configuration in this repo — where
something is missing or manual, it is called out as a **gap** rather than invented.

New to the project? Read in order:

1. **[Getting Started](getting-started.md)** — prerequisites, install the CLI, scaffold a product, first run.
2. **[Concepts & Architecture](architecture.md)** — kernel/module/app, tenancy + RLS, the request path,
   async platform, folder structure.
3. **[Configuration](configuration.md)** — layered config, `WOWAPI__*` env vars, `secretref://` secrets, DSNs.
4. **[Building & extending modules](modules.md)** — the module SDK, `module.Context`, `new-module`,
   `gen crud`, a full worked example, ports.
5. **[Database & migrations](database-migrations.md)** — RLS/roles, writing migrations, the reversibility drill.
6. **[Authentication & authorization](auth.md)** — actors, `RouteMeta`, the route gate, API keys, step-up.
7. **[Validation & error handling](validation-errors.md)** — request decoding, the error taxonomy, problem details.
8. **[Testing](testing.md)** — `testkit`, the suites, the authoritative gate, regression, fuzzing.
9. **[Build & deploy](build-deploy.md)** — binaries, `deploy render`, the compose stack, the checklist.
10. **[CLI reference](cli-reference.md)** — every `wowapi` subcommand + `Makefile` target.
11. **[Troubleshooting & FAQ](troubleshooting-faq.md)** — common errors, fixes, and answers.

Deeper material lives elsewhere: [design rationale](../blueprint/README.md) (the blueprint), [operations
runbooks](../operations/deployment-checklist.md), [design decisions](../implementation/decisions.md), and
the [contributor working layer](../working/README.md).

## Two mental models

- **Framework consumer** (most readers): you build a product in *your own repo* that depends on
  `github.com/qatoolist/wowapi`. The CLI scaffolds it; you write modules. Start with
  [Getting Started](getting-started.md).
- **Framework contributor**: you work on *this* repo. Use `make ci-container` as the gate and follow the
  [working layer](../working/README.md).

## Conventions in this guide

- `$ command` blocks are copy-paste-ready; product-repo commands assume you ran `wowapi init` first.
- "Framework repo" = a clone of `github.com/qatoolist/wowapi`. "Product repo" = what `wowapi init` scaffolds.
- **Gaps/manual steps are labeled explicitly** so you never mistake a recommendation for a shipped feature.
