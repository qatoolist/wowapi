---
id: W07-E02-S001-T001
type: task
title: Build the version-pinned control map
status: blocked
parent_story: W07-E02-S001
owner: W07-Phase-A-Execution.W07E02S001
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E02-S001-01
artifacts:
  - ART-W07-E02-S001-001
  - ART-W07-E02-S001-003
evidence:
  - EV-W07-E02-S001-001
  - EV-W07-E02-S001-003
---

# W07-E02-S001-T001 — Build the version-pinned control map

## Task Definition

### Task objective

Enumerate every applicable control from ASVS 5.0.0, OWASP API Security Top 10 2023, and NIST 800-63-4, and link each to an executable test or an approved waiver.

### Parent story

W07-E02-S001

### Owner

W07-Phase-A-Execution.W07E02S001

### Status

blocked

### Dependencies

SEC-01 (W03-E01), SEC-06 (W03-E02), SEC-03 (W03-E03), and SEC-04 (W05-E04) must all be `accepted`. Execution-time verification found 0/7 checked story/closure pairs consistently accepted; EV-W07-E02-S001-003 records the failed dependency.

### Detailed work

1. Enumerate every applicable control from the three named standards, informed by the framework's
   own actual capability surface.
2. For each control, identify an existing executable test, or add a small bounded new test, or record
   an approved waiver.
3. Assemble the control map document.

### Expected files or components affected

SEC-05/control-map.md (new); possibly small, bounded new test files per genuine gap.

### Expected output

A complete, version-pinned control map.

### Required artifacts

ART-W07-E02-S001-001 (the control map).

### Required evidence

EV-W07-E02-S001-001 (control-map completeness report).

### Related acceptance criteria

AC-W07-E02-S001-01.

### Completion criteria

Every applicable control is linked to an executable test or an approved waiver.

### Verification method

Direct inspection of the control map against the three named standards.

### Risks

None beyond the general risk that a genuinely missing test surfaces a larger-than-expected gap — mitigated by this task's own bounded-scope framing (small additions only; a large gap is scoped as its own follow-up item).

### Rollback or recovery considerations

If a control-map entry is later found incorrect, correct it directly; a documentation artifact does not require the same rollback discipline as a code change.

## Implementation Record

### What was actually implemented

Built the pinned 412-entry control map, local source inventories, validator, validator regression
tests, and accepted-state prerequisite checker. No production file changed.

### Components and files changed

`SEC-05/control-map.{json,md}`, `SEC-05/sources/*`, `SEC-05/validate_control_map.py`,
`SEC-05/test_validate_control_map.py`, and `SEC-05/verify_prerequisites.py`.

### Interfaces, configuration, schema, security behavior, and observability

No production interface/configuration/schema/security-behavior/observability change. A verification-only
CLI was added: `python3 SEC-05/validate_control_map.py [--run-tests]`.

### Tests added or modified

Six validator regression tests added. Existing focused Go tests were referenced and executed but not
modified.

### Commits and pull requests

No story-owned commit or PR. Execution observed at
`733ef3e930cbb3f89f5bbc53d8f562c60e426513` in a shared dirty workspace.

### Implementation dates

2026-07-14.

### Technical debt and known limitations

No production debt. External applicability confirmation is still required. The upstream accepted-state
precondition fails and final evidence must be re-pinned to a clean integration commit.

### Follow-up items

Reconcile upstream lifecycle records and re-run at the integration commit.

### Relationship to the approved plan

Plan steps 1–3 matched. The planning assumption that upstream stories were accepted was false in the
current lifecycle records; see DEV-W07-E02-S001-002.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E02-S001-01 | Validate catalogs and run all mapped tests | Python 3.14.2; Go 1.26.5; required DB/S3 env | Every applicable mapping resolves and passes | EV-W07-E02-S001-001; EV-W07-E02-S001-004 | W05ReviewGateFinal — PASS |

### Actual result

Map validation passed with `total=412 applicable=33 not-applicable=379 waived=0`; five focused Go
package invocations and six Python tests passed. Prerequisite verification failed 7/7 lifecycle pairs.

### Pass or fail

Functional map check: pass. Task lifecycle: blocked by failed hard dependency and final revision pin.

### Evidence identifiers

EV-W07-E02-S001-001 and EV-W07-E02-S001-003.

### Execution date and revision

2026-07-13T21:17:50Z at observed HEAD
`733ef3e930cbb3f89f5bbc53d8f562c60e426513`; artifact hashes recorded in EV-001.

### Environment

Darwin arm64, Go 1.26.5, Python 3.14.2, local PostgreSQL with `WOWAPI_REQUIRE_DB=1` and
`WOWAPI_REQUIRE_S3=1`.

### Reviewer

W05ReviewGateFinal: PASS; no open actionable story-scope issue.

### Findings and retest status

The map itself has no open validator finding. Upstream lifecycle acceptance and clean-commit retest are
pending.

### Final conclusion

Implementation complete and focused behavior passes; task remains `blocked`, not `done`.

## Deviations Record

DEV-W07-E02-S001-002 records the failed planning assumption that SEC-01/03/04/06 lifecycle records were
already accepted. No exception is approved; the checker and failed evidence are compensating
truthfulness controls until upstream owners resolve the state.
