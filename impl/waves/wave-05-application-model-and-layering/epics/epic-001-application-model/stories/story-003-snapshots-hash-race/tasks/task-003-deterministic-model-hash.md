---
id: W05-E01-S003-T003
type: task
title: Deterministic model hash
status: todo
parent_story: W05-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E01-S003-T001
  - W05-E01-S003-T002
acceptance_criteria:
  - AC-W05-E01-S003-03
artifacts:
  - ART-W05-E01-S003-003
evidence:
  - EV-W05-E01-S003-003
---

# W05-E01-S003-T003 — Deterministic model hash

## Task Definition

### Task objective

Implement a deterministic hash function over the sealed `ApplicationModel`, emitted at
startup/readiness: two identical compiles produce a byte-identical hash; one changed declaration
produces a different hash; non-deterministic inputs (map order, timestamps) are excluded.

### Parent story

W05-E01-S003 — Snapshot immutability, post-seal rejection, model hash, and race safety.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E01-S003-T001, W05-E01-S003-T002 (PLAN T9's own dependency row: "T1-T8" — the fuller model
surface should be stable before hashing it).

### Detailed work

1. Design the model-hash algorithm: a canonical serialization of the sealed model's declarations,
   excluding non-deterministic inputs (map iteration order, timestamps), followed by a cryptographic
   hash.
2. Implement the hash function.
3. Wire the hash into startup/readiness reporting.
4. Write `AR-01/model_hash_determinism_test.go`: two identical compiles produce a byte-identical
   hash; one changed declaration produces a different hash.
5. Document the hash algorithm and its determinism guarantees.

### Expected files or components affected

The `ApplicationModel`/`Compiler` from S001 (exact file paths TBD per `plan.md`); readiness-reporting
integration point (exact location TBD).

### Expected output

A deterministic model hash, emitted at startup/readiness, proven by the named test.

### Required artifacts

ART-W05-E01-S003-003.

### Required evidence

EV-W05-E01-S003-003.

### Related acceptance criteria

AC-W05-E01-S003-03.

### Completion criteria

Two identical compiles emit a byte-identical hash; one changed declaration emits a different hash —
proven by the named test passing.

### Verification method

Direct execution of `AR-01/model_hash_determinism_test.go`.

### Risks

Low, per PLAN T9's own risk column — the main risk (non-deterministic inputs leaking into the hash)
is explicitly named and mitigated by design (exclude map order, timestamps).

### Rollback or recovery considerations

If the determinism test reveals a non-deterministic input leaking into the hash (e.g. a map
iteration order dependency not caught by design), fix the serialization before proceeding — do not
ship a hash function that is not genuinely deterministic.

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

*Not applicable.*

### Observability changes

*Not yet implemented — the model hash is itself an observability artifact; recorded here once
implemented.*

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
| AC-W05-E01-S003-03 | Run `AR-01/model_hash_determinism_test.go` | Local dev or CI, Go toolchain | Byte-identical hash for identical compiles; different hash on change | unit-test report | unassigned |

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
