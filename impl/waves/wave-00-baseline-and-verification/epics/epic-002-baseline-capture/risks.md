---
id: W00-E02-RISKS
type: epic-risks
epic: W00-E02
wave: W00
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W00-E02 — Risks

Epic-scoped subset and elaboration of the wave-level register (`../../risks.md`). IDs are shared
with the wave-level register where the risk is identical (per naming-conventions.md — never reuse
or renumber an identifier); this file adds epic-specific detail and affected-story granularity.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W00-003 | Bench-budget baseline captured against stale (pre-#25) budgets if the sweep-bench recalibration (SD-03) is not correctly reflected at the commit this epic runs against | Low | Medium — later waves' "improvement over baseline" perf claims measured against the wrong starting point | Medium | W00-E02-S001 (T003) | Explicitly confirm `bench-budgets.txt` entry count (43) and values match the post-#25 state before capturing the baseline artifact; record the confirming diff in the evidence record | Re-capture once confirmed correct; mark the earlier capture `superseded`, not deleted (evidence-policy.md) | unassigned | open | Low |
| RISK-W00-004 | ADR-ification (S003) inadvertently introduces new design content beyond what D-01..D-09 already state in REVIEW §F/§U, silently resolving an ambiguity the mandate requires to stay explicit | Low | Medium — would violate mandate §18 ("do not silently resolve ambiguous architecture decisions") | Medium | W00-E02-S003 (all 3 tasks) | Each ADR cites its REVIEW §F/§U source verbatim for recommendation/safe-default/owner; any elaboration beyond the source is flagged as a Wave-00-added clarification, never folded in as if original | Independent review checks each ADR line-by-line against its REVIEW §F/§U source before the story moves to `accepted` | unassigned | open | Low |
| RISK-W00-005 | CI wall-clock / coverage baseline captured without correctly accounting for SD-01/SD-02's pipeline changes (parallelized gate, path-scoped bench), producing a baseline that misdescribes current CI shape | Low | Low-medium — affects descriptive accuracy of the baseline only, not correctness of later waves' work | Low | W00-E02-S001 (T001, T003) | Capture baseline against the current `.github/workflows/ci.yml` (3-leg parallelized, docs-only skip) explicitly; note SD-01/SD-02 facts in the evidence record | Re-capture if a discrepancy is found between the recorded baseline and the actual workflow file at capture time | unassigned | open | Low |
| RISK-W00-E02-001 | S002's approved-dependency cross-check misses a drift because `go list -m all` output is not diffed line-by-line against REVIEW §L's named list, only spot-checked | Low | Medium — an unapproved or license-incompatible dependency could go unflagged | Medium | W00-E02-S002 (T001) | Cross-check is a full enumeration, not a sample: every direct dependency in `go.mod` must appear in REVIEW §L's approved list or be flagged as new/undocumented drift requiring explicit disposition | Flag any undocumented dependency as a new finding requiring its own disposition before the story can move to `accepted` | unassigned | open | Low |

## Notes

RISK-W00-001 and RISK-W00-002 (wave-level) apply to W00-E01, not W00-E02 — this epic performs no
test re-execution against claimed-EXECUTED slices, so those two wave-level risks are out of scope
here and are not duplicated into this file.
