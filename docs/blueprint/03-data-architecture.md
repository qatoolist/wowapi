# 03 — Data Architecture, ERD, PostgreSQL DDL Skeleton

Framework foundation tables only. Domain modules add their own tables (own migrations, own sqlc),
always with `tenant_id` + RLS, and register resource rows for anything needing kernel services.

## 1. Cross-cutting conventions (normative)

| Convention | Decision |
|---|---|
| Primary keys | `uuid` (UUIDv7, app-generated via injected `IDGen` — testable, time-ordered, index-friendly). |
| Audit columns | `created_at timestamptz not null default now()`, `created_by uuid not null`, `updated_at timestamptz`, `updated_by uuid` on mutable tables. Append-only tables: `created_*` only. |
| Optimistic locking | `version int not null default 1`; every UPDATE: `SET version = version + 1 WHERE id=$1 AND version=$2`; 0 rows → `ErrVersionConflict` (HTTP 409/412 with `ETag`/`If-Match`). On all user-editable aggregates; NOT on append-only or counters. |
| Temporal validity | `valid_from timestamptz not null`, `valid_to timestamptz` (null = open) on assignments, relationships, capacities, rule versions, access grants. End rows, don't delete them. |
| Soft delete | `status` lifecycle (`active|inactive|archived`) preferred over `deleted_at` — deletion is a business state. True erasure (GDPR) = dedicated redaction jobs, not DELETE in handlers. |
| Append-only | `audit_logs`, `events_outbox`, `document_versions`, `job_runs`, `webhook_events`, `notification_deliveries`: INSERT-only; app role gets no UPDATE/DELETE grant (except narrow status columns via `SECURITY DEFINER` functions or a separate relay role where noted). |
| RLS | Tenant-scoped tables: `ENABLE` + `FORCE ROW LEVEL SECURITY`, policy `tenant_id = current_setting('app.tenant_id')::uuid` for USING and WITH CHECK. Global tables: no RLS, kernel-service access only. |
| Indexes | Every FK gets an index. Tenant-scoped hot paths use composite `(tenant_id, …)` leading indexes. Keyset pagination indexes `(tenant_id, created_at DESC, id DESC)` where listed. |
| JSONB — yes | payloads whose schema is owned elsewhere (event payloads, rule values validated by JSON Schema, workflow definitions/context, provider configs, metadata/extension bags). |
| JSONB — no | anything you filter/join/aggregate on, money, statuses, foreign keys, dates driving logic. If a module keeps querying into a JSONB field → promote to a column (expand-contract migration). |
| Money | `numeric(18,4)` + `currency char(3)` columns (kernel `Money` type). Never float, never JSONB. |
| Migrations | goose, per-module dirs, expand-contract for breaking changes (add-new → dual-write/backfill → cut-over → drop-old). |

## 2. Table matrix

Scope: **G**=global, **T**=tenant-scoped(RLS). Flags: **A**=append-only, **V**=optimistic version, **Tm**=temporal validity, **S**=status lifecycle.

| Table | Scope | Flags | Purpose / notes |
|---|---|---|---|
| tenants | G | V,S | isolation root; unique `slug` |
| user_tenant_access | G | Tm,S | user↔tenant membership incl. cross-tenant grants; unique `(user_id,tenant_id,kind)` active |
| users | G | V,S | identity; unique `idp_subject`, unique `email` |
| organizations | T | V,S | org tree; unique `(tenant_id, parent_org_id, name)` |
| parties / persons / legal_entities | T | V,S | party supertype + 1:1 subtypes (shared PK) |
| party_contacts | T | V,S | typed contact points; unique `(party_id,kind,value)` |
| acting_capacities | T | Tm,V,S | user's hats; unique active `(tenant_id,user_id,label)` |
| resource_types | G | — | registry; PK `key` (`module.name`), owner module |
| resources | T | V,S | thin kernel mirror of module rows; PK `id` shared with module table; idx `(tenant_id,resource_type,status)` |
| relationship_types | G | — | registry; PK `key`; subject/object kinds, cardinality |
| relationships | T | Tm,V | edges; idx `(tenant_id,object_type,object_id,rel_type)` + subject mirror; exclusion on overlap for cardinality=1 types |
| roles | G+T | V,S | `tenant_id null` = platform template; unique `(coalesce(tenant_id),key)` |
| permissions | G | — | catalog; PK `key`; synced from module registration at boot |
| role_permissions | G+T | — | PK `(role_id,permission_key)` |
| actor_assignments | T | Tm,V | role grants at scope; idx `(tenant_id,capacity_id,valid_to)`, `(tenant_id,scope_kind,scope_id)` |
| policies / policy_conditions | G+T | V,S | ABAC layer; conditions FK policy |
| rule_definitions | G | — | persisted rule points; PK `key`; JSON Schema |
| rule_versions | T+G | A(status transitions only),Tm | values; exclusion constraint prevents overlapping active versions per scope |
| feature_flags | — | — | implemented as `feature.*` rule points (no separate table) |
| workflow_definitions | G+T | — | immutable per `(key,version,tenant_id?)` |
| workflow_instances | T | V | pins def version; idx `(tenant_id,resource_type,resource_id)`, `(tenant_id,status)` |
| workflow_tasks | T | V,S | idx `(tenant_id,status,due_at)`, assignee GIN or child table `workflow_task_assignees` |
| documents | T | V,S | metadata + class + resource ref |
| document_versions | T | A | immutable file pointers; unique `(document_id,version_no)` |
| document_access_grants | T | Tm,V | explicit grants beyond policy |
| comments | T | V,S(voided) | threaded on resource ref |
| attachments | T | V,S | file↔resource link (+ comment/task context) |
| notification_templates | G+T | V,S | key+channel+locale; tenant override rows |
| notifications | T | S | logical message; idx `(tenant_id,recipient_party_id,created_at)` |
| notification_deliveries | T | A(+status via relay) | per-channel attempts, provider msg id |
| audit_logs | T | A | partition by month (declarative); idx `(tenant_id,occurred_at)`, `(tenant_id,resource_type,resource_id)`, `(tenant_id,actor_user_id)` |
| events_outbox | T | A(+dispatch status) | idx `(status,occurred_at)` partial WHERE pending |
| processed_events | T | A | consumer inbox; PK `(handler,event_id)` |
| jobs / job_runs | G(tenant in payload) | S / A | River-managed core + kernel `job_runs` mirror for reporting |
| idempotency_keys | T | A + expiry sweep | PK `(tenant_id,actor_scope,key)`; stores request_hash, response, status |
| integration_providers | G+T | V,S | provider registry + per-tenant config/credential *references* |
| webhook_endpoints | T | V,S | inbound + outbound endpoints; secrets as secret-provider refs |
| webhook_events | T | A(+delivery status) | received/sent payloads, signature result, attempts |

## 3. Text ERD (core spine)

```text
tenants ─┬─< organizations >── parent (self)
         ├─< user_tenant_access >── users (G) ──?── persons
         ├─< parties ─┬─ persons (1:1)             
         │            └─ legal_entities (1:1)
         │            └─< party_contacts
         ├─< acting_capacities >── users, ?parties
         ├─< resources >── resource_types (G registry)
         ├─< relationships >── relationship_types (G) ── ends: (party|resource) refs
         ├─< actor_assignments >── roles ──< role_permissions >── permissions (G)
         │        └── scope: tenant|org|resource_type|resource
         ├─< policies ──< policy_conditions
         ├─< rule_versions >── rule_definitions (G)
         ├─< workflow_instances >── workflow_definitions ──< workflow_tasks
         ├─< documents ──< document_versions ; ──< document_access_grants
         ├─< comments ; ─< attachments        (both → resource ref)
         ├─< notifications ──< notification_deliveries ; templates (G+T)
         ├─< audit_logs (A) ; ─< events_outbox (A) ; ─< processed_events
         ├─< idempotency_keys ; ─< webhook_endpoints ──< webhook_events
         └─< integration_providers (config refs)
```

**Invariants:** (1) no tenant-scoped row without valid `tenant_id`; (2) `resources.id` equals the
owning module row id; (3) an assignment's scope refs must live in the same tenant; (4) one active rule
version per (key,scope) instant; (5) audit/outbox rows are immutable; (6) relationship ends must match
their type's declared kinds; (7) workflow instance's definition version never changes after start.

## 4. DDL skeleton (representative; full set follows the same patterns)

```sql
-- ============ roles & RLS scaffolding (000_bootstrap.sql) ============
-- app_migrate: owner, runs goose. app_rt: runtime, RLS-forced. app_platform: support ops.
CREATE FUNCTION app_tenant_id() RETURNS uuid LANGUAGE sql STABLE AS
$$ SELECT current_setting('app.tenant_id')::uuid $$;
-- helper applied to every tenant table by convention:
--   ALTER TABLE t ENABLE ROW LEVEL SECURITY;
--   ALTER TABLE t FORCE ROW LEVEL SECURITY;
--   CREATE POLICY t_tenant_isolation ON t
--     USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id());

-- ============ global spine ============
CREATE TABLE tenants (
  id            uuid PRIMARY KEY,
  slug          text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
  display_name  text NOT NULL,
  parent_tenant_id uuid REFERENCES tenants(id),
  status        text NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended','closed')),
  settings      jsonb NOT NULL DEFAULT '{}',
  version       int  NOT NULL DEFAULT 1,
  created_at    timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at    timestamptz, updated_by uuid
);

CREATE TABLE users (
  id           uuid PRIMARY KEY,
  idp_subject  text NOT NULL UNIQUE,
  email        citext NOT NULL UNIQUE,
  status       text NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
  person_party_id uuid,           -- resolved per-tenant via capacities; optional global hint
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);

CREATE TABLE user_tenant_access (
  id uuid PRIMARY KEY,
  user_id   uuid NOT NULL REFERENCES users(id),
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  kind      text NOT NULL DEFAULT 'member' CHECK (kind IN ('member','support','federated_admin')),
  status    text NOT NULL DEFAULT 'active',
  valid_from timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL
);
CREATE UNIQUE INDEX uta_active ON user_tenant_access(user_id, tenant_id, kind)
  WHERE valid_to IS NULL;

CREATE TABLE resource_types (          -- global registry, synced at boot
  key text PRIMARY KEY,                -- 'requests.request'
  module text NOT NULL, description text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE relationship_types (
  key text PRIMARY KEY,                -- 'core.owner_of'
  module text NOT NULL,
  subject_kind text NOT NULL CHECK (subject_kind IN ('party','resource','capacity')),
  object_kind  text NOT NULL CHECK (object_kind  IN ('party','resource')),
  cardinality  text NOT NULL DEFAULT 'many' CHECK (cardinality IN ('one','many')),
  description text NOT NULL
);
CREATE TABLE permissions (
  key text PRIMARY KEY,                -- 'document.read'
  module text NOT NULL, description text NOT NULL,
  sensitive boolean NOT NULL DEFAULT false      -- denials always audited when true
);

-- ============ tenant-scoped spine (RLS applied to each) ============
CREATE TABLE organizations (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL REFERENCES tenants(id),
  parent_org_id uuid REFERENCES organizations(id),
  name text NOT NULL, kind text NOT NULL DEFAULT 'org',
  status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid,
  UNIQUE (tenant_id, parent_org_id, name)
);

CREATE TABLE parties (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL REFERENCES tenants(id),
  kind text NOT NULL CHECK (kind IN ('person','legal_entity')),
  display_name text NOT NULL, status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE TABLE persons (
  party_id uuid PRIMARY KEY REFERENCES parties(id),
  tenant_id uuid NOT NULL, given_name text NOT NULL, family_name text,
  dob date, locale text
);
CREATE TABLE legal_entities (
  party_id uuid PRIMARY KEY REFERENCES parties(id),
  tenant_id uuid NOT NULL, legal_name text NOT NULL,
  registration_no text, jurisdiction text
);
CREATE TABLE party_contacts (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  party_id uuid NOT NULL REFERENCES parties(id),
  kind text NOT NULL CHECK (kind IN ('email','phone','address','other')),
  value text NOT NULL, is_primary boolean NOT NULL DEFAULT false, verified_at timestamptz,
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid,
  UNIQUE (tenant_id, party_id, kind, value)
);

CREATE TABLE acting_capacities (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  user_id uuid NOT NULL REFERENCES users(id),
  party_id uuid REFERENCES parties(id),
  label text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  valid_from timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE UNIQUE INDEX cap_active ON acting_capacities(tenant_id, user_id, label) WHERE valid_to IS NULL;

CREATE TABLE resources (               -- kernel mirror; id == module row id
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  resource_type text NOT NULL REFERENCES resource_types(key),
  org_id uuid REFERENCES organizations(id),
  label text NOT NULL,                 -- human-readable pointer for audit/UI
  status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE INDEX res_by_type ON resources(tenant_id, resource_type, status);

CREATE TABLE relationships (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  rel_type text NOT NULL REFERENCES relationship_types(key),
  subject_kind text NOT NULL, subject_id uuid NOT NULL,
  object_kind  text NOT NULL, object_id  uuid NOT NULL,
  attributes jsonb NOT NULL DEFAULT '{}',
  valid_from timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE INDEX rel_obj ON relationships(tenant_id, object_kind, object_id, rel_type) WHERE valid_to IS NULL;
CREATE INDEX rel_sub ON relationships(tenant_id, subject_kind, subject_id, rel_type) WHERE valid_to IS NULL;

CREATE TABLE roles (
  id uuid PRIMARY KEY,
  tenant_id uuid REFERENCES tenants(id),         -- NULL = platform template (global row: RLS policy OR tenant match)
  key text NOT NULL, name text NOT NULL,
  is_system boolean NOT NULL DEFAULT false,
  status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE UNIQUE INDEX roles_key ON roles(COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid), key);
CREATE TABLE role_permissions (
  role_id uuid NOT NULL REFERENCES roles(id),
  permission_key text NOT NULL REFERENCES permissions(key),
  PRIMARY KEY (role_id, permission_key)
);

CREATE TABLE actor_assignments (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  capacity_id uuid REFERENCES acting_capacities(id),
  system_actor text,                              -- exactly one of capacity_id/system_actor
  role_id uuid NOT NULL REFERENCES roles(id),
  scope_kind text NOT NULL CHECK (scope_kind IN ('tenant','org','resource_type','resource')),
  scope_id uuid,                                  -- null for tenant scope; resource_type keyed via scope_type
  scope_type text,                                -- resource_type key when scope_kind='resource_type'
  valid_from timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
  granted_by uuid NOT NULL, delegated_from uuid REFERENCES actor_assignments(id),
  reason text,
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  CHECK ((capacity_id IS NULL) <> (system_actor IS NULL))
);
CREATE INDEX asg_actor ON actor_assignments(tenant_id, capacity_id) WHERE valid_to IS NULL;
CREATE INDEX asg_scope ON actor_assignments(tenant_id, scope_kind, scope_id) WHERE valid_to IS NULL;

CREATE TABLE policies (
  id uuid PRIMARY KEY, tenant_id uuid REFERENCES tenants(id),
  key text NOT NULL, effect text NOT NULL CHECK (effect IN ('allow','deny')),
  applies_to_permission text REFERENCES permissions(key),
  applies_to_resource_type text REFERENCES resource_types(key),
  priority int NOT NULL DEFAULT 100, status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE TABLE policy_conditions (
  id uuid PRIMARY KEY, policy_id uuid NOT NULL REFERENCES policies(id),
  attribute text NOT NULL,        -- 'resource.status', 'actor.relationship', 'env.time_of_day'
  op text NOT NULL CHECK (op IN ('eq','neq','in','not_in','contains','within','gte','lte')),
  value jsonb NOT NULL
);

-- ============ rules ============
CREATE TABLE rule_definitions (
  key text PRIMARY KEY, module text NOT NULL,
  value_schema jsonb NOT NULL,            -- JSON Schema
  default_value jsonb NOT NULL,
  allowed_scopes text[] NOT NULL DEFAULT '{platform,tenant,org}',
  requires_approval boolean NOT NULL DEFAULT false,
  description text NOT NULL
);
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE TABLE rule_versions (
  id uuid PRIMARY KEY,
  rule_key text NOT NULL REFERENCES rule_definitions(key),
  tenant_id uuid,                          -- NULL for platform scope
  scope_kind text NOT NULL CHECK (scope_kind IN ('platform','tenant','org')),
  scope_id uuid,
  value jsonb NOT NULL,
  effective_from timestamptz NOT NULL, effective_to timestamptz,
  status text NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft','pending_approval','active','superseded','rejected')),
  approved_by uuid, workflow_instance_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  EXCLUDE USING gist (
    rule_key WITH =, scope_kind WITH =,
    COALESCE(scope_id,'00000000-0000-0000-0000-000000000000'::uuid) WITH =,
    tstzrange(effective_from, effective_to) WITH &&
  ) WHERE (status = 'active')
);

-- ============ workflow ============
CREATE TABLE workflow_definitions (
  id uuid PRIMARY KEY,
  key text NOT NULL, version int NOT NULL,
  tenant_id uuid,                          -- NULL = module template
  applies_to text NOT NULL REFERENCES resource_types(key),
  definition jsonb NOT NULL,               -- validated graph
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  UNIQUE (key, version, COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid))
);
CREATE TABLE workflow_instances (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  definition_id uuid NOT NULL REFERENCES workflow_definitions(id),
  resource_type text NOT NULL, resource_id uuid NOT NULL,
  current_step text NOT NULL,
  status text NOT NULL DEFAULT 'running'
    CHECK (status IN ('running','completed','rejected','cancelled','overridden')),
  context jsonb NOT NULL DEFAULT '{}',
  started_by uuid NOT NULL, ended_at timestamptz,
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE INDEX wfi_resource ON workflow_instances(tenant_id, resource_type, resource_id);
CREATE INDEX wfi_open ON workflow_instances(tenant_id, status) WHERE status = 'running';
CREATE TABLE workflow_tasks (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  instance_id uuid NOT NULL REFERENCES workflow_instances(id),
  step_key text NOT NULL, task_type text NOT NULL,
  status text NOT NULL DEFAULT 'open'
    CHECK (status IN ('open','done','approved','rejected','skipped','expired','delegated')),
  due_at timestamptz, remind_after timestamptz, last_reminded_at timestamptz,
  decided_by uuid, decided_at timestamptz, decision_comment text,
  delegated_to uuid, output jsonb,
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE TABLE workflow_task_assignees (
  task_id uuid NOT NULL REFERENCES workflow_tasks(id),
  tenant_id uuid NOT NULL,
  assignee_kind text NOT NULL CHECK (assignee_kind IN ('capacity','role','relationship','system')),
  assignee_ref text NOT NULL,              -- capacity uuid / role key / rel key
  PRIMARY KEY (task_id, assignee_kind, assignee_ref)
);
CREATE INDEX wft_due ON workflow_tasks(tenant_id, status, due_at) WHERE status = 'open';

-- ============ documents / comments / attachments ============
CREATE TABLE documents (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  document_class text NOT NULL,            -- registered by modules
  resource_type text, resource_id uuid,    -- optional anchor
  title text NOT NULL, sensitivity text NOT NULL DEFAULT 'internal'
    CHECK (sensitivity IN ('public','internal','confidential','restricted')),
  retention_until timestamptz, status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE TABLE document_versions (           -- append-only
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  document_id uuid NOT NULL REFERENCES documents(id),
  version_no int NOT NULL,
  storage_key text NOT NULL, mime_type text NOT NULL,
  size_bytes bigint NOT NULL, checksum_sha256 text NOT NULL,
  scan_status text NOT NULL DEFAULT 'pending' CHECK (scan_status IN ('pending','clean','infected','skipped')),
  uploaded_by uuid NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (document_id, version_no)
);
CREATE TABLE document_access_grants (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  document_id uuid NOT NULL REFERENCES documents(id),
  grantee_kind text NOT NULL CHECK (grantee_kind IN ('capacity','role','relationship')),
  grantee_ref text NOT NULL, access text NOT NULL CHECK (access IN ('read','write')),
  valid_from timestamptz NOT NULL DEFAULT now(), valid_to timestamptz,
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL
);
CREATE TABLE comments (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  resource_type text NOT NULL, resource_id uuid NOT NULL,
  parent_comment_id uuid REFERENCES comments(id),
  author_capacity_id uuid NOT NULL, body text NOT NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','edited','voided')),
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE INDEX cmt_resource ON comments(tenant_id, resource_type, resource_id, created_at DESC);
CREATE TABLE attachments (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  resource_type text NOT NULL, resource_id uuid NOT NULL,
  document_version_id uuid NOT NULL REFERENCES document_versions(id),
  comment_id uuid REFERENCES comments(id), workflow_task_id uuid REFERENCES workflow_tasks(id),
  status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL
);

-- ============ notifications ============
CREATE TABLE notification_templates (
  id uuid PRIMARY KEY, tenant_id uuid,     -- NULL = platform/module default
  key text NOT NULL, channel text NOT NULL CHECK (channel IN ('inapp','email','sms','whatsapp','push')),
  locale text NOT NULL DEFAULT 'en',
  subject text, body text NOT NULL,        -- Go text/template with allowlisted vars
  status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid,
  UNIQUE (COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid), key, channel, locale)
);
CREATE TABLE notifications (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  template_key text NOT NULL, recipient_party_id uuid NOT NULL,
  variables jsonb NOT NULL DEFAULT '{}',
  resource_type text, resource_id uuid,
  importance text NOT NULL DEFAULT 'normal' CHECK (importance IN ('normal','important','legal')),
  status text NOT NULL DEFAULT 'pending',
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL
);
CREATE TABLE notification_deliveries (     -- append-only; relay updates status column only
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  notification_id uuid NOT NULL REFERENCES notifications(id),
  channel text NOT NULL, destination text NOT NULL,
  status text NOT NULL DEFAULT 'queued'
    CHECK (status IN ('queued','sent','delivered','failed','dead')),
  attempts int NOT NULL DEFAULT 0, provider_message_id text, last_error text,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz
);

-- ============ audit / outbox / jobs / idempotency ============
CREATE TABLE audit_logs (                  -- append-only, monthly partitions
  id uuid NOT NULL, tenant_id uuid NOT NULL,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  actor_kind text NOT NULL, actor_user_id uuid, actor_capacity_id uuid, actor_system text,
  impersonator_user_id uuid, cross_tenant boolean NOT NULL DEFAULT false,
  action text NOT NULL,                    -- permission key or 'authz.denied', 'auth.breakglass', …
  resource_type text, resource_id uuid,
  result text NOT NULL CHECK (result IN ('success','denied','error')),
  request_id text, ip inet,
  detail jsonb NOT NULL DEFAULT '{}',      -- before/after digest, rule version ids, reason
  PRIMARY KEY (tenant_id, occurred_at, id)
) PARTITION BY RANGE (occurred_at);
-- GRANT INSERT, SELECT ON audit_logs TO app_rt;  (no UPDATE/DELETE — immutability enforced)

CREATE TABLE events_outbox (
  id uuid PRIMARY KEY,                     -- uuidv7 == event id
  tenant_id uuid NOT NULL,
  event_type text NOT NULL, schema_version int NOT NULL DEFAULT 1,
  resource_type text, resource_id uuid,
  actor jsonb NOT NULL, payload jsonb NOT NULL,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  dispatch_status text NOT NULL DEFAULT 'pending'
    CHECK (dispatch_status IN ('pending','dispatched','failed')),
  dispatched_at timestamptz, attempts int NOT NULL DEFAULT 0
);
CREATE INDEX outbox_pending ON events_outbox(occurred_at) WHERE dispatch_status = 'pending';

CREATE TABLE processed_events (            -- consumer inbox (idempotent handlers)
  handler text NOT NULL, event_id uuid NOT NULL,
  tenant_id uuid NOT NULL, processed_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (handler, event_id)
);

CREATE TABLE job_runs (                    -- reporting mirror; queue itself is River's tables
  id uuid PRIMARY KEY, tenant_id uuid,
  job_kind text NOT NULL, job_id bigint,
  status text NOT NULL CHECK (status IN ('running','succeeded','failed','dead')),
  started_at timestamptz NOT NULL DEFAULT now(), finished_at timestamptz,
  error text, progress jsonb
);

CREATE TABLE idempotency_keys (
  tenant_id uuid NOT NULL,
  actor_scope text NOT NULL,               -- capacity id or system actor
  idem_key text NOT NULL,
  request_hash text NOT NULL,
  status text NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress','completed')),
  response_status int, response_body jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  PRIMARY KEY (tenant_id, actor_scope, idem_key)
);

-- ============ integrations / webhooks ============
CREATE TABLE integration_providers (
  id uuid PRIMARY KEY, tenant_id uuid,      -- NULL = platform-registered provider kind
  key text NOT NULL, kind text NOT NULL,    -- 'payment','messaging','identity','storage','device'
  config jsonb NOT NULL DEFAULT '{}',       -- non-secret config
  credential_ref text,                      -- secret-provider key; NEVER plaintext
  status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid,
  UNIQUE (COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid), key)
);
CREATE TABLE webhook_endpoints (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  direction text NOT NULL CHECK (direction IN ('inbound','outbound')),
  provider_id uuid REFERENCES integration_providers(id),
  url text,                                 -- outbound only
  secret_ref text NOT NULL, signature_scheme text NOT NULL DEFAULT 'hmac-sha256',
  subscribed_events text[],                 -- outbound only
  status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
CREATE TABLE webhook_events (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL,
  endpoint_id uuid NOT NULL REFERENCES webhook_endpoints(id),
  direction text NOT NULL,
  external_event_id text,                   -- provider's id → replay protection
  event_type text NOT NULL, payload jsonb NOT NULL,
  signature_ok boolean, received_at timestamptz NOT NULL DEFAULT now(),
  delivery_status text NOT NULL DEFAULT 'pending'
    CHECK (delivery_status IN ('pending','processed','delivered','failed','dead')),
  attempts int NOT NULL DEFAULT 0, next_attempt_at timestamptz, last_error text,
  UNIQUE (endpoint_id, external_event_id)
);
```

## 5. Migration order (Phase 0)

1. `000` bootstrap: extensions (`citext`, `btree_gist`), roles, `app_tenant_id()` helper.
2. `001` tenants, users, user_tenant_access.
3. `002` organizations, parties, persons, legal_entities, party_contacts, acting_capacities.
4. `003` resource_types, resources, relationship_types, relationships.
5. `004` permissions, roles, role_permissions, actor_assignments, policies, policy_conditions.
6. `005` rule_definitions, rule_versions.
7. `006` workflow_definitions/instances/tasks/assignees.
8. `007` documents, document_versions, document_access_grants, comments, attachments.
9. `008` notifications trio.
10. `009` audit_logs (+ first partitions), events_outbox, processed_events, idempotency_keys, job_runs; River migrations.
11. `010` integration_providers, webhook_endpoints, webhook_events.
12. `011` RLS enablement pass over all tenant tables (kept as one reviewable migration) + grants.

Module migrations run after kernel migrations, ordered by module dependency graph, tracked in
goose's table with a per-module prefix (`society/0001_…`).

## 6. How modules add tables

A module migration creates `requests_requests(id uuid pk, tenant_id …, business columns…)` with the
same conventions (RLS, audit cols, version). Its service upserts the kernel `resources` row (same id,
type `requests.request`) inside the same tx via `resource.Registrar`. Kernel services (comments,
documents, workflow, relationships, authz record-scope) now work against it with zero kernel changes.
Cross-module access: only via the other module's Go port, or read-only views explicitly exported by
the owning module — never direct SQL joins across module tables (lint + review enforced).
