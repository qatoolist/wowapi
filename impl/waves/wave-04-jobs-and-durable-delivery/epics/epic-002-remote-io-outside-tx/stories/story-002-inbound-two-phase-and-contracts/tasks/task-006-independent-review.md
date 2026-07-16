---
id: W04-E02-S002-T006
type: task
title: Independent review
status: done
parent_story: W04-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S002-T001
  - W04-E02-S002-T002
  - W04-E02-S002-T003
  - W04-E02-S002-T004
  - W04-E02-S002-T005
acceptance_criteria:
  - AC-W04-E02-S002-01
  - AC-W04-E02-S002-02
  - AC-W04-E02-S002-03
  - AC-W04-E02-S002-04
artifacts: []
evidence: []
---

# W04-E02-S002-T006 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all four acceptance criteria
are proven with valid evidence; T4's breaking signature change to `HandleInbound`'s
transaction-ownership contract is recorded as an explicit compatibility consideration, not silently
absorbed; T7 is correctly treated as cross-reference-only against DATA-08 W0-T2's already-executed
evidence, not re-implemented anywhere in this story or epic; the 6-boundary chaos test genuinely
reuses W04-E01-S003's harness rather than reimplementing one; no source requirement (DATA-03 T4, T5,
T6, T8) was silently dropped or narrowed.

### Parent story

W04-E02-S002 — Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos
test.

### Owner

unassigned

### Status

done (review outcome: story is unimplemented, not yet ready for the T001–T005 review this task
describes — see Verification Record)

### Dependencies

W04-E02-S002-T001 through T005 (review requires all five to be implemented first).

### Detailed work

1. Confirm T001's two-phase protocol matches PLAN DATA-03 T4's acceptance criterion ("Secret
   rotation/deactivation between phases cannot cause accept-under-stale-policy") and that the
   breaking-change note is genuinely recorded in `story.md`/`plan.md`, not silently folded into
   implementation detail.
2. Confirm T002's failed-signature audit matches PLAN DATA-03 T5's acceptance criterion ("No raw
   body ever persisted on failed verification"), verified by direct inspection of the
   empty-body-field test's actual assertion, not merely its name.
3. Confirm T003's adapter-contract mechanism matches PLAN DATA-03 T6's acceptance criterion
   ("Adapter cannot be registered for a non-idempotent high-impact operation without declaring
   duplicate-safety") and that the `Sender` inventory (step 1 of T003) is genuinely complete against
   the actual set of registered adapters in the repository at review time.
4. Confirm T004's chaos test matches PLAN DATA-03 T8's acceptance criterion ("Zero duplicate
   external effects across all 6 fault points") for both notify and webhook, and that it genuinely
   imports/reuses W04-E01-S003's harness rather than containing a parallel reimplementation.
5. Confirm T005's T7 cross-reference record correctly cites `DATA-08/wave0/legal-audit/` and
   contains no re-implementation of T7's scope anywhere in this story or epic.
6. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item —
   evidence without this must not be treated as final proof (mandate §10).
7. Confirm this story's `story.md` acceptance criteria are not narrower than PLAN DATA-03 T4/T5/T6/
   T8's own acceptance-criteria columns.
8. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record itself is captured in this task's own Verification Record.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E02-S002-01, AC-W04-E02-S002-02, AC-W04-E02-S002-03, AC-W04-E02-S002-04 (confirms all four,
does not itself prove any new one).

### Completion criteria

The review record confirms all four acceptance criteria are proven with valid evidence, that the
T4 breaking-change note and T7 cross-reference are both correctly recorded, or lists findings that
must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001–T005's evidence.

### Risks

None beyond the review itself missing a genuine gap — mitigated by requiring the review to
specifically re-check the "genuinely, not merely claimed" points above (breaking-change recording,
T7 cross-reference correctness, chaos-harness reuse vs. reimplementation) rather than trusting
T001–T005's own self-reported completion.

### Rollback or recovery considerations

Not applicable — a review task has no code to roll back; a failed review blocks story acceptance
until its findings are resolved.

## Implementation Record

*Not applicable — this is a review task, not an implementation task.*

### What was actually implemented

*Not applicable.*

### Components changed

*Not applicable.*

### Files changed

*Not applicable.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable.*

### Commits

*Not applicable.*

### Pull requests

*Not applicable.*

### Implementation dates

*Not applicable.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not applicable.*

### Follow-up items

*Not yet executed.*

### Relationship to the approved plan

*Not applicable.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E02-S002-01 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: rotation race closed, breaking-change note genuinely recorded | review report | unassigned |
| AC-W04-E02-S002-02 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: audit row genuinely body-free | review report | unassigned |
| AC-W04-E02-S002-03 | Independent review against mandate §14 checklist | Code review + inventory inspection | Confirmed: contract enforced, inventory genuinely complete | review report | unassigned |
| AC-W04-E02-S002-04 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: chaos test genuinely reuses W04-E01-S003's harness, zero duplicate effects | review report | unassigned |

### Actual result

Confirmed the 2026-07-16 status-honesty remediation (autopsy R-1) accurately reverted this story:
`story.md` frontmatter now reads `status: planned` (was falsely `accepted`); `closure.md` carries a
correction note ("frontmatter `status: accepted` was false ... reverted to `planned`") and its body
is unchanged, still reading "This story has not been implemented, verified, or closed." Confirmed no
partial-claim regression: `find . -type d -iname chaos` shows only `kernel/jobs/chaos` and
`foundation/bulk/chaos` — still no chaos test directory for notify or webhook anywhere in the repo,
matching the prior adversarial finding. Read `foundation/webhook/service.go`'s `HandleInbound`
(lines 34-127): it still runs signature verification, secret resolution, and event persistence
entirely inside the caller's single open tenant transaction, by design and by its own doc comment —
this is the pre-two-phase behavior; T001's claimed two-phase (short read-tx snapshot / verify outside
tx / short write-tx re-check) protocol for inbound verification does not exist in the codebase.
`grep`-style search found no adapter-contract idempotency-declaration mechanism (T003) and no
explicit-status per-class DSR reporting tied to this story (that capability exists but lives in
W04-E04-S002's `kernel/retention`, an unrelated story/epic). None of T001–T005 have real code; the
task list itself (not independently re-read line-by-line here, but consistent with `story.md`'s
`planned` status and `closure.md`'s unfilled sections) is consistent with "not started."

### Pass or fail

FAIL (as originally scoped) / status-honesty check PASSES. The story's four acceptance criteria
(AC-W04-E02-S002-01 through -04) are not met by any code in the repository. Separately, the specific
thing this re-review was asked to confirm — that the 2026-07-16 revert from `accepted` to `planned`
is accurate and that nothing now partially claims completion — is CONFIRMED TRUE: `story.md`,
`closure.md`, `verification.md`, and `evidence/index.md` are self-consistent in stating the story is
not implemented, and `impl/tracking/status-register.md` line 96 correctly lists
`W04-E02-S002 | story | ... | planned`.

### Evidence identifier

No new evidence produced (none exists to produce — the implementation does not exist). This record
itself is the evidence for the status-honesty confirmation.

### Execution date

2026-07-16.

### Commit or revision

HEAD 43b6e12 + remediation working tree 2026-07-16.

### Environment

macOS (darwin), local repository inspection; no test execution required (no test code exists for
this story's scope).

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3).

### Findings

1. (Confirmed, not a new finding) AC-W04-E02-S002-01 through -04 are unmet — no two-phase inbound
   verification, no adapter idempotency-safety contract, no chaos test for notify/webhook exists.
2. (New, positive) The 2026-07-16 status revert is accurate and complete: no document in this
   story's tree, nor `impl/tracking/status-register.md`, claims partial or full acceptance.

### Retest status

Not applicable — nothing to retest; the implementation has not been produced.

### Final conclusion

Recommend: **not-ready**. This story is correctly labeled `planned` (future work) as of 2026-07-16;
the revert from the prior false `accepted` claim is accurate and should be preserved. No part of
this story should be counted toward W04 acceptance. Recommend the conductor keep `status: planned`
and schedule this as a fresh, real implementation pass when prioritized.

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
