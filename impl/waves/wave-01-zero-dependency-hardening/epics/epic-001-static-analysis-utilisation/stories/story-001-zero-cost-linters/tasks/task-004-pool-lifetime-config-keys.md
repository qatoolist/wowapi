---
id: W01-E01-S001-T004
type: task
title: MaxConnLifetime/MaxConnIdleTime pool config keys
status: done
parent_story: W01-E01-S001
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S001-04
artifacts:
  - ART-W01-E01-S001-005
  - ART-W01-E01-S001-006
evidence:
  - EV-W01-E01-S001-004
---

# W01-E01-S001-T004 — MaxConnLifetime/MaxConnIdleTime pool config keys

## Task Definition

### Task objective

Expose `MaxConnLifetime` and `MaxConnIdleTime` as new database connection-pool configuration keys,
defaulted to pgx's own internal defaults, motivated by credential-rotation and load-balancer-rebalance
operational hygiene (not by any linter finding — this is new capability, not a triage output).

### Parent story

W01-E01-S001 — Enable the zero-cost leak-detection linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001/T002/T003 (disjoint files); shares only the epic's CS-10/pgx-pool
context, not a code dependency.

### Detailed work

1. Read `kernel/config/config.go` around the existing `MaxConns` field (lines 99, 211-212 per
   `story.md`'s current-state citation) to confirm the existing naming/validation/env-var-binding
   convention.
2. Add `MaxConnLifetime` and `MaxConnIdleTime` fields to the config struct, following that same
   convention, typed to match pgx's own `pgxpool.Config` field types (`time.Duration`).
3. Determine pgx's actual internal default values for these two fields (from the pinned pgx version
   in `go.mod`) and set them as this config's defaults, so omitting the new keys is behavior-neutral.
4. Add validation for the new fields consistent with the existing `MaxConns` validation pattern
   (fail-closed on invalid values, e.g. negative durations).
5. Locate and update the pool-construction call site (expected within `kernel/database/`, exact
   file/line to be confirmed at implementation time) to thread the new config values into
   `pgxpool.Config`.
6. Write a unit test confirming: (a) default value applied when the new keys are unset and pool
   behavior matches pre-task behavior; (b) an explicit value is accepted and threaded through to the
   constructed `pgxpool.Config`; (c) an invalid value is rejected per the validation rule from step 4.
7. Document the new keys alongside the existing `MaxConns` documentation.

### Expected files or components affected

`kernel/config/config.go`; the pool-construction call site within `kernel/database/` (file to be
confirmed); the config documentation location (to be confirmed).

### Expected output

Two new, validated, documented, backward-compatible config keys with a passing unit test.

### Required artifacts

ART-W01-E01-S001-005, ART-W01-E01-S001-006.

### Required evidence

EV-W01-E01-S001-004 (unit-test report).

### Related acceptance criteria

AC-W01-E01-S001-04.

### Completion criteria

The config-validation unit test passes, covering default/explicit/invalid-value cases; the pool
constructor demonstrably consumes the new values when set.

### Verification method

`go test ./kernel/config/... -run TestMaxConn -v` (or the actual test name chosen at implementation
time), logged output retained as evidence.

### Risks

Low — additive, optional config keys with defaults chosen to reproduce current behavior. Primary risk
is picking a default that does not actually match pgx's internal default, which the unit test's
default-value case is designed to catch.

### Rollback or recovery considerations

Revert the config-key addition independently of T001-T003 if the pgx-default-matching assumption
proves wrong and produces unexpected pool behavior in an environment that does not set the new keys
explicitly.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

Added `MaxConnLifetime` (default 1h) / `MaxConnIdleTime` (default 30m) to `config.Pool` — defaults verified equal to pgx v5.10.0's internal defaults from the pinned module source. Validation: 0 = pgx-default sentinel (back-compat); non-zero bounded 1m–24h / 30s–24h. `database.NewPool` wires non-zero values into `pgxpool.Config`. Documented in `docs/user-guide/configuration.md` (coordinated with W01Docs). Tests: `TestPoolLifetimeKeysValidate`, 4 reject-matrix entries, `TestIntegrationPoolLifetimeConfigWiring` (explicit values reach `pool.Config()`; omitted values keep pgx defaults).

## Verification Record

AC-W01-E01-S001-04: all new tests pass against the real compose DB (`kernel/config` ok, `kernel/database` ok — EV-004, `evidence/tests/touched-package-test-sweep.log`). Fail-first by construction (keys/tests do not compile at HEAD). **pass**

### Final conclusion

Keys shipped with pgx-default-preserving semantics, validated and integration-tested.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
