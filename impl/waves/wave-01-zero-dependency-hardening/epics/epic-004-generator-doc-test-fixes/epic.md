---
id: W01-E04
type: epic
title: Generator, documentation, and test-diagnosis fixes
status: verification
wave: W01
owner: unassigned
reviewer: unassigned
priority: P0
created_at: 2026-07-12
updated_at: 2026-07-16
source_requirements:
  - DX-01
  - DX-02
  - DX-05
  - T-DOC-01
  - FBL-03
  - T-TEST-01
depends_on: []
stories:
  - W01-E04-S001
  - W01-E04-S002
  - W01-E04-S003
decisions: []
risks:
  - RISK-W01-004
  - RISK-W01-005
---

# W01-E04 — Generator, documentation, and test-diagnosis fixes

## Epic objective

Make three independent, developer-facing correctness and traceability gaps right: the CLI generator
must produce a module that actually boots (not merely one that hand-inspection judged plausible);
the programme's own planning documents must not contradict themselves about what has already
shipped; and the one open question about test-infrastructure reliability (an intermittent
`internal/e2e` full-suite failure) must be genuinely diagnosed, not re-labelled with a diagnosis that
was asserted without checking the facts. All three problems share a common thread — each is a case
where an existing artifact (a generator, a plan document, a test suite) makes a claim about its own
correctness that turns out not to hold up when actually checked — but they are otherwise unrelated in
mechanism, code area, and owner, which is why they are grouped as one epic rather than split into
three (mandate §4.2 permits multiple related stories per epic where they do not need decomposing into
separate epics; §4.3's story-boundedness discipline is what actually separates them into three
stories below).

## Problem being solved

`impl/analysis/requirement-inventory.md` records three independent findings targeting this epic,
each already given a disposition and target by the canonical allocation:

- **DX-01 + DX-02** (row DX-01, row DX-02) — `wowapi init` on a `devel` build silently templates an
  unresolvable `v0.0.0` into generated `go.mod`, and separately, `gen crud`'s emitted permission verb
  (`.delete`) is outside the kernel's closed authorization-verb set, so every `gen crud` invocation is
  dead-on-arrival at `Register`. Both are generator-correctness defects; DX-01's fix (version
  resolution) and DX-02's fix (verb correction) are unrelated in mechanism but share a fail-first
  proof method — a real generate→build→boot→smoke cycle in an isolated temp directory — which is why
  they are combined into one story (S001) rather than split, per DX-01's own row note: "T5 harness =
  shared primitive for DX-02/DX-04."
- **T-DOC-01 + DX-05 (residual) + FBL-03** (rows T-DOC-01, DX-05, FBL-03) — the implementation plan
  document's own traceability matrix (§6) disagrees with its own execution record (§9) about DX-05's
  status; DX-05 itself has three residual sub-tasks (T3/T4/T5) beyond the already-executed T1/T2; and
  the wowsociety upstream finding register has entries that are stale relative to this program's
  fixes. All three are documentation/traceability-integrity problems, not code-correctness problems —
  grouped into one story (S002).
- **T-TEST-01** (row T-TEST-01) — a review previously asserted a "shared-DB concurrency flake" cause
  for one observed `internal/e2e` full-suite failure, without first checking whether `testkit` already
  provides per-test DB isolation (it does — `testkit/db.go:83-144,313` clones a per-test database via
  `CREATE DATABASE ... TEMPLATE` from a content-hashed migrated template, dropped in `t.Cleanup`). The
  cause attribution is withdrawn; the observed fact (one failure, 4/4 isolated-run passes) stands and
  needs an honest, reproduction-first diagnosis — grouped into its own story (S003) because it is an
  investigation, not a pre-committed fix, and mixing it with S001/S002's implementation-shaped work
  would violate mandate §12's "mix foundational implementation and broad migration" decomposition
  trigger (here: mixing implementation with investigation).

## Scope

- DX-01 T1–T5: `--framework-version` flag with `go list -m` pre-write verification, `--local-framework`
  flag with explicit `replace` directive and dev-mode warning, VCS-derived pseudo-version as the
  fail-closed default when neither flag is passed, deletion of the `v0.0.0` fallback path, and a real
  generate→build→boot→smoke test harness in an isolated temp directory covering both the released-CLI
  and source-built-CLI paths.
- DX-02 Wave-0 slice only: the generator template's permission-verb fix (`.delete` → `.deactivate`)
  and the fix to the generator's own test that currently locks in the buggy verb as correct
  (`TestGenCRUDPermissionKeys`), plus a generator-output-boots CI test reusing DX-01 T5's harness.
- T-DOC-01: the PLAN document's §6 traceability-matrix row for DX-05, corrected to show T1/T2 as
  `EXECUTED` (matching §9's accurate execution record).
- DX-05 residual: T3 (blueprint-11 CLI example reconciliation against `internal/cli/cli.go`'s real
  commands/flags), T4 (`wowapi version` compatibility gate on mutating generator commands, sharing
  DX-01's version-verification plumbing), and T5 scoped narrowly or flagged as coordinating with W06
  (public API/config/event compatibility gates, explicitly shared with REL-03 per the plan).
- FBL-03: reconciling the wowsociety upstream finding register's PF-2/PF-6/RFF-001 entries — PF-2's
  closure is contingent on this epic's own S001 (DX-02) landing; PF-6/RFF-001 are corrected to
  already-resolved status per REVIEW Answer 18.
- T-TEST-01 re-scoped per MATRIX CS-13: a 3-step reproduction-and-diagnosis investigation (reproduce
  under `-count`+parallel; determine whether `internal/e2e` uses `testkit.NewDB` cloning or its own DB
  wiring; fix what the reproduction shows) with the fix step explicitly conditional on the
  investigation's findings, not invented in advance.

## Out of scope

- **DX-02's full P1/Wave-4 generator rewrite** (the disable-vs-minimal-slice decision, a status
  column, replacing TODO handlers) — `requirement-inventory.md`'s DX-02 row targets only the Wave-0
  slice at `W01-E04-S001`; the remainder is deferred to future work (W06 or later), consistent with
  MATRIX CS-14's own "one template token + one harness test" framing of what belongs in this wave.
- **DX-05's T1/T2** (README status banner rewrite, upgrade-policy rewrite) — already `EXECUTED` per
  `requirement-inventory.md` row DX-05 ("T1/T2 EXECUTED"), verified in W00-E01, not re-verified here.
- **DX-05 T5's full compat-gate build** — explicitly "shared with REL-03" per the plan, and REL-03 is
  W06 scope; this epic's S002 records T5 as deferred-to-W06 with the cross-reference, rather than
  attempting the full build or silently dropping it.
- **The wowsociety-side upstream register edit itself** — the register lives in a different repository
  (`wowsociety`, not `wowapi`); per mandate §2.3's framework/product boundary discipline, this epic can
  only plan/recommend that edit (a PROD-level coordination note, following the pattern in
  `requirement-inventory.md` §D's product-items table), not execute it directly.
- **Two additional gaps folded into FBL-07 by MATRIX CS-13** — hosted fuzzing never running real
  `-fuzz=` coverage-guided generation, and the pre-push hook's DB-silent-skip gap — both already
  covered by `W01-E01-S003`'s scope. S003 (this epic's e2e-flake story) explicitly excludes both with
  a cross-reference, to avoid duplicating tracked work.
- **`RouteMeta`/central-validation enforcement, HTTP timeout hardening, observability correlation,
  linter enablement** — these are `W01-E01`/`W01-E02`/`W01-E03`'s scope respectively; this epic does
  not touch them even though all four W01 epics land in the same wave.

## Source requirements

DX-01, DX-02 (Wave-0 slice only), DX-05 (residual: T3/T4, T5-deferred-to-W06), T-DOC-01, FBL-03,
T-TEST-01 (re-scoped). Cross-referenced: DX-04 (W06 scope — golden consumer reuses this epic's DX-01
T5 harness, but DX-04 itself is not implemented here), REL-03 (W06 scope — DX-05 T5 shares its
compat-gate plumbing but REL-03 itself is not implemented here).

## Architectural context

This epic sits entirely in `internal/cli/` (generator templates, generator commands, generator tests)
plus the programme's own documentation tree (`docs/implementation/premier-framework-implementation-plan.md`,
the wowsociety upstream register) and the test-infrastructure layer (`internal/e2e`, `testkit/`). It
does not touch `kernel/` capability packages at all — the generator produces code that *targets* the
kernel's closed authorization-verb set (`kernel/authz/registry.go:15-19`'s
`{create, read, list, update, deactivate, restore, approve, reject, assign, export, admin, ingest,
activate}`), but S001's DX-02 fix is a one-token change on the generator side, deliberately not
widening that closed set — the closed-set discipline is correct and intentional per the epic's own
scoping instruction. The version-resolution work (DX-01) touches `internal/cli`'s `init` command and
Go's own module-resolution tooling (`go list -m`, VCS metadata inspection) but produces no kernel
change either. This is why the epic has no upstream dependency on AR-01/AR-02/SEC-01/DATA-09 and can
land in W01 (`wave.md`'s "zero-dependency" framing) even though DX-05 T4/T5 and FBL-03's PF-2 sub-task
carry *internal*, intra-wave/intra-epic dependencies on this epic's own S001 (documented in
`dependencies.md`).

wowsociety impact is explicitly assessed per finding: DX-01 does not affect wowsociety (its
`replace => ../wowapi` path-replace mechanism never touches the CLI-generated dependency line that
DX-01 fixes); DX-02 does not affect wowsociety (its `docs/CONVENTIONS.md:10` governance — "never
bypass the generator, file an RFF instead" — kept existing wowsociety modules immune from the buggy
verb, since none was generated fresh since the bug was introduced); T-TEST-01 is wowapi-internal test
infrastructure, not applicable to wowsociety at all.

## Included stories

- **W01-E04-S001 — generator-correctness** (DX-01, DX-02 Wave-0 slice): source-built CLI path
  validity (version-resolution flags, VCS-derived pseudo-version, fallback removal, isolated-temp-dir
  E2E harness) plus the generator's permission-verb fix and its own test-lock fix, plus a
  generator-output-boots CI test reusing the harness.
- **W01-E04-S002 — documentation-reconciliation** (T-DOC-01, DX-05 residual, FBL-03): the plan
  document's §6-vs-§9 DX-05 status fix, DX-05's T3/T4/T5-deferred residual items, and the wowsociety
  upstream register's PF-2/PF-6/RFF-001 reconciliation (PF-2 contingent on S001).
- **W01-E04-S003 — e2e-flake-diagnosis** (T-TEST-01, re-scoped): a reproduction-first investigation of
  the intermittent `internal/e2e` full-suite failure, structured as an investigation task (T001)
  followed by a conditional fix task (T002) whose content depends on T001's findings.

## Dependencies

- **S002 → S001 (internal, cross-story, within this epic)**: S002's FBL-03 task item for PF-2's
  closure depends on S001's DX-02 task (the permission-verb fix) actually landing — PF-2 cannot be
  marked closed in the upstream register until the fix it references exists. Recorded in S002's
  `story.md` front matter as `depends_on: ["W01-E04-S001"]` and elaborated in `dependencies.md`.
- **S002's DX-05 T4 → S001's DX-01 version-verification plumbing (internal, cross-story, within this
  epic)**: DX-05 T4 ("`wowapi version` fails mutating generator commands on incompatible major/minor
  pairing") shares version-verification plumbing with DX-01, which this epic's S001 builds. This is a
  soft/implementation-plumbing dependency, not a hard blocking dependency — S002 can be planned in
  parallel with S001, but S002's T002 (which owns DX-05 T4) should not be implemented before S001's
  T001 (which owns DX-01 T1–T4) lands, since T002 reuses artifacts T001 produces.
- **This epic → W00**: entry criteria per `../../wave.md` — W00's exit gate (8 executed finding-slices
  re-verified at current HEAD, baselines captured, D-01..D-09 ratified). No epic-specific dependency
  beyond the wave-level W00 gate.
- **No dependency on W01-E01/E02/E03**: this epic's three stories target disjoint files
  (`internal/cli/`, documentation, `internal/e2e`/`testkit`) from the other three W01 epics' scope
  (lint config, observability, HTTP/validation) and can proceed in any order relative to them.

## Risks

RISK-W01-004 (T-TEST-01's reproduction step fails to reproduce the intermittent failure at all,
leaving the diagnosis inconclusive — an accepted, honestly-recorded possible outcome, not a story
failure) and RISK-W01-005 (the generator fix must also fix the generator's own test-locking assertion,
or the bug re-surfaces immediately) both originate at wave scope (`../../risks.md`) and land entirely
within this epic's stories (S003 and S001 respectively). See `risks.md` for the epic-scoped
elaboration.

## Required decisions

None. This epic requires no new ADR: DX-01/DX-02's fixes are mechanical corrections to existing,
already-decided generator behavior; T-DOC-01/DX-05/FBL-03 are documentation corrections, not design
decisions; T-TEST-01's diagnosis produces a decision-shaped *output* (what the fix in T002 should be),
but that decision is recorded as a task-level decision output inside S003's T001 task, not a
programme-level ADR — no story in this epic carries a `decisions/` directory.

## Epic acceptance criteria

- **AC-W01-E04-01**: `wowapi init` with an unresolvable `--framework-version` fails before any file
  write, with an exact remediation command in the error; `--local-framework` with a non-absolute or
  nonexistent path is rejected; a clean, reachable-commit `init` with neither flag derives an exact
  VCS pseudo-version; a dirty/unreachable-commit `init` with neither flag fails closed with
  remediation (never falls back to `v0.0.0`, which is deleted as a code path entirely). Traces to
  W01-E04-S001.
- **AC-W01-E04-02**: The DX-01 T5 isolated-temp-dir harness runs a real generate→build→boot→smoke
  cycle for both the released-CLI and source-built-CLI paths, ending in success end-to-end; this
  harness is reused (not reimplemented) by the DX-02 generator-output-boots test. Traces to
  W01-E04-S001.
- **AC-W01-E04-03**: `gen crud`'s emitted permission verb is `deactivate` (in the closed authorization
  set), not `delete`; `TestGenCRUDPermissionKeys` asserts the corrected verb, not the buggy one; the
  generator-output-boots test fails before the fix (closed-verb-set rejection at boot) and passes
  after. Traces to W01-E04-S001.
- **AC-W01-E04-04**: The PLAN document's §6 traceability-matrix row for DX-05 shows T1/T2 as
  `EXECUTED`, matching §9's execution record; DX-05 T3's blueprint-11 CLI examples are reconciled
  (each example implemented or deleted, no example left silently wrong); DX-05 T4's version-
  compatibility gate on mutating generator commands is planned with its dependency on S001's version
  plumbing explicit; DX-05 T5 is explicitly recorded as deferred-to-W06 with the REL-03 cross-
  reference, not silently dropped. Traces to W01-E04-S002.
- **AC-W01-E04-05**: FBL-03's target register plan marks PF-2 as closeable contingent on S001's DX-02
  task, and PF-6/RFF-001 as corrected to already-resolved status per REVIEW Answer 18; because the
  register lives in the wowsociety repository, this criterion is satisfied by a documented, precise
  PROD-level coordination recommendation (per `requirement-inventory.md` §D's pattern), not a direct
  edit to a wowsociety-repository file. Traces to W01-E04-S002.
- **AC-W01-E04-06**: T-TEST-01's reproduction is attempted under `-count`+parallel full-suite runs; a
  determination is recorded of whether `internal/e2e` uses `testkit.NewDB` cloning or its own DB
  wiring; the resulting diagnosis (confirmed cause, or an honest "not reproducible, downgraded to
  monitoring") is recorded without re-asserting the withdrawn "shared-DB concurrency" cause. Traces to
  W01-E04-S003.
- **AC-W01-E04-07**: All three stories have passed independent review per mandate §14, with S001
  specifically checked for the test-lock fix (RISK-W01-005) and S003 specifically checked for not
  pre-committing to a fix mechanism before its reproduction step completes.

## Closure conditions

All three stories (S001, S002, S003) reach `accepted` per `governance/definition-of-done.md`; all
seven epic acceptance criteria above are verified with registered evidence, not merely implemented,
per mandate §2.5; `closure-report.md` for this epic is completed with reviewer conclusion and
acceptance date; no unresolved regression or silently-dropped scope item (particularly DX-02's
excluded Wave-0 sub-tasks and T-TEST-01's excluded FBL-07-duplicate items) remains open.

## Status update (2026-07-16)

`status: verification` (was `planned`; the parent wave claimed `accepted`) — set by the
2026-07-16 hierarchy reconciliation (**DEV-PROG-006**). All of this epic's stories are
`accepted` with story-level evidence, but this epic's own `closure-report.md` body was never
populated: its acceptance-criteria/story-completion tables still read "not started"/"planned"
while a reviewer-conclusion section appended 2026-07-13 claims acceptance. Until the closure
report is completed honestly against the epic's acceptance criteria, the epic sits in
`verification`; see DEV-PROG-006 in `impl/tracking/programme-deviations.md` for the disposition.
