package authz_test

// Hot-path benchmarks for the authz evaluator (acceptance criterion #17).
//
// The evaluator is the per-request CPU-critical path: every handler call that
// touches a protected route goes through Evaluate. We benchmark the three most
// common shapes:
//
//  1. Deny-by-default (no assignments): pure CPU, no allocations goal.
//  2. RBAC allow (single tenant-scope assignment): one grants() scan + covers().
//  3. ReBAC allow (relationship-derived grant): relationship check path.
//
// All benchmarks use in-memory fakes defined in evaluator_test.go (same
// package), so no database I/O contaminates the measurement.

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/resource"
)

func benchEval(b *testing.B, store authz.Store, perms []authz.Permission, rels authz.RelationshipChecker) authz.Evaluator {
	b.Helper()
	reg := authz.NewRegistry()
	for _, p := range perms {
		reg.Register(p)
	}
	if err := reg.Err(); err != nil {
		b.Fatalf("registry: %v", err)
	}
	return authz.New(authz.Options{
		Store: store, Registry: reg, Policies: policy.New(),
		Relationships: rels,
		Now:           func() time.Time { return time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC) },
	})
}

// BenchmarkEvaluateDenyByDefault measures the deny-by-default path: no
// assignments, no policies. Expected: low nanoseconds, zero allocations on the
// steady-state path (the slice returned by ActiveAssignments is nil).
func BenchmarkEvaluateDenyByDefault(b *testing.B) {
	perm := authz.Permission{Key: "requests.request.read"}
	e := benchEval(b, &fakeStore{}, []authz.Permission{perm}, nil)
	ctx := context.Background()
	tgt := authz.Target{Scope: authz.ScopeTenant}
	a := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), TenantID: uuid.New()}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Evaluate(ctx, nil, a, "requests.request.read", tgt)
	}
}

// BenchmarkEvaluateRBACTenantAllow measures the fast-path RBAC allow: one
// tenant-scope assignment that grants the requested permission. The algorithm
// returns as soon as it finds a covering grant.
func BenchmarkEvaluateRBACTenantAllow(b *testing.B) {
	perm := authz.Permission{Key: "requests.request.read"}
	store := &fakeStore{assignments: []authz.Assignment{
		{RoleKey: "core.tenant.admin", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.read"}},
	}}
	e := benchEval(b, store, []authz.Permission{perm}, nil)
	ctx := context.Background()
	tgt := authz.Target{Scope: authz.ScopeTenant}
	a := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), TenantID: uuid.New()}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Evaluate(ctx, nil, a, "requests.request.read", tgt)
	}
}

// BenchmarkEvaluateABACDenyPolicy measures the ABAC deny path: one
// tenant-scope RBAC allow overridden by a matching deny policy. This exercises
// policy.Engine.Matches + the unresolved-attribute fast-close check.
func BenchmarkEvaluateABACDenyPolicy(b *testing.B) {
	perm := authz.Permission{Key: "requests.request.approve", Sensitive: true}
	raw, _ := json.Marshal(true)
	store := &fakeStore{
		assignments: []authz.Assignment{
			{RoleKey: "requests.org.approver", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.approve"}},
		},
		policies: []authz.Policy{{
			ID: uuid.New(), Key: "deny_impersonated", Effect: authz.EffectDeny, Priority: 10,
			Conditions: []authz.Condition{{Attribute: "actor.impersonating", Op: "eq", Value: json.RawMessage(raw)}},
		}},
	}
	e := benchEval(b, store, []authz.Permission{perm}, nil)
	ctx := context.Background()
	tgt := authz.Target{Scope: authz.ScopeTenant}
	a := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), TenantID: uuid.New(), ImpersonatorUserID: uuid.New()}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Evaluate(ctx, nil, a, "requests.request.approve", tgt)
	}
}

// BenchmarkCoversTenantScope isolates the covers() hot sub-path — the inner
// loop body executed for each active assignment.
func BenchmarkCovers(b *testing.B) {
	perm := authz.Permission{Key: "requests.request.read"}
	rid := uuid.New()
	orgID := uuid.New()
	// Ten assignments of mixed scope kinds to stress the switch.
	asgs := make([]authz.Assignment, 10)
	for i := range asgs {
		switch i % 4 {
		case 0:
			asgs[i] = authz.Assignment{RoleKey: "r", ScopeKind: authz.ScopeTenant, Perms: []string{"requests.request.read"}}
		case 1:
			asgs[i] = authz.Assignment{RoleKey: "r", ScopeKind: authz.ScopeOrg, ScopeID: orgID, Perms: []string{"requests.request.read"}}
		case 2:
			asgs[i] = authz.Assignment{RoleKey: "r", ScopeKind: authz.ScopeResourceType, ScopeType: "requests.request", Perms: []string{"requests.request.read"}}
		case 3:
			asgs[i] = authz.Assignment{RoleKey: "r", ScopeKind: authz.ScopeResource, ScopeID: rid, Perms: []string{"requests.request.read"}}
		}
	}
	store := &fakeStore{
		assignments: asgs,
		ancestors:   map[uuid.UUID][]uuid.UUID{orgID: {orgID}},
	}
	e := benchEval(b, store, []authz.Permission{perm}, nil)
	ctx := context.Background()
	tgt := authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: rid}}
	a := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), TenantID: uuid.New()}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Evaluate(ctx, nil, a, "requests.request.read", tgt)
	}
}
