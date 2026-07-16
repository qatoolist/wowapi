---
id: W03-E01-S003-T003
type: task
title: Independent review
status: done
parent_story: W03-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E01-S003-T001
  - W03-E01-S003-T002
acceptance_criteria:
  - AC-W03-E01-S003-01
  - AC-W03-E01-S003-02
artifacts: []
evidence:
  - EV-W03-E01-S003-003
---

# W03-E01-S003-T003 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story per mandate §14, specifically confirming: the
"expired step-up" required test class (PLAN §6 SEC-05) is genuinely exercised; the credential-scheme
distinction mechanism correctly rejects a mismatched scheme without over- or under-restricting
legitimate combinations; and — specifically for this story — that the DX-03 cross-cut coordination
note (`plan.md`'s "Unresolved questions") was recorded explicitly rather than silently resolved or
silently ignored.

### Parent story

W03-E01-S003 — Assurance freshness and credential-scheme distinction.

### Owner

unassigned

### Status

done

### Dependencies

W03-E01-S003-T001, W03-E01-S003-T002.

### Detailed work

1. Confirm implementation matches `plan.md`, or that every divergence is recorded in
   `deviations.md`.
2. Confirm both acceptance criteria are each backed by a passing test with logged evidence in
   `evidence/index.md`, referencing the correct commit SHA.
3. Confirm the "expired step-up" required test class is genuinely exercised (stale `auth_time` +
   valid `amr` → step-up fails), not merely asserted in prose.
4. Confirm the credential-scheme distinction test genuinely proves a `CredentialUser`-scoped
   permission rejects a valid API-key actor, and spot-check that a correctly-scoped user-credential
   request is not incorrectly rejected by the same mechanism (a false-positive-rejection check).
5. **Confirm the DX-03 cross-cut coordination note was recorded, not silently resolved.** Verify
   that this story's `CredentialScheme` mechanism is documented as a candidate for reconciliation
   with DX-03 (W06-E01-S001), per `plan.md`'s "Unresolved questions," and that no part of this
   story's implementation or documentation claims to have made DX-03's eventual design decision.
6. Record findings; resolve or explicitly accept before this story can move to `accepted`.

### Expected files or components affected

None (review-only task).

### Expected output

A completed review report confirming the checklist above, recorded as evidence.

### Required artifacts

None (review report is evidence, not an artifact per mandate §9's artifact/evidence distinction).

### Required evidence

EV-W03-E01-S003-003 (review report).

### Related acceptance criteria

AC-W03-E01-S003-01, AC-W03-E01-S003-02.

### Completion criteria

Review report completed with no open finding, or every finding resolved or explicitly accepted per
mandate §14.

### Verification method

Manual independent review against the checklist in "Detailed work," conducted by a reviewer who did
not implement T001/T002.

### Risks

None beyond the story's own inherited scope — this task's own risk is limited to the DX-03
cross-cut note being overlooked during review; mitigated by making it an explicit checklist item
(step 5 above).

### Rollback or recovery considerations

Not applicable — a review-only task has no code to roll back.

## Implementation Record

### What was actually implemented

This entry previously recorded review prose without a filled Verification Record, reviewer
identity, execution date, or `status: done` — i.e. drafted but never formally executed or closed
out (matching the pattern flagged elsewhere in this wave). A genuine independent review is now
completed and recorded below, superseding the prior draft prose. Checklist:

1. Implementation matches `plan.md`; no deviations found in `deviations.md`.
2. Both acceptance criteria are backed by tests in `kernel/authz/assurance_freshness_test.go` and
   `kernel/authz/credential_scheme_test.go` — re-run below with a live DB (the prior draft review
   explicitly noted it relied on non-DB tests only; this review closes that gap).
3. The "expired step-up" required test class is exercised: stale `auth_time` + valid `amr` →
   `step_up_freshness_required` (`TestStepUpFreshnessStaleAuthTimeFails`).
4. The credential-scheme test proves a `CredentialUser`-scoped permission rejects a valid API-key
   actor (`TestCredentialSchemeUserPermissionRejectsAPIKey`), and the positive-path test
   (`TestCredentialSchemeUserPermissionAllowsUser`) confirms correctly-scoped actors are not
   falsely rejected.
5. The DX-03 cross-cut coordination note is recorded in `plan.md`, `story.md`, and the story's
   "Accepted risks" — not silently resolved.

### Components changed

None (review-only).

### Files changed

None (review-only).

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None directly.

### Observability changes

None.

### Tests added or modified

None (reviewed existing tests).

### Commits

None.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

DB-backed tests are skipped without `DATABASE_URL`. Review relied on package builds and passing
non-DB tests.

### Follow-up items

- Re-run DB-backed tests with `DATABASE_URL` set.

### Relationship to the approved plan

Review confirms implementation matches `plan.md` with no deviations.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S003-01, -02 | Independent review checklist per mandate §14 | N/A (documentation/code review) | No open finding, or every finding resolved/accepted | review report | unassigned |

### Actual result

`go test ./kernel/authz/... -run 'TestStepUpFreshness|TestCredentialScheme' -count=1 -v` and
`go test ./kernel/auth/... -run 'TestActorInternal_AssuranceFieldsPropagate' -count=1 -v` (DB up):
all 11 named tests PASS. This closes the prior draft review's own noted limitation ("DB-backed
tests are skipped without `DATABASE_URL`").

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E01-S003-003 (this review report, superseding the prior unexecuted draft).

### Execution date

2026-07-16.

### Commit or revision

HEAD `43b6e12` + remediation working tree 2026-07-16.

### Environment

Local dev; DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable;
WOWAPI_REQUIRE_DB=1; Go per repo `go.mod`.

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3). This reviewer did not implement T001/T002.

### Findings

**Finding (Low, carried and now confirmed)**: `../closure.md`'s "Evidence completeness" section
cites `tmp/s003_smoke.go` as evidence backing EV-W03-E01-S003-001/002. Confirmed by search
(`find . -name s003_smoke.go`) that this file does not exist anywhere in the repository — a
referenced-but-missing evidence artifact, a violation of `governance/evidence-policy.md`'s
requirement that cited evidence be real/reproducible. This does not affect the AC verdict (the
tests actually cited by `evidence/index.md`, which does NOT reference `tmp/s003_smoke.go`, are real
and pass), but `closure.md` must be corrected to drop or replace that citation before this story is
treated as fully closed. Separately, `closure.md`'s frontmatter `status: draft` contradicts its own
prose `## Final status: accepted (pending formal product-security lead sign-off and DB-backed test
re-run)` — an oxymoron per `governance/status-model.md`. This review's DB-backed re-run
requirement is now satisfied (see "Actual result" above); the sign-off and status-field
contradiction remain for the conductor/product-security lead to resolve.

### Retest status

Retested against current HEAD + working tree with a live DB (2026-07-16), closing the prior
review's own noted DB-test gap.

### Final conclusion

Acceptance criteria AC-W03-E01-S003-01 and -02 satisfied by genuine, DB-backed test evidence. One
Low-severity documentation defect remains open in `../closure.md` (missing evidence file citation +
contradictory status field) — recommend `accept-with-conditions`: fix `closure.md` before the
story is marked `accepted` outright.

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
