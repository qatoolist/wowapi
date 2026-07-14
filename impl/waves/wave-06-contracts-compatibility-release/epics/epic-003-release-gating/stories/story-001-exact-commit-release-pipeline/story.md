---
id: W06-E03-S001
type: story
title: Exact-commit release pipeline — REL-01 T1-T8 buildable-now set
status: verified
wave: W06
epic: W06-E03
owner: W06E03Impl
reviewer: independent-review-gate
priority: P0
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - REL-01
depends_on: []
blocks:
  - W06-E03-S002
acceptance_criteria:
  - AC-W06-E03-S001-01
  - AC-W06-E03-S001-02
  - AC-W06-E03-S001-03
  - AC-W06-E03-S001-04
  - AC-W06-E03-S001-05
  - AC-W06-E03-S001-06
  - AC-W06-E03-S001-07
  - AC-W06-E03-S001-08
artifacts:
  - ART-W06-E03-S001-001
  - ART-W06-E03-S001-002
  - ART-W06-E03-S001-003
  - ART-W06-E03-S001-004
  - ART-W06-E03-S001-005
  - ART-W06-E03-S001-006
  - ART-W06-E03-S001-007
  - ART-W06-E03-S001-008
  - ART-W06-E03-S001-009
evidence:
  - EV-W06-E03-S001-001
  - EV-W06-E03-S001-002
  - EV-W06-E03-S001-003
  - EV-W06-E03-S001-004
  - EV-W06-E03-S001-005
  - EV-W06-E03-S001-006
  - EV-W06-E03-S001-007
  - EV-W06-E03-S001-008
  - EV-W06-E03-S001-009
decisions:
  - ADR-W00-E02-S003-005
risks:
  - RISK-W06-E03-001
---

# W06-E03-S001 — Exact-commit release pipeline — REL-01 T1-T8 buildable-now set

## Story ID

W06-E03-S001

## Title

Exact-commit release pipeline — REL-01 T1-T8 buildable-now set

## Objective

Build the ~85%-buildable-now REL-01 release pipeline: the gate-manifest schema and validator, the
Wave-0 manifest entries, `required-gates.yml`, the `ci.yml` wiring, a `verify` job in `release.yml`, the
`build-candidate`/`publish` split via GoReleaser `--skip=publish` per D-05 (`ADR-W00-E02-S003-005`,
consumed by reference — not re-decided here), `verify_release.sh` with
golden-failure tests, and a `verify-published` job — all authorable and testable against a scratch or
throwaway repository, without needing the protected GitHub Environment that REL-01 T9/DEC-Q10 requires.

## Value to the framework

REVIEW §G's own verdict is the direct source of this story's value proposition: "REL-01/REL-02 are
**~85% implementable and fully testable now**; only the last-mile *activation* of branch/env/tag
protection is genuinely human. Do not classify the whole workstream blocked." Without this split, the
entire release-gating finding would sit un-started, waiting on a human action a coding agent cannot
perform, when in fact the overwhelming majority of the actual trust-boundary logic (manifest schema,
exact-SHA gating, tamper detection, artifact verification) can be built and proven today. This story
converts "the release process is a single atomic `goreleaser release --clean` step with no independent
verification" into "the release process is gated on the exact commit being published, with tamper
detection at every stage" — everything short of the final admin-only lock being turned.

## Problem statement

PLAN's own REL-01 evidence: "**Confirmed hard blockers requiring GitHub repo-admin action** (verified via
`gh api`, not assumed): `gh api repos/qatoolist/wowapi/branches/main/protection` → 404, no branch
protection exists. `gh api repos/qatoolist/wowapi/environments` → `{"total_count":0}`, no GitHub
Environments exist." But PLAN's own directive-selected design is explicit that the pipeline mechanics
themselves do not require those admin actions to exist first: "(1) a reusable `required-gates.yml`
(`workflow_call`) runs a versioned `ci/release-gates.yaml` manifest against an exact SHA; (2) both
PR/main CI and release call it... emitting an attested `gate-results.json`; (3) `build-candidate` (no
publish permissions) verifies gate results, builds once, emits artifacts + `release-manifest.json`,
never pushes; (4) `publish` (protected `release` environment, the only job with write permissions)
copies exactly the manifested bytes, never rebuilds; (5) `verify-published` re-verifies everything from
a clean runner post-publish." PLAN's own "Machine acceptance" floor: "a deliberately failing check
prevents `build-candidate`; changing the tag target changes both manifest SHAs; tampering with gate
results or candidate bytes is detected; publish rejects any artifact/digest absent from the manifest;
post-publish verification succeeds from a clean runner with no build workspace."

## Source requirements

REL-01 (T1–T8; T9's own protection-activation remainder is W06-E03-S002's scope).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit): no `ci/release-
gates.yaml` manifest exists. No `required-gates.yml` reusable workflow exists. No `verify` job in
`release.yml` checks the tag's exact target SHA independently. Today's `goreleaser release --clean` is a
single atomic build-and-publish step with no independent build-candidate/publish separation and no
post-publish clean-runner re-verification. No `scripts/validation/verify_release.sh` exists (§13.1
explicitly requires one: "a prose checklist is not sufficient").

## Desired state

A versioned `ci/release-gates.yaml` manifest (ID, command/job ref, owner, `required_from_wave`, timeout,
evidence-artifact path per entry) validated by a JSON Schema validator that rejects a malformed entry.
Wave-0 manifest entries mapping to today's `workflow-lint`/`unit`/`gate`/`coverage`/`reference-smoke`
jobs plus `vuln.yml` and REL-02's blocking scanners, with every job currently required for green
`ci-container` represented — none silently dropped. A `required-gates.yml` reusable workflow, called
identically by both PR/main CI and release, emitting an attested `gate-results.json`, checked out at
the exact SHA (never branch HEAD). A `verify` job in `release.yml` that never trusts a same-named check
on another ref, proven via a seeded-failure fixture: tag a commit with a deliberately broken test, prove
`verify` fails and `build-candidate` never runs. A `build-candidate` job (permissions:
`contents:read`, `id-token:write`, `attestations:write` only — no write) using GoReleaser
`release --skip=publish` per ADR-005, emitting archives/checksums/SBOMs/OCI layout plus an attested
`release-manifest.json`, proven via a tamper test (hand-edit one artifact byte, prove mismatch
detected). A `scripts/validation/verify_release.sh <version> <source-sha>` with golden failure tests,
one per verified property (wrong SHA, stripped signature, missing SBOM attestation, wrong platforms,
tampered manifest hash). A `verify-published` job invoking that script on a clean runner, proven via an
end-to-end dry run against a disposable throwaway repo. SLSA 1.2 guarantee documentation that states
exactly which build-track requirements are met, with no over-claim.

## Scope

- **T1** — Design `ci/release-gates.yaml` manifest schema + JSON Schema validator; schema rejects a
  manifest entry missing a required field.
- **T2** — Populate Wave-0 manifest entries mapping to today's `workflow-lint`/`unit`/`gate`/
  `coverage`/`reference-smoke` jobs + `vuln.yml` + REL-02's blocking scanners; every job currently
  required for green `ci-container` has a manifest entry.
- **T3** — Build `required-gates.yml` (`workflow_call`, parameterized on SHA), emitting attested
  `gate-results.json`.
- **T4** — Update `ci.yml` to call `required-gates.yml`, so PR CI and release use the identical
  execution path.
- **T5** — Add a `verify` job to `release.yml` calling `required-gates.yml` with the tag event's exact
  SHA; never trusting a same-named check on another ref; seeded-failure fixture proving `verify` fails
  and `build-candidate` never runs.
- **T6** — Split into `build-candidate` (no write permissions) using GoReleaser `--skip=publish` per
  ADR-005, emitting attested `release-manifest.json`; tamper test proving mismatch detection.
- **T7** — Add a `publish` job (`needs: build-candidate`, protected `release` environment) copying only
  manifested artifacts; unmanifested-artifact test proving rejection (this task's own scaffolding can be
  built and tested against a stub environment; it cannot be end-to-end proven without the real protected
  environment, which is W06-E03-S002's own scope).
- **T8** — Write `scripts/validation/verify_release.sh <version> <source-sha>`; golden failure tests, one
  per verified property.
- **T9's evidence-generation half only** — a `verify-published` job invoking T8 on a clean runner, with
  an end-to-end dry run against a disposable throwaway repo (T9's own PLAN acceptance criterion:
  "Corrupted publish (in a disposable test repo) caught, `latest` not moved" — this is fully buildable
  and testable against a scratch repo, distinct from T9's dependency on the real protected environment
  which W06-E03-S002 handles).

## Out of scope

- **T9's dependency on the real protected `release` GitHub Environment existing** — S002's own scope;
  this story's T7/T9 work builds and tests the mechanics against a scratch/throwaway repo and a stub
  environment, it does not perform the actual GitHub Environment creation.
- **REL-01 T9's own branch/tag protection remainder** — S002's scope entirely.
- **REL-02's own blocking-scanner implementation** — S003's scope; this story's T2 only references
  REL-02's blocking scanners as manifest entries once they exist, per T2's own dependency framing.

## Assumptions

- ADR-005's own unresolved caveat ("verify against the pinned GoReleaser version at implementation
  time... not yet independently confirmed") is inherited by this story's T6, not re-decided here — see
  RISK-W06-E03-001 (epic-scoped).
- T7's "stub environment" testing approach (since the real protected environment does not yet exist) is
  not specified by any source document beyond REVIEW §G's own framing ("publish job runs unprotected in
  scratch until set") — the exact stub mechanism (a scratch repo's own unprotected environment, a local
  mock) is recorded as an implementation-time decision.
- The exact manifest-entry field list beyond what PLAN T1's own acceptance criterion names (ID,
  command/job ref, owner, `required_from_wave`, timeout, evidence-artifact path) is confirmed from
  source, not invented.

## Dependencies

Depends on W00-E02-S003's ADR-ification of D-05 (`ADR-W00-E02-S003-005`, already ratified) for T6. No dependency within W06-E03 for this
story's own entry (it is the epic's foundational story). Blocks W06-E03-S002 (the protection-activation
story practically requires this pipeline's mechanics to exist before the final admin lock has anything
to protect, though S002's own front matter records its dependency as DEC-Q10, not this story, per its
own human-gated framing).

## Affected packages or components

New: `ci/release-gates.yaml`; a JSON Schema validator (exact location TBD); `.github/workflows/
required-gates.yml`; `scripts/validation/verify_release.sh`. Extended: `.github/workflows/ci.yml`,
`.github/workflows/release.yml`, the GoReleaser configuration (`.goreleaser.yaml` or equivalent).

## Compatibility considerations

T4's change to `ci.yml` (calling `required-gates.yml`) must not regress PR CI latency, per PLAN T4's own
risk note. This is a required non-functional constraint on this story's own implementation, not an
optional nicety.

## Security considerations

This entire story is a security-boundary-hardening story: T6's `build-candidate` job's permission
scoping (no write access) is itself a required security control, not an implementation detail — PLAN's
own risk note: "Job's token literally cannot push/release (permission-scoped, not conventional)." T7's
manifested-artifact-only copying is likewise a required control against supply-chain tampering.

## Performance considerations

T4's own constraint (no PR CI latency regression) is this story's one performance-adjacent concern; no
other performance budget applies.

## Observability considerations

`gate-results.json` (T3) and `release-manifest.json` (T6) are themselves observability/audit artifacts
— they must be attested and inspectable, not merely produced silently.

## Migration considerations

Not applicable — no schema or data migration is involved.

## Documentation requirements

T8's own SLSA-guarantee documentation (T10 in PLAN's own numbering, folded into this story's scope as
the final documentation step) must state exactly which build-track requirements T6/T7's builder meets,
with no false claim — PLAN's own risk note: "a false SLSA claim is itself a supply-chain trust defect."

## Acceptance criteria

- **AC-W06-E03-S001-01**: The manifest schema validator rejects a manifest entry missing a required field.
- **AC-W06-E03-S001-02**: Every job currently required for green `ci-container` has a Wave-0 manifest entry;
  none is silently dropped.
- **AC-W06-E03-S001-03**: `required-gates.yml`, called with a failing entry, attests failure in `gate-
  results.json`, with each entry individually reported.
- **AC-W06-E03-S001-04**: The same SHA through both the PR/main CI path and the release path produces
  byte-identical results (excluding run ID/timestamp).
- **AC-W06-E03-S001-05**: A seeded-failure fixture (tag a commit with a deliberately broken test) proves
  `verify` fails and `build-candidate` never runs.
- **AC-W06-E03-S001-06**: A tamper test (hand-edit one artifact byte) proves `build-candidate`'s mismatch
  detection; the job's own token cannot push or release (permission-scoped).
- **AC-W06-E03-S001-07**: Golden failure tests, one per verified property (wrong SHA, stripped signature,
  missing SBOM attestation, wrong platforms, tampered manifest hash), all pass against
  `verify_release.sh`.
- **AC-W06-E03-S001-08**: An end-to-end dry run against a disposable throwaway repo catches a corrupted publish
  and confirms `latest` is not moved; SLSA-guarantee documentation states exactly which build-track
  requirements are met with no over-claim.

## Required artifacts

- `ci/release-gates.yaml` manifest schema + validator (T1).
- Wave-0 manifest entries (T2).
- `required-gates.yml` (T3).
- `ci.yml` wiring (T4).
- `release.yml`'s `verify` job (T5).
- `build-candidate` job (T6).
- `publish` job scaffolding tested against a stub environment (T7).
- `verify_release.sh` (T8).
- `verify-published` job + SLSA documentation (T9's buildable half).
See `artifacts/index.md`.

## Required evidence

- Malformed-manifest-fixture test output (T1).
- Manifest-entry-count diff-review output (T2).
- Seeded-failure gate-results attestation output (T3).
- Diff-based same-SHA-both-paths test output (T4).
- Seeded-failure tag test output (T5).
- Tamper-test output (T6).
- Golden failure test output, one per verified property (T8).
- End-to-end dry-run output against a disposable repo (T9's buildable half).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all eight acceptance criteria numbered and measurable, dependency on ADR-005
recorded, owner/reviewer assignment pending, T7's stub-environment testing approach recorded as an
unresolved question rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all eight acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming this story's own scope boundary is honored — T7/T9's
own acceptance criteria are proven against a scratch/stub environment, and no claim is made that the
real protected-environment end-to-end path (W06-E03-S002's scope) has been proven by this story.

## Risks

RISK-W06-E03-001 (ADR-005's GoReleaser split-mode caveat surfacing a real incompatibility only at T6's
implementation time) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once T6's version-confirmation step and the full T1-T8 acceptance criteria are verified, residual risk
is expected to be low for this story's own buildable-now scope — the genuinely irreducible risk (DEC-Q10
human activation) is explicitly out of this story's scope and belongs to W06-E03-S002.

## Plan

See `plan.md`.
