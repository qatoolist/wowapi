package database

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// versionTablePrefix namespaces each migration source's goose history table.
const versionTablePrefix = "goose_version_"

// MigrateResult reports what a Migrate call did.
type MigrateResult struct {
	// Version is the source's highest applied migration version afterwards.
	Version int64
	// Applied counts migrations run by THIS call — 0 means the source was
	// already up to date (the idempotent rerun case).
	Applied int
}

// Migrate applies every pending migration from src (an fs.FS of goose
// NNNNN_name.sql files, e.g. migrations.Kernel()) under a history table
// dedicated to source. Because each source has its own history table,
// independently-numbered sources (kernel "wowapi", each product module)
// coexist without version collisions (blueprint 03 §5; review finding
// ARCH-16). Reruns are no-ops (Applied == 0) — migration idempotency is a
// Phase 2 acceptance criterion.
//
// The pool should carry migration-owner credentials (app_migrate /
// config.DB.MigrateDSN); runtime processes never hold them (12 §7).
func Migrate(ctx context.Context, pool *pgxpool.Pool, src fs.FS, source string) (MigrateResult, error) {
	if source == "" {
		return MigrateResult{}, fmt.Errorf("database: migrate source name is required")
	}
	db := stdlib.OpenDBFromPool(pool)
	// Closing the *sql.DB releases its adapter conns back to the pool; the
	// pool itself stays open for the caller.
	defer func() { _ = db.Close() }()

	p, err := goose.NewProvider(goose.DialectPostgres, db, src,
		goose.WithTableName(versionTablePrefix+source))
	if err != nil {
		return MigrateResult{}, fmt.Errorf("database: migration provider (%s): %w", source, err)
	}
	applied, err := p.Up(ctx)
	if err != nil {
		return MigrateResult{}, fmt.Errorf("database: migrate up (%s): %w", source, err)
	}
	v, err := p.GetDBVersion(ctx)
	if err != nil {
		return MigrateResult{}, fmt.Errorf("database: read migration version (%s): %w", source, err)
	}
	return MigrateResult{Version: v, Applied: len(applied)}, nil
}
