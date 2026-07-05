package authz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/resource"
)

var errBoom = errors.New("boom")

// errStore is a Store whose individual reads can be made to fail, so every
// error-propagation branch of Evaluate/Filter is reachable without a database.
type errStore struct {
	assignments []authz.Assignment

	errAssignments error
	errAncestors   error
	errSubtree     error
	errPolicies    error
	errResourceOrg error

	subtree []uuid.UUID
}

func (s *errStore) ActiveAssignments(context.Context, database.TenantDB, authz.Actor, time.Time) ([]authz.Assignment, error) {
	return s.assignments, s.errAssignments
}

func (s *errStore) OrgAncestors(_ context.Context, _ database.TenantDB, id uuid.UUID) ([]uuid.UUID, error) {
	if s.errAncestors != nil {
		return nil, s.errAncestors
	}
	return []uuid.UUID{id}, nil
}

func (s *errStore) OrgSubtree(_ context.Context, _ database.TenantDB, id uuid.UUID) ([]uuid.UUID, error) {
	if s.errSubtree != nil {
		return nil, s.errSubtree
	}
	if s.subtree != nil {
		return s.subtree, nil
	}
	return []uuid.UUID{id}, nil
}

func (s *errStore) Policies(context.Context, database.TenantDB, authz.Actor, string, string) ([]authz.Policy, error) {
	return nil, s.errPolicies
}

func (s *errStore) ResourceOrg(context.Context, database.TenantDB, resource.Ref) (uuid.UUID, error) {
	return uuid.Nil, s.errResourceOrg
}

// errRels is a RelationshipChecker that always fails.
type errRels struct{}

func (errRels) Has(context.Context, database.TenantDB, authz.Actor, string, resource.Ref, time.Time) (bool, error) {
	return false, errBoom
}

// errPolicy is a PolicyEngine that fails on every Matches call.
type errPolicy struct{}

func (errPolicy) Matches([]authz.Condition, map[string]any) (bool, error) { return false, errBoom }

func evalWith(store authz.Store, reg *authz.Registry, rels authz.RelationshipChecker, pe authz.PolicyEngine) authz.Evaluator {
	return authz.New(authz.Options{
		Store: store, Registry: reg, Policies: pe,
		Relationships: rels,
		Now:           func() time.Time { return time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC) },
	})
}

// ---------- Evaluate error propagation ----------

func TestEvaluateAssignmentsError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	e := evalWith(&errStore{errAssignments: errBoom}, reg, nil, policy.New())
	if _, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant}); err == nil {
		t.Fatal("ActiveAssignments error must propagate")
	}
}

func TestEvaluateResourceOrgError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	e := evalWith(&errStore{errResourceOrg: errBoom}, reg, nil, policy.New())
	// A resource target with no explicit OrgID forces a ResourceOrg lookup.
	target := authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: uuid.New()}}
	if _, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", target); err == nil {
		t.Fatal("targetOrg/ResourceOrg error must propagate")
	}
}

func TestEvaluateOrgAncestorsError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	e := evalWith(&errStore{errAncestors: errBoom}, reg, nil, policy.New())
	// An explicit org target skips ResourceOrg but still resolves ancestry.
	target := authz.Target{Scope: authz.ScopeOrg, OrgID: uuid.New()}
	if _, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", target); err == nil {
		t.Fatal("OrgAncestors error must propagate")
	}
}

func TestEvaluateRelationshipError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read", GrantedVia: "requests.assigned_to"})
	e := evalWith(&errStore{}, reg, errRels{}, policy.New())
	target := authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: uuid.New()}}
	if _, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", target); err == nil {
		t.Fatal("relationship-check error must propagate")
	}
}

func TestEvaluatePoliciesError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	e := evalWith(&errStore{errPolicies: errBoom}, reg, nil, policy.New())
	if _, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant}); err == nil {
		t.Fatal("Policies load error must propagate")
	}
}

// TestEvaluateDenyPolicyMatchError drives the deny-pass Matches error: a deny
// policy whose attribute IS resolvable reaches policies.Matches, which fails.
func TestEvaluateDenyPolicyMatchError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	store := &fakeStore{policies: []authz.Policy{{
		ID: uuid.New(), Key: "deny_x", Effect: authz.EffectDeny, Priority: 10,
		Conditions: []authz.Condition{{Attribute: "actor.kind", Op: "eq", Value: mustJSON(t, "user")}},
	}}}
	e := evalWith(store, reg, nil, errPolicy{})
	if _, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant}); err == nil {
		t.Fatal("deny-policy Matches error must propagate")
	}
}

// TestEvaluateAllowPolicyMatchError drives the allow-pass Matches error: with no
// matching deny, the allow loop calls Matches, which fails.
func TestEvaluateAllowPolicyMatchError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	store := &fakeStore{policies: []authz.Policy{{
		ID: uuid.New(), Key: "allow_x", Effect: authz.EffectAllow, Priority: 10,
		Conditions: []authz.Condition{{Attribute: "actor.kind", Op: "eq", Value: mustJSON(t, "user")}},
	}}}
	e := evalWith(store, reg, nil, errPolicy{})
	if _, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant}); err == nil {
		t.Fatal("allow-policy Matches error must propagate")
	}
}

// TestCoversUnknownScopeKindDenies exercises the covers() default arm: an
// assignment carrying an unrecognized scope kind grants nothing.
func TestCoversUnknownScopeKindDenies(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "r", ScopeKind: authz.ScopeKind("bogus"), Perms: []string{"requests.request.read"}},
	}}
	e := newEval(t, store, reg, nil, nil)
	d, err := e.Evaluate(context.Background(), nil, actor, "requests.request.read", authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if d.Allowed {
		t.Fatalf("an unknown scope kind must not cover any target, got allowed (%s)", d.Reason)
	}
}

// ---------- break-glass audit ----------

// TestBreakGlassDecisionAudited proves every break-glass decision is audited,
// even when the outcome is an allow (bannered per 01 §3).
func TestBreakGlassDecisionAudited(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "core.tenant.admin", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.read"}},
	}}
	audit := &captureAudit{}
	e := newEval(t, store, reg, nil, audit)

	bg := actor
	bg.BreakGlass = true
	d, err := e.Evaluate(context.Background(), nil, bg, "requests.request.read", authz.Target{Scope: authz.ScopeTenant})
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allowed {
		t.Fatalf("break-glass actor with a tenant grant should be allowed: %+v", d)
	}
	if len(audit.denials) != 1 {
		t.Fatalf("every break-glass decision must be audited, got %d records", len(audit.denials))
	}
}

// ---------- New wiring guards ----------

func TestNewPanicsOnMissingCollaborators(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.read"})
	cases := map[string]authz.Options{
		"no store":    {Registry: reg, Policies: policy.New()},
		"no registry": {Store: &fakeStore{}, Policies: policy.New()},
		"no policies": {Store: &fakeStore{}, Registry: reg},
	}
	for name, o := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("New(%s) must panic on a missing required collaborator", name)
				}
			}()
			authz.New(o)
		})
	}
}

// ---------- Filter branches ----------

func TestFilterUnregisteredPermissionErrors(t *testing.T) {
	reg := registry(t)
	e := newEval(t, &fakeStore{}, reg, nil, nil)
	if _, err := e.Filter(context.Background(), nil, actor, "unknown.thing.list", "unknown.thing"); err == nil {
		t.Fatal("Filter on an unregistered permission must error")
	}
}

func TestFilterAssignmentsError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.list"})
	e := evalWith(&errStore{errAssignments: errBoom}, reg, nil, policy.New())
	if _, err := e.Filter(context.Background(), nil, actor, "requests.request.list", "requests.request"); err == nil {
		t.Fatal("Filter must propagate an ActiveAssignments error")
	}
}

func TestFilterResourceTypeScopeMatches(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.list"})
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "r", ScopeKind: authz.ScopeResourceType, ScopeType: "requests.request", Perms: []string{"requests.request.list"}},
	}}
	e := newEval(t, store, reg, nil, nil)
	f, err := e.Filter(context.Background(), nil, actor, "requests.request.list", "requests.request")
	if err != nil {
		t.Fatal(err)
	}
	if !f.All {
		t.Fatalf("a resource_type grant matching the listed type is unrestricted: %+v", f)
	}
	// A grant for a DIFFERENT resource type yields no coverage → deny-all.
	f, err = e.Filter(context.Background(), nil, actor, "requests.request.list", "other.thing")
	if err != nil {
		t.Fatal(err)
	}
	if f.All || len(f.OrgIDs) != 0 || len(f.ResourceIDs) != 0 {
		t.Fatalf("a resource_type grant for a different type must not widen the filter: %+v", f)
	}
}

func TestFilterResourceScopeCollectsIDs(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.list"})
	r1, r2 := uuid.New(), uuid.New()
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "r", ScopeKind: authz.ScopeResource, ScopeID: r1, Perms: []string{"requests.request.list"}},
		{RoleKey: "r", ScopeKind: authz.ScopeResource, ScopeID: r2, Perms: []string{"requests.request.list"}},
		{RoleKey: "r", ScopeKind: authz.ScopeResource, ScopeID: r1, Perms: []string{"requests.request.list"}}, // dup
	}}
	e := newEval(t, store, reg, nil, nil)
	f, err := e.Filter(context.Background(), nil, actor, "requests.request.list", "requests.request")
	if err != nil {
		t.Fatal(err)
	}
	if f.All || len(f.OrgIDs) != 0 {
		t.Fatalf("resource-scope grants must restrict to resource ids: %+v", f)
	}
	if len(f.ResourceIDs) != 2 {
		t.Fatalf("resource ids must be de-duplicated: got %v, want 2 unique", f.ResourceIDs)
	}
}

func TestFilterOrgSubtreeError(t *testing.T) {
	reg := registry(t, authz.Permission{Key: "requests.request.list"})
	store := &errStore{
		assignments: []authz.Assignment{
			{RoleKey: "r", ScopeKind: authz.ScopeOrg, ScopeID: uuid.New(), Perms: []string{"requests.request.list"}},
		},
		errSubtree: errBoom,
	}
	e := evalWith(store, reg, nil, policy.New())
	if _, err := e.Filter(context.Background(), nil, actor, "requests.request.list", "requests.request"); err == nil {
		t.Fatal("Filter must propagate an OrgSubtree error")
	}
}

// ---------- registry multi-error join ----------

func TestRegistryErrJoinsMultiple(t *testing.T) {
	r := authz.NewRegistry()
	r.Register(authz.Permission{Key: "BAD.KEY.read"})       // invalid key
	r.Register(authz.Permission{Key: "requests.request.x"}) // invalid verb
	err := r.Err()
	if err == nil {
		t.Fatal("two bad registrations must surface an error")
	}
	// The joined message must mention both failures (the "; " join branch).
	msg := err.Error()
	if !contains(msg, "BAD.KEY.read") || !contains(msg, "requests.request.x") {
		t.Fatalf("joined error must reference both failures: %q", msg)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
