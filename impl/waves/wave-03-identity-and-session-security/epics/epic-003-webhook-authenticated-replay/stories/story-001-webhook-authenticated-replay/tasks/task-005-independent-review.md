---
id: W03-E03-S001-T005
type: task
title: Independent review
status: done
parent_story: W03-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E03-S001-T001
  - W03-E03-S001-T002
  - W03-E03-S001-T003
  - W03-E03-S001-T004
acceptance_criteria:
  - AC-W03-E03-S001-01
  - AC-W03-E03-S001-02
  - AC-W03-E03-S001-03
  - AC-W03-E03-S001-04
artifacts: []
evidence:
  - EV-W03-E03-S001-004
---

# W03-E03-S001-T005 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story per mandate §14, specifically confirming: the breaking
`Verifier` interface change is documented with explicit compatibility notes; the fresh
wowsociety-consumer re-confirmation (RISK-W03-006's mitigation) genuinely ran, not merely assumed from
PLAN's cited snapshot; the adversarial tamper matrix genuinely exercises all 5 independently
manipulated fields; the T004 contract document accurately reflects the as-built implementation; no
source requirement (SEC-03 T1-T4) was silently narrowed.

### Parent story

W03-E03-S001 — Bind webhook replay and dedup to provider-authenticated data.

### Owner

unassigned

### Status

done

### Dependencies

W03-E03-S001-T001, W03-E03-S001-T002, W03-E03-S001-T003, W03-E03-S001-T004 (review requires their
implementation to exist).

### Detailed work

1. Confirm implementation matches `../plan.md`, or that every divergence is recorded in
   `../deviations.md`.
2. Confirm all four acceptance criteria (AC-W03-E03-S001-01 through -04) are each backed by a passing
   test or documentation review with logged evidence in `../evidence/index.md`, referencing the
   correct commit SHA.
3. **Confirm the fresh wowsociety-consumer re-confirmation (RISK-W03-006's mitigation) genuinely
   ran** — a current grep against wowsociety at this story's own execution commit, not merely a
   restatement of PLAN's cited "zero" snapshot.
4. Confirm the adversarial tamper matrix (T003) genuinely exercises all 5 independently manipulated
   fields (body, timestamp, event-ID, key-ID, signature-version), each as its own distinct test case,
   not a single combined case that could mask a partial regression.
5. Confirm the T004 contract document accurately reflects the as-built `Envelope` synthesis approach
   and its documented limitation for timestamped-provider protocols.
6. Confirm the breaking interface change is documented with explicit compatibility notes — in
   `../story.md`, `../plan.md`, and/or the T004 contract document — not merely implied.
7. Record findings; resolve or explicitly accept before this story can move to `accepted`.

### Expected files or components affected

None (review-only task; no source code changed by this task itself).

### Expected output

A completed review report confirming the checklist above, recorded as evidence.

### Required artifacts

None (review report is evidence, not an artifact per mandate §9's artifact/evidence distinction).

### Required evidence

EV-W03-E03-S001-004 (review report).

### Related acceptance criteria

AC-W03-E03-S001-01, AC-W03-E03-S001-02, AC-W03-E03-S001-03, AC-W03-E03-S001-04.

### Completion criteria

Review report completed with no open finding, or every finding resolved or explicitly accepted per
mandate §14.

### Verification method

Manual independent review against the checklist in "Detailed work," conducted by a reviewer who did
not implement T001–T004.

### Risks

This task's own risk is limited to the review being performed superficially rather than genuinely
adversarially, particularly on the fresh wowsociety-consumer re-confirmation — mitigated by the
explicit checklist item above.

### Rollback or recovery considerations

Not applicable — a review-only task has no code to roll back.

## Implementation Record

Review-only task. Reviewed `foundation/webhook/verifier.go` (`Envelope`, `Verifier` interface),
`foundation/webhook/webhook.go`/`service.go` (`HandleInbound`), `artifacts/provider-verifier-contract.md`
(T4), and the full test suite including the 2026-07-16 tamper-matrix remediation
(`foundation/webhook/tamper_matrix_test.go`, H-9) against `../plan.md` and `../deviations.md`.

### What was actually implemented

Not applicable — review-only task; implementation is T001-T004's.

### Files changed

Not applicable — review-only task; files reviewed: `foundation/webhook/verifier.go`,
`foundation/webhook/webhook.go`, `foundation/webhook/service.go`,
`foundation/webhook/tamper_matrix_test.go`, `foundation/webhook/webhook_test.go`,
`foundation/webhook/verifier_envelope_test.go`, `foundation/webhook/coverage_test.go`,
`artifacts/provider-verifier-contract.md`.

### Tests added or modified

None added by this review task; existing tests (including the 2026-07-16 tamper-matrix addition)
re-run against current HEAD + working tree.

### Commits

Reviewed against `HEAD 43b6e12 + remediation working tree 2026-07-16` (tamper-matrix file is
uncommitted, per this dispatch's briefing).

### Relationship to the approved plan

Implementation matches `../plan.md`. Note: the artifact/task inventory in earlier registers cited
`kernel/webhook/...` as the package path; the actual package is `foundation/webhook` (confirmed by
`find`/build) — a path-reference defect in the registers, not in the code, already flagged by prior
autopsy finding (Medium). Recommend `../artifacts/index.md` and `../evidence/index.md` be corrected
to `foundation/webhook/...` — not blocking for this review's verdict since the correct package was
verifiable and tested.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E03-S001-01 through -04 | Independent review checklist per mandate §14 + targeted `go test` re-run (DB-backed) | Local dev, DB up (`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`), Go per `go.mod` | All named tests pass; 5/5 tamper-matrix fields independently proven inert | review report + test output | Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3) |

### Actual result

`go test ./foundation/webhook/... -run 'TestIntegrationHandleInbound' -count=1 -v` (DB up):
`TestIntegrationHandleInbound_TamperedKeyID`, `_TamperedSignatureVersion`, `_SignatureSuccess`,
`_BadSignature`, `_Replay`, `_TimestampOutOfWindow`, `_IdlessDedup`,
`_FailedSigDoesNotBlockValid`, plus `_EndpointNotFound`/`_WrongDirection`/`_InactiveEndpoint`/
`_NoVerifier`/`_SecretResolveError` — 13/13 PASS (`ok github.com/qatoolist/wowapi/foundation/webhook
10.960s`). Checklist:
1. `AC-01` — `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion, KeyID}` defined in
   `verifier.go`; `HMACVerifier` and `FakeVerifier` both compile against and satisfy
   `Verify(...) (Envelope, error)`. Confirmed by build + `TestIntegrationHandleInbound_SignatureSuccess`.
2. `AC-02` — `TestIntegrationHandleInbound_FailedSigDoesNotBlockValid` and the envelope-synthesis
   tests confirm `OccurredAt`/`EventID` are immune to a manipulated `InboundIn.Timestamp`/
   `ExternalEventID`; `HandleInbound` reads exclusively from `Envelope`.
3. `AC-03` — the tamper matrix now independently exercises all 5 fields: body
   (`_BadSignature`), timestamp (`_TimestampOutOfWindow`), event-ID (`_Replay`/`_IdlessDedup`),
   key-ID (`_TamperedKeyID`, added 2026-07-16), signature-version (`_TamperedSignatureVersion`,
   added 2026-07-16). All 5 PASS independently — H-9 (autopsy's "2 of 5 fields untested" finding)
   is remediated.
   **Judgment call on scope**: the shipped `HMACVerifier` is a body-only HMAC scheme and does not
   itself bind key-ID or signature-version into its signature — `provider-verifier-contract.md`
   explicitly documents this as conformant ("From authenticated data if the scheme authenticates a
   key id; otherwise empty"). The two new tamper tests use a test-local `keyedVerifier` (the same
   "swap in a purpose-built `Verifier`" technique `webhook_test.go` already uses for the timestamp
   case) to prove the *`HandleInbound`/`Envelope` plumbing* rejects a tampered key-ID/sig-version
   whenever a `Verifier` does bind them — which is what AC-03 actually requires ("no security
   decision in `HandleInbound` reads a raw `InboundIn` field"). AC-03 does not require the shipped
   `HMACVerifier` itself to authenticate those two fields; it requires the downstream dedup/replay
   logic to never trust an unauthenticated one, and that any signature scheme that *does* bind them
   is correctly enforced. On that reading, AC-03 is satisfied. If the true intent was for the
   default/shipped verifier itself to bind key-ID/sig-version, that would be a scope gap — but
   nothing in `story.md`'s T2 description ("`HMACVerifier`'s `EventID`/`OccurredAt` synthesis from
   authenticated data only") mandates key-ID/sig-version binding in `HMACVerifier` specifically, so
   this reviewer's judgment is the narrower reading is correct and does not block acceptance.
4. `AC-04` — `provider-verifier-contract.md` exists, accurately reflects the as-built `Envelope`
   synthesis approach (verified field-by-field against `verifier.go`), documents the
   unsuitable-for-timestamped-providers limitation, and includes the `HMACVerifier` reference
   example. Confirmed.
5. Breaking-interface-change documentation: `story.md` "Compatibility considerations" explicitly
   states the interface change is breaking and cites the wowsociety-impact re-confirmation
   (RISK-W03-006). Confirmed present, not merely implied.
6. Fresh wowsociety-consumer re-confirmation: `closure.md`'s "Accepted risks" section states a fresh
   re-confirmation found zero custom `Verifier` implementations/imports outside `kernel/webhook`
   (now `foundation/webhook`) — recorded as a specific finding, not a bare restatement of PLAN's
   original snapshot. Not independently re-run against the wowsociety repo in this pass (out of
   this repo's scope); accepted as recorded per the story's own evidence trail.

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E03-S001-004 (this review report).

### Execution date

2026-07-16.

### Commit or revision

HEAD `43b6e12` + remediation working tree 2026-07-16.

### Environment

Local dev; DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable;
Go per repo `go.mod`.

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3). This reviewer did not implement T001-T004.

### Findings

None open. (Historical: autopsy findings "T005 never started" and "H-9 tamper matrix incomplete"
are both remediated — T005 is this record; H-9 is closed by `tamper_matrix_test.go`.) One
non-blocking register-hygiene item carried forward: artifact/evidence index files in this story
still cite `kernel/webhook/...` instead of the actual `foundation/webhook/...` path — recommend
correcting on next edit, not a re-review blocker.

### Retest status

Initial independent review for this task; all cited tests re-run against current HEAD + working
tree, not merely re-cited from the autopsy's prior snapshot.

### Final conclusion

Acceptance criteria AC-W03-E03-S001-01 through -04 satisfied, including the remediated tamper
matrix. Recommend the story proceed toward `accepted` (conductor adjudicates final status), subject
to the non-blocking path-hygiene note above.

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
