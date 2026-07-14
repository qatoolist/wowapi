---
id: W06-E03-S003
type: story
title: Blocking security scans — Trivy flip, waiver schema, visibility-guard regression check, private fallback
status: verified
wave: W06
epic: W06-E03
owner: W06E03Impl
reviewer: independent-review-gate
priority: P0
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - REL-02
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W06-E03-S003-01
  - AC-W06-E03-S003-02
  - AC-W06-E03-S003-03
  - AC-W06-E03-S003-04
  - AC-W06-E03-S003-05
artifacts:
  - ART-W06-E03-S003-001
  - ART-W06-E03-S003-002
  - ART-W06-E03-S003-003
  - ART-W06-E03-S003-004
  - ART-W06-E03-S003-005
evidence:
  - EV-W06-E03-S003-001
  - EV-W06-E03-S003-002
  - EV-W06-E03-S003-003
  - EV-W06-E03-S003-004
  - EV-W06-E03-S003-005
decisions: []
risks: []
---

# W06-E03-S003 — Blocking security scans — Trivy flip, waiver schema, visibility-guard regression check, private fallback

## Story ID

W06-E03-S003

## Title

Blocking security scans — Trivy flip, waiver schema, visibility-guard regression check, private fallback

## Objective

Flip Trivy to blocking (`exit-code: "1"`), scoping `ignore-unfixed` to a reviewed allowlist only; build
a waiver mechanism with owner/rationale/expiry/remediation-link per entry, CI-validated; add a
regression meta-check confirming `dependency-review`/`codeql`/`scorecard` actually ran whenever the
repository is public; build a local-scanner fallback for "the repository goes private again"; and wire
all REL-02 blocking checks into REL-01's Wave-0 manifest.

## Value to the framework

MATRIX CS-23 gives the exact current gap: "Trivy soft-fails by design (`:75`) = plan REL-02's exact
scope, now with line citations." PLAN's own REL-02 evidence, corrected against a mid-review live-state
finding: "wowapi's repository visibility flipped to public on 2026-07-03... Live verification (`gh api
repos/qatoolist/wowapi --jq '.visibility'` → `public`; `gh run view` on the two most recent CodeQL/
Scorecard runs → both `success`, not `skipped`) confirms these gates are already active today, not
skipped... Trivy's `exit-code: 0`/`ignore-unfixed: true` remains a real, live gap regardless of
visibility." This story converts the one remaining live gap (Trivy's soft-fail posture) into a blocking
gate, while also building the regression safety net (the visibility-guard meta-check) that ensures the
*other* three scanners' currently-passing state is not silently lost if the repository's visibility
changes again — exactly the kind of drift the source review's own pre-flight correction caught once
already.

## Problem statement

MATRIX CS-23's exact evidence: "Trivy soft-fails by design (`:75`)." PLAN's own REL-02 corrected
baseline: "only Trivy's non-blocking config is a currently-live gap, plus the absence of a documented
fallback for 'repo reverts to private' — CodeQL/Scorecard/dependency-review are already effectively
active." PLAN's own T1 acceptance criterion: "Trivy fails on CRITICAL/HIGH findings with an available
fix." T3's own framing: "Meta-check: assert `dependency-review`/`codeql`/`scorecard` actually *ran*
whenever the repo is public, as a regression safety net against the currently-passing live state." T4's
own framing: "Local-scanner fallback for 'repo goes private again' (local SAST substitute +
scorecard-equivalent, auto-activating on `guard.outputs.public == 'false'`)."

## Source requirements

REL-02 (T1–T5).

## Current-state assessment

Per PLAN's own corrected evidence (to be re-confirmed at this story's own execution commit, since the
repository's visibility state has already drifted once mid-review and could drift again): Trivy runs
with `exit-code: 0` and `ignore-unfixed: true`, meaning it never fails the build regardless of findings.
CodeQL, Scorecard, and dependency-review are confirmed currently active (the repository is public) but
no regression meta-check exists to catch a future silent reversion to their skipped state. No waiver
mechanism exists for Trivy findings today (since Trivy never blocks, there is nothing to waive). No
local-scanner fallback exists for a hypothetical future private-repo state.

## Desired state

Trivy fails the build on CRITICAL/HIGH findings with an available fix, with `ignore-unfixed` scoped
only to a reviewed allowlist (not blanket-applied). A waiver-schema file format exists, CI-validated,
requiring owner/rationale/expiry/remediation-link per entry; a missing-field or expired entry fails CI.
A meta-check confirms `dependency-review`/`codeql`/`scorecard` genuinely ran (not merely configured to
run) whenever the repository is public, tested against a forced-private test branch to confirm the guard
logic itself, not just current visibility. A local-scanner fallback (local SAST substitute +
scorecard-equivalent) auto-activates on `guard.outputs.public == 'false'`, tested via a seeded unsafe
pattern caught by the fallback in a forced-private test branch. All of the above are wired into
W06-E03-S001's Wave-0 gate manifest.

## Scope

- **T1** — Flip Trivy to blocking (`exit-code: "1"`); scope `ignore-unfixed` to a reviewed allowlist
  only; run once report-only to baseline before flipping, per PLAN T1's own risk note (avoiding an
  immediate `main` break on latent findings).
- **T2** — Waiver mechanism: a reviewed allowlist file format with owner/rationale/expiry/remediation-
  link per entry, CI-validated; missing-field or expired entry fails.
- **T3** — Meta-check: assert `dependency-review`/`codeql`/`scorecard` actually ran whenever the
  repository is public, as a regression safety net; tested against a forced-private test branch to
  confirm the guard logic itself.
- **T4** — Local-scanner fallback for "repo goes private again" (local SAST substitute + scorecard-
  equivalent), auto-activating on `guard.outputs.public == 'false'`; seeded unsafe pattern caught in a
  forced-private test branch; documented coverage gap vs. CodeQL rather than claimed parity.
- **T5** — Wire all REL-02 blocking checks into W06-E03-S001's Wave-0 manifest.

## Out of scope

- **REL-01's own manifest schema/mechanics** — W06-E03-S001's own scope; this story's T5 only adds
  entries to that manifest, it does not build the manifest mechanism itself.
- **GHAS licensing** — PLAN's own evidence confirms "no GHAS license active (CodeQL/Scorecard run only
  because the repo is public, not because of GHAS)" — this story does not attempt to acquire or
  configure a GHAS license, it works within the existing public-repo-driven scanner activation model.

## Assumptions

- The repository's visibility state at this story's own execution commit must be re-confirmed (via the
  same live `gh api` calls PLAN's own pre-flight correction used) before trusting the "already active"
  baseline for CodeQL/Scorecard/dependency-review — this is a fail-first re-confirmation, consistent
  with this programme's convention applied elsewhere.
- The exact local-scanner fallback's coverage gap versus CodeQL is expected to be real and must be
  documented, not claimed as parity — PLAN T4's own risk note: "document coverage gap vs. CodeQL rather
  than claim parity."

## Dependencies

None within W06-E03 for T1-T4 (these target Trivy/CodeQL/Scorecard/dependency-review configuration,
disjoint from W06-E03-S001's own release-pipeline mechanics). **T5 depends on W06-E03-S001's T1/T2**
(the manifest schema and Wave-0 entries must exist before REL-02's checks can be wired into it).

## Affected packages or components

CI workflow configuration: the Trivy scanning workflow (exact file TBD, likely `.github/workflows/
security-scan.yml` per MATRIX CS-23's own line citations), a new waiver-schema file and its CI
validator, a new visibility-guard meta-check workflow step, a new local-scanner fallback workflow.

## Compatibility considerations

T1's Trivy flip is an intentional, compatibility-breaking improvement over the current soft-fail
posture: any latent CRITICAL/HIGH finding with an available fix that exists today would newly block the
build once flipped — PLAN T1's own risk note explicitly recommends "run once report-only to baseline
before flipping, or it can immediately break `main` on latent findings." This is a required
implementation-sequencing step, not optional.

## Security considerations

This entire story is a security-hardening story — see "Value to the framework" and "Problem statement"
above. T3's own regression-safety-net purpose is itself a defense against a silent security-posture
regression (the repository's own visibility flipping and the currently-passing scanners silently going
dormant without anyone noticing).

## Performance considerations

Not applicable — these are CI-time gates.

## Observability considerations

The waiver mechanism (T2) is itself an observability/audit artifact — every waived finding must be
traceable to an owner, rationale, expiry, and remediation link, not silently suppressed with no record.

## Migration considerations

Not applicable.

## Documentation requirements

Document the waiver-schema file format and how to add a waiver entry; document the local-scanner
fallback's documented coverage gap versus CodeQL, so a reader understands its limitations rather than
assuming parity.

## Acceptance criteria

- **AC-W06-E03-S003-01**: A seeded-vulnerability fixture proves Trivy fails then passes after removal (or after
  a properly-waived entry is added); `ignore-unfixed` is scoped to a reviewed allowlist, not
  blanket-applied.
- **AC-W06-E03-S003-02**: A waiver-schema fixture test confirms well-formed entries pass, and missing-field or
  expired entries fail CI.
- **AC-W06-E03-S003-03**: A forced-private test branch confirms the visibility-guard meta-check's own logic (not
  merely current visibility) — the meta-check genuinely asserts `dependency-review`/`codeql`/`scorecard`
  ran whenever the repository is public.
- **AC-W06-E03-S003-04**: A seeded unsafe-pattern fixture, run against a forced-private test branch, is caught
  by the local-scanner fallback; the fallback's documented coverage gap versus CodeQL is recorded, not
  claimed as parity.
- **AC-W06-E03-S003-05**: Every REL-02 blocking check (T1-T4) has exactly one manifest entry in W06-E03-S001's
  Wave-0 manifest, confirmed via a cross-reference test.

## Required artifacts

- The Trivy blocking-flip configuration (T1).
- The waiver-schema file format + CI validator (T2).
- The visibility-guard regression meta-check (T3).
- The local-scanner fallback workflow (T4).
- The manifest wiring for all four checks (T5).
See `artifacts/index.md`.

## Required evidence

- Seeded-vulnerability fail-then-pass test output (T1).
- Waiver-schema fixture test output, well-formed/missing-field/expired (T2).
- Forced-private-test-branch guard-regression test output (T3).
- Seeded-SAST-fixture fallback test output, forced-private branch (T4).
- Cross-reference test output confirming manifest wiring (T5).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all five acceptance criteria numbered and measurable, T5's dependency on
W06-E03-S001's T1/T2 recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all five acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T1's report-only baseline step was genuinely performed
before the blocking flip (not skipped to save time) and that T4's coverage-gap documentation is honest,
not a false parity claim.

## Risks

None recorded at this story's own scope beyond the general "flipping Trivy to blocking may immediately
break `main` on latent findings" risk, mitigated by T1's own report-only-baseline-first requirement —
this is a well-bounded, source-derived closure story with a clear MATRIX CS-23/PLAN REL-02 acceptance
bar.

## Residual-risk expectations

Once T1's report-only baseline step is honored and all five acceptance criteria are verified, residual
risk is expected to be low.

## Plan

See `plan.md`.
