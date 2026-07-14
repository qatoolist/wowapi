---
id: W04-E02-RISKS
type: epic-risks
epic: W04-E02
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02 — Risks

Epic-scoped risks specific to W04-E02's own stories. Neither risk originates at wave scope in
`../../risks.md`; both are added here because they are specific to this epic's task content.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W04-E02-001 | S002's T4 (inbound two-phase webhook verification) introduces a breaking signature change to `HandleInbound`'s transaction-ownership contract — the function can no longer assume it owns a single enclosing transaction for its entire body, since verification must now occur between two separate short transactions | Medium (confirmed by the source's own framing of T4's risk column: "Breaking signature change to `HandleInbound`'s transaction-ownership contract; bound retry attempts") | Medium — any caller of `HandleInbound` that assumed the old single-tx contract (in-repo or, per `wave.md`'s wowsociety-impact note, a future wowsociety direct caller) would break silently without an explicit compatibility note | Medium | W04-E02-S002 | S002's `story.md`/`plan.md` records this as an explicit compatibility consideration/breaking-change note, not silently absorbed into "implementation detail"; in-repo callers of `HandleInbound` are enumerated and updated as part of T4's own task scope; the bound-retry-attempts requirement (also named in T4's risk column) is implemented as an explicit, documented ceiling, mirroring W02-E01-S001-T002's bounded-retry pattern | If a caller outside this epic's own visibility is found to depend on the old contract after this epic lands, treat it as a standard breaking-change-adoption cycle (version bump + migration guide), consistent with how DATA-02 T5's worker-signature change is handled at wave scope (RISK-W04-003) | unassigned | open | Low once the compatibility-consideration note is recorded and in-repo callers are confirmed updated; wowsociety-side risk is tracked, not resolved, per `wave.md`'s non-blocking framing |
| RISK-W04-E02-002 | This entire epic has a hard dependency on W04-E01-S001 (shared lease primitive) landing before S001's claim-row work can begin, and on W04-E01-S003 (shared chaos harness) landing before S002's T8 chaos test can run — if either W04-E01 story is delayed, this epic's own timeline slips in lockstep rather than having independent schedule flexibility | Medium (a real, structural dependency, not merely a scheduling preference — confirmed in `dependencies.md`) | Medium — a delay in W04-E01 blocks this epic's S001 and the chaos-test portion of S002 outright; S002's T4/T5/T6 (excluding T8) and S003 (FBL-04) are not blocked and may proceed independently | Medium | W04-E02-S001, W04-E02-S002 (T8 specifically) | This epic's own story sequencing allows S002's non-chaos tasks (T4, T5, T6) and S003 (FBL-04) to proceed without waiting on W04-E01, minimizing the blocked surface to exactly S001 and S002-T8; S001's `plan.md` should track W04-E01-S001's own status explicitly rather than assuming a landing date | If W04-E01-S001/S003 slip materially, re-sequence this epic's own story order (S003 and S002's non-chaos tasks first) rather than blocking all work on this epic until W04-E01 lands | unassigned | open | Low for S002-T4/T5/T6 and S003 (no dependency); remains Medium for S001 and S002-T8 until W04-E01-S001/S003 are confirmed accepted |

## Residual risk after mitigation

RISK-W04-E02-001 is expected to reduce to low residual risk once the compatibility-consideration
note is recorded and in-repo `HandleInbound` callers are confirmed updated as part of T4's own task
scope — full resolution of the wowsociety-side dimension remains tracked, not owned, by this epic,
consistent with `wave.md`'s "flag for future, not now" framing. RISK-W04-E02-002 is expected to
resolve for S001 and S002-T8 specifically once W04-E01-S001 and W04-E01-S003 reach `accepted`; it
does not block S002's T4/T5/T6 or S003, which carry no such dependency.
