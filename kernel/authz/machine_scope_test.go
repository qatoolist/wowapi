package authz_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/policy"
)

// A machine principal (ActorSystem with Scopes) is authorized by its scopes — a
// scope acts like an RBAC grant (roadmap S1). Deny-by-default and the ABAC deny
// pass are preserved.
func TestMachineScopeAuthorizes(t *testing.T) {
	reg := registry(t,
		authz.Permission{Key: "gate.device.read"},
		authz.Permission{Key: "gate.device.update"})
	e := newEval(t, &fakeStore{}, reg, nil, nil)

	machine := authz.Actor{
		Kind:     authz.ActorSystem,
		System:   "apikey:gate-1",
		TenantID: uuid.New(),
		Scopes:   []string{"gate.device.read"},
	}

	// In-scope permission is allowed via the machine fast-path.
	d, err := e.Evaluate(context.Background(), nil, machine, "gate.device.read", authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed || d.Reason != "machine_scope" {
		t.Fatalf("in-scope perm = %v (%s), want allowed/machine_scope", d.Allowed, d.Reason)
	}

	// A permission NOT in the key's scopes is denied (deny-by-default holds).
	d, err = e.Evaluate(context.Background(), nil, machine, "gate.device.update", authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed {
		t.Fatalf("out-of-scope perm must be denied, got allowed (%s)", d.Reason)
	}
}

// TestMachineScopeStillSubjectToABACDeny is the CA-3 regression the review found
// missing: a machine principal whose SCOPE would authorize a permission must
// still be DENIED by a matching ABAC deny policy — a scope authorizes like an
// RBAC grant but never bypasses the deny-first ABAC pass.
func TestMachineScopeStillSubjectToABACDeny(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "gate.device.read", Sensitive: true})
	store := &fakeStore{
		policies: []authz.Policy{{
			ID: uuid.New(), Key: "deny_machine_read", Effect: authz.EffectDeny, Priority: 10,
			Conditions: []authz.Condition{{Attribute: "actor.kind", Op: "eq", Value: mustJSON(t, "system")}},
		}},
	}
	audit := &captureAudit{}
	e := newEval(t, store, reg, nil, audit)

	machine := authz.Actor{
		Kind:     authz.ActorSystem,
		System:   "apikey:gate-1",
		TenantID: uuid.New(),
		Scopes:   []string{"gate.device.read"}, // the scope WOULD authorize…
	}
	d, err := e.Evaluate(context.Background(), nil, machine, "gate.device.read", authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed {
		t.Fatalf("ABAC deny must override a machine scope, got allowed (%s)", d.Reason)
	}
	if d.Reason != "policy:deny_machine_read" {
		t.Errorf("reason = %q, want policy:deny_machine_read", d.Reason)
	}
	if len(audit.denials) == 0 {
		t.Error("an explicit deny on a sensitive permission must be audited")
	}
}

// An internal system actor carries no scopes and must NOT be granted anything by
// the machine fast-path — it stays deny-by-default (this guards against the
// change widening the relay/webhook actors' authority).
func TestMachineScopeEmptyScopesStillDenied(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "gate.device.read"})
	e := authz.New(authz.Options{Store: &fakeStore{}, Registry: reg, Policies: policy.New()})

	internal := authz.Actor{Kind: authz.ActorSystem, System: "outbox-relay", TenantID: uuid.New()}
	d, err := e.Evaluate(context.Background(), nil, internal, "gate.device.read", authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed {
		t.Fatalf("a scopeless system actor must be denied, got allowed (%s)", d.Reason)
	}
}
