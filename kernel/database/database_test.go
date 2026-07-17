package database_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
)

func TestTenantContextRoundtrip(t *testing.T) {
	id := uuid.New()
	ctx := database.WithTenantID(context.Background(), id)
	got, ok := database.TenantIDFrom(ctx)
	if !ok || got != id {
		t.Fatalf("TenantIDFrom = %v/%v", got, ok)
	}
	if _, ok := database.TenantIDFrom(context.Background()); ok {
		t.Fatal("empty context must not carry a tenant")
	}

	aid := uuid.New()
	ctx = database.WithActorID(ctx, aid)
	gotA, ok := database.ActorIDFrom(ctx)
	if !ok || gotA != aid {
		t.Fatalf("ActorIDFrom = %v/%v", gotA, ok)
	}
}

// The tenant check must fail closed BEFORE any connection is touched: a
// Manager with a nil pool proves no database work happens without a tenant.
func TestWithTenantFailsClosedWithoutTenant(t *testing.T) {
	m := database.NewManager(nil, config.Defaults().DB)
	err := m.WithTenant(context.Background(), func(ctx context.Context, db database.TenantDB) error {
		t.Fatal("fn must not run")
		return nil
	})
	if !errors.Is(err, database.ErrNoTenantContext) {
		t.Fatalf("err = %v, want ErrNoTenantContext", err)
	}
	err = m.WithTenantRO(context.Background(), func(ctx context.Context, db database.TenantDB) error {
		t.Fatal("fn must not run")
		return nil
	})
	if !errors.Is(err, database.ErrNoTenantContext) {
		t.Fatalf("RO err = %v, want ErrNoTenantContext", err)
	}
}

func TestExpectOneRow(t *testing.T) {
	if err := database.ExpectOneRow(pgconn.NewCommandTag("UPDATE 1"), "thing"); err != nil {
		t.Fatalf("1 row: %v", err)
	}
	err := database.ExpectOneRow(pgconn.NewCommandTag("UPDATE 0"), "thing")
	if !errors.Is(err, database.ErrVersionConflict) {
		t.Fatalf("0 rows: %v, want ErrVersionConflict", err)
	}
	if err.Error() != "thing: database: version conflict" {
		t.Fatalf("message = %q", err.Error())
	}
}

func TestNewPoolRejectsBadDSNWithoutEchoingIt(t *testing.T) {
	const dsn = "postgres://user:hunter2@[bad/db" // malformed on purpose
	_, err := database.NewPool(context.Background(), dsn, config.Defaults().DB)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if got := err.Error(); strings.Contains(got, "hunter2") || strings.Contains(got, dsn) {
		t.Fatalf("pool error echoed DSN credentials: %q", got)
	}
}
