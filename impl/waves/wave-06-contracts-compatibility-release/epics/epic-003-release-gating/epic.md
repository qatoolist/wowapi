---
id: W06-E03
type: epic
title: Release gating
status: partially-verified
wave: W06
owner: W06E03Impl
reviewer: W06-E01-E04-Execution.W06E03ReviewR
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - REL-01
  - REL-02
depends_on: []
stories:
  - W06-E03-S001
  - W06-E03-S002
  - W06-E03-S003
decisions: []
risks:
  - RISK-W06-001
---

# W06-E03 — Release gating

## Epic objective

Build the ~85%-buildable-now release-gating pipeline that gates a published release on the exact commit
being tested (REL-01 T1–T8), track its final admin-only activation (branch protection, protected release
Environment, tag protection ruleset — DEC-Q10) as a distinct, explicitly human-gated story, and make
security scanning blocking instead of soft-failing (REL-02).

## Problem being solved

`requirement-inventory.md` row REL-01 states: "Release gated on exact published commit (T1-T9) | IMPL |
P0 | planned | W06-E03-S001..S002 | ~85% buildable now; final activation = DEC-Q10 (admin)." Row REL-02
states: "Security checks blocking-or-replaced | QG | P0/P1 | planned | W06-E03-S003 | Trivy soft-fail +
visibility-guard dormancy cited at exact lines (MATRIX CS-23)." REVIEW §G's own blocker-resolution table
is the authoritative source for this epic's own two-story split within REL-01: "Workflow authoring
(`required-gates.yml`, gate manifest, `verify_release.sh` + golden-failure tests) | **No** — author +
unit-test now"; "Local validation (exact-SHA gate logic, tamper fixtures) | **No**"; "Blocking-scanner
rollout (Trivy `exit-code:1`, waiver schema) | **No** — code now"; but "Branch protection on `main` |
**Enforcement only** | repo admin (human)"; "Protected `release` GitHub Environment | **Rollout only** |
repo admin (human)"; "Tag protection ruleset | **Rollout only** | repo admin (human)." REVIEW §G's own
verdict: "REL-01/REL-02 are **~85% implementable and fully testable now**; only the last-mile
*activation* of branch/env/tag protection is genuinely human. Do not classify the whole workstream
blocked."

## Scope

- REL-01 T1–T8: manifest schema design + JSON Schema validator; Wave-0 manifest entries; `required-
  gates.yml` (`workflow_call`, parameterized on SHA) emitting attested `gate-results.json`; `ci.yml`
  calling `required-gates.yml`; a `verify` job in `release.yml` with a seeded-failure fixture proving
  `verify` fails and `build-candidate` never runs; the `build-candidate`/`publish` split via GoReleaser
  `--skip=publish` per ADR-005; `scripts/validation/verify_release.sh` with golden failure tests, one
  per verified property; a `verify-published` job with an end-to-end dry run against a disposable
  throwaway repo; SLSA-guarantee documentation (S001).
- REL-01 T9 (the branch/tag/environment protection remainder) and DEC-Q10 (S002 — human-gated).
- REL-02 T1–T5: Trivy blocking flip with a reviewed waiver allowlist; the waiver-schema CI validator; a
  visibility-guard regression meta-check; a private-repo local-scanner fallback; wiring into REL-01's
  manifest (S003).

## Out of scope

- **Actually performing the DEC-Q10 repo-admin action** — no coding agent can create a protected GitHub
  Environment, set branch protection, or configure a tag protection ruleset; this remains genuinely
  human, tracked (not performed) by S002.
- **REL-03's own compatibility-gate content** — W06-E02's scope; this epic's S001 folds in REL-03 T9's
  SBOM/provenance-verify naming as shared evidence, it does not itself build REL-03's gates.
- **AR-05's documentation gates** — W06-E04's scope.

## Source requirements

REL-01, REL-02. MATRIX CS-15 and REVIEW §F/§G are the primary consolidated sources; ADR-005
(`ADR-W00-E02-S003-005`, already ratified at W00) governs T6's GoReleaser split-mode implementation.

## Architectural context

This epic groups REL-01 and REL-02 because both are the framework's release-time trust boundary — one
gates *what* gets published and proves it matches the exact tested commit (REL-01), the other gates
*whether the published artifact is known-safe* (REL-02) — and both share the same gate-manifest
infrastructure (REL-01's `ci/release-gates.yaml`, which REL-02 T5 wires its own blocking checks into).
`impl/analysis/wave-allocation-detail.md`'s own W06-E03 grouping states this exactly: "S001
exact-commit-release-pipeline (REL-01 T1-T8 buildable set); S002 protection-activation (REL-01 remainder
+ DEC-Q10 — human-gated story, explicit blocked status allowed); S003 blocking-security-scans (REL-02:
Trivy exit-code flip, waiver schema, visibility-guard review)." This three-way split — buildable-now
pipeline, human-gated activation, and blocking-scan closure — is fixed by the canonical allocation and
mirrors REVIEW §G's own layer-by-layer authorable-vs-admin-only distinction exactly.

## Included stories

- **W06-E03-S001 — exact-commit-release-pipeline** (PLAN REL-01 T1–T8): the buildable-now, fully
  testable-against-a-scratch-repo release pipeline.
- **W06-E03-S002 — protection-activation** (PLAN REL-01 T9 remainder + DEC-Q10): the human-gated final
  activation, explicitly blocked-entry until a repo administrator acts.
- **W06-E03-S003 — blocking-security-scans** (PLAN REL-02 T1–T5): Trivy blocking flip, waiver mechanism,
  visibility-guard regression meta-check, private-repo fallback, manifest wiring.

## Dependencies

No dependency on any other W06 epic — this epic's release-gating scope is disjoint from W06-E01
(consumer/DSL) and W06-E02 (contract gates), except that W06-E02-S002's own T005 (container architecture
smoke) depends on this epic's S001 (the `build-candidate` split producing the candidate image it
smoke-tests against) and W06-E02-S002's T006 shares evidence with this epic's S001 T8/T9. This epic
depends on W00-E02-S003 (ADR-005, already ratified) for S001's T6, and transitively on this wave's own
W05 entry gate.

## Risks

RISK-W06-001 (DEC-Q10's human-gated activation blocking S002's own closure indefinitely absent
repo-admin action) originates at wave scope and lands entirely within this epic's S002. See `risks.md`
for the epic-scoped elaboration.

## Required decisions

None new in the D-0N sense — ADR-005 is already ratified at W00 and consumed, not re-decided, by S001.
DEC-Q10 itself is not a D-0N architecture decision but a tracked human-blocked operational decision per
`requirement-inventory.md` §B; it does not warrant a `decisions/` directory under S002 (it is a
repo-admin action to be performed, not a design decision to be recorded as an ADR).

## Epic acceptance criteria

- **AC-W06-E03-01**: REL-01's machine-acceptance floor is satisfied for T1–T8: a deliberately failing
  check prevents `build-candidate`; changing the tag target changes both manifest SHAs; tampering with
  gate results or candidate bytes is detected; publish rejects any artifact/digest absent from the
  manifest; post-publish verification succeeds from a clean runner with no build workspace — all proven
  against a scratch/throwaway repo.
- **AC-W06-E03-02**: S002 is correctly recorded as blocked-entry on DEC-Q10, with no false "done" claim
  for REL-01 as a whole while the admin-only activation remains unperformed.
- **AC-W06-E03-03**: REL-02's Trivy blocking flip, waiver mechanism, visibility-guard regression
  meta-check, and private-repo fallback are complete, evidenced, and wired into REL-01's manifest.
- **AC-W06-E03-04**: S001 and S003 have passed independent review per mandate §14; S002's review (once
  DEC-Q10 resolves and the remaining activation work is done) specifically confirms the activation
  genuinely required repo-admin action and was not silently bypassed.

## Closure conditions

S001 and S003 reach `accepted`; S002 remains `planned`/`blocked` honestly until DEC-Q10 resolves — this
epic may close with S002 in a documented blocked state, per `governance/definition-of-done.md`'s
partially-accepted framing at epic scope, PROVIDED the blocked state is explicitly recorded with DEC-Q10
restated as the exact unblocking condition, not silently presented as complete.
