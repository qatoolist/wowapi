package authz_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationNoSelfGrantViaAssignments proves the RBAC self-grant backstop
// (SEC-13): a module running as the runtime app_rt role — which is what module
// SQL executes as — cannot INSERT an actor_assignment to grant itself a role.
// The whole authz spine is SELECT-only to app_rt (writes are app_platform).
func TestIntegrationNoSelfGrantViaAssignments(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()
	seedTenant(t, h, context.Background(), tenant)
	ctx := database.WithTenantID(context.Background(), tenant)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx, `INSERT INTO actor_assignments
			(id, tenant_id, capacity_id, role_id, scope_kind, granted_by, created_by)
			VALUES ($1, app_tenant_id(), $2, $3, 'tenant', $4, $5)`,
			uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
		return e
	})
	if err == nil {
		t.Fatal("app_rt must NOT be able to write actor_assignments (self-grant escalation)")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected a permission-denied error, got: %v", err)
	}
}

// TestIntegrationNoSelfGrantViaRelationships proves the ReBAC self-grant
// backstop (SEC-24): app_rt cannot INSERT a relationship edge, so a module
// cannot forge a granted_via edge naming its own capacity to grant itself a
// permission on a resource. Edge writes are app_platform-only.
func TestIntegrationNoSelfGrantViaRelationships(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()
	seedTenant(t, h, context.Background(), tenant)
	seedResourceType(t, h, context.Background(), "requests.request")
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('core.owner_of','core','capacity','resource','owner')`); err != nil {
		t.Fatal(err)
	}
	ctx := database.WithTenantID(context.Background(), tenant)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := db.Exec(ctx, `INSERT INTO relationships
			(id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, valid_from, version, created_by)
			VALUES ($1, app_tenant_id(), 'core.owner_of', 'capacity', $2, 'resource', $3, now(), 1, $4)`,
			uuid.New(), uuid.New(), uuid.New(), uuid.New())
		return e
	})
	if err == nil {
		t.Fatal("app_rt must NOT be able to write relationships (ReBAC self-grant escalation)")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected a permission-denied error, got: %v", err)
	}
}

// TestIntegrationRuntimeRoleNotMemberOfPlatform is the CF-1 backstop: the runtime
// role app_rt must NOT be a member of app_platform. Role membership is
// cluster-global, so if some dev script ran `GRANT app_platform TO app_rt` on this
// cluster, app_rt would inherit app_platform's writes on the authorization spine
// in EVERY database (incl. testkit clones) and the self-grant guards above would
// silently pass through. This turns that environment poisoning into a red test.
func TestIntegrationRuntimeRoleNotMemberOfPlatform(t *testing.T) {
	h := testkit.NewDB(t)
	var isMember bool
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT pg_has_role('app_rt', 'app_platform', 'MEMBER')`).Scan(&isMember); err != nil {
		t.Fatal(err)
	}
	if isMember {
		t.Fatal("app_rt is a member of app_platform — runtime/platform privilege separation is broken " +
			"(CF-1). Some script likely ran `GRANT app_platform TO app_rt` on this cluster; run " +
			"`REVOKE app_platform FROM app_rt;` as the superuser. Role membership is cluster-global.")
	}
}

// TestIntegrationScopeCheckConstraints proves the DB CHECKs (SEC-26/SEC-29): a
// resource_type-scoped assignment with a NULL scope_type, or an org/resource
// scope with a NULL scope_id, is rejected — closing the covers() over-grant at
// the source. Seeded via the Admin pool (which otherwise bypasses grants) so it
// is the CHECK, not a privilege error, that rejects.
func TestIntegrationScopeCheckConstraints(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	role := seedRole(t, h, ctx, &tenant, "r")

	bad := []struct {
		name      string
		scopeKind string
		scopeID   any
		scopeType any
	}{
		{"resource_type without type", "resource_type", nil, nil},
		{"org without id", "org", nil, nil},
		{"resource without id", "resource", nil, "requests.request"},
	}
	for _, c := range bad {
		t.Run(c.name, func(t *testing.T) {
			_, err := h.Admin.Exec(ctx, `INSERT INTO actor_assignments
				(id, tenant_id, capacity_id, role_id, scope_kind, scope_id, scope_type, granted_by, created_by)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
				uuid.New(), tenant, uuid.New(), role, c.scopeKind, c.scopeID, c.scopeType, uuid.Nil, uuid.Nil)
			if err == nil {
				t.Fatalf("scope %q with missing id/type must violate a CHECK constraint", c.scopeKind)
			}
		})
	}
}
