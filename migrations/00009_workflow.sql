-- Blueprint 03 §5 "006" (on-disk 00009): the workflow engine tables.
-- workflow_definitions is a platform+tenant hybrid (module templates + tenant
-- overrides); instances/tasks/assignees are tenant-scoped with RLS.

-- +goose Up

CREATE TABLE workflow_definitions (
    id          uuid PRIMARY KEY,
    key         text NOT NULL,                  -- 'requests.approval'
    version     int  NOT NULL,
    tenant_id   uuid,                            -- NULL = module template
    applies_to  text NOT NULL,                  -- resource_type key
    definition  jsonb NOT NULL,                 -- validated graph
    status      text NOT NULL DEFAULT 'active',
    created_at  timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL
);
-- One definition per (key, version, scope); COALESCE folds the platform NULL
-- tenant into a sentinel (an expression index, not an inline UNIQUE constraint).
CREATE UNIQUE INDEX workflow_definitions_key
    ON workflow_definitions (key, version, COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid));

CREATE TABLE workflow_instances (
    id            uuid PRIMARY KEY,
    tenant_id     uuid NOT NULL,
    definition_id uuid NOT NULL REFERENCES workflow_definitions(id),
    resource_type text NOT NULL,
    resource_id   uuid NOT NULL,
    current_step  text NOT NULL,
    status        text NOT NULL DEFAULT 'running'
                      CHECK (status IN ('running','completed','rejected','cancelled','overridden')),
    context       jsonb NOT NULL DEFAULT '{}',
    started_by    uuid NOT NULL, ended_at timestamptz,
    version       int NOT NULL DEFAULT 1,
    created_at    timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at    timestamptz, updated_by uuid
);
CREATE INDEX wfi_resource ON workflow_instances (tenant_id, resource_type, resource_id);
CREATE INDEX wfi_open ON workflow_instances (tenant_id, status) WHERE status = 'running';

CREATE TABLE workflow_tasks (
    id            uuid PRIMARY KEY,
    tenant_id     uuid NOT NULL,
    instance_id   uuid NOT NULL REFERENCES workflow_instances(id),
    step_key      text NOT NULL,
    task_type     text NOT NULL,
    status        text NOT NULL DEFAULT 'open'
                      CHECK (status IN ('open','done','approved','rejected','skipped','expired','delegated')),
    due_at        timestamptz, remind_after timestamptz, last_reminded_at timestamptz,
    decided_by    uuid, decided_at timestamptz, decision_comment text,
    delegated_to  uuid, output jsonb,
    version       int NOT NULL DEFAULT 1,
    created_at    timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
    updated_at    timestamptz, updated_by uuid
);
CREATE INDEX wft_due ON workflow_tasks (tenant_id, status, due_at) WHERE status = 'open';

CREATE TABLE workflow_task_assignees (
    task_id       uuid NOT NULL REFERENCES workflow_tasks(id),
    tenant_id     uuid NOT NULL,
    assignee_kind text NOT NULL CHECK (assignee_kind IN ('capacity','role','relationship','system')),
    assignee_ref  text NOT NULL,
    PRIMARY KEY (task_id, assignee_kind, assignee_ref)
);

-- workflow_definitions: platform+tenant hybrid RLS (templates + tenant overrides).
ALTER TABLE workflow_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_definitions FORCE ROW LEVEL SECURITY;
CREATE POLICY workflow_definitions_tenant ON workflow_definitions
    USING (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null())
    WITH CHECK (tenant_id IS NULL OR tenant_id = app_tenant_id_or_null());
GRANT SELECT ON workflow_definitions TO app_rt;
GRANT SELECT, INSERT, UPDATE ON workflow_definitions TO app_platform;

-- instances/tasks/assignees: strict tenant RLS.
-- +goose StatementBegin
DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['workflow_instances','workflow_tasks','workflow_task_assignees']
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('CREATE POLICY %I ON %I USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id())', t||'_tenant_isolation', t);
        EXECUTE format('GRANT SELECT, INSERT, UPDATE ON %I TO app_rt', t);
    END LOOP;
END
$$;
-- +goose StatementEnd
GRANT DELETE ON workflow_task_assignees TO app_rt;  -- reassignment replaces the assignee set

-- +goose Down

DROP TABLE IF EXISTS workflow_task_assignees;
DROP TABLE IF EXISTS workflow_tasks;
DROP TABLE IF EXISTS workflow_instances;
DROP TABLE IF EXISTS workflow_definitions;
