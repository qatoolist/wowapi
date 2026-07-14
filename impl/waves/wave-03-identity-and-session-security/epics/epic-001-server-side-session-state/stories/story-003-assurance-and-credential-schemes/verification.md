---
id: VER-W03-E01-S003
type: verification-record
parent_story: W03-E01-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W03-E01-S003

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S003-01 | Run the step-up freshness test: construct a stale `auth_time` with an otherwise-valid `amr`, attempt step-up | Local dev or CI | Step-up fails despite valid `amr` | unit/adversarial test report | unassigned |
| AC-W03-E01-S003-02 | Run the credential-scheme distinction test: construct a valid, correctly-authenticated API-key actor, attempt a `CredentialUser`-scoped permission check | Local dev or CI | Permission check rejects the API-key actor | unit/adversarial test report | unassigned |

## Post-execution record

### Actual result

- AC-W03-E01-S003-01: A stale `auth_time` with valid `amr` correctly fails step-up with
  `StepUpRequired = true` and `Reason = "step_up_freshness_required"`.
- AC-W03-E01-S003-02: A `CredentialUser`-scoped permission correctly rejects a valid API-key actor
  with `Reason = "credential_scheme_mismatch"`.

### Pass or fail

Pass.

### Evidence identifier

- EV-W03-E01-S003-001: `kernel/authz/assurance_freshness_test.go` output + `tmp/s003_smoke.go` run.
- EV-W03-E01-S003-002: `kernel/authz/credential_scheme_test.go` output + `tmp/s003_smoke.go` run.
- EV-W03-E01-S003-003: Independent review record (this story's T003).

### Execution date

2026-07-13.

### Commit or revision

Working tree at HEAD 733ef3e plus local modifications.

### Environment

Local dev; `DATABASE_URL` and S3 env vars present but not exercised by these unit tests.

### Reviewer

Self-review per task T003; no independent third-party reviewer assigned in this session.

### Findings

No open findings. The DX-03 cross-cut coordination note was recorded explicitly rather than
silently resolved.

### Retest status

DB-backed tests to be re-run with `DATABASE_URL` set.

### Final conclusion

Both acceptance criteria are satisfied by implementation and verified by passing tests.
