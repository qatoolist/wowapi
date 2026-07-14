package testkit

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/workflow"
)

// linearDefCov: approval -> auto -> terminal(completed); reject -> terminal(rejected).
const linearDefCov = `
key: cov.approval
version: 1
applies_to: cov.request
initial_step: manager_review
steps:
  manager_review:
    type: approval
    assignees:
      - kind: resolver
        resolver: cov.approver
    on_approve: { next: provision }
    on_reject:  { next: end_rejected }
  provision:
    type: auto
    action: cov.provision
    next: { next: end_done }
    on_error: { then: end_rejected }
  end_done:
    type: terminal
    outcome: completed
  end_rejected:
    type: terminal
    outcome: rejected
`

// buildCovRuntime wires a workflow.Runtime whose single resolver resolves to
// approverCap and whose auto action is a no-op — mirroring the production wiring
// so the WorkflowSim driver exercises real transitions.
func buildCovRuntime(t *testing.T, h *DBHandle, approverCap uuid.UUID, raws ...string) *workflow.Runtime {
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
	reg.RegisterAssigneeResolver("cov.approver", func(_ context.Context, _ workflow.ResolveInput) ([]workflow.Assignee, error) {
		return []workflow.Assignee{{Kind: workflow.KindCapacity, Ref: approverCap.String()}}, nil
	})
	reg.RegisterAutoAction("cov.provision", func(_ context.Context, _ workflow.AutoInput) (map[string]any, error) {
		return map[string]any{"provisioned": true}, nil
	})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry.Err(): %v", err)
	}
	return workflow.NewRuntime(h.TxM, reg, covEvaluator(), outbox.NewWriter(model.UUIDv7()), model.UUIDv7(), audit.New(model.UUIDv7(), nil))
}

func covActor(tenant, userID, cap uuid.UUID) authz.Actor {
	return authz.Actor{Kind: authz.ActorUser, UserID: userID, CapacityID: cap, TenantID: tenant}
}

// TestIntegrationWorkflowSimApproveFlow drives WorkflowSim through its full
// approve path: NewWorkflowSim, Start (which resolves tenantOf + creates the
// instance), Approve (which locates openTask and Decides), and the ExpectStep /
// ExpectStatus / InstanceID assertions — proving the harness's fluent driver
// against a real Runtime and database.
func TestIntegrationWorkflowSimApproveFlow(t *testing.T) {
	h := NewDB(t)
	tn := CreateTenant(t, h)
	userID := CreateUser(t, h)
	approverCap := CreateCapacity(t, h, tn.ID, userID)
	res := CreateResourceTypeAndResource(t, h, tn.ID, "cov.request")

	rt := buildCovRuntime(t, h, approverCap, linearDefCov)
	SeedWorkflowDefinition(t, h, &tn.ID, "cov.approval", 1, "cov.request", nil)

	sim := NewWorkflowSim(t, h, rt)
	sim.Start("cov.approval", res, map[string]any{"amount": 10})
	sim.ExpectStep("manager_review").ExpectStatus("running")

	if sim.InstanceID() == uuid.Nil {
		t.Fatal("InstanceID must be set after Start")
	}

	sim.Approve("manager_review", covActor(tn.ID, userID, approverCap))
	sim.ExpectStep("end_done").ExpectStatus("completed")
}

// TestIntegrationWorkflowSimRejectFlow drives the Reject transition of the sim.
func TestIntegrationWorkflowSimRejectFlow(t *testing.T) {
	h := NewDB(t)
	tn := CreateTenant(t, h)
	userID := CreateUser(t, h)
	approverCap := CreateCapacity(t, h, tn.ID, userID)
	res := CreateResourceTypeAndResource(t, h, tn.ID, "cov.request")

	rt := buildCovRuntime(t, h, approverCap, linearDefCov)
	SeedWorkflowDefinition(t, h, &tn.ID, "cov.approval", 1, "cov.request", nil)

	NewWorkflowSim(t, h, rt).
		Start("cov.approval", res, nil).
		Reject("manager_review", covActor(tn.ID, userID, approverCap), "not this time").
		ExpectStep("end_rejected").
		ExpectStatus("rejected")
}

// TestSeedWorkflowDefinitionTemplate seeds a module-template definition
// (tenant_id NULL) with an empty raw body, exercising the nil-tenant and
// default-raw branches of SeedWorkflowDefinition, and asserts the row landed.
func TestIntegrationSeedWorkflowDefinitionTemplate(t *testing.T) {
	h := NewDB(t)
	id := SeedWorkflowDefinition(t, h, nil, "cov.template", 2, "cov.request", nil)

	var tenantID *uuid.UUID
	var version int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT tenant_id, version FROM workflow_definitions WHERE id = $1`, id).Scan(&tenantID, &version); err != nil {
		t.Fatalf("load seeded definition: %v", err)
	}
	if tenantID != nil {
		t.Fatalf("template definition tenant_id = %v, want NULL", tenantID)
	}
	if version != 2 {
		t.Fatalf("version = %d, want 2", version)
	}
}
