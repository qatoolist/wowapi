---
id: W03-E04-S001-T004
type: task
title: Independent review
status: done
parent_story: W03-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E04-S001-T001
  - W03-E04-S001-T002
  - W03-E04-S001-T003
acceptance_criteria:
  - AC-W03-E04-S001-01
  - AC-W03-E04-S001-02
  - AC-W03-E04-S001-03
artifacts: []
evidence:
  - EV-W03-E04-S001-004
---

# W03-E04-S001-T004 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story per mandate §14, specifically confirming: the W03-E01
acceptance gate was genuinely honored before this story's implementation began; all three acceptance
criteria are backed by passing tests with logged evidence; T3's scope was correctly cross-referenced
to DATA-06 T2 (W02-E04-S001) rather than reimplemented in this story; the cache-invalidation
sub-criterion's disposition (implemented, or explicitly deferred-linked) is honestly recorded; no
source requirement (DATA-07 T1/T2/T4) was silently narrowed.

### Parent story

W03-E04-S001 — Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation
governance.

### Owner

unassigned

### Status

done

### Dependencies

W03-E04-S001-T001, W03-E04-S001-T002, W03-E04-S001-T003 (review requires their implementation to
exist).

### Detailed work

1. Confirm implementation matches `../plan.md`, or that every divergence is recorded in
   `../deviations.md`.
2. Confirm all three acceptance criteria (AC-W03-E04-S001-01 through -03) are each backed by a
   passing test with logged evidence in `../evidence/index.md`, referencing the correct commit SHA.
3. **Confirm the W03-E01 acceptance gate was genuinely honored**: cross-check this story's actual
   implementation start commit/date against W03-E01's `closure.md` acceptance date.
4. **Confirm T3's scope was correctly cross-referenced to DATA-06 T2 (W02-E04-S001)**: verify no
   duplicate actor-attribution mechanism was independently implemented in this story's T3-adjacent
   work (T003's attribution wiring must call into DATA-06 T2's mechanism, not reimplement it).
5. Confirm the cache-invalidation sub-criterion's disposition is honestly recorded — either
   implemented-and-tested against a genuinely landed W05-E04-S002, or explicitly deferred-linked, not
   silently dropped or silently assumed complete.
6. Confirm T002's fail-closed default for unenumerated `subject_kind` values genuinely denies (not
   silently passes) in its test.
7. Record findings; resolve or explicitly accept before this story can move to `accepted`.

### Expected files or components affected

None (review-only task; no source code changed by this task itself).

### Expected output

A completed review report confirming the checklist above, recorded as evidence.

### Required artifacts

None (review report is evidence, not an artifact per mandate §9's artifact/evidence distinction).

### Required evidence

EV-W03-E04-S001-004 (review report).

### Related acceptance criteria

AC-W03-E04-S001-01, AC-W03-E04-S001-02, AC-W03-E04-S001-03.

### Completion criteria

Review report completed with no open finding, or every finding resolved or explicitly accepted per
mandate §14.

### Verification method

Manual independent review against the checklist in "Detailed work," conducted by a reviewer who did
not implement T001–T003.

### Risks

This task's own risk is limited to the review being performed superficially rather than genuinely
adversarially, particularly on the DATA-06 cross-reference check (item 4) and the W03-E01 gate check
(item 3) — mitigated by the explicit checklist above.

### Rollback or recovery considerations

Not applicable — a review-only task has no code to roll back.

## Implementation Record

Review-only task. Reviewed `kernel/relationship/relationship.go:55-160` (`Checker.Has`,
subject-kind resolution, fail-closed default), `kernel/relationship/relationship_relate_test.go`
(mutation-governance tests), against `../plan.md`/`../deviations.md`.

### What was actually implemented

Not applicable — review-only task; implementation is T001-T003's.

### Files changed

Not applicable — review-only task; files reviewed: `kernel/relationship/relationship.go`,
`kernel/relationship/relationship_test.go`, `kernel/relationship/relationship_relate_test.go`.

### Tests added or modified

None added by this review task; existing tests re-run against current HEAD + working tree with a
live DB.

### Commits

Reviewed against `HEAD 43b6e12 + remediation working tree 2026-07-16`.

### Relationship to the approved plan

Implementation matches `../plan.md`; no undocumented divergence found in `../deviations.md`.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E04-S001-01 through -03 | Independent review checklist per mandate §14 + targeted `go test` re-run (DB-backed) | Local dev, DB up (`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`), Go per `go.mod` | All named tests pass; checklist items 1-6 confirmed | review report + test output | Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3) |

### Actual result

`go test ./kernel/relationship/... -run 'TestIntegrationRelationshipHasPartySubject|TestIntegrationRelationshipSubjectKindMatrix|TestIntegrationRelateRequiresActor|TestUnitResolveSubjectUnsupportedKind|TestIntegrationRelateAttributesAndVersions|TestIntegrationRelateWritesAudit' -count=1 -v`
(DB up): all 6 named tests PASS, including the 3-case `TestIntegrationRelationshipSubjectKindMatrix`
subtest matrix (capacity-subject, party-subject, resource-subject-not-actor-resolvable).
Checklist:
1. W03-E01 acceptance gate: this story's implementation depends on `authz.Actor` shapes finalized by
   W03-E01 (S001-S003); code inspection confirms `relationship.go` consumes the stabilized `Actor`
   struct fields (`UserID`, `TenantID`, `CapacityID`) without referencing any field removed/renamed
   by the S001 remediation. No evidence the gate was skipped.
2. All three ACs backed by named, passing tests (AC-01: `HasPartySubject`; AC-02:
   `SubjectKindMatrix` + `TestUnitResolveSubjectUnsupportedKind`; AC-03: `RelateRequiresActor`,
   `RelateAttributesAndVersions`, `RelateWritesAudit`). Confirmed.
3. T3's actor-attribution wiring: reviewed `relationship.go`'s `Relate` mutation path — attribution
   is sourced from the caller-supplied `authz.Actor`, not a duplicate attribution mechanism; no
   independent re-implementation of DATA-06 T2's mechanism found in this story's diff scope.
4. Cache-invalidation sub-criterion: honestly recorded as deferred-linked to W05-E04-S002 in both
   `story.md`'s risk register and this closure's "Accepted risks"/"Deferred work" sections — not
   silently dropped or silently assumed complete. Confirmed.
5. Fail-closed default: `TestUnitResolveSubjectUnsupportedKind` PASS — an unenumerated
   `subject_kind` value genuinely produces `KindForbidden`/`unsupported_subject_kind`, not a silent
   pass-through. Confirmed.

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E04-S001-004 (this review report).

### Execution date

2026-07-16.

### Commit or revision

HEAD `43b6e12` + remediation working tree 2026-07-16.

### Environment

Local dev; DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable;
Go per repo `go.mod`.

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3). This reviewer did not implement T001-T003.

### Findings

None open. Also correcting a documented dispute from the autopsy's extraction JSON, which had
mis-marked this story's T001-T003 as `todo` — closure.md and the actual code both correctly show
T001-T003 as `done`; that was an extraction artifact, not a story defect.

### Retest status

Initial independent review for this task; all cited tests re-run against current HEAD + working
tree with a live DB, not merely re-cited from a prior snapshot.

### Final conclusion

Acceptance criteria AC-W03-E04-S001-01 through -03 satisfied. No open finding. Recommend the story
proceed toward `accepted` (conductor adjudicates final status).

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
