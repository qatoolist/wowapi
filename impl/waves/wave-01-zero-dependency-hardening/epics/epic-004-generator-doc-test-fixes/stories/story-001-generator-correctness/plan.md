---
id: PLAN-W01-E04-S001
type: plan
parent_story: W01-E04-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W01-E04-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information.

## Proposed architecture

No new architectural layer or package. Both defects are fixed within `internal/cli/`'s existing command
and template structure:

- DX-01's fix is a **pre-write verification step** inserted into `init`'s existing command flow: today,
  `init_cmd.go` computes `fwVer` (lines 122-123, the buggy fallback) and then proceeds directly to
  writing files via `initData`/the template file list. The fix restructures this into an explicit
  sequence — parse flags → resolve version (by one of three mutually exclusive paths: explicit
  `--framework-version` verified via `go list -m`, explicit `--local-framework` with a `replace`
  directive, or VCS-derived pseudo-version) → **only then** write any file. This is a control-flow
  change within the existing `runInit`-shaped function (exact function name to be confirmed at
  implementation time), not a new package or interface. The invariant this establishes: no file write
  occurs before version resolution either succeeds or the command has already exited non-zero with a
  remediation message.
- DX-02's fix is a one-token change to an existing template string plus a corresponding one-string
  change to an existing test assertion. No structural change.
- The isolated-temp-dir E2E harness (DX-01 T5) is a **planned design choice**, not yet a determined file
  path: it is expected to live in `internal/cli/`'s test infrastructure (e.g. as a Go test helper
  function or small internal test-support package invoked from `_test.go` files in that directory), but
  per mandate §8.5's instruction not to invent precise file paths the repository does not yet determine,
  its exact location (a single helper function in an existing test file, vs. a new
  `internal/cli/testharness` package, vs. something else) is left for implementation time. What is
  fixed by this plan: it must be reusable — callable from both this story's own T004 (generator-output-
  boots test) and, in the future, DX-04's golden-consumer test (W06 scope, not part of this story) —
  which means it must accept parameters for "which generator command to run" and "which CLI binary to
  exercise" (source-built vs. released) rather than hardcoding either.

## Implementation strategy

1. **DX-01 T1** — Add a `--framework-version` string flag to `init`'s flag set. When set, before any
   file write, shell out to (or otherwise invoke) `go list -m <module>@<version>` for the framework
   module and fail with a remediation message if it does not resolve.
2. **DX-01 T2** — Add a `--local-framework` string flag. When set, validate the path is absolute and
   exists before any file write; on success, populate `initData` (or its equivalent) so the generated
   `go.mod` template emits an explicit `replace` directive, and print a visible dev-mode warning to
   stderr/stdout.
3. **DX-01 T3** — When neither flag is set, derive a VCS pseudo-version: determine the current commit,
   check the working tree is clean and the commit is reachable (not purely local/unpushed-only — exact
   reachability check to be determined at implementation time, e.g. via `git merge-base --is-ancestor`
   against a known remote ref, or an equivalent Go-native VCS-status inspection), and construct the
   pseudo-version string. If the tree is dirty or the commit is not reachable, fail closed with
   remediation — do not fall back to any placeholder version.
4. **DX-01 T4** — Delete `init_cmd.go:122-123`'s `if fwVer == "devel" { fwVer = "v0.0.0" }` block
   entirely, once T1-T3 collectively cover every case that block used to handle.
5. **DX-01 T5** — Build the isolated-temp-dir harness described above. Use it to write a test that: (a)
   creates a temp dir, (b) invokes `init` (via both the currently-built/source CLI binary and, where
   feasible, a stand-in for a "released" CLI binary — exact mechanism for distinguishing the two paths
   in a test environment to be determined at implementation time, e.g. building the CLI twice with
   different `-ldflags` version-injection values), (c) runs `go mod download`, (d) runs `go build`, (e)
   runs whatever contract/smoke test the generated module ships with, and (f) asserts success at every
   step, failing loudly with the captured output of whichever step failed.
6. **DX-02 T003** — Change `resource.go.tmpl:54`'s `"{{.PermPrefix}}.delete"` to
   `"{{.PermPrefix}}.deactivate"`. Change `scaffold_test.go:949`'s `"widgets.widget.delete"` to
   `"widgets.widget.deactivate"`. Run `TestGenCRUDPermissionKeys` before the change (to confirm it
   currently passes with the buggy string, establishing the fail-first baseline for the *test's own*
   correctness) and after (to confirm it passes with the corrected string).
7. **DX-02 T004** — Write the generator-output-boots test, reusing T002's (DX-01 T5's) harness: generate
   a `gen crud` module, boot it, assert no closed-verb-set rejection. Run this test against the
   pre-T003 template first (fail-first: it must fail with the closed-verb-set rejection at
   `kernel/authz/registry.go:88-90`), then against the post-T003 template (must pass).

## Expected package or module changes

`internal/cli` (the `init` command's flag parsing and version-resolution logic; the CRUD generator
template; the generator's own test file; a new test-infrastructure harness). No `kernel/` package is
modified — `kernel/authz/registry.go`'s closed verb set is read/relied-upon, not changed.

## Expected file changes where determinable

- `internal/cli/init_cmd.go` — add `--framework-version`/`--local-framework` flags, restructure the
  version-resolution logic per the pre-write-verification sequence above, delete lines 122-123's
  fallback.
- `internal/cli/templates/crud/resource.go.tmpl:54` — `.delete` → `.deactivate`.
- `internal/cli/scaffold_test.go:949` (`TestGenCRUDPermissionKeys`) — `"widgets.widget.delete"` →
  `"widgets.widget.deactivate"`.
- A new isolated-temp-dir E2E test harness file, and a new generator-output-boots test file — exact
  paths not yet determined (see "Proposed architecture" above); to be located precisely at
  implementation time within `internal/cli/`'s existing test-file layout.
- Possibly `internal/cli/init/go.mod.tmpl` (or wherever the generated `go.mod`'s template lives) if the
  `replace` directive (T2) or version-line templating (T1/T3) requires a template-side change beyond
  the data passed into it — to be confirmed at implementation time by reading that template.

## Contracts and interfaces

`wowapi init`'s CLI flag surface gains two new flags (`--framework-version`, `--local-framework`) —
additive, does not change any existing flag's behavior when neither new flag is passed except for the
version-resolution *default* changing from "always succeed with a possibly-broken value" to "succeed
with a verified value, or fail closed" (see `story.md` "Compatibility considerations" for why this is
treated as a strict improvement, not a breaking contract change). No other public interface (Go API,
generated code's own interface) changes.

## Data structures

`initData` (or its equivalent internal struct feeding the `go.mod` template) gains whatever fields are
needed to represent "an explicit replace directive is present" (for `--local-framework`) versus "a plain
require-line version" (for `--framework-version` or the VCS-derived default) — exact field shape to be
determined at implementation time by reading the current `initData` struct and its template.

## APIs

None affected — this story is CLI-internal; no HTTP or programmatic API changes.

## Configuration changes

None beyond the two new CLI flags described above; no environment-variable or persisted-configuration
change.

## Persistence changes

None.

## Migration strategy

Not applicable — no schema or data migration.

## Concurrency implications

None. `init`'s version resolution and file-writing are single-threaded, sequential CLI operations; this
story does not introduce concurrency.

## Error-handling strategy

Fail-closed throughout: every one of DX-01's three resolution paths (explicit version, explicit local
path, VCS-derived default) either succeeds and produces a verified version/replace-directive, or the
command exits non-zero with an exact remediation command, before any file write occurs. This is the
single governing invariant for T1-T4 (see "Task decomposition" below for why they are one task). DX-02's
fix does not add new error-handling — the kernel's existing `Register`/`Err()` behavior is unchanged and
continues to be the boot-time backstop; the fix simply makes the generator stop producing input that
trips it.

## Security controls

None new. See `story.md` "Security considerations" — fail-closed-on-invalid-input is a general hygiene
improvement but not a security control this story is responsible for.

## Observability changes

None required by this story's acceptance criteria. The fail-closed remediation messages are the primary
UX surface (see "Documentation requirements" in `story.md`), delivered via the CLI's existing
stdout/stderr error-reporting path, not via a new logging/metrics/tracing mechanism.

## Testing strategy

Fail-first for both DX-01 and DX-02:

- **DX-01 fail-first**: before T4 deletes the `v0.0.0` fallback, a test exercises `init` on a `devel`-
  version build with neither `--framework-version` nor `--local-framework` passed and confirms the
  current (buggy) behavior — `init` succeeds and writes `v0.0.0` into `go.mod`. After T1-T4 land, the
  same invocation (devel build, no flags, and — critically — a commit state that is either dirty or
  unreachable, since a clean/reachable devel-build commit should now succeed via the VCS-derived path
  from T3) is re-run and confirmed to fail closed with a remediation message instead, and no file is
  written. A second test confirms the clean/reachable-commit case now succeeds with a real,
  non-`v0.0.0` pseudo-version.
- **DX-02 fail-first**: before T003's template fix, `TestGenCRUDPermissionKeys` is run and confirmed to
  pass with the (bug-asserting) `"widgets.widget.delete"` string — this establishes that the test
  currently locks in the bug, motivating why the test itself must change. After T003, the same test
  (now asserting `"widgets.widget.deactivate"`) is re-run and confirmed to pass. Independently, T004's
  generator-output-boots test is run against the pre-T003 template and confirmed to fail with the exact
  closed-verb-set rejection at `kernel/authz/registry.go:88-90`, then re-run against the post-T003
  template and confirmed to pass.
- T002's (DX-01 T5's) harness itself is proven by T002's own completion criteria: a full
  generate→build→boot→smoke cycle succeeding end-to-end for both the released-CLI and source-built-CLI
  paths, which is itself the acceptance evidence for AC-W01-E04-S001-02.

## Regression strategy

The generator-output-boots test (T004), once landed, becomes a permanent CI regression guard: any future
change to `resource.go.tmpl` (or to `kernel/authz/registry.go`'s verb set) that reintroduces an
out-of-set permission verb in generated CRUD output will fail this test before merge. Similarly, once
T002's harness exists and is wired into CI (exact CI-wiring mechanism — a new job vs. an addition to an
existing job — to be determined at implementation time), any future regression in `init`'s version-
resolution behavior that produces an unbuildable module is caught by the same harness DX-01 itself is
verified against.

## Compatibility strategy

See `story.md` "Compatibility considerations" — both fixes are treated as strict improvements over
behavior that never actually worked, not as compatibility breaks requiring a deprecation window or
migration path.

## Rollout strategy

Single PR/commit per this story is expected to be feasible, given the bounded scope (one CLI command's
version-resolution logic, one template token, one test assertion, one new test harness). No phased
rollout or feature flag is required — these are CLI-tool-invocation-time fixes, not runtime-service
behavior changes requiring gradual exposure.

## Rollback strategy

Revert the `init_cmd.go` changes (T1-T4) independently of the template/test changes (T003) if either
surfaces an unexpected regression; the two fixes are unrelated in mechanism and can be reverted
independently without affecting each other. If T002's harness itself has a defect that produces false
failures, it can be reverted or bypassed without reverting the underlying T001/T003 fixes, provided the
acceptance criteria's evidence is instead captured by a manual run of the same steps until the harness
defect is fixed.

## Implementation sequence

T001 (DX-01 T1-T4, the version-resolution flags and fallback removal) and T002 (DX-01 T5, the harness)
are expected to be implemented together or with T002 immediately following T001, since T002's harness is
the only way to prove T001's acceptance criteria (AC-W01-E04-S001-01/-02) end-to-end. T003 (DX-02 verb
fix) is independent of T001/T002 and may be implemented in parallel. T004 (generator-output-boots test)
depends on both T002 (reuses its harness) and T003 (tests its fix) and must be implemented after both.

## Task breakdown

- **W01-E04-S001-T001** — Version-resolution flags (DX-01 T1-T4): `--framework-version`,
  `--local-framework`, VCS-derived pseudo-version default, `v0.0.0` fallback deletion.
- **W01-E04-S001-T002** — Isolated-temp-dir E2E scaffold harness (DX-01 T5), the shared primitive.
- **W01-E04-S001-T003** — Generator verb fix (DX-02): `.delete` → `.deactivate` plus the
  `TestGenCRUDPermissionKeys` fix.
- **W01-E04-S001-T004** — Generator-output-boots CI test, reusing T002's harness, testing T003's fix.

## Expected artifacts

Updated `internal/cli/init_cmd.go`; a new isolated-temp-dir E2E harness (exact file TBD); updated
`internal/cli/templates/crud/resource.go.tmpl`; updated `internal/cli/scaffold_test.go`; a new
generator-output-boots test file (exact file TBD, reusing the harness).

## Expected evidence

`DX-01/t1-flag-verify.json` through `DX-01/t5-e2e-temp-dir.json`; `DX-02/w0-t2-verb-fix.json`
(equivalent evidence naming) — see `evidence/index.md` for the full planned register.

## Unresolved questions

- Exact new-file location(s) for the isolated-temp-dir harness and the generator-output-boots test
  within `internal/cli/`'s test-file layout (a single helper function vs. a small internal test-support
  package) — to be determined at implementation time per mandate §8.5's instruction not to invent file
  paths the repository does not yet determine.
- Exact wording of the three fail-closed remediation messages (unresolvable `--framework-version`,
  invalid `--local-framework` path, dirty/unreachable-commit default) — drafted during implementation,
  constrained by the "user-facing-clear, exact copy-pasteable command" requirement in `story.md`
  "Documentation requirements."
- Exact reachability check for DX-01 T3's "commit is reachable" condition (e.g. `git merge-base
  --is-ancestor` against a remote ref, vs. an alternative VCS-status inspection) — to be determined at
  implementation time.
- Exact mechanism for distinguishing "released CLI" from "source-built CLI" within T002's harness in a
  test environment (e.g. building the test binary twice with different `-ldflags` version-injection
  values) — to be determined at implementation time.
- Whether `initData`'s existing struct needs new fields, or whether the `replace`-directive/version-line
  distinction can be expressed via the existing template-and-data shape — to be confirmed by reading the
  current `initData` struct and `go.mod.tmpl` at implementation time.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above are answered by a
first read of `init_cmd.go`'s full current implementation, `initData`'s struct definition, and the
`go.mod` template at story start; (b) the fail-first tests described under "Testing strategy" are
written and confirmed to reproduce the current-broken state for both DX-01 and DX-02 before any fix is
applied; and (c) the owner and reviewer are assigned.
