---
id: W03-E03-S001-T004
type: task
title: Provider-verifier contract document (SEC-03 T4)
status: done
parent_story: W03-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W03-E03-S001-T001
  - W03-E03-S001-T002
  - W03-E03-S001-T003
acceptance_criteria:
  - AC-W03-E03-S001-04
artifacts:
  - ART-W03-E03-S001-005
evidence: []
---

# W03-E03-S001-T004 — Provider-verifier contract document (SEC-03 T4)

## Task Definition

### Task objective

Document the provider-verifier contract: what any `Verifier` implementation
must guarantee, including a reference example.

### Parent story

W03-E03-S001 — Bind webhook replay and dedup to provider-authenticated data.

### Owner

unassigned

### Status

done

### Dependencies

W03-E03-S001-T001, W03-E03-S001-T002, W03-E03-S001-T003 — PLAN's own
Depends-on column for T4: "T1-T3." Sequenced last so the document accurately
describes the final, as-built implementation shape.

### Detailed work

1. Identify the target documentation location currently covering the webhook
   module's provider-integration contract (or create a new, appropriately-located
   document if none exists).
2. Write the contract: `Envelope`'s fields must always be derived from
   authenticated data, never from caller-supplied request fields; `Verify` must
   return a non-nil `error` on failure and callers must not read `Envelope` on
   failure; document `HMACVerifier`'s specific
   receipt-time/authenticated-body synthesis approach and its stated limitation
   for timestamped-provider protocols.
3. Include a reference example (e.g., `HMACVerifier`'s own implementation,
   referenced or excerpted, as the canonical example of a compliant
   `Verifier`).

### Expected files or components affected

The identified/created documentation file.

### Expected output

A provider-verifier contract document exists, with a reference example,
accurately describing the `Envelope` guarantee this story establishes.

### Required artifacts

ART-W03-E03-S001-005 (the provider-verifier contract document).

### Required evidence

None beyond the documentation review recorded in `../verification.md` — this
task produces no test output of its own.

### Related acceptance criteria

AC-W03-E03-S001-04.

### Completion criteria

The document exists, is accurate against the actual T001-T003 implementation,
and includes a working reference example.

### Verification method

Documentation review against the actual implementation, confirming accuracy.

### Risks

Low, per PLAN's own T4 risk note.

### Rollback or recovery considerations

Not applicable — documentation-only task.

## Implementation Record

### What was actually implemented

- Created `artifacts/provider-verifier-contract.md` under the story directory.
- Documented the authenticated-fields-only core guarantee.
- Documented the failure contract (`Envelope` undefined on non-nil error).
- Documented each `Envelope` field's semantics and source rule.
- Included a reference example based on `HMACVerifier`.
- Included a limitation example for timestamped-provider protocols.
- Included an implementation checklist for future verifiers.
- Referenced the contract from `kernel/webhook/webhook.go` in the `Verifier`
  interface godoc.

### Components changed

Story artifacts; `kernel/webhook/webhook.go` (interface godoc reference).

### Files changed

- `impl/waves/wave-03-identity-and-session-security/epics/epic-003-webhook-authenticated-replay/stories/story-001-webhook-authenticated-replay/artifacts/provider-verifier-contract.md` (new)
- `kernel/webhook/webhook.go` (godoc reference)

### Interfaces introduced or changed

None.

### Configuration changes

Not applicable.

### Schema or migration changes

Not applicable.

### Security changes

Not applicable — documents an existing security guarantee, does not itself
change behavior.

### Observability changes

Not applicable.

### Tests added or modified

Not applicable.

### Commits

TBD at story closure.

### Pull requests

TBD at story closure.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

None.

### Follow-up items

None.

### Relationship to the approved plan

Implemented as planned.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E03-S001-04 | Review the provider-verifier contract document and its reference example for accuracy against the actual implementation | Documentation review | Contract document exists, accurately describes the `Envelope` guarantee, and includes a working reference example | document review record | unassigned |

### Actual result

Reviewed `artifacts/provider-verifier-contract.md` against
`kernel/webhook/verifier.go` and `kernel/webhook/service.go`. The contract
accurately describes the `Envelope` guarantee and the `HMACVerifier` reference
example matches the implementation.

### Pass or fail

PASS.

### Evidence identifier

N/A — documentation review recorded here and in `../verification.md`.

### Execution date

2026-07-13.

### Commit or revision

TBD at story closure.

### Environment

Local dev.

### Reviewer

unassigned.

### Findings

None.

### Retest status

Not retested.

### Final conclusion

Task complete; AC-W03-E03-S001-04 satisfied.

## Deviations Record

No deviations recorded.
