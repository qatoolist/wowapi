---
id: TRACK-REQ-TRACE-MATRIX
type: matrix
title: Requirement traceability matrix — requirement to wave/epic/story to AC/task/artifact/evidence
status: active
created_at: 2026-07-12
updated_at: 2026-07-12
derived: true
---

# Requirement traceability matrix

DERIVED VIEW. Per mandate §11.5: `Requirement → Wave → Epic → Story → Acceptance criterion →
Task → Artifact → Evidence → Final result`. Canonical source today =
`impl/analysis/requirement-inventory.md` Target column; once story files exist, the AC/Task/
Artifact/Evidence columns become canonical from each story's own front matter and index files.

## AC-ID convention

Acceptance criteria IDs follow `AC-<story-id>-NN` (e.g. `AC-W05-E01-S001-01`), defined in each
story's own `story.md` once created — not yet enumerated here (no story files exist at this
planning stage). Task IDs follow `<story-id>-T<NNN>`, registered in the story's own
`tasks/index.md`. Artifact and evidence IDs are registered in the story's own
`artifacts/index.md` and `evidence/index.md` respectively (mandate §9.2/§10).

## Scope of this table

Covers every `requirement-inventory.md` row with disposition `planned`, `partial`, `blocked→planned`,
or `implemented-needs-verification` (INV) — i.e. every row that maps to an implementation wave.
Rows with disposition `deferred`, `rejected`, `not-applicable`, `duplicate`, `superseded`, or
`blocked (human, tracked-only)` are excluded here; see the exclusion note at the bottom and
`findings-disposition.md` for their full disposition.

| Requirement | Wave | Epic | Story | AC | Task | Artifact | Evidence | Final result |
|---|---|---|---|---|---|---|---|---|
| AR-01 | W05 | W05-E01 | W05-E01-S001..S004 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| AR-02 | W05 | W05-E02 | W05-E02-S001..S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| AR-03 | W05 | W05-E03 | W05-E03-S001..S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| AR-04 | W05 | W05-E03 | W05-E03-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — T1 already EXECUTED (verified ×2); T2–T5 pending |
| AR-05 | W06 | W06-E04 | W06-E04-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — T1/T2 already EXECUTED; T3–T5 pending |
| AR-06 | W05 | W05-E04 | W05-E04-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — T1 already EXECUTED; T2/T3 pending |
| SEC-01 | W03 | W03-E01 | W03-E01-S001..S004 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — DEC-Q1 safe default unblocks build |
| SEC-02 | W03 | W03-E05 | W03-E05-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — T1–T3 already EXECUTED (verified ×2); T4/T5 pending |
| SEC-03 | W03 | W03-E03 | W03-E03-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| SEC-04 | W05 | W05-E04 | W05-E04-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| SEC-05 | W07 | W07-E02 | W07-E02-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| SEC-06 | W03 | W03-E02 | W03-E02-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| DATA-01 | W02 | W02-E02 | W02-E02-S001..S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — dep DATA-09 T1–T5 |
| DATA-02 | W04 | W04-E01 | W04-E01-S001..S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| DATA-03 | W04 | W04-E02 | W04-E02-S001..S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| DATA-04 | W04 | W04-E03 | W04-E03-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| DATA-05 | W02 | W02-E03 | W02-E03-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| DATA-06 | W02 | W02-E04 | W02-E04-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| DATA-07 | W03 | W03-E04 | W03-E04-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — hard dep SEC-01 |
| DATA-08 | W04 | W04-E04 | W04-E04-S001..S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — W0 slice EXECUTED (verified ×2); W6-T1 + T2–T5 pending |
| DATA-09 | W02 | W02-E01 | W02-E01-S001..S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| DX-01 | W01 | W01-E04 | W01-E04-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — T5 harness shared w/ DX-02/DX-04 |
| DX-02 | W01 | W01-E04 | W01-E04-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — re-verified at HEAD (CS-14) |
| DX-04 | W06 | W06-E01 | W06-E01-S002 | AC-01..AC-05 pass in verification.md | T001..T006 done | ART-W06-E01-S002-001..005 current, complete, and content-pinned | EV-W06-E01-S002-001..014 registered; EV-014 final independent PASS | accepted 2026-07-14 |
| DX-05 | W01 | W01-E04 | W01-E04-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — T1/T2 already EXECUTED; §6-vs-§9 inconsistency = T-DOC-01 |
| DX-06 | W06 | W06-E02 | W06-E02-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — single owner of AR-03 T2 scope |
| DX-07 | W04 | W04-E04 | W04-E04-S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — T4 dep AR-04 T5 |
| PERF-01 | W00 | W00-E01 | W00-E01-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending W00 re-verification — EXECUTED, #25 recalibrated budgets |
| PERF-02 | W07 | W07-E01 | W07-E01-S001 | AC-01..AC-05 accepted in closure.md | T001..T006 complete | ART-W07-E01-S001-001..005 produced | EV-W07-E01-S001-001..005 registered; clean review | accepted 2026-07-14 — relative/container scope; absolute SLO conditional on DEC-Q9 |
| PERF-03 | W07 | W07-E01 | W07-E01-S002 | AC-01..AC-06 accepted in closure.md | T001..T007 done | ART-W07-E01-S002-001..006 produced | EV-W07-E01-S002-001..007 registered; clean review | accepted 2026-07-14 |
| PERF-04 | W07 | W07-E01 | W07-E01-S003 | AC-01..AC-07 accepted in closure.md | T001..T009 complete | ART-W07-E01-S003-001..008 produced | EV-W07-E01-S003-001..008 registered; clean review | accepted 2026-07-14 — absolute SLO conditional on DEC-Q9 |
| PERF-05 | W07 | W07-E01 | W07-E01-S004 | AC-01..AC-07 accepted in closure.md | T001..T007 complete | ART-W07-E01-S004-001..006 produced | EV-W07-E01-S004-001..007 registered; clean review | accepted 2026-07-14 — behavioral/budget proof; absolute SLO conditional on DEC-Q9 |
| PERF-06 | W00 | W00-E01 | W00-E01-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending W00 re-verification — T1 EXECUTED; T3/T4 fuzz owned by REL-04 T8 (W07-E02-S002) |
| REL-01 | W06 | W06-E03 | W06-E03-S001..S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — ~85% buildable now; final activation = DEC-Q10 |
| REL-02 | W06 | W06-E03 | W06-E03-S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| REL-03 | W06 | W06-E02 | W06-E02-S002..S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — a-leg (T1,T2,T4,T6,T8,T9) now; b-leg (T3/T5/T7) dep DX-06/AR-03/DX-03/DX-04 |
| REL-04 | W07 | W07-E02 | W07-E02-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — T1–T4 EXECUTED (verified ×2); T5–T8 pending |
| FBL-01 | W05 | W05-E05 | W05-E05-S001..S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — dep AR-01/AR-02 |
| FBL-02 | W02 | W02-E05 | W02-E05-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| FBL-03 | W01 | W01-E04 | W01-E04-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| FBL-04 | W04 | W04-E02 | W04-E02-S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| FBL-05 | W01 | W01-E01 | W01-E01-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| FBL-06 | W01 | W01-E02 | W01-E02-S001..S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| FBL-07 | W01 | W01-E01 | W01-E01-S002..S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — nightly ci schedule EXISTS since #24, fuzz portion still seed-replay only |
| FBL-08 | W01 | W01-E03 | W01-E03-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| FBL-09 | W01 | W01-E03 | W01-E03-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| D-01..D-09 | W00 | W00-E02 | W00-E02-S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — decisions ratified in REVIEW; ADR-ification pending |
| T-DOC-01 | W01 | W01-E04 | W01-E04-S002 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending |
| T-TEST-01 | W01 | W01-E04 | W01-E04-S003 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — reproduce-first; original shared-DB diagnosis withdrawn |
| CS-10 | W01 | W01-E01 | W01-E01-S001 | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — decided constraint (keep raw pgx.Rows), enforcement via FBL-05 |
| CS-25 | — | — | (D-09 documentation) | defined in story file | defined in story's tasks/ index | registered in story's artifacts/index.md | registered in story's evidence/index.md | pending — restart-based rotation documented; file-provider deferred (DEF-01) |

## Excluded rows and why

Rows excluded from the table above because they do not map to an implementation wave (disposition
`deferred`, `rejected`, `not-applicable`, `duplicate`, or `blocked (human, tracked-only)` with no
story target of its own):

- **DX-03** — disposition `deferred` (design-investigation story only, Wave-4-class per plan); still
  has a nominal target `W06-E01-S001` for the design-investigation story itself, but the DSL build-out
  is deferred beyond this programme. See `findings-disposition.md`.
- **DEC-Q1, DEC-Q9, DEC-Q10** — disposition `blocked (human)`, tracked against a wave/epic but not a
  story (they gate stories, they are not implementation stories). See `decision-register.md`.
- **CS-03, CS-19, CS-24** — disposition `INV→verified` with no new work; recorded as W00 evidence
  pointers only, not stories. See `findings-disposition.md`.
- **K-RETAIN** — disposition `not-applicable`; justified retentions, no work planned.
- **K-P2** — disposition `deferred`; see `deferred-items-register.md` DEF-02/DEF-03.
- **M-REJ** — disposition `rejected`; rationale in REVIEW §M, no implementation item.
- **B11/B12/B13** — disposition `deferred`; see `deferred-items-register.md` DEF-04/DEF-05/DEF-06.
- **PROD-01..05** — product-level items, excluded from framework implementation per mandate §2.3;
  see `source-traceability-matrix.md` product-boundary rows.
- **SD-01..SD-04** — informational session-delta facts, not requirements; see
  `source-traceability-matrix.md` session-delta rows.

Full disposition for every excluded ID is recorded in `impl/analysis/findings-disposition.md`.
