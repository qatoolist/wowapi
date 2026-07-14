---
id: W01-E04-S001-T004
type: task
title: Generator-output-boots CI test
status: done
parent_story: W01-E04-S001
owner: W01Gen
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W01-E04-S001-T002
  - W01-E04-S001-T003
acceptance_criteria:
  - AC-W01-E04-S001-04
artifacts:
  - ART-W01-E04-S001-005
evidence:
  - EV-W01-E04-S001-004
---

# W01-E04-S001-T004 — Generator-output-boots CI test

## Task Definition

### Task objective

Add a CI test that generates a `gen crud` module (reusing T002's isolated-temp-dir harness), attempts to
boot it, and asserts no closed-authorization-verb-set rejection occurs. This is the PF-2-suggested fix
vehicle from the wowsociety upstream register, and is explicitly fail-first: it must fail today, before
T003's template fix, with exactly the closed-verb-set rejection at `kernel/authz/registry.go:88-90`, and
pass after T003 lands.

### Parent story

W01-E04-S001 — Generator correctness — source-built CLI path validity and boot-safe CRUD generation.

### Owner

W01Gen (wave-01 generator worker)

### Status

done

### Dependencies

**W01-E04-S001-T002** (reuses its isolated-temp-dir generate→build→boot→smoke harness — this task does
not reimplement the harness, only calls it with `gen crud` as the generator command under test).
**W01-E04-S001-T003** (this task exists specifically to prove T003's fix; it must be run in both the
pre-T003 state, to establish the fail-first baseline, and the post-T003 state, to confirm the fix).

### Detailed work

1. Confirm the exact failure mode expected before the fix: `kernel/authz/registry.go:83-90`'s `Register`
   method rejects any permission whose action segment is outside the closed verb set
   (`registry.go:15-19`), appending a `kerr.KindInternal` "invalid_permission... not in the closed verb
   set" error, surfaced via `Registry.Err()` — the module's own boot sequence is expected to gate on this
   error (exact boot call site that checks `Registry.Err()` to be confirmed at implementation time by
   reading the generated module's own `main.go`/bootstrap template).
2. Using T002's harness, write a test that: generates a `gen crud` module (with at least one resource,
   e.g. the harness's own default or an explicit `--resource widget` invocation), attempts to boot it
   (run the generated module's startup path far enough to reach permission registration), and asserts the
   boot either succeeds cleanly or fails with a *different* error than the closed-verb-set rejection.
3. Run this test against the **pre-T003** template state (i.e., run it before T003's fix lands, or by
   temporarily checking out/reverting to the pre-fix template) and confirm it fails, with the failure
   message matching the closed-verb-set rejection at `kernel/authz/registry.go:88-90` specifically — not
   some other unrelated failure. This is the fail-first proof.
4. Run the same test against the **post-T003** template state and confirm it passes.
5. Wire the test into CI (exact CI job/step to be determined at implementation time) so it runs on every
   future change to `internal/cli/templates/crud/` or `kernel/authz/registry.go`'s verb set, serving as a
   permanent regression guard against either side reintroducing this class of defect.

### Expected files or components affected

A new generator-output-boots test file (exact location TBD, within `internal/cli/`'s test infrastructure,
alongside or reusing T002's harness location); CI configuration (exact file/job TBD).

### Expected output

A CI-wired test proven fail-first: red before T003, green after, and permanently guarding against a
future regression in either the generator template or the kernel's verb set producing a boot-time
rejection.

### Required artifacts

ART-W01-E04-S001-005 (the generator-output-boots test).

### Required evidence

EV-W01-E04-S001-004 (functional-test report, fail-before/pass-after pair, with the pre-fix failure
message captured verbatim to confirm it matches the expected closed-verb-set rejection — recorded under
the `DX-02/w0-t2-verb-fix.json`-equivalent evidence naming per `story.md` "Required evidence").

### Related acceptance criteria

AC-W01-E04-S001-04.

### Completion criteria

The test fails, pre-T003, with the exact closed-verb-set rejection at `kernel/authz/registry.go:88-90`
(not a different, unrelated failure); the test passes, post-T003; the test is wired into CI.

### Verification method

Direct test execution against both the pre-T003 and post-T003 template state, logged output (including
the exact pre-fix failure message) retained as evidence.

### Risks

Low-medium — the main risk is a test that fails for the *wrong* reason pre-T003 (e.g. a harness bug, a
different template defect, or a network/environment issue unrelated to the permission verb), which would
falsely appear to validate T003's fix without actually proving it. Mitigation: step 3's explicit
requirement that the pre-fix failure message be confirmed to match the closed-verb-set rejection
specifically, not merely "the test failed somehow."

### Rollback or recovery considerations

If this test proves too brittle for CI (e.g. inherits flakiness from T002's harness), it can be
temporarily marked as a manual-only verification step while the underlying harness flakiness is
addressed, rather than being deleted — since it is this story's primary regression guard against DX-02
resurfacing.

## Implementation Record

### What was actually implemented

New test `TestGenCRUDOutputBoots` in `internal/cli/gen_crud_boots_test.go`. Flow: (1) scaffold a
product wired to this checkout via the existing shared scaffold primitive `buildRenderedProduct`
(`internal/cli/scaffold_test.go:568` — init → replace-to-local-checkout → `go mod tidy`; NOT
reimplemented, per this task's dependency note); (2) `wowapi new-module --name widgets`; (3)
`wowapi gen crud --resource widget` into it; (4) extract the RouteMeta permission keys VERBATIM from
the generated `widget.go` and declare them in the module's `seeds/permissions.yaml` — derived from
generator output, never hardcoded, so the boot proof always tracks what the template actually emitted;
(5) wire `RegisterWidgetRoutes(mc)` into `Module.Register` (the module template's own TODO step) and
the module into `internal/wire/modules.go`; (6) write a product-side `internal/boottest/boot_test.go`
that boots exactly as the generated binaries do (`app.New()` → `Register(wire.Modules()...)` →
`Boot`), with a no-op `TxManager` stub satisfying `kernel.Deps.Tx` (Boot's registration validation —
permission key shape, closed verb set, route-permission declaration — runs before any pool use, and
nil pools skip the RLS liveness assertions, so no database is needed); (7) `go test
./internal/boottest/` inside the product, failing the outer test with the verbatim boot error and a
targeted message when it matches the closed-verb-set rejection. A trailing assertion additionally
fails if any generated permission ends in `.delete`, guarding the specific bad key.

### Components changed

`internal/cli` test infrastructure only. No production code changed by this task.

### Files changed

- `internal/cli/gen_crud_boots_test.go` — new file (the test + `extractRoutePermissions` helper).

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

None needed: CI wiring falls out of test placement. `make test-unit` runs `go test ./...` with no
`-short` (Makefile:163-165) and is invoked by the containerized CI legs (Makefile:325-326), so this
test runs on every future change to `internal/cli/templates/crud/` or `kernel/authz/registry.go`'s
verb set. Locally it skips under `-short` like the sibling rendered-product compile tests.

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

One new test, `TestGenCRUDOutputBoots`, reusing the shared scaffold primitive as anticipated.

### Commits

None yet — new uncommitted file on top of HEAD `05dce5c8a548f7dce3222637ab2c82024236a2a0`; the wave
conductor owns commits.

### Pull requests

*Not yet implemented.*

### Implementation dates

2026-07-13 (implemented and verified same day).

### Technical debt introduced

*None anticipated.*

### Known limitations

- The boot runs without a database, so it proves the registration-validation gate (where the DX-02
  rejection lives), not migrations/seed-sync against Postgres — those are testkit.RunModuleContract /
  internal/e2e territory (DB-gated).
- T002's full T5 harness (released-CLI vs source-built-CLI paths, `go mod download`, smoke cycle) is
  NOT delivered by this task; this test consumes the same scaffold primitive T002 will build on.
- The first kernel.New attempt with empty Deps panicked (workflow.NewRuntime requires a TxManager) —
  resolved with the no-op stub; recorded here so a future harness author doesn't rediscover it.

### Follow-up items

None for this task.

### Relationship to the approved plan

Matches `plan.md` step 7 (generate → boot → assert no closed-verb rejection, fail-first against the
pre-T003 template, pass after). The plan left the file location TBD; resolved to a single new test
file in `internal/cli/` alongside the primitive it reuses, the lightest shape that satisfies reuse.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S001-04 | Run the generator-output-boots test against pre-T003 and post-T003 template state | Isolated temp dir (via T002's harness), Go toolchain, `kernel/authz` boot path exercised | Fails pre-T003 with the exact closed-verb-set rejection at `kernel/authz/registry.go:88-90`; passes post-T003 | functional-test report (fail-before/pass-after pair) | unassigned |

### Actual result

Pre-T003 (unfixed template): FAILED with the verbatim boot error `boot: app: boot validation failed:
permission registration failed: permission action %q is not in the closed verb set:
widgets.widget.delete` — exactly the `kernel/authz/registry.go:88-90` rejection, not a generic
failure (`evidence/DX-02/t004-boots-prefix-failfirst.log`). Post-T003: PASSED
(`evidence/DX-02/t004-boots-postfix.log`). Full `go test ./internal/cli/ -count=1` passes (15.2s,
`evidence/DX-02/pkg-internal-cli-full.log`).

### Pass or fail

PASS (fail-before/pass-after pair complete, pre-fix failure message verified against the expected
rejection).

### Evidence identifier

EV-W01-E04-S001-004 (`evidence/DX-02/w0-t2-boots-test.json`).

### Execution date

2026-07-13 (07:26 UTC).

### Commit or revision

HEAD `05dce5c8a548f7dce3222637ab2c82024236a2a0`; test + fix uncommitted on top (conductor commits).

### Environment

macOS Darwin 25.5.0 arm64, go1.26.5, isolated `t.TempDir()` product scaffold; offline after the
module-cache-priming tidy.

### Reviewer

Pending — wave-level review gate (conductor assigns).

### Findings

An intermediate fail-for-the-wrong-reason run occurred exactly as this task's "Risks" section warned:
the first boot attempt panicked in `kernel.New` (nil TxManager), NOT the closed-verb rejection. The
test's failure-message discrimination caught it ("not the closed-verb rejection — investigate
separately"), the stub TxManager fixed it, and the re-run produced the genuine closed-verb failure.
The mitigation designed into the task worked as intended.

### Retest status

Not required — fail-first and pass runs both captured at the pinned revision.

### Final conclusion

AC-W01-E04-S001-04 satisfied. The test is a permanent CI regression guard on both the CRUD template
and the kernel verb set.

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
