---
id: W01-E04-S001-T005
type: task
title: Scaffold config validates under the framework-only path (scope addition)
status: done
parent_story: W01-E04-S001
owner: W01Gen
created_at: 2026-07-13
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E04-S001-05
artifacts:
  - ART-W01-E04-S001-006
evidence:
  - EV-W01-E04-S001-005
---

# W01-E04-S001-T005 — Scaffold config validates under the framework-only path

## Task Definition

### Task objective

**Scope addition, conductor-approved 2026-07-13** (recorded as deviation DEV-W01-E04-S001-03; origin:
W01-E04-S002's DEV-03(a) finding, escalated to Main and routed back to this story under its
generator-correctness charter). A pristine `wowapi init` scaffold's `configs/*.yaml` must pass
`wowapi config validate --env local`. At HEAD it fails: the scaffolded `configs/base.yaml` carries an
ACTIVE `i18n:` block (`default_locale`/`supported_locales`/`locales_dir`/`go_bundles`) whose keys are
product-owned (`appcfg.Config.I18n`), unknown to `config.Framework` — so the prebuilt binary's
framework-only validation fallback (D-0002 / `config_delegate.go`: used whenever the product-local
`tools/configcheck` delegation is unavailable) strict-rejects the scaffold's own output. Fail-first:
capture the current failure, then fix at the true source of drift, then pass.

### Parent story

W01-E04-S001.

### Owner

W01Gen (wave-01 generator worker)

### Status

done

### Dependencies

None (disjoint from T001-T004; same story charter).

### Detailed work

1. Write a fail-first test: scaffold a pristine product (`callInit` into `t.TempDir()`), run
   `config validate --dir <dir>/configs --env local` in-process (test cwd `internal/cli` has no
   `tools/configcheck`, exercising exactly the framework-only fallback), assert exit 0 — capture the
   current failure.
2. Determine the true source of drift: template vs schema. Judged TEMPLATE: the i18n keys are
   legitimately product-owned (`appcfg.Config.I18n`); `config.Framework` cannot gain a duplicate
   `conf:"i18n"` section without colliding with every product's embedded-Framework composition; and the
   scaffold's own convention already ships every other product-owned/optional section (`auth`,
   `security`, `storage`, `privileged`, `concurrency`) as a COMMENTED example — the active `i18n:` block
   was the sole violation.
3. Fix: convert the `i18n:` block in `configs_base.yaml.tmpl` to a commented example. Behavior audit:
   every active value except `supported_locales` was the compiled-in default; `supported_locales` is
   read by nothing at runtime (`I18nConfig.Layers()` uses only DefaultLocale/LocalesDir/GoBundles;
   locale negotiation uses the loaded catalog's locales; `wowapi i18n validate` takes `--supported` as
   a flag). The `I18nConfig` zero value is documented valid.
4. Re-run the test (pass) and the full `internal/cli` package (no regression, incl. every
   `TestInitI18n*` test and the end-to-end i18n acceptance test).

### Expected files or components affected

`internal/cli/templates/init/configs_base.yaml.tmpl`; test in `internal/cli/gen_crud_boots_test.go`.

### Expected output / Completion criteria

`TestInitScaffoldConfigValidates` fails pre-fix with the exact `i18n.*: unknown key` rejections and
passes post-fix; the i18n documentation block remains in `base.yaml` (commented); no i18n runtime
behavior changes.

### Required artifacts

ART-W01-E04-S001-006 (updated `configs_base.yaml.tmpl` + the new test).

### Required evidence

EV-W01-E04-S001-005 (`evidence/DX-02/scaffold-config-validate-fix.json`).

### Related acceptance criteria

AC-W01-E04-S001-05 (added with this scope addition): a pristine scaffold's configs pass
`wowapi config validate --env local` under the framework-only path.

### Risks

Low. Residual: `config validate` run INSIDE a pristine scaffold still fails via the OTHER leg — the
`tools/configcheck` delegation cannot compile while the scaffolded `go.mod` carries DX-01's
unresolvable version. That is T001's open defect, not this task's; this task fixes the framework-only
leg and the schema drift.

## Implementation Record

### What was actually implemented

As specified above: `configs_base.yaml.tmpl`'s active `i18n:` block converted to a commented example
(comment text updated to explain WHY it ships commented and that the shown values are the compiled-in
defaults); new fail-first test `TestInitScaffoldConfigValidates` in
`internal/cli/gen_crud_boots_test.go`.

### Files changed

- `internal/cli/templates/init/configs_base.yaml.tmpl` (i18n block hunk only; the file's combined
  working-tree diff also carries sibling W01-E03-S002's http timeout keys — disjoint hunks,
  coordinated over IRC with that owner before editing)
- `internal/cli/gen_crud_boots_test.go` (one new test appended)

### Tests added or modified

`TestInitScaffoldConfigValidates` (new; permanent regression guard that everything init writes ACTIVE
into `configs/*.yaml` stays framework-schema-valid).

### Commits

None yet — uncommitted working-tree delta on top of HEAD `05dce5c8a548f7dce3222637ab2c82024236a2a0`;
the wave conductor owns commits.

### Implementation dates

2026-07-13.

### Known limitations

The delegation leg of in-scaffold `config validate` remains broken by DX-01's go.mod defect (T001,
open). TestInitI18nConfigSection's `assertFileContains(base.yaml, "i18n:")` still passes (the
commented block contains the substring).

### Relationship to the approved plan

Not in the original `plan.md` — scope addition per conductor instruction, recorded as
DEV-W01-E04-S001-03.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S001-05 | Fail-before/pass-after run of `TestInitScaffoldConfigValidates`; full `internal/cli` package regression | Local dev or CI, `go test ./internal/cli/` | Fails pre-fix with `i18n.default_locale/go_bundles/locales_dir/supported_locales: unknown key`; passes post-fix with `config OK` | functional-test report (fail-before/pass-after pair) | pending — wave-level review gate |

### Actual result

Pre-fix: FAIL with exactly the four `i18n.*: unknown key` rejections
(`evidence/DX-02/t005-scaffold-config-validate-prefix-failfirst.log`). Post-fix: PASS
(`.../t005-scaffold-config-validate-postfix.log`); full package `ok` 13.0s
(`.../pkg-internal-cli-full-2.log`).

### Pass or fail

PASS (fail-before/pass-after pair complete).

### Evidence identifier

EV-W01-E04-S001-005.

### Execution date

2026-07-13 (07:45 UTC).

### Commit or revision

HEAD `05dce5c8a548f7dce3222637ab2c82024236a2a0`; fix uncommitted on top (conductor commits).

### Environment

macOS Darwin 25.5.0 arm64, go1.26.5, local dev workstation.

### Reviewer

Pending — wave-level review gate (conductor assigns).

### Findings

None beyond the recorded residual (DX-01 delegation leg, T001's scope).

### Retest status

Not required — first-pass verification succeeded at the pinned revision.

### Final conclusion

AC-W01-E04-S001-05 satisfied; scaffold config is framework-schema-valid again, with the i18n
documentation value preserved as a commented example.

## Deviations Record

This task IS a recorded deviation (scope addition) — see story `deviations.md` DEV-W01-E04-S001-03.
No further intra-task deviation.
