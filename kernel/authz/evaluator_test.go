package authz_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// ---------- fakes ----------
// The fakes ignore the TenantDB (they answer from in-memory state), so the
// evaluator can be exercised with a nil db.

type fakeStore struct {
	assignments []authz.Assignment
	ancestors   map[uuid.UUID][]uuid.UUID
	subtree     map[uuid.UUID][]uuid.UUID
	policies    []authz.Policy
	resourceOrg map[uuid.UUID]uuid.UUID
}

func (f *fakeStore) ActiveAssignments(context.Context, database.TenantDB, authz.Actor, time.Time) ([]authz.Assignment, error) {
	return f.assignments, nil
}
func (f *fakeStore) OrgAncestors(_ context.Context, _ database.TenantDB, id uuid.UUID) ([]uuid.UUID, error) {
	return f.ancestors[id], nil
}
func (f *fakeStore) OrgSubtree(_ context.Context, _ database.TenantDB, id uuid.UUID) ([]uuid.UUID, error) {
	return f.subtree[id], nil
}
func (f *fakeStore) Policies(context.Context, database.TenantDB, authz.Actor, string, string) ([]authz.Policy, error) {
	return f.policies, nil
}
func (f *fakeStore) ResourceOrg(_ context.Context, _ database.TenantDB, ref resource.Ref) (uuid.UUID, error) {
	return f.resourceOrg[ref.ID], nil
}

type fakeRels struct{ has map[string]bool }

func (f fakeRels) Has(_ context.Context, _ database.TenantDB, s authz.Actor, relType string, obj resource.Ref, _ time.Time) (bool, error) {
	return f.has[relType+":"+obj.ID.String()], nil
}

type captureAudit struct{ denials []string }

func (c *captureAudit) AuthzDenial(_ context.Context, _ authz.Actor, perm string, _ authz.Target, reason string) {
	c.denials = append(c.denials, perm+"/"+reason)
}

func registry(t *testing.T, perms ...authz.Permission) *authz.Registry {
	t.Helper()
	r := authz.NewRegistry()
	for _, p := range perms {
		r.Register(p)
	}
	if err := r.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}
	return r
}

func newEval(t *testing.T, store authz.Store, reg *authz.Registry, rels authz.RelationshipChecker, audit authz.AuditSink) authz.Evaluator {
	t.Helper()
	return authz.New(authz.Options{
		Store: store, Registry: reg, Policies: policy.New(),
		Relationships: rels, Audit: audit,
		Now: func() time.Time { return time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC) },
	})
}

var actor = authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), CapacityID: uuid.New(), TenantID: uuid.New()}

// ---------- deny by default ----------

func TestDenyByDefault(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	e := newEval(t, &fakeStore{}, reg, nil, nil)
	d, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed {
		t.Fatal("no assignment must mean denied")
	}
	if d.Reason != "default_deny" {
		t.Errorf("reason = %q", d.Reason)
	}
}

func TestUnregisteredPermissionIsError(t *testing.T) {
	reg := registry(t)
	e := newEval(t, &fakeStore{}, reg, nil, nil)
	_, err := e.Evaluate(context.Background(), nil, actor, "unknown.thing.read", authz.Target{})
	if err == nil {
		t.Fatal("an unregistered permission must error, never silently allow")
	}
}

// ---------- RBAC ----------

func TestRBACTenantScopeAllows(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "core.tenant.admin", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.read"}},
	}}
	e := newEval(t, store, reg, nil, nil)
	d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant})
	if !d.Allowed || d.Reason != "role:core.tenant.admin" {
		t.Fatalf("tenant-scope grant should allow: %+v", d)
	}
}

func TestRBACWrongPermissionDenied(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.approve"}, authz.Permission{Key: "requests.request.read"})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "requests.org.member", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.read"}},
	}}
	e := newEval(t, store, reg, nil, nil)
	d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.approve", authz.Target{Scope: authz.ScopeTenant})
	if d.Allowed {
		t.Fatal("a role without the permission must not allow it")
	}
}

func TestRBACOrgScopeSubtreeCoverage(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	parent, child := uuid.New(), uuid.New()
	store := &fakeStore{
		assignments: []authz.Assignment{
			{RoleKey: "core.org.admin", ScopeKind: authz.ScopeOrg, ScopeID: parent, Perms: []string{"requests.request.read"}},
		},
		ancestors: map[uuid.UUID][]uuid.UUID{child: {child, parent}},
	}
	e := newEval(t, store, reg, nil, nil)
	// Grant at parent org covers a target in the child org.
	d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeOrg, OrgID: child})
	if !d.Allowed {
		t.Fatal("org grant must cover descendant orgs")
	}
	// But not an unrelated org.
	other := uuid.New()
	store.ancestors[other] = []uuid.UUID{other}
	d, _ = e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeOrg, OrgID: other})
	if d.Allowed {
		t.Fatal("org grant must NOT cover unrelated orgs")
	}
}

func TestRBACResourceScopeExact(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	rid := uuid.New()
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "r", ScopeKind: authz.ScopeResource, ScopeID: rid, Perms: []string{"requests.request.read"}},
	}}
	e := newEval(t, store, reg, nil, nil)
	target := authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: rid}}
	if d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.read", target); !d.Allowed {
		t.Fatal("resource-scope grant must allow its exact resource")
	}
	other := authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: uuid.New()}}
	if d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.read", other); d.Allowed {
		t.Fatal("resource-scope grant must not allow a different resource")
	}
}

// ---------- ReBAC ----------

func TestReBACRelationshipGrant(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read", GrantedVia: "requests.assigned_to"})
	rid := uuid.New()
	rels := fakeRels{has: map[string]bool{"requests.assigned_to:" + rid.String(): true}}
	e := newEval(t, &fakeStore{}, reg, rels, nil)
	target := authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: rid}}
	d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.read", target)
	if !d.Allowed || d.Reason != "rel:requests.assigned_to" {
		t.Fatalf("relationship grant should allow: %+v", d)
	}
	// No relationship on a different resource → denied.
	other := authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: uuid.New()}}
	if d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.read", other); d.Allowed {
		t.Fatal("no relationship must mean denied")
	}
}

// ---------- ABAC ----------

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestABACDenyOverridesRBAC(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.approve", Sensitive: true})
	store := &fakeStore{
		assignments: []authz.Assignment{
			{RoleKey: "requests.org.approver", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.approve"}},
		},
		policies: []authz.Policy{{
			ID: uuid.New(), Key: "deny_impersonated_approve", Effect: authz.EffectDeny, Priority: 10,
			Conditions: []authz.Condition{{Attribute: "actor.impersonating", Op: "eq", Value: mustJSON(t, true)}},
		}},
	}
	audit := &captureAudit{}
	e := newEval(t, store, reg, nil, audit)

	// Impersonating actor: RBAC would allow, but the deny policy kills it.
	imp := actor
	imp.ImpersonatorUserID = uuid.New()
	d, _ := e.Evaluate(context.Background(), nil, imp, "requests.request.approve", authz.Target{Scope: authz.ScopeTenant})
	if d.Allowed {
		t.Fatal("deny policy must override an RBAC allow")
	}
	if d.Reason != "policy:deny_impersonated_approve" {
		t.Errorf("reason = %q", d.Reason)
	}
	if len(audit.denials) == 0 {
		t.Error("an explicit deny on a sensitive permission must be audited")
	}

	// Non-impersonating actor: policy does not match, RBAC allows.
	d, _ = e.Evaluate(context.Background(), nil, actor, "requests.request.approve", authz.Target{Scope: authz.ScopeTenant})
	if !d.Allowed {
		t.Fatal("without the deny condition, RBAC should allow")
	}
}

// SEC-25: a deny policy gating on an attribute the evaluator cannot resolve
// must fail CLOSED (deny), never silently not-match and let an RBAC allow stand.
func TestABACDenyUnresolvedAttributeFailsClosed(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.update", Sensitive: true})
	store := &fakeStore{
		assignments: []authz.Assignment{
			{RoleKey: "requests.org.editor", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.update"}},
		},
		// The canonical "locked record" deny gates on resource.status, which the
		// Phase 4 attribute bag does not populate.
		policies: []authz.Policy{{
			ID: uuid.New(), Key: "deny_locked", Effect: authz.EffectDeny, Priority: 10,
			Conditions: []authz.Condition{{Attribute: "resource.status", Op: "eq", Value: mustJSON(t, "locked")}},
		}},
	}
	audit := &captureAudit{}
	e := newEval(t, store, reg, nil, audit)
	d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.update", authz.Target{Scope: authz.ScopeTenant})
	if d.Allowed {
		t.Fatal("a deny policy on an unresolvable attribute must fail closed, not be skipped")
	}
	if len(audit.denials) == 0 {
		t.Error("the fail-closed deny should be audited (sensitive permission)")
	}
}

// SEC-26: a resource_type-scoped assignment with an empty ScopeType must not
// match a typeless target via "" == "".
func TestResourceTypeScopeEmptyTypeNoOverGrant(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "r", ScopeKind: authz.ScopeResourceType, ScopeType: "", Perms: []string{"requests.request.read"}},
	}}
	e := newEval(t, store, reg, nil, nil)
	// Tenant-scope target has no resource type; empty scope type must not match.
	if d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant}); d.Allowed {
		t.Fatal("empty resource-type scope must not grant a typeless target (SEC-26)")
	}
}

func TestABACAllowPolicyGrants(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	store := &fakeStore{policies: []authz.Policy{{
		ID: uuid.New(), Key: "allow_business_hours", Effect: authz.EffectAllow, Priority: 50,
		Conditions: []authz.Condition{{Attribute: "env.hour", Op: "gte", Value: mustJSON(t, 9)}},
	}}}
	e := newEval(t, store, reg, nil, nil)
	// now() is 12:00 → hour 12 >= 9 → allow.
	d, _ := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant})
	if !d.Allowed || d.Reason != "policy:allow_business_hours" {
		t.Fatalf("allow policy should grant when nothing else did: %+v", d)
	}
}

// ---------- sensitive denial audit ----------

func TestSensitiveDenialAudited(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "users.user.admin", Sensitive: true})
	audit := &captureAudit{}
	e := newEval(t, &fakeStore{}, reg, nil, audit)
	d, _ := e.Evaluate(context.Background(), nil, actor, "users.user.admin", authz.Target{Scope: authz.ScopeTenant})
	if d.Allowed {
		t.Fatal("no grant → denied")
	}
	if len(audit.denials) != 1 {
		t.Fatalf("a sensitive-permission denial must be audited, got %d", len(audit.denials))
	}
}

// ---------- Filter ----------

func TestFilterTenantScopeUnrestricted(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.list"})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "core.tenant.admin", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.list"}},
	}}
	e := newEval(t, store, reg, nil, nil)
	f, _ := e.Filter(context.Background(), nil, actor, "requests.request.list", "requests.request")
	if !f.All {
		t.Fatal("tenant-scope grant should yield an unrestricted filter")
	}
}

func TestFilterNoGrantDeniesAll(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.list"})
	e := newEval(t, &fakeStore{}, reg, nil, nil)
	f, _ := e.Filter(context.Background(), nil, actor, "requests.request.list", "requests.request")
	if f.All {
		t.Fatal("no grant must not yield unrestricted")
	}
	if len(f.OrgIDs) != 0 || len(f.ResourceIDs) != 0 {
		t.Fatal("no grant must yield an empty (deny-all) filter")
	}
}

func TestFilterOrgScopeExpandsSubtree(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.list"})
	org, child := uuid.New(), uuid.New()
	store := &fakeStore{
		assignments: []authz.Assignment{
			{RoleKey: "core.org.admin", ScopeKind: authz.ScopeOrg, ScopeID: org, Perms: []string{"requests.request.list"}},
		},
		subtree: map[uuid.UUID][]uuid.UUID{org: {org, child}},
	}
	e := newEval(t, store, reg, nil, nil)
	f, _ := e.Filter(context.Background(), nil, actor, "requests.request.list", "requests.request")
	if f.All || len(f.OrgIDs) != 2 {
		t.Fatalf("org filter should list the subtree orgs: %+v", f)
	}
}
