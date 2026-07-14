---
id: TRACK-CHANGE-LOG
type: register
title: Change log — programme-structure changes (append-only)
status: active
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Change log

This is the one **non-derived** file in `impl/tracking/`: it is itself the source of record for
programme-structure changes. Records every programme-structure change (new/removed waves, epics,
stories; governance-policy changes; register-schema changes). Does not record per-story
implementation progress — that lives in each story's own files and rolls up via
`status-register.md`.

Reverse-chronological (newest first).

## 2026-07-13 — W01 executed and accepted

- Wave 01 executed and accepted 2026-07-13: all 10 stories (W01-E01-S001..S003, W01-E02-S001..S002,
  W01-E03-S001..S002, W01-E04-S001..S003) accepted; ~40 production files touched.
- Independent review gate passed: reviewer W01ReviewGate (independent reviewer agent); accepted by
  conductor 2026-07-13.
- Conductor integration fixes applied post-review: e2e harness --local-framework wiring
  (internal/e2e/e2e_test.go), init-template query-tracer wiring (database.WithQueryTracer in both
  api/worker templates), blueprint-11 init example documents --framework-version/--local-framework,
  stray root mandate file relocated to impl/premier-framework-implementation-programme-mandate.md.

## 2026-07-13 — W00 executed and accepted
- Wave 00 executed and accepted 2026-07-13: all 6 stories (W00-E01-S001..S003, W00-E02-S001..S003)
  accepted with commit-pinned evidence registered in each story's `evidence/index.md`.
- 9 ADRs (D-01..D-09) ratified; `decision-register.md` rows moved to `ratified` with ADR paths;
  `deviation-register.md` updated.
- Independent review gate passed: reviewer W00ReviewGate (independent reviewer agent); accepted
  by conductor 2026-07-13.
- Conductor adjudications: DEV-W00-E01-S001-002 (AC-04 re-scoped to executed AR-05 T1/T2 slice;
  future-state blueprint hits routed to AR-05 T5 / W06-E04-S002) and DEV-W00-E02-S001-001
  (lint-baseline pass-as-capture ratified; 18-of-25 analyzer-name gap acknowledged).

## 2026-07-12 — Implementation programme created

- Source documents reconciled: architecture-directive-2026-07-11.md, fable5-final-architecture-review-2026-07-11.md,
  fable5-closure-depth-matrix-2026-07-11.md, premier-framework-implementation-plan.md (+ decisions.md,
  framework-backlog-p2-decisions.md as secondary sources).
- Repository state at planning time: HEAD = 0a31186 (post #22-#25 merges).
- impl/ directory structure established per premier-framework-implementation-programme-mandate.md.md:
  governance/ (+ templates/), analysis/, tracking/ populated with full derived content; waves/ not yet
  populated (wave-by-wave execution begins after this programme is accepted).
- 38 PLAN findings + 15 REVIEW findings/decisions + 9 MATRIX verify-outcomes + 5 PROD-boundary items +
  4 session-delta facts allocated across 8 waves (W00-W07) per requirement-inventory.md.

## 2026-07-12 — generation-recovery + corrections
- Session-limit interruption killed the W00/W01 and W02/W03 generators mid-write; gap survey
  identified 5 incomplete stories (W02-E02-S002 tasks/artifacts/evidence; W03 E02–E05 story
  file sets) — all filled by a dedicated gap-fill pass without touching existing files.
- W04/W05 trees generated (385 files, 24 stories, 89 tasks); W06/W07 generation in progress.
- CORRECTION: requirement-inventory.md AR-04 target fixed W05-E03-S003 → W05-E03-S002 to match
  the canonical wave-allocation-detail.md grouping (generator flagged the discrepancy; the
  allocation file was followed; inventory row was the typo).
- Pending on W06/W07 completion: waves/index.md roll-up refresh (stale "not yet generated"
  notes for W02+), then programme-wide mechanical validation + independent review gate.

## 2026-07-12 — programme generation complete + mechanical validation
- All 8 wave trees complete: 33 epics, 75 stories, 297 tasks, 1,277 files under waves/.
- waves/index.md roll-up refreshed: all rows linked with real epic/story counts; stale
  "subsequent batch" notes removed.
- Mechanical validation: structural completeness (every wave 7-file set, epic 7-file set,
  story 9-file set + ≥1 task) PASS; story-ID uniqueness PASS; zero execution-state statuses on
  waves/epics/stories/tasks (the 9 `status: accepted` hits are the D-01..D-09 ADR files —
  correct: those decisions were ratified 2026-07-11 in the review; the W00-E02-S003 story
  remains `planned` as the registration/verification of them); every W##-E##-S### target in
  requirement-inventory + wave-allocation-detail resolves to an existing story directory.
- Pending: independent review gate, then final report.

## 2026-07-12 — review-gate fixes (iteration 1: FAIL → corrections applied)
- [Medium] Dead ID W05-E03-S003 propagated-out: replaced with W05-E03-S002 in all 16 affected
  files (4 tracking registers, 2 analysis files, 10 tree files); the two disambiguation notes
  rewritten to the resolved state; change-log history preserved unmodified.
- [Low→adjudicated] AR-05 in W00-E01-S001: the generator had excluded AR-05 verification from the
  story body while leaving it in front matter, flagging the choice for the acceptance authority.
  RESOLVED by the programme author: AR-05 executed T1/T2 re-verification is IN scope (AC-04 +
  task T004 added; the W00 exit gate requires all 8 slices); AR-05 T3–T5 remain owned by W06-E04.
  AR-05 appearing in both stories' source_requirements is the same executed-slice/remainder split
  as SEC-02/AR-04/AR-06 — not a single-ownership violation.
- [Low] waves/index.md roll-up: W00 3→6, W01 9→10 story counts corrected (a collateral W07
  9→10 mis-substitution during the fix was caught and repaired to 9).
- [Low] FBL-01 story W05-E05-S001 depends_on populated with the seven W05-E01/E02 story IDs
  (was epic-prose only).
