---
id: W04-E02-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W04-E02-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02-S001 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` are created on first real content, not pre-populated
empty.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W04-E02-S001-001 | test report | W04-E02-S001-T001 | AC-W04-E02-S001-01 | `go test ./kernel/lease/... -count=1` | HEAD 43b6e12 + remediation working tree 2026-07-16 | `ok github.com/qatoolist/wowapi/kernel/lease 0.479s` | resolved |
| EV-W04-E02-S001-002 | integration test report (no-send-while-tx-open assertion, notify+webhook) | W04-E02-S001-T002 | AC-W04-E02-S001-02 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./foundation/notify/... ./foundation/webhook/... -count=1` | HEAD 43b6e12 + remediation working tree 2026-07-16 | notify `ok` (9.347s); webhook `ok` (10.049s), including the boundary tests below | retested — supersedes the pre-remediation `implemented-incorrectly` verdict for the webhook leg recorded in `/private/tmp/claude-502/-Users-qatoolist-go-home-src-github-com-qatoolist-wowapi/97aeaae9-840e-4c51-bf72-b17540116e23/scratchpad/autopsy/verification/wave-04-jobs-and-durable-delivery.json` (`W04-E02-S001-T003`) |
| EV-W04-E02-S001-003 | integration test report (no-network-call-while-tx-open assertion, txDepthTracker) | W04-E02-S001-T003 | AC-W04-E02-S001-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./foundation/webhook/... -run 'TestIntegrationDispatchOutbound_NoTxOpenDuringRemoteIO|TestIntegrationRetryOutbound_NoTxOpenDuringRemoteIO' -count=1 -v` | HEAD 43b6e12 + remediation working tree 2026-07-16 | Both PASS; `txDepthTracker` asserts open-tx depth == 0 at the exact `secrets.Resolve`/`sender.Post` call sites | resolved — file `foundation/webhook/tx_boundary_test.go`, added 2026-07-16 (uncommitted at review time) |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once. EV-002/EV-003
above use `retested`/`resolved` because they supersede the pre-remediation state where the webhook
leg failed this same acceptance criterion (see `commands_run`/`reasoning` for `W04-E02-S001` and
`W04-E02-S001-T003` in the cited verification JSON).

Reviewer: Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor
(autopsy remediation R-3). Date: 2026-07-16. Full command output recorded in
`tasks/task-004-independent-review.md`'s Verification Record.
