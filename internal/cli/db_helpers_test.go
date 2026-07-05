package cli

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/testkit"
)

// requireDSN gives a DB-backed CLI test a migrated database and points the
// command under test at it via DATABASE_URL.
//
// The apikey/audit/dlq commands connect to DATABASE_URL directly and need a
// migrated schema. The CI compose database is pristine — only testkit's template
// clones are migrated — so instead of the raw base DSN we provision an exclusive,
// migrated clone with testkit.NewDB (dropped automatically at the end of the
// test) and rewrite DATABASE_URL to point at it. This keeps the base database
// untouched and mirrors how the rest of the suite acquires a database.
//
// It mirrors testkit's policy: skip locally when no DSN is configured, but FAIL
// when WOWAPI_REQUIRE_DB is set (the CI/release gate) so DB-backed tests can
// never silently vanish.
func requireDSN(t *testing.T) string {
	t.Helper()
	base := os.Getenv("DATABASE_URL")
	if base == "" {
		base = os.Getenv("WOWAPI_TEST_DSN")
	}
	if base == "" {
		if os.Getenv("WOWAPI_REQUIRE_DB") != "" {
			t.Fatal("WOWAPI_REQUIRE_DB is set but neither DATABASE_URL nor WOWAPI_TEST_DSN is available")
		}
		t.Skip("no DATABASE_URL/WOWAPI_TEST_DSN configured; skipping DB-backed CLI test")
	}

	// Migrated, exclusive clone (uses the base DSN to provision, dropped on cleanup).
	h := testkit.NewDB(t)
	u, err := url.Parse(base)
	if err != nil {
		t.Fatalf("parse base DSN: %v", err)
	}
	u.Path = "/" + h.Name
	dsn := u.String()

	// The apikey/audit/dlq commands read DATABASE_URL directly; point them at the clone.
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
// throwaway tenant. The clone database is dropped at test end regardless, so
// this is belt-and-suspenders for any test that shares a database.
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
