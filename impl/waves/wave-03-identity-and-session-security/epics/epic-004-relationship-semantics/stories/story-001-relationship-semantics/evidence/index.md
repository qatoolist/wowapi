---
id: W03-E04-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W03-E04-S001
status: recorded
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E04-S001 — Evidence index

Per mandate §10. Evidence recorded from `go test ./kernel/relationship/...` executed
against the working tree on 2026-07-13. Commit SHA will be updated once the W03
implementation branch is committed.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W03-E04-S001-001 | party-subject-edge test report | W03-E04-S001-T001 | AC-W03-E04-S001-01 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./kernel/relationship/... -run TestIntegrationRelationshipHasPartySubject -count=1 -v` | working tree | PASS | produced |
| EV-W03-E04-S001-002 | subject-kind matrix test report | W03-E04-S001-T002 | AC-W03-E04-S001-02 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./kernel/relationship/... -run TestIntegrationRelationshipSubjectKindMatrix -count=1 -v` | working tree | PASS | produced |
| EV-W03-E04-S001-002a | fail-closed default unit test | W03-E04-S001-T002 | AC-W03-E04-S001-02 (unenumerated kind) | `go test ./kernel/relationship/... -run TestUnitResolveSubjectUnsupportedKind -count=1 -v` | working tree | PASS | produced |
| EV-W03-E04-S001-003 | mutation-governance test report | W03-E04-S001-T003 | AC-W03-E04-S001-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./kernel/relationship/... -run 'TestIntegrationRelate(RequiresActor|AttributesAndVersions|WritesAudit)' -count=1 -v` | working tree | PASS | produced |
| EV-W03-E04-S001-004 | review report | W03-E04-S001-T004 | AC-W03-E04-S001-01, AC-W03-E04-S001-02, AC-W03-E04-S001-03 | Independent review checklist per mandate §14 | TBD | pending | not yet produced |

Cache-invalidation sub-criterion (AC-W03-E04-S001-03) remains deferred-linked to
W05-E04-S002 per plan; no test fabricated against a non-existent epoch table.
