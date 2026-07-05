package testkit

import (
	"context"
	"strings"
	"testing"
)

// badDSN is malformed so pgx's ParseConfig rejects it — used to drive the
// dsn-parse error branches of the pool/connection helpers.
const badDSN = "postgres://user:pass@host:notaport/db"

// TestConnectDBParseError covers connectDB's parse-error branch.
func TestConnectDBParseError(t *testing.T) {
	if _, err := connectDB(context.Background(), badDSN, ""); err == nil {
		t.Fatal("connectDB(badDSN) = nil error, want parse failure")
	}
}

// TestConnectDBWithExplicitDatabase covers connectDB's dbname!="" branch (the
// swap of the target database). Connecting to a non-existent database errors,
// but the branch under test is the database-name assignment, which runs first.
func TestConnectDBWithExplicitDatabase(t *testing.T) {
	dsn := adminDSN(t)
	if _, err := connectDB(context.Background(), dsn, "wt_definitely_not_a_real_db_xyz"); err == nil {
		t.Fatal("connectDB to a missing database = nil error, want connect failure")
	}
}

// TestBuildPoolParseError covers buildPool's parse-error branch.
func TestBuildPoolParseError(t *testing.T) {
	if _, err := newPoolDB(context.Background(), badDSN, "db", 2); err == nil {
		t.Fatal("newPoolDB(badDSN) = nil error, want parse failure")
	}
}

// TestBuildPoolPingFailure covers buildPool's ping-failure path (pool built, but
// the first ping fails) and pingWithRoleRetry's non-28000 "real error" return:
// a pool aimed at a non-existent database pings with SQLSTATE 3D000, which is
// not the provisioning race, so it returns immediately instead of retrying.
func TestBuildPoolPingFailure(t *testing.T) {
	dsn := adminDSN(t)
	if _, err := newPoolDB(context.Background(), dsn, "wt_definitely_not_a_real_db_xyz", 2); err == nil {
		t.Fatal("newPoolDB to a missing database = nil error, want ping failure")
	}
}

// TestAlterRoleWithRetryRealError covers alterRoleWithRetry's non-transient
// branch: a syntactically invalid statement fails with an error that is NOT the
// "tuple concurrently updated" race, so it returns without retrying.
func TestAlterRoleWithRetryRealError(t *testing.T) {
	dsn := adminDSN(t)
	ctx := context.Background()
	conn, err := connectDB(ctx, dsn, "")
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	err = alterRoleWithRetry(ctx, conn, "THIS IS NOT VALID SQL")
	if err == nil {
		t.Fatal("alterRoleWithRetry(invalid sql) = nil error, want syntax error")
	}
}

// TestDropTestDBConnectFailure covers dropTestDB's connect-failure early return:
// a bad DSN cannot connect, so the helper returns quietly (best-effort cleanup).
func TestDropTestDBConnectFailure(t *testing.T) {
	// Must not panic or block; simply returns after the failed connect.
	dropTestDB(context.Background(), badDSN, "wt_whatever")
}

// TestBuildTemplateConnectFailure covers buildTemplate's connect-error branch
// (the admin connect fails before any template work).
func TestBuildTemplateConnectFailure(t *testing.T) {
	_, err := buildTemplate(context.Background(), badDSN)
	if err == nil || !strings.Contains(err.Error(), "connect admin") {
		t.Fatalf("buildTemplate(badDSN) = %v, want a connect-admin error", err)
	}
}
