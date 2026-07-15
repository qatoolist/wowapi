---
id: PLAN-W04-E04-S003
type: plan
parent_story: W04-E04-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E04-S003

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. **DX-07 T4 is not planned by this document in any form** — it is explicitly
excluded, forward-referenced by requirement ID (AR-04 T5) and target story (W05-E03-S002) only.

## Proposed architecture

Three independent, additive fixes to the existing, already-correct readiness mechanism
(`kernel/httpx/health.go:52-79`, confirmed by MATRIX CS-21 as "correct and fail-closed"): (1) wiring
a migration-currency check into the generated readiness template, closing the contract-by-comment gap
at `app/health.go:9-14`; (2) extending the readiness payload with seed/rule/model-hash reporting; (3)
replacing `config_delegate.go`'s CWD-relative `os.Stat` product-root discovery with a `go env
GOMOD`/`--project`-based discovery. None of the three requires changing the underlying health-check
execution mechanism itself — each is a new check registration or a discovery-logic fix layered onto
the existing, sound infrastructure.

## Implementation strategy

1. Re-read the generated `cmd/api/main.go.tmpl`'s readiness map, `app/health.go:9-14`'s documented
   DB/migration-check contract comment, and `config_delegate.go`'s current discovery logic at this
   story's actual start commit to confirm the current-state assessment holds.
2. **T1**: Determine the "expected migration version" source (PLAN T1's own risk note: "needs a
   stable 'expected migration version' source") — likely the migration directory's own highest-
   numbered/latest migration, to be confirmed against the existing migration-registration mechanism.
   Implement the migration-currency check in the generated readiness template; wire it into the
   readiness map alongside the existing `"db"`/`"seeds"` checks.
3. Write the stale-migration 503 integration test: boot against a database at a lagging migration
   version, assert the readiness endpoint returns 503.
4. **T2**: Confirm AR-01's model hash's availability at this story's implementation time (dependency
   noted in `story.md` "Assumptions"). Implement seed/rule-hash reporting (independent of AR-01) and
   model-hash reporting (contingent on AR-01's model hash) as additions to the readiness payload.
5. Write the full-readiness-payload integration test: confirm the payload reports migration version,
   seed/rule hash, and (if available) model hash.
6. **T3**: Replace `config_delegate.go`'s CWD-relative `os.Stat` discovery with `go env GOMOD`/
   `--project`-based discovery; add explicit reporting of whether product validation ran (both success
   and fallback cases).
7. Write the nested-subdirectory and outside-repo-with-`--project` discovery tests.
8. Document all three changes; explicitly document T4's exclusion and its forward reference.

## Expected package or module changes

The generated `cmd/api/main.go.tmpl` template; `app/health.go` (or its readiness-check registration
logic); `config_delegate.go`.

## Expected file changes where determinable

- The generated `cmd/api/main.go.tmpl` — add the migration-currency check registration and the seed/
  rule/model-hash reporting fields to the readiness map.
- `app/health.go` — the DB/migration-check contract's own comment (`:9-14`) becomes an actually-wired
  check rather than a comment-only contract.
- `config_delegate.go` — replace CWD-relative `os.Stat` discovery with `go env GOMOD`/`--project`-
  based discovery; add explicit product-validation-ran reporting.
- New integration test files for the stale-migration 503 test and the full-readiness-payload test;
  new unit test files for the nested-subdirectory and outside-repo-`--project` discovery cases.

## Contracts and interfaces

`/readyz`'s response payload gains new fields (migration version, seed/rule hash, model hash) — an
additive, non-breaking payload change for any consumer that does not assume a fixed field set.
`config doctor`'s CLI output gains explicit product-validation-ran reporting — an additive output
change.

## Data structures

The readiness payload's own struct gains fields for migration version, seed/rule hash, and model
hash. No other data-structure change.

## APIs

`/readyz`'s HTTP response shape is extended (additive); its pass/fail (200/503) behavior changes for
any deployment currently running against a stale-migrated database, per `story.md` "Compatibility
considerations" — this is the intended, contract-restoring behavior, not an API addition requiring a
version bump beyond the payload extension itself.

## Configuration changes

The "expected migration version" source (T1, step 2 above) may require a new configuration surface
or may be derivable entirely from the existing migration-registration mechanism without new
configuration — to be determined at implementation time.

## Persistence changes

None. This story's three tasks are readiness/diagnostics logic changes; no schema or table change is
anticipated.

## Migration strategy

Not applicable in the schema-migration sense. This story's own T1 task concerns *detecting* migration
currency, not performing a migration itself.

## Concurrency implications

None identified beyond what `kernel/httpx/health.go`'s existing per-check timeout/concurrency model
already handles — the new migration-currency and hash checks are expected to run within that existing
model, not introduce a new one.

## Error-handling strategy

The migration-currency check must fail closed (503) on a genuine version-lag detection, and must not
itself introduce a new failure mode that could cause a false 503 on a correctly-migrated database
(e.g. a flaky "expected version" source). `config doctor`'s discovery fix must explicitly report
which mode it operated in (product validation ran vs. fell back to framework-only) rather than
silently choosing one.

## Security controls

None beyond the readiness-truthfulness property itself (see `story.md` "Security considerations") —
a `/readyz` endpoint that stops silently masking a stale-migration or missing-product-validation
condition.

## Observability changes

The readiness payload's new migration-version/seed-rule-hash/model-hash fields and `config doctor`'s
explicit product-validation-ran reporting are themselves the observability improvements this story
delivers — not a separate addition on top of the functional fix.

## Testing strategy

- T1: stale-migration 503 integration test — boot against a database at a lagging migration version,
  assert 503, per PLAN T1's own test column exactly.
- T2: full-readiness-payload integration test — confirm migration version, seed/rule hash, and (if
  available) model hash are all reported.
- T3: nested-subdirectory and outside-repo-with-`--project` unit tests — confirm discovery works
  correctly in both cases and explicitly reports whether product validation ran, per PLAN T3's own
  test column exactly.

## Regression strategy

Each of the three tests above becomes the regression guard for its own task's scope going forward.

## Compatibility strategy

T1/T2's payload and pass/fail-behavior changes are additive/contract-restoring, not breaking, per
`story.md` "Compatibility considerations." T3's discovery fix is confirmed non-breaking for
wowsociety (its own `tools/configcheck/main.go` already exists and engages correctly). No transition
period or compatibility flag is planned for any of the three tasks, since none removes existing
behavior — each adds a previously-missing, already-documented-as-required check or reporting field.

## Rollout strategy

Single story, landed as its own reviewable unit. T1-T3 have no forced internal order beyond T2's
model-hash portion depending on AR-01's availability — T1 can ship independently per PLAN's own
framing, and T3 is fully independent of T1/T2.

## Rollback strategy

Each of T1/T2/T3's changes is independently revertible without a data-migration concern — none
introduces a schema change. A revert of T1 would restore the previous (contract-violating) behavior
of `/readyz` reporting healthy despite a migration lag; this would be a deliberate, documented
regression if ever exercised, not a silent one.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–8). No hard sequencing constraint between
T1/T2/T3's own implementation order beyond T2's model-hash portion's AR-01 contingency.

## Task breakdown

- **W04-E04-S003-T001** — Migration-currency readiness check (T1; steps 2–3 above).
- **W04-E04-S003-T002** — Seed/rule/model-hash readiness reporting (T2; steps 4–5 above).
- **W04-E04-S003-T003** — `config doctor` product-root discovery fix (T3; steps 6–7 above).
- **W04-E04-S003-T004** — Independent review (per mandate §14, scoped to this story, with specific
  attention to T4's correct exclusion).

**DX-07 T4 has no task in this breakdown.** It is explicitly out of scope — see `story.md` "Out of
scope."

## Expected artifacts

The migration-currency readiness check; the seed/rule/model-hash readiness reporting change; the
`config doctor` discovery fix; documentation of all three, including the explicit T4 out-of-scope
note.

## Expected evidence

Stale-migration 503 integration-test output; full-readiness-payload integration-test output; nested-
subdirectory and outside-repo-`--project` config-doctor discovery test output.

## Unresolved questions

- The exact "expected migration version" source for T1 (PLAN T1's own risk note flags this as needing
  "a stable 'expected migration version' source") — to be determined at implementation time.
- Whether AR-01's model hash is available at this story's implementation time; if not, T2's model-
  hash reporting portion may need to be sequenced separately or its status recorded honestly as
  partial in `deviations.md` — not silently claimed complete.
- The exact readiness-payload field names/shape for migration version, seed/rule hash, and model
  hash — to be determined at implementation time.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above — most centrally,
the expected-migration-version source and AR-01's model-hash availability — are answered or their
contingency explicitly recorded, and (b) the owner and reviewer are assigned. This plan does not
require W02-E01, W04-E04-S001, or W04-E04-S002 to have landed first, per `story.md` "Dependencies."
