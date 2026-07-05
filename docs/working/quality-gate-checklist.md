# wowapi — Review & Quality Gate Checklist

Run this before declaring any goal/PR done. It is the project-specific instantiation of the
`independent-review-gate` skill. Tick each with **evidence** (file:line, test name, or command output) —
"I think so" is a fail. Prefer running `miscellaneous/review_gate.sh` to automate the mechanical checks.

## A. Does it satisfy the goal?
- [ ] Re-read the goal/roadmap item verbatim; every clause maps to a concrete verified deliverable.
- [ ] Every sub-requirement marked **Done / Partial / Missing**. No Partial/deferral counts as Done unless
      the user explicitly scoped it out.

## B. Built-but-not-wired (the #1 recurring miss)
- [ ] Every new kernel service is a `Kernel` field (`kernel/kernel.go`), exposed on `module.Context`
      (`module/module.go` + `app/context.go`), and wired in `app/boot.go` — if modules must use it.
- [ ] Every new primitive/function is actually invoked on the real runtime path (trace entry→effect).
      Run `miscellaneous/check_unwired.sh` to flag exported symbols with no non-test caller.
- [ ] Security/validation runs at **runtime** (a request/tx exercises it), not only at boot.

## C. Data, migrations, tenancy
- [ ] New tenant tables have `ENABLE`+`FORCE` RLS, a tenant-isolation policy, and correct grants
      (append-only = `SELECT,INSERT` only; cross-tenant = an `app_platform USING(true)` policy + grant).
- [ ] New migration is registered in `migrations/migrations_test.go` and has a correct `-- +goose Down`.
      Run `miscellaneous/check_migrations.sh`.
- [ ] Reversibility drill passes (part of `make ci-container`).

## D. Generated / rendered artifacts actually work
- [ ] Scaffolds/manifests/SQL/config the code emits are parsed/run/booted in a test (e.g. `wowapi init`
      output compiles; `deploy render` output passes config validation).

## E. Tests real & sufficient
- [ ] TDD used; new/changed behavior has a test that fails without the change (prove by revert for subtle
      fixes: concurrency, off-by-one, security).
- [ ] Real Postgres via `testkit`; **no skips masking coverage** — run `miscellaneous/check_test_skips.sh`
      and confirm DB tests run (not skip) under `WOWAPI_REQUIRE_DB=1`.
- [ ] Boundaries + adversarial cases covered (page boundary, parallel/concurrency, rollback, RLS
      isolation, append-only denial, expiry/revocation, injection/fuzz).
- [ ] No duplicate tests — `miscellaneous/find_duplicate_tests.sh`.

## F. Regressions & the authoritative gate
- [ ] `make ci` green (vet, boundary lint, unit, race, perf budgets, build; golangci-lint via `make lint-new`).
- [ ] `make ci-container` green — **0 FAIL, 0 SKIP**, DB tests forced. (`miscellaneous/review_gate.sh`
      can run and grep this.)
- [ ] `gofmt -l` clean; `make lint-boundaries` clean; pre-existing tests still pass.

## G. Required infra & production readiness
- [ ] Any supporting infra the feature needs is delivered (compose service, config knob, migration, role/
      grant) — not just the code.
- [ ] Production config path checked (dedicated DSNs/roles, secretrefs, fail-closed defaults). Not just
      "works in the test DB".

## H. Docs & traceability
- [ ] `decisions.md` entry (before the code) for any deviation; evidence bundle updated; `CHANGELOG.md`
      `[Unreleased]` updated. No "complete" claim next to a deferral —
      run `miscellaneous/check_overclaims.sh`.

## I. One-pass reviewer test
- [ ] Spawned a fresh reviewer agent (did not write the code) on the diff + goal; its findings triaged and
      closed.
- [ ] "First thing a sharp external reviewer would flag" — named and fixed.

## On any finding
Classify severity (Critical/High/Medium/Low) + impact → locate affected requirement/module/test/artifact →
fix per conventions → add/strengthen a test → re-run the gate → **re-run this whole checklist** until no
third-party-review-level issue remains. Log the learning in
[review-learning-register.md](review-learning-register.md).

## Mandatory completion report (attach to the goal)
1. Gate result (passed / iterations) · 2. issues found · 3. severity+impact · 4. fixes · 5. tests added/
updated · 6. re-test output (fail/skip/pass counts) · 7. docs/traceability updates · 8. explicit "no open
third-party-review-level issues remain".
