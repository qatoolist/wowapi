// Command migrate applies the kernel migrations to the database named by
// DATABASE_URL — the framework repo's local/CI migration runner behind
// `make migrate`. Product repositories get a real cmd/migrate process via
// app.RunMigrate (Phase 5+); this private tool exists so the compose stack
// can be migrated without one.
//
// SECURITY: this tool intentionally connects as the raw DATABASE_URL login
// (the local compose superuser) and uses config.Defaults() — acceptable ONLY
// because it is a dev/CI convenience. Product code (app.RunMigrate) MUST use
// the dedicated app_migrate owner DSN from the narrowed MigrateConfig and MUST
// NOT copy this DATABASE_URL/Defaults() shortcut (SEC-15).
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/migrations"
)

func main() {
	// Subcommand: default "up" applies all pending migrations; "reset" rolls the
	// kernel source back to version 0 (goose Down, newest-first). "reset" exists
	// only for local/CI drills — notably the migration reversibility drill
	// (scripts/migration_reversibility_drill.sh, backlog B-4) which needs a
	// shell-drivable down-to-0. Like Up here, it connects as the raw DATABASE_URL
	// login and MUST NEVER be pointed at a production database (see database.MigrateReset).
	mode := "up"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "migrate: DATABASE_URL is not set (run `make up`, or use `make migrate` which supplies the compose default)")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.NewPool(ctx, dsn, config.Defaults().DB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	switch mode {
	case "up":
		res, err := database.Migrate(ctx, pool, migrations.Kernel(), migrations.SourceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("kernel migrations applied (%d this run); %s source at version %d\n",
			res.Applied, migrations.SourceName, res.Version)
	case "reset":
		v, err := database.MigrateReset(ctx, pool, migrations.Kernel(), migrations.SourceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("kernel migrations rolled back; %s source at version %d\n",
			migrations.SourceName, v)
	default:
		fmt.Fprintf(os.Stderr, "migrate: unknown mode %q (want \"up\" or \"reset\")\n", mode)
		os.Exit(2)
	}
}
