package authz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationStoreSystemActorAssignment covers the non-human matching arm of
// ActiveAssignments: an assignment keyed by system_actor (not capacity_id) is
// loaded for a matching ActorSystem, and a resource_type scope round-trips its
// scope_type. A capacity actor with the same system name must NOT pick it up.
func TestIntegrationStoreSystemActorAssignment(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	seedResourceType(t, h, ctx, "gate.device")
	seedPermission(t, h, ctx, "gate.device.read")
	role := seedRole(t, h, ctx, &tenant, "gate-reader")
	seedRolePerm(t, h, ctx, role, "gate.device.read")

	const sysName = "apikey:gate-1"
	scopeType := "gate.device"
	seedSystemAssignment(t, h, ctx, tenant, sysName, role, "resource_type", nil, &scopeType,
		"now() - interval '1 hour'", "NULL")

	store := authz.NewStore()

	asgs := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Assignment, error) {
		return store.ActiveAssignments(ctx, db,
			authz.Actor{Kind: authz.ActorSystem, System: sysName, TenantID: tenant}, time.Now())
	})
	if len(asgs) != 1 {
		t.Fatalf("system actor must load its own assignment, got %d", len(asgs))
	}
	a := asgs[0]
	if a.RoleKey != "gate-reader" || a.ScopeKind != authz.ScopeResourceType || a.ScopeType != scopeType {
		t.Fatalf("assignment = %+v; want gate-reader/resource_type/%s", a, scopeType)
	}
	if !hasAll(a.Perms, "gate.device.read") {
		t.Fatalf("perms = %v; want gate.device.read", a.Perms)
	}

	// A human capacity actor (no system) must not match the system_actor row.
	none := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Assignment, error) {
		return store.ActiveAssignments(ctx, db,
			authz.Actor{Kind: authz.ActorUser, CapacityID: uuid.New(), TenantID: tenant}, time.Now())
	})
	if len(none) != 0 {
		t.Fatalf("a capacity actor must not match a system_actor assignment, got %d", len(none))
	}
}

// TestIntegrationStoreOrgSubtreeNil covers the zero-org short-circuit of
// OrgSubtree (mirrors the OrgAncestors(nil) case already asserted).
func TestIntegrationStoreOrgSubtreeNil(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	store := authz.NewStore()

	got := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]uuid.UUID, error) {
		return store.OrgSubtree(ctx, db, uuid.Nil)
	})
	if got != nil {
		t.Fatalf("OrgSubtree(nil) = %v; want nil", got)
	}
}

// TestIntegrationStoreResourceOrgNullOrg covers the NULL org_id arm: a resource
// with no owning org resolves to the zero uuid (not an error).
func TestIntegrationStoreResourceOrgNullOrg(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	seedResourceType(t, h, ctx, "requests.request")

	id := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO resources (id, tenant_id, resource_type, org_id, label, status, created_by)
		VALUES ($1,$2,$3,NULL,'r','active',$4)`, id, tenant, "requests.request", uuid.Nil)

	store := authz.NewStore()
	got := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) (uuid.UUID, error) {
		return store.ResourceOrg(ctx, db, resource.Ref{Type: "requests.request", ID: id})
	})
	if got != uuid.Nil {
		t.Fatalf("ResourceOrg(null org) = %s; want zero uuid", got)
	}
}

// TestIntegrationStorePoliciesTypeAgnostic covers the type-agnostic policy arm
// (applies_to_resource_type IS NULL and applies_to_permission IS NULL) plus a
// multi-condition load, and asserts such a policy also applies to a typeless
// (empty rt) check.
func TestIntegrationStorePoliciesTypeAgnostic(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	seedPermission(t, h, ctx, "requests.request.approve")

	pid := uuid.New()
	// NULL permission and NULL resource type → applies to every check.
	mustExec(t, h, ctx, `INSERT INTO policies
		(id, tenant_id, key, effect, applies_to_permission, applies_to_resource_type, priority, created_by)
		VALUES ($1,$2,$3,'deny',NULL,NULL,5,$4)`,
		pid, tenant, "deny_break_glass", uuid.Nil)
	mustExec(t, h, ctx, `INSERT INTO policy_conditions (id, policy_id, attribute, op, value)
		VALUES ($1,$2,$3,$4,$5)`, uuid.New(), pid, "actor.break_glass", "eq", `true`)
	mustExec(t, h, ctx, `INSERT INTO policy_conditions (id, policy_id, attribute, op, value)
		VALUES ($1,$2,$3,$4,$5)`, uuid.New(), pid, "actor.kind", "eq", `"user"`)

	store := authz.NewStore()

	// With a resource type: the type-agnostic policy applies.
	pols := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Policy, error) {
		return store.Policies(ctx, db, authz.Actor{TenantID: tenant}, "requests.request.approve", "requests.request")
	})
	if len(pols) != 1 || pols[0].Key != "deny_break_glass" {
		t.Fatalf("type-agnostic policy must apply with a resource type, got %+v", pols)
	}
	if len(pols[0].Conditions) != 2 {
		t.Fatalf("both conditions must load, got %d", len(pols[0].Conditions))
	}

	// Typeless (empty rt) check: the type-agnostic policy still applies (SEC-27
	// only excludes type-BOUND policies from typeless checks).
	pols = inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Policy, error) {
		return store.Policies(ctx, db, authz.Actor{TenantID: tenant}, "requests.request.approve", "")
	})
	if len(pols) != 1 {
		t.Fatalf("type-agnostic policy must also apply to a typeless check, got %d", len(pols))
	}
}

// TestIntegrationStorePoliciesTypeBoundExcludedFromTypeless covers the SEC-27
// exclusion arm: a policy bound to a specific resource type must NOT leak into a
// typeless (empty rt) check.
func TestIntegrationStorePoliciesTypeBoundExcludedFromTypeless(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	seedResourceType(t, h, ctx, "requests.request")
	seedPermission(t, h, ctx, "requests.request.approve")

	pid := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO policies
		(id, tenant_id, key, effect, applies_to_permission, applies_to_resource_type, priority, created_by)
		VALUES ($1,$2,$3,'deny',$4,$5,10,$6)`,
		pid, tenant, "deny_typed", "requests.request.approve", "requests.request", uuid.Nil)

	store := authz.NewStore()
	pols := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Policy, error) {
		return store.Policies(ctx, db, authz.Actor{TenantID: tenant}, "requests.request.approve", "")
	})
	if len(pols) != 0 {
		t.Fatalf("a type-bound policy must not leak into a typeless check (SEC-27), got %d", len(pols))
	}
}

// seedSystemAssignment inserts an actor_assignment keyed by system_actor (rather
// than capacity_id). validFrom/validTo are raw SQL expressions.
func seedSystemAssignment(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID,
	system string, role uuid.UUID, scopeKind string, scopeID *uuid.UUID, scopeType *string, validFrom, validTo string,
) {
	t.Helper()
	var scopeIDArg, scopeTypeArg any
	if scopeID != nil {
		scopeIDArg = *scopeID
	}
	if scopeType != nil {
		scopeTypeArg = *scopeType
	}
	sql := `INSERT INTO actor_assignments
		(id, tenant_id, system_actor, role_id, scope_kind, scope_id, scope_type, valid_from, valid_to, granted_by, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7, ` + validFrom + `, ` + validTo + `, $8, $9)`
	mustExec(t, h, ctx, sql, uuid.New(), tenant, system, role, scopeKind, scopeIDArg, scopeTypeArg, uuid.Nil, uuid.Nil)
}
