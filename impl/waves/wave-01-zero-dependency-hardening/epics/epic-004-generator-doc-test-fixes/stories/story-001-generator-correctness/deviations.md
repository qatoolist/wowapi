---
id: DEV-W01-E04-S001
type: deviations-record
parent_story: W01-E04-S001
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W01-E04-S001

Per mandate §8.9/§2.6. The approved `plan.md` is not altered; divergences are recorded here.

## DEV-W01-E04-S001-01 — DX-01 slice (T001/T002) not implemented this session

- **Approved plan**: four tasks; T001 (version-resolution flags, `v0.0.0` fallback deletion) and T002
  (isolated-temp-dir E2E harness) implement DX-01, verified by AC-W01-E04-S001-01/-02.
- **Actual implementation**: only the DX-02 slice (T003 + T004) was implemented and verified. The
  wave conductor's W01Gen assignment scoped this worker to "the generator tool (gen crud), its
  templates (resource.go.tmpl), and the generator-output-boots test harness," with new generator
  features explicitly a non-goal — which excludes T001's `--framework-version`/`--local-framework`
  flag work and T002's full released-vs-source-CLI harness.
- **Reason**: conductor task-assignment scope, not a technical blocker. Nothing discovered this session
  prevents T001/T002 as planned.
- **Impact**: AC-W01-E04-S001-01 and AC-W01-E04-S001-02 remain unverified; the story cannot move to
  `verified` as a whole. T004's dependency on T002 was satisfied without scope creep by reusing the
  ALREADY-EXISTING `buildRenderedProduct` scaffold primitive (`internal/cli/scaffold_test.go:568`,
  init → replace-to-local-checkout → tidy), which is the natural seed of T002's harness; T002 still owes
  the released-CLI vs source-built-CLI distinction, `go mod download`, and the full smoke cycle.
- **Risks**: DX-01's silent `v0.0.0` defect remains live at HEAD (independently re-confirmed this
  session by W01-E04-S002's owner during doc verification). Mitigated only by awareness, not by code.
- **Approval**: implicit in the conductor's assignment boundaries; flagged in this worker's final
  report for explicit conductor disposition (assign T001/T002 to a follow-up worker or session).
- **Compensating controls**: none needed for DX-02; for DX-01, the wowsociety `replace`-directive
  workaround documented in the story's current-state assessment remains the operative mitigation.
- **Follow-up work**: implement T001 + T002 per `plan.md` steps 1-5.
- **Resolution (2026-07-13)**: CLOSED — T001 + T002 implemented and verified by follow-up worker
  W01GenDX01 in this same working tree (EV-W01-E04-S001-001/-002). AC-01/-02 now pass; the story is
  whole. The wowsociety `replace`-workaround compensating control is superseded by the first-class
  `--local-framework` flag.

## DEV-W01-E04-S001-02 — line-number drift against story/task citations (informational)

- **Approved plan/story citations**: `scaffold_test.go:937-953` (`TestGenCRUDPermissionKeys`, assertion
  at 949); baseline SHA cited by the wave brief as 0a31186.
- **Actual**: at execution the test sat at `scaffold_test.go:985-1001` (assertion at 997) because
  sibling story W01-E03-S002's owner added tests earlier in the same file in the shared working tree;
  live HEAD was `05dce5c8a548f7dce3222637ab2c82024236a2a0` (the brief's baseline had advanced).
  `resource.go.tmpl:54` did not drift. Content matched all citations exactly; per the wave constraint,
  targets were re-derived against the live tree and the drift recorded here rather than forcing stale
  citations.
- **Impact/Risks**: none — cosmetic.

## DEV-W01-E04-S001-03 — scope addition: scaffold config-validate fix (T005)

- **Approved plan**: four tasks (T001-T004); no scaffold-config-schema work.
- **Actual implementation**: a fifth task, W01-E04-S001-T005, added and completed. Origin:
  W01-E04-S002's DEV-03(a) finding (a pristine scaffold's `configs/base.yaml` carries an active
  `i18n:` block whose product-owned keys the framework-only `config validate` fallback rejects, so
  `config validate --env local` fails on generator output). Escalated to Main by S002's owner;
  conductor approved the scope addition into this story under its generator-correctness charter on
  2026-07-13 and directed fail-first execution.
- **Reason**: the defect is generator-output correctness — the same class this story exists for —
  and the fix's true source (the scaffold template's convention violation) sits in files this story's
  owner already holds.
- **Impact**: one new acceptance criterion (AC-W01-E04-S001-05), task file, artifact
  (ART-W01-E04-S001-006), and evidence record (EV-W01-E04-S001-005). No change to T001-T004 scope.
- **Risks**: low; residual — the `tools/configcheck` delegation leg of in-scaffold `config validate`
  remains broken by DX-01's go.mod defect until T001 lands (recorded in T005's risk section).
- **Approval**: conductor (Main), 2026-07-13, via IRC scope-add instruction.
- **Compensating controls / Follow-up**: none beyond T001.

## DEV-W01-E04-S001-04 — DX-01 default-path scope extension: Go 1.24+ stamped-version shapes (T001)

- **Approved plan**: `plan.md` step 3 frames the no-flags default as "when neither flag is set, derive
  a VCS pseudo-version", with the defect cited as the `devel` → `v0.0.0` fallback (`buildinfo.Version()
  == "devel"`).
- **Actual implementation**: at execution, the defect's dominant live shape was found to be broader
  than the plan's framing: on Go 1.24+, a locally `go build`-built CLI is STAMPED with a VCS
  pseudo-version (e.g. `v1.0.1-0.20260713072141-05dce5c8a548+dirty`), so `buildinfo.Version()` never
  returns `"devel"` for it and the old fallback never fired — init wrote the unresolvable stamp
  verbatim (this is exactly W01-E04-S002 DEV-03 / wowsociety SF-7's "+dirty pseudo-version" finding,
  reproduced in `evidence/DX-01/t1-t4-prefix-failfirst.log`). T001's no-flags path therefore classifies
  the stamped version by shape: tagged release → used as-is; `…+dirty` → fail closed; clean stamped
  pseudo-version → verified resolvable pre-write; unstamped `devel` → derived from `vcs.revision`,
  verified, else fail closed. Additionally, the test suite's `callInit` helper now pins
  `--local-framework <this checkout>` for the ~35 pre-existing init tests that are not about version
  resolution (a devel test binary has no VCS stamp, so their flag-less invocations would otherwise
  correctly fail closed), and `buildRenderedProduct` builds on the new flag instead of hand-appending a
  replace directive.
- **Reason**: fixing only the plan's literal `devel` arm would have left the defect fully live for
  every `go build`-built CLI — the most common contributor workflow. Same defect class, same file, same
  acceptance criterion; the plan's own resolution mechanism (`go list -m` verification) covers it.
- **Impact**: AC-W01-E04-S001-01 is proven against BOTH defect shapes; no task scope changed. Plan
  "Unresolved questions" resolved: reachability check = `go list -m <module>@<revision>` (resolution ≡
  reachability, returning the canonical version); released-vs-source harness distinction = two
  `-ldflags` builds; harness location = `internal/cli/e2e_scaffold_harness_test.go`; `initData` gained
  one `LocalFramework` field and `go.mod.tmpl` one conditional replace block.
- **Risks**: none identified; the shape classifier is regression-tested per shape (including the SF-7
  `+dirty` test) and every failure test asserts zero files written.
- **Approval**: within the T001 assignment's charter ("fix at the source ... so a pristine scaffold's
  go mod download/build works"); recorded here for reviewer visibility rather than silently absorbed.
- **Compensating controls**: not needed.
- **Follow-up work**: none. (Informational: W01-E04-S002's FBL-03 can now recommend closing the
  wowsociety SF-7 upstream finding once the conductor commits this tree.)
