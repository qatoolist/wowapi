---
id: VER-W06-E02-S003
type: verification-record
parent_story: W06-E02-S003
status: blocked
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W06-E02-S003

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S003-01 | Once W06-E02-S001 is accepted, run the OpenAPI semantic-diff gate against a seeded breaking-OpenAPI fixture | CI | The breaking fixture fails the gate, classified per DX-06's 3.1/2020-12 baseline | CI gate test report | unassigned |
| AC-W06-E02-S003-02 | Once W06-E01-S001 and W05-E03 are both accepted, run the event/schema compatibility check against a seeded breaking-event fixture | CI | The breaking fixture fails when the compatibility mode is declared | CI gate test report | unassigned |
| AC-W06-E02-S003-03 | Once W06-E01-S002 is accepted, run the generated-consumer N-1-to-N upgrade check, reusing DX-04's drill | CI, real infrastructure | Contracts re-pass after the upgrade | two-pass integration-test report | unassigned |

## Post-execution record

Entry criteria were inspected; no blocked verification was executed or claimed.

### Actual result

All three legs remain blocked at their documented entry gates.

### Pass or fail

BLOCKED, not failed.

### Evidence identifier

EV-W06-E02-S003-001 (entry-criterion record).

### Execution date

2026-07-13.

### Commit or revision

Working tree based on `733ef3e`.

### Environment

Repository lifecycle records.

### Reviewer

W06-E01-E04-Execution.W06E02ReviewFinal — PASS; blockers confirmed honest.

### Findings

See `evidence/unblocking-status.txt` for exact per-leg statuses.

### Retest status

Not applicable until a leg becomes eligible.

### Final conclusion

Contract-compliant blocked state independently reviewed; no open issues.
