package testkit

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/workflow"
)

// WorkflowSim is a fluent driver that exercises a workflow.Runtime against a
// real test database (blueprint §1.3):
//
//	NewWorkflowSim(t, h, rt).
//	    Start("requests.approval", res, input).
//	    Approve("manager_review", approver).
//	    ExpectStep("auto_provision").
//	    ExpectStatus("completed")
//
// Every step runs the transition through the Runtime and fails the test on
// error, so a test reads as the state machine it drives.
type WorkflowSim struct {
	t          *testing.T
	h          *DBHandle
	rt         *workflow.Runtime
	tenant     uuid.UUID
	instanceID uuid.UUID
}

// NewWorkflowSim binds a sim to a runtime and DB handle. The tenant is inferred
// from the first Start call's resource (looked up via the resources mirror).
func NewWorkflowSim(t *testing.T, h *DBHandle, rt *workflow.Runtime) *WorkflowSim {
	t.Helper()
	return &WorkflowSim{t: t, h: h, rt: rt}
}

// Start begins an instance in its own tenant transaction and remembers the id.
func (s *WorkflowSim) Start(defKey string, res resource.Ref, input map[string]any) *WorkflowSim {
	s.t.Helper()
	s.tenant = s.tenantOf(res)
	ctx := TenantCtx(s.tenant)
	var id uuid.UUID
	err := s.h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = s.rt.StartIn(ctx, db, defKey, res, input)
		return e
	})
	if err != nil {
		s.t.Fatalf("WorkflowSim.Start(%s): %v", defKey, err)
	}
	s.instanceID = id
	return s
}

// Approve records an approval on the open task at stepKey.
func (s *WorkflowSim) Approve(stepKey string, asActor authz.Actor) *WorkflowSim {
	s.t.Helper()
	taskID := s.openTask(stepKey)
	if err := s.rt.Decide(TenantCtx(s.tenant), taskID, workflow.Decision{
		Actor: asActor, Type: workflow.DecisionApprove,
	}); err != nil {
		s.t.Fatalf("WorkflowSim.Approve(%s): %v", stepKey, err)
	}
	return s
}

// Reject records a rejection (with comment) on the open task at stepKey.
func (s *WorkflowSim) Reject(stepKey string, asActor authz.Actor, comment string) *WorkflowSim {
	s.t.Helper()
	taskID := s.openTask(stepKey)
	if err := s.rt.Decide(TenantCtx(s.tenant), taskID, workflow.Decision{
		Actor: asActor, Type: workflow.DecisionReject, Comment: comment,
	}); err != nil {
		s.t.Fatalf("WorkflowSim.Reject(%s): %v", stepKey, err)
	}
	return s
}

// ExpectStep asserts the instance's current_step.
func (s *WorkflowSim) ExpectStep(stepKey string) *WorkflowSim {
	s.t.Helper()
	var got string
	if err := s.h.Admin.QueryRow(context.Background(),
		`SELECT current_step FROM workflow_instances WHERE id = $1`, s.instanceID).Scan(&got); err != nil {
		s.t.Fatalf("WorkflowSim.ExpectStep: load instance: %v", err)
	}
	if got != stepKey {
		s.t.Fatalf("WorkflowSim.ExpectStep: current_step = %q, want %q", got, stepKey)
	}
	return s
}

// ExpectStatus asserts the instance's status.
func (s *WorkflowSim) ExpectStatus(status string) *WorkflowSim {
	s.t.Helper()
	var got string
	if err := s.h.Admin.QueryRow(context.Background(),
		`SELECT status FROM workflow_instances WHERE id = $1`, s.instanceID).Scan(&got); err != nil {
		s.t.Fatalf("WorkflowSim.ExpectStatus: load instance: %v", err)
	}
	if got != status {
		s.t.Fatalf("WorkflowSim.ExpectStatus: status = %q, want %q", got, status)
	}
	return s
}

// InstanceID returns the started instance id.
func (s *WorkflowSim) InstanceID() uuid.UUID { return s.instanceID }

// openTask finds the single open task for a step, failing if absent.
func (s *WorkflowSim) openTask(stepKey string) uuid.UUID {
	s.t.Helper()
	var id uuid.UUID
	if err := s.h.Admin.QueryRow(context.Background(),
		`SELECT id FROM workflow_tasks
		  WHERE instance_id = $1 AND step_key = $2 AND status = 'open'
		  ORDER BY created_at DESC LIMIT 1`, s.instanceID, stepKey).Scan(&id); err != nil {
		s.t.Fatalf("WorkflowSim: no open task at step %q: %v", stepKey, err)
	}
	return id
}

// tenantOf resolves the tenant owning a resource via the mirror row.
func (s *WorkflowSim) tenantOf(res resource.Ref) uuid.UUID {
	s.t.Helper()
	var tenant uuid.UUID
	if err := s.h.Admin.QueryRow(context.Background(),
		`SELECT tenant_id FROM resources WHERE id = $1`, res.ID).Scan(&tenant); err != nil {
		s.t.Fatalf("WorkflowSim: resolve tenant for resource %s: %v", res.ID, err)
	}
	return tenant
}
