---
id: W02-E05-S001-T001
type: task
title: Catalog manifest design investigation
status: done
parent_story: W02-E05-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W02-E05-S001-01
artifacts:
  - ART-W02-E05-S001-001
evidence:
  - EV-W02-E05-S001-001
---

# W02-E05-S001-T001 — Catalog manifest design investigation

## Task Definition

### Task objective

Resolve, with documented rationale, the seven design questions MATRIX CS-21 explicitly leaves open
for FBL-02 ("design detail to be ratified in Phase 5"): catalog manifest format; versioning scheme;
CLI command shape; idempotency mechanism; RLS/role posture; dry-run output format; audit-record
integration. Produce a decision record before any implementation task (T002–T005) begins.

### Parent story

W02-E05-S001 — Production seed-sync path — design investigation and implementation.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Re-read `kernel/httpx/health.go`, `app/health.go`, and any existing seed-related tooling at this
   task's actual start commit to confirm the story's current-state assessment still holds (no
   production seed-sync path exists; the readiness mechanism itself is sound and fail-closed).
2. Draft options for the catalog manifest's storage format and schema (e.g. a versioned file per
   catalog domain, a single consolidated manifest, or a registry-backed approach), with trade-offs;
   select one and document the rationale.
3. Resolve the versioning scheme (e.g. semantic version, monotonic integer, or content hash) the
   manifest uses, and document why it is sufficient for both idempotency checking and the
   readiness-payload hash-reporting requirement.
4. Resolve the exact CLI command shape beyond CS-21's own sketch (`wowapi seed sync --env prod`) —
   subcommand structure, flags, exit-code conventions.
5. Resolve the idempotency mechanism (e.g. content-hash comparison against an applied-version
   tracking table, or another approach) that a repeated seed-sync run against the same manifest
   version must satisfy.
6. Resolve the RLS/role posture the seed-sync path runs under, with an explicit, written
   justification for why that posture is safe in a bootstrap scenario (an empty catalog database, by
   construction, cannot yet enforce the tenancy controls those catalogs would otherwise configure).
   This is the single most safety-critical decision in this task — see RISK-W02-E05-001. State
   plainly which role the sync runs as and why "RLS-respecting" is genuinely true of that role's
   behavior, not merely true by label.
7. Resolve the dry-run output format (structured diff, human-readable summary, or both).
8. Confirm or reject the `kernel/audit` integration plausibility noted in `epic.md`'s "Architectural
   context" for the audit-record requirement; document the decision either way.
9. Resolve whether concurrent seed-sync invocation (e.g. two deploy processes racing during a rolling
   production deployment) is in-scope for this story's idempotency guarantee, or is explicitly
   assumed away as single-invoker — document the choice, do not leave it implicit.
10. Resolve whether the new readiness check (T004's scope) applies only in prod profile, per CS-21's
    own "a prod-profile boot" wording, or more broadly — document the choice.
11. Assemble all of the above into a single decision record (ART-W02-E05-S001-001), each decision
    stated with its rationale, distinguishing confirmed source facts (CS-21's fixed acceptance bar)
    from this task's own design choices.
12. If any of the above decisions is, on reflection, of genuinely D-0N ADR caliber (a new
    framework-wide convention intended to outlive this story, or a new external dependency), flag it
    explicitly in this task's own record and escalate per `epic.md`'s "Required decisions" process
    safeguard, rather than silently absorbing it as an ordinary story-scoped decision.

### Expected files or components affected

None (this task produces a design decision record; it does not itself change code). The decision
record's own file location is expected under this story's `artifacts/` structure once produced,
exact path TBD.

### Expected output

A documented design decision (ART-W02-E05-S001-001) resolving all seven named open questions plus
the two sequencing questions (concurrent-invocation scope, prod-profile-only readiness check), each
with stated rationale, recorded before T002–T005 begin.

### Required artifacts

ART-W02-E05-S001-001 (design-investigation decision record).

### Required evidence

EV-W02-E05-S001-001 (the decision record itself, evidenced by its existence, completeness, and a
commit/date predating any T002–T005 commit).

### Related acceptance criteria

AC-W02-E05-S001-01.

### Completion criteria

The decision record exists, is complete against all nine questions listed under "Detailed work," and
predates (by commit timestamp) the first commit of any implementation task.

### Verification method

Inspection of the decision record for completeness against the nine named questions; git-history
inspection confirming the record's commit predates T002–T005's first commits.

### Risks

RISK-W02-004 (the investigation may surface a need for new infrastructure not anticipated by this
wave's planning) and RISK-W02-E05-001 (the RLS-respecting bootstrap tension) — see epic-level
`risks.md`. This task is the origin point for both risks' actual resolution or escalation.

### Rollback or recovery considerations

If a later task (T002–T006) discovers this task's decision was materially flawed (e.g. the resolved
RLS/role posture turns out not to be safe under a concrete test in T002), the correction is recorded
as a deviation in this task's own Deviations Record (or the consuming task's, whichever first
discovers it) — the original decision record is not silently rewritten to hide that the flaw existed,
per mandate §2.6.

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

*Not applicable — this task produces a design decision, not a schema change.*

### Security changes

*Not applicable — the RLS/role-posture decision is documented here; its actual implementation (and
any resulting security change) is T002's scope.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable — this task produces a decision record, not code; it has no test of its own.*

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

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E05-S001-01 | Inspect decision record for completeness against all nine named questions; confirm commit predates implementation-task commits | Documentation review, git history inspection | Complete, dated decision record exists and predates implementation | design-decision record | unassigned |

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
