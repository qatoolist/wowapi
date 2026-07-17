package workflow_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
	"github.com/qatoolist/wowapi/v2/kernel/workflow"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// failingAuditRedactor mutates an audit Entry so that Record's downstream
// JSON canonicalization fails. This injects a real audit-write failure without
// touching the database layer, proving Override rolls back when audit cannot
// be durably written.
func failingAuditRedactor(e *audit.Entry) {
	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}
	// A channel is not JSON-serializable, so canonicalizeMetadata returns an
	// error and audit.Record fails before any INSERT executes.
	e.Metadata["injected_failure"] = make(chan int)
}

// buildRTWithAudit wires a runtime with the provided evaluator and optional
// audit redactor.
func buildRTWithAudit(t *testing.T, h *testkit.DBHandle, approverCap uuid.UUID, txm database.TxManager, ev authz.Evaluator, redact audit.Redactor, raws ...string) *workflow.Runtime {
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
	return workflow.NewRuntimeWithCompliance(txm, reg, ev, outbox.NewWriter(model.UUIDv7()), model.UUIDv7(), audit.New(model.UUIDv7(), redact))
}

const ratifyDef = `
key: requests.ratify
version: 1
applies_to: requests.request
initial_step: manager_review
ratify_by: approvers
steps:
  manager_review:
    type: approval
    assignees: [ { kind: resolver, resolver: test.approver } ]
    on_approve: { next: end_done }
    on_reject:  { next: end_rejected }
  end_done:     { type: terminal, outcome: completed }
  end_rejected: { type: terminal, outcome: rejected }
`

const ratifyStepDef = `
key: requests.ratify_step
version: 1
applies_to: requests.request
initial_step: manager_review
steps:
  manager_review:
    type: approval
    ratify_by: approvers
    assignees: [ { kind: resolver, resolver: test.approver } ]
    on_approve: { next: end_done }
    on_reject:  { next: end_rejected }
  end_done:     { type: terminal, outcome: completed }
  end_rejected: { type: terminal, outcome: rejected }
`

// TestRatifyByDefinitionRejected proves AC-01: ratify_by-declaring definitions
// are rejected at validation time as an interim, Wave-0-compatible posture.
func TestRatifyByDefinitionRejected(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"definition-level ratify_by", ratifyDef},
		{"step-level ratify_by", ratifyStepDef},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def, err := workflow.ParseDefinition([]byte(tc.raw))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			err = def.Validate(nil, map[string]bool{"test.approver": true})
			if err == nil {
				t.Fatal("expected ratify_by definition to be rejected, got nil error")
			}
			if !strings.Contains(err.Error(), "ratify_by is not yet supported") {
				t.Fatalf("expected ratify_by rejection message, got %v", err)
			}
		})
	}
}

// TestOverrideAuditRowPresent proves that a successful override writes a
// complete audit row (actor, impersonator, grant ID, source/target states,
// reason, ratification outcome) in the same transaction as the state jump.
func TestOverrideAuditRowPresent(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	grantID := uuid.New()
	a := authz.Actor{
		Kind:               authz.ActorUser,
		UserID:             userID,
		CapacityID:         cap,
		TenantID:           tn.ID,
		ImpersonatorUserID: uuid.New(),
		GrantID:            grantID,
	}

	rt := buildRTWithAudit(t, h, cap, h.TxM, fakeEvaluator{allow: map[string]bool{"workflow.instance.override": true}}, nil, linearDef)
	testkit.SeedWorkflowDefinition(t, h, &tn.ID, "requests.approval", 1, "requests.request", nil)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)
	instID := sim.InstanceID()

	ctx := testkit.TenantCtx(tn.ID)
	if err := rt.Override(ctx, a, instID, "end_rejected", "manual intervention"); err != nil {
		t.Fatalf("Override: %v", err)
	}

	var actorID, entityID, impID *uuid.UUID
	var actorKind, reason, oldVal, newVal string
	var metadata map[string]any
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT actor_id, actor_kind, impersonator_id, entity_id, reason, old_value, new_value, metadata
		   FROM audit_logs
		  WHERE tenant_id = $1 AND action = 'workflow.instance.override'`, tn.ID).Scan(
		&actorID, &actorKind, &impID, &entityID, &reason, &oldVal, &newVal, &metadata); err != nil {
		t.Fatalf("query audit log: %v", err)
	}
	if actorID == nil || *actorID != userID {
		t.Fatalf("audit actor_id = %v, want %v", actorID, userID)
	}
	if actorKind != "user" {
		t.Fatalf("audit actor_kind = %q, want user", actorKind)
	}
	if impID == nil || *impID != a.ImpersonatorUserID {
		t.Fatalf("audit impersonator_id = %v, want %v", impID, a.ImpersonatorUserID)
	}
	if entityID == nil || *entityID != instID {
		t.Fatalf("audit entity_id = %v, want %v", entityID, instID)
	}
	if reason != "manual intervention" {
		t.Fatalf("audit reason = %q, want %q", reason, "manual intervention")
	}
	if oldVal != "manager_review" {
		t.Fatalf("audit old_value (source_state) = %q, want manager_review", oldVal)
	}
	if newVal != "end_rejected" {
		t.Fatalf("audit new_value (target_state) = %q, want end_rejected", newVal)
	}
	if metadata == nil {
		t.Fatal("audit metadata nil")
	}
	if got := metadata["grant_id"]; got != grantID.String() {
		t.Fatalf("audit metadata grant_id = %v, want %v", got, grantID)
	}
	if got := metadata["ratification_outcome"]; got != "rejected_interim" {
		t.Fatalf("audit metadata ratification_outcome = %v, want rejected_interim", got)
	}
	if got := metadata["source_state"]; got != "manager_review" {
		t.Fatalf("audit metadata source_state = %v, want manager_review", got)
	}
	if got := metadata["target_state"]; got != "end_rejected" {
		t.Fatalf("audit metadata target_state = %v, want end_rejected", got)
	}
}

// TestOverrideAuditFailureRollsBack proves that an injected audit-write failure
// rolls back the entire override transaction: the instance remains in its
// original step with running status and the open task is not skipped.
func TestOverrideAuditFailureRollsBack(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	userID := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, userID)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	a := actor(tn.ID, userID, cap)
	rt := buildRTWithAudit(t, h, cap, h.TxM, fakeEvaluator{allow: map[string]bool{"workflow.instance.override": true}}, failingAuditRedactor, linearDef)
	testkit.SeedWorkflowDefinition(t, h, &tn.ID, "requests.approval", 1, "requests.request", nil)

	sim := testkit.NewWorkflowSim(t, h, rt)
	sim.Start("requests.approval", res, nil)
	instID := sim.InstanceID()

	ctx := testkit.TenantCtx(tn.ID)
	err := rt.Override(ctx, a, instID, "end_rejected", "should fail")
	if err == nil {
		t.Fatal("expected override to fail on injected audit-write failure")
	}

	var status, step string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status, current_step FROM workflow_instances WHERE id = $1`, instID).Scan(&status, &step); err != nil {
		t.Fatal(err)
	}
	if status != "running" || step != "manager_review" {
		t.Fatalf("instance state changed despite audit failure: status=%q step=%q", status, step)
	}

	var taskStatus string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM workflow_tasks WHERE instance_id = $1 AND step_key = 'manager_review'`, instID).Scan(&taskStatus); err != nil {
		t.Fatal(err)
	}
	if taskStatus != "open" {
		t.Fatalf("open task was mutated despite audit failure: status=%q", taskStatus)
	}

	var auditCount int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM audit_logs WHERE tenant_id = $1 AND action = 'workflow.instance.override'`, tn.ID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 0 {
		t.Fatalf("expected 0 override audit rows after rollback, got %d", auditCount)
	}
}
