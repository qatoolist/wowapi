package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// CompatibilityFlag is an observable schema-version posture flag. It is stored
// in migration.compat_flag so operators and application code can read the
// current posture without relying on process-local state.
type CompatibilityFlag struct {
	Key       string
	Version   string
	UpdatedAt time.Time
}

// EnsureCompatFlagTable creates the migration.compat_flag table.
func EnsureCompatFlagTable(ctx context.Context, conn *pgx.Conn) error {
	if _, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS migration"); err != nil {
		return err
	}
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migration.compat_flag (
			key text PRIMARY KEY,
			version text NOT NULL,
			updated_at timestamptz NOT NULL DEFAULT now()
		)
	`)
	return err
}

// SetCompatibility updates the observable compatibility flag to version.
func SetCompatibility(ctx context.Context, conn *pgx.Conn, key, version string) error {
	if version == "" {
		return fmt.Errorf("migration: compatibility version must not be empty")
	}
	_, err := conn.Exec(ctx, `
		INSERT INTO migration.compat_flag (key, version, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (key) DO UPDATE
		SET version = EXCLUDED.version, updated_at = EXCLUDED.updated_at
	`, key, version)
	return err
}

// GetCompatibility reads the current observable compatibility flag.
func GetCompatibility(ctx context.Context, conn *pgx.Conn, key string) (CompatibilityFlag, error) {
	var f CompatibilityFlag
	err := conn.QueryRow(ctx, `
		SELECT key, version, updated_at FROM migration.compat_flag WHERE key = $1
	`, key).Scan(&f.Key, &f.Version, &f.UpdatedAt)
	if err != nil {
		return CompatibilityFlag{}, err
	}
	return f, nil
}

// RollbackAfterSwitch returns the application to previousVersion without
// running a destructive schema Down. It only moves the compatibility flag;
// the expanded schema remains in place so N-1 binaries can read it.
func RollbackAfterSwitch(ctx context.Context, conn *pgx.Conn, key, previousVersion string) error {
	return SetCompatibility(ctx, conn, key, previousVersion)
}
