---
id: IMPL-W07-E02-S001
type: implementation-record
parent_story: W07-E02-S001
status: blocked
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Implementation record — W07-E02-S001

## What was actually implemented

T001 is implemented: a canonical machine-readable map inventories all 412 pinned source units (345
ASVS 5.0.0 requirements, ten OWASP API Security Top 10 2023 categories, and 57 normative units from
the final NIST SP 800-63-4 main publication). All 33 applicable entries resolve to executable focused
tests; all 379 non-applicable entries carry scope rationales; no waiver is asserted.

The map validator rejects catalog omissions, unknown/duplicate controls, version mismatch, dangling
tests, rationale-free N/A entries, and any waiver without genuine approval/owner/rationale/expiry.
The accepted-state prerequisite checker records the actual upstream lifecycle inconsistency.

T002 is not implemented: no external professional-services assessor or engagement/report was
available. The exact blocker is recorded without substituting the internal story review.

## Components changed

- New repository-root `SEC-05/` verification profile and pinned source inventories.
- W07-E02-S001 lifecycle, artifact, evidence, verification, and blocker records.
- No production package, API, configuration, schema, migration, or wowsociety repository change.

## Files changed

- `SEC-05/control-map.{json,md}`
- `SEC-05/sources/*`
- `SEC-05/validate_control_map.py`
- `SEC-05/test_validate_control_map.py`
- `SEC-05/verify_prerequisites.py`
- `SEC-05/external-assessment-status.md`
- W07-E02-S001 story package and registered evidence under `evidence/security/` and
  `evidence/reviews/`.

## Interfaces introduced or changed

No production interface. The verification-only CLI is
`python3 SEC-05/validate_control_map.py [--run-tests]`.

## Configuration changes

None.

## Schema or migration changes

None.

## Security changes

No security behavior changed. Existing SEC behavior was mapped and focused tests re-executed.

## Observability changes

None.

## Tests added or modified

Added six standard-library Python regression tests for the machine-check validator. No existing test
or production file was modified.

## Commits

Execution observed at base commit `733ef3e930cbb3f89f5bbc53d8f562c60e426513` in a shared dirty
workspace. A clean integration commit is not yet available; evidence explicitly requires a retest when
it exists.

## Pull requests

None.

## Implementation dates

2026-07-14.

## Technical debt introduced

None in production code. The control map still requires the mandated external assessor's independent
applicability confirmation.

## Known limitations

- AC-W07-E02-S001-02 is blocked because no external engagement/report exists.
- The hard dependency on consistently accepted SEC-01/03/04/06 lifecycle records fails 7/7 checks.
- Passing map evidence is provisional until repeated against the clean integration commit.

## Follow-up items

1. Upstream owners reconcile and accept the seven inconsistent SEC story/closure lifecycle pairs.
2. Product-security lead commissions the external assessor and registers the genuine report.
3. Resolve or genuinely waive every Critical/High finding, if any.
4. Re-run map validation and focused tests against the clean integration commit.

## Relationship to the approved plan

Plan steps 1–3 (enumeration, executable mapping, control-map assembly) were implemented. Steps 4–5
(commission assessment and dispose findings) are blocked by an unavailable human/vendor engagement.
This divergence is recorded as DEV-W07-E02-S001-001 in `deviations.md`.
