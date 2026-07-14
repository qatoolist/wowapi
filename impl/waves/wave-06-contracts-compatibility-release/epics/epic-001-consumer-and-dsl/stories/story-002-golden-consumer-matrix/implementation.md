---
id: IMPL-W06-E01-S002
type: implementation-record
parent_story: W06-E01-S002
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Implementation record — W06-E01-S002

## What was actually implemented

The completed fixture:

- installs the candidate CLI with real `go install` into an isolated `GOBIN` and verifies binary
  provenance;
- reuses W01's `scaffoldPipeline` and resolves the framework from a versioned module proxy without a
  checkout `replace`;
- generates `catalog` and `fulfillment`, two CRUD resources, and the required rule, workflow, event
  handler, recurring job, document flow, notification, and webhook;
- verifies every expected generated file, automatic module registration, tidy/build, and generated
  boot tests;
- boots generated API and worker processes against real Postgres, MinIO, Mailpit, and Jaeger/OTel;
- proves authenticated CRUD, cross-tenant RLS denial, transactional-outbox dispatch, and worker
  stop/restart recovery;
- generates and proves the fixture with tagged `v1.1.0`, upgrades its framework dependency and
  generated scaffold to candidate `v1.2.0-w06e01s002.11`, reruns build/boot contracts, and exercises
  the upgraded fixture against real infrastructure; and
- registers `make golden-consumer` in ordinary CI and the Wave-4 required-gates manifest, with the
  exact-SHA runner provisioning all required services including Jaeger.

## Components changed

Golden-consumer test infrastructure, generator coverage supplied by the wider W06 implementation,
CI/release-gate configuration, user-facing target documentation, and W06-E01-S002 governance/evidence.

## Files changed

- `internal/cli/golden_consumer_test.go`
- `internal/cli/golden_consumer_infra_test.go`
- `internal/cli/e2e_scaffold_harness_test.go`
- `internal/cli/cli.go`
- `internal/cli/gen_cmd.go`
- `internal/cli/gen_crud_boots_test.go`
- `internal/cli/new_module_cmd.go`
- `internal/cli/templates/crud/migration.sql.tmpl`
- `internal/cli/templates/crud/resource.go.tmpl`
- `internal/cli/templates/init/internal_wire_modules.go.tmpl`
- `internal/cli/templates/init/cmd_migrate_main.go.tmpl`
- `internal/cli/templates/module/module.go.tmpl`
- `internal/cli/templates/subsystem/subsystem.go.tmpl`
- `Makefile`
- `.github/workflows/ci.yml`
- `.github/workflows/required-gates.yml`
- `ci/release-gates.yaml`
- `ci/release-gates.schema.json`
- `testkit/rls_isolation_all_test.go`
- `docs/user-guide/cli-reference.md`
- story task/artifact/evidence/deviation/verification/closure records

## Interfaces introduced or changed

No production interface was introduced by this story's closure work. The fixture consumes the
installed public CLI and generated product surface.

## Configuration changes

The golden-consumer target is a required Wave-4 gate. Both CI runners start Postgres, MinIO, Mailpit,
and Jaeger before invoking it.

## Schema or migration changes

None in the framework. Generated temporary consumers execute their generated migrations in disposable
test databases.

## Security changes

No authorization implementation change in this story. The real-infrastructure contract proves API-key
authentication and tenant isolation.

## Observability changes

No tracing implementation change in this story. The gate requires Jaeger, sets the OTLP endpoint, and
checks the Jaeger service before generated processes start.

## Tests added or modified

- `TestGoldenConsumerInstalledBinaryTwoModules`
- `TestGoldenConsumerFailingFixture`
- `TestGoldenConsumerRealInfrastructure`
- `TestGoldenConsumerUpgradeReplay`
- `make golden-consumer` selection also includes `TestIntegrationRLSCensusComplete`

## Commits

Shared worktree snapshot based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; authoritative file
checksums are recorded in `artifacts/index.md`.

## Pull requests

None.

## Implementation dates

2026-07-13 through 2026-07-14.

## Technical debt introduced

None identified.

## Known limitations

The N side of the replay is a locally packaged release candidate, intentionally not a falsely claimed
published tag. Hosted GitHub execution of the exact-SHA gate remains the responsibility of the release
pipeline; local evidence proves the same command with the same service set.

## Follow-up items

None for this story. W06-E02-S003 may consume the completed replay as planned.

## Relationship to the approved plan

T001 through T005 are implemented as approved. The earlier DEV-W06-E01-S002-001 blocker was resolved
by the completed generator surface and automatic module wiring; it is retained as historical evidence,
not treated as an approved scope reduction.
