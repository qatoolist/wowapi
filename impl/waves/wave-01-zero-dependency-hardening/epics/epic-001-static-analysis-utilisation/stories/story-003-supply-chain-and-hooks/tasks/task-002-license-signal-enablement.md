---
id: W01-E01-S003-T002
type: task
title: License-scanning signal enablement
status: done
parent_story: W01-E01-S003
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S003-02
artifacts:
  - ART-W01-E01-S003-002
evidence:
  - EV-W01-E01-S003-002
---

# W01-E01-S003-T002 — License-scanning signal enablement

## Task Definition

### Task objective

Enable a license-scanning signal in CI — per `plan.md`'s "License signal decision," the planned choice
is adding `license` to the existing `trivy` job's `scanners:` list in
`.github/workflows/security-scan.yml`, since Trivy's `vuln,secret,misconfig` scanners are already
enabled there and `license` is a native addition to that same list, with `go-licenses` retained as a
documented fallback if the fresh re-read at this task's start finds Trivy's approach inadequate.

### Parent story

W01-E01-S003 — Close supply-chain and pre-push hook hygiene gaps.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001/T003/T004 (disjoint files: this task touches `security-scan.yml` only).

### Detailed work

1. Re-read `.github/workflows/security-scan.yml` fresh, at this task's actual start commit, to confirm
   the `trivy` job's current `scanners:` list and the `dependency-review` job's `license-check: true`
   gating (per `story.md`'s "Current-state assessment," `license-check: true` was observed around line
   93, gated to `pull_request` events, as of 2026-07-12 — re-confirm, do not assume unchanged).
2. Confirm the license-signal choice from `plan.md` still holds: Trivy's `license` scanner as the
   planned default. If the fresh re-read finds a materially different situation (e.g. the `trivy` job
   has been restructured, or Trivy's license-detection coverage for this repository's actual dependency
   mix proves inadequate), record the revised choice as a deviation in `deviations.md` per mandate
   §2.6 rather than silently picking `go-licenses` without documenting why.
3. Add `license` to the `trivy` job's `scanners:` list (or add a new `go-licenses` step, if the choice
   is revised).
4. Determine and explicitly configure whether the license-scan output should fail the build on a real
   violation or remain informational (matching the existing `trivy` job's `exit-code: "0"` posture for
   its other scanners) — per `plan.md`'s "Unresolved questions," this must be an explicit, documented
   choice, not left ambiguous.
5. Run the updated workflow to confirm the license-scan step executes and produces a license report or
   equivalent output.
6. Triage any real license finding the scan surfaces (a genuine violation, if one exists in the current
   dependency set, is a real finding to record and address or explicitly accept — not to be suppressed
   merely to make the task pass).

### Expected files or components affected

`.github/workflows/security-scan.yml`.

### Expected output

An updated `security-scan.yml` with a license-scanning signal enabled and running; the choice
(Trivy license scanner vs. `go-licenses`) and its rationale recorded in `implementation.md`.

### Required artifacts

ART-W01-E01-S003-002 (updated `security-scan.yml`).

### Required evidence

EV-W01-E01-S003-002 (security-scan report).

### Related acceptance criteria

AC-W01-E01-S003-02.

### Completion criteria

The license-scanning step runs in CI and produces output; the choice and rationale are documented; any
real license finding is triaged, not silently suppressed.

### Verification method

Direct CI execution of the updated `security-scan.yml`, logged output retained as evidence per
`evidence/index.md`.

### Risks

Low-to-moderate — the primary risk is that enabling license scanning surfaces a real finding (an
unexpected or incompatible dependency license) that requires triage effort not fully scoped by this
task's plan; per "Detailed work" step 6, any such finding is recorded and addressed or explicitly
accepted, not silently dropped.

### Rollback or recovery considerations

Revert the `scanners:` list change (or the `go-licenses` step, if chosen instead) if it produces an
unexpected volume of false positives or breaks the existing `trivy` job in an unrelated way; escalate a
genuine license-compliance finding rather than reverting to suppress it.

## Implementation Record

Implemented 2026-07-13 by W01Lint.

### What was actually implemented

Carried through the planned choice unchanged: added `license` to the `trivy` job's `scanners:` list
in `security-scan.yml` (step renamed to "trivy filesystem + config + license scan"; comment
documents the signal-not-gate posture). Validated the choice before committing to it: a local Trivy
license scan against a pristine HEAD copy enumerated all 70 Go dependency licenses (all LOW:
MIT/Apache-2.0/BSD-*), disproving the "gomod not scanned" first impression (the `gomod` analyzer row
shows `-`; the license findings attach to the `go.mod` target's license row). With the job's
existing `severity: CRITICAL,HIGH` filter, the job log reports exactly forbidden/restricted licenses
(currently zero). `go-licenses` fallback not needed.

### Files changed

`.github/workflows/security-scan.yml` (+7/−2).

### Commits

Conductor owns commits; delivered as a working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`.

### Pull requests

None (conductor owns wave integration).

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

## Verification Record

| AC | Actual result | Pass/fail | Evidence |
|---|---|---|---|
| AC-W01-E01-S003-02 | Local `trivy fs --scanners license .` (pristine HEAD): 70 dependency licenses enumerated, 0 CRITICAL/HIGH; choice + rationale recorded in story `implementation.md`; `actionlint` clean | pass (in-CI run log pending conductor push) | EV-W01-E01-S003-002 (`evidence/logs/trivy-license-local-report.txt`) |

### Retest status

Wave gate / next weekly security-scan run to be registered as `retested` evidence.

### Final conclusion

AC satisfied; license signal live in workflow config and proven non-hollow locally.

## Deviations Record

None — planned choice (Trivy) carried through unchanged.
