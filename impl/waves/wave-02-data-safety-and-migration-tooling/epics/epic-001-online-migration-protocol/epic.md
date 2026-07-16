---
id: W02-E01
type: epic
title: Online migration protocol
status: accepted
wave: W02
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - DATA-09
depends_on: []
stories:
  - W02-E01-S001
  - W02-E01-S002
  - W02-E01-S003
decisions: []
risks:
  - RISK-W02-001
  - RISK-W02-003
---

# W02-E01 — Online migration protocol

## Epic objective

Build, from zero, a general-purpose online expand/backfill/validate/contract migration protocol —
a migration manifest schema, lock-timeout enforcement, expand-phase tooling, a resumable backfill-
job harness, validation-phase tooling, canary/N-alongside-N-1 deploy tooling, switch-phase tooling
with application rollback, contract-phase tooling gated on evidenced safety, and a full CI drill
pipeline — so that this wave's own DATA-01 composite-FK rollout (W02-E02), and later W03's grant-
table migration and W04's audit-hash-widening migration, can ship risky schema changes without a
maintenance-window outage or an ad hoc, unreviewed migration procedure.

## Problem being solved

`requirement-inventory.md` row DATA-09 records: "Online expand/backfill/validate/contract protocol
(T1–T9) | IMPL | P0 | planned | W02-E01-S001..S003 | Precedes DATA-01 T4/T5 + DATA-08 W6-T1; T9 CI
drills." PLAN's own DATA-09 section states the problem bluntly: "Reality check: this is new tooling
from zero. No expand/contract discipline, no online-DDL lock-timeout classification, no backfill-
job harness exists anywhere in wowapi today. `Makefile`'s `migrate` target is a plain forward-apply;
`check_migrations.sh` checks only registration/markers/numbering, nothing about lock duration or
backfill." MATRIX CS-21 does not separately re-derive DATA-09 (it is folded into CS-21's
deployment-readiness discussion by cross-reference: "**IDs:** DATA-01, DATA-09 (T1–T5 precede the
risky steps)" under CS-18). The gap this epic closes is structural: the framework has no mechanism
today to classify a migration's risk, enforce a lock-timeout budget, run a resumable backfill, or
prove a schema change is safe to roll forward and back — every migration today is a single
forward-apply with no safety net beyond a human's own judgment.

## Scope

- The migration manifest schema (online/maintenance classification, rows/bytes estimate, lock/
  statement timeout, N/N-1 compatibility flag, backfill owner, validation query, rollback/forward-
  fix plan) and its CI-enforced validation (S001, PLAN DATA-09 T1).
- The 2-second online-DDL lock-timeout enforcement mechanism with abort-and-retry and a bounded
  retry ceiling (S001, PLAN DATA-09 T2).
- Expand-phase tooling: nullable/default-safe columns, new tables/indexes/compatibility views,
  `NOT VALID` constraints, non-transactional `CREATE INDEX CONCURRENTLY` (S002, PLAN DATA-09 T3).
- The resumable, tenant-scoped, keyset-paginated, checkpointed backfill-job harness with bounded
  batch/tx time and rate controls, including its interim checkpoint-lease mechanism pending
  DATA-02's shared primitive (S002, PLAN DATA-09 T4).
- Validation-phase tooling: `VALIDATE CONSTRAINT` orchestration plus reconciliation queries with
  machine-checked artifact capture (S002, PLAN DATA-09 T5).
- Canary/deploy-N tooling with soak metrics proving N alongside N-1 (S003, PLAN DATA-09 T6).
- Switch-phase tooling: observable compatibility flag, dual-schema-version consumer support, proven
  application rollback after switch (S003, PLAN DATA-09 T7).
- Contract-phase tooling gated on an evidenced no-N-1-remains precondition, proving forward recovery
  from every failed phase (S003, PLAN DATA-09 T8).
- The full CI drill pipeline covering all 6 directive-named drills (S003, PLAN DATA-09 T9).

## Out of scope

- **DATA-01's own composite-FK migration content** (which parent tables get a unique index, which
  8 edges get a composite FK) — that is W02-E02's scope. This epic builds the protocol DATA-01
  T4/T5 will run through; it does not itself perform DATA-01's schema change.
- **DATA-02's full shared lease/fencing primitive** (generations, heartbeats, job-claim fencing) —
  W04-E01-S001's scope. This epic's S002 builds only the minimal checkpoint-lease subset DATA-09
  T4's backfill harness needs (see `stories/story-002-expand-backfill-validate/story.md` and
  `risks.md` RISK-W02-001); it does not build the general-purpose primitive PLAN DATA-02 T1
  describes as "a reusable kernel building block for DATA-02/03/04."
- **DATA-08 W6-T1's audit hash-widening migration content** — W04-E04-S001's scope. This epic's
  protocol is a prerequisite tool for that migration, not the migration itself.
- **Choosing which real migration is the first live exercise of this protocol** — PLAN DATA-09 T9's
  own risk note names DATA-01's composite-FK rollout as "the natural first candidate," and this
  wave's W02-E02 does become that first real consumer, but the decision of scheduling and executing
  a live production rollout through this protocol is an operational decision outside this epic's own
  scope (this epic delivers the tooling and its CI-drill proof, not a live production migration
  event).
- **The soak-duration/threshold numeric calibration for canary tooling** — PLAN's own risk note
  states this is "a genuine, currently unresolvable judgment gap" given no production telemetry
  baseline exists. This epic's S003 builds T6's tooling to accept configurable thresholds; it does
  not — cannot — derive the "correct" threshold values from evidence this epic has access to. See
  RISK-W02-003.

## Source requirements

DATA-09. No MATRIX CS-ID directly owns DATA-09 as a dedicated closure spec (it is referenced from
within CS-18's "IDs" line as the tooling DATA-01 depends on); DATA-09 is confirmed to have no D-0N
architecture-decision dependency in either `requirement-inventory.md` §B or REVIEW §F/§U.

## Architectural context

DATA-09 is new infrastructure, not a fix to an existing subsystem — PLAN's own framing is "this is
new tooling from zero." It sits architecturally as migration-execution tooling layered over the
existing PostgreSQL/pgx toolchain and the existing `Makefile`/`check_migrations.sh` migration
registration mechanism; it does not replace that registration mechanism, it adds a risk-
classification, lock-budget-enforcement, backfill, validation, canary, switch, and contract
discipline on top of it. The nine tasks (T1–T9) form a strict phase pipeline matching the protocol's
own name — expand, backfill, validate, canary/switch, contract — with T9 (the CI drill pipeline)
depending on all eight preceding tasks. This epic's three stories are grouped by phase-cluster, not
by task count alone: S001 covers the manifest-and-budget foundation every later phase consumes
(T1, T2); S002 covers the "make the schema change and move the data safely" phase (T3, T4, T5); S003
covers the "prove it's safe to go live and safe to roll back" phase plus the CI-drill proof (T6, T7,
T8, T9). This grouping is fixed by `impl/analysis/wave-allocation-detail.md`'s canonical allocation
and is not to be regrouped.

The forward dependency between S002's backfill harness (T4) and DATA-02 T1's shared lease primitive
(W04 scope, not yet built) is the epic's principal architectural risk — see "Risks" below and
`stories/story-002-expand-backfill-validate/story.md`'s own treatment.

## Included stories

- **W02-E01-S001 — manifest-and-lock-budget** (PLAN DATA-09 T1, T2): the migration manifest schema
  and its CI validation; the 2-second lock-timeout enforcement with bounded abort-and-retry.
- **W02-E01-S002 — expand-backfill-validate** (PLAN DATA-09 T3, T4, T5): expand-phase tooling; the
  resumable backfill-job harness (with its interim checkpoint-lease, forward-dependency-flagged);
  validation-phase tooling with machine-checked artifacts.
- **W02-E01-S003 — canary-switch-contract-drills** (PLAN DATA-09 T6, T7, T8, T9): canary/deploy-N
  soak tooling; switch-phase tooling with application rollback; contract-phase tooling gated on
  evidenced safety; the full 6-drill CI pipeline, plus an evidence-aggregation task consolidating
  T6–T9's individually-named drill outputs into one consolidated evidence record (see
  `stories/story-003-canary-switch-contract-drills/tasks/index.md` for the rationale).

## Dependencies

No dependency on any other W02 epic — this epic is the wave's foundation and every other W02 epic
either depends on it (W02-E02) or is independent of it (W02-E03, W02-E04, W02-E05). This epic
depends only on W00's exit gate (baseline state captured) per `wave.md`'s entry criteria. Downstream:
W02-E02 (DATA-01 T4/T5), and — beyond this wave — W03-E01-S001 (SEC-01 grant-table migration) and
W04-E04-S001 (DATA-08 W6-T1 audit-hash migration) both depend on this epic's protocol existing, per
`../../dependencies.md` (wave-level).

## Risks

RISK-W02-001 (the S002 minimal-checkpoint-lease deviation, superseded by W04-E01-S001) and
RISK-W02-003 (the S003 canary soak-threshold judgment gap, PLAN's own "currently unresolvable" note)
both originate at wave scope and land entirely within this epic's stories. See `risks.md` for the
epic-scoped elaboration.

## Required decisions

None. DATA-09 has no D-0N architecture-decision dependency in the source (confirmed — see `wave.md`
"Assumptions"). This epic's stories accordingly carry no `decisions/` directory.

## Epic acceptance criteria

- **AC-W02-E01-01**: The migration manifest schema validates every migration's required fields
  (online/maintenance classification, rows/bytes, lock/statement timeout, N/N-1 flag, backfill
  owner, validation query, rollback/forward-fix plan); a migration missing a required field fails
  CI. The 2-second online-DDL lock-timeout enforcement aborts cleanly on a concurrently-locked table
  and retries within a bounded ceiling — no unbounded retry.
- **AC-W02-E01-02**: Expand-phase tooling issues `CREATE INDEX CONCURRENTLY` and `NOT VALID`
  constraints without blocking traffic, proven by an old-reader-compatibility test. The backfill-job
  harness passes its named interrupted/resumed test with no reprocessing or skipping. Validation-
  phase tooling produces a machine-checked, artifact-schema-conformant zero-mismatch report, not
  prose.
- **AC-W02-E01-03**: Canary/deploy-N tooling proves N-1 code runs correctly against N-expanded
  schema both before and after backfill. Switch-phase tooling proves application rollback after
  switch with no destructive `Down`. Contract-phase tooling proves forward recovery from every
  failed phase and gates the contract step on an evidenced no-N-1-remains precondition. All six
  directive-named drills run in the CI/scheduled pipeline with a passing run artifact.
- **AC-W02-E01-04**: All three stories have passed independent review per mandate §14, with S002
  specifically checked for the interim-checkpoint-lease deviation being correctly recorded (not
  silently absorbed as if it were DATA-02's full primitive) and S003 specifically checked for the
  soak-threshold judgment gap being honestly recorded as an accepted residual risk, not silently
  resolved with an invented number.

## Closure conditions

All three stories reach `accepted` (each satisfying its own `closure.md`); AC-W02-E01-01 through
AC-W02-E01-04 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; the RISK-W02-001 deviation is recorded with a clear pointer for
W04-E01-S001 to consume, and RISK-W02-003 is recorded as an accepted residual risk — neither is
silently dropped at closure.

## Status update (2026-07-16)

`status: accepted` — all three stories independently reviewed and accepted, superseding this epic's
prior `planned` state. See the per-story independent-review task files under this wave's
`epic-002-tenant-fk-integrity` (task-006) and sibling stories, dated 2026-07-16.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
