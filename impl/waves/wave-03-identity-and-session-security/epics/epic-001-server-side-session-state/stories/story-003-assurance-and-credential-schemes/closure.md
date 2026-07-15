---
id: CLOSURE-W03-E01-S003
type: closure-record
parent_story: W03-E01-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Closure — W03-E01-S003

## Acceptance-criteria completion

- AC-W03-E01-S003-01 — Pass. Stale `auth_time` with valid `amr` fails step-up.
- AC-W03-E01-S003-02 — Pass. `CredentialUser`-scoped permission rejects a valid API-key actor.

## Task completion

- W03-E01-S003-T001 — Complete.
- W03-E01-S003-T002 — Complete.
- W03-E01-S003-T003 — Complete (review performed, no open findings).

## Artifact completeness

Artifacts registered in `artifacts/index.md`:

- ART-W03-E01-S003-001: Assurance-freshness binding/enforcement code change — produced.
- ART-W03-E01-S003-002: Credential-scheme distinction implementation — produced.

## Evidence completeness

- EV-W03-E01-S003-001: Step-up freshness test report (`kernel/authz/assurance_freshness_test.go`,
  `kernel/auth/assurance_internal_test.go`, `tmp/s003_smoke.go`).
- EV-W03-E01-S003-002: Credential-scheme distinction test report
  (`kernel/authz/credential_scheme_test.go`, `kernel/apikey/apikey_test.go`, `tmp/s003_smoke.go`).
- EV-W03-E01-S003-003: Independent review record (T003).

## Unresolved findings

None. The "expired step-up" required test class is covered.

## Accepted risks

DX-03 cross-cut: this story's `CredentialScheme` mechanism is provisional and may need
reconciliation once DX-03 (W06-E01-S001) lands. This is accepted and recorded.

## Deferred work

- `CredentialScheme` reconciliation with DX-03 (W06-E01-S001).
- DB-backed test re-run with `DATABASE_URL` set.

## Reviewer conclusion

Review checklist completed with no open findings. Implementation matches `plan.md`; the DX-03
cross-cut note was explicitly recorded.

## Acceptance authority

product-security lead (per epic-level acceptance convention).

## Closure date

2026-07-13.

## Final status

accepted (pending formal product-security lead sign-off and DB-backed test re-run).
