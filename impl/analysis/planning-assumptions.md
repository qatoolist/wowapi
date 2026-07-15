---
id: ANALYSIS-PLANNING-ASSUMPTIONS
type: analysis
title: Planning assumptions — explicitly recorded per mandate §18
status: complete
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Planning assumptions

Per mandate §18 ("Record assumptions explicitly" — a hard requirement, not optional). Each assumption
below states: ID, statement, basis (which source), risk-if-wrong, and re-validation trigger. These
are the assumptions this whole programme's disposition and target allocation rest on; none are
invented — each traces to a named source.

## ASSUMPTION-01 — Planning-time HEAD pin

- **Statement:** `main @ 0a31186` (post #22–#25 merges) is the repository state every disposition and
  evidence-verification-requirement in this programme assumes as its baseline. Every "already
  executed" / INV / partial claim in `requirement-inventory.md` is a claim about the code at (or
  before) this commit, not necessarily about whatever commit is HEAD when a wave actually executes.
- **Basis:** `impl/index.md` §"Source documents" item 5 ("Repository state at planning time: `main @
  0a31186` (post #22–#25 merges) — see `analysis/requirement-inventory.md` §E session delta");
  `requirement-inventory.md` §E (SD-01..SD-04, the four session-delta facts newer than the primary
  documents).
- **Risk if wrong:** if HEAD has advanced materially by the time W00 executes (further merges, reverts,
  or regressions to any of the 8 executed slices), W00's re-verification could either (a) find a slice
  has regressed, silently invalidating a "partial"/"INV" disposition that assumed the slice still
  holds, or (b) find additional session-delta facts (an "SD-05" equivalent) not accounted for anywhere
  in this analysis layer.
- **Re-validation trigger:** W00's own baseline-capture story (W00-E02-S001) must re-pin against
  whatever commit SHA it actually executes against and explicitly note any delta from `0a31186`; if a
  delta exists, W00-E01's re-verification stories treat any newly-discovered regression as a finding
  in its own right, not as an assumption the wave can silently proceed under (per `wave.md`'s own
  "Assumptions" section for W00).

## ASSUMPTION-02 — One-file-per-task instead of 4-file task directories

- **Statement:** where the mandate's illustrative directory structure (§3) shows a task directory
  containing four separate files (`task.md`, `implementation.md`, `verification.md`,
  `deviations.md`), this programme's actual `impl/` structure may consolidate a task's planning
  content into fewer files where the task is small enough that four separate files would be
  placeholder-only, rather than mechanically creating all four for every task regardless of size.
- **Basis:** mandate §18 itself: "Do not create placeholder files containing only headings where
  meaningful planning content can be derived" — this is the governing rule the one-file adaptation
  serves; the concrete governance rationale for how this project applies it is intended to live in
  `impl/governance/naming-conventions.md`, which does not yet exist in this repository as of this
  analysis (only `lifecycle.md`, `status-model.md`, `definition-of-ready.md`,
  `definition-of-done.md`, and the `templates/` directory exist under `impl/governance/` today).
- **Risk if wrong:** if a later governance pass authors `naming-conventions.md` with a stricter
  four-file-always rule, any task directories already created under this adaptation would need
  restructuring — a mechanical, low-risk rework, not a content loss, since no meaningful content would
  have been lost, only refiled.
- **Re-validation trigger:** the moment `impl/governance/naming-conventions.md` is authored, re-read it
  and confirm this adaptation is consistent with whatever it specifies; if it specifies something
  different, follow the newer governance document and update any already-created task directories to
  match, recording the change in `tracking/change-log.md` per `impl/index.md`'s progress-maintenance
  rule.

## ASSUMPTION-03 — Evidence/artifact subdirectories created on first use

- **Statement:** the mandate's illustrative directory tree (§3) pre-lists a full set of `evidence/`
  subdirectories (`baselines/`, `tests/`, `coverage/`, `logs/`, `screenshots/`, `benchmarks/`,
  `security/`, `static-analysis/`, `compatibility/`, `regression/`, `reviews/`, `acceptance/`) and
  `artifacts/` subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) under
  every story. This programme creates each subdirectory only when a story's first artifact/evidence
  item of that type is actually produced, rather than pre-populating all subdirectories empty for
  every story regardless of whether that story will ever produce that evidence type.
- **Basis:** mandate §18: "Do not create placeholder files containing only headings where meaningful
  planning content can be derived" (the same governing rule as ASSUMPTION-02) and mandate §9.3's own
  framing of the subdirectories as organizational categories ("Organise artifacts into...") rather than
  a mandatory-populate-all-twelve instruction; the concrete rationale is intended to live in
  `impl/governance/naming-conventions.md` (not yet authored — see ASSUMPTION-02's basis note on its
  absence).
- **Risk if wrong:** a story that never creates, say, a `screenshots/` subdirectory could look
  incomplete to a reviewer expecting to see all twelve subdirectories present-but-empty; this is a
  presentation risk, not a content-loss risk — the story's `evidence/index.md` (once created) is the
  actual source of truth for what evidence exists, not the presence of empty directories.
- **Re-validation trigger:** same as ASSUMPTION-02 — re-check against `naming-conventions.md` once
  authored; if it mandates pre-population, backfill empty subdirectories for already-created stories
  and record the change in `tracking/change-log.md`.

## ASSUMPTION-04 — Safe defaults for the three open human decisions

- **Statement:** for each of the three genuine human decisions (DEC-Q1, DEC-Q9, DEC-Q10), this
  programme proceeds with the REVIEW-specified safe/provisional default rather than blocking any wave
  on the human decision landing first.
  - **DEC-Q1 (IdP `grant_id` claim contract):** safe default per REVIEW §F Q1 — the framework builds
    and owns the server-side `identity_grant` table keyed on grant-ID now; the IdP's exact claim shape
    is tuning, not a blocker. SEC-01 implementation (W03-E01-S001..S004) proceeds without waiting on
    the human decision; only the final claim-shape wiring is gated on it.
  - **DEC-Q9 (reference-perf-env ownership):** provisional default per REVIEW §F Q9 / §12 — a
    GitHub-hosted Linux amd64 runner plus a committed `perf/reference-v1.json` baseline; advisory/
    relative benchmark comparisons (PERF-02..05, target W07-E01) proceed now; only absolute-SLO gating
    waits on the human decision to name a dedicated bare-metal reference environment.
  - **DEC-Q10 (repo-admin activation of branch/tag/env protection):** per REVIEW §G's decomposition —
    all agent-completable YAML/workflow authoring (REL-01 is "~85% buildable now" per
    `requirement-inventory.md`) proceeds now; only the final GitHub-settings activation step (branch
    protection on `main`, the protected `release` environment, the tag-protection ruleset) is
    human-gated. Session fact: merge-queue rulesets are unavailable on this user-owned repo tier
    (`impl/index.md` programme risk register), so REL-01/REL-02's design must not assume a merge-queue
    ruleset is available even after admin activation.
- **Basis:** REVIEW §F ("Resolution of the 10 unresolved questions — reduced to 3 genuine human
  decisions", table rows 1/9/10) and §G ("Blocker-resolution plan"); `requirement-inventory.md` table
  B rows DEC-Q1/DEC-Q9/DEC-Q10.
- **Risk if wrong:** if the eventual human decision diverges from the safe default (e.g., the IdP
  cannot key by grant-ID at all, or the reference-perf-env decision picks a fundamentally different
  measurement methodology than the committed `perf/reference-v1.json` baseline), the affected story's
  implementation may need rework at the point the human decision actually lands — but per REVIEW's own
  analysis this is bounded rework (claim-shape wiring only for Q1; SLO-gating wiring only for Q9;
  rollout-activation only for Q10), not a redesign of the underlying implementation.
- **Re-validation trigger:** the moment any of the three human decisions is actually made, the
  affected story (W03-E01 for Q1, W07-E01 for Q9, W06-E03 for Q10) must record the decision in
  `tracking/decision-register.md`, compare it against the safe default assumed here, and — if it
  diverges — open a deviation record per mandate §2.6 rather than silently rewriting the story's
  already-approved plan.

## ASSUMPTION-05 — PLAN §5's task tables remain the task-level source of record

- **Statement:** the 38 PLAN findings' own §5 per-task breakdown (acceptance criteria, tests, evidence
  paths, risk — organized by work package PF-ARCH/PF-SEC/PF-DATA/PF-DX/PF-PERF/PF-REL) is not
  re-derived from scratch inside this programme's `impl/waves/` story/task files. Where a story's task
  breakdown is created under `impl/waves/`, it cites the corresponding PLAN §5 task IDs (e.g. "AR-01
  T1–T11") rather than renumbering or restating them independently.
- **Basis:** `impl/index.md` §"Source documents" (item 3: PLAN "task-level source of record");
  `source-inventory.md`'s PRIMARY-tier framing of PLAN's continuing authority for task-level detail;
  `requirement-inventory.md`'s own header note ("per-task detail lives in the plan's own tables, which
  remain the task-level source of record").
- **Risk if wrong:** if a story's task breakdown drifts from PLAN §5's original task numbering or
  scope without an explicit cross-reference, later readers lose the ability to trace "why does this
  task exist" back to its original acceptance/test/evidence specification, defeating the mandate §2.4
  traceability chain (source requirement → task).
- **Re-validation trigger:** at the point any `impl/waves/.../tasks/task-NNN-*/task.md` is authored for
  a finding with an existing PLAN §5 table, confirm the task.md's own body cites the specific PLAN §5
  task ID(s) it elaborates; if a story's scope genuinely requires a task with no PLAN §5 counterpart
  (a MATRIX-only "new" item per `duplicate-analysis.md` §a), that absence should be stated explicitly
  in the task.md rather than left implicit.

## ASSUMPTION-06 — wowsociety facts are pinned at review date

- **Statement:** every wowsociety-side fact cited anywhere in this analysis layer (import counts, file
  counts, product-side impact statements) — for example CS-01's "kernel/mfa: 5 files in
  `internal/modules/identity/`, other 8 packages = 0 imports" (REVIEW §29 answer 17), or PLAN §8's
  "confirmed zero `kernel/attachment`/`kernel/notify` usage in wowsociety" — is a point-in-time,
  grep-verified fact as of the review date (2026-07-11, the MATRIX/REVIEW date), not a fact this
  programme re-verifies for itself.
- **Basis:** REVIEW §29 answer 17 (explicit grep-verified count); PLAN §8/§9's repeated "confirmed via
  a dedicated sanity check" / "confirmed zero ... usage in wowsociety" claims for each Wave-0/second-
  batch executed slice; the review's own methodology section (§B, "Three-commit and repository-state
  inventory") establishes that these are live-repository checks performed at review time, not
  historical or cached facts.
- **Risk if wrong:** if wowsociety's codebase changes materially before a PROD-01..05 coordination item
  actually executes (new imports of a re-homed package, new usage of `kernel/attachment`/`kernel/
  notify`, schema changes to `policy_override`), the "zero impact" or "5 files" claims this programme's
  scope-boundary and target-allocation decisions rely on could be stale, understating the real
  coordination surface.
- **Re-validation trigger:** immediately before any PROD-0X coordination item's cross-repo work begins
  (per `scope-boundary.md`'s enabling-framework-capability column), re-run the same grep/import-count
  checks against wowsociety's then-current HEAD rather than trusting the 2026-07-11 figures; record any
  delta as a new finding, not as a silent update to this assumption's stated numbers.
