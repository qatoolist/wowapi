---
id: W01-E02-RISKS
type: epic-risks
epic: W01-E02
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E02 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W01-E02-001 | S002's implementation begins before D-08 is actually ratified in W00-E02-S003 (the ADR file did not exist at the time this epic's planning documents were authored) | Medium | Medium — S002 would implement against an assumed decision that could differ from the final ratified wording | Medium | W01-E02-S002 | S002's `plan.md` explicitly flags D-08's exact wording as an assumption to confirm, not a presumed-unchanged fact, before implementation work starts; the epic's entry criteria (inherited from `wave.md`) require D-08 ratified before S002 begins | If D-08's ratified wording differs materially from the "thin in-kernel tracer, not otelpgx" assumption used in this epic's planning, record a deviation in S002's `deviations.md` rather than silently reconciling the plan to match | unassigned | open | Low once D-08 is confirmed ratified before S002 starts |
| RISK-W01-E02-002 | The negative-case test for S001 ("no active span → trace_id/span_id genuinely absent") is implemented as an empty-string check instead of an absent-key check, silently weakening the acceptance criterion into noise rather than a real assertion | Low | Medium — a passing-but-wrong test would give false confidence that correlation degrades gracefully when it actually leaks empty-string attributes into every log record | Low-medium | W01-E02-S001 | `story.md`/`plan.md`/task T002 explicitly specify "absent — not empty-string noise" as the acceptance bar; independent review checks this specifically per mandate §14 | Independent reviewer rejects the task's verification record if the negative-case test only checks for `""` rather than key absence, forcing a rewrite | unassigned | open | Low |
| RISK-W01-E02-003 | The `pgx.QueryTracer` implementation (S002/T001) inadvertently imports OTel SDK types into `kernel/database`, breaking the port-discipline boundary D-08 exists to protect | Low | Medium — would defeat the entire rationale for choosing a hand-rolled tracer over `otelpgx` in the first place | Low-medium | W01-E02-S002 | The tracer is specified to consume only the `observability.Tracer`/`Span` port (already vendor-neutral); independent review checks `kernel/database`'s imports for OTel SDK packages | Reviewer rejects the implementation and requires the OTel-specific logic to move into `adapters/tracing/otel` only | unassigned | open | Low |

## Epic-level risk summary

No risk here is rated above Medium severity; both stories are additive integrations against existing,
working infrastructure rather than new subsystem builds. The dominant risk (RISK-W01-E02-001) is a
sequencing risk inherited from this epic's one external dependency (D-08's ratification), not a
technical-implementation risk.
