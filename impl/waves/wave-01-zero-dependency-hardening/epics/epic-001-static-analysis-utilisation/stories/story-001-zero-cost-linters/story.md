---
id: W01-E01-S001
type: story
title: Enable the zero-cost leak-detection linter set
status: accepted
wave: W01
epic: W01-E01
owner: W01Lint
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-05
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W01-E01-S001-01
  - AC-W01-E01-S001-02
  - AC-W01-E01-S001-03
  - AC-W01-E01-S001-04
artifacts:
  - ART-W01-E01-S001-001
  - ART-W01-E01-S001-002
  - ART-W01-E01-S001-003
  - ART-W01-E01-S001-004
  - ART-W01-E01-S001-005
  - ART-W01-E01-S001-006
evidence:
  - EV-W01-E01-S001-001
  - EV-W01-E01-S001-002
  - EV-W01-E01-S001-003
  - EV-W01-E01-S001-004
decisions: []
risks:
  - RISK-W01-E01-002
---

# W01-E01-S001 — Enable the zero-cost leak-detection linter set

## Story ID

W01-E01-S001

## Title

Enable the zero-cost leak-detection linter set

## Objective

Enable `sqlclosecheck`, `rowserrcheck`, `bodyclose`, `wastedassign`, `makezero`, `musttag`, and
`testifylint` in `.golangci.yml`; fix `noctx`'s 2 named production hits and `copyloopvar`'s 1 named
production hit; add `MaxConnLifetime`/`MaxConnIdleTime` as new database connection-pool configuration
keys.

## Value to the framework

The framework's pinned `golangci-lint` v2.11.4 toolchain already ships these seven analyzers; they
detect real resource-leak and correctness classes (unclosed `sql.Rows`, unchecked `rows.Err()`,
unclosed HTTP response bodies, wasted assignments, `slices.Grow`/`append` misuse, struct-tag
mismatches, and testify assertion misuse) that are cheap to prevent mechanically once turned on.
Enabling them converts "the framework happens to be clean today" into "the framework cannot regress
on these specific defect classes without CI catching it" — a durable quality-gate improvement with
(per the cited evidence) zero current-state fix cost for five of the seven analyzers.

## Problem statement

`requirement-inventory.md` row FBL-05 records: "Enable zero-cost leak linters (sqlclosecheck etc.)" —
disposition `planned`, priority P1, target `W01-E01-S001`, with the note "Counts measured at HEAD
(CS-23); noctx 2 prod fixes." The gap is that `.golangci.yml` does not currently enable these
analyzers, even though the codebase is already clean against them (per the cited evidence below) —
this is pure enablement-plus-two-small-fixes work, not a defect-remediation project.

## Source requirements

FBL-05. Cross-referenced constraint: CS-10 (pgx rows contract — the decided, closed question that
`kernel/database/txmanager.go:165,181`'s raw `pgx.Rows`/`pgx.Row` returns stay as-is; this story
enforces that contract mechanically via sqlclosecheck/rowserrcheck rather than reopening it).

## Current-state assessment

Per the source evidence cited for this story (to be re-confirmed at this story's own execution commit
— see "Assumptions" and `plan.md`'s fail-first verification note):

- All 26 production `.Query(` call sites across 15 files close their `rows` and check `rows.Err()` —
  zero violations reported when sqlclosecheck and rowserrcheck were actually run against the codebase.
- `kernel/database/txmanager.go:165,181` return raw `pgx.Rows`/`pgx.Row` to the caller (caller-owned
  close) — this is the idiomatic `database/sql`-shaped contract and is a **decided, closed** design
  choice (CS-10): wrapper types were explicitly considered and rejected as reinventing
  `database/sql`'s own contract for no benefit. This story does not reopen that question.
- The database connection pool currently exposes only `MaxConns` as a configuration key (default 16,
  range 2-200, `kernel/config/config.go:99,211-212`). `MaxConnLifetime` and `MaxConnIdleTime` are left
  at pgx's internal defaults — no configuration keys exist for them yet.
- `noctx` has 2 named production hits: `internal/cli/config_delegate.go:34` and
  `internal/cli/lint_cmd.go:129` — both call `exec.Command` without a context, where
  `exec.CommandContext` should be used instead.
- `copyloopvar` has 1 named production hit: `app/maintenance.go:148` — a dead pre-1.22 loop-variable-
  capture idiom that is now unnecessary given the module's Go version.

**This assessment reflects the state cited in `requirement-inventory.md`/MATRIX CS-23 at the time
those documents were written.** Per this story's own plan (`plan.md`), the first implementation step
re-runs sqlclosecheck/rowserrcheck/bodyclose/wastedassign/makezero/musttag/testifylint/noctx/
copyloopvar fresh at this story's actual start commit — it does not simply trust this cited snapshot.
W00-E01/E02's baseline work is the formal re-verification mechanism for the programme as a whole; this
story's own re-run is the story-local fail-first check that must pass before the story can claim its
acceptance criteria are met.

## Desired state

`.golangci.yml` enables all seven zero-cost analyzers; a full-module-tree `golangci-lint run` against
them exits 0. `noctx`'s 2 named sites use `exec.CommandContext`. `app/maintenance.go:148`'s loop no
longer uses the pre-1.22 capture idiom. `kernel/config/config.go` exposes `MaxConnLifetime` and
`MaxConnIdleTime` as configuration keys with pgx-default defaults, documented for credential-rotation
and load-balancer-rebalance hygiene.

## Scope

- Enabling `sqlclosecheck`, `rowserrcheck`, `bodyclose`, `wastedassign`, `makezero`, `musttag`,
  `testifylint` in `.golangci.yml`.
- Fixing the 2 named `noctx` production hits.
- Fixing the 1 named `copyloopvar` production hit.
- Adding `MaxConnLifetime` and `MaxConnIdleTime` as new pool-configuration keys, defaulted to pgx's
  own internal defaults, with corresponding config validation/documentation.

## Out of scope

- Any wrapper type around `pgx.Rows`/`pgx.Row` — CS-10 is a decided, closed question (see "Current-
  state assessment"); this story does not revisit it.
- The judged linter set (gosec, errorlint, exhaustive, forcetypeassert, usestdlibvars) — that is
  W01-E01-S002's scope.
- `go mod verify`, license-scanning, nightly-fuzz confirmation, and the pre-push hook fix — that is
  W01-E01-S003's scope.
- Any non-`noctx`/non-`copyloopvar` fix outside the seven named zero-cost analyzers, unless this
  story's fresh re-run (see "Current-state assessment") surfaces a new hit not covered by the cited
  snapshot — in which case it is handled per RISK-W01-E01-002's contingency (a 5th task, not a
  silent scope absorption).

## Assumptions

- The "26 production `.Query(` sites, zero violations" and "2 noctx / 1 copyloopvar" counts are
  assumed to still hold at this story's actual execution commit, subject to the fresh re-run required
  by `plan.md`'s fail-first verification step. If drift is found, it is recorded, not silently
  reconciled by editing this story's own current-state claims after the fact.
- pgx's internal defaults for `MaxConnLifetime`/`MaxConnIdleTime` are assumed appropriate as the new
  keys' default values (i.e., adding the keys should not change pool behavior for any deployment that
  does not explicitly set them) — to be confirmed against the actual pgxpool version pinned in
  `go.mod` at implementation time.

## Dependencies

None within W01-E01 (S001/S002/S003 target disjoint files — see epic-level `dependencies.md`). Depends
on W00's exit gate at wave scope (baseline lint-hit-count capture).

## Affected packages or components

`.golangci.yml`; `internal/cli/config_delegate.go`; `internal/cli/lint_cmd.go`; `app/maintenance.go`;
`kernel/config/config.go`; `kernel/database/` (pool-configuration wiring consuming the new keys — the
exact construction site, e.g. `pgxpool.Config` population, to be identified at implementation time).

## Compatibility considerations

Adding `MaxConnLifetime`/`MaxConnIdleTime` as new, optional configuration keys with pgx-default
defaults is additive and backward-compatible: any deployment (including wowsociety) that does not set
them observes unchanged pool behavior. The `noctx`/`copyloopvar` fixes are internal-behavior-preserving
(passing a context to `exec.Command`/removing a now-unnecessary loop-variable capture do not change
observable CLI or maintenance-job behavior).

## Security considerations

`noctx`'s fix (passing a context to `exec.CommandContext`) allows the invoking code path to cancel or
time-bound the subprocess it launches — a minor hardening side effect, not the primary motivation
(the primary motivation is linter-clean status), but worth recording since an uncancellable subprocess
is a minor resource-exhaustion surface.

## Performance considerations

`MaxConnLifetime`/`MaxConnIdleTime` directly affect connection-pool churn behavior once a deployment
opts into non-default values — documented as motivated by "credential-rotation/LB-rebalance hygiene,"
i.e. operational rather than throughput tuning. No performance regression is expected from adding the
keys with defaults that reproduce current (pgx-default) behavior.

## Observability considerations

None beyond what already exists — this story does not add new metrics or logs. If the new pool-config
keys are considered operationally significant enough to warrant a startup log line confirming their
effective values, that is a judgment call for implementation time, not a required scope item here.

## Migration considerations

None. No schema or data migration; the new config keys are additive with safe defaults.

## Documentation requirements

- Document the new `MaxConnLifetime`/`MaxConnIdleTime` config keys (purpose, defaults, valid ranges if
  any) alongside the existing `MaxConns` documentation in whatever doc currently covers
  `kernel/config/config.go`'s database pool settings.
- Record the `.golangci.yml` change (which analyzers were enabled, and why they were previously off —
  simply unconfigured, not intentionally disabled) in this story's `implementation.md` once executed.

## Acceptance criteria

- **AC-W01-E01-S001-01**: `golangci-lint run` with `sqlclosecheck`, `rowserrcheck`, `bodyclose`,
  `wastedassign`, `makezero`, `musttag`, `testifylint` enabled exits 0 across the full module tree.
- **AC-W01-E01-S001-02**: `noctx` exits 0 against `internal/cli/config_delegate.go` and
  `internal/cli/lint_cmd.go`, evidenced by a fail-before/pass-after run (the sites use
  `exec.CommandContext` after the fix, `exec.Command` before).
- **AC-W01-E01-S001-03**: `copyloopvar` exits 0 against `app/maintenance.go`, evidenced by a
  fail-before/pass-after run.
- **AC-W01-E01-S001-04**: `kernel/config/config.go` exposes `MaxConnLifetime` and `MaxConnIdleTime`
  config keys; a config-validation unit test confirms both accept pgx-default values and that omitting
  them preserves current (pre-story) pool behavior.

## Required artifacts

- Updated `.golangci.yml`.
- Updated `internal/cli/config_delegate.go`, `internal/cli/lint_cmd.go`, `app/maintenance.go`.
- Updated `kernel/config/config.go` (new keys) and its documentation.
See `artifacts/index.md`.

## Required evidence

- Per-analyzer enablement run logs (zero-hit state) for the seven zero-cost analyzers.
- Fail-before/pass-after run logs for `noctx` (2 sites) and `copyloopvar` (1 site).
- Config-validation unit test output for the new pool-lifetime keys.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none) recorded,
owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14.

## Risks

RISK-W01-E01-002 (a fresh re-run surfaces a hit the cited snapshot did not record) — see epic-level
`risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once the fresh re-run confirms the cited zero-hit/2-hit/1-hit state (or the story absorbs whatever
drift is found, per RISK-W01-E01-002's contingency), no residual risk is expected to remain open at
acceptance — this is a mechanical enablement story with a small, bounded, and already largely-verified
fix surface.

## Plan

See `plan.md`.
