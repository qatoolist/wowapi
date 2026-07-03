# 02 — Workflow Engine & Rule/Configuration Engine

## 1. Workflow engine

### 1.1 Engine choice: small custom Postgres-backed engine — **recommended**

| Option | Verdict | Why |
|---|---|---|
| **Custom Postgres engine (kernel)** | ✅ v1 | Our workflows are approval/state-machine shaped (steps, assignees, decisions, SLAs), tenant-configurable from seed/JSON, and must share the business transaction + RLS + audit + outbox. A few thousand lines, fully testable. |
| Temporal | later, selectively | Superb for *durable code orchestration* (long sagas, external retries). Wrong shape for declarative, tenant-editable approval graphs; adds a cluster + separate persistence outside RLS. Adopt later behind a `workflow.Engine` port if a use case demands it. |
| Camunda / Zeebe | ❌ | BPMN authoring + JVM/cluster ops for a Go monolith; per-tenant overrides awkward. |
| Conductor | ❌ | Microservice-orchestration shaped; infra overhead. |

The kernel owns the runtime; modules own the *definitions* (seeded JSON), tenants own *overrides*.

### 1.2 Definition model (JSON, versioned, seedable)

```yaml
# seeds/workflows.yaml (module-owned; generic example)
key: requests.approval            # module.name
version: 3                        # immutable once instances exist
applies_to: requests.request      # resource_type
initial_step: manager_review
steps:
  manager_review:
    type: approval                # approval | task | auto | gateway | vote
    assignees:
      - kind: role                # role | actor | relationship | resource_owner | expr
        role: requests.org.approver
        scope: org_of_resource
    policy: { min_approvals: 1, self_approval: false }
    sla: { due: "P2D", remind_after: "P1D", escalate_to: step:escalation }
    on_approve: { next: auto_provision }
    on_reject:  { next: end_rejected, require_comment: true }
  auto_provision:
    type: auto
    action: requests.provision    # registered auto-action (module Go code)
    on_error: { retry: default, then: manual_fixup }
  vote_example:
    type: vote                    # generic quorum capability (no AGM semantics in core)
    electorate: { kind: relationship, rel: core.member_of, of: org_of_resource }
    quorum: { kind: fraction, value: "1/2" }
    pass:   { kind: fraction, value: "2/3" }
    window: "P14D"
  end_rejected: { type: terminal, outcome: rejected }
```

- **Step types (closed set in kernel):** `approval`, `task` (do-something + mark done), `auto`
  (invoke registered module action), `gateway` (branch on rule/expression over instance context),
  `vote` (quorum/threshold), `terminal`.
- **Assignee resolution kinds:** explicit actor, role-at-scope, relationship-holder
  (e.g. `core.owner_of` the resource), resource owner, or module-registered resolver func.
- **Versioning:** definitions are immutable per version; running instances pin their version;
  tenants override by cloning a template version (`tenant_id` set) — resolution: tenant override → module template.

### 1.3 Runtime behavior (normative)

- `Start(ctx, defKey, resourceRef, input)` → creates `workflow_instances` row + first tasks, **inside the caller's tenant tx** — a request and its workflow start commit or roll back together.
- Transitions are the only way state changes: `Decide(taskID, actor, decision, comment?)`,
  `CompleteTask`, `RunAuto`, `Escalate`, `Override`. Each transition:
  1. re-validates the actor may act (assignee check + `workflow.task.decide` permission),
  2. mutates instance/task rows with optimistic locking (`version` column),
  3. writes audit row + outbox event `workflow.<def>.{task_created|approved|rejected|escalated|completed|overridden}` — same tx.
- **Delegation:** task-level (`Delegate(taskID, toCapacity, until)`) recorded on the task; original assignee retains visibility.
- **Escalation/SLA:** a scheduled sweeper job (per tenant) finds tasks past `remind_after`/`due`; reminders → notification framework; breaches → escalation step/assignee. Sweep is idempotent (`last_reminded_at` guard).
- **Emergency override:** privileged transition requiring `workflow.instance.override` + reason; jumps to a step or terminal outcome AND auto-creates a `ratification` task assigned per definition (`ratify_by` role). Ratification refusal fires `workflow.<def>.override_ratification_rejected` — the *domain module* decides consequences (kernel stays neutral).
- **Comments/attachments:** reuse kernel comment/attachment services against the task's `ResourceRef`.
- **State validation:** on definition load — graph connectivity, no orphan steps, terminals reachable, unknown auto-actions fail boot.
- **Test runner:** testkit `WorkflowSim` drives a definition in-memory over a real test DB: `sim.Start(...).Approve("manager_review", asActor).ExpectStep("auto_provision")`.

### 1.4 Kernel interfaces

```go
type Runtime interface {
    Start(ctx context.Context, defKey string, res resource.Ref, input map[string]any) (InstanceID, error)
    Decide(ctx context.Context, taskID uuid.UUID, d Decision) error // approve/reject/abstain + comment
    CompleteTask(ctx context.Context, taskID uuid.UUID, output map[string]any) error
    Delegate(ctx context.Context, taskID uuid.UUID, to uuid.UUID, until time.Time) error
    Override(ctx context.Context, instanceID uuid.UUID, to StepKey, reason string) error
    Instance(ctx context.Context, id uuid.UUID) (Instance, error)
    OpenTasksFor(ctx context.Context, a authz.Actor, p page.Cursor) (page.Page[Task], error)
}

type Registry interface {           // module-facing (via ModuleContext)
    RegisterDefinition(def Definition) error       // from seeds
    RegisterAutoAction(key string, fn AutoAction) error
    RegisterAssigneeResolver(key string, fn AssigneeResolver) error
}
type AutoAction func(ctx context.Context, in AutoInput) (map[string]any, error)
```

Generic examples the kernel ships tests for: access-request approval, document approval,
payment approval (as a neutral "high-value action" flow), vendor onboarding, policy-change approval,
data-correction approval. **Society flows (membership approval, notice approval, bill approval,
complaint escalation) are just YAML in the future society module — no engine changes.**

## 2. Rule / configuration engine

### 2.1 Model

Three pieces, all core:

1. **Rule point** (code-registered, from modules): `key`, value **JSON Schema**, default value, scope
   levels allowed, `requires_approval bool`, description. Naming `module.area.name`:
   `core.retention.audit_days`, `core.upload.max_size_mb`, `notify.digest.hour`,
   `workflow.escalation.default_sla`, `billing.late_fee.percent` (generic), `feature.requests.bulk_ops`.
2. **Rule definition row** (`rule_definitions`): persisted mirror of the registered point (schema, default, module) — makes points introspectable/auditable in the DB.
3. **Rule version** (`rule_versions`): a value at a scope with temporal validity:
   `(rule_key, scope_kind platform|tenant|org, scope_id?, value jsonb, effective_from, effective_to, status draft|pending_approval|active|superseded|rejected, approved_by?, created_by)`.

### 2.2 Resolution algorithm (normative)

```
Resolve(key, tenant, org?, at time.Time):
  candidates = active versions of key where effective_from <= at < effective_to
  pick first of: org-scope (walking org ancestry upward) → tenant-scope → platform-scope → code default
  validate value against schema (defense in depth; also validated at write time)
  return TypedValue + provenance{version_id, scope}
```
- **Historical evaluation:** pass any `at` — versions are never mutated, only superseded
  (`effective_to` set + new row). "What was the late fee on 2025-03-31?" is a lookup, not archaeology.
- **Fallback is total:** a registered point always resolves (code default is the floor). Missing
  registration = boot failure, not runtime nil.
- **Caching:** in-memory per-process cache keyed `(tenant, key)`, invalidated by
  `core.rule.version_activated` events + 60s TTL backstop. Rules are read-heavy, hot-path safe.

### 2.3 Change control

- Writes create `draft` versions. Points with `requires_approval` route activation through the
  workflow engine (definition `core.rule_change` — the rule engine and workflow engine compose;
  neither imports the other's internals, coupling is via events + a kernel service call).
- Activation checks overlap: no two active versions of one key may overlap at the same scope
  (DB exclusion constraint on `(rule_key, scope_kind, scope_id, tstzrange(effective_from, effective_to))`).
- Every activation/deactivation/supersession → audit row + outbox event.
- **Template bundles:** modules ship `rules.yaml` seed bundles (platform-scope defaults); tenant
  onboarding can apply a named bundle ("standard", "strict") = batch of tenant-scope versions.
- **Feature flags:** rule points of schema `{"type":"boolean"}` (or enum variants) under `feature.*`,
  resolved by the same path, with a sugar API `flags.Enabled(ctx, "feature.x")`.
- **Safe rollout:** flags support `{"rollout_pct": n}` variant; deterministic hash of tenant/actor id.

### 2.4 Interfaces

```go
type Registry interface {                        // module-facing
    RegisterPoint(p Point) error                 // key, schema, default, scopes, requiresApproval
}
type Resolver interface {
    Resolve(ctx context.Context, key string, at time.Time) (Value, error) // tenant/org from ctx
    ResolveAs(ctx context.Context, key string, at time.Time, out any) error // json.Unmarshal into typed struct
}
type Value struct {
    Raw        json.RawMessage
    VersionID  uuid.UUID   // provenance for audit ("decision used rule version X")
    Scope      ScopeKind
}
```

Anywhere a service makes a rule-driven decision it records `VersionID` in the audit payload —
compliance answers "which config produced this outcome" for free.

### 2.5 Example rows

```jsonc
// platform default
{ "rule_key":"core.upload.max_size_mb", "scope_kind":"platform", "value":25,
  "effective_from":"2026-01-01T00:00:00Z", "status":"active" }
// tenant override, approved, future-dated
{ "rule_key":"core.upload.max_size_mb", "scope_kind":"tenant", "scope_id":"<tenant>",
  "value":100, "effective_from":"2026-08-01T00:00:00Z", "status":"active",
  "approved_by":"<actor>", "workflow_instance_id":"<wf>" }
```

**Society mapping (illustration):** the society module registers points like
`society.agm.notice_period_days`, `society.billing.frequency`, `society.defaulter.threshold_amount`,
`society.parking.eligibility` — schema-validated, tenant-overridable, workflow-approved, historically
resolvable. Kernel unchanged.
