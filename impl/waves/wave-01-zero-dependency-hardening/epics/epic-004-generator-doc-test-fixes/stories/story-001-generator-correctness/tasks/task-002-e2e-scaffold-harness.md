---
id: W01-E04-S001-T002
type: task
title: Isolated-temp-dir E2E scaffold harness (DX-01 T5)
status: done
parent_story: W01-E04-S001
owner: W01GenDX01
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E04-S001-02
artifacts:
  - ART-W01-E04-S001-002
evidence:
  - EV-W01-E04-S001-002
---

# W01-E04-S001-T002 — Isolated-temp-dir E2E scaffold harness (DX-01 T5)

## Task Definition

### Task objective

Build a real generate→build→boot→smoke test harness that runs in an isolated temporary directory:
`init` → `go mod download` → `go build` → contract/smoke tests → success, end to end, covering both the
released-CLI and source-built-CLI (`devel`) invocation paths. This is DX-01's own T5, and is explicitly
a **shared primitive**: it is reused by this story's own T004 (generator-output-boots test) and is
expected, in the future, to be reused by DX-04 (golden consumer + upgrade matrix, W06 scope, out of this
story's scope) — per `requirement-inventory.md`'s DX-01 row note, "T5 harness = shared primitive for
DX-02/DX-04."

### Parent story

W01-E04-S001 — Generator correctness — source-built CLI path validity and boot-safe CRUD generation.

### Owner

unassigned

### Status

todo

### Dependencies

None as a precondition to *start* (does not require T001 to be complete first, since the harness itself
is generic infrastructure — it invokes whatever `init`/`gen` behavior currently exists). However, this
task's own completion evidence for AC-W01-E04-S001-02 (proving `init` actually boots end-to-end) is most
meaningfully captured once T001 has landed, since a run against the pre-T001 `init` would still exhibit
the `v0.0.0` failure this story is fixing. Sequenced immediately after or alongside T001 per `plan.md`
"Implementation sequence."

### Detailed work

1. Design the harness's reusable shape: a Go test helper/package that accepts parameters for which
   generator command to run (`init`, `gen crud`, etc.), which CLI binary to exercise (built fresh from
   source vs. a stand-in for a released binary), and returns/asserts on the outcome of each pipeline
   step. Exact file location is a planned design choice, to be determined precisely at implementation
   time — expected to live within `internal/cli/`'s test infrastructure (see `plan.md` "Proposed
   architecture" and "Unresolved questions"), not invented here.
2. Implement the isolated-temp-dir creation/cleanup (a fresh `t.TempDir()`-rooted directory per
   invocation, or equivalent, ensuring no cross-test or cross-run contamination).
3. Implement the "released CLI" vs. "source-built CLI" distinction — expected via building the CLI test
   binary twice with different `-ldflags` version-injection values (one simulating a tagged release
   version, one simulating `devel`), exact mechanism to be confirmed at implementation time.
4. Implement each pipeline step: invoke `init` (or `gen crud`) against the temp dir; run `go mod
   download`; run `go build ./...` (or the appropriate build target) against the generated module; run
   whatever contract/smoke test the generated module ships with (the generated module's own test suite,
   or a minimal boot-and-health-check smoke test if the generated module has no tests of its own — to be
   confirmed at implementation time by inspecting what `init` actually scaffolds).
5. Ensure each step's failure is captured with full output and surfaced clearly (which step failed, with
   what output) rather than a bare pass/fail signal — this is what makes the harness usable as a
   diagnostic tool, not just a gate.
6. Prove the harness itself: run it end-to-end for both CLI paths and confirm success (this is the
   AC-W01-E04-S001-02 evidence).

### Expected files or components affected

A new isolated-temp-dir E2E harness (Go test helper/package under `internal/cli/`'s test infrastructure;
exact file location to be determined at implementation time — see `plan.md`).

### Expected output

A reusable harness, proven end-to-end for both the released-CLI and source-built-CLI paths, callable by
this story's own T004 and structured so a future DX-04 story can call it without modification to its
core pipeline logic (only different generator-command/assertion parameters).

### Required artifacts

ART-W01-E04-S001-002 (the harness itself).

### Required evidence

EV-W01-E04-S001-002 (functional-test report — the harness's own end-to-end run log, both CLI paths;
recorded under `DX-01/t5-e2e-temp-dir.json` per `story.md` "Required evidence").

### Related acceptance criteria

AC-W01-E04-S001-02.

### Completion criteria

The harness runs a full generate→build→boot→smoke cycle to success, end to end, for both the
released-CLI and source-built-CLI invocation paths, with each pipeline step's outcome captured in the
evidence log.

### Verification method

Direct execution of the harness's own test entry point (exact test name to be determined at
implementation time), logged output retained as evidence per `evidence/index.md`.

### Risks

Medium — this is the highest-complexity task in this story: it depends on real network access (`go mod
download`), real subprocess execution (`go build`), and correctly distinguishing two CLI-build
configurations in a test environment. A flaky or environment-sensitive harness would undermine its value
as a shared primitive for T004 and future DX-04 reuse. Mitigation: capture full output on every step
failure (see "Detailed work" step 5) so any flakiness is diagnosable rather than a silent red/green
signal.

### Rollback or recovery considerations

If the harness proves unreliable in CI (e.g. network flakiness from `go mod download`), the harness
design should isolate that step so it can be retried or the test skipped with a clear reason (not
silently passed) rather than the whole harness being reverted — since T004 and future DX-04 depend on
its existence.

## Implementation Record

Implemented 2026-07-13 by W01GenDX01, immediately after T001 per `plan.md` "Implementation sequence".

### What was actually implemented

`internal/cli/e2e_scaffold_harness_test.go` (new), containing the reusable pipeline primitive plus its
two proving tests:

- **`scaffoldPipeline(t, cli, modulePath, initArgs, goEnv)`** — the shared primitive: runs a REAL
  `wowapi` binary (subprocess, not an in-process call) through `init` → `go mod tidy` → `go mod
  download` → `go build ./...` → a written-in boot-and-validate smoke test (`kernel.New` →
  `app.Register(wire.Modules())` → `app.Boot`, no DB via a no-op TxManager — the scaffold ships no
  tests of its own, so the harness supplies the minimal boot smoke task-002 step 4 anticipated). Every
  step failure reports its STEP NAME plus the full command output (step 5's diagnostic contract).
  Parameterized by CLI binary, init/generator args, and go-env, so T004-style and future DX-04 callers
  plug in different generator commands/assertions without touching the pipeline core.
- **`buildWowapiCLI(t, ldflagsVersion)`** — the released-vs-source distinction, exactly as `plan.md`
  anticipated: the same checkout built twice, once plain (`devel`) and once with
  `-ldflags -X …buildinfo.version=v0.1.0` (the release pipeline's own stamping mechanism);
  `-buildvcs=false` keeps both deterministic regardless of the checkout's momentary VCS state.
- **`buildFrameworkProxy(t, version)`** — packages THIS checkout as `wowapi@v0.1.0` in a local
  `file://` GOPROXY (list/.info/.mod/.zip, stdlib archive/zip — no new dependency in the
  zero-dependency wave), so the released path's `go mod download` exercises real proxy resolution
  hermetically.
- **`hermeticGoEnv(proxy)`** — pins `GOPROXY` to `file://` proxies (framework proxy + the local module
  cache's download dir for dependencies), `GOSUMDB=off`, `GOWORK=off`, `GOFLAGS=-mod=mod`, and
  neutralizes developer-machine `GOPRIVATE`/`GONOPROXY` overrides (this workstation routes
  `github.com/qatoolist/*` straight to VCS, which would have bypassed the proxy — T001 finding 2).
- **`TestE2EScaffoldSourceBuiltCLI`** — devel binary: first asserts the OLD failure mode is impossible
  (flag-less init fails closed pre-write: non-zero exit, zero files, remediation), then runs the full
  pipeline via the sanctioned `--local-framework` workflow and asserts the replace directive landed.
- **`TestE2EScaffoldReleasedCLI`** — v0.1.0-stamped binary: flag-less init pins `wowapi v0.1.0` (no
  replace directive), and the full pipeline resolves the framework at that version from the proxy.

### Components changed

`internal/cli` test infrastructure only; no production code.

### Files changed

- `internal/cli/e2e_scaffold_harness_test.go` (new)

### Interfaces introduced or changed

Test-internal only: `scaffoldPipeline`, `buildWowapiCLI`, `buildFrameworkProxy`, `hermeticGoEnv`,
`runPipelineStep` — the reusable harness surface for T004-class and future DX-04 callers.

### Configuration changes

None. CI wiring falls out of test placement (`make test-unit` = `go test ./...`, no `-short`), the same
mechanism T004 recorded.

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

Per-step named logging (`pipeline step %q ok` / full failure output) — the harness-as-diagnostic-tool
requirement, not a runtime change.

### Tests added or modified

`TestE2EScaffoldSourceBuiltCLI`, `TestE2EScaffoldReleasedCLI` (both skipped in `-short`, like the other
rendered-product compile tests).

### Commits

None — uncommitted working-tree delta on HEAD 05dce5c8 (conductor owns commits).

### Pull requests

None (conductor owns commits/PRs).

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

- The released path serves this checkout AS v0.1.0 from a local proxy — it proves the resolution/
  download/build/boot pipeline for a release-stamped CLI, not the availability of any actually
  published tag (none exists to test against).
- The proxy zip whitelists the module's Go source trees plus metadata; a future top-level Go package
  not in the whitelist fails the harness build step loudly with the missing import named.

### Follow-up items

Forward reference only (not owed): DX-04 (W06) calls `scaffoldPipeline` with its own generator
args/assertions.

### Relationship to the approved plan

Matches `plan.md` step 5 and resolves its two harness-related "Unresolved questions": the harness lives
in a single new test file in `internal/cli/` (consistent with the T004 resolution), and the
released-vs-source distinction is two `-ldflags` builds of the same checkout — plus the file-proxy
mechanism the plan did not anticipate but the released path's `go mod download` requires to be hermetic.
DX-02's T004 keeps consuming the same underlying scaffold primitive (`buildRenderedProduct`, itself now
built on `init --local-framework`), satisfying AC-02's reuse clause without rework.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S001-02 | Run the harness end-to-end for both released-CLI and source-built-CLI paths | Isolated temp dirs, go1.26.5, file:// module proxies (no network) | `init` → `go mod tidy` → `go mod download` → `go build ./...` → boot smoke all succeed, both CLI paths | functional-test report (harness run log) | pending — wave review gate |

### Actual result

Both proving tests PASS with all five pipeline steps individually logged ok:
`TestE2EScaffoldSourceBuiltCLI` (5.39s, incl. the fail-closed guard) and `TestE2EScaffoldReleasedCLI`
(6.60s, framework fetched from the file proxy at v0.1.0) — `evidence/DX-01/t5-e2e-both-paths.log`.
Full package regression `go test ./internal/cli/ -count=1` → ok (24.4s) —
`evidence/DX-01/pkg-internal-cli-full-3.log`.

### Pass or fail

PASS.

### Evidence identifier

EV-W01-E04-S001-002 (`evidence/DX-01/t5-e2e-temp-dir.json`).

### Execution date

2026-07-13 (~08:00–08:05 UTC).

### Commit or revision

HEAD 05dce5c8a548f7dce3222637ab2c82024236a2a0; harness uncommitted on top (conductor commits).

### Environment

macOS Darwin 25.5.0 arm64, go1.26.5, local dev workstation; fully offline module resolution.

### Reviewer

Pending — wave-level review gate (conductor assigns).

### Findings

1. First released-path run failed: this workstation's `GOPRIVATE=github.com/qatoolist/*` made
   `GONOPROXY` bypass the file proxy (direct VCS → `unknown revision v0.1.0`). Fixed at the harness
   level by neutralizing `GOPRIVATE`/`GONOPROXY`/`GONOSUMDB` in `hermeticGoEnv` — exactly the
   environment-sensitivity class task-002's risk section flagged, now impossible by construction.
2. The pipeline includes `go mod tidy` before `go mod download` (the scaffold's own README/next-steps
   instruction) — a scaffold go.mod carries only the framework requirement, so tidy is the step that
   materializes the transitive requirements.

### Retest status

Not required — verified first-pass at the pinned revision (after the finding-1 harness fix).

### Final conclusion

AC-W01-E04-S001-02 verified: a real generate→build→boot→smoke cycle succeeds end to end for both CLI
invocation paths, with per-step diagnostics, hermetically.

## Deviations Record

No deviation from the approved plan; the plan's deferred design choices (harness location, released-CLI
mechanism) are recorded as resolutions above, and the file-proxy addition is design detail within the
task's own charter. Test-environment accommodation shared with T001 is recorded at story level
(DEV-W01-E04-S001-04).
