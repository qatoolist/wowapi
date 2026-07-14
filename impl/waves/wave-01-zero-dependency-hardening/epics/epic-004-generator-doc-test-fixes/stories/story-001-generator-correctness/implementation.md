---
id: IMPL-W01-E04-S001
type: implementation-record
parent_story: W01-E04-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E04-S001

*This record aggregates the implementation reality of the story across all of its tasks. As of
2026-07-13 all five tasks are implemented and verified: the DX-02 slice (T003 + T004) and the
conductor-approved scope addition (T005) by W01Gen, and the DX-01 slice (T001 + T002) by the
follow-up worker W01GenDX01 (DEV-W01-E04-S001-01, now closed).*

## What was actually implemented

- **T001 (DX-01 T1-T4, version-resolution flags)**: `runInit` now resolves and VERIFIES the framework
  version before any file write, via a new `internal/cli/init_version.go`. Three mutually exclusive
  paths: `--framework-version` (verified via `go list -m` in a throwaway module; fail-closed with the
  exact `go list -m -versions` discovery command); `--local-framework` (absolute existing wowapi
  checkout required; emits an explicit `replace` directive — new conditional block in `go.mod.tmpl` —
  with an inert canonical placeholder require line, plus a visible dev-mode stderr warning); and the
  no-flags default, which classifies the version stamped into the binary BY SHAPE: tagged release used
  as-is, `…+dirty` stamp fails closed (the SF-7 shape — see DEV-W01-E04-S001-04), clean stamped
  pseudo-version verified resolvable pre-write, unstamped `devel` derived from `vcs.revision` via
  `go list -m <module>@<revision>` (deriving the canonical version AND proving reachability in one
  step) or failing closed. The `init_cmd.go:122-123` `v0.0.0` fallback is DELETED; every failure path
  is proven to write zero files. Fail-first order observed: both live defect shapes (`v0.0.0` from
  `go run`; `v1.0.1-…+dirty` from a dirty-tree `go build`) captured succeeding silently before any edit.
- **T002 (DX-01 T5, E2E scaffold harness)**: new `internal/cli/e2e_scaffold_harness_test.go` with the
  reusable `scaffoldPipeline` primitive — a REAL `wowapi` binary driven through `init` → `go mod tidy`
  → `go mod download` → `go build ./...` → a written-in boot-and-validate smoke test, each step's
  failure reporting its name + full output. Both CLI paths proven: source-built (devel binary,
  fail-closed guard + `--local-framework` workflow) and released (binary release-stamped via
  `-ldflags`, framework fetched at that version from a hermetic local `file://` GOPROXY packaged from
  this checkout). Fully offline (GOSUMDB=off, GOPRIVATE/GONOPROXY neutralized, module cache as dep
  proxy).
- **T003 (DX-02 verb fix)**: `internal/cli/templates/crud/resource.go.tmpl:54`'s generated DELETE-route
  permission changed `"{{.PermPrefix}}.delete"` → `"{{.PermPrefix}}.deactivate"` (one token), and
  `TestGenCRUDPermissionKeys`'s test-locking assertion changed `"widgets.widget.delete"` →
  `"widgets.widget.deactivate"` (one line) — landed together as one unit per RISK-W01-005. Fail-first
  order observed: the pre-fix test run (PASS on the buggy string) was captured before any edit.
- **T004 (generator-output-boots test)**: new `internal/cli/gen_crud_boots_test.go` with
  `TestGenCRUDOutputBoots`: scaffold product (via the existing shared `buildRenderedProduct` primitive)
  → `new-module` → `gen crud` → seed-declare the permission keys extracted VERBATIM from the generated
  file → wire routes exactly as the module template's TODO instructs → boot the product's module set
  through `app.Boot`'s full registration-validation gate (no DB needed; a no-op TxManager stub
  satisfies `kernel.Deps.Tx`). Fail-first proven: pre-T003 it failed with exactly the
  `kernel/authz/registry.go:88-90` closed-verb-set rejection; post-T003 it passes.
- **T005 (scope addition, conductor-approved 2026-07-13)**: pristine-scaffold config validation fixed
  at the true source of drift — `configs_base.yaml.tmpl`'s ACTIVE product-owned `i18n:` block (the
  file's sole violation of its own commented-example convention for product-owned sections) became a
  commented example; new fail-first test `TestInitScaffoldConfigValidates` proves
  `config validate --env local` passes on a pristine scaffold under the framework-only fallback.
  Fail-first captured: four `i18n.*: unknown key` rejections pre-fix. No i18n runtime behavior change.

## Components changed

`internal/cli` only (command logic, templates, tests). `kernel/authz/registry.go` read, not modified —
the closed verb set was not widened. `internal/buildinfo` read (Version/ModulePath/FindGoMod), not
modified.

## Files changed

- `internal/cli/init_version.go` (new — T001 resolution logic, seams, remediation wording)
- `internal/cli/init_cmd.go` (T001: two new flags, usage text, pre-write resolution, fallback deleted)
- `internal/cli/templates/init/go.mod.tmpl` (T001: conditional dev-mode `replace` block)
- `internal/cli/init_version_test.go` (new — T001 test suite, 15 tests incl. SF-7 regression tests)
- `internal/cli/e2e_scaffold_harness_test.go` (new — T002 harness + both proving tests)
- `internal/cli/templates/crud/resource.go.tmpl` (T003: 1 insertion, 1 deletion)
- `internal/cli/scaffold_test.go` (T003: TestGenCRUDPermissionKeys assertion; T001: `callInit` pins
  `--local-framework` for non-resolution tests, `buildRenderedProduct` simplified onto the new flag,
  new `wowapiCheckoutRoot` helper)
- `internal/cli/gen_crud_boots_test.go` (new file: TestGenCRUDOutputBoots + TestInitScaffoldConfigValidates)
- `internal/cli/templates/init/configs_base.yaml.tmpl` (i18n block hunk only — T005; the file's
  combined working-tree diff also carries sibling W01-E03-S002's http timeout keys, disjoint hunks)

## Interfaces introduced or changed

Two new CLI flags on `wowapi init`: `--framework-version vX.Y.Z` and `--local-framework /abs/path` —
additive. The no-flags default changes from "always succeed, possibly with an unresolvable value" to
"verified value or fail closed pre-write" (an intentional strict improvement per `story.md`
"Compatibility considerations").

## Configuration changes

None beyond the two CLI flags. CI wiring for T002/T004 falls out of test placement: `make test-unit` =
`go test ./...` with no `-short` (Makefile:163-165), invoked by the containerized CI legs
(Makefile:325-326).

## Schema or migration changes

*Not applicable — this story has no schema or migration changes (see `story.md` "Migration
considerations").*

## Security changes

*Not applicable — see `story.md` "Security considerations."* The generator output now complies with the
existing closed-verb-set control (unchanged), and init fails loudly on unverifiable input.

## Observability changes

The three fail-closed remediation messages (T001) are the story's observability surface — each contains
a copy-pasteable command (`go list -m -versions …`, the two flag forms), finalized in
`init_version.go` per `story.md` "Documentation requirements".

## Tests added or modified

- `TestGenCRUDPermissionKeys` (assertion corrected).
- `TestGenCRUDOutputBoots` (new; permanent CI regression guard on the CRUD template + kernel verb set).
- `TestInitScaffoldConfigValidates` (new; T005).
- `init_version_test.go` suite (new; 15 tests + 2 subtests covering every resolution path's success and
  fail-closed cases, incl. the SF-7 `+dirty` regression test; every failure test asserts zero files
  written).
- `TestE2EScaffoldSourceBuiltCLI` / `TestE2EScaffoldReleasedCLI` (new; the T002 harness's proving runs,
  permanent regression guards on the whole generate→build→boot→smoke pipeline).

## Commits

None yet — all changes are an uncommitted working-tree delta on top of HEAD
`05dce5c8a548f7dce3222637ab2c82024236a2a0`; the wave conductor owns commits.

## Pull requests

None (conductor owns commits/PRs).

## Implementation dates

2026-07-13 (DX-02 slice + T005: W01Gen; DX-01 slice: W01GenDX01).

## Technical debt introduced

None.

## Known limitations

- T004's boot proof exercises the registration-validation gate (where DX-02's rejection lives) without
  a database; DB-backed migration/seed-sync proof remains testkit/e2e territory.
- The generated DELETE handler body remains the soft-delete TODO stub — DX-02's P1/Wave-4 scope,
  explicitly out of this story.
- T002's released path proves the pipeline against a locally packaged v0.1.0 proxy, not a published tag
  (none exists); `--local-framework` validates module identity, not version compatibility (DX-05 T4's
  concern, W01-E04-S002).

## Follow-up items

- None owed by this story. Forward references: future DX-04 (W06) calls `scaffoldPipeline`;
  W01-E04-S002's FBL-03 can recommend closing wowsociety's SF-7 upstream finding once this tree is
  committed.

## Relationship to the approved plan

All plan steps 1-7 implemented, with the plan's "Unresolved questions" resolved as recorded in
task-001/task-002 and DEV-W01-E04-S001-04 (reachability = `go list -m` resolution; harness in one new
test file; released-vs-source = two `-ldflags` builds; `initData` + template gained the replace shape).
One same-class scope extension recorded (DEV-W01-E04-S001-04: the Go 1.24+ stamped-version shapes,
including the SF-7 `+dirty` stamp). Prior scope deviations DEV-W01-E04-S001-01 (DX-01 deferred) is
closed; DEV-W01-E04-S001-03 (T005 scope addition) stands as recorded.
