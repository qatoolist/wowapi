---
id: W03-E01-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W03-E01-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E01-S003 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` are created on first real content, not pre-populated
empty. All entries below are `not yet produced`.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W03-E01-S003-001 | unit/adversarial test report | W03-E01-S003-T001 | AC-W03-E01-S003-01 | `go test ./kernel/authz/... -run 'TestStepUpFreshness'`; `go test ./kernel/auth/...` | 733ef3e + local | PASS | produced |
| EV-W03-E01-S003-002 | unit/adversarial test report | W03-E01-S003-T002 | AC-W03-E01-S003-02 | `go test ./kernel/authz/... -run 'TestCredentialScheme'`; `go test ./kernel/apikey/...` | 733ef3e + local | PASS | produced |
| EV-W03-E01-S003-003 | review report | W03-E01-S003-T003 | AC-W03-E01-S003-01, AC-W03-E01-S003-02 | Independent review checklist per mandate §14 | 733ef3e + local | PASS | produced |

Note: DB-backed tests in `./kernel/authz/...` and `./kernel/apikey/...` are skipped when
`DATABASE_URL` is unavailable. All non-DB tests pass.

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
