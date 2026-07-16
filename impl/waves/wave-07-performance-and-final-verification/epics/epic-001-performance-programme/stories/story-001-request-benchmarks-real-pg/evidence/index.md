---
id: W07-E01-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W07-E01-S001
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S001 — Evidence index

Per mandate §10. Complete machine-readable records live under `evidence/benchmarks/`; checksums pin each cited artifact.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status | Re-pin addendum (H-8/R-6) |
|---|---|---|---|---|---|---|---|---|
| EV-W07-E01-S001-001 | artifact inspection report (field completeness) | W07-E01-S001-T001 | AC-W07-E01-S001-01 | focused full-field contract test + actionlint | working tree based on `1626b11` | PASS: full field set and workflow | produced | `benchmarks/EV-W07-E01-S001-001-repin-2026-07-16.json` |
| EV-W07-E01-S001-002 | benchmark run report (real PostgreSQL, all 6 profiles) | W07-E01-S001-T002 | AC-W07-E01-S001-02 | exact-env focused request suite | working tree based on `1626b11` | PASS: six real-DB profiles, app_rt/RLS | produced | `benchmarks/EV-W07-E01-S001-002-repin-2026-07-16.json` |
| EV-W07-E01-S001-003 | benchmark run report (concurrency matrix) | W07-E01-S001-T003 | AC-W07-E01-S001-03 | pinned Linux/amd64 container benchmark | working tree based on `1626b11` | PASS: 36/36 cells | produced | `benchmarks/EV-W07-E01-S001-003-repin-2026-07-16.json` |
| EV-W07-E01-S001-004 | attribution report (per component) | W07-E01-S001-T004 | AC-W07-E01-S001-04 | pinned Linux/amd64 container benchmark | working tree based on `1626b11` | PASS: all six components per cell | produced | `benchmarks/EV-W07-E01-S001-004-repin-2026-07-16.json` |
| EV-W07-E01-S001-005 | published relative/container report | W07-E01-S001-T005 | AC-W07-E01-S001-05 | pinned Go/PostgreSQL container capture | working tree based on `1626b11` | PASS: relative ratios; absolute conditional on DEC-Q9 | produced | `benchmarks/EV-W07-E01-S001-005-repin-2026-07-16.json` |

Fresh passing evidence uses index status `produced`; the failure/supersession vocabulary remains applicable to retained failed or replacement records.

**Revision re-pin note (autopsy finding H-8, remediation R-6, 2026-07-16):** the `1626b11` pin above
is not an ancestor of current HEAD (a side effect of the e8cda6b squash). Per
`impl/governance/evidence-policy.md`'s revision-pinning rule, each record above has been re-run
against current HEAD (`43b6e128672f0b0997adcebc92703884deba5684`); the results are recorded, without
overwriting these original rows, in the sibling `*-repin-2026-07-16.json` addendum files listed above
(status `retested`). No divergence from the original claims was found.
