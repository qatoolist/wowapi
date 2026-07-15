---
id: W07-E02-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W07-E02-S001
status: blocked
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E02-S001 — Evidence index

Per mandate §10. Each linked record contains the full required environment, tool-version, timestamp,
file/checksum, reviewer, and supersession fields. Failed/blocker evidence is preserved rather than
overwritten by a completion claim.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status | Record |
|---|---|---|---|---|---|---|---|---|
| EV-W07-E02-S001-001 | control-map completeness + focused executable tests | W07-E02-S001-T001 | AC-W07-E02-S001-01 (functional result; final clean-commit retest still required) | `python3 SEC-05/validate_control_map.py`; validator unit tests; `... validate_control_map.py --run-tests` with required DB/S3 env | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + artifact checksums | PASS: 412 mapped with pinned inventory digests; 33 applicable tested; five Go packages + six validator tests pass | accepted with explicit shared-working-tree revision caveat | [record](security/EV-W07-E02-S001-001-control-map-validation.md) |
| EV-W07-E02-S001-002 | external professional-services assessment status | W07-E02-S001-T002 | None; AC-W07-E02-S001-02 remains unverified | Not applicable — requires actual external engagement | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | BLOCKED/FAIL: no assessor, engagement, report, findings, or approved waivers | failed | [record](security/EV-W07-E02-S001-002-external-assessment-status.md) |
| EV-W07-E02-S001-003 | SEC accepted-state prerequisite check | W07-E02-S001-T001 | None; hard dependency remains unsatisfied | `python3 SEC-05/verify_prerequisites.py` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | FAIL: 0/7 story/closure pairs consistently accepted | failed | [record](security/EV-W07-E02-S001-003-sec-accepted-state.md) |
| EV-W07-E02-S001-004 | independent story-artifact review | story gate | mandate §14 review only; never substitutes for AC-02 external assessment | reviewer inspection + focused re-execution | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + artifact checksums | PASS: no open actionable story-scope issue; external assessment remains blocker | accepted | [record](security/EV-W07-E02-S001-004-independent-review.md) |
