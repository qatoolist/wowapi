package authz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationStoreActiveAssignments seeds a role with two permissions and an
// org-scoped assignment for a capacity, then asserts the store returns one
// assignment with both permission keys and the correct scope.
func TestIntegrationStoreActiveAssignments(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	cap1 := seedCapacity(t, h, ctx, tenant)
	org := seedOrg(t, h, ctx, tenant, nil, "Root")

	seedPermission(t, h, ctx, "requests.request.read")
	seedPermission(t, h, ctx, "requests.request.approve")
	role := seedRole(t, h, ctx, &tenant, "approver")
	seedRolePerm(t, h, ctx, role, "requests.request.read")
	seedRolePerm(t, h, ctx, role, "requests.request.approve")
	// Active org-scoped assignment.
	seedAssignment(t, h, ctx, tenant, cap1, role, "org", &org, nil, "now() - interval '1 hour'", "NULL")
	// Expired assignment for the same capacity (must be excluded).
	roleOld := seedRole(t, h, ctx, &tenant, "old")
	seedAssignment(t, h, ctx, tenant, cap1, roleOld, "tenant", nil, nil,
		"now() - interval '2 days'", "now() - interval '1 day'")

	store := authz.NewStore()
	asgs := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Assignment, error) {
		return store.ActiveAssignments(ctx, db, authz.Actor{Kind: authz.ActorUser, CapacityID: cap1, TenantID: tenant}, time.Now())
	})
	if len(asgs) != 1 {
		t.Fatalf("got %d active assignments, want 1 (expired excluded)", len(asgs))
	}
	a := asgs[0]
	if a.RoleKey != "approver" || a.ScopeKind != authz.ScopeOrg || a.ScopeID != org {
		t.Fatalf("assignment = %+v; want role=approver scope=org id=%s", a, org)
	}
	if !hasAll(a.Perms, "requests.request.read", "requests.request.approve") {
		t.Fatalf("perms = %v; want both read and approve", a.Perms)
	}
}

// TestIntegrationStoreOrgTree builds a 3-level org tree and asserts ancestors
// (self-first, upward) and subtree (self + descendants) resolve correctly.
func TestIntegrationStoreOrgTree(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)

	root := seedOrg(t, h, ctx, tenant, nil, "Root")
	mid := seedOrg(t, h, ctx, tenant, &root, "Mid")
	leaf := seedOrg(t, h, ctx, tenant, &mid, "Leaf")
	sibling := seedOrg(t, h, ctx, tenant, &root, "Sibling")

	store := authz.NewStore()

	anc := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]uuid.UUID, error) {
		return store.OrgAncestors(ctx, db, leaf)
	})
	if len(anc) != 3 || anc[0] != leaf {
		t.Fatalf("ancestors = %v; want [leaf mid root] self-first", anc)
	}
	if !hasAllUUID(anc, leaf, mid, root) {
		t.Fatalf("ancestors = %v; want leaf, mid, root", anc)
	}

	sub := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]uuid.UUID, error) {
		return store.OrgSubtree(ctx, db, root)
	})
	if len(sub) != 4 || sub[0] != root {
		t.Fatalf("subtree = %v; want 4 nodes, self-first", sub)
	}
	if !hasAllUUID(sub, root, mid, leaf, sibling) {
		t.Fatalf("subtree = %v; want root, mid, leaf, sibling", sub)
	}

	// Empty org yields empty result, no error.
	got := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]uuid.UUID, error) {
		return store.OrgAncestors(ctx, db, uuid.Nil)
	})
	if got != nil {
		t.Fatalf("OrgAncestors(nil) = %v; want nil", got)
	}
}

// TestIntegrationStoreResourceOrg asserts ResourceOrg returns the owning org and
// the zero uuid for an unknown resource.
func TestIntegrationStoreResourceOrg(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	seedResourceType(t, h, ctx, "requests.request")
	org := seedOrg(t, h, ctx, tenant, nil, "Root")

	id := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO resources (id, tenant_id, resource_type, org_id, label, status, created_by)
		VALUES ($1,$2,$3,$4,'r','active',$5)`, id, tenant, "requests.request", org, uuid.Nil)

	store := authz.NewStore()

	got := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) (uuid.UUID, error) {
		return store.ResourceOrg(ctx, db, resource.Ref{Type: "requests.request", ID: id})
	})
	if got != org {
		t.Fatalf("ResourceOrg(known) = %s; want %s", got, org)
	}
	got = inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) (uuid.UUID, error) {
		return store.ResourceOrg(ctx, db, resource.Ref{Type: "requests.request", ID: uuid.New()})
	})
	if got != uuid.Nil {
		t.Fatalf("ResourceOrg(unknown) = %s; want zero", got)
	}
}

// TestIntegrationStorePolicies seeds an active deny policy with a condition and
// asserts the store loads it (with the condition) and filters by permission.
func TestIntegrationStorePolicies(t *testing.T) {
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
		pid, tenant, "deny_locked", "requests.request.approve", "requests.request", uuid.Nil)
	mustExec(t, h, ctx, `INSERT INTO policy_conditions (id, policy_id, attribute, op, value)
		VALUES ($1,$2,$3,$4,$5)`, uuid.New(), pid, "resource.status", "eq", `"locked"`)

	store := authz.NewStore()

	pols := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Policy, error) {
		return store.Policies(ctx, db, authz.Actor{TenantID: tenant}, "requests.request.approve", "requests.request")
	})
	if len(pols) != 1 {
		t.Fatalf("got %d policies, want 1", len(pols))
	}
	p := pols[0]
	if p.Key != "deny_locked" || p.Effect != authz.EffectDeny || p.Priority != 10 {
		t.Fatalf("policy = %+v; want deny_locked/deny/10", p)
	}
	if len(p.Conditions) != 1 || p.Conditions[0].Attribute != "resource.status" || string(p.Conditions[0].Value) != `"locked"` {
		t.Fatalf("conditions = %+v; want resource.status eq \"locked\"", p.Conditions)
	}

	// A different permission does not match this policy.
	pols = inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Policy, error) {
		return store.Policies(ctx, db, authz.Actor{TenantID: tenant}, "requests.request.read", "requests.request")
	})
	if len(pols) != 0 {
		t.Fatalf("got %d policies for unrelated permission, want 0", len(pols))
	}
}

// inTx runs fn inside a read-only tenant transaction and returns its result,
// failing the test on error — the harness through which store methods (which
// now take the caller's TenantDB) are exercised.
func inTx[T any](t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, fn func(ctx context.Context, db database.TenantDB) (T, error)) T {
	t.Helper()
	var out T
	err := h.TxM.WithTenantRO(database.WithTenantID(context.Background(), tenant),
		func(ctx context.Context, db database.TenantDB) error {
			var e error
			out, e = fn(ctx, db)
			return e
		})
	if err != nil {
		t.Fatalf("inTx: %v", err)
	}
	return out
}

// --- seed helpers (Admin pool bypasses RLS as the superuser login) ---

func seedTenant(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID) {
	t.Helper()
	mustExec(t, h, ctx, `INSERT INTO tenants (id, slug, display_name, created_by) VALUES ($1,$2,$3,$4)`,
		tenant, "t-"+uuid.New().String()[:8], "Tenant", uuid.Nil)
}

func seedResourceType(t *testing.T, h *testkit.DBHandle, ctx context.Context, key string) {
	t.Helper()
	mustExec(t, h, ctx, `INSERT INTO resource_types (key, module, description) VALUES ($1,$2,$3)
		ON CONFLICT (key) DO NOTHING`, key, "requests", "request")
}

func seedCapacity(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO users (id, idp_subject, email, created_by) VALUES ($1,$2,$3,$4)`,
		userID, "idp-"+uuid.New().String()[:8], uuid.New().String()[:8]+"@example.test", uuid.Nil)
	capID := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO acting_capacities (id, tenant_id, user_id, label, created_by)
		VALUES ($1,$2,$3,$4,$5)`, capID, tenant, userID, "member", uuid.Nil)
	return capID
}

func seedOrg(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID, parent *uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var parentArg any
	if parent != nil {
		parentArg = *parent
	}
	mustExec(t, h, ctx, `INSERT INTO organizations (id, tenant_id, parent_org_id, name, created_by)
		VALUES ($1,$2,$3,$4,$5)`, id, tenant, parentArg, name, uuid.Nil)
	return id
}

func seedPermission(t *testing.T, h *testkit.DBHandle, ctx context.Context, key string) {
	t.Helper()
	mustExec(t, h, ctx, `INSERT INTO permissions (key, module, description) VALUES ($1,$2,$3)
		ON CONFLICT (key) DO NOTHING`, key, "requests", key)
}

func seedRole(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant *uuid.UUID, key string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var tArg any
	if tenant != nil {
		tArg = *tenant
	}
	mustExec(t, h, ctx, `INSERT INTO roles (id, tenant_id, key, name, created_by) VALUES ($1,$2,$3,$4,$5)`,
		id, tArg, key, key, uuid.Nil)
	return id
}

func seedRolePerm(t *testing.T, h *testkit.DBHandle, ctx context.Context, role uuid.UUID, perm string) {
	t.Helper()
	mustExec(t, h, ctx, `INSERT INTO role_permissions (role_id, permission_key) VALUES ($1,$2)`, role, perm)
}

// seedAssignment inserts an actor_assignment. validFrom/validTo are raw SQL
// expressions (e.g. "now()", "NULL") interpolated as trusted test literals.
func seedAssignment(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant, capID, role uuid.UUID,
	scopeKind string, scopeID *uuid.UUID, scopeType *string, validFrom, validTo string,
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
		(id, tenant_id, capacity_id, role_id, scope_kind, scope_id, scope_type, valid_from, valid_to, granted_by, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7, ` + validFrom + `, ` + validTo + `, $8, $9)`
	mustExec(t, h, ctx, sql, uuid.New(), tenant, capID, role, scopeKind, scopeIDArg, scopeTypeArg, uuid.Nil, uuid.Nil)
}

func mustExec(t *testing.T, h *testkit.DBHandle, ctx context.Context, sql string, args ...any) {
	t.Helper()
	if _, err := h.Admin.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("seed exec: %v\n%s", err, sql)
	}
}

func hasAll(got []string, want ...string) bool {
	set := map[string]bool{}
	for _, g := range got {
		set[g] = true
	}
	for _, w := range want {
		if !set[w] {
			return false
		}
	}
	return true
}

func hasAllUUID(got []uuid.UUID, want ...uuid.UUID) bool {
	set := map[uuid.UUID]bool{}
	for _, g := range got {
		set[g] = true
	}
	for _, w := range want {
		if !set[w] {
			return false
		}
	}
	return true
}
