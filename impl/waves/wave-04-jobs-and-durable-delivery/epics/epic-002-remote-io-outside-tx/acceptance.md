---
id: W04-E02-ACCEPTANCE
type: epic-acceptance
epic: W04-E02
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../risks.md`/
`../../wave.md` pattern (this epic's ACs roll up into W04's own exit criteria for DATA-03/FBL-04).

## AC-W04-E02-01 — Three-stage protocol removes remote I/O from open transactions

`kernel/notify` and `kernel/webhook.deliverToEndpoint` both reuse W04-E01's shared lease/fencing
primitive for their claim rows (not a bespoke copy). No `sender.Send` call (notify) and no
DNS/secret-resolve/POST call (webhook) executes while a database transaction is open, proven by the
T8 boundary-matrix tests covering the claim-tx → effect-outside-tx → fenced-finalize-tx protocol.
The self-documented "should move outside tx" comment at `notify/service.go:456-586` (446-449) is
deleted or updated to reflect the new protocol as part of the same change. Traces to W04-E02-S001.

## AC-W04-E02-02 — Inbound verification, failed-signature audit, and adapter contracts proven

Inbound webhook verification's two-phase protocol (short read-tx snapshot → verify outside tx →
short write-tx re-check with discard+retry on mismatch) is immune to a secret rotation/deactivation
race between snapshot and verification, proven by a dedicated rotation-during-verification test.
Failed-signature audit rows are written in their own short transaction and never persist a raw
body, proven by a test asserting an empty body field. No adapter can be registered for a
non-idempotent high-impact operation without declaring its duplicate-safety mechanism, proven by a
boot-time fixture test rejecting an undeclared adapter. Traces to W04-E02-S002.

## AC-W04-E02-03 — Named 6-boundary chaos test passes with zero duplicate effects

The named chaos test — before send, during send, after success/before finalize, lease expiry,
duplicate workers, provider timeout — passes for both `kernel/notify` and `kernel/webhook` with
zero duplicate external effects observed across all 6 fault points. The test reuses
W04-E01-S003's chaos harness (built for DATA-02 T7, `DATA-02/chaos/
duplicate_worker_lease_expiry_test.go`) by direct cross-reference and dependency; it is not a
reimplementation of a new harness. Traces to W04-E02-S002.

## AC-W04-E02-04 — Retry-library adoption with parity and fault-injection proof

`cenkalti/backoff/v5` replaces both of the framework's hand-rolled retry implementations
(REVIEW §K's reuse-opportunity register: "Retry/backoff (hand-rolled ×2)"). A retry-schedule-parity
test proves the new library's schedule matches or improves on each prior hand-rolled schedule's
behavior; a fault-injection test proves correct retry/backoff behavior under induced remote-call
failure. Traces to W04-E02-S003.

## AC-W04-E02-05 — Independent review passed

S001 and S002 (both P0, per DATA-03's own priority) have passed independent review per mandate §14.
S002's review specifically confirms T4's breaking signature change to `HandleInbound`'s
transaction-ownership contract is recorded as an explicit compatibility consideration (not silently
absorbed), and that T7 (the stale-comment/legal-audit item) is correctly treated as
cross-reference-only against DATA-08 W0-T2's already-executed evidence
(`DATA-08/wave0/legal-audit/`), not re-implemented. S003 (FBL-04, P1) has passed a lighter,
documented review approach appropriate to its low-risk, well-bounded scope, per that story's own
`tasks/index.md` "Grouping rationale."

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA, applied to DATA-03/FBL-04).
