package testkit

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/policy"
)

// covEvaluator wires the real pg-backed authz stack (Store + Registry + policy
// engine) the way production composes it, so AssertAllowed/AssertDenied run the
// genuine evaluation path against seeded assignments.
func covEvaluator(perms ...string) authz.Evaluator {
	reg := authz.NewRegistry()
	for _, p := range perms {
		reg.Register(authz.Permission{Key: p})
	}
	return authz.New(authz.Options{
		Store:    authz.NewStore(),
		Registry: reg,
		Policies: policy.New(),
	})
}

// TestIntegrationAssertAllowedAndDenied proves the two authz assertion helpers
// on real decisions: a capacity holding a tenant-scoped role that grants the
// permission is ALLOWED; a capacity with no grant is DENIED. Both are exercised
// through their happy paths (the underlying evaluate helper, which both wrap,
// is covered end to end).
func TestIntegrationAssertAllowedAndDenied(t *testing.T) {
	h := NewDB(t)
	tn := CreateTenant(t, h)

	const perm = "cov.authz.read"

	// Granted actor: role → perm, assigned tenant-wide to its capacity.
	grantedUser := CreateUser(t, h)
	grantedCap := CreateCapacity(t, h, tn.ID, grantedUser)
	role := CreateRole(t, h, tn.ID, "cov.authz.role", perm)
	GrantRole(t, h, tn.ID, grantedCap, role, "tenant", nil, "")

	// Ungranted actor: a capacity with no assignment at all.
	deniedUser := CreateUser(t, h)
	deniedCap := CreateCapacity(t, h, tn.ID, deniedUser)

	e := covEvaluator(perm)
	granted := authz.Actor{Kind: authz.ActorUser, UserID: grantedUser, CapacityID: grantedCap, TenantID: tn.ID}
	denied := authz.Actor{Kind: authz.ActorUser, UserID: deniedUser, CapacityID: deniedCap, TenantID: tn.ID}
	target := authz.Target{Scope: authz.ScopeTenant}

	AssertAllowed(t, h, e, granted, perm, target)
	AssertDenied(t, h, e, denied, perm, target)

	// Also drive the unexported evaluate directly to assert the Decision the
	// wrappers key off, covering both allow and deny outcomes explicitly.
	if d := evaluate(t, h, e, granted, perm, target); !d.Allowed {
		t.Fatalf("evaluate(granted) = deny (%s), want allow", d.Reason)
	}
	if d := evaluate(t, h, e, denied, perm, target); d.Allowed {
		t.Fatalf("evaluate(denied) = allow (%s), want deny", d.Reason)
	}
}
