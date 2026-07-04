package authz_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/authz"
)

// A StepUp permission challenges an otherwise-allowed actor that lacks a strong
// auth factor, and admits it once the AMR carries one (roadmap S3).
func TestStepUpChallenge(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	// The actor HAS an RBAC grant for the permission at tenant scope.
	store := &fakeStore{
		assignments: []authz.Assignment{{RoleKey: "biller", ScopeKind: "tenant", Perms: []string{perm}}},
	}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeTenant}

	// Without a strong factor: allowed-but-for-MFA → step-up challenge, not allowed.
	noMFA := authz.Actor{Kind: authz.ActorUser, TenantID: actor.TenantID, UserID: actor.UserID, CapacityID: actor.CapacityID}
	d, err := e.Evaluate(context.Background(), nil, noMFA, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || !d.StepUpRequired || d.Reason != "step_up_required" {
		t.Fatalf("no-MFA on a StepUp perm = %+v, want denied+StepUpRequired", d)
	}

	// With a strong factor in the AMR: allowed.
	withMFA := noMFA
	withMFA.AMR = []string{"pwd", "mfa"}
	d, err = e.Evaluate(context.Background(), nil, withMFA, perm, target)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed || d.StepUpRequired {
		t.Fatalf("MFA-satisfied on a StepUp perm = %+v, want allowed", d)
	}
}

// A StepUp permission the actor was NOT granted is a plain deny — step-up only
// gates an otherwise-allowed decision, it never surfaces for an unauthorized one.
func TestStepUpDoesNotMaskPlainDeny(t *testing.T) {
	const perm = "billing.export.read"
	reg := registry(t, authz.Permission{Key: perm, StepUp: true})
	e := newEval(t, &fakeStore{}, reg, nil, nil) // no assignments → not granted

	d, err := e.Evaluate(context.Background(), nil, actor, perm, authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed || d.StepUpRequired || d.Reason != "default_deny" {
		t.Fatalf("ungranted StepUp perm = %+v, want plain default_deny (no step-up)", d)
	}
}
