---
id: PLAN-W07-E02-S001
type: plan
parent_story: W07-E02-S001
status: blocked
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Plan — W07-E02-S001

Per mandate §8.5. This story's own scope is explicitly bounded to building the control map and
commissioning/recording the external assessment — it is not a general security-remediation story.
Confirmed facts, planned changes, and assumptions are distinguished explicitly below.

> **Execution note (2026-07-14):** The approved plan is preserved, but both approval assumptions were
> false at execution: upstream lifecycle records are not consistently accepted and no external
> assessor/vendor was available. DEV-W07-E02-S001-001/002 record the deviations; the story is blocked.

## Proposed architecture

A single control-map document, `SEC-05/control-map.md` (or equivalent path), enumerating every
applicable control from ASVS 5.0.0, OWASP API Security Top 10 2023, and NIST 800-63-4, each linked to an
executable test (existing or newly added, in a small, bounded way) or an approved waiver. An external
assessment report evaluating the framework's actual implementation against that control map.

## Implementation strategy

1. Enumerate every applicable control from the three named standards, informed by the framework's own
   actual capability surface (an API-only framework, no direct browser UI).
2. For each control, identify an existing executable test that proves it, or add a small, bounded new
   test if a genuine gap exists, or record an approved waiver with owner/rationale/expiry.
3. Assemble the control map document.
4. Commission the independent external assessment.
5. Record the assessment's own findings; resolve or waive each Critical/High finding.

## Expected package or module changes

No production code package change expected beyond small, bounded test additions per any genuine gap the
control-mapping exercise itself identifies (exact scope, if any, determined during step 2 above).

## Expected file changes where determinable

- `SEC-05/control-map.md` (new).
- The external assessment's own report (new, exact location/format TBD).
- Possibly small, bounded new test files, one per genuinely-identified gap (exact scope TBD).

## Contracts and interfaces

None new.

## Data structures

None new.

## APIs

None affected.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

Not applicable.

## Security controls

This entire story is itself a security-verification exercise; see `story.md` "Security considerations."

## Observability changes

None.

## Testing strategy

- The control map's own completeness is verified by direct inspection against the three named
  standards' own control lists.
- The external assessment's own findings are verified by direct inspection of its report for zero open
  Critical/High findings or approved waivers.

## Regression strategy

The control map, once established, becomes the ongoing reference for future security work to check
against — though maintaining it current against future framework changes is not itself mandated as a
recurring task by this story's own scope (a future finding, if any, would establish that ongoing
maintenance obligation).

## Compatibility strategy

Not applicable.

## Rollout strategy

Single story, landed as its own reviewable unit — though the external assessment's own timeline is
outside this programme's own direct control (a professional-services engagement).

## Rollback strategy

Not applicable — this story produces documentation and an assessment record, not a reversible code
change (beyond any small test additions, which follow standard rollback discipline if found incorrect).

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–5).

## Task breakdown

- **W07-E02-S001-T001** — Build the version-pinned control map.
- **W07-E02-S001-T002** — Commission and record the external assessment; resolve or waive findings.

## Expected artifacts

`SEC-05/control-map.md`; the external assessment's own report.

## Expected evidence

The control map itself; the external assessment's own report.

## Unresolved questions

- The exact scope of any small, bounded test additions the control-mapping exercise identifies as
  genuinely missing — not knowable until the enumeration (step 1) is actually performed.
- The external assessment's own exact timeline — outside this programme's own direct control.

## Approval conditions

This plan is approved for implementation once: (a) SEC-01/03/04/06 are confirmed `accepted` (already
satisfied by this wave's own entry gate), and (b) the owner and reviewer are assigned.
