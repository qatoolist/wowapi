---
id: DEV-W07-E02-S001
type: deviations-record
parent_story: W07-E02-S001
status: blocked
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Deviations record — W07-E02-S001

## DEV-W07-E02-S001-001 — External assessment unavailable

- **Approved plan:** Steps 4–5 commission an independent professional-services assessment and dispose
  every Critical/High finding.
- **Actual implementation:** No assessor/vendor, engagement owner/identifier, report, findings
  register, or approved waiver was supplied or discoverable. A status record was produced, explicitly
  not an assessment report.
- **Reason:** The engagement requires a human product-security lead and an external professional
  party. A coding agent cannot perform or approve it.
- **Impact:** T002 and AC-W07-E02-S001-02 remain blocked; the story cannot be verified or accepted.
- **Risks:** External review could find a Critical/High issue not visible to internal verification.
- **Approval:** None. No exception or waiver is asserted.
- **Compensating controls:** Machine-checked control map and independent story-artifact review reduce
  preparation/review risk but explicitly do not substitute for the external assessment.
- **Follow-up work:** Product-security lead commissions the engagement, registers the genuine report,
  and owns finding disposition.

## DEV-W07-E02-S001-002 — Assumed accepted precondition is not recorded consistently

- **Approved plan:** SEC-01/03/04/06 were assumed accepted by the wave entry gate.
- **Actual implementation:** `python3 SEC-05/verify_prerequisites.py` found all seven checked
  `story.md`/`closure.md` pairs inconsistent and none consistently `accepted`.
- **Reason:** Upstream lifecycle records diverge from the planning-time assumption.
- **Impact:** The story's explicit hard dependency is unsatisfied even though focused tests can
  exercise the implementation currently in the working tree.
- **Risks:** Treating the implementation as accepted would bypass missing upstream independent review
  and closure records.
- **Approval:** None; the deviation is a blocker, not an accepted exception.
- **Compensating controls:** Preserve the failed prerequisite evidence and do not modify upstream
  lifecycle records from this story.
- **Follow-up work:** Upstream owners reconcile each story through the governed lifecycle, then re-run
  the prerequisite checker.
