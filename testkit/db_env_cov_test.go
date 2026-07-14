package testkit

import (
	"context"
	"os"
	"testing"
)

// TestAdminDSNFallbackToDatabaseURL covers adminDSN's DATABASE_URL fallback: when
// WOWAPI_TEST_DSN is unset but DATABASE_URL is present, the latter is returned.
func TestAdminDSNFallbackToDatabaseURL(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		if RequireDB() {
			t.Fatal("WOWAPI_REQUIRE_DB is set but DATABASE_URL is empty; the fallback branch must run in this gate")
		}
		t.Skip("no DATABASE_URL to exercise the fallback branch")
	}
	// Clear the primary var for this test only (restored automatically). The
	// helper must then fall through to DATABASE_URL.
	t.Setenv("WOWAPI_TEST_DSN", "")
	if got := adminDSN(t); got != os.Getenv("DATABASE_URL") {
		t.Fatalf("adminDSN fallback = %q, want DATABASE_URL", got)
	}
}

// TestBuildPoolConfigRejected covers buildPool's pool-construction error branch:
// a valid DSN but an out-of-range MaxConns (0) is accepted by ParseConfig yet
// rejected by pgxpool.NewWithConfig.
func TestBuildPoolConfigRejected(t *testing.T) {
	dsn := adminDSN(t)
	if _, err := newPoolDB(context.Background(), dsn, "postgres", 0); err == nil {
		t.Fatal("newPoolDB with MaxConns=0 = nil error, want NewWithConfig rejection")
	}
}
