---
id: VER-W06-E03-S001
type: verification-record
parent_story: W06-E03-S001
status: verified
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W06-E03-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E03-S001-01 | Run the manifest validator against a malformed-entry fixture | Local dev or CI | Malformed entry rejected | unit-test report | unassigned |
| AC-W06-E03-S001-02 | Diff-review the manifest entries against the current green-ci-container required-check set | CI config review | Every required check has a manifest entry | diff-review report | unassigned |
| AC-W06-E03-S001-03 | Run required-gates.yml with a seeded failing entry | CI | gate-results.json attests failure, per-entry | workflow test report | unassigned |
| AC-W06-E03-S001-04 | Run the same SHA through both PR/main CI and release paths, diff the results | CI | Byte-identical results excluding run ID/timestamp | diff-based test report | unassigned |
| AC-W06-E03-S001-05 | Tag a commit with a deliberately broken test, run the pipeline | Scratch/throwaway repo | verify fails, build-candidate never runs | seeded-failure fixture report | unassigned |
| AC-W06-E03-S001-06 | Hand-edit one artifact byte, run build-candidate's verification | Scratch/throwaway repo | Mismatch detected; job token cannot push/release | tamper-test report | unassigned |
| AC-W06-E03-S001-07 | Run each golden failure test against verify_release.sh | Local dev or CI | Each verified property's failure is caught | golden-failure test report | unassigned |
| AC-W06-E03-S001-08 | Run an end-to-end dry run against a disposable throwaway repo; inspect SLSA documentation | Disposable throwaway repo | Corrupted publish caught, latest not moved; SLSA doc states exact guarantees with no over-claim | end-to-end dry-run report + doc review | unassigned |

## Post-execution record

Focused verification was executed against scratch directories, temporary Git repositories, and a
temporary immutable registry. Real protected-environment activation is deliberately excluded (S002).

### Actual result

All 10 release-contract tests passed; actionlint reported no diagnostics for all affected workflows. The tests observed malformed schema, failing exact-SHA gate, moved tag, altered gate/manifest/archive/image/candidate bytes, missing security reports, unmanifested input, clean verification golden failures, and immutable promotion behavior.

### Pass or fail

Pass for S001's authorable/scratch scope.

### Evidence identifier

EV-W06-E03-S001-001 through EV-W06-E03-S001-009; see `evidence/index.md`.

### Execution date

2026-07-13.

### Commit or revision

`733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus the dirty shared W05/W06 workspace.

### Environment

Local Darwin runner; temporary repositories/registries; workflow syntax validation. No live release or protected environment was claimed.

### Reviewer

Implementer verification; independent review recorded separately by T009.

### Findings

ADR-005's OSS publisher command was unavailable. The authorized exact-byte `gh`/ORAS deviation and stronger draft-first/artifact-scan controls are recorded in `deviations.md`.

### Retest status

Focused suites re-run after candidate artifact security-report enforcement and clean manifest-attestation wiring: pass.

### Final conclusion

S001 satisfies its buildable-now acceptance scope. Live protected-environment proof remains correctly blocked under S002/DEC-Q10.
