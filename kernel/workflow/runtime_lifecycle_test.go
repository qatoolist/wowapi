package workflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/pagination"
	"github.com/qatoolist/wowapi/testkit"
)

// runtime_lifecycle_test.go — QA G2 (workflow runtime): the existing suite covers
// definition validation + the approval decide/reject/optimistic-lock/SLA paths,
// but NOT the task-step lifecycle (CompleteTask), Delegate, the privileged
// Override jump, or gateway routing. These are core runtime behaviors driven here
// through the same buildRuntime/WorkflowSim harness (no new scaffolding).

// taskDef: a single TASK step (not an approval) → terminal.
const taskDef = `
key: requests.task
version: 1
applies_to: requests.request
initial_step: do_work
steps:
  do_work:
    type: task
    assignees: [ { kind: resolver, resolver: test.approver } ]
    next: { next: end_done }
  end_done: { type: terminal, outcome: completed }
`

// gatewayDef: a gateway routes on the instance context's `tier`.
const gatewayDef = `
key: requests.gateway
version: 1
applies_to: requests.request
initial_step: route
steps:
  route:
    type: gateway
    branches:
      - when: { key: tier, equals: gold }
        next: end_gold
      - next: end_other
  end_gold:  { type: terminal, outcome: completed }
  end_other: { type: terminal, outcome: rejected }
`

func openTaskID(t *testing.T, h *testkit.DBHandle, instanceID uuid.UUID, step string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT id FROM workflow_tasks WHERE instance_id=$1 AND step_key=$2 AND status='open'`,
		instanceID, step).Scan(&id); err != nil {
		t.Fatalf("load open task for step %q: %v", step, err)
	}
	return id
}

func TestIntegrationWorkflowCompleteTask(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, cap, taskDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.task", res, nil)
	sim.ExpectStep("do_work").ExpectStatus("running")

	taskID := openTaskID(t, h, sim.InstanceID(), "do_work")
	if err := rt.CompleteTask(testkit.TenantCtx(tn.ID), taskID, map[string]any{"result": "ok"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}
	sim.ExpectStep("end_done").ExpectStatus("completed")

	// The task output was merged into the instance context.
	var result string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT context->>'result' FROM workflow_instances WHERE id=$1`, sim.InstanceID()).Scan(&result); err != nil {
		t.Fatal(err)
	}
	if result != "ok" {
		t.Fatalf("task output not merged into context: result=%q", result)
	}

	// Completing an already-done task is a state conflict, not a double-advance.
	err := rt.CompleteTask(testkit.TenantCtx(tn.ID), taskID, nil)
	if kerr.KindOf(err) != kerr.KindWorkflowState {
		t.Fatalf("re-completing a done task must be a workflow-state error, got %v", err)
	}
}

func TestIntegrationWorkflowCompleteTaskRejectsNonTaskStep(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, cap, linearDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "manager_review")

	// manager_review is an APPROVAL step — CompleteTask must refuse it (an
	// approval advances via Decide, not CompleteTask).
	err := rt.CompleteTask(testkit.TenantCtx(tn.ID), taskID, nil)
	if kerr.KindOf(err) != kerr.KindWorkflowState {
		t.Fatalf("CompleteTask on an approval step must fail with a workflow-state error, got %v", err)
	}
}

func TestIntegrationWorkflowDelegate(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, cap, linearDef)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "manager_review")

	delegateUser := testkit.CreateUser(t, h)
	delegateCap := testkit.CreateCapacity(t, h, tn.ID, delegateUser)
	if err := rt.Delegate(testkit.TenantCtx(tn.ID), taskID, delegateCap, time.Now().Add(24*time.Hour)); err != nil {
		t.Fatalf("Delegate: %v", err)
	}

	// The task stays OPEN, records delegated_to, and the delegate is added as an
	// assignee (the original assignee retains visibility).
	var status string
	var delegatedTo uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status, delegated_to FROM workflow_tasks WHERE id=$1`, taskID).Scan(&status, &delegatedTo); err != nil {
		t.Fatal(err)
	}
	if status != "open" || delegatedTo != delegateCap {
		t.Fatalf("delegate: status=%q delegated_to=%s (want open / %s)", status, delegatedTo, delegateCap)
	}
	var assigneeCount int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM workflow_task_assignees WHERE task_id=$1 AND assignee_ref=$2`,
		taskID, delegateCap.String()).Scan(&assigneeCount); err != nil {
		t.Fatal(err)
	}
	if assigneeCount != 1 {
		t.Fatalf("delegate not added as assignee (count=%d)", assigneeCount)
	}
	if n := countEvents(t, h, tn.ID, "workflow.requests.approval.delegated"); n != 1 {
		t.Fatalf("delegated events = %d, want 1", n)
	}

	// The delegate can now approve (it is an assignee).
	sim.Approve("manager_review", actor(tn.ID, delegateUser, delegateCap))
	sim.ExpectStep("end_done").ExpectStatus("completed")
}

func TestIntegrationWorkflowOverride(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, cap, linearDef) // permissive fake evaluator; mechanics tested here, the gate itself in TestIntegrationOverrideAuthzGate

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)
	instID := sim.InstanceID()
	ctx := testkit.TenantCtx(tn.ID)
	act := actor(tn.ID, userID, cap)

	// Negative: an override with no reason is rejected (auditability).
	if err := rt.Override(ctx, act, instID, "end_rejected", ""); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("empty reason must be a validation error, got %v", err)
	}
	// Negative: an unknown target step is rejected.
	if err := rt.Override(ctx, act, instID, "does_not_exist", "manual"); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("unknown override target must be a validation error, got %v", err)
	}

	// Valid override: jump to the reject terminal; the open manager_review task is
	// skipped and an overridden event is emitted.
	if err := rt.Override(ctx, act, instID, "end_rejected", "manual intervention"); err != nil {
		t.Fatalf("Override: %v", err)
	}
	var taskStatus string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM workflow_tasks WHERE instance_id=$1 AND step_key='manager_review'`, instID).Scan(&taskStatus); err != nil {
		t.Fatal(err)
	}
	if taskStatus != "skipped" {
		t.Fatalf("override should skip the open task, got status=%q", taskStatus)
	}
	if n := countEvents(t, h, tn.ID, "workflow.requests.approval.overridden"); n != 1 {
		t.Fatalf("overridden events = %d, want 1", n)
	}

	// Negative: overriding a no-longer-running instance is a state error.
	if err := rt.Override(ctx, act, instID, "end_done", "again"); kerr.KindOf(err) != kerr.KindWorkflowState {
		t.Fatalf("override on a non-running instance must be a workflow-state error, got %v", err)
	}
}

// TestIntegrationWorkflowOpenTasksForNoSkip is the regression for the keyset
// pagination bug: paging N assignee tasks in pages of `limit` must return EVERY
// task exactly once — the old cursor (encoded from the dropped lookahead row)
// skipped one task per page boundary.
func TestIntegrationWorkflowOpenTasksForNoSkip(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)

	rt := buildRuntime(t, h, cap, linearDef)
	testkit.CreateResourceType(t, h, "requests.request")

	// Start 5 instances → 5 open manager_review tasks assigned to `cap`.
	const total = 5
	for i := 0; i < total; i++ {
		res := testkit.CreateResource(t, h, tn.ID, "requests.request", nil)
		testkit.NewWorkflowSim(t, h, rt).Start("requests.approval", res, nil)
	}

	// Page through with a small limit; collect every returned task id.
	act := actor(tn.ID, userID, cap)
	def := pagination.Defaults{PerPage: 2, MaxPerPage: 100}
	seen := map[uuid.UUID]int{}
	cursor := ""
	pages := 0
	for {
		req, err := pagination.Parse("2", cursor, def)
		if err != nil {
			t.Fatalf("pagination.Parse: %v", err)
		}
		pg, err := rt.OpenTasksFor(testkit.TenantCtx(tn.ID), act, req)
		if err != nil {
			t.Fatalf("OpenTasksFor: %v", err)
		}
		for _, task := range pg.Items {
			seen[task.ID]++
		}
		pages++
		if !pg.HasMore {
			break
		}
		cursor = pg.NextCursor
		if pages > total+2 {
			t.Fatal("pagination did not terminate")
		}
	}

	if len(seen) != total {
		t.Fatalf("paged tasks = %d, want %d (a task was skipped or duplicated across pages)", len(seen), total)
	}
	for id, n := range seen {
		if n != 1 {
			t.Fatalf("task %s returned %d times across pages, want exactly 1", id, n)
		}
	}
}

// TestIntegrationWorkflowOpenTasksForRejectsForeignCursor is the R7/CA-2
// regression: OpenTasksFor now mints signed (sort-versioned) cursors, so a
// cursor carrying a different sort-spec signature must be rejected loudly with
// KindValidation instead of silently mis-paging.
func TestIntegrationWorkflowOpenTasksForRejectsForeignCursor(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	rt := buildRuntime(t, h, cap, linearDef)
	act := actor(tn.ID, userID, cap)

	foreign, err := pagination.EncodeCursorWithSig("some-other-sort", map[string]any{
		"t.created_at": time.Now(), "t.id": uuid.New(),
	})
	if err != nil {
		t.Fatalf("mint foreign cursor: %v", err)
	}
	req, err := pagination.Parse("2", foreign, pagination.Defaults{PerPage: 2, MaxPerPage: 100})
	if err != nil {
		t.Fatalf("pagination.Parse: %v", err)
	}
	if _, err := rt.OpenTasksFor(testkit.TenantCtx(tn.ID), act, req); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("foreign-sort cursor should be rejected with KindValidation, got %v", err)
	}
}

func TestIntegrationWorkflowGatewayRouting(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	rt := buildRuntime(t, h, cap, gatewayDef)

	// tier=gold routes to end_gold (completed).
	simGold := testkit.NewWorkflowSim(t, h, rt)
	simGold.Start("requests.gateway", res, map[string]any{"tier": "gold"})
	simGold.ExpectStep("end_gold").ExpectStatus("completed")

	// A distinct resource with tier=silver falls through to end_other (rejected).
	res2 := testkit.CreateResource(t, h, tn.ID, "requests.request", nil)
	simOther := testkit.NewWorkflowSim(t, h, rt)
	simOther.Start("requests.gateway", res2, map[string]any{"tier": "silver"})
	simOther.ExpectStep("end_other").ExpectStatus("rejected")
}
