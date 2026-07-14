---
id: W01-E04-S001-TASKS-INDEX
type: tasks-index
parent_story: W01-E04-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E04-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task definition,
implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W01-E04-S001-T001](task-001-version-resolution-flags.md) | Version-resolution flags (DX-01 T1-T4) | W01GenDX01 | done | none | `--framework-version`/`--local-framework` flags, shape-classified VCS default (incl. the SF-7 `+dirty` stamp), `v0.0.0` fallback deleted | AC-W01-E04-S001-01 | done (uncommitted at 05dce5c8; conductor commits) | verified — EV-W01-E04-S001-001 |
| [W01-E04-S001-T002](task-002-e2e-scaffold-harness.md) | Isolated-temp-dir E2E scaffold harness (DX-01 T5) | W01GenDX01 | done | none | Reusable generate→build→boot→smoke harness (`scaffoldPipeline`, both CLI paths, hermetic file:// proxies) | AC-W01-E04-S001-02 | done (new file `internal/cli/e2e_scaffold_harness_test.go`) | verified — EV-W01-E04-S001-002 |
| [W01-E04-S001-T003](task-003-generator-verb-fix.md) | Generator verb fix (DX-02) | W01Gen | done | none | `.delete`→`.deactivate` template fix + corrected `TestGenCRUDPermissionKeys` | AC-W01-E04-S001-03 | done (uncommitted at 05dce5c8; conductor commits) | verified — EV-W01-E04-S001-003 |
| [W01-E04-S001-T004](task-004-generator-output-boots-test.md) | Generator-output-boots CI test | W01Gen | done | T002 (satisfied by reusing the existing `buildRenderedProduct` scaffold primitive), T003 | Fail-first CI test proving generated CRUD modules boot | AC-W01-E04-S001-04 | done (new file `internal/cli/gen_crud_boots_test.go`) | verified — EV-W01-E04-S001-004 |
| [W01-E04-S001-T005](task-005-scaffold-config-validate-fix.md) | Scaffold config validates under the framework-only path (scope addition, conductor-approved 2026-07-13) | W01Gen | done | none | `configs_base.yaml.tmpl` i18n block commented + fail-first `TestInitScaffoldConfigValidates` | AC-W01-E04-S001-05 (added) | done (uncommitted at 05dce5c8; conductor commits) | verified — EV-W01-E04-S001-005 |

## Grouping rationale

Per mandate §12 ("avoid excessive fragmentation into trivial tasks that provide no tracking value," and
"tasks must be decomposed when they... produce multiple unrelated outputs... need separate ownership...
need separate evidence... can block independently... have materially different risks"):

- **T001 keeps DX-01's T1-T4 as one task**, not four. All four sub-items are sequential steps in the
  *same* version-resolution code path within `init_cmd.go`: T1 (explicit-version flag) and T2
  (explicit-local-path flag) are two branches of the same "did the caller supply an override" decision;
  T3 (VCS-derived default) is what runs when neither override is supplied; T4 (deleting the old
  fallback) is only safe to do once T1-T3 collectively cover every case the fallback used to handle —
  T4 cannot be verified independently of T1-T3 having landed first. They share one owner, one file
  (`init_cmd.go`), one evidence pattern (fail-closed-before-write, proven per-path), and one risk
  profile (an incomplete case leaves a silent-bad-version gap). Splitting them into four tasks would
  multiply tracking overhead (four task files, four sets of front matter, four "related AC" entries all
  pointing at the same AC-W01-E04-S001-01) without adding independent verifiability — none of T1/T2/T3/T4
  can be meaningfully marked "done" in isolation from the others, since T4's completion criterion (no
  code path can produce `v0.0.0`) is only true once T1-T3 exist to replace it. This is a judgment call;
  a reasonable alternative would split T4 out as a final "delete the fallback" task once T1-T3 are
  proven, but even then T4 would depend on all three of T1/T2/T3, making it a poor candidate for
  independent tracking value. One task is kept.
- **T002 is kept separate from T001** because it is a materially different kind of work (test
  infrastructure, not command logic) with its own evidence (a harness run log, not a flag-verification
  log) and because it is explicitly a *shared primitive* — reused by T004 within this story and, in the
  future, by DX-04 (W06 scope) — which the mandate's "need separate ownership" and "can block
  independently" decomposition triggers both support treating as its own task rather than folding into
  T001.
- **T003 is kept separate from T001/T002** because it is a wholly unrelated defect in a different file
  (`resource.go.tmpl`, not `init_cmd.go`) with a different risk profile (RISK-W01-005's test-lock trap,
  specific to T003) and no shared code path with DX-01's work.
- **T004 is kept separate from T003** because it has its own evidence (a fail-before/pass-after boots-
  test run, distinct from T003's unit-test run) and an explicit two-way dependency (on T002's harness and
  T003's fix) that would be obscured if folded into either.

Four tasks is the natural grain: T001's internal coherence argues against further splitting; T002/T003/
T004 are each independently ownable, independently verifiable, and carry distinct evidence and risk.
