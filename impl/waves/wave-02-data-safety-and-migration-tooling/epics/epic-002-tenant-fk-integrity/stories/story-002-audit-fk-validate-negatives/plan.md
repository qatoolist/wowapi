---
id: PLAN-W02-E02-S002
type: plan
parent_story: W02-E02-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W02-E02-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent a remediation procedure for a mismatch-audit finding
that has not yet occurred, per mandate §18 — where the source is genuinely silent (T3's outcome,
T8's exact consumer-verification scope), this plan states what must be determined during execution
rather than inventing specifics.

## Proposed architecture

A mismatch-audit tool that connects as the platform role (bypassing RLS by design, per T3's own risk
note) and scans all 8 tenant-scoped FK edges for `child.tenant_id != parent.tenant_id`, producing a
dated, inspectable report. A set of 8 composite-FK migrations, each performed in two sequenced
steps per edge (`ADD CONSTRAINT ... NOT VALID`, then a separate `VALIDATE CONSTRAINT`), run
per-table as independent statements so each statement's lock duration is independently measurable
against the DATA-09 budget this wave's W02-E01 provides. A catalog-driven RLS matrix test extended
to cover cross-tenant FK-violating inserts under both `app_rt` and `app_platform`. Optionally, a
cleanup migration removing the 8 now-redundant single-column FKs, gated on a verified absence of any
remaining consumer/rollback dependency on the old FK's cascade behavior.

## Implementation strategy

1. Re-confirm, at this story's actual start commit: the current absence of any composite FK on the 8
   edges; that W02-E02-S001's catalog scanner is producing the expected 8-edge inventory; and the
   acceptance status of W02-E01-S001 and W02-E01-S002 (required before T002/T003 may begin — see
   "Implementation sequence" below for the hard checkpoint).
2. Build the mismatch-audit tool (T001/PLAN T3): a platform-role-connected scan across all 8 edges,
   producing a dated report. Write the integration test that seeds a deliberate cross-tenant
   mismatch and confirms the audit detects it (this test's fixture is independent of, and does not
   substitute for, the real-data audit run).
3. Run the real-data mismatch audit against staging/prod-shaped data. **Branch on outcome:**
   - Zero-mismatch: record the report, proceed.
   - Mismatch found: halt, escalate to the acceptance authority per RISK-W02-002's documented path,
     record the finding and its eventual resolution in `deviations.md`, and do not proceed to step
     5 (`VALIDATE CONSTRAINT`) until a second zero-mismatch audit passes. This branch's exact
     remediation procedure is not pre-specified here — it is genuinely undecidable in advance per
     mandate §18.
4. **Checkpoint: confirm W02-E01-S001 and W02-E01-S002 have both reached `accepted`.** Do not begin
   T002 (composite FK `NOT VALID` add, PLAN T4) or T003 (`VALIDATE CONSTRAINT`, PLAN T5) before this
   checkpoint passes, regardless of whether steps 1–3 above are otherwise complete. This checkpoint
   is restated in "Implementation sequence" below as its own numbered gate, not merely folded into
   the task list, because it is a hard blocking condition rather than an ordinary task dependency.
5. Add composite FK `NOT VALID` for all 8 edges (T002/PLAN T4), one statement per table, and measure
   each statement's lock duration against the DATA-09 2-second budget (migration lock-duration
   test).
6. `VALIDATE CONSTRAINT` each of the 8 new composite FKs (T003/PLAN T5), scheduled per
   W02-E01-S002's own backfill/validate-phase tooling, under a load test confirming no concurrent-DML
   blocking; produce the second zero-mismatch confirmation as part of this step.
7. Extend the catalog-driven RLS matrix test with seeded cross-tenant insert negative-test cases
   under both `app_rt` and `app_platform` (T004/PLAN T7); explicitly assert on the platform-role
   result rather than assuming it, per T7's own risk note.
8. If pursued: verify no code or migration relies on the old single-column FK's cascade behavior
   (grep sweep + full regression run), then remove the 8 redundant single-column FKs (T005/PLAN
   T8). This step is optional and its non-completion does not block story or epic closure.
9. Independent review (T006), scoped specifically to confirming the W02-E01 gate was genuinely
   honored and the mismatch-audit outcome was honestly recorded.

## Expected package or module changes

The mismatch-audit tool (exact package location TBD, expected adjacent to W02-E02-S001's catalog
scanner tool). The 8 composite-FK migrations (2 statements each: `NOT VALID` add, `VALIDATE
CONSTRAINT`), located under the existing migration directory structure. Extensions to the existing
catalog-driven RLS matrix test suite for T004's negative-test cases. If T005 proceeds: 8 further
migrations removing the redundant single-column FKs.

## Expected file changes where determinable

- A new mismatch-audit tool and its report format (exact file path TBD).
- 8 new migration files (or migration statement groups) adding `NOT VALID` composite FKs, one per
  edge, each with its own `VALIDATE CONSTRAINT` follow-up statement.
- Extensions to the existing catalog-driven RLS matrix test file(s) for the T004 cross-tenant
  negative-test cases under both roles.
- If T005 proceeds: 8 further migration files removing the old single-column FKs.

## Contracts and interfaces

The mismatch-audit report format (fields: edge identifier, row count scanned, mismatch count,
timestamp, connection role used) — exact schema to be determined at implementation time, informed
by whatever reporting convention W02-E02-S001's catalog scanner already establishes, for
consistency. No new runtime API — this story's outputs are migration/tooling artifacts and test
coverage, not application-facing interfaces.

## Data structures

The mismatch-audit report's own record structure, per "Contracts and interfaces" above. No
application data model change beyond the composite-FK constraint definitions themselves (a schema
change, not a data-structure change in the application-code sense).

## APIs

None affected — this story is schema/tooling-internal, not a runtime API change.

## Configuration changes

None anticipated beyond whatever configuration the mismatch-audit tool needs to obtain a
platform-role connection (expected to reuse an existing platform-role connection mechanism already
present in the framework for other RLS-bypassing administrative operations, rather than introducing
a new one — to be confirmed at implementation time).

## Persistence changes

The 8 composite-FK constraints themselves (`NOT VALID` add, then `VALIDATE CONSTRAINT`) are this
story's core persistence change. No table, column, or index is added or removed by T002–T004 (T001's
mismatch audit is read-only). T005, if pursued, removes the 8 redundant single-column FK
constraints — a persistence change, but strictly a removal of now-redundant metadata, not a data
change.

## Migration strategy

Per-table separate statements for both the `NOT VALID` add and the `VALIDATE CONSTRAINT` step (T4's
own risk note: "Run per-table as separate statements"), so that a single edge's migration failure or
excessive lock duration does not block or obscure the others. `VALIDATE CONSTRAINT` is scheduled
per W02-E01-S002's backfill/validate-phase tooling (T5's own risk note) rather than run as an
ordinary inline migration step, because it is I/O-bound and must not block concurrent DML.

## Concurrency implications

`VALIDATE CONSTRAINT` (T003/PLAN T5) is the step with the most direct concurrency implication: PLAN's
own acceptance criterion requires it not block concurrent DML, verified by "Load test under
concurrent writer load." The mismatch audit (T001/PLAN T3) must itself be safe to run against a live,
concurrently-written database — it is a read-only scan, so the primary concurrency concern is
producing a consistent-enough snapshot for the report to be meaningful, not correctness of the scan
itself under concurrent writes (a row written mid-scan is expected to be caught by the *second*
zero-mismatch confirmation performed alongside T5, not necessarily the first).

## Error-handling strategy

The mismatch audit must fail deployment (not merely warn) on any detected mismatch, per T3's own
acceptance criterion — this is a hard gate, not an advisory report. The composite FK `NOT VALID` add
(T002) must fail cleanly per-table if a statement would exceed the DATA-09 lock-timeout budget,
consuming W02-E01-S001's own lock-timeout enforcement mechanism rather than reimplementing it. If
`VALIDATE CONSTRAINT` (T003) itself surfaces a mismatch that the T001 audit missed (e.g. a row
written between the audit and the validate step), that failure must be treated with the same
severity as a T001-detected mismatch — escalated per RISK-W02-002's path, not silently retried.

## Security controls

T004's cross-tenant negative test under `app_platform` is itself a required security control, not
merely a test: PLAN's own risk note ("Confirm platform role doesn't bypass FK constraints — don't
assume") means this story cannot claim AC-W02-E02-S002-04 satisfied without an actual assertion on
the platform-role result. If the platform role is found to bypass the new FK constraints, that
outcome is itself a new finding requiring escalation, not a result this plan may silently discard or
treat as expected.

## Observability changes

The mismatch-audit report (T001) must be dated and inspectable, per "Observability considerations"
in `story.md`. The `VALIDATE CONSTRAINT` step (T003), being I/O-bound, should emit progress/duration
visibility during its run — exact mechanism informed by whatever observability convention
W02-E01-S002's backfill/validate tooling already establishes, to avoid this story inventing a
parallel one.

## Testing strategy

- T001: integration test seeding a deliberate cross-tenant mismatch via a platform-role connection,
  confirming the audit tool detects it (fixture-based fail-first test, independent of the real-data
  audit run against staging/prod-shaped data).
- T002: migration lock-duration test confirming each per-table `NOT VALID` add stays under the
  DATA-09 2-second lock-timeout budget.
- T003: load test under concurrent writer load confirming `VALIDATE CONSTRAINT` does not block
  concurrent DML, plus the second zero-mismatch confirmation.
- T004: new catalog-driven RLS matrix test asserting a seeded cross-tenant insert fails under both
  `app_rt` and `app_platform`, with an explicit assertion on the platform-role outcome (not an
  assumption).
- T005 (if pursued): full regression suite run plus a grep sweep for any reference to the old
  single-column FK constraint name.

## Regression strategy

The extended catalog-driven RLS matrix test (T004) becomes a permanent regression guard against a
future change silently reopening the cross-tenant insert gap. If T005 proceeds, the full regression
run required by its own acceptance criterion is itself the regression-safety check for the FK
removal.

## Compatibility strategy

T002's `NOT VALID` add is designed to be non-disruptive to existing writes (metadata-only at add
time, per "Compatibility considerations" in `story.md`). T005's FK removal is the one compatibility-
sensitive change in this story — it must not proceed until verified that no code or migration
depends on the old FK's cascade behavior, which is precisely why T005 is scoped as optional and
sequenced strictly last.

## Rollout strategy

Sequenced per "Implementation sequence" below — T001 (audit) and the W02-E01 acceptance checkpoint
must both clear before T002/T003 begin; T004 follows T003; T005 is optional and, if pursued, follows
T004. No phased/canary rollout beyond what W02-E01-S002's own protocol provides for T002/T003, which
this story consumes rather than re-implements.

## Rollback strategy

If `VALIDATE CONSTRAINT` (T003) surfaces an unexpected mismatch or blocks concurrent DML despite the
load test passing pre-production, the composite FK can be dropped per-edge without affecting the
other 7 edges (per the per-table-separate-statements migration strategy) while the issue is
investigated — the `NOT VALID` add itself is not destructive to existing data, so a rollback here
is a schema-metadata change, not a data-recovery operation. T005's removal of the old single-column
FKs is the one step in this story that is harder to reverse cleanly (recreating a dropped FK
constraint requires re-validating it); this is exactly why T005 requires full consumer/rollback-path
verification before proceeding, and why it remains optional.

## Implementation sequence

Steps 1–9 under "Implementation strategy" above, with the following as a **hard checkpoint, not
merely a dependency-list entry**: T002 (composite FK `NOT VALID` add) and T003 (`VALIDATE
CONSTRAINT`) must not begin until W02-E01-S001 and W02-E01-S002 have both reached `accepted`. This
checkpoint sits between step 3 (the mismatch audit resolving to zero-mismatch or a resolved
remediation decision) and step 5 (the composite FK add) in the sequence above — both conditions
(audit clean, W02-E01 accepted) must hold before T002 starts; neither alone is sufficient. This
checkpoint is deliberately restated here, in this story's own `depends_on` front matter, in
`dependencies.md` (epic and wave scope), and in T002/T003's own task files' "Dependencies" sections
— per this story's own design goal that a task-level reader picking up "the next todo task" cannot
miss it by reading only the task file.

## Task breakdown

- **W02-E02-S002-T001** — Mismatch audit (PLAN T3).
- **W02-E02-S002-T002** — Composite FK `NOT VALID` add for all 8 edges (PLAN T4) — **gated: cannot
  start before W02-E01-S001 and W02-E01-S002 reach `accepted`.**
- **W02-E02-S002-T003** — `VALIDATE CONSTRAINT` for each new composite FK (PLAN T5) — **gated: same
  condition as T002, plus depends on T002 itself.**
- **W02-E02-S002-T004** — Seeded cross-tenant insert negative tests under both roles (PLAN T7).
- **W02-E02-S002-T005** — Optional, best-effort redundant single-column FK cleanup (PLAN T8).
- **W02-E02-S002-T006** — Independent review (mandate §14, P0 priority), specifically scoped to
  confirming the W02-E01 gate was genuinely honored on T002/T003 and the mismatch-audit outcome was
  honestly recorded.

## Expected artifacts

The mismatch-audit tool and its report; the 8 composite-FK `NOT VALID`/`VALIDATE CONSTRAINT`
migrations; the extended cross-tenant negative-test suite; if T005 proceeds, the FK-removal
migrations and their consumer/rollback verification record.

## Expected evidence

The zero-mismatch report (or resolved remediation-decision record); the seeded-mismatch integration
test output; the migration lock-duration test output for T002; the concurrent-writer-load test
output plus second zero-mismatch confirmation for T003; the cross-tenant negative-test output for
both roles; if T005 proceeds, the regression + grep sweep output.

## Unresolved questions

- The mismatch-audit report's exact schema/format — to be determined at implementation time,
  informed by W02-E02-S001's own catalog-scanner reporting convention for consistency.
- The exact platform-role connection mechanism the audit tool reuses (existing framework mechanism
  vs. a new one) — to be confirmed at implementation time.
- **The mismatch-audit's actual outcome, and — if a mismatch is found — its remediation procedure.**
  Genuinely undecidable in advance per mandate §18; this plan records the escalation path (halt,
  escalate to acceptance authority, record in `deviations.md`, re-audit before proceeding) but does
  not invent the remediation itself.
- T8's exact consumer/rollback-path verification scope beyond "grep + full regression" (PLAN's own
  Tests column) — to be determined at implementation time if T005 is pursued.
- Whether the mismatch-audit tool and the cross-tenant negative-test suite (T004) share test
  infrastructure with W02-E02-S001's fixture-schema test — an efficiency question to be resolved at
  implementation time, not a scope question.

## Approval conditions

This plan is approved for implementation once: (a) the owner and reviewer are assigned; (b) T001's
audit tool and integration test are ready to execute against staging/prod-shaped data; and (c) —
specifically for T002/T003 — W02-E01-S001 and W02-E01-S002 have both reached `accepted`. Condition
(c) is a precondition for T002/T003 specifically, not for this plan's overall approval or for T001/
T004/T005's work, which may proceed independently of W02-E01's acceptance status.
