---
id: PLAN-W06-E03-S003
type: plan
parent_story: W06-E03-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E03-S003

Per mandate §8.5. Confirmed facts, planned changes, and assumptions are distinguished explicitly below.

## Proposed architecture

Four independent security-scan-hardening mechanisms (Trivy blocking flip, waiver schema, visibility-
guard meta-check, local-scanner fallback), each wired into CI as its own configuration change or new
workflow step, plus a fifth task wiring all four into W06-E03-S001's Wave-0 manifest.

## Implementation strategy

1. **T1** — Run Trivy report-only first to baseline current findings; then flip to `exit-code: "1"`,
   scoping `ignore-unfixed` to a reviewed allowlist; write a seeded-vulnerability fixture proving
   fail-then-pass.
2. **T2** — Design and implement the waiver-schema file format (owner/rationale/expiry/remediation-link
   per entry); write a CI validator; write well-formed/missing-field/expired fixture tests.
3. **T3** — Implement the visibility-guard meta-check asserting `dependency-review`/`codeql`/`scorecard`
   ran whenever the repository is public; test against a forced-private test branch to confirm the guard
   logic itself.
4. **T4** — Implement the local-scanner fallback (local SAST substitute + scorecard-equivalent),
   auto-activating on `guard.outputs.public == 'false'`; write a seeded unsafe-pattern fixture tested
   against a forced-private test branch; document the coverage gap versus CodeQL.
5. **T5** — Wire all four checks into W06-E03-S001's Wave-0 manifest, with a cross-reference test
   confirming exactly one manifest entry per enumerated scanner class.

## Expected package or module changes

CI workflow configuration changes only: the Trivy scanning workflow, a new waiver-schema file + CI
validator, a new visibility-guard meta-check step, a new local-scanner fallback workflow.

## Expected file changes where determinable

- The Trivy scanning workflow configuration (likely `.github/workflows/security-scan.yml`, per MATRIX
  CS-23's own line citations).
- A new waiver-schema file (exact path TBD) and its CI validator.
- A new visibility-guard meta-check workflow step.
- A new local-scanner fallback workflow.
- W06-E03-S001's `ci/release-gates.yaml` (extended with REL-02's manifest entries).

## Contracts and interfaces

The waiver-schema file's own format (owner/rationale/expiry/remediation-link per entry) is the primary
new contract.

## Data structures

The waiver-schema entry structure itself.

## APIs

None affected — CI-configuration-internal.

## Configuration changes

Trivy's own scan configuration (`exit-code`, `ignore-unfixed` scoping); no application configuration
change.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

Each gate must fail with a clear, specific error identifying the exact finding/entry that caused the
failure.

## Security controls

This entire story is itself a set of security controls — see `story.md` "Security considerations."

## Observability changes

The waiver mechanism (T2) is itself the primary observability/audit artifact this story adds.

## Testing strategy

- T1: seeded-vulnerability fixture, fail-then-pass-after-removal (or after a properly-waived entry).
- T2: well-formed/missing-field/expired waiver-schema fixture tests.
- T3: forced-private test branch confirming the guard logic itself.
- T4: seeded unsafe-pattern fixture caught by the fallback, forced-private test branch.
- T5: cross-reference test confirming exactly one manifest entry per scanner class.

## Regression strategy

Once wired into CI, all four checks become ongoing regression guards for their respective security
classes; T3 specifically is itself a regression guard against a silent visibility-driven scanner
dormancy.

## Compatibility strategy

T1's flip is an intentional, compatibility-breaking improvement, sequenced with a report-only baseline
step first per `story.md` "Compatibility considerations."

## Rollout strategy

T1's report-only baseline must land before the blocking flip; T2-T4 may proceed independently; T5 lands
once T1-T4 exist and W06-E03-S001's manifest schema is available.

## Rollback strategy

If T1's blocking flip produces excessive false positives, temporarily widen the reviewed
`ignore-unfixed` allowlist (via T2's own waiver mechanism, with rationale recorded) rather than
reverting to the unscoped soft-fail posture.

## Implementation sequence

T1 (report-only baseline, then flip) → T2 (waiver mechanism, needed to manage T1's own findings) → T3
and T4 may proceed independently of T1/T2 → T5 (manifest wiring, requires W06-E03-S001's schema).

## Task breakdown

- **W06-E03-S003-T001** — Trivy blocking flip with reviewed allowlist scoping (T1).
- **W06-E03-S003-T002** — Waiver mechanism (T2).
- **W06-E03-S003-T003** — Visibility-guard regression meta-check (T3).
- **W06-E03-S003-T004** — Local-scanner fallback (T4).
- **W06-E03-S003-T005** — Manifest wiring (T5).
- **W06-E03-S003-T006** — Independent review.

## Expected artifacts

The Trivy blocking-flip configuration; the waiver-schema file format + CI validator; the visibility-
guard regression meta-check; the local-scanner fallback workflow; the manifest wiring.

## Expected evidence

Seeded-vulnerability fail-then-pass test output; waiver-schema fixture test output; forced-private-test-
branch guard-regression test output; seeded-SAST-fixture fallback test output; cross-reference test
output.

## Unresolved questions

- Exact waiver-schema file path and format details beyond the required fields (owner/rationale/expiry/
  remediation-link) — to be decided at implementation time.
- Exact local-scanner substitute tool choice for the SAST fallback — not specified by any source
  document beyond "local SAST substitute + scorecard-equivalent."

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned; T1's report-only
baseline step must be performed and its results reviewed before the blocking flip itself is approved to
proceed.
