package testkit

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// countAdmin runs a scalar count query through the admin pool.
func countAdmin(t *testing.T, h *DBHandle, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := h.Admin.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("count query: %v\n%s", err, sql)
	}
	return n
}

// TestIntegrationFixtureBuilders exercises every catalog/tenant fixture builder
// and asserts each wrote the row it claims. Coverage aside, this is the harness
// self-test: a product repo relies on these fixtures behaving exactly so.
func TestIntegrationFixtureBuilders(t *testing.T) {
	h := NewDB(t)

	tn := CreateTenant(t, h)
	if n := countAdmin(t, h, `SELECT count(*) FROM tenants WHERE id=$1`, tn.ID); n != 1 {
		t.Fatalf("tenant rows = %d, want 1", n)
	}

	userID := CreateUser(t, h)
	if n := countAdmin(t, h, `SELECT count(*) FROM users WHERE id=$1`, userID); n != 1 {
		t.Fatalf("user rows = %d, want 1", n)
	}

	cap := CreateCapacity(t, h, tn.ID, userID)
	if n := countAdmin(t, h, `SELECT count(*) FROM acting_capacities WHERE id=$1 AND user_id=$2`, cap, userID); n != 1 {
		t.Fatalf("capacity rows = %d, want 1", n)
	}

	// Root org (no parent) then a child org (parent set) — covers both branches.
	root := CreateOrg(t, h, tn.ID, nil, "Root")
	child := CreateOrg(t, h, tn.ID, &root, "Child")
	var parent *uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT parent_org_id FROM organizations WHERE id=$1`, child).Scan(&parent); err != nil {
		t.Fatalf("load child org: %v", err)
	}
	if parent == nil || *parent != root {
		t.Fatalf("child parent_org_id = %v, want %v", parent, root)
	}

	// Permission (dotted key → module prefix via moduleOf) plus a role granting it.
	CreatePermission(t, h, "cov.thing.read", false)
	CreatePermission(t, h, "cov.thing.read", false) // ON CONFLICT DO NOTHING path
	if n := countAdmin(t, h, `SELECT count(*) FROM permissions WHERE key=$1 AND module='cov'`, "cov.thing.read"); n != 1 {
		t.Fatalf("permission rows = %d, want 1 (dedup)", n)
	}

	role := CreateRole(t, h, tn.ID, "cov.role", "cov.thing.read", "cov.thing.write")
	if n := countAdmin(t, h, `SELECT count(*) FROM role_permissions WHERE role_id=$1`, role); n != 2 {
		t.Fatalf("role_permissions = %d, want 2", n)
	}

	// GrantRole at tenant scope (nil scopeID, empty scopeType).
	GrantRole(t, h, tn.ID, cap, role, "tenant", nil, "")
	// GrantRole at org scope (scopeID set).
	GrantRole(t, h, tn.ID, cap, role, "org", &root, "")
	if n := countAdmin(t, h, `SELECT count(*) FROM actor_assignments WHERE capacity_id=$1 AND role_id=$2`, cap, role); n != 2 {
		t.Fatalf("assignments = %d, want 2", n)
	}

	// Resource type + resource (with org), and the combined convenience builder.
	CreateResourceType(t, h, "cov.widget")
	ref := CreateResource(t, h, tn.ID, "cov.widget", &root)
	if ref.Type != "cov.widget" || ref.ID == uuid.Nil {
		t.Fatalf("CreateResource ref = %+v", ref)
	}
	if n := countAdmin(t, h, `SELECT count(*) FROM resources WHERE id=$1 AND org_id=$2`, ref.ID, root); n != 1 {
		t.Fatalf("resource-with-org rows = %d, want 1", n)
	}
	// GrantRole at resource_type and resource scope to exercise those branches.
	GrantRole(t, h, tn.ID, cap, role, "resource_type", nil, "cov.widget")
	GrantRole(t, h, tn.ID, cap, role, "resource", &ref.ID, "cov.widget")

	combined := CreateResourceTypeAndResource(t, h, tn.ID, "cov.gadget")
	if combined.Type != "cov.gadget" {
		t.Fatalf("combined ref type = %q", combined.Type)
	}
	if n := countAdmin(t, h, `SELECT count(*) FROM resources WHERE id=$1`, combined.ID); n != 1 {
		t.Fatalf("combined resource rows = %d, want 1", n)
	}
}

// TestModuleOf covers the pure prefix helper on both the dotted and no-dot paths.
func TestModuleOf(t *testing.T) {
	if got := moduleOf("requests.request.approve"); got != "requests" {
		t.Fatalf("moduleOf dotted = %q, want requests", got)
	}
	if got := moduleOf("nodot"); got != "nodot" {
		t.Fatalf("moduleOf no-dot = %q, want nodot", got)
	}
}
