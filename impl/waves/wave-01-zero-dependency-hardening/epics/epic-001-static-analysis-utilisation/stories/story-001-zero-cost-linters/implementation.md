---
id: IMPL-W01-E01-S001
type: implementation-record
parent_story: W01-E01-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E01-S001

Implemented 2026-07-13 by W01Lint at HEAD `0a31186cada5c275a588c74081cf977adf346e61` (working diff;
conductor owns the wave commit). Phase 1 ran the story-local fail-first re-run (fresh triage at HEAD);
Phase 2 (conductor green-light, after sibling W01 workers finished their kernel/http/generator edits)
applied enablement + fixes over the combined working tree and re-verified against it.

## What was actually implemented

1. **`.golangci.yml` enablement (T001)**: `sqlclosecheck`, `rowserrcheck`, `bodyclose`,
   `wastedassign`, `makezero`, `musttag`, `testifylint` (zero-cost set) plus `noctx` and
   `copyloopvar` added to `linters.enable`. The analyzers were previously simply unconfigured, not
   deliberately disabled. Fresh triage at HEAD confirmed the cited zero-hit state for the seven
   zero-cost analyzers; one post-baseline `musttag` hit appeared in sibling-new
   `internal/cli/init_version.go` (W01Gen) and was fixed (json tag on the `go list -m -json` decode
   struct). Config judgment recorded: `noctx` excluded for `_test.go` files (145 of 146 hits were
   context-less `httptest` request construction in tests — request cancellation is meaningless
   there); the exclusion is documented inline in the config.
2. **noctx fixes (T002)** — with recorded drift: noctx v2.11.4 does **not** report the two named CLI
   `exec.Command` sites (upstream noctx checks net/http + database/sql, not os/exec) — see
   `deviations.md` DEV-001. The named sites were fixed anyway per story scope:
   `internal/cli/config_delegate.go` and `internal/cli/lint_cmd.go` now use
   `exec.CommandContext(context.Background(), …)`. The one hit noctx does report in non-test code
   (`testkit/i18n.go:33`, drift vs the cited 2-site list) was fixed with
   `httptest.NewRequestWithContext`.
3. **copyloopvar fix (T003)**: `app/maintenance.go:148`'s dead `rj := rj` removed (named site),
   plus the 6 equally-dead pre-1.22 captures in test files (unsafe_config_matrix_test.go,
   jobs/rls_test.go, outbox_test.go, retention/coverage_test.go, migrations_test.go,
   testkit/rls_isolation_all_test.go) — mechanical deletions, zero semantic change on Go ≥1.22.
4. **Pool lifetime keys (T004)**: `kernel/config/config.go` `Pool` gains `MaxConnLifetime`
   (`conf:"max_conn_lifetime"`, default `1h`) and `MaxConnIdleTime` (`conf:"max_conn_idle_time"`,
   default `30m`) — defaults are exactly pgx v5.10.0's own internal defaults (verified in the pinned
   module source: `pgxpool/pool.go` `defaultMaxConnLifetime = time.Hour`,
   `defaultMaxConnIdleTime = time.Minute * 30`). `Validate()` accepts 0 as "use the pgx default"
   (back-compat for hand-built Pool literals) and bounds non-zero values to 1m–24h / 30s–24h.
   `kernel/database/database.go` `NewPool` wires non-zero values into `pgxpool.Config`; zero leaves
   pgx defaults untouched, so unset deployments observe unchanged pool behavior. Documented in
   `docs/user-guide/configuration.md` (yaml sample, key table, ranges row) — coordinated with
   W01Docs before editing (file confirmed unowned).

## Components changed

`.golangci.yml`; `internal/cli` (2 files); `app/maintenance.go`; `testkit` (i18n.go + 1 test);
`kernel/config` (config.go + 2 test files); `kernel/database` (database.go + coverage_test.go);
5 test files for copyloopvar; `docs/user-guide/configuration.md`.

## Interfaces introduced or changed

`config.Pool` gains two exported fields (additive, defaulted). No behavioral interface change.

## Configuration changes

`db.max_conn_lifetime`, `db.max_conn_idle_time` — new optional keys, pgx-default defaults.

## Tests added or modified

- `kernel/config/config_test.go` `TestPoolLifetimeKeysValidate` (defaults are pgx values; explicit
  pgx-default values validate; zero validates for back-compat).
- `kernel/config/unsafe_config_matrix_test.go`: 4 new reject-matrix entries (below-floor/above-ceiling
  per key).
- `kernel/database/coverage_test.go` `TestIntegrationPoolLifetimeConfigWiring` (explicit values reach
  `pool.Config()`; omitted values keep pgx defaults 1h/30m — the "omission preserves pre-story
  behavior" half of AC-04).

## Commits

Conductor owns commits; delivered as the W01 wave working diff on `0a31186`.

## Known limitations

- The `noctx` fail-before/pass-after pair for the two named CLI sites is evidenced via gosec G204
  (which did flag them) plus the code diff, not via noctx itself — see `deviations.md`.

## Relationship to the approved plan

Matched `plan.md`, with the noctx-detection drift and the test-file exclusion judgment recorded in
`deviations.md` rather than silently absorbed. RISK-W01-E01-002's contingency was exercised for the
new `musttag`/`init_version.go` hit (fixed in-story, recorded).
