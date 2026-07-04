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
