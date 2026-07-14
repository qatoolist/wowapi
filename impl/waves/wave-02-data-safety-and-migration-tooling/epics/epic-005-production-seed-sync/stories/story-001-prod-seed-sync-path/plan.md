---
id: PLAN-W02-E05-S001
type: plan
parent_story: W02-E05-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W02-E05-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan deliberately does NOT invent design specifics the source defers — MATRIX
CS-21 is explicit that the "design detail [is] to be ratified in Phase 5, but the acceptance bar is
fixed now," so this plan fixes the acceptance-bar-facing structure (what must be proven) and routes
every deferred design question through T001's investigation (what must be determined during the
story, per mandate §18). T001's documented decisions become a recorded plan revision — not a silent
rewrite of this file.

## Proposed architecture

**Confirmed-by-source skeleton:** a seed-sync command path (CS-21's sketch: `wowapi seed sync --env
prod`) consuming versioned catalog manifests, executing idempotent and RLS-respecting catalog
writes, offering dry-run, producing an audit record; plus a readiness check registered through the
existing fail-closed readiness mechanism (`kernel/httpx/health.go:52-79` runs each check with a 3s
timeout, 503 on any failure — CS-21 evidence refinement) that fails with a named check until sync
has run and reports the seed/catalog hash afterward.

**Deferred-by-source detail:** everything inside that skeleton — manifest format, versioning,
idempotency mechanism, role posture, hash computation, audit shape — is T001's output. This plan
does not pre-draw component boundaries beyond the skeleton.

## Implementation strategy

1. **T001 (design investigation) first, gating everything else.** Re-confirm the current-state
   absence at the actual start commit (no seed-sync path; readiness silent on empty catalogs;
   inventory the existing dev/test seed mechanism for reuse potential). Resolve each question in
   "Unresolved questions" below with a documented decision + rationale in a design document.
   Escalate any D-0N-caliber decision per the epic's process safeguard. Record the resulting task-
   breakdown confirmation (or revision) as a plan revision.
2. Capture the fail-first evidence: boot prod-profile against an empty catalog DB at the pre-fix
   commit; record today's silent-ready, deny-everything behavior (CS-21 fail-first).
3. **T003**: implement the catalog manifest schema + parser/validator per T001's ratified format;
   schema-invalid manifests are rejected before any write.
4. **T002**: implement the seed-sync core path — idempotent apply per T001's mechanism, under
   T001's documented role posture, with dry-run mode producing a change plan without writes.
5. **T005**: implement audit-record production per T001's integration decision.
6. **T004**: wire readiness — named check failing until sync has run; seed/catalog hash computed
   and reported in the readiness payload per T001's hash design; prove the pass-after half against
   the step-2 fail-first baseline.
7. **T006**: independent review (P0-prod, mandate §14).

(T003 before T002 in execution order because the sync core consumes the manifest parser; the task
numbering follows the story's functional decomposition, not execution order — execution order is
governed by the dependency column in `tasks/index.md`.)

## Expected package or module changes

A new seed-sync command in the CLI layer (`internal/cli/`-adjacent — exact location per T001); a
new catalog-manifest schema/parser package; a readiness-check addition at the `app/health.go`
`extra`-supplied-checks seam and/or the readiness template (per CS-21's "contract-by-comment at the
seam + template omission at the product end" diagnosis — exact wiring point per T001); audit
integration per T001 (expected: `kernel/audit` usage, not a new subsystem).

## Expected file changes where determinable

Not determinable beyond the seam citations above (`kernel/httpx/health.go` readiness mechanics are
consumed, likely not modified; `app/health.go`'s check-supply seam is the expected registration
point). Every concrete file path is a T001 output. Listing speculative paths here would violate
mandate §8.5's instruction not to invent precise changes the repository does not yet support.

## Contracts and interfaces

- The catalog manifest schema — the story's central new contract; format per T001.
- The readiness-check contract — existing (`kernel/httpx/health.go` check functions); this story
  adds a check implementation, not a contract change.
- The seed-sync CLI command surface — new; shape per T001.

## Data structures

The manifest's parsed representation; a sync-state/hash record (whether an in-DB table, a catalog-
derived computation, or another mechanism is a T001 decision — see "Unresolved questions" on hash
computation and pre-existing-database detection).

## APIs

No HTTP API changes beyond the readiness payload gaining the seed/catalog-hash field (additive).

## Configuration changes

None confirmed. If T001's design needs configuration (e.g. manifest search path, sync-role
credentials source), it follows the existing `kernel/config` conventions — recorded as a T001
output.

## Persistence changes

Possibly a sync-state/hash table (T001 decision). Catalog tables themselves are written by the sync
but not schema-changed by this story.

## Migration strategy

If T001 concludes a new table is needed, its migration is authored per existing conventions (and
classified against DATA-09's manifest schema if W02-E01-S001 has landed — a timing convenience, not
a dependency). Otherwise none.

## Concurrency implications

Two concurrent sync runs against the same database must not corrupt catalogs or the sync-state
record — the idempotency mechanism T001 selects must address concurrent invocation (advisory lock,
CAS on the sync-state record, or equivalent). Explicitly in T001's scope.

## Error-handling strategy

Fail closed throughout: schema-invalid manifest → reject before any write; partial sync failure →
no readiness pass (the hash/state record must only reflect a completed sync); readiness check
errors → 503 per the existing mechanism. Dry-run must be side-effect-free even on error paths.

## Security controls

The RLS/role posture (RISK-W02-E05-002) is the central control: T001 must document which role the
sync runs as and why that does not undermine tenancy controls — a silent superuser bypass fails
independent review. The audit record and dry-run mode are the compensating controls CS-21's own fix
sketch includes.

## Observability changes

The named readiness failure (replacing today's silent deny-everything) and the seed/catalog hash in
the readiness payload are themselves the story's observability deliverables. Sync runs log manifest
version and outcome; the audit record is the durable trail.

## Testing strategy

- **Fail-first**: prod-profile boot against empty catalogs at the pre-fix commit — silent-ready
  captured as the "before" artifact (CS-21: "currently silently deny-everything").
- Idempotency: repeat-run test — second run converges, no spurious writes.
- Dry-run: no-writes assertion against an unsynced database.
- Manifest validation: accept/reject pair (schema-valid and schema-invalid fixtures).
- RLS posture: test verifying the sync runs under the documented role and tenant-table RLS
  enforcement is preserved.
- Readiness: post-fix boot against empty catalogs → named 503 until sync; after sync → ready with
  hash reported.
- Audit: audit-row assertion per sync run.

## Regression strategy

The readiness check itself is the durable regression guard for the deny-everything failure mode.
Existing boot/readiness tests must stay green for dev/test profiles (the named failure is
prod-profile behavior); an already-populated database must continue to pass readiness (the
pre-existing-database predicate is a T001 question precisely because of this regression surface).

## Compatibility strategy

Additive: existing healthy deployments keep passing readiness; dev/test profiles unaffected; the
readiness payload change is a new field, not a changed one. wowsociety gains the fix via backport
(PROD-03), out of scope here.

## Rollout strategy

Single story landing. The readiness gate activates for prod-profile boots at upgrade; a deployment
upgrading with already-populated catalogs must pass the T001-designed predicate without operator
action — this is an explicit design constraint on T001, not an afterthought.

## Rollback strategy

Reverting the story's commits restores today's behavior (no sync path, silent readiness). The sync
itself is additive data-writing; a synced catalog does not need "un-syncing" on rollback. The
sync-state record (if a table) is dropped by its migration's documented rollback plan.

## Implementation sequence

Steps 1–7 under "Implementation strategy." Step 1 (T001) strictly precedes all implementation;
step 2's fail-first capture precedes step 6's pass-after proof.

## Task breakdown

- **W02-E05-S001-T001** — Design investigation: catalog manifest format, versioning scheme, CLI
  shape, idempotency mechanism, RLS posture, hash design, dry-run format, audit integration.
  Output: design document + recorded plan revision. Gates T002–T005.
- **W02-E05-S001-T002** — Seed-sync core path: idempotent, RLS-respecting sync with dry-run mode.
- **W02-E05-S001-T003** — Versioned catalog manifest schema + parser/validator.
- **W02-E05-S001-T004** — Readiness wiring: named empty-catalog failure, sync gate, seed/catalog-
  hash reporting (includes the fail-first capture).
- **W02-E05-S001-T005** — Audit-record production per T001's integration decision.
- **W02-E05-S001-T006** — Independent review (mandate §14; P0-prod).

## Expected artifacts

The T001 design document; the seed-sync command; the manifest schema/parser; the readiness wiring;
operator documentation. See `artifacts/index.md`.

## Expected evidence

Design-decision record; idempotency + dry-run test outputs; manifest accept/reject outputs; RLS-
posture test output; readiness fail-first/pass-after pair with hash assertion; audit-row assertion.
See `evidence/index.md`.

## Unresolved questions

Each of the following must be resolved by T001 with a documented decision + rationale before any
implementation task begins (mandate §18):

1. **Catalog manifest schema/format** — file format (YAML/JSON/other), field set, one manifest per
   catalog domain vs. a single manifest, and where manifests live in the repository/deployment.
2. **Versioning scheme** — how manifest versions are declared, ordered, and recorded; whether
   partial upgrades between versions are supported or sync is always to-latest.
3. **Seed/catalog hash** — how the hash is computed (over manifest content? over resulting catalog
   rows? canonicalized how — noting DATA-08's jsonb-canonicalization lesson that hashing stored
   forms is non-reproducible), where it is persisted, and where in the readiness payload it is
   reported.
4. **CLI command shape** — whether CS-21's `wowapi seed sync --env prod` sketch is adopted as-is,
   and how it relates to the existing seed/dev tooling surface.
5. **Idempotency mechanism** — upsert semantics vs. content-hash-based skip vs. full
   diff-and-apply; and how concurrent invocation is made safe.
6. **RLS posture** — which role the sync runs as, and how "RLS-respecting" is honored in a
   bootstrap context where the catalogs RLS presupposes are empty (cf. DATA-01 T3's analogous
   platform-role-bypass caution: "Requires a platform-role connection to bypass RLS for the scan" —
   the same class of tension, RISK-W02-E05-002).
7. **Pre-existing-database predicate** — how a deployment whose catalogs were populated before this
   feature existed satisfies the readiness gate without operator intervention.
8. **Dry-run output format** — what the change plan shows and in what form (human-readable,
   machine-readable, both); whether dry-runs are themselves audited.
9. **Audit record shape/location** — whether it reuses `kernel/audit` (the infrastructure
   DATA-06/DATA-08 already exercise) or is separate; what fields it carries (manifest version,
   hash, actor, outcome).

## Approval conditions

This plan is approved for implementation once: (a) T001's design document resolves every question
above with rationale, and the resulting plan revision is recorded; (b) any D-0N-caliber decision
has been escalated per the epic's process safeguard rather than made in-story; and (c) the owner
and reviewer are assigned.
