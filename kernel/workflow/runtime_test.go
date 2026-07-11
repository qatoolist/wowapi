package workflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/workflow"
	"github.com/qatoolist/wowapi/testkit"
)

// linearDef: approval -> auto -> terminal(completed); reject -> terminal(rejected).
const linearDef = `
key: requests.approval
version: 1
applies_to: requests.request
initial_step: manager_review
steps:
  manager_review:
    type: approval
    assignees:
      - kind: resolver
        resolver: test.approver
    on_approve: { next: provision }
    on_reject:  { next: end_rejected }
  provision:
    type: auto
    action: requests.provision
    next: { next: end_done }
    on_error: { then: end_rejected }
  end_done:
    type: terminal
    outcome: completed
  end_rejected:
    type: terminal
    outcome: rejected
`

// twoStepDef: two sequential approvals -> terminal(completed) (WorkflowSim self-test).
const twoStepDef = `
key: requests.twostep
version: 1
applies_to: requests.request
initial_step: review1
steps:
  review1:
    type: approval
    assignees: [ { kind: resolver, resolver: test.approver } ]
    on_approve: { next: review2 }
    on_reject:  { next: end_rejected }
  review2:
    type: approval
    assignees: [ { kind: resolver, resolver: test.approver } ]
    on_approve: { next: end_done }
    on_reject:  { next: end_rejected }
  end_done:     { type: terminal, outcome: completed }
  end_rejected: { type: terminal, outcome: rejected }
`

// buildRuntime registers the given raw definitions plus a resolver bound to
// approverCap and a no-op provision auto-action, and returns a wired Runtime.
func buildRuntime(t *testing.T, h *testkit.DBHandle, approverCap uuid.UUID, raws ...string) *workflow.Runtime {
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
	reg.RegisterAutoAction("requests.provision", func(_ context.Context, _ workflow.AutoInput) (map[string]any, error) {
		return map[string]any{"provisioned": true}, nil
	})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry.Err(): %v", err)
	}
	// NewRuntime requires a non-nil evaluator (SEC-02: Override's permission
	// check is unconditional). This suite doesn't exercise the authz gate
	// itself (see runtime_extra_test.go's TestIntegrationOverrideAuthzGate for
	// that), so wire a permissive fake that allows the one permission Override
	// checks.
	ev := fakeEvaluator{allow: map[string]bool{"workflow.instance.override": true}}
	return workflow.NewRuntime(h.TxM, reg, ev, outbox.NewWriter(model.UUIDv7()), model.UUIDv7())
}

func countEvents(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, typ string) int {
	t.Helper()
	var n int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM events_outbox WHERE tenant_id = $1 AND event_type = $2`, tenant, typ).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func actor(tenant, userID, cap uuid.UUID) authz.Actor {
	return authz.Actor{Kind: authz.ActorUser, UserID: userID, CapacityID: cap, TenantID: tenant}
}

// TestIntegrationWorkflowLinearApproval drives the full happy path: Start creates
// an instance + task + task_created event; a non-assignee is denied; the assignee
// approves, the auto action runs, and the instance reaches terminal completed —
// with the outbox events written in the same tx.
func TestIntegrationWorkflowLinearApproval(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	approverCap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, approverCap, linearDef)
	testkit.SeedWorkflowDefinition(t, h, &tn.ID, "requests.approval", 1, "requests.request", nil)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, map[string]any{"amount": 10})
	sim.ExpectStep("manager_review").ExpectStatus("running")

	if n := countEvents(t, h, tn.ID, "workflow.requests.approval.task_created"); n != 1 {
		t.Fatalf("task_created events = %d, want 1", n)
	}

	// A non-assignee cannot decide.
	strangerUser := testkit.CreateUser(t, h)
	strangerCap := testkit.CreateCapacity(t, h, tn.ID, strangerUser)
	var taskID uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT id FROM workflow_tasks WHERE instance_id=$1 AND step_key='manager_review' AND status='open'`,
		sim.InstanceID()).Scan(&taskID); err != nil {
		t.Fatal(err)
	}
	err := rt.Decide(testkit.TenantCtx(tn.ID), taskID, workflow.Decision{
		Actor: actor(tn.ID, strangerUser, strangerCap), Type: workflow.DecisionApprove,
	})
	if err == nil {
		t.Fatal("expected non-assignee decide to be denied")
	}

	// The assignee approves → auto runs → terminal completed.
	sim.Approve("manager_review", actor(tn.ID, userID, approverCap))
	sim.ExpectStep("end_done").ExpectStatus("completed")

	if n := countEvents(t, h, tn.ID, "workflow.requests.approval.approved"); n != 1 {
		t.Fatalf("approved events = %d, want 1", n)
	}
	if n := countEvents(t, h, tn.ID, "workflow.requests.approval.completed"); n < 1 {
		t.Fatalf("completed events = %d, want >=1", n)
	}
	// The auto action's output landed in the instance context.
	var provisioned bool
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT (context->>'provisioned')::bool FROM workflow_instances WHERE id=$1`, sim.InstanceID()).Scan(&provisioned); err != nil {
		t.Fatal(err)
	}
	if !provisioned {
		t.Fatal("auto action output not merged into instance context")
	}
}

// TestIntegrationWorkflowReject drives the reject path to terminal rejected.
func TestIntegrationWorkflowReject(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	approverCap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, approverCap, linearDef)
	testkit.SeedWorkflowDefinition(t, h, &tn.ID, "requests.approval", 1, "requests.request", nil)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil).
		Reject("manager_review", actor(tn.ID, userID, approverCap), "not allowed").
		ExpectStep("end_rejected").ExpectStatus("rejected")

	if n := countEvents(t, h, tn.ID, "workflow.requests.approval.rejected"); n < 1 {
		t.Fatalf("rejected events = %d, want >=1", n)
	}
}

// TestIntegrationWorkflowOptimisticLock proves two concurrent Decides on the same
// task resolve to exactly one success (the other hits a version/state conflict).
func TestIntegrationWorkflowOptimisticLock(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	approverCap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, approverCap, linearDef)
	testkit.SeedWorkflowDefinition(t, h, &tn.ID, "requests.approval", 1, "requests.request", nil)

	var instanceID uuid.UUID
	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)
	instanceID = sim.InstanceID()

	var taskID uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT id FROM workflow_tasks WHERE instance_id=$1 AND status='open'`, instanceID).Scan(&taskID); err != nil {
		t.Fatal(err)
	}

	a := actor(tn.ID, userID, approverCap)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			errs <- rt.Decide(testkit.TenantCtx(tn.ID), taskID, workflow.Decision{Actor: a, Type: workflow.DecisionApprove})
		}()
	}
	e1, e2 := <-errs, <-errs
	nilCount := 0
	if e1 == nil {
		nilCount++
	}
	if e2 == nil {
		nilCount++
	}
	if nilCount != 1 {
		t.Fatalf("expected exactly one successful Decide, got %d (e1=%v e2=%v)", nilCount, e1, e2)
	}
}

// TestIntegrationWorkflowSimTwoStep is the WorkflowSim self-test over a simple
// two-approval definition.
func TestIntegrationWorkflowSimTwoStep(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	approverCap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, approverCap, twoStepDef)
	testkit.SeedWorkflowDefinition(t, h, &tn.ID, "requests.twostep", 1, "requests.request", nil)

	a := actor(tn.ID, userID, approverCap)
	testkit.NewWorkflowSim(t, h, rt).
		Start("requests.twostep", res, nil).
		ExpectStep("review1").
		Approve("review1", a).
		ExpectStep("review2").
		Approve("review2", a).
		ExpectStep("end_done").
		ExpectStatus("completed")
}

// TestIntegrationWorkflowSweepSLA proves the SLA sweeper is idempotent: a second
// sweep does not double-remind.
func TestIntegrationWorkflowSweepSLA(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	approverCap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, approverCap, linearDef)
	testkit.SeedWorkflowDefinition(t, h, &tn.ID, "requests.approval", 1, "requests.request", nil)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)

	// Force the open task past its remind_after.
	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE workflow_tasks SET remind_after = now() - interval '1 hour' WHERE instance_id=$1 AND status='open'`,
		sim.InstanceID()); err != nil {
		t.Fatal(err)
	}

	sweep := func() (int, int) {
		var rem, esc int
		if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
			var e error
			rem, esc, e = rt.SweepSLA(ctx, db, time.Now())
			return e
		}); err != nil {
			t.Fatalf("sweep: %v", err)
		}
		return rem, esc
	}

	if rem, _ := sweep(); rem != 1 {
		t.Fatalf("first sweep reminders = %d, want 1", rem)
	}
	// Idempotent: the last_reminded_at guard prevents a second reminder.
	if rem, _ := sweep(); rem != 0 {
		t.Fatalf("second sweep reminders = %d, want 0 (idempotent)", rem)
	}
	if n := countEvents(t, h, tn.ID, "workflow.requests.approval.reminded"); n != 1 {
		t.Fatalf("reminded events = %d, want 1", n)
	}
}
