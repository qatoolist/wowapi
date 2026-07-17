# Sixth review — Comprehensive baseline/release-readiness review (2026-07-17)

Source: ComprehensiveReport.md (untracked, user-owned). Verdict: APPROVED as a
remediation/baseline work order; NO-GO for release or clean-V1 cutover at
8b89412. Four classes of work: (1) open correctness blockers C-01..C-11;
(2) a release-identity decision; (3) V1-compatibility residue cleanup; (4)
disconnected/incomplete V2 surfaces.

## Verified against source before acting

Spot-checked and confirmed real: C-01 (registry invalidated before the sealed
check), C-04 (all resolvers in a step share one context copy), C-05
(newArtifactWriter falls back to retention.TestKey() on missing/invalid key),
C-08 (Dockerfile/devbox ldflags used the pre-/v2 buildinfo path), C-06
(compatibility-gates.yml + ci.yml baseline default v1.0.0).

## Remediation this round (commit 7): self-contained correctness blockers

Fixed the code-level defects that hold regardless of the release-identity
decision, each with a discriminating regression:

- C-01: registration invalidates validation ONLY on a successful mutation —
  a rejected duplicate/invalid registration and a recovered post-seal panic
  leave generation, contents, and validation unchanged
  (TestRejectedMutationsDoNotInvalidate).
- C-02: resolveValidated checks validation AND resolves the graph under one
  read lock — the executed graph is covered by a single validated generation;
  defForInstance uses it (TestResolveValidatedIsAtomicWithValidation).
- C-04: each assignee resolver invocation receives its OWN deep canonical
  context copy (TestResolversEachGetIsolatedContext).
- C-05: a missing/malformed DSR artifact key FAILS BOOT in production instead
  of silently using the deterministic test key; non-prod keeps the warned
  convenience (TestArtifactWriterFailsClosedInProd).
- C-08: Dockerfile and devbox entrypoint ldflags stamp the /v2 buildinfo
  path (.goreleaser.yaml already correct from the cutover).

Gates: full host suite, ci-container, mechanical batch green.

## NOT done autonomously — requires the user's decision / larger programme

- **C-03 (High)** persisted-definition identity: the smallest safe fix
  (persist+verify a definition digest, or execute the persisted snapshot) is a
  SCHEMA change that the report couples to the migration-baseline decision.
  Deferred pending that decision; the round-5 parseAndValidateDefinition
  already validates persisted graphs against the current callback sets, so a
  persisted def can never execute unvalidated — the residual gap is
  registry-vs-persisted divergence for the same (key,version), which needs the
  digest column.
- **Release identity (section 5, blocking):** v1.1.0 is published in the Go
  proxy and cannot be reused; v1.0.0 is reserved pending proxy preflight. The
  choice — root-module higher-V1, a new module path for a fresh v1.0.0, or
  keep /v2 — is the user's and must precede any second module-path rewrite,
  or the repo is rewritten twice.
- **CI single-owner redesign (C-06/09/10/11):** the required-gates matrix runs
  alongside duplicate native legs; release-gate evidence declarations don't
  bind real artifacts; exact-tag compatibility wiring is absent. A CI/release
  workflow programme, best done after the identity decision.
- **V1-residue cleanup (L-01..L-18, D-01..D-11, M-*, R-*):** forwarding
  shims, parallel constructors, unsigned cursors, ignored claim fields,
  markerless generator rewrite, migration squash, history archive. Many gated
  on product decisions (MFA scope, checksum-repair scope, clean-DB policy) and
  all coupled to the identity decision. A staged programme, not an autonomous
  sweep — the report's own phase 1 is "decide the release identity."

## C-03 lifecycle trace + canonical-source decision (2026-07-17)

Exhaustive trace of the workflow_definitions lifecycle (per the sequencing
correction — C-03 precedes the baseline):

- **Registration/declaration:** in-memory `workflow.Registry` (boot-validated,
  RWMutex + generation-keyed). Source of truth for the graph.
- **Seed/sync source:** NONE in production. `kernel/rules.SyncDefinitions`
  syncs the rule registry → `rule_definitions` (wired in the generated migrate
  main, cmd_migrate_main.go.tmpl:193). There is NO `workflow.SyncDefinitions`
  and the generated migrate main does not sync workflow definitions.
- **Production insert/update ownership:** NONE. `00009_workflow.sql:72` GRANTs
  INSERT/UPDATE on workflow_definitions to app_platform, but no code writes it.
  Only `testkit.SeedWorkflowDefinition` inserts rows (test-only).
- **Instance creation (StartIn):** reads `SELECT id, version FROM
  workflow_definitions WHERE key=$1 ORDER BY version DESC` (definitionRow) —
  would return NO rows in a real product, failing StartIn.
- **Runtime reload (defForInstance):** `SELECT key, version, definition WHERE
  id=$1`, prefers the registry def, falls back to parseAndValidateDefinition.
- **SLA sweep:** batch-reads definitions the same way.
- **Generated-product provisioning:** nothing syncs workflow definitions.

**Decision (canonical source of truth):** registered module workflow
definitions are canonical, and the production writer is a MISSING INTEGRATION,
not a testkit gap. Tests mask it by seeding rows directly. C-03 is therefore:
(1) implement `workflow.SyncDefinitions(registry → workflow_definitions)`
mirroring rules, computing an immutable canonical digest at sync; (2) wire it
into the generated migrate main after rules sync; (3) add a definition_digest
column (schema delta folded into the baseline); (4) verify the digest at
StartIn, reload, task execution, and SLA sweep, rejecting before state
mutation on missing/mismatched identity; (5) cross-tenant and
same-key/version-divergent regressions.

Evidence retained: migrations/baseline/census-reference.txt (framework-owned
schema inventory of the proven 49-chain head, with PG/extension versions,
source commit, normalization rules; regenerated + drift-guarded by
scripts/baseline_census.sh via `make baseline-census-check`).


## Design-blocker resolutions (2026-07-17)

### Blocker 1 — tenant workflow-definition overrides: REMOVE (clean-V1 decision)

Verified: `workflow_definitions.tenant_id` exists (NULL=module template,
non-NULL=tenant override) and `definitionRow` prefers overrides
(`ORDER BY (tenant_id IS NOT NULL) DESC`). But there is NO production writer for
overrides (no SyncDefinitions; only testkit seeds), the registry has no tenant
dimension, and no product requirement is evidenced. A global SyncDefinitions
cannot produce or validate tenant-specific definitions.

DECISION: remove tenant workflow-definition overrides. Registered module
definitions are the SOLE canonical source; `workflow_definitions` becomes a
global platform catalog like `rule_definitions`. Schema delta (folded into the
baseline): drop `workflow_definitions.tenant_id` and its RLS policy; simplify
`definitionRow` to `WHERE key=$1 AND status='active' ORDER BY version DESC`.
The instance→definition binding stays immutable via `workflow_instances.
definition_id` (FK) plus the C-03 digest.

### Blocker 2 — census strengthened into a semantic manifest

The first census counted objects only; it could not prove object-for-object
equivalence and wrongly counted 238 functions (98% extension-provided). The
strengthened census (scripts/baseline_census.sql, shared by the check and the
mutation harness) captures: extension name+version+schema; column full type +
identity + generated + collation; constraint full def (FK actions, referenced
cols, check exprs) + validated/deferrable/deferred; policy cmd + permissive +
roles + using + check; function signature + return + language + volatility +
strict + security-definer + config + body-md5, EXCLUDING extension-provided
functions via pg_depend (now 3 framework functions, not 238); grants across
table/column/sequence/function/schema with grant-option. Concurrency-safe
(PID-suffixed scratch DB, mktemp).

NEGATIVE PROOF (make baseline-census-discriminates,
scripts/baseline_census_discriminates.sh): mutating one object per class —
FK deferrability, policy role, function security-definer, an added grant, an
added extension, a generated column — changes the manifest in every case. The
oracle is therefore genuinely discriminating, not an "equal counts" false pass.
`make baseline-census-check` is now sound as the squash equivalence guard.

## C-03 implementation plan (scoped, turnkey — 2026-07-17)

Prerequisites cleared (tenant-override decision, discriminating census). The
C-03 coding unit, in dependency order, each DB-verified:

1. Migration 00050_workflow_definition_identity (folds into the baseline):
   - DROP POLICY workflow_definitions_tenant; DISABLE + NO FORCE RLS (becomes a
     global platform catalog, matching rule_definitions).
   - DROP INDEX workflow_definitions_key; DROP COLUMN tenant_id;
     CREATE UNIQUE INDEX workflow_definitions_key ON (key, version).
   - ADD COLUMN definition_digest text NOT NULL DEFAULT ''.
   - Reversible Down restores tenant_id + the COALESCE unique index + RLS +
     policy exactly as 00009 (verify with the reversibility drill).
2. Canonical serialization + digest (kernel/workflow): deterministic JSON of
   the validated Definition (sorted keys) -> sha256 hex. Stable across
   marshal/unmarshal.
3. workflow.SyncDefinitions(ctx, db, reg) mirroring rules.SyncDefinitions:
   upsert each registered def with its digest. Contract:
   - (key,version) immutable; re-sync of same digest = idempotent no-op;
   - a DIFFERENT digest for an existing (key,version) FAILS LOUDLY (never
     overwrite) — semantics change requires a new version;
   - existing row IDs stable (workflow_instances reference them);
   - definitions absent from a later binary are NOT auto-deleted while
     instances may reference them.
4. Wire SyncDefinitions into the generated migrate template after rules sync
   (cmd_migrate_main.go.tmpl), and into testkit setup so tests get real rows.
5. ONE shared verified-definition loader used by StartIn, Decide, CompleteTask,
   Delegate, Override, and SweepSLA (replace the ad-hoc resolveValidated /
   parseAndValidateDefinition calls): resolves the graph, recomputes its
   canonical digest, compares to the persisted row's definition_digest, and
   REJECTS before any state mutation on missing or mismatched identity. No
   fallback executes persisted JSON absent from the canonical registry (no
   historical-execution policy in clean V1). Simplify definitionRow to
   `WHERE key=$1 AND status='active' ORDER BY version DESC` (tenant_id gone).
6. testkit.SeedWorkflowDefinition: drop the tenant param, compute+store the
   digest (or route through SyncDefinitions); update its 6 callers.
7. Regressions: sync idempotent-same-digest; sync divergent-digest-fails-loud;
   load rejects a digest mismatch before mutation; StartIn/Decide/CompleteTask/
   SweepSLA all go through the shared loader (one check, not four); missing
   definition rejected.
8. Regenerate migrations/baseline/census-reference.txt for the 00050 delta.

## Census completion (before baseline generation — reviewer's remaining gaps)

Add before generating 00001_baseline: a declared framework-owned schema set
(>= public + migration); enumerate all relations in those schemas incl.
migration.backfill_checkpoint (columns, composite PK, defaults, indexes),
sequences, views/matviews/partitioned/foreign tables; triggers + rules;
enum/domain/range/composite user types; role existence + intended membership
(normalizing LOGIN creds); pg_default_acl; ACL via aclexplode with object
identity + overloaded-function signatures + PUBLIC + grantor + grantee +
privilege + grant-option; fuller function attrs (kind, leakproof, parallel,
cost, rows). Add negative tests for: migration checkpoint table, a sequence, a
view, a trigger, a custom type, a default privilege, an overloaded-function
grant, and a PUBLIC grant/revoke. Remove 2>/dev/null from the gate (preserve
stderr for diagnosis). The largest current omission: baseline_census.sql is
public-only and never inventories the migration schema — a baseline could drop
the entire checkpoint schema and still pass.
