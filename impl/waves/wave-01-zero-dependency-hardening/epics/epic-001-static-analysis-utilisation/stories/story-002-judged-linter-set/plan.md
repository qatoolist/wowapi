---
id: PLAN-W01-E01-S002
type: plan
parent_story: W01-E01-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W01-E01-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information — most notably, G115's exact site list and usestdlibvars's site list,
neither of which exist as a file:line enumeration in the source material.

## Proposed architecture

No architectural change. This story is linter configuration plus per-site annotations and small,
behavior-preserving code fixes. No new package, interface, or contract is introduced. The one
judgment-bearing design element is *how* each gosec/exhaustive/nilerr hit is annotated (comment syntax
and content), not a structural change to the code.

## Implementation strategy

1. Re-run gosec, errorlint, exhaustive, forcetypeassert, and usestdlibvars fresh, at this story's
   actual start commit, force-enabled via a one-off `golangci-lint run --enable=...` invocation (not
   yet via `.golangci.yml`) — the fail-first evidence step, exactly as `W01-E01-S001` performed for
   its own zero-cost set.
2. Compare the fresh run's output against the cited "38 hits" gosec count and the named
   G704/G304/errorlint/exhaustive/forcetypeassert sites. Record any drift (a site that no longer
   reproduces, a new site, a changed line number) in `deviations.md`, not silently absorbed into this
   plan's own claims.
3. From the fresh gosec run, extract the exact G115 site list (file:line for every G115 hit across
   the audit/database/jobs/mfa/pagination packages the source material names as the affected areas).
   This list does not exist yet — it is a product of this step, not an input to it.
4. Annotate G704 (`kernel/auth/jwks.go:204,210`) with an inline `#nosec G704` (or the pinned gosec
   version's actual supported directive syntax — to be confirmed) justification comment referencing
   SEC-06.
5. Review each G115 site individually. For each: confirm whether a prior validation step in the same
   call path already bounds the value being converted. If yes, annotate with a justification
   referencing that specific validation. If no, add an explicit bounds check immediately before the
   conversion, then re-run gosec to confirm the site no longer flags.
6. Annotate G304 (buildinfo file read) as tool-only, low-risk.
7. Fix errorlint's named site (`kernel/httpx/middleware.go:54`): replace `==` with `errors.Is` against
   `http.ErrAbortHandler`.
8. Annotate exhaustive's 2 named sites (`kernel/workflow/definition.go:313`,
   `kernel/workflow/runtime.go:170`) using the pinned golangci-lint v2.11.4 exhaustive analyzer's
   suppression mechanism (exact directive to be confirmed at implementation time), with a comment at
   each site explaining that the `default:` arm is an intentional fail-closed design, not a gap.
9. Fix forcetypeassert's 2 named sites (`kernel/auth/jwks.go:112`, `kernel/config/bind.go:150`):
   convert each unchecked type assertion to a checked (comma-ok) form with explicit handling of the
   false-ok path (exact handling — error return, log-and-default, panic-with-message — to be
   determined per site from the surrounding function's existing error-handling convention).
10. Annotate `kernel/policy/policy.go:166`'s `nilerr` hit with a comment explaining the fail-closed
    intent (unparseable runtime value evaluates the governing condition to `false`/deny) and its
    relationship to the malformed-policy-error handling already present at line 161. Do not alter the
    underlying condition logic.
11. Enable usestdlibvars via `.golangci.yml`, run it fresh, and fix each site the run enumerates by
    replacing the literal with the equivalent stdlib constant.
12. Update `.golangci.yml` to permanently enable gosec, errorlint, exhaustive, forcetypeassert, and
    usestdlibvars. Confirm `wrapcheck`/`revive` remain absent from the `enable:` list (they were never
    enabled and this story does not add them) and record their rejection rationale in this story's
    documentation per "Documentation requirements" in `story.md`.
13. Re-run the full `golangci-lint run` against the full module tree to confirm zero hits across all
    five newly-enabled analyzers, and separately confirm `wrapcheck`/`revive` are not present in the
    enabled analyzer set.

## Expected package or module changes

`kernel/auth` (G704 annotation, forcetypeassert fix), `kernel/config` (forcetypeassert fix),
`kernel/httpx` (errorlint fix), `kernel/workflow` (exhaustive annotations, 2 files), `kernel/policy`
(nilerr annotation), the audit/database/jobs/mfa/pagination packages (G115 sites — exact packages
within these to be confirmed by the fresh run), whatever package(s) usestdlibvars's fresh run
enumerates, root `.golangci.yml`.

## Expected file changes where determinable

- `.golangci.yml` — enable gosec, errorlint, exhaustive, forcetypeassert, usestdlibvars.
- `kernel/auth/jwks.go:204,210` — G704 `#nosec` annotation (comment only, no logic change).
- `kernel/auth/jwks.go:112` — forcetypeassert fix (checked type assertion).
- `kernel/config/bind.go:150` — forcetypeassert fix (checked type assertion).
- `kernel/httpx/middleware.go:54` — errorlint fix (`errors.Is` in place of `==`).
- `kernel/workflow/definition.go:313` — exhaustive annotation (comment only).
- `kernel/workflow/runtime.go:170` — exhaustive annotation (comment only).
- `kernel/policy/policy.go:166` — nilerr annotation (comment only, no logic change).
- G304 buildinfo-read site — annotation only; exact file to be confirmed at implementation time (the
  source material characterizes it as "buildinfo file read," not a specific path).
- G115 sites across audit/database/jobs/mfa/pagination packages — exact files and line numbers not
  yet known; to be produced by the fresh run (see "Implementation strategy" step 3).
- usestdlibvars sites — exact files and line numbers not yet known; to be produced by the fresh run
  (see "Implementation strategy" step 11).

## Contracts and interfaces

No public interface changes anticipated. The forcetypeassert fixes change an unchecked assertion into
a checked one at each site; if the surrounding function does not already return an error, this may
require widening that function's own signature to return an error on the false-ok path — to be
confirmed per site at implementation time. If a signature change is required at either site, that is
recorded as a plan-vs-actual note in `implementation.md`, not silently treated as "no interface
change."

## Data structures

None anticipated.

## APIs

None affected, unless a forcetypeassert fix's false-ok handling requires a call-site signature change
that propagates to an exported API — to be confirmed at implementation time; not expected based on the
two named sites' described locations (`jwks.go`, `bind.go`, both internal parsing/binding paths).

## Configuration changes

`.golangci.yml`'s `linters.enable` list gains `gosec`, `errorlint`, `exhaustive`, `forcetypeassert`,
`usestdlibvars`. No runtime/application configuration changes.

## Persistence changes

None.

## Migration strategy

Not applicable — no schema or data migration.

## Concurrency implications

None. All fixes and annotations in this story are single-goroutine, non-concurrent code paths per the
named sites' descriptions (JWKS parsing, config binding, panic-recovery middleware, workflow switch
statements, policy evaluation) — to be confirmed at implementation time that none of the as-yet-
unenumerated G115/usestdlibvars sites introduce a concurrency-relevant change (not expected, since
integer-conversion and literal-replacement fixes are not concurrency-shaped by nature).

## Error-handling strategy

The forcetypeassert fixes introduce explicit error handling on the false-ok path at each of the 2
named sites, replacing an implicit panic-on-bad-assertion with an explicit, handled error — the exact
handling shape (return an error, log and use a zero value, etc.) follows the surrounding function's
existing convention, to be determined per site at implementation time. Any G115 site lacking prior
validation gets an explicit bounds check with a defined rejection behavior (error return or fail-closed
default) — exact shape per site, to be determined at implementation time from the surrounding
function's convention.

## Security controls

This story's core content *is* security-control triage: G704's annotation preserves SEC-06's governed
pattern; each G115 site's bounds check (where added) is itself a new security control preventing
integer-overflow-driven misbehavior; the `nilerr` annotation preserves an existing fail-closed
authorization-adjacent control. No new security control class is introduced beyond what the gosec/
exhaustive/nilerr triage itself produces at the per-site level.

## Observability changes

None required. If a G115 bounds-check rejection path is judged to warrant a log line, that is an
implementation-time judgment call per site, not a required scope item (see `story.md` "Observability
considerations").

## Testing strategy

- Fail-first: run gosec, errorlint, exhaustive, forcetypeassert, and usestdlibvars against today's
  `.golangci.yml` state (force-enabled via `--enable=`) — confirms the "before" state (the cited "38
  hits" for gosec, and the named single-site hits for the other four, or whatever the fresh run
  actually finds). Then run again after each fix/annotation, and finally after full `.golangci.yml`
  enablement — confirms the "after" state (all five exit 0).
- For each G115 site that receives an explicit bounds check (rather than an annotation): a targeted
  unit test exercising both the in-bounds (pass) and out-of-bounds (rejected) cases, added at
  whichever existing test file covers that site's function — exact test placement to be determined
  per site at implementation time.
- For the forcetypeassert fixes: a targeted unit test exercising both the successful-assertion and
  failed-assertion (false-ok) paths at each of the 2 named sites.
- For the errorlint fix: confirm the existing `kernel/httpx/middleware.go` test suite (if one exists
  covering the panic-recovery path) still passes; add a targeted test if no existing test exercises
  the `http.ErrAbortHandler` comparison path.
- No new test is required for the annotation-only fixes (G704, G304, exhaustive, nilerr) — these are
  comment-only changes with no behavioral delta to test; the linter's own pass/fail state is the
  verification for these.
- No new integration, concurrency, or race tests are required — this story does not change concurrent
  behavior (see "Concurrency implications").

## Regression strategy

The `golangci-lint run` itself, wired into CI (already the case — this story enables analyzers within
the existing CI-gated `.golangci.yml`, it does not add a new CI step), is the regression guard for the
five newly-enabled analyzers going forward. For sites that receive an explicit bounds check or checked
type assertion, the new unit tests are the regression guard for the underlying logic.

## Compatibility strategy

All fixes in this story are internal-behavior-preserving on the success path (see `story.md`
"Compatibility considerations"). The false-ok/out-of-bounds paths introduced by the forcetypeassert
and G115-bounds-check fixes are new, explicit handling of a condition that previously either panicked
implicitly (forcetypeassert) or was silently unbounded (G115) — this is a compatibility-neutral or
compatibility-improving change (an explicit, handled failure mode replacing an implicit one), not a
breaking change, but each site's exact behavior must be confirmed at implementation time.

## Rollout strategy

Single PR/commit, or a small number of PRs grouped by task (see "Task breakdown" below); no phased
rollout required — this is a CI-config and source-annotation/small-fix change, not a runtime-behavior
change requiring gradual exposure.

## Rollback strategy

Revert the `.golangci.yml` change and the associated fixes/annotations per task if a false positive,
an incorrect annotation, or unexpected CI breakage is found. Each task's changes are independently
revertible since they touch disjoint files (see "Task breakdown").

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-13). Steps 1-3 (fail-first re-run and G115
site enumeration) must occur before steps 4-12 (fixes/annotations/enablement) begin, per the
mandate's fail-first evidence requirement and per the fact that the G115 site list is itself a
required input to that task's per-site review.

## Task breakdown

- **W01-E01-S002-T001** — gosec triage: G704 annotation (step 4).
- **W01-E01-S002-T002** — gosec triage: G115 multi-site review (steps 3, 5).
- **W01-E01-S002-T003** — gosec triage: G304 annotation (step 6).
- **W01-E01-S002-T004** — errorlint fix (step 7).
- **W01-E01-S002-T005** — exhaustive annotations, 2 sites (step 8).
- **W01-E01-S002-T006** — forcetypeassert fixes, 2 sites (step 9).
- **W01-E01-S002-T007** — usestdlibvars fixes (step 11), nilerr annotation (step 10), and the final
  permanent `.golangci.yml` enablement/confirmation run plus wrapcheck/revive-absence check
  (steps 12-13). Depends on T001-T006; consumes T001's fresh-run baseline as its usestdlibvars site
  list.

See "Task grouping rationale" in `tasks/index.md` for why this story departs from the epic-level
suggestion of 5 tasks (G115 split out from the rest of the gosec triage; T007 added as a closure task
so that AC-W01-E01-S002-01/-06/-07 each have an owning task rather than being tracked only at story
level).

## Expected artifacts

Updated `.golangci.yml`; updated `kernel/auth/jwks.go`, `kernel/config/bind.go`,
`kernel/httpx/middleware.go`, `kernel/workflow/definition.go`, `kernel/workflow/runtime.go`,
`kernel/policy/policy.go`; updated G304 site and G115 sites (files TBD); updated usestdlibvars sites
(files TBD); the per-hit triage record.

## Expected evidence

Fail-first/pass-after `golangci-lint run` logs (per analyzer and combined); the full gosec fresh-run
output; the G115 per-site review record; targeted unit-test output for the forcetypeassert fixes and
any G115 bounds checks added.

## Unresolved questions

- The exact G115 site list (files, line numbers, count) — not enumerable until the fresh run in
  "Implementation strategy" step 3 executes.
- The exact usestdlibvars site list (files, line numbers, count) — not enumerable until the fresh run
  in step 11 executes.
- Whether the pinned gosec/golangci-lint v2.11.4 version's supported `#nosec` directive syntax
  requires a specific rule-ID suffix (e.g. `#nosec G704`) or accepts a bare `#nosec` with trailing
  justification text — to be confirmed at implementation time.
- The exact suppression/annotation mechanism the pinned golangci-lint v2.11.4 exhaustive analyzer
  supports (a `//exhaustive:ignore` directive, a `//nolint:exhaustive` directive, or a
  default-case-exempts-from-exhaustiveness configuration option) — to be confirmed at implementation
  time.
- Whether either forcetypeassert site's false-ok handling requires widening the surrounding function's
  signature to return an error (see "Contracts and interfaces" above) — to be determined per site at
  implementation time.
- Whether any G115 site is in a hot/performance-sensitive path such that an added bounds check
  warrants a benchmark rather than only a unit test — not expected, but to be confirmed once the G115
  site list exists.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above that block a
specific task (most importantly, the G115 and usestdlibvars site enumerations) are answered by the
fresh re-run at story start, (b) the fail-first re-run (steps 1-2) confirms or corrects the cited
current-state assessment in `story.md`, and (c) the owner and reviewer are assigned.
