---
id: W04-E01-S001-T001
type: task
title: Shared primitive design, implementation, and cross-consumer field-set review
status: done
parent_story: W04-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E01-S001-01
  - AC-W04-E01-S001-02
artifacts:
  - ART-W04-E01-S001-001
  - ART-W04-E01-S001-003
evidence:
  - EV-W04-E01-S001-001
  - EV-W04-E01-S001-002
---

# W04-E01-S001-T001 — Shared primitive design, implementation, and cross-consumer field-set review

## Task Definition

### Task objective

Design and implement the shared lease/fencing primitive (`lease_token`, monotonic
`lease_generation`, `lease_expires_at`, optional heartbeat) as a reusable kernel building block,
prove its token/generation comparison semantics with unit tests, and validate its field set against
DATA-03's (W04-E02) and DATA-04's (W04-E03) own stated needs before treating the design as locked.

### Parent story

W04-E01-S001 — Shared lease/fencing primitive.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Re-read `kernel/jobs`'s claim/finalize/reclaim SQL and W02-E01-S002's interim checkpoint-lease
   implementation at this task's actual start commit to confirm no shared lease/fencing primitive
   currently exists (resolving `plan.md`'s current-state re-confirmation step).
2. Draft the primitive's package-location and API-shape options with trade-offs; select one and
   document the rationale (resolves `plan.md`'s "Unresolved questions" item on package location).
3. Implement the primitive: `lease_token` generation, monotonic `lease_generation` semantics,
   `lease_expires_at` handling, and the optional heartbeat extension point.
4. Write unit tests on token/generation comparison semantics: a current token/generation pair
   compares as valid; a stale (superseded generation) or expired pair compares as rejected.
5. Read DATA-03's (PLAN, W04-E02 scope) and DATA-04's (PLAN, W04-E03 scope) own task rows and this
   wave's `wave.md` framework-capabilities list for what each requires from a shared lease type;
   confirm the primitive's field set covers both, or record any gap found.
6. Document the primitive's contract (field set, comparison semantics, package location).

### Expected files or components affected

A new lease/fencing primitive package and its unit tests (exact location TBD per `plan.md`).

### Expected output

A locked, cross-consumer-reviewed shared lease/fencing primitive with unit-tested comparison
semantics.

### Required artifacts

ART-W04-E01-S001-001 (shared primitive package), ART-W04-E01-S001-003 (documentation, shared with
T002).

### Required evidence

EV-W04-E01-S001-001 (unit-test report), EV-W04-E01-S001-002 (cross-consumer field-set review
record).

### Related acceptance criteria

AC-W04-E01-S001-01, AC-W04-E01-S001-02.

### Completion criteria

The primitive is implemented with passing unit tests on comparison semantics, and a dated,
attributed review record confirms its field set covers DATA-03's and DATA-04's own stated needs
before the design is treated as locked.

### Verification method

Direct execution of the unit-test suite; inspection of the cross-consumer review record for
existence, date, and attribution against DATA-03/DATA-04's own PLAN task rows.

### Risks

RISK-W04-E01-001 (an under-specified design, once consumed by S002 and eventually W04-E02/W04-E03,
is costly to retrofit) — see epic-level `risks.md`.

### Rollback or recovery considerations

If the cross-consumer review surfaces a material field-set gap after implementation has begun,
revert the design's "locked" status (not necessarily the code) and extend the field set before
re-submitting for review — do not lock a design the review has flagged as incomplete.

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

*Not applicable — this task implements a kernel-level type; it does not itself add a database
schema change (see `plan.md` "Persistence changes").*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

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

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E01-S001-01 | Run unit tests on token/generation comparison semantics | Local dev or CI, Go toolchain | Current pair validates; stale/expired pair rejects | unit-test report | unassigned |
| AC-W04-E01-S001-02 | Inspect cross-consumer field-set review record | Documentation review | Dated, attributed review record exists, confirms field set covers DATA-03/DATA-04 needs, predates locking | review report | unassigned |

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
