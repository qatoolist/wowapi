package cli

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// requireDSN resolves the admin DSN for DB-backed CLI tests and ensures the
// command under test reads the same value from DATABASE_URL. It mirrors testkit's
// policy: skip locally when no DSN is configured, but FAIL when WOWAPI_REQUIRE_DB
// is set (the CI/release gate) so DB-backed tests can never silently vanish.
func requireDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("WOWAPI_TEST_DSN")
	}
	if dsn == "" {
		if os.Getenv("WOWAPI_REQUIRE_DB") != "" {
			t.Fatal("WOWAPI_REQUIRE_DB is set but neither DATABASE_URL nor WOWAPI_TEST_DSN is available")
		}
		t.Skip("no DATABASE_URL/WOWAPI_TEST_DSN configured; skipping DB-backed CLI test")
	}
	// The apikey/audit/dlq commands read DATABASE_URL directly; keep it aligned.
	t.Setenv("DATABASE_URL", dsn)
	return dsn
}

// adminPool opens a superuser pool (the DSN's own login) for test fixtures and
// teardown. It is closed automatically at the end of the test.
func adminPool(t *testing.T, dsn string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open admin pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// execAdmin runs a statement on the admin pool, failing the test on error.
func execAdmin(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("admin exec failed: %v\nsql: %s", err, sql)
	}
}

// cleanupTenant removes every row a DB-backed test could have written for a
// throwaway tenant so the shared database is left as it was found.
func cleanupTenant(t *testing.T, pool *pgxpool.Pool, tenant uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	for _, sql := range []string{
		`DELETE FROM api_keys WHERE tenant_id = $1`,
		`DELETE FROM audit_logs WHERE tenant_id = $1`,
		`DELETE FROM audit_chain WHERE tenant_id = $1`,
	} {
		if _, err := pool.Exec(ctx, sql, tenant); err != nil {
			t.Logf("cleanup %s: %v", sql, err)
		}
	}
}
