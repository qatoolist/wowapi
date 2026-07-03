# Phase 4 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-03.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `make migrate` (00004–00006 on fresh schema) | 0 | 6 migrations applied; org/party/capacity, resource/relationship, authz spine |
| 2 | `go test ./kernel/authz/ ./kernel/policy/` (evaluator + policy units) | 0 | deny-by-default matrix, RBAC scope coverage, ReBAC, ABAC deny-first, sensitive-denial audit, policy operators |
| 3 | `go get golang-jwt/jwt/v5` + `go test ./kernel/auth/` (agent) | 0 | 11 tests: verify→Actor, expired/wrong-iss/aud/kid, HS256 alg-confusion rejected, tampered sig, token never leaks |
| 4 | `DATABASE_URL=… go test -run Integration ./kernel/authz/ ./kernel/relationship/ ./kernel/resource/` (agent pg stores) | 0 | assignment perms+scope, org ancestors/subtree, ResourceOrg, relationship Has, registrar upsert version bump |
| 5 | RBAC self-grant probe: as app_rt, INSERT actor_assignments | denied | pre-emptive fix — `permission denied for table actor_assignments` (spine SELECT-only) |
| 6 | ARCH-36 refactor: thread TenantDB through Store/Checker/Evaluator | 0 | one snapshot in the request tx; stores stateless; unit + integration tests updated |
| 7 | SEC fixes (24/25/26/27/29/30/46) in evaluator/store/migrations | 0 | ReBAC self-grant backstop, deny fail-closed, scope over-grant guards, CTE cycle cap, ResourceOrg type match |
| 8 | `DATABASE_URL=… go test -run Integration ./kernel/authz/` (escalation backstops) | 0 | `TestIntegrationNoSelfGrantVia{Assignments,Relationships}` (permission denied), `TestIntegrationScopeCheckConstraints` |
| 9 | `unset DATABASE_URL; make ci` | 0 | vet, boundary lint, unit (incl. authz matrix + SEC regressions), race, build all green |
| 10 | `make test-integration` (all packages) | 0 | authz/relationship/resource/testkit integration green against the SEC-fix migration template |
