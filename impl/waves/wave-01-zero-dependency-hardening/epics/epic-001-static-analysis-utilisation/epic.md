---
id: W01-E01
type: epic
title: Static-analysis utilisation
status: verification
wave: W01
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-16
source_requirements:
  - FBL-05
  - FBL-07
  - CS-10
  - CS-23
depends_on: []
stories:
  - W01-E01-S001
  - W01-E01-S002
  - W01-E01-S003
decisions: []
risks:
  - RISK-W01-001
---

# W01-E01 — Static-analysis utilisation

## Epic objective

Enable the analyzers the framework's own pinned `golangci-lint` v2.11.4 binary already ships but does
not currently run, triage every hit each newly-enabled analyzer surfaces (fix or annotate with
justification), and close the two remaining supply-chain/hook hygiene gaps (`go mod verify` in CI, a
license-scanning signal, and a pre-push hook fix) that `requirement-inventory.md` groups under the
same FBL-07 finding as the judged-linter triage.

## Problem being solved

`requirement-inventory.md` §B records two review findings targeting this epic:

- **FBL-05** — "Enable zero-cost leak linters (sqlclosecheck etc.)" — disposition `planned`, priority
  P1, target `W01-E01-S001`. MATRIX CS-23 measured the codebase at zero hits for the zero-cost set at
  the time of the matrix pass; the gap is purely that these analyzers are configured off in
  `.golangci.yml`, not that the code has leaks.
- **FBL-07** — "Utilisation closure (gosec triage, go mod verify, license signal, nightly fuzz, hook
  DB-skip)" — disposition `partial` (the nightly CI schedule already exists since PR #24 / session
  delta SD-02; the fuzz-coverage remainder is W07 scope), priority split P1/P2, target
  `W01-E01-S002..S003`.

Both findings describe the same underlying gap: this is a **utilisation** problem, not a **tooling**
problem. `fable5-final-architecture-review-2026-07-11.md`'s reuse-test framing treats "is a capability
the toolchain already ships actually turned on and enforced" as a distinct axis from "does the right
tool exist at all" — MATRIX CS-10 and CS-23 both close against this axis. The pinned golangci-lint
binary ships 25 analyzers total; this epic's three stories account for enabling the ones currently
off and triaging every hit that surfaces, including analyzers whose hits require a human judgment
call (annotate vs. fix vs. reject) rather than a mechanical flip.

## Scope

- Enabling `sqlclosecheck`, `rowserrcheck`, `bodyclose`, `wastedassign`, `makezero`, `musttag`,
  `testifylint` (the zero-cost set) in `.golangci.yml`, plus fixing `noctx`'s 2 named production hits
  and `copyloopvar`'s 1 named production hit (S001).
- Enabling `gosec`, `errorlint`, `exhaustive`, `forcetypeassert`, `usestdlibvars` (the judged set),
  triaging every hit — fix, or annotate with an inline justification comment — per the named triage
  list in S002's scope, and explicitly recording `wrapcheck`/`revive` as a rejected recommendation
  (S002).
- Adding a `go mod verify` step to `ci.yml`, enabling a license-scanning signal, confirming/extending
  the nightly fuzz-schedule wiring within its stated boundary, and fixing the pre-push hook's silent
  DB-test skip (S003).
- New configuration keys `MaxConnLifetime`/`MaxConnIdleTime` on the database connection pool (S001),
  motivated by credential-rotation/load-balancer-rebalance hygiene, not by any linter hit — grouped
  into S001 because it shares CS-10's pgx-pool-configuration context.

## Out of scope

- **FBL-09's G120 unbounded form-parse fix** (`kernel/httpx/csrf.go:118`) — this is FBL-09's own fix,
  targeted at `W01-E03-S001` (`http-hardening` epic). This epic's S002 gosec triage cross-references
  it but does not implement it.
- **The wrapper-type question over `kernel/database/txmanager.go:165,181`'s raw `pgx.Rows`/`pgx.Row`
  returns** — CS-10 records this as a **decided, closed question**: wrapper types are explicitly
  rejected as reinventing `database/sql`'s own caller-owned-close contract for no benefit. This epic's
  S001 enforces the existing idiomatic contract mechanically (via sqlclosecheck/rowserrcheck); it does
  not reopen the wrapper-type debate.
- **REL-04 T8 / PERF-06 T3/T4's real `-fuzz=` coverage-guided generation wiring** — W07 scope (shared
  ownership, assigned to PF-REL per `premier-framework-implementation-plan.md`). S003 confirms the
  nightly *schedule* exists and is correctly wired; it does not add the `-fuzz=` flag itself.
- **`kernel/policy/policy.go:166`'s `nilerr` hit** — recorded as a deliberate fail-closed design,
  personally adjudicated by Fable 5 as not a bug. S002 records this as an explicit non-finding
  (annotate only) so it is not silently dropped from the triage record, but it is not "fixed."
- Any linter not already present in the pinned golangci-lint v2.11.4 binary — this epic is enablement
  of existing capability, not adoption of new tooling.

## Source requirements

FBL-05, FBL-07. Cross-referenced constraint: CS-10 (pgx rows contract — decided, closed), CS-23
(closure-depth matrix triage list snapshot this epic's stories re-confirm rather than blindly trust).

## Architectural context

This epic is about the **utilisation principle** — `fable5-final-architecture-review-2026-07-11.md`'s
second reuse-test axis (the first being "does a reusable primitive already exist that a new feature
should be built on top of"; the second being "is a capability the toolchain already provides actually
turned on"). The pinned `golangci-lint` v2.11.4 binary already ships all 25 analyzers this epic
enables or triages — `sqlclosecheck`, `rowserrcheck`, `bodyclose`, `wastedassign`, `makezero`,
`musttag`, `testifylint`, `noctx`, `copyloopvar`, `gosec`, `errorlint`, `exhaustive`,
`forcetypeassert`, `usestdlibvars`, plus the CI-level `go mod verify` and license-scan capabilities
already present in the Go toolchain and the `security-scan.yml`/`dependency-review` GitHub Actions
already wired into the repository. None of this epic's stories introduce a new binary, a new
dependency, or a new external service — every task is "flip a config flag and fix or annotate what it
finds," which is precisely why this epic is sequenced into W01 (no upstream dependency on AR-01/
AR-02/SEC-01/DATA-09) rather than a later wave.

The affected layers span the database layer (`kernel/database/`), the CLI (`internal/cli/`), the
application-lifecycle layer (`app/maintenance.go`), the auth layer (`kernel/auth/jwks.go`), the config
layer (`kernel/config/bind.go`), the workflow layer (`kernel/workflow/`), CI configuration
(`.github/workflows/ci.yml`, `.github/workflows/security-scan.yml`), and the local git hooks
(`.githooks/`). This breadth is why the epic decomposes into three stories by triage-cost, not by
package — S001's zero-cost set has (per source evidence) no fix burden beyond two named sites; S002's
judged set requires named per-site human review; S003's supply-chain/hook work is CI- and
process-configuration, not code-analyzer-driven.

## Included stories

- **W01-E01-S001 — zero-cost-linters** (FBL-05): enable the zero-cost leak-detection linter set,
  fix noctx's 2 named prod hits and copyloopvar's 1 named prod hit, add
  `MaxConnLifetime`/`MaxConnIdleTime` config keys.
- **W01-E01-S002 — judged-linter-set** (FBL-07, enablement+triage half): enable gosec/errorlint/
  exhaustive/forcetypeassert/usestdlibvars, triage every hit per the named site list, record
  wrapcheck/revive as rejected.
- **W01-E01-S003 — supply-chain-and-hooks** (FBL-07, remainder half): `go mod verify` CI step,
  license-scanning signal, nightly-fuzz-schedule confirmation, pre-push hook DB-silent-skip fix.

## Dependencies

No dependency on any other W01 epic — S001/S002/S003 target disjoint files (database config vs.
CLI/app/auth/config/workflow files vs. CI/hook configuration) and can proceed in any order or in
parallel. This epic depends only on W00's exit gate (baseline lint/coverage state captured, D-01..D-09
ratified) per `wave.md`'s entry criteria — no W01-E01-specific blocking dependency beyond that.
Downstream: `dependencies.md` (wave-level) records that "all later waves' CI runs" depend on this
epic's S001/S002/S003 landing, since every later wave's PR/CI gate runs against the linter/supply-chain
configuration this epic establishes.

## Risks

RISK-W01-001 (judged-linter enablement surfacing more hits at this epic's actual start commit than
MATRIX CS-23's snapshot recorded, since the codebase has moved since the matrix pass) is the epic's
primary risk, inherited from `../../risks.md` (wave-level). See `risks.md` (epic-level) for the
epic-scoped elaboration and any epic-specific risk this creates beyond the wave-level entry.

## Required decisions

None. CS-10's pgx-rows-contract question is already a decided, closed question (wrapper types
rejected) that this epic enforces mechanically, not a decision this epic must make. No new ADR is
required — this epic's stories accordingly carry no `decisions/` directory (see each story's
`story.md` front matter, `decisions: []`).

## Epic acceptance criteria

- **AC-W01-E01-01**: `.golangci.yml` enables `sqlclosecheck`, `rowserrcheck`, `bodyclose`,
  `wastedassign`, `makezero`, `musttag`, `testifylint` and a full-module-tree `golangci-lint run`
  exits 0 against them; `noctx`'s 2 named prod hits and `copyloopvar`'s 1 named prod hit are fixed.
- **AC-W01-E01-02**: `.golangci.yml` enables `gosec`, `errorlint`, `exhaustive`, `forcetypeassert`,
  `usestdlibvars` and every hit surfaced is either fixed or carries an inline justification annotation
  traceable to this epic's triage record; `wrapcheck`/`revive` are recorded as a rejected
  recommendation with rationale, not enabled.
- **AC-W01-E01-03**: `ci.yml` runs `go mod verify`; a license-scanning signal (Trivy license scanner
  or `go-licenses`) is enabled and the choice is documented; the nightly fuzz-schedule wiring is
  confirmed to exist and correctly invoke the seed-corpus replay (with the coverage-guided `-fuzz=`
  gap explicitly recorded as W07 scope, not silently closed or silently duplicated); the pre-push hook
  no longer silently skips DB tests without `WOWAPI_REQUIRE_DB` set.
- **AC-W01-E01-04**: All three stories have passed independent review per mandate §14, with S002
  specifically checked for the completeness of its per-site triage record (no gosec/errorlint/
  exhaustive/forcetypeassert hit silently dropped) and S003 specifically checked for the nightly-fuzz
  scope boundary being honestly stated rather than either silently closed or silently duplicated
  against W07.

## Closure conditions

All three stories reach `accepted` (each satisfying its own `closure.md`); AC-W01-E01-01 through
AC-W01-E01-04 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; no unresolved regression or silently-dropped triage item remains open
against any of the three stories.

## Status update (2026-07-16)

`status: verification` (was `planned`; the parent wave claimed `accepted`) — set by the
2026-07-16 hierarchy reconciliation (**DEV-PROG-006**). All of this epic's stories are
`accepted` with story-level evidence, but this epic's own `closure-report.md` body was never
populated: its acceptance-criteria/story-completion tables still read "not started"/"planned"
while a reviewer-conclusion section appended 2026-07-13 claims acceptance. Until the closure
report is completed honestly against the epic's acceptance criteria, the epic sits in
`verification`; see DEV-PROG-006 in `impl/tracking/programme-deviations.md` for the disposition.
