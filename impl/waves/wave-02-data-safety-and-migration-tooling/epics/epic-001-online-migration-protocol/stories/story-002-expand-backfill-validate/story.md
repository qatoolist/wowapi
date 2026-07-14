---
id: W02-E01-S002
type: story
title: Expand-phase tooling, resumable backfill harness, and validation-phase tooling
status: accepted
wave: W02
epic: W02-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - DATA-09
depends_on:
  - W02-E01-S001
blocks:
  - W02-E01-S003
  - W02-E02-S002
acceptance_criteria:
  - AC-W02-E01-S002-01
  - AC-W02-E01-S002-02
  - AC-W02-E01-S002-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W02-001
---

# W02-E01-S002 — Expand-phase tooling, resumable backfill harness, and validation-phase tooling

## Story ID

W02-E01-S002

## Title

Expand-phase tooling, resumable backfill harness, and validation-phase tooling

## Objective

Build expand-phase tooling (nullable/default-safe columns, new tables/indexes/compatibility views,
`NOT VALID` constraints, non-transactional `CREATE INDEX CONCURRENTLY`), a resumable/tenant-scoped/
keyset-paginated/checkpointed backfill-job harness with bounded batch/tx time and rate controls, and
validation-phase tooling (`VALIDATE CONSTRAINT` orchestration plus reconciliation queries with
machine-checked artifact capture) — the "make the schema change and move the data safely" phase of
the online-migration protocol this epic builds from zero.

## Value to the framework

This story is the protocol's actual data-movement machinery: S001 (already planned) gives every
migration a manifest and a lock budget, but nothing yet lets a migration actually expand the schema
without blocking traffic, backfill existing rows safely, or prove the backfill completed correctly.
Without this story, the manifest and lock-budget mechanism from S001 has nothing to enforce beyond
metadata bookkeeping. This story is also this epic's highest-risk story: PLAN's own risk column for
T4 states plainly, "Largest risk surface in DATA-09" — an incorrectly-resumable backfill harness
that silently reprocesses or skips rows is a data-correctness defect, not merely a missed feature.

## Problem statement

PLAN DATA-09's task table gives three rows for this story's scope. T3: "Expand-phase tooling:
nullable/default-safe columns, new tables/indexes/compatibility views, `NOT VALID` constraints,
non-transactional `CREATE INDEX CONCURRENTLY` | T1 | Expand migrations don't block traffic; old and
new readers both accept | Old-reader-compatibility test | `DATA-09/expand-phase/` | Confirm current
tooling supports issuing statements outside the wrapping transaction." T4: "Backfill job harness:
resumable, tenant-scoped, keyset-paginated, checkpointed, bounded batch/tx time, rate controls —
reuses DATA-02's lease primitive for checkpoint safety | T3; DATA-02 T1 | **Interrupted/resumed
backfill test (explicitly required)** — no reprocessing or skipping | This is the test |
`DATA-09/backfill-interrupt-resume/` | Largest risk surface in DATA-09 | Code for the harness;
human decision on batch size/rate/window per migration." T5: "Validation-phase tooling:
`VALIDATE CONSTRAINT` + reconciliation queries, artifact capture | T4 | Zero-mismatch reports are
machine-checked artifacts, not prose | Artifact-schema test | `DATA-09/validation-artifacts/` |
Code for the harness; human review of the report before canary." None of this tooling exists today
(per DATA-09's own "reality check": "No expand/contract discipline, no online-DDL lock-timeout
classification, no backfill-job harness exists anywhere in wowapi today").

## Source requirements

DATA-09 (T3, T4, T5).

## Current-state assessment

Per PLAN's evidence (to be re-confirmed at this story's own execution commit, following this
programme's fail-first convention): no expand-phase tooling, backfill-job harness, or validation-
phase tooling exists anywhere in the repository today. T4's own "Depends-on" column names two
prerequisites: "T3; DATA-02 T1." T3 is this story's own first task (sequenced internally before T4).
DATA-02 T1 is the confirmed problem this story must resolve explicitly, not silently route around:
DATA-02 ("Add lease generations/fencing and effect idempotency to jobs") is W04 scope
(`impl/analysis/wave-allocation-detail.md`'s W04-E01-S001 row), meaning its "shared lease/fencing
primitive... as a reusable kernel building block for DATA-02/03/04" does not exist at the time this
story executes. `impl/analysis/wave-allocation-detail.md`'s own W02-E01 entry states the resolution
explicitly: "T4 backfill harness reuses DATA-02 T1's lease primitive: forward-dependency, so S002
builds a minimal checkpoint lease and W04-E01 replaces it — record as planned deviation-risk." This
is confirmed programme-level guidance, not an assumption this story invented — it is the authoritative
allocation this story must follow.

## Desired state

Expand-phase tooling issues `CREATE INDEX CONCURRENTLY` and `NOT VALID` constraints without blocking
traffic, proven by an old-reader-compatibility test (old-version application code continues to
function against the expanded schema). A backfill-job harness resumes correctly after an interruption
— no row reprocessed, no row skipped — using a minimal, purpose-bounded checkpoint-lease mechanism
built by this story as an explicit interim measure (not DATA-02's full shared primitive).
Validation-phase tooling orchestrates `VALIDATE CONSTRAINT` and reconciliation queries, producing a
machine-checked, artifact-schema-conformant zero-mismatch report.

## Scope

- Expand-phase tooling: nullable/default-safe column support, new table/index/compatibility-view
  creation, `NOT VALID` constraint addition, non-transactional `CREATE INDEX CONCURRENTLY` issuance
  (PLAN T3).
- The backfill-job harness: resumable, tenant-scoped, keyset-paginated, checkpointed, with bounded
  batch/transaction time and rate controls (PLAN T4).
- **The interim checkpoint-lease mechanism**: a minimal, purpose-bounded lease primitive providing
  exactly the checkpoint-token and resumability semantics the backfill harness needs — explicitly
  not DATA-02 T1's full shared lease/fencing primitive (which additionally provides job-claim
  fencing generations and heartbeats that this story's backfill harness does not require). This is
  a deliberate, planned, scope-bounded substitute, recorded as RISK-W02-001, pending
  W04-E01-S001's replacement of it with the full shared primitive once DATA-02 lands.
- Validation-phase tooling: `VALIDATE CONSTRAINT` orchestration, reconciliation queries, and
  machine-checked artifact capture (PLAN T5).

## Out of scope

- **DATA-02 T1's full shared lease/fencing primitive** (generations, heartbeats, job-claim fencing
  reused across DATA-02/03/04) — W04-E01-S001's scope. This story does not attempt to build the
  general-purpose primitive; it builds only the minimal checkpoint-lease subset its own backfill
  harness needs, explicitly scoped to avoid overlapping with W04's eventual build (see "Assumptions"
  and RISK-W02-001).
- **The manifest schema and lock-timeout enforcement mechanism** — W02-E01-S001's scope (already
  planned; this story depends on it, per `depends_on`).
- **Canary, switch, and contract-phase tooling** — W02-E01-S003's scope.
- **Any specific migration's actual expand/backfill/validate content** (e.g. DATA-01's own composite-
  FK migration) — that is each consuming story's (e.g. W02-E02) own responsibility when it authors
  its migrations using this story's tooling.
- **The human decision on batch size/rate/window per migration** — PLAN T4's own classification
  column states this is a per-migration human judgment call; this story builds the harness's
  configurable controls, it does not itself set a specific batch/rate/window value for any real
  migration.

## Assumptions

- The interim checkpoint-lease's scope-bounding (checkpoint token + resumability only, no fencing
  generations, no heartbeats) is assumed sufficient for the backfill harness's own interrupted/
  resumed test to pass cleanly, without needing DATA-02's additional job-claim-fencing semantics —
  this is a design assumption this story's `plan.md` must validate at implementation time, not a
  confirmed fact from the source (the source only confirms that DATA-02 T1 does not yet exist, not
  that a minimal subset suffices; if implementation reveals the minimal subset is insufficient, that
  is a deviation to record, not a silent scope expansion).
- T3's own risk note ("Confirm current tooling supports issuing statements outside the wrapping
  transaction") is treated as an implementation-time confirmation step, not a pre-confirmed fact —
  this story's plan records it as an assumption to verify, consistent with mandate §18.
- The backfill harness's tenant-scoping assumes the existing tenant-context mechanism
  (`plat.WithTenant`-shaped, per other DATA findings' evidence in the same PLAN §5.3 section) is
  reusable for a backfill job's own row-scoped iteration — to be confirmed at implementation time
  against the actual tenant-context API surface.

## Dependencies

Depends on W02-E01-S001 (this story's expand-phase tooling classifies its own migrations against
S001's manifest schema; T3 depends on T1 per PLAN's own "Depends-on" column). Blocks
W02-E01-S003 (T6 depends on T5's validation-phase tooling) and W02-E02-S002 (DATA-01 T4/T5, gated on
this epic's S001+S002 acceptance per `impl/analysis/wave-allocation-detail.md`'s cross-wave
sequencing note).

## Affected packages or components

New: expand-phase tooling (exact package location TBD, expected adjacent to the migration-execution
code S001 establishes); the backfill-job harness and its interim checkpoint-lease mechanism (new
package, location TBD); validation-phase tooling (new package, location TBD). No existing kernel
package is modified by this story beyond wiring these new tools into the migration-execution flow
S001 establishes.

## Compatibility considerations

Expand-phase tooling's entire purpose is compatibility: old and new application readers must both
accept the expanded schema during the migration window, proven by the old-reader-compatibility test
(PLAN T3's own acceptance criterion). This is the compatibility contract this story exists to
deliver, not a separate concern layered on top.

## Security considerations

None beyond what S001's lock-timeout mechanism already provides for the DDL statements this story's
expand-phase tooling issues. The backfill harness's rate controls are an operational-safety
mechanism (protecting production load), not a security control in the access-control sense.

## Performance considerations

The backfill harness's bounded batch/tx time and rate controls are themselves the performance
safety mechanism — they exist specifically to prevent a large backfill from monopolizing database
resources or holding a transaction open longer than acceptable. The exact batch size/rate/window
values are a per-migration human decision (PLAN T4's own classification column), not a fixed value
this story's harness hardcodes.

## Observability considerations

The backfill harness's checkpoint state should be observable (at minimum, queryable) so an operator
can confirm a backfill's progress and confirm a resume picked up correctly after an interruption —
this is a reasonable implementation-time addition supporting the interrupted/resumed test's own
verification, not separately mandated beyond that test's own requirements.

## Migration considerations

This story is itself migration-tooling. Its "migration considerations" are the subject of the story:
expand-phase tooling changes how future migrations may be authored (compatibly, without blocking
traffic); the backfill harness changes how a data-migration's row-by-row work is safely executed;
validation-phase tooling changes how a migration proves its own correctness before proceeding.

## Documentation requirements

Document the expand-phase tooling's usage (which schema changes it supports, e.g. nullable/default-
safe columns, `NOT VALID` constraints, concurrent index creation); document the backfill harness's
configuration surface (batch size, rate, window) and its interim-checkpoint-lease scope-boundary
(explicitly flagged as a temporary substitute pending W04-E01-S001); document the validation-phase
tooling's artifact schema.

## Acceptance criteria

- **AC-W02-E01-S002-01**: Expand-phase tooling issues `CREATE INDEX CONCURRENTLY` and `NOT VALID`
  constraints without blocking traffic; an old-reader-compatibility test confirms both old and new
  application readers accept the expanded schema during the migration window.
- **AC-W02-E01-S002-02**: The backfill-job harness's named interrupted/resumed test passes: no row
  is reprocessed, no row is skipped, when a backfill is interrupted and resumed. The interim
  checkpoint-lease mechanism's scope is documented as bounded to checkpoint-token/resumability only,
  explicitly distinguished from DATA-02 T1's full shared primitive, with a recorded forward
  reference to W04-E01-S001 as its planned replacement.
- **AC-W02-E01-S002-03**: Validation-phase tooling's zero-mismatch report is a machine-checked,
  artifact-schema-conformant record (proven by an artifact-schema test), not free-form prose.

## Required artifacts

- Expand-phase tooling (code).
- The backfill-job harness and its interim checkpoint-lease mechanism (code).
- Validation-phase tooling and its artifact schema (code).
- Documentation for all three, including the interim-lease scope-boundary note.
See `artifacts/index.md`.

## Required evidence

- Old-reader-compatibility test output.
- Interrupted/resumed backfill test output (the named, explicitly-required test).
- Artifact-schema test output for the validation-phase zero-mismatch report.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W02-E01-S001
recorded, the interim-checkpoint-lease scope-bounding decision recorded as an assumption to validate
(not a silently-assumed fact), owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the interim-checkpoint-lease deviation is recorded
honestly (bounded scope, explicit forward reference to W04-E01-S001), not silently presented as a
complete DATA-02 T1 substitute.

## Risks

RISK-W02-001 (the interim checkpoint-lease is a genuine, planned technical-debt-bearing deviation,
superseded by W04-E01-S001) — see wave-level `risks.md` and epic-level `risks.md` for full detail
and mitigation/contingency.

## Residual-risk expectations

RISK-W02-001's residual risk remains medium until W04-E01-S001 actually lands and migrates this
story's interim lease onto the full shared primitive — this story's own closure does not eliminate
the risk, it merely bounds and records it correctly. Beyond that, once the interrupted/resumed
backfill test (AC-W02-E01-S002-02) and the old-reader-compatibility test (AC-W02-E01-S002-01) pass,
residual risk for this story's own scope is expected to be low.

## Plan

See `plan.md`.
