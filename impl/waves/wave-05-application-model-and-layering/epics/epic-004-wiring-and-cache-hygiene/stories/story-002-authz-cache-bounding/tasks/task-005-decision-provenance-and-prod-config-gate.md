---
id: W05-E04-S002-T005
type: task
title: Decision provenance and prod-config gate
status: todo
parent_story: W05-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E04-S002-T004
acceptance_criteria:
  - AC-W05-E04-S002-03
artifacts:
  - ART-W05-E04-S002-005
evidence:
  - EV-W05-E04-S002-005
  - EV-W05-E04-S002-006
---

# W05-E04-S002-T005 — Decision provenance and prod-config gate

## Task Definition

### Task objective

Expose `CacheHit`/epoch-observed metadata on `Decision`, and require an explicit max-size +
stale-allow bound in `prod` config, failing boot without both.

### Parent story

W05-E04-S002 — Bounded, epoch-invalidated authorization cache.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E04-S002-T004 (PLAN T5's own dependency row: "T1-T4"; T6's own: "T1-T5").

### Detailed work

1. Extend `Decision`'s own type with `CacheHit`/epoch-observed fields.
2. Write a test confirming `Decision` metadata differs for a cache-hit vs. a cache-miss/
   epoch-observed scenario, producing `SEC-04/decision-provenance-tests.md`.
3. Implement prod-config validation: a `prod` profile with the cache enabled but no explicit
   max-size and stale-allow bound fails boot, per `config.go`'s existing pattern for this kind of
   validation.
4. Write a negative config test, producing `SEC-04/prod-config-gate-tests.md`.
5. Document both additions.

### Expected files or components affected

The `Decision` type; `config.go`.

### Expected output

`Decision` metadata correctly distinguishes hit/miss/epoch-observed; `prod` boot fails without an
explicit cache bound.

### Required artifacts

ART-W05-E04-S002-005.

### Required evidence

EV-W05-E04-S002-005, EV-W05-E04-S002-006.

### Related acceptance criteria

AC-W05-E04-S002-03.

### Completion criteria

Both named tests pass.

### Verification method

Direct execution of both tests.

### Risks

Low, per PLAN T5/T6's own risk columns — "an established pattern already exists in `config.go`" for
T6 specifically.

### Rollback or recovery considerations

If the prod-config gate is found to be bypassable (e.g. one bound present but not both, still
passing), fix before proceeding.

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

*Not yet implemented — new prod-config keys for max-size/stale-allow bound; recorded here once
implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented — the prod-config gate is a required security control; recorded here once
implemented.*

### Observability changes

*Not yet implemented — the Decision provenance metadata; recorded here once implemented.*

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
| AC-W05-E04-S002-03 | Run the decision-provenance test and the prod-config negative test | Local dev or CI, Go toolchain | Decision metadata differs hit vs. miss; prod+cache-enabled+missing-bound fails boot | test report | unassigned |

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
