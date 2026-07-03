# 01 — Generic Domain Model, Multi-Tenancy, Authorization

## 1. Framework glossary (all concepts are core unless marked *module*)

Format per concept: **Meaning · Is-not · Key fields · Rules · Example**.

### Tenant — core
- **Meaning:** the hard isolation boundary for data, config, rules, workflows, billing of the SaaS itself.
- **Is-not:** a customer's business unit (that's Organization), not a user group.
- **Fields:** `id`, `slug` (unique, immutable), `display_name`, `parent_tenant_id?`, `status(active|suspended|closed)`, `settings jsonb`, audit cols.
- **Rules:** every tenant-scoped row carries its `tenant_id`; RLS keys off it; parent/child only grants *administrative* visibility via explicit cross-tenant grants, never implicit data access.
- **Example:** "Sunrise Heights CHS" is a tenant; "Greenfield International School" is another tenant of the same deployment.

### Organization — core
- **Meaning:** a business entity *inside* a tenant (legal entity, branch, chapter, department).
- **Is-not:** the isolation boundary; orgs share the tenant's RLS scope but form an authz scope level.
- **Fields:** `id`, `tenant_id`, `parent_org_id?`, `name`, `kind` (module-defined string), `status`.
- **Rules:** tree within a tenant; authz scopes may target an org subtree.
- **Example:** a facility-management tenant with orgs "West Region" → "Site 14".

### Workspace / Project / Account — **not in v1 core**
Organization subtree covers the need. Add later only if a domain needs a collaboration container that is *not* a legal/structural org. Decision recorded so nobody bolts a fourth hierarchy on ad hoc.

### Party — core
- **Meaning:** abstract subject that can hold relationships: a Person or a Legal Entity.
- **Is-not:** a login (User), not an authorization subject (Actor).
- **Fields:** `id`, `tenant_id`, `kind(person|legal_entity)`, `display_name`, `status`.
- **Rules:** contacts hang off party; persons/legal_entities are 1:1 extensions (same id).
- **Example:** person "Asha Rao"; legal entity "Acme Facility Services Pvt Ltd".

### Person / Legal entity — core
- Person: `party_id (pk/fk)`, `given_name`, `family_name`, `dob?`, `locale?`. Legal entity: `party_id`, `legal_name`, `registration_no?`, `jurisdiction?`.
- **Is-not:** person ≠ user. A person may exist with no login (a vendor contact); a user may map to a person later.

### User — core
- **Meaning:** an authenticated identity (mapped to IdP subject).
- **Fields:** `id`, `idp_subject` (unique), `email`, `status`, `person_party_id?`. **Global**, not tenant-scoped; tenant membership lives in `user_tenant_access`.
- **Rules:** a user acts inside a tenant only if an active `user_tenant_access` row exists.
- **Example:** asha@example.com is one user with access to two tenants.

### Actor — core
- **Meaning:** the *authorization subject* of any action: a user-in-capacity, or a system principal (`system:outbox-relay`, `system:scheduler`, `webhook:providerX`).
- **Is-not:** a table by itself — it's a resolved runtime value `Actor{Kind, UserID?, CapacityID?, SystemName?}` persisted by reference in audit/assignments.
- **Rules:** every audited action, assignment, workflow task, and job records its actor. System and webhook actors get roles like any other actor — no bypass path.

### Acting capacity — core
- **Meaning:** the hat a user wears in a tenant: a link `(user, tenant, party?, label)` under which assignments are granted.
- **Is-not:** a role. Capacity says *who you are acting as*; roles say *what that capacity may do*.
- **Fields:** `id`, `tenant_id`, `user_id`, `party_id?`, `label`, `status`, `valid_from/valid_to`.
- **Rules:** a user with >1 active capacity in a tenant must select one per session/request (`X-Acting-Capacity`); with exactly 1, it's implicit. All assignments attach to the capacity, not the raw user.
- **Example:** the same user is capacity "self" (a requester) and capacity "authorized signatory of Acme Ltd" — approvals granted to the signatory capacity are invisible when acting as "self".

### Role — core mechanism; role *templates* come from modules
- **Fields:** `id`, `tenant_id?` (null = platform role template), `key`, `name`, `is_system`.
- **Rules:** roles are bags of permissions; granted to capacities via assignments *at a scope*. Naming: `module.audience.function` → `core.tenant.admin`, `requests.org.approver`.

### Permission — core mechanism; catalog entries from modules
- **Fields:** `key` (unique, `module.resource.action` e.g. `document.read`, `requests.request.approve`), `description`, `module`.
- **Rules:** registered at startup from module catalogs; unknown keys fail boot. Deny-by-default: nothing is permitted without a matching assignment.

### Policy / Policy condition — core
- **Meaning:** a conditional layer on top of role grants: constraints (ABAC) that can *deny* or *narrow* an otherwise-allowed action.
- **Fields:** policy: `id`, `tenant_id?`, `key`, `effect(allow|deny)`, `priority`; condition rows: `attribute`, `op`, `value jsonb` (e.g. `resource.status != "locked"`, `env.time within business_hours`, `actor.relationship contains "assigned_to"`).
- **Rules:** deny overrides allow; policies attach to permissions or resource types; conditions evaluate against a small, documented attribute bag — not a general expression VM in v1.

### Relationship / Relationship type — core
- **Meaning:** typed edge `(subject party|resource) —type→ (object resource|party)` with validity window. The ReBAC substrate.
- **Fields:** type: `key` (`module.verb`, e.g. `core.owner_of`, `requests.assigned_to`), `subject_kind`, `object_kind`, `cardinality`; relationship: both ends as `(kind,type,id)` refs + `valid_from/valid_to`, `attributes jsonb`.
- **Rules:** types are registered by modules; the kernel stores/queries edges and exposes `HasRelationship(actor, type, resource)` to the policy evaluator. Temporal: never physically deleted, ended by `valid_to`.
- **Example:** party P `core.owner_of` resource R (an asset); capacity C `requests.assigned_to` request Q.

### Resource / Resource type / Resource instance — core registry
- **Meaning:** the kernel's way to point at *any* domain object without knowing it. Modules own their real tables; they register a `resource_type` and upsert a thin row in `resources` (id, type, tenant, org, label, status) for everything that needs kernel services (authz scope, comments, attachments, documents, workflow, relationships, audit).
- **Is-not:** an EAV store for domain data. Domain fields live in module tables.
- **Rules:** `resources.id` == the module row's id (shared UUID). Kernel services accept `ResourceRef{Type, ID}` only.

### Assignment — core
- **Meaning:** grant of a role to an actor (capacity or system principal) at a scope, optionally time-boxed, optionally delegated.
- **Fields:** `id`, `tenant_id`, `capacity_id|system_actor`, `role_id`, `scope_kind(tenant|org|resource_type|resource)`, `scope_id?`, `valid_from/valid_to`, `granted_by`, `delegated_from?`, `reason?`.
- **Rules:** append + end (`valid_to`), never hard-deleted (audit). Temporary access = assignment with `valid_to`. Delegation = new assignment linked via `delegated_from`, bounded by delegator's own validity.

### Scope — core (value concept)
Hierarchy `tenant ⊃ organization(subtree) ⊃ resource_type ⊃ resource`. An assignment at a scope covers everything beneath it. Record scope is the `resource` level.

### Workflow definition / instance / task — core (see [02-workflow-rules.md](02-workflow-rules.md))
Definition: versioned JSON graph of steps registered per tenant (or template). Instance: one run bound to a `ResourceRef`. Task: a unit of human/auto work with assignees, due dates, decisions.

### Rule definition / Rule version — core (see [02-workflow-rules.md](02-workflow-rules.md))
Definition = a typed *rule point* (key + JSON Schema + default). Version = a value with `effective_from/to`, scope (platform|tenant|org), approval status. Historical evaluation = resolve *as of* a timestamp.

### Feature flag — core
`key`, scope (platform/tenant), `enabled`, optional variant jsonb. Implemented as a thin special case of the rule engine (rule point of type boolean/variant) — one resolution path, one audit trail.

### Document / Document version / Document access grant — core
Document: metadata + class (module-registered `document_class`) + link to `ResourceRef`. Version: immutable file pointer (storage key, checksum, size, mime, uploaded_by). Grant: explicit read/write access for an actor/role/relationship beyond default policy. Files live in object storage; Postgres holds metadata only.

### Comment — core
Threaded note on any `ResourceRef`: `id`, `tenant_id`, resource ref, `author_actor`, `body`, `parent_comment_id?`, edited/redacted flags. Soft-void, never hard-delete (audit).

### Attachment — core
Link table binding an uploaded file (document version or raw file object) to a `ResourceRef` + context (`comment_id?`, `workflow_task_id?`).

### Event — core
Immutable fact: `id (uuidv7)`, `tenant_id`, `type` (`module.resource.verb` past tense, e.g. `requests.request.approved`), `schema_version`, `resource ref`, `actor`, `payload jsonb`, `occurred_at`. Written to outbox in the business tx.

### Job — core
Queued unit of async work: `kind`, `tenant_id?`, `payload`, schedule/priority, attempts, status. Job ≠ event: events are facts, jobs are intentions.

### Notification — core
A message to a party/user through channels (in-app, email, SMS, WhatsApp, push): template key + variables + per-channel `notification_deliveries` with status/retries.

### Audit log — core
Append-only record of every sensitive action and every permission denial: actor, capacity, tenant, action key, `ResourceRef`, before/after digest or diff, request id, ip, result. No UPDATE/DELETE grants to the app role — enforced in DB, tested in testkit.

### Integration provider — core
Registered external system (`kind`: payment, identity-verification, messaging, storage, device), config + credential ref (secret provider key, never plaintext), status, per-tenant enablement.

### Webhook endpoint / Webhook event — core
Endpoint: inbound (provider → us: secret, signature scheme, per-provider verifier) or outbound (us → customer: URL, secret, subscribed event types). Webhook event: stored received/sent payload with signature-verification result, idempotency key, delivery attempts.

### Idempotency key — core
`(tenant_id, actor, key)` → request hash + stored response + status + expiry. Same key + same hash ⇒ replay stored response; same key + different hash ⇒ `409 conflict`.

### How a society module maps onto this (illustration only — lives entirely in a future society product repo's `internal/modules/society`, importing wowapi)
building/wing/unit → module tables + `resource_types` (`society.building`, `society.unit`); flat owner/tenant → `relationship_types` (`society.owner_of_unit`, `society.occupier_of_unit`) from party→unit; member/associate/nominal → module-defined membership records + capacities; committee/chairman/secretary/treasurer → roles (`society.tenant.chairman`…) granted via assignments with validity = term; AGM/notice/bill approvals → workflow definitions; maintenance rates, defaulter thresholds, notice periods → rule points; society notices/minutes → document classes; gate entries/complaints → module resources with workflows. Zero kernel changes.

## 2. Multi-tenancy design

### Model
- **Shared database, shared schema, `tenant_id` column + PostgreSQL RLS** — v1 choice. Cheapest to operate, easiest to migrate/backup, strong enough isolation when RLS is FORCED and tested.
- **Escape hatch (design now, build later):** the `TenantDB` abstraction and per-module sqlc queries never name a schema; a future `schema-per-tenant` or `db-per-tenant` router can be introduced behind `TxManager` for regulated customers without touching module code.

### Mechanics (normative)
1. Two DB roles: `app_migrate` (owner, runs DDL) and `app_rt` (runtime). Every tenant-scoped table:
   `ALTER TABLE t ENABLE ROW LEVEL SECURITY; ALTER TABLE t FORCE ROW LEVEL SECURITY;` with policy
   `USING (tenant_id = current_setting('app.tenant_id')::uuid)` (+ identical `WITH CHECK`).
2. Every unit of work runs `SET LOCAL app.tenant_id = $1` (and `app.actor_id`) as the first statement
   inside the transaction. `SET LOCAL` dies with the tx — no leakage across pooled connections.
3. The application never exposes a raw pool to modules. Modules get `TenantDB` (see [05-http-and-persistence.md](05-http-and-persistence.md)) whose only constructor is `TxManager.WithTenant(ctx, fn)`, which requires tenant context in `ctx`. **No tenant in ctx → error, not fallback.**
4. Platform-global tables (`tenants`, `users`, `permissions`, role templates, platform rule versions) have no RLS; they are reachable only through kernel services, never through module repositories.
5. Cross-tenant admin operations (platform support) use a separate `app_platform` role + explicit break-glass flow (audited, time-boxed), never by skipping `SET LOCAL`.

### Tenant-aware everything
| Concern | Mechanism |
|---|---|
| AuthN | JWT is tenant-agnostic; middleware resolves tenant from path `/v1/t/{tenantSlug}/…` (browser) or `X-Tenant` header (API), then requires an active `user_tenant_access` row. |
| AuthZ | all assignments/policies carry `tenant_id`; evaluator only loads rows for the resolved tenant. |
| Jobs/events | payload carries `tenant_id`; the worker's first act is `SET LOCAL`. Jobs without tenant are explicitly `platform`-scoped kinds. |
| Documents/files | storage keys prefixed `tenants/{tenant_id}/…`; presigned URLs generated per-tenant; bucket policies deny cross-prefix. |
| Rules/workflows/config | resolution order org → tenant → platform default (see 02). |
| Rate limiting | token buckets keyed `(tenant)`, `(tenant, actor)`; per-tenant quota from rule point `core.rate_limit.rpm`. |
| Observability | `tenant_id` is a log field + trace attribute + metric label (bounded: use tenant *tier* for high-cardinality metrics; exact tenant only in logs/traces). |
| Migrations | schema is shared: one migration run serves all tenants. Data backfills iterate tenants inside tenant-scoped txs. |
| Isolation testing | testkit `AssertRLSIsolation(t, tableName)`: creates 2 tenants, inserts as A, asserts count-as-B == 0 and cross-write fails. Runs against every registered tenant-scoped table automatically (catalog-driven). |

### Parent/child tenants & cross-tenant grants
`tenants.parent_tenant_id` models franchise/federation structures. It grants *nothing* by itself.
Explicit `user_tenant_access` rows (kind `member|support|federated_admin`, time-boxed) are the only
cross-tenant door, and every such session is flagged in audit (`cross_tenant=true`).

## 3. Actor / relationship / role / permission / policy framework

### The layered decision (evaluation algorithm, normative)
```
Evaluate(actor, permissionKey, target ScopeOrResource) Decision:
 1. deny-by-default; permission must exist in registry (else 500 at boot, not runtime)
 2. load active assignments for actor (capacity or system principal) in tenant
        where scope covers target (tenant ⊃ org-subtree ⊃ resource_type ⊃ resource)
        and now() within [valid_from, valid_to)
 3. RBAC: any assignment whose role contains permissionKey → candidate ALLOW
 4. ReBAC: if no RBAC allow, check relationship-derived grants:
        permission registry may declare "granted_via relationship R on target"
        (e.g. document.read granted_via core.owner_of) → candidate ALLOW
 5. ABAC/policies: evaluate attached policies (deny first, priority order)
        against attribute bag {actor.*, capacity.*, resource.*, env.time, request.*}
        → any matching DENY kills the decision; ALLOW policies may *add* narrow grants
 6. record Decision{effect, matched assignment/policy/relationship} 
 7. if DENY on a sensitive permission or any explicit deny → audit the denial
```
Caching: per-request memoization always; per-actor assignment snapshot cached 30s with
version-stamp invalidation (assignments table `xmin`/updated_at max). No cross-request policy caching in v1.

### Go interfaces (kernel `authz` package)
```go
type Actor struct {
    Kind       ActorKind // user | system | webhook
    UserID     uuid.UUID
    CapacityID uuid.UUID // zero for system actors
    System     string    // "outbox-relay", "webhook:razorpay"
    TenantID   uuid.UUID
}

type Target struct {
    Scope    ScopeKind    // tenant | org | resource_type | resource
    OrgID    uuid.UUID
    Resource resource.Ref // {Type string; ID uuid.UUID}
}

type Decision struct {
    Allowed   bool
    Reason    string     // "role:requests.org.approver", "rel:core.owner_of", "policy:deny_locked"
    PolicyIDs []uuid.UUID
}

type Evaluator interface {
    Evaluate(ctx context.Context, a Actor, perm string, t Target) (Decision, error)
    // Filter narrows a list query: returns predicate the store layer translates to SQL
    Filter(ctx context.Context, a Actor, perm string, rt string) (ListFilter, error)
}

type RelationshipChecker interface {
    Has(ctx context.Context, subject authz.Actor, relType string, obj resource.Ref, at time.Time) (bool, error)
}
```
`Filter` matters: record-level lists must not load-then-filter. The evaluator returns structured
constraints (`org IN (…)`, `resource.id IN (SELECT … relationships …)`) that sqlc queries embed.

### Naming conventions
- Permission: `module.resource.action`, lower snake dot; actions from the closed verb set
  `create|read|list|update|deactivate|restore|approve|reject|assign|export|admin|ingest|activate`
  (`ingest` = inbound webhook/event ingestion, `activate` = break-glass/feature activation; D-0038).
- Role: `module.audience.function` (`core.tenant.admin`, `core.org.auditor`, `requests.org.approver`).
- Relationship type: `module.verb_phrase` (`core.owner_of`, `core.member_of`, `requests.assigned_to`).

### Generic permission matrix (illustrative)
| Actor (capacity+role) | `requests.request.create` | `.approve` | `.read` (own) | `.read` (all org) | `users.user.admin` |
|---|---|---|---|---|---|
| Org member (`core.org.member`) | ✅ | ❌ | ✅ via `rel:created_by` | ❌ | ❌ |
| Manager (`requests.org.approver`) | ✅ | ✅ (policy: not own request) | ✅ | ✅ | ❌ |
| Org admin (`core.org.admin`) | ✅ | ❌ | ✅ | ✅ | ✅ (org scope) |
| External auditor (`core.org.auditor`, valid_to=+30d) | ❌ | ❌ | ✅ | ✅ read-only | ❌ |
| Vendor capacity (`work.vendor.assignee`) | ❌ | ❌ | ✅ via `rel:assigned_to` only | ❌ | ❌ |
| `system:sla-sweeper` | ❌ | ❌ | ✅ | ✅ | ❌ |
| `webhook:payments` | only `payments.callback.ingest` | | | | |

### Special flows (all core, all audited)
- **Delegation:** approver creates a time-boxed assignment `delegated_from=own_assignment`; revoked automatically when the source assignment ends.
- **Break-glass:** platform role `core.platform.breakglass` grants nothing until activated via a dedicated endpoint that (a) requires reason, (b) creates a ≤60min assignment, (c) fires `core.security.breakglass_activated` event + notification to tenant admins, (d) banners every audit row it produces.
- **Impersonation:** support actor gets `impersonating_user_id` in ctx; every audit row records both identities; impersonation cannot approve workflows or change security settings (kernel policy `deny`).
- **Emergency workflow override:** see workflow engine — an override records mandatory reason + creates a ratification task.

### Society mapping (illustration)
Owner/occupier → relationship types; secretary/treasurer/chairman → roles with term-limited assignments; auditor → `core.org.auditor` time-boxed; vendor → vendor capacity + `assigned_to` relationships; "member can view own society docs" → `document.read granted_via society.member_of`.
