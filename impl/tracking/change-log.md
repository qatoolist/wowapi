---
id: TRACK-CHANGE-LOG
type: register
title: Change log — programme-structure changes (append-only)
status: active
created_at: 2026-07-12
updated_at: 2026-07-16
derived: false
---

# Change log

This is the one **non-derived** file in `impl/tracking/`: it is itself the source of record for
programme-structure changes. Records every programme-structure change (new/removed waves, epics,
stories; governance-policy changes; register-schema changes). Does not record per-story
implementation progress — that lives in each story's own files and rolls up via
`status-register.md`.

Reverse-chronological (newest first).

## 2026-07-16 — Findings-remediation pass R-1 follow-up: task front-matter and register reconciliation

- **Task front-matter status normalization:** 31 W02/W03 task files reconciled: todo→done vocabulary normalization (34 files); front-matter status values matched to mapped task evidence. W02-E02-S002 T001/T002/T003 reclassified `implemented` (schema delivered and re-verified live, but three specific proof artifacts never built — noted in task caveat entries).
- **Story and index demotion:** W02-E02-S002 story.md acceptance rolled back to `implemented` per findings-remediation adjudication (discovery: three tasks' named proof artifacts in acceptance criteria never built); story.md given status note and closure.md given dated correction line; parent epic W02-E02 and wave W02 set to `partially-accepted`. Tasks/index.md files for 10 stories (8 W02 + 2 W03) updated to match current task front-matter status values; W06-E03-S001 T006 status cell changed from `done-with-deviation` to `done` (deviation tracking preserved in other columns).
- **New deviation and decision records:** DEV-PROG-005 (W02-E02-S002 acceptance rollback with three deferred proof artifacts) and DEC-PROG-003 (proposed, human ratification pending — options for W02-E02-S002 completion: build the 3 artifacts + re-accept, or formally descope via decision).
- **Verification records updated:** W03-E01-S003 tasks T001/T002 "Verification Record" sections replaced with real evidence from closure.md re-run on 2026-07-16 (DB-backed tests PASS).
- **W03 caveat notes added:** Task status notes added to W02-E01-S001-T001 and W02-E01-S003-T004 per findings-remediation adjudication (evidence reconciliation).
- **Status-register ready for regeneration:** vocabulary normalization and roll-up corrections in place; `python3 miscellaneous/regen_status_register.py --check` run in this pass (see verification output below).

## 2026-07-16 — Conductor adjudication applied per the six review-gate-2026-07-16.md records

Conductor adjudication (Fable 5), per review-gate-2026-07-16.md records. Applied the conductor's
final status adjudications across Waves 00, 01, 02, 03, 04, and 06 following the independent
review gates executed 2026-07-16 (`review-gate-2026-07-16.md` under each wave's directory; W02 has
no wave-level file of that name — its per-story independent-review task files, dated 2026-07-16,
served as the basis instead, flagged as a conflict with the dispatch instructions rather than
improvised). W00: all 6 stories accepted, including a conductor adjudication note for
W00-E01-S001's AC-04/DEV-02 and a cross-reference from W00-E02-S001 to the coverage-floor
supersession (DEV-PROG-001/DEC-PROG-001). W01: all 10 stories remain accepted; W01-E01-S003 gets a
condition note (AC-W01-03's CI-execution leg outstanding) and a new technical-debt-register.md row
(TD-005). W02: wave and all 5 epics and all 8 stories set `accepted`; the "8 edges" → "9 edges"
count corrected in W02-E02-S002's closure.md and in `testkit/tenant_fk_cross_tenant_test.go`'s
comment; W02-E03-S001's false "T006"/"W02ReviewGate" citation corrected. W03: mixed —
E01-S001/S002 accepted, E01-S003 verified (not accepted, human sign-off gate — new
`deferred-items-register.md` row DEF-07), E01-S004 downgraded accepted → implemented (cross-repo
sign-off unverifiable), E02/E03/E04/E05-S001 accepted; wave and E02-E05 epics accepted, E01 epic
in-progress, wave in-progress. W04: E01-S001/S002/S003 closure.md template sections filled and
accepted; E02-S001 accepted (C-1 remediation reviewed and passed); E02-S002 stays planned; E04-S001/
S002 `closed-pending-review` normalized to `accepted`; E04-S003 accepted; epic-002 in-progress,
other epics accepted; wave and closure-report.md set/annotated `in-progress`. W06: wave-level
roll-ups (wave.md, progress.md, acceptance.md, closure-report.md) corrected from
`planned`/"not begun" to `in-progress` with an honest summary; W06-E04-S002 closure.md given a
scoping note (T5-only acceptance; T4/AC-01 open pending W05-E03); all 4 epics set `in-progress`
per their stories. `CHANGELOG.md` `[Unreleased]` remediation note updated from in-progress to
completed.

## 2026-07-16 — Autopsy remediation R-1: truth-reconciliation pass

- Ref: `impl/reports/implementation-autopsy-report-2026-07-16.md` (independent implementation
  autopsy of Waves 00-07; verdict: programme not complete, `e8cda6b`'s finalization claim
  rejected). This entry records remediation R-1 (of R-1..R-10), the truth-reconciliation pass over
  `impl/` tracking/governance/wave markdown — the gate for all further acceptance activity.
- **False statuses reverted:** `W04-E02-S002` story.md/closure.md `accepted` → `planned` (C-2,
  closure body already honestly said unimplemented); `W03-E03-S001` closure.md `accepted` →
  `implemented` (C-3, story.md was already honestly `ready`); `W03-E02-S001` closure.md `accepted`
  → `implemented` (H-5, self-review only); `W02` closure-report.md `accepted` → `verification`
  (C-4, review gate falsely claimed — all 6 epic-level review tasks were `todo`); `W04` wave
  closure-report.md `accepted` → `in-progress`, template body replaced with an honest interim
  state summary (C-5); `W04-E02-S001` story.md/closure.md `accepted` → `implemented` (C-1 code
  defect — webhook outbound I/O inside an open DB transaction — remediated 2026-07-16 in the
  working tree, re-review pending).
- **Template closures filled honestly (M-2):** `W04-E01-S001/S002/S003` and `W04-E02-S001/S002`
  closure.md bodies given short current-state paragraphs reflecting the autopsy's §4 matrix
  verdicts; story.md statuses left unchanged except `W04-E02-S001` (see above).
- **W05 contradictory-tracking stories annotated, not re-statused:** `W05-E01-S001`,
  `W05-E02-S001`, `W05-E03-S001`, `W05-E05-S001` each got a dated note recording that related code
  landed outside their tracked execution (H-6/H-7), pointing to `DEV-PROG-002`.
- **New deviation records:** `DEV-PROG-001` (coverage floor 90.0→84.0 silently lowered in
  `e8cda6b`, H-1), `DEV-PROG-002` (FBL-01/AR-01/AR-02/authz_epoch executed outside W05's story
  lifecycle, H-6/H-7), `DEV-PROG-003` (W02/W03/W04/W05 sequencing-gate bypasses, H-7),
  `DEV-PROG-004` (commit `e8cda6b` message overclaims finalization, H-2) — full records in
  `impl/tracking/programme-deviations.md`, summary rows in `impl/tracking/deviation-register.md`.
- **New decision records (both `proposed`, human ratification pending):** `DEC-PROG-001` (interim
  84.0% coverage floor acknowledged as a regression, ratchet plan to 90.0%), `DEC-PROG-002`
  (disposition of AR-01/AR-02/SEC-04 deferred to the Wave 05 execution owner) — full records in
  `impl/tracking/programme-decisions.md`, summary rows in `impl/tracking/decision-register.md`.
- **Evidence-gap acknowledgments (H-4):** dated notes added to `W00`, `W01`, and `W06` wave
  closure-report.md files stating the review gate outcome was claimed without a compliant evidence
  record; re-review scheduled 2026-07-16. No reviews were fabricated to fill these gaps.
- **Not done in this pass (explicitly out of scope for R-1):** `impl/tracking/status-register.md`
  was not hand-edited (it is script-regenerated); no acceptance/closure activity was performed; no
  Go code was changed as part of this pass (the C-1/H-9 code fixes referenced above were made
  separately, outside R-1's own scope, and are only cross-referenced here).

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
