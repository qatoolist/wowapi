package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/filtering"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/pagination"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// decidePermission is the permission a transition re-checks when an authz
// Evaluator is wired and the task carries role-based assignees. It is a
// best-effort secondary gate; the primary gate is the assignee check.
const decidePermission = "workflow.task.decide"

// DecisionType is the outcome an actor records on an approval/vote task.
type DecisionType string

const (
	// DecisionApprove advances via the step's on_approve transition.
	DecisionApprove DecisionType = "approve"
	// DecisionReject advances via the step's on_reject transition.
	DecisionReject DecisionType = "reject"
	// DecisionAbstain records a non-committal vote (vote steps).
	DecisionAbstain DecisionType = "abstain"
)

// Decision is the input to Decide: who acted, the outcome, and an optional
// comment (required when the transition sets require_comment).
type Decision struct {
	Actor   authz.Actor
	Type    DecisionType
	Comment string
}

// Instance is a running (or ended) workflow instance.
type Instance struct {
	ID           uuid.UUID
	DefinitionID uuid.UUID
	Resource     resource.Ref
	CurrentStep  string
	Status       string
	Context      map[string]any
	Version      int
}

// Task is a unit of work in an instance (an approval, a to-do, etc.).
type Task struct {
	ID          uuid.UUID
	InstanceID  uuid.UUID
	StepKey     string
	TaskType    string
	Status      string
	DueAt       *time.Time
	RemindAfter *time.Time
	DecidedBy   *uuid.UUID
	DelegatedTo *uuid.UUID
	Output      map[string]any
	Version     int
}

// Runtime is the workflow engine. Every method mutates state inside a tenant
// transaction (RLS-scoped), writing the matching outbox event in the same tx.
// StartIn joins the caller's transaction; the other mutators open their own.
type Runtime struct {
	txm      database.TxManager
	registry *Registry
	authz    authz.Evaluator // optional secondary gate
	outbox   outbox.Writer
	audit    *audit.Writer
	idgen    model.IDGen
	now      func() time.Time
	metrics  observability.Metrics
}

// RuntimeOption customizes the workflow runtime.
type RuntimeOption func(*Runtime)

// WithRuntimeMetrics wires the bounded-cardinality worker timing metrics used by
// SweepSLA. A nil sink leaves the safe NoOp default in place.
func WithRuntimeMetrics(metrics observability.Metrics) RuntimeOption {
	return func(rt *Runtime) {
		if metrics != nil {
			rt.metrics = metrics
		}
	}
}

// NewRuntime wires the runtime. All dependencies, including the authz
// Evaluator and audit Writer, are required: Override's privileged permission
// check is unconditional, and every override must be durable-audited in the
// same transaction (blueprint §1.3, review finding SEC-02).
func NewRuntime(txm database.TxManager, reg *Registry, ev authz.Evaluator, ob outbox.Writer, idgen model.IDGen, aud *audit.Writer, opts ...RuntimeOption) *Runtime {
	if txm == nil || reg == nil || ev == nil || ob == nil || idgen == nil || aud == nil {
		panic("workflow.NewRuntime: txm, registry, authz evaluator, outbox, idgen, and audit writer are required")
	}
	rt := &Runtime{
		txm: txm, registry: reg, authz: ev, outbox: ob, audit: aud, idgen: idgen,
		now: time.Now, metrics: observability.NoOp,
	}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

// StartIn creates an instance and enters its initial step INSIDE the caller's
// tenant transaction, so a business write and its workflow start commit or roll
// back together (blueprint §1.3).
func (rt *Runtime) StartIn(ctx context.Context, db database.TenantDB, defKey string, res resource.Ref, input map[string]any) (uuid.UUID, error) {
	if res.IsZero() {
		return uuid.Nil, kerr.E(kerr.KindValidation, "workflow_start_invalid", "workflow start requires a resource ref")
	}
	// Resolve the definition row (for the definition_id FK) + the registered
	// graph. The version is pinned to the DB row's version.
	defID, version, err := rt.definitionRow(ctx, db, defKey)
	if err != nil {
		return uuid.Nil, err
	}
	def, ok := rt.registry.definition(defKey, version)
	if !ok {
		return uuid.Nil, kerr.E(kerr.KindInternal, "workflow_definition_unregistered",
			fmt.Sprintf("workflow definition %s v%d is not registered", defKey, version))
	}
	if def.AppliesTo != "" && def.AppliesTo != res.Type {
		return uuid.Nil, kerr.E(kerr.KindValidation, "workflow_applies_to_mismatch",
			fmt.Sprintf("definition %s applies to %q, not %q", defKey, def.AppliesTo, res.Type))
	}

	actor := actorFromCtx(ctx)
	instanceID := rt.idgen.New()
	if input == nil {
		input = map[string]any{}
	}
	ctxJSON, err := json.Marshal(input)
	if err != nil {
		return uuid.Nil, kerr.E(kerr.KindValidation, "workflow_context_invalid", "instance context not JSON-encodable")
	}
	_, err = db.Exec(ctx,
		`INSERT INTO workflow_instances
		    (id, tenant_id, definition_id, resource_type, resource_id, current_step, status, context, started_by, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, 'running', $6, $7, $7)`,
		instanceID, defID, res.Type, res.ID, def.InitialStep, ctxJSON, actor)
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "workflow.StartIn", "insert instance")
	}

	inst := Instance{ID: instanceID, DefinitionID: defID, Resource: res, CurrentStep: def.InitialStep, Status: "running", Context: input, Version: 1}
	if err := rt.enterStep(ctx, db, &inst, def, def.InitialStep, actor); err != nil {
		return uuid.Nil, err
	}
	return instanceID, nil
}

// Decide records an approve/reject on a task and drives the resulting
// transition, in its own tenant transaction.
func (rt *Runtime) Decide(ctx context.Context, taskID uuid.UUID, d Decision) error {
	return rt.txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		task, err := rt.loadTask(ctx, db, taskID)
		if err != nil {
			return err
		}
		if task.Status != "open" {
			return kerr.E(kerr.KindWorkflowState, "task_not_open",
				fmt.Sprintf("task %s is %s, not open", taskID, task.Status))
		}
		inst, def, err := rt.loadInstanceAndDef(ctx, db, task.InstanceID)
		if err != nil {
			return err
		}
		step := def.Steps[task.StepKey]
		if step.Type != StepApproval && step.Type != StepVote {
			return kerr.E(kerr.KindWorkflowState, "not_a_decision_step",
				fmt.Sprintf("step %q is a %s, not a decision step", task.StepKey, step.Type))
		}
		if err := rt.authorize(ctx, db, task, inst, d.Actor); err != nil {
			return err
		}

		var transition *Transition
		var newStatus, verb string
		// Fail-closed by design (adjudicated, MATRIX CS-23): the default: arm
		// rejects DecisionAbstain — and any future unknown decision type — with
		// an invalid_decision error. The deny-by-default arm IS the safety
		// property; do not convert to an exhaustive enumeration.
		//exhaustive:ignore
		switch d.Type {
		case DecisionApprove:
			transition, newStatus, verb = step.OnApprove, "approved", "approved"
		case DecisionReject:
			transition, newStatus, verb = step.OnReject, "rejected", "rejected"
		default:
			return kerr.E(kerr.KindValidation, "invalid_decision",
				fmt.Sprintf("unsupported decision type %q", d.Type))
		}
		if transition != nil && transition.RequireComment && d.Comment == "" {
			return kerr.E(kerr.KindValidation, "comment_required",
				"this transition requires a decision comment")
		}

		decidedBy := d.Actor.CapacityID
		if err := rt.closeTask(ctx, db, task, newStatus, &decidedBy, d.Comment, nil); err != nil {
			return err
		}
		if err := rt.emit(ctx, db, inst, def, verb, map[string]any{
			"instance_id": inst.ID.String(), "task_id": task.ID.String(), "step": task.StepKey,
		}); err != nil {
			return err
		}
		return rt.advance(ctx, db, &inst, def, transition, d.Actor.CapacityID)
	})
}

// CompleteTask marks a `task`-type task done (with optional output) and advances
// via the step's next transition, in its own tenant transaction.
func (rt *Runtime) CompleteTask(ctx context.Context, taskID uuid.UUID, output map[string]any) error {
	return rt.txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		task, err := rt.loadTask(ctx, db, taskID)
		if err != nil {
			return err
		}
		if task.Status != "open" {
			return kerr.E(kerr.KindWorkflowState, "task_not_open",
				fmt.Sprintf("task %s is %s, not open", taskID, task.Status))
		}
		inst, def, err := rt.loadInstanceAndDef(ctx, db, task.InstanceID)
		if err != nil {
			return err
		}
		step := def.Steps[task.StepKey]
		if step.Type != StepTask {
			return kerr.E(kerr.KindWorkflowState, "not_a_task_step",
				fmt.Sprintf("step %q is a %s, not a task step", task.StepKey, step.Type))
		}
		if err := rt.closeTask(ctx, db, task, "done", nil, "", output); err != nil {
			return err
		}
		if err := rt.emit(ctx, db, inst, def, "completed", map[string]any{
			"instance_id": inst.ID.String(), "task_id": task.ID.String(), "step": task.StepKey,
		}); err != nil {
			return err
		}
		// Merge output into instance context so downstream gateways can branch.
		if len(output) > 0 {
			for k, v := range output {
				inst.Context[k] = v
			}
			if err := rt.saveContext(ctx, db, &inst); err != nil {
				return err
			}
		}
		return rt.advance(ctx, db, &inst, def, step.Next, uuid.Nil)
	})
}

// Delegate records a delegate on an OPEN task: delegated_to is set and the
// delegate is ADDED as an assignee so the original assignee retains visibility.
// The task stays open (blueprint §1.3).
func (rt *Runtime) Delegate(ctx context.Context, taskID, to uuid.UUID, until time.Time) error {
	return rt.txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		task, err := rt.loadTask(ctx, db, taskID)
		if err != nil {
			return err
		}
		if task.Status != "open" {
			return kerr.E(kerr.KindWorkflowState, "task_not_open",
				fmt.Sprintf("task %s is %s, not open", taskID, task.Status))
		}
		tag, err := db.Exec(ctx,
			`UPDATE workflow_tasks SET delegated_to = $2, version = version + 1, updated_at = now()
			 WHERE id = $1 AND version = $3`, taskID, to, task.Version)
		if err != nil {
			return kerr.Wrapf(err, "workflow.Delegate", "update task")
		}
		if tag.RowsAffected() == 0 {
			return versionConflict(taskID)
		}
		if _, err := db.Exec(ctx,
			`INSERT INTO workflow_task_assignees (task_id, tenant_id, assignee_kind, assignee_ref)
			 VALUES ($1, app_tenant_id(), 'capacity', $2)
			 ON CONFLICT DO NOTHING`, taskID, to.String()); err != nil {
			return kerr.Wrapf(err, "workflow.Delegate", "add delegate assignee")
		}
		inst, def, err := rt.loadInstanceAndDef(ctx, db, task.InstanceID)
		if err != nil {
			return err
		}
		return rt.emit(ctx, db, inst, def, "delegated", map[string]any{
			"instance_id": inst.ID.String(), "task_id": taskID.String(),
			"delegated_to": to.String(), "until": until.UTC().Format(time.RFC3339),
		})
	})
}

// Override is a privileged transition: it requires a reason and jumps the
// instance to a step or terminal, emitting workflow.<def>.overridden. Any open
// tasks on the current step are marked skipped.
//
// Ratification is explicitly rejected as an interim, Wave-0-compatible posture
// (W03-E05-S001): definitions declaring ratify_by are rejected at validation
// time, so every override records ratification_outcome="rejected_interim".
func (rt *Runtime) Override(ctx context.Context, actor authz.Actor, instanceID uuid.UUID, to string, reason string) error {
	if reason == "" {
		return kerr.E(kerr.KindValidation, "override_reason_required", "override requires a reason")
	}
	return rt.txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		inst, def, err := rt.loadInstanceAndDef(ctx, db, instanceID)
		if err != nil {
			return err
		}
		// Override is a privileged state jump — it MUST be gated by the
		// workflow.instance.override permission (blueprint §1.3, review finding
		// SEC-02/SEC-39), unlike a normal Decide which is gated by task
		// assignment. NewRuntime requires a non-nil evaluator, so this check is
		// unconditional: there is no construction path that can bypass it.
		d, aerr := rt.authz.Evaluate(ctx, db, actor, "workflow.instance.override",
			authz.Target{Scope: authz.ScopeResource, Resource: inst.Resource})
		if aerr != nil {
			return aerr
		}
		if !d.Allowed {
			return kerr.E(kerr.KindForbidden, "permission_denied", "not permitted to override this workflow")
		}
		if inst.Status != "running" {
			return kerr.E(kerr.KindWorkflowState, "instance_not_running",
				fmt.Sprintf("instance %s is %s", instanceID, inst.Status))
		}
		target, ok := def.Steps[to]
		if !ok {
			return kerr.E(kerr.KindValidation, "override_target_unknown",
				fmt.Sprintf("override target step %q does not exist", to))
		}
		// Durable audit, written inside the same transaction as the state jump.
		// If this write fails, the entire override (including any state mutation
		// that follows) rolls back.
		ctx = withAuditActor(ctx, actor)
		if err := rt.audit.Record(ctx, db, audit.Entry{
			Action:         "workflow.instance.override",
			EntityType:     "workflow_instance",
			EntityID:       inst.ID,
			OldValue:       inst.CurrentStep,
			NewValue:       to,
			Reason:         reason,
			ActorKind:      string(actor.Kind),
			ImpersonatorID: actor.ImpersonatorUserID,
			Metadata: map[string]any{
				"source_state":         inst.CurrentStep,
				"target_state":         to,
				"grant_id":             grantIDStr(actor.GrantID),
				"ratification_outcome": "rejected_interim",
			},
		}); err != nil {
			return kerr.Wrapf(err, "workflow.Override", "audit override")
		}
		// Skip any open tasks on the current step — the override supersedes them.
		if _, err := db.Exec(ctx,
			`UPDATE workflow_tasks SET status = 'skipped', version = version + 1, updated_at = now()
			 WHERE instance_id = $1 AND status = 'open'`, instanceID); err != nil {
			return kerr.Wrapf(err, "workflow.Override", "skip open tasks")
		}
		if err := rt.emit(ctx, db, inst, def, "overridden", map[string]any{
			"instance_id": inst.ID.String(), "to": to, "reason": reason,
		}); err != nil {
			return err
		}
		if target.Type == StepTerminal {
			return rt.enterStep(ctx, db, &inst, def, to, uuid.Nil)
		}
		// Jump: set current_step then enter it (create tasks / run auto).
		if err := rt.setCurrentStep(ctx, db, &inst, to); err != nil {
			return err
		}
		return rt.enterStep(ctx, db, &inst, def, to, uuid.Nil)
	})
}

// Instance loads an instance by id in a read-only tenant transaction.
func (rt *Runtime) Instance(ctx context.Context, id uuid.UUID) (Instance, error) {
	var inst Instance
	err := rt.txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		inst, e = rt.loadInstance(ctx, db, id)
		return e
	})
	return inst, err
}

// openTasksSort is the fixed (created_at, id) keyset order for OpenTasksFor.
// Building it through the filtering allowlist gives the cursor a sort-spec
// signature so a forged or stale cursor is rejected loudly on decode (roadmap
// R7/CA-2) instead of the previous legacy unsigned cursor.
var openTasksSort = mustSort(filtering.SortAllowlist{
	"created_at": {Col: "t.created_at"},
	"id":         {Col: "t.id"},
})

func mustSort(allow filtering.SortAllowlist) filtering.Sort {
	s, err := filtering.ParseSort("created_at,id", allow)
	if err != nil {
		panic("workflow: invalid openTasksSort spec: " + err.Error())
	}
	return s
}

// OpenTasksFor lists the open tasks an actor may act on (capacity assignee or
// delegate), cursor-paginated by (created_at, id) with a signed (versioned)
// keyset cursor.
func (rt *Runtime) OpenTasksFor(ctx context.Context, a authz.Actor, cur pagination.Request) (pagination.CursorPage[Task], error) {
	var page pagination.CursorPage[Task]
	limit := cur.Limit
	if limit <= 0 {
		limit = 50
	}
	err := rt.txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		args := []any{a.CapacityID.String(), a.CapacityID, limit + 1}
		// KeysetClause verifies the cursor's sort-spec signature and builds the
		// injection-safe "rows after the cursor" predicate; a forged/stale cursor
		// (wrong sort) fails with KindValidation instead of silently mis-paging.
		clause, cargs, _, cerr := filtering.KeysetClause(openTasksSort, cur.Cursor, 4)
		if cerr != nil {
			return cerr
		}
		where := ""
		if clause != "" {
			where = " AND " + clause
			args = append(args, cargs...)
		}
		rows, err := db.Query(ctx,
			`SELECT DISTINCT t.id, t.instance_id, t.step_key, t.task_type, t.status,
			        t.due_at, t.remind_after, t.decided_by, t.delegated_to, t.output, t.version, t.created_at
			   FROM workflow_tasks t
			   JOIN workflow_task_assignees wa ON wa.task_id = t.id
			  WHERE t.status = 'open'
			    AND ((wa.assignee_kind = 'capacity' AND wa.assignee_ref = $1) OR t.delegated_to = $2)`+where+`
			  ORDER BY t.created_at, t.id
			  LIMIT $3`, args...)
		if err != nil {
			return kerr.Wrapf(err, "workflow.OpenTasksFor", "query tasks")
		}
		defer rows.Close()
		// createdAts parallels page.Items (Task carries no created_at field) so the
		// keyset cursor can be encoded from the last RETURNED item, not the
		// lookahead row.
		var createdAts []time.Time
		for rows.Next() {
			var t Task
			var createdAt time.Time
			var out []byte
			if err := rows.Scan(&t.ID, &t.InstanceID, &t.StepKey, &t.TaskType, &t.Status,
				&t.DueAt, &t.RemindAfter, &t.DecidedBy, &t.DelegatedTo, &out, &t.Version, &createdAt); err != nil {
				return kerr.Wrapf(err, "workflow.OpenTasksFor", "scan task")
			}
			t.Output = decodeJSONMap(out)
			page.Items = append(page.Items, t)
			createdAts = append(createdAts, createdAt)
		}
		if err := rows.Err(); err != nil {
			return kerr.Wrapf(err, "workflow.OpenTasksFor", "iterate tasks")
		}
		if len(page.Items) > limit {
			page.Items = page.Items[:limit]
			page.HasMore = true
			// The cursor must point at the last item ACTUALLY RETURNED (index
			// limit-1). Encoding the dropped lookahead row (index limit) would make
			// the next `> cursor` query skip it entirely.
			c, err := filtering.NextCursor(openTasksSort, map[string]any{
				"t.created_at": createdAts[limit-1], "t.id": page.Items[limit-1].ID,
			})
			if err != nil {
				return kerr.Wrapf(err, "workflow.OpenTasksFor", "encode cursor")
			}
			page.NextCursor = c
		}
		return nil
	})
	return page, err
}

// ---------------------------------------------------------------------------
// step machinery
// ---------------------------------------------------------------------------

// enterStep performs the work of arriving at a step: create tasks (approval/
// task/vote), run the action (auto), branch (gateway), or end the instance
// (terminal). It emits the matching outbox event. It assumes current_step is
// already set to stepKey (StartIn/Override handle that).
func (rt *Runtime) enterStep(ctx context.Context, db database.TenantDB, inst *Instance, def Definition, stepKey string, actor uuid.UUID) error {
	step, ok := def.Steps[stepKey]
	if !ok {
		return kerr.E(kerr.KindInternal, "unknown_step", "definition has no step "+stepKey)
	}
	switch step.Type {
	case StepApproval, StepTask, StepVote:
		if err := rt.createTask(ctx, db, inst, stepKey, step, actor); err != nil {
			return err
		}
		return rt.emit(ctx, db, *inst, def, "task_created", map[string]any{
			"instance_id": inst.ID.String(), "step": stepKey, "task_type": string(step.Type),
		})
	case StepAuto:
		return rt.runAuto(ctx, db, inst, def, stepKey, step, actor)
	case StepGateway:
		next := rt.gatewayTarget(step, inst.Context)
		if next == "" {
			return kerr.E(kerr.KindWorkflowState, "gateway_no_branch",
				"gateway step "+stepKey+" matched no branch and has no default")
		}
		return rt.advanceTo(ctx, db, inst, def, next, actor)
	case StepTerminal:
		return rt.terminate(ctx, db, inst, def, step)
	default:
		return kerr.E(kerr.KindInternal, "unknown_step_type", "unhandled step type "+string(step.Type))
	}
}

// advance follows a decision/next transition (may be nil = dead end) to its
// target step, setting current_step first.
func (rt *Runtime) advance(ctx context.Context, db database.TenantDB, inst *Instance, def Definition, t *Transition, actor uuid.UUID) error {
	tgt := t.target()
	if tgt == "" {
		return kerr.E(kerr.KindWorkflowState, "no_transition",
			"step "+inst.CurrentStep+" has no transition for this outcome")
	}
	return rt.advanceTo(ctx, db, inst, def, tgt, actor)
}

// advanceTo sets current_step to tgt and enters it.
func (rt *Runtime) advanceTo(ctx context.Context, db database.TenantDB, inst *Instance, def Definition, tgt string, actor uuid.UUID) error {
	if _, ok := def.Steps[tgt]; !ok {
		return kerr.E(kerr.KindInternal, "unknown_step", "transition targets unknown step "+tgt)
	}
	if err := rt.setCurrentStep(ctx, db, inst, tgt); err != nil {
		return err
	}
	return rt.enterStep(ctx, db, inst, def, tgt, actor)
}

// runAuto invokes the registered action; on success merges output + advances via
// Next, on error follows on_error.
func (rt *Runtime) runAuto(ctx context.Context, db database.TenantDB, inst *Instance, def Definition, stepKey string, step Step, actor uuid.UUID) error {
	fn, ok := rt.registry.auto(step.Action)
	if !ok {
		return kerr.E(kerr.KindInternal, "auto_action_unregistered",
			"auto step "+stepKey+" references unregistered action "+step.Action)
	}
	out, err := fn(ctx, AutoInput{InstanceID: inst.ID.String(), Resource: inst.Resource, Step: stepKey, Context: inst.Context})
	if err != nil {
		if step.OnError.target() == "" {
			return kerr.Wrapf(err, "workflow.runAuto", "auto action %s failed with no on_error", step.Action)
		}
		if e := rt.emit(ctx, db, *inst, def, "auto_failed", map[string]any{
			"instance_id": inst.ID.String(), "step": stepKey, "error": err.Error(),
		}); e != nil {
			return e
		}
		return rt.advance(ctx, db, inst, def, step.OnError, actor)
	}
	if len(out) > 0 {
		for k, v := range out {
			inst.Context[k] = v
		}
		if err := rt.saveContext(ctx, db, inst); err != nil {
			return err
		}
	}
	if err := rt.emit(ctx, db, *inst, def, "completed", map[string]any{
		"instance_id": inst.ID.String(), "step": stepKey,
	}); err != nil {
		return err
	}
	return rt.advance(ctx, db, inst, def, step.Next, actor)
}

// terminate ends the instance with the step's outcome.
func (rt *Runtime) terminate(ctx context.Context, db database.TenantDB, inst *Instance, def Definition, step Step) error {
	status := statusForOutcome(step.Outcome)
	tag, err := db.Exec(ctx,
		`UPDATE workflow_instances
		    SET status = $2, current_step = $3, ended_at = now(), version = version + 1, updated_at = now()
		  WHERE id = $1 AND version = $4`,
		inst.ID, status, inst.CurrentStep, inst.Version)
	if err != nil {
		return kerr.Wrapf(err, "workflow.terminate", "update instance")
	}
	if tag.RowsAffected() == 0 {
		return versionConflict(inst.ID)
	}
	inst.Status = status
	inst.Version++
	verb := "completed"
	if status == "rejected" {
		verb = "rejected"
	}
	return rt.emit(ctx, db, *inst, def, verb, map[string]any{
		"instance_id": inst.ID.String(), "outcome": step.Outcome, "status": status,
	})
}

// gatewayTarget evaluates branches against the instance context; first match
// wins, a When==nil branch is the default.
func (rt *Runtime) gatewayTarget(step Step, ctxMap map[string]any) string {
	def := ""
	for _, b := range step.Branches {
		if b.When == nil {
			def = b.Next
			continue
		}
		if fmt.Sprint(ctxMap[b.When.Key]) == fmt.Sprint(b.When.Equals) {
			return b.Next
		}
	}
	return def
}

// createTask inserts a task and its resolved assignees, applying the step SLA.
func (rt *Runtime) createTask(ctx context.Context, db database.TenantDB, inst *Instance, stepKey string, step Step, actor uuid.UUID) error {
	assignees, err := rt.resolveAssignees(ctx, step.Assignees, ResolveInput{
		InstanceID: inst.ID.String(), Resource: inst.Resource, Step: stepKey, Context: inst.Context,
	})
	if err != nil {
		return err
	}
	taskID := rt.idgen.New()
	dueAt, remindAfter := rt.slaTimes(step.SLA)
	if _, err := db.Exec(ctx,
		`INSERT INTO workflow_tasks
		    (id, tenant_id, instance_id, step_key, task_type, status, due_at, remind_after, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, 'open', $5, $6, $7)`,
		taskID, inst.ID, stepKey, string(step.Type), dueAt, remindAfter, actor); err != nil {
		return kerr.Wrapf(err, "workflow.createTask", "insert task")
	}
	for _, as := range assignees {
		if _, err := db.Exec(ctx,
			`INSERT INTO workflow_task_assignees (task_id, tenant_id, assignee_kind, assignee_ref)
			 VALUES ($1, app_tenant_id(), $2, $3) ON CONFLICT DO NOTHING`,
			taskID, string(as.Kind), as.Ref); err != nil {
			return kerr.Wrapf(err, "workflow.createTask", "insert assignee")
		}
	}
	return nil
}

// resolveAssignees turns AssigneeSpecs into concrete assignee rows.
func (rt *Runtime) resolveAssignees(ctx context.Context, specs []AssigneeSpec, in ResolveInput) ([]Assignee, error) {
	var out []Assignee
	for _, spec := range specs {
		switch spec.Kind {
		case SpecActor:
			out = append(out, Assignee{Kind: KindCapacity, Ref: spec.Actor})
		case SpecRole:
			out = append(out, Assignee{Kind: KindRole, Ref: spec.Role})
		case SpecRelationship:
			out = append(out, Assignee{Kind: KindRelationship, Ref: spec.Rel})
		case SpecResourceOwner:
			// Minimal: recorded as a relationship assignee against the resource.
			out = append(out, Assignee{Kind: KindRelationship, Ref: "resource_owner"})
		case SpecResolver:
			fn, ok := rt.registry.resolver(spec.Resolver)
			if !ok {
				return nil, kerr.E(kerr.KindInternal, "resolver_unregistered",
					"assignee resolver not registered: "+spec.Resolver)
			}
			resolved, err := fn(ctx, in)
			if err != nil {
				return nil, kerr.Wrapf(err, "workflow.resolveAssignees", "resolver %s", spec.Resolver)
			}
			out = append(out, resolved...)
		default:
			return nil, kerr.E(kerr.KindValidation, "unknown_assignee_kind",
				"unknown assignee kind: "+spec.Kind)
		}
	}
	return out, nil
}

// authorize is the transition actor re-check: the actor must be a capacity
// assignee (or the delegate) of the task. If an authz Evaluator is wired and the
// task carries role assignees, a best-effort workflow.task.decide check is a
// secondary gate. Denial is KindForbidden.
func (rt *Runtime) authorize(ctx context.Context, db database.TenantDB, task Task, inst Instance, a authz.Actor) error {
	rows, err := db.Query(ctx,
		`SELECT assignee_kind, assignee_ref FROM workflow_task_assignees WHERE task_id = $1`, task.ID)
	if err != nil {
		return kerr.Wrapf(err, "workflow.authorize", "load assignees")
	}
	defer rows.Close()
	var hasRole bool
	for rows.Next() {
		var kind, ref string
		if err := rows.Scan(&kind, &ref); err != nil {
			return kerr.Wrapf(err, "workflow.authorize", "scan assignee")
		}
		if kind == string(KindCapacity) && ref == a.CapacityID.String() {
			return nil // primary gate satisfied
		}
		if kind == string(KindRole) {
			hasRole = true
		}
	}
	if err := rows.Err(); err != nil {
		return kerr.Wrapf(err, "workflow.authorize", "iterate assignees")
	}
	if task.DelegatedTo != nil && *task.DelegatedTo == a.CapacityID {
		return nil
	}
	// Secondary gate: role assignees resolved via authz (best-effort — an
	// unregistered permission or evaluator error falls through to deny).
	if hasRole && rt.authz != nil {
		if dec, err := rt.authz.Evaluate(ctx, db, a, decidePermission,
			authz.Target{Scope: authz.ScopeResource, Resource: inst.Resource}); err == nil && dec.Allowed {
			return nil
		}
	}
	return kerr.E(kerr.KindForbidden, "not_an_assignee",
		fmt.Sprintf("actor %s is not an assignee of task %s", a.CapacityID, task.ID))
}

// closeTask sets a task's terminal status with optimistic locking.
func (rt *Runtime) closeTask(ctx context.Context, db database.TenantDB, task Task, status string, decidedBy *uuid.UUID, comment string, output map[string]any) error {
	var out any
	if output != nil {
		b, err := json.Marshal(output)
		if err != nil {
			return kerr.E(kerr.KindValidation, "output_invalid", "task output not JSON-encodable")
		}
		out = b
	}
	var commentArg any
	if comment != "" {
		commentArg = comment
	}
	var decidedByArg any
	if decidedBy != nil {
		decidedByArg = *decidedBy
	}
	tag, err := db.Exec(ctx,
		`UPDATE workflow_tasks
		    SET status = $2,
		        decided_by = $3::uuid,
		        decided_at = CASE WHEN $3::uuid IS NULL THEN decided_at ELSE now() END,
		        decision_comment = $4::text, output = $5::jsonb,
		        version = version + 1, updated_at = now()
		  WHERE id = $1 AND version = $6`,
		task.ID, status, decidedByArg, commentArg, out, task.Version)
	if err != nil {
		return kerr.Wrapf(err, "workflow.closeTask", "update task")
	}
	if tag.RowsAffected() == 0 {
		return versionConflict(task.ID)
	}
	return nil
}

// setCurrentStep advances the instance pointer with optimistic locking.
func (rt *Runtime) setCurrentStep(ctx context.Context, db database.TenantDB, inst *Instance, step string) error {
	tag, err := db.Exec(ctx,
		`UPDATE workflow_instances SET current_step = $2, version = version + 1, updated_at = now()
		 WHERE id = $1 AND version = $3`, inst.ID, step, inst.Version)
	if err != nil {
		return kerr.Wrapf(err, "workflow.setCurrentStep", "update instance")
	}
	if tag.RowsAffected() == 0 {
		return versionConflict(inst.ID)
	}
	inst.CurrentStep = step
	inst.Version++
	return nil
}

// saveContext persists a mutated instance context (no version bump: context is
// additive scratch, not a locking surface).
func (rt *Runtime) saveContext(ctx context.Context, db database.TenantDB, inst *Instance) error {
	b, err := json.Marshal(inst.Context)
	if err != nil {
		return kerr.E(kerr.KindValidation, "context_invalid", "instance context not JSON-encodable")
	}
	if _, err := db.Exec(ctx,
		`UPDATE workflow_instances SET context = $2 WHERE id = $1`, inst.ID, b); err != nil {
		return kerr.Wrapf(err, "workflow.saveContext", "update context")
	}
	return nil
}

// emit writes the workflow.<def>.<verb> outbox event in the caller's tx.
func (rt *Runtime) emit(ctx context.Context, db database.TenantDB, inst Instance, def Definition, verb string, payload map[string]any) error {
	evType := "workflow." + def.Key + "." + verb
	return rt.outbox.Write(ctx, db, outbox.Event{
		Type:     evType,
		Resource: inst.Resource,
		Payload:  payload,
	})
}

// ---------------------------------------------------------------------------
// loaders
// ---------------------------------------------------------------------------

func (rt *Runtime) definitionRow(ctx context.Context, db database.TenantDB, key string) (uuid.UUID, int, error) {
	var id uuid.UUID
	var version int
	// Prefer a tenant override, else the module template (NULL tenant), highest
	// version — RLS already scopes visible rows to this tenant + templates.
	err := db.QueryRow(ctx,
		`SELECT id, version FROM workflow_definitions
		  WHERE key = $1 AND status = 'active'
		  ORDER BY (tenant_id IS NOT NULL) DESC, version DESC
		  LIMIT 1`, key).Scan(&id, &version)
	if err != nil {
		return uuid.Nil, 0, kerr.E(kerr.KindNotFound, "workflow_definition_not_found",
			"no active workflow definition for key "+key)
	}
	return id, version, nil
}

func (rt *Runtime) loadInstance(ctx context.Context, db database.TenantDB, id uuid.UUID) (Instance, error) {
	var inst Instance
	var ctxBytes []byte
	err := db.QueryRow(ctx,
		`SELECT id, definition_id, resource_type, resource_id, current_step, status, context, version
		   FROM workflow_instances WHERE id = $1`, id).
		Scan(&inst.ID, &inst.DefinitionID, &inst.Resource.Type, &inst.Resource.ID,
			&inst.CurrentStep, &inst.Status, &ctxBytes, &inst.Version)
	if err != nil {
		return Instance{}, kerr.E(kerr.KindNotFound, "workflow_instance_not_found",
			"workflow instance not found: "+id.String())
	}
	inst.Context = decodeJSONMap(ctxBytes)
	return inst, nil
}

func (rt *Runtime) loadInstanceAndDef(ctx context.Context, db database.TenantDB, id uuid.UUID) (Instance, Definition, error) {
	inst, err := rt.loadInstance(ctx, db, id)
	if err != nil {
		return Instance{}, Definition{}, err
	}
	def, err := rt.defForInstance(ctx, db, inst.DefinitionID)
	if err != nil {
		return Instance{}, Definition{}, err
	}
	return inst, def, nil
}

// defForInstance resolves the registered Definition for an instance's pinned
// definition row (falling back to parsing the stored jsonb if the registry does
// not carry it — the graph is authoritative either way; auto actions/resolvers
// only exist in the registry).
func (rt *Runtime) defForInstance(ctx context.Context, db database.TenantDB, defID uuid.UUID) (Definition, error) {
	var key string
	var version int
	var raw []byte
	if err := db.QueryRow(ctx,
		`SELECT key, version, definition FROM workflow_definitions WHERE id = $1`, defID).
		Scan(&key, &version, &raw); err != nil {
		return Definition{}, kerr.E(kerr.KindNotFound, "workflow_definition_not_found",
			"workflow definition row not found: "+defID.String())
	}
	if def, ok := rt.registry.definition(key, version); ok {
		return def, nil
	}
	return ParseDefinition(raw)
}

func (rt *Runtime) loadTask(ctx context.Context, db database.TenantDB, id uuid.UUID) (Task, error) {
	var t Task
	var out []byte
	err := db.QueryRow(ctx,
		`SELECT id, instance_id, step_key, task_type, status, due_at, remind_after,
		        decided_by, delegated_to, output, version
		   FROM workflow_tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.InstanceID, &t.StepKey, &t.TaskType, &t.Status, &t.DueAt, &t.RemindAfter,
			&t.DecidedBy, &t.DelegatedTo, &out, &t.Version)
	if err != nil {
		return Task{}, kerr.E(kerr.KindNotFound, "workflow_task_not_found",
			"workflow task not found: "+id.String())
	}
	t.Output = decodeJSONMap(out)
	return t, nil
}

// slaTimes computes due_at/remind_after from now + the SLA ISO durations.
func (rt *Runtime) slaTimes(sla *SLA) (dueAt, remindAfter *time.Time) {
	if sla == nil {
		return nil, nil
	}
	now := rt.now().UTC()
	if d, err := parseISODuration(sla.Due); err == nil && d > 0 {
		t := now.Add(d)
		dueAt = &t
	}
	if d, err := parseISODuration(sla.RemindAfter); err == nil && d > 0 {
		t := now.Add(d)
		remindAfter = &t
	}
	return dueAt, remindAfter
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func versionConflict(id uuid.UUID) error {
	return kerr.E(kerr.KindVersionConflict, "version_conflict",
		"optimistic lock conflict on "+id.String())
}

// statusForOutcome maps a terminal step outcome to an instance status
// (the DB CHECK allows running|completed|rejected|cancelled|overridden).
func statusForOutcome(outcome string) string {
	switch outcome {
	case "rejected":
		return "rejected"
	case "cancelled", "canceled":
		return "cancelled"
	case "overridden":
		return "overridden"
	default:
		return "completed"
	}
}

// actorFromCtx derives the acting principal id for created_by/started_by from
// the tenant context (uuid.Nil when absent — a valid non-NULL value).
func actorFromCtx(ctx context.Context) uuid.UUID {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return uuid.Nil
}

func decodeJSONMap(b []byte) map[string]any {
	if len(b) == 0 {
		return map[string]any{}
	}
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	if m == nil {
		return map[string]any{}
	}
	return m
}

// withAuditActor returns ctx with the actor id bound for audit attribution.
// User actors are attributed by UserID; other actor kinds by CapacityID when
// present, otherwise no actor id is bound (audit row stores NULL actor_id but
// retains actor_kind).
func withAuditActor(ctx context.Context, actor authz.Actor) context.Context {
	switch actor.Kind {
	case authz.ActorUser:
		if actor.UserID != uuid.Nil {
			return database.WithActorID(ctx, actor.UserID)
		}
	case authz.ActorSystem, authz.ActorWebhook:
		if actor.CapacityID != uuid.Nil {
			return database.WithActorID(ctx, actor.CapacityID)
		}
	}
	return ctx
}

func grantIDStr(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}
