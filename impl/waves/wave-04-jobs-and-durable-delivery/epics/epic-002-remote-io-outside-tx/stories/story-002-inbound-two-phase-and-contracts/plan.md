---
id: PLAN-W04-E02-S002
type: plan
parent_story: W04-E02-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E02-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan does not invent precise code changes where the repository does not yet
provide enough information — `HandleInbound`'s exact current transaction structure has not yet been
inspected in detail by this plan, and the exact retry-bound figure for T4's discard+retry loop is
recorded as an implementation-time decision, not invented here.

## Proposed architecture

**T4 — inbound two-phase verification.** `HandleInbound` is restructured into: (1) a short read-tx
that snapshots the target endpoint's verification state (secret, version, status) and closes; (2)
signature verification performed entirely outside any open transaction, using the snapshot; (3) a
short write-tx that re-checks the endpoint's current version/status against the snapshot — if they
match, commit the verification outcome; if they mismatch (a rotation or deactivation occurred in the
window), discard this attempt's result and retry from step 1, bounded by an explicit retry ceiling.

**T5 — failed-signature audit.** On any failed verification (including a T4 discard, if it
ultimately resolves to failure), write a body-free audit row in its own short transaction, separate
from the verification transactions themselves.

**T6 — per-adapter idempotency-safety contract declaration.** A boot-time registration check
requiring every adapter registered for a non-idempotent high-impact operation to declare its
duplicate-safety mechanism (inbox/effect ledger, domain CAS, or provider idempotency key, per
`wave.md`'s framework-capabilities framing); the framework's boot sequence rejects registration of
an adapter that has not made this declaration.

**T8 — 6-boundary chaos test.** A chaos-test suite, built on W04-E01-S003's shared harness, that
injects failure at each of the 6 named boundaries (before send, during send, after success/before
finalize, lease expiry, duplicate workers, provider timeout) for both notify and webhook, asserting
zero duplicate external effects at each.

## Implementation strategy

1. Re-read `HandleInbound`'s current transaction structure at this story's actual start commit,
   confirming the current-state assessment (single enclosing transaction across snapshot and
   verification) still holds.
2. Design the two-phase protocol's exact snapshot data shape (secret, version, status — and any
   other field whose change during the window would invalidate a verification) and the
   re-check comparison logic.
3. Implement the read-tx snapshot stage, the out-of-tx verification stage, and the write-tx
   re-check-and-commit stage, with discard+retry on mismatch, bounded by an explicit retry ceiling
   (exact bound TBD, see "Unresolved questions").
4. Record the breaking-change note (T4's transaction-ownership contract change) explicitly in this
   plan and in `story.md`'s "Compatibility considerations"; enumerate and update in-repo callers of
   `HandleInbound`.
5. Write the rotation-during-verification test: rotate or deactivate the endpoint's secret in the
   window between snapshot and verification, confirming the re-check discards and retries rather
   than accepting under the stale policy.
6. Design and implement the failed-signature audit table/row write, in its own short transaction, on
   every failed verification, confirming no raw body is captured.
7. Write the empty-body-field test.
8. Inventory all existing `Sender` implementations (and any other adapter registered for a
   high-impact operation) to determine their current duplicate-safety posture before enforcement is
   turned on.
9. Design and implement the per-adapter idempotency-safety contract-declaration API and its
   boot-time enforcement check.
10. Write the boot-time fixture test asserting an undeclared adapter is rejected.
11. Confirm W04-E01-S003's shared chaos harness's API/fixture shape; do not redesign it.
12. Implement the 6-boundary chaos-test suite for notify, reusing the harness, exercising each of
    the 6 named boundaries against S001's three-stage protocol.
13. Implement the same 6-boundary chaos-test suite for webhook.
14. Record the T7 cross-reference note pointing at `DATA-08/wave0/legal-audit/`.
15. Document all of T4/T5/T6/T8's behavior per "Documentation requirements" in `story.md`.

## Expected package or module changes

`kernel/webhook` (`HandleInbound`'s transaction structure; a new failed-signature-audit write path);
`kernel/notify` and `kernel/webhook`'s adapter-registration mechanism (new idempotency-safety
contract-declaration API and boot-time enforcement); a new chaos-test suite package/location reusing
W04-E01-S003's harness (exact location TBD, expected alongside or importing that harness's own
package).

## Expected file changes where determinable

- `webhook/service.go` or its `HandleInbound`-owning file — restructured into the two-phase
  protocol.
- A new failed-signature-audit write path (exact file path TBD).
- The adapter-registration mechanism (exact file path TBD, expected in or near
  `kernel/notify`/`kernel/webhook`'s existing `Sender` interface definition).
- New chaos-test files for notify and webhook (exact paths TBD, expected under `DATA-03/chaos/` per
  the source's own evidence-path convention).

## Contracts and interfaces

A new idempotency-safety-declaration interface/contract each `Sender` (and other high-impact
adapter) implementation must satisfy — exact shape TBD, expected to require declaring one of:
inbox/effect ledger, domain CAS, or provider idempotency key (per `wave.md`'s framework-capabilities
framing of DATA-03's per-adapter contract). `HandleInbound`'s new transaction-ownership contract —
exact signature change TBD, to be recorded here once designed, with the breaking-change note
carried into `deviations.md` if it differs from what this plan anticipates.

## Data structures

A new failed-signature-audit row schema (body-free by construction — likely storing signature,
timestamp, endpoint reference, and failure reason, but explicitly never the raw request body).
Exact schema TBD at implementation time.

## APIs

`HandleInbound`'s signature is confirmed to change in a breaking way per T4's own risk note — exact
new signature TBD, to be documented here once implemented, with the change flagged in the
framework's own changelog/migration-guide process per this epic's `risks.md` mitigation.

## Configuration changes

The retry ceiling for T4's discard+retry loop may be a hardcoded constant or a configuration key —
to be determined at implementation time, mirroring W02-E01-S001-T002's own unresolved-question
treatment of its retry ceiling.

## Persistence changes

A new failed-signature-audit table (T5). No other schema change anticipated for T4/T6/T8.

## Migration strategy

The failed-signature-audit table is new (additive); no backfill required. No other migration
anticipated.

## Concurrency implications

T4's two-phase split is itself a concurrency-safety mechanism: the write-tx re-check exists
specifically to detect a concurrent secret rotation/deactivation that occurred during the
out-of-tx verification window. T8's chaos test is the concurrency-correctness proof for the whole
epic's claim/effect/finalize protocol under simulated worker failure and duplication.

## Error-handling strategy

A T4 mismatch (rotation/deactivation detected at re-check) must discard the verification attempt's
result and retry, not silently accept the stale result nor crash — bounded by the retry ceiling. A
T6 boot-time rejection of an undeclared adapter must fail the boot sequence with a clear,
adapter-identifying error, not merely log a warning and continue.

## Security controls

T4, T5, and T6 are each themselves security controls (see `story.md` "Security considerations") —
this plan does not add separate hardening beyond correctly implementing what T4/T5/T6 already
specify.

## Observability changes

Log lines for T4's discard+retry events (mirroring W02-E01-S001-T002's bounded-retry observability
pattern) and for T6's boot-time adapter-rejection events, per `story.md` "Observability
considerations."

## Testing strategy

- T4: rotation-during-verification test (rotate/deactivate the secret in the snapshot-to-
  verification window, confirm discard+retry, confirm no accept-under-stale-policy).
- T5: empty-body-field test on failed verification.
- T6: boot-time fixture test rejecting an undeclared adapter; positive fixture confirming a
  correctly-declared adapter registers successfully.
- T8: the 6-boundary chaos test itself, applied independently to notify and to webhook, asserting
  zero duplicate external effects at each of the 6 named boundaries.

## Regression strategy

The rotation-during-verification test, the empty-body-field test, the boot-time adapter-rejection
fixture, and the 6-boundary chaos test all become permanent regression guards once landed.

## Compatibility strategy

T4's breaking signature change is recorded explicitly (see "APIs" above and `story.md`
"Compatibility considerations"), with in-repo callers of `HandleInbound` enumerated and updated as
part of this story's own implementation, and the wowsociety-side non-blocking risk tracked per
`wave.md`'s framing, not resolved here.

## Rollout strategy

Single story, landed as its own reviewable unit. T4/T5 (inbound verification + audit) and T6
(adapter contracts) are logically separable and may land as separate commits/PRs within this story;
T8 (chaos test) necessarily lands last, since it depends on T2-T4 (S001's protocol plus this story's
T4) all being in place.

## Rollback strategy

Revert T4's two-phase protocol if it produces false-positive discard+retry loops under legitimate
(non-adversarial, non-rotation) load; revert T6's boot-time enforcement if the `Sender` inventory
(step 8) surfaces an adapter that cannot be feasibly declared before this story's own deadline —
escalate that finding rather than silently disabling enforcement without recording why.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–15). Step 11 (confirming W04-E01-S003's
harness shape) is a hard gate before steps 12–13 (the chaos-test implementations) can proceed.

## Task breakdown

- **W04-E02-S002-T001** — Inbound two-phase verification for `HandleInbound` (steps 2–5 above).
- **W04-E02-S002-T002** — Failed-signature audit path (steps 6–7 above).
- **W04-E02-S002-T003** — Per-adapter idempotency-safety contract declaration (steps 8–10 above).
- **W04-E02-S002-T004** — Named 6-boundary chaos test for notify and webhook (steps 11–13 above).
- **W04-E02-S002-T005** — Evidence aggregation (consolidating T001–T004's evidence, including the
  T7 cross-reference record, into one story-scope acceptance package — see `tasks/index.md`
  "Grouping rationale" for why this is warranted here but was not warranted in W02-E01-S001).
- **W04-E02-S002-T006** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The inbound two-phase verification implementation; the failed-signature audit path; the per-adapter
idempotency-safety contract-declaration mechanism and `Sender` inventory; the 6-boundary chaos-test
suite for notify and webhook; the T7 cross-reference record.

## Expected evidence

Rotation-during-verification test output; empty-body-field test output; boot-time
undeclared-adapter-rejection fixture test output; the 6-boundary chaos test output for both notify
and webhook.

## Unresolved questions

- Exact retry-ceiling bound for T4's discard+retry loop.
- Exact new signature for `HandleInbound` after the breaking transaction-ownership contract change.
- Exact failed-signature-audit table schema.
- Exact shape of the per-adapter idempotency-safety-declaration contract/interface.
- Exact package location for the chaos-test suite (expected under `DATA-03/chaos/` per the source's
  own evidence-path convention, but not yet confirmed against the actual repository layout).
- Whether any inventoried `Sender` implementation (step 8) is found non-idempotent-and-undeclared,
  and if so, what the resolution path is (a follow-up item vs. blocking this story).

## Approval conditions

This plan is approved for implementation once: (a) W04-E02-S001 and W04-E01-S003 have both reached
a stable, reviewable state (need not be fully `accepted`, but S001's three-stage protocol and
W04-E01-S003's chaos harness must both be stable enough for this story's own dependent tasks to
proceed), and (b) the owner and reviewer are assigned.
