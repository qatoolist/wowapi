---
id: W04
type: wave
title: Jobs and durable delivery
status: in-progress
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
included_epics:
  - W04-E01
  - W04-E02
  - W04-E03
  - W04-E04
depends_on:
  - W02
blocks: []
source_requirements:
  - DATA-02
  - DATA-03
  - DATA-04
  - DATA-08
  - DX-07
  - FBL-04
  - CS-11
  - CS-20
  - CS-21
  - D-04
---

# W04 — Jobs and durable delivery

## Objective

Build the shared lease/fencing primitive that closes the confirmed duplicate-effect race in the
jobs queue (DATA-02), reuse that primitive to move remote provider/secret I/O outside database
transactions for notify and webhook delivery (DATA-03) and to make bulk multi-worker processing
actually safe rather than documented-safe (DATA-04); adopt `cenkalti/backoff/v5` in place of the
framework's duplicated hand-rolled retry logic (FBL-04); widen the audit hash chain to cover every
persisted field with a version-branched verification scheme (DATA-08 Wave-6, enacting D-04); and
make the framework's own readiness and configuration diagnostics truthful (DX-07). This wave
converts "the framework's at-least-once job/delivery story has no fencing, so it is
at-least-once-with-overwrites" (MATRIX CS-11) into a framework whose jobs, remote deliveries, and
bulk processors cannot silently duplicate an external effect on worker failover, and whose audit
trail and own health signals are no longer partially blind to tampering or lying about readiness.

## Rationale

`impl/index.md`'s wave map assigns W04 "Shared lease/fencing primitive → DATA-02/03/04; FBL-04
retry adoption; DATA-08 W6 audit integrity (D-04); DX-07 readiness truthfulness," depending on W02
("DATA-09 for W6-T1 migration"). `requirement-inventory.md` row DATA-02 states plainly: "T1 shared
primitive is keystone" — PLAN's own PF-DATA cross-cutting note (1) is unambiguous: "The shared
lease primitive (DATA-02 T1) is the single highest-leverage build in this package — staff and
design-review it first." DATA-02, DATA-03, and DATA-04 are grouped into one wave because they are
architecturally one build, not three: MATRIX CS-11 frames them as a single closure spec ("Jobs,
outbox, lease/fencing, drain") with "the single shared lease/fencing primitive (DATA-02 T1) first,
then three-stage claim→effect-outside-tx→fenced-finalize for notify/webhook, SKIP-LOCKED bounded
batch for bulk, chaos tests at every named boundary." FBL-04 rides in this wave because it is a
small, well-bounded item sharing no dependency with anything else and fitting naturally alongside
DATA-03's remote-I/O retry paths. DATA-08's Wave-6 tasks are P0/P1 and MATRIX CS-20 fixes their
target state on D-04 (ratified in `requirement-inventory.md` §B: "D-01..D-09 | Nine ratified
architecture decisions ... enacted inside their target stories"); W6-T1 is graded "Single
highest-risk task in PF-DATA's Wave-6 scope" (PLAN DATA-08 W6-T1 risk column) precisely because it
is a breaking audit-hash format change touching wowsociety's live rows, and it depends on W02-E01's
online-migration protocol to ship safely — the one dependency this wave actually has on W02, not a
blanket dependency of every W04 epic. DX-07 closes this wave because MATRIX CS-21 ties it to the
same deployment-readiness closure spec as FBL-02 (W02 scope) and because its T4 is itself
deferred-linked to AR-04 T5's waiver mechanism (W05 scope) — DX-07 T1-T3 are buildable now and are
grouped here per `wave-allocation-detail.md`'s canonical allocation.

## Framework capabilities delivered

- A shared, reusable lease/fencing primitive (`lease_token`, monotonic `lease_generation`,
  `lease_expires_at`, optional heartbeat) used identically by jobs, notify/webhook delivery, and
  bulk processing — not three independent copies (PLAN DATA-02 T1's own acceptance criterion: "One
  primitive reused ≥3 times").
- Fenced job claim, finalize, and reclaim paths on `jobs_queue`, closing the confirmed race where a
  reclaimed worker's late finalize silently overwrites the reclaiming worker's outcome.
- A stable job idempotency contract: every worker declares exactly one duplicate-safety mechanism
  (inbox/effect ledger, domain CAS, or provider idempotency key) and cannot register without doing
  so.
- A three-stage claim→effect-outside-tx→fenced-finalize protocol for `kernel/notify` and
  `kernel/webhook.deliverToEndpoint`, removing remote network/secret I/O from open database
  transactions.
- An inbound two-phase webhook-verification protocol immune to a secret rotation/deactivation race
  between snapshot and verification, plus a body-free audit row on failed signature verification.
- A per-adapter idempotency-safety contract declaration, boot-time-enforced.
- A corrected, fenced, `SKIP LOCKED`-honest bulk multi-worker claim path replacing today's
  documented-safe-but-actually-unsafe plain unlocked `SELECT`, plus pause/resume/cancel lifecycle
  controls.
- A shared chaos-test harness (DATA-02 T7), reused — not reimplemented — by DATA-03's 6-boundary
  chaos test and DATA-04's multi-worker chaos test.
- `cenkalti/backoff/v5` adopted in place of the framework's two duplicated hand-rolled retry
  implementations, with retry-schedule parity and fault-injection tests proving the replacement is
  behaviorally equivalent or better.
- An audit hash chain (`kernel/audit.chainHash`) widened to cover every persisted field — including
  the previously-excluded canonicalized `metadata` and `tx_id` — with a `hash_version smallint`
  discriminator column (D-04) so historical rows verify under a v1 branch and new rows under v2.
- External anchor verification for the audit chain, encrypted immutable DSR export artifacts,
  central legal-hold enforcement every dispose/erase callback must pass through, and explicit
  partial/not-applicable DSR results per registered record class.
- Truthful readiness diagnostics: a migration-currency check, seed/rule/model-hash reporting, and a
  `config doctor` that discovers the product root via `go env GOMOD`/`--project` rather than a
  CWD-relative `os.Stat` that silently falls back to framework-only validation.

## Included epics

- **W04-E01 — lease-fencing-primitive-and-jobs**: the DATA-02 shared lease/fencing primitive
  (the wave's keystone build) and its application to the jobs queue's claim, finalize, and reclaim
  paths, plus the idempotency contract and the shared chaos harness.
- **W04-E02 — remote-io-outside-tx**: the DATA-03 three-stage protocol for notify/webhook delivery,
  inbound two-phase verification, the adapter idempotency contract, and FBL-04's retry-library
  adoption.
- **W04-E03 — bulk-multi-worker-safety**: the DATA-04 stopgap fix and the full leased-claim,
  fenced-finalize, lifecycle-controlled rewrite of bulk multi-worker processing.
- **W04-E04 — compliance-and-readiness**: the DATA-08 Wave-6 audit-hash widening (enacting D-04),
  external anchoring, DSR export/legal-hold/explicit-status work, and DX-07's readiness/config
  diagnostics truthfulness (T1-T3; T4 deferred-linked to W05).

## Entry criteria

- W02's exit gate satisfied for **W02-E01 specifically** (the DATA-09 online-migration protocol) —
  this is the one concrete predecessor capability this wave's own stories require, and only for
  W04-E04-S001 (DATA-08 W6-T1's audit-hash migration). Per `impl/index.md`'s wave map ("Depends on:
  W02 (DATA-09 for W6-T1 migration)") and confirmed at story grain in `dependencies.md`: DATA-02,
  DATA-03, and DATA-04 (E01, E02, E03) have **no** dependency on W02's protocol and may enter as
  soon as this wave's own entry gate (W00's baseline, per the strict W00→W07 ordering) is satisfied.
- E03-S001 (DATA-04 T1, the immediate stopgap) may start at wave entry independent of E01's
  primitive landing — `wave-allocation-detail.md`: "S001 T1 stopgap (can start at wave entry)."

## Exit criteria

- DATA-02's shared lease/fencing primitive is implemented and reused, not copied, across jobs
  (E01), notify/webhook (E02), and bulk (E03); the named chaos test
  `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` passes, proving exactly one logical effect
  recorded and the stale worker's writes rejected at all three named boundaries — PLAN DATA-02
  T1–T7 acceptance criteria satisfied.
- DATA-03's three-stage protocol removes every remote network/secret-resolution call from an open
  database transaction for `kernel/notify` and `kernel/webhook`; the inbound two-phase verification
  protocol is immune to a rotation-during-verification race; the 6-boundary chaos test passes with
  zero duplicate external effects — PLAN DATA-03 T1–T6, T8 satisfied (T7 is a cross-reference to
  DATA-08 W0-T2, already executed, not a W04 deliverable).
- FBL-04's `cenkalti/backoff/v5` adoption replaces both hand-rolled retry implementations with
  retry-schedule parity proven under fault injection.
- DATA-04's corrected migration comment and fenced, `SKIP LOCKED`-honest claim path pass a
  concurrency test with ≥2 processors; the named chaos test
  `DATA-04/chaos/duplicate_worker_test.go` passes; pause/resume/cancel lifecycle controls behave
  correctly mid-run — PLAN DATA-04 T1–T6 satisfied.
- DATA-08 W6-T1's widened `chainHash` breaks verification on mutation of any declared field
  (including metadata and tx_id) under a per-field tamper test; historical rows verify under the
  `hash_version=1` branch; W6-T2 through W6-T5 (external anchoring, encrypted DSR export, central
  legal-hold enforcement, explicit per-class status) are all evidenced per their own named tests —
  PLAN DATA-08 W6-T1–T5 satisfied.
- DX-07's readiness payload fails when applied-migration version lags expected, reports seed/rule/
  model-hash, and `config doctor` discovers the product root correctly regardless of invocation
  directory — PLAN DX-07 T1–T3 satisfied; T4 is explicitly out of scope, deferred-linked to
  W05-E03-S002's AR-04 T5 waiver mechanism (not yet built).

## Dependencies

Depends on W02 — but narrowly: only W04-E04-S001 (DATA-08 W6-T1) depends on W02-E01's online-
migration protocol, per `impl/analysis/wave-allocation-detail.md`'s own row ("dep W02-E01
protocol") and per `impl/waves/wave-02-data-safety-and-migration-tooling/dependencies.md`'s
downstream table, which names this exact edge. DATA-02 (E01), DATA-03 (E02), and DATA-04 (E03) have
no dependency on W02's protocol or on W02's DATA-01 tenant-FK work — confirmed by
`requirement-inventory.md`'s own notes column for each row, none of which cites a W02 dependency
beyond the wave-level entry-ordering convention. See `dependencies.md` for the full statement,
including the E01-S001 supersession of W02-E01-S002's interim checkpoint lease.

## Assumptions

- DATA-02 T1 (this wave's E01-S001) is confirmed, not assumed, to supersede W02-E01-S002's minimal
  checkpoint lease: `impl/analysis/wave-allocation-detail.md`'s W04 block states this explicitly
  ("S001 shared-primitive (T1 — replaces W02-E01-S002's minimal checkpoint lease; migration
  note)"), and the "Cross-wave sequencing notes" at that file's bottom record it a second time:
  "W02-E01-S002's minimal checkpoint lease is intentionally superseded by W04-E01-S001 (recorded
  here so the deviation is planned, not silent)." This is a planned transition, not a silent scope
  reduction — see E01-S001's own `story.md`/`plan.md` and `risks.md` (RISK-W04-001).
- DX-07 T4 is confirmed, not assumed, to require AR-04 T5's waiver mechanism before it can be
  implemented: PLAN DX-07 T4's own dependency column states "T1-T3, AR-04's waiver framework," and
  `wave-allocation-detail.md`'s W05-E03-S002 row confirms T5 "builds the shared waiver mechanism
  consumed by SEC-06/DX-07." W05 has not been built as of this wave's planning; T4 is recorded as
  explicitly out of scope for E04-S003, forward-referenced by requirement ID, not by a file path
  that may not yet exist.
- A scan of `requirement-inventory.md` §B for any D-0N row targeting DATA-02, DATA-03, DATA-04, or
  DX-07 finds none — only DATA-08 W6 (D-04) enacts a decision in this wave. This is confirmed, not
  assumed, from the source text: the nine ADR-consuming rows are D-01→SEC-01/W03, D-02→AR-02/W05,
  D-03→AR-01/W05, D-04→DATA-08 W6/W04, D-05→REL-01/W06, D-06→SEC-04/W05, D-07→SEC-06/W03,
  D-08→FBL-06/W01, D-09→secrets docs/W01. Accordingly, only W04-E04-S001 carries a `decisions/`
  directory in this wave — no other W04 story enacts a D-0N decision.
- DATA-02 T5's worker-signature change is confirmed by the source to be a breaking change requiring
  wowsociety coordination ("worker signature change is breaking — coordinate with wowsociety even
  though it has zero current job usage," PLAN DATA-02 T5 risk column) — recorded as an unresolved
  coordination point in E01-S003's `plan.md`, not silently absorbed.

## Risks

See `risks.md`. Headline risks: RISK-W04-001 (E01-S001's supersession of W02-E01-S002's interim
checkpoint lease carries a migration-correctness risk on the receiving side, mirroring
RISK-W02-001's sending-side risk); DATA-08 W6-T1's breaking audit-hash format change is confirmed
"single highest-risk task in PF-DATA's Wave-6 scope" and hits wowsociety's live audit rows directly
(PROD-05 staging-drill coordination, product-level, tracked not blocking); DATA-02 T5's worker-
signature change is a confirmed-breaking coordination point with wowsociety.

## Quality gates

- DATA-02's fail-first evidence is the named chaos test itself (`DATA-02/chaos/
  duplicate_worker_lease_expiry_test.go`), constructible today and failing (duplicate effect
  observed) before the fix — per MATRIX CS-11's own framing. It is required as stated, not
  paraphrased into a generic "chaos test."
- DATA-03's fail-first evidence is its own named 6-boundary chaos test (`DATA-03/chaos/`), applied
  to both notify and webhook, per PLAN DATA-03 T8's acceptance criterion ("Zero duplicate external
  effects across all 6 fault points").
- DATA-04's fail-first evidence is the named chaos test `DATA-04/chaos/duplicate_worker_test.go`,
  matching "the Wave-3 exit gate wording verbatim" per PLAN's own acceptance criterion.
- DATA-08 W6-T1's fail-first evidence is a tamper test mutating `metadata` on a chained row, which
  "passes verification today (that's the defect), fails after" — per MATRIX CS-20's own framing;
  this exact tamper test, not a substitute, is required, applied independently to every declared
  field per PLAN's own acceptance criterion.
- DX-07's fail-first evidence is an integration test that boots against a stale-migrated DB and
  asserts a 503, per PLAN DX-07 T1's own test column.
- FBL-04's evidence is retry-schedule parity plus fault injection, per REVIEW §O's own task text —
  not a substitute test.

## Required artifacts

- DATA-02: the shared lease/fencing primitive (kernel building block); `jobs_queue` lease-column
  migration; fenced claim/finalize/reclaim code; the idempotency-declaration contract; the shared
  chaos harness.
- DATA-03: the three-stage protocol implementation for notify and webhook; the inbound two-phase
  verification code; the failed-signature audit path; the per-adapter idempotency-contract
  declaration mechanism.
- DATA-04: the corrected migration comment; the advisory-lock/CAS stopgap; the leased, `SKIP
  LOCKED`-honest claim SQL; the fencing/idempotency/retry/cancellation code; the lifecycle-control
  API.
- DATA-08 W6: the widened `chainHash` implementation and `hash_version` migration; the external
  anchor mechanism; the encrypted DSR export artifact writer; the central legal-hold wrapper; the
  explicit per-class status reporting.
- DX-07: the migration-currency readiness check; seed/rule/model-hash readiness reporting; the
  `config doctor` discovery fix.
- FBL-04: the `cenkalti/backoff/v5` integration replacing both hand-rolled retry call sites.

## Required evidence

- DATA-02: the named chaos test output (`duplicate_worker_lease_expiry_test.go`); claim/finalize/
  reclaim unit-test output; idempotency-declaration boot-time-rejection test output.
- DATA-03: three-stage protocol test output (no send-while-tx-open assertion); rotation-during-
  verification test output; failed-signature empty-body test output; the 6-boundary chaos test
  output.
- DATA-04: 2-processor concurrency test output; `EXPLAIN`-plan `SKIP LOCKED` assertion; the named
  chaos test output (`duplicate_worker_test.go`); lifecycle integration-test output.
- DATA-08 W6: per-field tamper test output (metadata, tx_id, and every other declared field
  independently); anchor-then-tamper detection test output; DSR export artifact-completion test
  output; central legal-hold negative test output; explicit-status test output.
- DX-07: stale-migration 503 integration-test output; full-readiness-payload integration-test
  output; nested-subdirectory and outside-repo `config doctor` discovery test output.
- FBL-04: retry-schedule parity test output; fault-injection test output.

## Expected implementation outcome

A framework whose jobs, remote deliveries, and bulk processors share one fencing primitive instead
of three inconsistent (or absent) ones, so a worker that stalls and is reclaimed can no longer
silently overwrite the reclaiming worker's outcome or double-fire an external effect; a framework
whose audit chain can no longer be tampered on `metadata` or `tx_id` without detection, with
historical rows still verifiable; a framework whose readiness endpoint tells the truth about
migration currency and seed/rule/model state instead of reporting healthy while missing an entire
documented check; and a framework with one retry implementation instead of two duplicated,
hand-rolled ones.

## Acceptance authority

Data/reliability lead — per PLAN §5.3's own "Accountable role: data/reliability lead" for PF-DATA,
applied to DATA-02/03/04/08; DX-07's readiness/diagnostics scope shares the same accountability
given its direct dependency-neighbor relationship to DATA-08's Wave-6 deployment-readiness closure
spec (MATRIX CS-21) within this wave.

## Closure conditions

All exit criteria satisfied; all four epics' `closure-report.md` accepted; `waves/index.md`'s W04
row updated to reflect `accepted` status; no unresolved regression from the lease-primitive
migration, the retry-library swap, or the audit-hash-widening breaking change; the RISK-W04-001
interim-lease-migration risk is resolved (E01-S001 has correctly migrated any state written under
W02-E01-S002's interim lease) before this wave can close; DATA-08 W6-T1's PROD-05 staging-drill
coordination note is recorded (not resolved — PROD-05 is product-level, outside this wave's
framework-side closure) with a clear pointer for the eventual product-side drill; DX-07 T4's
deferred-link to W05-E03-S002's AR-04 T5 waiver mechanism is recorded, not silently dropped.

## Status update (2026-07-16)

`status: in-progress` — corrected from the prior unsound `accepted` claim. Independent review
(`review-gate-2026-07-16.md`) accepted E01 (all 3 stories), E03 (both stories), and E04 (all 3
stories, including normalizing the non-vocabulary `closed-pending-review` token). E02 remains
`in-progress`: S001/S003 accepted, S002 genuinely `planned` (not started) — no false completion
claim survives. Wave cannot reach `accepted` until E02-S002 is either implemented or formally
excluded from this wave's acceptance scope.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
