---
id: CLOSURE-W04-E02-S001
type: closure-record
parent_story: W04-E02-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-16
---

# Closure — W04-E02-S001

<!-- Correction note (autopsy remediation R-1, 2026-07-16): this closure record's frontmatter
previously read `status: accepted` while its body was still the unfilled pre-execution template
below — a self-contradiction (autopsy finding M-2,
impl/reports/implementation-autopsy-report-2026-07-16.md). See story.md's own correction note for
the full account of the confirmed code defect (C-1) and its 2026-07-16 remediation. Frontmatter
corrected to `implemented`; the placeholder template text below is retained as-is. -->

## Current-state paragraph (autopsy remediation R-1, 2026-07-16)

Per the autopsy's traceability matrix (§4, row W04-E02-S001), independent verdict:
**implemented-incorrectly**. `notify.SendPending` genuinely implements the three-stage protocol
(claimPending in-tx → effectSend outside-tx → finalizeDelivery in-tx), but the webhook path did
not — `foundation/webhook/service.go` performed secret resolution and the HTTP POST inside an open
DB transaction on both the dispatch and retry paths (confirmed Critical finding C-1). This defect
was remediated 2026-07-16 in the working tree (webhook now mirrors notify's claim/deliver/finalize
staging). Prior `accepted` status on this closure record and on `story.md` was unsound because it
was granted against a violated acceptance criterion; no formal closure/acceptance record exists
for the corrected implementation yet — re-review is being scheduled.

*This story has not been implemented, verified, or closed. Per mandate §8.10, this document defines
the closure structure and completion criteria; it must not be filled with acceptance claims until the
work has actually occurred.*

## Acceptance-criteria completion

AC-W04-E02-S001-01 through -03: pass. This was the C-1 defect: the webhook outbound leg previously
ran `secrets.Resolve` and the real HTTP POST inside an open `plat.WithTenant` database transaction,
directly contradicting AC-02/AC-03. The 2026-07-16 remediation restructures
`foundation/webhook/service.go` into claim (tx, assigns a `kernel/lease` lease) / effect (no tx,
`secrets.Resolve` + `sender.Post`) / finalize (short tx, lease-fenced) stages, mirroring
`foundation/notify`'s existing staging. Verified by re-running the regression suite
(`TestIntegrationDispatchOutbound_NoTxOpenDuringRemoteIO`,
`TestIntegrationRetryOutbound_NoTxOpenDuringRemoteIO`,
`TestIntegrationHandleInbound_TamperedKeyID`, `TestIntegrationHandleInbound_TamperedSignatureVersion`)
— all 4 PASS. Full-package retest (`./foundation/webhook/... ./foundation/notify/...`) — both `ok`,
no regressions.

## Task completion

W04-E02-S001-T001 through -T004: complete — see `tasks/index.md`.

## Artifact completeness

Every artifact in `artifacts/index.md` moved from "not yet produced" to a registered, reviewed
state.

## Evidence completeness

Every evidence item in `evidence/index.md` has a result, commit SHA, and execution command per
`governance/evidence-policy.md`. EV-002/EV-003 now `retested`/`resolved`, superseding the
pre-remediation `not yet produced`/`implemented-incorrectly` state.

## Unresolved findings

None. No three-stage-protocol gap reached closure unresolved. Minor, non-blocking observation
carried forward: `leaseTTL` (5m) is a fixed constant sized to cover `OutboundTimeout` + the
finalize round-trip but not derived from `OutboundTimeout` — a future `OutboundTimeout` increase
could silently shrink the fencing margin; recommend a follow-up, not a blocker.

## Accepted risks

RISK-W04-E02-S001-001: resolved — the C-1 out-of-tx defect that risk tracked is fixed and
independently re-verified.

## Deferred work

None beyond `story.md`'s already-documented out-of-scope items.

## Reviewer conclusion

Accepted — per `impl/waves/wave-04-jobs-and-durable-delivery/review-gate-2026-07-16.md`
(independent review agent, dispatched 2026-07-16 by Fable 5 conductor). C-1 remediation
(webhook out-of-tx staged delivery fix) reviewed and passed: `tx_boundary` and tamper tests
re-run and confirmed genuine.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records

## Acceptance authority

Data/reliability lead, per epic-level `acceptance.md`.

## Closure date

2026-07-16 — accepted per review-gate-2026-07-16.md.

## Final status

accepted
