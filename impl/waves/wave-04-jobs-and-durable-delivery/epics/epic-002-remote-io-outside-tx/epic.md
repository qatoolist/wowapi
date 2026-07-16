---
id: W04-E02
type: epic
title: Remote I/O outside transactions
status: in-progress
wave: W04
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-03
  - FBL-04
depends_on:
  - W04-E01
stories:
  - W04-E02-S001
  - W04-E02-S002
  - W04-E02-S003
decisions: []
risks:
  - RISK-W04-E02-001
  - RISK-W04-E02-002
---

# W04-E02 — Remote I/O outside transactions

## Epic objective

Move every remote provider/secret network call in `kernel/notify` and `kernel/webhook` outside an
open database transaction, replacing today's claim-and-call-inside-tx pattern with a three-stage
claim-tx → effect-outside-tx → fenced-finalize-tx protocol built on the shared lease/fencing
primitive this wave's W04-E01 delivers; close the inbound webhook-verification race between a
secret snapshot and its use by splitting verification into a two-phase read-verify-recheck
sequence; require every adapter to declare its own duplicate-safety mechanism before it can be
registered; and adopt `cenkalti/backoff/v5` in place of the framework's two duplicated hand-rolled
retry implementations (FBL-04). This epic converts the framework's own self-documented admission
that it does the wrong thing — `notify/service.go:456-586`'s comment at lines 446-449 already
states "Real production deployments should move the network call outside the tx" — into an
enforced architectural guarantee, and the named 6-boundary chaos test proves it holds under
worker failure, not merely under normal operation.

## Problem being solved

`requirement-inventory.md` row DATA-03 states: "Remote I/O outside DB transactions (T1–T8) | IMPL |
P0 | planned | W04-E02-S001..S002 | Scope refined by MATRIX CS-11: external effects only; T7 =
DATA-08 W0-T2 duplicate (done)." The source evidence is self-documented in wowapi's own code:
`notify/service.go:456-586`'s own comment (446-449) already states "Real production deployments
should move the network call outside the tx." `webhook/service.go`'s delivery loop and secret
resolution both run inside `plat.WithTenant(...)` — meaning a slow, hung, or failed remote call
(a notification provider, a webhook endpoint, or a secret store) currently holds a database
transaction open for as long as that remote call takes, and a transaction rollback/retry after a
partial remote effect can duplicate that effect (send the same notification or POST twice) with no
protection today. FBL-04 is a smaller, unrelated-but-adjacent problem riding in the same epic
because it touches the same remote-I/O retry paths this epic is already restructuring: "Replace
duplicated hand-rolled retry with `cenkalti/backoff/v5`. Tests: retry-schedule parity + fault
injection" (REVIEW §O), justified by REVIEW §K's reuse-opportunity register: "Retry/backoff
(hand-rolled ×2) | custom, duplicated | Replace → `cenkalti/backoff/v5` (already in module graph,
unused) | Duplication + a mature lib already transitively present."

## Scope

- Reusing W04-E01's shared lease/fencing primitive for notify/webhook claim rows, not a bespoke
  copy (S001, DATA-03 T1).
- The three-stage claim-tx → effect-outside-tx → fenced-finalize-tx protocol for `kernel/notify`,
  including deleting/updating the self-documented "should move outside tx" comment at
  `notify/service.go:456-586` (446-449) as part of the same change (S001, DATA-03 T2).
- The same three-stage protocol for `kernel/webhook.deliverToEndpoint`, with the current-row-state
  check moved into the claim stage so the effect stage needs no mid-flight DB reads (S001,
  DATA-03 T3).
- Inbound two-phase webhook-verification: a short read-tx endpoint snapshot that closes, then
  verification outside any tx, then a short write-tx re-check of version/status with discard+retry
  on mismatch (S002, DATA-03 T4).
- The failed-signature audit path: a body-free audit row written in its own short transaction,
  proving no raw body is ever persisted on a failed verification (S002, DATA-03 T5).
- The per-adapter idempotency-safety contract declaration, boot-time enforced so an adapter cannot
  be registered for a non-idempotent high-impact operation without declaring its duplicate-safety
  mechanism (S002, DATA-03 T6).
- The named 6-boundary chaos test — before send, during send, after success/before finalize, lease
  expiry, duplicate workers, provider timeout — applied to both notify and webhook, reusing (not
  reimplementing) the chaos harness W04-E01-S003 builds for DATA-02 T7 (S002, DATA-03 T8).
- `cenkalti/backoff/v5` adoption replacing both of the framework's hand-rolled retry
  implementations, with retry-schedule-parity and fault-injection tests (S003, FBL-04).

## Out of scope

- **DATA-03 T7** (removing the stale "app_platform lacks INSERT on events_outbox" comment and
  wiring legal-delivery audit) is explicitly **not** implemented in this epic. Per
  `requirement-inventory.md`'s DATA-08 row, this task shares scope with DATA-08 W0-T2, which is
  already "W0 slice EXECUTED (verified ×2)" — the legal-delivery audit write using migration
  00011's already-granted permission is done, at `DATA-08/wave0/legal-audit/`. S002's `story.md`
  carries a cross-reference to that evidence, not a re-implementation.
- **W04-E01's shared lease/fencing primitive itself** (generations, heartbeats, job-claim fencing,
  the primitive's own kernel package) — that is W04-E01-S001's scope. This epic reuses that
  primitive; it does not build or extend it.
- **W04-E01-S003's chaos-test harness** — this epic's T8 chaos test reuses that harness by
  cross-reference; it does not design or reimplement a new harness.
- **DATA-04's bulk multi-worker safety work** — W04-E03's scope, architecturally parallel to this
  epic (both consume W04-E01's primitive), not a dependency of this epic.
- **`kernel/webhook.HandleInbound`'s full downstream consumer migration** — this epic changes the
  transaction-ownership contract of `HandleInbound` itself (T4); it does not audit or migrate every
  caller of that function beyond what this epic's own compatibility-consideration note requires.

## Source requirements

DATA-03 (T1–T6, T8; T7 cross-referenced only, not implemented here), FBL-04.

## Architectural context

DATA-03 sits directly on top of W04-E01's shared lease/fencing primitive — it is not an independent
build. PLAN's own DATA-03 T1 acceptance criterion states "Lease columns via shared primitive, not a
bespoke copy," making the dependency on W04-E01-S001 a hard architectural constraint, not a
scheduling convenience: this epic's claim-tx stage for both notify and webhook must use the same
`lease_token`/`lease_generation`/`lease_expires_at` columns and claim/finalize semantics W04-E01
defines for jobs, so that a worker crash mid-delivery is fenced identically to a worker crash
mid-job. The three-stage protocol itself (claim-tx → effect-outside-tx → fenced-finalize-tx) is the
same shape MATRIX CS-11 names for the whole PF-DATA package: "the single shared lease/fencing
primitive (DATA-02 T1) first, then three-stage claim→effect-outside-tx→fenced-finalize for
notify/webhook." T4's inbound two-phase verification is architecturally distinct from T1–T3's
outbound three-stage protocol — it addresses a different race (a secret rotating or being
deactivated between when a webhook-verification snapshot is read and when it is used to verify an
inbound signature), not the outbound duplicate-effect race T1–T3 close. T4 also introduces a
breaking signature change to `HandleInbound`'s transaction-ownership contract, since the function
can no longer assume it owns a single enclosing transaction for its entire body. FBL-04 (S003) is
architecturally unrelated to T1–T6/T8's transaction-boundary work — it shares only the fact that
both touch remote-I/O call sites — and is grouped into this epic by `wave-allocation-detail.md`'s
canonical allocation ("S003 retry-adoption"), not by any task dependency.

## Included stories

- **W04-E02-S001 — notify-and-webhook-three-stage** (DATA-03 T1, T2, T3): reuse the shared lease
  primitive for notify/webhook claim rows; the three-stage protocol for `kernel/notify`; the same
  protocol for `kernel/webhook.deliverToEndpoint`.
- **W04-E02-S002 — inbound-two-phase-and-contracts** (DATA-03 T4, T5, T6, T8; T7 cross-reference
  only): inbound two-phase verification; failed-signature audit; per-adapter idempotency-contract
  declaration; the named 6-boundary chaos test.
- **W04-E02-S003 — retry-adoption** (FBL-04): `cenkalti/backoff/v5` adoption with retry-schedule
  parity and fault-injection tests.

## Dependencies

Depends on W04-E01 (specifically W04-E01-S001, the shared lease/fencing primitive, and
W04-E01-S003, the shared chaos harness) — see `dependencies.md` for the full statement. No
dependency on W02 or W03. Downstream: none within this wave depends on this epic by name; this
epic's own exit is one of W04's wave-level exit-criteria bullets (DATA-03 T1–T6, T8 satisfied).

## Risks

RISK-W04-E02-001 (T4's breaking signature change to `HandleInbound`'s transaction-ownership
contract) and RISK-W04-E02-002 (this epic's hard dependency on W04-E01-S001/S003 landing first,
since neither the lease columns nor the chaos harness can be built independently) — see `risks.md`
for full detail and mitigation/contingency.

## Required decisions

None. A scan of `requirement-inventory.md` §B finds no D-0N row targeting DATA-03 or FBL-04 — only
DATA-08 W6 (D-04, W04-E04-S001's scope) enacts a decision in this wave, per `wave.md`
"Assumptions." This epic's stories accordingly carry no `decisions/` directory.

## Epic acceptance criteria

- **AC-W04-E02-01**: `kernel/notify` and `kernel/webhook.deliverToEndpoint` both reuse W04-E01's
  shared lease/fencing primitive for their claim rows (not a bespoke copy), and no `sender.Send`
  (notify) or DNS/secret-resolve/POST call (webhook) executes while a database transaction is open,
  proven by the T8 boundary-matrix tests. The self-documented "should move outside tx" comment at
  `notify/service.go:456-586` (446-449) is deleted or updated to reflect the new protocol.
- **AC-W04-E02-02**: Inbound webhook verification is immune to a secret rotation/deactivation race
  between snapshot and verification, proven by a rotation-during-verification test; failed-signature
  audit rows never persist a raw body, proven by a test asserting an empty body field; no adapter
  can be registered for a non-idempotent high-impact operation without declaring its duplicate-safety
  mechanism, proven by a boot-time fixture test that rejects an undeclared adapter.
- **AC-W04-E02-03**: The named 6-boundary chaos test (before send, during send, after
  success/before finalize, lease expiry, duplicate workers, provider timeout) passes for both
  notify and webhook with zero duplicate external effects across all 6 fault points, reusing
  W04-E01-S003's chaos harness by cross-reference, not reimplementing it.
- **AC-W04-E02-04**: `cenkalti/backoff/v5` replaces both of the framework's hand-rolled retry
  implementations; retry-schedule parity is proven against the prior hand-rolled schedules, and
  fault-injection tests confirm correct retry/backoff behavior under induced failure.
- **AC-W04-E02-05**: All three stories have passed independent review (S001, S002 per mandate §14;
  S003 per a lighter, documented review approach appropriate to its P1/low-risk profile), with S002
  specifically checked for T4's breaking-change note being recorded as a compatibility
  consideration (not silently absorbed) and for T7 being correctly treated as cross-reference-only,
  not re-implemented.

## Closure conditions

All three stories reach `accepted` (each satisfying its own `closure.md`); AC-W04-E02-01 through
AC-W04-E02-05 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; RISK-W04-E02-001 (T4's breaking-change note) is recorded as an
explicit compatibility consideration at closure, not silently dropped; RISK-W04-E02-002 (the
W04-E01 dependency) is confirmed resolved (W04-E01-S001 and W04-E01-S003 both accepted) before this
epic itself can close.

## Status update (2026-07-16)

`status: in-progress` — S001 accepted (C-1 out-of-tx defect remediated and independently
re-verified) and S003 accepted; S002 (inbound two-phase verification) remains genuinely `planned`
— no chaos-test directory exists for notify/webhook and `HandleInbound` still runs entirely inside
the caller's open transaction, by design, per `review-gate-2026-07-16.md`. Epic cannot reach
`accepted` until S002 is implemented or formally excluded from this epic's acceptance scope.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
