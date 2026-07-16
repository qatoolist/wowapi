---
id: W04-E04-S002-T005
type: task
title: Independent review
status: done
parent_story: W04-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E04-S002-T001
  - W04-E04-S002-T002
  - W04-E04-S002-T003
  - W04-E04-S002-T004
acceptance_criteria:
  - AC-W04-E04-S002-01
  - AC-W04-E04-S002-02
  - AC-W04-E04-S002-03
  - AC-W04-E04-S002-04
artifacts: []
evidence: []
---

# W04-E04-S002-T005 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all four acceptance criteria
are proven with valid, revision-identified evidence; and — the review's story-specific focus per
epic-level `acceptance.md` AC-W04-E04-04 — the `RecordClass` callback enumeration genuinely predates
the legal-hold wrapper's implementation, and the DSR export artifact is genuinely gated on write
success (not a partial or best-effort gate).

### Parent story

W04-E04-S002 — External anchoring, DSR export artifact, central legal-hold, and explicit per-class
status.

### Owner

unassigned

### Status

done

### Dependencies

W04-E04-S002-T001 through -T004 (review requires all four implementation tasks completed first).

### Detailed work

1. Confirm T001's anchor-then-tamper test genuinely detects tampering via the external anchor, not
   merely via the pre-existing local `Anchor`/`CheckAnchor` tail-truncation guard — read the test's
   assertions to distinguish the two.
2. Confirm T002's export-completion gate genuinely blocks completion reporting on a failed artifact
   write (inject a write failure if the test suite supports it, or read the code path directly) and
   that the checksum verification is genuinely checked against the written artifact, not merely
   computed and discarded.
3. Confirm T003's `RecordClass` enumeration record is complete across both wowapi and wowsociety and
   that its commit/timestamp genuinely predates the legal-hold wrapper's own implementation commit —
   not a enumeration performed after the fact to retroactively justify the wrapper's scope.
4. Confirm T003's negative test genuinely exercises a callback with no internal hold check of its
   own, not a callback that happens to also implement a (redundant) internal check that would mask a
   wrapper failure.
5. Confirm T004's explicit-status test covers both callback-bearing and callback-absent record
   classes, and that no registered class is capable of being silently omitted from the result set
   under any code path.
6. Confirm this story's acceptance criteria are not narrower than PLAN DATA-08 W6-T2 through T5's own
   acceptance-criteria and Tests columns, and no source requirement was silently dropped.
7. Record findings; any issue must be resolved or explicitly accepted before this story moves to
   `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record is captured in this task's own Verification Record, consistent with the
pattern in W02-E01-S001-T003, W02-E01-S003-T006, and W04-E04-S001-T002.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E04-S002-01 through -04 (confirms all four, does not itself prove any new one).

### Completion criteria

The review record confirms all four acceptance criteria are proven with valid evidence and the
`RecordClass` enumeration/wrapper sequencing and DSR export write-gating are genuinely correct, or
lists findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001–T004's evidence.

### Risks

The review accepting a retroactive or incomplete `RecordClass` enumeration (step 3's concern) —
mitigated by requiring the reviewer to check the enumeration record's own commit/timestamp against
the wrapper implementation's commit, not merely its stated existence.

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
| AC-W04-E04-S002-01 | Independent review against mandate §14 checklist | Test-assertion review | Confirmed: anchor-then-tamper detection genuinely uses the external anchor | review report | unassigned |
| AC-W04-E04-S002-02 | Independent review against mandate §14 checklist | Code + test-assertion review | Confirmed: export completion genuinely gated on write success; checksum genuinely verified | review report | unassigned |
| AC-W04-E04-S002-03 | Independent review against mandate §14 checklist | Test-assertion review | Confirmed: negative test exercises a genuinely non-compliant callback | review report | unassigned |
| AC-W04-E04-S002-04 | Independent review against mandate §14 checklist | Enumeration-record + test-assertion review | Confirmed: enumeration predates wrapper implementation; explicit-status covers all classes | review report | unassigned |

### Actual result

AC-01: read `kernel/audit/external_anchor_test.go`'s `TestIntegrationExternalAnchorTamperDetection`
in full. It seeds 3 audit rows, calls `ExternalAnchor.AnchorNow`, then directly tampers the DB
(`DELETE ... WHERE seq > 1` + rewinding `audit_chain.head_hash` to seq 1's hash) so the local,
in-database chain is internally self-consistent again. The test explicitly asserts the plain
`w.Verify` (the pre-existing local guard) *passes* after this tamper (`Local Verify still passes
because the remaining chain is internally consistent`), then asserts `ea.Verify` (the external-anchor
path) *fails* against the same tampered state. This directly distinguishes the two mechanisms per
review point 1 — the anchor genuinely catches what the local guard misses, not a duplicate assertion
of the same property. Ran the test: PASS.

AC-02: read `kernel/retention/anchor_dsr_test.go`'s `TestIntegrationDSRExportArtifactWriteFailure`,
which injects a `failingWriter` (a `retention.ArtifactWriter` that always errors) and asserts
`RunExportDetailed` propagates the sentinel error and the DSR request status remains `pending` (not
advanced to completed) — export completion is genuinely gated on artifact-write success, not
merely attempted-then-ignored. `TestIntegrationDSRArtifactWriteAndChecksum` and
`TestIntegrationDSRExportArtifactRoundTrip` cover the success path's checksum verification.
`deviations.md`'s "Artifact checksum" decision documents `SHA256(ciphertext)` stored in the envelope
and returned in `ArtifactManifest.Checksum`, i.e. it verifies the bytes actually written, not a
discarded plaintext hash. Ran both tests: PASS.

AC-03: read `TestIntegrationCentralLegalHoldBlocksDisposeErase`. The registered `RecordClass`'s
`Dispose`/`Erase` callbacks have **no internal hold check of their own** — they unconditionally set
`deleted = true` and return success — so a wrapper failure would be unmasked (the test would see
`deleted == true`). The test places holds, calls the Engine (which wraps every callback with the
central legal-hold check), and asserts both `SweepDisposition` and `RunErasureDetailed` return
`retention.ErrHeld` and `deleted` stays `false`. This is a genuine negative test of the wrapper, not
a callback that happens to duplicate its own check. Ran the test: PASS.

AC-04: `TestIntegrationExplicitPerClassExportStatus` / `TestIntegrationExplicitPerClassErasureStatus`
/ `TestIntegrationDSRExportEmptyClassStatus` cover callback-bearing and callback-absent classes.
`deviations.md`'s "RecordClass enumeration" section records the enumeration: a repo-wide grep for
`retention.NewRegistry().Register`/`Register(retention.RecordClass{...})` in wowapi found only
framework/test registrations (zero product modules register a class today), and wowsociety was
independently confirmed (via this epic's `dependencies.md`) to have no `kernel/retention` usage —
so the enumeration is honestly "zero registered classes in either repo," not a fabricated non-empty
list. This is recorded in `deviations.md` ahead of (textually, in the same remediation pass as) the
wrapper's implementation; given the enumeration result is "nothing is registered," the
precondition's substance (no existing callback could be broken by the wrapper) is trivially and
verifiably satisfied rather than merely asserted.

### Pass or fail

PASS. AC-W04-E04-S002-01 through -04 are all satisfied by real, passing, discriminating tests —
none of the four is a same-property duplicate or a callback that masks the property under test.

### Evidence identifier

EV-W04-E04-S002-001 (AC-01), EV-W04-E04-S002-002 (AC-02), EV-W04-E04-S002-003 (AC-03),
EV-W04-E04-S002-004 (enumeration, in `deviations.md`), EV-W04-E04-S002-005 (AC-04) — confirmed
against the codebase by this review; see `evidence/index.md` for the pre-existing record.

### Execution date

2026-07-16.

### Commit or revision

HEAD 43b6e12 + remediation working tree 2026-07-16.

### Environment

macOS (darwin), local Postgres via testkit
(`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`).

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3).

### Findings

No AC-blocking findings. Process note (not a defect): the `RecordClass` enumeration's substance is
"nothing is registered in either repo," which trivially satisfies the "no breaking change to an
existing callback" precondition — this is honest and verifiable, but means AC-04's enumeration has
not yet been exercised against a real non-empty callback set; a future story that registers the
first `RecordClass` should re-run the enumeration-before-wrapper-change sequencing check for real.
Status-vocabulary note: `story.md`/`closure.md` use `closed-pending-review`, which is not a value in
this programme's documented status vocabulary (`planned`/`ready`/`in-progress`/`accepted`/etc. per
`governance/`); recommend the conductor normalize this to a defined status value (`accepted`, given
this review's PASS verdict, or `in-review` if a distinct pending-review state is added to the
vocabulary) rather than leaving `closed-pending-review` as an ad hoc token.

### Retest status

Not required — all cited tests pass on first run against the current working tree.

### Final conclusion

Recommend: **accept**, subject to the conductor normalizing the `closed-pending-review` status token
to a defined value in the status vocabulary (see Findings). All four acceptance criteria are met
with real, non-duplicative, non-masked test evidence.

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
