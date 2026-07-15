---
id: W02-E01-RISKS
type: epic-risks
epic: W02-E01
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W02-001 and
RISK-W02-003 originate at wave scope and are reproduced/elaborated here because they land entirely
within this epic's stories. One further epic-specific risk is added below.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W02-001 | S002's interim checkpoint lease (built because DATA-02 T1's shared lease/fencing primitive does not exist yet) creates a genuine, planned technical-debt window until W04-E01-S001 replaces it | Medium | Medium — migration-checkpoint state written under the interim lease must be correctly readable/translatable once W04 migrates to the shared primitive | Medium | W02-E01-S002, W04-E01-S001 (forward reference) | S002's `plan.md` bounds the interim lease's surface to exactly what T4's backfill harness needs (checkpoint token + resumability only, no job-claim fencing or heartbeat) | If W04-E01-S001 is delayed, record continued interim-lease use as an accepted, time-bounded technical-debt item, not a silent permanent fork | unassigned | open | Medium until W04-E01-S001 lands |
| RISK-W02-003 | S003's canary/soak tooling (T6) has no production telemetry baseline for calibrating soak-duration/threshold values — PLAN's own risk note calls this "a genuine, currently unresolvable judgment gap" | High (confirmed gap, not merely likely) | Medium — does not block the tooling's own deterministic tests, but leaves the first real production rollout through this protocol without a pre-validated threshold | Medium | W02-E01-S003 | T6's tooling accepts configurable soak-duration/threshold parameters rather than hardcoding a guess | Accept as residual risk explicitly, recorded at closure, not silently dropped | unassigned | open | Accepted — mechanism delivered, calibration is a per-rollout operational judgment call |
| RISK-W02-E01-002 | S001's manifest schema, once locked, becomes the contract every subsequent migration in the repository must satisfy — if the schema is under-specified (e.g. missing a field a later phase's tooling needs), retrofitting it onto migrations written against an earlier schema version is more costly than getting the schema right the first time | Low-medium | Medium — a schema gap discovered during S002/S003 implementation would require either a manifest-schema migration of its own or a compatibility shim | Low-medium | W02-E01-S001 (as author), W02-E01-S002/S003 (as consumers) | PLAN DATA-09 T1's own risk note: "Get external review before locking the format" — S001's task record requires this review step explicitly, not as an optional nicety | If a gap is found during S002/S003, treat it as a manifest-schema extension task recorded in `deviations.md`, not a silent field addition with no record of why the original schema was incomplete | unassigned | open | Low once the external-review step is honored |

## Residual risk after mitigation

RISK-W02-E01-002 is expected to reduce to low residual risk once S001's external-review step is
executed as planned. RISK-W02-001 and RISK-W02-003 are wave-level risks with epic-scoped mitigation
already described in `../../risks.md`; neither is expected to fully resolve within this epic's own
closure — RISK-W02-001 resolves only when W04-E01-S001 lands, and RISK-W02-003 is accepted, not
resolved, per PLAN's own framing.
