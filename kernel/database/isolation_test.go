package database_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/testkit"
)

func TestIntegration_TenantIsolation_Adversarial(t *testing.T) {
	h := testkit.NewDB(t)
	// Seed valid resource type for FK
	_, err := h.Admin.Exec(context.Background(), `INSERT INTO resource_types (key, module, description) VALUES ('test-resource', 'kernel', 'test') ON CONFLICT DO NOTHING`)
	if err != nil {
		t.Fatalf("failed to seed resource_types: %v", err)
	}

	// 1. Setup two distinct tenants
	tenantA := testkit.CreateTenant(t, h)
	tenantB := testkit.CreateTenant(t, h)

	// 2. Provision data for Tenant B
	// We use the Tenant B handle to insert data so it is correctly scoped.
	ctxB := testkit.TenantCtx(tenantB.ID)
	resourceID := uuid.New()
	err = h.TxM.WithTenant(ctxB, func(ctx context.Context, db database.TenantDB) error {
		_, err := db.Exec(ctx, `INSERT INTO resources (id, tenant_id, resource_type, label, created_by) VALUES ($1, $2, 'test-resource', 'secret-data', $3)`,
			resourceID, tenantB.ID, uuid.New())
		return err
	})
	if err != nil {
		t.Fatalf("failed to seed tenant B: %v", err)
	}

	// 3. Attempt access as Tenant A (The Adversarial Vector)
	// We simulate an attacker attempting to access resourceID by switching the context to Tenant A.
	ctxA := testkit.TenantCtx(tenantA.ID)

	err = h.TxM.WithTenantRO(ctxA, func(ctx context.Context, db database.TenantDB) error {
		var label string
		// This query intentionally targets a record that exists in the DB but belongs to Tenant B.
		err := db.QueryRow(ctx, `SELECT label FROM resources WHERE id = $1`, resourceID).Scan(&label)

		if err == nil {
			t.Errorf("SECURITY BREACH: Tenant A accessed Tenant B data! Label: %s", label)
		}
		return nil
	})
	if err != nil {
		t.Logf("Expected access failure (as intended): %v", err)
	}
}
