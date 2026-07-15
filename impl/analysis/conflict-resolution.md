---
id: ANALYSIS-CONFLICT-RESOLUTION
type: analysis
title: Conflict resolution — duplicate-scope and inconsistent-guidance conflicts across PLAN/REVIEW/MATRIX
status: complete
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Conflict resolution

Per mandate §1.2 ("document conflicts... do not duplicate the same work across multiple waves or
stories"). Eight conflicts are documented: the six given by the task (CONFLICT-01..06, numbering
preserved exactly as given, not renumbered), plus two further genuine conflicts found by scanning
PLAN §7 (cross-cutting risks/assumptions, 14 numbered items), REVIEW §T (risk register), REVIEW §U
(decision register), REVIEW §F (unresolved-questions), and MATRIX §3 (adjudication log) —
CONFLICT-07 (given by the task as a known seventh item to add) and CONFLICT-08 (independently
derived). No further genuine conflicts were found beyond these eight; see the closing note.

Format per item: ID | Items in conflict | Nature of conflict | Sources | Resolution | Rationale |
Status.

---

## CONFLICT-01

- **Items in conflict:** AR-03 T2 vs DX-06 (T1)
- **Nature of conflict:** Duplicate scope — both findings independently specify the identical closure
  contract for OpenAPI full-field merge, complete-or-loud behaviour.
- **Sources:** PLAN §7 cross-cutting note #11 ("AR-03 T2 and DX-06 T1 (OpenAPI full-field merge) —
  identical closure contract"); PLAN §6 traceability matrix (DX-06 row: "overlaps AR-03 T2, assign
  one owner"); `requirement-inventory.md` AR-03 row notes.
- **Resolution:** Single owner DX-06. AR-03's target story (W05-E03-S001..S002) proceeds without T2;
  DX-06 (W06-E02-S001) owns the merge-completeness work in full.
- **Rationale:** PLAN §7 explicitly flags this as a duplicate-effort risk requiring single ownership
  before implementing twice; DX-06 is the more specific/scoped finding (3-task OpenAPI-only scope vs.
  AR-03's broader 5-task declaration/projection scope), so it is the natural single owner.
- **Status:** resolved.

## CONFLICT-02

- **Items in conflict:** PERF-06 T3/T4 vs REL-04 T8
- **Nature of conflict:** Duplicate scope — identical fuzz scope (time-bounded coverage-guided
  fuzzing), identical evidence text and fix per PLAN §7.
- **Sources:** PLAN §7 cross-cutting note #12 ("PERF-06 T3/T4 and REL-04 T8 (time-bounded
  coverage-guided fuzzing) — identical evidence text and fix"); PLAN §9 REL-04 section ("T7/T8 ...
  explicitly overlaps and is deferred to PERF-06 T3/T4's single-ownership resolution"); PLAN §6
  traceability matrix REL-04 notes.
- **Resolution:** Single owner REL-04 T8 (target W07-E02-S002, per its "T8 owns fuzz, shared w/
  PERF-06 T3/T4" note). PERF-06's target story (W00-E01-S002) proceeds without T3/T4.
- **Rationale:** REL-04's own PLAN §9 text already defers to PERF-06's resolution, and
  `requirement-inventory.md`'s REL-04 row states the single-owner assignment explicitly — the
  programme adopts that pre-existing assignment rather than re-deciding it.
- **Status:** resolved.

## CONFLICT-03

- **Items in conflict:** DATA-08 W0-T2 vs DATA-03 T7
- **Nature of conflict:** Same underlying fix, already implemented once. DATA-08's Wave-0 slice
  (W0-T1+W0-T2) was EXECUTED per PLAN §8; DATA-03 T7 is a separate finding's task that targets the
  identical defect.
- **Sources:** PLAN §8 ("DATA-08 — Wave-0 slice (W0-T1+W0-T2) — EXECUTED, independently reviewed,
  PASS"); `requirement-inventory.md` DATA-03 row notes ("T7 = DATA-08 W0-T2 duplicate (done)").
- **Resolution:** DATA-03's task breakdown (target W04-E02-S001..S002) excludes T7 as already
  satisfied by the executed DATA-08 W0-T2 slice; the W00 verification wave re-confirms the shared
  evidence covers both findings' claims over that code path.
- **Rationale:** Re-implementing an already-shipped, independently-reviewed fix would be pure waste
  and would also risk clobbering the existing, verified DATA-08 change; the correct action is to
  point DATA-03 T7 at the same evidence rather than re-deriving it.
- **Status:** resolved.

## CONFLICT-04

- **Items in conflict:** PLAN §6 (traceability matrix) vs PLAN §9 (second batch, "what was executed")
  — both describe DX-05's status, but disagree.
- **Nature of conflict:** Status inconsistency within the same source document. §6 marks DX-05 as
  PLANNED (flat table, no execution annotation). §9 records AR-05 T1/T2 + DX-05 T1/T2
  documentation-drift fixes as EXECUTED and independently reviewed twice.
- **Sources:** PLAN §6 traceability matrix; PLAN §9 "AR-05 T1/T2 + DX-05 T1/T2 — documentation drift
  fixes — EXECUTED, independently reviewed twice (PASS both times)".
- **Resolution:** T-DOC-01 (a `requirement-inventory.md` table B row) exists specifically to fix this
  §6-vs-§9 status inconsistency in the plan document itself. The prose in §9 — later in document
  order and more specific (it names exact files changed and the two independent review passes) — wins
  over the flat table in §6, per the programme's stated rule: "where sources overlap, the
  later/stricter statement wins" (`impl/index.md`). `requirement-inventory.md`'s own DX-05 row already
  reflects this correctly: disposition `partial`, notes "T1/T2 EXECUTED; §6-vs-§9 status
  inconsistency = T-DOC-01".
- **Rationale:** A flat status table generated early in a large document is more prone to going stale
  than a narrative execution record written after the fact describing exactly what shipped and how it
  was verified; the later, more detailed statement is the more trustworthy one, and T-DOC-01 exists to
  correct the table itself so future readers of PLAN §6 aren't misled.
- **Status:** resolved.

## CONFLICT-05

- **Items in conflict:** Pre-#25 sweep-budget values (as they existed before session-delta #25) vs.
  the #25 recalibration.
- **Nature of conflict:** Numeric/factual conflict — the token-bucket sweep-budget values referenced
  anywhere in PLAN/REVIEW/MATRIX prose predate a code fix and honest re-measurement; the code and
  `bench-budgets.txt` have since moved past those numbers.
- **Sources:** `requirement-inventory.md` §E SD-03 ("Sweep-bench O(n²)+empty-map fix; budgets
  recalibrated"); `impl/index.md` (planning HEAD `0a31186` is after #22–#25 merges); `bench-budgets.txt`
  (live, post-#25 values).
- **Resolution:** #25 wins — it is later and reflects an honest full-map measurement, superseding any
  earlier budget numbers referenced anywhere in PLAN/REVIEW/MATRIX. PERF-01's W00-E01-S002
  verification story must verify against the NEW post-#25 budgets in `bench-budgets.txt`, not any
  value quoted in the three primary documents (which predate #25 in git history even though the
  inventory/plan already accounts for the code being current).
- **Rationale:** The three primary documents' prose may still cite pre-#25 numbers as historical
  context, but the actual enforcement artifact (`bench-budgets.txt`) and the actual code (the O(n²)
  fix) are already current at planning HEAD — treating a stale prose number as authoritative over the
  live enforced config would be backwards.
- **Status:** resolved.

## CONFLICT-06

- **Items in conflict:** DATA-06 T2 vs DATA-07 T3
- **Nature of conflict:** Duplicate scope — same file/fix. Both findings' task breakdowns independently
  target the identical resource-mirror aggregate write-contract defect.
- **Sources:** `requirement-inventory.md` DATA-06 row notes ("T2 shared fix w/ DATA-07 T3 (one
  owner)").
- **Resolution:** Single owner DATA-06 (target W02-E04-S001). DATA-07's task breakdown (target
  W03-E04-S001) excludes T3 as covered by DATA-06 T2.
- **Rationale:** DATA-06 is sequenced earlier (W02, no upstream dependency) than DATA-07 (W03, HARD
  dep on SEC-01) — assigning ownership to the earlier-landing finding means the shared fix is
  available sooner and DATA-07 simply consumes it rather than both findings racing to implement the
  same file change independently at different points in the programme.
- **Status:** resolved.

## CONFLICT-07

- **Items in conflict:** AR-04 T5 vs SEC-06 vs DX-07 T4/T5
- **Nature of conflict:** Duplicate scope, but of a different shape than CONFLICT-01/02/06 — not two
  findings racing to fix the same file, but **three closure contracts each independently describing
  the same underlying primitive**: "a no-op adapter fails readiness in prod without a waiver."
- **Sources:** PLAN §7 cross-cutting note #13 ("AR-04 T5, SEC-06, and DX-07 T4/T5 (the 'no-op adapter
  fails readiness in prod without a waiver' mechanism) — three closure contracts describing the same
  waiver/readiness primitive"); `requirement-inventory.md` AR-04 row ("T5 waiver shared w/
  SEC-06/DX-07"); `requirement-inventory.md` DX-07 row ("T4 dep AR-04 T5 waiver mechanism").
- **Resolution:** Unlike CONFLICT-01/02/03/06 (exclusive single ownership — one finding does the work,
  the other's task is dropped), this is a **shared primitive**: AR-04 T5 builds the shared
  waiver/readiness-failure primitive once (target W05-E03-S002). SEC-06 (target W03-E02-S001) and
  DX-07 T4 (target W04-E04-S003) *consume* that primitive rather than re-implementing it — their own
  task breakdowns keep their scope but depend on AR-04 T5 rather than each building their own waiver
  mechanism.
- **Rationale:** SEC-06 (outbound-security escape-hatch governance) and DX-07 (truthful
  readiness/config diagnostics) are genuinely different findings with different closure bars beyond
  the shared primitive — dropping either finding's task entirely (as with the exclusive-ownership
  conflicts) would lose real, distinct scope. The correct resolution is "build once, consume twice,"
  not "own once, drop the duplicate."
- **Status:** resolved.

## CONFLICT-08

- **Items in conflict:** MATRIX §1 dedup mapping's CS-09 "new (candidate)" anchor-task designation vs.
  REVIEW/PLAN's assignment of the same underlying HTTP-hygiene defect to FBL-09 as an established
  finding.
- **Nature of conflict:** Contradictory-guidance conflict (not duplicate scope) — MATRIX §1 lists CS-09
  ("HTTP transport hygiene (timeouts, body limits)", consolidating §H rows H10/H11) with anchor task
  "new (candidate)", implying the matrix's own dedup pass did not yet know of an assigned owner
  finding when it wrote that row. But `requirement-inventory.md` (drawing on REVIEW §H/§I and PLAN)
  already assigns this exact capability area to FBL-09 ("HTTP server timeouts + CSRF body bound...
  CS-09"), with a fully-specified task target (W01-E03-S001). If read literally, MATRIX's "new
  (candidate)" phrasing could suggest FBL-09 doesn't yet exist as an owner, when in fact it already
  does in the inventory that consolidates all three documents.
- **Sources:** MATRIX §1 dedup mapping (CS-09 row, anchor tasks column: "new (candidate)");
  `requirement-inventory.md` FBL-09 row ("CS-09; template-delivery model (wowsociety backport =
  PROD-03)"); REVIEW §H (H10/H11, HTTP hygiene) and §I (mandatory-capability readiness).
  Checked for: whether REVIEW's severity/priority for any finding disagrees with PLAN's (found none —
  every cross-referenced finding's priority is consistent between REVIEW §T/§O and PLAN §4/§6); and
  whether MATRIX's CS dedup mapping (§1) implies a different owner than PLAN/REVIEW assign for the
  same defect (found this one case, CS-09/FBL-09).
- **Resolution:** `requirement-inventory.md`'s FBL-09 assignment stands as the actual owner — MATRIX's
  "new (candidate)" notation is read as "this closure spec required a *new* task where none existed in
  the original 38 PLAN findings," which REVIEW then supplied by elevating it to a named finding
  (FBL-09) in its own §O task register, not as MATRIX overriding or contradicting that assignment.
  MATRIX (dated 2026-07-11, same day as REVIEW) is later in the document pipeline but does not itself
  reassign CS-09 away from FBL-09 anywhere in §2 (closure specifications) — it is silent on ownership
  beyond the §1 summary column, so there is no actual textual contradiction, only an ambiguous label
  that a naive reader could misparse as one. No inventory change is required; this entry exists so the
  ambiguity is recorded and future readers of MATRIX §1 in isolation are not misled into thinking
  CS-09 is unowned.
- **Rationale:** The three-document reconciliation rule ("later/stricter wins") only applies where
  there is a genuine disagreement about disposition or ownership; here there is no disagreement once
  MATRIX §2's full closure spec is read (it does not repeat or contradict the "new" label from §1) —
  this is a documentation-clarity issue in MATRIX §1's summary table, not a planning conflict requiring
  re-adjudication.
- **Status:** resolved.

---

## Further-conflict scan — result

PLAN §7 cross-cutting section (14 numbered items) was reviewed in full: items #1-#10 are the
genuinely-undecided/human-decision items (already tracked as DEC-Q1/Q9/Q10 and D-01..D-09, not
conflicts), item #14 is a sequencing-constraint list (not a conflict — dependency edges, already
reflected in `requirement-inventory.md`'s Notes columns and the wave map), and items #11/#12/#13 are
exactly CONFLICT-01/02/07 above. REVIEW §T (risk register) and §U (decision register) were checked
line-by-line against PLAN §4/§6's severity/priority assignments for every finding they both reference
(SEC-01, DATA-01, DATA-08, FBL-01, FBL-02) — no disagreement found; severities are consistent across
both documents. MATRIX §3's adjudication log (5 entries) was checked for any adjudication that
reassigns ownership away from an existing PLAN/REVIEW finding — only the DX-02/PF-2 entry does
substantive reassignment work, and it *confirms* DX-02 as real and correctly attributed (an
"OVERTURNED a worker refutation" entry, not a conflict with PLAN/REVIEW's own DX-02 assignment).
MATRIX §1's dedup table was checked row-by-row against PLAN's finding-to-anchor-task mapping (see
`duplicate-analysis.md` §a for the full table) — only CS-09 showed an ambiguous "new (candidate)"
label warranting CONFLICT-08's clarifying entry.

**No further genuine conflicts found beyond the eight documented above.**
