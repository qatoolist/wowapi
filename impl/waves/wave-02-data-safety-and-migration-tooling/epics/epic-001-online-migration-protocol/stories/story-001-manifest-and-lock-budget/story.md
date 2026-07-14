---
id: W02-E01-S001
type: story
title: Migration manifest schema and online-DDL lock budget
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
depends_on: []
blocks:
  - W02-E01-S002
acceptance_criteria:
  - AC-W02-E01-S001-01
  - AC-W02-E01-S001-02
  - AC-W02-E01-S001-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W02-E01-002
---

# W02-E01-S001 — Migration manifest schema and online-DDL lock budget

## Story ID

W02-E01-S001

## Title

Migration manifest schema and online-DDL lock budget

## Objective

Define and CI-enforce a migration manifest schema (online/maintenance classification, rows/bytes
estimate, lock/statement timeout, N/N-1 compatibility flag, backfill owner, validation query,
rollback/forward-fix plan) that every migration must satisfy, and implement a 2-second online-DDL
lock-timeout enforcement mechanism with abort-and-retry and a bounded retry ceiling.

## Value to the framework

Every later phase of the online-migration protocol this epic builds (expand, backfill, validate,
canary, switch, contract — W02-E01-S002/S003) depends on migrations first being classified and
budgeted: without a manifest, there is no machine-checkable statement of what a migration is
allowed to do, how long it is allowed to hold a lock, or who owns its backfill. This story is the
protocol's foundation — PLAN DATA-09's own task table orders T1 and T2 first for exactly this
reason. Today, per PLAN's evidence, "`Makefile`'s `migrate` target is a plain forward-apply;
`check_migrations.sh` checks only registration/markers/numbering, nothing about lock duration or
backfill" — this story converts an unclassified, unbudgeted migration process into one where a
migration's risk profile is declared, validated, and (for lock duration) mechanically enforced.

## Problem statement

`requirement-inventory.md` row DATA-09 states: "Adopt an online expand/backfill/validate/contract
protocol (T1–T9) | IMPL | P0 | planned | W02-E01-S001..S003 | Precedes DATA-01 T4/T5 + DATA-08 W6-T1;
T9 CI drills." PLAN's DATA-09 T1 row: "Migration manifest schema (online/maintenance, rows/bytes,
lock/statement timeout, N/N-1 flag, backfill owner, validation query, rollback/forward-fix plan) |
— | Every migration has a validated manifest entry; missing fields fail CI | Schema validation +
negative fixture | `DATA-09/manifest-schema/` | Get external review before locking the format |
Code + per-migration classification is human judgment every time." T2's row: "2-second online-DDL
lock-timeout enforcement with abort-and-retry | T1 | A statement exceeding budget aborts cleanly, no
partial DDL | Test against a concurrently-locked table | `DATA-09/lock-timeout/` | Bound total
retries — unbounded retry is a deploy-time DoS | Code, with human-set retry ceiling." No manifest
schema or lock-budget enforcement exists anywhere in the repository today.

## Source requirements

DATA-09 (T1, T2).

## Current-state assessment

Per PLAN's own evidence for DATA-09 as a whole (to be re-confirmed at this story's own execution
commit): "No expand/contract discipline, no online-DDL lock-timeout classification, no backfill-job
harness exists anywhere in wowapi today. `Makefile`'s `migrate` target is a plain forward-apply;
`check_migrations.sh` checks only registration/markers/numbering, nothing about lock duration or
backfill." This is a confirmed absence, not a partial implementation — there is no prior manifest
concept, prior lock-timeout wrapper, or prior classification scheme to extend or migrate away from.
This story's own re-confirmation step (per this programme's fail-first convention applied elsewhere,
e.g. W01-E01-S001) is to read `Makefile`'s `migrate` target and `check_migrations.sh` at this
story's actual start commit and confirm they still lack any manifest/lock-budget concept before
building one from zero.

## Desired state

Every migration in the repository has a manifest entry (format TBD by this story's own design work,
per "Unresolved questions" in `plan.md`) declaring: online-vs-maintenance classification, an
estimated row/byte count for the affected table(s), a lock timeout and statement timeout, an N/N-1
compatibility flag, a named backfill owner (if the migration requires one), a validation query, and
a rollback/forward-fix plan. CI fails a migration missing any required manifest field. A migration
whose DDL statement would exceed a 2-second lock-timeout budget aborts cleanly (no partial DDL
applied) and retries within a bounded ceiling, not indefinitely.

## Scope

- Designing and documenting the manifest schema's exact format (a new file per migration, an
  extension to existing migration files, or a separate registry — to be determined at
  implementation time per `plan.md`'s "Unresolved questions").
- CI validation of the manifest schema against every migration, failing on a missing required
  field, proven via a negative fixture test.
- The 2-second lock-timeout enforcement mechanism, applied to online-classified DDL statements,
  with abort-and-retry and an explicit, bounded retry ceiling.
- External review of the manifest schema's format before it is locked, per PLAN T1's own risk note.

## Out of scope

- Actually classifying and manifest-ing every existing historical migration in the repository —
  this story defines the schema and its enforcement mechanism; retrofitting manifest entries onto
  every pre-existing migration (if required by the CI gate's design) is scoped at implementation
  time and, if it is a large undertaking, may be split into a follow-up task rather than silently
  absorbed here (mandate §12).
- Expand-phase, backfill, validation, canary, switch, or contract tooling — W02-E01-S002/S003's
  scope. This story produces the manifest schema and lock-budget mechanism those later phases
  consume; it does not implement the phases themselves.
- Any specific migration's actual manifest content (e.g. DATA-01's own manifest entries) — that is
  each consuming story's own responsibility when it authors its migrations.

## Assumptions

- The manifest schema's exact storage format (inline in the migration file's header comment, a
  sibling YAML/JSON file, or a database-backed registry) is not yet determined by the source
  documents — PLAN's own risk note ("Get external review before locking the format") confirms this
  is a genuinely open design question this story must resolve, not one this plan can pre-answer.
  Recorded as an unresolved question in `plan.md`, not invented here.
- The 2-second lock-timeout figure is taken directly from PLAN's own acceptance criterion for T2
  ("2-second online-DDL lock-timeout enforcement") and MATRIX CS-18's cross-reference ("DATA-09 (T1
  –T5 precede the risky steps)") — this is a confirmed source figure, not an assumption this story
  invented.
- The retry ceiling's exact bound (number of retries, backoff schedule) is not specified in the
  source beyond "human-set retry ceiling" (PLAN T2's own classification column) — this story's plan
  records the exact figure as an implementation-time decision, per mandate §18.

## Dependencies

None within W02-E01 (this is the epic's first story). Depends on W00's exit gate at wave scope.
Blocks W02-E01-S002 (expand-phase tooling classifies its own migrations against this story's
manifest schema; the backfill harness's bounded-batch/tx-time controls are informed by this story's
lock-budget mechanism).

## Affected packages or components

New: a migration-manifest validation tool/library (exact package location to be determined —
expected under a new or existing `internal/tools/` or `kernel/migrate`-adjacent location, per
`plan.md`'s "Unresolved questions"). Extended: `Makefile`'s `migrate` target and
`check_migrations.sh` (or their replacement/successor, if the manifest schema's design supersedes
the current script rather than extending it — to be determined at implementation time).

## Compatibility considerations

This story adds a new validation gate (the manifest-schema CI check) that, once enforced, will
reject any future migration lacking a manifest entry. Per this programme's compat-flag-first pattern
used elsewhere (e.g. W01-E03-S002's FBL-08 boot-time enforcement), this story's plan should consider
whether the CI gate is enforced immediately or behind a transition period — to be resolved in
`plan.md` given no source guidance exists on this specific point for DATA-09 T1.

## Security considerations

The lock-timeout enforcement mechanism's abort-and-retry logic must not create a deploy-time denial-
of-service via unbounded retry — PLAN T2's own risk note states this explicitly: "Bound total
retries — unbounded retry is a deploy-time DoS." This is a required security control for this
story's T2 scope, not an optional hardening add-on.

## Performance considerations

The lock-timeout budget (2 seconds) directly bounds how long an online-classified migration may
hold a lock before aborting — this is itself the performance control DATA-09 exists to provide
(protecting concurrent traffic from a long-held DDL lock), not a separate performance concern this
story must additionally address.

## Observability considerations

A lock-timeout abort-and-retry event should be observable (logged, at minimum) so an operator can
distinguish "the migration succeeded on retry N" from "the migration is silently retrying
indefinitely" — this is a reasonable implementation-time addition given T2's DoS risk note, though
not separately mandated by the source beyond the retry-ceiling requirement itself.

## Migration considerations

This story is itself migration-tooling — it has no data or schema migration of its own to perform
(it does not touch any application table). Its own "migration" is process/tooling change only.

## Documentation requirements

Document the manifest schema's format, required fields, and validation rules; document the
lock-timeout budget, abort-and-retry behavior, and retry ceiling, so that a future migration author
knows how to author a compliant migration without re-reading this story's own planning documents.

## Acceptance criteria

- **AC-W02-E01-S001-01**: The migration manifest schema is defined and documented; a migration with
  a complete manifest entry validates successfully; a migration missing any required field
  (online/maintenance classification, rows/bytes, lock/statement timeout, N/N-1 flag, backfill
  owner, validation query, rollback/forward-fix plan) fails CI via a negative fixture test.
- **AC-W02-E01-S001-02**: The manifest schema format has been externally reviewed (per PLAN T1's own
  risk note) before being locked as the schema every subsequent DATA-09 story and every future
  migration must satisfy.
- **AC-W02-E01-S001-03**: A DDL statement exceeding the 2-second lock-timeout budget against a
  concurrently-locked table aborts cleanly with no partial DDL applied, and retries within an
  explicit, bounded ceiling — proven by a test against a deliberately concurrently-locked table.

## Required artifacts

- The manifest schema definition (format TBD, see "Unresolved questions" in `plan.md`).
- The manifest-schema CI validation tool.
- The lock-timeout enforcement mechanism (code).
- Manifest-schema and lock-budget documentation.
See `artifacts/index.md`.

## Required evidence

- Schema validation test output (positive and negative fixture).
- Concurrently-locked-table lock-timeout abort/retry test output.
- External review record for the manifest schema format.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none within this
epic) recorded, owner/reviewer assignment pending, unresolved questions (manifest storage format,
retry ceiling exact bound) explicitly recorded rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the manifest schema received external review before
being locked (AC-W02-E01-S001-02).

## Risks

RISK-W02-E01-002 (the manifest schema, once locked, becomes a contract every subsequent migration
must satisfy — an under-specified schema is costly to retrofit) — see epic-level `risks.md` for full
detail and mitigation/contingency.

## Residual-risk expectations

Once the external-review step (AC-W02-E01-S001-02) is executed as planned, residual risk is expected
to be low — this is a foundational but well-bounded design-and-enforcement story with a clear,
source-derived acceptance bar.

## Plan

See `plan.md`.
