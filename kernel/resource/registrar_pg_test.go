package resource_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationRegistrarUpsertBumpsVersion drives the Postgres registrar: a
// first Upsert inserts the mirror row at version 1; a second Upsert for the same
// id updates label/status/org and bumps the version to 2.
func TestIntegrationRegistrarUpsertBumpsVersion(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()

	// Seed the FK spine: tenant, resource type, org.
	rt := "requests.request"
	org := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO tenants (id, slug, display_name, created_by) VALUES ($1,$2,$3,$4)`,
		tenant, "t-"+shortHex(), "Tenant", uuid.Nil)
	mustExec(t, h, ctx, `INSERT INTO resource_types (key, module, description) VALUES ($1,$2,$3)`,
		rt, "requests", "request")
	mustExec(t, h, ctx, `INSERT INTO organizations (id, tenant_id, name, created_by) VALUES ($1,$2,$3,$4)`,
		org, tenant, "Org", uuid.Nil)

	reg := resource.NewRegistrar()
	ref := resource.Ref{Type: rt, ID: uuid.New()}
	tctx := database.WithTenantID(ctx, tenant)

	// First upsert: insert.
	if err := h.TxM.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		return reg.Bind(db).Upsert(ctx, ref, &org, "First", "active")
	}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	label, status, version := readResource(t, h, ctx, ref.ID)
	if label != "First" || status != "active" || version != 1 {
		t.Fatalf("after insert: label=%q status=%q version=%d, want First/active/1", label, status, version)
	}

	// Second upsert: update + version bump.
	if err := h.TxM.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		return reg.Bind(db).Upsert(ctx, ref, nil, "Second", "locked")
	}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	label, status, version = readResource(t, h, ctx, ref.ID)
	if label != "Second" || status != "locked" || version != 2 {
		t.Fatalf("after re-upsert: label=%q status=%q version=%d, want Second/locked/2", label, status, version)
	}
}

func readResource(t *testing.T, h *testkit.DBHandle, ctx context.Context, id uuid.UUID) (label, status string, version int) {
	t.Helper()
	if err := h.Admin.QueryRow(ctx,
		`SELECT label, status, version FROM resources WHERE id = $1`, id).Scan(&label, &status, &version); err != nil {
		t.Fatalf("read resource: %v", err)
	}
	return
}

func mustExec(t *testing.T, h *testkit.DBHandle, ctx context.Context, sql string, args ...any) {
	t.Helper()
	if _, err := h.Admin.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("seed exec: %v\n%s", err, sql)
	}
}

func shortHex() string { return uuid.New().String()[:8] }
