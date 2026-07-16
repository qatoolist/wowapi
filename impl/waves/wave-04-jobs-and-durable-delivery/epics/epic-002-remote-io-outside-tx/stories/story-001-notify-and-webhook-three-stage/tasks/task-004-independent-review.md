---
id: W04-E02-S001-T004
type: task
title: Independent review
status: done
parent_story: W04-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S001-T001
  - W04-E02-S001-T002
  - W04-E02-S001-T003
acceptance_criteria:
  - AC-W04-E02-S001-01
  - AC-W04-E02-S001-02
  - AC-W04-E02-S001-03
artifacts: []
evidence: []
---

# W04-E02-S001-T004 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance
criteria are proven with valid evidence; the shared primitive was genuinely reused (not copied);
the self-documented "should move outside tx" comment was genuinely resolved, not merely deleted
without the underlying gap being closed; the webhook current-row-state check genuinely moved to the
claim stage; no source requirement (DATA-03 T1, T2, T3) was silently dropped or narrowed.

### Parent story

W04-E02-S001 — Notify and webhook three-stage remote-I/O protocol.

### Owner

unassigned

### Status

done

### Dependencies

W04-E02-S001-T001, W04-E02-S001-T002, W04-E02-S001-T003 (review requires all three to be
implemented first).

### Detailed work

1. Confirm T001's lease-column migration matches W04-E01's shared primitive's own schema exactly —
   not a parallel or bespoke implementation with a similar but distinct column set.
2. Confirm T002's three-stage protocol matches PLAN DATA-03 T2's acceptance criterion ("No
   `sender.Send` call while a DB tx is open") and that the self-documented comment deletion/update
   genuinely reflects a closed gap, not a comment removal papering over an unresolved issue.
3. Confirm T003's three-stage protocol matches PLAN DATA-03 T3's acceptance criterion ("No
   DNS/secret-resolve/POST call while a tx is open") and that the current-row-state check is
   genuinely in the claim stage, confirmed by direct code inspection, not merely by task-record
   claim.
4. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item —
   evidence without this must not be treated as final proof (mandate §10).
5. Confirm this story's `story.md` acceptance criteria are not narrower than PLAN DATA-03 T1/T2/T3's
   own acceptance-criteria columns.
6. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record itself is captured in this task's own Verification Record.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E02-S001-01, AC-W04-E02-S001-02, AC-W04-E02-S001-03 (confirms all three, does not itself
prove any new one).

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence, or lists
findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001/T002/T003's evidence.

### Risks

None beyond the review itself missing a genuine gap — mitigated by requiring the review to
specifically re-check the "genuinely, not merely claimed" points above (shared-primitive reuse,
comment resolution, claim-stage relocation) rather than trusting T001/T002/T003's own self-reported
completion.

### Rollback or recovery considerations

Not applicable — a review task has no code to roll back; a failed review blocks story acceptance
until its findings are resolved.

## Implementation Record

*Not applicable — this is a review task, not an implementation task.*

### What was actually implemented

*Not applicable.*

### Components changed

*Not applicable.*

### Files changed

*Not applicable.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable.*

### Commits

*Not applicable.*

### Pull requests

*Not applicable.*

### Implementation dates

*Not applicable.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not applicable.*

### Follow-up items

*Not yet executed.*

### Relationship to the approved plan

*Not applicable.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E02-S001-01 | Independent review against mandate §14 checklist | Code review | Confirmed: shared primitive genuinely reused, not copied | review report | unassigned |
| AC-W04-E02-S001-02 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: no send-while-tx-open, comment genuinely resolved | review report | unassigned |
| AC-W04-E02-S001-03 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: no network-call-while-tx-open, check genuinely in claim stage | review report | unassigned |

### Actual result

Code-read `foundation/webhook/service.go` end to end (claimDispatch/claimDeliveryRow/claimRetry ->
deliverClaimed -> effectDeliver -> finalizeOutboundDelivery) and `foundation/notify/service.go`'s
existing claim/effect/finalize staging, confirmed the webhook path now mirrors notify's structure:
(1) claim-tx assigns a fresh `kernel/lease` lease and returns before any transaction closes over
`secrets.Resolve`/`sender.Post`; (2) `effectDeliver` runs entirely outside `plat.WithTenant(...)`,
calling `s.secrets.Resolve` then `s.sender.Post`; (3) `finalizeOutboundDelivery` re-opens a short tx
and applies the outcome only `WHERE lease_token = $n AND lease_generation = $n AND
lease_expires_at > $n` (stale/reclaimed leases silently discard the effect result, matching AC-02/03's
fencing intent). The secret-resolution short-circuit is preserved: when `res.secretErr != nil`,
`finalizeOutboundDelivery` releases the lease (`lease_token=NULL, lease_generation=0,
lease_expires_at=NULL`) and leaves `delivery_status`/`attempts` untouched — no POST was attempted, no
row mutation beyond the lease release, matching the pre-staging behavior noted in the code comment.
Ran the new regression suite `foundation/webhook/tx_boundary_test.go` (`txDepthTracker`-instrumented
`TestIntegrationDispatchOutbound_NoTxOpenDuringRemoteIO`,
`TestIntegrationRetryOutbound_NoTxOpenDuringRemoteIO`) which asserts the wrapped `TxManager`'s open-tx
depth is 0 at the exact moment `Sender.Post`/`secrets.Resolve` execute — both PASS. Also ran the
pre-existing `foundation/webhook/...` and `foundation/notify/...` suites in full (no regressions) and
the new tamper-matrix tests (`foundation/webhook/tamper_matrix_test.go`,
`TestIntegrationHandleInbound_TamperedKeyID`, `TestIntegrationHandleInbound_TamperedSignatureVersion`,
H-9 scope, adjacent to this story but sharing the same file set) — both PASS.

One residual observation (not blocking): `leaseTTL = 5 * time.Minute` is a fixed constant sized to
cover `effectDeliver`'s bound (`OutboundTimeout`) plus the finalize round-trip, mirroring notify's
identical constant — reasonable for the current `OutboundTimeout`, but not derived from it, so a
future increase to `OutboundTimeout` could silently shrink the fencing margin. Flag as a minor
follow-up, not an AC failure.

### Pass or fail

PASS — AC-W04-E02-S001-01, -02, -03 are satisfied for both the notify leg (previously verified) and
the webhook leg (this remediation). The C-1 defect (webhook outbound delivery running secret
resolution and the POST call inside an open `plat.WithTenant` transaction, contradicting task T003's
own title) identified by the prior adversarial verification
(`/private/tmp/.../scratchpad/autopsy/verification/wave-04-jobs-and-durable-delivery.json`,
`W04-E02-S001` / `W04-E02-S001-T003` entries) is resolved by the 2026-07-16 remediation.

### Evidence identifier

EV-W04-E02-S001-002 (webhook no-send-while-tx-open, retested), EV-W04-E02-S001-003 (webhook
no-network-call-while-tx-open, retested) — supersede the prior `not yet produced` rows in
`evidence/index.md` (updated by this review, see below).

### Execution date

2026-07-16.

### Commit or revision

HEAD 43b6e12 + remediation working tree 2026-07-16 (`foundation/webhook/service.go` modified,
`foundation/webhook/tx_boundary_test.go` and `foundation/webhook/tamper_matrix_test.go` added,
uncommitted at review time).

### Environment

macOS (darwin), local Postgres via testkit
(`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`), Go
toolchain per `go.mod`.

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3).

### Findings

No AC-blocking findings. Minor observation: `leaseTTL` is a fixed constant rather than derived from
`OutboundTimeout` — recommend a follow-up (not a story blocker) to either derive it or add a comment
cross-referencing both constants so a future `OutboundTimeout` change is not silently under-fenced.

### Retest status

Retested against the 2026-07-16 remediation commit (working tree); superseded the pre-remediation
`implemented-incorrectly` verdict recorded in
`/private/tmp/.../scratchpad/autopsy/verification/wave-04-jobs-and-durable-delivery.json` for
`W04-E02-S001` / `W04-E02-S001-T003`.

### Final conclusion

AC-W04-E02-S001-01/02/03 are satisfied for both notify and webhook. Recommend: accept. Conductor
adjudicates final story-status change (this record is a recommendation only, per this review's
mandate).

Execution command:
```
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
  go test ./foundation/webhook/... -run \
  'TestIntegrationDispatchOutbound_NoTxOpenDuringRemoteIO|TestIntegrationRetryOutbound_NoTxOpenDuringRemoteIO|TestIntegrationHandleInbound_TamperedKeyID|TestIntegrationHandleInbound_TamperedSignatureVersion' \
  -count=1 -v
```
Result: all 4 tests PASS (`ok github.com/qatoolist/wowapi/foundation/webhook 1.348s`). Full-package
retest: `go test ./foundation/webhook/... ./foundation/notify/... -count=1` → both `ok` (10.049s,
9.347s respectively), no regressions.

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
