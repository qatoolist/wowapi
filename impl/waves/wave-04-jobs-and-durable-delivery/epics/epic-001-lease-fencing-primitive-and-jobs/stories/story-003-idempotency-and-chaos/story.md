---
id: W04-E01-S003
type: story
title: Worker idempotency contract and the shared duplicate-worker chaos harness
status: accepted
wave: W04
epic: W04-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-02
depends_on:
  - W04-E01-S002
blocks: []
acceptance_criteria:
  - AC-W04-E01-S003-01
  - AC-W04-E01-S003-02
  - AC-W04-E01-S003-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W04-003
---

# W04-E01-S003 — Worker idempotency contract and the shared duplicate-worker chaos harness

## Story ID

W04-E01-S003

## Title

Worker idempotency contract and the shared duplicate-worker chaos harness

## Objective

Establish a stable job idempotency key and lease context passed to workers, requiring every worker
to declare exactly one duplicate-safety mechanism; prove that fencing the queue row alone does not
undo an already-committed stale-worker domain transaction; and build the named chaos test
`DATA-02/chaos/duplicate_worker_lease_expiry_test.go` as a **reusable chaos harness explicitly
shared with W04-E02 (DATA-03) and W04-E03 (DATA-04)** — not reimplemented by either.

## Value to the framework

This story completes the epic's end-to-end proof: S001 built the primitive, S002 applied it to
`jobs_queue`'s claim/finalize/reclaim paths, and this story is where the whole chain gets exercised
under the exact adversarial scenario the epic exists to prevent — "pause worker A after claim,
expire, reclaim via B, B completes, resume A and attempt finalize at every domain/external/finalize
boundary — exactly one logical effect recorded, A's writes rejected" (PLAN DATA-02 T7's own row,
verbatim). This is also this story's second load-bearing responsibility, stated explicitly in the
epic's own scope: T7's chaos harness is built "as a reusable chaos harness shared with DATA-03/
DATA-04" (PLAN DATA-02 T7's own risk column) — `wave-allocation-detail.md` confirms this exactly
("S003 idempotency-and-chaos (T5, T6, T7 chaos harness — harness shared with E02/E03)"). **Any
worker building W04-E02's 6-boundary chaos test or W04-E03's chaos test must consume this story's
harness, not reimplement it** — this is the explicit design intent this story's own task record must
make visible to those parallel-building teams.

## Problem statement

PLAN DATA-02's task table gives the exact acceptance bar this story must satisfy:

- T5: "Stable job idempotency key + lease context passed to workers; each worker declares exactly
  one of: inbox/effect ledger unique on `(job_id, effect_name)`, domain CAS, or provider idempotency
  key | T2 | Worker cannot register without declaring its mechanism | Duplicate-effect test |
  `DATA-02/idempotency/` | Likely needs PF-ARCH's typed operation model to enforce at compile time;
  worker signature change is breaking — coordinate with wowsociety even though it has zero current
  job usage."
- T6: "Document/test: fencing the queue row does not undo an already-committed stale-worker domain
  transaction | T3, T5 | Testable claim, not prose | Test proving effect ledger still catches an
  idempotency-ignoring worker | `DATA-02/worker-contract/` | Low."
- T7: "**Named chaos test:** pause worker A after claim, expire, reclaim via B, B completes, resume
  A and attempt finalize at every domain/external/finalize boundary — exactly one logical effect
  recorded, A's writes rejected | T3-T5 | Matches closure contract verbatim |
  `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` | Must exercise all 3 named boundaries; build
  as a reusable chaos harness shared with DATA-03/DATA-04."

MATRIX CS-11's own evidence refinement is the honest framing this story must preserve, not
re-litigate: "the current at-least-once posture is explicitly documented as an accepted
idempotent-worker tradeoff (`kernel/jobs/runner.go:437-438,108-113`); DATA-03's exposure is scoped
to external side effects only — DB effects are already exactly-once via the outbox inbox-dedup
(`kernel/outbox/relay.go:191,205-219`)... The fix contract stands; the honest framing is 'make the
documented assumption enforceable' not 'fix an unacknowledged race.'" No idempotency-declaration
contract, no effect-ledger-vs-fencing test, and no chaos test exist anywhere in the repository
today.

## Source requirements

DATA-02 (T5, T6, T7).

## Current-state assessment

Per PLAN's own DATA-02 evidence (to be re-confirmed at this story's own execution commit): no
worker idempotency-declaration mechanism exists — a worker may register with `kernel/jobs` without
declaring any duplicate-safety strategy. No test exists proving the effect ledger (not the queue
row) is the actual source of truth for whether an effect already happened. No chaos test exists at
all. This story's own re-confirmation step (per this programme's fail-first convention, e.g.
W02-E01-S001) is to read `kernel/jobs`'s worker-registration mechanism and `kernel/outbox/relay.go`'s
inbox-dedup logic at this story's actual start commit and confirm these facts still hold before
building the contract and the chaos test.

## Desired state

Every job worker declares exactly one duplicate-safety mechanism (inbox/effect ledger unique on
`(job_id, effect_name)`, domain CAS, or provider idempotency key) at registration time, and cannot
register without doing so. A test proves that fencing the `jobs_queue` row alone — S002's own
mechanism — does not by itself undo an already-committed stale-worker domain transaction; the effect
ledger is what catches an idempotency-ignoring worker, not queue-row fencing. The named chaos test
`DATA-02/chaos/duplicate_worker_lease_expiry_test.go` exists, exercises all three named boundaries
(domain, external, finalize), proves exactly one logical effect is recorded and worker A's writes
are rejected, and is built as a reusable harness that W04-E02 and W04-E03 explicitly consume for
their own chaos tests rather than reimplementing.

## Scope

- A stable job idempotency key and lease context passed to every worker at invocation.
- A worker-registration-time contract requiring exactly one declared duplicate-safety mechanism
  (inbox/effect ledger, domain CAS, or provider idempotency key); registration fails without a
  declaration.
- A duplicate-effect test proving the registration contract is enforced (a worker without a
  declaration cannot register).
- A test proving fencing the queue row alone does not undo an already-committed stale-worker domain
  transaction — the effect ledger genuinely catches an idempotency-ignoring worker.
- The named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`, exercising all three
  named boundaries (domain, external, finalize) exactly as PLAN T7 states, built as a reusable
  harness.
- Documenting and structuring the harness so W04-E02's 6-boundary chaos test and W04-E03's chaos
  test can consume it directly, without reimplementing the pause/expire/reclaim/resume mechanics.
- Recording DATA-02 T5's worker-signature-change breaking-change coordination note explicitly, per
  RISK-W04-003, not silently resolving or hiding it.

## Out of scope

- **Actually building W04-E02's 6-boundary chaos test or W04-E03's chaos test** — those epics'/
  stories' own scope. This story delivers the reusable harness; it does not consume it on their
  behalf.
- **PF-ARCH's typed operation model** (referenced by T5's own risk note as a possible compile-time
  enforcement mechanism) — out of this story's scope unless it already exists in the repository;
  this story's plan records whether registration-time enforcement is compile-time or runtime as an
  implementation-time decision, not a dependency on unbuilt PF-ARCH tooling.
- **Actually announcing or shipping the T5 worker-signature-change migration guidance to
  wowsociety** — recorded as an open coordination note (RISK-W04-003) per `story.md` and `plan.md`;
  not resolved or acted upon by this story.
- **The shared primitive's own design or the `jobs_queue` fencing mechanics** — W04-E01-S001's and
  W04-E01-S002's scope respectively; this story consumes both as already-built prerequisites.

## Assumptions

- PF-ARCH's typed operation model (T5's own risk note: "Likely needs PF-ARCH's typed operation
  model to enforce at compile time") is not confirmed to exist in the repository as of this story's
  planning — if it does not exist, registration-time enforcement is implemented as a runtime check
  instead, recorded as an implementation-time decision in `plan.md`, not silently assumed to be
  compile-time.
- The exact idempotency key's derivation (job ID alone, job ID + attempt number, or another scheme)
  is not specified by the source beyond "stable job idempotency key" — recorded as an
  implementation-time decision.
- The chaos harness's exact mechanism for "pausing" worker A after claim (a test hook, a
  synchronization primitive, or a controlled delay) is not specified by the source beyond the
  scenario description itself — this story's plan records the chosen mechanism, not invents it here.

## Dependencies

Depends on W04-E01-S002 (T5's lease context passed to workers, T6's fencing-vs-effect-ledger test,
and T7's chaos test all operate on S002's fenced claim/finalize/reclaim chain). Does not block any
further story within this epic (this is the epic's final story) but is depended upon by W04-E02
(DATA-03's 6-boundary chaos test) and W04-E03 (DATA-04's chaos test) for the shared harness, per
`../../dependencies.md`.

## Affected packages or components

`kernel/jobs` — worker-registration mechanism (idempotency-declaration contract), the idempotency
key/lease-context invocation path. New: the chaos harness itself, expected at
`DATA-02/chaos/duplicate_worker_lease_expiry_test.go` per PLAN T7's own required artifact path
(exact repository-relative location TBD at implementation time — the source's own path notation
mirrors this programme's PLAN-ID/task-topic convention, not necessarily a literal existing directory
today).

## Compatibility considerations

T5's worker-signature change is confirmed breaking by the source itself ("worker signature change is
breaking — coordinate with wowsociety even though it has zero current job usage," PLAN DATA-02 T5
risk column). Per PLAN's own wowsociety-impact note, this is "not affected" today (zero
`kernel/jobs` import, zero job registration in wowsociety) but "would become breaking... the moment
wowsociety registers a job." This story's plan must record this as an explicit, unresolved
coordination note — not resolve it, not hide it, not silently absorb it as if it were an additive
change.

## Security considerations

The idempotency-declaration contract is itself a security/correctness control: a worker able to
register without declaring a duplicate-safety mechanism is exactly the gap that makes fencing alone
insufficient (per T6's own concern) — the queue row can be perfectly fenced while a worker still
double-fires an external effect if it has no idempotency mechanism of its own.

## Performance considerations

None separately mandated by the source. The chaos harness itself is a test-infrastructure concern,
not a runtime performance concern.

## Observability considerations

None separately mandated by the source beyond what S002 already requires for finalize-rejection
observability; T7's chaos test itself is the primary observability proof mechanism for this story's
scope (it must be able to observe "exactly one logical effect recorded" and "A's writes rejected" at
each boundary).

## Migration considerations

None — this story does not introduce a schema or data migration of its own. T5's worker-signature
change is a code-contract change, not a data migration.

## Documentation requirements

Document the idempotency-declaration contract (the three allowed mechanisms, registration-time
enforcement), the effect-ledger-vs-fencing distinction (T6's own testable claim), and the chaos
harness's structure and reuse contract for W04-E02/W04-E03 to consume. Document the T5
breaking-change coordination note explicitly, including its current non-blocking status
(wowsociety's zero current usage) and its future-blocking trigger (wowsociety's first job
registration).

## Acceptance criteria

- **AC-W04-E01-S003-01**: Every job worker must declare exactly one of: inbox/effect ledger unique
  on `(job_id, effect_name)`, domain CAS, or provider idempotency key at registration time; a worker
  attempting to register without a declaration is rejected — proven by a duplicate-effect test.
- **AC-W04-E01-S003-02**: A test proves that fencing the `jobs_queue` row alone does not undo an
  already-committed stale-worker domain transaction — the effect ledger, not queue-row fencing,
  catches an idempotency-ignoring worker. This is a testable claim per PLAN T6's own wording, not
  prose documentation alone.
- **AC-W04-E01-S003-03**: The named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`
  exists and exercises all three named boundaries (domain, external, finalize): pausing worker A
  after claim, expiring its lease, reclaiming via worker B, B completing, resuming A, and A
  attempting to finalize at each boundary — exactly one logical effect is recorded and A's writes
  are rejected at every boundary. The harness is built reusably, with a documented reuse contract,
  for W04-E02's and W04-E03's own chaos tests to consume without reimplementation.

## Required artifacts

- The idempotency-declaration contract (registration-time enforcement code).
- The effect-ledger-vs-fencing test.
- The named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`, structured as a
  reusable harness.
- Documentation of the contract, the fencing/effect-ledger distinction, the harness reuse contract,
  and the T5 coordination note.
See `artifacts/index.md`.

## Required evidence

- Duplicate-effect / registration-rejection test output.
- Effect-ledger-catches-idempotency-ignoring-worker test output.
- The named chaos test's own passing-run output, covering all three named boundaries.
- A consolidated evidence bundle aggregating the above (see `tasks/index.md` "Grouping rationale").
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W04-E01-S002
recorded, owner/reviewer assignment pending, unresolved questions (PF-ARCH typed-operation-model
availability, idempotency-key derivation, chaos-harness pause mechanism) and the T5 breaking-change
coordination note explicitly recorded rather than silently assumed or resolved.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the chaos test genuinely exercises all three named
boundaries (not a subset) and that the T5 breaking-change coordination note is recorded honestly as
an open item, not silently resolved with an invented migration decision.

## Risks

RISK-W04-003 (T5's worker-signature change is confirmed breaking, wowsociety coordination required)
— see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once the coordination note (RISK-W04-003) is recorded as planned and the chaos test genuinely proves
all three named boundaries, residual risk for this story's own scope is expected to be low — the
remaining risk (RISK-W04-003 itself) is explicitly accepted as open until wowsociety's own roadmap
intersects job registration, not something this story's own closure can eliminate.

## Plan

See `plan.md`.
