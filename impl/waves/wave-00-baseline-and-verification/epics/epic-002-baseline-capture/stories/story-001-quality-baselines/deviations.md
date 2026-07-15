---
id: DEV-W00-E02-S001
type: deviation-log
parent_story: W00-E02-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations — W00-E02-S001

One deviation recorded. Per mandate §2.6, `plan.md` has not been rewritten; the divergence is
recorded here.

## Deviation ID

DEV-W00-E02-S001-001 (arising in task W00-E02-S001-T002).

## Approved plan

`plan.md` implementation-sequence step 2 and T002 step 2: "transcribe the verbatim list of all 25
analyzer names it queried" from MATRIX CS-23, then enable all 25 in the throwaway config.
`story.md` AC-W00-E02-S001-02 likewise says "all 25 MATRIX CS-23 analyzers".

## Actual implementation

The MATRIX document (`docs/implementation/fable5-closure-depth-matrix-2026-07-11.md` §CS-23)
states "All 25 queried analyzers ship unenabled" but **names only 18 analyzers verbatim**
(sqlclosecheck, rowserrcheck, bodyclose, wastedassign, makezero, musttag, testifylint, noctx,
copyloopvar, gocritic, nilerr, exhaustive, errorlint, gosec, forcetypeassert, usestdlibvars,
wrapcheck, revive). An exhaustive analyzer-name sweep of the MATRIX document, plus corroborating
reads of `fable5-final-architecture-review-2026-07-11.md` and `docs/working/lint-backlog.md`
(which name only subsets of the same 18), recovered no further names; the per-linter enablement
run logs the MATRIX's own closure-evidence register points at (`evidence/premier/FBL-05/`,
`FBL-07/`) do not exist in the repository. The throwaway config therefore enabled the **18
recoverable named analyzers**, and the drift comparison covers all 18, with the gap stated
explicitly in EV-W00-E02-S001-002 — per T002 step 6's own instruction to "state that explicitly
rather than inventing an expected value".

## Reason

The remaining 7 analyzer names are not recorded in the authoritative source (or anywhere else in
the repository). Inventing them would corrupt the apples-to-apples comparison the story exists to
provide.

## Impact

The lint baseline covers every analyzer the MATRIX actually names and every count/site it actually
records. Nothing the MATRIX recorded is left uncompared. The only loss is against the MATRIX's
own unsubstantiated "25" headcount.

## Risks

If the MATRIX author's missing 7 analyzers surface later (e.g. from session notes), they will
lack a W00 baseline count and would need a supplementary capture pinned to a then-current SHA.
Low severity: no recorded MATRIX expectation exists for them, so no downstream claim can cite one.

## Approval

Recorded by the executing worker 2026-07-13; approval is the conductor's story-review gate
(self-approval not claimed, per status discipline).

## Compensating controls

The throwaway config and both raw run outputs are preserved as artifacts, so the run is exactly
reproducible and extensible: adding any later-recovered analyzer names to the preserved config and
re-running yields a comparable supplementary baseline.

## Follow-up work

None mandated. Optional: W01-E01-S001 (FBL-05) may recover the intended full list when it
permanently enables analyzers; if so, it should note the delta against this baseline's 18.
