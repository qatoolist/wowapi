---
id: W06-E02-RISKS
type: epic-risks
epic: W06-E02
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E02 — Risks

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W06-E02-001 | DX-06 T2's OpenAPI 3.1 validator dependency decision (`pb33f/libopenapi` candidate) is made without the security/licence review MATRIX CS-15's own risk note requires, if the implementation-time task is rushed | Medium | Medium — an unreviewed third-party validator in the OpenAPI-merge closure path could introduce a supply-chain or licensing issue in a release-adjacent tool | Medium | W06-E02-S001 | S001's own task record makes the security/licence review an explicit, separately-checkable sub-step of the validator-decision task, not an implicit assumption | If the evaluated candidate fails review, record the rejection and the alternative chosen as a deviation, not a silent substitution | unassigned | open | Low once the review step is honored |
| RISK-W06-E02-002 | S003's three REL-03b legs may remain blocked past this epic's own closure attempt if their unblocking stories (S001 within this epic; W06-E01-S001/S002; W05-E03) land later than expected | Medium | Medium — a still-blocked leg at closure drives a `partially-accepted` disposition for S003, not a full block on this epic's other two stories | Medium | W06-E02-S003 | S003's own `story.md` records explicit per-leg blocked-entry criteria, making a partial-acceptance disposition honestly recordable rather than forcing a false "all blocked/all done" binary | Record any still-blocked leg in `closure-report.md` with its unblocking condition restated | unassigned | open | Low-medium — this is a scheduling risk, not a design gap; every leg's unblocking condition is already fully specified |

## Residual risk after mitigation

RISK-W06-E02-001 is expected to reduce to low residual risk once S001's review step is honored as
planned. RISK-W06-E02-002 is a scheduling risk expected to resolve naturally as S003's three unblocking
stories land in their own course, tracked honestly via partial-acceptance status if any does not by this
epic's own closure attempt.
