---
id: W04-E02-S002
type: story
title: Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos test
status: planned
wave: W04
epic: W04-E02
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-03
depends_on:
  - W04-E02-S001
  - W04-E01-S003
blocks: []
acceptance_criteria:
  - AC-W04-E02-S002-01
  - AC-W04-E02-S002-02
  - AC-W04-E02-S002-03
  - AC-W04-E02-S002-04
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W04-E02-S002-001
---

# W04-E02-S002 — Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos test

## Story ID

W04-E02-S002

## Title

Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos test

## Objective

Implement inbound webhook verification as a two-phase protocol immune to a secret rotation/
deactivation race between snapshot and verification; write a body-free audit row for every failed
signature verification, in its own short transaction; require every adapter to declare its
duplicate-safety mechanism before it can be registered; and prove, with the named 6-boundary chaos
test (reusing W04-E01-S003's shared chaos harness), that zero duplicate external effects occur
across all 6 fault points for both notify and webhook.

## Value to the framework

T4's two-phase verification closes a race this programme has not yet had to close anywhere else:
a webhook endpoint's secret can be rotated or deactivated in the window between when
`HandleInbound` reads a snapshot of the endpoint's verification state and when it uses that
snapshot to verify an inbound signature — without this story's fix, an attacker (or a legitimate
but oddly-timed operational rotation) could cause a signature to be accepted under a since-revoked
policy. T8's named 6-boundary chaos test is this epic's own proof obligation for everything W04-E02
built in S001 and this story: it is, per the wave-level `wave.md`, "the most labor-intensive
requirement in PF-DATA," and it is the test that actually demonstrates the three-stage protocol and
the fencing primitive hold under real worker failure, not merely under normal operation.

## Problem statement

`requirement-inventory.md` row DATA-03 groups this story's tasks with S001 under one W04-E02-S001..
S002 target. This story's source task rows, verbatim:

- **T4**: "Inbound two-phase verification: short read-tx (endpoint snapshot) closes → verify
  outside tx → short write-tx re-checks version/status, discard+retry on mismatch | Depends on: T1
  | Acceptance: Secret rotation/deactivation between phases cannot cause accept-under-stale-policy |
  Tests: Rotation-during-verification test | Evidence: `DATA-03/webhook/inbound-two-phase/` | Risk:
  Breaking signature change to `HandleInbound`'s transaction-ownership contract; bound retry
  attempts."
- **T5**: "Failed-signature audit: body-free audit row in its own short tx | Depends on: T4 |
  Acceptance: No raw body ever persisted on failed verification | Tests: Test asserting empty body
  field | Evidence: `DATA-03/webhook/failed-sig-audit/` | Risk: Low."
- **T6**: "Per-adapter idempotency-safety contract declaration | Depends on: T2, T3 | Acceptance:
  Adapter cannot be registered for a non-idempotent high-impact operation without declaring
  duplicate-safety | Tests: Boot-time fixture rejecting undeclared adapter | Evidence:
  `DATA-03/adapter-contract/` | Risk: Inventory all existing `Sender` implementations first."
- **T7**: "Remove the stale 'app_platform lacks INSERT on events_outbox' comment; wire
  legal-delivery audit — shares scope with DATA-08 W0-T2, implement once, cross-reference | Depends
  on: — | Acceptance: — | Tests: — | Evidence: Cross-reference only | Risk: Avoid
  double-implementation."
- **T8**: "Named chaos test at 6 boundaries: before send, during send, after success/before
  finalize, lease expiry, duplicate workers, provider timeout — applied to both notify and webhook |
  Depends on: T2-T4 | Acceptance: Zero duplicate external effects across all 6 fault points | Tests:
  This is the test | Evidence: `DATA-03/chaos/` | Risk: Most labor-intensive requirement in
  PF-DATA; reuse DATA-02's chaos harness."

**T7 is already executed and is not implemented in this story.** Per `requirement-inventory.md`'s
DATA-08 row, T7's scope — removing the stale deferral comment and implementing the legal-delivery
audit write using migration 00011's already-granted permission — is identical to DATA-08's own
W0-T2: "W0-T2. Remove the stale deferral comment; implement legal-delivery audit write using
migration 00011's already-granted permission | — | `ImportanceLegal` deliveries produce a durable
audit/outbox record with provider receipt in the same transaction as the `sent` status update |
Test: legal-importance send → audit row with provider msg ID; negative test for non-legal |
`DATA-08/wave0/legal-audit/`" — status EXECUTED, verified ×2, per `requirement-inventory.md`. This
story's own scope explicitly excludes T7; see "Out of scope" below.

## Source requirements

DATA-03 (T4, T5, T6, T8; T7 cross-referenced only against DATA-08 W0-T2, not implemented here).

## Current-state assessment

Per the source evidence: no two-phase inbound-verification protocol exists today —
`HandleInbound`'s current transaction-ownership contract is assumed (by S001's own current-state
assessment of the wider DATA-03 defect) to hold a single transaction across snapshot-read and
verification, the same class of defect T1–T3 close for outbound delivery. No failed-signature audit
path with a body-free guarantee exists today. No per-adapter idempotency-safety contract-declaration
mechanism exists today — PLAN T6's own risk note ("Inventory all existing `Sender` implementations
first") confirms this is new enforcement, not an extension of an existing declaration mechanism. No
named 6-boundary chaos test exists today for notify/webhook — it depends on this epic's own S001
output existing first. This story's own re-confirmation step is to re-read `HandleInbound`'s current
transaction structure and the current (absent) audit/adapter-contract mechanisms at this story's
actual start commit.

## Desired state

`HandleInbound` reads a short read-tx snapshot of the endpoint's verification state (secret,
version, status), closes that transaction, verifies the inbound signature entirely outside any open
transaction, then opens a short write-tx that re-checks the endpoint's version/status against the
snapshot — discarding and retrying the verification if a mismatch is found (indicating a rotation or
deactivation occurred in the window). Every failed signature verification writes a body-free audit
row in its own short transaction. Every `Sender` (and other high-impact adapter) implementation
declares its duplicate-safety mechanism at registration time, and the framework's boot sequence
rejects any adapter registered for a non-idempotent high-impact operation without such a
declaration. The named 6-boundary chaos test — before send, during send, after success/before
finalize, lease expiry, duplicate workers, provider timeout — passes for both notify and webhook
with zero duplicate external effects observed, reusing W04-E01-S003's shared chaos harness.

## Scope

- Inbound two-phase verification for `kernel/webhook.HandleInbound`: short read-tx snapshot → verify
  outside tx → short write-tx re-check with discard+retry on mismatch (T4).
- The failed-signature audit path: a body-free audit row written in its own short transaction (T5).
- The per-adapter idempotency-safety contract declaration mechanism, boot-time enforced, including
  an inventory of all existing `Sender` implementations before enforcement is turned on (T6).
- The named 6-boundary chaos test, applied to both notify and webhook, reusing (not reimplementing)
  W04-E01-S003's shared chaos harness (T8).
- A cross-reference note (not an implementation) confirming T7 is already executed under DATA-08
  W0-T2.

## Out of scope

- **DATA-03 T7 itself.** T7's scope (removing the stale "app_platform lacks INSERT on
  events_outbox" comment and wiring legal-delivery audit) is not implemented in this story or
  anywhere in this epic. It is already executed and verified twice under DATA-08 W0-T2, evidenced at
  `DATA-08/wave0/legal-audit/`. This story records the cross-reference only, per PLAN DATA-03 T7's
  own risk note: "Avoid double-implementation."
- **W04-E02-S001's outbound three-stage protocol** (T1, T2, T3) — that story's own scope. This
  story's T8 chaos test exercises S001's protocol, but does not re-implement it.
- **W04-E01-S003's chaos harness itself** — this story's T8 reuses that harness by dependency and
  cross-reference; it does not design or build a new one.
- **Any adapter's actual idempotency-safety implementation redesign** — T6's contract-declaration
  mechanism requires adapters to *declare* their duplicate-safety mechanism; it does not itself
  redesign any adapter's internal duplicate-safety logic if an inventoried adapter is found to have
  none. An adapter found non-idempotent-and-undeclared during T6's inventory is a finding to record
  (in `deviations.md` or a follow-up item), not silently fixed as an unplanned scope expansion of
  this story.

## Assumptions

- T4's dependency on T1 (per its own dependency column) is satisfied by W04-E02-S001-T001's
  lease-column migration and shared-primitive reuse, not by T2/T3's three-stage protocol
  specifically — T4 addresses a different race (inbound verification timing) than T2/T3's outbound
  duplicate-effect race, and its dependency on T1 is about sharing the same claim-row
  infrastructure where relevant, not a functional dependency on the outbound protocol.
- T6's dependency on T2 and T3 (per its own dependency column) is confirmed, not assumed — the
  contract-declaration mechanism is validated against the concrete `Sender`/webhook-delivery
  adapters T2/T3 (W04-E02-S001) produce, so T6 cannot be meaningfully tested until those adapters
  exist in their post-three-stage-protocol form.
- T8's dependency on "T2-T4" (per its own dependency column) is confirmed — the chaos test exercises
  the three-stage protocol (T2, T3, this epic's S001) together with T4's two-phase verification, so
  T8 cannot start until all three are implemented, in addition to its dependency on W04-E01-S003's
  harness.
- The exact retry-attempt bound for T4's discard+retry-on-mismatch loop ("bound retry attempts," per
  T4's own risk note) is not specified numerically by the source — this story's plan records the
  exact figure as an implementation-time decision, per mandate §18, mirroring
  W02-E01-S001-T002's bounded-retry pattern for lock-timeout enforcement.

## Dependencies

Depends on **W04-E02-S001** (T4's own dependency on T1; T6's dependency on T2, T3; T8's dependency
on T2-T4) and on **W04-E01-S003** (the shared chaos harness T8 reuses). No dependency within this
epic beyond S001. Does not block any other story in this epic (S003/FBL-04 has no dependency on this
story).

## Affected packages or components

`kernel/webhook` (specifically `HandleInbound`'s transaction structure, a new failed-signature audit
path, and its endpoint-secret snapshot/verification logic); `kernel/notify` and `kernel/webhook`'s
`Sender`/delivery-adapter registration mechanism (the new idempotency-safety contract-declaration
enforcement); a new or extended chaos-test suite reusing W04-E01-S003's harness, applied against both
`kernel/notify` and `kernel/webhook`.

## Compatibility considerations

T4 is a **confirmed breaking change** to `HandleInbound`'s transaction-ownership contract — per PLAN
DATA-03 T4's own risk column: "Breaking signature change to `HandleInbound`'s transaction-ownership
contract." This is recorded here as an explicit compatibility consideration, not silently absorbed:
any caller of `HandleInbound` that assumed the prior single-enclosing-transaction contract must be
identified and updated. Per `wave.md`'s wowsociety-impact note for DATA-03: "Not affected today;
conditionally breaking in the future... If wowsociety ever calls `webhook.HandleInbound` directly,
T4's transaction-ownership contract change would need integration review — flag for future, not
now." This story's `plan.md` must record the breaking-change note explicitly, not fold it silently
into "implementation detail."

## Security considerations

T4's entire purpose is a security fix: closing the accept-under-stale-policy race between a secret
rotation/deactivation and its use in verification. T5's body-free audit guarantee is itself a
security/compliance control (no raw payload persisted on a failed, potentially-malicious,
verification attempt). T6's boot-time enforcement is a security-relevant control preventing
silent registration of an adapter that could duplicate a high-impact external effect.

## Performance considerations

T4's two-phase split trades a single transaction for two short transactions plus an out-of-tx
verification step — expected to be performance-neutral or better (shorter individual lock hold
times), to be confirmed by this story's own testing, not assumed without evidence.

## Observability considerations

T4's discard+retry-on-mismatch events should be observable (logged) so an operator can distinguish
"verification succeeded on retry N due to a genuine rotation race" from "verification is retrying
indefinitely" — mirroring W02-E01-S001-T002's bounded-retry observability pattern. T6's boot-time
rejection of an undeclared adapter must produce a clear, actionable error identifying which adapter
and which operation triggered the rejection.

## Migration considerations

No schema migration is anticipated for T4/T5/T6 beyond whatever failed-signature-audit table T5
requires (exact schema TBD at implementation time) — this story does not touch application data
migration beyond that new audit table.

## Documentation requirements

Document the two-phase inbound-verification protocol and its discard+retry-on-mismatch behavior;
document the failed-signature audit's body-free guarantee; document the per-adapter
idempotency-safety contract declaration requirement and how to satisfy it; document the 6-boundary
chaos test's named boundaries and what each proves; record the T7 cross-reference explicitly in this
story's own documentation, not merely in `story.md`'s prose.

## Acceptance criteria

- **AC-W04-E02-S002-01**: Inbound webhook verification's two-phase protocol (short read-tx snapshot
  → verify outside tx → short write-tx re-check, discard+retry on mismatch) prevents
  accept-under-stale-policy when a secret rotation or deactivation occurs between the snapshot and
  verification phases — proven by a dedicated rotation-during-verification test. Retry attempts are
  explicitly bounded, not unbounded.
- **AC-W04-E02-S002-02**: Every failed signature verification produces a body-free audit row written
  in its own short transaction — proven by a test asserting the audit row's body field is empty for
  every failed-verification case exercised.
- **AC-W04-E02-S002-03**: No adapter can be registered for a non-idempotent high-impact operation
  without declaring its duplicate-safety mechanism — proven by a boot-time fixture test that rejects
  an undeclared adapter; all existing `Sender` implementations are inventoried and correctly
  declared before enforcement is enabled.
- **AC-W04-E02-S002-04**: The named 6-boundary chaos test — before send, during send, after
  success/before finalize, lease expiry, duplicate workers, provider timeout — passes for both
  `kernel/notify` and `kernel/webhook` with zero duplicate external effects observed across all 6
  fault points, reusing W04-E01-S003's chaos harness (`DATA-02/chaos/
  duplicate_worker_lease_expiry_test.go`'s underlying harness) by direct cross-reference, not
  reimplementation.

## Required artifacts

- The inbound two-phase verification implementation for `HandleInbound`.
- The failed-signature audit path implementation.
- The per-adapter idempotency-safety contract declaration mechanism, plus the `Sender`
  implementation inventory.
- The 6-boundary chaos-test suite for notify and webhook.
- The T7 cross-reference record.
See `artifacts/index.md`.

## Required evidence

- Rotation-during-verification test output (`DATA-03/webhook/inbound-two-phase/`).
- Empty-body-field test output (`DATA-03/webhook/failed-sig-audit/`).
- Boot-time undeclared-adapter-rejection fixture test output (`DATA-03/adapter-contract/`).
- The 6-boundary chaos test output for both notify and webhook (`DATA-03/chaos/`).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies on W04-E02-S001
and W04-E01-S003 recorded and their status tracked, owner/reviewer assignment pending, the T4
breaking-change note and the T7 cross-reference both explicitly recorded rather than silently
assumed or dropped.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T4's breaking-change note is recorded as a
compatibility consideration (not silently absorbed) and T7 is correctly treated as cross-reference-
only (not re-implemented).

## Risks

RISK-W04-E02-S002-001 (T4's breaking signature change to `HandleInbound`'s transaction-ownership
contract) — see epic-level `risks.md` (RISK-W04-E02-001) for full detail and mitigation/contingency;
this story-level risk entry narrows that epic-level risk to this story's own T4 scope.

## Residual-risk expectations

Once T4's compatibility-consideration note is recorded and in-repo `HandleInbound` callers are
confirmed updated, residual risk is expected to be low for the compatibility dimension. T8's chaos
test, once passing, is expected to bring the duplicate-effect risk this whole epic addresses to low
residual risk for both notify and webhook.

## Plan

See `plan.md`.

## Correction note (autopsy remediation R-1, 2026-07-16)

`status: accepted` was false. The implementation-autopsy report
(`impl/reports/implementation-autopsy-report-2026-07-16.md`, finding **C-2**) found this story's
own `closure.md`, `verification.md`, and evidence records still stating "not implemented,
verified, or closed" — no two-phase inbound verification, no chaos test for notify/webhook, and
inbound verification runs in a single transaction, not the claimed two-phase protocol — while
`story.md` front matter claimed `accepted`. Per the programme's own Definition of Done, a story
must not be accepted without verification and reviewer acceptance (mandate §7); none occurred
here. Status reverted to `planned`, the honest status-model value matching the story's actual,
unstarted state. — autopsy remediation R-1, 2026-07-16.
