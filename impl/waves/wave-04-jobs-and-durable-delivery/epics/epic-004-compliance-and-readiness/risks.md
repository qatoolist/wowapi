---
id: W04-E04-RISKS
type: epic-risks
epic: W04-E04
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W04-002 and
RISK-W04-004 originate at wave scope and are reproduced/elaborated here because they land entirely
within this epic's stories (S001 and S003 respectively). Two further epic-specific risks are added
below for S002.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W04-002 | DATA-08 W6-T1 (audit hash-widening) is PLAN's own highest-risk task in this wave's Wave-6 scope: "No `hash_version` column exists today; widening makes every historical row unverifiable under new-scheme verification unless a version discriminator is added in the same migration and verification branches by row version... Single highest-risk task in PF-DATA's Wave-6 scope, and directly hits wowsociety's live audit rows" | Medium (a breaking format change on a live-production table is inherently execution-risky even with D-04's design resolved) | High — an incorrect version-branch implementation could make historical wowsociety audit rows permanently unverifiable, which is itself a compliance failure the fix was meant to prevent | High | W04-E04-S001 | D-04's ratified design (already resolved, not this epic's open question: `hash_version smallint NOT NULL DEFAULT 1`, canonicalized pre-serialization metadata hashing, version-branched verification) is implemented exactly as ratified, with a tamper test proving both the v1 historical-row branch and the v2 new-row branch verify correctly, and mutating any declared field independently breaks verification; the migration itself ships through W02-E01's online-migration protocol (expand/backfill/validate/contract), not an ad hoc one-off | PROD-05 (staging-drill coordination, product-level) requires a dedicated wowsociety-side staging verification pass before `FRAMEWORK_VERSION` is bumped past this migration's commit — tracked, not resolved, by this epic | unassigned | open | Cannot be fully eliminated within this epic's framework-side scope — final confidence requires the product-side staging drill (PROD-05), which is outside this epic's closure authority |
| RISK-W04-004 | DX-07 T4 (production-profile capacity/backpressure enforcement) cannot be implemented in this epic because it depends on AR-04 T5's waiver mechanism, which is W05 scope and does not yet exist. Deferring T4 leaves `CapacityMode` defaulting to `"advisory"` (never enforced) and `HTTPMaxInFlight` defaulting to `0` (backpressure fully disabled) unresolved through this epic's own closure | High (confirmed: the dependency is real and W05 has not been built as of this epic's planning) | Medium — a production deployment remains able to boot with capacity/backpressure effectively disabled and no readiness signal calling that out, until W05-E03-S002 and a later follow-on close the gap | Medium | W04-E04-S003 | S003's `story.md`/`plan.md` explicitly scope T4 out and record the forward dependency by requirement ID (AR-04 T5) and by target story (W05-E03-S002), not silently dropped; T1-T3 (migration-currency, hash reporting, config-doctor discovery fix) still close real, independently valuable gaps in this epic | Track as a deferred item in the deferred-items register with W05-E03-S002 named as the unblocking story; do not claim DX-07 "complete" at this epic's closure — only T1-T3 | unassigned | open | Medium until W05-E03-S002 lands and a follow-on task implements DX-07 T4 |
| RISK-W04-E04-001 | DATA-08 W6-T4's central legal-hold enforcement wrapper is a breaking change to the `DisposeFunc`/`EraseFunc` contract every registered `RecordClass` callback must satisfy — an incomplete enumeration of currently-registered callbacks (in both wowapi and any consuming product) before the wrapper lands risks silently breaking a callback the wrapper's negative test did not anticipate | Medium | Medium — a missed callback would either fail closed (blocking a legitimate dispose/erase) or, worse, fail open (defeating the wrapper's own purpose) depending on how the contract change is implemented | Medium | W04-E04-S002 | S002's `plan.md` requires enumerating every registered `RecordClass` in both repos before implementing the wrapper, per PLAN DATA-08 W6-T4's own risk note ("Breaking change to the `DisposeFunc`/`EraseFunc` contract — enumerate every registered `RecordClass` in both repos first") | If a callback is discovered post-landing that the wrapper does not correctly handle, treat it as a regression requiring an immediate follow-up fix, not a silently-accepted gap | unassigned | open | Low once the enumeration step is completed as planned |
| RISK-W04-E04-002 | DATA-08 W6-T3's encrypted DSR export artifact introduces a new encryption-key-management dependency not previously present anywhere in this epic's scope | Low-medium | Medium — an under-specified key-management design could leave exported DSR artifacts either unrecoverable (key loss) or insufficiently protected (weak key handling) | Medium | W04-E04-S002 | S002's `plan.md` records the key-management design as an implementation-time decision requiring explicit documentation of key custody, rotation, and recovery, rather than an incidental implementation detail | If the design proves materially under-specified during implementation, escalate as a story-level unresolved question before the export mechanism is locked, not silently defaulted | unassigned | open | Low-medium pending the implementation-time key-management design |

## Residual risk after mitigation

RISK-W04-002 is a wave-level risk with epic-scoped mitigation already described in `../../risks.md`;
it is not expected to fully resolve within this epic's own closure — final resolution requires the
product-side PROD-05 staging drill. RISK-W04-004 is expected to remain open at this epic's closure by
design — it resolves only once W05-E03-S002 lands and a follow-on task implements DX-07 T4, which
this epic explicitly does not attempt. RISK-W04-E04-001 and RISK-W04-E04-002 are expected to reduce
to low/low-medium residual risk once their respective mitigations (the callback enumeration step; the
key-management design decision) are executed as planned in S002.
