---
id: W07-E02-RISKS
type: epic-risks
epic: W07-E02
wave: W07
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07-E02 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W07-002 originates at
wave scope and is reproduced/elaborated here because it lands entirely within this epic's S001.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W07-002 | SEC-05's external assessment surfaces an open Critical/High finding with no immediate remediation path available within this epic's own timeline | Low-medium | High if it occurs | High (conditional on occurrence) | W07-E02-S001 | The control map, built against SEC-01-04's already-`accepted` implementation, is expected to find few genuinely new gaps | If found, escalate to the product-security lead for emergency remediation or an approved, time-bounded waiver | unassigned | open | Low-medium — SEC-01-04's own prior acceptance reduces but does not eliminate this risk |
| RISK-W07-E02-001 | S002's T8 real-fuzz work discovers a genuine, previously-undetected bug via coverage-guided fuzzing (the entire point of the mechanism working correctly) that requires remediation outside this story's own scope | Medium (a working fuzzer finding a real bug is a sign of success, not failure) | Medium — a genuine bug found by fuzzing needs its own remediation, which may not fit cleanly within S002's own task boundaries | Low-medium | W07-E02-S002 | S002's own task record distinguishes "the fuzz infrastructure works" (this story's own acceptance bar) from "every bug the fuzzer finds is fixed within this story" (not this story's own scope, per mandate §12's own task-boundary discipline) | If a genuine bug is found, file it as its own tracked item (a new task or technical-debt entry) rather than either silently absorbing an open-ended remediation into this story's scope or silently ignoring the finding | unassigned | open | Low — this is the fuzzing infrastructure working as intended, not a defect in this epic's own planning |

## Residual risk after mitigation

RISK-W07-002 is expected to reduce to low-medium residual risk given SEC-01-04's own prior acceptance.
RISK-W07-E02-001 is expected to resolve naturally via this epic's own task-boundary discipline (fuzz
infrastructure vs. bug remediation are tracked separately) — finding a real bug is evidence the
mechanism works, not a planning failure.
