---
id: ANALYSIS-SRC-INV
type: analysis
title: Source inventory — every document consulted, its authority, and its relationships
status: complete
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Source inventory

Per mandate §11.1. Every document that materially fed `requirement-inventory.md`'s allocation, or
that a reader needs to understand provenance, is listed here with: document, version/date, status
(active/superseded/historical), authority, relevant sections, and relationships to other documents.
Classification into PRIMARY / SECONDARY / HISTORICAL follows the mandate's own framing (§1: "at a
minimum, thoroughly review and reconcile" the three primaries; "also inspect all related documents
... referenced by these files or materially affect the implementation").

## PRIMARY — required minimum (mandate §1) plus the directive that defines terms

| Document | Version/date | Status | Authority | Relevant sections | Relationships |
|---|---|---|---|---|---|
| `docs/implementation/fable5-closure-depth-matrix-2026-07-11.md` (MATRIX) | 2026-07-11 | active | Latest and most specific — closure specs CS-01..CS-25 are the dedup layer over §H/§I capability rows; its adjudication log (§3) overturns/accepts prior worker findings with personal verification | §0 spec template, §1 dedup mapping (50→25), §2 closure specs, §2.1 evidence register, §3 adjudication log | Consumes REVIEW §H/§I as raw capability rows; several CS specs map onto PLAN findings (e.g. CS-01→FBL-01, CS-07→SEC-01); `impl/index.md` lists it first in its authority order |
| `docs/implementation/fable5-final-architecture-review-2026-07-11.md` (REVIEW) | 2026-07-11 | active | Final verdict + FBL/D/T items + decision register (§U); the review-of-record for the whole programme | §A verdict, §D claim-verification, §F 10-questions resolution, §H/§I capability matrices, §O task register, §T risk register, §U decisions, §29 22-answers, §30 approval gate | Reviews and corrects the PLAN; feeds the MATRIX's raw rows; `requirement-inventory.md` table B derives from its FBL/D/T items |
| `docs/implementation/premier-framework-implementation-plan.md` (PLAN) | 2026-07-11 | active | Task-level source of record — §5's per-finding task tables (T1..Tn, acceptance, tests, evidence paths) remain authoritative task detail; not superseded by later docs, only status-corrected by them | §4 findings register (38 findings), §5 task breakdown, §6 traceability matrix, §7 cross-cutting risks (14 items), §8 what-was-executed (Wave-0), §9 second-batch executed, §10 explicit accounting | Reviewed and corrected by REVIEW; further consolidated by MATRIX; `requirement-inventory.md` table A is a direct derivation of PLAN §4/§6/§8/§9 |
| `docs/implementation/architecture-directive-2026-07-11.md` (directive) | 2026-07-11 | active | Normative — defines what a finding even means, the severity model, and the phased-implementation blueprint the other three respond to | §4 severity model and release posture, §5-§9 must-refactor findings, §10 must-haves, §11 release-engineering directive, §12 phased blueprint, §13 acceptance matrix, §17 final architectural decision | Upstream of PLAN/REVIEW/MATRIX — they are responses to this directive's findings, not independent sources |

## SECONDARY — consulted, materially affect scope

| Document | Version/date | Status | Authority | Relevant sections | Relationships |
|---|---|---|---|---|---|
| `docs/implementation/decisions.md` | running log, D-0001..D-0090+ | active | Historical record of ratified decisions across the whole programme (Goal 2, hardening, backlog-P2); referenced but not itself a requirement source — it records outcomes, it does not generate new requirements | D-0090 (B11/B12/B13 re-verification, 2026-07-11) is the entry directly relevant here | Referenced by `requirement-inventory.md`'s B11/B12/B13 row; corroborates `framework-backlog-p2-decisions.md`'s parked status with an independent re-measurement |
| `docs/implementation/framework-backlog-p2-decisions.md` | 2026-07-10, re-verified 2026-07-11 per D-0090 | active | The B11/B12/B13 parked-decision record with reopen triggers — authoritative for those three items' disposition (`deferred`, DEF-04..06) | Full document (B11 radix router, B12 schema unification, B13 hot overlays), each with gate/evidence/decision/reopen-trigger | Referenced by `requirement-inventory.md` table C row "B11/B12/B13"; its evidence was independently reproduced by D-0090, not merely re-cited |

## HISTORICAL — context only, not requirement sources for this programme

Retired working files (all retired from the repository 2026-07-11, durable content folded into
`docs/SRS.md` and `docs/GOALS-TRACKER.md`, physically preserved in the `wowapi2` documentation
archive):

| Document | Version/date | Status | Authority | Relevant sections | Relationships |
|---|---|---|---|---|---|
| `Goal.md`, `Goal 1.1.md`, `Goal 1.2.md`, `Goal 2.md`, `goal-test.md` | pre-2026-07-11 | historical | Superseded-by-SRS-and-tracker — these were the AI-prompt/conversation files that seeded the original Goal-2 build | n/a (retired) | Content now lives in `docs/SRS.md`; preserved verbatim in `wowapi2` archive `archive/prompts-and-mandates/` |
| `ROADMAP-wowapi.md` | pre-2026-07-11 | historical | Superseded-by-SRS-and-tracker — was the hardening-tranche roadmap (S/R/E/O items, CA-1..CA-15) | n/a (retired) | Durable content folded into `docs/GOALS-TRACKER.md` §3/§4; preserved in `wowapi2` archive `archive/plans/` |
| `VERIFICATION-wowapi-hardening.md` | pre-2026-07-11 | historical | Superseded-by-SRS-and-tracker — was the hardening closure matrix (source of CA-1..CA-15 in GOALS-TRACKER §4) | n/a (retired) | Preserved in `wowapi2` archive `archive/reviews/` |
| `WOW-Review.md` | pre-2026-07-11 | historical | Superseded-by-SRS-and-tracker; also directly cited inside REVIEW (§F Q2, §G) as a prior review this programme's REVIEW responds to | n/a (retired) | Preserved in `wowapi2` archive `archive/reviews/`; REVIEW's own text still references "WOW-Review §1/§11" as an internal citation to this document's now-archived content |
| Archived evidence bundles: `phase-00/`..`phase-12/`, `hardening-H1/`..`hardening-H5/`+`hardening-P1/`, `wowsociety-gaps/` | closed, archived 2026-07-11 | historical | Closed-goal proof record — required per-bundle files (`proof-bundle.md`, `review-findings.md`, `command-log.md`, `acceptance-map.md`) per `docs/implementation/evidence/README.md` | n/a (archived under `wowapi2/archive/evidence/`) | Referenced by `decisions.md`'s historical entries and by `docs/GOALS-TRACKER.md`'s per-goal tables; not consulted for new requirements in this programme — they prove prior, already-closed goals |

## Canonical status/vision documents (pre-programme authority)

| Document | Version/date | Status | Authority | Relevant sections | Relationships |
|---|---|---|---|---|---|
| `docs/SRS.md` | current | active | Canonical vision for the framework, for everything *before* this implementation programme | whole document | Per the "Status reconciliation" callout in `docs/GOALS-TRACKER.md` §6-preface: this `impl/` programme is the authoritative *outstanding-work* register going forward; SRS remains the vision document it always was |
| `docs/GOALS-TRACKER.md` | current, "Status reconciliation" callout dated 2026-07-11 | active | Canonical status for everything *before* this implementation programme — records completed goals (Goal 2, hardening tranche CA-1..CA-15, backlog B-1..B-9) | §3 hardening tranche, §4 CA-1..CA-15, §5 pending/deferred (all closed), §6 DoD status + Status-reconciliation callout, §7 retired-working-files note | §6-preface explicitly states this `impl/` programme "supersedes" GOALS-TRACKER as the forward-looking backlog: "That programme is the authoritative outstanding-work register; this tracker's per-goal tables remain the record of *completed* goals" |

## Live configuration referenced as an evidence-artifact source (not a planning document)

| Document | Version/date | Status | Authority | Relevant sections | Relationships |
|---|---|---|---|---|---|
| `bench-budgets.txt` | live, 139 lines, recalibrated by session-delta #25 | active | Live perf-gate configuration enforced by `make bench-budget` — not a planning doc, but the evidence-artifact source PERF-01/PERF-06 verification cites | Format header + per-benchmark `max_ns_per_op`/`max_allocs_per_op` rows | Referenced by `requirement-inventory.md` PERF-01 (INV — "#25 recalibrated sweep budgets to honest full-map measurements — verify at current HEAD") and PERF-06 (T1 EXECUTED, enforces this file's contents at CI gate time); see also CONFLICT-05 in `conflict-resolution.md` for the pre-#25-vs-post-#25 recalibration relationship; also feeds CS-16 (performance verification programme) |

## Summary framing

The mandate's minimum-three (MATRIX, REVIEW, PLAN) plus the directive form the PRIMARY tier — all
four are "active" and none is superseded by another; where they overlap, the programme's own rule
("the later/stricter statement wins unless a decision record says otherwise" — `impl/index.md`)
governs, and MATRIX (2026-07-11, latest closure-depth pass) is explicitly the most specific for
capability-area dedup while PLAN §5 remains the task-level source of record per its own unchanged
authority. `decisions.md` and `framework-backlog-p2-decisions.md` are SECONDARY: they don't introduce
new requirements but they do fix disposition for specific rows (B11/B12/B13) and are cited by
`requirement-inventory.md`. Everything else here is HISTORICAL — read for provenance/context only,
never as a source of new, undispositioned requirements for this programme.
