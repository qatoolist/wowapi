# Warnings

> The tool has detected code relationships. It has not assumed whether the project follows layered architecture, clean architecture, hexagonal architecture, MVC, CQRS, or any other pattern unless explicitly configured.

51 warning(s) across 10 code(s).

## CALLGRAPH_PARTIAL (1)

- **[medium]** Some calls are dynamic (function values) and could not be statically resolved.
  - Evidence: Unresolved dynamic call sites: 503
  - Suggestion: Treat the call graph as a lower bound for dynamic dispatch; review affected symbols manually.

## HIGH_FAN_IN_SYMBOL (16)

- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/audit.deref`
  - Evidence: Direct incoming call edges: 15
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/database.ActorIDFrom`
  - Evidence: Direct incoming call edges: 18
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/database.WithTenantID`
  - Evidence: Direct incoming call edges: 21
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/errors.E`
  - Evidence: Direct incoming call edges: 250
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/errors.Op`
  - Evidence: Direct incoming call edges: 16
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/errors.Wrapf`
  - Evidence: Direct incoming call edges: 247
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/httpx.WriteError`
  - Evidence: Direct incoming call edges: 14
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/model.UUIDv7`
  - Evidence: Direct incoming call edges: 14
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/testkit.quoteIdent`
  - Evidence: Direct incoming call edges: 13
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/database.DBTX.Exec`
  - Evidence: Direct incoming call edges: 74
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/database.DBTX.Query`
  - Evidence: Direct incoming call edges: 24
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/database.DBTX.QueryRow`
  - Evidence: Direct incoming call edges: 52
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/database.TxManager.WithTenant`
  - Evidence: Direct incoming call edges: 29
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/model.IDGen.New`
  - Evidence: Direct incoming call edges: 29
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/module.Module.Name`
  - Evidence: Direct incoming call edges: 15
  - Suggestion: Changes here affect many callers; review carefully and add tests.
- **[medium]** Symbol is called by many other symbols; changes have wide impact.
  - Node: `github.com/qatoolist/wowapi/kernel/config.binder.errf`
  - Evidence: Direct incoming call edges: 18
  - Suggestion: Changes here affect many callers; review carefully and add tests.

## HIGH_FAN_OUT_SYMBOL (10)

- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/app.registerMaintenance`
  - Evidence: Direct outgoing call edges: 28
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/kernel.New`
  - Evidence: Direct outgoing call edges: 38
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/testkit.RunModuleContract`
  - Evidence: Direct outgoing call edges: 29
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/app.App.Boot`
  - Evidence: Direct outgoing call edges: 32
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/kernel/audit.Writer.Record`
  - Evidence: Direct outgoing call edges: 23
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/kernel/authz.engine.Evaluate`
  - Evidence: Direct outgoing call edges: 24
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/kernel/config.binder.bindField`
  - Evidence: Direct outgoing call edges: 26
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/kernel/document.Service.ConfirmUpload`
  - Evidence: Direct outgoing call edges: 29
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/kernel/notify.Service.Send`
  - Evidence: Direct outgoing call edges: 22
  - Suggestion: Review this symbol before making changes because it has many dependencies.
- **[medium]** Symbol calls many other symbols and may be complex.
  - Node: `github.com/qatoolist/wowapi/kernel/webhook.Service.deliverToEndpoint`
  - Evidence: Direct outgoing call edges: 26
  - Suggestion: Review this symbol before making changes because it has many dependencies.

## ISOLATED_PACKAGE (2)

- **[low]** Package has no import relationships with other packages in this module.
  - Node: `github.com/qatoolist/wowapi/internal/e2e`
  - Evidence: No local PACKAGE_IMPORTS_PACKAGE edges to or from this package.
  - Suggestion: Confirm this package is intentionally standalone.
- **[low]** Package has no import relationships with other packages in this module.
  - Node: `github.com/qatoolist/wowapi/internal/tools/benchbudget`
  - Evidence: No local PACKAGE_IMPORTS_PACKAGE edges to or from this package.
  - Suggestion: Confirm this package is intentionally standalone.

## LARGE_FILE (4)

- **[low]** File context.go defines 43 top-level symbols.
  - Node: `app/context.go`
  - Evidence: Symbol count: 43 (threshold 25)
  - Suggestion: Large files can be harder to navigate; consider splitting.
- **[low]** File service.go defines 31 top-level symbols.
  - Node: `kernel/document/service.go`
  - Evidence: Symbol count: 31 (threshold 25)
  - Suggestion: Large files can be harder to navigate; consider splitting.
- **[low]** File runner.go defines 32 top-level symbols.
  - Node: `kernel/jobs/runner.go`
  - Evidence: Symbol count: 32 (threshold 25)
  - Suggestion: Large files can be harder to navigate; consider splitting.
- **[low]** File runtime.go defines 37 top-level symbols.
  - Node: `kernel/workflow/runtime.go`
  - Evidence: Symbol count: 37 (threshold 25)
  - Suggestion: Large files can be harder to navigate; consider splitting.

## LARGE_INTERFACE (1)

- **[low]** Interface Context declares 38 methods.
  - Node: `github.com/qatoolist/wowapi/module.Context`
  - Evidence: Method count: 38 (threshold 10)
  - Suggestion: Large interfaces are harder to implement; consider splitting.

## LARGE_PACKAGE (12)

- **[low]** Package defines 81 symbols.
  - Node: `github.com/qatoolist/wowapi/app`
  - Evidence: Symbol count: 81 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 74 symbols.
  - Node: `github.com/qatoolist/wowapi/internal/cli`
  - Evidence: Symbol count: 74 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 58 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/authz`
  - Evidence: Symbol count: 58 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 66 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/config`
  - Evidence: Symbol count: 66 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 50 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/database`
  - Evidence: Symbol count: 50 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 53 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/document`
  - Evidence: Symbol count: 53 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 86 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/httpx`
  - Evidence: Symbol count: 86 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 59 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/jobs`
  - Evidence: Symbol count: 59 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 43 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/notify`
  - Evidence: Symbol count: 43 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 52 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/webhook`
  - Evidence: Symbol count: 52 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 75 symbols.
  - Node: `github.com/qatoolist/wowapi/kernel/workflow`
  - Evidence: Symbol count: 75 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.
- **[low]** Package defines 76 symbols.
  - Node: `github.com/qatoolist/wowapi/testkit`
  - Evidence: Symbol count: 76 (threshold 40)
  - Suggestion: Large packages may benefit from decomposition.

## LARGE_STRUCT (3)

- **[low]** Struct moduleContext has 33 fields.
  - Node: `github.com/qatoolist/wowapi/app.moduleContext`
  - Evidence: Field count: 33 (threshold 20)
  - Suggestion: Consider whether this struct has too many responsibilities.
- **[low]** Struct moduleDeps has 30 fields.
  - Node: `github.com/qatoolist/wowapi/app.moduleDeps`
  - Evidence: Field count: 30 (threshold 20)
  - Suggestion: Consider whether this struct has too many responsibilities.
- **[low]** Struct Kernel has 32 fields.
  - Node: `github.com/qatoolist/wowapi/kernel.Kernel`
  - Evidence: Field count: 32 (threshold 20)
  - Suggestion: Consider whether this struct has too many responsibilities.

## SSA_DISABLED (1)

- **[low]** SSA export is disabled in config; SSA was still built internally to produce the call graph.
  - Evidence: analysis.ssa = false
  - Suggestion: Enable analysis.ssa to expose SSA-based outputs in future phases.

## TESTS_EXCLUDED (1)

- **[low]** Test files were excluded from this scan.
  - Evidence: scan.include_tests = false
  - Suggestion: Pass --include-tests to include *_test.go files.

