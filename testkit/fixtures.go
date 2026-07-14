package testkit

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// Fixtures create catalog and tenant rows for tests. Catalog rows (tenants,
// users, permissions, roles) and tenant spine rows are seeded through the Admin
// pool (superuser login, bypassing RLS/grants), matching how a platform service
// or migration would provision them. Runtime assertions then run through
// h.TxM (app_rt, RLS-enforced) — the production path.

// TenantHandle is a created tenant plus convenient ids.
type TenantHandle struct {
	ID uuid.UUID
}

// CreateTenant inserts a tenant and returns its handle.
func CreateTenant(t testing.TB, h *DBHandle) TenantHandle {
	t.Helper()
	id := uuid.New()
	execAdmin(t, h, `INSERT INTO tenants (id, slug, display_name, created_by) VALUES ($1,$2,$3,$4)`,
		id, "t-"+uuid.New().String()[:8], "Tenant", uuid.Nil)
	return TenantHandle{ID: id}
}

// CreateUser inserts a global user and returns its id.
func CreateUser(t *testing.T, h *DBHandle) uuid.UUID {
	t.Helper()
	id := uuid.New()
	execAdmin(t, h, `INSERT INTO users (id, idp_subject, email, created_by) VALUES ($1,$2,$3,$4)`,
		id, "idp-"+uuid.New().String()[:8], uuid.New().String()[:8]+"@example.test", uuid.Nil)
	return id
}

// CreateCapacity inserts an active acting capacity for a user in a tenant.
func CreateCapacity(t *testing.T, h *DBHandle, tenant, userID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	execAdmin(t, h, `INSERT INTO acting_capacities (id, tenant_id, user_id, label, created_by)
		VALUES ($1,$2,$3,$4,$5)`, id, tenant, userID, "member", uuid.Nil)
	return id
}

// CreateOrg inserts an organization (optionally under a parent) and returns its id.
func CreateOrg(t *testing.T, h *DBHandle, tenant uuid.UUID, parent *uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var parentArg any
	if parent != nil {
		parentArg = *parent
	}
	execAdmin(t, h, `INSERT INTO organizations (id, tenant_id, parent_org_id, name, created_by)
		VALUES ($1,$2,$3,$4,$5)`, id, tenant, parentArg, name, uuid.Nil)
	return id
}

// CreatePermission inserts a permission into the global catalog.
func CreatePermission(t *testing.T, h *DBHandle, key string, sensitive bool) {
	t.Helper()
	execAdmin(t, h, `INSERT INTO permissions (key, module, description, sensitive) VALUES ($1,$2,$3,$4)
		ON CONFLICT (key) DO NOTHING`, key, moduleOf(key), key, sensitive)
}

// CreateRole inserts a tenant role granting the given permissions and returns id.
func CreateRole(t *testing.T, h *DBHandle, tenant uuid.UUID, key string, perms ...string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	execAdmin(t, h, `INSERT INTO roles (id, tenant_id, key, name, created_by) VALUES ($1,$2,$3,$4,$5)`,
		id, tenant, key, key, uuid.Nil)
	for _, p := range perms {
		CreatePermission(t, h, p, false)
		execAdmin(t, h, `INSERT INTO role_permissions (role_id, permission_key) VALUES ($1,$2)
			ON CONFLICT DO NOTHING`, id, p)
	}
	return id
}

// GrantRole assigns a role to a capacity at the given scope (tenant/org/
// resource_type/resource). scopeID/scopeType may be zero/empty per scope.
func GrantRole(t *testing.T, h *DBHandle, tenant, capacity, role uuid.UUID, scopeKind string, scopeID *uuid.UUID, scopeType string) {
	t.Helper()
	var scopeIDArg, scopeTypeArg any
	if scopeID != nil {
		scopeIDArg = *scopeID
	}
	if scopeType != "" {
		scopeTypeArg = scopeType
	}
	// valid_from is backdated a minute so an assignment is immediately effective:
	// the DB clock (DEFAULT now()) can run slightly ahead of the host clock the
	// authz evaluator uses for time.Now(), and a grant seeded at exactly now()
	// could otherwise be excluded by a tight grant→evaluate window (flaky 403).
	execAdmin(t, h, `INSERT INTO actor_assignments
		(id, tenant_id, capacity_id, role_id, scope_kind, scope_id, scope_type, granted_by, created_by, valid_from)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now() - interval '1 minute')`,
		uuid.New(), tenant, capacity, role, scopeKind, scopeIDArg, scopeTypeArg, uuid.Nil, uuid.Nil)
}

// CreateResourceType registers a resource type in the global catalog.
func CreateResourceType(t *testing.T, h *DBHandle, key string) {
	t.Helper()
	execAdmin(t, h, `INSERT INTO resource_types (key, module, description) VALUES ($1,$2,$3)
		ON CONFLICT (key) DO NOTHING`, key, moduleOf(key), key)
}

// CreateResource inserts a kernel resources mirror row (via Admin) and returns
// its Ref. resType must be a registered resource type.
func CreateResource(t *testing.T, h *DBHandle, tenant uuid.UUID, resType string, org *uuid.UUID) resource.Ref {
	t.Helper()
	id := uuid.New()
	var orgArg any
	if org != nil {
		orgArg = *org
	}
	execAdmin(t, h, `INSERT INTO resources (id, tenant_id, resource_type, org_id, label, status, created_by)
		VALUES ($1,$2,$3,$4,'r','active',$5)`, id, tenant, resType, orgArg, uuid.Nil)
	return resource.Ref{Type: resType, ID: id}
}

// CreateResourceTypeAndResource registers a resource type (if needed) and
// inserts a resources mirror row, returning its Ref — convenient for tests that
// need a resource-scoped target or aggregate.
func CreateResourceTypeAndResource(t *testing.T, h *DBHandle, tenant uuid.UUID, resType string) resource.Ref {
	t.Helper()
	CreateResourceType(t, h, resType)
	return CreateResource(t, h, tenant, resType, nil)
}

// TenantCtx returns a context scoped to the tenant, for driving h.TxM.
func TenantCtx(tenant uuid.UUID) context.Context {
	return database.WithTenantID(context.Background(), tenant)
}

func execAdmin(t testing.TB, h *DBHandle, sql string, args ...any) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("fixture exec: %v\n%s", err, sql)
	}
}

func moduleOf(key string) string {
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			return key[:i]
		}
	}
	return key
}
