---
id: W02
type: wave
title: Data-safety and migration tooling
status: partially-accepted
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-16
included_epics:
  - W02-E01
  - W02-E02
  - W02-E03
  - W02-E04
  - W02-E05
depends_on:
  - W00
blocks:
  - W03
  - W04
source_requirements:
  - DATA-09
  - DATA-01
  - DATA-05
  - DATA-06
  - FBL-02
  - CS-18
  - CS-21
---

# W02 — Data-safety and migration tooling

## Objective

Build the online expand/backfill/validate/contract migration protocol from zero (DATA-09), then use
it to close the highest-severity confirmed data-integrity gap in the framework — tenant-scoped
foreign keys that reference only a parent's `id` and never `(tenant_id, id)` (DATA-01) — followed by
two independent persistence-correctness findings (version-allocation races/blob GC, DATA-05; the
resource-mirror aggregate write contract, DATA-06) and the production seed-sync path that currently
does not exist at all (FBL-02). This wave converts "the framework has no discipline for shipping a
risky schema change safely" into "the framework has a protocol, and its first two real consumers —
the composite tenant FK and the production catalog seed path — are built on it."

## Rationale

`impl/index.md`'s wave map assigns W02 "DATA-09 online-migration protocol, then DATA-01 tenant FKs
over it; DATA-05/06; FBL-02 prod seed-sync," depending only on W00. `requirement-inventory.md` row
DATA-09 states the dependency explicitly: "Precedes DATA-01 T4/T5 + DATA-08 W6-T1; T9 CI drills."
PLAN's own PF-DATA cross-cutting note (6) makes the sequencing argument directly: "DATA-09 is new
infrastructure that DATA-01 and DATA-08 W6-T1 both need before their riskiest steps ship safely —
sequence DATA-09 T1-T5 ahead of DATA-01 T4/T5 and DATA-08 W6-T1 in the real release plan, even
though they're presented finding-by-finding [in the plan]." DATA-01 is P0 (per
`requirement-inventory.md`: "Composite tenant FKs (T1–T8) | IMPL | P0 | planned") because it is a
live tenant-isolation gap: MATRIX CS-18's fail-first framing states plainly, "platform-role seeded
cross-tenant parent/child insert succeeds today; fails after." DATA-05 and DATA-06 are independent
P1 persistence-correctness findings that do not depend on DATA-09 or DATA-01 and are grouped into
this wave because they share the wave's persistence-and-migration-tooling theme
(`impl/index.md`'s W02 title) and have no dependency forcing them into a later wave. FBL-02 is
graded P0-prod in `requirement-inventory.md` ("Production seed-sync path (PF-9) | IMPL | P0-prod |
planned | W02-E05-S001 | CS-21 acceptance bar fixed") because, per MATRIX CS-21, a production boot
against an empty catalog database silently produces a deny-everything system today — this is a
production-blocking gap, not a theoretical one, and it is grouped into W02 because its design
depends on nothing from W01/W03 and because its "versioned catalog manifest" concept sits naturally
alongside DATA-09's own manifest-driven migration discipline.

## Framework capabilities delivered

- A durable, general-purpose online expand/backfill/validate/contract migration protocol: a
  migration manifest schema, lock-timeout enforcement with abort-and-retry, expand-phase tooling,
  a resumable/checkpointed backfill-job harness, validation-phase tooling with machine-checked
  artifacts, canary/N-alongside-N-1 deploy tooling, switch-phase tooling with application rollback,
  contract-phase tooling gated on evidenced absence of N-1 traffic, and a full CI drill pipeline —
  MATRIX CS-21's dedup anchor and PLAN DATA-09's T1–T9.
- Composite, tenant-aware foreign keys on every tenant-scoped child table that today references
  only its parent's bare `id` — closing a real, confirmed cross-tenant data-integrity gap (MATRIX
  CS-18, PLAN DATA-01).
- Race-free, concurrency-safe version allocation for `kernel/artifact` and `kernel/document`
  (replacing inline `MAX(version)+1` reads), plus durable upload-session tracking and scheduled
  garbage collection of orphaned blobs (PLAN DATA-05).
- A framework-enforced aggregate write contract: a typed repository/unit-of-work helper that makes
  it structurally impossible for a module to write its business row without the framework also
  writing the resource mirror, audit row, and outbox entry in the same transaction, with real actor
  attribution replacing today's `uuid.Nil` placeholder (PLAN DATA-06).
- A production seed-sync path (`wowapi seed sync --env prod` or equivalent): idempotent,
  RLS-respecting, versioned catalog manifests, dry-run plus audit, wired so that a production-profile
  boot against an empty catalog database reaches `ready` only after seed-sync has run, and readiness
  reports the seed/catalog hash (MATRIX CS-21, FBL-02).

## Included epics

- **W02-E01 — online-migration-protocol**: the DATA-09 manifest schema, lock-timeout enforcement,
  expand/backfill/validate/canary/switch/contract tooling, and the full CI drill pipeline.
- **W02-E02 — tenant-fk-integrity**: the DATA-01 composite tenant-FK rollout, built on E01's
  protocol for its riskiest steps (T4/T5).
- **W02-E03 — version-allocation-and-gc**: the DATA-05 version-allocation race fix and blob GC.
- **W02-E04 — aggregate-write-contract**: the DATA-06 resource-mirror aggregate write contract and
  actor attribution.
- **W02-E05 — production-seed-sync**: the FBL-02 production catalog seed-sync path.

## Entry criteria

- W00's exit gate satisfied: the 8 executed finding-slices re-verified at current HEAD, baselines
  captured, D-01..D-09 ratified as ADRs. No W02 story depends on a specific D-0N ADR (see
  "Dependencies" below and each epic's "Required decisions" — none identified for W02), but W02
  still depends on W00's baseline/coverage/lint state as its starting point per the strict
  W00→W07 entry ordering.

## Exit criteria

- DATA-09's manifest schema, lock-timeout enforcement, expand-phase tooling, backfill-job harness
  (interrupted/resumed test explicitly passing), validation-phase tooling, canary/switch tooling
  (application-rollback-after-switch test explicitly passing), contract-phase tooling (forward
  recovery from every failed phase test explicitly passing), and the full 6-drill CI pipeline are
  all in place and evidenced — PLAN DATA-09 T1–T9's acceptance criteria satisfied in full.
- DATA-01's composite tenant FKs are `VALIDATE CONSTRAINT`-clean on all 8 confirmed edges, the
  tenant-FK catalog scanner is wired as a permanent CI gate, the mismatch audit reports zero
  cross-tenant mismatches (or a documented remediation decision if it does not — see `risks.md`),
  and seeded cross-tenant insert negative tests fail under both `app_rt` and `app_platform` — PLAN
  DATA-01 T1–T7 satisfied (T8, FK cleanup, is explicitly optional per its own acceptance row and is
  not a wave exit blocker).
- DATA-05's version allocation is race-free under concurrent load for both `kernel/artifact` and
  `kernel/document`, upload sessions are durable, confirmation is atomic (CAS), and scheduled GC
  never removes a referenced object — PLAN DATA-05 T1–T5 satisfied.
- DATA-06's aggregate write helper makes a module unable to write its business row without also
  writing the mirror in the same transaction under fault injection at each of 4 stages, real actor
  attribution replaces `uuid.Nil`, and the reference handler is migrated onto the new helper — PLAN
  DATA-06 T1–T4 satisfied.
- FBL-02's seed-sync path exists and is idempotent and RLS-respecting; a prod-profile boot against
  an empty catalog database reaches readiness only after seed-sync runs; the readiness payload
  reports the seed/catalog hash — MATRIX CS-21's fixed acceptance bar satisfied.

## Dependencies

Depends on W00 (full wave, per `impl/index.md`'s wave map: "W02 | data-safety-and-migration-tooling
| ... | Depends on | W00 | Epics | 5"). No dependency on W01 — W02's scope is disjoint from W01's
zero-dependency-hardening work (linters, observability, HTTP hardening, generator/doc fixes); the
two waves may in principle execute in parallel once each has independently satisfied W00's exit
gate, though the program's strict W00→W01→W02 entry ordering (`impl/index.md`: "Execution order:
strictly W00→W07 for wave entry") sequences W02 after W01 by convention, not by a technical
dependency this wave's own stories require. See `dependencies.md` for the full upstream/downstream
detail, including the forward-dependency risk recorded for E01-S002's minimal checkpoint lease
(superseded by W04-E01-S001) and the internal E02-on-E01 gating.

## Assumptions

- The 8 tenant-scoped child tables named in PLAN DATA-01's evidence (`persons`, `legal_entities`,
  `party_contacts`, `acting_capacities` → `parties`; `resources` → `organizations`;
  `document_versions`, `document_access_grants`, `attachments` → `documents`/`document_versions`)
  are assumed to still be exactly the affected set at this wave's actual start commit, subject to
  E02-S001-T2's own catalog-scanner re-confirmation (the scanner's whole purpose, per PLAN DATA-01
  T2, is to enumerate the known FKs mechanically rather than trust a hand-maintained list).
  Consistent with the mandate's fail-first re-confirmation posture applied elsewhere in this
  programme (e.g. W01-E01-S001's re-run-before-trusting-the-cited-snapshot pattern) — illustrative
  of programme convention, not a load-bearing dependency of this wave.
- DATA-09's own protocol design (T1–T9) is assumed to require no D-0N architecture decision — a
  scan of `requirement-inventory.md` §B and REVIEW §F/§U finds no D-0N row citing DATA-09 as its
  consumer (the nine ADR-consuming rows in W00-E02-S003's story are D-01→SEC-01/W03, D-02→AR-02/W05,
  D-03→AR-01/W05, D-04→DATA-08 W6/W04, D-05→REL-01/W06, D-06→SEC-04/W05, D-07→SEC-06/W03,
  D-08→FBL-06/W01, D-09→secrets docs/W01 — none is DATA-09, DATA-01, DATA-05, or DATA-06). This is
  confirmed, not assumed, from the source text; it is recorded here because the task brief for this
  wave explicitly asked for confirmation. See "Required decisions" in each epic (`epic.md`) — all
  state none.
- FBL-02's seed-sync design detail (catalog manifest format, versioning scheme) is explicitly
  unresolved in the source: MATRIX CS-21 states "design detail to be ratified in Phase 5, but the
  acceptance bar is fixed now." W02-E05-S001 accordingly contains a design-investigation task before
  any implementation task, per this wave's task brief and per mandate §18 ("Where implementation
  details cannot yet be known, state what must be determined during the story rather than inventing
  specifics").

## Risks

See `risks.md`. Headline risks: E01-S002's minimal checkpoint lease is a planned, recorded
technical-debt-bearing deviation superseded by W04-E01-S001 (not a silent shortcut); DATA-01's
mismatch audit (E02-S002-T3/PLAN DATA-01 T3) may find real cross-tenant data requiring a
remediation decision before `VALIDATE CONSTRAINT` can proceed, which would block E02-S002's
completion on a decision this wave cannot make unilaterally; DATA-09's own soak-duration/
canary-threshold judgment gap (PLAN DATA-09 T6: "No production telemetry baseline exists — soak
duration/thresholds are a genuine, currently unresolvable judgment gap") has no resolution path
inside this wave and is recorded as an accepted residual risk, not silently closed.

## Quality gates

- Every DATA-09 tooling task's acceptance criterion that names an explicit test (the interrupted/
  resumed backfill test, the N-1/N canary soak test, the application-rollback-after-switch test, the
  forward-recovery-from-every-failed-phase test) is proven with that named test as fail-first
  evidence, per PLAN DATA-09's own "Tests" column — not asserted from code review alone.
- DATA-01's fail-first evidence is MATRIX CS-18's own framing: "platform-role seeded cross-tenant
  parent/child insert succeeds today; fails after" — this exact adversarial test is required, not a
  substitute.
- DATA-05's concurrency claim ("N concurrent callers → N unique monotonic versions") is proven with
  a ≥20-concurrent-caller test per PLAN DATA-05 T1's own "Tests" column.
- DATA-06's atomicity claim is proven with fault injection at each of 4 stages independently, full
  rollback at every stage, per PLAN DATA-06 T1's own "Tests" column.
- FBL-02's acceptance bar (MATRIX CS-21) is proven with the named fail-first test: "prod boot with
  empty catalogs → currently silently deny-everything, after: named readiness failure [until
  seed-sync runs]."

## Required artifacts

- DATA-09: migration manifest schema definition; lock-timeout enforcement module; expand-phase
  tooling; backfill-job harness (a reusable kernel building block); validation-phase artifact
  schema; canary/soak tooling; switch-phase tooling; contract-phase gate; CI drill pipeline
  definition.
- DATA-01: `UNIQUE (tenant_id, id)` migrations on 8 parent tables; the tenant-FK catalog scanner
  (permanent CI gate); the mismatch-audit tool; `NOT VALID` composite FK migrations; the CI gate
  test fixture.
- DATA-05: the locked-counter/sequence-row version-allocation change for `kernel/artifact` and
  `kernel/document`; the upload-session schema and table; the GC sweep job.
- DATA-06: the typed aggregate repository/unit-of-work helper; the `registrar_pg.go` actor-
  attribution fix; the migrated reference handler; updated `kernel/resource` documentation.
- FBL-02: the seed-sync command/path; the catalog manifest format (design artifact, per the
  investigation task); readiness-payload seed/catalog-hash reporting.

## Required evidence

- DATA-09: interrupted/resumed backfill test output; lock-duration migration test output;
  old-reader-compatibility test output; zero-mismatch validation artifacts; N-1/N canary soak
  metrics; application-rollback-after-switch test output; forward-recovery test output; full CI
  drill pipeline run artifact.
- DATA-01: `pg_indexes` migration test output; fixture-schema catalog-scanner test output;
  cross-tenant mismatch-audit report; migration lock-duration test output; concurrent-writer load
  test output; negative fixture migration (CI gate) output; seeded cross-tenant insert negative
  test output under both roles.
- DATA-05: ≥20-concurrent-caller allocation test output; crash-simulation upload-session test
  output; racing-confirm CAS test output; mixed confirmed/expired/pending GC test output.
- DATA-06: per-stage fault-injection test output (4 stages); actor-attribution test output
  (with/without actor, system vs. user path); reference-handler regression test output.
- FBL-02: seed-sync dry-run and audit-log output; empty-catalog fail-first/pass-after readiness test
  output; seed/catalog hash reporting test output.

## Expected implementation outcome

A framework that can ship a risky schema change (like the composite tenant FK this wave itself
adds) without a maintenance-window outage, using its own general-purpose expand/backfill/validate/
contract discipline instead of ad hoc per-migration judgment calls; a framework in which the
parent-child tenant relationship a row's own row-level-security policy assumes is now actually
enforced by a database constraint, not merely implied; a framework whose version allocation and
resource-mirror write path cannot silently race or silently omit the mirror/audit/outbox rows; and
a framework that no longer silently boots into a deny-everything production system when its catalog
database is empty.

## Acceptance authority

Data/reliability lead — per PLAN §5.3's own "Accountable role: data/reliability lead" for PF-DATA,
applied uniformly across this wave's five epics (FBL-02, though sourced from the REVIEW/MATRIX
rather than PLAN's PF-DATA table, is a deployment-readiness/seed-sync concern that shares the same
data/reliability accountability).

## Closure conditions

All exit criteria satisfied; all five epics' `closure-report.md` accepted; `waves/index.md`'s W02
row updated to reflect `accepted` status; no unresolved regression from the composite-FK rollout or
the aggregate write contract change; the E01-S002 minimal-checkpoint-lease deviation is recorded
(not resolved — its resolution is W04-E01-S001's responsibility) with a clear pointer for W04 to
pick up; DATA-01's mismatch-audit outcome (zero-mismatch, or a documented remediation decision) is
resolved before this wave's E02-S002 can be marked `accepted`, per `risks.md`.

## Status update (2026-07-16)

`status: partially-accepted` — 7 of 8 W02 stories across the 5 epics independently reviewed and
accepted per `review-gate-2026-07-16.md`-basis records (this wave's per-story independent-review
task files, dated 2026-07-16), superseding the prior uncorroborated `W02ReviewGate, 2026-07-13`
citation. The exception is W02-E02-S002, rolled back from `accepted` to `implemented` because
three of its named proof artifacts were never built (**DEV-PROG-005**); its disposition —
build-and-re-accept vs. formal descope — is the open decision **DEC-PROG-003**. The wave returns
to `accepted` when that decision is executed and E02 re-accepts.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
