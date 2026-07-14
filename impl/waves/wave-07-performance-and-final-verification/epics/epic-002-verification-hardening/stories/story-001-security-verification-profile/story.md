---
id: W07-E02-S001
type: story
title: Security verification profile — version-pinned control map + external assessment
status: blocked
wave: W07
epic: W07-E02
owner: W07-Phase-A-Execution.W07E02S001
reviewer: W05ReviewGateFinal
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - SEC-05
depends_on:
  - W03-E01
  - W03-E02
  - W03-E03
  - W05-E04
blocks: []
acceptance_criteria:
  - AC-W07-E02-S001-01
  - AC-W07-E02-S001-02
artifacts:
  - ART-W07-E02-S001-001
  - ART-W07-E02-S001-002
  - ART-W07-E02-S001-003
  - ART-W07-E02-S001-004
evidence:
  - EV-W07-E02-S001-001
  - EV-W07-E02-S001-002
  - EV-W07-E02-S001-003
  - EV-W07-E02-S001-004
decisions: []
risks:
  - RISK-W07-002
---

# W07-E02-S001 — Security verification profile — version-pinned control map + external assessment

## Story ID

W07-E02-S001

## Title

Security verification profile — version-pinned control map + external assessment

## Objective

Build a version-pinned control map linking every applicable control from ASVS 5.0.0, OWASP API Security
Top 10 2023, and NIST 800-63-4 to an executable test or an approved waiver, and obtain an independent
external assessment leaving zero open Critical/High findings (or each with an approved waiver). **This
story is explicitly a closure gate, not implementable until SEC-01-04 substantially complete.**
Execution-time verification on 2026-07-14 disproved the planning-time claim that this entry condition
was already satisfied: none of the seven checked upstream `story.md`/`closure.md` pairs is consistently
`accepted` (EV-W07-E02-S001-003).

## Value to the framework

PLAN's own SEC-05 framing states its role precisely: "Standards adoption (ASVS 5.0.0, OWASP API Security
Top 10 2023, NIST 800-63-4), not a source-citation finding. Its role is supplying the required
test-class checklist SEC-01–04 already inherit." Without this story, the framework's own security
posture across SEC-01 (server-side session state), SEC-03 (webhook replay), SEC-04 (authz cache
bounding), and SEC-06 (outbound-security governance) has been individually verified against each
finding's own acceptance criteria, but never independently mapped against an external, version-pinned
standard and assessed by a party that did not write the implementation. This story is the framework's
own equivalent of the independent-review-gate this whole programme already applies at the story level —
applied instead to the framework's entire security posture as a single subject.

## Problem statement

PLAN's own SEC-05 T1 task row, quoted in full: "Version-pinned control map linking every applicable
control to an executable test or an approved waiver | SEC-01–04 substantially complete | Independent
assessment leaves zero open Critical/High | External assessment | `SEC-05/control-map.md` + report |
**Closure gate**, not implementable until SEC-01–04 exist to map against — Wave 6." PLAN's own
wowsociety baseline note gives a head start on the test-infrastructure side: "real adversarial-test
infrastructure already exists to plug into a future control map (`abac_test.go`, `authz_matrix_test.go`,
`rls_test.go`, `stepup_test.go`, `otp_test.go`/`totp_test.go`, `whoami_impersonation_test.go`) — but
these validate wowsociety's *own* product-layer workarounds, so expect rework once SEC-01 ships, not
pure addition."

## Source requirements

SEC-05 (T1).

## Current-state assessment

The version-pinned map and validator now exist under `SEC-05/` and focused mapped tests pass. No
external assessment has been performed. `python3 SEC-05/verify_prerequisites.py` found every checked
SEC-01/03/04/06 story/closure pair lifecycle-inconsistent and 0/7 consistently `accepted`; the
closure-gate precondition is therefore not satisfied.

## Desired state

A version-pinned control map (`SEC-05/control-map.md` or equivalent) exists, linking every applicable
ASVS 5.0.0 / OWASP API Security Top 10 2023 / NIST 800-63-4 control to either an executable test already
existing in the framework's own test suite (per PLAN's own wowsociety-baseline note, potentially reusing
or adapting wowsociety's own adversarial-test infrastructure) or an approved, documented waiver. An
independent external assessment has been performed, leaving zero open Critical/High findings, or each
open finding carries an approved, time-bounded waiver.

## Scope

- Build the version-pinned control map, enumerating every applicable control from the three named
  standards and linking each to an executable test or an approved waiver.
- Commission and complete an independent external assessment against that control map.
- Resolve or waive every Critical/High finding the assessment surfaces.

## Out of scope

- **SEC-01/03/04/06's own implementation** — already built and accepted; this story maps and assesses
  their existing state, it does not re-implement or extend any of them.
- **Remediating a genuinely new architectural gap the assessment surfaces beyond what an approved waiver
  can cover** — if the assessment finds something requiring substantial new implementation work, that
  work is scoped as its own follow-up item (per mandate §12's own task-boundary discipline), not silently
  absorbed into this story's own bounded scope.

## Assumptions

- The exact set of "applicable" controls from ASVS 5.0.0/OWASP API Security Top 10 2023/NIST 800-63-4
  is not pre-enumerated by any source document available to this planning generation — this story's own
  control-map-building work performs that enumeration, informed by the framework's own actual
  capabilities (an API-only framework with no direct end-user browser surface, for instance, may find
  some ASVS controls genuinely not applicable).
- The external assessment is assumed to be a professional-services engagement performed by a party
  outside this programme's own execution — this story's own tasks build the control map the assessor
  consumes and record the assessment's own outcome, they do not themselves constitute the assessment.

## Dependencies

**Hard dependency on W03-E01 (SEC-01), W03-E02 (SEC-06), W03-E03 (SEC-03), and W05-E04 (SEC-04) all
being `accepted`** — PLAN's own explicit framing: "SEC-01–04 substantially complete." Execution-time
verification found this dependency unresolved; see EV-W07-E02-S001-003 and
DEV-W07-E02-S001-002. No dependency within W07-E02 or on another W07 epic.

## Affected packages or components

No production code package is modified by this story — its own artifact is the control map document and
the external assessment's own report. Existing test files across `kernel/auth`, `kernel/authz`,
`kernel/webhook`, and related packages may be referenced (and, where the control map identifies a gap, a
new executable test may be added — though this is scoped as a small, bounded addition per finding, not a
broad new test-suite build).

## Compatibility considerations

Not applicable — this story produces a document and an assessment report, not a code change (beyond any
small, bounded test additions the control map itself identifies as missing).

## Security considerations

This entire story IS a security-verification exercise; see "Objective" and "Value to the framework"
above.

## Performance considerations

Not applicable.

## Observability considerations

Not applicable.

## Migration considerations

Not applicable.

## Documentation requirements

The control map itself (`SEC-05/control-map.md` or equivalent) is this story's own primary
documentation output; the external assessment's own report is a second required documentation artifact.

## Acceptance criteria

- **AC-W07-E02-S001-01**: The version-pinned control map links every applicable ASVS 5.0.0/OWASP API Security
  Top 10 2023/NIST 800-63-4 control to an executable test or an approved, documented waiver.
- **AC-W07-E02-S001-02**: The independent external assessment leaves zero open Critical/High findings, or each
  open finding carries an approved, time-bounded waiver.

## Required artifacts

- `SEC-05/control-map.md` (or equivalent) — the version-pinned control map.
- The external assessment's own report.
See `artifacts/index.md`.

## Required evidence

- The control map itself, inspected for completeness against the three named standards.
- The external assessment's own report, inspected for zero open Critical/High findings or approved
  waivers.
See `evidence/index.md`.

## Definition of ready

Rechecked against `governance/definition-of-ready.md` on 2026-07-14: `story.md` and `plan.md` are
complete, both acceptance criteria are numbered and measurable, and owner/reviewer are assigned.
The SEC-01/03/04/06 accepted-state hard dependency fails 7/7 lifecycle pairs, so the story remains blocked.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance criteria
verified with evidence in `evidence/index.md`; `closure.md` completed; independent review passed per
mandate §14, specifically confirming the external assessment was genuinely performed by an independent
party (not an internal self-assessment presented as external) and that every waived finding carries a
genuine owner/rationale/expiry, not a placeholder.

## Risks

RISK-W07-002 (the external assessment surfacing an open Critical/High finding with no immediate
remediation path) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once the external assessment is complete and every finding is either resolved or genuinely waived,
residual risk is expected to be low — SEC-01-04's own prior, independently-reviewed acceptance across
earlier waves substantially reduces the likelihood of a major new gap surfacing here.

## Plan

See `plan.md`.
