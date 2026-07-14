---
id: W04-E01-RISKS
type: epic-risks
epic: W04-E01
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W04-001 and
RISK-W04-003 originate at wave scope and are reproduced/elaborated here because they land entirely
within this epic's stories. One further epic-specific risk is added below.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W04-001 | W04-E01-S001 (this epic's shared lease/fencing primitive) supersedes W02-E01-S002's interim checkpoint lease, per `wave-allocation-detail.md`'s explicit note and the mirrored entry RISK-W02-001 in W02's own risk register. Any migration checkpoint state written under the interim lease (by DATA-09's backfill harness, in production or in test fixtures, before this epic lands) must be correctly read or translated once the shared primitive replaces it — an incomplete migration would either lose in-flight checkpoint state or silently misinterpret it | Medium | Medium-high — an incorrectly translated checkpoint could cause a backfill job to reprocess rows it already completed, or to skip rows it had not yet reached, undermining DATA-09's own "no reprocessing or skipping" acceptance bar for the harness that used the interim lease | Medium-high | W04-E01-S001, W02-E01-S002 (receiving-side mirror of RISK-W02-001) | S001's `plan.md` requires an explicit migration step (not a big-bang cutover) that reads any existing interim-lease checkpoint state and re-expresses it under the shared primitive's schema before the interim lease code path is removed, plus a test proving no in-flight backfill's checkpoint state is lost or duplicated across the cutover | If a live backfill is genuinely in flight at cutover time, pause it, complete the migration, and resume rather than cutting over underneath a running job; record the pause/resume as a deviation if it occurs | unassigned | open | Medium until the migration step is implemented and its test passes |
| RISK-W04-003 | DATA-02 T5's worker-signature change ("stable job idempotency key + lease context passed to workers") is confirmed breaking: "worker signature change is breaking — coordinate with wowsociety even though it has zero current job usage" (PLAN DATA-02 T5 risk column) | Low near-term (wowsociety has zero current `kernel/jobs` usage, confirmed by import grep per PLAN's wowsociety-impact note), but certain to matter the moment wowsociety registers its first job | Low near-term / Medium future — no current breakage, but an unannounced signature change would surprise any future wowsociety job registration | Low (time-bounded by wowsociety's current non-usage) | W04-E01-S003 | S003's `plan.md` records the breaking change explicitly in its own "Unresolved questions"/coordination-notes section rather than silently shipping it as if it were additive; the change is flagged for the framework's own changelog/migration-guide process | If wowsociety registers a job before this coordination note is acted on, treat it as a standard breaking-change-adoption cycle (version bump + migration guide), not an emergency | unassigned | open | Low today; escalates only when wowsociety's own roadmap intersects |
| RISK-W04-E01-001 | The shared primitive (S001), once locked as the contract W04-E02 and W04-E03 both consume, becomes load-bearing across three epics at once — a design gap discovered only after E02/E03 begin consuming it (e.g. a missing field one of them needs) is materially more expensive to retrofit than a gap found within a single-consumer story, because it would require re-touching S002/S003's already-built jobs application plus whichever of E02/E03 had already started consuming it | Low-medium | Medium-high — a retrofit after multi-epic consumption has begun risks a coordinated multi-story rework, not a single story's fix | Medium | W04-E01-S001 (as author), W04-E02/W04-E03 (as consumers) | PLAN DATA-02 T1's own risk note ("Architecturally load-bearing across all three findings") is treated as a design-review trigger: S001's task record requires the primitive's field set to be validated against DATA-03's and DATA-04's own stated needs (not just DATA-02's) before being treated as locked, even though this epic does not implement E02/E03 itself | If a gap is found during E02/E03 implementation, treat it as a primitive-extension task recorded in `deviations.md` at the consuming epic, with a cross-reference back to this epic's own closure record — not a silent field addition with no record of why the original design was incomplete | unassigned | open | Low once the cross-consumer field-set review is honored during S001 |

## Residual risk after mitigation

RISK-W04-001 and RISK-W04-003 are wave-level risks with epic-scoped mitigation already described in
`../../risks.md`; neither is expected to fully resolve within this epic's own closure —
RISK-W04-001 resolves once its migration step is implemented and evidenced, and RISK-W04-003 remains
open (by design) until wowsociety's own roadmap intersects job registration. RISK-W04-E01-001 is
expected to reduce to low residual risk once S001's cross-consumer field-set review is executed as
planned, before the primitive is treated as locked.
