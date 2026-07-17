# Sixth review — Comprehensive baseline/release-readiness review (2026-07-17)

Source: ComprehensiveReport.md (untracked, user-owned). Verdict: APPROVED as a
remediation/baseline work order; NO-GO for release or clean-V1 cutover at
8b89412. Four classes of work: (1) open correctness blockers C-01..C-11;
(2) a release-identity decision; (3) V1-compatibility residue cleanup; (4)
disconnected/incomplete V2 surfaces.

## Verified against source before acting

Spot-checked and confirmed real: C-01 (registry invalidated before the sealed
check), C-04 (all resolvers in a step share one context copy), C-05
(newArtifactWriter falls back to retention.TestKey() on missing/invalid key),
C-08 (Dockerfile/devbox ldflags used the pre-/v2 buildinfo path), C-06
(compatibility-gates.yml + ci.yml baseline default v1.0.0).

## Remediation this round (commit 7): self-contained correctness blockers

Fixed the code-level defects that hold regardless of the release-identity
decision, each with a discriminating regression:

- C-01: registration invalidates validation ONLY on a successful mutation —
  a rejected duplicate/invalid registration and a recovered post-seal panic
  leave generation, contents, and validation unchanged
  (TestRejectedMutationsDoNotInvalidate).
- C-02: resolveValidated checks validation AND resolves the graph under one
  read lock — the executed graph is covered by a single validated generation;
  defForInstance uses it (TestResolveValidatedIsAtomicWithValidation).
- C-04: each assignee resolver invocation receives its OWN deep canonical
  context copy (TestResolversEachGetIsolatedContext).
- C-05: a missing/malformed DSR artifact key FAILS BOOT in production instead
  of silently using the deterministic test key; non-prod keeps the warned
  convenience (TestArtifactWriterFailsClosedInProd).
- C-08: Dockerfile and devbox entrypoint ldflags stamp the /v2 buildinfo
  path (.goreleaser.yaml already correct from the cutover).

Gates: full host suite, ci-container, mechanical batch green.

## NOT done autonomously — requires the user's decision / larger programme

- **C-03 (High)** persisted-definition identity: the smallest safe fix
  (persist+verify a definition digest, or execute the persisted snapshot) is a
  SCHEMA change that the report couples to the migration-baseline decision.
  Deferred pending that decision; the round-5 parseAndValidateDefinition
  already validates persisted graphs against the current callback sets, so a
  persisted def can never execute unvalidated — the residual gap is
  registry-vs-persisted divergence for the same (key,version), which needs the
  digest column.
- **Release identity (section 5, blocking):** v1.1.0 is published in the Go
  proxy and cannot be reused; v1.0.0 is reserved pending proxy preflight. The
  choice — root-module higher-V1, a new module path for a fresh v1.0.0, or
  keep /v2 — is the user's and must precede any second module-path rewrite,
  or the repo is rewritten twice.
- **CI single-owner redesign (C-06/09/10/11):** the required-gates matrix runs
  alongside duplicate native legs; release-gate evidence declarations don't
  bind real artifacts; exact-tag compatibility wiring is absent. A CI/release
  workflow programme, best done after the identity decision.
- **V1-residue cleanup (L-01..L-18, D-01..D-11, M-*, R-*):** forwarding
  shims, parallel constructors, unsigned cursors, ignored claim fields,
  markerless generator rewrite, migration squash, history archive. Many gated
  on product decisions (MFA scope, checksum-repair scope, clean-DB policy) and
  all coupled to the identity decision. A staged programme, not an autonomous
  sweep — the report's own phase 1 is "decide the release identity."
