package workflow_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/pagination"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/workflow"
	"github.com/qatoolist/wowapi/testkit"
)

// runtime_extra_test.go — coverage for the runtime error paths, assignee-kind
// resolution, the authz-gated override + role secondary gate, SLA escalation,
// and auto-action failure handling that the existing suites do not reach. All
// tests drive the real Runtime against the test database.

// fakeEvaluator is a deterministic authz.Evaluator: it allows exactly the
// permissions in `allow`, or returns `err` for every call when set.
type fakeEvaluator struct {
	allow map[string]bool
	err   error
}

func (f fakeEvaluator) Evaluate(_ context.Context, _ database.TenantDB, _ authz.Actor, perm string, _ authz.Target) (authz.Decision, error) {
	if f.err != nil {
		return authz.Decision{}, f.err
	}
	return authz.Decision{Allowed: f.allow[perm]}, nil
}

func (f fakeEvaluator) Filter(_ context.Context, _ database.TenantDB, _ authz.Actor, _ string, _ string) (authz.ListFilter, error) {
	return authz.ListFilter{All: true}, nil
}

// buildRT wires a runtime with an optional evaluator, the standard test.approver
// resolver (bound to approverCap), a succeeding provision action, and a failing
// action. Any subset of raws may be registered.
func buildRT(t *testing.T, h *testkit.DBHandle, approverCap uuid.UUID, ev authz.Evaluator, raws ...string) *workflow.Runtime {
	t.Helper()
	reg := workflow.NewRegistry()
	for _, raw := range raws {
		def, err := workflow.ParseDefinition([]byte(raw))
		if err != nil {
			t.Fatalf("parse def: %v", err)
		}
		if err := reg.RegisterDefinition(def); err != nil {
			t.Fatalf("register def: %v", err)
		}
	}
	reg.RegisterAssigneeResolver("test.approver", func(_ context.Context, _ workflow.ResolveInput) ([]workflow.Assignee, error) {
		return []workflow.Assignee{{Kind: workflow.KindCapacity, Ref: approverCap.String()}}, nil
	})
	reg.RegisterAssigneeResolver("test.failresolver", func(_ context.Context, _ workflow.ResolveInput) ([]workflow.Assignee, error) {
		return nil, errors.New("resolver blew up")
	})
	reg.RegisterAutoAction("requests.provision", func(_ context.Context, _ workflow.AutoInput) (map[string]any, error) {
		return map[string]any{"provisioned": true}, nil
	})
	reg.RegisterAutoAction("requests.failing", func(_ context.Context, _ workflow.AutoInput) (map[string]any, error) {
		return nil, errors.New("boom")
	})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry.Err(): %v", err)
	}
	if err := workflow.SyncDefinitions(context.Background(), h.Platform, reg); err != nil {
		t.Fatalf("SyncDefinitions: %v", err)
	}
	return workflow.NewRuntime(h.TxM, reg, ev, outbox.NewWriter(model.UUIDv7()), model.UUIDv7(), audit.New(model.UUIDv7(), nil))
}

// ---------------------------------------------------------------------------
// StartIn validation / resolution error paths.
// ---------------------------------------------------------------------------

func TestIntegrationStartInErrors(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef)
	ctx := testkit.TenantCtx(tn.ID)

	start := func(defKey string, r resource.Ref, input map[string]any) error {
		return h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			_, e := rt.StartIn(ctx, db, defKey, r, input)
			return e
		})
	}

	// Zero resource ref.
	if err := start("requests.approval", resource.Ref{}, nil); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("zero resource must be a validation error, got %v", err)
	}
	// Unknown definition key (no DB row).
	if err := start("requests.ghost", res, nil); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("unknown definition must be not-found, got %v", err)
	}
	// applies_to mismatch: definition applies to requests.request, resource is a
	// different type.
	other := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "other.thing")
	if err := start("requests.approval", other, nil); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("applies_to mismatch must be a validation error, got %v", err)
	}
	// Context not JSON-encodable (a channel cannot marshal).
	if err := start("requests.approval", res, map[string]any{"bad": make(chan int)}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("non-encodable context must be a validation error, got %v", err)
	}
}

func TestIntegrationStartInUnregisteredDefinition(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	// Runtime knows only linearDef. Start selection is registry-authoritative;
	// an unregistered key is rejected before any database lookup or insertion.
	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef)
	ctx := testkit.TenantCtx(tn.ID)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := rt.StartIn(ctx, db, "requests.unreg", res, nil)
		return e
	})
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("unregistered definition key must be a not-found error, got %v", err)
	}
}

// StartInStartedBy: an actor id in context is persisted as started_by/created_by.
func TestIntegrationStartInRecordsActor(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef)

	actorID := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tn.ID), actorID)
	var instID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		instID, e = rt.StartIn(ctx, db, "requests.approval", res, nil)
		return e
	}); err != nil {
		t.Fatalf("StartIn: %v", err)
	}
	var startedBy uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT started_by FROM workflow_instances WHERE id=$1`, instID).Scan(&startedBy); err != nil {
		t.Fatal(err)
	}
	if startedBy != actorID {
		t.Fatalf("started_by = %s, want %s", startedBy, actorID)
	}
}

// ---------------------------------------------------------------------------
// Decide / CompleteTask / Delegate error paths.
// ---------------------------------------------------------------------------

func TestIntegrationDecideErrors(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef, taskDef)
	ctx := testkit.TenantCtx(tn.ID)
	a := actor(tn.ID, userID, cap)

	// Task not found.
	if err := rt.Decide(ctx, uuid.New(), workflow.Decision{Actor: a, Type: workflow.DecisionApprove}); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("decide on missing task must be not-found, got %v", err)
	}

	// A task-type step rejects Decide (not a decision step).
	simTask := testkit.NewWorkflowSim(t, h, rt)
	simTask.Start("requests.task", res, nil)
	taskStepID := openTaskID(t, h, simTask.InstanceID(), "do_work")
	if err := rt.Decide(ctx, taskStepID, workflow.Decision{Actor: a, Type: workflow.DecisionApprove}); kerr.KindOf(err) != kerr.KindWorkflowState {
		t.Fatalf("decide on a task step must be a workflow-state error, got %v", err)
	}

	// An approval task: an unsupported decision type is a validation error.
	res2 := testkit.CreateResource(t, h, tn.ID, "requests.request", nil)
	simAppr := testkit.NewWorkflowSim(t, h, rt)
	simAppr.Start("requests.approval", res2, nil)
	apprID := openTaskID(t, h, simAppr.InstanceID(), "manager_review")
	if err := rt.Decide(ctx, apprID, workflow.Decision{Actor: a, Type: workflow.DecisionType("unsupported")}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("abstain on an approval must be a validation error, got %v", err)
	}

	// Approve once, then a second Decide sees a closed task (not open).
	if err := rt.Decide(ctx, apprID, workflow.Decision{Actor: a, Type: workflow.DecisionApprove}); err != nil {
		t.Fatalf("first approve: %v", err)
	}
	if err := rt.Decide(ctx, apprID, workflow.Decision{Actor: a, Type: workflow.DecisionApprove}); kerr.KindOf(err) != kerr.KindWorkflowState {
		t.Fatalf("second decide on a closed task must be a workflow-state error, got %v", err)
	}
}

func TestIntegrationDecideRequireComment(t *testing.T) {
	const commentDef = `
key: requests.comment
version: 1
applies_to: requests.request
initial_step: review
steps:
  review:
    type: approval
    assignees: [ { kind: resolver, resolver: test.approver } ]
    on_approve: { next: end_done, require_comment: true }
    on_reject:  { next: end_rejected }
  end_done:     { type: terminal, outcome: completed }
  end_rejected: { type: terminal, outcome: rejected }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, commentDef)
	ctx := testkit.TenantCtx(tn.ID)
	a := actor(tn.ID, userID, cap)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.comment", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "review")

	// The on_approve transition requires a comment.
	if err := rt.Decide(ctx, taskID, workflow.Decision{Actor: a, Type: workflow.DecisionApprove}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("missing required comment must be a validation error, got %v", err)
	}
	// With a comment it advances to completed.
	if err := rt.Decide(ctx, taskID, workflow.Decision{Actor: a, Type: workflow.DecisionApprove, Comment: "ok"}); err != nil {
		t.Fatalf("approve with comment: %v", err)
	}
	sim.ExpectStep("end_done").ExpectStatus("completed")
}

func TestIntegrationCompleteTaskNotFound(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	rt := buildRT(t, h, cap, fakeEvaluator{}, taskDef)
	if err := rt.CompleteTask(testkit.TenantCtx(tn.ID), uuid.New(), nil); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("CompleteTask on missing task must be not-found, got %v", err)
	}
}

func TestIntegrationDelegateErrors(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef)
	ctx := testkit.TenantCtx(tn.ID)

	// Missing task.
	if err := rt.Delegate(ctx, uuid.New(), cap, time.Now().Add(time.Hour)); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("delegate on missing task must be not-found, got %v", err)
	}

	// Delegating a closed task fails (not open).
	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "manager_review")
	sim.Approve("manager_review", actor(tn.ID, userID, cap)) // closes the task
	if err := rt.Delegate(ctx, taskID, cap, time.Now().Add(time.Hour)); kerr.KindOf(err) != kerr.KindWorkflowState {
		t.Fatalf("delegate on a closed task must be a workflow-state error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Instance loader.
// ---------------------------------------------------------------------------

func TestIntegrationInstanceLoad(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef)
	ctx := testkit.TenantCtx(tn.ID)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, map[string]any{"amount": 42})

	inst, err := rt.Instance(ctx, sim.InstanceID())
	if err != nil {
		t.Fatalf("Instance: %v", err)
	}
	if inst.ID != sim.InstanceID() || inst.CurrentStep != "manager_review" || inst.Status != "running" {
		t.Fatalf("unexpected instance: %+v", inst)
	}
	if inst.Resource.Type != "requests.request" || inst.Resource.ID != res.ID {
		t.Fatalf("unexpected instance resource: %+v", inst.Resource)
	}

	// Not found.
	if _, err := rt.Instance(ctx, uuid.New()); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("Instance for missing id must be not-found, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Assignee-kind resolution + SLA times.
// ---------------------------------------------------------------------------

func TestIntegrationResolveAssigneeKinds(t *testing.T) {
	const multiDef = `
key: requests.multiassign
version: 1
applies_to: requests.request
initial_step: work
steps:
  work:
    type: task
    assignees:
      - { kind: actor, actor: cap-actor-123 }
      - { kind: role, role: approvers }
      - { kind: relationship, rel: manager_of }
      - { kind: resource_owner }
    next: { next: end_done }
  end_done: { type: terminal, outcome: completed }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, multiDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.multiassign", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "work")

	got := map[string]string{}
	rows, err := h.Admin.Query(context.Background(),
		`SELECT assignee_kind, assignee_ref FROM workflow_task_assignees WHERE task_id=$1`, taskID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var kind, ref string
		if err := rows.Scan(&kind, &ref); err != nil {
			t.Fatal(err)
		}
		got[kind+":"+ref] = ref
	}
	for _, want := range []string{
		"capacity:cap-actor-123",
		"role:approvers",
		"relationship:manager_of",
		"relationship:resource_owner",
	} {
		if _, ok := got[want]; !ok {
			t.Fatalf("missing resolved assignee %q; got %v", want, got)
		}
	}
}

func TestIntegrationSLATimesSet(t *testing.T) {
	const slaDef = `
key: requests.sla
version: 1
applies_to: requests.request
initial_step: review
steps:
  review:
    type: approval
    assignees: [ { kind: resolver, resolver: test.approver } ]
    sla: { due: PT2H, remind_after: PT1H }
    on_approve: { next: end_done }
    on_reject:  { next: end_rejected }
  end_done:     { type: terminal, outcome: completed }
  end_rejected: { type: terminal, outcome: rejected }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, slaDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.sla", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "review")

	var dueAt, remindAfter *time.Time
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT due_at, remind_after FROM workflow_tasks WHERE id=$1`, taskID).Scan(&dueAt, &remindAfter); err != nil {
		t.Fatal(err)
	}
	if dueAt == nil || remindAfter == nil {
		t.Fatalf("SLA times not set: due=%v remind=%v", dueAt, remindAfter)
	}
	if !dueAt.After(*remindAfter) {
		t.Fatalf("due_at (%v) should be after remind_after (%v)", dueAt, remindAfter)
	}
}

// ---------------------------------------------------------------------------
// Authz: override gate + role secondary gate.
// ---------------------------------------------------------------------------

func TestIntegrationOverrideAuthzGate(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
	act := actor(tn.ID, userID, cap)
	ctx := testkit.TenantCtx(tn.ID)

	// Denied evaluator: override is forbidden.
	denyRT := buildRT(t, h, cap, fakeEvaluator{allow: map[string]bool{}}, linearDef)
	simDeny := testkit.NewWorkflowSim(t, h, denyRT)
	simDeny.Start("requests.approval", res, nil)
	if err := denyRT.Override(ctx, act, simDeny.InstanceID(), "end_rejected", "why"); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("override without permission must be forbidden, got %v", err)
	}

	// Evaluator error propagates.
	errRT := buildRT(t, h, cap, fakeEvaluator{err: errors.New("evaluator down")}, linearDef)
	res2 := testkit.CreateResource(t, h, tn.ID, "requests.request", nil)
	simErr := testkit.NewWorkflowSim(t, h, errRT)
	simErr.Start("requests.approval", res2, nil)
	if err := errRT.Override(ctx, act, simErr.InstanceID(), "end_rejected", "why"); err == nil {
		t.Fatal("evaluator error must propagate from Override")
	}

	// Allowed evaluator: override JUMPS to a non-terminal step (provision, an
	// auto step) which runs and drives the instance to completion.
	okRT := buildRT(t, h, cap, fakeEvaluator{allow: map[string]bool{"workflow.instance.override": true}}, linearDef)
	res3 := testkit.CreateResource(t, h, tn.ID, "requests.request", nil)
	simOK := testkit.NewWorkflowSim(t, h, okRT)
	simOK.Start("requests.approval", res3, nil)
	if err := okRT.Override(ctx, act, simOK.InstanceID(), "provision", "manual provision"); err != nil {
		t.Fatalf("permitted override: %v", err)
	}
	// provision (auto) ran → advanced to end_done completed.
	simOK.ExpectStep("end_done").ExpectStatus("completed")
	if n := countEvents(t, h, tn.ID, "workflow.requests.approval.overridden"); n < 1 {
		t.Fatalf("overridden events = %d, want >=1", n)
	}
}

// TestIntegrationOverrideFailsClosedWithoutPermission is the SEC-02 regression
// test: Override must never grant a privileged state jump to an actor who
// lacks workflow.instance.override, and NewRuntime must never accept a nil
// evaluator that could silently disable the check. Before the SEC-02 fix,
// NewRuntime tolerated a nil authz.Evaluator and Override's permission check
// ran only `if rt.authz != nil`, so a runtime built with a nil evaluator let
// ANY actor override ANY running instance. This test proves both halves of
// the fix: (1) construction with a nil evaluator now panics rather than
// silently degrading, and (2) a Runtime built the normal way (non-nil,
// permission-denying evaluator) fails the override closed with KindForbidden.
func TestIntegrationOverrideFailsClosedWithoutPermission(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
	act := actor(tn.ID, userID, cap)
	ctx := testkit.TenantCtx(tn.ID)

	// Half 1: a nil evaluator must never reach a constructed Runtime.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("NewRuntime with a nil evaluator must panic (SEC-02 fail-closed guard)")
			}
		}()
		workflow.NewRuntime(h.TxM, workflow.NewRegistry(), nil, outbox.NewWriter(model.UUIDv7()), model.UUIDv7(), audit.New(model.UUIDv7(), nil))
	}()

	// Half 2: a normally-constructed Runtime (real, non-nil evaluator) with an
	// actor who holds no permissions must fail Override closed, not open.
	rt := buildRT(t, h, cap, fakeEvaluator{allow: map[string]bool{}}, linearDef)
	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)

	err := rt.Override(ctx, act, sim.InstanceID(), "end_rejected", "adversarial: no permission")
	if kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("Override by an unpermitted actor must fail closed with KindForbidden, got %v", err)
	}
	// The instance must NOT have moved — a failed authz check must not have
	// any side effect on workflow state.
	var status, step string
	if qerr := h.Admin.QueryRow(context.Background(),
		`SELECT status, current_step FROM workflow_instances WHERE id=$1`, sim.InstanceID()).Scan(&status, &step); qerr != nil {
		t.Fatal(qerr)
	}
	if status != "running" || step != "manager_review" {
		t.Fatalf("unauthorized override must not mutate instance state, got status=%q step=%q", status, step)
	}
}

func TestIntegrationAuthorizeRoleSecondaryGate(t *testing.T) {
	// An approval step assigned to a ROLE (not a capacity): the deciding actor is
	// not a capacity assignee, so authorization falls to the workflow.task.decide
	// secondary gate resolved via the evaluator.
	const roleDef = `
key: requests.rolegate
version: 1
applies_to: requests.request
initial_step: review
steps:
  review:
    type: approval
    assignees: [ { kind: role, role: approvers } ]
    on_approve: { next: end_done }
    on_reject:  { next: end_rejected }
  end_done:     { type: terminal, outcome: completed }
  end_rejected: { type: terminal, outcome: rejected }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
	ctx := testkit.TenantCtx(tn.ID)
	act := actor(tn.ID, userID, cap)

	// Denied secondary gate: the role holder is not permitted → forbidden.
	denyRT := buildRT(t, h, cap, fakeEvaluator{allow: map[string]bool{}}, roleDef)
	simDeny := testkit.NewWorkflowSim(t, h, denyRT)
	simDeny.Start("requests.rolegate", res, nil)
	denyTask := openTaskID(t, h, simDeny.InstanceID(), "review")
	if err := denyRT.Decide(ctx, denyTask, workflow.Decision{Actor: act, Type: workflow.DecisionApprove}); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("role decide without the decide permission must be forbidden, got %v", err)
	}

	// Allowed secondary gate: the evaluator grants workflow.task.decide.
	okRT := buildRT(t, h, cap, fakeEvaluator{allow: map[string]bool{"workflow.task.decide": true}}, roleDef)
	res2 := testkit.CreateResource(t, h, tn.ID, "requests.request", nil)
	simOK := testkit.NewWorkflowSim(t, h, okRT)
	simOK.Start("requests.rolegate", res2, nil)
	okTask := openTaskID(t, h, simOK.InstanceID(), "review")
	if err := okRT.Decide(ctx, okTask, workflow.Decision{Actor: act, Type: workflow.DecisionApprove}); err != nil {
		t.Fatalf("role decide with the decide permission: %v", err)
	}
	simOK.ExpectStep("end_done").ExpectStatus("completed")
}

// ---------------------------------------------------------------------------
// Auto-action failure handling.
// ---------------------------------------------------------------------------

func TestIntegrationRunAutoOnError(t *testing.T) {
	const autoErrDef = `
key: requests.autoerr
version: 1
applies_to: requests.request
initial_step: review
steps:
  review:
    type: approval
    assignees: [ { kind: resolver, resolver: test.approver } ]
    on_approve: { next: do_fail }
    on_reject:  { next: end_rejected }
  do_fail:
    type: auto
    action: requests.failing
    next: { next: end_done }
    on_error: { then: end_rejected }
  end_done:     { type: terminal, outcome: completed }
  end_rejected: { type: terminal, outcome: rejected }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, autoErrDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.autoerr", res, nil)
	sim.Approve("review", actor(tn.ID, userID, cap))

	// The failing auto action followed on_error → end_rejected, emitting auto_failed.
	sim.ExpectStep("end_rejected").ExpectStatus("rejected")
	if n := countEvents(t, h, tn.ID, "workflow.requests.autoerr.auto_failed"); n != 1 {
		t.Fatalf("auto_failed events = %d, want 1", n)
	}
}

func TestIntegrationRunAutoNoErrorHandler(t *testing.T) {
	// An auto step with NO on_error: a failing action surfaces as an error from
	// the driving call (the instance does not silently advance).
	const autoNoHandlerDef = `
key: requests.autohard
version: 1
applies_to: requests.request
initial_step: do_fail
steps:
  do_fail:
    type: auto
    action: requests.failing
    next: { next: end_done }
  end_done: { type: terminal, outcome: completed }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, autoNoHandlerDef)
	ctx := testkit.TenantCtx(tn.ID)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := rt.StartIn(ctx, db, "requests.autohard", res, nil)
		return e
	})
	if err == nil {
		t.Fatal("auto failure with no on_error must surface as an error")
	}
}

// ---------------------------------------------------------------------------
// Gateway with no matching branch and no default.
// ---------------------------------------------------------------------------

func TestIntegrationGatewayNoBranch(t *testing.T) {
	const gwDef = `
key: requests.gwnodefault
version: 1
applies_to: requests.request
initial_step: route
steps:
  route:
    type: gateway
    branches:
      - when: { key: tier, equals: gold }
        next: end_gold
  end_gold: { type: terminal, outcome: completed }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, gwDef)
	ctx := testkit.TenantCtx(tn.ID)

	// tier=silver matches no branch and there is no default → runtime error.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := rt.StartIn(ctx, db, "requests.gwnodefault", res, map[string]any{"tier": "silver"})
		return e
	})
	if kerr.KindOf(err) != kerr.KindWorkflowState {
		t.Fatalf("gateway with no matching branch must be a workflow-state error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// SLA escalation sweep (expiry + escalate_to target task creation).
// ---------------------------------------------------------------------------

func TestIntegrationSweepSLAEscalation(t *testing.T) {
	const escDef = `
key: requests.escalate
version: 1
applies_to: requests.request
initial_step: review
steps:
  review:
    type: approval
    assignees: [ { kind: resolver, resolver: test.approver } ]
    sla: { due: PT1H, escalate_to: "step:fallback" }
    on_approve: { next: fallback }
    on_reject:  { next: end_rejected }
  fallback:
    type: task
    assignees: [ { kind: resolver, resolver: test.approver } ]
    next: { next: end_done }
  end_done:     { type: terminal, outcome: completed }
  end_rejected: { type: terminal, outcome: rejected }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, escDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.escalate", res, nil)
	reviewTask := openTaskID(t, h, sim.InstanceID(), "review")

	// Force the review task past its due_at.
	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE workflow_tasks SET due_at = now() - interval '1 hour' WHERE id=$1`, reviewTask); err != nil {
		t.Fatal(err)
	}

	var esc int
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		var e error
		_, esc, e = rt.SweepSLA(ctx, db, time.Now())
		return e
	}); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if esc != 1 {
		t.Fatalf("escalations = %d, want 1", esc)
	}

	// The review task is expired.
	var reviewStatus string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM workflow_tasks WHERE id=$1`, reviewTask).Scan(&reviewStatus); err != nil {
		t.Fatal(err)
	}
	if reviewStatus != "expired" {
		t.Fatalf("review task status = %q, want expired", reviewStatus)
	}
	// The instance moved to the escalation target and a fallback task is open there.
	sim.ExpectStep("fallback")
	openTaskID(t, h, sim.InstanceID(), "fallback") // fails if no open fallback task
	if n := countEvents(t, h, tn.ID, "workflow.requests.escalate.escalated"); n != 1 {
		t.Fatalf("escalated events = %d, want 1", n)
	}

	// Idempotent: a second sweep does not re-escalate the now-expired task.
	var esc2 int
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		var e error
		_, esc2, e = rt.SweepSLA(ctx, db, time.Now())
		return e
	}); err != nil {
		t.Fatalf("second sweep: %v", err)
	}
	if esc2 != 0 {
		t.Fatalf("second sweep escalations = %d, want 0 (idempotent)", esc2)
	}
}

// ---------------------------------------------------------------------------
// Terminal outcome → cancelled status (terminate + statusForOutcome live path).
// ---------------------------------------------------------------------------

func TestIntegrationTerminalCancelled(t *testing.T) {
	const cancelDef = `
key: requests.cancel
version: 1
applies_to: requests.request
initial_step: work
steps:
  work:
    type: task
    assignees: [ { kind: resolver, resolver: test.approver } ]
    next: { next: end_cancelled }
  end_cancelled: { type: terminal, outcome: cancelled }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, cancelDef)
	ctx := testkit.TenantCtx(tn.ID)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.cancel", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "work")
	if err := rt.CompleteTask(ctx, taskID, nil); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}
	sim.ExpectStep("end_cancelled").ExpectStatus("cancelled")
}

// ---------------------------------------------------------------------------
// CompleteTask with a non-encodable output → closeTask validation error.
// ---------------------------------------------------------------------------

func TestIntegrationCompleteTaskBadOutput(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, taskDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.task", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "do_work")

	// A channel cannot be JSON-encoded → the output is rejected before the task
	// is closed (the task stays open).
	err := rt.CompleteTask(testkit.TenantCtx(tn.ID), taskID, map[string]any{"bad": make(chan int)})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("non-encodable output must be a validation error, got %v", err)
	}
	sim.ExpectStep("do_work").ExpectStatus("running") // unchanged
}

// ---------------------------------------------------------------------------
// OpenTasksFor with a non-positive limit falls back to the default page size.
// ---------------------------------------------------------------------------

func TestIntegrationOpenTasksForDefaultLimit(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef)
	testkit.NewWorkflowSim(t, h, rt).Start("requests.approval", res, nil)

	// Limit 0 exercises the default-limit branch (limit <= 0 → 50).
	page, err := rt.OpenTasksFor(testkit.TenantCtx(tn.ID), actor(tn.ID, userID, cap), pagination.Request{Limit: 0})
	if err != nil {
		t.Fatalf("OpenTasksFor: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("open tasks = %d, want 1", len(page.Items))
	}
}

// ---------------------------------------------------------------------------
// An assignee resolver that errors aborts the transition.
// ---------------------------------------------------------------------------

func TestIntegrationResolverError(t *testing.T) {
	const failDef = `
key: requests.failresolve
version: 1
applies_to: requests.request
initial_step: work
steps:
  work:
    type: task
    assignees: [ { kind: resolver, resolver: test.failresolver } ]
    next: { next: end_done }
  end_done: { type: terminal, outcome: completed }
`
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, failDef)

	err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		_, e := rt.StartIn(ctx, db, "requests.failresolve", res, nil)
		return e
	})
	if err == nil {
		t.Fatal("a failing assignee resolver must abort StartIn")
	}
}

// ---------------------------------------------------------------------------
// authorize: a task whose delegate is NOT also a capacity assignee is still
// authorized via the delegated_to branch.
// ---------------------------------------------------------------------------

func TestIntegrationAuthorizeDelegateOnly(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "manager_review")

	// A distinct capacity that is NOT an assignee of the task. Set delegated_to
	// directly (without also adding it as a capacity assignee, which Delegate
	// would do) so authorization must rely on the delegated_to branch alone.
	delUser := testkit.CreateUser(t, h)
	delCap := testkit.CreateCapacity(t, h, tn.ID, delUser)
	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE workflow_tasks SET delegated_to = $2 WHERE id = $1`, taskID, delCap); err != nil {
		t.Fatal(err)
	}

	if err := rt.Decide(testkit.TenantCtx(tn.ID), taskID, workflow.Decision{
		Actor: actor(tn.ID, delUser, delCap), Type: workflow.DecisionApprove,
	}); err != nil {
		t.Fatalf("delegate (delegated_to only) must be authorized: %v", err)
	}
	sim.ExpectStep("end_done").ExpectStatus("completed")
}

// TestIntegrationInstanceNullContext exercises the decodeJSONMap defense: a row
// whose context is JSON null must load as an empty (non-nil) map, not nil.
func TestIntegrationInstanceNullContext(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, linearDef)
	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)

	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE workflow_instances SET context = 'null'::jsonb WHERE id=$1`, sim.InstanceID()); err != nil {
		t.Fatal(err)
	}
	inst, err := rt.Instance(testkit.TenantCtx(tn.ID), sim.InstanceID())
	if err != nil {
		t.Fatalf("Instance: %v", err)
	}
	if inst.Context == nil {
		t.Fatal("decodeJSONMap must return a non-nil empty map for a JSON-null context")
	}
	if len(inst.Context) != 0 {
		t.Fatalf("expected empty context, got %v", inst.Context)
	}
}

// Fifth closure-audit regression (2026-07-17): auto actions and assignee
// resolvers receive a deep canonical COPY of the instance context — mutating
// or retaining the input must not steer downstream routing, desynchronize the
// persisted context, or race framework readers (the -race gate covers the
// retained-map variant).
func TestIntegrationCallbacksReceiveIsolatedContext(t *testing.T) {
	const isoDef = `{"key":"requests.iso","version":1,"applies_to":"requests.request",` +
		`"initial_step":"mutate","steps":{` +
		`"mutate":{"type":"auto","action":"requests.mutator","next":{"next":"gate"}},` +
		`"gate":{"type":"gateway","branches":[{"when":{"key":"tier","equals":"gold"},"next":"end_fast"},{"next":"end_slow"}]},` +
		`"end_fast":{"type":"terminal","outcome":"completed"},` +
		`"end_slow":{"type":"terminal","outcome":"rejected"}}}`
	const resDef = `{"key":"requests.reswork","version":1,"applies_to":"requests.request",` +
		`"initial_step":"work","steps":{` +
		`"work":{"type":"task","assignees":[{"kind":"resolver","resolver":"test.mutres"}],"next":{"next":"end_done"}},` +
		`"end_done":{"type":"terminal","outcome":"completed"}}}`

	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	var retained map[string]any
	reg := workflow.NewRegistry()
	for _, raw := range []string{isoDef, resDef} {
		def, err := workflow.ParseDefinition([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if err := reg.RegisterDefinition(def); err != nil {
			t.Fatal(err)
		}
	}
	reg.RegisterAutoAction("requests.mutator", func(_ context.Context, in workflow.AutoInput) (map[string]any, error) {
		in.Context["tier"] = "platinum" // sabotage the input map
		retained = in.Context           // and retain it for later mutation
		return nil, nil                 // no output: the only supported mutation channel stays empty
	})
	reg.RegisterAssigneeResolver("test.mutres", func(_ context.Context, in workflow.ResolveInput) ([]workflow.Assignee, error) {
		in.Context["tier"] = "sabotaged"
		return []workflow.Assignee{{Kind: workflow.KindCapacity, Ref: cap.String()}}, nil
	})
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	if err := workflow.SyncDefinitions(context.Background(), h.Platform, reg); err != nil {
		t.Fatal(err)
	}
	rt := workflow.NewRuntime(h.TxM, reg, fakeEvaluator{}, outbox.NewWriter(model.UUIDv7()), model.UUIDv7(), audit.New(model.UUIDv7(), nil))
	ctx := testkit.TenantCtx(tn.ID)

	var isoInst uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		isoInst, e = rt.StartIn(ctx, db, "requests.iso", res, map[string]any{"tier": "gold"})
		return e
	}); err != nil {
		t.Fatalf("StartIn iso: %v", e2s(err))
	}
	var status, step string
	var ctxJSON []byte
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status, current_step, context FROM workflow_instances WHERE id=$1`, isoInst).Scan(&status, &step, &ctxJSON); err != nil {
		t.Fatal(err)
	}
	// The mutating auto action must NOT have steered the gateway: tier stayed
	// gold, so routing took end_fast.
	if status != "completed" || step != "end_fast" {
		t.Fatalf("auto-action input mutation steered routing: status=%q step=%q", status, step)
	}
	if !strings.Contains(string(ctxJSON), `"gold"`) || strings.Contains(string(ctxJSON), "platinum") {
		t.Fatalf("persisted context altered by callback input mutation: %s", ctxJSON)
	}
	// Mutating the RETAINED map after the fact must be inert (and race-free
	// under the -race gate: it aliases nothing framework-owned).
	if retained == nil {
		t.Fatal("mutator action never ran")
	}
	retained["tier"] = "post-hoc"

	var resInst uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		resInst, e = rt.StartIn(ctx, db, "requests.reswork", res, map[string]any{"tier": "gold"})
		return e
	}); err != nil {
		t.Fatalf("StartIn reswork: %v", err)
	}
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT context FROM workflow_instances WHERE id=$1`, resInst).Scan(&ctxJSON); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(ctxJSON), "sabotaged") {
		t.Fatalf("resolver input mutation reached the persisted context: %s", ctxJSON)
	}
}

func e2s(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

// statefulMarshaler serializes DIFFERENTLY on every call — the sharpest probe
// for double serialization.
type statefulMarshaler struct{ calls *int }

func (m statefulMarshaler) MarshalJSON() ([]byte, error) {
	*m.calls++
	return []byte(fmt.Sprintf("%q", fmt.Sprintf("call-%d", *m.calls))), nil
}

// Fifth closure-audit regression (2026-07-17): CompleteTask canonicalizes its
// output EXACTLY ONCE and uses that same value for task persistence and the
// context merge — a stateful marshaler must not be able to record one output
// on the task row while gateways route on another.
func TestIntegrationCompleteTaskSerializesOutputOnce(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRT(t, h, cap, fakeEvaluator{}, taskDef)
	ctx := testkit.TenantCtx(tn.ID)

	var instID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		instID, e = rt.StartIn(ctx, db, "requests.task", res, nil)
		return e
	}); err != nil {
		t.Fatalf("StartIn: %v", err)
	}
	taskID := openTaskID(t, h, instID, "do_work")

	calls := 0
	if err := rt.CompleteTask(ctx, taskID, map[string]any{"v": statefulMarshaler{calls: &calls}}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}
	if calls != 1 {
		t.Fatalf("output was serialized %d times, want exactly once", calls)
	}
	var taskOut, instCtx []byte
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT output FROM workflow_tasks WHERE id=$1`, taskID).Scan(&taskOut); err != nil {
		t.Fatal(err)
	}
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT context FROM workflow_instances WHERE id=$1`, instID).Scan(&instCtx); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(taskOut), "call-1") || !strings.Contains(string(instCtx), "call-1") {
		t.Fatalf("task output (%s) and merged context (%s) diverged from the single canonical serialization", taskOut, instCtx)
	}
}
