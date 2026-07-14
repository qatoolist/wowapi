package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/seeds"
)

// CatalogsSeeded is the GAP-003 "clear failure mode" check: it turns an empty
// authorization/resource catalog into one loud, actionable readiness failure
// instead of the scattered per-request 403s and resource-mirror FK violations
// that would otherwise be the only symptom (docs/upstream PF-9 in the product
// repo this was upstreamed from).
//
// It compares what the booted module seeds DECLARE (b, in-memory — every
// module's Register already ran) against what the DATABASE actually holds. A
// mismatch — seeds declared but the catalog table is empty — means the
// deploy's migrate step never ran seeds.Sync, which is exactly the gap this
// check exists to catch: wire it as a readiness check (see app.Readiness's
// `extra` parameter) on a platform-privileged connection, so an unseeded pod
// never reports ready and never takes traffic.
//
// A product that declares no seeds at all (b is the zero Bundle) has nothing
// to sync, so an empty catalog is expected and not an error.
func CatalogsSeeded(ctx context.Context, db database.DBTX, b seeds.Bundle) error {
	_, err := catalogsSeededState(ctx, db, b)
	return err
}

// catalogsSeededState returns the populated-catalog check result and, when the
// database already records a successful sync run, the latest manifest hash.
func catalogsSeededState(ctx context.Context, db database.DBTX, b seeds.Bundle) (string, error) {
	checks := []struct {
		declared bool
		table    string
	}{
		{len(b.Permissions) > 0, "permissions"},
		{len(b.ResourceTypes) > 0, "resource_types"},
		{len(b.RelationshipTypes) > 0, "relationship_types"},
		{len(b.Roles) > 0, "roles"},
	}
	for _, c := range checks {
		if !c.declared {
			continue
		}
		var n int
		if err := db.QueryRow(ctx, "SELECT count(*) FROM "+pgx.Identifier{c.table}.Sanitize()).Scan(&n); err != nil {
			return "", fmt.Errorf("seed catalog check: query %s: %w", c.table, err)
		}
		if n == 0 {
			return "", fmt.Errorf(
				"seed catalog %q is empty but modules declare %d %s seed(s): "+
					"the database migration ran without a seed sync — run `wowapi seed sync` "+
					"(or your generated `cmd/migrate`, which now runs it automatically) before serving traffic",
				c.table, seedCount(b, c.table), c.table)
		}
	}
	return latestSeedHash(ctx, db)
}

// latestSeedHash returns the most recent successful sync manifest hash, or an
// empty string if no run has been recorded (e.g. a pre-FBL-02 database that was
// populated by seeds.Sync before the audit table existed).
func latestSeedHash(ctx context.Context, db database.DBTX) (string, error) {
	var hash string
	err := db.QueryRow(ctx,
		`SELECT manifest_hash FROM seed_sync_runs
		  WHERE outcome IN ('applied','noop')
		  ORDER BY created_at DESC LIMIT 1`).Scan(&hash)
	if err != nil {
		return "", nil // no row = no hash to report
	}
	return hash, nil
}

// seedCount reports how many entries the bundle declares for the named table,
// for the error message only.
func seedCount(b seeds.Bundle, table string) int {
	switch table {
	case "permissions":
		return len(b.Permissions)
	case "resource_types":
		return len(b.ResourceTypes)
	case "relationship_types":
		return len(b.RelationshipTypes)
	case "roles":
		return len(b.Roles)
	default:
		return 0
	}
}
