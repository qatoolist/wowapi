---
id: W01-E01-S002
type: story
title: Enable and triage the judged linter set
status: accepted
wave: W01
epic: W01-E01
owner: W01Lint
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-07
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W01-E01-S002-01
  - AC-W01-E01-S002-02
  - AC-W01-E01-S002-03
  - AC-W01-E01-S002-04
  - AC-W01-E01-S002-05
  - AC-W01-E01-S002-06
  - AC-W01-E01-S002-07
artifacts:
  - ART-W01-E01-S002-001
  - ART-W01-E01-S002-002
  - ART-W01-E01-S002-003
  - ART-W01-E01-S002-004
  - ART-W01-E01-S002-005
  - ART-W01-E01-S002-006
  - ART-W01-E01-S002-007
  - ART-W01-E01-S002-008
  - ART-W01-E01-S002-009
  - ART-W01-E01-S002-010
  - ART-W01-E01-S002-011
  - ART-W01-E01-S002-012
evidence:
  - EV-W01-E01-S002-001
  - EV-W01-E01-S002-002
  - EV-W01-E01-S002-003
  - EV-W01-E01-S002-004
  - EV-W01-E01-S002-005
  - EV-W01-E01-S002-006
  - EV-W01-E01-S002-007
  - EV-W01-E01-S002-008
  - EV-W01-E01-S002-009
  - EV-W01-E01-S002-010
decisions: []
risks:
  - RISK-W01-E01-002
---

# W01-E01-S002 — Enable and triage the judged linter set

## Story ID

W01-E01-S002

## Title

Enable and triage the judged linter set

## Objective

Enable `gosec`, `errorlint`, `exhaustive`, `forcetypeassert`, and `usestdlibvars` in `.golangci.yml`;
triage every hit each surfaces — fix it, or annotate it with an inline justification comment — per the
named triage list in "Source content" below; explicitly record `kernel/policy/policy.go:166`'s
`nilerr` hit as a deliberate non-finding; and explicitly record `wrapcheck`/`revive` as a rejected
recommendation, not enabled.

## Value to the framework

Unlike the zero-cost set (`W01-E01-S001`), these five analyzers surface real judgment calls: gosec
flags security-shaped patterns that may be governed/intentional rather than bugs, errorlint and
forcetypeassert flag correctness idioms with a spectrum of actual risk, and exhaustive flags missing
switch arms that may already be covered by a fail-closed default. Turning them on and recording an
explicit, reviewable disposition for every hit — rather than leaving them off because "they'd need
triage" — converts implicit tribal knowledge (this gosec hit is fine because SEC-06 governs it; this
exhaustive hit is fine because the default arm fails closed) into an auditable, CI-enforced record.
This is the same utilisation principle as S001 (turn on what the toolchain already ships), applied to
the harder, judgment-bearing half of the epic's linter surface.

## Problem statement

`requirement-inventory.md` row FBL-07 records: "Utilisation closure (gosec triage, go mod verify,
license signal, nightly fuzz, hook DB-skip)" — disposition `partial`, priority split P1/P2, target
`W01-E01-S002..S003`. This single source finding spans two disjoint concerns: (1) enabling and
triaging the judgment-bearing linter set (gosec/errorlint/exhaustive/forcetypeassert/usestdlibvars),
and (2) supply-chain and git-hook hygiene (`go mod verify`, license scanning, nightly fuzz-schedule
confirmation, pre-push hook fix). This story, `W01-E01-S002`, owns only the first half — the
linter-enablement-and-triage half. The second half is `W01-E01-S003`'s scope; see "Out of scope"
below. The epic-level `epic.md` records this same split under "Included stories."

The historical evidence citation for this finding uses the path convention
`evidence/premier/FBL-07/` — this is the source review material's own evidence path from the earlier
architecture-review pass, not a path inside this programme's `impl/` tree. This story's own evidence
lives at `evidence/index.md` (this directory), per this programme's structure (see "Required
evidence" below). The two are distinct: the historical citation path is referenced here for
provenance; it is not where this story's own verification evidence will be produced or stored.

## Source requirements

FBL-07 (enablement-and-triage half only). Cross-referenced constraint: CS-24 (SSRF dial-time guard —
already verified strength; the gosec G704 annotation task inside this story's scope references CS-24's
verified design, it does not reopen it).

## Current-state assessment

Per the source evidence cited for this finding — **to be re-confirmed fresh at this story's own
execution commit**, following the same fail-first re-run discipline `W01-E01-S001` applied to its own
"26 sites clean / 2 noctx / 1 copyloopvar" citation (do not trust the cited counts below without a
fresh run first):

- **gosec** was reported at **38 hits** at MATRIX-pass time. A named subset of those hits has an
  established disposition (below); the remainder — principally the G115 int-overflow-conversion class
  — has not been reduced to an exact file:line list in the source material and must be enumerated by
  this story's own fresh run, not invented here.
  - **G704** (taint via a JWKS fetch) — 2 named sites, `kernel/auth/jwks.go:204,210`. This is a
    deliberate, governed pattern: SEC-06 (outbound-security escape-hatch governance, D-07 ratified)
    governs exactly this trusted-issuer JWKS fetch. It is not a bug.
  - **G115** (potentially unsafe int-to-int conversions that may overflow) — a multi-site class
    spanning the audit, database, jobs, mfa, and pagination packages, per the source material's
    characterization. Most conversions are believed bounded by validation performed earlier in the
    same call path, but no exact file:line enumeration exists yet in the source material — this
    story's own fresh run must produce that list before triage can proceed per-site.
  - **G304** (file read via variable path) — 1 named site, the buildinfo file read. Tool-only
    (build-time diagnostic tooling), low production-risk.
- **errorlint** — 1 named site, `kernel/httpx/middleware.go:54`, compares a recovered panic value to
  `http.ErrAbortHandler` using `==`. `net/http` documents `ErrAbortHandler` as a panicked sentinel
  value, so `==` comparison against it is technically defensible as written; `errors.Is` is a harmless
  mechanical improvement, not a defect fix.
- **exhaustive** — 2 named sites, `kernel/workflow/definition.go:313` and
  `kernel/workflow/runtime.go:170`. Both switch statements are covered by a fail-closed `default:`
  arm. These were reviewed and rejected as bugs (personally verified by Fable 5 during the
  architecture-review pass that produced the source material).
- **forcetypeassert** — 2 named sites, `kernel/auth/jwks.go:112` and `kernel/config/bind.go:150`, both
  currently use unchecked (non-comma-ok) type assertions.
- **usestdlibvars** — no specific sites are named in the source material; this analyzer is currently
  off and must be enumerated by a fresh run at implementation time.
- **nilerr** (already-enabled analyzer, not part of this story's enablement scope, but its one
  reviewed hit belongs in this story's triage record per the epic's closure conditions) —
  `kernel/policy/policy.go:166` has a `nilerr` hit that is a deliberate fail-closed design: an
  unparseable runtime value makes the governing condition evaluate `false` (deny), and malformed
  policy errors are already handled separately at line 161. Personally adjudicated by Fable 5 as not
  a bug.
- **wrapcheck** and **revive** — approximately 50 hits each at MATRIX-pass time, per the source
  material, judged noise-dominant without a heavy per-project tuning investment. Not currently
  enabled.

**This assessment reflects the state cited in the source review material at the time it was written.**
Per this story's own plan (`plan.md`), the first implementation step re-runs gosec, errorlint,
exhaustive, forcetypeassert, and usestdlibvars fresh, at this story's actual start commit — it does
not simply trust the cited "38 hits" or the named-site lists above. Any drift (a site count, a
file:line, a hit that no longer reproduces, or a new hit not covered by the lists above) is recorded,
not silently reconciled into this document after the fact.

## Desired state

`.golangci.yml` enables `gosec`, `errorlint`, `exhaustive`, `forcetypeassert`, and `usestdlibvars`.
Every hit these five analyzers surface at this story's fresh-run commit has one of two dispositions,
both traceable in this story's triage record: fixed (code changed to eliminate the hit), or annotated
(an inline justification comment, referencing the governing decision or design rationale, that
satisfies the linter's suppression mechanism where one exists, or is otherwise recorded in this
story's evidence if the analyzer has no native suppression syntax). `kernel/policy/policy.go:166`'s
`nilerr` hit is explicitly annotated and recorded, not silently left uncommented. `wrapcheck` and
`revive` remain disabled, with their rejection recorded as a permanent decision record in this story's
scope section (see "Out of scope"), not merely as an absence.

## Scope

- Enabling `gosec`, `errorlint`, `exhaustive`, `forcetypeassert`, `usestdlibvars` in `.golangci.yml`.
- Triaging every gosec hit at the fresh-run commit:
  - G704 (`kernel/auth/jwks.go:204,210`) — annotate with an inline `#nosec` justification comment
    referencing SEC-06; do not "fix" (change) this governed pattern.
  - G115 (int-overflow conversions across audit/database/jobs/mfa/pagination) — enumerate the exact
    site list at implementation time via the fresh run; review each site individually; either
    annotate (bounded-by-prior-validation justification, referencing the specific validation that
    bounds it) or add an explicit bounds check where no such prior validation exists.
  - G304 (buildinfo file read) — annotate as tool-only, low-risk.
  - Any additional gosec hit the fresh run surfaces beyond the named list above — triaged the same
    way (fix or annotate), not silently dropped.
- Fixing errorlint's 1 named site (`kernel/httpx/middleware.go:54`): mechanically adopt `errors.Is`
  in place of `==` against `http.ErrAbortHandler`. This is a low-risk mechanical fix, not a defect
  remediation — the `==` comparison was already technically defensible.
- Annotating exhaustive's 2 named sites (`kernel/workflow/definition.go:313`,
  `kernel/workflow/runtime.go:170`) to satisfy the linter while preserving and explicitly documenting
  the fail-closed `default:` arm's intentional design — not converting either switch to an exhaustive
  enumeration of cases.
- Fixing forcetypeassert's 2 named sites (`kernel/auth/jwks.go:112`, `kernel/config/bind.go:150`) by
  converting each to a checked (comma-ok) type assertion with explicit error handling on the
  false-ok path.
- Enabling usestdlibvars and mechanically fixing whatever sites the fresh run enumerates (none are
  named in the source material; this story does not invent site names in advance of that run).
- Explicitly recording `kernel/policy/policy.go:166`'s `nilerr` hit as a deliberate fail-closed
  non-finding: annotated (inline comment explaining the fail-closed intent and the line-161 malformed-
  error handling it complements), not "fixed" — the underlying logic is correct as written and must
  not be altered.
- Explicitly recording `wrapcheck` and `revive` as a **rejected** recommendation: classification `REJ`
  per mandate §1.3, disposition `rejected` per mandate §1.4, rationale noise-dominant (~50 hits each
  at MATRIX-pass time) without a heavy per-project tuning investment disproportionate to the benefit.
  Neither analyzer is enabled by this story.

## Out of scope

- **`W01-E01-S003`'s remainder of FBL-07** — `go mod verify` in CI, the license-scanning signal,
  nightly-fuzz-schedule confirmation, and the pre-push hook's DB-silent-skip fix. These are a
  disjoint concern (CI/supply-chain/hook configuration, not linter-hit triage) tracked entirely under
  `W01-E01-S003`. This story does not duplicate or silently assume any of that scope handled.
- **FBL-09's G120 unbounded form-parse fix** (`kernel/httpx/csrf.go:118`). This gosec hit belongs to
  a different epic entirely — `W01-E03-S001` (http-hardening). This story's gosec triage task
  reviews and records this hit's existence as a cross-reference only; it does not fix it, and it must
  not be silently assumed already handled by this story's own gosec-triage work.
- **`W01-E01-S001`'s zero-cost linter set** (sqlclosecheck, rowserrcheck, bodyclose, wastedassign,
  makezero, musttag, testifylint, plus the noctx/copyloopvar fixes and the pool-lifetime config
  keys) — that is `W01-E01-S001`'s scope, already a sibling story.
- **`wrapcheck`/`revive` enablement** — explicitly rejected (see "Scope" above), not deferred for a
  future story. If either analyzer is reconsidered later, that requires a new decision record, not a
  reopening of this story.
- **Re-litigating exhaustive's 2 named sites or the `nilerr` hit as bugs** — both were personally
  adjudicated by Fable 5 during the architecture-review pass as intentional, correct, fail-closed
  design. This story's task for each is annotation only; changing the underlying switch/condition
  logic is out of scope and would itself be a deviation requiring justification.
- **Any gosec hit the fresh run does not reproduce** — if a cited hit no longer exists at this
  story's execution commit (e.g. already fixed by unrelated work), that is recorded as a finding in
  `deviations.md`, not silently treated as this story's own completed work.

## Assumptions

- The "38 hits" gosec count, and the named G704/G304/errorlint/exhaustive/forcetypeassert sites, are
  assumed to still reproduce at this story's actual execution commit, subject to the fresh re-run
  required by `plan.md`'s fail-first verification step. If drift is found, it is recorded, not
  silently reconciled by editing this story's own current-state claims after the fact.
- G115's exact site list is assumed to be enumerable by a single fresh `gosec`/`golangci-lint
  --enable=gosec` run at implementation time; this story does not assume any specific count or
  file:line in advance of that run.
- usestdlibvars's site list is assumed to be small and mechanical (the analyzer flags direct use of
  values like HTTP status-code integers or `"GET"` string literals where an `net/http`/`http.Method*`
  stdlib constant exists) based on the analyzer's documented behavior; this is not confirmed against
  this repository's actual code until the fresh run occurs.
- gosec's native `#nosec` suppression-comment mechanism (with a following justification and, where
  supported, a specific rule-ID reference such as `#nosec G704`) is assumed to be the correct
  annotation mechanism for the G704/G304/G115-bounded-by-validation cases; to be confirmed against
  the pinned gosec/golangci-lint version's actual supported comment syntax at implementation time.
- exhaustive's own annotation mechanism (e.g. an `//exhaustive:ignore` or `//nolint:exhaustive`
  directive, or an explicit `default:` case recognized by the analyzer's own default-case handling
  configuration) is assumed to exist and be sufficient to satisfy the linter while preserving the
  fail-closed default arm; the exact mechanism is to be confirmed at implementation time against the
  pinned golangci-lint v2.11.4 exhaustive analyzer's actual configuration options.

## Dependencies

None within W01-E01 (S001/S002/S003 target disjoint files — see epic-level `dependencies.md`).
Depends on W00's exit gate at wave scope (baseline lint-hit-count capture).

## Affected packages or components

`.golangci.yml`; `kernel/auth/jwks.go`; `kernel/config/bind.go`; `kernel/httpx/middleware.go`;
`kernel/workflow/definition.go`; `kernel/workflow/runtime.go`; `kernel/policy/policy.go`; the
audit/database/jobs/mfa/pagination packages touched by the G115 site enumeration (exact files to be
identified at implementation time via the fresh run); the buildinfo-reading tool/site for G304 (exact
location to be confirmed at implementation time); whatever sites the fresh usestdlibvars run
enumerates.

## Compatibility considerations

The forcetypeassert fixes (comma-ok assertions with explicit false-ok handling) and the errorlint fix
(`errors.Is` in place of `==`) are internal-behavior-preserving for the success path; the false-ok /
non-match paths must be reviewed at implementation time to confirm they do not change observable
behavior beyond making an existing implicit panic-on-bad-assertion path into an explicit, handled
error path (which is a compatibility-neutral or compatibility-improving change, not a regression, but
must be confirmed per site). The gosec annotations and exhaustive annotations are comment-only changes
with no behavioral effect. usestdlibvars fixes (replacing a literal with an equivalent stdlib
constant) are behavior-preserving by construction, since the stdlib constant's value equals the
literal it replaces.

## Security considerations

This is a security-linter-triage story by nature. The G704 annotation preserves SEC-06's governed
outbound-fetch pattern without weakening it. The G115 per-site review is itself a security review —
each site's disposition (annotate as bounded, or add an explicit bounds check) must be individually
justified, not blanket-annotated, since a G115 hit that is not actually bounded by prior validation
is a real integer-overflow risk. The G304 annotation is justified as tool-only/low-risk but the
justification itself is a security judgment that must be recorded, not assumed. The
`kernel/policy/policy.go:166` `nilerr` non-finding preserves a fail-closed authorization-adjacent
design; the annotation must make this fail-closed intent explicit so a future reader does not mistake
the silenced linter warning for an oversight.

## Performance considerations

None expected. All of this story's changes are either comment-only annotations, single-site mechanical
fixes (errorlint, forcetypeassert), or bounds checks added at G115 sites lacking prior validation
(a bounds check is O(1) and not expected to be performance-material at any of the affected sites, to
be confirmed at implementation time if any site is in a hot path).

## Observability considerations

None required by this story's acceptance criteria. If a G115 site's added bounds check is judged to
warrant a log line or metric on the rejection path (value out of bounds), that is a judgment call for
implementation time at that specific site, not a required scope item here.

## Migration considerations

None. No schema or data migration; this story is linter configuration plus source-level annotations
and small, behavior-preserving code fixes.

## Documentation requirements

- Record the `.golangci.yml` change (which analyzers were enabled) in this story's `implementation.md`
  once executed.
- Record the full per-hit triage disposition (fixed or annotated, with rationale) for every gosec,
  errorlint, exhaustive, and forcetypeassert hit surfaced by the fresh run, including any hit beyond
  the named sites in "Current-state assessment," in `implementation.md` and cross-referenced from the
  relevant task file(s) under `tasks/`.
- Record the `wrapcheck`/`revive` rejection as a permanent decision record (rationale: noise-dominant,
  disproportionate tuning cost) — this is already stated in this story's "Scope" section as the
  authoritative record; no separate ADR is required since this is a documented rejection, not an
  architectural decision requiring the `decisions/` structure (this story's front matter carries
  `decisions: []`, matching the epic-level statement that no new ADR is required for this epic).

## Acceptance criteria

- **AC-W01-E01-S002-01**: `.golangci.yml` enables `gosec`, `errorlint`, `exhaustive`,
  `forcetypeassert`, `usestdlibvars`, and a full-module-tree `golangci-lint run` exits 0 against all
  five.
- **AC-W01-E01-S002-02**: Every gosec hit surfaced by the fresh run at this story's execution commit
  has a recorded disposition (fixed or annotated) in this story's triage record; G704
  (`kernel/auth/jwks.go:204,210`) is specifically annotated with an inline justification referencing
  SEC-06, not fixed; G115 sites are specifically enumerated, individually reviewed, and each
  disposed as annotated (bounded-by-prior-validation) or fixed (explicit bounds check added); G304
  is specifically annotated as tool-only/low-risk.
- **AC-W01-E01-S002-03**: `errorlint`'s named site (`kernel/httpx/middleware.go:54`) uses `errors.Is`
  in place of `==` against `http.ErrAbortHandler`, evidenced by a fail-before/pass-after run.
- **AC-W01-E01-S002-04**: `exhaustive`'s 2 named sites (`kernel/workflow/definition.go:313`,
  `kernel/workflow/runtime.go:170`) are annotated to satisfy the linter, with the fail-closed
  `default:` arm's intentional design explicitly preserved and documented in the annotation comment
  at each site — not converted to an exhaustive case enumeration.
- **AC-W01-E01-S002-05**: `forcetypeassert`'s 2 named sites (`kernel/auth/jwks.go:112`,
  `kernel/config/bind.go:150`) use checked (comma-ok) type assertions with explicit handling of the
  false-ok path, evidenced by a fail-before/pass-after run.
- **AC-W01-E01-S002-06**: `usestdlibvars` is enabled and a full-module-tree `golangci-lint run` exits
  0 against it, with every site the fresh run enumerates recorded in the triage record (fixed
  mechanically).
- **AC-W01-E01-S002-07**: `kernel/policy/policy.go:166`'s `nilerr` hit is explicitly recorded in this
  story's triage record as a deliberate fail-closed non-finding (annotated, not fixed), and
  `wrapcheck`/`revive` are explicitly recorded as a rejected recommendation with rationale in this
  story's scope record — neither is enabled in `.golangci.yml`.

## Required artifacts

- Updated `.golangci.yml`.
- Updated `kernel/auth/jwks.go` (G704 annotation, forcetypeassert fix), `kernel/config/bind.go`
  (forcetypeassert fix), `kernel/httpx/middleware.go` (errorlint fix), `kernel/workflow/definition.go`
  and `kernel/workflow/runtime.go` (exhaustive annotations), `kernel/policy/policy.go` (nilerr
  annotation), plus whatever G115 sites and usestdlibvars sites the fresh run enumerates.
- A per-hit triage record (gosec/errorlint/exhaustive/forcetypeassert/usestdlibvars) covering every
  hit surfaced by the fresh run, with disposition and rationale.
See `artifacts/index.md`.

## Required evidence

- Per-analyzer enablement run logs (fail-before/pass-after) for the five judged analyzers.
- The full gosec fresh-run output (the actual enumerated hit list at this story's execution commit),
  retained as evidence regardless of whether it matches the cited "38 hits."
- Fail-before/pass-after run logs for errorlint (1 site) and forcetypeassert (2 sites).
- The G115 per-site review record (site list, disposition, rationale per site).
- Confirmation that `wrapcheck`/`revive` remain disabled in the final `.golangci.yml`.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none) recorded,
owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all seven acceptance
criteria verified with evidence in `evidence/index.md`; every gosec/errorlint/exhaustive/
forcetypeassert/usestdlibvars hit surfaced at the execution commit has a recorded disposition (none
silently dropped); the `nilerr` non-finding and the `wrapcheck`/`revive` rejection are both explicitly
recorded; `closure.md` completed; independent review passed per mandate §14.

## Risks

RISK-W01-E01-002 (a fresh re-run surfaces more hits at this story's actual execution commit than the
cited "38 hits" MATRIX snapshot recorded, since the codebase has moved since the matrix pass) — see
epic-level `risks.md` for full detail and mitigation/contingency. This risk is materially more likely
to manifest for this story than for `W01-E01-S001`, since the judged set's G115 site list is not yet
enumerated at all in the source material (unlike S001's fully-named zero-cost-analyzer sites).

## Residual-risk expectations

Once the fresh re-run enumerates and this story's triage disposes of every gosec/errorlint/exhaustive/
forcetypeassert/usestdlibvars hit, no residual *undocumented* risk is expected to remain open at
acceptance. Residual risk that is expected to remain, by design, is the accepted risk inherent in the
G704/G304/G115-bounded/`nilerr` annotations themselves — these are governed, reviewed, deliberate
acceptances of a linter-flagged pattern, not eliminated risk, and closure must record them as accepted
risk rather than imply they have been "fixed away."

## Plan

See `plan.md`.
