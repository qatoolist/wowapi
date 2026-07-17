# Platform Kernel Blueprint — Index

Coding-ready architecture & implementation blueprint for the reusable, domain-agnostic Go enterprise
backend framework specified in `Goal.md`, refined per `Goal 1.1.md` (distribution model) and
`Goal 1.2.md` (configuration & deployment) — the original prompt/vision files, retired from this
repository and archived in the `wowapi2` documentation archive (`archive/prompts-and-mandates/`);
their durable requirements live in [docs/SRS.md](../SRS.md).
The housing society product is a reference domain only; the core is fully reusable for schools,
clubs, facilities, vendor/case management, and other multi-tenant, workflow-heavy SaaS.

**Consumption model:** wowapi is a versioned third-party Go dependency
(`go get github.com/qatoolist/wowapi/v2@vX.Y.Z`) with an installable CLI
(`go install github.com/qatoolist/wowapi/v2/cmd/wowapi@vX.Y.Z`); product applications live in their
own repositories and register domain modules via the public module SDK — see
[11-framework-distribution-and-consumption.md](11-framework-distribution-and-consumption.md).

| File | Covers (Goal.md sections) |
|---|---|
| [00-overview.md](00-overview.md) | Executive recommendation, principles, architecture style + diagrams, layering, what-lives-where (§1–3) |
| [01-domain-model.md](01-domain-model.md) | Generic glossary, multi-tenancy + RLS design, actor/capacity/role/permission/policy/relationship framework (§4–6) |
| [02-workflow-rules.md](02-workflow-rules.md) | Workflow engine (custom PG vs Temporal), rule/configuration engine (§7–8) |
| [03-data-architecture.md](03-data-architecture.md) | Conventions, table matrix, ERD, PostgreSQL DDL skeleton, migration order (§9) |
| [04-project-and-primitives.md](04-project-and-primitives.md) | Project structure, package map, base model primitives, DTO/response primitives, error/validation framework (§10–11, 19, 33) |
| [05-http-and-persistence.md](05-http-and-persistence.md) | Handler helpers, repository/tx/UoW/RLS/idempotency helpers, service conventions, CRUD scaffolding (§12–15) |
| [06-module-sdk.md](06-module-sdk.md) | Module starter template, registration contract, DI/bootstrap, hook system (§16–18, 24, 34) |
| [07-platform-services.md](07-platform-services.md) | Security, performance, concurrency, documents, notifications, webhooks/integrations, events/jobs, REST conventions, observability (§20–22, 25–30) |
| [08-testing-and-tooling.md](08-testing-and-tooling.md) | Testkit, testing strategy, codegen/CLI (§31–32) |
| [09-patterns.md](09-patterns.md) | Pattern catalog with use/avoid guidance, anti-patterns, decision matrix, recommended stack (Additional Requirement §1–11) |
| [10-delivery.md](10-delivery.md) | NFR matrix, acceptance criteria, Phase 0 backlog, boundary check, first 10 files (§36–40) |
| [11-framework-distribution-and-consumption.md](11-framework-distribution-and-consumption.md) | Framework-as-dependency: public vs internal packages, product-repo usage flow, combined migrations, installable `wowapi` CLI, boundary rules (Goal 1.1) |
| [12-configuration-and-deployment.md](12-configuration-and-deployment.md) | Config layers (framework/product/module/deployment/tenant-runtime), typed contracts, precedence, secrets-by-reference, prod safety checks, per-process views, CLI config tooling, compose/k8s deployment (Goal 1.2) |
