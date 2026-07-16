---
id: W02-E04-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W02-E04-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-16
---

# W02-E04-S001 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` are created on first real content, not pre-populated
empty.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W02-E04-S001-001 | fault-injection test report (`TestIntegrationAggregateWriteFaultInjection`, 4 stage subtests) | W02-E04-S001-T001 | AC-W02-E04-S001-01 | `go test ./kernel/resource/aggregate/... -run TestIntegrationAggregateWriteFaultInjection -v -count=1 -tags=integration` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced (corrected 2026-07-16 from stale "TBD"/"not yet produced" — see EV-005 Finding 2) |
| EV-W02-E04-S001-002 | unit/integration-test report (actor-attribution: `TestIntegrationAggregateWriteUserWithoutActorFailsFast`, `TestIntegrationAggregateWriteSystemActorPathsSucceed`) | W02-E04-S001-T002 | AC-W02-E04-S001-02 | `go test ./kernel/resource/aggregate/... -run 'TestIntegrationAggregateWriteUserWithoutActorFailsFast\|TestIntegrationAggregateWriteSystemActorPathsSucceed' -v -count=1 -tags=integration` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced (corrected 2026-07-16, same as above) |
| EV-W02-E04-S001-003 | regression-test report (`TestIntegrationRequestsModuleContract`) | W02-E04-S001-T003 | AC-W02-E04-S001-03 | `go test ./testkit/... -run TestIntegrationRequestsModuleContract -v -count=1 -tags=integration` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced (corrected 2026-07-16, same as above) |
| EV-W02-E04-S001-004 | documentation review (manual) | W02-E04-S001-T004 | AC-W02-E04-S001-04 | Not applicable (manual documentation review, not a command) | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced (corrected 2026-07-16, same as above) |
| EV-W02-E04-S001-005 | review report (independent review, task-005, re-verifying AC-01..04) | W02-E04-S001-T005 | AC-W02-E04-S001-01, AC-W02-E04-S001-02, AC-W02-E04-S001-03, AC-W02-E04-S001-04 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./kernel/resource/aggregate/... -run 'TestIntegrationAggregateWriteFaultInjection\|TestIntegrationAggregateWriteUserWithoutActorFailsFast\|TestIntegrationAggregateWriteSystemActorPathsSucceed' -v -count=1 -tags=integration` | HEAD 43b6e12 + remediation working tree 2026-07-16 | pass | produced |

Corrections applied 2026-07-16 by Independent review agent (Claude Sonnet 4.5), dispatched by
Fable 5 conductor (autopsy remediation R-3): EV-001..004's "TBD"/"not yet produced" placeholders
were stale — the underlying `.txt` output files in `evidence/tests/` were genuinely produced at
commit 1626b1132622aacc3e85475e4190e16a457ad1f6 but the index table itself was never updated to
reflect it. Environment for EV-005: macOS (darwin/arm64), go1.26.5.
