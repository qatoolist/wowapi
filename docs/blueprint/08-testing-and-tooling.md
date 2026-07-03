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

## 2. `/internal/testkit` layout & helpers

```text
/internal/testkit/
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
                  # every event payload round-trips, no cross-module imports (AST check)
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
**Anti-lock-in rules:** generated code is committed and diffable; generators are `tools/` templates
(text/template) owned in-repo; regeneration is always optional; generated files carry a
`// Code generated — edits allowed, regen overwrites` or `edits preserved` marker per kind.

```make
make new-module name=requests    # scaffold /internal/modules/requests from template
make gen                         # sqlc + mocks + mappers (idempotent)
make migrate-create name=x mod=requests
make seed-validate               # schema-check all seeds/*.yaml against registries
make openapi-generate            # merge fragments, lint spec, diff against routes
make lint                        # golangci-lint
make lint-boundaries             # import rules + kernel vocabulary denylist
make test / test-integration / test-race / bench
make db-reset                    # local compose db recreate + migrate + seed
```
