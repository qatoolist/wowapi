---
id: W06-E03-S002-T001
type: task
title: Repo-admin activation (branch protection, release Environment, tag protection ruleset)
status: todo
parent_story: W06-E03-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W06-E03-S002-01
artifacts:
  - ART-W06-E03-S002-001
  - ART-W06-E03-S002-002
  - ART-W06-E03-S002-003
evidence:
  - EV-W06-E03-S002-001
---

# W06-E03-S002-T001 — Repo-admin activation (branch protection, release Environment, tag protection ruleset)

## Task Definition

### Task objective

Configure branch protection on main, create the protected release GitHub Environment with required reviewers, and configure a tag protection ruleset. THIS TASK IS HUMAN-ONLY and cannot be performed by a coding agent.

### Parent story

W06-E03-S002

### Owner

unassigned

### Status

todo

### Dependencies

DEC-Q10 must be resolved — a human with repo-admin access must commit to and perform this activation. This task cannot begin, be simulated, or be worked around by any coding agent.

### Detailed work

1. [Human, repo-admin] Re-confirm the current absence of branch protection, the release environment,
   and tag protection via the same `gh api` calls the source review used.
2. [Human, repo-admin] Configure branch protection on `main`, referencing W06-E03-S001's own gate
   manifest's required status checks where applicable.
3. [Human, repo-admin] Create the protected `release` GitHub Environment with required reviewers.
4. [Human, repo-admin] Configure a tag protection ruleset for release tags, working within this
   repository's own platform constraints (REVIEW's own session-fact note: "Merge-queue rulesets
   unavailable on user-owned repo").

### Expected files or components affected

None — this task changes GitHub repository settings, not repository files.

### Expected output

Active branch protection on main, an active protected release Environment with required reviewers, and an active tag protection ruleset.

### Required artifacts

ART-W06-E03-S002-001 (branch-protection configuration record), ART-W06-E03-S002-002 (release environment configuration record), ART-W06-E03-S002-003 (tag-protection ruleset configuration record).

### Required evidence

EV-W06-E03-S002-001 (live gh api call output confirming all three controls are active).

### Related acceptance criteria

AC-W06-E03-S002-01.

### Completion criteria

All three controls are confirmed active via live gh api calls, not a console screenshot or an unverified claim.

### Verification method

Live gh api call re-run, mirroring the exact calls used to confirm the controls' prior absence.

### Risks

RISK-W06-001 (this task cannot begin until DEC-Q10 resolves) — see epic-level `risks.md`.

### Rollback or recovery considerations

Not applicable in the code-rollback sense — if a configured rule proves too restrictive, the repo administrator adjusts it directly.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented.*

### Observability changes

*Not yet implemented.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented. Once implementation occurs, record whether it matched `plan.md`; if not,
reference the corresponding entry in `deviations.md`.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E03-S002-01 | Live gh api call re-run | Live GitHub API, post-activation | All three controls confirmed active | live API call output | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
