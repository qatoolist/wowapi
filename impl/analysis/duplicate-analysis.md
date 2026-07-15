---
id: ANALYSIS-DUPLICATE
type: analysis
title: Duplicate analysis — CS-layer consolidation mapping and conflict-file overlaps
status: complete
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Duplicate analysis

Per mandate §1.2 ("merge compatible requirements... identify obsolete or superseded guidance... do
not duplicate the same work across multiple waves or stories"). Two distinct layers of duplication
exist in the source material, and this document separates them:

- **(a) Capability-row duplication** — the MATRIX's own dedup pass, mapping 50 raw capability-
  assessment rows (30 §H rows + 20 §I rows) onto 25 consolidated closure specs (CS-01..CS-25). This
  is duplication *within the review/matrix layer itself* — the same underlying capability gap
  observed twice (once from the "complete capability matrix" angle in §H, once from the "mandatory-
  capability readiness" angle in §I).
- **(b) Scope-overlap duplication** — two distinct PLAN findings independently specifying the
  identical closure contract for the same code change. This is duplication *between PLAN findings*,
  documented in full in `conflict-resolution.md` (CONFLICT-01 through CONFLICT-08).

Both layers matter for the same reason: the mandate requires that the same defect not be tracked
three times (once in §H, once in §I, once in PLAN) or fixed twice (once under each of two PLAN
findings). The CS number is the single deduplicated identifier that `requirement-inventory.md`'s
Target column ultimately traces through — several CS specs map onto existing PLAN findings, so a
reader tracing "why is this in Wave X" from a raw §H/§I row goes: **§H/§I row → CS spec → PLAN
finding (or "new") → requirement-inventory.md Target → wave/epic/story**.

---

## (a) CS-layer consolidation mapping

Reproduced from MATRIX §1 ("Dedup mapping — 50 rows → consolidated closure specs"). Every §H row
(H1–H30) and §I row (I1–I20) maps to exactly one CS; overlaps are merged only where the underlying
defect is identical. MATRIX states traceability is total: no row is dropped.

| CS ID | Title | §H rows consolidated | §I rows consolidated | Anchor task(s)/finding(s) |
|---|---|---|---|---|
| CS-01 | Kernel layering & module structure | H1, H20, H27(part) | I1 | FBL-01 |
| CS-02 | Registration model, DI, extensibility, lifecycle manifest | H2, H4, H19 | I2, I15 | AR-01/02/03 |
| CS-03 | Configuration (verify-Ready) | H3 | I3 | — |
| CS-04 | Structured errors (verify-Ready + error-wrapping hygiene) | H5 | I4 | new: lint utilisation |
| CS-05 | Logging ↔ trace correlation & observability | H6, H7 | I5, I6(part) | new FBL-06 |
| CS-06 | Health/readiness migration-currency | H29(part) | I6(part) | DX-07 |
| CS-07 | Identity & session security | H8 | I7 | SEC-01 |
| CS-08 | Validation enforcement path (verify-Ready) | H9 | I8 | — |
| CS-09 | HTTP transport hygiene (timeouts, body limits) | H10, H11 | — | new (candidate) — see CONFLICT-08 for the FBL-09 ownership clarification |
| CS-10 | Data access & pgx resource contract | H12, H13 | I10 | new FBL-05 |
| CS-11 | Jobs, outbox, lease/fencing, drain | H14, H24 | I9, I14 | DATA-02/03 |
| CS-12 | Resilience primitives (retry, breaker, limiter) | H15 | I11 | FBL-04, SEC-04(part) |
| CS-13 | Test infrastructure (e2e isolation, fuzz/race utilisation) | H16 | I12 | T-TEST-01 |
| CS-14 | Generator correctness | H17 | — | DX-02 |
| CS-15 | API contract & compatibility gates | H18, H21 | I13, I18 | DX-06, REL-03 |
| CS-16 | Performance verification programme | H22 | — | PERF-02..05 (§12 constrained) |
| CS-17 | Authz cache bounding & invalidation | H23 | — | SEC-04 |
| CS-18 | Tenant FK integrity | H25 | — | DATA-01 |
| CS-19 | i18n runtime behaviour (verify-retain) | H26 | — | — |
| CS-20 | Audit hash-chain completeness | H28 | I19 | DATA-08 W6 |
| CS-21 | Deployment readiness & seed-sync | H29 | I17 | FBL-02 |
| CS-22 | Documentation gates | H30 | I20 | new: doc-example gate spec |
| CS-23 | Static-analysis & CI-gate utilisation | — | I16 (downgrade candidate) | new FBL-05/07 |
| CS-24 | Outbound-HTTP SSRF guard depth | H8(part), H15(part) | — | verify; new if gap |
| CS-25 | Secrets lifecycle (rotation, provider surface) | H3(part) | I3(part) | verify; new if gap |

Rows H1–H30 / I1–I20 are all accounted for above (several §I "Ready" rows appear as *verify-Ready*
specs per MATRIX's own note: a Ready verdict must be re-earned with evidence, not presumed — this is
why CS-03/CS-08/CS-19 appear in `requirement-inventory.md` table C as verify-outcomes rather than as
net-new implementation work).

**Framing required by the mandate:** this is the "duplicate-analysis" layer because it maps 50 raw
capability-assessment rows onto 25 consolidated closure specs, several of which further map onto
existing PLAN findings (CS-01→FBL-01, CS-02→AR-01/02/03, CS-05→FBL-06, CS-07→SEC-01, CS-10→FBL-05,
CS-11→DATA-02/03, CS-12→FBL-04+SEC-04, CS-14→DX-02, CS-15→DX-06+REL-03, CS-17→SEC-04, CS-18→DATA-01,
CS-20→DATA-08 W6, CS-21→FBL-02) — so the same defect is not tracked three times (once in §H, once in
§I, once in PLAN). Where a CS has no PLAN-finding anchor ("—" or "new" in the Anchor column), MATRIX
itself is flagging either a verify-only outcome (CS-03/CS-08/CS-19) or a genuinely new task that PLAN
did not originally carry (CS-04, CS-06 → already folded into DX-07, CS-09, CS-22, CS-23, CS-24,
CS-25) — these "new" items are the ones `requirement-inventory.md` had to fold into existing findings
(e.g. CS-06 into DX-07, CS-23 into FBL-05/07) or record as their own table-B/table-C items (T-TEST-01
for CS-13, T-DOC-01/FBL-03 lineage for CS-22-adjacent doc work) rather than inventing a 39th PLAN
finding number.

---

## (b) Conflict-file overlaps

The following are duplicate **scope** overlaps between two (or three) PLAN findings — not §H/§I
capability-row duplicates from part (a) above. Each is documented in full, with resolution and
rationale, in `conflict-resolution.md`. Listed here only as pointers, not re-derived:

- **CONFLICT-01** — AR-03 T2 vs DX-06 T1 (OpenAPI full-field merge, identical closure contract).
  Single owner: DX-06. See `conflict-resolution.md`.
- **CONFLICT-02** — PERF-06 T3/T4 vs REL-04 T8 (time-bounded coverage-guided fuzzing, identical
  evidence text and fix). Single owner: REL-04 T8. See `conflict-resolution.md`.
- **CONFLICT-03** — DATA-08 W0-T2 vs DATA-03 T7 (same fix, already implemented once via DATA-08's
  executed Wave-0 slice). DATA-03 T7 excluded as already satisfied. See `conflict-resolution.md`.
- **CONFLICT-04** — PLAN §6 vs PLAN §9 (DX-05 status inconsistency within the same source document,
  not a cross-finding scope overlap, but a duplicate/conflicting *status claim* for the same item).
  Resolved by T-DOC-01, later statement (§9) wins. See `conflict-resolution.md`.
- **CONFLICT-05** — pre-#25 sweep-budget values vs. the #25 recalibration (a temporal-duplication /
  superseded-fact case, not a two-finding scope overlap). #25 (later, honest measurement) wins. See
  `conflict-resolution.md`.
- **CONFLICT-06** — DATA-06 T2 vs DATA-07 T3 (same file/fix). Single owner: DATA-06. See
  `conflict-resolution.md`.
- **CONFLICT-07** — AR-04 T5 vs SEC-06 vs DX-07 T4/T5 (three closure contracts describing the same
  no-op-adapter/readiness-waiver primitive). Shared primitive, built once by AR-04 T5, consumed by
  SEC-06 and DX-07 T4. See `conflict-resolution.md`.
- **CONFLICT-08** — MATRIX §1's CS-09 "new (candidate)" label vs. FBL-09's already-assigned ownership
  of the same HTTP-hygiene capability area (a documentation-clarity conflict, not a genuine
  reassignment). Resolved: FBL-09 ownership stands. See `conflict-resolution.md`.

All eight are fully resolved with no open scope-ownership ambiguity remaining in
`requirement-inventory.md`'s Target column.
