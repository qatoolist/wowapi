---
id: W01-E04-RISKS
type: epic-risks
epic: W01-E04
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E04 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W01-004 and
RISK-W01-005 originate at wave scope and are reproduced/elaborated here because they land entirely
within this epic's stories. Three further epic-specific risks are added below, scoped narrowly to
this epic's own stories.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W01-004 | T-TEST-01's reproduction step fails to reproduce the intermittent `internal/e2e` failure at all, leaving the diagnosis inconclusive | Medium | Low — this is an investigation story; an inconclusive result is itself a valid, honestly-recorded outcome, not a story failure | Low | W01-E04-S003 | The story's completion criteria explicitly allow "reproduce, then diagnose what the reproduction shows" rather than pre-committing to a mechanism (MATRIX CS-13's own re-scoping) | If unreproducible after a reasonable `-count`+parallel budget, record that finding and downgrade to a monitoring item rather than open-ended investigation | unassigned | open | Low — accepted as a legitimate story outcome, not a residual project risk |
| RISK-W01-005 | Generator fix (DX-02 `.delete`→`.deactivate`) needs to also fix the generator's own test that currently asserts the buggy verb as correct (`TestGenCRUDPermissionKeys`) — missing this sub-fix would leave the test suite red or, worse, silently reverted | Low | Medium — a missed fix here means the bug is test-locked again immediately | Low-medium | W01-E04-S001 | Explicitly named as a distinct task item (T003) rather than folded silently into the template-fix task, so it cannot be dropped without the task itself being visibly incomplete | Independent review specifically checks this test was updated, not just the template | unassigned | open | Low |
| RISK-W01-E04-001 | DX-01's fail-closed default (VCS-derived pseudo-version, refusing dirty/unreachable commits) could be perceived as a regression by a developer used to the old (silently-wrong) `v0.0.0` fallback always "succeeding," generating friction even though the old behavior was never actually correct | Low | Low — this is intended, documented behavior; the risk is developer confusion, not a functional defect | Low | W01-E04-S001 | S001's task explicitly requires the failure message to include the exact remediation command (`--framework-version` or `--local-framework`), so the fail-closed path is self-service, not a dead end | If the remediation message proves insufficient in practice, expand it as a follow-up rather than reintroducing the fallback | unassigned | open | Low |
| RISK-W01-E04-002 | S002's DX-05 T3 (blueprint-11 CLI example reconciliation) requires a per-example implement-or-delete judgment call against `internal/cli/cli.go`'s real commands/flags; a large number of stale examples could expand this task's scope beyond what was estimated at planning time | Low-medium | Low-medium — could expand S002's T002 task's bounded scope if the blueprint has drifted significantly since it was authored | Low-medium | W01-E04-S002 | The task explicitly re-confirms `internal/cli/cli.go`'s actual current commands/flags at implementation time rather than trusting the blueprint's age; if the count of stale examples is large, the task can be split into a follow-up rather than silently expanding | Record any split as a deviation in `deviations.md` rather than quietly absorbing extra scope | unassigned | open | Low-medium |
| RISK-W01-E04-003 | FBL-03's wowsociety-register reconciliation is, by cross-repository necessity, a recommendation this epic cannot verify was actually applied — there is a structural risk that the register drifts further out of sync if the wowsociety-side edit is never actually made | Low | Low — the framework-side fix (DX-02) is unaffected either way; the risk is purely to the accuracy of a downstream document this repository does not own | Low | W01-E04-S002 | S002's acceptance criteria are scoped to "a documented, precise PROD-level coordination recommendation exists," not "the wowsociety register was edited" — this makes the epic's own closure independent of an action outside its control, per mandate §2.3 | If the coordination recommendation is not acted on, this is tracked as a deferred cross-repository item in `impl/tracking/deferred-items-register.md` (programme level), not a reopened W01-E04 story | unassigned | open | Low |

## Residual risk after mitigation

RISK-W01-004 and RISK-W01-005 reduce to Low residual risk once S003's honest-inconclusive-outcome
allowance and S001's explicit test-lock-fix task item are actually honored as planned. The three
epic-specific risks are all Low or Low-medium and do not block closure — each has a scoping or
recording mitigation that keeps the epic's acceptance boundary independent of an uncontrollable
external factor (developer perception, blueprint drift volume, or a downstream repository's actual
edit).
