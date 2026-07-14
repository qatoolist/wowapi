---
id: ART-W02-E05-S001-001
type: design-document
title: Catalog manifest design investigation — decision record
parent_story: W02-E05-S001
task: W02-E05-S001-T001
status: complete
created_at: 2026-07-13
updated_at: 2026-07-13
base_commit: 1626b1132622aacc3e85475e4190e16a457ad1f6
---

# W02-E05-S001-T001 — Design decision record

Produced BEFORE any T002–T005 implementation began (this wave runs under a
single-conductor-commit model; sequencing within the worker session is attested by this record's
creation preceding every code change listed in `../../implementation.md`, and by the fail-first
evidence in `../../evidence/`). Base commit re-confirmed: `1626b113`.

## Current-state re-confirmation (T001 step 1, at 1626b113)

The CS-21 assessment is PARTIALLY stale; re-confirmed state:

- `kernel/seeds` exists: `Load` (strict YAML, module-prefix ownership) and `Sync` (idempotent
  ON CONFLICT upserts of permissions / resource_types / relationship_types / roles +
  role_permissions reconcile), platform-privileged by convention (SEC-13/D-0026).
- `wowapi seed sync` exists (`internal/cli/seed_cmd.go`) — an explicitly-labelled low-level
  escape hatch: bare DATABASE_URL, `--module name=dir`, connects `WithSetRole("app_platform")`.
- The generated `cmd/migrate` (init template) runs migrations → `seeds.Sync` → `rules.SyncDefinitions`.
- `app.CatalogsSeeded` (`app/seed_health.go`) is a count-based check: fails when the booted bundle
  declares seeds but a catalog table is empty. It is wired ONLY by the generated `cmd/api`
  template via `app.Readiness`'s `extra` map — the framework seam itself (`app/health.go`)
  registers nothing (contract-by-comment, exactly as CS-21's evidence refinement located it).
- `kernel/httpx/health.go:52-79` readiness runner confirmed sound and fail-closed (3s/check, 503).

**Confirmed absent (the FBL-02 gap at this commit):**
1. No manifest **version** concept anywhere in the seed path.
2. No **seed/catalog hash** — computed, persisted, or reported.
3. No **dry-run** mode.
4. No durable **audit record** of a sync run.
5. No **sync-state record** — readiness infers "synced" heuristically from row counts only.
6. Framework-level readiness does not register any catalog check itself: a product (or the
   framework's own default assembly) calling `app.Readiness(b, fp, nil)` on a prod-profile boot
   with empty catalogs reaches `ready` silently — the CS-21 fail-first defect, re-confirmed.

## Decisions

### Q1 — Catalog manifest schema/format
**Decision:** The catalog manifest IS the existing per-module seed YAML bundle (`kernel/seeds`
schema: `permissions`, `roles`, `resource_types`, `relationship_types`), extended with one new
OPTIONAL top-level field `version: "<label>"` (string). One module directory = one manifest; at
most one file per module may declare a non-empty `version` (two conflicting labels fail `Load`).
Manifests live where seeds live today: embedded module FSes (`mc.Seeds(...)`) for the production
path, directories for the CLI escape hatch.
**Rationale:** The framework already has a declarative, strict-decoded, ownership-validated
catalog format with two production consumers (boot, migrate). Inventing a second manifest format
would create the parallel-convention violation the programme forbids; extending the existing one
gives versioning + hashing to every existing deployment with zero migration of seed content.
Rejected alternatives: separate `manifest.yaml` registry file (second file to keep in sync with
the content it describes); consolidated single manifest (breaks per-module ownership validation).

### Q2 — Versioning scheme
**Decision:** Content-addressed. The authoritative manifest version is the canonical content hash
(Q3); the optional `version:` label is an operator-facing annotation recorded alongside, never
compared. Sync is always whole-manifest to-latest; partial upgrades are not supported.
**Rationale:** Catalog sync is convergent (upsert + reconcile), so ordering/monotonicity carries
no meaning — only identity does. A content hash cannot drift from content (a declared version
can); the acceptance bar needs exactly an identity ("the readiness payload reports the
seed/catalog hash"). Labels stay because deploy pipelines want a human-readable marker.

### Q3 — Seed/catalog hash
**Decision:** `seeds.Hash(Bundle)` = SHA-256 (hex) over a canonical JSON serialization of the
PARSED, NORMALIZED bundle: entries sorted by key, role grant-lists sorted, version labels
EXCLUDED (content identity only). Persisted in the new `seed_sync_runs` table (Q9) by every
successful apply; reported in the readiness payload as `details.seed_catalog_hash`, read back by
a single indexed-row lookup (well inside the 3s check budget — never a catalog re-scan).
**Rationale:** Hashing the parsed-normalized form (not file bytes, not stored rows) is precisely
the DATA-08 jsonb-canonicalization lesson: whitespace/comment/ordering churn must not change the
identity, and stored forms are not reproducibly serializable. Label exclusion keeps "content
unchanged, label bumped" a no-op (idempotency follows content, not annotation); the run record
still logs the new label (Q9).

### Q4 — CLI command shape
**Decision:** Keep `wowapi seed sync` (CS-21's sketch adopted minus `--env`), extended with
`--dry-run`. The `--env prod` flag is NOT adopted: this binary intentionally has no product
config layering, so an `--env` flag would be theater — environment comes from the product config
in the real production path, which remains the generated `cmd/migrate` (now calling the new
recorded apply, Q5). Exit codes unchanged: 0 success/no-op, 1 failure, 2 usage.
**Rationale:** Two command surfaces already exist with documented roles (escape hatch vs.
generated migrate); this story upgrades both rather than adding a third.

### Q5 — Idempotency mechanism
**Decision:** Two-layer. (1) Existing per-row `ON CONFLICT` upserts + grant reconcile (kept as
`seeds.Sync`, the write primitive). (2) New orchestration entrypoint `seeds.Apply(ctx, db, b,
opts) (Report, error)`: computes the content hash, wraps the whole run in one transaction
guarded by `pg_advisory_xact_lock` (Q5b), short-circuits to outcome `noop` — no catalog writes at
all — when the latest recorded successful run has the same hash AND the declared catalogs are
non-empty, otherwise runs `Sync` and appends the run record in the same transaction (atomic:
partial failure rolls back state with writes — fail-closed).
**Rationale:** The hash short-circuit makes "second run is a no-op" provable in the strongest
form (row `xmin`s untouched, not merely value-identical after rewrite), and makes the state
record trustworthy: it commits iff the writes did.

### Q5b — Concurrent invocation
**Decision:** IN SCOPE. Concurrent `Apply` calls serialize on a constant
`pg_advisory_xact_lock` inside the apply transaction; the loser observes the winner's committed
state (typically no-ops). Callers passing an already-open transaction get the same lock; a bare
non-transactional DBTX is not a supported Apply surface (all real callers hold a pool or tx).
**Rationale:** Rolling deployments genuinely race migrate jobs; an advisory xact lock costs one
row-less statement and removes the race entirely rather than documenting it away.

### Q6 — RLS/role posture (RISK-W02-E05-001/002)
**Decision:** `app_platform` — the existing dedicated catalog-maintenance role — via
`database.WithSetRole("app_platform")` + `WithConnRLSGuard()` (CLI), or the migrate process's
schema-owner pool (generated migrate; already today's posture). NEVER a superuser/BYPASSRLS DSN.
"RLS-respecting" is honored structurally: the sync writes ONLY the five global catalog tables
(permissions, roles, role_permissions, resource_types, relationship_types) plus the append-only
`seed_sync_runs` — none of which is tenant-scoped, so no tenant-data RLS policy is bypassed
because none applies to what the sync touches. `app_platform` remains RLS-BOUND on every
tenant-scoped table (policies are role-unqualified; the role has no BYPASSRLS), so even a
compromised manifest cannot reach tenant rows through this path — the SQL surface is fixed, and
the role's tenant-table access is policy-gated exactly as before. The bootstrap tension resolves
cleanly: empty catalogs configure AUTHORIZATION (deny-everything at the app layer), not database
RLS — tenancy isolation is enforced by Postgres policies that exist from migration time,
independent of catalog content, so populating catalogs under app_platform cannot weaken tenancy.
**Adversarial proof (T002 test):** assert the sync connection is neither superuser nor BYPASSRLS
(`pg_roles`), assert cross-tenant isolation on a tenant table under app_rt is still enforced
after a sync, and assert `seed_sync_runs` is not writable by app_rt.

### Q7 — Pre-existing-database predicate
**Decision:** The readiness check keeps `CatalogsSeeded`'s populated-catalog predicate as the
gate (declared-but-empty catalog table → named failure), with the recorded hash as an ADDITIVE
report: latest successful `seed_sync_runs` row → `details.seed_catalog_hash`; no row (deployment
populated before this feature) → check still passes on populated catalogs, hash omitted until
the next sync records one. Zero operator action on upgrade.
**Rationale:** Gating on hash presence would 503 every healthy pre-feature deployment at
upgrade — exactly the regression the plan's rollout constraint forbids.

### Q8 — Dry-run output format; dry-run auditing
**Decision:** Human-readable deterministic change plan on stdout: per-catalog-table counts
(`insert/update/unchanged`, plus role-grant add/prune) followed by per-key lines, computed by
read-only diffing; manifest hash printed. Dry-runs are NOT audited and write NOTHING — full
side-effect freedom (including on error paths) outweighs an audit trail of intents; the run log
line is the dry-run's trace.
**Rationale:** The plan's own error-handling strategy demands side-effect-free dry-runs; an
audit INSERT is a side effect. Machine-readable output deferred until a consumer exists.

### Q9 — Audit record shape/location
**Decision:** NOT `kernel/audit`. A new append-only global table `seed_sync_runs`
(migration `00042_seed_sync_runs.sql`, symmetric Down): identity PK, `manifest_hash`,
`version_label`, `actor`, `outcome` (`applied`|`noop`|`failed`), `counts` jsonb,
`error` text, `created_at`. Grants: `app_platform` SELECT+INSERT only (append-only even for the
writer, mirroring `audit_anchors`); no app_rt access. The latest `applied`/`noop` row doubles as
the sync-state record (Q3/Q5). Failed runs are recorded best-effort AFTER rollback (outside the
aborted tx); the readiness gate reads only successful outcomes.
**Rationale:** `kernel/audit.Writer.Record` requires a TenantDB and extends a per-TENANT hash
chain; seed-sync is a global, tenant-less operation — forcing a synthetic tenant would corrupt
the chain's semantics. The `audit_anchors` pattern (global, append-only, platform-written) is
the established precedent for exactly this shape. This is a story-scoped table, not a D-0N
convention: no escalation required (checked against epic.md's safeguard — no new framework-wide
convention, no new dependency).

### Q10 — Readiness check environment scope
**Decision:** The framework-supplied check registers in ALL environments (new assembly
`app.ReadinessWithCatalogs`, which the generated api template now uses; check name
`seed_catalogs`). CS-21's bar names prod-profile because that is where the gap is blocking, not
because dev should be exempt — a dev boot with declared-but-unsynced seeds is equally broken,
and `CatalogsSeeded` has never been env-conditional. Dev/test remain unaffected in practice
(their lifecycle runs migrate→seed; a no-seed product passes vacuously). The fail-first pair is
captured under a prod-profile boot per the bar's wording.

## Task-breakdown confirmation

T002–T005 confirmed as planned; execution order T003(manifest/hash)→T002(apply)→T005(audit —
folded into Apply per Q5/Q9, its table and assertions land with T002/T003's package)→T004
(readiness + hash reporting). Recorded as a plan revision in `../../deviations.md`. No D-0N
escalation needed (Q9 rationale).
