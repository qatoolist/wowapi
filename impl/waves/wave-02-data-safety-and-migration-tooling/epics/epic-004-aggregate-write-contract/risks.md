---
id: W02-E04-RISKS
type: epic-risks
epic: W02-E04
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E04 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W02-E04-001 | This epic's T1 aggregate repository/unit-of-work helper overlaps conceptually with AR-03's future "one authoritative declaration, derived projections" work (W05-E03) — PLAN DATA-06 T1's own risk note states this explicitly: "Overlaps AR-03 — coordinate to avoid a parallel one-off mechanism." Because AR-03 is sequenced much later (W05), this epic's helper design may need reconciliation once AR-03 actually lands, if AR-03's eventual shape conflicts with decisions this epic makes now | Medium | Medium — if AR-03's design diverges materially from this epic's helper shape, W05-E03 may need a migration/adapter step to reconcile the two, adding scope there that this epic's own closure cannot prevent | Medium | W02-E04-S001, W05-E03 (forward reference) | This epic's `plan.md` documents the helper's design decisions explicitly (not merely as code) so a future AR-03 implementer can evaluate compatibility deliberately rather than discovering the overlap by surprise; the helper is scoped narrowly to DATA-06's own T1 acceptance criterion (aggregate write + mirror + audit + outbox atomicity) rather than attempting to pre-empt AR-03's broader declaration-and-projection model | If W05-E03 finds the two designs incompatible, record a deviation at that point (in W05-E03's own `deviations.md`, not retroactively rewritten into this epic's plan) describing the reconciliation approach chosen | unassigned | open | Medium until W05-E03 either confirms compatibility or completes its own reconciliation |

## Residual risk after mitigation

RISK-W02-E04-001 cannot be fully resolved within this epic's own scope, since its resolution depends
on a design decision (AR-03's own shape) that belongs to a much later wave. This epic's mitigation —
explicit, reviewable documentation of its own design decisions — is expected to reduce the risk to
a well-understood, trackable coordination item rather than a silent architectural conflict
discovered late in W05.
