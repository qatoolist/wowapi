---
id: W07-E02
type: epic
title: Verification hardening
status: blocked
wave: W07
owner: W07-Phase-A-Execution
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - SEC-05
  - REL-04
  - PERF-06
depends_on: []
stories:
  - W07-E02-S001
  - W07-E02-S002
decisions: []
risks:
  - RISK-W07-002
---

# W07-E02 — Verification hardening

## Epic objective

Establish SEC-05's versioned security-verification profile — a closure gate over SEC-01/03/04/06's own
substantially-complete implementation, backed by an external assessment, not implementable until those
findings exist to map against; and complete REL-04's remaining coverage-truthfulness work (T5-T8: fail-
not-skip E2E prerequisites, a machine-checked skip manifest, a race-integration test schedule, and real
time-bounded coverage-guided fuzzing on PR and scheduled runs), which also owns PERF-06's own identical
T3/T4 fuzz scope per `impl/analysis/conflict-resolution.md` CONFLICT-02.

## Problem being solved

PLAN's own SEC-05 framing: "Standards adoption (ASVS 5.0.0, OWASP API Security Top 10 2023, NIST
800-63-4), not a source-citation finding. Its role is supplying the required test-class checklist
SEC-01–04 already inherit." PLAN's own SEC-05 T1 task row: "Version-pinned control map linking every
applicable control to an executable test or an approved waiver | SEC-01–04 substantially complete |
Independent assessment leaves zero open Critical/High | External assessment | `SEC-05/control-map.md` +
report | **Closure gate**, not implementable until SEC-01–04 exist to map against — Wave 6." (This
programme's own W07 is where that "Wave 6"-labeled dependency resolves, since SEC-01/03/04/06 have all
been built and accepted by prior waves per this wave's own all-prior-waves entry gate.) `requirement-
inventory.md`'s own REL-04 row states: "Truthful integration coverage | QG | P1 | partial | W07-E02-S002
| T1–T4 EXECUTED (verified ×2); T5–T8 planned (T8 owns fuzz, shared w/ PERF-06 T3/T4)." `impl/analysis/
conflict-resolution.md` CONFLICT-02 confirms: "PERF-06 T3/T4 vs REL-04 T8 — Duplicate scope — identical
fuzz scope... Resolution: Single owner REL-04 T8 (target W07-E02-S002)... PERF-06's target story
(W00-E01-S002) proceeds without T3/T4."

## Scope

- SEC-05 T1: the version-pinned control map linking every applicable ASVS 5.0.0/OWASP API Security Top
  10 2023/NIST 800-63-4 control to an executable test or an approved waiver, backed by an external
  assessment (S001).
- REL-04 T5: make E2E prerequisite failures fail, not skip, in the authoritative E2E job (S002).
- REL-04 T6: a machine-checked skip manifest extending `check_test_skips.sh` (S002).
- REL-04 T7: race tests over integration-relevant packages (S002).
- REL-04 T8 (owning PERF-06 T3/T4's identical scope): actual time-bounded coverage-guided fuzzing on
  PRs and scheduled runs (S002).

## Out of scope

- **SEC-01/03/04/06's own implementation** — already built and accepted in W03/W05; this epic's S001
  consumes their `accepted` state as SEC-05's own closure-gate precondition, it does not re-implement
  any of them.
- **PERF-06's own T1/T2** — already `EXECUTED` and verified at W00-E01-S002; not re-implemented here.
- **REL-04's own T1-T4** — already `EXECUTED` and verified twice per `requirement-inventory.md`; not
  re-implemented here, only T5-T8 remain.

## Source requirements

SEC-05 (T1); REL-04 (T5-T8); PERF-06 (T3/T4's scope, owned by REL-04 T8 per CONFLICT-02).

## Architectural context

This epic groups SEC-05 and REL-04 because both are, in different senses, *verification-of-verification*
work: SEC-05 verifies that the framework's own security controls (built across SEC-01/03/04/06) actually
satisfy an external, version-pinned standard, not merely an internal self-assessment; REL-04 verifies
that the framework's own integration-test coverage claims are actually true (fail, not silently skip)
rather than merely appearing green. `impl/analysis/wave-allocation-detail.md`'s own W07-E02 grouping
states this exactly: "S001 security-verification-profile (SEC-05); S002 coverage-truthfulness-completion
(REL-04 T5 fail-not-skip, T6 skip manifest, T7 race-integration schedule, T8 real fuzz — owns PERF-06
T3/T4 scope)." This two-story split is fixed by the canonical allocation.

## Included stories

- **W07-E02-S001 — security-verification-profile** (PLAN SEC-05 T1): the version-pinned control map and
  external assessment, a closure gate over SEC-01-04.
- **W07-E02-S002 — coverage-truthfulness-completion** (PLAN REL-04 T5-T8): fail-not-skip E2E, skip
  manifest, race-integration schedule, real fuzz (owning PERF-06 T3/T4's identical scope).

## Dependencies

**S001 hard-depends on SEC-01 (W03-E01), SEC-03 (W03-E03), SEC-06 (W03-E02), and SEC-04 (W05-E04-S002)
all being `accepted`** — PLAN's own explicit dependency: "SEC-01–04 substantially complete." No
dependency on any other W07 epic. This epic depends transitively on this wave's own all-prior-waves
entry gate, which already subsumes the SEC-01/03/04/06 dependency.

## Risks

RISK-W07-002 (SEC-05's external assessment surfacing an open Critical/High finding with no immediate
remediation path) originates at wave scope and lands entirely within this epic's S001. See `risks.md`
for the epic-scoped elaboration.

## Required decisions

None new. SEC-05 and REL-04 carry no D-0N architecture-decision dependency in `requirement-
inventory.md` §B or REVIEW §F/§U.

## Epic acceptance criteria

- **AC-W07-E02-01**: SEC-05's control map leaves zero open Critical/High findings per the external
  assessment, or each open finding has an approved, documented waiver.
- **AC-W07-E02-02**: REL-04 T5's fail-not-skip enforcement means an unmet E2E prerequisite exits
  non-zero, not "0 tests ran, green."
- **AC-W07-E02-03**: REL-04 T6's machine-checked skip manifest fails CI on a new/unapproved skip and
  passes an approved skip with rationale.
- **AC-W07-E02-04**: REL-04 T7's race tests run over DB/S3-backed packages in CI.
- **AC-W07-E02-05**: REL-04 T8's real fuzzing runs PR + scheduled `-fuzz` with fuzz artifacts proving
  non-zero time beyond seed replay — owning PERF-06 T3/T4's identical scope, with no duplicate
  implementation under a different name.
- **AC-W07-E02-06**: Both stories have passed independent review per mandate §14.

## Closure conditions

Both stories reach `accepted`; AC-W07-E02-01 through AC-W07-E02-06 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date.
