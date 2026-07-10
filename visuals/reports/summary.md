# Code Graph Report

## Repository Summary

- **Module:** github.com/qatoolist/wowapi
- **Repository root:** /Users/qatoolist/go_home/src/github.com/qatoolist/wowapi
- **Go version:** go1.26.4
- **Tool:** wow-codegraph 1.0.0 (schema v1)
- **Generated at:** 2026-07-05T19:46:46Z

## Summary Counts

- **Packages:** 51
- **Files:** 145
- **Types:** 358
- **Structs:** 267
- **Interfaces:** 33
- **Named types:** 58
- **Functions:** 471
- **Methods:** 442
- **Graph nodes:** 6044
- **Graph edges:** 15722

## Call Graph Summary

- **Call edges:** 2062

See [calls.md](calls.md) for the full call relationship summary.

## Highly Connected Symbols

### Top fan-out (calls the most other symbols)

- `github.com/qatoolist/wowapi/kernel.New` — 38 outgoing call(s)
- `github.com/qatoolist/wowapi/app.App.Boot` — 32 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/document.Service.ConfirmUpload` — 29 outgoing call(s)
- `github.com/qatoolist/wowapi/testkit.RunModuleContract` — 29 outgoing call(s)
- `github.com/qatoolist/wowapi/app.registerMaintenance` — 28 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/config.binder.bindField` — 26 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/webhook.Service.deliverToEndpoint` — 26 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/authz.engine.Evaluate` — 24 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/audit.Writer.Record` — 23 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/notify.Service.Send` — 22 outgoing call(s)
- `github.com/qatoolist/wowapi/internal/cli.runApikey` — 20 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/notify.Service.SendPending` — 20 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/workflow.Runtime.SweepSLA` — 19 outgoing call(s)
- `github.com/qatoolist/wowapi/kernel/audit.Writer.Verify` — 18 outgoing call(s)
- `github.com/qatoolist/wowapi/testkit.NewDB` — 18 outgoing call(s)

### Top fan-in (called by the most other symbols)

- `github.com/qatoolist/wowapi/kernel/errors.E` — 250 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/errors.Wrapf` — 247 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/database.DBTX.Exec` — 74 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/database.DBTX.QueryRow` — 52 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/database.TxManager.WithTenant` — 29 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/model.IDGen.New` — 29 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/database.DBTX.Query` — 24 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/database.WithTenantID` — 21 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/config.binder.errf` — 18 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/database.ActorIDFrom` — 18 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/errors.Op` — 16 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/audit.deref` — 15 incoming call(s)
- `github.com/qatoolist/wowapi/module.Module.Name` — 15 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/httpx.WriteError` — 14 incoming call(s)
- `github.com/qatoolist/wowapi/kernel/model.UUIDv7` — 14 incoming call(s)

## Possible Package Cycles

_No import cycles were detected among module-local packages._

## Isolated Packages

- `github.com/qatoolist/wowapi/internal/e2e` — Package has no import relationships with other packages in this module.
  - No local PACKAGE_IMPORTS_PACKAGE edges to or from this package.
- `github.com/qatoolist/wowapi/internal/tools/benchbudget` — Package has no import relationships with other packages in this module.
  - No local PACKAGE_IMPORTS_PACKAGE edges to or from this package.

## Large Packages

- `github.com/qatoolist/wowapi/app` — Package defines 81 symbols.
  - Symbol count: 81 (threshold 40)
- `github.com/qatoolist/wowapi/internal/cli` — Package defines 74 symbols.
  - Symbol count: 74 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/authz` — Package defines 58 symbols.
  - Symbol count: 58 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/config` — Package defines 66 symbols.
  - Symbol count: 66 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/database` — Package defines 50 symbols.
  - Symbol count: 50 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/document` — Package defines 53 symbols.
  - Symbol count: 53 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/httpx` — Package defines 86 symbols.
  - Symbol count: 86 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/jobs` — Package defines 59 symbols.
  - Symbol count: 59 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/notify` — Package defines 43 symbols.
  - Symbol count: 43 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/webhook` — Package defines 52 symbols.
  - Symbol count: 52 (threshold 40)
- `github.com/qatoolist/wowapi/kernel/workflow` — Package defines 75 symbols.
  - Symbol count: 75 (threshold 40)
- `github.com/qatoolist/wowapi/testkit` — Package defines 76 symbols.
  - Symbol count: 76 (threshold 40)

## Large Interfaces

- `github.com/qatoolist/wowapi/module.Context` — Interface Context declares 38 methods.
  - Method count: 38 (threshold 10)

## Large Structs

- `github.com/qatoolist/wowapi/app.moduleContext` — Struct moduleContext has 33 fields.
  - Field count: 33 (threshold 20)
- `github.com/qatoolist/wowapi/app.moduleDeps` — Struct moduleDeps has 30 fields.
  - Field count: 30 (threshold 20)
- `github.com/qatoolist/wowapi/kernel.Kernel` — Struct Kernel has 32 fields.
  - Field count: 32 (threshold 20)

## Generated-File Summary

_No generated files were included in this scan._

## Test-File Summary

_No test files were included in this scan._

## Visualizations

When visualizations have been generated they are written to the sibling `../visualizations/` directory:

- [Graphviz DOT: ../visualizations/graph.dot](../visualizations/graph.dot)
- [Mermaid: ../visualizations/graph.mmd](../visualizations/graph.mmd)
- Package-level diagrams: `../visualizations/packages.dot`, `../visualizations/packages.mmd`

## AI Context Suggestions

Generate a compact, evidence-based context pack for any symbol before editing it:

```bash
wow-codegraph ai-context --symbol <name>
```

Suggested starting points (most connected symbols):

- `wow-codegraph ai-context --symbol E`
- `wow-codegraph ai-context --symbol Wrapf`
- `wow-codegraph ai-context --symbol Exec`
- `wow-codegraph ai-context --symbol QueryRow`
- `wow-codegraph ai-context --symbol WithTenant`

## Warnings and Limitations

51 warning(s) were produced. See [warnings.md](warnings.md) for full detail.

- **CALLGRAPH_PARTIAL**: 1
- **HIGH_FAN_IN_SYMBOL**: 16
- **HIGH_FAN_OUT_SYMBOL**: 10
- **ISOLATED_PACKAGE**: 2
- **LARGE_FILE**: 4
- **LARGE_INTERFACE**: 1
- **LARGE_PACKAGE**: 12
- **LARGE_STRUCT**: 3
- **SSA_DISABLED**: 1
- **TESTS_EXCLUDED**: 1

## Architecture Neutrality

The tool has detected code relationships. It has not assumed whether the project follows layered architecture, clean architecture, hexagonal architecture, MVC, CQRS, or any other pattern unless explicitly configured.
