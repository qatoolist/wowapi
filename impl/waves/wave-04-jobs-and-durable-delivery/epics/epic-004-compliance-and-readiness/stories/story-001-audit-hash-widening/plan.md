---
id: PLAN-W04-E04-S001
type: plan
parent_story: W04-E04-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E04-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information beyond what `kernel/audit/audit.go`'s cited line ranges already confirm.

## Proposed architecture

`chainHash`'s field-coverage widening and the `hash_version` discriminator are additive to
`kernel/audit`'s existing chain-of-custody design — no replacement of the hash-chain concept itself,
no change to `Anchor`/`CheckAnchor`'s tail-truncation guard beyond making it version-aware where it
depends on the hash input set. The core architectural addition is: (1) a canonicalization function
producing a reproducible pre-serialization form of `metadata` for hashing, distinct from the stored
jsonb; (2) the `hash_version` column and its version-branch dispatch inside `Verify`.

## Implementation strategy

1. Re-read `kernel/audit/audit.go:130-311` fresh at this story's actual start commit to confirm the
   current-state assessment (15-field `chainHash`, `metadata`/`tx_id` excluded, no `hash_version`
   column, `Verify`/`Anchor`/`CheckAnchor` behavior) still holds.
2. Design the `metadata` canonicalization function: a deterministic pre-serialization form (e.g.
   canonical key-ordering, stable numeric/string encoding) that reproduces identically across reads,
   distinct from the stored jsonb's own round-trip-unstable form. Document the chosen canonicalization
   approach and its rationale.
3. Confirm `tx_id`'s exact field representation as currently stored (via `pg_current_xact_id()` at
   `:140`) and how it is included in the widened hash input.
4. Confirm every other currently-in-scope field (all nullable fields, sequence, ID, timestamps,
   previous hash) is genuinely included in the widened `chainHash`, not merely assumed included
   because it was in the original 15-field list — re-derive the full field list from the actual code,
   not from this plan's own paraphrase of it.
5. Choose the exact `hash_version` value for the new scheme (D-04 reserves `1` for the historical
   scheme; the new scheme's value, e.g. `2`, is an implementation-time choice) and document it.
6. Implement the `hash_version smallint NOT NULL DEFAULT 1` column migration, shipped through
   W02-E01's online-migration protocol: classify it via the manifest schema (online/maintenance,
   lock/statement timeout, N/N-1 compatibility flag, backfill owner if applicable, validation query,
   rollback/forward-fix plan), and run it through expand/backfill/validate/contract phases as
   applicable to a single additive column with a default value.
7. Implement the widened `chainHash` (including canonicalized metadata and tx_id) alongside the
   version-branch dispatch in `Verify`: `hash_version = 1` rows verify under the original 15-field
   scheme; new rows verify under the widened scheme.
8. Write the per-field tamper test: for each declared field (all fields in the widened scheme,
   independently), mutate it on a chained row and assert verification fails. This must exercise every
   field independently, not a single combined mutation across several fields at once.
9. Write the version-branch verification test: confirm a `hash_version = 1` row created before this
   story's change still verifies correctly under the v1 branch, and a new row created after the change
   verifies correctly under the v2 branch.
10. Document the widened field list, the canonicalization approach, the `hash_version` value chosen,
    and the version-branch semantics.

## Expected package or module changes

`kernel/audit` (the `chainHash`, `Verify`, `Anchor`/`CheckAnchor` functions and their supporting
types); a new migration file adding the `hash_version` column, authored and classified per W02-E01's
protocol tooling (manifest schema, lock-timeout mechanism).

## Expected file changes where determinable

- `kernel/audit/audit.go` — widen `chainHash`'s field-coverage list (lines around `:130-179`); add
  the metadata-canonicalization function; add version-branch dispatch to `Verify` (`:195-248`);
  confirm whether `Anchor`/`CheckAnchor` (`:253-311`) require any version-awareness of their own.
- A new migration file (exact path/number TBD, following the repository's existing migration
  numbering convention) adding `hash_version smallint NOT NULL DEFAULT 1`, with a manifest entry per
  W02-E01-S001's schema.
- New tamper-test and version-branch-verification test files under `kernel/audit`'s existing test
  package.

## Contracts and interfaces

`chainHash`'s function signature may change if the widened field set requires additional parameters
(e.g. the canonicalized metadata form) — exact signature TBD at implementation time; not assumed to
be call-shape-breaking for wowsociety per the "Compatibility considerations" in `story.md` (internal
hash computation change only, no `kaudit.Writer`/`Record` call-shape change expected). `Verify`'s
public contract gains version-branch dispatch internally; its own external call shape is not expected
to change.

## Data structures

The audit row's own struct gains a `HashVersion` field (or equivalent) corresponding to the new
`hash_version` column. No other data-structure change anticipated.

## APIs

None affected at the HTTP/API layer — this story is internal to `kernel/audit`'s hash computation and
verification logic.

## Configuration changes

None anticipated. The `hash_version` new-scheme value is a compile-time/implementation constant, not
a runtime-configurable value, consistent with D-04's own framing (a fixed discriminator scheme, not a
tunable one).

## Persistence changes

The `hash_version smallint NOT NULL DEFAULT 1` column addition to the audit table, shipped via
migration per D-04 and W02-E01's protocol. No other schema change.

## Migration strategy

Per D-04's decision text verbatim: "Add a `hash_version smallint NOT NULL DEFAULT 1` column in the
same migration that widens `chainHash`'s field coverage; verification branches on it." This story's
migration and its `chainHash` widening land as one atomic unit — not the column added in one
migration and the widening implemented separately in a following change, which would create a window
where the column exists but does not yet discriminate anything real. The migration itself runs
through W02-E01's expand/backfill/validate/contract protocol given the confirmed breaking-change risk
against a live-production table (see "Migration considerations" in `story.md`).

## Concurrency implications

None beyond what W02-E01's protocol already handles for the migration itself (lock-timeout budget,
online-DDL classification). The hash computation and verification logic themselves are not expected
to introduce new concurrency concerns beyond what already exists in `kernel/audit`'s current design.

## Error-handling strategy

A verification failure under either version branch must produce a clear, version-identified error —
distinguishing "this row failed v1 verification" from "this row failed v2 verification" so an
operator investigating a failure knows which scheme was in play (see `story.md` "Observability
considerations"). A row with an unrecognized `hash_version` value (neither the historical nor the new
scheme) must fail closed with an explicit error, not silently fall through to one branch or the
other.

## Security controls

The per-field tamper test is itself the acceptance-defining security control — AC-W04-E04-S001-01
requires it explicitly, not a generic tamper test. The canonicalization requirement (never hash the
stored jsonb directly) is also a required security control, not optional hardening, per `story.md`'s
"Security considerations": hashing the stored jsonb would reintroduce non-reproducibility, defeating
the tamper-evidence purpose the widening exists to serve.

## Observability changes

Version-identified verification-failure logging/errors, per "Error-handling strategy" above — an
implementation-time addition, not separately mandated by the source beyond the version-branching
requirement itself.

## Testing strategy

- Per-field tamper test: mutate `metadata`, `tx_id`, and every other declared field independently on
  a chained row; assert every one fails verification. This is the exact test named in the source
  (PLAN DATA-08 W6-T1's own Tests column: "Tamper test: mutate each field independently, assert every
  one fails"), not a substitute or generic tamper test.
- Version-branch verification test: a `hash_version = 1` row (created before this story's change, or
  a fixture representing one) verifies correctly under the v1 branch; a new row (created after the
  change) verifies correctly under the v2 branch.
- Migration-classification test: confirm the `hash_version` migration has a complete manifest entry
  per W02-E01-S001's schema and complies with its lock-timeout budget.

## Regression strategy

The per-field tamper test and the version-branch verification test, once landed, become the
regression guard for this story's own scope: any future change to `chainHash`'s field coverage or to
`Verify`'s version-branch logic that silently drops a field or misroutes a version would be caught by
these tests failing.

## Compatibility strategy

Historical rows (hash_version = 1, implicitly via the column's own default) must continue to verify
correctly under the original 15-field scheme — this is the entire purpose of the version-branch
design and is directly tested by the version-branch verification test above. No transition period is
needed beyond the version-branch dispatch itself, since existing rows are automatically
`hash_version = 1` via the column's `DEFAULT 1` and new rows are written with the new value from the
moment the widened `chainHash` implementation lands.

## Rollout strategy

Single story, landed as its own reviewable unit, sequenced after W02-E01's exit gate is satisfied
(this story's own upstream dependency). The migration itself follows W02-E01's expand/backfill/
validate/contract rollout discipline rather than a single-step rollout, given its confirmed breaking-
change risk.

## Rollback strategy

If the widened `chainHash` implementation or version-branch verification proves incorrect after
landing (e.g. false-positive tamper detection on legitimate historical rows), the rollback path is
constrained by the fact that new rows will already have been written with the new `hash_version`
value and widened hash — a naive code revert alone would break verification for those new rows.
Rollback must therefore be handled per W02-E01's own protocol rollback/forward-fix discipline (the
manifest's own required "rollback/forward-fix plan" field), not a bespoke ad hoc revert. This is
recorded as an implementation-time requirement on the migration's own manifest entry, not invented
independently by this plan.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–10). Step 6 (the migration, shipped through
W02-E01's protocol) and step 7 (the widened `chainHash` + version-branch `Verify`) must land as one
atomic unit per D-04's decision text, not sequenced across two separate changes.

## Task breakdown

- **W04-E04-S001-T001** — Audit hash-chain widening, hash_version migration, and version-branched
  verification (steps 1–10 above).
- **W04-E04-S001-T002** — Independent review (per mandate §14, scoped to this story, mandatory given
  this is the single highest-risk task in the epic's scope).

## Expected artifacts

The widened `chainHash` implementation and metadata-canonicalization function; the `hash_version`
migration (via W02-E01's protocol); the version-branched `Verify` implementation; documentation of
the widened field list, canonicalization approach, and version-branch semantics.

## Expected evidence

Per-field tamper test output (every declared field independently); version-branch verification test
output (v1 historical-row branch; v2 new-row branch); confirmation the migration was classified and
shipped through W02-E01's protocol.

## Unresolved questions

- The exact `hash_version` value assigned to the new scheme (D-04 only confirms it is not `1`) — to
  be chosen and documented at implementation time.
- The exact canonicalization approach for `metadata` (canonical key-ordering scheme, numeric/string
  encoding stability guarantees) — to be designed at implementation time, following the constraint
  that it must never hash the stored jsonb directly.
- Whether `Anchor`/`CheckAnchor` require any version-awareness of their own beyond what `Verify`'s
  version-branch dispatch already provides — to be confirmed by re-reading `:253-311` at
  implementation time.
- The exact migration phase classification (whether this single additive-column-with-default change
  needs the full expand/backfill/validate/contract sequence, or a narrower subset of W02-E01's
  protocol phases) — to be determined against W02-E01-S001's manifest schema at implementation time.

## Approval conditions

This plan is approved for implementation once: (a) W02-E01's exit gate is satisfied (this story's
upstream dependency), (b) the unresolved questions above — most centrally, the exact `hash_version`
value and the metadata-canonicalization approach — are answered and documented, and (c) the owner and
reviewer are assigned.
