---
id: W07-E02-S002
type: story
title: Coverage truthfulness completion — fail-not-skip E2E, skip manifest, race schedule, real fuzz
status: accepted
wave: W07
epic: W07-E02
owner: W07-E02-S002 executor
reviewer: W05ReviewGateFinal
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - REL-04
  - PERF-06
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W07-E02-S002-01
  - AC-W07-E02-S002-02
  - AC-W07-E02-S002-03
  - AC-W07-E02-S002-04
artifacts:
  - ART-W07-E02-S002-001
  - ART-W07-E02-S002-002
  - ART-W07-E02-S002-003
  - ART-W07-E02-S002-004
evidence:
  - EV-W07-E02-S002-001
  - EV-W07-E02-S002-002
  - EV-W07-E02-S002-003
  - EV-W07-E02-S002-004
decisions: []
risks:
  - RISK-W07-E02-001
---

# W07-E02-S002 — Coverage truthfulness completion — fail-not-skip E2E, skip manifest, race schedule, real fuzz

## Story ID

W07-E02-S002

## Title

Coverage truthfulness completion — fail-not-skip E2E, skip manifest, race schedule, real fuzz

## Objective

Make unmet E2E prerequisites fail, not skip, in the authoritative E2E job (T5); build a machine-checked
skip manifest extending `check_test_skips.sh` (T6); run race tests over integration-relevant packages in
CI (T7); and implement actual time-bounded coverage-guided fuzzing on PRs and scheduled runs (T8) —
**this story owns PERF-06's own identical T3/T4 fuzz scope**, per `impl/analysis/conflict-
resolution.md` CONFLICT-02's single-ownership resolution.

## Value to the framework

PLAN's own REL-04 evidence for T5: "Make E2E prerequisite failures fail, not skip, in the authoritative
E2E job | — | Unmet prerequisite exits non-zero, not '0 tests ran, green' | Kill a required E2E
dependency, confirm failure | `REL-04/e2e-fail-closed/` | Medium — requires classifying which of the 22
inventoried skip sites are legitimately optional vs. mask required coverage." This story converts "the
test suite reported green" into "the test suite reported green because it actually ran the required
tests," closing the exact class of false-confidence gap a CI system can silently accumulate over time.
T8's own dual ownership (REL-04 and PERF-06) is resolved here per CONFLICT-02's own rationale:
"REL-04's own PLAN §9 text already defers to PERF-06's resolution, and `requirement-inventory.md`'s
REL-04 row states the single-owner assignment explicitly."

## Problem statement

PLAN's own REL-04 task table: T6, "Machine-checked skip manifest extending `check_test_skips.sh` | T1-T5
(only meaningful once known-bad skips are fixed) | New/unapproved skip fails CI; approved skip (with
rationale) passes | Fixture: add an unguarded `t.Skip()`, prove failure | `REL-04/skip-manifest/` |
Medium." T7, "Race tests over integration-relevant packages | — | `go test -race` runs over DB/S3-backed
packages in CI | CI run + seeded data-race fixture | `REL-04/race-integration/` | Medium — may need a
separate scheduled job, not every PR." T8, "Actual time-bounded coverage-guided fuzzing on PRs +
scheduled — **identical scope to PERF-06 T3/T4, same evidence text, assign single ownership (recommend
PF-REL, coverage-truthfulness framing) to avoid duplicate implementation** | — | PR + scheduled `-fuzz`
runs, fuzz artifacts prove non-zero time beyond seed replay | Fuzz-duration/corpus-mtime test |
`REL-04/fuzz-time-bounded/` | Medium — **shared with PERF-06, coordinate ownership before implementing**."
MATRIX CS-13's own evidence confirms the current gap: "hosted fuzzing never runs — CI replays seed
corpus only (`ci.yml:98-101`, `-run '^Fuzz'`, no `-fuzz=`), `make test-fuzz` exists un-wired."

## Source requirements

REL-04 (T5–T8); PERF-06 (T3/T4's identical scope, owned here per CONFLICT-02).

## Current-state assessment

Per PLAN's own evidence and MATRIX CS-13 (to be re-confirmed at this story's own execution commit): an
unmet E2E prerequisite today produces "0 tests ran, green" rather than a failure — a false-positive
signal. No machine-checked skip manifest exists beyond `check_test_skips.sh`'s own current scope (T1-T5's
own EXECUTED work, per `requirement-inventory.md`). Race tests may not run consistently over DB/S3-backed
packages in CI. Hosted CI replays only the existing seed corpus for the repository's 2 existing fuzz
test files — never real `-fuzz` coverage-guided generation; `make test-fuzz` exists but is un-wired from
CI.

## Desired state

Killing a required E2E dependency and re-running the authoritative E2E job produces a non-zero exit, not
a silent "0 tests ran, green." A machine-checked skip manifest, extending `check_test_skips.sh`, fails
CI on a new/unapproved skip (proven by a fixture: add an unguarded `t.Skip()`, prove failure) and passes
an approved skip carrying a documented rationale. `go test -race` runs over DB/S3-backed packages in CI
(possibly as a separate scheduled job, per PLAN's own risk note, not necessarily every PR). PR and
scheduled CI runs invoke real `-fuzz=<Name> -fuzztime=Ns` coverage-guided fuzzing, with fuzz artifacts
proving non-zero fuzzing time beyond seed replay, and a longer scheduled fuzzing job retains its corpus
across runs.

## Scope

- **T5** — Fail-not-skip E2E prerequisites: classify each of the 22 inventoried skip sites as
  legitimately optional or masking required coverage; make masking-required-coverage cases fail, not
  skip.
- **T6** — A machine-checked skip manifest extending `check_test_skips.sh`; fixture proving a new
  unapproved skip fails CI.
- **T7** — Race tests (`go test -race`) over DB/S3-backed packages in CI, possibly as a separate
  scheduled job.
- **T8** — Actual time-bounded coverage-guided fuzzing on PR (short) and scheduled (longer, corpus-
  retained) runs — **owns PERF-06 T3/T4's identical scope**.

## Out of scope

- **REL-04's own T1-T4** — already `EXECUTED` and independently reviewed twice; not re-implemented here.
- **PERF-06's own T1/T2** — already `EXECUTED` at W00-E01-S002; this story owns only PERF-06's T3/T4
  fuzz scope, folded into this story's own T8, per CONFLICT-02.
- **Remediating every bug the real fuzzer finds once wired in** — per RISK-W07-E02-001's own framing
  (epic-scoped), a genuine bug found by working fuzz infrastructure is tracked as its own separate item,
  not silently absorbed into this story's own scope.

## Assumptions

- The exact classification of each of the 22 inventoried skip sites (legitimately optional vs. masking
  required coverage) is not pre-determined by any source document — PLAN's own T5 risk note frames this
  as the task's own central work, not a fact this planning document can state in advance.
- T7's own "possibly a separate scheduled job, not every PR" framing is confirmed from PLAN's own risk
  note as a genuine open implementation choice, not a pre-decided fact.
- T8's own fuzz-time-bound value (the exact `-fuzztime` duration for PR vs. scheduled runs) is not
  specified by any source document beyond "time-bounded" — this story's own implementation determines
  the exact duration, balancing CI-time cost against fuzzing depth.

## Dependencies

None within W07-E02 (independent of S001). No dependency on any other W07 epic. Depends transitively on
this wave's own all-prior-waves entry gate, and specifically on REL-04's own T1-T4 (already `EXECUTED`)
and PERF-06's own T1/T2 (already `EXECUTED`) as its own starting point, though neither is a blocking
entry criterion since both are already satisfied.

## Affected packages or components

The authoritative E2E job's own workflow configuration (T5); `check_test_skips.sh` and its own manifest
extension (T6); CI workflow configuration for the race-test job (T7); CI workflow configuration for the
PR and scheduled fuzz jobs, plus `make test-fuzz`'s own wiring (T8).

## Compatibility considerations

T5's own fail-not-skip conversion is an intentional behavior change: any workflow relying today on an
unmet E2E prerequisite silently producing a "green" result will now see that same condition correctly
fail. This is a strict correctness improvement — the old behavior was a false-positive signal, not a
legitimate "pass."

## Security considerations

T8's own real coverage-guided fuzzing is itself a security-adjacent capability — fuzz testing is a
standard technique for surfacing memory-safety and input-validation defects; wiring it from seed-replay-
only to genuine `-fuzz` execution meaningfully increases the framework's own defect-discovery surface.

## Performance considerations

T7 and T8 both have real CI-time-cost implications (race tests and fuzz runs are both slower than
standard unit tests) — PLAN's own risk notes for both tasks acknowledge this explicitly ("may need a
separate scheduled job, not every PR" for T7; "CI runtime budget impact" is the underlying concern for
T8, inherited from PERF-06 T3's own risk note).

## Observability considerations

T8's own fuzz artifacts (proving non-zero fuzzing time beyond seed replay) are themselves an
observability/audit artifact confirming the fuzz job actually did real work, not merely replayed a seed
corpus and reported success.

## Migration considerations

Not applicable.

## Documentation requirements

Document the skip-manifest's own approval process (T6) — how a legitimate skip gets approved with
rationale, versus how an unapproved skip fails CI; document the fuzz-corpus-retention mechanism (T8) —
artifact vs. commit, per PLAN's own noted implementation choice.

## Acceptance criteria

- **AC-W07-E02-S002-01**: Killing a required E2E dependency and re-running the authoritative E2E job produces a
  non-zero exit, not "0 tests ran, green."
- **AC-W07-E02-S002-02**: A fixture adding an unguarded `t.Skip()` fails CI via the machine-checked skip
  manifest; an approved skip with documented rationale passes.
- **AC-W07-E02-S002-03**: `go test -race` runs over DB/S3-backed packages in CI (as a per-PR or scheduled job,
  per this story's own implementation-time decision), proven by a seeded data-race fixture.
- **AC-W07-E02-S002-04**: PR and scheduled CI runs invoke real `-fuzz=<Name> -fuzztime=Ns` coverage-guided
  fuzzing, with fuzz artifacts proving non-zero fuzzing time beyond seed replay; a longer scheduled job
  retains its corpus across runs — this AC is the single, owned closure of both REL-04 T8 and PERF-06
  T3/T4's identical scope, with no duplicate implementation under either name.

## Required artifacts

- The fail-not-skip E2E job configuration + the 22-skip-site classification record (T5).
- The machine-checked skip manifest extension (T6).
- The race-test CI job configuration (T7).
- The real-fuzz PR and scheduled CI job configuration, plus the corpus-retention mechanism (T8).
See `artifacts/index.md`.

## Required evidence

- Kill-a-required-dependency test output confirming non-zero exit (T5).
- Unguarded-`t.Skip()` fixture fail-test output (T6).
- Seeded data-race fixture test output (T7).
- Fuzz-duration/corpus-mtime test output confirming non-zero fuzzing time (T8).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all four acceptance criteria numbered and measurable, no dependency, owner/
reviewer assignment pending, the 22-skip-site classification work and T8's exact fuzz-time-bound value
recorded as unresolved questions rather than silently pre-decided.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T8's own single-ownership resolution against PERF-06
T3/T4 is genuine (no separate, duplicate fuzz-wiring implementation exists anywhere else in the
repository under PERF-06's own name).

## Risks

RISK-W07-E02-001 (T8's real-fuzz work may discover a genuine bug requiring remediation outside this
story's own bounded scope) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once all four acceptance criteria are verified and T8's single-ownership resolution against PERF-06 is
confirmed genuine, residual risk is expected to be low, with the caveat that a genuine bug found by the
now-working fuzz infrastructure is expected and should be filed as its own tracked item, not treated as
this story's own defect.

## Plan

See `plan.md`.
