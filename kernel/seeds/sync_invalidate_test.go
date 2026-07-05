package seeds_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationSyncInvalidatesAuthzCache proves the CA-2 wiring: a seed sync
// that changes a platform role's permission set invalidates the in-process authz
// cache, so the change is reflected IMMEDIATELY (not after the TTL). The control
// arm — the same re-sync WITHOUT passing the invalidator — stays stale within the
// TTL, so the assertion isolates the invalidation (not the clock) as the cause.
//
// The cache holds ActiveAssignments, which pre-join role_permissions; the store
// runs on the caller's tenant tx, and the platform role is assigned to a
// capacity, so the cached entry carries the role's granted permission.
func TestIntegrationSyncInvalidatesAuthzCache(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	tenant := uuid.New()
	seedTenant(t, h, tenant)
	cap1 := seedCapacity(t, h, tenant)

	// A platform role (tenant_id NULL) granting one permission, synced as
	// app_platform — the real seed posture. The permission must exist for the
	// role_permissions FK, so the bundle declares it too.
	bundleWith := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{{Key: "core.thing.read"}, {Key: "core.thing.write"}},
		Roles:       []seeds.RoleSeed{{Key: "core.editor", Name: "Editor", Permissions: []string{"core.thing.read", "core.thing.write"}}},
	}
	if err := seeds.Sync(ctx, h.Platform, bundleWith); err != nil {
		t.Fatalf("initial seed sync: %v", err)
	}

	// Assign the freshly-synced platform role to the capacity at tenant scope so
	// ActiveAssignments returns it (with role_permissions pre-joined).
	roleID := roleIDByKey(t, h, "core.editor")
	assignRole(t, h, tenant, cap1, roleID)

	// A long TTL means only invalidation — never expiry — can refresh within the
	// test window: if the cache reloads, it is because InvalidateAll fired.
	cache := authz.NewCachingStore(authz.NewStore(), time.Hour)
	actor := authz.Actor{Kind: authz.ActorUser, CapacityID: cap1, TenantID: tenant}

	perms := func() []string {
		asgs := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Assignment, error) {
			return cache.ActiveAssignments(ctx, db, actor, time.Now())
		})
		if len(asgs) != 1 {
			t.Fatalf("want exactly one assignment for the capacity, got %d", len(asgs))
		}
		return asgs[0].Perms
	}

	// Warm the cache: the role grants write.
	if !contains(perms(), "core.thing.write") {
		t.Fatal("precondition: role must grant core.thing.write before the prune")
	}

	// Re-sync with the write grant pruned. First WITHOUT an invalidator: within
	// the TTL the cache still serves the stale grant (bounded staleness — the
	// documented trade-off, and proof the value really is cached).
	bundlePruned := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{{Key: "core.thing.read"}, {Key: "core.thing.write"}},
		Roles:       []seeds.RoleSeed{{Key: "core.editor", Name: "Editor", Permissions: []string{"core.thing.read"}}},
	}
	if err := seeds.Sync(ctx, h.Platform, bundlePruned); err != nil {
		t.Fatalf("prune seed sync (no invalidator): %v", err)
	}
	if !contains(perms(), "core.thing.write") {
		t.Fatal("without an invalidator a spine change must be bounded-stale within the TTL (still cached)")
	}

	// Now re-sync passing the live cache as the invalidator. The write is already
	// pruned in the DB; InvalidateAll drops the stale entry so the next read
	// reloads and no longer sees write — immediately, with no clock advance.
	if err := seeds.Sync(ctx, h.Platform, bundlePruned, cache); err != nil {
		t.Fatalf("prune seed sync (with invalidator): %v", err)
	}
	if contains(perms(), "core.thing.write") {
		t.Fatal("after a spine sync with the cache invalidator, the pruned grant must be gone immediately (stale-allow!)")
	}
	if !contains(perms(), "core.thing.read") {
		t.Fatal("the surviving grant (read) must still resolve after invalidation")
	}
}

// TestIntegrationSyncCachingOffUnaffected proves the caching-off default is
// untouched: Sync with no invalidator succeeds and a plain PgStore reflects the
// pruned grant on the next read (nothing to invalidate, nothing stale).
func TestIntegrationSyncCachingOffUnaffected(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	tenant := uuid.New()
	seedTenant(t, h, tenant)
	cap1 := seedCapacity(t, h, tenant)

	bundleWith := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{{Key: "core.thing.read"}, {Key: "core.thing.write"}},
		Roles:       []seeds.RoleSeed{{Key: "core.editor", Name: "Editor", Permissions: []string{"core.thing.read", "core.thing.write"}}},
	}
	if err := seeds.Sync(ctx, h.Platform, bundleWith); err != nil {
		t.Fatalf("initial seed sync: %v", err)
	}
	roleID := roleIDByKey(t, h, "core.editor")
	assignRole(t, h, tenant, cap1, roleID)

	store := authz.NewStore() // no cache — the default posture
	actor := authz.Actor{Kind: authz.ActorUser, CapacityID: cap1, TenantID: tenant}
	perms := func() []string {
		asgs := inTx(t, h, tenant, func(ctx context.Context, db database.TenantDB) ([]authz.Assignment, error) {
			return store.ActiveAssignments(ctx, db, actor, time.Now())
		})
		if len(asgs) != 1 {
			t.Fatalf("want one assignment, got %d", len(asgs))
		}
		return asgs[0].Perms
	}
	if !contains(perms(), "core.thing.write") {
		t.Fatal("precondition: role grants write")
	}

	bundlePruned := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{{Key: "core.thing.read"}, {Key: "core.thing.write"}},
		Roles:       []seeds.RoleSeed{{Key: "core.editor", Name: "Editor", Permissions: []string{"core.thing.read"}}},
	}
	if err := seeds.Sync(ctx, h.Platform, bundlePruned); err != nil {
		t.Fatalf("prune seed sync (caching off): %v", err)
	}
	if contains(perms(), "core.thing.write") {
		t.Fatal("with caching off a pruned grant must be gone on the next read")
	}
}

// --- helpers (Admin/Platform pools bypass RLS as owner / act as app_platform) ---

func seedTenant(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO tenants (id, slug, display_name, created_by) VALUES ($1,$2,$3,$4)`,
		tenant, "t-"+uuid.New().String()[:8], "Tenant", uuid.Nil); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}
}

func seedCapacity(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO users (id, idp_subject, email, created_by) VALUES ($1,$2,$3,$4)`,
		userID, "idp-"+uuid.New().String()[:8], uuid.New().String()[:8]+"@example.test", uuid.Nil); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	capID := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO acting_capacities (id, tenant_id, user_id, label, created_by) VALUES ($1,$2,$3,$4,$5)`,
		capID, tenant, userID, "member", uuid.Nil); err != nil {
		t.Fatalf("seed capacity: %v", err)
	}
	return capID
}

func roleIDByKey(t *testing.T, h *testkit.DBHandle, key string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT id FROM roles WHERE key = $1 AND tenant_id IS NULL`, key).Scan(&id); err != nil {
		t.Fatalf("lookup role %s: %v", key, err)
	}
	return id
}

func assignRole(t *testing.T, h *testkit.DBHandle, tenant, capID, roleID uuid.UUID) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO actor_assignments
			(id, tenant_id, capacity_id, role_id, scope_kind, scope_id, scope_type, granted_by, created_by, valid_from)
			VALUES ($1,$2,$3,$4,'tenant',NULL,NULL,$5,$6, now() - interval '1 minute')`,
		uuid.New(), tenant, capID, roleID, uuid.Nil, uuid.Nil); err != nil {
		t.Fatalf("assign role: %v", err)
	}
}

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

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
