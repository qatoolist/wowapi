------- ACTIVE IMPLEMENTATION GOAL - DO NOT MARK ACCOMPLISHED UNTIL THE FRAMEWORK IS COMPLETE --------

# Goal 2: Implement The Complete WowAPI Framework

You are the implementation lead for `wowapi`.

Your job is to implement the entire `wowapi` framework until it is genuinely usable as a production-ready, reusable, domain-agnostic Go backend framework dependency.

This is no longer only an architecture exercise. The blueprint must become working code, tests, tooling, documentation, and containerized development workflows.

## Source Material

Before implementing anything, read and understand:

- `Goal.md`
- `Goal 1.1.md`
- `Goal 1.2.md`
- every file under `docs/blueprint/`
- `docs/graphify.md`

The older goal files are marked accomplished because their architecture/design tasks are done. Do not redo those goals. Use them as reference material for this implementation goal.

## Non-Negotiable Product Boundary

`wowapi` is the framework repository.

It must remain:

- domain-neutral,
- reusable as a third-party Go dependency,
- free of real product modules,
- free of housing-society-specific implementation,
- importable by separate product repositories,
- usable through public packages and an installable `wowapi` CLI.

Do not implement the housing society product in this repository.

Neutral fixtures and examples are allowed only where the blueprint allows them:

- private contract-test fixtures under `internal/testmodules`,
- standalone non-contractual examples under `examples`,
- no real product/domain module in the framework core.

## Primary Objective

Implement all framework capabilities described by the goals and blueprint:

- public package layout,
- kernel primitives,
- module SDK,
- app composition,
- configuration and deployment model,
- migration and seed system,
- tenant/RLS database foundation,
- authn/authz/policy/relationship/resource framework,
- workflow engine,
- rule/config engine,
- audit logging,
- outbox/events/jobs,
- document/file/comment/attachment framework,
- notification/webhook/integration framework,
- HTTP helpers and API conventions,
- validation/error/pagination/filtering helpers,
- testkit,
- CLI and code generators,
- containerized development/test workflow,
- Makefile targets,
- examples and contract fixtures,
- documentation and runbooks.

The framework is complete only when the acceptance criteria in `docs/blueprint/10-delivery.md`, `docs/blueprint/11-framework-distribution-and-consumption.md`, and `docs/blueprint/12-configuration-and-deployment.md` are implemented and verified by tests.

## Working Style: Act Like A Professional Engineering Team

Work as a disciplined engineering team, not as a single-pass coding assistant.

You must:

1. Build a phase plan before coding.
2. Split work into coherent epics and implementation slices.
3. Use multiple agents/models in parallel wherever this materially reduces time and risk.
4. Assign work to cost-effective agents:
   - smaller/cheaper agents for localized code generation, docs, fixtures, Makefile edits, and straightforward tests,
   - stronger agents for architecture-critical, security-sensitive, concurrency-heavy, database/RLS, authz, workflow, and review tasks.
5. Keep task ownership clear so parallel agents do not edit the same files unnecessarily.
6. Use critique/review agents to challenge design, code, security, performance, and test adequacy.
7. Cross-check generated work before integration.
8. Treat tests, lint, review, documentation, and operability as part of the implementation, not follow-up chores.
9. Keep the repository coherent at every phase.
10. Commit only when the working tree is in a clean, verified state unless explicitly doing a checkpoint branch/commit.

When using agents, assign them concrete responsibilities such as:

- package implementation,
- SQL/migration design,
- testkit development,
- CLI/codegen implementation,
- security review,
- performance review,
- documentation consistency review,
- acceptance-test review,
- Graphify graph update/review.

Do not blindly trust agent output. Review it like code from a teammate.

## Required Planning Artifacts

At the start of implementation, create and maintain planning artifacts such as:

```text
docs/implementation/
  phase-plan.md
  progress.md
  decisions.md
  test-strategy.md
  risk-register.md
  evidence/
    README.md
    phase-XX/
      proof-bundle.md
      review-findings.md
      command-log.md
      acceptance-map.md
```

The plan must map blueprint epics to implementation phases, dependencies, tests, and acceptance criteria.

Update these files as phases complete.

## Evidence Gate: No Memory-Only Proofs

Do not rely on memory, chat claims, or unstored reasoning as proof that work happened.

Every important decision, discussion, criticism, review, fix, and acceptance claim must produce durable, verifiable evidence in the repository or in command output that is summarized into repository artifacts.

For each phase, maintain a proof bundle under `docs/implementation/evidence/phase-XX/`.

Each proof bundle must include:

1. **Decision evidence**
   - ADRs, entries in `docs/implementation/decisions.md`, or phase notes.
   - Each decision must include context, options considered, selected option, tradeoffs, and links to the files or tests affected.

2. **Discussion evidence**
   - Summaries of meaningful design discussions or agent debates.
   - Include the question being resolved, positions considered, final decision, and implementation consequence.

3. **Critique/review evidence**
   - Review findings with severity, file/line references, owner, status, and final resolution.
   - If a finding is rejected, record the reason and reviewer/implementer agreement.

4. **Implementation evidence**
   - File paths changed.
   - Package/API/migration/CLI/test artifacts added.
   - Mapping from review findings and acceptance criteria to concrete commits or patches.

5. **Verification evidence**
   - Exact commands run.
   - Exit status.
   - Relevant summarized output.
   - Links to generated reports where applicable.
   - If a command could not run, document why, what risk remains, and what must be run later.

6. **Acceptance evidence**
   - A phase acceptance checklist.
   - Each item must point to code, tests, command output, or review notes.
   - No unchecked acceptance item may be silently ignored.

Memory-only statements such as "reviewed", "tested", "looks good", "implemented", or "verified" are not acceptable unless backed by proof artifacts.

If a proof artifact is missing, the work is not done.

## Commit Discipline

Commit code whenever it is appropriate and safe.

Use small, coherent commits at phase or slice boundaries, especially after:

- repository/bootstrap setup,
- a package reaches a working tested state,
- migrations and matching tests land,
- a CLI command and its tests land,
- a review finding is fixed,
- an acceptance criterion is satisfied,
- a phase is completed.

Before each commit:

1. Run the relevant Makefile/verification targets for the changed scope.
2. Update the phase proof bundle.
3. Ensure generated files are either intentionally committed or intentionally ignored.
4. Check `git status`.
5. Confirm no local-only secrets, graph artifacts, editor state, or assistant scratch files are staged.

Commit messages must be specific and evidence-oriented, for example:

```text
Implement tenant transaction manager and RLS tests
Add module SDK registration contract and contract tests
Fix authz denial audit review findings
```

Do not create a commit just to hide incomplete or failing work. If checkpoint commits are needed for a long phase, clearly label them and keep the proof bundle honest about what remains.

## Required Use Of Graphify

This project is expected to grow substantially. Use Graphify regularly.

At the start of major phases:

```bash
scripts/graphify_refresh.sh check
```

After substantial code or architecture changes:

```bash
scripts/graphify_refresh.sh update
```

After substantial documentation or design changes, if an LLM backend key is available:

```bash
scripts/graphify_refresh.sh extract
```

Use the graph to review architecture, hidden coupling, and package-boundary drift. Do not commit generated `graphify-out/` artifacts unless explicitly instructed.

## Engineering Standards

Use practical Go engineering.

Required:

- Go modules with idiomatic package names.
- Public packages only where needed by consuming product repositories.
- Private implementation under `internal`.
- No circular imports.
- No service locator.
- Constructor injection.
- Interfaces at boundaries, concrete structs inside packages.
- No reflection-heavy runtime magic.
- No ORM hiding SQL.
- `pgx` and `sqlc` where specified by the blueprint.
- PostgreSQL RLS for tenant isolation.
- Transactional outbox for side effects.
- River or the selected Postgres-backed job runner behind framework interfaces.
- Typed configuration, strict validation, and secret references.
- Immutable hot-path config.
- Route metadata enforcement.
- Deny-by-default authorization.
- Stable error contracts.
- Context propagation.
- Graceful shutdown.
- Observability hooks.
- Redaction by design.
- Benchmarks for hot paths.

Forbidden:

- product-domain leakage into framework core,
- housing society implementation,
- runtime low-code CRUD engine,
- generic repositories hiding SQL,
- untyped global config maps,
- raw secrets in config or logs,
- security behavior disabled by config,
- dynamic config reads on request/job hot paths,
- direct cross-module SQL joins,
- module imports of another module's internals,
- product imports from `wowapi/internal`,
- test-only hooks in production code,
- bypassing RLS or tenant context for tenant data.

## Container-First Execution

All development, tests, and validation must be runnable inside containers.

Provide and maintain:

- `Dockerfile`,
- `docker-compose.yaml` or equivalent under `deployments/`,
- Postgres service,
- MinIO or S3-compatible local storage,
- Mailpit or equivalent local mail sink,
- any required local fake providers,
- containerized test runner,
- containerized lint/check runner,
- containerized migration runner.

Local host execution can be supported, but container execution is the standard.

Do not require developers to install Postgres, MinIO, or external services locally.

## Makefile Requirements

Create a serious `Makefile` that makes common work simple.

It should include targets such as:

```text
make help
make setup
make tools
make up
make down
make reset
make logs
make shell
make db-shell
make migrate
make seed
make lint
make lint-boundaries
make fmt
make test
make test-unit
make test-integration
make test-contract
make test-security
make test-race
make bench
make coverage
make gen
make new-module name=requests
make gen-crud module=requests resource=request
make openapi
make config-validate
make config-doctor
make graph-check
make graph-update
make ci
```

Make targets should call the `wowapi` CLI or project scripts where appropriate.

The Makefile must be usable by humans and CI.

## CLI Requirements

Implement the installable CLI:

```bash
go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z
```

It must provide, at minimum:

```text
wowapi init --module example.com/acme-ops --wowapi-version vX.Y.Z
wowapi new-module requests
wowapi gen
wowapi gen crud --module requests --resource request
wowapi migrate create --module requests --name create_requests
wowapi seed validate
wowapi openapi merge
wowapi lint boundaries
wowapi version
wowapi config init
wowapi config validate
wowapi config doctor
wowapi config print --redacted
wowapi config diff --from dev --to prod
wowapi config schema
wowapi deploy render --env prod
```

The CLI must:

- work on Linux, macOS, and Windows,
- embed generator templates,
- write generated files into consuming product repositories,
- avoid business-logic generation,
- produce diffable output,
- warn on CLI/framework dependency version mismatch,
- support CI use,
- never require cloning the framework repo to use generators.

## Testing Requirements

Testing must be comprehensive and executable.

Implement and maintain:

1. Unit tests.
2. Integration tests using real Postgres in containers.
3. Contract tests for modules.
4. RLS isolation tests.
5. Authz matrix tests.
6. Route metadata tests.
7. Error envelope tests.
8. Config validation and redaction tests.
9. Secret leakage tests.
10. Migration idempotency tests.
11. Seed idempotency tests.
12. Outbox crash/retry/idempotency tests.
13. Job retry/DLQ tests.
14. Workflow simulation tests.
15. Rule versioning and historical resolution tests.
16. Document/file flow tests.
17. Webhook signature/replay tests.
18. OpenAPI route/spec consistency tests.
19. CLI golden-file tests.
20. Code generator golden-file tests.
21. Boundary lint tests.
22. Race tests.
23. Benchmarks for hot paths.
24. End-to-end functional smoke tests through API, worker, migrate, and DB.

Tests must run through Makefile targets and inside containers.

Use fakes only at real external boundaries such as:

- mail/SMS/push providers,
- malware scanner,
- object storage where a local emulator is not suitable,
- external IdP token minting,
- third-party webhooks.

Do not mock repositories or RLS-sensitive database behavior when the real database is required to prove correctness.

## Review And Critique Requirements

Every phase must include rigorous review.

Use critiques for:

- package architecture,
- public API surface,
- dependency cycles,
- tenant isolation,
- RLS correctness,
- authz correctness,
- configuration security,
- secret redaction,
- migration safety,
- event/job reliability,
- workflow/rule correctness,
- performance/hot-path allocation,
- generator output quality,
- test adequacy,
- documentation consistency.

Review output should be actionable:

- findings first,
- severity,
- file/line references,
- concrete fixes,
- residual risk.

Do not treat "tests pass" as sufficient review.

Every review must be captured in the phase proof bundle. Every accepted finding must be traceable to a code/doc/test change or a recorded decision explaining why no change was made.

## Implementation Phases

The AI team must create a detailed phase plan. At minimum, use phases like these unless a better dependency-aware plan is justified.

### Phase 0: Bootstrap And Planning

Deliver:

- `go.mod`,
- repository structure,
- Makefile,
- Dockerfile,
- deployments compose stack,
- lint tooling,
- boundary lint scaffold,
- CI/local command plan,
- implementation phase plan,
- risk register,
- initial Graphify check.

Exit criteria:

- `make help` works,
- `make up` starts local infra,
- `make lint` and `make test-unit` run, even if initial packages are skeletal,
- package graph rules are encoded.

### Phase 1: Config, Logging, App Skeleton

Deliver:

- `kernel/config`,
- typed framework config,
- product config composition contracts,
- secret refs and redaction,
- config CLI commands,
- `kernel/logging`,
- `app` skeleton,
- process config views,
- startup/shutdown skeleton.

Exit criteria:

- invalid config fails with complete errors,
- secrets never print,
- API/worker/migrate config views are tested,
- no package cycles.

### Phase 2: Database, Migrations, Tenant Foundation

Deliver:

- `kernel/database`,
- pgx pool,
- transaction manager,
- tenant-bound transaction API,
- RLS helpers,
- kernel migrations,
- migration runner,
- tenant/user/access tables,
- testkit DB helpers.

Exit criteria:

- fresh DB migrates idempotently,
- tenant-scoped query without tenant context fails,
- RLS isolation tests pass.

### Phase 3: HTTP, Errors, Validation, Pagination

Deliver:

- `kernel/httpx`,
- middleware chain,
- route metadata,
- problem details errors,
- validator wrapper,
- pagination/filtering/sorting helpers,
- idempotency helpers.

Exit criteria:

- routes without metadata fail registration,
- error contracts tested,
- list helpers tested,
- idempotency behavior tested.

### Phase 4: Identity, Actor, Authorization, Policy, Relationship, Resource

Deliver:

- auth middleware/OIDC verifier,
- principal and actor model,
- capacities,
- roles/permissions/assignments,
- policy evaluator,
- relationship framework,
- resource registry,
- record-level filter hooks,
- break-glass/impersonation flows if in blueprint scope.

Exit criteria:

- deny by default,
- authz matrix tests pass,
- sensitive denials audited,
- module route permissions register from seeds.

### Phase 5: Module SDK, Seeds, Testkit, Example Fixture

Deliver:

- public `module` package,
- module registries,
- lifecycle validation,
- migrations/seeds/OpenAPI registration,
- private neutral test module under `internal/testmodules`,
- public `testkit`,
- module contract suite.

Exit criteria:

- external scratch product repo can import `wowapi`,
- module contract tests pass,
- product module can register everything without framework edits.

### Phase 6: Outbox, Events, Jobs

Deliver:

- transactional outbox,
- dispatcher,
- inbox/idempotent handlers,
- job runner integration,
- retry policies,
- DLQ,
- worker process.

Exit criteria:

- events commit atomically with business writes,
- crash/retry tests pass,
- jobs are tenant-aware,
- graceful worker shutdown works.

### Phase 7: Rules And Workflow Engine

Deliver:

- rule point registry,
- rule version storage/resolution,
- approval-gated rule activation,
- workflow definitions,
- workflow runtime,
- tasks/decisions/delegation/override,
- SLA sweeper,
- workflow simulator.

Exit criteria:

- rule historical resolution works,
- tenant overrides work,
- workflow simulations pass,
- rule/workflow changes are audited/versioned.

### Phase 8: Documents, Files, Comments, Attachments

Deliver:

- document metadata,
- file versions,
- storage adapter/fake,
- presigned upload/download flow,
- scan hook,
- grants,
- comments,
- attachments,
- retention/redaction jobs.

Exit criteria:

- upload-confirm-download flow tested,
- grants enforced,
- audit rows produced,
- unsafe file behavior rejected.

### Phase 9: Notifications, Webhooks, Integrations

Deliver:

- notification templates/preferences/deliveries,
- provider ports/fakes,
- outbound webhook signing/delivery/retry,
- inbound webhook verification/replay protection,
- integration provider registry,
- circuit breaker behavior.

Exit criteria:

- replayed inbound events rejected,
- outbound webhook retries/DLQ work,
- credentials are secret refs only,
- provider payloads do not leak into kernel types.

### Phase 10: CLI, Codegen, OpenAPI

Deliver:

- installable `cmd/wowapi`,
- init generator,
- module generator,
- CRUD generator,
- migration helper,
- seed validator,
- OpenAPI merge/check,
- boundary lint,
- config tooling,
- deploy render tooling,
- golden tests.

Exit criteria:

- CLI works from a scratch product repo,
- generated module compiles and passes contract tests,
- generated CRUD slice has tests,
- OpenAPI check catches drift.

### Phase 11: Observability, Performance, Security Hardening

Deliver:

- metrics,
- tracing,
- health/readiness,
- log redaction,
- config fingerprinting,
- dashboards/runbook starter,
- benchmarks,
- race tests,
- security test suite,
- performance budgets.

Exit criteria:

- hot-path benchmark budgets met,
- race tests pass,
- secret leak tests pass,
- readiness reflects real dependencies.

### Phase 12: End-To-End Acceptance And Release Readiness

Deliver:

- external scratch product repo acceptance test,
- API/worker/migrate E2E smoke,
- generated module acceptance,
- all acceptance criteria mapped and checked,
- release notes,
- final Graphify update,
- final critique/review pass.

Exit criteria:

- `make ci` passes in containers,
- all blueprint acceptance criteria are satisfied,
- no critical/high review findings remain,
- public API surface is documented,
- framework can be consumed as a dependency by a separate product repo.

## Deployment Scope

Do not spend excessive time building a full robust production deployment platform.

However, the framework itself must be production-ready.

Required:

- container images,
- local compose stack,
- migration job support,
- clear config/deployment docs,
- Kubernetes/plain manifest rendering if in blueprint scope,
- health/readiness endpoints,
- graceful shutdown,
- logs/metrics/traces,
- backup/restore runbook starter,
- deployment notes.

Not required for this implementation goal:

- full managed-cloud production automation,
- multi-region deployment,
- advanced progressive delivery system,
- complete Helm chart ecosystem,
- external SaaS operations platform.

## Definition Of Done

This goal is complete only when all of the following are true:

1. The framework builds.
2. The CLI builds and installs.
3. Public packages can be imported from an external product repo.
4. The external product repo can define a module and register it.
5. Kernel migrations and product module migrations run together.
6. Tenant RLS is enforced and tested.
7. Authz denies by default and is tested.
8. Rules and workflows run with versioning/audit.
9. Events/jobs/outbox are reliable and tested.
10. Documents/files/comments/attachments work and are tested.
11. Notifications/webhooks/integrations work with fakes/adapters and are tested.
12. Config is typed, strict, redacted, and process-scoped.
13. Generators produce compiling, tested code.
14. OpenAPI merge/check works.
15. Testkit supports external product modules.
16. Makefile targets work.
17. Containerized test and development flow works.
18. Unit, integration, contract, security, race, benchmark, and E2E tests exist and pass.
19. Boundary lint proves no forbidden imports or domain leakage.
20. Graphify has been updated.
21. Documentation explains how to build, test, consume, and extend the framework.
22. Final review finds no unresolved critical or high severity issues.
23. Every phase has a proof bundle with decisions, reviews, criticisms, fixes, commands, and acceptance evidence.
24. Every accepted review finding is traceably resolved.
25. Commit history contains coherent implementation checkpoints and no local-only/generated noise.

Do not mark this goal accomplished until the implementation is actually complete and verified.

## Final Output Required From The Implementing AI

When complete, report:

1. Implemented phases.
2. Key packages and binaries.
3. Public API surface.
4. CLI commands.
5. Makefile targets.
6. Container workflow.
7. Tests and verification results.
8. Remaining known limitations, if any.
9. How a product repo consumes `wowapi`.
10. Proof bundle index and acceptance evidence summary.
11. Commit hash or PR summary.

Keep the final report factual and evidence-based.
