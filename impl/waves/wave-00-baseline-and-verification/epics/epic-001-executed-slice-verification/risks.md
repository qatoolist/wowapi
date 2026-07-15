---
id: W00-E01-RISKS
type: epic-risks
epic: W00-E01
wave: W00
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W00-E01 — Risk register

Per mandate §11.7. These risks specialize the wave-level register (`../../risks.md`) to this epic's
three stories; RISK IDs are inherited from the wave-level register where the risk is identical in
substance (mandate §5: never mint a duplicate ID for the same risk under a new scope).

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W00-001 | A finding-slice claimed EXECUTED for SEC-02/AR-04/AR-06/PERF-01/PERF-06/DATA-08/REL-04 fails to re-verify at current HEAD | Low-medium | High — blocks the downstream waves listed in `dependencies.md` | High | S001, S002, S003 (all 9 tasks) | Re-run the exact named test files/commands, not a paraphrase; register failing evidence per `evidence-policy.md` rather than silently re-trying until green | Open a new remediation task under the owning story; do not mark the story `accepted` with a known regression (see `epic.md` epic acceptance criteria AC-W00-E01-01) | unassigned | open | Some — a slice could be flaky under CI conditions not reproduced locally |
| RISK-W00-002 | Test infrastructure (Postgres, MinIO) unavailable in the execution environment, producing a false-negative "fail" that looks like a genuine regression | Medium | Medium — wastes investigation time, risks a false regression being escalated | Medium | S003 (DATA-08 fault injection, REL-04 S3/TOTP); secondarily S001 if its DB-backed tests need testkit Postgres | Confirm `make ci-container`/`docker compose` health before treating any failure as genuine; capture environment state as part of the evidence record | Re-run in a known-good environment (e.g., CI itself) before concluding a genuine regression | unassigned | open | Low once mitigated |
| RISK-W00-003 | Bench-budget baseline (S002) captured against stale, pre-#25 values if the sweep-bench recalibration (SD-03) is not correctly reflected at the commit this epic runs against | Low | Medium — later waves' perf-improvement claims measured against the wrong starting point | Medium | S002 | Explicitly confirm `bench-budgets.txt` entry count and values match the post-#25 state (43 budgeted entries) before treating S002's evidence as valid | Re-capture once confirmed correct; mark the earlier evidence record `superseded`, not deleted | unassigned | open | Low |
| RISK-W00-E01-004 | AR-05 scope conflict between `wave.md`/`epics/index.md` (claims W00-E01 covers AR-05 T1/T2) and `requirement-inventory.md` (canonically targets AR-05 to W06-E04-S002) is left unresolved, causing this epic to be closed with an ambiguous scope boundary | Medium | Low-medium — does not block any of the 9 defined tasks, but leaves AC-W00-E01-03 unsatisfiable until resolved | Medium | Epic-level closure (AC-W00-E01-03) | Flagged explicitly in `epic.md` "Out of scope" and this epic's creation report rather than silently resolved; acceptance authority must rule on it before epic acceptance | Escalate to the acceptance authority; if unresolved by the time S001-S003 are otherwise ready to close, epic moves to `partially-accepted` with the gap stated in `closure-report.md`, not `accepted` | unassigned | open | None once resolved — this is a documentation-consistency risk, not a technical one |

## Notes

RISK-W00-001, -002, -003 are the wave-level risks (`../../risks.md`) as they apply to this epic's
scope; they are not epic-specific re-inventions. RISK-W00-E01-004 is genuinely epic-specific (it did
not exist at wave-level because the wave-level risk register predates this epic's detailed task
content being reconciled against `requirement-inventory.md`) and is scoped with the epic's own ID
suffix per `naming-conventions.md`.
