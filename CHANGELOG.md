# Changelog

wowapi follows [Semantic Versioning](https://semver.org) from its first supported clean-line
release. The module is `github.com/qatoolist/wowapi`.

The published `v1.0.0` and `v1.1.0` identities are abandoned and immutable. They are not supported
predecessors, compatibility baselines, or database-upgrade sources and will never be reused. The
first supported release is `v1.2.0`; compatibility commitments begin there.

## [Unreleased — target v1.2.0]

### Clean baseline

- Establishes the root Go module and opaque, boot-validated `app.Booted` runtime view.
- Treats registered module declarations as the canonical application definition and seals every
  extension registry after successful boot.
- Generates products whose imports all derive from `buildinfo.ModulePath`; a repository lint rejects
  hard-coded framework import paths in templates.
- Defines `v1.2.0` as the only clean-line bootstrap release. Later releases require an explicit older
  supported predecessor and exact-tag compatibility evidence.
- Removes support for abandoned-release APIs, generated-project layouts, databases, object metadata,
  cursors, claim shapes, migration replay, and compatibility allowlists.

### Correctness and reliability

- Serializes retry schedule state and rejects nil backoff policies.
- Supervises critical workers and process hooks: asynchronous child failure cancels siblings,
  performs bounded reverse-order shutdown, and is returned to the process.
- Uses tenant-bound, generation-fenced migration checkpoint leases with monotonic writes.
- Makes bulk lifecycle transitions transactional and prevents cancelled operations from regaining
  pending items through recovery or reclaim paths.
- Binds document confirmation to the reserved document, version, key, checksum, expiry, and active
  document state. Upload hooks receive an explicit transaction and retry-stable delivery ID.
- Runs storage I/O outside document row locks and uses checksum-aware uploads as the sole storage
  contract.
- Requeues failed outbox work on its own due schedule under sustained traffic and surfaces bounded
  requeue failure.
- Reports per-tenant recurring-maintenance failures without stopping attempts for other tenants.
- Makes invalid pagination defaults fail loudly and accepts only signed opaque cursor envelopes.
- Requires explicit job idempotency, actor attribution, credential schemes for restricted
  permissions, request contracts for mutating routes, and exact webhook tenant identity.
- Rejects missing or malformed production DSR artifact keys at boot; deterministic convenience keys
  are private to non-production composition and test code.

### Workflow identity

- Makes registered workflow definitions the sole canonical source; tenant definition overrides are
  not part of the clean model.
- Adds deterministic canonical JSON and SHA-256 definition identity, atomic definition
  synchronization, and a single verified loader for start, task mutation, override, and SLA paths.
- Rejects missing, malformed, divergent, or registry-absent definitions before any workflow state or
  side effect changes.
- Isolates callback and resolver inputs with canonical deep copies and preserves exact JSON numeric
  semantics.
- Removes public workflow fields and step types that had no implemented state-machine behavior.

### API and repository consolidation

- Removes the nine forwarding `kernel/*` domain shims; canonical domain services live under
  `foundation/*`.
- Consolidates lifecycle onto `app.Hook`/`app.RunHooks`, upload initiation onto the checksum-bound
  method, retention onto compliance-aware result-bearing operations, and workflow construction onto
  the compliance-aware runtime.
- Removes disconnected typed-port, app-model projection, advisory lifecycle-manifest, DB-i18n mode,
  dead CLI branch, comparability-only tests, markerless generator rewrite, and checksum-repair
  surfaces.
- Moves transport fakes to `testkit/fakes` and makes chaos harnesses test-only.
- Generates semantic `uuid.UUID` and `time.Time` CRUD fields.
- Keeps independently necessary foundation capabilities: boot validation/sealing, tenant
  transactions and RLS, audit, outbox/recovery, online migration machinery, observability,
  localization fallback, MFA primitives, generated consumers, and future compatibility engines.

### Database and release operations

- Replaces historical upgrade choreography with a clean current-state migration baseline, verified
  by a semantic catalog manifest covering schemas, relations, columns, constraints, indexes, RLS,
  policies, functions, types, triggers, sequences, extensions, roles, and ACLs.
- Proves the catalog oracle with negative mutations for every supported object class.
- Gives each CI gate one execution owner. Release evidence is hashed and bound to the exact source,
  manifest, command, status, and candidate commit; missing, stale, or modified evidence fails.
- Path-scopes expensive pull-request jobs while protected-main, merge-queue, scheduled, and release
  execution retain the complete required set.

### Documentation

- Reconciles the README, SRS, blueprints, operations policy, generator output, invariant ledger, and
  review report around the clean `v1.2.0` baseline.
- Retains independent protocol and data-format version fields; a `v1`/`v2` token alone is never used
  as evidence that a capability is legacy.

[Unreleased — target v1.2.0]: https://github.com/qatoolist/wowapi/compare/v1.2.0...HEAD
