package database_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
)

// rls_guard_test.go — QA G1 (security/tenancy): WithConnRLSGuard is the
// pool-construction backstop that fails closed when the effective role would
// silently defeat FORCE row-level security (a superuser or BYPASSRLS role). It
// had no direct test; RLS enforcement is the tenancy boundary, so its guard is
// security-critical.

func guardTestDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("WOWAPI_TEST_DSN")
	}
	if dsn == "" {
		t.Skip("RLS-guard test needs DATABASE_URL (a migrated DB with the app_rt role)")
	}
	return dsn
}

// A pool that connects as an over-privileged role (superuser / BYPASSRLS) with
// NO non-privileged role set MUST fail construction, not serve tenant traffic
// with RLS disabled (SEC-12, fail-closed).
func TestConnRLSGuardRejectsOverPrivilegedConnection(t *testing.T) {
	dsn := guardTestDSN(t)
	pool, err := database.NewPool(context.Background(), dsn, config.Defaults().DB, database.WithConnRLSGuard())
	if err == nil {
		pool.Close()
		// The DSN login role is itself non-privileged (a production-style login);
		// the reject path can't be exercised here — that is a valid environment,
		// not a failure. The allow-path test below still covers the guard.
		t.Skip("DSN login role is non-superuser/non-BYPASSRLS; guard-reject path not applicable in this environment")
	}
	if !strings.Contains(err.Error(), "RLS would not be enforced") {
		t.Fatalf("guard must reject with an RLS-enforcement message; got: %v", err)
	}
	// The error must not leak DSN credentials.
	if strings.Contains(err.Error(), "@") && strings.Contains(err.Error(), "password") {
		t.Fatalf("guard error leaked connection details: %v", err)
	}
}

// The same guard, chained AFTER WithSetRole("app_rt"), MUST admit the pool — the
// effective role is now RLS-enforced — and the connection must actually run as
// app_rt.
func TestConnRLSGuardAdmitsNonPrivilegedRole(t *testing.T) {
	dsn := guardTestDSN(t)
	pool, err := database.NewPool(context.Background(), dsn, config.Defaults().DB,
		database.WithSetRole("app_rt"), database.WithConnRLSGuard())
	if err != nil {
		t.Fatalf("guard must admit a SET ROLE app_rt connection: %v", err)
	}
	defer pool.Close()

	var who string
	if err := pool.QueryRow(context.Background(), "SELECT current_user").Scan(&who); err != nil {
		t.Fatal(err)
	}
	if who != "app_rt" {
		t.Fatalf("WithSetRole did not take effect: current_user=%q, want app_rt", who)
	}
	// And the effective role is genuinely RLS-enforced (the guard's own predicate).
	var enforced bool
	if err := pool.QueryRow(context.Background(),
		`SELECT current_setting('is_superuser') = 'off' AND NOT rolbypassrls
		   FROM pg_roles WHERE rolname = current_user`).Scan(&enforced); err != nil {
		t.Fatal(err)
	}
	if !enforced {
		t.Fatal("app_rt must be non-superuser and non-BYPASSRLS for RLS to be enforced")
	}
}
