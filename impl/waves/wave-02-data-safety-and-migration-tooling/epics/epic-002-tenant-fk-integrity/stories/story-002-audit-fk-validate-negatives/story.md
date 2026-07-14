---
id: W02-E02-S002
type: story
title: Cross-tenant mismatch audit, composite FK validation, and negative tests
status: accepted
wave: W02
epic: W02-E02
owner: W02FKVerAgg
reviewer: W02ReviewGate
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-01
depends_on:
  - W02-E02-S001
  - W02-E01-S001
  - W02-E01-S002
blocks: []
acceptance_criteria:
  - AC-W02-E02-S002-01
  - AC-W02-E02-S002-02
  - AC-W02-E02-S002-03
  - AC-W02-E02-S002-04
  - AC-W02-E02-S002-05
artifacts:
  - ART-W02-E02-S002-001
  - ART-W02-E02-S002-002
  - ART-W02-E02-S002-003
  - ART-W02-E02-S002-004
  - ART-W02-E02-S002-005
evidence:
  - EV-W02-E02-S002-001
  - EV-W02-E02-S002-002
  - EV-W02-E02-S002-003
decisions: []
risks:
  - RISK-W02-002
  - RISK-W02-E02-002
---

# W02-E02-S002 — Cross-tenant mismatch audit, composite FK validation, and negative tests

## Story ID

W02-E02-S002

## Title

Cross-tenant mismatch audit, composite FK validation, and negative tests

## Objective

Prove `child.tenant_id = parent.tenant_id` for every existing row across all 8 tenant-scoped FK
edges (failing deployment on any mismatch); add composite `NOT VALID` foreign keys for all 8 edges
and `VALIDATE CONSTRAINT` each one — both gated on this wave's W02-E01 online-migration protocol
reaching `accepted`; prove a seeded cross-tenant insert fails under both `app_rt` and
`app_platform` roles; and, as an optional, non-blocking best-effort task, remove the redundant
single-column FKs once every consumer and rollback path has been verified.

## Value to the framework

This story closes the epic's higher-risk half: MATRIX CS-18's own fail-first framing is the direct
test this story exists to flip — "platform-role seeded cross-tenant parent/child insert succeeds
today; fails after." Where W02-E02-S001 built the purely-additive foundation (parent unique
indexes, the catalog scanner, the CI gate), this story performs the actual data-integrity proof
(the mismatch audit) and the actual schema enforcement (the composite FK add and validation) that
converts "RLS proves the child row's own tenant; nothing proves parent and child agree" (PLAN's own
DATA-01 evidence) into a database-enforced guarantee. This is the framework's highest-severity
confirmed structural gap in tenant isolation, graded P0 in every source that discusses it — this
story is where that gap is actually closed, not merely scaffolded for closure.

## Problem statement

PLAN's own DATA-01 task table gives this story's five tasks directly, quoted exactly:

- "T3. Mismatch audit: prove `child.tenant_id = parent.tenant_id` for every existing row; fail
  deployment on any mismatch | T2 | Zero-mismatch report against staging/prod-shaped data |
  Integration seeding a deliberate cross-tenant mismatch via platform role | `DATA-01/mismatch-
  audit/` | Requires a platform-role connection to bypass RLS for the scan."
- "T4. Add composite FK `NOT VALID` for all 8 edges | T1, T3 | Metadata-only add stays under the
  DATA-09 2-second lock-timeout budget | Migration lock-duration test | `DATA-01/composite-fk-
  notvalid/` | Run per-table as separate statements."
- "T5. `VALIDATE CONSTRAINT` each new composite FK | T4 | Validation doesn't block concurrent DML;
  second zero-mismatch confirmation | Load test under concurrent writer load | `DATA-01/validate-
  constraint/` | I/O-bound — schedule per DATA-09's backfill/validate phases."
- "T7. Seeded cross-tenant insert negative tests under both `app_rt` and `app_platform` | T5 |
  Insert violating tenant equality fails under both roles | New catalog-driven RLS matrix test |
  `DATA-01/cross-tenant-fk-negative/` | Confirm platform role doesn't bypass FK constraints — don't
  assume."
- "T8. Remove redundant single-column FKs, only after all consumers/rollback paths verified | T5,
  T7 | No code relies on the old FK name for cascade behavior | Full regression + grep |
  `DATA-01/fk-cleanup/` | Optional — don't block P0 closure on it."

T4's own acceptance criterion — "Metadata-only add stays under the DATA-09 2-second lock-timeout
budget" — is the direct textual evidence that T4 depends on W02-E01-S001's lock-budget mechanism
existing and having been accepted, not merely a general sequencing preference. T5's own risk note
— "schedule per DATA-09's backfill/validate phases" — is the direct textual evidence that T5
depends on W02-E01-S002's backfill/validate-phase tooling specifically. PLAN's own PF-DATA
cross-cutting note (6), quoted exactly: "DATA-09 is new infrastructure that DATA-01 and DATA-08
W6-T1 both need before their riskiest steps ship safely — sequence DATA-09 T1-T5 ahead of DATA-01
T4/T5... in the real release plan, even though they're presented finding-by-finding here." MATRIX
CS-18 states the fail-first framing this story is directly answerable to: "platform-role seeded
cross-tenant parent/child insert succeeds today; fails after."

## Source requirements

DATA-01 (T3, T4, T5, T7, T8). Cross-referenced constraint: MATRIX CS-18 ("Tenant FK integrity"),
PLAN's PF-DATA cross-cutting note (6) (DATA-09 sequencing ahead of DATA-01 T4/T5).

## Current-state assessment

**The mismatch-audit outcome is unknown and must not be assumed.** Per mandate §18 ("Record
assumptions explicitly") and per RISK-W02-002 (epic-level `risks.md`), whether any existing row
across the 8 tenant-scoped FK edges currently has `child.tenant_id != parent.tenant_id` is a fact
about the actual data that has not yet been discovered — the audit has not yet run. PLAN's own
evidence describes the current state as structural ("nothing proves parent and child agree"), which
is evidence of an unenforced invariant, not evidence that the invariant currently holds or
currently fails. This story's T3 task is exactly the mechanism that answers the open question, and
this document does not pre-judge the answer either direction. If the audit finds a mismatch, the
remediation decision (delete the offending row, reassign its tenant, or escalate as a security
incident) is outside this story's own authority to make unilaterally — see "Risks" below and
epic-level `risks.md`'s RISK-W02-002 entry for the documented escalation path: halt T4/T5
immediately, escalate to the acceptance authority (data/reliability lead) for a remediation
decision, record the finding and its resolution in `deviations.md`, and do not proceed to
`VALIDATE CONSTRAINT` until a second zero-mismatch audit passes.

Beyond the audit's own unknown outcome, the following are confirmed absences (to be re-confirmed
at this story's actual start commit, per this programme's fail-first re-confirmation convention):
no composite FK exists on any of the 8 edges today (all 8 reference only the parent's bare `id`);
no cross-tenant negative test exists under either `app_rt` or `app_platform`; whether the platform
role (`app_platform`) bypasses FK constraints the way it is documented to bypass RLS is itself an
open question T7's own risk note flags explicitly ("Confirm platform role doesn't bypass FK
constraints — don't assume") — this story must not assume the answer either direction.

## Desired state

A zero-mismatch report (or a documented, resolved remediation decision if a mismatch was found)
exists for all 8 edges against staging/prod-shaped data. All 8 edges carry a composite FK added
`NOT VALID` and subsequently `VALIDATE CONSTRAINT`-clean, both steps performed only after
W02-E01-S001 and W02-E01-S002 have reached `accepted`. A seeded cross-tenant insert fails under
both `app_rt` and `app_platform` roles, with the platform-role result specifically confirmed rather
than assumed. As an optional, best-effort activity that does not block this story's or the epic's
closure, the redundant single-column FKs are removed once every consumer and rollback path has been
verified not to depend on the old FK's cascade behavior.

## Scope

- The mismatch audit (T3): a platform-role-connected scan proving `child.tenant_id = parent.tenant_id`
  for every existing row across all 8 edges, against staging/prod-shaped data, failing deployment
  on any mismatch found.
- Composite FK `NOT VALID` add for all 8 edges (T4), run per-table as separate statements, gated on
  W02-E01-S001+S002 acceptance.
- `VALIDATE CONSTRAINT` for each new composite FK (T5), scheduled per W02-E01-S002's backfill/
  validate-phase tooling, gated the same way, under a documented concurrent-writer-load test.
- Seeded cross-tenant insert negative tests under both `app_rt` and `app_platform` (T7), explicitly
  confirming (not assuming) whether the platform role bypasses the new FK constraints.
- Optional, best-effort removal of redundant single-column FKs (T8), performed only after all
  consumers and rollback paths are verified, explicitly not blocking this story's or the epic's P0
  closure.

## Out of scope

- **The parent `UNIQUE (tenant_id, id)` indexes, the catalog scanner, and the CI gate** —
  W02-E02-S001's scope (PLAN DATA-01 T1, T2, T6). This story depends on S001's outputs; it does not
  build them.
- **The online-migration protocol's own tooling** (manifest schema, lock budget, expand/backfill/
  validate/canary/switch/contract mechanics) — W02-E01's scope (S001, S002). This story is a
  *consumer* of that protocol for T4/T5 specifically, gated on its acceptance; it does not build
  any part of the protocol itself.
- **The actual remediation of any cross-tenant mismatch the T3 audit finds** — genuinely
  undecidable in advance, per mandate §18. If a mismatch is found, the remediation decision (delete,
  reassign tenant, or escalate as a security incident) is escalated to the acceptance authority per
  RISK-W02-002's documented path; this story's own scope is to run the audit honestly and record
  whichever outcome results, not to pre-specify or invent a remediation procedure.
- **wowsociety's own `policy_override.rule_version_id` composite-FK migration** — tracked as
  `PROD-01`, product-level, excluded from this framework-side story per mandate §2.3.

## Assumptions

- **The mismatch-audit outcome is explicitly not assumed to be zero-mismatch.** This is the single
  most important assumption boundary in this story: RISK-W02-002 exists precisely because PLAN's
  own evidence gives no basis to predict the audit's result, and this document does not manufacture
  one. See "Current-state assessment" above.
- The 8 edges and their parent/child tables are assumed to still match PLAN's own enumeration
  (persons/legal_entities/party_contacts/acting_capacities → parties; resources → organizations;
  document_versions/document_access_grants/attachments → documents/document_versions) at this
  story's actual start commit, subject to re-confirmation against W02-E02-S001's own T2 scanner
  output, which is the mechanical source of truth this story's T3/T4/T5/T7 tasks should key off
  rather than re-deriving the list by hand.
- Whether `app_platform` bypasses the new composite FK constraints is explicitly not assumed either
  direction — T7's own risk note requires this be confirmed, not assumed, and this story's plan
  records it as an open question resolved only by the T7 test's actual result.
- T8's exact verification scope (which "consumers/rollback paths" must be checked before removing a
  redundant single-column FK) is not enumerated by the source beyond "no code relies on the old FK
  name for cascade behavior" — this story's plan records the verification method (grep + full
  regression, per PLAN T8's own Tests column) rather than inventing a specific consumer list.

## Dependencies

**Intra-epic: depends on W02-E02-S001.** T3 (mismatch audit) is more useful once S001's T2 catalog
scanner confirms the exact 8-edge FK inventory is complete and current — an audit against an
incomplete inventory could silently miss an edge. T4 additionally has an intra-epic dependency on
S001's T1 (`UNIQUE (tenant_id, id)` on every parent) — PLAN's own Depends-on column for T4 lists
"T1, T3," and T1 is S001's own task.

**Cross-wave: depends on W02-E01-S001 and W02-E01-S002 — a hard gate, not a sequencing preference.**
T4 and T5 specifically (the composite FK `NOT VALID` add and its `VALIDATE CONSTRAINT`) must not
start before W02-E01-S001 and W02-E01-S002 both reach `accepted`. This is stated at wave scope
(`../../../../dependencies.md`), at epic scope (`../../dependencies.md`, `../../risks.md`'s
RISK-W02-E02-002), and is repeated here in this story's own front matter (`depends_on:
[W02-E02-S001, W02-E01-S001, W02-E01-S002]`) precisely so the gate is machine-checkable at the
story level, not merely documented in prose. The textual basis is PLAN's own PF-DATA cross-cutting
note (6): "DATA-09 is new infrastructure that DATA-01 and DATA-08 W6-T1 both need before their
riskiest steps ship safely — sequence DATA-09 T1-T5 ahead of DATA-01 T4/T5... in the real release
plan." T4's own acceptance criterion depends on W02-E01-S001's lock-timeout budget mechanism
specifically ("stays under the DATA-09 2-second lock-timeout budget"); T5's own risk note depends
on W02-E01-S002's backfill/validate-phase tooling specifically ("schedule per DATA-09's backfill/
validate phases"). T3 and T7 are not subject to this cross-wave gate — T3 is an audit (no schema
change), and T7 depends only on T5 within this story per PLAN's own Depends-on column.

T7 depends on T5 (PLAN's own Depends-on column). T8 depends on T5 and T7 (PLAN's own Depends-on
column: "T5, T7") and is explicitly optional — see "Scope" and "Acceptance criteria."

## Affected packages or components

New: the mismatch-audit tool (platform-role-connected scan, exact package location TBD, expected
adjacent to W02-E02-S001's catalog scanner); 8 composite-FK migrations (one set of `NOT VALID` add
+ `VALIDATE CONSTRAINT` statements per edge, run per-table as separate statements per T4's own risk
note); the cross-tenant negative-test suite (catalog-driven RLS matrix test, per T7's Tests column).
Potentially modified, only if T8 proceeds: the 8 existing single-column FK definitions (removal),
and any migration or code path that names the old FK constraint directly.

## Compatibility considerations

The composite FK add is `NOT VALID` specifically so it is metadata-only at add time (no existing-row
scan, no lock beyond a brief `ACCESS EXCLUSIVE` for the DDL itself) — this is precisely why T4's own
acceptance criterion ties it to "the DATA-09 2-second lock-timeout budget": a metadata-only add is
expected to comfortably clear that budget, unlike `VALIDATE CONSTRAINT` (T5), which scans existing
rows and is explicitly I/O-bound, hence scheduled per W02-E01-S002's backfill/validate phases rather
than run inline. T8's FK removal is the one change in this story with a real backward-compatibility
question — PLAN's own acceptance criterion ("No code relies on the old FK name for cascade
behavior") is the compatibility bar, verified by "Full regression + grep" per PLAN's own Tests
column, which is why T8 is optional and explicitly not gating P0 closure.

## Security considerations

T3's mismatch audit is itself a security-relevant proof: a confirmed cross-tenant mismatch would be
a live tenant-isolation breach, and RISK-W02-002 documents that possibility explicitly rather than
assuming it away. T7's cross-tenant negative test is the story's other core security control,
explicitly required to confirm — not assume — that the platform role (`app_platform`), which is
documented elsewhere in the framework to bypass RLS for legitimate cross-tenant administrative
operations, does not also bypass the new FK constraints; PLAN's own risk note states this plainly:
"Confirm platform role doesn't bypass FK constraints — don't assume." If the platform role were
found to bypass composite FK enforcement, that would itself be a new finding requiring escalation,
not a silently-accepted gap.

## Performance considerations

T5's `VALIDATE CONSTRAINT` step is I/O-bound (it scans every existing row on all 8 edges) and must
not block concurrent DML — its own acceptance criterion states this directly ("Validation doesn't
block concurrent DML"), verified by a load test under concurrent writer load. PLAN's own risk note
directs that T5 be scheduled per W02-E01-S002's backfill/validate-phase tooling specifically because
that tooling is what provides the safe scheduling/checkpointing mechanism for an I/O-bound
validation pass against a live table — this story does not invent its own scheduling mechanism, it
consumes W02-E01-S002's.

## Observability considerations

The mismatch audit (T3) must produce an inspectable, dated report — "zero-mismatch report against
staging/prod-shaped data" per its own acceptance criterion — not merely a pass/fail exit code, since
a future reader (including this story's own independent-review task) must be able to confirm what
was actually checked and what was found. The `VALIDATE CONSTRAINT` step (T5) should be observable
during its I/O-bound run (progress/duration visibility) so an operator can distinguish "still
validating" from "stalled," though the exact mechanism is an implementation-time decision informed
by W02-E01-S002's own backfill/validate observability tooling.

## Migration considerations

T4 and T5 are themselves the schema migrations this story performs: 8 composite FK `NOT VALID` adds
(T4), each subsequently `VALIDATE CONSTRAINT`-ed (T5), run per-table as separate statements per T4's
own risk note (not batched into one multi-table migration, to keep each statement's lock duration
independently measurable against the DATA-09 budget). T8's optional FK removal, if performed, is a
further migration removing the now-redundant single-column FK definitions — sequenced strictly
after T5 and T7 per PLAN's own Depends-on column.

## Documentation requirements

Document the mismatch audit's method and how to re-run it; document the composite-FK migration
sequence (`NOT VALID` add, then `VALIDATE CONSTRAINT`) and its gating on W02-E01 acceptance, so a
future reader of this story does not need to re-derive why T4/T5 could not simply run immediately
after T1/T3; document the cross-tenant negative-test suite's coverage (both roles); if T8 proceeds,
document the verification performed before FK removal (grep results, regression scope) so the
"optional, don't block P0 closure" decision and its later resolution are both traceable.

## Acceptance criteria

- **AC-W02-E02-S002-01**: The mismatch audit (T3) produces a zero-mismatch report against
  staging/prod-shaped data for all 8 edges, using a platform-role connection to bypass RLS for the
  scan, proven by an integration test that seeds a deliberate cross-tenant mismatch and confirms the
  audit detects it. If a mismatch is found in the real audit run (as opposed to the seeded test
  fixture), the finding and its resolution are recorded per RISK-W02-002's documented escalation
  path before this criterion is considered satisfied — a mismatch finding is not itself an
  acceptance-criterion failure of the *audit tool*, but T4/T5 do not proceed until it resolves to
  zero-mismatch.
- **AC-W02-E02-S002-02**: Composite FK `NOT VALID` is added for all 8 edges, run per-table as
  separate statements, each add staying under the DATA-09 2-second lock-timeout budget, proven by a
  migration lock-duration test. **This criterion cannot be satisfied — the underlying task cannot
  even start — until W02-E01-S001 and W02-E01-S002 have both reached `accepted`.**
- **AC-W02-E02-S002-03**: `VALIDATE CONSTRAINT` is run for each of the 8 new composite FKs without
  blocking concurrent DML, proven by a load test under concurrent writer load, and a second
  zero-mismatch confirmation is produced as part of this step. **This criterion cannot be satisfied
  — the underlying task cannot even start — until W02-E01-S001 and W02-E01-S002 have both reached
  `accepted`**, and it additionally cannot start before AC-W02-E02-S002-02 (T5 depends on T4).
- **AC-W02-E02-S002-04**: A seeded cross-tenant insert fails under both `app_rt` and `app_platform`
  roles, proven by a new catalog-driven RLS matrix test, with the platform-role result specifically
  confirmed (not assumed) per T7's own risk note.
- **AC-W02-E02-S002-05 (optional, non-blocking)**: The redundant single-column FKs are removed only
  after all consumers and rollback paths have been verified not to depend on the old FK's cascade
  behavior, proven by full regression plus a grep-based sweep for any reference to the old FK name.
  Per PLAN's own acceptance framing for T8 ("Optional — don't block P0 closure on it"), this
  criterion is not a precondition for this story's or the epic's `accepted` status; if not completed,
  it is recorded at closure as intentionally deferred, not as an unresolved gap.

## Required artifacts

- The mismatch-audit tool and its zero-mismatch report (or documented remediation-decision record).
- The 8 composite-FK `NOT VALID` migrations.
- The 8 `VALIDATE CONSTRAINT` migrations/statements.
- The cross-tenant negative-test suite (catalog-driven RLS matrix test).
- If T8 proceeds: the FK-removal migration and its consumer/rollback verification record.
See `artifacts/index.md`.

## Required evidence

- Zero-mismatch report (or resolved remediation-decision record) against staging/prod-shaped data.
- Integration test output for the seeded cross-tenant mismatch fixture.
- Migration lock-duration test output for the composite FK `NOT VALID` adds.
- Concurrent-writer-load test output for `VALIDATE CONSTRAINT`, plus the second zero-mismatch
  confirmation.
- Cross-tenant negative-test output under both `app_rt` and `app_platform`.
- If T8 proceeds: regression + grep sweep output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies recorded
(`depends_on: [W02-E02-S001, W02-E01-S001, W02-E01-S002]`), owner/reviewer assignment pending, the
W02-E01 gate on T4/T5 explicitly recorded (not silently assumed satisfied), the mismatch-audit
outcome explicitly recorded as unknown rather than assumed zero, T8's optional status explicitly
recorded.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; AC-W02-E02-S002-01
through -04 verified with evidence in `evidence/index.md` (AC-W02-E02-S002-05 recorded as completed
or intentionally deferred, per its own optional status); `closure.md` completed; independent review
passed per mandate §14, specifically confirming: (a) the W02-E01 gate on T4/T5 was genuinely honored
— T4/T5 were not started before W02-E01-S001 and W02-E01-S002 reached `accepted` — and (b) the
mismatch-audit outcome is honestly recorded whichever way it resolved, not silently assumed clean.

## Risks

- **RISK-W02-002** (the mismatch audit, T3, may find real cross-tenant data requiring a remediation
  decision this story's own scope cannot make unilaterally) — originates at wave scope, elaborated
  at epic scope in `../../risks.md`, and lands entirely within this story. See "Current-state
  assessment" above for this story's own explicit non-assumption of a clean result, and
  `../../risks.md` for the full mitigation/contingency: "If found: halt T4/T5 immediately, escalate
  to the acceptance authority (data/reliability lead) for a remediation decision, record in
  `deviations.md`, do not `VALIDATE CONSTRAINT` until a second zero-mismatch audit passes."
- **RISK-W02-E02-002** (T4/T5 are gated on W02-E01's acceptance; if that gate is not honored, the
  framework's highest-severity confirmed data-integrity fix would ship without the safety tooling
  DATA-09 exists to provide) — elaborated at epic scope in `../../risks.md`. This story's own
  `depends_on` front matter (`[W02-E02-S001, W02-E01-S001, W02-E01-S002]`) is the machine-checkable
  form of this risk's mitigation; this story's tasks T002/T003 (the T4/T5-equivalent tasks) each
  additionally state the gate in their own "Dependencies" section per this story's `tasks/index.md`.

## Residual-risk expectations

RISK-W02-E02-002 is expected to reduce to low residual risk as long as the `depends_on` gate is
honored at both story and task level — this is a process-discipline risk, not a technical
uncertainty. RISK-W02-002 cannot be pre-resolved by this story's planning; its outcome is a fact
about the actual data, discovered only when T3 actually runs, and remains a genuine blocking risk to
this story's own closure until resolved one way or the other and recorded honestly.

## Plan

See `plan.md`.
