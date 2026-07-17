-- Clean v1.2.0 kernel baseline.
--
-- The database-owned DDL below is generated from PostgreSQL's catalog after
-- applying the proven historical 00001..00050 chain to an empty database. It
-- is intentionally a direct final-state construction: no abandoned upgrade,
-- staged constraint, data-rewrite, or N/N-1 choreography is retained.

-- +goose Up

-- Roles are cluster-global and therefore absent from pg_dump. Preserve the
-- bootstrap contract without resetting an operator-owned LOGIN attribute.
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_rt') THEN
        CREATE ROLE app_rt NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_platform') THEN
        CREATE ROLE app_platform NOLOGIN;
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        IF SQLERRM LIKE '%tuple concurrently updated%' THEN
            NULL;
        ELSE
            RAISE;
        END IF;
END
$$;
-- +goose StatementEnd

-- CATALOG-DERIVED UP DDL

CREATE SCHEMA migration;

CREATE EXTENSION IF NOT EXISTS btree_gist WITH SCHEMA public;

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;

CREATE FUNCTION public.app_actor_id() RETURNS uuid
    LANGUAGE sql STABLE
    AS $$
    SELECT current_setting('app.actor_id')::uuid
$$;

CREATE FUNCTION public.app_tenant_id() RETURNS uuid
    LANGUAGE sql STABLE
    AS $$ SELECT current_setting('app.tenant_id')::uuid $$;

CREATE FUNCTION public.app_tenant_id_or_null() RETURNS uuid
    LANGUAGE sql STABLE
    AS $$ SELECT nullif(current_setting('app.tenant_id', true), '')::uuid $$;

CREATE TABLE migration.backfill_checkpoint (
    job_id text NOT NULL,
    tenant_id uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid NOT NULL,
    last_key bigint DEFAULT 0 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    lease_token text,
    lease_generation bigint DEFAULT 0 NOT NULL,
    lease_expires_at timestamp with time zone
);

CREATE TABLE public.acting_capacities (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    party_id uuid,
    label text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    valid_from timestamp with time zone DEFAULT now() NOT NULL,
    valid_to timestamp with time zone,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid
);

ALTER TABLE ONLY public.acting_capacities FORCE ROW LEVEL SECURITY;

CREATE TABLE public.actor_assignments (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    capacity_id uuid,
    system_actor text,
    role_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid,
    scope_type text,
    valid_from timestamp with time zone DEFAULT now() NOT NULL,
    valid_to timestamp with time zone,
    granted_by uuid NOT NULL,
    delegated_from uuid,
    reason text,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT actor_assignments_check CHECK (((capacity_id IS NULL) <> (system_actor IS NULL))),
    CONSTRAINT actor_assignments_check1 CHECK (((scope_kind <> 'resource_type'::text) OR (scope_type IS NOT NULL))),
    CONSTRAINT actor_assignments_check2 CHECK (((scope_kind <> 'resource'::text) OR (scope_id IS NOT NULL))),
    CONSTRAINT actor_assignments_check3 CHECK (((scope_kind <> 'org'::text) OR (scope_id IS NOT NULL))),
    CONSTRAINT actor_assignments_scope_kind_check CHECK ((scope_kind = ANY (ARRAY['tenant'::text, 'org'::text, 'resource_type'::text, 'resource'::text])))
);

ALTER TABLE ONLY public.actor_assignments FORCE ROW LEVEL SECURITY;

CREATE TABLE public.api_keys (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    name text NOT NULL,
    key_prefix text NOT NULL,
    key_hash text NOT NULL,
    scopes text[] DEFAULT '{}'::text[] NOT NULL,
    expires_at timestamp with time zone,
    revoked_at timestamp with time zone,
    last_used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid
);

ALTER TABLE ONLY public.api_keys FORCE ROW LEVEL SECURITY;

CREATE TABLE public.artifacts (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    kind text NOT NULL,
    version integer NOT NULL,
    content_hash text NOT NULL,
    content bytea NOT NULL,
    content_type text DEFAULT 'application/pdf'::text NOT NULL,
    sidecar jsonb DEFAULT '{}'::jsonb NOT NULL,
    template_version text,
    effective_date timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid
);

ALTER TABLE ONLY public.artifacts FORCE ROW LEVEL SECURITY;

CREATE TABLE public.attachments (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    resource_type text NOT NULL,
    resource_id uuid NOT NULL,
    document_version_id uuid NOT NULL,
    comment_id uuid,
    workflow_task_id uuid,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT attachments_status_check CHECK ((status = ANY (ARRAY['active'::text, 'voided'::text])))
);

ALTER TABLE ONLY public.attachments FORCE ROW LEVEL SECURITY;

CREATE TABLE public.audit_anchors (
    id bigint NOT NULL,
    tenant_id uuid NOT NULL,
    anchor_seq bigint NOT NULL,
    chain_head_hash text NOT NULL,
    row_count bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.audit_anchors FORCE ROW LEVEL SECURITY;

ALTER TABLE public.audit_anchors ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.audit_anchors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

CREATE TABLE public.audit_chain (
    tenant_id uuid NOT NULL,
    next_seq bigint DEFAULT 1 NOT NULL,
    head_hash text DEFAULT ''::text NOT NULL
);

ALTER TABLE ONLY public.audit_chain FORCE ROW LEVEL SECURITY;

CREATE TABLE public.audit_logs (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    actor_id uuid,
    actor_kind text,
    impersonator_id uuid,
    request_id text,
    action text NOT NULL,
    entity_type text,
    entity_id uuid,
    field text,
    old_value text,
    new_value text,
    reason text,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    seq bigint NOT NULL,
    row_hash text NOT NULL,
    prev_hash text DEFAULT ''::text NOT NULL,
    tx_id text,
    hash_version smallint DEFAULT 1 NOT NULL
);

ALTER TABLE ONLY public.audit_logs FORCE ROW LEVEL SECURITY;

CREATE TABLE public.authz_epoch (
    tenant_id uuid NOT NULL,
    epoch integer DEFAULT 1 NOT NULL
);

ALTER TABLE ONLY public.authz_epoch FORCE ROW LEVEL SECURITY;

CREATE TABLE public.bulk_items (
    id uuid NOT NULL,
    bulk_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    seq integer NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    last_error text,
    processed_at timestamp with time zone,
    lease_token text,
    lease_generation bigint DEFAULT 0 NOT NULL,
    lease_expires_at timestamp with time zone,
    idempotency_key uuid,
    CONSTRAINT bulk_items_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'done'::text, 'failed'::text, 'cancelled'::text])))
);

ALTER TABLE ONLY public.bulk_items FORCE ROW LEVEL SECURITY;

CREATE TABLE public.bulk_operations (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    kind text NOT NULL,
    total_items integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_at timestamp with time zone,
    max_attempts integer DEFAULT 3 NOT NULL,
    CONSTRAINT bulk_operations_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'paused'::text, 'completed'::text, 'cancelled'::text])))
);

ALTER TABLE ONLY public.bulk_operations FORCE ROW LEVEL SECURITY;

CREATE TABLE public.comments (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    resource_type text NOT NULL,
    resource_id uuid NOT NULL,
    parent_comment_id uuid,
    author_capacity_id uuid NOT NULL,
    body text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT comments_status_check CHECK ((status = ANY (ARRAY['active'::text, 'edited'::text, 'voided'::text])))
);

ALTER TABLE ONLY public.comments FORCE ROW LEVEL SECURITY;

CREATE TABLE public.document_access_grants (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    document_id uuid NOT NULL,
    grantee_kind text NOT NULL,
    grantee_ref text NOT NULL,
    access text NOT NULL,
    valid_from timestamp with time zone DEFAULT now() NOT NULL,
    valid_to timestamp with time zone,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT document_access_grants_access_check CHECK ((access = ANY (ARRAY['read'::text, 'write'::text]))),
    CONSTRAINT document_access_grants_grantee_kind_check CHECK ((grantee_kind = ANY (ARRAY['capacity'::text, 'role'::text, 'relationship'::text])))
);

ALTER TABLE ONLY public.document_access_grants FORCE ROW LEVEL SECURITY;

CREATE TABLE public.document_upload_sessions (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    document_id uuid NOT NULL,
    version_no integer NOT NULL,
    storage_key text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    checksum_sha256 text,
    size_bytes bigint,
    mime_type text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_upload_sessions_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'confirmed'::text, 'expired'::text])))
);

ALTER TABLE ONLY public.document_upload_sessions FORCE ROW LEVEL SECURITY;

CREATE TABLE public.document_versions (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    document_id uuid NOT NULL,
    version_no integer NOT NULL,
    storage_key text NOT NULL,
    mime_type text NOT NULL,
    size_bytes bigint NOT NULL,
    checksum_sha256 text NOT NULL,
    scan_status text DEFAULT 'pending'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    voided_at timestamp with time zone,
    uploaded_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT document_versions_scan_status_check CHECK ((scan_status = ANY (ARRAY['pending'::text, 'clean'::text, 'infected'::text, 'skipped'::text]))),
    CONSTRAINT document_versions_status_check CHECK ((status = ANY (ARRAY['active'::text, 'voided'::text])))
);

ALTER TABLE ONLY public.document_versions FORCE ROW LEVEL SECURITY;

CREATE TABLE public.documents (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    document_class text NOT NULL,
    resource_type text,
    resource_id uuid,
    title text NOT NULL,
    sensitivity text DEFAULT 'internal'::text NOT NULL,
    retention_until timestamp with time zone,
    legal_hold boolean DEFAULT false NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT documents_sensitivity_check CHECK ((sensitivity = ANY (ARRAY['public'::text, 'internal'::text, 'confidential'::text, 'restricted'::text]))),
    CONSTRAINT documents_status_check CHECK ((status = ANY (ARRAY['active'::text, 'voided'::text])))
);

ALTER TABLE ONLY public.documents FORCE ROW LEVEL SECURITY;

CREATE TABLE public.dsr_requests (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subject_ref text NOT NULL,
    kind text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    override_reason text,
    requested_at timestamp with time zone DEFAULT now() NOT NULL,
    requested_by uuid,
    completed_at timestamp with time zone,
    CONSTRAINT dsr_requests_kind_check CHECK ((kind = ANY (ARRAY['export'::text, 'erasure'::text]))),
    CONSTRAINT dsr_requests_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'completed'::text, 'rejected'::text])))
);

ALTER TABLE ONLY public.dsr_requests FORCE ROW LEVEL SECURITY;

CREATE TABLE public.events_outbox (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    event_type text NOT NULL,
    schema_version integer DEFAULT 1 NOT NULL,
    resource_type text,
    resource_id uuid,
    actor jsonb DEFAULT '{}'::jsonb NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    occurred_at timestamp with time zone DEFAULT now() NOT NULL,
    dispatch_status text DEFAULT 'pending'::text NOT NULL,
    dispatched_at timestamp with time zone,
    failed_at timestamp with time zone,
    attempts integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 10 NOT NULL,
    last_error text,
    created_by uuid NOT NULL,
    trace_context text,
    lease_token text,
    lease_generation bigint DEFAULT 0 NOT NULL,
    lease_expires_at timestamp with time zone,
    CONSTRAINT events_outbox_dispatch_status_check CHECK ((dispatch_status = ANY (ARRAY['pending'::text, 'dispatched'::text, 'failed'::text, 'dead'::text])))
);

ALTER TABLE ONLY public.events_outbox FORCE ROW LEVEL SECURITY;

CREATE TABLE public.idempotency_keys (
    tenant_id uuid NOT NULL,
    actor_scope text NOT NULL,
    idem_key text NOT NULL,
    request_hash text NOT NULL,
    status text DEFAULT 'in_progress'::text NOT NULL,
    response_status integer,
    response_body bytea,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    CONSTRAINT idempotency_keys_status_check CHECK ((status = ANY (ARRAY['in_progress'::text, 'completed'::text])))
);

ALTER TABLE ONLY public.idempotency_keys FORCE ROW LEVEL SECURITY;

CREATE TABLE public.identity_grant (
    id uuid NOT NULL,
    status text NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    impersonated_user_id uuid,
    approver_id uuid,
    reason text,
    activated_at timestamp with time zone,
    expires_at timestamp with time zone,
    revoked_at timestamp with time zone,
    CONSTRAINT identity_grant_status_check CHECK ((status = ANY (ARRAY['active'::text, 'revoked'::text, 'expired'::text])))
);

ALTER TABLE ONLY public.identity_grant FORCE ROW LEVEL SECURITY;

CREATE TABLE public.integration_providers (
    id uuid NOT NULL,
    tenant_id uuid,
    key text NOT NULL,
    kind text NOT NULL,
    config jsonb DEFAULT '{}'::jsonb NOT NULL,
    credential_ref text,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid
);

ALTER TABLE ONLY public.integration_providers FORCE ROW LEVEL SECURITY;

CREATE TABLE public.job_runs (
    id uuid NOT NULL,
    tenant_id uuid,
    job_kind text NOT NULL,
    job_id bigint,
    status text NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    error text,
    progress jsonb,
    CONSTRAINT job_runs_status_check CHECK ((status = ANY (ARRAY['running'::text, 'succeeded'::text, 'failed'::text, 'dead'::text])))
);

ALTER TABLE ONLY public.job_runs FORCE ROW LEVEL SECURITY;

CREATE TABLE public.jobs_queue (
    id bigint NOT NULL,
    kind text NOT NULL,
    tenant_id uuid,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    status text DEFAULT 'available'::text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 5 NOT NULL,
    run_at timestamp with time zone DEFAULT now() NOT NULL,
    locked_at timestamp with time zone,
    last_error text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    trace_context text,
    lease_token text,
    lease_generation bigint DEFAULT 0 NOT NULL,
    lease_expires_at timestamp with time zone,
    CONSTRAINT jobs_queue_status_check CHECK ((status = ANY (ARRAY['available'::text, 'running'::text, 'completed'::text, 'discarded'::text])))
);

ALTER TABLE ONLY public.jobs_queue FORCE ROW LEVEL SECURITY;

ALTER TABLE public.jobs_queue ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.jobs_queue_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

CREATE TABLE public.legal_entities (
    party_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    legal_name text NOT NULL,
    registration_no text,
    jurisdiction text
);

ALTER TABLE ONLY public.legal_entities FORCE ROW LEVEL SECURITY;

CREATE TABLE public.legal_holds (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    entity_type text NOT NULL,
    entity_id uuid NOT NULL,
    reason text NOT NULL,
    placed_at timestamp with time zone DEFAULT now() NOT NULL,
    placed_by uuid,
    released_at timestamp with time zone,
    released_by uuid
);

ALTER TABLE ONLY public.legal_holds FORCE ROW LEVEL SECURITY;

CREATE TABLE public.notification_channel_prefs (
    tenant_id uuid NOT NULL,
    party_id uuid NOT NULL,
    channel text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT notification_channel_prefs_channel_check CHECK ((channel = ANY (ARRAY['inapp'::text, 'email'::text, 'sms'::text, 'whatsapp'::text, 'push'::text])))
);

ALTER TABLE ONLY public.notification_channel_prefs FORCE ROW LEVEL SECURITY;

CREATE TABLE public.notification_deliveries (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    notification_id uuid NOT NULL,
    channel text NOT NULL,
    destination text NOT NULL,
    status text DEFAULT 'queued'::text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    next_attempt_at timestamp with time zone,
    provider_message_id text,
    last_error text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    trace_context text,
    lease_token text,
    lease_generation bigint DEFAULT 0 NOT NULL,
    lease_expires_at timestamp with time zone,
    CONSTRAINT notification_deliveries_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'sent'::text, 'delivered'::text, 'failed'::text, 'dead'::text])))
);

ALTER TABLE ONLY public.notification_deliveries FORCE ROW LEVEL SECURITY;

CREATE TABLE public.notification_templates (
    id uuid NOT NULL,
    tenant_id uuid,
    key text NOT NULL,
    channel text NOT NULL,
    locale text DEFAULT 'en'::text NOT NULL,
    subject text,
    body text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT notification_templates_channel_check CHECK ((channel = ANY (ARRAY['inapp'::text, 'email'::text, 'sms'::text, 'whatsapp'::text, 'push'::text])))
);

ALTER TABLE ONLY public.notification_templates FORCE ROW LEVEL SECURITY;

CREATE TABLE public.notifications (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    template_key text NOT NULL,
    recipient_party_id uuid NOT NULL,
    variables jsonb DEFAULT '{}'::jsonb NOT NULL,
    resource_type text,
    resource_id uuid,
    importance text DEFAULT 'normal'::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT notifications_importance_check CHECK ((importance = ANY (ARRAY['normal'::text, 'important'::text, 'legal'::text])))
);

ALTER TABLE ONLY public.notifications FORCE ROW LEVEL SECURITY;

CREATE TABLE public.organizations (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    parent_org_id uuid,
    name text NOT NULL,
    kind text DEFAULT 'org'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid
);

ALTER TABLE ONLY public.organizations FORCE ROW LEVEL SECURITY;

CREATE TABLE public.parties (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    kind text NOT NULL,
    display_name text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT parties_kind_check CHECK ((kind = ANY (ARRAY['person'::text, 'legal_entity'::text])))
);

ALTER TABLE ONLY public.parties FORCE ROW LEVEL SECURITY;

CREATE TABLE public.party_contacts (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    party_id uuid NOT NULL,
    kind text NOT NULL,
    value text NOT NULL,
    is_primary boolean DEFAULT false NOT NULL,
    verified_at timestamp with time zone,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT party_contacts_kind_check CHECK ((kind = ANY (ARRAY['email'::text, 'phone'::text, 'address'::text, 'other'::text])))
);

ALTER TABLE ONLY public.party_contacts FORCE ROW LEVEL SECURITY;

CREATE TABLE public.permissions (
    key text NOT NULL,
    module text NOT NULL,
    description text NOT NULL,
    sensitive boolean DEFAULT false NOT NULL,
    step_up boolean DEFAULT false NOT NULL
);

CREATE TABLE public.persons (
    party_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    given_name text NOT NULL,
    family_name text,
    dob date,
    locale text
);

ALTER TABLE ONLY public.persons FORCE ROW LEVEL SECURITY;

CREATE TABLE public.policies (
    id uuid NOT NULL,
    tenant_id uuid,
    key text NOT NULL,
    effect text NOT NULL,
    applies_to_permission text,
    applies_to_resource_type text,
    priority integer DEFAULT 100 NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT policies_effect_check CHECK ((effect = ANY (ARRAY['allow'::text, 'deny'::text])))
);

ALTER TABLE ONLY public.policies FORCE ROW LEVEL SECURITY;

CREATE TABLE public.policy_conditions (
    id uuid NOT NULL,
    policy_id uuid NOT NULL,
    attribute text NOT NULL,
    op text NOT NULL,
    value jsonb NOT NULL,
    CONSTRAINT policy_conditions_op_check CHECK ((op = ANY (ARRAY['eq'::text, 'neq'::text, 'in'::text, 'not_in'::text, 'contains'::text, 'within'::text, 'gte'::text, 'lte'::text])))
);

CREATE TABLE public.processed_events (
    handler text NOT NULL,
    event_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    processed_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.processed_events FORCE ROW LEVEL SECURITY;

CREATE TABLE public.relationship_types (
    key text NOT NULL,
    module text NOT NULL,
    subject_kind text NOT NULL,
    object_kind text NOT NULL,
    cardinality text DEFAULT 'many'::text NOT NULL,
    description text NOT NULL,
    CONSTRAINT relationship_types_cardinality_check CHECK ((cardinality = ANY (ARRAY['one'::text, 'many'::text]))),
    CONSTRAINT relationship_types_object_kind_check CHECK ((object_kind = ANY (ARRAY['party'::text, 'resource'::text]))),
    CONSTRAINT relationship_types_subject_kind_check CHECK ((subject_kind = ANY (ARRAY['party'::text, 'resource'::text, 'capacity'::text])))
);

CREATE TABLE public.relationships (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    rel_type text NOT NULL,
    subject_kind text NOT NULL,
    subject_id uuid NOT NULL,
    object_kind text NOT NULL,
    object_id uuid NOT NULL,
    attributes jsonb DEFAULT '{}'::jsonb NOT NULL,
    valid_from timestamp with time zone DEFAULT now() NOT NULL,
    valid_to timestamp with time zone,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid
);

ALTER TABLE ONLY public.relationships FORCE ROW LEVEL SECURITY;

CREATE TABLE public.resource_types (
    key text NOT NULL,
    module text NOT NULL,
    description text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.resources (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    resource_type text NOT NULL,
    org_id uuid,
    label text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid
);

ALTER TABLE ONLY public.resources FORCE ROW LEVEL SECURITY;

CREATE TABLE public.role_permissions (
    role_id uuid NOT NULL,
    permission_key text NOT NULL
);

CREATE TABLE public.roles (
    id uuid NOT NULL,
    tenant_id uuid,
    key text NOT NULL,
    name text NOT NULL,
    is_system boolean DEFAULT false NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid
);

ALTER TABLE ONLY public.roles FORCE ROW LEVEL SECURITY;

CREATE TABLE public.rule_definitions (
    key text NOT NULL,
    module text NOT NULL,
    value_schema jsonb NOT NULL,
    default_value jsonb NOT NULL,
    allowed_scopes text[] DEFAULT '{platform,tenant,org}'::text[] NOT NULL,
    requires_approval boolean DEFAULT false NOT NULL,
    description text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.rule_versions (
    id uuid NOT NULL,
    rule_key text NOT NULL,
    tenant_id uuid,
    scope_kind text NOT NULL,
    scope_id uuid,
    value jsonb NOT NULL,
    effective_from timestamp with time zone NOT NULL,
    effective_to timestamp with time zone,
    status text DEFAULT 'draft'::text NOT NULL,
    approved_by uuid,
    workflow_instance_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT rule_versions_scope_kind_check CHECK ((scope_kind = ANY (ARRAY['platform'::text, 'tenant'::text, 'org'::text]))),
    CONSTRAINT rule_versions_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'pending_approval'::text, 'active'::text, 'superseded'::text, 'rejected'::text])))
);

ALTER TABLE ONLY public.rule_versions FORCE ROW LEVEL SECURITY;

CREATE TABLE public.schedules (
    name text NOT NULL,
    interval_seconds integer NOT NULL,
    next_run_at timestamp with time zone DEFAULT now() NOT NULL,
    last_run_at timestamp with time zone,
    enabled boolean DEFAULT true NOT NULL,
    CONSTRAINT schedules_interval_seconds_check CHECK ((interval_seconds >= 1))
);

CREATE TABLE public.seed_sync_runs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    manifest_hash text NOT NULL,
    version_label text DEFAULT ''::text NOT NULL,
    actor text DEFAULT ''::text NOT NULL,
    outcome text NOT NULL,
    counts jsonb DEFAULT '{}'::jsonb NOT NULL,
    error text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT seed_sync_runs_outcome_check CHECK ((outcome = ANY (ARRAY['applied'::text, 'noop'::text, 'failed'::text])))
);

CREATE TABLE public.sequence_allocations (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    series_key text NOT NULL,
    value bigint NOT NULL,
    allocated_at timestamp with time zone DEFAULT now() NOT NULL,
    allocated_by uuid,
    voided_at timestamp with time zone,
    void_reason text
);

ALTER TABLE ONLY public.sequence_allocations FORCE ROW LEVEL SECURITY;

CREATE TABLE public.sequences (
    tenant_id uuid NOT NULL,
    series_key text NOT NULL,
    next_value bigint DEFAULT 0 NOT NULL
);

ALTER TABLE ONLY public.sequences FORCE ROW LEVEL SECURITY;

CREATE TABLE public.tenants (
    id uuid NOT NULL,
    slug text NOT NULL,
    display_name text NOT NULL,
    parent_tenant_id uuid,
    status text DEFAULT 'active'::text NOT NULL,
    settings jsonb DEFAULT '{}'::jsonb NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT tenants_slug_check CHECK ((slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'::text)),
    CONSTRAINT tenants_status_check CHECK ((status = ANY (ARRAY['active'::text, 'suspended'::text, 'closed'::text])))
);

CREATE TABLE public.user_tenant_access (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    kind text DEFAULT 'member'::text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    valid_from timestamp with time zone DEFAULT now() NOT NULL,
    valid_to timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    CONSTRAINT user_tenant_access_kind_check CHECK ((kind = ANY (ARRAY['member'::text, 'support'::text, 'federated_admin'::text])))
);

CREATE TABLE public.users (
    id uuid NOT NULL,
    idp_subject text NOT NULL,
    email public.citext NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    person_party_id uuid,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT users_status_check CHECK ((status = ANY (ARRAY['active'::text, 'disabled'::text])))
);

CREATE TABLE public.version_counters (
    tenant_id uuid NOT NULL,
    scope text NOT NULL,
    value integer DEFAULT 0 NOT NULL
);

ALTER TABLE ONLY public.version_counters FORCE ROW LEVEL SECURITY;

CREATE TABLE public.webhook_endpoints (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    direction text NOT NULL,
    provider_id uuid,
    url text,
    secret_ref text NOT NULL,
    signature_scheme text DEFAULT 'hmac-sha256'::text NOT NULL,
    subscribed_events text[],
    status text DEFAULT 'active'::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT webhook_endpoints_direction_check CHECK ((direction = ANY (ARRAY['inbound'::text, 'outbound'::text])))
);

ALTER TABLE ONLY public.webhook_endpoints FORCE ROW LEVEL SECURITY;

CREATE TABLE public.webhook_events (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    endpoint_id uuid NOT NULL,
    direction text NOT NULL,
    external_event_id text,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    signature_ok boolean,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    delivery_status text DEFAULT 'pending'::text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    lease_token text,
    lease_generation bigint DEFAULT 0 NOT NULL,
    lease_expires_at timestamp with time zone,
    CONSTRAINT webhook_events_delivery_status_check CHECK ((delivery_status = ANY (ARRAY['pending'::text, 'processed'::text, 'delivered'::text, 'failed'::text, 'dead'::text])))
);

ALTER TABLE ONLY public.webhook_events FORCE ROW LEVEL SECURITY;

CREATE TABLE public.webhook_failed_signature_audit (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    endpoint_id uuid NOT NULL,
    event_type text NOT NULL,
    signature_header text,
    failure_reason text NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.webhook_failed_signature_audit FORCE ROW LEVEL SECURITY;

CREATE TABLE public.workflow_definitions (
    id uuid NOT NULL,
    key text NOT NULL,
    version integer NOT NULL,
    applies_to text NOT NULL,
    definition jsonb NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    definition_digest text NOT NULL,
    CONSTRAINT workflow_definitions_definition_digest_check CHECK ((definition_digest ~ '^[0-9a-f]{64}$'::text))
);

CREATE TABLE public.workflow_instances (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    definition_id uuid NOT NULL,
    resource_type text NOT NULL,
    resource_id uuid NOT NULL,
    current_step text NOT NULL,
    status text DEFAULT 'running'::text NOT NULL,
    context jsonb DEFAULT '{}'::jsonb NOT NULL,
    started_by uuid NOT NULL,
    ended_at timestamp with time zone,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT workflow_instances_status_check CHECK ((status = ANY (ARRAY['running'::text, 'completed'::text, 'rejected'::text, 'cancelled'::text, 'overridden'::text])))
);

ALTER TABLE ONLY public.workflow_instances FORCE ROW LEVEL SECURITY;

CREATE TABLE public.workflow_task_assignees (
    task_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    assignee_kind text NOT NULL,
    assignee_ref text NOT NULL,
    CONSTRAINT workflow_task_assignees_assignee_kind_check CHECK ((assignee_kind = ANY (ARRAY['capacity'::text, 'role'::text, 'relationship'::text, 'system'::text])))
);

ALTER TABLE ONLY public.workflow_task_assignees FORCE ROW LEVEL SECURITY;

CREATE TABLE public.workflow_tasks (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    instance_id uuid NOT NULL,
    step_key text NOT NULL,
    task_type text NOT NULL,
    status text DEFAULT 'open'::text NOT NULL,
    due_at timestamp with time zone,
    remind_after timestamp with time zone,
    last_reminded_at timestamp with time zone,
    decided_by uuid,
    decided_at timestamp with time zone,
    decision_comment text,
    delegated_to uuid,
    output jsonb,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid NOT NULL,
    updated_at timestamp with time zone,
    updated_by uuid,
    CONSTRAINT workflow_tasks_status_check CHECK ((status = ANY (ARRAY['open'::text, 'done'::text, 'approved'::text, 'rejected'::text, 'skipped'::text, 'expired'::text, 'delegated'::text])))
);

ALTER TABLE ONLY public.workflow_tasks FORCE ROW LEVEL SECURITY;

ALTER TABLE ONLY migration.backfill_checkpoint
    ADD CONSTRAINT backfill_checkpoint_pkey PRIMARY KEY (job_id, tenant_id);

ALTER TABLE ONLY public.acting_capacities
    ADD CONSTRAINT acting_capacities_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.actor_assignments
    ADD CONSTRAINT actor_assignments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_key_prefix_key UNIQUE (key_prefix);

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.artifacts
    ADD CONSTRAINT artifacts_tenant_id_kind_version_key UNIQUE (tenant_id, kind, version);

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.audit_anchors
    ADD CONSTRAINT audit_anchors_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.audit_chain
    ADD CONSTRAINT audit_chain_pkey PRIMARY KEY (tenant_id);

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.authz_epoch
    ADD CONSTRAINT authz_epoch_pkey PRIMARY KEY (tenant_id);

ALTER TABLE ONLY public.bulk_items
    ADD CONSTRAINT bulk_items_bulk_id_seq_key UNIQUE (bulk_id, seq);

ALTER TABLE ONLY public.bulk_items
    ADD CONSTRAINT bulk_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.bulk_operations
    ADD CONSTRAINT bulk_operations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.document_access_grants
    ADD CONSTRAINT document_access_grants_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.document_upload_sessions
    ADD CONSTRAINT document_upload_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.document_versions
    ADD CONSTRAINT document_versions_document_id_version_no_key UNIQUE (document_id, version_no);

ALTER TABLE ONLY public.document_versions
    ADD CONSTRAINT document_versions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.dsr_requests
    ADD CONSTRAINT dsr_requests_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.events_outbox
    ADD CONSTRAINT events_outbox_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.idempotency_keys
    ADD CONSTRAINT idempotency_keys_pkey PRIMARY KEY (tenant_id, actor_scope, idem_key);

ALTER TABLE ONLY public.identity_grant
    ADD CONSTRAINT identity_grant_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.integration_providers
    ADD CONSTRAINT integration_providers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.job_runs
    ADD CONSTRAINT job_runs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.jobs_queue
    ADD CONSTRAINT jobs_queue_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.legal_entities
    ADD CONSTRAINT legal_entities_pkey PRIMARY KEY (party_id);

ALTER TABLE ONLY public.legal_holds
    ADD CONSTRAINT legal_holds_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.notification_channel_prefs
    ADD CONSTRAINT notification_channel_prefs_pkey PRIMARY KEY (tenant_id, party_id, channel);

ALTER TABLE ONLY public.notification_deliveries
    ADD CONSTRAINT notification_deliveries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.notification_templates
    ADD CONSTRAINT notification_templates_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_tenant_id_parent_org_id_name_key UNIQUE (tenant_id, parent_org_id, name);

ALTER TABLE ONLY public.parties
    ADD CONSTRAINT parties_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.party_contacts
    ADD CONSTRAINT party_contacts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.party_contacts
    ADD CONSTRAINT party_contacts_tenant_id_party_id_kind_value_key UNIQUE (tenant_id, party_id, kind, value);

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (key);

ALTER TABLE ONLY public.persons
    ADD CONSTRAINT persons_pkey PRIMARY KEY (party_id);

ALTER TABLE ONLY public.policies
    ADD CONSTRAINT policies_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.policy_conditions
    ADD CONSTRAINT policy_conditions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.processed_events
    ADD CONSTRAINT processed_events_pkey PRIMARY KEY (handler, event_id);

ALTER TABLE ONLY public.relationship_types
    ADD CONSTRAINT relationship_types_pkey PRIMARY KEY (key);

ALTER TABLE ONLY public.relationships
    ADD CONSTRAINT relationships_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.resource_types
    ADD CONSTRAINT resource_types_pkey PRIMARY KEY (key);

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_key);

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.rule_definitions
    ADD CONSTRAINT rule_definitions_pkey PRIMARY KEY (key);

ALTER TABLE ONLY public.rule_versions
    ADD CONSTRAINT rule_versions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.rule_versions
    ADD CONSTRAINT rule_versions_rule_key_scope_kind_coalesce_coalesce1_tstzr_excl EXCLUDE USING gist (rule_key WITH =, scope_kind WITH =, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'::uuid) WITH =, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid) WITH =, tstzrange(effective_from, effective_to) WITH &&) WHERE ((status = 'active'::text));

ALTER TABLE ONLY public.schedules
    ADD CONSTRAINT schedules_pkey PRIMARY KEY (name);

ALTER TABLE ONLY public.seed_sync_runs
    ADD CONSTRAINT seed_sync_runs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.sequence_allocations
    ADD CONSTRAINT sequence_allocations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.sequence_allocations
    ADD CONSTRAINT sequence_allocations_tenant_id_series_key_value_key UNIQUE (tenant_id, series_key, value);

ALTER TABLE ONLY public.sequences
    ADD CONSTRAINT sequences_pkey PRIMARY KEY (tenant_id, series_key);

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_slug_key UNIQUE (slug);

ALTER TABLE ONLY public.user_tenant_access
    ADD CONSTRAINT user_tenant_access_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_idp_subject_key UNIQUE (idp_subject);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.version_counters
    ADD CONSTRAINT version_counters_pkey PRIMARY KEY (tenant_id, scope);

ALTER TABLE ONLY public.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.webhook_failed_signature_audit
    ADD CONSTRAINT webhook_failed_signature_audit_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workflow_definitions
    ADD CONSTRAINT workflow_definitions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workflow_instances
    ADD CONSTRAINT workflow_instances_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workflow_task_assignees
    ADD CONSTRAINT workflow_task_assignees_pkey PRIMARY KEY (task_id, assignee_kind, assignee_ref);

ALTER TABLE ONLY public.workflow_tasks
    ADD CONSTRAINT workflow_tasks_pkey PRIMARY KEY (id);

CREATE INDEX api_keys_tenant ON public.api_keys USING btree (tenant_id, created_at DESC);

CREATE INDEX artifacts_kind ON public.artifacts USING btree (tenant_id, kind, version DESC);

CREATE INDEX asg_actor ON public.actor_assignments USING btree (tenant_id, capacity_id, valid_from, valid_to);

CREATE INDEX asg_scope ON public.actor_assignments USING btree (tenant_id, scope_kind, scope_id, valid_to);

CREATE INDEX asg_system ON public.actor_assignments USING btree (tenant_id, system_actor, valid_from, valid_to);

CREATE INDEX att_resource ON public.attachments USING btree (tenant_id, resource_type, resource_id);

CREATE INDEX audit_anchors_tenant_seq ON public.audit_anchors USING btree (tenant_id, anchor_seq);

CREATE INDEX audit_logs_actor ON public.audit_logs USING btree (tenant_id, actor_id, occurred_at DESC);

CREATE UNIQUE INDEX audit_logs_chain ON public.audit_logs USING btree (tenant_id, seq);

CREATE INDEX audit_logs_entity ON public.audit_logs USING btree (tenant_id, entity_type, entity_id, occurred_at DESC);

CREATE INDEX bulk_items_pending ON public.bulk_items USING btree (bulk_id) WHERE (status = 'pending'::text);

CREATE UNIQUE INDEX cap_active ON public.acting_capacities USING btree (tenant_id, user_id, label) WHERE (valid_to IS NULL);

CREATE INDEX cap_user ON public.acting_capacities USING btree (tenant_id, user_id) WHERE (valid_to IS NULL);

CREATE INDEX cmt_resource ON public.comments USING btree (tenant_id, resource_type, resource_id, created_at DESC);

CREATE INDEX contacts_party ON public.party_contacts USING btree (tenant_id, party_id);

CREATE INDEX doc_class ON public.documents USING btree (tenant_id, document_class);

CREATE INDEX doc_resource ON public.documents USING btree (tenant_id, resource_type, resource_id);

CREATE INDEX doc_retention ON public.documents USING btree (tenant_id, retention_until) WHERE ((status = 'active'::text) AND (legal_hold = false));

CREATE INDEX docgrant_doc ON public.document_access_grants USING btree (tenant_id, document_id);

CREATE UNIQUE INDEX document_upload_sessions_confirmed_version ON public.document_upload_sessions USING btree (document_id, version_no) WHERE (status = 'confirmed'::text);

CREATE INDEX document_upload_sessions_document_id_idx ON public.document_upload_sessions USING btree (document_id);

CREATE INDEX document_upload_sessions_gc ON public.document_upload_sessions USING btree (tenant_id, expires_at) WHERE (status = 'pending'::text);

CREATE UNIQUE INDEX document_versions_tenant_id_id_uidx ON public.document_versions USING btree (tenant_id, id);

CREATE UNIQUE INDEX documents_tenant_id_id_uidx ON public.documents USING btree (tenant_id, id);

CREATE INDEX docver_doc ON public.document_versions USING btree (tenant_id, document_id, version_no DESC);

CREATE INDEX dsr_requests_subject ON public.dsr_requests USING btree (tenant_id, subject_ref, requested_at DESC);

CREATE INDEX idempotency_keys_expiry ON public.idempotency_keys USING btree (expires_at);

CREATE UNIQUE INDEX identity_grant_one_active_per_actor ON public.identity_grant USING btree (actor_id) WHERE (status = 'active'::text);

CREATE UNIQUE INDEX integration_providers_key ON public.integration_providers USING btree (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid), key);

CREATE INDEX job_runs_kind ON public.job_runs USING btree (job_kind, started_at);

CREATE INDEX jobs_available ON public.jobs_queue USING btree (run_at) WHERE (status = 'available'::text);

CREATE UNIQUE INDEX legal_holds_active ON public.legal_holds USING btree (tenant_id, entity_type, entity_id) WHERE (released_at IS NULL);

CREATE INDEX notif_recipient ON public.notifications USING btree (tenant_id, recipient_party_id, created_at DESC);

CREATE INDEX notifdel_pending ON public.notification_deliveries USING btree (tenant_id, next_attempt_at) WHERE (status = ANY (ARRAY['queued'::text, 'failed'::text]));

CREATE UNIQUE INDEX notification_templates_key ON public.notification_templates USING btree (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid), key, channel, locale);

CREATE INDEX org_parent ON public.organizations USING btree (tenant_id, parent_org_id);

CREATE UNIQUE INDEX organizations_tenant_id_id_uidx ON public.organizations USING btree (tenant_id, id);

CREATE INDEX outbox_aggregate ON public.events_outbox USING btree (tenant_id, resource_type, resource_id, occurred_at);

CREATE INDEX outbox_pending ON public.events_outbox USING btree (occurred_at) WHERE (dispatch_status = 'pending'::text);

CREATE UNIQUE INDEX parties_tenant_id_id_uidx ON public.parties USING btree (tenant_id, id);

CREATE INDEX rel_obj ON public.relationships USING btree (tenant_id, object_kind, object_id, rel_type, valid_from, valid_to);

CREATE INDEX rel_sub ON public.relationships USING btree (tenant_id, subject_kind, subject_id, rel_type, valid_from, valid_to);

CREATE INDEX res_by_type ON public.resources USING btree (tenant_id, resource_type, status);

CREATE UNIQUE INDEX roles_key ON public.roles USING btree (COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid), key);

CREATE INDEX rule_versions_history_resolution_idx ON public.rule_versions USING btree (rule_key, scope_kind, scope_id, tenant_id, effective_from DESC) WHERE (status = ANY (ARRAY['active'::text, 'superseded'::text]));

CREATE INDEX seed_sync_runs_created_at_idx ON public.seed_sync_runs USING btree (created_at DESC);

CREATE INDEX seed_sync_runs_hash_idx ON public.seed_sync_runs USING btree (manifest_hash);

CREATE UNIQUE INDEX uta_active ON public.user_tenant_access USING btree (user_id, tenant_id, kind) WHERE (valid_to IS NULL);

CREATE UNIQUE INDEX webhook_endpoints_tenant_id_id_uidx ON public.webhook_endpoints USING btree (tenant_id, id);

CREATE UNIQUE INDEX webhook_events_dedup ON public.webhook_events USING btree (endpoint_id, external_event_id) WHERE (external_event_id IS NOT NULL);

CREATE INDEX wfi_open ON public.workflow_instances USING btree (tenant_id, status) WHERE (status = 'running'::text);

CREATE INDEX wfi_resource ON public.workflow_instances USING btree (tenant_id, resource_type, resource_id);

CREATE INDEX wfsa_endpoint ON public.webhook_failed_signature_audit USING btree (tenant_id, endpoint_id, received_at DESC);

CREATE INDEX wft_due ON public.workflow_tasks USING btree (tenant_id, status, due_at) WHERE (status = 'open'::text);

CREATE INDEX wft_remind_after ON public.workflow_tasks USING btree (tenant_id, remind_after, id) WHERE ((status = 'open'::text) AND (remind_after IS NOT NULL) AND ((last_reminded_at IS NULL) OR (last_reminded_at < remind_after)));

CREATE INDEX whep_direction ON public.webhook_endpoints USING btree (tenant_id, direction, status);

CREATE INDEX whev_pending ON public.webhook_events USING btree (tenant_id, delivery_status, next_attempt_at) WHERE (delivery_status = ANY (ARRAY['pending'::text, 'failed'::text]));

CREATE UNIQUE INDEX workflow_definitions_key_version ON public.workflow_definitions USING btree (key, version);

ALTER TABLE ONLY public.acting_capacities
    ADD CONSTRAINT acting_capacities_party_id_tenant_fkey FOREIGN KEY (tenant_id, party_id) REFERENCES public.parties(tenant_id, id);

ALTER TABLE ONLY public.acting_capacities
    ADD CONSTRAINT acting_capacities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.actor_assignments
    ADD CONSTRAINT actor_assignments_capacity_id_fkey FOREIGN KEY (capacity_id) REFERENCES public.acting_capacities(id);

ALTER TABLE ONLY public.actor_assignments
    ADD CONSTRAINT actor_assignments_delegated_from_fkey FOREIGN KEY (delegated_from) REFERENCES public.actor_assignments(id);

ALTER TABLE ONLY public.actor_assignments
    ADD CONSTRAINT actor_assignments_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id);

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.comments(id);

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_document_version_id_tenant_fkey FOREIGN KEY (tenant_id, document_version_id) REFERENCES public.document_versions(tenant_id, id);

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_workflow_task_id_fkey FOREIGN KEY (workflow_task_id) REFERENCES public.workflow_tasks(id);

ALTER TABLE ONLY public.bulk_items
    ADD CONSTRAINT bulk_items_bulk_id_fkey FOREIGN KEY (bulk_id) REFERENCES public.bulk_operations(id);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_parent_comment_id_fkey FOREIGN KEY (parent_comment_id) REFERENCES public.comments(id);

ALTER TABLE ONLY public.document_access_grants
    ADD CONSTRAINT document_access_grants_document_id_tenant_fkey FOREIGN KEY (tenant_id, document_id) REFERENCES public.documents(tenant_id, id);

ALTER TABLE ONLY public.document_upload_sessions
    ADD CONSTRAINT document_upload_sessions_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id);

ALTER TABLE ONLY public.document_versions
    ADD CONSTRAINT document_versions_document_id_tenant_fkey FOREIGN KEY (tenant_id, document_id) REFERENCES public.documents(tenant_id, id);

ALTER TABLE ONLY public.legal_entities
    ADD CONSTRAINT legal_entities_party_id_tenant_fkey FOREIGN KEY (tenant_id, party_id) REFERENCES public.parties(tenant_id, id);

ALTER TABLE ONLY public.notification_deliveries
    ADD CONSTRAINT notification_deliveries_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES public.notifications(id);

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_parent_org_id_fkey FOREIGN KEY (parent_org_id) REFERENCES public.organizations(id);

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);

ALTER TABLE ONLY public.parties
    ADD CONSTRAINT parties_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);

ALTER TABLE ONLY public.party_contacts
    ADD CONSTRAINT party_contacts_party_id_tenant_fkey FOREIGN KEY (tenant_id, party_id) REFERENCES public.parties(tenant_id, id);

ALTER TABLE ONLY public.persons
    ADD CONSTRAINT persons_party_id_tenant_fkey FOREIGN KEY (tenant_id, party_id) REFERENCES public.parties(tenant_id, id);

ALTER TABLE ONLY public.policies
    ADD CONSTRAINT policies_applies_to_permission_fkey FOREIGN KEY (applies_to_permission) REFERENCES public.permissions(key);

ALTER TABLE ONLY public.policies
    ADD CONSTRAINT policies_applies_to_resource_type_fkey FOREIGN KEY (applies_to_resource_type) REFERENCES public.resource_types(key);

ALTER TABLE ONLY public.policies
    ADD CONSTRAINT policies_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);

ALTER TABLE ONLY public.policy_conditions
    ADD CONSTRAINT policy_conditions_policy_id_fkey FOREIGN KEY (policy_id) REFERENCES public.policies(id);

ALTER TABLE ONLY public.relationships
    ADD CONSTRAINT relationships_rel_type_fkey FOREIGN KEY (rel_type) REFERENCES public.relationship_types(key);

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_org_id_tenant_fkey FOREIGN KEY (tenant_id, org_id) REFERENCES public.organizations(tenant_id, id);

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_resource_type_fkey FOREIGN KEY (resource_type) REFERENCES public.resource_types(key);

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_key_fkey FOREIGN KEY (permission_key) REFERENCES public.permissions(key);

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id);

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);

ALTER TABLE ONLY public.rule_versions
    ADD CONSTRAINT rule_versions_rule_key_fkey FOREIGN KEY (rule_key) REFERENCES public.rule_definitions(key);

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_parent_tenant_id_fkey FOREIGN KEY (parent_tenant_id) REFERENCES public.tenants(id);

ALTER TABLE ONLY public.user_tenant_access
    ADD CONSTRAINT user_tenant_access_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);

ALTER TABLE ONLY public.user_tenant_access
    ADD CONSTRAINT user_tenant_access_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.integration_providers(id);

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_endpoint_id_fkey FOREIGN KEY (endpoint_id) REFERENCES public.webhook_endpoints(id);

ALTER TABLE ONLY public.webhook_failed_signature_audit
    ADD CONSTRAINT webhook_failed_signature_audit_tenant_id_endpoint_id_fkey FOREIGN KEY (tenant_id, endpoint_id) REFERENCES public.webhook_endpoints(tenant_id, id);

ALTER TABLE ONLY public.workflow_instances
    ADD CONSTRAINT workflow_instances_definition_id_fkey FOREIGN KEY (definition_id) REFERENCES public.workflow_definitions(id);

ALTER TABLE ONLY public.workflow_task_assignees
    ADD CONSTRAINT workflow_task_assignees_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.workflow_tasks(id);

ALTER TABLE ONLY public.workflow_tasks
    ADD CONSTRAINT workflow_tasks_instance_id_fkey FOREIGN KEY (instance_id) REFERENCES public.workflow_instances(id);

ALTER TABLE public.acting_capacities ENABLE ROW LEVEL SECURITY;

CREATE POLICY acting_capacities_tenant_isolation ON public.acting_capacities USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.actor_assignments ENABLE ROW LEVEL SECURITY;

CREATE POLICY actor_assignments_tenant_isolation ON public.actor_assignments USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.api_keys ENABLE ROW LEVEL SECURITY;

CREATE POLICY api_keys_platform_all ON public.api_keys TO app_platform USING (true) WITH CHECK (true);

CREATE POLICY api_keys_tenant_isolation ON public.api_keys USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.artifacts ENABLE ROW LEVEL SECURITY;

CREATE POLICY artifacts_tenant_isolation ON public.artifacts USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.attachments ENABLE ROW LEVEL SECURITY;

CREATE POLICY attachments_tenant_isolation ON public.attachments USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.audit_anchors ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_anchors_platform_write ON public.audit_anchors TO app_platform USING (true) WITH CHECK (true);

CREATE POLICY audit_anchors_tenant_read ON public.audit_anchors FOR SELECT TO app_rt USING ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.audit_chain ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_chain_platform_read ON public.audit_chain FOR SELECT TO app_platform USING (true);

CREATE POLICY audit_chain_tenant_isolation ON public.audit_chain USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.audit_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_logs_tenant_isolation ON public.audit_logs USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.authz_epoch ENABLE ROW LEVEL SECURITY;

CREATE POLICY authz_epoch_tenant_isolation ON public.authz_epoch USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.bulk_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY bulk_items_tenant_isolation ON public.bulk_items USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.bulk_operations ENABLE ROW LEVEL SECURITY;

CREATE POLICY bulk_operations_tenant_isolation ON public.bulk_operations USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.comments ENABLE ROW LEVEL SECURITY;

CREATE POLICY comments_tenant_isolation ON public.comments USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.document_access_grants ENABLE ROW LEVEL SECURITY;

CREATE POLICY document_access_grants_owner_write ON public.document_access_grants AS RESTRICTIVE USING (true) WITH CHECK ((EXISTS ( SELECT 1
   FROM public.documents d
  WHERE ((d.id = document_access_grants.document_id) AND (d.created_by = public.app_actor_id())))));

CREATE POLICY document_access_grants_tenant_isolation ON public.document_access_grants USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.document_upload_sessions ENABLE ROW LEVEL SECURITY;

CREATE POLICY document_upload_sessions_tenant_isolation ON public.document_upload_sessions USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.document_versions ENABLE ROW LEVEL SECURITY;

CREATE POLICY document_versions_tenant_isolation ON public.document_versions USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY documents_tenant_isolation ON public.documents USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.dsr_requests ENABLE ROW LEVEL SECURITY;

CREATE POLICY dsr_requests_tenant_isolation ON public.dsr_requests USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.events_outbox ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.idempotency_keys ENABLE ROW LEVEL SECURITY;

CREATE POLICY idempotency_keys_platform_sweep ON public.idempotency_keys TO app_platform USING (true) WITH CHECK (true);

CREATE POLICY idempotency_keys_tenant_isolation ON public.idempotency_keys USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.identity_grant ENABLE ROW LEVEL SECURITY;

CREATE POLICY identity_grant_platform_all ON public.identity_grant TO app_platform USING (true) WITH CHECK (true);

CREATE POLICY identity_grant_tenant_isolation ON public.identity_grant USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.integration_providers ENABLE ROW LEVEL SECURITY;

CREATE POLICY integration_providers_platform_write ON public.integration_providers AS RESTRICTIVE USING (true) WITH CHECK (((tenant_id IS NOT NULL) OR (public.app_tenant_id_or_null() IS NULL)));

CREATE POLICY integration_providers_tenant ON public.integration_providers USING (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null()))) WITH CHECK (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null())));

ALTER TABLE public.job_runs ENABLE ROW LEVEL SECURITY;

CREATE POLICY job_runs_platform_all ON public.job_runs TO app_platform USING (true) WITH CHECK (true);

CREATE POLICY job_runs_tenant_isolation ON public.job_runs USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.jobs_queue ENABLE ROW LEVEL SECURITY;

CREATE POLICY jobs_queue_platform_all ON public.jobs_queue TO app_platform USING (true) WITH CHECK (true);

CREATE POLICY jobs_queue_tenant_isolation ON public.jobs_queue USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.legal_entities ENABLE ROW LEVEL SECURITY;

CREATE POLICY legal_entities_tenant_isolation ON public.legal_entities USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.legal_holds ENABLE ROW LEVEL SECURITY;

CREATE POLICY legal_holds_tenant_isolation ON public.legal_holds USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.notification_channel_prefs ENABLE ROW LEVEL SECURITY;

CREATE POLICY notification_channel_prefs_tenant_isolation ON public.notification_channel_prefs USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.notification_deliveries ENABLE ROW LEVEL SECURITY;

CREATE POLICY notification_deliveries_tenant_isolation ON public.notification_deliveries USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.notification_templates ENABLE ROW LEVEL SECURITY;

CREATE POLICY notification_templates_platform_write ON public.notification_templates AS RESTRICTIVE USING (true) WITH CHECK (((tenant_id IS NOT NULL) OR (public.app_tenant_id_or_null() IS NULL)));

CREATE POLICY notification_templates_tenant ON public.notification_templates USING (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null()))) WITH CHECK (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null())));

ALTER TABLE public.notifications ENABLE ROW LEVEL SECURITY;

CREATE POLICY notifications_tenant_isolation ON public.notifications USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.organizations ENABLE ROW LEVEL SECURITY;

CREATE POLICY organizations_tenant_isolation ON public.organizations USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

CREATE POLICY outbox_relay_all ON public.events_outbox TO app_platform USING (true) WITH CHECK (true);

CREATE POLICY outbox_tenant_isolation ON public.events_outbox USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.parties ENABLE ROW LEVEL SECURITY;

CREATE POLICY parties_tenant_isolation ON public.parties USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.party_contacts ENABLE ROW LEVEL SECURITY;

CREATE POLICY party_contacts_tenant_isolation ON public.party_contacts USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.persons ENABLE ROW LEVEL SECURITY;

CREATE POLICY persons_tenant_isolation ON public.persons USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.policies ENABLE ROW LEVEL SECURITY;

CREATE POLICY policies_tenant_read ON public.policies USING (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null()))) WITH CHECK (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null())));

ALTER TABLE public.processed_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY processed_events_tenant_isolation ON public.processed_events USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.relationships ENABLE ROW LEVEL SECURITY;

CREATE POLICY relationships_tenant_isolation ON public.relationships USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.resources ENABLE ROW LEVEL SECURITY;

CREATE POLICY resources_tenant_isolation ON public.resources USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY roles_tenant_read ON public.roles USING (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null()))) WITH CHECK (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null())));

ALTER TABLE public.rule_versions ENABLE ROW LEVEL SECURITY;

CREATE POLICY rule_versions_platform_all ON public.rule_versions TO app_platform USING (true) WITH CHECK (true);

CREATE POLICY rule_versions_tenant ON public.rule_versions USING (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null()))) WITH CHECK (((tenant_id IS NULL) OR (tenant_id = public.app_tenant_id_or_null())));

ALTER TABLE public.sequence_allocations ENABLE ROW LEVEL SECURITY;

CREATE POLICY sequence_allocations_tenant_isolation ON public.sequence_allocations USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.sequences ENABLE ROW LEVEL SECURITY;

CREATE POLICY sequences_tenant_isolation ON public.sequences USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.version_counters ENABLE ROW LEVEL SECURITY;

CREATE POLICY version_counters_tenant_isolation ON public.version_counters USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.webhook_endpoints ENABLE ROW LEVEL SECURITY;

CREATE POLICY webhook_endpoints_tenant_isolation ON public.webhook_endpoints USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.webhook_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY webhook_events_tenant_isolation ON public.webhook_events USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.webhook_failed_signature_audit ENABLE ROW LEVEL SECURITY;

CREATE POLICY webhook_failed_signature_audit_tenant_isolation ON public.webhook_failed_signature_audit USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.workflow_instances ENABLE ROW LEVEL SECURITY;

CREATE POLICY workflow_instances_tenant_isolation ON public.workflow_instances USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.workflow_task_assignees ENABLE ROW LEVEL SECURITY;

CREATE POLICY workflow_task_assignees_tenant_isolation ON public.workflow_task_assignees USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

ALTER TABLE public.workflow_tasks ENABLE ROW LEVEL SECURITY;

CREATE POLICY workflow_tasks_tenant_isolation ON public.workflow_tasks USING ((tenant_id = public.app_tenant_id())) WITH CHECK ((tenant_id = public.app_tenant_id()));

GRANT USAGE ON SCHEMA public TO app_rt;
GRANT USAGE ON SCHEMA public TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.acting_capacities TO app_rt;
GRANT SELECT ON TABLE public.acting_capacities TO app_platform;

GRANT SELECT ON TABLE public.actor_assignments TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.actor_assignments TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.api_keys TO app_rt;
GRANT SELECT,UPDATE ON TABLE public.api_keys TO app_platform;

GRANT SELECT,INSERT ON TABLE public.artifacts TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.attachments TO app_rt;

GRANT SELECT ON TABLE public.audit_anchors TO app_rt;
GRANT SELECT,INSERT ON TABLE public.audit_anchors TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.audit_chain TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.audit_chain TO app_platform;

GRANT SELECT,INSERT ON TABLE public.audit_logs TO app_rt;
GRANT SELECT,INSERT ON TABLE public.audit_logs TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.authz_epoch TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.authz_epoch TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.bulk_items TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.bulk_operations TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.comments TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.document_access_grants TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.document_upload_sessions TO app_rt;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.document_upload_sessions TO app_platform;

GRANT SELECT,INSERT ON TABLE public.document_versions TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.document_versions TO app_platform;

GRANT SELECT,INSERT ON TABLE public.documents TO app_rt;
GRANT SELECT,UPDATE ON TABLE public.documents TO app_platform;

GRANT UPDATE(title) ON TABLE public.documents TO app_rt;

GRANT UPDATE(sensitivity) ON TABLE public.documents TO app_rt;

GRANT UPDATE(version) ON TABLE public.documents TO app_rt;

GRANT UPDATE(updated_at) ON TABLE public.documents TO app_rt;

GRANT UPDATE(updated_by) ON TABLE public.documents TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.dsr_requests TO app_rt;

GRANT SELECT,INSERT ON TABLE public.events_outbox TO app_rt;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.events_outbox TO app_platform;

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.idempotency_keys TO app_rt;
GRANT SELECT,DELETE ON TABLE public.idempotency_keys TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.identity_grant TO app_platform;

GRANT SELECT ON TABLE public.integration_providers TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.integration_providers TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.job_runs TO app_platform;

GRANT INSERT ON TABLE public.jobs_queue TO app_rt;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.jobs_queue TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.legal_entities TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.legal_holds TO app_rt;

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.notification_channel_prefs TO app_rt;

GRANT SELECT,INSERT ON TABLE public.notification_deliveries TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.notification_deliveries TO app_platform;

GRANT SELECT ON TABLE public.notification_templates TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.notification_templates TO app_platform;

GRANT SELECT,INSERT ON TABLE public.notifications TO app_rt;
GRANT SELECT,UPDATE ON TABLE public.notifications TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.organizations TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.parties TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.party_contacts TO app_rt;

GRANT SELECT ON TABLE public.permissions TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.permissions TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.persons TO app_rt;

GRANT SELECT ON TABLE public.policies TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.policies TO app_platform;

GRANT SELECT ON TABLE public.policy_conditions TO app_rt;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.policy_conditions TO app_platform;

GRANT SELECT,INSERT ON TABLE public.processed_events TO app_rt;

GRANT SELECT ON TABLE public.relationship_types TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.relationship_types TO app_platform;

GRANT SELECT ON TABLE public.relationships TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.relationships TO app_platform;

GRANT SELECT ON TABLE public.resource_types TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.resource_types TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.resources TO app_rt;
GRANT SELECT ON TABLE public.resources TO app_platform;

GRANT SELECT ON TABLE public.role_permissions TO app_rt;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.role_permissions TO app_platform;

GRANT SELECT ON TABLE public.roles TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.roles TO app_platform;

GRANT SELECT ON TABLE public.rule_definitions TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.rule_definitions TO app_platform;

GRANT SELECT,INSERT ON TABLE public.rule_versions TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.rule_versions TO app_platform;

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.schedules TO app_platform;

GRANT SELECT,INSERT ON TABLE public.seed_sync_runs TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.sequence_allocations TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.sequences TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.tenants TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.user_tenant_access TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.users TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.version_counters TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.version_counters TO app_platform;

GRANT SELECT ON TABLE public.webhook_endpoints TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.webhook_endpoints TO app_platform;

GRANT SELECT,INSERT ON TABLE public.webhook_events TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.webhook_events TO app_platform;

GRANT SELECT,INSERT ON TABLE public.webhook_failed_signature_audit TO app_rt;
GRANT SELECT,INSERT ON TABLE public.webhook_failed_signature_audit TO app_platform;

GRANT SELECT ON TABLE public.workflow_definitions TO app_rt;
GRANT SELECT,INSERT,UPDATE ON TABLE public.workflow_definitions TO app_platform;

GRANT SELECT,INSERT,UPDATE ON TABLE public.workflow_instances TO app_rt;

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.workflow_task_assignees TO app_rt;

GRANT SELECT,INSERT,UPDATE ON TABLE public.workflow_tasks TO app_rt;


-- goose creates the source history table before executing this migration.
GRANT SELECT ON public.goose_version_wowapi TO app_platform;

-- +goose Down

-- The history table belongs to the migration runner, not the kernel schema.
REVOKE SELECT ON public.goose_version_wowapi FROM app_platform;

-- CATALOG-DERIVED DOWN DDL
DROP POLICY IF EXISTS workflow_tasks_tenant_isolation ON public.workflow_tasks;
DROP POLICY IF EXISTS workflow_task_assignees_tenant_isolation ON public.workflow_task_assignees;
DROP POLICY IF EXISTS workflow_instances_tenant_isolation ON public.workflow_instances;
DROP POLICY IF EXISTS webhook_failed_signature_audit_tenant_isolation ON public.webhook_failed_signature_audit;
DROP POLICY IF EXISTS webhook_events_tenant_isolation ON public.webhook_events;
DROP POLICY IF EXISTS webhook_endpoints_tenant_isolation ON public.webhook_endpoints;
DROP POLICY IF EXISTS version_counters_tenant_isolation ON public.version_counters;
DROP POLICY IF EXISTS sequences_tenant_isolation ON public.sequences;
DROP POLICY IF EXISTS sequence_allocations_tenant_isolation ON public.sequence_allocations;
DROP POLICY IF EXISTS rule_versions_tenant ON public.rule_versions;
DROP POLICY IF EXISTS rule_versions_platform_all ON public.rule_versions;
DROP POLICY IF EXISTS roles_tenant_read ON public.roles;
DROP POLICY IF EXISTS resources_tenant_isolation ON public.resources;
DROP POLICY IF EXISTS relationships_tenant_isolation ON public.relationships;
DROP POLICY IF EXISTS processed_events_tenant_isolation ON public.processed_events;
DROP POLICY IF EXISTS policies_tenant_read ON public.policies;
DROP POLICY IF EXISTS persons_tenant_isolation ON public.persons;
DROP POLICY IF EXISTS party_contacts_tenant_isolation ON public.party_contacts;
DROP POLICY IF EXISTS parties_tenant_isolation ON public.parties;
DROP POLICY IF EXISTS outbox_tenant_isolation ON public.events_outbox;
DROP POLICY IF EXISTS outbox_relay_all ON public.events_outbox;
DROP POLICY IF EXISTS organizations_tenant_isolation ON public.organizations;
DROP POLICY IF EXISTS notifications_tenant_isolation ON public.notifications;
DROP POLICY IF EXISTS notification_templates_tenant ON public.notification_templates;
DROP POLICY IF EXISTS notification_templates_platform_write ON public.notification_templates;
DROP POLICY IF EXISTS notification_deliveries_tenant_isolation ON public.notification_deliveries;
DROP POLICY IF EXISTS notification_channel_prefs_tenant_isolation ON public.notification_channel_prefs;
DROP POLICY IF EXISTS legal_holds_tenant_isolation ON public.legal_holds;
DROP POLICY IF EXISTS legal_entities_tenant_isolation ON public.legal_entities;
DROP POLICY IF EXISTS jobs_queue_tenant_isolation ON public.jobs_queue;
DROP POLICY IF EXISTS jobs_queue_platform_all ON public.jobs_queue;
DROP POLICY IF EXISTS job_runs_tenant_isolation ON public.job_runs;
DROP POLICY IF EXISTS job_runs_platform_all ON public.job_runs;
DROP POLICY IF EXISTS integration_providers_tenant ON public.integration_providers;
DROP POLICY IF EXISTS integration_providers_platform_write ON public.integration_providers;
DROP POLICY IF EXISTS identity_grant_tenant_isolation ON public.identity_grant;
DROP POLICY IF EXISTS identity_grant_platform_all ON public.identity_grant;
DROP POLICY IF EXISTS idempotency_keys_tenant_isolation ON public.idempotency_keys;
DROP POLICY IF EXISTS idempotency_keys_platform_sweep ON public.idempotency_keys;
DROP POLICY IF EXISTS dsr_requests_tenant_isolation ON public.dsr_requests;
DROP POLICY IF EXISTS documents_tenant_isolation ON public.documents;
DROP POLICY IF EXISTS document_versions_tenant_isolation ON public.document_versions;
DROP POLICY IF EXISTS document_upload_sessions_tenant_isolation ON public.document_upload_sessions;
DROP POLICY IF EXISTS document_access_grants_tenant_isolation ON public.document_access_grants;
DROP POLICY IF EXISTS document_access_grants_owner_write ON public.document_access_grants;
DROP POLICY IF EXISTS comments_tenant_isolation ON public.comments;
DROP POLICY IF EXISTS bulk_operations_tenant_isolation ON public.bulk_operations;
DROP POLICY IF EXISTS bulk_items_tenant_isolation ON public.bulk_items;
DROP POLICY IF EXISTS authz_epoch_tenant_isolation ON public.authz_epoch;
DROP POLICY IF EXISTS audit_logs_tenant_isolation ON public.audit_logs;
DROP POLICY IF EXISTS audit_chain_tenant_isolation ON public.audit_chain;
DROP POLICY IF EXISTS audit_chain_platform_read ON public.audit_chain;
DROP POLICY IF EXISTS audit_anchors_tenant_read ON public.audit_anchors;
DROP POLICY IF EXISTS audit_anchors_platform_write ON public.audit_anchors;
DROP POLICY IF EXISTS attachments_tenant_isolation ON public.attachments;
DROP POLICY IF EXISTS artifacts_tenant_isolation ON public.artifacts;
DROP POLICY IF EXISTS api_keys_tenant_isolation ON public.api_keys;
DROP POLICY IF EXISTS api_keys_platform_all ON public.api_keys;
DROP POLICY IF EXISTS actor_assignments_tenant_isolation ON public.actor_assignments;
DROP POLICY IF EXISTS acting_capacities_tenant_isolation ON public.acting_capacities;
ALTER TABLE IF EXISTS ONLY public.workflow_tasks DROP CONSTRAINT IF EXISTS workflow_tasks_instance_id_fkey;
ALTER TABLE IF EXISTS ONLY public.workflow_task_assignees DROP CONSTRAINT IF EXISTS workflow_task_assignees_task_id_fkey;
ALTER TABLE IF EXISTS ONLY public.workflow_instances DROP CONSTRAINT IF EXISTS workflow_instances_definition_id_fkey;
ALTER TABLE IF EXISTS ONLY public.webhook_failed_signature_audit DROP CONSTRAINT IF EXISTS webhook_failed_signature_audit_tenant_id_endpoint_id_fkey;
ALTER TABLE IF EXISTS ONLY public.webhook_events DROP CONSTRAINT IF EXISTS webhook_events_endpoint_id_fkey;
ALTER TABLE IF EXISTS ONLY public.webhook_endpoints DROP CONSTRAINT IF EXISTS webhook_endpoints_provider_id_fkey;
ALTER TABLE IF EXISTS ONLY public.user_tenant_access DROP CONSTRAINT IF EXISTS user_tenant_access_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.user_tenant_access DROP CONSTRAINT IF EXISTS user_tenant_access_tenant_id_fkey;
ALTER TABLE IF EXISTS ONLY public.tenants DROP CONSTRAINT IF EXISTS tenants_parent_tenant_id_fkey;
ALTER TABLE IF EXISTS ONLY public.rule_versions DROP CONSTRAINT IF EXISTS rule_versions_rule_key_fkey;
ALTER TABLE IF EXISTS ONLY public.roles DROP CONSTRAINT IF EXISTS roles_tenant_id_fkey;
ALTER TABLE IF EXISTS ONLY public.role_permissions DROP CONSTRAINT IF EXISTS role_permissions_role_id_fkey;
ALTER TABLE IF EXISTS ONLY public.role_permissions DROP CONSTRAINT IF EXISTS role_permissions_permission_key_fkey;
ALTER TABLE IF EXISTS ONLY public.resources DROP CONSTRAINT IF EXISTS resources_resource_type_fkey;
ALTER TABLE IF EXISTS ONLY public.resources DROP CONSTRAINT IF EXISTS resources_org_id_tenant_fkey;
ALTER TABLE IF EXISTS ONLY public.relationships DROP CONSTRAINT IF EXISTS relationships_rel_type_fkey;
ALTER TABLE IF EXISTS ONLY public.policy_conditions DROP CONSTRAINT IF EXISTS policy_conditions_policy_id_fkey;
ALTER TABLE IF EXISTS ONLY public.policies DROP CONSTRAINT IF EXISTS policies_tenant_id_fkey;
ALTER TABLE IF EXISTS ONLY public.policies DROP CONSTRAINT IF EXISTS policies_applies_to_resource_type_fkey;
ALTER TABLE IF EXISTS ONLY public.policies DROP CONSTRAINT IF EXISTS policies_applies_to_permission_fkey;
ALTER TABLE IF EXISTS ONLY public.persons DROP CONSTRAINT IF EXISTS persons_party_id_tenant_fkey;
ALTER TABLE IF EXISTS ONLY public.party_contacts DROP CONSTRAINT IF EXISTS party_contacts_party_id_tenant_fkey;
ALTER TABLE IF EXISTS ONLY public.parties DROP CONSTRAINT IF EXISTS parties_tenant_id_fkey;
ALTER TABLE IF EXISTS ONLY public.organizations DROP CONSTRAINT IF EXISTS organizations_tenant_id_fkey;
ALTER TABLE IF EXISTS ONLY public.organizations DROP CONSTRAINT IF EXISTS organizations_parent_org_id_fkey;
ALTER TABLE IF EXISTS ONLY public.notification_deliveries DROP CONSTRAINT IF EXISTS notification_deliveries_notification_id_fkey;
ALTER TABLE IF EXISTS ONLY public.legal_entities DROP CONSTRAINT IF EXISTS legal_entities_party_id_tenant_fkey;
ALTER TABLE IF EXISTS ONLY public.document_versions DROP CONSTRAINT IF EXISTS document_versions_document_id_tenant_fkey;
ALTER TABLE IF EXISTS ONLY public.document_upload_sessions DROP CONSTRAINT IF EXISTS document_upload_sessions_document_id_fkey;
ALTER TABLE IF EXISTS ONLY public.document_access_grants DROP CONSTRAINT IF EXISTS document_access_grants_document_id_tenant_fkey;
ALTER TABLE IF EXISTS ONLY public.comments DROP CONSTRAINT IF EXISTS comments_parent_comment_id_fkey;
ALTER TABLE IF EXISTS ONLY public.bulk_items DROP CONSTRAINT IF EXISTS bulk_items_bulk_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attachments DROP CONSTRAINT IF EXISTS attachments_workflow_task_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attachments DROP CONSTRAINT IF EXISTS attachments_document_version_id_tenant_fkey;
ALTER TABLE IF EXISTS ONLY public.attachments DROP CONSTRAINT IF EXISTS attachments_comment_id_fkey;
ALTER TABLE IF EXISTS ONLY public.actor_assignments DROP CONSTRAINT IF EXISTS actor_assignments_role_id_fkey;
ALTER TABLE IF EXISTS ONLY public.actor_assignments DROP CONSTRAINT IF EXISTS actor_assignments_delegated_from_fkey;
ALTER TABLE IF EXISTS ONLY public.actor_assignments DROP CONSTRAINT IF EXISTS actor_assignments_capacity_id_fkey;
ALTER TABLE IF EXISTS ONLY public.acting_capacities DROP CONSTRAINT IF EXISTS acting_capacities_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.acting_capacities DROP CONSTRAINT IF EXISTS acting_capacities_party_id_tenant_fkey;
DROP INDEX IF EXISTS public.workflow_definitions_key_version;
DROP INDEX IF EXISTS public.whev_pending;
DROP INDEX IF EXISTS public.whep_direction;
DROP INDEX IF EXISTS public.wft_remind_after;
DROP INDEX IF EXISTS public.wft_due;
DROP INDEX IF EXISTS public.wfsa_endpoint;
DROP INDEX IF EXISTS public.wfi_resource;
DROP INDEX IF EXISTS public.wfi_open;
DROP INDEX IF EXISTS public.webhook_events_dedup;
DROP INDEX IF EXISTS public.webhook_endpoints_tenant_id_id_uidx;
DROP INDEX IF EXISTS public.uta_active;
DROP INDEX IF EXISTS public.seed_sync_runs_hash_idx;
DROP INDEX IF EXISTS public.seed_sync_runs_created_at_idx;
DROP INDEX IF EXISTS public.rule_versions_history_resolution_idx;
DROP INDEX IF EXISTS public.roles_key;
DROP INDEX IF EXISTS public.res_by_type;
DROP INDEX IF EXISTS public.rel_sub;
DROP INDEX IF EXISTS public.rel_obj;
DROP INDEX IF EXISTS public.parties_tenant_id_id_uidx;
DROP INDEX IF EXISTS public.outbox_pending;
DROP INDEX IF EXISTS public.outbox_aggregate;
DROP INDEX IF EXISTS public.organizations_tenant_id_id_uidx;
DROP INDEX IF EXISTS public.org_parent;
DROP INDEX IF EXISTS public.notification_templates_key;
DROP INDEX IF EXISTS public.notifdel_pending;
DROP INDEX IF EXISTS public.notif_recipient;
DROP INDEX IF EXISTS public.legal_holds_active;
DROP INDEX IF EXISTS public.jobs_available;
DROP INDEX IF EXISTS public.job_runs_kind;
DROP INDEX IF EXISTS public.integration_providers_key;
DROP INDEX IF EXISTS public.identity_grant_one_active_per_actor;
DROP INDEX IF EXISTS public.idempotency_keys_expiry;
DROP INDEX IF EXISTS public.dsr_requests_subject;
DROP INDEX IF EXISTS public.docver_doc;
DROP INDEX IF EXISTS public.documents_tenant_id_id_uidx;
DROP INDEX IF EXISTS public.document_versions_tenant_id_id_uidx;
DROP INDEX IF EXISTS public.document_upload_sessions_gc;
DROP INDEX IF EXISTS public.document_upload_sessions_document_id_idx;
DROP INDEX IF EXISTS public.document_upload_sessions_confirmed_version;
DROP INDEX IF EXISTS public.docgrant_doc;
DROP INDEX IF EXISTS public.doc_retention;
DROP INDEX IF EXISTS public.doc_resource;
DROP INDEX IF EXISTS public.doc_class;
DROP INDEX IF EXISTS public.contacts_party;
DROP INDEX IF EXISTS public.cmt_resource;
DROP INDEX IF EXISTS public.cap_user;
DROP INDEX IF EXISTS public.cap_active;
DROP INDEX IF EXISTS public.bulk_items_pending;
DROP INDEX IF EXISTS public.audit_logs_entity;
DROP INDEX IF EXISTS public.audit_logs_chain;
DROP INDEX IF EXISTS public.audit_logs_actor;
DROP INDEX IF EXISTS public.audit_anchors_tenant_seq;
DROP INDEX IF EXISTS public.att_resource;
DROP INDEX IF EXISTS public.asg_system;
DROP INDEX IF EXISTS public.asg_scope;
DROP INDEX IF EXISTS public.asg_actor;
DROP INDEX IF EXISTS public.artifacts_kind;
DROP INDEX IF EXISTS public.api_keys_tenant;
ALTER TABLE IF EXISTS ONLY public.workflow_tasks DROP CONSTRAINT IF EXISTS workflow_tasks_pkey;
ALTER TABLE IF EXISTS ONLY public.workflow_task_assignees DROP CONSTRAINT IF EXISTS workflow_task_assignees_pkey;
ALTER TABLE IF EXISTS ONLY public.workflow_instances DROP CONSTRAINT IF EXISTS workflow_instances_pkey;
ALTER TABLE IF EXISTS ONLY public.workflow_definitions DROP CONSTRAINT IF EXISTS workflow_definitions_pkey;
ALTER TABLE IF EXISTS ONLY public.webhook_failed_signature_audit DROP CONSTRAINT IF EXISTS webhook_failed_signature_audit_pkey;
ALTER TABLE IF EXISTS ONLY public.webhook_events DROP CONSTRAINT IF EXISTS webhook_events_pkey;
ALTER TABLE IF EXISTS ONLY public.webhook_endpoints DROP CONSTRAINT IF EXISTS webhook_endpoints_pkey;
ALTER TABLE IF EXISTS ONLY public.version_counters DROP CONSTRAINT IF EXISTS version_counters_pkey;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_idp_subject_key;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE IF EXISTS ONLY public.user_tenant_access DROP CONSTRAINT IF EXISTS user_tenant_access_pkey;
ALTER TABLE IF EXISTS ONLY public.tenants DROP CONSTRAINT IF EXISTS tenants_slug_key;
ALTER TABLE IF EXISTS ONLY public.tenants DROP CONSTRAINT IF EXISTS tenants_pkey;
ALTER TABLE IF EXISTS ONLY public.sequences DROP CONSTRAINT IF EXISTS sequences_pkey;
ALTER TABLE IF EXISTS ONLY public.sequence_allocations DROP CONSTRAINT IF EXISTS sequence_allocations_tenant_id_series_key_value_key;
ALTER TABLE IF EXISTS ONLY public.sequence_allocations DROP CONSTRAINT IF EXISTS sequence_allocations_pkey;
ALTER TABLE IF EXISTS ONLY public.seed_sync_runs DROP CONSTRAINT IF EXISTS seed_sync_runs_pkey;
ALTER TABLE IF EXISTS ONLY public.schedules DROP CONSTRAINT IF EXISTS schedules_pkey;
ALTER TABLE IF EXISTS ONLY public.rule_versions DROP CONSTRAINT IF EXISTS rule_versions_rule_key_scope_kind_coalesce_coalesce1_tstzr_excl;
ALTER TABLE IF EXISTS ONLY public.rule_versions DROP CONSTRAINT IF EXISTS rule_versions_pkey;
ALTER TABLE IF EXISTS ONLY public.rule_definitions DROP CONSTRAINT IF EXISTS rule_definitions_pkey;
ALTER TABLE IF EXISTS ONLY public.roles DROP CONSTRAINT IF EXISTS roles_pkey;
ALTER TABLE IF EXISTS ONLY public.role_permissions DROP CONSTRAINT IF EXISTS role_permissions_pkey;
ALTER TABLE IF EXISTS ONLY public.resources DROP CONSTRAINT IF EXISTS resources_pkey;
ALTER TABLE IF EXISTS ONLY public.resource_types DROP CONSTRAINT IF EXISTS resource_types_pkey;
ALTER TABLE IF EXISTS ONLY public.relationships DROP CONSTRAINT IF EXISTS relationships_pkey;
ALTER TABLE IF EXISTS ONLY public.relationship_types DROP CONSTRAINT IF EXISTS relationship_types_pkey;
ALTER TABLE IF EXISTS ONLY public.processed_events DROP CONSTRAINT IF EXISTS processed_events_pkey;
ALTER TABLE IF EXISTS ONLY public.policy_conditions DROP CONSTRAINT IF EXISTS policy_conditions_pkey;
ALTER TABLE IF EXISTS ONLY public.policies DROP CONSTRAINT IF EXISTS policies_pkey;
ALTER TABLE IF EXISTS ONLY public.persons DROP CONSTRAINT IF EXISTS persons_pkey;
ALTER TABLE IF EXISTS ONLY public.permissions DROP CONSTRAINT IF EXISTS permissions_pkey;
ALTER TABLE IF EXISTS ONLY public.party_contacts DROP CONSTRAINT IF EXISTS party_contacts_tenant_id_party_id_kind_value_key;
ALTER TABLE IF EXISTS ONLY public.party_contacts DROP CONSTRAINT IF EXISTS party_contacts_pkey;
ALTER TABLE IF EXISTS ONLY public.parties DROP CONSTRAINT IF EXISTS parties_pkey;
ALTER TABLE IF EXISTS ONLY public.organizations DROP CONSTRAINT IF EXISTS organizations_tenant_id_parent_org_id_name_key;
ALTER TABLE IF EXISTS ONLY public.organizations DROP CONSTRAINT IF EXISTS organizations_pkey;
ALTER TABLE IF EXISTS ONLY public.notifications DROP CONSTRAINT IF EXISTS notifications_pkey;
ALTER TABLE IF EXISTS ONLY public.notification_templates DROP CONSTRAINT IF EXISTS notification_templates_pkey;
ALTER TABLE IF EXISTS ONLY public.notification_deliveries DROP CONSTRAINT IF EXISTS notification_deliveries_pkey;
ALTER TABLE IF EXISTS ONLY public.notification_channel_prefs DROP CONSTRAINT IF EXISTS notification_channel_prefs_pkey;
ALTER TABLE IF EXISTS ONLY public.legal_holds DROP CONSTRAINT IF EXISTS legal_holds_pkey;
ALTER TABLE IF EXISTS ONLY public.legal_entities DROP CONSTRAINT IF EXISTS legal_entities_pkey;
ALTER TABLE IF EXISTS ONLY public.jobs_queue DROP CONSTRAINT IF EXISTS jobs_queue_pkey;
ALTER TABLE IF EXISTS ONLY public.job_runs DROP CONSTRAINT IF EXISTS job_runs_pkey;
ALTER TABLE IF EXISTS ONLY public.integration_providers DROP CONSTRAINT IF EXISTS integration_providers_pkey;
ALTER TABLE IF EXISTS ONLY public.identity_grant DROP CONSTRAINT IF EXISTS identity_grant_pkey;
ALTER TABLE IF EXISTS ONLY public.idempotency_keys DROP CONSTRAINT IF EXISTS idempotency_keys_pkey;
ALTER TABLE IF EXISTS ONLY public.events_outbox DROP CONSTRAINT IF EXISTS events_outbox_pkey;
ALTER TABLE IF EXISTS ONLY public.dsr_requests DROP CONSTRAINT IF EXISTS dsr_requests_pkey;
ALTER TABLE IF EXISTS ONLY public.documents DROP CONSTRAINT IF EXISTS documents_pkey;
ALTER TABLE IF EXISTS ONLY public.document_versions DROP CONSTRAINT IF EXISTS document_versions_pkey;
ALTER TABLE IF EXISTS ONLY public.document_versions DROP CONSTRAINT IF EXISTS document_versions_document_id_version_no_key;
ALTER TABLE IF EXISTS ONLY public.document_upload_sessions DROP CONSTRAINT IF EXISTS document_upload_sessions_pkey;
ALTER TABLE IF EXISTS ONLY public.document_access_grants DROP CONSTRAINT IF EXISTS document_access_grants_pkey;
ALTER TABLE IF EXISTS ONLY public.comments DROP CONSTRAINT IF EXISTS comments_pkey;
ALTER TABLE IF EXISTS ONLY public.bulk_operations DROP CONSTRAINT IF EXISTS bulk_operations_pkey;
ALTER TABLE IF EXISTS ONLY public.bulk_items DROP CONSTRAINT IF EXISTS bulk_items_pkey;
ALTER TABLE IF EXISTS ONLY public.bulk_items DROP CONSTRAINT IF EXISTS bulk_items_bulk_id_seq_key;
ALTER TABLE IF EXISTS ONLY public.authz_epoch DROP CONSTRAINT IF EXISTS authz_epoch_pkey;
ALTER TABLE IF EXISTS ONLY public.audit_logs DROP CONSTRAINT IF EXISTS audit_logs_pkey;
ALTER TABLE IF EXISTS ONLY public.audit_chain DROP CONSTRAINT IF EXISTS audit_chain_pkey;
ALTER TABLE IF EXISTS ONLY public.audit_anchors DROP CONSTRAINT IF EXISTS audit_anchors_pkey;
ALTER TABLE IF EXISTS ONLY public.attachments DROP CONSTRAINT IF EXISTS attachments_pkey;
ALTER TABLE IF EXISTS ONLY public.artifacts DROP CONSTRAINT IF EXISTS artifacts_tenant_id_kind_version_key;
ALTER TABLE IF EXISTS ONLY public.artifacts DROP CONSTRAINT IF EXISTS artifacts_pkey;
ALTER TABLE IF EXISTS ONLY public.api_keys DROP CONSTRAINT IF EXISTS api_keys_pkey;
ALTER TABLE IF EXISTS ONLY public.api_keys DROP CONSTRAINT IF EXISTS api_keys_key_prefix_key;
ALTER TABLE IF EXISTS ONLY public.actor_assignments DROP CONSTRAINT IF EXISTS actor_assignments_pkey;
ALTER TABLE IF EXISTS ONLY public.acting_capacities DROP CONSTRAINT IF EXISTS acting_capacities_pkey;
ALTER TABLE IF EXISTS ONLY migration.backfill_checkpoint DROP CONSTRAINT IF EXISTS backfill_checkpoint_pkey;
DROP TABLE IF EXISTS public.workflow_tasks;
DROP TABLE IF EXISTS public.workflow_task_assignees;
DROP TABLE IF EXISTS public.workflow_instances;
DROP TABLE IF EXISTS public.workflow_definitions;
DROP TABLE IF EXISTS public.webhook_failed_signature_audit;
DROP TABLE IF EXISTS public.webhook_events;
DROP TABLE IF EXISTS public.webhook_endpoints;
DROP TABLE IF EXISTS public.version_counters;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.user_tenant_access;
DROP TABLE IF EXISTS public.tenants;
DROP TABLE IF EXISTS public.sequences;
DROP TABLE IF EXISTS public.sequence_allocations;
DROP TABLE IF EXISTS public.seed_sync_runs;
DROP TABLE IF EXISTS public.schedules;
DROP TABLE IF EXISTS public.rule_versions;
DROP TABLE IF EXISTS public.rule_definitions;
DROP TABLE IF EXISTS public.roles;
DROP TABLE IF EXISTS public.role_permissions;
DROP TABLE IF EXISTS public.resources;
DROP TABLE IF EXISTS public.resource_types;
DROP TABLE IF EXISTS public.relationships;
DROP TABLE IF EXISTS public.relationship_types;
DROP TABLE IF EXISTS public.processed_events;
DROP TABLE IF EXISTS public.policy_conditions;
DROP TABLE IF EXISTS public.policies;
DROP TABLE IF EXISTS public.persons;
DROP TABLE IF EXISTS public.permissions;
DROP TABLE IF EXISTS public.party_contacts;
DROP TABLE IF EXISTS public.parties;
DROP TABLE IF EXISTS public.organizations;
DROP TABLE IF EXISTS public.notifications;
DROP TABLE IF EXISTS public.notification_templates;
DROP TABLE IF EXISTS public.notification_deliveries;
DROP TABLE IF EXISTS public.notification_channel_prefs;
DROP TABLE IF EXISTS public.legal_holds;
DROP TABLE IF EXISTS public.legal_entities;
DROP TABLE IF EXISTS public.jobs_queue;
DROP TABLE IF EXISTS public.job_runs;
DROP TABLE IF EXISTS public.integration_providers;
DROP TABLE IF EXISTS public.identity_grant;
DROP TABLE IF EXISTS public.idempotency_keys;
DROP TABLE IF EXISTS public.events_outbox;
DROP TABLE IF EXISTS public.dsr_requests;
DROP TABLE IF EXISTS public.documents;
DROP TABLE IF EXISTS public.document_versions;
DROP TABLE IF EXISTS public.document_upload_sessions;
DROP TABLE IF EXISTS public.document_access_grants;
DROP TABLE IF EXISTS public.comments;
DROP TABLE IF EXISTS public.bulk_operations;
DROP TABLE IF EXISTS public.bulk_items;
DROP TABLE IF EXISTS public.authz_epoch;
DROP TABLE IF EXISTS public.audit_logs;
DROP TABLE IF EXISTS public.audit_chain;
DROP TABLE IF EXISTS public.audit_anchors;
DROP TABLE IF EXISTS public.attachments;
DROP TABLE IF EXISTS public.artifacts;
DROP TABLE IF EXISTS public.api_keys;
DROP TABLE IF EXISTS public.actor_assignments;
DROP TABLE IF EXISTS public.acting_capacities;
DROP TABLE IF EXISTS migration.backfill_checkpoint;
DROP FUNCTION IF EXISTS public.app_tenant_id_or_null();
DROP FUNCTION IF EXISTS public.app_tenant_id();
DROP FUNCTION IF EXISTS public.app_actor_id();
DROP SCHEMA IF EXISTS migration;


-- Cluster-global roles intentionally survive database reset. Extensions are
-- database-scoped but are also retained: another schema or product module in
-- this database may depend on them, so extension removal is operator-owned.
