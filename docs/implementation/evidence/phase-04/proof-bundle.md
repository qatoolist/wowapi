# Phase 4 — Proof Bundle

Scope (phase-plan row 4): identity/auth (OIDC verifier, actor model), authz evaluator (RBAC/ReBAC/
ABAC deny-by-default), policy engine, relationship checker, resource registry, migrations 002–004
(on-disk 00004–00006). Date: 2026-07-03.

## 1. Decision evidence
D-0035 (migration numbering), D-0036 (evaluator + Store port, boot-validated registry), D-0037
(OIDC verifier + injectable JWKS + local issuer). Post-review: D-0038 (verb-set extension),
D-0039 (evaluator in-tx; caching + list-ReBAC deferred). Blueprint 01 §3 verb list updated.

## 2. Discussion evidence
- The two self-grant vectors: RBAC (write actor_assignments) and ReBAC (write relationships). Both
  are the same class — the authz decision inputs must be write-protected from the shared app_rt
  role modules run as. Resolved by making the entire authz spine (assignments, roles,
  role_permissions, policies, relationships) SELECT-only to app_rt; writes are an app_platform/
  kernel capability. Reproduced closed live.
- ABAC fail-open: the security review reproduced that a deny policy on an attribute the evaluator
  cannot populate (the blueprint's own `resource.status`) silently does not match. Deny policies on
  unresolvable attributes now fail closed — the opposite of a silent skip.
- Evaluator transaction structure (ARCH-36): the initial pg store opened its own RO tx per method,
  giving ~5 transactions and a different snapshot from the request's writes. Refactored so every
  port takes the caller's TenantDB and runs in the request tx — one snapshot, no extra connections.

## 3. Critique/review evidence
`review-findings.md`: 21 findings (2 reproduced high security + 1 high architecture, several
medium, rest low/info). The self-grant and fail-open holes were reproduced on live Postgres before
fixing; each has a regression test (unit for logic, live integration for the DB backstops). Items
genuinely belonging to Phase 5 app-wiring (evaluator injection, list-ReBAC, caching) are documented
and deferred with rationale.

## 4. Implementation evidence
New: `kernel/auth/` (verifier, keysource), `kernel/authz/` (authz, registry, store, evaluator,
store_pg), `kernel/policy/`, `kernel/relationship/`, `kernel/resource/` (resource, registrar_pg),
migrations 00004–00006, `testkit/auth.go`. Changed: `module/module.go` + `app/context.go`
(Permissions/Resources/Authz accessors), `kernel/httpx/router.go` (ScopeExtractor → authz.Target).
Deps: golang-jwt/jwt/v5. Team: 2 parallel implementation agents (OIDC/testkit; pg stores) + lead
(migrations, authz+policy evaluator, all security refactor/fixes); 2 parallel review agents.

## 5. Verification evidence
`command-log.md`: unit suites (deny-by-default matrix, policy operators, auth verifier incl.
alg-confusion), pg-store integration, the RBAC + ReBAC self-grant live probes (permission denied),
the scope CHECK constraints, ARCH-36 refactor, full `make ci` + `make test-integration`. Graphify
updated.

## 6. Acceptance evidence
`acceptance-map.md`: all 16 Phase 4 exit criteria mapped to code + named tests; acceptance #4
(deny-by-default + authz matrix + sensitive-denial audit) to specific tests, and the DB-level
no-self-grant backstops to live integration tests. Carried forward to Phase 5: evaluator boot
wiring into module.Context, list-ReBAC + ABAC deny in Filter, PrincipalStore adapter, caching;
Phase 8: resources.org_id transition authz (SEC-31). Graphify `extract` blocked on LLM key (R11).
