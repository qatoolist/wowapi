---
id: W01-E04-S001
type: story
title: Generator correctness — source-built CLI path validity and boot-safe CRUD generation
status: accepted
wave: W01
epic: W01-E04
owner: W01Gen (DX-02 slice + T005), W01GenDX01 (DX-01 slice)
reviewer: unassigned
priority: P0
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DX-01
  - DX-02
depends_on: []
blocks:
  - W01-E04-S002
  - W06-E01-S002
acceptance_criteria:
  - AC-W01-E04-S001-01
  - AC-W01-E04-S001-02
  - AC-W01-E04-S001-03
  - AC-W01-E04-S001-04
  - AC-W01-E04-S001-05
artifacts:
  - ART-W01-E04-S001-001
  - ART-W01-E04-S001-002
  - ART-W01-E04-S001-003
  - ART-W01-E04-S001-004
  - ART-W01-E04-S001-005
  - ART-W01-E04-S001-006
evidence:
  - EV-W01-E04-S001-001
  - EV-W01-E04-S001-002
  - EV-W01-E04-S001-003
  - EV-W01-E04-S001-004
  - EV-W01-E04-S001-005
decisions: []
risks:
  - RISK-W01-005
  - RISK-W01-E04-001
---

# W01-E04-S001 — Generator correctness — source-built CLI path validity and boot-safe CRUD generation

## Story ID

W01-E04-S001

## Title

Generator correctness — source-built CLI path validity and boot-safe CRUD generation

## Objective

Make `wowapi init` fail closed instead of silently writing an unresolvable framework-version pin when
run from a source (`devel`) build, and make `wowapi gen crud` emit a permission verb that is actually
inside the kernel's closed authorization-verb set — so that both of the CLI's two primary code-
generation paths produce a module that boots, proven by a real generate→build→boot→smoke cycle, not by
hand-inspection of the generated output.

## Value to the framework

A generator whose output doesn't boot is worse than no generator — it actively erodes trust. The whole
value proposition of `wowapi init`/`wowapi gen crud` is that a developer can run one command and get a
working starting point; every defect that makes generated output fail at `go build` or at boot converts
that value proposition into a liability, because the developer's first interaction with the framework is
a failure they did not cause and cannot diagnose from the generator's own output. Both defects fixed
here are exactly this class of failure: DX-01 silently writes a dependency line that cannot resolve, and
DX-02 silently writes a permission key that the kernel's own registration path rejects at boot. Fixing
them, and proving the fix with an end-to-end generate→build→boot→smoke harness rather than a unit test
that only inspects template output, converts "the generator produced text that looked right" into "the
generator produced a module that actually runs" — the only claim a generator should be allowed to make.

## Problem statement

Two independent, unrelated-in-mechanism defects, both classified as generator-correctness gaps and both
targeted at this story by the canonical allocation in `requirement-inventory.md`:

**DX-01 — source-built CLI path validity.** `internal/cli/init_cmd.go:122-123`:

```go
fwVer := buildinfo.Version()
if fwVer == "devel" {
    fwVer = "v0.0.0"
}
```

When `wowapi init` runs from a source (non-release) build — the exact situation for any contributor
building the CLI from a git checkout rather than downloading a tagged release — `buildinfo.Version()`
returns `"devel"`, and the fallback unconditionally substitutes the literal string `v0.0.0`. This value
is templated into the generated `go.mod`'s framework-module require line with no resolvability check
against the actual module-proxy/VCS state (`go list -m` is never invoked before the file write). `v0.0.0`
is not a version any real tagged or pseudo-versioned commit of `github.com/qatoolist/wowapi` will ever
resolve to, so `go mod download`/`go build` on the generated module fails immediately, unless the
developer already happens to have a manual `replace` directive in place for unrelated reasons. The
defect is a silent one: `init` reports success, writes all files, and only the *next* command
(`go build`) surfaces the failure, by which point the developer has no direct signal connecting the
build failure back to the version string `init` chose.

**DX-02 — generator emits an out-of-set permission verb (Wave-0 slice).**
`internal/cli/templates/crud/resource.go.tmpl:54`:

```go
r.Handle("DELETE", "/{{.Resource}}/{id}", httpx.RouteMeta{Permission: "{{.PermPrefix}}.delete"}, h.delete)
```

emits a `RouteMeta.Permission` whose action segment is the literal string `delete`. The kernel's closed
authorization-verb set, `kernel/authz/registry.go:15-19`:

```go
var verbs = map[string]bool{
    "create": true, "read": true, "list": true, "update": true,
    "deactivate": true, "restore": true, "approve": true, "reject": true,
    "assign": true, "export": true, "admin": true, "ingest": true, "activate": true,
}
```

does not contain `delete`. `registry.go:83-90`'s `Register` method checks every permission's action
segment against this set and appends a `kerr.KindInternal` error — surfaced via `Registry.Err()` at
boot — for any action outside it. The consequence: every single `gen crud` invocation produces a module
that is dead-on-arrival at boot, unconditionally, because the DELETE route's permission registration
fails the closed-set check every time. This is not an edge case triggered by unusual input; it is the
generator's own, always-present output.

Compounding this, the generator's own test currently locks the bug in as the *expected* behavior:
`internal/cli/scaffold_test.go:937-953`'s `TestGenCRUDPermissionKeys` asserts that the generated file
contains the literal string `"widgets.widget.delete"` (line 949) as one of five permission keys the test
requires to be present. A template-only fix would leave this test red (correctly flagging the fix as a
"regression" against its own now-wrong assertion) unless the test is fixed in the same change — or worse,
if the test is treated as ground truth and the fix is reverted to make the (wrong) test pass again. Both
the template and its own test must be corrected together, or the defect resurfaces immediately. This is
exactly RISK-W01-005 (see "Risks" below).

The template's own DELETE handler already carries a TODO acknowledging the correct semantics:
`resource.go.tmpl:146`, `// TODO: soft-delete/status transition inside h.tx.WithTenant.` — the intended
behavior was always a status transition (a `deactivate`), never a hard delete. The permission string is
the only place the template disagrees with its own documented intent; the fix is a one-token correction
that brings the permission key into alignment with behavior the template author had already decided on.

## Source requirements

DX-01, DX-02 (Wave-0 slice only, per this epic's `epic.md` scoping — the full P1/Wave-4 generator
rewrite is explicitly out of scope here, see "Out of scope" below).

## Current-state assessment

- `internal/cli/init_cmd.go` has no `--framework-version` or `--local-framework` flag today (confirmed
  by inspection: no such flag is registered anywhere in `internal/cli/`). The only version-selection
  logic is the two-line `devel` → `v0.0.0` fallback at lines 122-123 quoted above, unconditional and
  unverified.
- `wowapi init` performs no `go list -m` (or any other resolvability) check before writing any generated
  file. Files are written first; resolvability is never checked at all by `init` itself — the developer
  discovers the problem only when they subsequently run `go build`/`go mod download` against the
  generated module.
- `internal/cli/templates/crud/resource.go.tmpl:54` emits `Permission: "{{.PermPrefix}}.delete"` for the
  generated DELETE route. `kernel/authz/registry.go:15-19` defines the closed verb set without `delete`;
  `registry.go:83-90`'s `Register` rejects any permission whose action segment is outside that set, and
  this rejection surfaces via `Registry.Err()`, which is expected to gate boot (per `registry.go`'s own
  doc comment: "an unknown permission can never silently allow" — the registry is designed to fail
  loudly, not silently, when this happens).
- `internal/cli/scaffold_test.go:937-953`'s `TestGenCRUDPermissionKeys` currently asserts
  `"widgets.widget.delete"` (line 949) as one of the five permission keys required to be present in
  generated output — i.e., the test is currently written to treat the bug as correct behavior.
- No generate→build→boot→smoke test exists today for either the released-CLI or the source-built-CLI
  path. `internal/cli/scaffold_test.go`'s existing generator tests inspect generated file *contents*
  (string assertions like `TestGenCRUDPermissionKeys` above) — none of them actually run `go build` or
  boot the generated module, so neither of these two defects would be caught by the existing test suite
  as it stands.
- wowsociety is confirmed unaffected by either defect: DX-01 does not affect wowsociety because its
  `replace => ../wowapi` path-replace mechanism (a manually maintained local-path replace, unrelated to
  `init`'s generated `require` line) never touches the CLI-generated dependency line DX-01 fixes.
  wowsociety already independently discovered and documented DX-01's exact underlying bug in
  `wowsociety/docs/upstream/12-sf-7-init-gomod-invalid-and-gitignored-local-overlay.md` — this is
  informational corroboration of the defect, not a dependency of this story. DX-02 does not affect
  wowsociety because `wowsociety/docs/CONVENTIONS.md:10`'s governance rule ("never bypass the generator,
  file an RFF instead") kept every existing wowsociety module immune — no module was freshly generated
  with `gen crud` since the buggy verb was introduced, so no wowsociety module carries the bad permission
  key today.

## Desired state

`wowapi init`, on any build (release or source/`devel`), resolves an exact framework version *before*
writing any file: either the caller supplies `--framework-version vX.Y.Z` and `init` verifies it via
`go list -m` before any write; or the caller supplies `--local-framework /absolute/path` and `init`
emits an explicit `replace` directive plus a visible dev-mode warning; or, with neither flag passed,
`init` derives an exact pseudo-version from VCS metadata when the current commit is reachable and the
working tree is clean. If none of these three paths succeeds — an unresolvable explicit version, an
invalid `--local-framework` path, or a dirty/unreachable commit with neither flag given — `init` fails
closed, before any file write, with an exact remediation command in the error. The `v0.0.0` fallback
path is deleted from the codebase entirely; there is no longer any code path that can silently write an
unresolvable version. Both the released-CLI and source-built-CLI paths are proven end-to-end by a real
generate→build→boot→smoke test running in an isolated temporary directory.

`gen crud` emits `{{.PermPrefix}}.deactivate` instead of `{{.PermPrefix}}.delete` for the DELETE route's
permission, matching the closed authorization-verb set and the template's own existing soft-delete TODO.
`TestGenCRUDPermissionKeys` asserts the corrected verb. A new generator-output-boots CI test — reusing
the same isolated-temp-dir harness DX-01's fix builds — generates a `gen crud` module, boots it, and
asserts no closed-verb-set rejection occurs; this test is written fail-first (it fails today, before the
template fix, with exactly the closed-verb-set rejection described above, and passes after).

## Scope

- **DX-01 T1** — Add a `--framework-version vX.Y.Z` flag to `wowapi init`; verify the supplied version
  resolves via `go list -m` *before* any generated file is written; on an unresolvable version, fail with
  an exact remediation command (e.g. the correct `go list -m` invocation to discover valid versions),
  before any write occurs.
- **DX-01 T2** — Add a `--local-framework /absolute/path` flag; on use, emit an explicit `replace`
  directive into the generated `go.mod` pointing at the given path, plus a visible dev-mode warning in
  `init`'s output; reject a non-absolute or nonexistent path before any write.
- **DX-01 T3** — When neither `--framework-version` nor `--local-framework` is passed, derive an exact
  pseudo-version from VCS metadata (matching Go's own pseudo-version scheme) when the current commit is
  reachable (pushed/tagged-ancestor, not purely local-only) and the working tree is clean. When the
  commit is dirty or unreachable, fail closed with remediation — this path never falls back to
  `v0.0.0`.
- **DX-01 T4** — Delete the `v0.0.0` fallback code path (`init_cmd.go:122-123`) entirely; no code path in
  `init` may ever write an unresolvable placeholder version again.
- **DX-01 T5** — Build a real generate→build→boot→smoke test harness that runs in an isolated temporary
  directory: `init` → `go mod download` → `go build` → contract/smoke tests → success, end to end,
  covering both the released-CLI invocation path and the source-built-CLI (`devel`) invocation path.
  This harness is the shared primitive this epic's own governing instruction calls out explicitly:
  DX-01's own row note in `requirement-inventory.md` states "T5 harness = shared primitive for
  DX-02/DX-04." This story's own T004 (the generator-output-boots test, below) reuses it for DX-02; a
  future DX-04 story (golden consumer + upgrade matrix, W06 scope, out of this story's scope) is expected
  to reuse it again — noted here for traceability, not implemented here.
- **DX-02 Wave-0 slice, strictly** — the permission-verb fix (`.delete` → `.deactivate` at
  `resource.go.tmpl:54`) and the fix to the test that currently locks the bug in as correct
  (`TestGenCRUDPermissionKeys`, `scaffold_test.go:937-953`), plus the generator-output-boots CI test
  (generate → boot → assert no closed-verb-set rejection), which reuses DX-01 T5's harness rather than
  building a second one.

## Out of scope

- **DX-02's full P1/Wave-4 generator rewrite** — the disable-vs-minimal-slice decision (whether the
  generated DELETE handler's TODO stub should be disabled by default or left as a minimal functioning
  slice), a status column for soft-deleted resources, and replacing the TODO handlers with real
  implementations are explicitly not part of this story. `requirement-inventory.md`'s DX-02 row and
  this epic's own `epic.md` ("Out of scope") both scope DX-02 at this story to the Wave-0 slice only —
  "one template token + one harness test," in MATRIX CS-14's own framing. The remainder is deferred to
  future work (W06 or later).
- **DX-02's W0-T1 (disable-vs-minimal-slice decision)** — not implemented here; this story's DX-02 scope
  is the verb-fix and the boots-test only, per the epic's explicit instruction not to silently expand
  into other DX-02 Wave-0 sub-tasks.
- **DX-02's W0-T3 (status column)** — not implemented here, for the same reason.
- **DX-02's W0-T4 (replace TODO handlers with real implementations)** — not implemented here, for the
  same reason.
- **DX-04 (golden consumer + upgrade matrix)** — W06 scope, a separate future story. This story's T002
  (the isolated-temp-dir harness) is built so that DX-04 can reuse it later, but DX-04 itself is not
  implemented, planned in detail, or scheduled by this story.
- **Editing the wowsociety upstream register** — the `12-sf-7-init-gomod-invalid-and-gitignored-local-
  overlay.md` finding lives in the wowsociety repository, not wowapi. Per mandate §2.3's framework/
  product boundary, this story can only record the informational observation that wowsociety already
  documented this exact bug (see "Current-state assessment") and recommend — not require — that finding
  be marked resolved once T001 ships; the actual edit to that wowsociety-repository file is out of scope
  for this story (and is FBL-03's concern, tracked in this epic's sibling story W01-E04-S002).

## Assumptions

- `go list -m` is available and correctly configured in the environment `wowapi init` runs in (the same
  assumption any Go-module-aware tool already makes); this story does not add network-availability
  fallback behavior beyond failing closed with a clear remediation message when resolution cannot
  complete.
- The exact wording of the fail-closed remediation messages (for an unresolvable
  `--framework-version`, an invalid `--local-framework` path, and a dirty/unreachable-commit default) is
  not yet fixed by this planning document — per mandate §8.5's instruction not to invent precise
  implementation details the repository does not yet determine, the exact message text is an
  implementation-time decision, constrained only by the requirement that it be user-facing-clear and
  contain an exact, copy-pasteable remediation command (see "Documentation requirements" below).
- Go's pseudo-version scheme (`vX.Y.Z-0.yyyymmddhhmmss-abcdefabcdef`) is assumed to be the correct target
  format for DX-01 T3's VCS-derived default, consistent with how `go mod` itself represents an untagged
  commit; the exact derivation mechanism (shelling out to `git`, or using an existing Go module-tooling
  library) is an implementation-time decision.

## Dependencies

None upstream. This story requires no prior work from any other epic or story in this wave or an earlier
wave beyond the wave-level W00 exit gate (per `../../epic.md`'s "Dependencies" section: "No dependency on
W01-E01/E02/E03... this epic's three stories target disjoint files... and can proceed in any order
relative to them"). This story is itself upstream of other work, not downstream of it:

- **W01-E04-S002** (this epic's sibling story) depends on this story for two of its own task items: its
  FBL-03 task item (PF-2's closure in the wowsociety upstream register is contingent on this story's
  DX-02 task landing) and its DX-05 T4 task item (the `wowapi version` compatibility gate on mutating
  generator commands shares this story's DX-01 version-verification plumbing). Both dependencies are
  recorded in S002's own `story.md`/`dependencies.md`, not duplicated here beyond this note.
- **A future DX-04 story** (golden consumer + upgrade matrix, W06 scope, not part of this wave or this
  epic) is expected to reuse this story's T002 isolated-temp-dir harness, per DX-01's own row note in
  `requirement-inventory.md` ("T5 harness = shared primitive for DX-02/DX-04"). This is a forward
  reference for traceability only — DX-04 is not scheduled or implemented by this story.

## Affected packages or components

`internal/cli/` — specifically `internal/cli/init_cmd.go` (the `init` command's version-resolution
logic), `internal/cli/templates/crud/resource.go.tmpl` (the CRUD generator template), and
`internal/cli/scaffold_test.go` (the generator's own test file, specifically `TestGenCRUDPermissionKeys`
at lines 937-953). A new isolated-temp-dir E2E test harness is added to `internal/cli/`'s test
infrastructure (exact new file location to be determined at implementation time — see `plan.md`).
`kernel/authz/registry.go` is read (to confirm the closed verb set and its exact rejection behavior) but
not modified — the closed-set discipline is intentional and this story does not widen it.

## Compatibility considerations

The fail-closed default (DX-01 T3/T4) is an intentional, compatibility-breaking improvement over the
silently-wrong `v0.0.0` fallback: any workflow that today relies on `init` "succeeding" with an
unresolvable version and only discovering the failure at `go build` time will now see the failure move
earlier, to `init` itself, with a remediation message instead of a bare build error. This is a strict
improvement — the old behavior was never actually correct (it never produced a buildable module from a
`devel` build without a pre-existing manual workaround), so there is no compatible prior behavior being
broken. Any caller that was passing an already-working manual `replace` directive to route around the
`v0.0.0` bug is unaffected, since `--local-framework` now provides the same capability as a first-class,
verified flag. DX-02's verb fix is likewise a strict improvement: no released module can have relied on
the `.delete` permission key actually working, because it never worked (every prior `gen crud` output
was rejected at boot) — there is no working prior state to preserve.

## Security considerations

Largely not applicable. DX-01's fail-closed default is itself a minor security-adjacent improvement in
the general sense that "fail loudly on an invalid/unverifiable input" is preferable to "silently proceed
with garbage input," but this story does not introduce or remove any authentication, authorization, or
data-handling control. DX-02's fix operates entirely within the kernel's own existing, unmodified
authorization-verb enforcement (`registry.go:83-90`) — the fix makes the generator's output comply with
an existing security-relevant control (the closed verb set), it does not alter that control.

## Performance considerations

Not applicable. `go list -m`/VCS-metadata inspection adds a bounded, one-time cost to `init` invocation
(a command that already performs file-system I/O and template execution); this is not a hot path and no
performance budget applies. The generate→build→boot→smoke harness (T002) runs in CI/test time, not in
any production request path.

## Observability considerations

Minor. The fail-closed remediation messages (DX-01) are themselves the primary observability surface
this story adds — a developer must be able to see exactly why `init` failed and exactly what command to
run next, without needing to inspect logs or source. No new metrics, structured logging, or tracing is
required by this story's acceptance criteria.

## Migration considerations

None. No schema, data, or persisted-state migration is involved; this story only affects CLI code-
generation logic and generator templates.

## Documentation requirements

The exact remediation-command wording for each fail-closed path (unresolvable `--framework-version`,
invalid `--local-framework` path, dirty/unreachable-commit default) needs to be user-facing-clear, since
it is the primary UX of the fail-closed path this story introduces — the entire value of failing closed
instead of silently writing `v0.0.0` depends on the developer being told, precisely and actionably, what
to do next. This wording is drafted during implementation (see "Assumptions" above for why it is not
fixed in this planning document) and recorded in `implementation.md` once finalized. No other
documentation-file update (README, CLI reference, etc.) is required by this story's own scope; DX-05 T3's
blueprint-11 CLI-example reconciliation (which would need to reflect the new `--framework-version`/
`--local-framework` flags once they exist) is this epic's sibling story W01-E04-S002's concern, not
duplicated here.

## Acceptance criteria

- **AC-W01-E04-S001-01**: `wowapi init --framework-version vX.Y.Z` with an unresolvable version fails
  before any file is written, with an exact remediation command in the error output. `wowapi init
  --local-framework <path>` with a non-absolute or nonexistent path is rejected before any file is
  written. A clean, reachable-commit `wowapi init` invocation with neither flag passed derives an exact
  VCS pseudo-version and writes it into the generated `go.mod`. A dirty or unreachable-commit `wowapi
  init` invocation with neither flag passed fails closed with an exact remediation command, before any
  file is written — this path never writes `v0.0.0`, and the `v0.0.0` fallback code path no longer
  exists anywhere in `internal/cli/init_cmd.go`.
- **AC-W01-E04-S001-02**: The isolated-temp-dir generate→build→boot→smoke harness (T002) runs a real
  `init` → `go mod download` → `go build` → contract/smoke-test cycle to success, end to end, for both
  the released-CLI invocation path and the source-built (`devel`) CLI invocation path. The harness is
  implemented once and reused (not reimplemented) by T004's generator-output-boots test.
- **AC-W01-E04-S001-03**: `internal/cli/templates/crud/resource.go.tmpl`'s generated DELETE route's
  permission action segment is `deactivate`, not `delete`; `TestGenCRUDPermissionKeys` asserts
  `"widgets.widget.deactivate"`, not `"widgets.widget.delete"`; a fail-before/pass-after run of this test
  demonstrates the fix.
- **AC-W01-E04-S001-04**: The generator-output-boots CI test (T004), reusing T002's harness, generates a
  `gen crud` module, attempts to boot it, and asserts no closed-authorization-verb-set rejection occurs.
  This test is proven fail-first: it fails today (before the T003 template fix) with exactly the
  closed-verb-set rejection at `kernel/authz/registry.go:88-90`, and passes after T003 lands.
- **AC-W01-E04-S001-05** *(added 2026-07-13 — conductor-approved scope addition, DEV-W01-E04-S001-03)*:
  a pristine `wowapi init` scaffold's `configs/` passes `wowapi config validate --env local` under the
  framework-only validation path (no active product-owned keys in scaffolded config files; product
  sections ship as commented examples). Proven fail-first by `TestInitScaffoldConfigValidates`.

## Required artifacts

- Updated `internal/cli/init_cmd.go` (version-resolution flags, VCS-pseudo-version derivation, fallback
  removal).
- New isolated-temp-dir E2E test harness (Go test helper/package under `internal/cli/`'s test
  infrastructure; exact file location determined at implementation time — see `plan.md`).
- Updated `internal/cli/templates/crud/resource.go.tmpl` (verb fix).
- Updated `internal/cli/scaffold_test.go` (`TestGenCRUDPermissionKeys` corrected assertion).
- New generator-output-boots CI test (reusing the harness above).
See `artifacts/index.md`.

## Required evidence

- `DX-01/t1-flag-verify.json` through `DX-01/t5-e2e-temp-dir.json` — per-task evidence for the
  version-resolution flags (T1/T2), the VCS-pseudo-version default and fallback removal (T3/T4), and the
  end-to-end harness run (T5), covering both released-CLI and source-built-CLI paths.
- `DX-02/w0-t2-verb-fix.json` (equivalent evidence naming) — fail-before/pass-after evidence for the
  template fix, the corrected `TestGenCRUDPermissionKeys` run, and the generator-output-boots test's
  fail-before/pass-after pair.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md` and
`plan.md` complete, all four acceptance criteria numbered and measurable, dependencies (none upstream)
recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`: implementation
matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance criteria verified
with evidence in `evidence/index.md`; `closure.md` completed; independent review passed per mandate §14,
with specific confirmation that `TestGenCRUDPermissionKeys` was actually updated (RISK-W01-005), not only
the template.

## Risks

RISK-W01-005 — the generator fix (DX-02 `.delete`→`.deactivate`) must also fix the generator's own test
that currently asserts the buggy verb as correct (`TestGenCRUDPermissionKeys`); missing this sub-fix
would leave the test suite red, or worse, invite a well-intentioned but wrong "fix" that reverts the
template change to make the (incorrect) test pass again. See epic-level `../../risks.md` for the full
wave-scoped risk entry (likelihood Low, impact Medium, severity Low-medium); this story's T003 task
explicitly calls out this risk in its own scope (see `tasks/task-003-generator-verb-fix.md`).

## Residual-risk expectations

Once T003's dual fix (template + test) and T004's fail-before/pass-after generator-output-boots test are
both verified, no residual risk is expected to remain open at acceptance for DX-02. For DX-01, the
primary residual-risk surface is the exact wording of the fail-closed remediation messages (see
"Assumptions" and "Documentation requirements") — this is expected to be resolved during implementation,
not left open at acceptance, but is flagged here as the one area where this planning document
intentionally defers a precise decision per mandate §8.5.

## Plan

See `plan.md`.
