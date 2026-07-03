# Phase 4 — Review Findings

Two parallel critique agents reviewed the identity/authz slice (2026-07-03): **S** = hostile
authorization-security reviewer (deny-by-default, self-grant, RLS — ran live Postgres probes) and
**A** = architecture/API reviewer. A pre-emptive RBAC self-grant fix (SEC-13-class) was applied
before review and confirmed closed live by S. Fixes carry regression tests; the two self-grant
backstops and the scope CHECKs have live integration tests.

| ID | Sev | Finding | Resolution | Status |
|---|---|---|---|---|
| (pre-empt) | critical | RBAC self-grant: app_rt could INSERT actor_assignments/roles and grant itself any role | authz spine SELECT-only to app_rt, writes app_platform-only; `TestIntegrationNoSelfGrantViaAssignments` (live: permission denied). Confirmed closed by S live | **fixed** |
| SEC-24 | high | ReBAC self-grant: app_rt could INSERT a `relationships` edge naming its own capacity → self-grant any granted_via permission — reproduced live | app_rt loses INSERT/UPDATE on relationships (edge creation is an app_platform/kernel capability); `TestIntegrationNoSelfGrantViaRelationships` (live: permission denied) | **fixed** |
| SEC-25 | high | deny policy gating on an un-populated attribute (e.g. `resource.status`) silently fails OPEN — a "locked record" deny never fires and an RBAC allow stands — reproduced | deny policies referencing an unresolvable attribute now fail CLOSED (deny + audit); `TestABACDenyUnresolvedAttributeFailsClosed` | **fixed** |
| SEC-26 | med/high | `covers()` resource_type scope with empty `scope_type` matches a typeless target (`""==""`) → over-grant — reproduced | `covers` requires `ScopeType != ""`; DB CHECK forbids NULL scope_type for resource_type scope; `TestResourceTypeScopeEmptyTypeNoOverGrant`, `TestIntegrationScopeCheckConstraints` | **fixed** |
| SEC-27 | med | allow policy leaks via the `$2=''` query broadening — a type-bound policy applies to a typeless check → grant leak — reproduced | `Policies` query drops the `$2=''` disjunct: a type-bound policy never matches a typeless check | **fixed** |
| SEC-28 | med | `Filter` omits ABAC deny (and ReBAC) → list endpoints under-authorize/over-return | Filter documented as RBAC-only and safe-but-incomplete (under-grants, never over-grants); list handlers must run per-row Evaluate for deny-policy permissions until the Phase 5 deny-aware/ReBAC seam lands (D-0039) | **documented / Phase 5** |
| SEC-29 | low | `ScopeResource` matches a NULL scope_id against a nil target id | `covers` requires `ScopeID != uuid.Nil`; DB CHECK forbids NULL scope_id for org/resource scope | **fixed** |
| SEC-30 | low | recursive org CTEs have no cycle guard → self-inflicted tenant-local DoS | CTEs use `UNION` (dedup) + a depth cap (`depth < 64`) | **fixed** |
| SEC-31 | low/med | `resources` mirror app_rt-writable incl. org_id (relocate to over-granted org) | accepted for Phase 4: the mirror is the module's own aggregate; cross-tenant blocked by RLS. org_id-transition authorization + audit tracked for the document/resource work (Phase 8) | **accepted (tracked)** |
| ARCH-36 | high | evaluator opened ~5 separate RO transactions per check (wrong snapshot vs the request's writes, N round-trips, deadlock risk) | Store/Checker/Evaluator methods now take the caller's `TenantDB`; one snapshot in the request tx; stores are stateless (D-0039) | **fixed** |
| ARCH-37 | high | ReBAC list visibility unimplemented → relationship-only users get empty lists | documented in `Filter`; `Store.RelationshipResourceIDs` seam lands with Phase 5 list endpoints (D-0039) | **documented / Phase 5** |
| ARCH-38 | med | blueprint caching (per-request memo, 30s snapshot) absent, deferral undocumented | deferred to Phase 5/6 in D-0039; the in-tx evaluator makes per-request memoization cheap to add there | **documented** |
| ARCH-39 | med | Context.Authz()/Permissions()/Resources() dangling; app never builds the evaluator | expected per D-0006; Phase 5 app boot wires the evaluator (build Registry from module perms + gate boot on Err(), PgStore, policy, relationship, AuditSink) — noted | **accepted (Phase 5)** |
| ARCH-40 | med | ABAC allow policy grants tenant-wide, not "narrow" | accepted with SEC-27 scoping: allow policies are constrained by the Policies query to their permission + resource type, so an allow grants only within that perm+type, not across all permissions. Broader per-target narrowing tracked for Phase 5 | **accepted (scoped)** |
| ARCH-41 | med | verb set adds `ingest`/`activate` beyond the closed set, undocumented | D-0038 records the extension + blueprint 01 §3 updated | **fixed** |
| ARCH-42 | med | partial indexes `WHERE valid_to IS NULL` miss time-boxed-active rows; no system_actor index | indexes cover the temporal predicate; added `asg_system` | **fixed** |
| ARCH-43 | low | `ScopeExtractor` still `any` | retyped to `func(*http.Request) (authz.Target, error)` | **fixed** |
| ARCH-44 | low | `PrincipalStore` has no DB impl (auth→actor dangling) | Phase 5 wiring (kernel/auth may not import a DB); noted | **accepted (Phase 5)** |
| ARCH-45 | low | `boundRegistrar` can outlive its tx | documented "valid only within the binding tx" on Bind | **documented** |
| ARCH-46 | low | `ResourceOrg` ignores `Ref.Type` | query now matches `resource_type = $2` | **fixed** |

Reviewer-confirmed solid (verified live): RBAC self-grant closed; `Evaluate` fails closed on every
error path (no allow-with-error); registry symmetry (Evaluate + Filter both reject unregistered
permissions); relationship-checker tenant isolation + time-bound; `ActiveAssignments` exact actor
match with nil-disabled branches; full RLS coverage on all tenant tables incl. platform-template
read admission.

Residual risk: the model has no intra-tenant, inter-module isolation (all modules share app_rt) —
cross-tenant is RLS-protected throughout, but a compromised module can act within its tenant with
app_rt's grants. The authz decision inputs (assignments, roles, edges) are now write-protected from
app_rt; the remaining `resources.org_id` transition (SEC-31) is tracked for Phase 8. Live authz
end-to-end (evaluator under a real request tx) is proven at the store/evaluator level but the full
app wiring lands in Phase 5.
