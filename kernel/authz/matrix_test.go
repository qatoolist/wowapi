package authz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/policy"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
)

// TestAuthzMatrix reproduces the blueprint 01 §3 permission matrix as a single
// table-driven test over (actor role, permission, target) → expect allow/deny.
// It exercises RBAC scope coverage, ReBAC relationship grants, and deny-by-
// default together, the way the illustrative matrix intends.
func TestAuthzMatrix(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	reqID := uuid.New()
	assignedReqID := uuid.New()

	reg := registry(t,
		authz.Permission{Key: "requests.request.create"},
		authz.Permission{Key: "requests.request.approve", Sensitive: true},
		authz.Permission{Key: "requests.request.read", GrantedVia: "requests.assigned_to"},
		authz.Permission{Key: "requests.request.list"},
		authz.Permission{Key: "users.user.admin", Sensitive: true},
	)

	// Role → permissions and the scope each actor holds.
	type roleGrant struct {
		perms []string
		scope authz.ScopeKind
		orgID uuid.UUID
	}
	roles := map[string]roleGrant{
		"org.member":   {perms: []string{"requests.request.create"}, scope: authz.ScopeOrg, orgID: orgA},
		"org.approver": {perms: []string{"requests.request.create", "requests.request.approve", "requests.request.read"}, scope: authz.ScopeOrg, orgID: orgA},
		"org.admin":    {perms: []string{"requests.request.create", "requests.request.read", "users.user.admin"}, scope: authz.ScopeOrg, orgID: orgA},
		"tenant.admin": {perms: []string{"users.user.admin", "requests.request.read"}, scope: authz.ScopeTenant},
		"vendor":       {perms: nil, scope: authz.ScopeOrg, orgID: orgA}, // relies on relationship only
	}

	// The request resource lives in orgA; a vendor is assigned_to assignedReqID.
	resourceOrg := map[uuid.UUID]uuid.UUID{reqID: orgA, assignedReqID: orgA}
	ancestors := map[uuid.UUID][]uuid.UUID{orgA: {orgA}, orgB: {orgB}}
	rels := fakeRels{has: map[string]bool{"requests.assigned_to:" + assignedReqID.String(): true}}

	reqTarget := func(id uuid.UUID) authz.Target {
		return authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: id}}
	}

	cases := []struct {
		name   string
		role   string
		perm   string
		target authz.Target
		want   bool
	}{
		{"member creates in own org", "org.member", "requests.request.create", authz.Target{Scope: authz.ScopeOrg, OrgID: orgA}, true},
		{"member cannot approve", "org.member", "requests.request.approve", reqTarget(reqID), false},
		{"member cannot admin users", "org.member", "users.user.admin", authz.Target{Scope: authz.ScopeTenant}, false},
		{"approver approves in org", "org.approver", "requests.request.approve", reqTarget(reqID), true},
		{"approver cannot admin users", "org.approver", "users.user.admin", authz.Target{Scope: authz.ScopeTenant}, false},
		{"org admin admins users", "org.admin", "users.user.admin", authz.Target{Scope: authz.ScopeOrg, OrgID: orgA}, true},
		{"org admin cannot approve", "org.admin", "requests.request.approve", reqTarget(reqID), false},
		{"tenant admin admins users tenant-wide", "tenant.admin", "users.user.admin", authz.Target{Scope: authz.ScopeTenant}, true},
		{"approver cannot reach other org", "org.approver", "requests.request.approve", authz.Target{Scope: authz.ScopeOrg, OrgID: orgB}, false},
		{"vendor reads only assigned via relationship", "vendor", "requests.request.read", reqTarget(assignedReqID), true},
		{"vendor cannot read unassigned", "vendor", "requests.request.read", reqTarget(reqID), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rg := roles[c.role]
			store := &fakeStore{
				assignments: []authz.Assignment{{RoleKey: c.role, ScopeKind: rg.scope, ScopeID: rg.orgID, Perms: rg.perms}},
				ancestors:   ancestors,
				resourceOrg: resourceOrg,
			}
			e := authz.New(authz.Options{
				Store: store, Registry: reg, Policies: policy.New(), Relationships: rels,
				Now: func() time.Time { return time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC) },
			})
			d, err := e.Evaluate(context.Background(), nil, actor, c.perm, c.target)
			if err != nil {
				t.Fatal(err)
			}
			if d.Allowed != c.want {
				t.Errorf("Evaluate(%s, %s) = %v (%s); want %v", c.role, c.perm, d.Allowed, d.Reason, c.want)
			}
		})
	}
}
