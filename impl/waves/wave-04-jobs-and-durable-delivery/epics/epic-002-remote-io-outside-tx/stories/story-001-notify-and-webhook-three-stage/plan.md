---
id: PLAN-W04-E02-S001
type: plan
parent_story: W04-E02-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E02-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information — the exact claim-row schema integration shape depends on W04-E01-S001's
own finalized primitive API, which this story's plan tracks as an unresolved question rather than
pre-guessing.

## Proposed architecture

A three-stage protocol replacing today's single-transaction claim-and-send pattern in both
`kernel/notify` and `kernel/webhook.deliverToEndpoint`:

1. **Claim-tx** (short transaction): select the pending delivery row, assign a lease via W04-E01's
   shared primitive (`lease_token`, `lease_generation`, `lease_expires_at`), commit.
2. **Effect stage** (no transaction open): perform the remote call — `sender.Send` for notify;
   DNS resolution, secret resolution, and the POST for webhook — using the delivery ID as the
   idempotency key so a retried effect stage does not duplicate the remote call's outcome.
3. **Finalize-tx** (short transaction): compare the lease token against the current row's lease
   state (fencing check — if the lease has expired or been reclaimed, discard this attempt's
   result rather than overwrite a reclaiming worker's outcome); commit the delivery outcome.

The self-documented "should move outside tx" comment at `notify/service.go:456-586` (446-449) is
deleted (or replaced with a comment describing the now-implemented protocol) as part of stage 2's
implementation for notify.

## Implementation strategy

1. Re-read `notify/service.go:446-586` and `webhook/service.go`'s `deliverToEndpoint` and secret-
   resolution code paths at this story's actual start commit, confirming the current-state
   assessment (remote calls inside an open tx) still holds.
2. Confirm W04-E01-S001's shared lease/fencing primitive's finalized API surface (claim, lease
   columns, finalize-with-token-comparison) — this story's own claim-row integration cannot be
   designed in detail until that API is stable.
3. Design and migrate notify/webhook's delivery-tracking tables to carry the shared primitive's
   lease columns, replacing any prior ad hoc claim/status tracking those tables may already have.
4. Implement the three-stage protocol for `kernel/notify`: claim-tx assigning a lease; `sender.Send`
   outside any tx, keyed by delivery ID; finalize-tx comparing the lease token.
5. Delete/update the self-documented "should move outside tx" comment at `notify/service.go:
   456-586` (446-449) as part of step 4.
6. Implement the same three-stage protocol for `kernel/webhook.deliverToEndpoint`, moving the
   current-row-state check into the claim stage so the effect stage (DNS/secret-resolve/POST)
   requires no mid-flight DB read.
7. Add observability (log lines) for claim/effect/finalize stage transitions.
8. Document both packages' three-stage protocol boundaries.
9. Coordinate with W04-E02-S002 on the T8 boundary-matrix test scope — this story's own protocol is
   what that test exercises, but the test itself is S002's task.

## Expected package or module changes

`kernel/notify` (claim/send/finalize restructuring in and around `notify/service.go:446-586`);
`kernel/webhook` (claim/deliver/finalize restructuring in `webhook/service.go`'s
`deliverToEndpoint` and its callers); a new or extended migration adding shared-primitive lease
columns to notify/webhook's delivery-tracking tables.

## Expected file changes where determinable

- `notify/service.go` — the claim/send/finalize logic around lines 446-586 restructured into the
  three-stage protocol; the self-documented comment at 446-449 deleted or updated.
- `webhook/service.go` — `deliverToEndpoint` and its delivery loop restructured into the three-stage
  protocol; the current-row-state check relocated into the claim stage.
- A new migration file adding lease columns to notify/webhook's delivery-tracking tables (exact
  table names and migration file path TBD at implementation time).

## Contracts and interfaces

The claim-tx/finalize-tx functions for both notify and webhook consume W04-E01's shared primitive's
claim/finalize API — exact function signatures TBD pending that primitive's own finalized design
(see "Unresolved questions").

## Data structures

Notify/webhook delivery-tracking rows gain the shared primitive's lease columns
(`lease_token`, `lease_generation`, `lease_expires_at`). No other data-structure change anticipated.

## APIs

No caller-facing API change expected for either package's external `Send`/deliver entry points —
this is an internal control-flow restructuring. To be confirmed at implementation time; if an API
change proves necessary, record it as a deviation, not a silent scope expansion.

## Configuration changes

None anticipated beyond whatever configuration the shared lease primitive itself already exposes
(e.g. lease duration) — this story does not introduce new notify/webhook-specific configuration.

## Persistence changes

The claim-row schema migration described above. No other persistence change.

## Migration strategy

The lease-column migration is expected to be additive (new nullable/defaulted columns), not
destructive to existing delivery-tracking data — to be confirmed at implementation time given the
exact current schema of notify/webhook's delivery-tracking tables, which this plan has not yet
inspected in detail.

## Concurrency implications

This is the story's central concern: the three-stage split means a worker crash between claim-tx
and finalize-tx must be correctly fenced by the shared primitive's lease-token comparison, or a
reclaimed worker's late finalize could silently overwrite the reclaiming worker's outcome — the
exact failure mode DATA-02's own shared primitive exists to prevent, and the exact scenario
W04-E02-S002's T8 chaos test proves does not occur.

## Error-handling strategy

A remote-call failure during the effect stage must not leave the delivery row permanently
unclaimed or permanently stuck — the finalize-tx stage must record the failure outcome (for retry
scheduling) using the same lease-token-fencing discipline as a success outcome.

## Security controls

None new beyond what the shared primitive itself provides.

## Observability changes

Log lines for claim/effect/finalize stage transitions, per "Implementation strategy" step 7.

## Testing strategy

- Fail-first: a test confirming the current-state defect (remote call executes while a tx is open)
  fails against today's code, before this story's change lands — re-confirming the current-state
  assessment as an executable test, not just a documentation citation.
- No-send-while-tx-open assertion for notify (T2's own acceptance criterion).
- No-network-call-while-tx-open assertion for webhook, plus confirmation the current-row-state check
  occurs in claim, not effect, stage (T3's own acceptance criterion and risk note).
- The full 6-boundary chaos-test matrix (T8) is W04-E02-S002's own task, not duplicated here — this
  story's own tests prove the protocol's structural correctness in isolation; S002's chaos test
  proves it under worker-failure conditions.

## Regression strategy

The no-send/no-network-call-while-tx-open assertions, once landed, become permanent regression
guards against a future change reintroducing a remote call inside the transaction boundary.

## Compatibility strategy

No caller-facing API change expected (see "APIs" above); if the claim-row schema migration requires
a transition period for in-flight deliveries claimed under the old (pre-lease-column) schema, that
transition is designed and recorded here at implementation time — not yet determined given the
current schema has not been inspected in this plan.

## Rollout strategy

Single story, landed as its own reviewable unit. Notify and webhook's three-stage protocols may land
as separate commits/PRs within this story (T2 and T3 are independently testable) or together — to be
determined at implementation time.

## Rollback strategy

Revert the three-stage protocol and lease-column migration if the change destabilizes delivery
throughput or introduces an unanticipated race; the shared primitive itself (W04-E01-S001) is not
rolled back by this story's own rollback — only this story's consumption of it.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–9). Step 2 (confirming W04-E01-S001's
finalized API) is a hard gate before step 3 can proceed in detail.

## Task breakdown

- **W04-E02-S001-T001** — Shared-primitive reuse: claim-row schema migration for notify/webhook
  (step 3 above).
- **W04-E02-S001-T002** — Three-stage protocol for `kernel/notify`, including the self-documented
  comment deletion/update (steps 4–5 above).
- **W04-E02-S001-T003** — Three-stage protocol for `kernel/webhook.deliverToEndpoint` (step 6
  above).
- **W04-E02-S001-T004** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The claim-row schema migration; the notify three-stage protocol implementation; the webhook
three-stage protocol implementation; protocol-boundary documentation for both packages.

## Expected evidence

Lease-column migration test output; notify no-send-while-tx-open test output; webhook
no-network-call-while-tx-open test output.

## Unresolved questions

- W04-E01-S001's exact claim/finalize API surface — this story's own claim-tx/finalize-tx
  implementations cannot be finalized until that primitive's API is confirmed stable.
- Exact current schema of notify/webhook's delivery-tracking tables (table names, existing
  claim/status columns) — not yet inspected in this plan; required before the lease-column migration
  can be designed in detail.
- Whether notify and webhook's three-stage protocols land as one combined change or two separable
  ones within this story.
- Exact retry/backoff behavior for a failed effect-stage remote call — this story's own scope is the
  three-stage transaction-boundary structure; the retry mechanism itself may be affected by
  W04-E02-S003's FBL-04 adoption, and the two stories' implementers should coordinate to avoid
  building two incompatible retry mechanisms.

## Approval conditions

This plan is approved for implementation once: (a) W04-E01-S001 has reached a stable, reviewable API
surface (need not be fully `accepted`, but its claim/finalize contract must be stable enough for
this story's plan's unresolved questions to be answered), and (b) the owner and reviewer are
assigned.
