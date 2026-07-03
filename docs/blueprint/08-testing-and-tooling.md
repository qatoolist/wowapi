# 08 — Testing Framework & Testkit, Code Generation / Templates / CLI

## 1. Testing strategy

| Level | Scope | Infra |
|---|---|---|
| Unit | domain funcs, validators, policy/rule evaluation, cursor encoding, backoff math | none; table-driven |
| Integration (default for modules) | service+repo+kernel services against real Postgres | testcontainers (one container per package, template-DB clone per test for speed) |
| Contract | every module passes the kernel's module-contract suite | testkit |
| Security | RLS isolation, authz matrix, route-metadata completeness, secret-leak scan of logs | testkit + CI |
| Race / bench | `go test -race ./...`; budget benchmarks | CI gates |

Real DB over mocks (per CLAUDE.md and Goal.md): repos and RLS are only meaningfully tested against
Postgres. Fakes are reserved for process/network boundaries we can't stand up locally: mail/SMS/
WhatsApp/push providers, malware scanner, external IdP (token minting faked), payment callbacks.

## 2. `wowapi/testkit` layout & helpers

The testkit is a **public package** (`github.com/qatoolist/wowapi/testkit`) precisely so external
product repositories can integration-test their modules with the same fixtures, fakes, and
assertions the framework uses on itself. Framework-repo-only helpers stay in its `_test.go` files
or `internal/`; anything a product module's tests need is exported here.

```text
wowapi/testkit/
  app.go          # NewApp(t) → real Kernel on testcontainer PG, fake clock/idgen/providers, registered test modules
  db.go           # NewDB(t), template-clone, per-test tx rollback option
  fixtures.go     # CreateTenant, CreateOrg, CreateUser, CreatePerson/Party, CreateCapacity,
                  # GrantRole, CreateAssignment, CreateResource, Relate(subject, rel, object)
  auth.go         # IssueToken(user, tenant, capacity) — locally-signed JWT the test IdP verifier accepts
  asserts.go      # AssertRLSIsolation(t, table) · AssertAllowed/Denied(actor, perm, target)
                  # AssertAuditRow(action, resource) · AssertOutboxEvent(type, matcher)
                  # AssertWorkflowStep(instance, step) · AssertRuleResolves(key, at, want)
                  # AssertIdempotentReplay(req) · AssertNoSecretsInLogs(t)
  fakes/          # clock.go (manual-advance), idgen.go (deterministic uuidv7 seq),
                  # notify.go (capture channel), storage.go (in-mem object store, presign stub),
                  # scanner.go, webhookverifier.go, integration.go
  workflowsim.go  # WorkflowSim fluent driver (see 02)
  contract.go     # RunModuleContract(t, module): registers module alone on a fresh kernel,
                  # asserts: boots, validates, migrates, seeds idempotently (run twice),
                  # every route has meta, every permission seeded, RLS on every module table,
                  # every event payload round-trips, no cross-module imports (AST check),
                  # module config: boots on empty namespace (complete defaults) + rejects invalid config
  http.go         # authenticated test client bound to a tenant/capacity
```

Conventions: table-driven tests everywhere; fixtures return typed handles (`TenantHandle.OrgID`…);
fake clock/idgen injected through the same constructors production uses — no test hooks in
production code. Every kernel guarantee in the acceptance criteria ([10-delivery.md](10-delivery.md))
has a named assertion here; the criteria are executable.

## 3. Code generation, templates, CLI

**Generate (mechanical, no decisions):** module skeleton, CRUD vertical slice (see 05 §4), sqlc
output, DTO↔domain mappers, mock/fake interfaces (`moq` for ports), OpenAPI merged doc + stubs,
migration file stamps, seed file skeletons, test fixture builders.
**Never generate:** services with real logic, workflow definitions' semantics, policy conditions,
anything money.
**Anti-lock-in rules:** generated code is committed and diffable; templates are text/template
assets **embedded in the `wowapi` CLI binary** (no cloning/copying the framework repo); generators
write into the *consuming product repository*, never into wowapi; regeneration is always optional;
generated files carry a `// Code generated — edits allowed, regen overwrites` or `edits preserved`
marker per kind.

### The `wowapi` CLI — the primary tooling workflow

Installable, cross-platform (macOS/Linux/Windows), also shipped as goreleaser release binaries:

```text
go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z

wowapi init --module example.com/acme-ops --wowapi-version vX.Y.Z
                                                   # scaffold a new product repo; flags make it repeatable
wowapi new-module requests                         # scaffold <repo>/internal/modules/requests
wowapi gen crud --module requests --resource request
wowapi gen                                         # sqlc + mocks + mappers (idempotent)
wowapi migrate create --module requests --name create_requests
wowapi seed validate                               # schema-check all seeds/*.yaml against registries
wowapi openapi merge [--check]                     # merge fragments, lint spec, diff against routes
wowapi lint boundaries                             # import rules; framework repo adds vocabulary denylist
wowapi version                                     # prints CLI version + go.mod wowapi version, warns on mismatch

wowapi config init                                 # scaffold configs/{base,<env>}.yaml + typed Config stub
wowapi config validate [--env prod]                # full load+validation incl. unsafe-in-prod checks (CI gate)
wowapi config doctor                               # effective config + per-key provenance, redacted
wowapi config print --redacted [--env stage]       # canonical effective config (always redacted)
wowapi config diff --from dev --to prod            # redacted effective diff between environments
wowapi config schema                               # JSON Schema from struct tags (framework+product+modules)
wowapi deploy render --env prod                    # render compose/k8s manifests from effective config
```

Config tooling semantics are specified in [12-configuration-and-deployment.md](12-configuration-and-deployment.md) §8–9.

The installed CLI version should match the product's `wowapi` dependency version (it reads `go.mod`
and warns on mismatch). CI runs `wowapi seed validate`, `wowapi openapi merge --check`, and
`wowapi lint boundaries` directly. `go run github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z <cmd>`
remains a no-install fallback for tightly pinned CI jobs — not the primary developer experience.

Product repos may keep thin Makefile wrappers (source of truth stays the CLI):

```make
new-module: ; wowapi new-module $(name)
gen:        ; wowapi gen
lint:       ; golangci-lint run && wowapi lint boundaries
test / test-integration / test-race / bench        # go test invocations
db-reset:   # local compose db recreate + migrate + seed
```
