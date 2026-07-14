---
id: PLAN-W01-E01-S001
type: plan
parent_story: W01-E01-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W01-E01-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information.

## Proposed architecture

No architectural change. This story is pure configuration (linter enablement, config-key addition)
plus two small, behavior-preserving code fixes. No new package, interface, or contract is introduced.

## Implementation strategy

1. Re-run the seven zero-cost analyzers (sqlclosecheck, rowserrcheck, bodyclose, wastedassign,
   makezero, musttag, testifylint) plus noctx and copyloopvar fresh, at this story's actual start
   commit, with the analyzers explicitly force-enabled via a one-off `golangci-lint run
   --enable=...` invocation (not yet via `.golangci.yml`) — this is the fail-first evidence step: the
   "off" config state today vs. the analyzers actually running.
2. Confirm the re-run reproduces the cited "26 sites clean / 2 noctx / 1 copyloopvar" state. Record
   any drift found (this is planned, not assumed).
3. Fix the 2 named `noctx` sites (`exec.Command` → `exec.CommandContext`, threading an existing
   context from the calling function where available, or `context.Background()`/`context.TODO()`
   with a comment if no context is currently threaded — to be determined at implementation time from
   the actual call site).
4. Fix the 1 named `copyloopvar` site (`app/maintenance.go:148`) by removing the pre-1.22
   loop-variable-capture idiom.
5. Add `MaxConnLifetime`/`MaxConnIdleTime` to `kernel/config/config.go`'s config struct, validation,
   and defaults (pgx-default values), and wire them into the pool-construction call site.
6. Update `.golangci.yml` to permanently enable all seven zero-cost analyzers plus confirm noctx/
   copyloopvar (already-enabled analyzers per the source evidence, or newly enabled here if not — to
   be confirmed at implementation time which of noctx/copyloopvar are already configured on vs. off).
7. Re-run the full `golangci-lint run` against the full module tree to confirm zero hits across all
   nine analyzers post-fix.

## Expected package or module changes

`kernel/config` (new config keys), `internal/cli` (2 files, noctx fix), `app` (1 file, copyloopvar
fix), root `.golangci.yml`.

## Expected file changes where determinable

- `.golangci.yml` — enable sqlclosecheck, rowserrcheck, bodyclose, wastedassign, makezero, musttag,
  testifylint (and confirm noctx/copyloopvar state).
- `internal/cli/config_delegate.go:34` — `exec.Command` → `exec.CommandContext`.
- `internal/cli/lint_cmd.go:129` — `exec.Command` → `exec.CommandContext`.
- `app/maintenance.go:148` — remove pre-1.22 loop-variable-capture idiom.
- `kernel/config/config.go` — add `MaxConnLifetime`, `MaxConnIdleTime` fields, validation, defaults
  (exact struct/field names to be determined at implementation time, following the existing
  `MaxConns` field's naming convention at line 99).
- The pgx pool-construction call site consuming `kernel/config/config.go`'s values (exact file to be
  identified at implementation time — expected to be within `kernel/database/`, not yet confirmed).

## Contracts and interfaces

No public interface changes. The config struct gains two new optional fields; this is additive.

## Data structures

`kernel/config` config struct: two new fields (`MaxConnLifetime time.Duration`,
`MaxConnIdleTime time.Duration`, or equivalent — exact typing to match pgx's own
`pgxpool.Config` field types).

## APIs

None affected.

## Configuration changes

New keys `MaxConnLifetime`/`MaxConnIdleTime` (naming to match existing `MaxConns` convention, e.g.
`WOWAPI_DB_MAX_CONN_LIFETIME`/`WOWAPI_DB_MAX_CONN_IDLE_TIME` env vars if that is the existing
convention — to be confirmed against `kernel/config/config.go`'s existing `MaxConns` binding at
implementation time), defaulted to pgx's internal defaults so omitting them is behavior-neutral.

## Persistence changes

None.

## Migration strategy

Not applicable — no schema or data migration.

## Concurrency implications

None beyond what pgx's own pool already handles internally; this story only exposes existing pgx pool
knobs as configuration, it does not change pool concurrency behavior.

## Error-handling strategy

Config validation for the new keys follows the existing pattern for `MaxConns` (range/type validation
at config load, fail-closed on invalid values) — exact validation rule (e.g. must be non-negative, or
zero meaning "use pgx default") to be determined at implementation time from pgx's own semantics for
these fields.

## Security controls

`noctx`'s fix is a minor hardening side effect (subprocess becomes cancellable/time-boundable via
context) — not a required security control for this story, but recorded as a beneficial side effect.

## Observability changes

None required by this story's acceptance criteria. A startup log line confirming effective pool-config
values is a reasonable implementation-time addition but not mandated.

## Testing strategy

- Fail-first: run the nine analyzers with today's `.golangci.yml` (all seven zero-cost + noctx +
  copyloopvar in their current configured state) — confirms the "before" state (either not-yet-
  enabled analyzers found nothing when force-run, or the 2/1 named hits, whichever applies).
  Then run again after each fix, and finally after full `.golangci.yml` enablement — confirms the
  "after" state (all nine exit 0).
- Unit test for the new config keys: default value applied when unset; explicit value accepted and
  threaded to the pool-construction call; invalid value rejected per whatever validation rule is
  implemented.
- No new integration or race tests are required — this story does not change concurrent behavior.

## Regression strategy

The `golangci-lint run` itself, wired into CI (already the case — this story enables analyzers within
the existing CI-gated `.golangci.yml`, it does not add a new CI step), is the regression guard: any
future PR introducing an unclosed `sql.Rows` or similar would now fail CI where it previously would
not have.

## Compatibility strategy

New config keys are optional with backward-compatible defaults (see "Compatibility considerations" in
`story.md`). The `noctx`/`copyloopvar` fixes are behavior-preserving.

## Rollout strategy

Single PR/commit; no phased rollout required — this is a CI-config and small-code-fix change, not a
runtime-behavior change requiring gradual exposure.

## Rollback strategy

Revert the `.golangci.yml` change and the two/one code fixes if a false positive or unexpected CI
breakage is found; the config-key addition can be reverted independently since it is additive.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-7). Steps 1-2 (fail-first re-run) must occur
before steps 3-6 (fixes/enablement) begin, per the mandate's fail-first evidence requirement.

## Task breakdown

- **W01-E01-S001-T001** — Zero-cost linter enablement (steps 1, 2, 6, 7 above; fail-first + final
  enablement run).
- **W01-E01-S001-T002** — noctx fix (step 3).
- **W01-E01-S001-T003** — copyloopvar fix (step 4).
- **W01-E01-S001-T004** — MaxConnLifetime/MaxConnIdleTime config keys (step 5).

## Expected artifacts

Updated `.golangci.yml`; updated `internal/cli/config_delegate.go`, `internal/cli/lint_cmd.go`,
`app/maintenance.go`, `kernel/config/config.go`; config documentation update.

## Expected evidence

Fail-first/pass-after `golangci-lint run` logs (per analyzer and combined); config-validation unit
test output.

## Unresolved questions

- Exact field/env-var naming for `MaxConnLifetime`/`MaxConnIdleTime` (to follow `MaxConns`'s existing
  convention — confirm exact convention at implementation time).
- Exact pool-construction call site consuming the new config values (expected within
  `kernel/database/`, not yet confirmed by file/line).
- Whether `noctx`/`copyloopvar` are today fully unconfigured (never mentioned in `.golangci.yml`) or
  configured-but-disabled — affects whether "enabling" them is a new `enable:` list entry or a
  removal from a `disable:` list. To be confirmed by reading `.golangci.yml` at implementation time.
- Whether the 2 named `noctx` sites currently have an available context to thread through, or need
  `context.Background()`/`context.TODO()` — to be determined per call site at implementation time.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above are answered by a
first re-read of `.golangci.yml`/the named call sites at story start, (b) the fail-first re-run
(steps 1-2) confirms or corrects the cited current-state assessment in `story.md`, and (c) the owner
and reviewer are assigned.
