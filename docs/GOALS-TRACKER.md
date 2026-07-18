# wowapi goals and release tracker

**Updated:** 2026-07-18

**Current line:** `github.com/qatoolist/wowapi`

**First supported release target:** `v1.2.0`

The published `v1.0.0` and `v1.1.0` identities are abandoned and immutable. They are not supported
applications, database predecessors, generated-project formats, or compatibility baselines.
Compatibility commitments begin at `v1.2.0`.

This tracker describes the current architecture and release programme. Historical phase evidence and
dated reviews are evidence, not current status sources; the executable gates and
[`invariant-ledger.md`](reference/invariant-ledger.md) are authoritative.

## Stable-baseline goals

| Goal | Required outcome | Current disposition |
|---|---|---|
| One module identity | Root module, one template source, exact build stamping, strict release identity | Implemented; locally contract-tested |
| One boot/runtime model | Opaque validated boot result, immutable snapshots, sealed extension registries | Implemented; adversarial regressions retained |
| One public API per capability | No forwarding shims, weak parallel constructors, hidden compatibility transport, or obsolete shapes | Consolidated for the reviewed surface |
| Workflow definition integrity | Registered definitions are canonical; atomic sync; canonical JSON/digest; one verified execution loader | Implemented in the clean-line candidate; DB and generated-product gates required |
| Fresh database baseline | One current-state migration with semantic equality to the proven historical head | Baseline generation/equivalence is a release gate |
| Reliable process operation | Supervised hooks/workers, bounded shutdown, fenced recovery, surfaced fan-out/requeue errors | Implemented; race/DB/container regressions retained |
| Misuse-resistant boundaries | Explicit tenant, actor, credential scheme, route contract, checksum, idempotency, and pagination identity | Implemented for reviewed entry points |
| Honest CI/release evidence | One gate owner; exact-SHA hashed evidence; clean bootstrap versus future predecessor policy | Implemented locally; hosted artifact/provenance behavior remains post-push proof |
| Independent generated consumer | Installed CLI generates, compiles, migrates, boots, serves, and shuts down a clean product | Required before push-ready verdict |
| Repository coherence | README/SRS/blueprints/operations/CHANGELOG/report/Graphify agree with source | Required before final verdict |

## Required foundation retained

- tenant-bound transactions, RLS, role/grant enforcement, and tenant FK gates;
- audit chains, actor attribution, privileged grant resolution, and step-up policy;
- transactional outbox, idempotent jobs, fenced leases, recovery, and leader-safe scheduling;
- current document, bulk, notification, webhook, integration, retention, rules, and workflow services;
- boot validation/sealing, dependency compilation, generated-product composition, and narrow runtime accessors;
- metrics/tracing boundaries, readiness, localization fallback, and safe non-production adapters;
- MFA factor primitives as a documented and tested leaf foundation capability;
- online expand/backfill/validate/canary/switch/contract machinery for future supported-line upgrades;
- generic API, configuration, OpenAPI, migration, and architecture compatibility engines for releases
  after `v1.2.0`; and
- independent event, config, signature, audit, workflow, and performance-dataset schema versions.

## Removed from the clean baseline

- abandoned-release forwarding packages, aliases, shape-freeze fixtures, and allowlist entries;
- old hook/runtime/retention/job/upload/storage/resource convenience paths with weaker invariants;
- unsigned cursors, ignored privileged claims, implicit credential inference, webhook tenant fallback,
  optional route-contract enforcement, and markerless generator mutation;
- checksum repair/backfill for abandoned objects;
- disconnected typed ports, app-model projections, advisory lifecycle manifest/lint, reserved DB i18n,
  dead CLI branches, comparability-only tests, and production-package test doubles;
- workflow vote, ratification, multi-approval, and self-approval policy fields that had no implemented
  state-machine behavior; and
- old migration replay and historical expand/backfill choreography from the fresh-install baseline.

## Release-candidate gate

The local candidate is push-ready only when all of the following are green on the same source state:

1. formatting, vet, tidy/module verification, boundaries, constructors, templates, docs, release
   identity, workflow-contract, and diff checks;
2. complete Go tests with PostgreSQL and MinIO required—no silent skips;
3. race tests for all touched concurrency-sensitive packages;
4. semantic catalog equivalence, mutation discrimination, clean up/down/up, and tenant FK/RLS/grant
   verification;
5. installed generated-product compile, migrate, API/worker boot, readiness, representative CRUD and
   not-found behavior, and graceful shutdown;
6. Docker/devbox/direct/GoReleaser sentinel version stamping and locally available reference smoke;
7. authoritative `make ci-container` with exact test accounting;
8. release-contract negative matrix for reused identity, wrong major/commit, dirty tree, missing or
   stale evidence, and invalid bootstrap mode;
9. final Google Gemini Graphify extraction with source/model provenance; and
10. a fresh Fable review that inspects source and evidence without relying on the implementation
    narrative.

Hosted artifact transfer/provenance and the public-proxy non-reuse query are tag-time gates after an
authorized push. No push, tag, PR, or release is part of the local candidate programme.
