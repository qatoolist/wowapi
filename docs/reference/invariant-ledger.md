# Invariant ledger

Machine-reviewable table of framework-wide invariants, their enforcement
points, consumer classes, and the regression that guards each one (third
closure audit 2026-07-17). **Rule:** a change that touches an enforcement
point, adds a consumer class, or adds a mutable field to a declaration type
must update the corresponding row (and its regression) in the same PR.

Every public runtime surface has three consumer classes to check: framework-
internal consumers, generated-project consumers (`internal/cli/templates`),
and independently maintained derived projects (stable-v1 source compatibility).

| Invariant | Enforcement | Consumers | Regression |
|---|---|---|---|
| Runtime state is isolated from module-owned declaration memory (no shared mutable storage, however nested) | deep `clone()` at Register and getters (authz, rules, document, notify, workflow); closed scalar model for workflow `Condition.Equals`; `seeds.Bundle.Clone()`; migration FS materialized to immutable byte snapshots at boot | api, worker, migrate, readiness | per-registry alias tests; `TestGatewayConditionRejectsMutableEqualsValues`; `TestGatewayRoutingImmuneToRetainedDefinitionMutation`; `TestSeedsI18nAndMigrationContentAreBootCaptured` |
| Post-boot registration is impossible | every registry sealed by `App.Boot` via `internal/sealer.Authority` (out-of-module boundary; in-repo callers limited to boot by review) | retained `module.Context`, `Booted` live pointers | `TestSealedExtensionModelRejectsEveryPostBootRegistration` (18 mutator classes) |
| The runtime view is the single source of truth; informational `Booted` fields are never authoritative | `runtimeView` captured in Boot; framework consumers + generated templates use `Runtime*` accessors only; template lint in `scripts/lint_boundaries.sh` | StartWorker, Readiness builders, generated api/migrate | `TestBootedFieldReplacementCannotAlterRuntimeState`; `TestSeedsI18nAndMigrationContentAreBootCaptured`; boundary lint |
| Only boot-produced values can run | `mustBeBooted` panic on accessors; `StartWorker` returns `ErrNotBooted`; no fallback to exported fields | all boot operations | `TestUnbootedBootedFailsLoudly` |
| Declarations are validated at boot, never deferred to first use | nil/typed-nil/empty/duplicate/nonpositive rejection across collectors and registries (recurring jobs, health, migrations/seeds FS, document hooks, integration providers, ports) | boot | `TestBootRejectsInvalidRecurringDeclarations`, `TestBootRejectsNilAndEmptyDeclarations`, `TestBootRejectsNilDocumentHooks`, `TestRegisterRejectsNilAndTypedNilProviders`, ports-enforcement suite |
| Stable-v1 consumers keep compiling (positional literals included) | frozen field sets (`app.Hook`, `document.UploadEvent`); new capabilities delivered via new types (`SupervisedHook`) or context (`UploadDeliveryFromContext`) | derived projects | `internal/compat/stable_v1_consumer_test.go`; golden consumer gate |
| Hook effects are atomic with the confirming transaction or deduplicable | delivery context carries Tx + retry-stable DeliveryID | document hooks | `TestIntegrationHookEffectsAtomicOrDeduplicatedAcrossRetry` |
| A cancelled bulk aggregate never regains pending items | one-tx Cancel; FOR SHARE aggregate reads in recovery paths; Process entry sweep | bulk workers | `cancel_recovery_test.go` (sequential + commit-window races) |
| Generated products consume only supported accessors | template lint (`scripts/lint_boundaries.sh`) | generated api/migrate/worker | boundary lint gate + golden consumer runtime |
