---
id: W01-E01-S002-T002
type: task
title: gosec G115 multi-site review (int-overflow conversions)
status: done
parent_story: W01-E01-S002
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S002-02
artifacts:
  - ART-W01-E01-S002-003
evidence:
  - EV-W01-E01-S002-003
---

# W01-E01-S002-T002 — gosec G115 multi-site review (int-overflow conversions)

## Task Definition

### Task objective

Enumerate the exact G115 site list (potentially-unsafe integer conversions that may overflow) across
the audit, database, jobs, mfa, and pagination packages via a fresh gosec run, then individually
review and dispose each site — either annotate (bounded-by-prior-validation justification, citing the
specific validation) or fix (add an explicit bounds check).

### Parent story

W01-E01-S002 — Enable and triage the judged linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T001/T003-T006 (disjoint files, once its own site list is enumerated). May
reuse T001's story-wide fresh-run baseline (EV-W01-E01-S002-001) if that has already been produced by
the time this task starts; otherwise this task produces its own gosec-scoped run.

### Detailed work

1. Run `golangci-lint run --enable=gosec ./...` (or reuse T001's EV-W01-E01-S002-001 baseline if
   available) and filter the output to G115 hits specifically. Produce the exact enumerated site list
   (file:line for every G115 hit) — this list does not exist in the source material and is this
   step's own output, not an input to it.
2. Record the enumerated site list and compare its size/location profile against the source material's
   characterization ("audit/database/jobs/mfa/pagination packages... most conversions are believed
   bounded by prior validation"). Record any material difference (a package not named in the source
   characterization, an unexpectedly large count, etc.) as a candidate deviation.
3. For each enumerated site, individually:
   a. Trace the value's origin within the same call path to determine whether a prior validation step
      (range check, type-narrowing guard, parsed-and-validated input, etc.) already bounds the value
      before this conversion.
   b. If a prior validation bounds it: annotate the site with a `#nosec G115` (or the pinned
      version's actual supported syntax) comment stating specifically which prior validation bounds
      it (e.g. "bounded by the range check at line N" or "bounded by the enum validation performed
      during config load") — a generic "bounded elsewhere" comment without a specific pointer is not
      sufficient disposition.
   c. If no prior validation bounds it: add an explicit bounds check immediately before the
      conversion (reject or clamp the value per the surrounding function's existing error-handling
      convention — to be judged per site), then re-run gosec to confirm the site no longer flags.
4. Aggregate all site dispositions (site, package, disposition, rationale/validation reference or
   description of the added bounds check) into a single per-site triage record.
5. Re-run `golangci-lint run --enable=gosec ./...` to confirm zero G115 hits remain unaddressed
   (either annotated or fixed) across the full module tree.
6. For any site that received a new explicit bounds check (not an annotation), add a targeted unit
   test exercising both the in-bounds (pass) and out-of-bounds (rejected) cases.

### Expected files or components affected

Packages within `kernel/audit/`, `kernel/database/`, `kernel/jobs/`, `kernel/mfa/`, and the pagination
package (exact package paths and files not yet known — to be enumerated by step 1).

### Expected output

Every G115 site is disposed (annotated or fixed) with a recorded rationale; a per-site triage record
exists as part of this story's aggregated triage record; gosec reports zero unaddressed G115 hits.

### Required artifacts

ART-W01-E01-S002-003 (G115 site fixes/annotations, aggregated across whatever files step 1
enumerates).

### Required evidence

EV-W01-E01-S002-003 (static-analysis report + per-site triage record).

### Related acceptance criteria

AC-W01-E01-S002-02.

### Completion criteria

Every enumerated G115 site has a recorded disposition (annotated with a specific validation reference,
or fixed with an explicit bounds check and a targeted unit test); `golangci-lint run --enable=gosec`
reports zero unaddressed G115 hits; the site count and disposition list are recorded in this story's
`implementation.md`, not silently summarized away.

### Verification method

Direct command execution (`golangci-lint run --enable=gosec`) plus `go test` for any site that
received a new bounds check, logged output retained as evidence; manual review of the per-site triage
record confirming no site received a generic, unspecific "bounded" annotation.

### Risks

Materially higher than T001/T003's single-site annotation tasks — this is an unenumerated,
multi-package review where an individual site could be mis-disposed (annotated as bounded when it is
not actually bounded). This task is the primary driver of RISK-W01-E01-002 for this story (see
epic-level `risks.md`) — the site count is unknown until step 1 executes, and could be materially
larger than what a triage pass planned in advance assumed.

### Rollback or recovery considerations

If a site's "bounded by prior validation" annotation is later found to be inaccurate (i.e. the prior
validation does not actually cover the range that could overflow), escalate and replace the annotation
with an explicit bounds check rather than silently leaving an inaccurate annotation in place — this is
a security-relevant recovery, not a routine rollback.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

G115 enumerated to 7 exact sites and individually reviewed (full rationale in story `implementation.md`): audit.go:164,176 / database.go:135 / jobs.go:105 / totp.go:116 annotated (bounded-by-prior-validation or bijective-reinterpretation, site-specific); pagination/cursor.go:202,210 FIXED with explicit fail-closed bounds checks (uint/uint64 > MaxInt64 now error instead of wrapping) + regression test `TestEncodeCursorUnsignedOverflow`.

## Verification Record

AC-W01-E01-S002-02 (G115 slice): per-site record complete (EV-003); cursor fix fail-first proven — the new test FAILS at pristine HEAD, passes after; kernel/pagination suite ok. **pass**

### Final conclusion

5 annotated with real justifications, 2 genuinely fixed — no blanket annotation.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
