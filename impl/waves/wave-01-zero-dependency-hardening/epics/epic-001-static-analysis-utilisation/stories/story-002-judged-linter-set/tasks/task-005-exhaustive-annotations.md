---
id: W01-E01-S002-T005
type: task
title: exhaustive annotations (2 sites, workflow package)
status: done
parent_story: W01-E01-S002
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S002-04
artifacts:
  - ART-W01-E01-S002-006
  - ART-W01-E01-S002-007
evidence:
  - EV-W01-E01-S002-006
---

# W01-E01-S002-T005 — exhaustive annotations (2 sites, workflow package)

## Task Definition

### Task objective

Annotate exhaustive's 2 named hits at `kernel/workflow/definition.go:313` and
`kernel/workflow/runtime.go:170` to satisfy the linter while explicitly preserving and documenting the
existing fail-closed `default:` arm's intentional design. Both were reviewed and rejected as bugs
(personally verified by Fable 5) — this task annotates, it does not convert either switch into an
exhaustive case enumeration or otherwise alter the underlying logic.

### Parent story

W01-E01-S002 — Enable and triage the judged linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001-T004/T006 (disjoint files).

### Detailed work

1. Run `golangci-lint run --enable=exhaustive ./kernel/workflow/...` to confirm the fail-first
   "before" state (2 hits at the named lines).
2. Read both switch statements to confirm each already has a `default:` arm that fails closed (i.e.
   rejects/errors on an unrecognized case rather than silently proceeding) — reconfirming the
   assessment already recorded in `story.md`, not assuming it without a fresh read.
3. Confirm the pinned golangci-lint v2.11.4 exhaustive analyzer's supported suppression mechanism
   (resolving `plan.md`'s "Unresolved questions" item on this point — a `//exhaustive:ignore`
   directive, a `//nolint:exhaustive` directive, or a default-case-exempts configuration option).
4. Apply the confirmed suppression mechanism at each of the two sites, with an accompanying comment
   explaining that the `default:` arm is an intentional fail-closed design (not a missing-case gap)
   and, where useful, briefly stating what the default arm does (reject/error) so a future reader does
   not mistake the suppression for an oversight.
5. Re-run `golangci-lint run --enable=exhaustive ./kernel/workflow/...` to confirm the "after" state
   (0 hits).
6. Run the existing `kernel/workflow` test suite (`go test ./kernel/workflow/...`) to confirm no
   behavioral regression — this task makes no logic change, so this is a confirmation step, not an
   expectation of new failures.

### Expected files or components affected

`kernel/workflow/definition.go`, `kernel/workflow/runtime.go`.

### Expected output

Both sites carry an accurate suppression annotation preserving and explaining the fail-closed
`default:` arm; `exhaustive` reports 0 hits against `kernel/workflow/...`.

### Required artifacts

ART-W01-E01-S002-006, ART-W01-E01-S002-007.

### Required evidence

EV-W01-E01-S002-006 (fail-before/pass-after static-analysis report + review note).

### Related acceptance criteria

AC-W01-E01-S002-04.

### Completion criteria

Both sites' `default:` arms are confirmed fail-closed by design; both carry the confirmed suppression
mechanism with an explanatory comment; `exhaustive` exits 0 against both files; the underlying switch
logic is unchanged (byte-for-byte identical apart from the added comment/directive, confirmed by
diff review).

### Verification method

Direct command execution (`golangci-lint run --enable=exhaustive`, `go test ./kernel/workflow/...`),
logged output retained as evidence; manual review confirming the annotation comment accurately
describes the fail-closed intent and that no logic was altered beyond the comment/directive addition.

### Risks

Low for the annotation mechanics; the primary risk this task must guard against is scope creep into
"fixing" the switch (e.g. adding explicit cases) when the design intent is that the `default:` arm
handles the unenumerated cases deliberately — the task's completion criteria explicitly require
confirming the underlying logic is unchanged.

### Rollback or recovery considerations

Revert if `go test ./kernel/workflow/...` regresses, which would indicate an unintended logic change
crept in alongside the annotation — investigate rather than silently reintroducing the change without
understanding the regression.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

definition.go:313 and runtime.go:170 annotated `//exhaustive:ignore` with comments documenting the fail-closed design (no-outgoing-for-unknown-types; deny-by-default invalid_decision arm) — NOT converted to enumerations. Drift sites bind.go:326 / schema.go:95 (fail-closed fall-through returns) annotated identically.

## Verification Record

AC-W01-E01-S002-04: fail-before in both enumerations (4 hits); exhaustive per-linter run exit 0 after (EV-006); annotation comments present and accurate at all 4 sites. **pass**

### Final conclusion

Fail-closed designs preserved, documented, linter satisfied.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
