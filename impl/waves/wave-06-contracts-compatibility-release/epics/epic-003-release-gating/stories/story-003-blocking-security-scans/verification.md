---
id: VER-W06-E03-S003
type: verification-record
parent_story: W06-E03-S003
status: verified
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W06-E03-S003

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E03-S003-01 | Run Trivy against a seeded-vulnerability fixture, before and after removal/waiver | CI | Fails with the vulnerability present, passes after removal or a properly-waived entry | seeded-vuln fixture report | unassigned |
| AC-W06-E03-S003-02 | Run the waiver-schema validator against well-formed, missing-field, and expired fixtures | CI | Well-formed passes; missing-field and expired both fail | fixture test report | unassigned |
| AC-W06-E03-S003-03 | Run the visibility-guard meta-check against a forced-private test branch | CI, forced-private test branch | The guard logic itself is confirmed, not just current visibility | test report | unassigned |
| AC-W06-E03-S003-04 | Run a seeded unsafe-pattern fixture against a forced-private test branch | CI, forced-private test branch | The fallback catches the seeded pattern; coverage gap vs. CodeQL is documented | seeded SAST fixture report | unassigned |
| AC-W06-E03-S003-05 | Cross-reference test confirming exactly one manifest entry per REL-02 scanner class | CI config review | Every enumerated scanner class has exactly one manifest entry | cross-reference test report | unassigned |

## Post-execution record

Focused local and workflow-contract verification was executed; raw outputs are registered in
`evidence/index.md`.

### Actual result

Eight security-contract tests passed; the seeded lodash vulnerability failed before removal and
passed afterward; and the complete private fallback passed local SAST, repository posture, actionlint,
govulncheck, and Trivy. The live hosted-scanner exact-SHA meta-check also passed for remote `main`
revision `0a31186cada5c275a588c74081cf977adf346e61`.

### Pass or fail

Pass.

### Evidence identifier

EV-W06-E03-S003-001 through EV-W06-E03-S003-005.

### Execution date

2026-07-13; sequential retest 2026-07-14.

### Commit or revision

`733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus dirty shared W05/W06 changes.

### Environment

Local Darwin runner, real Trivy v0.72.0 database, temporary seeded fixtures, forced visibility inputs,
and the live GitHub hosted-scanner run index for remote `main`.

### Reviewer

Implementer verification; independent review recorded by T006.

### Findings

Trivy baseline found only the scoped Dockerfile `AVD-DS-0002` exception, synchronized to an active reviewed waiver. No blanket ignore exists.

### Retest status

Sequential rerun passed: 8/8 focused tests; seeded Trivy reject-then-accept; full private fallback;
and live hosted-scanner meta-check for remote `main`.

### Final conclusion

All five REL-02 acceptance criteria are satisfied in the authorable scope.
