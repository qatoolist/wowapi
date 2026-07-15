---
id: W01-E04-S001-T001
type: task
title: Version-resolution flags (DX-01 T1-T4)
status: done
parent_story: W01-E04-S001
owner: W01GenDX01
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E04-S001-01
artifacts:
  - ART-W01-E04-S001-001
evidence:
  - EV-W01-E04-S001-001
---

# W01-E04-S001-T001 — Version-resolution flags (DX-01 T1-T4)

## Task Definition

### Task objective

Replace `internal/cli/init_cmd.go`'s unconditional `devel` → `v0.0.0` fallback (lines 122-123) with a
pre-write version-resolution sequence: an explicit `--framework-version` flag verified via `go list -m`,
an explicit `--local-framework` flag emitting a `replace` directive, and — when neither is supplied — a
VCS-derived pseudo-version default that fails closed (never falls back to a placeholder) on a dirty or
unreachable commit.

### Parent story

W01-E04-S001 — Generator correctness — source-built CLI path validity and boot-safe CRUD generation.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Read `internal/cli/init_cmd.go`'s current `runInit`-shaped function in full (the exact function name
   and its surrounding flag-parsing/file-writing structure) to confirm the exact insertion point for a
   pre-write verification step, and read the `initData` struct plus the `go.mod` template it feeds to
   confirm what data shape a `replace` directive or a verified version-line requires.
2. Add a `--framework-version` string flag. When set: before any file is written, invoke `go list -m
   <framework-module>@<version>` (or the equivalent resolution check) and fail with a non-zero exit and
   an exact remediation command (e.g. how to list available versions) if it does not resolve. On
   success, use the verified version as the `go.mod` require-line version.
3. Add a `--local-framework` string flag. When set: before any file is written, validate the path is
   absolute and exists; fail with remediation if not. On success, populate the template data so the
   generated `go.mod` emits an explicit `replace` directive pointing at the given path, and print a
   visible dev-mode warning.
4. When neither flag is set, derive a VCS pseudo-version: determine the current commit, confirm the
   working tree is clean and the commit is reachable (exact reachability mechanism to be determined —
   see `plan.md` "Unresolved questions"). On success, use the derived pseudo-version. On a dirty or
   unreachable commit, fail closed before any file write, with an exact remediation command (e.g. commit
   or pass `--framework-version`/`--local-framework` explicitly).
5. Delete `init_cmd.go:122-123`'s `if fwVer == "devel" { fwVer = "v0.0.0" }` block once steps 2-4
   collectively cover every case it used to handle. Confirm by inspection (and by the fail-first tests
   below) that no code path in `init_cmd.go` can produce the literal string `v0.0.0` any longer.
6. Write fail-first tests: (a) before this task's changes, a test confirming a `devel`-build `init` with
   no flags currently succeeds and writes `v0.0.0` (documents the bug being removed); (b) after the
   changes, the same invocation on a dirty/unreachable commit fails closed with remediation and writes no
   file; (c) the same invocation on a clean/reachable commit succeeds with a real, non-`v0.0.0`
   pseudo-version; (d) an unresolvable `--framework-version` fails closed pre-write with remediation; (e)
   an invalid `--local-framework` path fails closed pre-write with remediation.

### Expected files or components affected

`internal/cli/init_cmd.go`; possibly the `go.mod` generation template (exact file to be confirmed at
implementation time) if the `replace`-directive/version-line distinction requires a template-side
change beyond the data passed into it.

### Expected output

An `init_cmd.go` with no `v0.0.0` code path remaining, three working version-resolution paths (explicit
version, explicit local path, VCS-derived default), all fail-closed pre-write, each proven by a
fail-before/pass-after (or success/failure) test pair.

### Required artifacts

ART-W01-E04-S001-001 (updated `internal/cli/init_cmd.go`).

### Required evidence

EV-W01-E04-S001-001 (functional-test report covering all four resolution paths: explicit-version-success,
explicit-version-failure, local-framework-success, local-framework-failure, VCS-default-success,
VCS-default-failure — recorded under `DX-01/t1-flag-verify.json` through the T3/T4-equivalent evidence
files per `story.md` "Required evidence").

### Related acceptance criteria

AC-W01-E04-S001-01.

### Completion criteria

Every one of the three resolution paths is proven, by test, to either produce a real resolvable version
(and, only then, proceed to file writes) or fail closed before any file write with an exact remediation
command; no code path can produce `v0.0.0`.

### Verification method

`go test ./internal/cli/... -run TestInit` (or the actual test name(s) chosen at implementation time),
logged output retained as evidence; manual CLI invocation against controlled git-commit states (clean,
dirty, unreachable) to confirm each path independently.

### Risks

Low — this task changes CLI-invocation-time logic only, no runtime-service behavior. The primary risk is
an incomplete reachability check for the VCS-derived default (T3) that either rejects a legitimately
clean/reachable commit or, worse, accepts a commit that should have been rejected — mitigated by the
fail-first test pair covering both the accept and reject cases explicitly.

### Rollback or recovery considerations

Revert `init_cmd.go`'s changes independently of T002/T003/T004 (disjoint files) if an unexpected
regression is found in any of the three resolution paths; the `v0.0.0` fallback deletion (step 5) should
only be reverted together with the rest of this task, since deleting it alone (without the replacement
paths) would leave `init` with no version-resolution behavior at all on a `devel` build with no flags.

## Implementation Record

Implemented 2026-07-13 by W01GenDX01 (DEV-W01-E04-S001-01's follow-up work).

### What was actually implemented

The pre-write version-resolution sequence, extracted into a new `internal/cli/init_version.go` and
invoked from `runInit` after the cheap arg/target checks and BEFORE the first file write:

- **T1** — `--framework-version vX.Y.Z`: verified via `go list -m -json <module>@<version>` (run inside
  a throwaway one-off module so the target dir and any enclosing module's replace directives cannot
  influence resolution; `GOWORK=off`). Unresolvable → exit 1 pre-write with the exact discovery command
  `go list -m -versions github.com/qatoolist/wowapi` in the error.
- **T2** — `--local-framework /abs/path`: rejects non-absolute, nonexistent, and non-wowapi-checkout
  paths (the path's own `go.mod` must declare the framework module) pre-write; on success the generated
  `go.mod` gains an explicit `replace` directive (new conditional block in
  `templates/init/go.mod.tmpl`) with the require line carrying Go's canonical inert placeholder
  `v0.0.0-00010101000000-000000000000` (only ever written together with the replace that satisfies it),
  plus a visible dev-mode warning on stderr. Both flags together → rejected (mutually exclusive).
- **T3** — no flags: the version stamped into the binary is classified BY SHAPE, which covers both the
  story's cited `devel` case and the Go 1.24+ stamped-pseudo-version case discovered at execution (the
  SF-7 `+dirty` shape — see Deviations below): a tagged release (`go install …@vX.Y.Z`) is used as-is;
  a `…+dirty` stamp fails closed (dirty tree); a clean stamped pseudo-version is verified resolvable
  via `go list -m` pre-write; an unstamped `devel` build derives the canonical version from the
  `vcs.revision` build setting via `go list -m <module>@<revision>` — which simultaneously derives the
  exact canonical pseudo-version (or tag) AND proves reachability, resolving `plan.md`'s open
  reachability-mechanism question — and fails closed when the binary has no VCS stamp, was built from a
  dirty tree (`vcs.modified`), or the commit does not resolve (unpushed).
- **T4** — `init_cmd.go`'s `if fwVer == "devel" { fwVer = "v0.0.0" }` block is deleted;
  `grep 'fwVer = "v0.0.0"' internal/cli/init_cmd.go` → 0 occurrences. No code path writes the bare
  unresolvable placeholder.

Test seams (`initBuildVersion`, `initVCSInfo`, `resolveModuleVersion` package vars) make every branch
drivable hermetically; the unresolvable-version CLI test additionally drives the REAL `go list`
subprocess under `GOPROXY=off`.

### Components changed

`internal/cli` only. `internal/buildinfo` read (Version, ModulePath, FindGoMod), not modified.

### Files changed

- `internal/cli/init_version.go` (new — resolution logic, seams, remediation wording)
- `internal/cli/init_cmd.go` (two new flags, usage text, pre-write resolution call, fallback deleted,
  `initData.LocalFramework` field)
- `internal/cli/templates/init/go.mod.tmpl` (conditional dev-mode `replace` block)
- `internal/cli/init_version_test.go` (new — 15 tests: all three paths, success + fail-closed each,
  plus the SF-7 stamped-shape regression tests)
- `internal/cli/scaffold_test.go` (`callInit` now pins `--local-framework <this checkout>` for tests
  that are not about version resolution — a devel test binary has no VCS stamp, so flag-less init
  correctly fails closed; `buildRenderedProduct` simplified onto the new flag, dropping its manual
  replace-append; new shared `wowapiCheckoutRoot` helper)

### Interfaces introduced or changed

Two new CLI flags on `wowapi init` (`--framework-version`, `--local-framework`) — additive. The
no-flags default changes from "always succeed, possibly with a broken value" to "verified value or
fail closed" per `story.md` "Compatibility considerations".

### Configuration changes

None beyond the two flags.

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable* (fail-closed hygiene improvement only, per `story.md`).

### Observability changes

The three fail-closed remediation messages (finalized wording lives in `init_version.go`), each
containing a copy-pasteable command, per `story.md` "Documentation requirements".

### Tests added or modified

New `init_version_test.go`: `TestInitDevelNoVCSInfoFailsClosed`, `TestInitDevelDirtyTreeFailsClosed`,
`TestInitDevelUnreachableCommitFailsClosed`, `TestInitDevelCleanReachableDerivesVersion`,
`TestInitFrameworkVersionUnresolvableFailsClosed` (real subprocess, GOPROXY=off),
`TestInitFrameworkVersionVerifiedIsWritten`, `TestInitLocalFrameworkRelativePathFailsClosed`,
`TestInitLocalFrameworkNonexistentPathFailsClosed`, `TestInitLocalFrameworkNonFrameworkDirFailsClosed`,
`TestInitLocalFrameworkWritesReplaceAndWarns`, `TestInitBothVersionFlagsRejected`,
`TestGoResolveModuleVersionFromModuleCache`, `TestInitStampedDirtyVersionFailsClosed` (SF-7),
`TestInitStampedPseudoVersionVerifiedBeforeWrite` (2 subtests), `TestInitStampedReleaseVersionUsedAsIs`.
Every failure test also asserts ZERO files were written (fail-closed-pre-write).

### Commits

None — uncommitted working-tree delta on HEAD 05dce5c8 (conductor owns commits).

### Pull requests

None (conductor owns commits/PRs).

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

- The clean-devel default path requires the go tool to resolve the framework module (network or
  module cache); by design it fails closed rather than degrading.
- `--local-framework` validates the checkout by module path (`go.mod` declares
  `github.com/qatoolist/wowapi`), not by content/version compatibility — DX-05 T4's version gate
  (W01-E04-S002) owns compatibility checking.

### Follow-up items

None owed by this task.

### Relationship to the approved plan

Matches `plan.md` steps 1-4 with two recorded resolutions of its "Unresolved questions": (a) the
reachability check is `go list -m <module>@<revision>` (resolution ≡ reachability, and it returns the
canonical version, making hand-construction of pseudo-versions unnecessary); (b) `initData` gained one
field (`LocalFramework`) and `go.mod.tmpl` one conditional block. One scope extension beyond the plan's
literal text, same defect class: the Go 1.24+ stamped-version (`+dirty`) shape — recorded as
DEV-W01-E04-S001-04.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S001-01 | Fail-first functional tests across all three resolution paths (explicit-version, explicit-local-path, VCS-default), success and failure cases each; manual CLI demos (go run, dirty-tree go build, --local-framework) | Local dev, go1.26.5, controllable stamped-version/VCS states via test seams; GOPROXY=off for the real-subprocess failure test | All failure paths fail closed pre-write (zero files) with remediation; all success paths write a real, verified version; no bare-`v0.0.0` code path remains | functional-test report | pending — wave review gate |

### Actual result

All 15 tests + 2 subtests PASS (`evidence/DX-01/t1-t4-tests-postfix.log`). Fail-first baseline captured
pre-fix: flag-less init SUCCEEDED and wrote `v0.0.0` (go run) and `v1.0.1-0.20260713072141-05dce5c8a548+dirty`
(dirty-tree go build), with `go mod download` failing only afterwards (`unknown revision`) —
`evidence/DX-01/t1-t4-prefix-failfirst.log`. Post-fix CLI demos: both shapes fail closed pre-write with
remediation, zero files; `--local-framework` succeeds with replace + warning —
`evidence/DX-01/t1-t4-postfix.log`. Full package regression `go test ./internal/cli/ -count=1` → ok.

### Pass or fail

PASS.

### Evidence identifier

EV-W01-E04-S001-001 (`evidence/DX-01/t1-flag-verify.json`).

### Execution date

2026-07-13 (~07:50–08:05 UTC).

### Commit or revision

HEAD 05dce5c8a548f7dce3222637ab2c82024236a2a0; fix uncommitted on top (conductor commits).

### Environment

macOS Darwin 25.5.0 arm64, go1.26.5, local dev workstation.

### Reviewer

Pending — wave-level review gate (conductor assigns).

### Findings

1. The live defect's dominant shape on modern Go is NOT the story's cited `devel`→`v0.0.0` arm but the
   Go 1.24+ stamped `+dirty` pseudo-version (SF-7): `buildinfo.Version()` returns the stamp, so the
   original fallback never fired for `go build`-built CLIs — init wrote the unresolvable stamp
   verbatim. Both shapes captured fail-first and both now fail closed (DEV-W01-E04-S001-04).
2. This machine's `GOPRIVATE=github.com/qatoolist/*` would bypass module proxies for the framework —
   irrelevant to T001 (resolution failure still fails closed) but material to T002's harness design.

### Retest status

Not required — verified first-pass at the pinned revision.

### Final conclusion

AC-W01-E04-S001-01 verified with preserved fail-first evidence; no `v0.0.0` fallback path remains.

## Deviations Record

See story-level `deviations.md` DEV-W01-E04-S001-04 (same-defect-class scope extension: the Go 1.24+
stamped-version shapes — tagged release pass-through, verified clean pseudo-version, fail-closed
`+dirty` — handled by the no-flags path beyond the plan's literal `devel`-only framing; plus the
test-suite accommodation: `callInit` pins `--local-framework` for non-resolution tests). No other
deviation from the approved plan.
